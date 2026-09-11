import { cn } from "cn"

import { Icon } from "./icon"

export interface TickBoxProps {
  picked: boolean
  className?: string
}

/**
 * The small square that says a row in a chooser is picked.
 *
 * Decorative: it is drawn inside a button that already says what it is and carries the
 * pressed state, so it never announces anything of its own.
 */
export function TickBox({ picked, className }: TickBoxProps) {
  return (
    <span
      className={cn(
        "grid size-4 shrink-0 place-items-center rounded-sm border-[1.5px] text-background",
        picked ? "border-shared bg-shared" : "border-control",
        className,
      )}
    >
      {picked && <Icon name="check" size="small" className="size-3" />}
    </span>
  )
}
