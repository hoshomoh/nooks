import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"
import type { Item } from "@nooks/api"

import { Checkbox } from "./checkbox"
import { DateField } from "./date-field"
import { EditableTitle } from "./editable-title"
import { NoteEditor } from "./note-editor"
import type { DueDate } from "@/lib/dates"
import { useDueLabel } from "@/lib/use-due-label"
import { Icon } from "./icon"
import { IconButton } from "./icon-button"

export interface NoteSheetProps {
  item: Item
  crumbs: string[]
  /** Which List the Item is on. Moving between Lists is not a sheet-level action. */
  listName: string
  /** Whether the Member may change anything here. */
  canEdit: boolean
  /**
   * Called on every change. Debouncing belongs to whoever owns the saving, so this
   * component stays a renderer.
   */
  onNoteChange: (markdown: string) => void
  onToggleDone: (done: boolean) => void
  /** Renames the Item. Its title is the field that renders it. */
  onRename: (label: string) => void
  onQuantityChange: (quantity: string) => void
  onDueChange: (dueOn: DueDate) => void
  onClose: () => void
  /** Opens the same Note at full width. */
  onOpenFull: () => void
  /** Shown at the right of the footer, e.g. "Saving…". */
  status?: string
}

/**
 * The side sheet, per DESIGN.md §9: 520px, pinned below the chrome bar.
 *
 * It sits on the sheet layer rather than at the top of the stack: a row's checkbox and
 * label are lifted above the row's own click target, and without a layer of its own the
 * sheet would be covered by the rows it is drawn over.
 *
 * Same order as full screen — chrome bar, checkbox and title, detail row, hairline,
 * Note, footer bar — so the two read as one thing at two widths. Opening an Item never
 * replaces the List with a page.
 *
 * Every field here is the control that changes it. There is no edit mode: the sheet is
 * where an Item is read and where it is rewritten.
 */
export function NoteSheet({
  item,
  crumbs,
  listName,
  canEdit,
  onNoteChange,
  onToggleDone,
  onRename,
  onQuantityChange,
  onDueChange,
  onClose,
  onOpenFull,
  status,
}: NoteSheetProps) {
  const { t } = useTranslation()
  const due = useDueLabel()

  return (
    <aside className="absolute inset-y-0 right-0 z-20 flex w-sheet animate-sheet-in flex-col border-l border-border bg-background">
      <div className="flex h-chrome items-center gap-2.5 border-b border-hair pr-4 pl-5.5 text-micro text-muted-foreground">
        <span className="text-secondary-foreground">{crumbs.join(" / ")}</span>
        <span className="flex-1" />

        <button
          type="button"
          onClick={onOpenFull}
          className="flex h-control-toolbar items-center gap-1 rounded-md bg-secondary px-2 text-micro text-secondary-foreground transition-colors hover:text-foreground"
        >
          <Icon name="fullScreen" size="small" className="size-3" />
          <span>{t("note.openFull")}</span>
        </button>

        {/* An icon, not a word: the crumb already says where this is, and a second
            label beside "Open full" would read as a second destination. */}
        <IconButton name="close" label={t("note.close")} onClick={onClose} className="ml-1" />
      </div>

      <div className="flex flex-1 flex-col gap-5 overflow-y-auto px-7.5 pt-7.5">
        <div className="grid grid-cols-[24px_1fr] items-start gap-3">
          <span className="mt-1">
            <Checkbox checked={item.done} onCheckedChange={onToggleDone} disabled={!canEdit} />
          </span>
          <EditableTitle
            value={item.label}
            onCommit={onRename}
            readOnly={!canEdit}
            label={t("note.itemName")}
            as="h2"
            className="text-sheet-title"
          />
        </div>

        <div className="flex flex-wrap items-center gap-x-4.5 gap-y-2.5 pl-9">
          <Detail label={t("note.list")}>
            <span className="text-small">{listName}</span>
          </Detail>

          <Detail label={t("note.quantity")}>
            <EditableTitle
              value={item.quantity}
              onCommit={onQuantityChange}
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
              label={item.dueOn ? due.label(item.dueOn) : t("note.addDate")}
              chosen={Boolean(item.dueOn)}
              onChange={onDueChange}
            />
          </Detail>

          <Detail label={t("note.addedBy")}>
            <span className="text-small">{item.addedByName}</span>
          </Detail>
        </div>

        <div className="h-px bg-hair" />

        <div className="flex min-h-60 flex-1 pl-9">
          {/* Keyed on the Item: opening a different one builds a fresh editor. */}
          <NoteEditor
            key={item.uid}
            initialValue={item.note}
            onChange={onNoteChange}
            readOnly={!canEdit}
          />
        </div>
      </div>

      <div className="flex h-sheet-footer items-center gap-4 border-t border-hair px-5.5 text-micro text-muted-foreground">
        <span>{t("note.hint")}</span>
        <span className="flex-1" />
        {status && <span>{status}</span>}
      </div>
    </aside>
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
