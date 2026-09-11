import type { ReactNode } from "react"
import { Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import { cn } from "cn"

import { ChromeBar } from "./chrome-bar"
import { Sidebar } from "./sidebar"
import { useCommandPalette } from "@/lib/use-command-palette"
import { useSignedInData } from "@/lib/use-signed-in-data"

/** Where a settings page lives. Only routes that exist are listed. */
export type SettingsRoute = "/settings/members" | "/settings/groups" | "/settings/public"

/** One entry in the settings column. */
interface SettingsSection {
  to: SettingsRoute
  labelKey: string
}

/**
 * The settings pages, in the order the design lists them.
 *
 * General, Access tokens, Public list and About arrive with the milestones that build
 * them. A nav entry that goes nowhere is worse than one that is not there yet.
 */
const SECTIONS: SettingsSection[] = [
  { to: "/settings/members", labelKey: "settings.members" },
  { to: "/settings/groups", labelKey: "settings.groups" },
  { to: "/settings/public", labelKey: "settings.publicList" },
]

export interface SettingsShellProps {
  /** The page being read, so its entry is marked. */
  active: SettingsRoute
  /** What the chrome bar says after "Settings". */
  crumb: string
  /** Counts beside the entries, keyed by route. */
  counts?: Partial<Record<SettingsRoute, number>>
  children: ReactNode
}

/**
 * The settings layout: the app's own sidebar, then a settings column, then a 720px
 * content column.
 *
 * The sidebar stays because settings is somewhere a Member goes for a minute in the
 * middle of using the app — taking their Lists away would make going back a journey
 * rather than a click. The content column is wider than a List's 660px because these
 * pages hold tables rather than sentences, per DESIGN.md §5.
 */
export function SettingsShell({ active, crumb, counts, children }: SettingsShellProps) {
  const { t } = useTranslation()
  const { instanceName, member, lists } = useSignedInData()
  const palette = useCommandPalette()

  return (
    <div className="flex min-h-dvh bg-background">
      <Sidebar
        instanceName={instanceName}
        memberName={member?.name ?? ""}
        lists={lists}
        onSearch={palette.open}
        onAddList={palette.openAddList}
      />

      <nav className="flex w-56 flex-col gap-4 border-r border-border bg-sidebar px-2.5 pt-3 pb-4">
        <span className="px-2 text-label text-muted-foreground uppercase">
          {t("settings.title")}
        </span>

        <div className="flex flex-col gap-px">
          {SECTIONS.map((section) => (
            <Link
              key={section.to}
              to={section.to}
              className={cn(
                "flex h-8 items-center rounded-md px-2 text-meta text-secondary-foreground",
                "transition-colors hover:bg-secondary",
                section.to === active && "bg-secondary font-medium text-foreground",
              )}
            >
              {t(section.labelKey)}
              {counts?.[section.to] !== undefined && (
                <span className="ml-auto text-[11.5px] text-muted-foreground">
                  {counts[section.to]}
                </span>
              )}
            </Link>
          ))}
        </div>

        <Link
          to="/"
          className="mt-auto flex h-8 items-center rounded-md px-2 text-meta text-muted-foreground transition-colors hover:bg-secondary"
        >
          {t("settings.backToLists")}
        </Link>
      </nav>

      <main className="flex min-w-0 flex-1 flex-col">
        <ChromeBar crumbs={[t("settings.title"), crumb]} />
        <div className="flex justify-center px-5.5 pt-5 pb-16">
          <div className="flex w-full max-w-settings flex-col gap-3.5">{children}</div>
        </div>
      </main>
    </div>
  )
}
