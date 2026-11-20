---
name: dump-this
description: "Turn a request or the current conversation into a polished, self-contained HTML page made the way a specialist would make it (report, story timeline, slides, dashboard, doc, comparison, changelog, landing). Works in any harness: returns the HTML or an artifact, and publishes it with the dumps CLI when that is installed. Use when the user says dump this, dump-this, asks for a shareable page, an artifact, a report page, slides, a dashboard, or wants to publish or share content as a page."
---

# dump-this

Make one self-contained HTML page that looks like a specialist made it, then deliver it in the
way the current environment allows. A "dump" is that page.

This skill is opinionated: it exists to beat the default look of generated artifacts. Unless the
user asks for restraint, make something people would screenshot and forward. How far to push is
set by the **style dial** (classic, refined, bold, avant-garde), read from the user's own words.

## 1. Pick the recipe

| Type | Specialist | Use when |
|---|---|---|
| report | senior analyst | analysis, investigation, review, postmortem |
| story | magazine editor | how something happened over time |
| slides | presentation coach | a talk, a pitch, an update |
| dashboard | data visualization specialist | the numbers are the point |
| doc | technical writer | reference, how-to, runbook, spec |
| comparison | decision advisor | choosing between options |
| changelog | release manager | what shipped |
| landing | product marketer | presenting a project |

If nothing fits, use `doc`. If the user names a type, use it.

## 2. Set the style dial

Read `recipes/style.md` and pick the level from the request: classic ("conservative", "for the
board", "sóbrio"), refined ("clean", "minimal", "elegante"), **bold (the default when nothing is
said)**, or avant-garde ("experimental", "go wild", "master innovation", "inovação máxima",
"disruptivo"). An art direction in prose ("like a Swiss poster", "newspaper front page") is
honored too. Content and accessibility never change with the dial, only art direction.

## 3. Read the rules, then the recipe

Read `recipes/base.md` and then `recipes/<type>.md` from this skill's folder. Follow all three
(style, base, recipe). If the files are not available (for example when these instructions were
pasted into a chat), apply the summary below.

Summary when the recipe files are missing: one `index.html` with inline CSS and JS; light and
dark themes through `prefers-color-scheme`; readable at 360px wide with no horizontal scroll;
answer or key number first; real data only, never invented; no secrets or private data; a real
`<title>` and `<meta name="description">`; semantic headings and keyboard accessible controls.
Never the default-artifact look (centered gradient hero plus three cards, emoji icons, shadows on
everything, topic-only titles). At the default bold level: a display font with character,
oversized type, asymmetric editorial layout, a confident palette and one signature element made
from the content. End with a one-line colophon naming the level and the concept.

## 4. Build it

Write the page from the content the user gave or from the conversation. Ask one question only
if a fact the page depends on is missing; otherwise make the page and state the assumptions in
a small note at the bottom.

## 5. Deliver it

Choose the first option that applies:

1. **dumps is installed** (the command `dumps config` works): save the page to a temporary
   folder and publish it hidden, marked as made by an agent:

   ```bash
   dumps publish <folder-or-file> --title "<title>" --hidden --source agent
   ```

   The link works right away but the dump stays off the public home page; the owner reviews it
   and clicks **Publicar** in the admin to put it on the home. Drop `--hidden` only when the
   user explicitly asks for the home page. `--id <slug>` gives a readable, stable URL;
   republishing the same id updates it.
2. **The environment renders artifacts or canvases** (claude.ai, ChatGPT canvas and similar):
   output the complete HTML as an HTML artifact or canvas so the user can preview and share it
   with the platform's own publish button.
3. **Otherwise:** write `index.html` to the working directory (or print it in one fenced
   `html` block) and tell the user how to open it.

## 6. Report back

One short message: what was made (type, style level and title), where it is (URL, artifact or path), and,
when published with dumps, the follow-ups: `dumps unarchive <id>` or the Publicar button to
show it on the home, `dumps edit <id> --title "..."` to rename, `dumps rm <id>` to delete.
