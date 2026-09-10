import { useMutation, useQueryClient } from "@tanstack/react-query"
import { getRouteApi } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import { format } from "date-fns"

import { AppShell } from "@/components/ds/app-shell"
import { ChromeBar } from "@/components/ds/chrome-bar"
import { EmptyState } from "@/components/ds/empty-state"
import { ListRow } from "@/components/ds/list-row"
import { SectionHeading } from "@/components/ds/section-heading"
import { listClient } from "@/lib/api"
import { isDueToday, isOverdue, today } from "@/lib/dates"
import { useCommandPalette } from "@/lib/use-command-palette"
import { useDueLabel } from "@/lib/use-due-label"
import { useLocale } from "@/lib/use-locale"

const route = getRouteApi("/today")

/**
 * Today: everything overdue, then everything due today.
 *
 * Undated Items never appear — they are still sitting on their own Lists, which is
 * where they belong.
 */
export function TodayScreen() {
  const { instance, member, lists, dated } = route.useLoaderData()
  const { t } = useTranslation()
  const { dateLocale } = useLocale()
  const palette = useCommandPalette()
  const queryClient = useQueryClient()
  const due = useDueLabel()

  const from = today()

  const setDone = useMutation({
    mutationFn: ({ itemUid, done }: { itemUid: string; done: boolean }) =>
      listClient.setItemDone({ itemUid, done }),
    onSuccess: () => queryClient.invalidateQueries(),
  })

  const overdue = dated.items.filter((entry) => isOverdue(entry.item?.dueOn ?? "", from))
  const dueToday = dated.items.filter((entry) => isDueToday(entry.item?.dueOn ?? "", from))

  return (
    <AppShell
      instanceName={instance.name}
      memberName={member.name}
      lists={lists.lists}
      onSearch={palette.open}
      onAddList={palette.openAddList}
    >
      <ChromeBar crumbs={[t("views.today")]} />

      <div className="flex justify-center px-5.5 pt-14 pb-22">
        <div className="w-full max-w-content">
          <header className="mb-8.5 flex flex-col gap-3.5">
            <h1 className="text-display">{t("views.today")}</h1>
            <div className="flex items-center gap-3 text-secondary text-secondary-foreground">
              <span>{format(from, "EEEE, d MMMM", { locale: dateLocale })}</span>
              <span className="h-3 w-px bg-border" />
              <span>
                {dated.items.length === 0
                  ? t("views.nothingDue")
                  : [
                      t("views.dueCount", { count: dueToday.length }),
                      overdue.length > 0 && t("views.overdueCount", { count: overdue.length }),
                    ]
                      .filter(Boolean)
                      .join(" · ")}
              </span>
            </div>
          </header>

          {dated.items.length === 0 ? (
            <EmptyState title={t("views.todayEmptyTitle")} body={t("views.todayEmptyBody")} />
          ) : (
            <div className="flex flex-col gap-6.5">
              {overdue.length > 0 && (
                <section>
                  <SectionHeading
                    label={t("views.overdue")}
                    note={t("views.overdueSince", { date: due.label(overdue[0].item?.dueOn ?? "") })}
                    overdue
                  />
                  {overdue.map((entry) => (
                    <DatedRow
                      key={entry.item?.uid}
                      entry={entry}
                      dueLabel={due.label(entry.item?.dueOn ?? "")}
                      overdue
                      onToggle={(done) =>
                        setDone.mutate({ itemUid: entry.item?.uid ?? "", done })
                      }
                    />
                  ))}
                </section>
              )}

              {dueToday.length > 0 && (
                <section>
                  <SectionHeading label={t("views.dueToday")} />
                  {dueToday.map((entry) => (
                    <DatedRow
                      key={entry.item?.uid}
                      entry={entry}
                      onToggle={(done) =>
                        setDone.mutate({ itemUid: entry.item?.uid ?? "", done })
                      }
                    />
                  ))}
                </section>
              )}
            </div>
          )}
        </div>
      </div>
    </AppShell>
  )
}

export type DatedRowProps = {
  entry: { item?: { uid: string; label: string; quantity: string }; listName: string }
  dueLabel?: string
  overdue?: boolean
  onToggle: (done: boolean) => void
}

/**
 * A row in a dated view. It names the List instead of who added it: away from its own
 * List, where an Item lives matters more than who put it there.
 */
export function DatedRow({ entry, dueLabel, overdue, onToggle }: DatedRowProps) {
  return (
    <ListRow
      label={entry.item?.label ?? ""}
      quantity={entry.item?.quantity}
      addedByName={entry.listName}
      dueLabel={dueLabel}
      overdue={overdue}
      onToggle={onToggle}
    />
  )
}
