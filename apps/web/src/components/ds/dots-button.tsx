import type { ComponentProps } from "react"
import { cn } from "cn"

/**
 * The `···` that opens a row's menu, per the design's hover spec: a 22px square at
 * radius 5, with the dots drawn rather than typed — `···` is punctuation, not an icon.
 *
 * Deliberately not a Button: a Button carries padding and a label's worth of width, and
 * this has to sit inside a 31px sidebar row without changing its shape.
 */
export function DotsButton({ className, ...props }: ComponentProps<"button">) {
  return (
    <button
      type="button"
      className={cn(
        "grid size-5.5 place-items-center rounded-[5px] text-secondary-foreground",
        "transition-colors hover:bg-chip",
        className,
      )}
      {...props}
    >
      <svg width="13" height="13" viewBox="0 0 14 14" fill="currentColor" aria-hidden>
        <circle cx="3" cy="7" r="1.3" />
        <circle cx="7" cy="7" r="1.3" />
        <circle cx="11" cy="7" r="1.3" />
      </svg>
    </button>
  )
}
