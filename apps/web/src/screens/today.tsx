import { useMutation, useQueryClient, useSuspenseQuery } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { AppShell } from "@/components/ds/app-shell"
import { ActivityControl } from "@/components/ds/activity-control"
import { ChromeBar } from "@/components/ds/chrome-bar"
import { DatedAddRow } from "@/components/ds/dated-add-row"
import { EmptyState } from "@/components/ds/empty-state"
import { ListRow, type ListRowLabels } from "@/components/ds/list-row"
import { SectionHeading } from "@/components/ds/section-heading"
import { listClient } from "@/lib/api"
import { todayQuery } from "@/lib/dated-queries"
import type { RenameItemVariables, SetDoneVariables } from "@/lib/item-mutations"
import type { DatedItem } from "@nooks/api"
import { dayFullHeading, isDueToday, isOverdue, today, toStored } from "@/lib/dates"
import { refreshLists } from "@/lib/refresh"
import { useCommandPalette } from "@/lib/use-command-palette"
import { useDueLabel } from "@/lib/use-due-label"
import { useLocale } from "@/lib/use-locale"
import { useSignedInData } from "@/lib/use-signed-in-data"

/**
 * Today: everything overdue, then everything due today.
 *
 * Undated Items never appear — they are still sitting on their own Lists, which is
 * where they belong.
 */
export function TodayScreen() {
  const { instanceName, member, lists } = useSignedInData()
  const { t } = useTranslation()
  const { dateLocale } = useLocale()
  const palette = useCommandPalette()
  const queryClient = useQueryClient()
  const due = useDueLabel()

  const from = today()
  const dated = useSuspenseQuery(todayQuery(from)).data

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

  const rowLabels: ListRowLabels = {
    name: t("note.itemName"),
    open: t("list.openItem"),
    quantity: t("note.quantity"),
    due: t("note.addDate"),
  }

  const overdue = dated.items.filter((entry) => isOverdue(entry.item?.dueOn ?? "", from))
  const dueToday = dated.items.filter((entry) => isDueToday(entry.item?.dueOn ?? "", from))

  return (
    <AppShell
      instanceName={instanceName}
      memberName={member?.name ?? ""}
      lists={lists}
      onSearch={palette.open}
      onAddList={palette.openAddList}
    >
      <ChromeBar crumbs={[t("views.today")]} actions={<ActivityControl />} />

      <div className="flex justify-center px-5.5 pt-14 pb-22">
        <div className="w-full max-w-content">
          <header className="mb-8.5 flex flex-col gap-3.5">
            <h1 className="text-display">{t("views.today")}</h1>
            <div className="flex items-center gap-3 text-meta text-secondary-foreground">
              <span>{dayFullHeading(from, dateLocale)}</span>
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
                      onRename={(label) =>
                        rename.mutate({ itemUid: entry.item?.uid ?? "", label })
                      }
                      labels={rowLabels}
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
                      onRename={(label) =>
                        rename.mutate({ itemUid: entry.item?.uid ?? "", label })
                      }
                      labels={rowLabels}
                    />
                  ))}
                </section>
              )}
            </div>
          )}

          <DatedAddRow lists={lists} defaultDue={toStored(from)} divided={dated.items.length > 0} />
        </div>
      </div>
    </AppShell>
  )
}

export interface DatedRowProps {
  entry: DatedItem
  dueLabel?: string
  overdue?: boolean
  onToggle: (done: boolean) => void
  onRename: (label: string) => void
  labels: ListRowLabels
}

/**
 * A row in a dated view. It names the List instead of who added it: away from its own
 * List, where an Item lives matters more than who put it there.
 */
export function DatedRow({ entry, dueLabel, overdue, onToggle, onRename, labels }: DatedRowProps) {
  const navigate = useNavigate()
  const item = entry.item
  if (!item) {
    return null
  }

  return (
    <ListRow
      label={item.label}
      quantity={item.quantity}
      addedByName={entry.listName}
      dueLabel={dueLabel}
      overdue={overdue}
      onToggle={onToggle}
      onRename={onRename}
      onOpen={() =>
        void navigate({
          to: "/lists/$listUid/items/$itemUid",
          params: { listUid: entry.listUid, itemUid: item.uid },
        })
      }
      labels={labels}
    />
  )
}
