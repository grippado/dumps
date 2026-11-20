# Autenticação do admin

Os dumps publicados são estáticos e não têm login próprio. Uma instalação pode servi-los publicamente ou restringir todo o host de artefatos com uma policy externa de SSO ou VPN. `DUMPS_AUTH` vale só para a API do admin: o `dumpsd serve` confere toda requisição sob `/_dumps/api/` com um de três modos, escolhido com `--auth` (ou `DUMPS_AUTH` na configuração do CLI). O padrão é `access`.

A página da listagem é HTML estático. Quem a protege é o login na frente do host do admin (Access, Tailscale), não o `dumpsd`. O host público não tem API. O `dumpsd serve` só escuta em loopback e recusa iniciar em outro endereço.

Para um serviço instalado, `DUMPS_AUTH` é `access` ou `tailscale`. O CLI recusa `DUMPS_AUTH=none` para serviço.

A checagem de identidade vem primeiro. Uma falha devolve `403 acesso negado`, e o motivo vai para o log.

## access (Cloudflare Access)

Configuração: `DUMPS_AUTH=access`, `DUMPS_ACCESS_TEAM`, `DUMPS_ACCESS_AUD`, `DUMPS_ADMIN_ORIGIN`. Guia: [exposure/cloudflare.md](exposure/cloudflare.md).

O `dumpsd` lê o header `Cf-Access-Jwt-Assertion` e o verifica por conta própria, então um pedido que não passou pelo Access não consegue forjá-lo. Qualquer problema nega o acesso.

O que é conferido:

- O token existe, tem no máximo 16 KiB e três partes separadas por ponto.
- O algoritmo do header é `RS256`. Qualquer outro (incluindo `none` e `HS256`) é recusado.
- A assinatura confere com a chave RSA do `kid` do token.
- `iss` é exatamente `https://<time>`.
- `aud` contém a tag AUD configurada (string ou lista).
- `exp` existe e não passou, e `nbf` (se houver) não está no futuro, os dois com 60 segundos de tolerância de relógio.

As chaves de assinatura vêm de `https://<time>/cdn-cgi/access/certs` e ficam em cache:

- As chaves são renovadas depois de uma hora, com pelo menos um minuto entre tentativas (10 segundos antes da primeira busca bem-sucedida).
- Se a renovação falhar, as chaves em cache continuam valendo por até 24 horas depois da última busca bem-sucedida. Depois disso, todo pedido é negado.
- Um `kid` desconhecido é negado.

A claim `email` só serve para rotular linhas de log. Não é checagem de autorização: quem pode entrar é decidido pela policy do Access. Veja [exposure/cloudflare.md](exposure/cloudflare.md).

## tailscale

Configuração: `DUMPS_AUTH=tailscale` e, opcionalmente, `DUMPS_TAILSCALE_ALLOW` (logins separados por espaço, uma flag `--allow` para cada).

Usa a identidade que o `tailscale serve` coloca no header `Tailscale-User-Login`.

- O par TCP precisa ser loopback. Qualquer outro é negado.
- O header precisa existir e não estar vazio.
- Com um ou mais `--allow <login>` (`DUMPS_TAILSCALE_ALLOW`), o login precisa estar na lista (sem diferenciar maiúsculas). Sem lista, qualquer login que chegue ao listener é aceito.

Este modo confia num header comum. Só é seguro quando o `tailscale serve` é o único caminho até o listener do admin. O `dumpsd` aceita o header de qualquer conexão em loopback, então não ponha nenhum outro proxy ou túnel na frente que possa repassar um `Tailscale-User-Login` enviado pelo cliente, e nunca chegue a ele por Funnel ou por um hostname público. Veja [exposure/alternatives.md](exposure/alternatives.md).

## none

Configuração: não há chave a definir para serviço. O `dumps serve-local` inicia o `dumpsd` com `--auth none` por conta própria e não lê `DUMPS_AUTH`.

`none` aceita só conexões diretas do loopback e não define identidade (`local` nos logs). Também recusa, com o mesmo `403`, qualquer pedido que traga um destes headers, que proxies e túneis adicionam:

- `X-Forwarded-For`
- `Forwarded`
- `Cf-Connecting-Ip`
- `X-Real-Ip`
- `Cf-Access-Jwt-Assertion`

O CLI recusa `DUMPS_AUTH=none` para um serviço instalado. Há um único uso para este modo: o modo local, em que só você alcança a porta.

Não ponha o `none` atrás de proxy ou túnel. A checagem de headers pega proxies que adicionam esses headers, mas um proxy na mesma máquina que não adicione nenhum deles ainda conecta pelo loopback e passaria. Numa máquina compartilhada, qualquer usuário local consegue chamar a API diretamente.

`--static` (servir o site pelo próprio `dumpsd`) só é permitido com `--auth none`, e exige `--admin-listen` numa porta loopback diferente. Nesse modo, o listener do admin também serve a página da listagem sem login, e só ele tem a API. Veja [local.md](local.md).

## Checagens em toda requisição à API

Depois da autenticação, nesta ordem:

1. O método precisa ser `POST` (senão `405`, com `Allow: POST`).
2. O header `Origin` precisa ser igual a `--origin` (`DUMPS_ADMIN_ORIGIN`), caractere a caractere: esquema, host e porta (senão `403`). No modo local a origem é `DUMPS_ADMIN_ORIGIN` ou `http://localhost:<porta de DUMPS_ADMIN_LISTEN>`, então abra o admin por `localhost`, não por `127.0.0.1`. Isso bloqueia requisições cross-site de um navegador logado. Um pedido sem `Origin` é recusado.
3. `Content-Type` precisa ser `application/json` (senão `415`).
4. O corpo tem no máximo 4 KiB e precisa ser um objeto JSON só com os campos `id` e `title`. Campos desconhecidos são recusados (`400`).
5. O `id` precisa casar com `^[a-z0-9][a-z0-9-]{0,62}$` (senão `400`).
6. As ações são `archive`, `unarchive`, `title` e `delete`. Outras devolvem `404`.
7. Em `title`, o valor é aparado e as sequências de espaços viram um só, precisa ser UTF-8 válido sem caracteres de controle nem de formatação invisíveis, e ter de 1 a 200 caracteres (senão `400`).

A checagem de `Origin` e o tipo JSON impedem que uma página de terceiros dispare ações com o seu login (CSRF). Um `POST` simples de formulário não passa.

Todas as respostas levam `Cache-Control: no-store` e `X-Content-Type-Options: nosniff`.

## Pôr outro login na frente

Não há modo embutido para outros proxies com identidade, como oauth2-proxy, um gateway OIDC ou basic auth. Pôr um na frente do host do admin não faz o `dumpsd` verificá-lo, e `--auth none` não é uma forma segura de combinar os dois (veja acima). Isso ainda não tem suporte nativo. Use `access` ou `tailscale`, ou gerencie os dumps pelo CLI por ssh (`dumps archive`, `edit`, `rm`), que não precisa de host do admin.

Uma forma segura de acrescentar outro provedor seria um modo novo no `dumpsd`, que valide a identidade dentro do próprio binário.
