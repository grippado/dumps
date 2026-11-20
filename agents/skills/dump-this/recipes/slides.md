# Recipe: slides (think like a presentation coach)

Use for a talk, a pitch, a weekly update or a training.

1. **One idea per slide.** A slide is a full-viewport `<section>`. Title states the takeaway as a
   sentence ("Latency dropped 40% after the cache"), not a topic ("Latency").
2. **Arc.** Title slide, the problem, 3 to 7 content slides, the ask or conclusion, a closing
   slide with a link or contact.
3. **Navigation.** Arrow keys, Page Up/Down, space, and swipe on touch. A slide counter and a
   thin progress bar. `#n` in the URL keeps the current slide on reload.
4. **Content density.** At most 3 bullets or one chart or one big number per slide. Text large
   enough to read from the back of a room (at least 28px on desktop).
5. **Speaker notes.** Optional `<aside class="notes">` hidden by default, toggled with the `N`
   key.
6. **Print.** `@media print` puts one slide per page so the deck exports to PDF.

Avoid: paragraphs on slides, more than one accent color, animations that delay content.
