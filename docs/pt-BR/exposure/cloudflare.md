# Expor com Cloudflare Tunnel e Access

Você precisa de um domínio no Cloudflare. O túnel leva o tráfego do servidor para a internet sem abrir portas no roteador, e o Access coloca o login na frente do admin.

## Dois hostnames, dois listeners

| Hostname | Quem atende | Acesso |
|---|---|---|
| `dumps.example.com` | Caddy `127.0.0.1:8081` (só `dumps/`) | público |
| `admin.example.com` | Caddy `127.0.0.1:8080` (listagem e API) | atrás do Access |

**Nunca aponte o hostname público para o listener do admin (`:8080`).** Isso publicaria a listagem (ids, títulos, tamanhos, nomes de máquina) para todo mundo.

## 1. Criar o túnel

No servidor, depois de instalar o `cloudflared`:

```bash
cloudflared tunnel login
```

```bash
cloudflared tunnel create dumps
```

O segundo comando imprime o UUID do túnel e grava o arquivo de credenciais em `~/.cloudflared/<UUID>.json`. Esse arquivo é segredo ([security.md](../security.md)).

## 2. Ingress

Copie `deploy/cloudflared.yml` para `~/.cloudflared/config.yml` e preencha o UUID e o caminho das credenciais:

```yaml
tunnel: <TUNNEL-UUID>
credentials-file: /home/you/.cloudflared/<TUNNEL-UUID>.json
ingress:
  - hostname: dumps.example.com
    service: http://127.0.0.1:8081
  - hostname: admin.example.com
    service: http://127.0.0.1:8080
  - service: http_status:404
```

A última regra devolve 404 para qualquer outro hostname.

## 3. Criar o app do Access ANTES do DNS do admin

No painel Cloudflare Zero Trust, crie uma aplicação self-hosted para `admin.example.com` (Access, Applications). Faça isso antes de criar o registro DNS do admin. Se o DNS existir primeiro, o admin fica público até o app existir.

Adicione uma policy de **Allow** só com o seu e-mail (ou os e-mails do time). Quem a policy não liberar nem chega ao `dumpsd`.

## 4. Achar o domínio do time e o AUD

- **Domínio do time**: o host `<nome>.cloudflareaccess.com` definido no Zero Trust (Settings, General). Esse valor vai em `DUMPS_ACCESS_TEAM`, **sem** `https://`.
- **Application Audience (AUD)**: um hash exibido na página do app do admin (Overview, "Application Audience (AUD) Tag"). Vai em `DUMPS_ACCESS_AUD`.

Os nomes dos menus mudam com o tempo; procure pelos termos acima.

## 5. Registrar o DNS

Só depois do app do Access:

```bash
cloudflared tunnel route dns dumps dumps.example.com
```

```bash
cloudflared tunnel route dns dumps admin.example.com
```

## 6. Configurar o dumpsd

No `~/.config/dumps/config`:

```
DUMPS_URL=https://dumps.example.com
DUMPS_ADMIN_ORIGIN=https://admin.example.com
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=seutime.cloudflareaccess.com
DUMPS_ACCESS_AUD=<AUD do app do admin>
```

Depois reinstale para o serviço pegar os valores:

```bash
cli/dumps install-dumpsd
```

O que o `dumpsd` confere está em [auth.md](../auth.md).

## 7. Subir o túnel

```bash
cloudflared tunnel run dumps
```

Em produção, rode como serviço (systemd, launchd ou runit; veja [hosts](../hosts/linux.md)).

## Opcional: ssh pelo Access

Se o servidor não tem IP alcançável, dá para fazer o ssh do CLI passar pelo próprio Cloudflare. Isso usa um terceiro hostname (por exemplo `ssh.example.com`) com ingress `ssh://localhost:22` e outro app do Access; não faz parte do `deploy/cloudflared.yml`. No notebook, em `~/.ssh/config`:

```
Host meuservidor
  HostName ssh.example.com
  ProxyCommand cloudflared access ssh --hostname %h
```

O CLI roda ssh em modo não interativo. Se o login do Access expirar, o `dumps` termina com `servidor inalcançável` (saída 3). Renove abrindo um `ssh meuservidor` interativo.

## Conferência

```bash
curl -i https://admin.example.com/_dumps/
```

Sem login, a resposta precisa ser um redirecionamento para o login do Access, nunca a listagem.
