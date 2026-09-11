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
 */
export const DIALOG_SURFACE = "w-full max-w-dialog gap-0 rounded-2xl p-0 sm:max-w-dialog"
