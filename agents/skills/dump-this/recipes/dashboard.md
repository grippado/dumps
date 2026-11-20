# Recipe: dashboard (think like a data visualization specialist)

Use when numbers are the point: metrics, KPIs, usage, costs, results.

1. **KPI row first.** 3 to 6 tiles: the value large, a label, the comparison (vs previous
   period, vs target) with direction and color that does not rely on color alone (arrow or
   sign).
2. **Choose the chart by the question.** Trend over time: line. Comparing categories: horizontal
   bar sorted by value. Part of a whole with few parts: stacked bar, not a pie. Distribution:
   histogram. Two variables: scatter. Never 3D, never dual axes.
3. **Draw with inline SVG.** No library needed for lines and bars: compute coordinates in a
   small script from a JSON array embedded in the page. Label axes, start bars at zero, show
   the data value on hover and as a `<title>` for screen readers.
4. **Color.** One series color plus a highlight color. Gray for context. Same color always
   means the same thing across charts.
5. **Raw data.** The full table below the charts, sortable by clicking a header, with units.
6. **Provenance.** "Source" and "as of" date under the title.

Avoid: more than 6 charts, legends far from the data, rounded numbers that hide changes.
