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
    <div className="flex h-chrome items-center gap-2.5 border-b border-hair pr-4 pl-5.5 text-micro text-muted-foreground">
      {crumbs.map((crumb, index) => (
        <span key={crumb} className={index === crumbs.length - 1 ? "text-secondary-foreground" : ""}>
          {index > 0 && <span className="pr-2.5 text-muted-foreground">/</span>}
          {crumb}
        </span>
      ))}
      <span className="flex-1" />
      {actions}
    </div>
  )
}
