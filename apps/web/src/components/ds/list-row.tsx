import { cn } from "cn"

import { Checkbox } from "./checkbox"

export type ListRowProps = {
  label: string
  /** Free text — "2", "1 kg". Rendered as a mono badge beside the label. */
  quantity?: string
  /** Who put it on the List. Sits at the right of the row. */
  addedByName?: string
  /** A date, already formatted for reading: "Fri", "Mon 1 Sep". */
  dueLabel?: string
  /** Overdue dates are the one thing in a row that takes a meaning colour. */
  overdue?: boolean
  done?: boolean
  /** Someone else ticked it a moment ago. */
  justTicked?: boolean
  onToggle?: (done: boolean) => void
}

/**
 * The list row — the core component; everything else is furniture around it.
 *
 * DESIGN.md §6: 44px minimum however much metadata it carries. The grid is
 * `20px 1fr auto`, so the label always starts at the same x, and truncation happens in
 * the label rather than in the metadata — a row never grows and never reflows.
 */
export function ListRow({
  label,
  quantity,
  addedByName,
  dueLabel,
  overdue,
  done,
  justTicked,
  onToggle,
}: ListRowProps) {
  return (
    <div
      className={cn(
        "group grid min-h-row grid-cols-[20px_1fr_auto] items-center gap-3.5 rounded-md px-2 py-1.5 -mx-2",
        "hover:bg-secondary",
        justTicked && "bg-done-bg",
      )}
    >
      <Checkbox checked={done} justTicked={justTicked} onCheckedChange={onToggle} />

      <span className="flex min-w-0 items-baseline gap-2.5">
        <span
          className={cn(
            "truncate text-body",
            done && "text-muted-foreground line-through decoration-[#C4C4BE]",
          )}
        >
          {label}
        </span>
        {quantity && (
          <span className="shrink-0 rounded-sm border border-border px-1.5 py-px font-mono text-[11.5px] text-muted-foreground">
            {quantity}
          </span>
        )}
      </span>

      <span className="flex items-center gap-3 whitespace-nowrap text-micro text-muted-foreground">
        {dueLabel && <span className={overdue ? "text-overdue" : "text-shared"}>{dueLabel}</span>}
        {addedByName && <span>{addedByName}</span>}
      </span>
    </div>
  )
}
