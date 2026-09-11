import type { ReactNode } from "react"
import { Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import { cn } from "cn"
import { Role } from "@nooks/api"

import { ChromeBar } from "./chrome-bar"
import { Sidebar } from "./sidebar"
import { SETTINGS_SECTIONS, type SettingsRoute } from "./settings-sections"
import { useCommandPalette } from "@/lib/use-command-palette"
import { useSignedInData } from "@/lib/use-signed-in-data"

/** The numbers beside the nav entries, read once by the shell. */
export interface SettingsCounts {
  tokens?: number
  members?: number
  groups?: number
}

export interface SettingsShellProps {
  /** The page being read, so its entry is marked. */
  active: SettingsRoute
  /** What the chrome bar says after "Settings". */
  crumb: string
  /**
   * The counts beside the entries.
   *
   * Passed by every settings page rather than by some of them: a column whose entries
   * grow a number depending on which page you are standing on is a column that changes
   * shape as you walk through it.
   */
  counts: SettingsCounts
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
  const isAdmin = member?.role === Role.ADMIN
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
          {SETTINGS_SECTIONS.filter((section) => !section.adminOnly || isAdmin).map((section) => (
            <Link
              key={section.to}
              to={section.to}
              className={cn(
                "flex h-8 items-center rounded-md px-2 text-meta text-secondary-foreground",
                "transition-colors hover:bg-secondary",
                section.to === active && "bg-secondary font-medium text-foreground",
              )}
            >
              <span>{t(section.labelKey)}</span>
              {section.counts && counts[section.counts] !== undefined && (
                <span className="ml-auto text-micro text-muted-foreground">
                  {counts[section.counts]}
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
