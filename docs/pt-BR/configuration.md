# Configuração

O CLI lê chaves `DUMPS_*` do ambiente e de um arquivo de configuração. O `dumpsd` não lê arquivo: recebe tudo por flags.

## Onde o CLI lê a configuração

O CLI lê `$DUMPS_CONFIG`, que por padrão é `${XDG_CONFIG_HOME:-$HOME/.config}/dumps/config`. Comece pelo exemplo:

```bash
mkdir -p ~/.config/dumps
```

```bash
cp examples/config ~/.config/dumps/config
```

Regras:

- Uma chave por linha, no formato `DUMPS_CHAVE=valor`. Linhas que não casam com esse formato (comentários, vazias) são ignoradas.
- Aspas simples ou duplas ao redor do valor são removidas. `$HOME` e um `~/` inicial são expandidos.
- Não há comentário no fim da linha: tudo depois do `=` é valor.
- A variável de ambiente vence o arquivo. Uma chave já definida no ambiente não é sobrescrita pelo arquivo.
- O próprio `DUMPS_CONFIG` só pode vir do ambiente.

Para ver a configuração efetiva:

```bash
cli/dumps config
```

## Chaves do CLI

| Chave | Padrão | Significado |
|-------|--------|-------------|
| `DUMPS_HOST` | `local` | `local` publica na própria máquina. Qualquer outro valor é um host do `~/.ssh/config` |
| `DUMPS_ROOT` | local: `${XDG_DATA_HOME:-~/.local/share}/dumps`; remoto: a home do usuário ssh | Pasta no servidor que guarda `dumps/`, `dumps-meta/` e `bin/`. Não pode conter aspa simples |
| `DUMPS_URL` | local: `http://localhost:8090`; remoto: vazio (obrigatório) | URL base pública, no formato `http(s)://host[:porta]`, sem caminho nem barra final |
| `DUMPS_LIST_URL` | local: `http://localhost:<porta de DUMPS_ADMIN_LISTEN>/_dumps/`; remoto: `$DUMPS_URL/_dumps/` | Onde fica a listagem privada (informativa, aparece no `dumps config`). Num host remoto o admin fica em outro hostname, então defina |
| `DUMPS_AUTHORS` | vazio | Autores do rodapé, `Nome\|URL`, vários separados por `;`. `--author` no `publish` substitui |
| `DUMPS_BLOCKED_PATHS` | vazio | Caminhos separados por `:`. Publicar o que estiver dentro deles, ou uma pasta que os contenha, termina com saída 2. Cada item expande o `~` por conta própria: `~` e `~/...` significam a sua home |
| `DUMPS_SITE` | vazio | Caminho de um `site.json`, enviado ao servidor no `install-dumpsd` (veja abaixo) |
| `DUMPS_SERVICE` | `none` | Serviço do admin a instalar: `none`, `termux-runit`, `systemd-user` ou `launchd` |
| `DUMPS_TARGET` | `linux/arm64` | `GOOS/GOARCH` do servidor remoto para o `install-dumpsd`. Ignorado no modo local (usa a própria máquina) |
| `DUMPS_LISTEN` | `127.0.0.1:8090` | Endereço do `dumpsd serve`. Precisa ser loopback. No `serve-local` é o listener dos dumps |
| `DUMPS_ADMIN_LISTEN` | `127.0.0.1:8091` | Listener do admin no `dumps serve-local`. Precisa ser loopback e numa porta diferente de `DUMPS_LISTEN`. Não é usado por serviço |
| `DUMPS_ADMIN_ORIGIN` | vazio | Origem exata da página do admin (por exemplo `https://admin.example.com`). Obrigatória para instalar um serviço. O `serve-local` usa `http://localhost:<porta de DUMPS_ADMIN_LISTEN>` quando está vazia |
| `DUMPS_AUTH` | `access` | Modo de login do serviço instalado: `access` ou `tailscale`. Veja [auth.md](auth.md). `none` é recusado para serviço, e o `serve-local` sempre usa `none` por conta própria |
| `DUMPS_ACCESS_TEAM` | vazio | Domínio do time no Cloudflare (por exemplo `seutime.cloudflareaccess.com`). Obrigatório com `access` |
| `DUMPS_ACCESS_AUD` | vazio | Tag Application Audience (AUD) do app do admin no Access. Obrigatória com `access` |
| `DUMPS_TAILSCALE_ALLOW` | vazio | Logins permitidos com `tailscale`, separados por espaço. Vazio aceita qualquer login do tailnet |
| `DUMPS_GA_ID` | vazio | ID de medição do GA4 (`G-XXXXXXX`). Vazio desliga. Qualquer outro valor é recusado. Começa com consentimento negado e mostra um banner de aceite (ver security.md) |
| `DUMPS_MACHINE` | `hostname -s` | Nome gravado no manifest e mostrado na listagem |
| `DUMPS_CHROME` | detecção automática | Caminho do Chrome ou Chromium usado nas miniaturas |

Os valores que vão para o serviço (a raiz resolvida, `DUMPS_LISTEN`, `DUMPS_URL`, `DUMPS_ADMIN_ORIGIN`, `DUMPS_ACCESS_TEAM`, `DUMPS_ACCESS_AUD` e cada login de `DUMPS_TAILSCALE_ALLOW`) só podem ter letras, números e `: / . _ @ -`. Espaços, aspas e qualquer outro caractere fazem o `install-dumpsd` parar.

O CLI depende de `bash`, `jq`, `rsync`, `python3` e `ssh` (para hosts remotos). O `go` é necessário para o `install-dumpsd`.

### Comandos do CLI

```text
dumps publish <arquivo.html|pasta> [--id <id>] [--title <texto>] [--author "Nome|URL"]... [--no-bar] [--no-footer] [--hidden] [--source cli|agent]
dumps list
dumps archive <id>
dumps unarchive <id>
dumps edit <id> --title <texto>
dumps rm <id>
dumps regen
dumps install-dumpsd
dumps serve-local
dumps config
dumps install-agent [claude|codex|all]
```

| Comando | O que faz |
|---------|-----------|
| `publish` | Publica um arquivo ou uma pasta. Exige `prepare` e `activate` no binário do servidor |
| `list` | Lista os dumps por data de atualização |
| `archive`, `unarchive` | Tira um dump da home sem tirar do ar, ou devolve |
| `edit` | Troca o título na home e na listagem. A barra dentro do HTML só muda no próximo publish |
| `rm` | Apaga o conteúdo, o manifest e a miniatura |
| `regen` | Refaz a home e a listagem a partir dos manifests |
| `install-dumpsd` | Roda os testes de Go, compila o `dumpsd` para o alvo, instala em `$DUMPS_ROOT/bin/dumpsd` (o anterior fica se o novo não rodar), envia o `DUMPS_SITE` se houver e instala ou atualiza o serviço |
| `serve-local` | Serve os dumps em `DUMPS_LISTEN` e o admin em `DUMPS_ADMIN_LISTEN`. Só com `DUMPS_HOST=local`. Veja [local.md](local.md) |
| `config` | Mostra a configuração efetiva: arquivo de config, host, raiz, URLs, serviço, alvo, auth, listeners, origem do admin, site, caminhos bloqueados, autores, máquina e GA |
| `install-agent` | Cria os symlinks da skill `dump-this` e do comando `/dump-this` para Claude Code, Codex ou os dois (padrão `all`). Veja [agent.md](agent.md) |

Notas sobre o `publish`:

- `--hidden` publica já arquivado: o link funciona, mas o dump fica fora da home.
- `--source` aceita `cli` ou `agent` e é gravado como `source` no manifest. Sem a flag, republicar mantém a origem anterior, e um dump novo recebe `cli`.
- `--title` usa por padrão o `<title>` do HTML e, na falta, o nome do arquivo. Um título definido com `--title` ou `edit` é mantido nas publicações seguintes.
- Ids casam com `^[a-z0-9][a-z0-9-]{0,62}$`.

Os códigos de saída estão em [architecture.md](architecture.md#códigos-de-saída-do-cli).

## site.json

O `site.json` personaliza a home. Aponte `DUMPS_SITE` para ele e rode `dumps install-dumpsd`. O arquivo é gravado como `dumps-meta/.site.json` e a home é refeita. Para mudar os textos depois, edite o arquivo e rode `install-dumpsd` de novo: o `dumps regen` não reenvia o arquivo. Veja `examples/site.json`.

O arquivo precisa ser um objeto JSON. Campos desconhecidos são recusados. Todos os campos são opcionais.

| Campo | Padrão | Significado |
|-------|--------|-------------|
| `title` | host de `DUMPS_URL` | Título da página e marca |
| `description` | `Artifacts published with dumps.` | Meta description |
| `tagline_html` | `description`, com HTML escapado | Frase sob o título (HTML cru) |
| `host_card_html` | vazio | Bloco opcional perto do fim da home (HTML cru) |
| `footer_html` | vazio | Acrescentado ao rodapé da home, depois do link do host (HTML cru) |
| `featured` | vazio | Links em destaque no topo da home: `[{ "id": "...", "label": "...", "note": "..." }]` |
| `listing_dir` | `admin` | Pasta, dentro da raiz, onde a listagem privada é gravada |

Como o `featured` funciona:

- Cada id listado que tem um dump vira um bloco de link no topo da home, com o rótulo, o título do dump e a nota opcional. A busca cobre todos os manifests, inclusive os dumps arquivados.
- `label` vazio vira `Em destaque`. `note` é opcional. Os dois têm o HTML escapado.
- Um id sem dump correspondente é ignorado. Um id que não é um id de dump válido invalida o arquivo inteiro.

Como o `listing_dir` é validado:

- Precisa ser um caminho relativo dentro da raiz. Não pode ser absoluto, ser `.` nem começar com `..`.
- O primeiro componente, comparado sem diferenciar maiúsculas, não pode ser `dumps`, `dumps-meta`, `dumps-staging` nem `bin`, e não pode começar com ponto.
- Mantenha fora da pasta pública. Se mudar, mude também o `root` do bloco do admin no Caddyfile.

Como o arquivo é enviado: o `install-dumpsd` confere que é um objeto JSON, envia e renderiza. Se o `dumpsd` recusar (campo desconhecido, `listing_dir` inválido, id inválido em `featured`), o `site.json` anterior continua no lugar, ou nenhum, se não havia, e o CLI termina com saída 4 e a mensagem `site.json recusado pelo dumpsd; o anterior foi mantido`.

Aviso: `tagline_html`, `host_card_html` e `footer_html` entram na home como HTML cru e confiável. Não são sanitizados. Coloque só conteúdo que você escreveu.

## dumpsd

O `dumpsd` fica instalado em `$DUMPS_ROOT/bin/dumpsd`. Erros de uso saem com 1 e falhas saem com 4.

`render`, `apply`, `serve` e `activate` aceitam `--root <dir>`. Sem a flag, o `dumpsd` usa o diretório atual só se ele tiver `dumps-meta/`. Caso contrário, falha e pede para rodar dentro da raiz ou passar `--root`. O `prepare` trabalha na pasta de `--dir` e não tem `--root`.

### version e capabilities

```bash
bin/dumpsd version
```

```bash
bin/dumpsd capabilities
```

`capabilities` imprime `render serve apply prepare activate`. O CLI usa isso para detectar um binário defasado no servidor: o `dumps publish` exige `prepare` e `activate`, e termina com saída 4 e a dica de rodar `dumps install-dumpsd` quando falta um deles.

### render

```text
dumpsd render --base-url <url> [--ga] [--no-lock] [--root <dir>]
```

Refaz `dumps/index.html` e a listagem a partir dos manifests, sob o lock.

- `--base-url`: obrigatória.
- `--ga`: faz a home carregar o `dumps-bar.js`, para o GA4 cobrir a home também. Liga sozinha quando a barra instalada já contém o snippet do GA4.
- `--no-lock`: pula o lock. Para quem já o segura. Não use no dia a dia.

### prepare

```text
dumpsd prepare --dir <pasta> [--manifest <arquivo>]
```

Lê uma especificação JSON do stdin (até 64 KiB) e injeta barra, rodapé e tags Open Graph em cada arquivo HTML de `--dir`. Com `--manifest`, grava também o tamanho total no manifest. Todos os campos da especificação são obrigatórios e campos desconhecidos são recusados: `title`, `url`, `version`, `authors` (um array JSON codificado como string), `bar`, `footer`, `og_image`, `og_authors`. O CLI monta isso por você.

### activate

```text
dumpsd activate --base-url <url> [--ga] [--root <dir>] <id>
```

Ativa uma publicação preparada. Sob o lock, faz backup do conteúdo, do manifest, da barra e da miniatura atuais, troca `dumps-staging/<id>` por `dumps/<id>`, move manifest, barra e miniatura para o lugar e refaz a home e a listagem. Se qualquer passo falhar, restaura os backups e sai com 4. Precisa de `dumps-staging/<id>` e do manifest temporário `dumps-meta/.<id>.json.tmp`, senão falha com `staging incompleto`.

O CLI chama esse subcomando por ssh (ou localmente), então o servidor não precisa do comando `flock`: o lock é do próprio `dumpsd`, e isso funciona no macOS. Veja [architecture.md](architecture.md#fluxo-do-publish).

### apply

```text
dumpsd apply --base-url <url> [--root <dir>] <archive|unarchive|title|delete> <id>
```

Executa uma ação do admin pelo shell. `title` lê o novo título do stdin. Além de 0, 1 e 4, sai com 5 quando a mudança foi aplicada mas a home e a listagem não foram regeneradas (rode `dumps regen`), e com 6 quando o dump não existe. O `delete` também remove uma pasta órfã sem manifest.

### serve

```text
dumpsd serve --base-url <url> --origin <url> [--listen <addr>] [--admin-listen <addr>] [--root <dir>] [--auth access|tailscale|none] [--team <host>] [--aud <tag>] [--allow <login>]... [--static]
```

Executa a API HTTP do admin.

| Flag | Padrão | Significado |
|------|--------|-------------|
| `--base-url` | obrigatória | URL pública |
| `--origin` | obrigatória | Origem exata aceita nos `POST`s: a origem da página do admin |
| `--listen` | `127.0.0.1:8090` | Precisa começar com `127.0.0.1:`, `localhost:` ou `[::1]:` |
| `--admin-listen` | vazio | Listener do admin, só loopback. Obrigatória com `--static`, numa porta diferente de `--listen` |
| `--root` | diretório atual com `dumps-meta/` | Raiz dos dumps |
| `--auth` | `access` | `access` exige `--team` e `--aud`. `tailscale` aceita `--allow` repetida, opcional. `none` é para o modo local. Veja [auth.md](auth.md) |
| `--team` | vazio | Host do time no Access (obrigatório com `access`) |
| `--aud` | vazio | AUD do app no Access (obrigatório com `access`) |
| `--allow` | vazio | Login permitido com `tailscale`. Repita a flag para vários |
| `--static` | desligada | Serve o site público e o admin por este processo. Só com `--auth none` e `--admin-listen`. É o que o `dumps serve-local` usa |

Sem `--static`, um listener serve a API em `/_dumps/api/*` e o `/healthz`, e qualquer outro caminho devolve 404. Em produção, quem serve os arquivos é o proxy.

Com `--static`, são dois listeners. O `--listen` serve só `dumps/`, e o `--admin-listen` serve a listagem em `/_dumps/` e a API. A API só aceita a origem do admin.

O `GET /healthz` responde `ok` sem login.

## Makefile

| Alvo | O que faz |
|------|-----------|
| `make test` | `go test -race ./...` em `server/` |
| `make lint` | gofmt, `go vet` e shellcheck |
| `make build` | Compila o `dumpsd` para esta máquina em `bin/` |
| `make check` | `lint` e `test`, o mesmo do CI |
