export type SectionHeadingProps = {
  /** The section's own label, e.g. "Overdue" or "Friday". */
  label: string
  /** A quiet note beside it, e.g. "since Friday" or "29 Aug". */
  note?: string
  /** Overdue is the one section that takes a meaning colour. */
  overdue?: boolean
}

/**
 * The rule above a group of Items, per DESIGN.md §6: an uppercase label, a quiet note,
 * and a hairline under both.
 */
export function SectionHeading({ label, note, overdue }: SectionHeadingProps) {
  return (
    <div className="mb-1.5 flex items-baseline gap-2.5 border-b border-hair pb-2">
      <span className={overdue ? "text-label uppercase text-overdue" : "text-label uppercase"}>
        {label}
      </span>
      {note && <span className="text-micro text-muted-foreground">{note}</span>}
    </div>
  )
}
