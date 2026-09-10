import { useMutation, useQueryClient, useSuspenseQuery } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { AppShell } from "@/components/ds/app-shell"
import { Button } from "@/components/ds/button"
import { ActivityControl } from "@/components/ds/activity-control"
import { ChromeBar } from "@/components/ds/chrome-bar"
import { EmptyState } from "@/components/ds/empty-state"
import { SectionHeading } from "@/components/ds/section-heading"
import { listClient } from "@/lib/api"
import { groupByDay, upcomingQuery } from "@/lib/dated-queries"
import type { RenameItemVariables, SetDoneVariables } from "@/lib/item-mutations"
import { today } from "@/lib/dates"
import { refreshLists } from "@/lib/refresh"
import { useCommandPalette } from "@/lib/use-command-palette"
import { useDueLabel } from "@/lib/use-due-label"
import { useSignedInData } from "@/lib/use-signed-in-data"
import type { ListRowLabels } from "@/components/ds/list-row"
import { DatedRow } from "./today"

/**
 * Upcoming: the next two weeks, grouped by day.
 *
 * Nothing here has to be done today, which is the point of having it here.
 */
export function UpcomingScreen() {
  const { instanceName, member, lists } = useSignedInData()
  const { t } = useTranslation()
  const palette = useCommandPalette()
  const queryClient = useQueryClient()
  const due = useDueLabel()
  const dated = useSuspenseQuery(upcomingQuery(today())).data

  const setDone = useMutation({
    mutationFn: ({ itemUid, done }: SetDoneVariables) =>
      listClient.setItemDone({ itemUid, done }),
    onSuccess: () => refreshLists(queryClient),
  })

  const rename = useMutation({
    mutationFn: ({ itemUid, label }: RenameItemVariables) =>
      listClient.updateItem({ itemUid, label }),
    onSuccess: () => refreshLists(queryClient),
  })

  const rowLabels: ListRowLabels = { name: t("note.itemName"), open: t("list.openItem") }

  const days = groupByDay(dated.items)

  return (
    <AppShell
      instanceName={instanceName}
      memberName={member?.name ?? ""}
      lists={lists}
      onSearch={palette.open}
      onAddList={palette.openAddList}
    >
      <ChromeBar
        crumbs={[t("views.upcoming")]}
        actions={
          <>
            <Link to="/calendar">
              <Button tone="quiet" scale="toolbar">
                {t("views.calendar")}
              </Button>
            </Link>
            <ActivityControl />
          </>
        }
      />

      <div className="flex justify-center px-5.5 pt-14 pb-22">
        <div className="w-full max-w-content">
          <header className="mb-8.5 flex flex-col gap-3.5">
            <h1 className="text-display">{t("views.upcoming")}</h1>
            <div className="flex items-center gap-3 text-meta text-secondary-foreground">
              <span>{t("views.upcomingSubtitle")}</span>
              <span className="h-3 w-px bg-border" />
              <span>{t("views.itemCount", { count: dated.items.length })}</span>
            </div>
          </header>

          {days.length === 0 ? (
            <EmptyState title={t("views.upcomingEmptyTitle")} body={t("views.upcomingEmptyBody")} />
          ) : (
            <div className="flex flex-col gap-6.5">
              {days.map((group) => (
                <section key={group.day}>
                  <SectionHeading label={due.heading(group.day)} note={due.label(group.day)} />
                  {group.items.map((entry) => (
                    <DatedRow
                      key={entry.item?.uid}
                      entry={entry}
                      onToggle={(done) =>
                        setDone.mutate({ itemUid: entry.item?.uid ?? "", done })
                      }
                      onRename={(label) =>
                        rename.mutate({ itemUid: entry.item?.uid ?? "", label })
                      }
                      labels={rowLabels}
                    />
                  ))}
                </section>
              ))}
            </div>
          )}
        </div>
      </div>
    </AppShell>
  )
}
