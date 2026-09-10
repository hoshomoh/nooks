import { useMutation, useQueryClient } from "@tanstack/react-query"
import { getRouteApi } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { AddRow } from "@/components/ds/add-row"
import { AppShell } from "@/components/ds/app-shell"
import { ChromeBar } from "@/components/ds/chrome-bar"
import { EmptyState } from "@/components/ds/empty-state"
import { ListRow } from "@/components/ds/list-row"
import { listClient } from "@/lib/api"
import { isOverdue, today } from "@/lib/dates"
import { useDueLabel } from "@/lib/use-due-label"
import { useCommandPalette } from "@/lib/use-command-palette"

const route = getRouteApi("/lists/$listUid")

/** What the tick mutation is told. */
type SetDoneVariables = {
  itemUid: string
  done: boolean
}

export function ListScreen() {
  const { instance, member, lists, list } = route.useLoaderData()
  const { listUid } = route.useParams()
  const queryClient = useQueryClient()
  const palette = useCommandPalette()
  const { t } = useTranslation()

  // The day is read once per render rather than per row, so every date in one paint is
  // measured against the same moment.
  const from = today()
  const due = useDueLabel()

  const refresh = () => queryClient.invalidateQueries()

  const addItem = useMutation({
    mutationFn: (label: string) => listClient.createItem({ listUid, label }),
    onSuccess: refresh,
  })

  const setDone = useMutation({
    mutationFn: ({ itemUid, done }: SetDoneVariables) => listClient.setItemDone({ itemUid, done }),
    onSuccess: refresh,
  })

  const open = list.items.filter((item) => !item.done)
  const done = list.items.filter((item) => item.done)
  const canEdit = list.list?.isOwner || list.list?.canEdit

  return (
    <AppShell
      instanceName={instance.name}
      memberName={member.name}
      lists={lists.lists}
      activeListUid={listUid}
      onSearch={palette.open}
      onAddList={palette.openAddList}
    >
      <ChromeBar crumbs={[t(list.list?.isOwner ? "list.myListsCrumb" : "list.sharedCrumb"), list.list?.name ?? ""]} />

      <div className="flex justify-center px-5.5 pt-14 pb-22">
        <div className="w-full max-w-content">
          <header className="mb-8.5 flex flex-col gap-3.5">
            <h1 className="text-display">{list.list?.name}</h1>
            <div className="flex items-center gap-3 text-secondary text-secondary-foreground">
              <span>{t(sharingKey(list.list?.sharing ?? 0, list.list?.canEdit ?? false))}</span>
              <span className="h-3 w-px bg-border" />
              <span>{open.length === 0 ? t("list.nothingYet") : t("list.openCount", { count: open.length })}</span>
            </div>
          </header>

          {list.items.length === 0 ? (
            <EmptyState title={t("list.emptyTitle")} body={t("list.emptyBody")} />
          ) : (
            <div className="flex flex-col">
              {open.map((item) => (
                <ListRow
                  key={item.uid}
                  label={item.label}
                  quantity={item.quantity}
                  addedByName={item.addedByName}
                  dueLabel={due.label(item.dueOn)}
                  overdue={isOverdue(item.dueOn, from)}
                  onToggle={(next) => setDone.mutate({ itemUid: item.uid, done: next })}
                />
              ))}
            </div>
          )}

          {canEdit && (
            <AddRow
              placeholder={t("list.addItem")}
              onAdd={(label) => addItem.mutate(label)}
              disabled={addItem.isPending}
            />
          )}

          {done.length > 0 && (
            <div className="mt-8.5 flex flex-col gap-1.5 border-t border-hair pt-4.5">
              <span className="text-secondary text-secondary-foreground">
                {t("list.doneCount", { count: done.length })}
              </span>
              {done.map((item) => (
                <ListRow
                  key={item.uid}
                  label={item.label}
                  quantity={item.quantity}
                  addedByName={item.doneByName || item.addedByName}
                  done
                  onToggle={(next) => setDone.mutate({ itemUid: item.uid, done: next })}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </AppShell>
  )
}

/** Which line under the title says who can reach this List. */
function sharingKey(sharing: number, canEdit: boolean): string {
  // 1 is SHARING_PRIVATE, 2 is SHARING_INSTANCE.
  if (sharing === 1) {
    return "list.onlyYou"
  }
  return canEdit ? "list.sharedCanEdit" : "list.sharedReadOnly"
}
