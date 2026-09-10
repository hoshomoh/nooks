import type { ReactNode } from "react"

/**
 * A quiet panel stating a fact — who has your request, when it was sent. Per
 * DESIGN.md §8's card: a fact, not an alert, and never a warning triangle.
 */
export type NotePanelProps = {
  label: string
  children: ReactNode
}

export function NotePanel({ label, children }: NotePanelProps) {
  return (
    <div className="border-border bg-sidebar flex flex-col gap-1 rounded-xl border px-4 py-3.5">
      <span className="text-muted-foreground text-micro">{label}</span>
      <span className="text-field">{children}</span>
    </div>
  )
}
