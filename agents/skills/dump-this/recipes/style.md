# The style dial

dump-this has opinions. Its job is to make pages people screenshot and forward, not pages that
look like every other generated artifact. The dial sets how far to push. Content and
accessibility never move with the dial: only the art direction does.

## Read the level from the request

| Level | Say it like | Default for |
|---|---|---|
| 1 classic | "conservative", "corporate", "safe", "for the board", "sóbrio", "conservador", "formal" | compliance, finance, legal |
| 2 refined | "clean", "polished", "minimal", "like a product page", "elegante", "limpo" | docs, changelogs |
| 3 bold | nothing said, "make it nice", "ousado", "com personalidade" | **everything else (the default)** |
| 4 avant-garde | "go wild", "experimental", "break the rules", "master innovation", "inovação máxima", "pode chutar o balde", "disruptivo" | when asked |

The user can also give an art direction in prose ("like a Swiss poster", "Japanese magazine",
"terminal aesthetic", "newspaper front page", "museum catalogue"). Honor it at the level it implies
(a named movement or medium is level 3 or 4). Combine: "level 4, brutalist" is valid. When in
doubt between two levels, pick the bolder one.

## Never, at any level (the default-artifact look)

- A centered hero with a purple to blue gradient, followed by three identical cards.
- Emoji as icons or bullets. Stock rounded cards with drop shadows on everything.
- Every section the same width, the same padding, the same card. Rhythm comes from contrast.
- Gray text on gray. Lorem ipsum. "Welcome to". Titles that name a topic instead of saying
  something ("Overview", "Results").
- Decorative charts with no question behind them.

## Level 1: classic

- Typography: one sans or a classic serif for headings, 2 to 3 sizes only, generous line height.
- Color: near-black on white (and the inverse in dark), one restrained accent (navy, oxblood,
  forest), no gradients.
- Layout: a strict single column or a 12-column grid, left aligned, thin rules as dividers.
- Motion: none.
- Signature: impeccable tables and alignment. It should look like a top consultancy printed it.

## Level 2: refined

- Typography: a modern grotesk, tight letter spacing on large headings, tabular numbers.
- Color: a neutral scale with warmth, one accent used sparingly, hairline borders instead of
  shadows.
- Layout: generous whitespace, asymmetric two-column moments, content max width around 720px.
- Motion: subtle (opacity and small translate on enter), respects reduced motion.
- Signature: one precise detail done perfectly (a beautiful KPI, a crafted timeline marker).

## Level 3: bold (default)

- Typography: a display face with character (load one family from fonts.googleapis.com, with a
  system fallback) for oversized headings, paired with a quiet text face. Headlines at 10vw or
  more are welcome. Mix weights and italics with intent.
- Color: a confident palette of 2 to 3 colors with one surprising accent; full-bleed color
  blocks; dark and light both designed, not inverted.
- Layout: editorial and asymmetric. Break the grid on purpose: a number that spills over a
  section edge, a pull quote in the margin, a sticky element that tells the story while the
  page scrolls.
- Motion: purposeful (numbers counting up, lines drawing in, sections revealing on scroll),
  always disabled under `prefers-reduced-motion`.
- Signature: one memorable element that only this page has, ideally made from the content
  itself (a timeline drawn as a skyline, a KPI as a giant type sculpture).

## Level 4: avant-garde

- Pick a strong concept first and state it in one line at the top of your thinking ("this report
  is a newspaper front page", "these slides are a Swiss poster series", "this dashboard is an
  instrument panel"). Everything follows the concept.
- Allowed: brutalist or Swiss poster layouts, kinetic and variable typography, generative SVG
  art computed from the actual data, horizontal or scroll-driven storytelling, unusual
  navigation (keyboard-first, scrubbing, a map instead of a menu), monospace or display-only
  type, raw HTML aesthetics on purpose.
- Still required: the content is readable and complete without the effects, contrast passes AA,
  everything works with a keyboard, motion stops under reduced motion, no horizontal scroll on
  phones (horizontal storytelling must fall back to vertical below 700px).
- Signature: the page itself is the idea. Someone should be able to describe it in one sentence
  after closing it.

## Say what you chose

End the page with a tiny colophon line: level, concept or art direction, fonts. Example:
"Level 3 bold · concept: the build as a skyline · Fraunces + Inter". It teaches the reader the
dial exists.
