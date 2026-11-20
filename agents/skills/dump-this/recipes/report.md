# Recipe: report (think like a senior analyst)

Use for an analysis, investigation, audit, review or postmortem.

1. **Answer first.** A highlighted TL;DR box with 2 to 4 bullets: what happened, why it matters,
   what to do. A reader who stops here must have the gist.
2. **Context.** One short paragraph: scope, period, sources. Name the data sources.
3. **Findings.** One section per finding, ordered by impact. Each finding has: a one-line claim
   as the heading, the evidence (numbers, a small table, a quote with source), and the
   confidence level when it is not certain.
4. **Evidence tables.** Right-align numbers, use thousands separators, units in the header, no
   more than 7 columns. Highlight the row that matters.
5. **Recommendations.** Numbered, each with an owner or a next step when known. Separate
   "do now" from "consider later".
6. **Appendix.** Method, raw links, glossary. Collapsible with `<details>`.

Avoid: walls of text, findings without evidence, charts that only decorate.
