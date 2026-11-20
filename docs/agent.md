# dump-this: a dump made by a specialist, in your favorite harness

`dump-this` is a skill that turns a request, a file or the current conversation into one polished,
self-contained HTML page. Before writing, it picks the specialist the job needs and follows that
specialist's recipe:

| Type | Specialist | What you get |
|---|---|---|
| report | senior analyst | TL;DR first, findings with evidence and confidence, recommendations |
| story | magazine editor | hero with key numbers, cast, real timeline, turning points |
| slides | presentation coach | one idea per slide, keyboard and swipe, notes, prints to PDF |
| dashboard | data visualization specialist | KPI tiles, the right chart per question in inline SVG, raw data |
| doc | technical writer | table of contents, steps with copy buttons, reference tables |
| comparison | decision advisor | verdict first, weighted criteria, option cards |
| changelog | release manager | grouped changes, breaking changes with migration |
| landing | product marketer | benefit headline, proof, features, call to action |

Every page also follows a shared base: light and dark themes, readable on a phone, accessible,
answer first, and real data only (it never invents numbers, dates or quotes).

**You do not need a dumps server to use it.** Without dumps, the skill hands you the HTML (or an
artifact or canvas, where the tool has one). With dumps, it also publishes the page to your own
domain and gives you the link.

The skill lives in [`agents/skills/dump-this`](../agents/skills/dump-this): `SKILL.md` plus the
recipes in `recipes/`.

## Claude Code

With the repo cloned:

```bash
cli/dumps install-agent claude
```

This links the skill to `~/.claude/skills/dump-this` and the command to
`~/.claude/commands/dump-this.md` (symlinks, so `git pull` updates them; `CLAUDE_CONFIG_DIR` is
respected). Without the CLI, copy the folder instead:

```bash
cp -R agents/skills/dump-this ~/.claude/skills/
```

Then, in any session:

```text
/dump-this a report of what changed in this PR
```

You can also just ask ("dump this as slides"): the skill activates on its own.

## Codex

```bash
cli/dumps install-agent codex
```

This links the skill to `~/.codex/skills/dump-this` and the prompt to
`~/.codex/prompts/dump-this.md` (`CODEX_HOME` is respected). Use it with `/prompts:dump-this`
followed by what you want, or ask for it in plain words.

## claude.ai (Skills)

1. Download [dump-this.zip](https://grippado.github.io/dumps/dump-this.zip) (built from this
   repo on every release of the landing page).
2. In claude.ai, open Settings, Capabilities, Skills, and upload the zip. Skills must be enabled
   for your account or organization.
3. In a chat, ask "dump this as a dashboard" or attach a file and ask for a report.

Claude builds the page as an HTML artifact, and you share it with the artifact's own publish
button. The output follows the same specialist recipes as in Claude Code.

## ChatGPT (custom GPT or Project)

ChatGPT does not load skill folders, so use the compact instructions instead:

1. Copy [dump-this-chatgpt.txt](https://grippado.github.io/dumps/dump-this-chatgpt.txt)
   (source: [`agents/chatgpt/instructions.md`](../agents/chatgpt/instructions.md)).
2. Paste it as the instructions of a custom GPT, or of a Project, so every chat in it uses them.
3. Ask for a dump. ChatGPT delivers the HTML in canvas, where you can preview and share it.

The compact version carries the same specialists and rules, condensed to fit the instruction
limits.

## Publishing with dumps

When the `dumps` command works on the machine, the skill publishes for you:

```bash
dumps publish <page> --title "<title>" --hidden --source agent
```

- The link works right away, but the dump stays off the public home page.
- In the admin listing the dump shows a "do agente" badge and a highlighted **Publicar** button:
  review the page, click it, and it goes to the home. From the terminal:
  `dumps unarchive <id>`.
- Ask for "on the home" or "public" to skip the review step.
- Republishing with the same `--id` updates the page in place.

## Tips for better dumps

- Name the type when you know it: "dump this as a comparison".
- Give it the data: attach the CSV, paste the notes, point to the file.
- Ask for a stable link: "use the id launch-deck".
- Iterate: "make the KPI row compare with last month" and republish with the same id.
