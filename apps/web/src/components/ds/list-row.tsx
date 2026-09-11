import type { ReactNode } from "react"
import { cn } from "cn"

import { Checkbox } from "./checkbox"
import { COVERING, INERT, RAISED } from "./covering"
import { DateField } from "./date-field"
import { EditableTitle } from "./editable-title"

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
  /** Renames the Item in place. Without it the label is read-only text. */
  onRename?: (label: string) => void
  /** Sets the quantity in place. Without it the quantity is a read-only badge. */
  onQuantityChange?: (quantity: string) => void
  /** Sets the due date in place. Without it the date is read-only text. */
  onDueChange?: (dueOn: string) => void
  /**
   * Which field something outside the row has asked to open.
   *
   * A menu entry called "Set a quantity" should put the caret in the quantity, not open
   * a panel somewhere else and leave the Member to find it.
   */
  editing?: ListRowField
  onEditingDone?: () => void
  /** The due date as stored, for the date control. */
  dueValue?: string
  /** What a screen reader calls the label field and the rest of the row. */
  labels: ListRowLabels
  /**
   * The row's own `···`, shown on hover.
   *
   * Passed in rather than built here: what can be done to an Item is the screen's
   * business, and a row that knew would need every mutation the screen has.
   */
  menu?: ReactNode
}

/** A field on the row that can be opened from outside it. */
export type ListRowField = "label" | "quantity" | "due"

/** The words the row needs that are not the Item's own. */
export interface ListRowLabels {
  /** Names the label field, e.g. "Item name". */
  name: string
  /** Names the area that opens the Item, e.g. "Open item". */
  open: string
  /** Names the quantity field. */
  quantity: string
  /** What an Item with no date offers, e.g. "Add a date". */
  due: string
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
  onRename,
  onQuantityChange,
  onDueChange,
  editing,
  onEditingDone,
  labels,
  menu,
  dueValue,
}: ListRowProps) {
  return (
    <div className="flex flex-col">
      <div
        className={cn(
          "group relative grid min-h-row grid-cols-[20px_1fr_auto] items-center gap-3.5 rounded-md px-2 py-1.5 -mx-2",
          "transition-colors duration-150 hover:bg-secondary",
          // The animation paints the highlight and takes it away again; the class
          // must not also paint it, or it would never settle.
          justTicked && "animate-settle",
        )}
      >
        {/* The whole row opens the Item — a 44px row is a target, and asking a Member
            to hit the label alone wastes most of it. It sits behind the row's own
            controls, so the checkbox still ticks and the label still edits. */}
        {onOpen && (
          <button type="button" onClick={onOpen} aria-label={labels.open} className={COVERING} />
        )}

        <Checkbox checked={done} justTicked={justTicked} onCheckedChange={onToggle} />

        <span className={cn(RAISED, "flex min-w-0 items-baseline gap-2.5")}>
          <EditableTitle
            value={label}
            onCommit={onRename}
            label={labels.name}
            as="span"
            autoFocus={editing === "label"}
            className={cn(
              "min-w-0 truncate text-body",
              done && "text-muted-foreground line-through decoration-[#C4C4BE]",
            )}
          />

          {/* The badge is the field. A quantity is one word, and asking for a panel to
              change one word is asking for more than the change is worth. */}
          {(quantity || editing === "quantity") && onQuantityChange ? (
            <EditableTitle
              value={quantity ?? ""}
              onCommit={(next) => {
                onQuantityChange(next)
                onEditingDone?.()
              }}
              label={labels.quantity}
              as="span"
              autoFocus={editing === "quantity"}
              placeholder={labels.quantity}
              className="w-16 shrink-0 rounded-sm border border-border px-1.5 py-px font-mono text-[11.5px] text-muted-foreground"
            />
          ) : (
            quantity && (
              <span className="shrink-0 rounded-sm border border-border px-1.5 py-px font-mono text-[11.5px] text-muted-foreground">
                {quantity}
              </span>
            )
          )}
        </span>

        {/* The metadata and the menu share the last column: the menu takes its place on
            hover rather than sitting beside it, because a row that widens as the
            pointer crosses it is a row nobody can aim at. */}
        <span className={cn(RAISED, "flex items-center justify-end gap-3")}>
          {onDueChange && (dueLabel || editing === "due") && (
            <DateField
              value={dueValue ?? ""}
              label={dueLabel || labels.due}
              chosen={Boolean(dueValue)}
              onChange={(next) => {
                onDueChange(next)
                onEditingDone?.()
              }}
              open={editing === "due" ? true : undefined}
              onOpenChange={(next) => !next && onEditingDone?.()}
            />
          )}
          <span className={cn(INERT, "flex items-center gap-3 whitespace-nowrap text-micro text-muted-foreground")}>
            {dueLabel && !onDueChange && (
              <span className={overdue ? "text-overdue" : "text-shared"}>{dueLabel}</span>
            )}
            {addedByName && <span>{addedByName}</span>}
          </span>

          {/* The `···` has a slot of its own, always the same size, so the row's
              metadata neither moves nor disappears when the pointer crosses it. Hidden
              by opacity rather than by display: a trigger that stops being rendered
              takes with it the element its menu is positioned against. */}
          {menu && (
            <span className="-mr-1 shrink-0 opacity-0 transition-opacity group-hover:opacity-100 focus-within:opacity-100 has-[[data-popup-open]]:opacity-100">
              {menu}
            </span>
          )}
        </span>
      </div>

      {/* The row shows the Note's own first line, truncated, and how much is left. */}
      {note && note.firstLine && (
        <div className="-mt-0.5 grid grid-cols-[20px_1fr] gap-3.5 px-2 pb-2 -mx-2">
          <span />
          <span className="flex min-w-0 items-baseline gap-2 border-l-2 border-border pl-2.5">
            <span className="truncate text-meta text-secondary-foreground">
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
