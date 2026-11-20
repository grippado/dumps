# Segurança

O dumps é um publicador auto-hospedado de HTML estático. O listener de artefatos não tem login próprio, então você decide se vai servi-lo publicamente ou restringi-lo por uma camada externa de SSO/VPN. Esta página descreve o que ele defende, o que não defende e o que você precisa fazer por conta própria. Para reportar um problema, veja o [SECURITY.md](../../SECURITY.md).

## Modelo

- **O host de artefatos serve apenas arquivos estáticos de `dumps/`, sem login próprio.** Manifests (`dumps-meta/`), staging, `bin/` e a listagem ficam fora da pasta servida. O Caddy de exemplo esconde arquivos com ponto no início e redireciona as URLs `index.html` e `.html` para as limpas. No modo local, o `dumpsd --static` aplica as mesmas regras: segmentos com ponto no início devolvem 404, `/_dumps` devolve 404 na origem dos dumps, e só `dumps/` é servido ali. A origem dos artefatos é pública, a menos que uma policy externa de SSO ou VPN proteja o host inteiro.
- **O host do admin fica atrás de login.** A página da listagem e `/_dumps/api/*` só devem ser alcançáveis por Cloudflare Access ou Tailscale. O `dumpsd` também verifica a identidade por conta própria (JWT ou header do Tailscale), o `Origin` e o formato do pedido. Veja [auth.md](auth.md).
- **O `dumpsd serve` escuta só em loopback.** Ele recusa qualquer outro endereço em `--listen` ou `--admin-listen`. Nada abre uma porta de entrada na internet; quem expõe é o túnel.
- **Um único escritor.** Home, listagem e manifests são gravados só pelo `dumpsd`, sob um lock de arquivo, com gravação seguida de rename. A publicação envia para o staging e troca de forma atômica, com rollback.
- **O modo local tem duas origens.** Os dumps são servidos numa porta e o admin noutra. A API só aceita a origem do admin, então o JavaScript de um dump não consegue chamá-la. Veja [local.md](local.md).

## Regras de caminho e de id

- Ids casam com `^[a-z0-9][a-z0-9-]{0,62}$`, conferido no CLI, no `dumpsd activate`, no `dumpsd apply` e na API. Assim um id não tem `/`, `..` nem ponto no início, e `_` (a pasta dos arquivos compartilhados) não pode ser um id.
- O CLI recusa origens dentro de `DUMPS_BLOCKED_PATHS`, ou que as contenham (saída 2). Use para pastas que nunca devem ser publicadas, como as de trabalho ou de clientes. Cada item expande o `~` por conta própria.
- Em pastas, symlinks precisam resolver para dentro da pasta e não para um caminho bloqueado. O conteúdo é copiado com os symlinks resolvidos, e `.git/` e `.DS_Store` ficam de fora. Um arquivo solto precisa ser `.html` ou `.htm`.
- O `listing_dir` do `site.json` não pode ser absoluto, ser `.` nem começar com `..`. O primeiro componente não pode ser `dumps`, `dumps-meta`, `dumps-staging` nem `bin` (comparado sem diferenciar maiúsculas), e não pode começar com ponto.
- `DUMPS_ROOT` não pode conter aspa simples. Os valores gravados num serviço (a raiz resolvida, o endereço de escuta, a URL, a origem do admin, o time e o AUD do Access e cada login do Tailscale) só podem ter letras, números e `: / . _ @ -`, porque passam por um shell.

## O que um atacante consegue e não consegue

| Atacante | Consegue | Não consegue |
|----------|----------|--------------|
| Qualquer pessoa que alcance um host de artefatos sem proteção | Ler os dumps publicados e a home. Adivinhar ids (ids aleatórios de 10 caracteres não são adivinháveis na prática, mas um link deixa de ser segredo depois de compartilhado) | Ler manifests, staging ou a listagem, chamar a API do admin, gravar arquivos |
| Uma página visitada por um admin (CSRF) | Enviar um pedido cross-site ao host do admin | Passar pela checagem de `Origin` nem carregar um token válido do Access |
| Quem chega à API do admin sem login | Nada: `403` | Arquivar, renomear nem apagar |
| Quem forja um JWT | Nada | Assinar com a chave da Cloudflare. `alg: none` e `HS256` são recusados |
| Um token de outro app do Access, do mesmo time | Nada | Passar na checagem de `aud` |
| Um token expirado, ou de outro time | Nada | Passar nas checagens de `exp` e `iss` |
| Um usuário logado e permitido pela sua policy | Arquivar, desarquivar, renomear e apagar qualquer dump | Publicar nem alterar conteúdo (isso exige ssh no servidor), nem executar comandos |
| JavaScript dentro de um dump, no modo local | Ler os outros dumps da origem dos dumps | Chamar a API do admin (o `Origin` dele não é o da origem do admin, e o listener dos dumps não tem API) |
| Um usuário ssh com a sua chave | Tudo | |

Dumps arquivados somem da home, mas podem ser lidos por qualquer pessoa que alcance o host de artefatos e tenha o link. `--hidden` serve para revisão, não para privacidade; o link fica no ar assim que a publicação termina. Use uma barreira externa de SSO ou VPN para material interno e `dumps rm` para tirar um dump do ar.

## HTML de terceiros divide a sua origem

Todo dump é servido da mesma origem, por exemplo `https://dumps.example.com`. Scripts de um dump conseguem ler os cookies e o `localStorage` de todos os outros dumps daquele host (a home guarda a preferência de visualização em `localStorage`, por exemplo) e buscar os arquivos deles. A barra injetada também grava ali um cookie pequeno, `dumps_tip`. Um dump malicioso também consegue se passar por uma página sua (a barra e o rodapé são HTML injetado, não uma fronteira de segurança), e uma má reputação pode alcançar o domínio inteiro.

O admin fica em outro hostname e a API exige um `Origin` exato, então um dump malicioso não consegue acionar o admin por si só. Mesmo assim, se o hostname do admin for um subdomínio do mesmo domínio registrável dos dumps, cookies definidos no domínio pai podem interferir.

Publique só HTML que você escreveu ou em que confia. Para conteúdo em que você não confia (gerado por outros, de origem desconhecida), use um domínio registrável separado e uma instância separada do `dumps`, com a própria raiz. O mesmo cuidado vale para páginas geradas por um agente.

No modo local, os dumps e o admin ficam em origens diferentes (portas diferentes), e a API só aceita a origem do admin. O JavaScript de um dump não consegue arquivar, renomear nem apagar outros dumps. Duas coisas continuam valendo:

- Os dumps ainda dividem a origem dos dumps entre si, como acima.
- `--auth none` confia em qualquer pedido direto do loopback. Ele recusa pedidos que tragam `X-Forwarded-For`, `Forwarded`, `Cf-Connecting-Ip`, `X-Real-Ip` ou `Cf-Access-Jwt-Assertion`, mas um processo local que não seja um navegador pode definir o `Origin` sozinho e chamar a API. Trate o modo local como de um usuário só.

## HTML confiável no site.json

`tagline_html`, `host_card_html` e `footer_html` do `site.json` entram na home como HTML cru ([configuration.md](configuration.md)). Trate o `site.json` como código: só o dono do servidor deve poder editá-lo.

## Segredos e arquivos que nunca entram em commit

- As credenciais do túnel (`~/.cloudflared/<TUNNEL-UUID>.json`) e o `cert.pem` do `cloudflared tunnel login`.
- Chaves privadas de ssh.
- O seu `~/.config/dumps/config` real e qualquer `site.json` com texto privado. A tag AUD do Access não é um segredo, já que o JWT ainda precisa de uma assinatura válida. Mesmo assim ela identifica a sua instalação, junto com o domínio do time e os hostnames do admin, então mantenha a configuração privada também.

Faça commit só de `examples/`, com hostnames `example.com`.

## Notas de privacidade

- Com `DUMPS_GA_ID` definido, cada dump e a home carregam `googletagmanager.com` e reportam visualizações de página e os eventos `share` e `dumps_clean_mode`. Deixe vazio para desligar. O ID do GA4 (`G-...`) aparece em claro no HTML servido.
- O analytics começa com o Consent Mode do Google em negado: nenhum cookie de analytics é gravado até o visitante clicar em Aceitar num banner pequeno (em inglês ou português, seguindo o idioma da página). A escolha fica no `localStorage` (`dumps-consent`) e o banner não volta naquele navegador. Recusar mantém os cookies desligados; o Google ainda pode receber pings sem cookie para modelagem.
- As tags Open Graph incluem a URL do dump e, quando existirem, os nomes dos autores.
- Os botões de compartilhar abrem links de compartilhamento do WhatsApp e do Telegram ao clicar. Nada é enviado antes disso.

## Boas práticas

- Crie o app do Access antes do registro DNS do admin.
- Nunca aponte o hostname público para o listener do admin.
- Rode `curl -i` no host do admin sem login e confira que ele não devolve a listagem.
- Reinstale o `dumpsd` ao atualizar o repositório (`dumps install-dumpsd`). O CLI avisa quando a versão do servidor está defasada.

## Reportar uma vulnerabilidade

Siga o [SECURITY.md](../../SECURITY.md). Não abra issue pública para falhas de segurança.
