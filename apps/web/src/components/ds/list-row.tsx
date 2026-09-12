import type { ReactNode } from "react"
import { cn } from "cn"

import { Checkbox } from "./checkbox"
import { COVERING, INERT, RAISED } from "./covering"
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
  /**
   * Which List it is on, for a view that gathers Items from several.
   *
   * Beside the label rather than in the metadata column: away from its own List, where
   * an Item lives is part of reading the Item, not a fact filed after it.
   */
  listName?: string
  /**
   * What it came through, when that was not a browser: "Anna · via Kitchen tablet".
   *
   * The Member is still named. A token is somebody's access narrowed, never an identity
   * of its own, so a row says who did it and only then what through.
   */
  addedVia?: string
  /** A date, already formatted for reading: "Fri", "Mon 1 Sep". */
  dueLabel?: string
  /** Overdue dates are the one thing in a row that takes a meaning colour. */
  overdue?: boolean
  done?: boolean
  /** Someone else ticked it a moment ago. */
  justTicked?: boolean
  /**
   * The change to this row has not reached the Instance yet.
   *
   * A tick made with no connection looks exactly like a saved one, which would be the
   * app claiming something it has not done. The row says so instead, and keeps saying
   * so until the change lands.
   */
  notSynced?: boolean
  /** The Note attached to the Item, previewed under the row. */
  note?: ListRowNote
  /** How "+N lines" reads in the Member's language. */
  moreLinesLabel?: (count: number) => string
  onToggle?: (done: boolean) => void
  /** Opening the Item slides the sheet in; the List keeps its place behind it. */
  onOpen?: () => void
  /** Renames the Item in place. Without it the label is read-only text. */
  onRename?: (label: string) => void
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
  /** What a change that has not reached the Instance is called, e.g. "not synced". */
  notSynced: string
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
  addedVia,
  listName,
  dueLabel,
  overdue,
  done,
  justTicked,
  notSynced,
  note,
  moreLinesLabel,
  onToggle,
  onOpen,
  onRename,
  labels,
  menu,
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
            className={cn(
              "min-w-0 truncate text-body",
              done && "text-muted-foreground line-through decoration-[#C4C4BE]",
            )}
          />

          {quantity && (
            <span className="shrink-0 rounded-sm border border-border px-1.5 py-px font-mono text-[11.5px] text-muted-foreground">
              {quantity}
            </span>
          )}

          {listName && (
            <span className="shrink-0 text-micro text-secondary-foreground">{listName}</span>
          )}

          {notSynced && (
            <span className="shrink-0 rounded-sm border border-offline-line bg-offline-bg px-1.5 py-px text-micro text-offline-text">
              {labels.notSynced}
            </span>
          )}
        </span>

        {/* The metadata and the menu share the last column: the menu takes its place on
            hover rather than sitting beside it, because a row that widens as the
            pointer crosses it is a row nobody can aim at. */}
        <span className={cn(RAISED, "flex items-center justify-end gap-3")}>
          <span className={cn(INERT, "flex items-center gap-3 whitespace-nowrap text-micro text-muted-foreground")}>
            {dueLabel && (
              <span className={overdue ? "text-overdue" : "text-shared"}>{dueLabel}</span>
            )}
            {addedByName && <span>{addedByName}</span>}
            {addedVia && <span className="text-muted-foreground">{addedVia}</span>}
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
