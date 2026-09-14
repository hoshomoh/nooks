import { useCallback, useRef, useState } from "react"
import type { ChangeEvent, KeyboardEvent, MouseEvent } from "react"
import { cn } from "cn"

/** The element a title is edited in: one line for a row, wrapping for a heading. */
type TitleField = HTMLInputElement | HTMLTextAreaElement

/** Which element the text is drawn as when it cannot be changed. */
export type TitleElement = "h1" | "h2" | "span"

export interface EditableTitleProps {
  /** What the title currently reads, as stored. */
  value: string
  /** Called with the new text once the Member has committed it. */
  onCommit?: (value: string) => void
  /** Read-only renders the text and nothing else. Implied when nothing can be saved. */
  readOnly?: boolean
  /** The element to render when read-only, so a row is not given a heading. */
  as?: TitleElement
  /** The type token the text is drawn at, so a row and a page can differ. */
  className?: string
  /** What a screen reader calls the field. */
  label: string
  /** What an empty field reads as, e.g. "Add". Without it, empty is invisible. */
  placeholder?: string
  /** Takes the caret on mount, for a field something else asked to open. */
  autoFocus?: boolean
  /**
   * What a single click on the text does.
   *
   * "edit" puts the caret in it, which is right for a page title: the title is the
   * only thing there. "open" is for a title inside a row that does something — a
   * single click does the row's job and a double click rewrites the title, because a
   * row somebody has to aim at the gaps of is a row with no target.
   */
  clickTo?: "edit" | "open"
  /** What "open" means. Required when clickTo is "open", ignored otherwise. */
  onOpen?: () => void
}

/**
 * A title that is also the field that changes it.
 *
 * An Item's name is one line of text, and the design gives it no separate edit mode:
 * the title is where it is read and where it is rewritten, in the row, the sheet and
 * the full-screen view alike. Enter and blur commit; Escape puts back what was there.
 *
 * The stored value is read into the field only when nobody is typing in it, so a save
 * coming back never moves the caret. That is why the draft is held in state rather than
 * synchronised to the prop with an effect.
 */
export function EditableTitle({
  value,
  onCommit,
  readOnly,
  as = "h1",
  className,
  label,
  placeholder,
  autoFocus,
  clickTo = "edit",
  onOpen,
}: EditableTitleProps) {
  const [draft, setDraft] = useState<string | null>(null)
  const fieldRef = useRef<TitleField | null>(null)

  /*
   * A title wraps; a row's label does not.
   *
   * They are the same control doing two jobs. In the sheet and full screen the name is
   * a heading and has to be readable in full, so it grows down the page. In a list row
   * it cannot: DESIGN.md §6 says a row never grows and never reflows, because a list
   * whose rows change height as somebody types is a list nobody can aim at.
   */
  const wraps = as !== "span"

  // A callback ref rather than one written into the JSX: the same assignment inline
  // reads to a linter as touching a ref while rendering, which this is not.
  const holdField = useCallback((node: TitleField | null) => {
    fieldRef.current = node
  }, [])

  if (readOnly || !onCommit) {
    const Text = as
    // Inert where the row has a job of its own: text that answers nothing is text in
    // the way of the covering element behind it. See covering.tsx.
    return (
      <Text className={cn(className, clickTo === "open" && "pointer-events-none")}>
        {value || placeholder}
      </Text>
    )
  }

  const commit = () => {
    const next = (draft ?? value).trim()
    setDraft(null)
    if (next && next !== value) {
      onCommit(next)
    }
  }

  const field = {
    autoFocus,
    "aria-label": label,
    placeholder,
    value: draft ?? value,
    onBlur: commit,
    // Three separate jobs, and they are separate on purpose. The mousedown stops the
    // caret landing here so a single click means the row. The click is what actually
    // does the row's job. The double click hands the field over — and does not select
    // anything, because somebody who double-clicked a word wants to edit at that word,
    // not to replace the whole line.
    onMouseDown: (event: MouseEvent<TitleField>) => {
      if (clickTo === "open" && event.detail === 1) {
        event.preventDefault()
      }
    },
    onClick: (event: MouseEvent<TitleField>) => {
      if (clickTo === "open" && event.detail === 1) {
        onOpen?.()
      }
    },
    onDoubleClick: () => {
      if (clickTo === "open") {
        fieldRef.current?.focus()
      }
    },
    onChange: (event: ChangeEvent<TitleField>) => setDraft(event.target.value),
    onKeyDown: (event: KeyboardEvent<TitleField>) => {
      // Enter commits rather than writing a line: a name is one line however it wraps.
      if (event.key === "Enter") {
        event.preventDefault()
        fieldRef.current?.blur()
      }
      if (event.key === "Escape") {
        setDraft(null)
        fieldRef.current?.blur()
      }
    },
  }

  // No border until it is being worked on: a title is text first and a field second.
  const look = cn(
    // No outline-none here. DESIGN.md §7 gives one focus treatment for everything
    // focusable, and a title that is also a field is focusable — the background fill
    // below says "you are editing this", which is not the same as saying "the keyboard
    // is here".
    "w-full rounded-md bg-transparent px-1 -mx-1",
    "transition-colors placeholder:text-muted-foreground hover:bg-secondary focus:bg-secondary",
    // A single click opens the row, so the pointer says so. An I-beam over something
    // that does not take the caret is the control lying about itself.
    clickTo === "open" && "cursor-pointer focus:cursor-text",
    className,
  )

  if (!wraps) {
    return <input ref={holdField} {...field} className={look} />
  }

  return (
    <textarea
      ref={holdField}
      rows={1}
      // Grows to its content instead of scrolling it out of sight, so a long name is
      // read rather than hunted for. `resize-none` because the size is the text's to
      // decide, not a corner handle's.
      {...field}
      className={cn(look, "field-sizing-content resize-none overflow-hidden")}
    />
  )
}
