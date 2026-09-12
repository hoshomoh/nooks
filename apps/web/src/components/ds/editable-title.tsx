import { useRef, useState } from "react"
import { cn } from "cn"

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
  const inputRef = useRef<HTMLInputElement | null>(null)

  if (readOnly || !onCommit) {
    const Text = as
    return <Text className={className}>{value || placeholder}</Text>
  }

  const commit = () => {
    const next = (draft ?? value).trim()
    setDraft(null)
    if (next && next !== value) {
      onCommit(next)
    }
  }

  return (
    <input
      ref={inputRef}
      autoFocus={autoFocus}
      aria-label={label}
      // Preventing the mousedown is what stops the caret landing here, so the click
      // reaches the row behind. The second click of a double still arrives, which is
      // what puts the field in hand.
      onMouseDown={(event) => {
        if (clickTo === "open" && event.detail === 1) {
          event.preventDefault()
          onOpen?.()
        }
      }}
      onDoubleClick={() => {
        if (clickTo === "open") {
          inputRef.current?.focus()
          inputRef.current?.select()
        }
      }}
      placeholder={placeholder}
      value={draft ?? value}
      onChange={(event) => setDraft(event.target.value)}
      onBlur={commit}
      onKeyDown={(event) => {
        if (event.key === "Enter") {
          event.preventDefault()
          inputRef.current?.blur()
        }
        if (event.key === "Escape") {
          setDraft(null)
          inputRef.current?.blur()
        }
      }}
      // No border until it is being worked on: a title is text first and a field
      // second.
      className={cn(
        "w-full rounded-md bg-transparent px-1 -mx-1 outline-none",
        "transition-colors placeholder:text-muted-foreground hover:bg-secondary focus:bg-secondary",
        className,
      )}
    />
  )
}
