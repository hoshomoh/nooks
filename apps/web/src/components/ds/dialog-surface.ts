/**
 * The one shape every dialog takes, per DESIGN.md §9: 560px, radius 10.
 *
 * Named once because the generated shadcn dialog sets a narrower width at the `sm`
 * breakpoint, and overriding it takes both the plain and the prefixed class. Passing
 * one leaves the dialog 560px on a phone and 384px on a desktop.
 *
 * Padding is zero because a Nooks dialog lays out its own header, body and footer.
 *
 * The height ceiling is here rather than in each dialog so the next one written gets
 * it: a dialog sized by its content grows past the viewport the moment somebody picks
 * enough things, taking the confirm button off the bottom. The body claims the space
 * with DIALOG_BODY.
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
