import { useCallback, useMemo, useState } from "react"
import { useMutation, useQueryClient, useSuspenseQuery } from "@tanstack/react-query"
import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { ActivityControl } from "@/components/ds/activity-control"
import { Presence } from "@/components/ds/presence"
import { AddRow, type AddRowSubmission } from "@/components/ds/add-row"
import { DoneSection } from "@/components/ds/done-section"
import { AppShell } from "@/components/ds/app-shell"
import { ChromeBar } from "@/components/ds/chrome-bar"
import { EmptyState } from "@/components/ds/empty-state"
import { ListRow, type ListRowLabels } from "@/components/ds/list-row"
import { NoteSheet, type NoteSheetField } from "@/components/ds/note-sheet"
import { ShareDialog, type ShareDecision } from "@/components/ds/share-dialog"
import { Button } from "@/components/ds/button"
import { listClient } from "@/lib/api"
import { listQuery } from "@/lib/list-queries"
import { refreshLists } from "@/lib/refresh"
import type { Translate } from "@/lib/translate"
import { debounce } from "@/lib/debounce"
import type { RenameItemVariables, SaveNoteVariables, SetDoneVariables } from "@/lib/item-mutations"
import { happenedToday, isOverdue, justHappened, today } from "@/lib/dates"
import { useDueLabel } from "@/lib/use-due-label"
import { useCommandPalette } from "@/lib/use-command-palette"
import { useLive } from "@/lib/use-live"
import { useSignedInData } from "@/lib/use-signed-in-data"
import { Sharing, type Item } from "@nooks/api"

const route = getRouteApi("/lists/$listUid")

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

  // A tick somebody else just made lands with a highlight and then settles. Read from
  // the Item itself rather than remembered between renders: the Item already says when
  // it was ticked and by whom, so there is nothing to keep in sync.
  const settled = new Date()
  const justTickedByAnother = (item: Item) =>
    item.done && item.doneByUid !== member?.uid && justHappened(item.doneAt, settled)

  // Which Item's sheet is open. The List keeps its place behind it, so this is the
  // screen's own state rather than a route.
  const [openItemUid, setOpenItemUid] = useState<string | null>(null)

  // Sharing is a decision the owner makes in a dialog, so whether it is open is the
  // screen's own state.
  const [sharingOpen, setSharingOpen] = useState(false)

  const refresh = () => refreshLists(queryClient)

  const addItem = useMutation({
    mutationFn: ({ label, quantity, dueOn }: AddRowSubmission) =>
      listClient.createItem({ listUid, label, quantity, dueOn }),
    onSuccess: refresh,
  })

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

  const setDone = useMutation({
    mutationFn: ({ itemUid, done }: SetDoneVariables) => listClient.setItemDone({ itemUid, done }),
    onSuccess: refresh,
  })

  const rename = useMutation({
    mutationFn: ({ itemUid, label }: RenameItemVariables) =>
      listClient.updateItem({ itemUid, label }),
    onSuccess: refresh,
  })

  const share = useMutation({
    mutationFn: (decision: ShareDecision) =>
      listClient.setListSharing({ listUid, ...decision }),
    onSuccess: refresh,
  })

  const rowLabels: ListRowLabels = { name: t("note.itemName"), open: t("list.openItem") }

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
      <ChromeBar
        crumbs={[t(list.list?.isOwner ? "list.myListsCrumb" : "list.sharedCrumb"), list.list?.name ?? ""]}
        actions={
          <>
            {list.list?.isOwner && (
              <Button tone="secondary" scale="toolbar" onClick={() => setSharingOpen(true)}>
                {t("share.action")}
              </Button>
            )}
            <ActivityControl />
          </>
        }
      />

      {list.list && (
        <ShareDialog
          list={list.list}
          instanceName={instanceName}
          open={sharingOpen}
          onOpenChange={setSharingOpen}
          onSave={(decision) => share.mutate(decision)}
        />
      )}

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
                  justTicked={justTickedByAnother(item)}
                  note={{ firstLine: item.noteFirstLine, remainingLines: item.noteRemainingLines }}
                  moreLinesLabel={(count) => t("note.moreLines", { count })}
                  onToggle={(next) => setDone.mutate({ itemUid: item.uid, done: next })}
                  onOpen={() => setOpenItemUid(item.uid)}
                  onRename={
                    canEdit ? (label) => rename.mutate({ itemUid: item.uid, label }) : undefined
                  }
                  labels={rowLabels}
                />
              ))}
            </div>
          )}

          {canEdit && (
            <AddRow
              placeholder={t("list.addItem")}
              onAdd={(item) => addItem.mutate(item)}
              disabled={addItem.isPending}
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
            fields={noteFields(t, openItem, list.list?.name ?? "", due.label(openItem.dueOn))}
            canEdit={Boolean(canEdit)}
            status={saveNote.isPending ? t("note.saving") : undefined}
            onNoteChange={onNoteChange}
            onToggleDone={(done) => setDone.mutate({ itemUid: openItem.uid, done })}
            onRename={(label) => rename.mutate({ itemUid: openItem.uid, label })}
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

/** The detail row in the sheet: what is known about the Item, and what is not yet. */
function noteFields(
  t: Translate,
  item: Item,
  listName: string,
  dueLabel: string,
): NoteSheetField[] {
  return [
    { label: t("note.list"), value: listName },
    { label: t("note.quantity"), value: item.quantity || t("note.add"), empty: !item.quantity },
    { label: t("note.due"), value: dueLabel || t("note.addDate"), empty: !item.dueOn },
    { label: t("note.addedBy"), value: item.addedByName },
  ]
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
