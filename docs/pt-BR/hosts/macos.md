# Host macOS

Use quando um Mac (um Mac mini, um notebook antigo) é o servidor sempre ligado e você publica a partir de outra máquina. O serviço do admin é um LaunchAgent do launchd e o Caddy vem do Homebrew. Se você só quer os dumps no Mac que está usando, veja o [modo local](../local.md).

## 1. Pré-requisitos do Mac

- Login remoto ligado (Ajustes do Sistema, Geral, Compartilhamento) para o CLI conseguir entrar por ssh
- Homebrew
- `rsync` e Caddy
- Uma sessão gráfica logada do usuário que vai rodar o serviço: o LaunchAgent fica no domínio `gui/<uid>`, que só existe com esse usuário logado

Instale:

```bash
brew install caddy rsync
```

Impeça o Mac de dormir se ele deve servir o dia todo (ajustes de energia).

## 2. Configuração do CLI na máquina de onde você publica

```text
DUMPS_HOST=meu-mac
DUMPS_URL=https://dumps.example.com
DUMPS_LIST_URL=https://admin.example.com/_dumps/
DUMPS_TARGET=darwin/arm64
DUMPS_SERVICE=launchd
DUMPS_ADMIN_ORIGIN=https://admin.example.com
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=seutime.cloudflareaccess.com
DUMPS_ACCESS_AUD=<tag AUD>
```

`DUMPS_TARGET` é `darwin/arm64` para Apple Silicon ou `darwin/amd64` para Intel. `meu-mac` é um alias de ssh em `~/.ssh/config`. Sem `DUMPS_ROOT`, a raiz é a home do usuário ssh.

Para um serviço, `DUMPS_AUTH` é `access` ou `tailscale`. `none` é recusado. Os valores que vão para o serviço só podem ter letras, números e `: / . _ @ -`. Veja [../auth.md](../auth.md).

Se a máquina de onde você publica e o servidor forem o mesmo Mac, use o [modo local](../local.md).

## 3. Instalar o dumpsd e o agente do launchd

```bash
cli/dumps install-dumpsd
```

O comando compila o `dumpsd` para o Mac, envia e grava `~/Library/LaunchAgents/io.github.grippado.dumpsd.plist` (com `RunAtLoad` e `KeepAlive`, log em `<raiz>/dumpsd.log`), depois carrega com `launchctl bootstrap gui/<uid>`.

É um LaunchAgent no domínio gráfico do usuário, então começa quando esse usuário faz login. Para um Mac sem ninguém por perto, ligue o login automático desse usuário.

A publicação funciona no macOS porque o `dumpsd activate` pega o lock por conta própria: o Mac não precisa de um comando `flock`.

Confira no Mac:

```bash
launchctl print gui/$(id -u)/io.github.grippado.dumpsd
```

## 4. Caddy

Copie `deploy/Caddyfile` para o Mac e troque `/home/you` pela raiz absoluta (por exemplo `/Users/voce`). Depois rode o Caddy com esse arquivo.

```bash
caddy run --config /caminho/do/Caddyfile
```

Para o Caddy subir no login, envolva esse comando num agente do launchd seu.

## 5. Expor na internet

Exponha os dois listeners em loopback com um túnel: [../exposure/cloudflare.md](../exposure/cloudflare.md) ou [../exposure/alternatives.md](../exposure/alternatives.md).

## 6. Publicar

```bash
cli/dumps publish ./pagina.html
```

As miniaturas são geradas na máquina que publica, não no Mac servidor.
