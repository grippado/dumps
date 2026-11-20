# Pacotes de estilo

Um pacote de estilo é um guia de apresentação reutilizável para a skill `dump-this`. Ele reúne tokens de design, voz, exemplos e regras de composição que o agente aplica ao artefato. O pacote declara quais formatos de conteúdo suporta e define instruções para os quatro níveis atuais do dial de estilo.

O formato tem versão própria, independente da skill. O primeiro exemplo público é [`field-notes.json`](../../agents/skills/dump-this/style-packs/field-notes.json), um pacote para landing pages que pode ser visualizado na [biblioteca de estilos](https://grippado.github.io/dumps/style-packs/).

## Estrutura do pacote

- `metadata`: nome e resumo em dois idiomas, autoria e licença.
- `supports`: um ou mais formatos entre `report`, `story`, `slides`, `dashboard`, `doc`, `comparison`, `changelog` e `landing`.
- `designTokens`: documento Design Tokens Community Group 2025.10 para valores como cores, tipografia e espaçamento.
- `voice`: princípios de escrita com exemplos de antes e depois.
- `guidance.global`: regras que valem em todos os níveis do dial.
- `guidance.levels`: direção, composição da prévia e regras para `classic`, `refined`, `bold` e `avant-garde`.
- `accessibility`: requisitos que o estilo nunca substitui.
- `examples`: conteúdo de exemplo, localizado, para um formato suportado.
- `assets` (opcional): recursos incluídos com licença e atribuição.

O formato DTCG cobre os valores de design, não o guia completo. Voz e composição ficam no pacote para que os agentes possam aplicá-las como instruções. Um pacote contém dados declarativos e prosa; ele não executa JavaScript nem CSS.

## Validar um pacote

O validador de referência usa apenas a biblioteca padrão do Python. Ele confere o contrato v1, a compatibilidade de formatos, os quatro níveis do dial, os tokens da prévia e o contraste WCAG AA das cores de texto do exemplo.

```bash
python3 tools/validate-style-pack.py agents/skills/dump-this/style-packs/field-notes.json
```

O JSON Schema fica em [`agents/skills/dump-this/style-packs/schema.json`](../../agents/skills/dump-this/style-packs/schema.json). A prévia atual renderiza um formato declarado, `landing`, nos quatro níveis do dial. O agente que cria pacotes e as bibliotecas privadas das organizações ficam para fatias futuras.

## Contribuir com um pacote público

Coloque o JSON em `agents/skills/dump-this/style-packs/`, rode o validador e abra um pull request. Declare uma licença para o pacote e para cada recurso incluído. A prévia só aceita valores seguros de cor e fonte; pacotes de estilo não fornecem código executável.
