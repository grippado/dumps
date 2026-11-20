# Style packs

A style pack is a reusable presentation guide for `dump-this`. It carries the design tokens, voice, examples and composition rules that an agent applies to an artifact. A pack declares which content formats it supports and defines guidance for all four existing style dial levels.

The format is versioned independently from the skill. The first public example is [`field-notes.json`](../agents/skills/dump-this/style-packs/field-notes.json), a landing-page pack that can be previewed at [Style packs](https://grippado.github.io/dumps/style-packs/).

## Pack structure

- `metadata`: localized name and summary, author and license.
- `supports`: one or more of `report`, `story`, `slides`, `dashboard`, `doc`, `comparison`, `changelog` or `landing`.
- `designTokens`: a Design Tokens Community Group 2025.10 document for values such as color, typography and spacing.
- `voice`: writing principles with before/after examples.
- `guidance.global`: rules that apply at every dial level.
- `guidance.levels`: a direction, preview composition and rules for `classic`, `refined`, `bold` and `avant-garde`.
- `accessibility`: invariants that the style never overrides.
- `examples`: localized sample content for a supported format.
- `assets` (optional): bundled assets with their license and attribution.

The DTCG token format covers design values, not the complete style guide. Voice and composition stay in the pack so agents can apply them as instructions. A pack contains declarative data and prose; it does not execute JavaScript or CSS.

## Validate a pack

The reference validator uses the Python standard library. It checks the v1 contract, format compatibility, all four dial levels, preview tokens and WCAG AA contrast for the sample foreground colors.

```bash
python3 tools/validate-style-pack.py agents/skills/dump-this/style-packs/field-notes.json
```

The JSON Schema is at [`agents/skills/dump-this/style-packs/schema.json`](../agents/skills/dump-this/style-packs/schema.json). The preview currently renders one declared format, `landing`, at each dial level. The pack authoring agent and private organization libraries are later product slices.

## Contribute a public pack

Put the JSON pack under `agents/skills/dump-this/style-packs/`, run the validator, and open a pull request. Include a clear license for the pack and for every bundled asset. The preview renderer only accepts safe color and font token values; style packs provide no executable code.
