# Proposta de roadmap Enterprise: object storage e Google Workspace

**Status:** proposta. O dumps ainda não implementa backend S3 nem login Google nativo.

Esta proposta amplia o fluxo Enterprise auto-hospedado com armazenamento de objetos e controle de
acesso pela organização. As duas capacidades podem ser entregues de forma independente.

## O que funciona hoje

- O CLI publica num filesystem local ou num filesystem remoto por SSH e rsync. O `dumpsd` monta a
  home, a listagem e o admin a partir de metadados no filesystem, sob um lock de escritor único.
- A origem dos artefatos serve HTML estático e fica pública, a menos que a organização aplique
  SSO, VPN ou uma rota privada diante da origem inteira.
- O admin pode validar identidade do Cloudflare Access ou do Tailscale. `DUMPS_AUTH` não protege
  a origem dos artefatos estáticos.
- A organização já pode usar Google Workspace como provedor de identidade pelo Cloudflare Access.
  O login Google e a política ficam na borda; o próprio dumps não valida tokens Google. Configure o
  Access separadamente nas origens dos artefatos e do admin.

Veja o [guia Enterprise atual](enterprise.md) e a
[integração Google Workspace no Cloudflare Access](https://developers.cloudflare.com/cloudflare-one/integrations/identity-providers/google-workspace/).

## Proposta: armazenamento de artefatos compatível com S3

Permitir que as organizações guardem o conteúdo dos artefatos e os arquivos estáticos gerados num
bucket S3 privado, mantendo `DUMPS_URL` como endereço estável para leitores. Uma origem S3 no
CloudFront pode permanecer privada com Origin Access Control (OAC); outro provedor precisa de um
mecanismo equivalente para proteger a origem. O bucket não pode ficar público diretamente.

O backend precisa preservar o comportamento do dumps: home e listagem geradas, estados oculto e
arquivado, atualização, remoção, rollback de ativação e garantia de escritor único. O desenho atual
depende de operações de filesystem local e `flock`; montar um bucket como filesystem não basta para
implementar isso. O contrato de storage e o modelo de ativação precisam de threat model e análise
de falhas antes da implementação.

Requisitos de design:

- Manter o backend de filesystem como padrão e preservar seus arquivos golden.
- Definir chaves de objetos, propriedade dos metadados, staging, ativação e limpeza após publishes
  interrompidos.
- Serializar mudanças de metadados ou usar uma estratégia comprovada de escrita condicional para
  evitar que publishers concorrentes percam atualizações da home ou da listagem.
- Manter o bucket privado atrás da origem de artefatos aprovada. Aplicar SSO, VPN ou rota privada na
  fronteira de leitura, não só diante do admin.
- Preferir credenciais temporárias por role ou workload, restritas ao prefixo do bucket. Não exigir
  chaves permanentes no perfil do agente.
- Cobrir upload, atualização, arquivamento, restauração, remoção, rollback, cache e recuperação de
  falhas parciais de rede.

As decisões em aberto incluem começar por AWS S3 ou definir um contrato S3-compatible mais amplo,
se o CLI envia diretamente ou entrega um pacote preparado a um serviço, e como o escritor de
metadados coordena a ativação. O Amazon S3 tem suporte a escritas condicionais com `If-Match` e
`If-None-Match`, que podem ser um componente; isso não define sozinho a transação nem os locks do
dumps. Veja [escritas condicionais no S3](https://docs.aws.amazon.com/AmazonS3/latest/userguide/conditional-writes.html)
e [acesso do CloudFront a uma origem S3 privada](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-restricting-access-to-s3.html).

## Proposta: acesso nativo com Google Workspace

A integração atual com Cloudflare Access é a opção mais próxima para login Google. Uma integração
nativa no dumps reduziria a necessidade de colocar esse provedor específico diante do serviço e
levaria a política de acesso para outros deploys auto-hospedados.

Manter permissões separadas:

1. **Leitor:** pode abrir a home e as páginas de artefatos protegidas.
2. **Publicador:** pode criar ou atualizar artefatos por um fluxo aprovado.
3. **Admin:** pode arquivar, renomear ou remover artefatos e alterar configurações do serviço.

O design deve decidir se a identidade Google Workspace protege a leitura dos artefatos, as ações de
admin ou ambas. Deve permitir uma política explícita de domínio da organização e avaliar política de
grupos como requisito de autorização separado. Autenticar no Google não deve, por si só, conceder
acesso.

Requisitos de segurança:

- Validar tokens de identidade no servidor, incluindo assinatura, emissor, audience e expiração.
- Ao restringir o acesso a domínios Workspace, usar a claim verificada de domínio hospedado (`hd`)
  do Google; não tratar apenas o sufixo do e-mail como prova de vínculo com o Workspace.
- Manter a fronteira de artefatos diante de todas as páginas e garantir que nenhuma rota direta ao
  bucket ou à origem contorne a autenticação.
- Usar sessões seguras e curtas, com logout, revogação e comportamento explícitos para mudanças de
  associação.
- Negar por padrão, separar permissões de leitura e escrita e evitar enviar credenciais ou conteúdo
  dos artefatos ao Google além do fluxo de identidade aprovado pela organização.

O [guia de autenticação backend do Google](https://developers.google.com/identity/sign-in/web/backend-auth)
descreve a validação de tokens no servidor e a claim de domínio hospedado. Se autorização por grupo
entrar no escopo, definir as permissões necessárias da API do Workspace e o consentimento de admin
antes de escolher a implementação.

## Fatias sugeridas e critérios de aceite

1. Threat model e revisão de design para acesso ao bucket, coordenação entre escritores, sessões
   Google, papéis, deploy entre origens e recuperação. Decidir o primeiro provedor de storage e a
   fronteira de identidade.
2. Adicionar uma interface de storage e uma implementação S3, mantendo o backend de filesystem e
   seus arquivos golden inalterados.
3. Implementar ativação atômica ou condicional, serialização de metadados, rollback, remoção e
   recuperação para o backend de objetos.
4. Documentar credenciais gerenciadas pela organização e exemplos de privilégio mínimo. Nunca
   colocar credenciais de nuvem ou hostnames da organização no pacote público da skill.
5. Adicionar autenticação Google Workspace e autorização para leitores, publicadores e admins na
   fronteira escolhida, ou manter Cloudflare Access como gateway de identidade para a primeira
   entrega com S3.
6. Documentar e verificar que leitores anônimos, domínios ou grupos não autorizados e usuários
   somente de leitura não acessem nem alterem artefatos protegidos. Testar publishers concorrentes
   e escritas interrompidas.

A camada de acesso à origem S3 e o login de usuários resolvem problemas diferentes: mantenha o
bucket inacessível diretamente e aplique a política de visualização da organização na origem de
leitura pública ou privada.
