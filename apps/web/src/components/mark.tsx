/**
 * The Nooks mark: a list title, the hairline rule under it, and two items.
 *
 * Geometry is fixed by DESIGN.md §14 and must not be adjusted per use — pass a size
 * instead. Below 20px the mark simplifies to two paths, because the four-line version
 * fills in at that size; that is an optical size, not a scale.
 */
type MarkProps = {
  /** Rendered width and height in pixels. */
  size?: number
  /** Colour of the title and the hairline rule. Defaults to the current text colour. */
  ink?: string
  /** Colour of the two item lines. */
  items?: string
  className?: string
}

/** Below this width the mark uses its simplified two-path form. */
const SIMPLIFY_BELOW = 20

export function Mark({
  size = 24,
  ink = "currentColor",
  items = "var(--control)",
  className,
}: MarkProps) {
  const simplified = size < SIMPLIFY_BELOW

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      role="img"
      aria-label="nooks"
      className={className}
    >
      {simplified ? (
        <>
          <path d="M4 8h8" stroke={ink} strokeWidth="3" strokeLinecap="square" />
          <path d="M4 14h16" stroke={ink} strokeWidth="1.5" />
        </>
      ) : (
        <>
          <path d="M3 6h9" stroke={ink} strokeWidth="2.2" strokeLinecap="square" />
          <path d="M3 11h18" stroke={ink} strokeWidth="1" />
          <path d="M3 15.5h14M3 19.5h9" stroke={items} strokeWidth="1.6" strokeLinecap="square" />
        </>
      )}
    </svg>
  )
}
