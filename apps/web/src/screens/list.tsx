import { useMutation, useQueryClient } from "@tanstack/react-query"
import { getRouteApi } from "@tanstack/react-router"

import { AddRow } from "@/components/ds/add-row"
import { AppShell } from "@/components/ds/app-shell"
import { ChromeBar } from "@/components/ds/chrome-bar"
import { EmptyState } from "@/components/ds/empty-state"
import { ListRow } from "@/components/ds/list-row"
import { listClient } from "@/lib/api"
import { dueLabel, isOverdue } from "@/lib/dates"
import { useCommandPalette } from "@/lib/use-command-palette"

const route = getRouteApi("/lists/$listUid")

export function ListScreen() {
  const { instance, member, lists, list } = route.useLoaderData()
  const { listUid } = route.useParams()
  const queryClient = useQueryClient()
  const palette = useCommandPalette()

  // Today is read once per render rather than per row, so every date in one paint is
  // measured against the same moment.
  const today = new Date()

  const refresh = () => queryClient.invalidateQueries()

  const addItem = useMutation({
    mutationFn: (label: string) => listClient.createItem({ listUid, label }),
    onSuccess: refresh,
  })

  const setDone = useMutation({
    mutationFn: ({ itemUid, done }: { itemUid: string; done: boolean }) =>
      listClient.setItemDone({ itemUid, done }),
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
      <ChromeBar crumbs={[list.list?.isOwner ? "My lists" : "Shared", list.list?.name ?? ""]} />

      <div className="flex justify-center px-5.5 pt-14 pb-22">
        <div className="w-full max-w-content">
          <header className="mb-8.5 flex flex-col gap-3.5">
            <h1 className="text-display">{list.list?.name}</h1>
            <div className="flex items-center gap-3 text-secondary text-secondary-foreground">
              <span>{sharingLine(list.list?.sharing ?? 0, list.list?.canEdit ?? false)}</span>
              <span className="h-3 w-px bg-border" />
              <span>{open.length === 0 ? "Nothing on it yet" : `${open.length} open`}</span>
            </div>
          </header>

          {list.items.length === 0 ? (
            <EmptyState
              title="This list is empty"
              body="Type below and press enter. Items can carry a quantity, a date and a note later — none of that is needed to start."
            />
          ) : (
            <div className="flex flex-col">
              {open.map((item) => (
                <ListRow
                  key={item.uid}
                  label={item.label}
                  quantity={item.quantity}
                  addedByName={item.addedByName}
                  dueLabel={dueLabel(item.dueOn, today)}
                  overdue={isOverdue(item.dueOn, today)}
                  onToggle={(next) => setDone.mutate({ itemUid: item.uid, done: next })}
                />
              ))}
            </div>
          )}

          {canEdit && (
            <AddRow
              placeholder="Add an item"
              onAdd={(label) => addItem.mutate(label)}
              disabled={addItem.isPending}
            />
          )}

          {done.length > 0 && (
            <div className="mt-8.5 flex flex-col gap-1.5 border-t border-hair pt-4.5">
              <span className="text-secondary text-secondary-foreground">
                {done.length} done
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

/** The line under the title saying who can reach this List. */
function sharingLine(sharing: number, canEdit: boolean): string {
  // 1 is SHARING_PRIVATE, 2 is SHARING_INSTANCE.
  if (sharing === 1) {
    return "Only you"
  }
  return canEdit ? "Shared with the household · anyone can edit" : "Shared with the household · read-only"
}
