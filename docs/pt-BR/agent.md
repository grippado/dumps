# dump-this: um dump feito por especialista, no seu harness favorito

`dump-this` é uma skill que transforma um pedido, um arquivo ou a conversa atual numa página HTML
caprichada e autocontida. Antes de escrever, ela escolhe o especialista que o trabalho pede e
segue a receita dele:

| Tipo | Especialista | O que você recebe |
|---|---|---|
| report | analista sênior | TL;DR primeiro, achados com evidência e confiança, recomendações |
| story | editor de revista | abertura com números-chave, elenco, linha do tempo real, viradas |
| slides | coach de apresentação | uma ideia por slide, teclado e swipe, notas, exporta para PDF |
| dashboard | especialista em visualização de dados | KPIs, o gráfico certo para cada pergunta em SVG inline, dados brutos |
| doc | redator técnico | sumário, passos com botão de copiar, tabelas de referência |
| comparison | conselheiro de decisão | veredito primeiro, critérios com peso, card por opção |
| changelog | gerente de release | mudanças agrupadas, quebras com o passo de migração |
| landing | profissional de marketing de produto | título com o benefício, prova, recursos, chamada para ação |

Toda página segue também uma base comum: tema claro e escuro, legível no celular, acessível,
resposta primeiro e só dados reais (a skill nunca inventa número, data ou citação).

**Você não precisa de um servidor dumps para usar.** Sem o dumps, a skill entrega o HTML (ou um
artefato ou canvas, onde a ferramenta tiver). Com o dumps, ela também publica a página no seu
domínio e devolve o link.

A skill mora em [`agents/skills/dump-this`](../../agents/skills/dump-this): o `SKILL.md` e as
receitas em `recipes/`.

## O dial de estilo: do conservador à inovação máxima

A dump-this é opinativa. Ela recusa o visual padrão de artefato (abertura centralizada com
gradiente, três cards iguais, emoji como ícone, sombra em tudo) e decide quanto ousar a partir das
suas próprias palavras:

| Nível | Como pedir | O que muda |
|---|---|---|
| 1 clássico | "sóbrio", "conservador", "para a diretoria" | grid rígido, quase preto no branco, um acento contido, sem animação |
| 2 refinado | "limpo", "minimalista", "elegante" | grotesca moderna, filetes no lugar de sombra, um detalhe perfeito |
| 3 ousado (padrão) | nada, ou "ousado" | tipografia de display, títulos enormes, layout editorial assimétrico, uma assinatura feita do conteúdo |
| 4 experimental | "pode chutar o balde", "experimental", "inovação máxima", "disruptivo" | um conceito forte primeiro (capa de jornal, painel de instrumentos, pôster suíço), SVG gerado dos dados, narrativa por rolagem |

Também dá para pedir uma direção de arte em prosa ("como um pôster suíço", "revista japonesa",
"estética de terminal"). Conteúdo e acessibilidade nunca mudam com o dial: contraste, teclado,
movimento reduzido e layout de celular valem em todo nível. Toda página termina com uma linha de
colofão com o nível e o conceito.

```text
/dump-this relatório sóbrio para a diretoria sobre os custos do trimestre
/dump-this slides para o meetup, com cara de pôster suíço
/dump-this dashboard do release, inovação máxima
```

Veja um exemplo de cada tipo e nível em
[grippado.github.io/dumps/examples](https://grippado.github.io/dumps/examples/).

## Claude Code

**Só a skill (sem clonar nada).** Rode num terminal qualquer; isso baixa a skill e descompacta na
pasta de skills do Claude Code:

```bash
mkdir -p ~/.claude/skills
```

```bash
curl -fsSL https://grippado.github.io/dumps/dump-this.zip -o /tmp/dump-this.zip
```

```bash
unzip -oq /tmp/dump-this.zip -d ~/.claude/skills
```

Abra uma sessão nova do Claude Code e digite:

```text
/dump-this um relatório do que mudou nesta PR
```

Também dá para só pedir ("faz um dump disso em slides"): a skill se ativa sozinha.

**Com o CLI do dumps (para também publicar no seu domínio).** Clone o repo uma vez e rode o
instalador de dentro da pasta do repo (`cli/dumps` é o script do CLI dentro do repo):

```bash
git clone https://github.com/grippado/dumps
```

```bash
cd dumps
```

```bash
cli/dumps install-agent claude
```

Isso cria symlinks de `agents/skills/dump-this` para `~/.claude/skills/dump-this` e do comando
para `~/.claude/commands/dump-this.md`, então um `git pull` no repo atualiza tudo
(`CLAUDE_CONFIG_DIR` é respeitado). Coloque o CLI no PATH para a skill conseguir publicar:
`ln -s "$PWD/cli/dumps" ~/.local/bin/dumps`.

## Codex

**Só a skill (sem clonar nada):**

```bash
mkdir -p ~/.codex/skills
```

```bash
curl -fsSL https://grippado.github.io/dumps/dump-this.zip -o /tmp/dump-this.zip
```

```bash
unzip -oq /tmp/dump-this.zip -d ~/.codex/skills
```

Abra uma sessão nova do Codex e peça um dump em palavras ("faz um dump disso como dashboard").

**Com o CLI do dumps:** de dentro da pasta do repo clonado (veja Claude Code acima), rode
`cli/dumps install-agent codex`. Ele cria o symlink da skill em `~/.codex/skills/dump-this` e o
prompt `/prompts:dump-this` em `~/.codex/prompts/dump-this.md` (`CODEX_HOME` é respeitado).

## claude.ai (Skills)

1. Baixe o [dump-this.zip](https://grippado.github.io/dumps/dump-this.zip) (montado a partir
   deste repo a cada publicação da landing).
2. No claude.ai, abra Configurações, Capacidades, Skills, e suba o zip. As Skills precisam estar
   habilitadas na sua conta ou organização.
3. Num chat, peça "faz um dump disso como dashboard" ou anexe um arquivo e peça um relatório.

O Claude monta a página como artefato HTML, e você compartilha pelo botão de publicar do próprio
artefato. A saída segue as mesmas receitas de especialista do Claude Code.

## ChatGPT (GPT personalizado ou Projeto)

O ChatGPT não carrega pastas de skill, então use as instruções compactas:

1. Copie o [dump-this-chatgpt.txt](https://grippado.github.io/dumps/dump-this-chatgpt.txt)
   (fonte: [`agents/chatgpt/instructions.md`](../../agents/chatgpt/instructions.md)).
2. Cole como instruções de um GPT personalizado, ou de um Projeto, para valer em todo chat dele.
3. Peça um dump. O ChatGPT entrega o HTML no canvas, onde você vê a prévia e compartilha.

A versão compacta leva os mesmos especialistas e regras, condensados para caber no limite de
instruções.

## Publicando com o dumps

Quando o comando `dumps` funciona na máquina, a skill publica por você:

```bash
dumps publish <página> --title "<título>" --hidden --source agent
```

- O link funciona na hora, mas o dump fica fora da home pública.
- Na listagem do admin o dump aparece com o selo "do agente" e um botão **Publicar** em destaque:
  revise a página, clique, e ela vai para a home. Pelo terminal: `dumps unarchive <id>`.
- Peça "na home" ou "público" para pular a revisão.
- Republicar com o mesmo `--id` atualiza a página no lugar.

## Dicas para dumps melhores

- Diga o tipo quando souber: "faz um dump disso como comparação".
- Dê os dados: anexe o CSV, cole as notas, aponte o arquivo.
- Peça um link estável: "usa o id launch-deck".
- Itere: "faz a linha de KPIs comparar com o mês passado" e republique com o mesmo id.
