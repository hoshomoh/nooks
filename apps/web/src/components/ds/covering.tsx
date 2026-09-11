import type { ReactNode } from "react"
import { cn } from "cn"

/**
 * Making a whole row one target, without its contents eating the click.
 *
 * A row that is a single target is drawn as a container with an element stretched over
 * it, and the row's own contents on top. The trap is that a positioned child paints
 * *above* that stretched element even with no z-index of its own — so the label, the
 * count or the description quietly swallows every click that lands on it, and the row
 * appears to work only in the gaps between its own text.
 *
 * That is one bug, and it has been found three times in three rows. So the three parts
 * are named here rather than written out again: what stretches, what is inert, and what
 * is raised above both.
 */

/** COVERING stretches an element over its container, behind the container's contents. */
export const COVERING = "absolute inset-0 rounded-[inherit]"

/**
 * INERT marks content that only shows something.
 *
 * Everything in a covered row is inert unless it has a job of its own. Text, counts,
 * dots, badges: none of them answer a click, and all of them are in the way of one.
 */
export const INERT = "pointer-events-none relative"

/** RAISED marks content that answers a click of its own — a checkbox, a `···`, a field. */
export const RAISED = "relative z-10"

export interface CoveringButtonProps {
  /** What a screen reader calls the row. */
  label: string
  onClick: () => void
}

/** The stretched button, for a row whose action is not a destination. */
export function CoveringButton({ label, onClick }: CoveringButtonProps) {
  return <button type="button" onClick={onClick} aria-label={label} className={COVERING} />
}

export interface CoveredRowProps {
  className?: string
  children: ReactNode
}

/** The container a covering element and its contents sit in. */
export function CoveredRow({ className, children }: CoveredRowProps) {
  return <div className={cn("group relative", className)}>{children}</div>
}
