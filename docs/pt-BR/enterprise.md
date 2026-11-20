# Publicação Enterprise e interna

A `dump-this` tem um modo Enterprise de entrega para organizações que querem artefatos num
endereço interno `dumps.<domínio-da-org>`. Ele combina segurança e entrega: o especialista e o
estilo pedidos continuam definindo a página. O modo usa o servidor dumps e a configuração de CLI
que a organização já mantém; ele não cria uma organização, um provedor de identidade nem um
endpoint de publicação.

Peça explicitamente ou diga que o material deve ficar dentro da empresa:

```text
/dump-this enterprise relatório do incidente para o time de segurança, publicar no dumps interno
```

## Como a publicação funciona

O CLI `dumps` prepara o HTML localmente e o envia ao servidor configurado por SSH e rsync.
`DUMPS_HOST` escolhe o host SSH e `DUMPS_URL` é o endereço entregue a quem vai ler. A máquina do
agente precisa ter o CLI, a configuração aprovada pela organização e acesso de rede ao servidor.
Veja um [guia de host](hosts/linux.md) para os detalhes da instalação.

Use uma configuração gerenciada em cada máquina de publicação. Por exemplo:

```text
DUMPS_HOST=org-dumps
DUMPS_URL=https://dumps.example.org
DUMPS_LIST_URL=https://admin-dumps.example.org/_dumps/
DUMPS_TARGET=linux/amd64
DUMPS_SERVICE=systemd-user
DUMPS_ADMIN_ORIGIN=https://admin-dumps.example.org
DUMPS_AUTH=access
DUMPS_ACCESS_TEAM=org.cloudflareaccess.com
DUMPS_ACCESS_AUD=<AUD da aplicação Access do admin>
```

`org-dumps` é um alias SSH configurado para o servidor aprovado. Mantenha hostnames da organização,
valores do Access, chaves e policies na configuração local gerenciada, fora deste repositório
público. Se a máquina tiver mais de um destino dumps, selecione um arquivo dedicado com
`DUMPS_CONFIG=/caminho/para/org/dumps.conf`. Confira `dumps config` antes de publicar.

Quem administra o servidor precisa configurar as origens de artefatos e do admin separadamente:

- `DUMPS_URL` serve a home e os artefatos HTML estáticos.
- `DUMPS_LIST_URL` serve a listagem e a API do admin.
- `DUMPS_AUTH=access` faz o `dumpsd` validar o JWT do Cloudflare Access para a API do admin. O AUD
  precisa corresponder à aplicação Access do admin.
- A rota SSH usada pelo CLI só pode ser alcançada pela rede ou policy aprovadas pela organização.

## Proteja os artefatos, não só o admin

Por padrão, os dumps estáticos são públicos. `--hidden` mantém um dump fora da home, mas não
adiciona autenticação: qualquer pessoa que alcance o host de artefatos e saiba o link pode ler a
página. Para material da empresa, aplique SSO corporativo ou VPN diante de toda a origem
`DUMPS_URL`, incluindo home e todos os artefatos, antes de publicar. Um hostname privado sem rota
pública também serve quando a organização tem DNS privado e VPN.

O admin tem uma fronteira de acesso separada. `DUMPS_AUTH=access` e `DUMPS_AUTH=tailscale` são os
modos de identidade que o `dumpsd` consegue validar para a API do admin. `DUMPS_AUTH=none` vale só
para o modo local e não pode ficar atrás de um proxy. O `dumpsd` não valida nativamente outros
proxies com identidade; nesse caso, mantenha a origem do admin sem exposição e use o CLI por SSH.
Veja o guia de [autenticação](auth.md), o guia de [Cloudflare](exposure/cloudflare.md) e as
[alternativas de exposição](exposure/alternatives.md).

Se a organização usa Google Workspace, o Cloudflare Access pode usá-lo como provedor de identidade.
Crie aplicações Access e policies separadas para toda a origem dos artefatos e para a origem do
admin. Mantenha `DUMPS_AUTH=access` no admin para o `dumpsd` validar o JWT do Access. Nesse fluxo,
o Cloudflare faz o login Google; o `dumpsd` não valida tokens Google diretamente. Veja o
[guia Google Workspace da Cloudflare](https://developers.cloudflare.com/cloudflare-one/integrations/identity-providers/google-workspace/).

Cada dump roda na mesma origem de navegador que os outros. O JavaScript de uma página consegue ler
os outros dumps daquele host, então SSO ou VPN não separa as equipes. Publique só HTML confiável,
restrinja quem pode publicar e use instâncias e origens de artefatos separadas para grupos que não
devem compartilhar essa fronteira de confiança. Mantenha `DUMPS_GA_ID` vazio e evite fontes,
scripts, analytics e outros pedidos a terceiros, a menos que a política de segurança da organização
os aprove.

Antes de publicar, confira se o `DUMPS_URL` ativo é exatamente a origem aprovada pela organização
e se a policy de acesso cobre o próprio host dos artefatos. Se o perfil, o domínio ou a policy não
estiverem claros, não publique material interno. Mantenha o HTML no workspace aprovado e peça ao
responsável pelo serviço a configuração gerenciada.

## Distribua a skill

Um backend de armazenamento S3-compatible e acesso Google Workspace nativo são propostas de
roadmap, ainda não recursos do dumps. Veja a [proposta de roadmap Enterprise](enterprise-roadmap.md).

No Claude Code e no Codex, instale `dump-this` junto do CLI com
[`cli/dumps install-agent`](agent.md). O pacote da skill inclui a receita Enterprise. A organização
pode distribuir o pacote por um espelho interno e acrescentar uma policy privada, mantendo
domínios, credenciais, dados de clientes e configurações do provedor de identidade fora do repo
público.

Skills do claude.ai e instruções do ChatGPT conseguem gerar o HTML, mas não executam o CLI local
nem alcançam uma rede privada por conta própria. A publicação direta desses harnesses exige um
conector aprovado pela organização. Sem ele, entregue o HTML apenas por um destino permitido pela
policy da empresa; não apresente um artefato hospedado como publicação no dumps interno.
