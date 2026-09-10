import { cn } from "cn"

import { Checkbox } from "./checkbox"

export type ListRowNote = {
  /** The Note's own first line — never a summary. */
  firstLine: string
  /** How many further lines it has, for the "+N lines" after it. */
  remainingLines: number
}

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
  /** The Note attached to the Item, previewed under the row. */
  note?: ListRowNote
  /** How "+N lines" reads in the Member's language. */
  moreLinesLabel?: (count: number) => string
  onToggle?: (done: boolean) => void
  /** Opening the Item slides the sheet in; the List keeps its place behind it. */
  onOpen?: () => void
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
  note,
  moreLinesLabel,
  onToggle,
  onOpen,
}: ListRowProps) {
  return (
    <div className="flex flex-col">
    <div
      className={cn(
        "group grid min-h-row grid-cols-[20px_1fr_auto] items-center gap-3.5 rounded-md px-2 py-1.5 -mx-2",
        "hover:bg-secondary",
        justTicked && "bg-done-bg",
      )}
    >
      <Checkbox checked={done} justTicked={justTicked} onCheckedChange={onToggle} />

      <span className="flex min-w-0 items-baseline gap-2.5">
        <button
          type="button"
          onClick={onOpen}
          disabled={!onOpen}
          className="min-w-0 text-left"
        >
        <span
          className={cn(
            "truncate text-body",
            done && "text-muted-foreground line-through decoration-[#C4C4BE]",
          )}
        >
          {label}
        </span>
        </button>
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

      {/* The row shows the Note's own first line, truncated, and how much is left. */}
      {note && note.firstLine && (
        <div className="-mt-0.5 grid grid-cols-[20px_1fr] gap-3.5 px-2 pb-2 -mx-2">
          <span />
          <span className="flex min-w-0 items-baseline gap-2 border-l-2 border-border pl-2.5">
            <span className="truncate text-secondary text-secondary-foreground">
              {note.firstLine}
            </span>
            {note.remainingLines > 0 && moreLinesLabel && (
              <span className="shrink-0 text-micro text-muted-foreground">
                {moreLinesLabel(note.remainingLines)}
              </span>
            )}
          </span>
        </div>
      )}
    </div>
  )
}
