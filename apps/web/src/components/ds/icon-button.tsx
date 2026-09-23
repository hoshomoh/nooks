import type { ComponentProps } from "react"
import { cn } from "cn"

import { Icon, type IconName, type IconSize } from "./icon"

/** How much room the button takes. */
export type IconButtonScale = "compact" | "regular"

interface ScaleStyle {
  box: string
  icon: IconSize
}

const SCALES: Record<IconButtonScale, ScaleStyle> = {
  // Fits a 31px sidebar row without changing its shape.
  compact: { box: "size-5.5 rounded-[5px]", icon: "small" },
  regular: { box: "size-6.5 rounded-md", icon: "medium" },
}

/*
The hit area, which is taller than the button is drawn.

DESIGN.md §4: a control in a row is as easy to hit as the row is tall. The square drawn
here is 22px or 26px, both under the 24px WCAG asks at AA, and growing the square would
push what sits beside it out of line and change every screen it appears on.

So the target grows and the button does not, the way the tick box already does it. Only
downwards and upwards: these sit at the end of a row beside other things, and a target
that reached sideways would take its neighbour's taps, which is an error nobody can
anticipate. Taller than the row it is in would do the same to the row above.
*/
const HIT_AREA =
  "relative after:absolute after:inset-x-0 after:top-1/2 after:h-control-compact " +
  "after:-translate-y-1/2 after:content-['']"

export interface IconButtonProps extends Omit<ComponentProps<"button">, "children"> {
  name: IconName
  /** What pressing it does, for anyone who cannot see the icon. Required. */
  label: string
  scale?: IconButtonScale
}

/**
 * A button that is only an icon.
 *
 * It is a square with a hover ground, so it reads as something that can be pressed
 * rather than as a character that happens to sit there. Every icon-only control in the
 * app is one of these — a `‹` in a span is not a button, however it is styled.
 */
export function IconButton({
  name,
  label,
  scale = "regular",
  className,
  ...props
}: IconButtonProps) {
  const style = SCALES[scale]

  return (
    <button
      type="button"
      aria-label={label}
      className={cn(
        "grid place-items-center text-secondary-foreground transition-colors",
        "hover:bg-secondary hover:text-foreground",
        HIT_AREA,
        style.box,
        className,
      )}
      {...props}
    >
      <Icon name={name} size={style.icon} />
    </button>
  )
}
