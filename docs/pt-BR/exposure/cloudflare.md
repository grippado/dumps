# Cloudflare Tunnel e Access

Este arranjo expõe o servidor sem porta de entrada aberta: um túnel leva dois hostnames para dois listeners em loopback, e o Cloudflare Access protege o do admin. Os nomes dos menus do painel mudam com o tempo, então consulte a documentação do Cloudflare se algum rótulo for diferente.

Você precisa de um domínio no Cloudflare e da configuração do Caddy do guia do seu host ([Linux](../hosts/linux.md), [macOS](../hosts/macos.md), [Android](../hosts/android-termux.md)).

## Dois hostnames, dois listeners

| Hostname | Vai para | Quem alcança |
|----------|----------|--------------|
| `dumps.example.com` (público) | Caddy `127.0.0.1:8081`, serve só `dumps/` | Qualquer pessoa |
| `admin.example.com` (privado) | Caddy `127.0.0.1:8080`, listagem e API do admin | Só usuários liberados pelo Access |

Nunca aponte o hostname público para o listener do admin (`:8080`). A página da listagem é um arquivo estático sem login próprio, então qualquer pessoa leria seus ids, títulos, tamanhos e nomes de máquina, inclusive os dos dumps arquivados. Só as ações da API são conferidas pelo `dumpsd`.

## 1. Criar o túnel

No servidor, depois de instalar o `cloudflared`:

```bash
cloudflared tunnel login
```

```bash
cloudflared tunnel create dumps
```

O comando imprime o UUID do túnel e grava o arquivo de credenciais `~/.cloudflared/<TUNNEL-UUID>.json`. Trate-o como segredo ([../security.md](../security.md)).

## 2. Ingress

Copie `deploy/cloudflared.yml` para `~/.cloudflared/config.yml` e preencha o UUID do túnel e o seu usuário:

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

## 3. Criar o app do Access primeiro

Crie o app do Access para `admin.example.com` antes de o registro DNS dele existir. Se o hostname resolver primeiro, o listener do admin fica público por um tempo.

1. No painel Zero Trust da Cloudflare, vá em Access, Applications, e adicione uma aplicação self-hosted.
2. Defina o domínio da aplicação como `admin.example.com`.
3. Adicione uma policy com ação Allow e uma regra de inclusão só para você (por exemplo o seu e-mail, ou um grupo do provedor de identidade). Todos os outros são negados por padrão.

Nesta configuração pública, não crie um app do Access para `dumps.example.com`; ele continua
público. Numa instalação Enterprise interna, crie também uma aplicação e uma policy separadas do
Access para o hostname de artefatos, antes de criar o DNS, ou mantenha esse hostname numa rede
privada com acesso só pela VPN. A chave `DUMPS_ACCESS_AUD` abaixo continua identificando a
aplicação do admin cujo JWT o `dumpsd` valida. Veja [publicação Enterprise](../enterprise.md).

## 4. Achar o domínio do time e a tag AUD

- Domínio do time: nas configurações do Zero Trust, General. Parece com `seutime.cloudflareaccess.com`. Esse valor vai em `DUMPS_ACCESS_TEAM`, sem `https://`.
- Tag Application Audience (AUD): abra a aplicação que você criou e copie a tag AUD da visão geral (ou das configurações de token). Esse valor vai em `DUMPS_ACCESS_AUD`.

A tag AUD não é um segredo, mas identifica a sua instalação, então mantenha a configuração fora de repositórios públicos.

## 5. Registros DNS

Agora aponte os dois hostnames para o túnel.

```bash
cloudflared tunnel route dns dumps dumps.example.com
```

```bash
cloudflared tunnel route dns dumps admin.example.com
```

Suba o túnel.

```bash
cloudflared tunnel run dumps
```

Numa máquina que fica ligada, rode como serviço (systemd, runit ou launchd; veja os [guias dos hosts](../hosts/linux.md)).

## 6. Avisar o dumps sobre o Access

Na configuração do CLI:

```text
DUMPS_URL=https://dumps.example.com
DUMPS_LIST_URL=https://admin.example.com/_dumps/
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=seutime.cloudflareaccess.com
DUMPS_ACCESS_AUD=<tag AUD>
DUMPS_ADMIN_ORIGIN=https://admin.example.com
```

`DUMPS_ADMIN_ORIGIN` precisa ser igual ao `Origin` do navegador: esquema e host, sem barra no fim (acrescente a porta só se não for a padrão).

Reinstale para o serviço pegar os argumentos:

```bash
cli/dumps install-dumpsd
```

Abra `https://admin.example.com/_dumps/`, entre pelo Access e tente arquivar um dump. O que o `dumpsd` verifica está descrito em [../auth.md](../auth.md).

## Opcional: ssh para o servidor pelo Access

Se você não quer abrir o ssh para a internet, dá para pôr o ssh atrás do Access também. É opcional e desnecessário quando o CLI roda na mesma rede. Isso usa um terceiro hostname, que não faz parte do `deploy/cloudflared.yml`.

1. Adicione uma regra de ingress antes da regra de 404, por exemplo `hostname: ssh.example.com` com `service: ssh://localhost:22`, e uma rota de DNS para ele.
2. Crie uma aplicação do Access para `ssh.example.com`, com policy própria.
3. No notebook, em `~/.ssh/config`:

```text
Host dumps-box
  HostName ssh.example.com
  User voce
  ProxyCommand cloudflared access ssh --hostname %h
```

Entre uma vez com `cloudflared access login https://ssh.example.com` para as execuções não interativas funcionarem. O CLI chama o ssh com `BatchMode=yes`, então não consegue responder a perguntas. Se o login tiver expirado, o `dumps` termina com saída 3 ("servidor inalcançável").

## Conferir

```bash
curl -i https://admin.example.com/_dumps/
```

Sem login, a resposta precisa ser um redirecionamento para o login do Access, nunca a listagem.
