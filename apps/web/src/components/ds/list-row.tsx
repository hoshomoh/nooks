import type { KeyboardEvent, ReactNode } from "react"
import { cn } from "cn"

import { initialOf } from "@/lib/initial"
import type { MoveDirection, Mover } from "@/lib/item-order"

import { Avatar } from "./avatar"
import { Checkbox } from "./checkbox"
import { COVERING, INERT, RAISED } from "./covering"
import { EditableTitle } from "./editable-title"
import { Icon } from "./icon"

export type ListRowProps = {
  label: string
  /** Free text — "2", "1 kg". Plain text on the right, capped so a long one ends in "…". */
  quantity?: string
  /** Whose circle sits at the right of the row: who added it, or who ticked it. */
  person?: RowPerson
  /**
   * Which List it is on, for a view that gathers Items from several.
   *
   * Beside the label rather than in the metadata column: away from its own List, where
   * an Item lives is part of reading the Item, not a fact filed after it.
   */
  listName?: string
  /** A date, already formatted for reading: "Fri", "Mon 1 Sep". */
  dueLabel?: string
  /** Overdue dates are the one thing in a row that takes a meaning colour. */
  overdue?: boolean
  /** When it was ticked, "08:12", in the place a date takes on an open row. */
  doneWhen?: string
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
  /**
   * What this row has to say about a change that did not land, under the row itself.
   *
   * DESIGN.md §11 puts it here rather than anywhere else: the question is about this
   * Item's words, and a banner at the top of the screen is a question about nothing in
   * particular.
   */
  trouble?: ReactNode
  /** Whether that trouble was a refusal, which the row itself is drawn for. */
  failed?: boolean
  /** The Item carries a Note. The Note itself is read in the sheet, not on the row. */
  hasNote?: boolean
  /** How the List lays its rows out, so the slots line up from one row to the next. */
  layout?: RowLayout
  onToggle?: (done: boolean) => void
  /** Opening the Item slides the sheet in; the List keeps its place behind it. */
  onOpen?: () => void
  /** Renames the Item in place. Without it the label is read-only text. */
  onRename?: (label: string) => void
  /** ⌥↑ or ⌥↓ pressed anywhere in the row. Absent where the order is not the Member's. */
  onMove?: Mover
  /** Picking the row up to drop it elsewhere. Absent where it cannot be moved. */
  drag?: RowDrag
  /** What a screen reader calls the label field and the rest of the row. */
  labels: ListRowLabels
  /**
   * The row's own `···`, shown on hover in the person's slot.
   *
   * Passed in rather than built here: what can be done to an Item is the screen's
   * business, and a row that knew would need every mutation the screen has.
   */
  menu?: ReactNode
}

/** What a row needs from the drag library to be picked up by its grip. */
export interface RowDrag {
  /** The row itself, which is what moves. */
  ref: (element: Element | null) => void
  /** The grip, the only part that picks the row up. */
  handleRef: (element: Element | null) => void
  /** This row is the one being carried. */
  lifted: boolean
  /** What the grip is called, to a screen reader and on hover. */
  label: string
}

/** A person drawn as a circle at the end of a row. */
export interface RowPerson {
  /** Whose letter the circle carries. */
  name: string
  /** What the circle says on hover and to a screen reader: "Anna via Kitchen tablet". */
  said: string
  /** It came in through an access token, which the circle marks with a spark. */
  viaToken?: boolean
}

/**
 * What the List decides for every row on it, so the slots line up down the page.
 *
 * DESIGN.md §6: only the person and the note keep their width. Everything else takes no
 * room on a row without it, so a label's width varies and the circles do not.
 */
export interface RowLayout {
  /** Some row on the List names a person, so every row keeps the slot. */
  people?: boolean
  /** Some Item on the List has a Note, so every row keeps the slot. */
  notes?: boolean
  /** The side sheet is open and the row is narrower, so the quantity gets less. */
  narrow?: boolean
  /** The completed section's shorter row. */
  dense?: boolean
}

/** The words the row needs that are not the Item's own. */
export interface ListRowLabels {
  /** Names the label field, e.g. "Item name". */
  name: string
  /** Names the area that opens the Item, e.g. "Open item". */
  open: string
  quantity: string
  /** What an Item with no date offers, e.g. "Add a date". */
  due: string
  /** What a change that has not reached the Instance is called, e.g. "not synced". */
  notSynced: string
}

/**
 * The list row — the core component; everything else is furniture around it.
 *
 * DESIGN.md §6: 44px however much metadata it carries. The grid is `20px 1fr auto` and
 * aligned to the top of a 24px line, so the label always starts at the same x, and
 * truncation happens in the label rather than in the metadata — a row never grows and
 * never reflows.
 */
export function ListRow({
  label,
  quantity,
  person,
  listName,
  dueLabel,
  overdue,
  doneWhen,
  done,
  justTicked,
  notSynced,
  trouble,
  failed,
  hasNote,
  layout = {},
  onToggle,
  onOpen,
  onRename,
  onMove,
  drag,
  labels,
  menu,
}: ListRowProps) {
  const personSlot = layout.people || person !== undefined || menu !== undefined
  const noteSlot = layout.notes || hasNote

  return (
    <div className="flex flex-col">
      <div
        ref={drag?.ref}
        onKeyDown={onMove && ((event) => moveOnArrow(event, onMove))}
        className={cn(
          "group relative grid grid-cols-[20px_1fr_auto] items-start gap-3.5 rounded-md border border-transparent px-2 -mx-2",
          layout.dense ? "min-h-9 py-1.5" : "min-h-row py-2.25",
          "transition-colors duration-150 hover:bg-secondary",
          // The animation paints the highlight and takes it away again; the class
          // must not also paint it, or it would never settle.
          justTicked && "animate-settle",
          // A refusal is the one thing that colours the row. A conflict does not: both
          // versions are somebody's words, and neither of them is an error.
          failed && "border-destructive-line bg-destructive-bg hover:bg-destructive-bg",
          // Carried: lifted off the List on the page's own ground, DESIGN.md §6.
          drag?.lifted && "cursor-grabbing border-border bg-background shadow-floating hover:bg-background",
        )}
      >
        {/* In the margin, outside the row, so showing it moves nothing inside the row. */}
        {drag && (
          <button
            type="button"
            ref={drag.handleRef}
            aria-label={drag.label}
            title={drag.label}
            className={cn(
              "absolute top-2.25 -left-5.5 grid h-6 w-4.5 cursor-grab place-items-center rounded-sm text-control opacity-0 transition-opacity",
              "group-hover:opacity-100 group-focus-within:opacity-100 focus-visible:opacity-100",
              drag.lifted && "cursor-grabbing text-secondary-foreground opacity-100",
            )}
          >
            <svg viewBox="0 0 8 14" className="h-3.5 w-2" fill="currentColor" aria-hidden>
              <circle cx="2" cy="2" r="1.3" />
              <circle cx="6" cy="2" r="1.3" />
              <circle cx="2" cy="7" r="1.3" />
              <circle cx="6" cy="7" r="1.3" />
              <circle cx="2" cy="12" r="1.3" />
              <circle cx="6" cy="12" r="1.3" />
            </svg>
          </button>
        )}

        {/* The whole row opens the Item — a 44px row is a target, and asking a Member
            to hit the label alone wastes most of it. It sits behind the row's own
            controls, so the checkbox still ticks and the label still edits. */}
        {onOpen && (
          <button type="button" onClick={onOpen} aria-label={labels.open} className={COVERING} />
        )}

        {/* Nothing to tick with is a checkbox that says what the Item is and does not
            pretend to be a control: a read-only List still shows what was done. */}
        <span className="grid h-6 place-items-center">
          <Checkbox
            checked={done}
            justTicked={justTicked}
            onCheckedChange={onToggle}
            disabled={!onToggle}
          />
        </span>

        <span className={cn(RAISED, "flex h-6 min-w-0 items-center gap-2.5")}>
          <EditableTitle
            value={label}
            onCommit={onRename}
            label={labels.name}
            as="span"
            // Where the row opens the Item, the first click does that and rewriting
            // the name is the second: asking a Member to find the gap beside the label
            // wastes most of a 44px target. Where there is nothing to open, a click
            // means what it always did.
            clickTo={onOpen ? "open" : "edit"}
            onOpen={onOpen}
            className={cn(
              "min-w-0 truncate leading-6",
              layout.dense ? "text-note" : "text-body",
              done && "text-muted-foreground line-through decoration-done-strike",
            )}
          />

          {listName && (
            <span className="shrink-0 text-micro text-secondary-foreground">{listName}</span>
          )}
        </span>

        {/* Status, date, quantity, then the two slots that keep their width. */}
        <span className={cn(RAISED, "flex h-6 items-center justify-end gap-3 whitespace-nowrap")}>
          <span className={cn(INERT, "flex min-w-0 items-center gap-3 text-micro text-muted-foreground")}>
            {notSynced && (
              <span className="shrink-0 rounded-sm border border-offline-line bg-offline-bg px-1.5 py-px text-offline-text">
                {labels.notSynced}
              </span>
            )}
            {dueLabel && (
              <span className={cn("shrink-0", overdue ? "text-overdue" : "text-shared")}>{dueLabel}</span>
            )}
            {doneWhen && <span className="shrink-0">{doneWhen}</span>}
            {quantity && (
              <span
                title={quantity}
                className={cn("min-w-0 truncate", layout.narrow ? "max-w-24" : "max-w-30")}
              >
                {quantity}
              </span>
            )}
          </span>

          {/* The `···` takes the person's place on hover rather than sitting beside it,
              because a row that widens as the pointer crosses it is a row nobody can aim
              at. Both stay mounted and trade by opacity: a trigger that stops being
              rendered takes with it the element its menu is positioned against. */}
          {personSlot && (
            <span className="relative grid h-6 w-6.5 shrink-0 place-items-center">
              {person && (
                <span
                  className={cn(
                    INERT,
                    "transition-opacity",
                    menu && "group-hover:opacity-0 group-focus-within:opacity-0 group-has-[[data-popup-open]]:opacity-0",
                  )}
                >
                  <PersonCircle person={person} />
                </span>
              )}
              {menu && (
                <span className="absolute inset-0 grid place-items-center opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100 group-has-[[data-popup-open]]:opacity-100">
                  {menu}
                </span>
              )}
            </span>
          )}

          {noteSlot && (
            <span className={cn(INERT, "grid h-6 w-4 shrink-0 place-items-center")}>
              {hasNote && (
                <Icon name="note" size="small" className="size-3.5 text-control" aria-hidden />
              )}
            </span>
          )}
        </span>
      </div>

      {/* Aligned with the label rather than the checkbox: the question is about the
          words, and it should start where the words start. */}
      {trouble && (
        <div className="grid grid-cols-[20px_1fr] gap-3.5 px-2 -mx-2">
          <span />
          {trouble}
        </div>
      )}
    </div>
  )
}

/** moveOnArrow answers ⌥↑ and ⌥↓, the keys DESIGN.md §12 gives to moving a row. */
function moveOnArrow(event: KeyboardEvent, onMove: Mover) {
  if (!event.altKey || event.metaKey || event.ctrlKey || event.shiftKey) {
    return
  }
  const direction = ARROWS[event.key]
  if (!direction) {
    return
  }
  event.preventDefault()
  onMove(direction)
}

const ARROWS: Partial<Record<string, MoveDirection>> = { ArrowUp: "up", ArrowDown: "down" }

interface PersonCircleProps {
  person: RowPerson
}

/** The small avatar a row ends in, with a spark when it came in through a token. */
function PersonCircle({ person }: PersonCircleProps) {
  return (
    <span role="img" title={person.said} aria-label={person.said} className="relative grid">
      <Avatar badge={initialOf(person.name)} size="small" />
      {person.viaToken && (
        <span
          aria-hidden
          className="absolute -right-0.75 -bottom-0.75 grid size-3 place-items-center rounded-full border-2 border-background bg-shared text-background"
        >
          <svg viewBox="0 0 6 6" className="size-1.5" fill="currentColor">
            <path d="M3 0l.8 2.2L6 3l-2.2.8L3 6l-.8-2.2L0 3l2.2-.8z" />
          </svg>
        </span>
      )}
    </span>
  )
}
