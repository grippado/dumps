# Alternativas ao Cloudflare

O dumps só precisa de duas coisas de uma camada de exposição: o host público chegar ao Caddy `127.0.0.1:8081`, e o host do admin (se você quiser um) chegar ao Caddy `127.0.0.1:8080` atrás de um login que o `dumpsd` consiga verificar. Comandos e limites dessas ferramentas de terceiros mudam, então trate os trechos como ponto de partida e confira a documentação delas.

## Tailscale

O Tailscale Funnel publica uma porta local num hostname público `*.ts.net`. O `tailscale serve` compartilha uma porta só dentro do seu tailnet.

- Público: Funnel no listener público (`8081`).
- Admin: `tailscale serve` no listener do admin (`8080`), para que só dispositivos do tailnet o abram.

```bash
tailscale funnel --bg 8081
```

```bash
tailscale serve --bg --https=8443 8080
```

Funnel e serve usam portas diferentes no mesmo hostname, e as flags exatas variam com a versão do Tailscale. Ajuste a configuração do CLI ao que você acabou usando:

```text
DUMPS_URL=https://box.tailnet-name.ts.net
DUMPS_ADMIN_ORIGIN=https://box.tailnet-name.ts.net:8443
DUMPS_AUTH=tailscale
DUMPS_TAILSCALE_ALLOW="ada@example.com"
```

O `tailscale serve` adiciona o header `Tailscale-User-Login` em que o `dumpsd` confia, e o `dumpsd` só o aceita vindo do loopback. O Funnel serve o seu nome `ts.net`. Limites do modo `tailscale` ([../auth.md](../auth.md)):

- O `dumpsd` confia no header vindo de qualquer conexão em loopback, então nada além do `tailscale serve` pode conseguir alcançar o listener do admin. Nunca chegue a ele por Funnel, por um hostname público nem por outro proxy ou túnel.
- `DUMPS_ADMIN_ORIGIN` precisa ser a origem exata que o navegador envia, com a porta.
- Use `DUMPS_TAILSCALE_ALLOW` se o tailnet tiver outras pessoas.

## ngrok

O ngrok pode publicar o `8081` numa URL pública (domínio próprio depende do plano). Os recursos de login dele (OAuth, basic auth) ficam na frente do host do admin, mas o `dumpsd` não tem modo que leia a identidade do ngrok, então a API do admin ficaria protegida só pelo login de fora e pela checagem de Origin. Isso não tem suporte nativo, veja "Pôr outro login na frente" em [../auth.md](../auth.md). Para o trabalho de admin, prefira `dumps archive`, `edit` e `rm` por ssh, e exponha só o listener público.

## frp

Rode o `frps` num VPS com IP público e o `frpc` na máquina de casa, encaminhando o `8081` para o VPS. A mesma ressalva do ngrok vale para o host do admin: o `dumpsd` não entende o login do frp, então exponha só o listener público, a menos que você acrescente uma camada autenticada própria.

## Caddy num VPS, com túnel até em casa

Rode o Caddy num VPS para ter HTTPS automático e domínio próprio, e chegue à máquina de casa por um túnel WireGuard ou um túnel reverso de ssh.

```bash
ssh -N -R 127.0.0.1:18081:127.0.0.1:8081 vps.example.com
```

No VPS, o Caddy faz proxy reverso de `dumps.example.com` para `127.0.0.1:18081`. Mantenha o listener do admin fora do VPS, ou proteja-o com um mecanismo que o `dumpsd` suporte. Você usa o domínio que quiser, sem terceiros no caminho do tráfego. É o que dá mais controle e mais trabalho.

## Admin pelo CLI por ssh

Se você não quer um host de admin, não exponha o listener do admin. `dumps archive`, `unarchive`, `edit` e `rm` executam o `dumpsd apply` no servidor por ssh, então quem já tem ssh não precisa de camada de login. Exponha só o listener público. Não use `DUMPS_AUTH=none` para isso: o CLI o recusa para serviço.

## Comparação

| Opção | Custo | Domínio próprio | Login do admin | Esforço |
|-------|-------|-----------------|----------------|---------|
| Cloudflare Tunnel e Access | Um domínio no Cloudflare; confira os limites atuais do plano | Sim | Nativo (`access`) | Médio: túnel, DNS, app do Access |
| Tailscale Funnel e serve | Confira os limites atuais do plano | Usa o nome `ts.net` | Nativo (`tailscale`) | Baixo |
| ngrok | Confira os limites atuais do plano | Depende do plano | Nenhum nativo | Baixo |
| frp num VPS | Um VPS | Sim | Nenhum nativo | Médio |
| Caddy num VPS com WireGuard ou túnel ssh | Um VPS | Sim, HTTPS automático | Nenhum nativo | Maior: VPS, túnel, manter ativo |

Para um site público sem interface de admin, qualquer opção serve. Para o admin no navegador, use Cloudflare Access ou Tailscale, ou gerencie os dumps pelo CLI.

## Outro provedor de login na frente

Não há suporte nativo para oauth2-proxy e similares. Veja [../auth.md](../auth.md) para o que isso significa e o risco.
