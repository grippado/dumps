# Arquitetura

O dumps tem três partes: um CLI em bash que roda na sua máquina, um binário Go (`dumpsd`) que roda no servidor e dois scripts de navegador que o `dumpsd` injeta nas páginas publicadas. O CLI só envia HTML cru. Tudo que altera o estado do servidor (injetar a barra, ativar, montar home e listagem) é feito pelo `dumpsd`, o único escritor.

## Componentes

| Parte | Onde | O que faz |
|-------|------|-----------|
| `cli/dumps` | sua máquina | Valida a origem, gera a miniatura, envia o HTML cru por ssh/rsync (ou copia local) e chama o `dumpsd` |
| `dumpsd` | servidor | `prepare` injeta barra, rodapé e Open Graph; `activate` troca um dump de forma atômica; `render` grava home e listagem; `apply` executa ações do admin pelo CLI; `serve` é a API HTTP do admin (veja [configuration.md](configuration.md)) |
| `web/dumps-bar.js` | navegador | Define `<dumps-bar>` (compartilhar, modo limpo) e `<dumps-footer>` (autores) |
| `web/dumps-ga.js` | navegador | Anexado à barra quando há `DUMPS_GA_ID`; carrega o GA4 e registra `share` e `dumps_clean_mode` |
| `agents/` | sua máquina | A skill `dump-this` e o comando `/dump-this`, ligados pelo `dumps install-agent` ([agent.md](agent.md)) |
| `site/` | GitHub Pages | A landing page do projeto, publicada em https://grippado.github.io/dumps/ |
| Proxy reverso | servidor | Serve `dumps/` ao público (exemplo de Caddy em `deploy/Caddyfile`) e o listener do admin |
| Túnel | servidor | Expõe o proxy na internet (exemplo de cloudflared em `deploy/cloudflared.yml`) |

Em produção, quem serve os arquivos é o proxy, e o `dumpsd serve` só atende a API do admin (`/_dumps/api/*`) e o `/healthz`. No modo local, o `dumps serve-local` liga `--static` e o `dumpsd` serve tudo no loopback, com os dumps e o admin em duas origens diferentes ([local.md](local.md)).

O CLI só envia para `dumps-staging/` e para arquivos temporários em `dumps-meta/`. O `dumpsd` é o único escritor da pasta pública `dumps/`, dos manifests e da listagem.

## Layout da raiz no servidor

Todos os caminhos são relativos a `DUMPS_ROOT` (padrão: a home do usuário ssh; no modo local, `~/.local/share/dumps`).

```text
bin/dumpsd                  o binário do servidor
dumps/                      pasta pública
  index.html                home (gravada pelo render)
  <id>/                     uma pasta por dump
  _/dumps-bar.js            barra, rodapé e GA4 opcional, compartilhados por todos os dumps
  _/thumbs/<id>.jpg         miniaturas
dumps-meta/
  <id>.json                 manifest de cada dump
  .site.json                configuração opcional do site (veja configuration.md)
  .render.lock              arquivo do flock
dumps-staging/              envios antes da ativação
admin/                      listagem privada (index.html), nome definido por listing_dir
```

O proxy público serve só `dumps/`. Os manifests, a área de staging e a pasta da listagem nunca ficam dentro dela. A pasta `_` guarda arquivos compartilhados e não pode ser um id de dump, porque ids precisam começar com letra ou número.

Um manifest tem este formato:

```json
{
  "v": 1,
  "id": "abc",
  "title": "My page",
  "kind": "file",
  "entry": "index.html",
  "source_name": "page.html",
  "bytes": 2048,
  "created": "2026-01-01T10:00:00-0300",
  "updated": "2026-01-01T10:00:00-0300",
  "url": "https://dumps.example.com/abc/",
  "machine": "laptop",
  "bar": true,
  "footer": true,
  "authors": [{ "name": "Ada Lovelace", "url": "https://example.com" }],
  "thumb": true,
  "archived": false,
  "title_locked": false,
  "source": "cli"
}
```

`kind` é `file` ou `dir`. `source` é `cli` ou `agent`: `dumps publish --source agent` grava `agent`, e republicar sem a flag mantém o valor anterior. Manifests gravados antes de esse campo existir não têm `source`.

## Fluxo do publish

`dumps publish <arquivo|pasta>` faz o seguinte.

1. **Validação local.** A origem precisa existir e ser um arquivo `.html`/`.htm` ou uma pasta com `index.html`, não pode estar dentro de (nem conter) um caminho de `DUMPS_BLOCKED_PATHS`, e symlinks de uma pasta precisam ficar dentro dela. O id precisa casar com `^[a-z0-9][a-z0-9-]{0,62}$`.
2. **Preflight.** Confere `DUMPS_URL`, alcança o servidor (saída 3 se não), roda `bin/dumpsd version` (saída 4 se não existir, aviso se diferir do repositório) e roda `bin/dumpsd capabilities`. `prepare` e `activate` são obrigatórios, senão saída 4 com a dica de rodar `dumps install-dumpsd`.
3. **Id.** Usa `--id` ou gera 10 caracteres aleatórios de `a-z2-7` e tenta até 3 vezes se o id estiver ocupado. Republicar com o mesmo id substitui o dump e mantém `created`, o estado de arquivado, o título renomeado e o `source`.
4. **Montar o pacote.** Copia o conteúdo para um diretório temporário (`rsync -aL`, sem `.DS_Store` e `.git/`), monta o `dumps-bar.js` (com o snippet do GA4 se houver `DUMPS_GA_ID`) e gera uma miniatura JPEG de 640 px com o Chrome headless. Se a miniatura falhar, o CLI avisa e a home usa um placeholder.
5. **Staging.** Limpa `dumps-staging/<id>` no servidor e envia o conteúdo para `dumps-staging/<id>/`, a barra para `dumps-staging/<id>.bar.tmp`, o manifest para `dumps-meta/.<id>.json.tmp` e a miniatura para `dumps-staging/<id>.thumb.tmp`. Nada público muda ainda.
6. **Prepare.** O `dumpsd prepare` lê uma especificação JSON do stdin e edita os arquivos em staging: injeta `<dumps-bar>` e a tag do script depois do `<body>`, `<dumps-footer>` antes do `</body>`, e as metatags Open Graph e Twitter card no `<head>`, em cada arquivo `.html`/`.htm` (arquivos UTF-16 são ignorados, tags já existentes são respeitadas). Também grava no manifest o total de bytes.
7. **Activate.** O CLI executa `dumpsd activate <id>`. Sob o lock, faz backup do conteúdo, do manifest, da barra e da miniatura atuais (`dumps-staging/<id>.old` e afins), renomeia o staging para `dumps/<id>`, move manifest, barra e miniatura para o lugar e refaz a home e a listagem. O lock é do próprio `dumpsd`, então o servidor não precisa do comando `flock` e um servidor macOS funciona.
8. **Rollback.** Se qualquer passo da troca ou do render falhar, o `activate` restaura os backups, remove as sobras do staging, refaz o render e sai com 4 ("versão anterior restaurada"). Em caso de sucesso, os backups são apagados.
9. **Resultado.** O CLI imprime a URL do dump: `<DUMPS_URL>/<id>/`.

Se o CLI falhar ou for interrompido antes da ativação, um trap apaga os arquivos do staging e avisa que nada mudou. Se a conexão cair durante a ativação, o CLI sai com 4 em estado incerto e pede para conferir com `dumps list`.

## Escritor único e lock

`render`, `activate` e toda ação do admin (`archive`, `unarchive`, `title`, `delete`) pegam um `flock` exclusivo em `dumps-meta/.render.lock` e esperam até 60 segundos. Isso serializa publicações e cliques de admin concorrentes. Home e listagem são gravadas num arquivo temporário na pasta de destino, sincronizadas e renomeadas, então o leitor nunca vê uma página pela metade. O `dumpsd` ignora SIGHUP, para que uma sessão ssh derrubada não o mate no meio da escrita.

O `render --no-lock` existe para quem já segura o lock. Não use no dia a dia.

Como usa `flock`, o `dumpsd` mira sistemas Unix (Linux, macOS, Android).

## Fluxo do admin

A página do admin é a listagem privada em `<DUMPS_LIST_URL>`. Em produção o padrão é `<DUMPS_URL>/_dumps/`, a menos que você defina `DUMPS_LIST_URL` com o host do admin. No modo local é `http://localhost:8091/_dumps/`. Os botões chamam `POST /_dumps/api/<archive|unarchive|title|delete>` com `{"id": "..."}` (mais `"title"` no caso de `title`).

1. O navegador chega ao host do admin. Sua camada de login (Cloudflare Access, Tailscale) autentica o usuário.
2. O proxy encaminha `/_dumps/api/*` para o `dumpsd serve` em `127.0.0.1:8090` e serve `/_dumps/*` a partir da pasta da listagem.
3. O `dumpsd` confere a identidade, o header `Origin`, o tipo de conteúdo e o corpo ([auth.md](auth.md)), depois altera o manifest sob o lock e refaz o render. Se o render falhar, o manifest anterior é restaurado.

No modo local não há proxy: o `dumpsd serve --auth none --static` tem dois listeners, os dumps em `DUMPS_LISTEN` e o admin (listagem e API) em `DUMPS_ADMIN_LISTEN`. A API só aceita a origem do admin, então o JavaScript de um dump não consegue chamá-la.

Pelo terminal, `dumps archive|unarchive|edit|rm <id>` executa o `dumpsd apply` no servidor por ssh, em vez da API HTTP. Não passa pelo login, porque quem chega ali já tem ssh.

`archive` tira um dump da home mas mantém o link no ar. `delete` remove o conteúdo, o manifest e a miniatura.

Na listagem, um dump publicado com `--source agent` mostra o selo **do agente**. Se ele também estiver arquivado, o botão de alternância diz **Publicar** (em destaque) em vez de **Desarquivar**, e chama a mesma ação `unarchive`. Veja [agent.md](agent.md).

## O que roda onde

| Onde | O quê |
|------|-------|
| Sua máquina | `cli/dumps`, Chrome headless (miniaturas), a montagem da barra e da especificação JSON, `go build` do `dumpsd` (o `dumps install-dumpsd` compila para `DUMPS_TARGET`) |
| Sua máquina para o servidor | O envio (rsync sobre ssh) |
| Servidor | `dumpsd` (`prepare`, `activate`, `render`, `apply` como comandos curtos por ssh, mais o serviço `serve`), proxy reverso, túnel |
| Navegador | `dumps-bar.js` e o script da página do admin |
| Modo local | Tudo isso numa só máquina; o `dumpsd serve --auth none --static` substitui o proxy e o túnel |

O servidor precisa de `rsync` e de `bin/dumpsd`. Ele não precisa de Go, Chrome nem jq.

## Códigos de saída do CLI

| Código | Significado |
|--------|-------------|
| 0 | ok |
| 1 | Erro de uso ou de validação |
| 2 | Caminho bloqueado (`DUMPS_BLOCKED_PATHS`) |
| 3 | Servidor inalcançável |
| 4 | Falha remota |

O `dumpsd apply` ainda usa 5 (ação aplicada, mas home e listagem não foram regeneradas; rode `dumps regen`) e 6 (o dump não existe). O CLI traduz os dois em mensagens.
