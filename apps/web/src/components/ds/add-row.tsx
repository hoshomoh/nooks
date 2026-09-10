import { useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { cn } from "cn"

import type { AddRowChip, ChipKind } from "@/lib/add-row"
import type { DueDate } from "@/lib/dates"
import { useAddRowParse } from "@/lib/use-add-row-parse"
import { useDueLabel } from "@/lib/use-due-label"

/** The Item a finished row describes. */
export interface AddRowSubmission {
  label: string
  quantity: string
  dueOn: DueDate
}

export interface AddRowProps {
  /** States the defaults in words, e.g. "Add to Groceries, due today". */
  placeholder: string
  /**
   * The date this view gives a new Item — today on Today, tomorrow on Upcoming.
   *
   * It applies unless the Member sets or clears one, so the row files things where the
   * view says they belong without being told twice.
   */
  defaultDue?: DueDate
  onAdd: (item: AddRowSubmission) => void
  disabled?: boolean
}

/** What the row is holding, before it becomes an Item. */
interface AddRowDraft {
  /** The name so far: everything the parser has left alone. */
  name: string
  /** The word being typed now, after the chips. */
  typing: string
  /** Values lifted out of the sentence, in the order they were recognised. */
  chips: AddRowChip[]
  /** A date chosen in the control rather than typed. Null means neither. */
  picked: DueDate | null
  /** Whether the Member has cleared the view's date. */
  cleared: boolean
}

const EMPTY: AddRowDraft = { name: "", typing: "", chips: [], picked: null, cleared: false }

/**
 * The add row, per DESIGN.md §6: the same grid as a list row, a `+` where the checkbox
 * goes, the date control and a `↵` keycap on the right.
 *
 * A quantity and a date are lifted out of the sentence as it is typed — `Tomatoes 1kg
 * sat` files an Item called "Tomatoes" — but the date control is in the row whether or
 * not one was typed: parsing is the shortcut, not the requirement.
 *
 * Enter adds and keeps the focus, so a Member can type a whole shopping list without
 * touching the mouse.
 */
export function AddRow({ placeholder, defaultDue = "", onAdd, disabled }: AddRowProps) {
  const { t } = useTranslation()
  const parse = useAddRowParse()
  const due = useDueLabel()
  const [draft, setDraft] = useState<AddRowDraft>(EMPTY)
  const inputRef = useRef<HTMLInputElement | null>(null)

  const kinds = draft.chips.map((chip) => chip.kind)
  const hasText = Boolean(draft.name) || draft.chips.length > 0
  const typedDue = draft.chips.find((chip) => chip.kind === "due")?.label ?? ""
  const quantity = draft.chips.find((chip) => chip.kind === "quantity")?.label ?? ""
  // The control and the typed date are one value, and the last action wins: absorbing
  // a due chip drops any picked date, and picking one drops the chip.
  const chosenDue = draft.picked ?? typedDue
  const effectiveDue = chosenDue || (draft.cleared ? "" : defaultDue)

  /** Reads the sentence and lifts out whatever it recognises. */
  const absorb = (typing: string): AddRowDraft => {
    const parsed = parse(sentenceOf(draft.name, typing), kinds)
    if (parsed.chips.length === 0) {
      return { ...draft, typing }
    }
    const typedADate = parsed.chips.some((chip) => chip.kind === "due")
    return {
      ...draft,
      name: parsed.name,
      typing: "",
      chips: [...draft.chips, ...parsed.chips],
      picked: typedADate ? null : draft.picked,
      cleared: typedADate ? false : draft.cleared,
    }
  }

  const submit = () => {
    const parsed = parse(sentenceOf(draft.name, draft.typing), kinds)
    const label = parsed.name.trim()
    if (!label) {
      return
    }
    const settled = chosenDue || parsed.due
    onAdd({
      label,
      quantity: quantity || parsed.quantity,
      dueOn: settled || (draft.cleared ? "" : defaultDue),
    })
    setDraft(EMPTY)
  }

  /** Returns the last chip to the text, which is the only undo the row needs. */
  const unchip = () => {
    const last = draft.chips[draft.chips.length - 1]
    if (!last) {
      return
    }
    const rest = draft.chips.slice(0, -1)
    setDraft({
      ...draft,
      chips: rest,
      // The last one out takes the name back with it, so there is one field again.
      name: rest.length > 0 ? draft.name : "",
      typing: rest.length > 0 ? last.source : sentenceOf(draft.name, last.source),
    })
  }

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault()
        submit()
      }}
      className="mt-1 grid min-h-row grid-cols-[20px_1fr_auto] items-center gap-3.5 border-t border-hair px-2 py-1.5 -mx-2"
    >
      <span className="text-center text-[15px] text-muted-foreground">+</span>

      {/* The sentence and its chips are one field: the chips sit where the words were,
          and the caret carries on after them. */}
      <span className="flex min-w-0 flex-wrap items-center gap-2">
        {draft.name && <span className="text-body">{draft.name}</span>}
        {draft.chips.map((chip) => (
          <Chip
            key={`${chip.kind}-${chip.source}`}
            kind={chip.kind}
            label={chip.kind === "due" ? due.label(chip.label) : chip.label}
            kindLabel={t(chip.kind === "due" ? "date.dueChip" : "date.quantityChip")}
          />
        ))}
        <input
          ref={inputRef}
          value={draft.typing}
          onChange={(event) => setDraft(absorb(event.target.value))}
          onKeyDown={(event) => {
            if (event.key === "Backspace" && draft.typing === "") {
              event.preventDefault()
              unchip()
            }
            if (event.key === "Escape") {
              if (isEmpty(draft)) {
                inputRef.current?.blur()
              }
              setDraft(EMPTY)
            }
          }}
          placeholder={hasText ? "" : placeholder}
          disabled={disabled}
          className="min-w-24 flex-1 bg-transparent text-body outline-none placeholder:text-muted-foreground"
        />
      </span>

      <span className="flex items-center gap-1.5">
        <DateControl
          value={effectiveDue}
          label={effectiveDue ? due.label(effectiveDue) : t("date.addDate")}
          set={Boolean(chosenDue)}
          onChange={(next) => setDraft({ ...draft, picked: next || null, cleared: !next })}
        />
        <span className="rounded-sm border border-border px-1.5 py-px font-mono text-keycap text-muted-foreground">
          ↵
        </span>
      </span>
    </form>
  )
}

interface ChipProps {
  kind: ChipKind
  label: string
  /** The kind in the Member's language, e.g. "due". */
  kindLabel: string
}

/** A value the row recognised, drawn where the words it replaced were. */
function Chip({ kind, label, kindLabel }: ChipProps) {
  return (
    <span className="flex shrink-0 items-baseline gap-1.5 rounded-sm bg-shared-bg px-1.5 py-0.5">
      <span className={cn("text-small text-shared", kind === "quantity" && "font-mono")}>
        {label}
      </span>
      <span className="text-keycap text-muted-foreground">{kindLabel}</span>
    </span>
  )
}

interface DateControlProps {
  /** The date it currently carries, stored. */
  value: DueDate
  /** How that date reads, or the words for having none. */
  label: string
  /** Whether the Member chose this date, rather than the view supplying it. */
  set: boolean
  onChange: (value: DueDate) => void
}

/**
 * The date control, always in the row.
 *
 * It wraps a real date input rather than drawing a calendar: the browser's own picker
 * is already in the Member's language, already reachable from the keyboard, and already
 * knows what a month looks like where they live.
 */
function DateControl({ value, label, set, onChange }: DateControlProps) {
  return (
    <label
      className={cn(
        "flex h-6.5 shrink-0 cursor-pointer items-center gap-1.5 rounded-md px-2.5 text-micro",
        "transition-colors",
        set
          ? "bg-shared-bg text-shared"
          : "border border-border text-secondary-foreground hover:bg-secondary",
      )}
    >
      <CalendarGlyph />
      <span className="whitespace-nowrap">{label}</span>
      <input
        type="date"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        // The pill is the control; the input is what opens the picker and holds the
        // value. Hidden rather than removed, so the keyboard still reaches it.
        className="sr-only"
      />
    </label>
  )
}

/** The calendar glyph, drawn at the size the row uses. */
function CalendarGlyph() {
  return (
    <svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden className="shrink-0">
      <rect x="1" y="2.2" width="10" height="9" rx="1.6" stroke="currentColor" strokeWidth="1.1" />
      <path
        d="M1 5h10M3.6 1v2.2M8.4 1v2.2"
        stroke="currentColor"
        strokeWidth="1.1"
        strokeLinecap="round"
      />
    </svg>
  )
}

/** sentenceOf joins the name and what is being typed back into one line. */
function sentenceOf(name: string, typing: string): string {
  return [name, typing].filter(Boolean).join(" ")
}

/** isEmpty reports whether the row is holding nothing at all. */
function isEmpty(draft: AddRowDraft): boolean {
  return !draft.name && !draft.typing && draft.chips.length === 0 && !draft.picked
}
