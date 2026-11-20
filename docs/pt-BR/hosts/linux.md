# Host Linux

Use para um Raspberry Pi, um VPS, um NAS ou qualquer máquina com systemd e ssh. O seu notebook roda o CLI, e o servidor roda o `dumpsd`, o Caddy e um túnel.

## 1. Pré-requisitos do servidor

- Login ssh (com chave, o CLI usa `BatchMode`)
- `rsync`
- Caddy, da sua distribuição ou da documentação do Caddy
- systemd com serviços de usuário

## 2. Alias de ssh no notebook

Em `~/.ssh/config`:

```text
Host dumps-box
  HostName box.example.com
  User voce
```

## 3. Configuração do CLI

Em `~/.config/dumps/config` no notebook:

```text
DUMPS_HOST=dumps-box
DUMPS_URL=https://dumps.example.com
DUMPS_LIST_URL=https://admin.example.com/_dumps/
DUMPS_TARGET=linux/arm64
DUMPS_SERVICE=systemd-user
DUMPS_ADMIN_ORIGIN=https://admin.example.com
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=seutime.cloudflareaccess.com
DUMPS_ACCESS_AUD=<tag AUD>
```

`DUMPS_TARGET` é a arquitetura do servidor: `linux/arm64` (Raspberry Pi 4 e 5 com sistema de 64 bits, a maioria dos VPS ARM) ou `linux/amd64` (VPS e NAS x86). Sem `DUMPS_ROOT`, a raiz é a home do usuário ssh, então os dumps ficam em `~/dumps`.

Para um serviço, `DUMPS_AUTH` é `access` ou `tailscale`. `none` é recusado. Os valores que vão para o serviço (raiz, endereço de escuta, URL, origem do admin, time, AUD e logins do Tailscale) só podem ter letras, números e `: / . _ @ -`. Sobre os modos de auth e a configuração do Access, veja [../auth.md](../auth.md) e [../exposure/cloudflare.md](../exposure/cloudflare.md).

## 4. Instalar o dumpsd e o serviço

No notebook (precisa de Go):

```bash
cli/dumps install-dumpsd
```

O comando compila o `dumpsd` para o alvo, envia para `~/bin/dumpsd` e grava `~/.config/systemd/user/dumpsd.service` com `Restart=on-failure` e `NoNewPrivileges=true`, depois roda `daemon-reload`, habilita e reinicia o serviço. A unit executa `dumpsd serve` em `DUMPS_LISTEN` (padrão `127.0.0.1:8090`). Ele também cria `bin/`, `dumps/`, `dumps-meta/` e `dumps-staging/` na raiz.

Confira no servidor:

```bash
systemctl --user status dumpsd
```

## 5. Manter os serviços de usuário rodando depois do logout

Serviços de usuário param quando a sua última sessão termina, a menos que o lingering esteja ligado. Para eles subirem no boot e sobreviverem ao logout, rode no servidor:

```bash
loginctl enable-linger $USER
```

## 6. Caddy

Copie `deploy/Caddyfile` para o servidor (por exemplo `/etc/caddy/Caddyfile`) e troque `/home/you` pela raiz absoluta. O arquivo define dois listeners em loopback:

- `127.0.0.1:8081`: público, serve só `dumps/`
- `127.0.0.1:8080`: admin, encaminha `/_dumps/api/*` para o `dumpsd` em `:8090` e serve a listagem de `admin/`

O arquivo desliga o `auto_https` e a API admin do Caddy (`admin off`), porque o TLS fica no túnel. Recarregue o Caddy.

```bash
sudo systemctl reload caddy
```

O usuário que roda o Caddy precisa de leitura em `dumps/` e `admin/`. Se você instalou o Caddy como serviço do sistema, ele roda como outro usuário, então confira as permissões das pastas na sua home.

Uma alternativa é rodar o Caddy como serviço de usuário, ao lado do `dumpsd`. Exemplo de `~/.config/systemd/user/caddy.service`:

```text
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

Ajuste o caminho do binário (`command -v caddy`). Com o lingering ligado, os dois sobem no boot.

## 7. Expor na internet

Exponha os dois listeners com um túnel: [../exposure/cloudflare.md](../exposure/cloudflare.md) ou [../exposure/alternatives.md](../exposure/alternatives.md).

## 8. Publicar

```bash
cli/dumps publish ./pagina.html
```

A home e a listagem são gravadas pela primeira publicação. Para criá-las antes:

```bash
cli/dumps regen
```
