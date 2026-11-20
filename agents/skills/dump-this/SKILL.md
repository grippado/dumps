---
name: dump-this
description: "Create a self-contained HTML artifact (report, story timeline, slides, dashboard, doc, comparison, changelog, landing) and publish it with the dumps CLI to the user's own dumps site, returning the URL. Use when the user says dump this, dump-this, asks to dump, publish, share as a page, or turn content into a shareable HTML page."
---

# dump-this

Turn what the user asked for into one self-contained HTML artifact and publish it with
`dumps publish`. The result is a URL on the user's own dumps site.

## 1. Check the CLI

```bash
dumps config
```

If `dumps` is missing, tell the user to clone github.com/grippado/dumps and put `cli/dumps` on
the PATH, then stop. Read `url` from the output: it is where the dump will live.

## 2. Pick the recipe

| Type | When | Shape |
|---|---|---|
| report | analysis, investigation, review | title, TL;DR box, sections with findings, evidence tables, next steps |
| story | how something happened over time | hero, numbered timeline with real timestamps, cast of actors, "where we are now" |
| slides | a talk or a pitch | one idea per slide, full viewport sections, keyboard arrows and swipe to navigate |
| dashboard | numbers that matter | KPI tiles, inline SVG charts (no chart library needed), table with the raw data |
| doc | reference or how-to | sticky table of contents, headings with anchors, code blocks with copy buttons |
| comparison | choosing between options | criteria table, verdict per option, recommendation box |
| changelog | what shipped | versions newest first, grouped changes, links |
| landing | presenting a project | hero, features, how it works, call to action |

When the request fits none, use `doc`.

## 3. Build the HTML

- One `index.html`. A folder is fine when there are images: put them next to it and use relative
  paths. Never reference files outside the folder.
- Inline all CSS and JS. External requests only to well known CDNs when truly needed
  (fonts.googleapis.com, cdn.jsdelivr.net), pinned to an exact version.
- `<html lang>` set to the language of the content. Include `<title>` and
  `<meta name="description">`: dumps uses them for the bar and the link preview.
- Light and dark themes through `prefers-color-scheme`, readable on a phone (no horizontal
  scroll at 360px), accessible contrast, semantic headings.
- Real data only. Never invent numbers, dates or quotes. Mark anything estimated as estimated.
- Never include secrets, tokens, private emails, internal hostnames or customer data. If the
  source has them, remove them or ask before publishing.
- Do not add the dumps top bar or footer yourself: the publish step injects them.

Write it to a temporary folder, for example `$(mktemp -d)/index.html`.

## 4. Publish

Publish hidden by default: the link works, but the dump stays off the public home page. Put it on
the home only when the user asks for it (for example "on the home", "public", "show it on the site").

```bash
dumps publish <path> --title "<title>" --hidden --source agent
```

`--source agent` marks the dump as made by an agent: the admin listing shows a **Publicar**
button for it, so the user can review the page and put it on the home with one click.
Drop `--hidden` only when the user explicitly wants it on the home page. Use `--id <slug>` when
the user wants a stable, readable URL (lowercase letters, digits and hyphens). Republishing the
same `--id` updates the dump in place.

## 5. Report back

Reply with the URL, the recipe used, and whether it is hidden. Mention the follow-ups:

- `dumps unarchive <id>` puts it on the home page.
- `dumps edit <id> --title "..."` renames it.
- `dumps rm <id>` deletes it.
