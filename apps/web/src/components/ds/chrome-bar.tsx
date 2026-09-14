import type { ReactNode } from "react"

/**
 * The chrome bar, per DESIGN.md §8: 44px, breadcrumb left, view-level actions right.
 *
 * It carries location and view-level actions only — never the page's primary action.
 */
export type ChromeBarProps = {
  /** The trail of where the Member is, e.g. ["Shared", "Groceries"]. */
  crumbs: string[]
  /** View-level actions only — never the page's primary action. */
  actions?: ReactNode
}

export function ChromeBar({ crumbs, actions }: ChromeBarProps) {
  return (
    <div className="relative flex h-chrome items-center gap-2.5 border-b border-hair pr-4 pl-5.5 text-micro text-muted-foreground">
      {/* The crumb is what gives way. It takes the room that is left and truncates in
          it; the controls beside it keep their size, because a button that shrinks
          wraps its own label onto two lines inside a 44px bar. */}
      <span className="min-w-0 flex-1 truncate">
        {crumbs.map((crumb, index) => (
          <span
            key={crumb}
            className={index === crumbs.length - 1 ? "text-secondary-foreground" : ""}
          >
            {index > 0 && <span className="pr-2.5 text-muted-foreground">/</span>}
            {crumb}
          </span>
        ))}
      </span>
      {actions && <span className="flex shrink-0 items-center gap-2.5">{actions}</span>}
    </div>
  )
}
