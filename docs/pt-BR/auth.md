# Autenticação do admin

A autenticação protege só a API do admin (`/_dumps/api/*`). A página da listagem é HTML estático servido pelo proxy: quem a protege é o login na frente do host do admin (Access, Tailscale). O host público não tem API.

O modo vem de `DUMPS_AUTH` (flag `--auth` do `dumpsd serve`). Padrão: `access`. O `dumpsd` só escuta em loopback; ele recusa iniciar com outro endereço.

## Modo access (Cloudflare Access)

Config: `DUMPS_AUTH=access`, `DUMPS_ACCESS_TEAM`, `DUMPS_ACCESS_AUD`, `DUMPS_ADMIN_ORIGIN`. Guia: [exposure/cloudflare.md](exposure/cloudflare.md).

O `dumpsd` lê o JWT do header `Cf-Access-Jwt-Assertion` e confere, nesta ordem:

1. O token existe, tem até 16 KiB e três partes.
2. O header do JWT declara `alg` exatamente `RS256` (qualquer outro, incluindo `none` e `HS256`, é recusado).
3. A chave pública do `kid` vem de `https://<time>/cdn-cgi/access/certs` (só chaves RSA). A assinatura RSA com SHA-256 confere.
4. `iss` é `https://<time>`.
5. `aud` contém o AUD configurado (aceita string ou lista).
6. `exp` existe e não passou (com 60 segundos de tolerância de relógio).
7. Se houver `nbf`, já vale (também com 60 segundos de tolerância).

O e-mail do token só vai para o log. Quem pode entrar é decidido pela policy do Access, não pelo `dumpsd`.

Cache das chaves: valem por 1 hora. A renovação é tentada no máximo uma vez por minuto (a cada 10 segundos enquanto não houver nenhuma chave). Se a renovação falhar, as chaves antigas continuam valendo até 24 horas; depois disso, e sem nenhuma chave, o acesso é negado. Em resumo, falha fechada: na dúvida, nega.

Qualquer falha responde `403 acesso negado`, sem detalhe; o motivo vai para o log do serviço.

## Modo tailscale

Config: `DUMPS_AUTH=tailscale`, opcionalmente `DUMPS_TAILSCALE_ALLOW` (logins separados por espaço; vira uma flag `--allow` para cada).

O `dumpsd` confere:

1. A conexão vem do loopback.
2. O header `Tailscale-User-Login` existe e não é vazio.
3. Se há lista `--allow`, o login (sem diferenciar maiúsculas) está nela. Sem lista, qualquer login do tailnet passa.

O header só é confiável porque o `tailscale serve` o preenche. Por isso o `tailscale serve` precisa ficar **direto** na frente do `dumpsd`. Qualquer outro proxy que repasse headers do cliente (inclusive o Caddy do admin) abre a chance de alguém forjar o header. Veja as restrições em [exposure/alternatives.md](exposure/alternatives.md).

## Modo none

Config: `DUMPS_AUTH=none`.

Aceita qualquer conexão do loopback, sem login, e recusa o resto. É o modo do `dumps serve-local`, em que só você alcança a porta. Num servidor compartilhado, qualquer usuário local passa.

A flag `--static` (serve `dumps/` e a listagem pelo próprio `dumpsd`) só existe com `--auth none`.

## Checagens em toda requisição à API

Depois do login, e nesta ordem:

| Checagem | Falha |
|---|---|
| Método `POST` | 405 (`Allow: POST`) |
| Header `Origin` igual, caractere a caractere, a `DUMPS_ADMIN_ORIGIN` (esquema, host e porta) | 403 |
| `Content-Type: application/json` | 415 |
| Corpo de até 4 KiB, só campos `id` e `title` | 400 |
| `id` casa com `^[a-z0-9][a-z0-9-]{0,62}$` | 400 |
| Ação em `archive`, `unarchive`, `title`, `delete` | 404 |
| `title` com 1 a 200 caracteres, sem caracteres de controle | 400 |

Todas as respostas levam `Cache-Control: no-store` e `X-Content-Type-Options: nosniff`.

A checagem de `Origin` e o tipo JSON evitam que uma página de terceiros dispare ações com o seu login (CSRF). Um `POST` simples de formulário não passa.

## Pôr outro login na frente (oauth2-proxy e similares)

Não há suporte nativo: os três modos acima são os únicos. Um arranjo possível, não testado neste repositório, é usar `DUMPS_AUTH=none` com o proxy de login rodando na mesma máquina e encaminhando para o Caddy do admin. Como `none` aceita tudo que vem do loopback, o `dumpsd` passa a confiar que o proxy de login é o único caminho até o listener `:8080`. Se qualquer outra rota chegar ali, o admin fica aberto. Além disso, `DUMPS_ADMIN_ORIGIN` precisa ser a origem pública do proxy de login.

Se você precisar de outro provedor de forma segura, a forma correta é um modo novo no `dumpsd`, que valide a identidade no próprio binário.
