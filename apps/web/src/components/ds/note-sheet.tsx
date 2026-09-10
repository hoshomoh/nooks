import { useTranslation } from "react-i18next"
import type { Item } from "@nooks/api"

import { Checkbox } from "./checkbox"
import { EditableTitle } from "./editable-title"
import { NoteEditor } from "./note-editor"

export type NoteSheetField = {
  label: string
  value: string
  /** An empty field reads as a prompt rather than a value. */
  empty?: boolean
}

export type NoteSheetProps = {
  item: Item
  crumbs: string[]
  fields: NoteSheetField[]
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
  onClose: () => void
  /** Opens the same Note at full width. */
  onOpenFull: () => void
  /** Shown at the right of the footer, e.g. "Saving…". */
  status?: string
}

/**
 * The side sheet, per DESIGN.md §9: 520px, pinned below the chrome bar.
 *
 * Same order as the full-screen view — chrome bar, checkbox and title, detail row,
 * hairline, Note, footer bar — so the two read as one thing at two widths. Opening an
 * Item never replaces the List with a page: the list keeps its place behind this.
 */
export function NoteSheet({
  item,
  crumbs,
  fields,
  canEdit,
  onNoteChange,
  onToggleDone,
  onRename,
  onClose,
  onOpenFull,
  status,
}: NoteSheetProps) {
  const { t } = useTranslation()

  return (
    <aside className="absolute inset-y-0 right-0 flex w-sheet animate-sheet-in flex-col border-l border-border bg-background">
      <div className="flex h-chrome items-center gap-2.5 border-b border-hair pr-4 pl-5.5 text-micro text-muted-foreground">
        <span className="text-secondary-foreground">{crumbs.join(" / ")}</span>
        <span className="flex-1" />
        <button type="button" onClick={onOpenFull} className="text-small hover:text-foreground">
          {t("note.openFull")}
        </button>
        <button type="button" onClick={onClose} className="text-small hover:text-foreground">
          {t("note.close")}
        </button>
      </div>

      <div className="flex flex-1 flex-col gap-5 overflow-y-auto px-7 pt-7">
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
            className="text-page"
          />
        </div>

        <div className="flex flex-wrap gap-x-4.5 gap-y-2.5 pl-9">
          {fields.map((field) => (
            <span key={field.label} className="flex items-center gap-2">
              <span className="text-micro text-muted-foreground">{field.label}</span>
              <span className={field.empty ? "text-small text-muted-foreground" : "text-small"}>
                {field.value}
              </span>
            </span>
          ))}
        </div>

        <div className="h-px bg-hair" />

        <div className="flex min-h-[240px] flex-1 pl-9">
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
