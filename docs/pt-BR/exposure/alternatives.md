# Alternativas ao Cloudflare

O `dumpsd` só precisa de duas coisas: um caminho público para o listener `:8081` do Caddy e um caminho protegido para o `:8080`. Qualquer ferramenta que entregue isso serve. Esta página resume as opções e o que muda em cada uma; comandos e limites de planos pertencem à documentação de cada ferramenta, que muda com frequência.

## Tailscale

- **Público**: Tailscale Funnel expõe a porta `8081` na internet, no hostname `<máquina>.<tailnet>.ts.net` (sem domínio próprio).
- **Admin**: `tailscale serve` publica o admin só dentro do tailnet, e o `dumpsd` usa `DUMPS_AUTH=tailscale`. Ele confia no header `Tailscale-User-Login`, que o `tailscale serve` preenche, e só o aceita vindo do loopback.

Restrições do modo `tailscale` ([auth.md](../auth.md)):

- O `tailscale serve` precisa ficar **direto** na frente do `dumpsd`, sem outro proxy que repasse headers do cliente (inclusive o Caddy do admin), ou o header pode ser forjado.
- O `dumpsd` sem `--static` só responde a API. A página da listagem é HTML estático e precisa ser servida por outro caminho; mapeie caminhos diferentes para backends diferentes no próprio `tailscale serve`. Este repositório não traz esse arranjo pronto.
- Funnel e serve no mesmo nó usam o mesmo hostname, então cada um fica numa porta diferente. `DUMPS_ADMIN_ORIGIN` precisa ser a origem exata, com a porta, que o navegador envia.
- Use `DUMPS_TAILSCALE_ALLOW` se o tailnet tiver outras pessoas.

## ngrok

Publica o `:8081` com um domínio do ngrok (domínio próprio depende do plano). Serve para o host público. Para o admin, o `dumpsd` não lê a identidade do ngrok: proteja o admin com outro meio (Tailscale, ou ssh, abaixo).

## frp

Um servidor frp num VPS e um cliente em casa. Você mantém o domínio e o certificado no VPS. Mesmo caso do ngrok: o `dumpsd` não entende o login do frp, então o admin precisa de outro meio.

## Caddy num VPS com HTTPS automático

Um VPS barato roda Caddy com HTTPS automático (Let's Encrypt) e encaminha para a sua casa por um túnel reverso (WireGuard, ou `ssh -R`). Você usa o domínio que quiser, sem terceiros no caminho do tráfego. É o que dá mais controle e mais trabalho.

## Admin só por ssh

Arranjo não testado neste repositório, derivado do comportamento do `dumpsd`. Sem nenhum serviço de login, dá para nunca expor o admin: use `DUMPS_AUTH=none` e abra o listener por um túnel ssh.

```bash
ssh -L 8080:127.0.0.1:8080 meuservidor
```

Com `DUMPS_ADMIN_ORIGIN=http://localhost:8080`, abra http://localhost:8080/_dumps/. Cuidado: `none` aceita **qualquer** conexão vinda do loopback do servidor, então qualquer usuário local dessa máquina alcança o admin. Só faz sentido em servidor de um usuário só.

## Comparação

| Opção | Custo | Domínio próprio | Login do admin | Esforço |
|---|---|---|---|---|
| Cloudflare Tunnel + Access | plano gratuito cobre uso pessoal (confira o atual) | sim (precisa de domínio no Cloudflare) | Access, nativo (`access`) | médio |
| Tailscale Funnel + serve | plano pessoal gratuito (confira o atual) | não (`*.ts.net`) | nativo (`tailscale`) | médio, o arranjo do admin é manual |
| ngrok | plano gratuito com limites | depende do plano | não nativo | baixo |
| frp (VPS) | custo do VPS | sim | não nativo | alto |
| Caddy num VPS + WireGuard/ssh -R | custo do VPS | sim | não nativo (ssh -L, ou Tailscale) | alto |

## Outro provedor de login na frente (oauth2-proxy e similares)

Não há suporte nativo. Veja [auth.md](../auth.md) para o que dá para fazer e o risco.
