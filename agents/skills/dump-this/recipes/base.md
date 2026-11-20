# Base rules for every dump

Read this before any recipe. A dump is one self-contained page that someone opens from a link,
often on a phone, with no context. It has to work offline once loaded and look finished.

## Structure

- One `index.html`. Images or data next to it in the same folder, referenced by relative path.
- Inline CSS and JS. If a library is truly needed, load it from a major CDN pinned to an exact
  version (for example `https://cdn.jsdelivr.net/npm/chart.js@4.4.6`). Prefer no library.
- `<html lang>` matching the content, `<meta charset="utf-8">`, a viewport meta, a real `<title>`
  and `<meta name="description">` (one sentence). Link previews use them.

## Design system (small and consistent)

- CSS custom properties on `:root` for colors, redefined in
  `@media (prefers-color-scheme: dark)`. Always set an explicit `body` background.
- One accent color, one neutral scale, one muted text color. Contrast at least WCAG AA.
- Type: follow the style dial. Classic and refined use the system stack or one family; bold and
  avant-garde load one display family from fonts.googleapis.com with a system fallback. Type
  scale with `clamp()`: body 16 to 18px, headings clearly larger (much larger from bold up).
- Max content width around 760px for reading, up to 1120px for dashboards and grids.
- 16px side padding on phones, no horizontal scroll at 360px wide. Test mentally at 360, 768
  and 1280.
- Spacing in steps of 4 or 8. Create rhythm through contrast (dense next to airy, wide next to
  narrow), never the same card repeated down the page.

## Content

- Lead with the answer: a TL;DR, a verdict or the key number at the top.
- Real data only. Never invent numbers, dates, names or quotes. Write "estimated" or "not
  measured" when that is the truth.
- Remove secrets, tokens, private emails, internal hostnames and customer data before writing.
- Write in the user's language, with correct accents and punctuation.

## Finish

- Headings are semantic (`h1` once, then `h2`, `h3`). Links have meaningful text.
- Images have `alt`. Interactive controls are buttons, reachable by keyboard, with visible focus.
- `@media print` hides navigation and keeps content readable.
- Motion only with purpose, and none under `prefers-reduced-motion: reduce`.
- No console errors. No placeholder text left behind.
- A one-line colophon at the end: style level, concept, fonts.
