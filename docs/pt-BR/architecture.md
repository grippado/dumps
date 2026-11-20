# Arquitetura

O CLI roda na máquina do autor e só envia HTML cru. Tudo que altera o estado do servidor (injetar barra, ativar, montar home e listagem) é feito pelo binário `dumpsd`, que é o único escritor.

## Componentes

| Componente | Onde roda | O que faz |
|---|---|---|
| `cli/dumps` (bash) | máquina do autor | valida a origem, gera a miniatura com Chrome headless, envia por ssh/rsync (ou copia local com `DUMPS_HOST=local`) |
| `dumpsd` (Go) | servidor | `prepare`, `activate`, `render`, `apply` e `serve` (veja [configuration.md](configuration.md)) |
| `web/dumps-bar.js` | navegador do leitor | define os elementos `<dumps-bar>` e `<dumps-footer>` (compartilhar, modo limpo, autores) |
| `web/dumps-ga.js` | navegador do leitor | anexado à barra quando há `DUMPS_GA_ID`; carrega o GA4 e registra `share` e `dumps_clean_mode` |
| Proxy (Caddy) | servidor | serve os arquivos estáticos; exemplo em `deploy/Caddyfile` |
| Túnel (cloudflared) | servidor | expõe os dois listeners do Caddy; exemplo em `deploy/cloudflared.yml` |

Em produção, quem serve os arquivos é o proxy. O `dumpsd serve` só atende a API do admin (`/_dumps/api/*`) e `/healthz`. No modo local, `dumps serve-local` liga `--static` e o `dumpsd` serve tudo no loopback.

## Layout da raiz no servidor

A raiz é `DUMPS_ROOT` (padrão: a home do usuário ssh; no modo local, `~/.local/share/dumps`).

```
<raiz>/
  bin/dumpsd                  binário instalado por `dumps install-dumpsd`
  dumps/                      o que o host público serve
    index.html                home
    _/dumps-bar.js            barra, rodapé (e GA4, se ligado)
    _/thumbs/<id>.jpg         miniaturas
    <id>/                     conteúdo de cada dump
  dumps-meta/
    <id>.json                 manifest de cada dump
    .site.json                textos do site (de DUMPS_SITE)
    .render.lock              lock de escrita
  dumps-staging/              área de preparo da publicação em curso
  admin/                      listagem privada (nome vem de listing_dir)
```

O id `_` é reservado: ids precisam casar com `^[a-z0-9][a-z0-9-]{0,62}$`.

Cada manifest tem: `v`, `id`, `title`, `kind` (`file` ou `dir`), `entry`, `source_name`, `bytes`, `created`, `updated`, `url`, `machine`, `bar`, `footer`, `authors`, `thumb`, `archived` e `title_locked`.

## Fluxo do publish

`dumps publish <arquivo|pasta>`:

1. **Validação local.** A origem precisa existir e ser `.html`/`.htm` ou uma pasta com `index.html`. `DUMPS_BLOCKED_PATHS` bloqueia (saída 2). Symlinks que saem da pasta são recusados.
2. **Preflight.** Testa o ssh (saída 3 se falhar), pede `bin/dumpsd version` e `bin/dumpsd capabilities` (precisa de `prepare` e `activate`, senão saída 4 com a dica de rodar `dumps install-dumpsd`).
3. **Id.** Usa `--id` ou gera 10 caracteres (`a-z2-7`) e confere que estão livres (3 tentativas). Republicar com o mesmo id substitui o dump e preserva `created`, o estado de arquivado e o título renomeado.
4. **Preparo local.** Copia o conteúdo para um diretório temporário (sem `.DS_Store` e `.git/`, symlinks resolvidos), monta o `dumps-bar.js` (com o snippet do GA4 se houver `DUMPS_GA_ID`) e gera a miniatura. Sem miniatura, a publicação segue e a home usa um placeholder.
5. **Staging.** Envia o conteúdo para `dumps-staging/<id>/`, a barra para `dumps-staging/<id>.bar.tmp`, a miniatura para `dumps-staging/<id>.thumb.tmp` e o manifest para `dumps-meta/.<id>.json.tmp`. Nada disso está no ar ainda.
6. **`dumpsd prepare`.** Lê uma especificação JSON do stdin e, em cada `.html`/`.htm` do staging, injeta `<dumps-bar>` após o `<body>` (na falta dele, depois de `</head>`, de `<html>` ou do doctype), `<dumps-footer>` antes do `</body>` e as metatags Open Graph e Twitter que ainda não existirem. Marcações já presentes não são duplicadas. Atualiza `bytes` no manifest.
7. **`dumpsd activate`.** Sob o lock: faz backup do estado anterior (conteúdo, manifest, barra, miniatura), troca o staging pelo conteúdo ativo, move manifest, barra e miniatura e chama o render da home e da listagem.
8. **Rollback.** Se qualquer passo da ativação falhar, o `activate` restaura tudo do backup, refaz o render e sai com 4. Se o publish falhar antes da ativação, o CLI apaga o staging e avisa que nada mudou no ar.
9. **Resultado.** O CLI imprime `<DUMPS_URL>/<id>/`.

## Escritor único e lock

Home (`dumps/index.html`), listagem e manifests só são escritos pelo `dumpsd`. `render`, `activate`, `apply` e a API do `serve` pegam um `flock` em `dumps-meta/.render.lock` e esperam até 60 segundos. Os arquivos são gravados em temporário, sincronizados e movidos com `rename`, então o leitor nunca vê uma página pela metade.

A flag `render --no-lock` existe para quem já segura o lock; não use no dia a dia.

## Fluxo do admin

Pelo CLI: `dumps archive|unarchive|edit|rm <id>` roda `dumpsd apply` no servidor por ssh. Não passa pelo login, porque quem chega ali já tem ssh.

Pelo browser: a listagem (HTML estático em `admin/`) chama `POST /_dumps/api/<archive|unarchive|title|delete>` com `{"id": "..."}` (e `"title"` no caso de `title`). O `dumpsd serve` autentica, confere método, `Origin`, tipo e tamanho do corpo ([auth.md](auth.md)), pega o lock, altera o manifest e refaz a home e a listagem. Se o render falhar, o manifest anterior é restaurado. Apagar remove a pasta do dump, o manifest e a miniatura.

## O que roda onde

| Etapa | Máquina |
|---|---|
| Miniatura (Chrome/Chromium) | autor |
| Montagem da barra e do `spec` JSON | autor |
| Envio (rsync sobre ssh) | autor para servidor |
| `prepare`, `activate`, `render`, `apply` | servidor |
| `serve` (API do admin) | servidor, só em loopback |
| Servir `dumps/` e `admin/` | proxy no servidor |

O servidor precisa de `rsync` e de `bin/dumpsd`. Ele não precisa de Go, Chrome nem jq.

## Códigos de saída do CLI

| Código | Significado |
|---|---|
| 0 | ok |
| 1 | uso ou validação |
| 2 | caminho bloqueado (`DUMPS_BLOCKED_PATHS`) |
| 3 | servidor inalcançável |
| 4 | falha remota |

O `dumpsd apply` ainda usa 5 (ação aplicada, mas home e listagem não foram regeneradas; rode `dumps regen`) e 6 (dump não existe). O CLI traduz esses dois em mensagens.
