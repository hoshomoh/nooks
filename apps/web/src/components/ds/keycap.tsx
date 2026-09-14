import type { ReactNode } from "react"
import { cn } from "cn"

export interface KeycapProps {
  /** The key as it is printed on the keyboard: `Esc`, `⌘K`, `↵`. */
  children: ReactNode
  className?: string
}

/**
 * A key, drawn as a key — DESIGN.md §3's mono keycap.
 *
 * Not translated, and deliberately: `Esc` is what is written on the key under the
 * Member's finger, in every language. Translating it would name a key that is not
 * there.
 *
 * Here rather than written out at each use because it was already written out twice,
 * and a shortcut hint that is drawn three slightly different ways is a shortcut hint
 * nobody trusts.
 */
export function Keycap({ children, className }: KeycapProps) {
  return (
    <span
      className={cn(
        "rounded-sm border border-border px-1.5 py-px font-mono text-keycap text-muted-foreground",
        className,
      )}
    >
      {children}
    </span>
  )
}
