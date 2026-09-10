import { useMutation, useQueryClient } from "@tanstack/react-query"
import { getRouteApi, Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { AppShell } from "@/components/ds/app-shell"
import { Button } from "@/components/ds/button"
import { ChromeBar } from "@/components/ds/chrome-bar"
import { EmptyState } from "@/components/ds/empty-state"
import { SectionHeading } from "@/components/ds/section-heading"
import { listClient } from "@/lib/api"
import { groupByDay } from "@/lib/dated-queries"
import { useCommandPalette } from "@/lib/use-command-palette"
import { useDueLabel } from "@/lib/use-due-label"
import { DatedRow } from "./today"

const route = getRouteApi("/upcoming")

/**
 * Upcoming: the next two weeks, grouped by day.
 *
 * Nothing here has to be done today, which is the point of having it here.
 */
export function UpcomingScreen() {
  const { instance, member, lists, dated } = route.useLoaderData()
  const { t } = useTranslation()
  const palette = useCommandPalette()
  const queryClient = useQueryClient()
  const due = useDueLabel()

  const setDone = useMutation({
    mutationFn: ({ itemUid, done }: { itemUid: string; done: boolean }) =>
      listClient.setItemDone({ itemUid, done }),
    onSuccess: () => queryClient.invalidateQueries(),
  })

  const days = groupByDay(dated.items)

  return (
    <AppShell
      instanceName={instance.name}
      memberName={member.name}
      lists={lists.lists}
      onSearch={palette.open}
      onAddList={palette.openAddList}
    >
      <ChromeBar
        crumbs={[t("views.upcoming")]}
        actions={
          <Link to="/calendar">
            <Button tone="quiet" scale="toolbar">
              {t("views.calendar")}
            </Button>
          </Link>
        }
      />

      <div className="flex justify-center px-5.5 pt-14 pb-22">
        <div className="w-full max-w-content">
          <header className="mb-8.5 flex flex-col gap-3.5">
            <h1 className="text-display">{t("views.upcoming")}</h1>
            <div className="flex items-center gap-3 text-secondary text-secondary-foreground">
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
