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
        style.box,
        className,
      )}
      {...props}
    >
      <Icon name={name} size={style.icon} />
    </button>
  )
}
