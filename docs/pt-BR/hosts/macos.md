# Host macOS

Um Mac que fica ligado serve bem de servidor doméstico. Aqui o serviço do admin é um LaunchAgent do `launchd`, e o Caddy vem do Homebrew.

## Preparo do servidor (o Mac)

- Ligue o login remoto em Ajustes do Sistema, Geral, Compartilhamento, Login Remoto.
- Impeça o Mac de dormir enquanto estiver servindo (Ajustes do Sistema, Bateria ou Economia de Energia).
- Faça login na sessão gráfica do usuário que vai rodar o serviço. O LaunchAgent é instalado no domínio `gui/<uid>`, que só existe com esse usuário logado.
- Instale o Caddy:

```bash
brew install caddy
```

## Configuração do CLI

Em `~/.config/dumps/config` no notebook:

```
DUMPS_HOST=meumac
DUMPS_ROOT=/Users/voce/dumps-root
DUMPS_URL=https://dumps.example.com
DUMPS_TARGET=darwin/arm64
DUMPS_SERVICE=launchd
DUMPS_ADMIN_ORIGIN=https://admin.example.com
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=seutime.cloudflareaccess.com
DUMPS_ACCESS_AUD=<AUD do app do admin>
```

`DUMPS_TARGET` é `darwin/arm64` (Apple Silicon) ou `darwin/amd64` (Intel). Se o notebook e o servidor forem o mesmo Mac, use o [modo local](../local.md).

## Instalar o dumpsd e o serviço

```bash
cli/dumps install-dumpsd
```

O comando envia o binário e grava `~/Library/LaunchAgents/io.github.grippado.dumpsd.plist` no servidor, com `RunAtLoad` e `KeepAlive`, e carrega o agente com `launchctl bootstrap`. O log fica em `<raiz>/dumpsd.log`.

Conferir:

```bash
launchctl print gui/$(id -u)/io.github.grippado.dumpsd
```

## Caddy

Copie `deploy/Caddyfile` para o Mac e troque `/home/you` pelo `DUMPS_ROOT` absoluto (por exemplo `/Users/voce/dumps-root`). Rode:

```bash
caddy run --config ~/Caddyfile
```

Para subir no login, use `brew services` com o Caddyfile padrão do Homebrew (`$(brew --prefix)/etc/Caddyfile`) apontando para o seu conteúdo, ou um LaunchAgent próprio.

## Expor na internet

Siga [exposure/cloudflare.md](../exposure/cloudflare.md) ou uma das [alternativas](../exposure/alternatives.md).
