import type { ReactElement, ReactNode } from "react"
import { cn } from "cn"

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

export interface MenuProps {
  /**
   * What opens it, usually the `···` control in the chrome bar.
   *
   * An element rather than any node: the trigger's own element is the one the menu is
   * anchored to and given its accessibility state, so there has to be exactly one.
   */
  trigger: ReactElement
  children: ReactNode
}

/**
 * A menu, per DESIGN.md §9: 248px, radius 9, 6px padding, 32px items at radius 6.
 *
 * Everything that is not worth a permanent control lives behind one of these. Groups
 * are split by a hairline inset 8px, and a destructive item takes `--overdue` — it is
 * never the menu's primary action and never sits with the rest.
 */
export function Menu({ trigger, children }: MenuProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger render={trigger} />
      <DropdownMenuContent
        align="end"
        className="w-62 rounded-menu border border-border p-1.5 shadow-[0_12px_32px_rgba(0,0,0,0.14)]"
      >
        {children}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export interface MenuItemProps {
  children: ReactNode
  /** The keyboard shortcut, shown right in muted text. */
  shortcut?: string
  /** Deleting is never the menu's primary action, and never reads like one. */
  destructive?: boolean
  disabled?: boolean
  onSelect: () => void
}

/** One thing a menu can do. */
export function MenuItem({
  children,
  shortcut,
  destructive,
  disabled,
  onSelect,
}: MenuItemProps) {
  return (
    <DropdownMenuItem
      disabled={disabled}
      onClick={onSelect}
      className={cn(
        "h-8 rounded-md px-2.5 text-chrome",
        destructive && "text-overdue focus:bg-destructive-bg focus:text-overdue",
      )}
    >
      {children}
      {shortcut && (
        <DropdownMenuShortcut className="text-micro tracking-normal">
          {shortcut}
        </DropdownMenuShortcut>
      )}
    </DropdownMenuItem>
  )
}

/** The hairline between one group of items and the next. */
export function MenuSeparator() {
  return <DropdownMenuSeparator className="-mx-0 my-1.5 bg-hair" />
}
