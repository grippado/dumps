# Configuração

O CLI lê chaves `DUMPS_*` do ambiente e de um arquivo de configuração. O `dumpsd` não lê arquivo: recebe tudo por flags.

## Arquivo de configuração

Local padrão: `$XDG_CONFIG_HOME/dumps/config`, ou `~/.config/dumps/config`. Para usar outro caminho, defina `DUMPS_CONFIG` no ambiente (essa chave não vale dentro do arquivo).

```bash
mkdir -p ~/.config/dumps
```

```bash
cp examples/config ~/.config/dumps/config
```

Regras do arquivo:

- Uma chave por linha, no formato `DUMPS_CHAVE=valor`. Linhas que não casam com esse formato (comentários, vazias) são ignoradas.
- Precedência: **variável de ambiente vence o arquivo**.
- Aspas simples ou duplas ao redor do valor são removidas. `$HOME` e um `~/` inicial são expandidos.
- Não há comentário no fim da linha: tudo depois do `=` é valor.

Para ver o resultado final:

```bash
cli/dumps config
```

## Chaves do CLI

| Chave | Padrão | Significado |
|---|---|---|
| `DUMPS_HOST` | `local` | `local` publica na própria máquina, sem ssh. Qualquer outro valor é um host do `~/.ssh/config` |
| `DUMPS_ROOT` | local: `~/.local/share/dumps` (ou `$XDG_DATA_HOME/dumps`); remoto: a home do usuário ssh | raiz do dumps no servidor. Não pode conter aspa simples |
| `DUMPS_URL` | local: `http://localhost:8090`; remoto: vazio (obrigatório) | URL base pública, no formato `http(s)://host[:porta]`, sem caminho |
| `DUMPS_LIST_URL` | `$DUMPS_URL/_dumps/` | URL da listagem privada (aparece no `dumps config` e na ajuda) |
| `DUMPS_AUTHORS` | vazio | autores do rodapé: `Nome\|URL`, vários separados por `;`. `--author` no `publish` substitui |
| `DUMPS_BLOCKED_PATHS` | vazio | caminhos separados por `:`; recusa publicar o que estiver dentro deles ou os contiver (saída 2) |
| `DUMPS_SITE` | vazio | caminho local de um `site.json`; enviado no `install-dumpsd` (veja abaixo) |
| `DUMPS_SERVICE` | `none` | serviço do admin: `none`, `termux-runit`, `systemd-user` ou `launchd` |
| `DUMPS_TARGET` | `linux/arm64` | `GOOS/GOARCH` do binário para servidor remoto. No modo local o CLI usa o da própria máquina |
| `DUMPS_LISTEN` | `127.0.0.1:8090` | endereço do `dumpsd serve`; precisa ser loopback |
| `DUMPS_ADMIN_ORIGIN` | vazio | origem exata de onde a listagem é servida (ex.: `https://admin.example.com`). Obrigatório para instalar o serviço |
| `DUMPS_AUTH` | `access` | `access`, `tailscale` ou `none` ([auth.md](auth.md)) |
| `DUMPS_ACCESS_TEAM` | vazio | domínio do time no Access, sem `https://` (ex.: `seutime.cloudflareaccess.com`). Obrigatório com `access` |
| `DUMPS_ACCESS_AUD` | vazio | Application Audience (AUD) do app do admin. Obrigatório com `access` |
| `DUMPS_TAILSCALE_ALLOW` | vazio | logins permitidos com `tailscale`, separados por espaço. Vazio aceita qualquer login do tailnet |
| `DUMPS_GA_ID` | vazio | ID de medição do GA4 (`G-XXXXXXX`). Vazio desliga. ID fora do formato é erro |
| `DUMPS_MACHINE` | `hostname -s` | nome da máquina gravado no manifest (coluna da listagem) |
| `DUMPS_CHROME` | vazio | caminho do Chrome/Chromium para a miniatura, antes da busca automática |

Os valores que vão para o serviço (`DUMPS_ADMIN_ORIGIN`, `DUMPS_LISTEN`, team, AUD, allow) só podem ter letras, números, espaço e `: / . _ @ -`; qualquer outro caractere aborta a instalação.

## site.json

Caminho no cliente: `DUMPS_SITE`. No servidor ele vira `dumps-meta/.site.json`. O `dumps install-dumpsd` envia o arquivo e refaz a home. Para mudar o texto do site, edite o arquivo e rode `dumps install-dumpsd` de novo (o `dumps regen` não reenvia o `site.json`).

O arquivo precisa ser um objeto JSON, e campos desconhecidos são recusados. Veja `examples/site.json`.

| Campo | Padrão | Uso |
|---|---|---|
| `title` | host de `DUMPS_URL` | título da página inicial (`<title>` e `og:title`) |
| `description` | `Artifacts published with dumps.` | meta description e `og:description` |
| `tagline_html` | `description` com HTML escapado | frase sob o nome do site na home |
| `host_card_html` | vazio | bloco opcional no fim da home |
| `footer_html` | vazio | trecho no rodapé da home, ao lado do link para a raiz |
| `featured` | vazio | lista de `{id, label, note}` com links em destaque no topo da home |
| `listing_dir` | `admin` | pasta (relativa à raiz) onde a listagem é gravada |

Detalhes:

- `featured`: `id` precisa ser um id válido. Ids que ainda não foram publicados são ignorados. `label` vazio vira `Em destaque`; `note` é opcional. `label` e `note` são escapados.
- `listing_dir`: caminho relativo e dentro da raiz. O `dumpsd` recusa caminho absoluto, `.`, qualquer coisa começando com `..`, `dumps`, `dumps/...` e `dumps-meta`. Use um nome próprio (como `admin`) e não aponte para `bin` nem `dumps-staging`. Se mudar, ajuste também o `root` do bloco do admin no Caddy.

**Aviso:** `tagline_html`, `host_card_html` e `footer_html` são HTML confiável, inserido cru na home pública. Quem controla o `site.json` controla o HTML (inclusive scripts) da sua home. Não monte esses campos a partir de entrada de terceiros.

## Subcomandos do CLI

| Comando | Faz |
|---|---|
| `dumps publish <arquivo.html\|pasta> [--id <id>] [--title <texto>] [--author "Nome\|URL"]... [--no-bar] [--no-footer] [--hidden]` | publica. `--hidden` publica já arquivado (o link funciona, fora da home) |
| `dumps list` | lista os dumps por data de atualização |
| `dumps archive <id>` | tira da home sem tirar do ar |
| `dumps unarchive <id>` | devolve à home |
| `dumps edit <id> --title <texto>` | troca o título na home e na listagem; a barra do HTML só muda no próximo `publish` |
| `dumps rm <id>` | apaga conteúdo, manifest e miniatura |
| `dumps regen` | refaz home e listagem a partir dos manifests |
| `dumps install-dumpsd` | roda os testes, compila o `dumpsd`, instala no servidor, envia o `site.json` e atualiza o serviço |
| `dumps serve-local` | serve home, dumps e admin em `DUMPS_LISTEN` (só com `DUMPS_HOST=local`) |
| `dumps config` | mostra a configuração efetiva |

Ids: `^[a-z0-9][a-z0-9-]{0,62}$`. Sem `--title`, o título vem do `<title>` do HTML e, na falta, do nome do arquivo.

O CLI depende de `bash`, `jq`, `rsync`, `python3`, `ssh` (modo remoto) e, para `install-dumpsd`, `go`.

## Subcomandos do dumpsd

Todos aceitam `--root <dir>` quando indicado (padrão: diretório atual). `--base-url` é a URL pública, a mesma de `DUMPS_URL`.

| Subcomando | Flags e argumentos |
|---|---|
| `dumpsd version` | imprime a versão |
| `dumpsd capabilities` | imprime `render serve apply prepare activate` |
| `dumpsd render` | `--base-url <url>` (obrigatória), `--ga`, `--no-lock`, `--root <dir>`. Refaz home e listagem |
| `dumpsd prepare` | `--dir <pasta>` (obrigatória), `--manifest <arquivo>`. Lê a especificação JSON do stdin (até 64 KiB) |
| `dumpsd activate` | `--base-url <url>`, `--ga`, `--root <dir>`, `<id>`. Ativa o staging com rollback |
| `dumpsd apply` | `--base-url <url>`, `--root <dir>`, `<archive\|unarchive\|title\|delete> <id>`. `title` lê o título do stdin |
| `dumpsd serve` | veja abaixo |

`--ga` faz a home carregar `/_/dumps-bar.js`. O snippet do GA4 em si é anexado à barra pelo CLI (`DUMPS_GA_ID`). Sem a flag, o render ainda detecta um `dumps-bar.js` que já contém o GA4.

Flags do `dumpsd serve`:

| Flag | Padrão | Significado |
|---|---|---|
| `--base-url` | obrigatória | URL pública |
| `--origin` | obrigatória | origem exata aceita nos POSTs |
| `--listen` | `127.0.0.1:8090` | só loopback (`127.0.0.1:`, `localhost:` ou `[::1]:`) |
| `--root` | diretório atual | raiz dos dumps |
| `--auth` | `access` | `access`, `tailscale` ou `none` |
| `--team` | vazio | host do time no Access (obrigatório com `access`) |
| `--aud` | vazio | AUD do app (obrigatório com `access`) |
| `--allow` | vazio | login permitido com `tailscale`; repita a flag para vários |
| `--static` | desligada | serve também `dumps/` e a listagem; só com `--auth none` |

O `GET /healthz` responde `ok` sem login.

## Makefile

| Alvo | Faz |
|---|---|
| `make test` | `go test -race ./...` em `server/` |
| `make lint` | gofmt, `go vet` e shellcheck |
| `make build` | compila o `dumpsd` para esta máquina em `bin/` |
| `make check` | `lint` e `test`, o mesmo do CI |
