import { useCallback, useMemo } from "react"
import type { ReactNode } from "react"
import { useMutation, useQueryClient, useSuspenseQuery } from "@tanstack/react-query"
import { getRouteApi, useNavigate } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ds/button"
import { Checkbox } from "@/components/ds/checkbox"
import { ChromeBar } from "@/components/ds/chrome-bar"
import { DateField } from "@/components/ds/date-field"
import { EditableTitle } from "@/components/ds/editable-title"
import { NoteEditor } from "@/components/ds/note-editor"
import { listClient } from "@/lib/api"
import { debounce } from "@/lib/debounce"
import { listQuery } from "@/lib/list-queries"
import { refreshLists } from "@/lib/refresh"
import { useDueLabel } from "@/lib/use-due-label"
import { useEscape } from "@/lib/use-escape"

const route = getRouteApi("/lists/$listUid/items/$itemUid")

/** The fields an edit may change. Either may be left alone. */
interface ItemFields {
  quantity?: string
  dueOn?: string
}

/** How long the typing has to settle before a Note is saved. */
const AUTOSAVE_DELAY_MS = 800

/**
 * The Note at full width.
 *
 * The same blocks as the side sheet, one step larger — a Note is a document, and this
 * is it with room to be one. Esc returns to the List.
 */
export function NoteScreen() {
  const { listUid, itemUid } = route.useParams()
  const listData = useSuspenseQuery(listQuery(listUid)).data
  const list = listData.list
  const item = listData.items.find((candidate) => candidate.uid === itemUid)
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const due = useDueLabel()

  const refresh = useCallback(() => refreshLists(queryClient), [queryClient])

  const save = useMutation({
    mutationFn: (note: string) => listClient.updateItem({ itemUid, note }),
    onSuccess: refresh,
  })
  const autosave = useMemo(() => debounce(save.mutate, AUTOSAVE_DELAY_MS), [save.mutate])

  const setDone = useMutation({
    mutationFn: (done: boolean) => listClient.setItemDone({ itemUid, done }),
    onSuccess: refresh,
  })

  const rename = useMutation({
    mutationFn: (label: string) => listClient.updateItem({ itemUid, label }),
    onSuccess: refresh,
  })

  // Quantity and due date are two fields of the same edit, so they are one mutation.
  const edit = useMutation({
    mutationFn: (fields: ItemFields) => listClient.updateItem({ itemUid, ...fields }),
    onSuccess: refresh,
  })

  const back = useCallback(() => {
    autosave.flush()
    void navigate({ to: "/lists/$listUid", params: { listUid } })
  }, [autosave, navigate, listUid])

  useEscape(back)

  // The Item can go while its Note is open — somebody else deleted it, or the List was
  // unshared. The route's loader has already ruled out the case of arriving at one that
  // was never there.
  if (!list || !item) {
    return null
  }

  const canEdit = Boolean(list.isOwner || list.canEdit)

  return (
    <div className="flex min-h-dvh flex-col bg-background">
      <ChromeBar
        crumbs={[list.name, t("note.crumb")]}
        actions={
          <>
            <span className="text-micro text-muted-foreground">
              {save.isPending ? t("note.saving") : t("note.saved")}
            </span>
            <Button tone="secondary" scale="toolbar" onClick={back}>
              {t("note.backToList")}
            </Button>
          </>
        }
      />

      <div className="flex flex-1 justify-center px-8 pt-11">
        <div className="flex w-full max-w-content flex-col gap-5.5">
          <div className="grid grid-cols-[24px_1fr] items-start gap-3.5">
            <span className="mt-2">
              <Checkbox
                checked={item.done}
                disabled={!canEdit}
                onCheckedChange={(done) => setDone.mutate(Boolean(done))}
              />
            </span>
            <EditableTitle
              value={item.label}
              onCommit={(label) => rename.mutate(label)}
              readOnly={!canEdit}
              label={t("note.itemName")}
              className="text-display"
            />
          </div>

          {/* The same fields as the sheet, at the same scale: an Item is read and
              rewritten in both, and a field that is a control in one and text in the
              other would be a rule a Member has to learn. */}
          <div className="flex flex-wrap items-center gap-4.5 pl-9.5">
            <Detail label={t("note.list")}>
              <span className="text-small">{list.name}</span>
            </Detail>

            <Detail label={t("note.quantity")}>
              <EditableTitle
                value={item.quantity}
                onCommit={(quantity) => edit.mutate({ quantity })}
                readOnly={!canEdit}
                label={t("note.quantity")}
                as="span"
                className="w-24 text-small"
                placeholder={t("note.add")}
              />
            </Detail>

            <Detail label={t("note.due")}>
              <DateField
                value={item.dueOn}
                label={due.label(item.dueOn) || t("note.addDate")}
                chosen={Boolean(item.dueOn)}
                onChange={(dueOn) => edit.mutate({ dueOn })}
              />
            </Detail>

            <Detail label={t("note.addedBy")}>
              <span className="text-small">{item.addedByName}</span>
            </Detail>
          </div>

          <div className="h-px bg-hair" />

          <div className="flex flex-1 pl-9.5">
            <NoteEditor
              key={item.uid}
              initialValue={item.note}
              onChange={autosave.call}
              readOnly={!canEdit}
              scale="full"
            />
          </div>
        </div>
      </div>

      <div className="flex h-sheet-footer items-center gap-4 border-t border-hair px-5.5 text-micro text-muted-foreground">
        <span>{t("note.hint")}</span>
        <span className="flex-1" />
        <span>{t("note.escapeReturns")}</span>
      </div>
    </div>
  )
}

interface DetailProps {
  label: string
  children: ReactNode
}

/** One fact about the Item, and the control that changes it. */
function Detail({ label, children }: DetailProps) {
  return (
    <span className="flex items-center gap-2">
      <span className="text-micro text-muted-foreground">{label}</span>
      {children}
    </span>
  )
}
