import { useState } from "react"

/**
 * The add row, per DESIGN.md §6: the same grid as a list row, a `+` where the checkbox
 * goes, and a `↵` keycap on the right.
 *
 * Enter adds and keeps the focus, so a Member can type a whole shopping list without
 * touching the mouse.
 */
export function AddRow({
  placeholder,
  onAdd,
  disabled,
}: {
  placeholder: string
  onAdd: (label: string) => void
  disabled?: boolean
}) {
  const [label, setLabel] = useState("")

  const submit = (event: React.FormEvent) => {
    event.preventDefault()
    const trimmed = label.trim()
    if (!trimmed) {
      return
    }
    onAdd(trimmed)
    setLabel("")
  }

  return (
    <form
      onSubmit={submit}
      className="mt-1 grid min-h-row grid-cols-[20px_1fr_auto] items-center gap-3.5 border-t border-hair px-2 py-1.5 -mx-2"
    >
      <span className="text-center text-[15px] text-muted-foreground">+</span>
      <input
        value={label}
        onChange={(event) => setLabel(event.target.value)}
        placeholder={placeholder}
        disabled={disabled}
        className="bg-transparent text-body outline-none placeholder:text-muted-foreground"
      />
      <span className="rounded-sm border border-border px-1.5 py-px font-mono text-keycap text-muted-foreground">
        ↵
      </span>
    </form>
  )
}
