# Host Linux

Vale para Raspberry Pi, VPS, NAS ou qualquer máquina com systemd. O CLI roda no seu notebook e fala com o servidor por ssh.

## Preparo do servidor

- ssh com chave (o CLI usa `BatchMode=yes`, sem pedir senha)
- `rsync`
- Caddy (pelo gerenciador de pacotes da distro)

No notebook, uma entrada em `~/.ssh/config` com o nome que você vai usar em `DUMPS_HOST`.

## Configuração do CLI

Em `~/.config/dumps/config`:

```
DUMPS_HOST=meuservidor
DUMPS_ROOT=/home/voce
DUMPS_URL=https://dumps.example.com
DUMPS_TARGET=linux/arm64
DUMPS_SERVICE=systemd-user
DUMPS_ADMIN_ORIGIN=https://admin.example.com
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=seutime.cloudflareaccess.com
DUMPS_ACCESS_AUD=<AUD do app do admin>
```

`DUMPS_TARGET` é `linux/arm64` (Raspberry Pi 64 bits, a maioria dos ARM) ou `linux/amd64` (VPS e NAS x86). Para o auth com Tailscale, veja [auth.md](../auth.md).

## Instalar o dumpsd e o serviço

```bash
cli/dumps install-dumpsd
```

O comando cria `bin/`, `dumps/`, `dumps-meta/` e `dumps-staging/` na raiz, envia o binário e grava `~/.config/systemd/user/dumpsd.service` no servidor, depois roda `daemon-reload`, `enable` e `restart` (`systemctl --user`). O serviço tem `Restart=on-failure` e `NoNewPrivileges=true`.

É um serviço de usuário, que por padrão só roda com o usuário logado. Para ele subir no boot e sobreviver ao logout, rode no servidor:

```bash
loginctl enable-linger $USER
```

Conferir:

```bash
systemctl --user status dumpsd
```

## Caddy

Copie `deploy/Caddyfile` para o servidor e troque `/home/you` pelo `DUMPS_ROOT` absoluto. O arquivo define dois listeners em loopback:

- `:8081`: público, serve só `<raiz>/dumps`
- `:8080`: admin, serve `<raiz>/admin` e encaminha `/_dumps/api/*` para o `dumpsd` em `127.0.0.1:8090`

O arquivo desliga o `auto_https` e a API admin do Caddy (`admin off`), porque o TLS fica no túnel. O usuário que roda o Caddy precisa de leitura em `dumps/` e `admin/`. Se usar o serviço do pacote da distro, que roda como outro usuário, confira as permissões da pasta.

Uma alternativa é rodar o Caddy como serviço de usuário, ao lado do `dumpsd`. Exemplo de `~/.config/systemd/user/caddy.service`:

```
[Unit]
Description=Caddy (dumps)
After=network-online.target

[Service]
ExecStart=/usr/bin/caddy run --config %h/Caddyfile
Restart=on-failure

[Install]
WantedBy=default.target
```

```bash
systemctl --user enable --now caddy
```

Ajuste o caminho do binário (`command -v caddy`). Com `enable-linger` ligado, os dois sobem no boot.

## Expor na internet

Siga [exposure/cloudflare.md](../exposure/cloudflare.md) ou uma das [alternativas](../exposure/alternatives.md).

## Primeiro publish

```bash
cli/dumps publish ./pagina.html
```
