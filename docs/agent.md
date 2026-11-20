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

**Publisher choice follows the active profile.** Set `DUMP_THIS_PUBLISHER` to `dumps`, `native`,
`file` or `auto`:

| Value | Delivery |
|---|---|
| `dumps` | Verify `dumps config`, publish through that configured server and return its URL. |
| `native` | Use the publisher exposed by the signed-in harness: Claude Artifacts, ChatGPT Sites, or another connected publisher. Never fall back to the dumps CLI. |
| `file` | Keep the HTML in the approved local workspace. |
| `auto` | Prefer the current harness's native publisher, then a verified dumps CLI, then an HTML file. |

A personal machine profile with CLI access can select its configured dumps server. A hosted work
profile can select the native publisher in the active Claude or ChatGPT account/workspace when
that harness exposes it. The skill never changes accounts or silently crosses from a native route
to a personal host. If the active harness cannot publish,
it leaves the HTML ready and reports that it was not published.

The skill lives in [`agents/skills/dump-this`](../agents/skills/dump-this): `SKILL.md` plus the
recipes in `recipes/`.

For the account and sharing limits of native publishers, see
[`recipes/publishing.md`](../agents/skills/dump-this/recipes/publishing.md).

## The style dial: from conservative to master innovation

dump-this is opinionated. It refuses the default artifact look (a centered gradient hero, three
identical cards, emoji icons, shadows on everything) and decides how far to push from your own
words:

| Level | Say it like | What changes |
|---|---|---|
| 1 classic | "conservative", "for the board", "sóbrio" | strict grid, near-black on white, one restrained accent, no motion |
| 2 refined | "clean", "minimal", "elegante" | modern grotesk, hairlines instead of shadows, one perfect detail |
| 3 bold (default) | nothing, or "ousado" | display type, oversized headlines, asymmetric editorial layout, a signature element made from the content |
| 4 avant-garde | "go wild", "experimental", "master innovation", "inovação máxima" | a strong concept first (newspaper front page, instrument panel, Swiss poster), generative SVG from the data, scroll-driven storytelling |

You can also name an art direction in prose ("like a Swiss poster", "Japanese magazine",
"terminal aesthetic"). Content and accessibility never move with the dial: contrast, keyboard,
reduced motion and phone layouts hold at every level. Every page ends with a one-line colophon
that names the level and the concept.

```text
/dump-this conservative report for the board about Q3 costs
/dump-this slides for the meetup, make it look like a Swiss poster
/dump-this dashboard of the release, go wild
```

See one example per type and level at
[grippado.github.io/dumps/examples](https://grippado.github.io/dumps/examples/).

## Style packs

A Style Pack v1 adds reusable design tokens, voice, examples and presentation rules to the built-in
style dial. Each pack keeps all four dial levels and declares the content formats it supports. The
[Field Notes preview](https://grippado.github.io/dumps/style-packs/) applies one landing-page pack at
all four levels. See the [pack format and validator](style-packs.md) to create a contribution.

## Claude Code

**Just the skill (nothing to clone).** Run this in any terminal; it downloads the skill and
unpacks it into Claude Code's skills folder:

```bash
mkdir -p ~/.claude/skills
```

```bash
curl -fsSL https://grippado.github.io/dumps/dump-this.zip -o /tmp/dump-this.zip
```

```bash
unzip -oq /tmp/dump-this.zip -d ~/.claude/skills
```

Open a new Claude Code session and type:

```text
/dump-this a report of what changed in this PR
```

You can also just ask ("dump this as slides"): the skill activates on its own.

**With the dumps CLI (to also publish to your own domain).** Clone the repo once, then run the
installer from the repo folder (`cli/dumps` is the CLI script inside the repo):

```bash
git clone https://github.com/grippado/dumps
```

```bash
cd dumps
```

```bash
cli/dumps install-agent claude
```

This symlinks `agents/skills/dump-this` to `~/.claude/skills/dump-this` and the command to
`~/.claude/commands/dump-this.md`, so a `git pull` in the repo updates them
(`CLAUDE_CONFIG_DIR` is respected). Put the CLI on your PATH so the skill can publish:
`ln -s "$PWD/cli/dumps" ~/.local/bin/dumps`.

## Codex

**Just the skill (nothing to clone):**

```bash
mkdir -p ~/.codex/skills
```

```bash
curl -fsSL https://grippado.github.io/dumps/dump-this.zip -o /tmp/dump-this.zip
```

```bash
unzip -oq /tmp/dump-this.zip -d ~/.codex/skills
```

Start a new Codex session and ask for a dump in plain words ("dump this as a dashboard").

**With the dumps CLI:** from the cloned repo folder (see Claude Code above), run
`cli/dumps install-agent codex`. It symlinks the skill to `~/.codex/skills/dump-this` and adds the
`/prompts:dump-this` prompt at `~/.codex/prompts/dump-this.md` (`CODEX_HOME` is respected).

## claude.ai (Skills)

1. Download [dump-this.zip](https://grippado.github.io/dumps/dump-this.zip) (built from this
   repo on every release of the landing page).
2. In claude.ai, open Settings, Capabilities, Skills, and upload the zip. Skills must be enabled
   for your account or organization.
3. In a chat, ask "dump this as a dashboard" or attach a file and ask for a report.

Claude builds the page as an HTML Artifact in the signed-in account. Sharing follows that
account's controls; Claude for Work Artifacts stay within the organization. The output follows
the same specialist recipes as in Claude Code.

## ChatGPT (custom GPT or Project)

ChatGPT does not load skill folders, so use the compact instructions instead:

1. Copy [dump-this-chatgpt.txt](https://grippado.github.io/dumps/dump-this-chatgpt.txt)
   (source: [`agents/chatgpt/instructions.md`](../agents/chatgpt/instructions.md)).
2. Paste it as the instructions of a custom GPT, or of a Project, so every chat in it uses them.
3. Ask for a dump. Use ChatGPT Sites in the active workspace when available and selected; otherwise
   ChatGPT delivers the HTML in canvas, where you can preview it.

The compact version carries the same specialists and rules, condensed to fit the instruction
limits.

## Publishing with dumps

When the active profile selects `DUMP_THIS_PUBLISHER=dumps` and the `dumps` command verifies the
intended server, the skill publishes for you:

```bash
dumps publish <page> --title "<title>" --hidden --source agent
```

- The link works right away, but the dump stays off the public home page.
- In the admin listing the dump shows a "do agente" badge and a highlighted **Publicar** button:
  review the page, click it, and it goes to the home. From the terminal:
  `dumps unarchive <id>`.
- Ask for "on the home" or "public" to skip the review step.
- Republishing with the same `--id` updates the page in place.

## Enterprise and internal publishing

For an organization's internal `dumps.<org-domain>` service, use the enterprise delivery overlay in
[`recipes/enterprise.md`](../agents/skills/dump-this/recipes/enterprise.md) and the
[enterprise setup guide](enterprise.md). It uses an organization-managed CLI profile and server;
`--hidden` is not access control. The guide covers corporate SSO or VPN protection for the full
artifact host, admin authentication and distributing the skill. Browser-only harnesses need an
approved connector before they can publish to a private service.

## Tips for better dumps

- Name the type when you know it: "dump this as a comparison".
- Give it the data: attach the CSV, paste the notes, point to the file.
- Ask for a stable link: "use the id launch-deck".
- Iterate: "make the KPI row compare with last month" and republish with the same id.
