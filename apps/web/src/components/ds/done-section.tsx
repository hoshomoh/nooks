import { useState } from "react"
import type { ReactNode } from "react"

export interface DoneSectionProps {
  /** How the row reads, e.g. "3 done today". */
  label: string
  /** The ticked rows, drawn only once the Member asks for them. */
  children: ReactNode
}

/**
 * The completed Items, per DESIGN.md §6.
 *
 * Collapsed, at the foot of the List, behind one line that says how many were ticked
 * today. What is done is not what a Member came to the List for, and a List that
 * doubles in length as the week goes on is a List nobody scrolls to the bottom of — but
 * ticking something should still visibly put it somewhere rather than delete it.
 */
export function DoneSection({ label, children }: DoneSectionProps) {
  const [open, setOpen] = useState(false)

  return (
    <div className="mt-8.5 flex flex-col gap-1.5 border-t border-hair pt-4.5">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        aria-expanded={open}
        className="flex items-center gap-2 self-start text-meta text-secondary-foreground"
      >
        <span className="text-[10px] text-muted-foreground">{open ? "▾" : "▸"}</span>
        <span>{label}</span>
      </button>

      {open && children}
    </div>
  )
}
