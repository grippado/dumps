# Segurança

Esta página descreve o modelo de ameaça do dumps como ele é hoje. Para reportar uma vulnerabilidade, veja o [SECURITY.md](../../SECURITY.md).

## Modelo

O dumps foi pensado para uma pessoa (ou um grupo pequeno e de confiança) que publica páginas suas e as mostra para o mundo.

1. **O host público serve só `dumps/`.** O Caddy de exemplo aponta o `root` do listener `:8081` para `<raiz>/dumps`. Manifests (`dumps-meta/`), staging, o binário (`bin/`) e a listagem (`admin/`) ficam fora dessa pasta. Arquivos e pastas que começam com ponto são escondidos (`hide .*` no Caddy e `hasHiddenSegment` no modo `--static`).
2. **O admin fica atrás de login.** O listener `:8080` só é alcançável pelo hostname do admin, protegido por Access ou Tailscale ([auth.md](auth.md)). O `dumpsd` só escuta em loopback.
3. **Escritor único.** Só o `dumpsd` escreve em `dumps/index.html`, na listagem e nos manifests, sempre sob `flock`. Quem publica não monta HTML da home à mão.
4. **Regras de caminho no CLI.**
   - ids casam com `^[a-z0-9][a-z0-9-]{0,62}$`, então não há `/`, `..` nem ponto no início. O id `_` é reservado.
   - `DUMPS_BLOCKED_PATHS` impede publicar o que estiver dentro de pastas sensíveis, ou o que as contenha (saída 2).
   - Symlinks dentro de uma pasta publicada precisam apontar para dentro dela; os que saem são recusados.
   - Arquivos soltos precisam ser `.html` ou `.htm`; `.git/` e `.DS_Store` não sobem.
   - `listing_dir` do `site.json` não pode ser absoluto, conter `..`, nem ser `dumps` ou `dumps-meta`.

## O que um atacante consegue e não consegue

| Situação | Consegue | Não consegue |
|---|---|---|
| Visitante anônimo | ver a home e os dumps cujo link conhece ou que estão na home; ver dumps arquivados se souber o id | ler manifests, listagem, staging ou `bin/`; chamar a API (403) |
| Página de terceiro no navegador de uma vítima logada | tentar um POST cross-site | passar pela checagem de `Origin` e de `Content-Type: application/json` |
| Quem forja um JWT | nada | assinar com a chave da Cloudflare; `alg: none` e HS256 são recusados |
| Token de outro app do Access, do mesmo time | nada | passar na checagem de `aud` |
| Token expirado ou de outro time | nada | passar em `exp` e `iss` |
| Quem tem ssh para o servidor | tudo (publicar, apagar, ler a raiz) | n/a: ssh é o nível de confiança máximo |

Atenção: arquivar tira o dump da home, mas o link continua funcionando. Arquivado não é privado. Para tirar do ar, use `dumps rm`.

## Publicar HTML de terceiros na mesma origem

Tudo que você publica roda na mesma origem que os outros dumps do host. Um dump com JavaScript malicioso consegue, para os visitantes:

- ler e escrever `localStorage` e cookies do host (a home guarda a preferência de visualização em `localStorage`, por exemplo);
- fazer requisições a qualquer outro dump do mesmo host, lendo as respostas;
- se passar por uma página sua (a barra e o rodapé são só HTML injetado, não uma fronteira de segurança), e fazer a reputação do domínio inteiro cair em listas de phishing.

O admin fica em outro hostname e a API exige `Origin` exata, então um dump malicioso não consegue acionar o admin por si só. Mesmo assim, se o hostname do admin for um subdomínio do mesmo domínio registrável dos dumps, cookies definidos pelo domínio pai podem interferir.

**Recomendação:** publique só o que é seu ou de quem você confia. Para conteúdo não confiável, use um domínio registrável separado (não apenas outro subdomínio) da home principal e do admin.

## HTML confiável no site.json

`tagline_html`, `host_card_html` e `footer_html` são inseridos crus na home pública ([configuration.md](configuration.md)). Trate o `site.json` como código: só você deve poder editá-lo.

## Segredos

Nunca entram em commit:

- o arquivo de credenciais do túnel (`~/.cloudflared/<UUID>.json`) e o `cert.pem` do `cloudflared tunnel login`;
- chaves privadas de ssh;
- o seu `~/.config/dumps/config` real, que tem hostnames, o domínio do time e o AUD.

O AUD do Access **não é um segredo**: ele só identifica o app, e o JWT continua precisando de assinatura válida. Mesmo assim, mantenha a sua configuração privada: o AUD, o domínio do time e os hostnames do admin dão mapa para quem quer atacar. O ID do GA4 (`G-...`) aparece em claro no HTML servido.

## Boas práticas

- Crie o app do Access antes do DNS do admin.
- Nunca aponte o hostname público para o listener do admin.
- Rode `curl -i` no host do admin sem login e confira que não devolve a listagem.
- Reinstale o `dumpsd` ao atualizar o repositório (`dumps install-dumpsd`); o CLI avisa quando a versão do servidor está defasada.

## Reportar uma vulnerabilidade

Siga o [SECURITY.md](../../SECURITY.md). Não abra issue pública para falhas de segurança.
