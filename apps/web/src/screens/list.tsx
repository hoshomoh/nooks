import { useCallback, useMemo, useState } from "react"
import { useMutation, useQueryClient, useSuspenseQuery } from "@tanstack/react-query"
import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { ActivityControl } from "@/components/ds/activity-control"
import { Presence } from "@/components/ds/presence"
import { AddRow } from "@/components/ds/add-row"
import { DoneSection } from "@/components/ds/done-section"
import { ItemMenu } from "@/components/ds/item-menu"
import { ListActions } from "@/components/ds/list-actions"
import { AppShell } from "@/components/ds/app-shell"
import { ChromeBar } from "@/components/ds/chrome-bar"
import { EmptyState } from "@/components/ds/empty-state"
import { ListRow, type ListRowLabels } from "@/components/ds/list-row"
import { PrintSheet } from "@/components/ds/print-sheet"
import { NoteSheet } from "@/components/ds/note-sheet"
import { listClient } from "@/lib/api"
import { listQuery } from "@/lib/list-queries"
import { refreshLists } from "@/lib/refresh"
import { useAddItem, useSetDone } from "@/lib/use-item-changes"
import { isProvisional } from "@/lib/queued-changes"
import { useQueuedChanges } from "@/lib/use-queued-changes"
import type { Translate } from "@/lib/translate"
import { debounce } from "@/lib/debounce"
import type { RenameItemVariables, SaveNoteVariables } from "@/lib/item-mutations"
import { dayFullHeading, happenedToday, isOverdue, justHappened, today } from "@/lib/dates"
import { useDueLabel } from "@/lib/use-due-label"
import { useLocale } from "@/lib/use-locale"
import { useCommandPalette } from "@/lib/use-command-palette"
import { useLive } from "@/lib/use-live"
import { useSignedInData } from "@/lib/use-signed-in-data"
import { Sharing, type Item } from "@nooks/api"

const route = getRouteApi("/lists/$listUid")

/** What the quantity-and-date edit is told. Either field may be left alone. */
interface EditItemVariables {
  itemUid: string
  quantity?: string
  dueOn?: string
}

/** How long the typing has to settle before a Note is saved. */
const AUTOSAVE_DELAY_MS = 800

export function ListScreen() {
  const { listUid } = route.useParams()
  const { instanceName, member, lists } = useSignedInData()
  const list = useSuspenseQuery(listQuery(listUid)).data
  const live = useLive()
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const palette = useCommandPalette()
  const { t } = useTranslation()

  // The day is read once per render rather than per row, so every date in one paint is
  // measured against the same moment.
  const from = today()
  const due = useDueLabel()
  const { dateLocale } = useLocale()
  const printedOn = dayFullHeading(from, dateLocale)

  // A tick somebody else just made lands with a highlight and then settles. Read from
  // the Item itself rather than remembered between renders: the Item already says when
  // it was ticked and by whom, so there is nothing to keep in sync.
  const settled = new Date()
  const justTickedByAnother = (item: Item) =>
    item.done && item.doneByUid !== member?.uid && justHappened(item.doneAt, settled)

  // Which Item's sheet is open. The List keeps its place behind it, so this is the
  // screen's own state rather than a route.
  const [openItemUid, setOpenItemUid] = useState<string | null>(null)


  const refresh = () => refreshLists(queryClient)

  const addItem = useAddItem()

  const saveNote = useMutation({
    mutationFn: ({ itemUid, note }: SaveNoteVariables) => listClient.updateItem({ itemUid, note }),
    onSuccess: refresh,
  })

  // Autosave waits for the typing to settle rather than firing per keystroke.
  // `mutate` is referentially stable, so the debounce is built once.
  const autosave = useMemo(() => debounce(saveNote.mutate, AUTOSAVE_DELAY_MS), [saveNote.mutate])

  // Closing flushes, so the last sentence is never lost to a timer that never ran.
  const closeSheet = useCallback(() => {
    autosave.flush()
    setOpenItemUid(null)
  }, [autosave])

  // Stable, so the editor is built once rather than on every render.
  const onNoteChange = useCallback(
    (note: string) => {
      if (openItemUid) {
        autosave.call({ itemUid: openItemUid, note })
      }
    },
    [autosave, openItemUid],
  )

  const setDone = useSetDone()
  const queued = useQueuedChanges()

  const rename = useMutation({
    mutationFn: ({ itemUid, label }: RenameItemVariables) =>
      listClient.updateItem({ itemUid, label }),
    onSuccess: refresh,
  })

  // Quantity and due date are two fields of the same edit, so they are one mutation.
  const editItem = useMutation({
    mutationFn: ({ itemUid, quantity, dueOn }: EditItemVariables) =>
      listClient.updateItem({ itemUid, quantity, dueOn }),
    onSuccess: refresh,
  })

  const removeItem = useMutation({
    mutationFn: (itemUid: string) => listClient.deleteItem({ itemUid }),
    onSuccess: refresh,
  })






  const rowLabels: ListRowLabels = {
    name: t("note.itemName"),
    open: t("list.openItem"),
    quantity: t("note.quantity"),
    due: t("note.addDate"),
    notSynced: t("list.notSynced"),
  }

  const openItem = list.items.find((item) => item.uid === openItemUid)
  const open = list.items.filter((item) => !item.done)
  const done = list.items.filter((item) => item.done)
  const canEdit = list.list?.isOwner || list.list?.canEdit

  return (
    <AppShell
      instanceName={instanceName}
      memberName={member?.name ?? ""}
      lists={lists}
      activeListUid={listUid}
      onSearch={palette.open}
      onAddList={palette.openAddList}
    >
      {/* Rendered into the page and hidden on screen, so ⌘P and the Print entry take
          the List rather than the browser's idea of the app. */}
      {list.list && (
        <PrintSheet
          instanceName={instanceName}
          listName={list.list.name}
          printedOn={printedOn}
          items={list.items}
        />
      )}

      <ChromeBar
        crumbs={[t(list.list?.isOwner ? "list.myListsCrumb" : "list.sharedCrumb"), list.list?.name ?? ""]}
        actions={
          <>
            <ActivityControl />
            {list.list && (
              <ListActions
                list={list.list}
                instanceName={instanceName}
                items={list.items}
                withControls
              />
            )}
          </>
        }
      />



      <div className="relative flex flex-1 justify-center px-5.5 pt-14 pb-22">
        <div className="w-full max-w-content">
          <header className="mb-8.5 flex flex-col gap-3.5">
            <h1 className="text-display">{list.list?.name}</h1>
            <div className="flex items-center gap-3 text-meta text-secondary-foreground">
              {/* While somebody else is reading it, that is the more useful of the
                  two facts, so it takes the line. */}
              {live.watchers.length > 0 ? (
                <Presence watchers={live.watchers} />
              ) : (
                <span>
                  {t(sharingKey(list.list?.sharing ?? Sharing.UNSPECIFIED, list.list?.canEdit ?? false))}
                </span>
              )}
              <span className="h-3 w-px bg-border" />
              <span>{open.length === 0 ? t("list.nothingYet") : t("list.openCount", { count: open.length })}</span>
            </div>
          </header>

          {/* Emptiness is about what is left to do, not about what the List holds: a
              List whose Items are all ticked has nothing on it to read, and without
              this the add row draws its top border against nothing at all. */}
          {open.length === 0 ? (
            <EmptyState
              title={t(done.length > 0 ? "list.allDoneTitle" : "list.emptyTitle")}
              body={t(done.length > 0 ? "list.allDoneBody" : "list.emptyBody")}
            />
          ) : (
            <div className="flex flex-col">
              {open.map((item) => (
                <ListRow
                  key={item.uid}
                  label={item.label}
                  quantity={item.quantity}
                  notSynced={queued.itemUids.has(item.uid) || isProvisional(item.uid)}
                  addedByName={item.addedByName}
                  addedVia={viaLabel(t, item.addedViaToken)}
                  dueLabel={due.label(item.dueOn)}
                  overdue={isOverdue(item.dueOn, from)}
                  justTicked={justTickedByAnother(item)}
                  note={{ firstLine: item.noteFirstLine, remainingLines: item.noteRemainingLines }}
                  moreLinesLabel={(count) => t("note.moreLines", { count })}
                  onToggle={(next) => setDone.mutate({ itemUid: item.uid, done: next })}
                  onOpen={() => setOpenItemUid(item.uid)}
                  onRename={
                    canEdit ? (label) => rename.mutate({ itemUid: item.uid, label }) : undefined
                  }
                  labels={rowLabels}
                  menu={
                    canEdit ? (
                      <ItemMenu
                        quantity={item.quantity}
                        dueOn={item.dueOn}
                        actions={{
                          onView: () => setOpenItemUid(item.uid),
                          onSetDate: (dueOn) => editItem.mutate({ itemUid: item.uid, dueOn }),
                          onSetQuantity: (quantity) =>
                            editItem.mutate({ itemUid: item.uid, quantity }),
                          onDuplicate: () =>
                            addItem.mutate({
                              listUid,
                              label: item.label,
                              quantity: item.quantity,
                              dueOn: item.dueOn,
                            }),
                          onDelete: () => removeItem.mutate(item.uid),
                        }}
                      />
                    ) : undefined
                  }
                />
              ))}
            </div>
          )}

          {canEdit && (
            <AddRow
              placeholder={t("list.addItem")}
              onAdd={(item) => addItem.mutate({ listUid, ...item })}
              disabled={addItem.isPending}
              divided={open.length > 0}
              hint={t("addRow.hint")}
            />
          )}

          {done.length > 0 && (
            <DoneSection label={doneLabel(t, done, from)}>
              {done.map((item) => (
                <ListRow
                  key={item.uid}
                  label={item.label}
                  quantity={item.quantity}
                  addedByName={item.doneByName || item.addedByName}
                  done
                  justTicked={justTickedByAnother(item)}
                  onToggle={(next) => setDone.mutate({ itemUid: item.uid, done: next })}
                  onOpen={() => setOpenItemUid(item.uid)}
                  onRename={
                    canEdit ? (label) => rename.mutate({ itemUid: item.uid, label }) : undefined
                  }
                  labels={rowLabels}
                />
              ))}
            </DoneSection>
          )}
        </div>

        {openItem && (
          <NoteSheet
            item={openItem}
            crumbs={[list.list?.name ?? "", t("note.crumb")]}
            listName={list.list?.name ?? ""}
            canEdit={Boolean(canEdit)}
            status={saveNote.isPending ? t("note.saving") : undefined}
            onNoteChange={onNoteChange}
            onToggleDone={(done) => setDone.mutate({ itemUid: openItem.uid, done })}
            onRename={(label) => rename.mutate({ itemUid: openItem.uid, label })}
            onQuantityChange={(quantity) =>
              editItem.mutate({ itemUid: openItem.uid, quantity })
            }
            onDueChange={(dueOn) => editItem.mutate({ itemUid: openItem.uid, dueOn })}
            onClose={closeSheet}
            onOpenFull={() => {
              autosave.flush()
              void navigate({
                to: "/lists/$listUid/items/$itemUid",
                params: { listUid, itemUid: openItem.uid },
              })
            }}
          />
        )}
      </div>
    </AppShell>
  )
}

/**
 * doneLabel is the line the completed Items sit behind.
 *
 * "3 done today" while the day is going, because that is the number a Member recognises
 * as theirs. Anything ticked before today is counted plainly — it is history, not an
 * account of the afternoon.
 */
function doneLabel(t: Translate, done: Item[], from: Date): string {
  const todays = done.filter((item) => happenedToday(item.doneAt, from)).length
  return todays > 0
    ? t("list.doneToday", { count: todays })
    : t("list.doneEarlier", { count: done.length })
}

/** Which line under the title says who can reach this List. */
function sharingKey(sharing: Sharing, canEdit: boolean): string {
  if (sharing === Sharing.PRIVATE) {
    return "list.onlyYou"
  }
  if (sharing === Sharing.SPECIFIC) {
    return canEdit ? "list.namedCanEdit" : "list.namedReadOnly"
  }
  return canEdit ? "list.sharedCanEdit" : "list.sharedReadOnly"
}

/**
 * viaLabel says what an Item came through, when that was not a browser.
 *
 * Nothing for the ordinary case, so a List nobody scripts against reads exactly as it
 * did before tokens existed.
 */
function viaLabel(t: Translate, tokenName: string): string | undefined {
  return tokenName ? t("list.addedVia", { name: tokenName }) : undefined
}
