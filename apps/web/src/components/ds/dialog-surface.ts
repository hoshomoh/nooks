/**
 * The one shape every dialog takes, per DESIGN.md §9: 560px, radius 10.
 *
 * Named once because the generated shadcn dialog sets its own narrower width at the
 * `sm` breakpoint, and overriding it takes both the plain and the prefixed class —
 * passing only one leaves the dialog 560px on a phone and 384px on a desktop, which is
 * exactly the inconsistency this constant exists to prevent.
 *
 * Padding is zero because a Nooks dialog lays out its own header, body and footer
 * against its own edges.
 *
 * It is also a column with a ceiling. A dialog whose content decides its height stops
 * being a dialog the moment somebody picks enough things: it grows past the viewport,
 * the footer with the confirm button in it goes off the bottom, and what was a
 * question becomes a page. The ceiling is here rather than in each dialog so that the
 * next one written gets it without anybody remembering — the body inside claims the
 * space with DIALOG_BODY.
 */
export const DIALOG_SURFACE =
  "flex max-h-[min(85vh,40rem)] w-full max-w-dialog flex-col gap-0 rounded-2xl p-0 sm:max-w-dialog"

/**
 * The scrolling middle of a dialog, between a pinned header and a pinned footer.
 *
 * `min-h-0` is what makes it scroll rather than push: a flex child's floor is its
 * content unless it is told otherwise, so without it the column grows and the ceiling
 * above does nothing.
 */
export const DIALOG_BODY = "min-h-0 flex-1 overflow-y-auto"
