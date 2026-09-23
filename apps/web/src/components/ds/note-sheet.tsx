import { useCallback } from "react"
import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { cn } from "cn"
import type { Item } from "@nooks/api"

import { Checkbox } from "./checkbox"
import { DateField } from "./date-field"
import { EditableTitle } from "./editable-title"
import { NoteEditor } from "./note-editor"
import type { DueDate } from "@/lib/dates"
import { useDueLabel } from "@/lib/use-due-label"
import { useEscape } from "@/lib/use-escape"
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
   * Who else has this Note open, if anybody.
   *
   * One Note has one editor and the first to open it keeps it, so this is a name to
   * show rather than a state to argue with. canEdit is already false when it is set;
   * this says why, because a Note that has quietly stopped accepting typing is worse
   * than one that says who is in it.
   */
  heldBy?: string
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
  /** The Member asked to close: the exit starts now, and the List has to move with it. */
  onLeave: () => void
  /** The exit has played and the sheet is gone. */
  onClose: () => void
  /** Opens the same Note at full width. */
  onOpenFull: () => void
  /** True once the Member has asked to close, while the exit plays. */
  leaving: boolean
  /** Shown at the right of the footer, e.g. "Saving…". */
  status?: string
}

/**
 * The side sheet, per DESIGN.md §9: 520px, pinned below the chrome bar.
 *
 * On the sheet layer rather than the top of the stack: a row lifts its checkbox and
 * label above its own click target, so a sheet without a layer would sit under the rows
 * it is drawn over.
 *
 * Same order as the full screen so the two read as one thing at two widths, and every
 * field is the control that changes it. There is no edit mode.
 *
 * Closing plays the entrance backwards and tells the List once the movement is over,
 * so a panel that took 180ms to arrive does not vanish between frames. Browser Back is
 * not this: the address holds the sheet open, and by the time it changes there is
 * nothing left to move.
 */
export function NoteSheet({
  item,
  crumbs,
  listName,
  canEdit,
  heldBy,
  onNoteChange,
  onToggleDone,
  onRename,
  onQuantityChange,
  onDueChange,
  onLeave,
  leaving,
  onClose,
  onOpenFull,
  status,
}: NoteSheetProps) {
  const { t } = useTranslation()
  const due = useDueLabel()

  // Esc backs out of the sheet the way it backs out of a Note or Settings. It goes
  // through onLeave rather than onClose so it plays the exit rather than cutting it,
  // and it is ignored once the sheet is already on its way out.
  useEscape(
    useCallback(() => {
      if (!leaving) {
        onLeave()
      }
    }, [leaving, onLeave]),
  )

  /*
   * Tells the List the sheet has gone, once the exit has played.
   *
   * A listener on the element rather than an `onAnimationEnd` prop: React names the
   * event by asking the style object which vendor prefixes it owns up to, which jsdom
   * cannot answer, so the prop fires in a browser and is silent in a test.
   */
  const leaveWhenDone = useCallback(
    (node: HTMLElement | null) => {
      if (!node || !leaving) {
        return
      }
      // Only the sheet's own movement ends the sheet. Everything inside it animates
      // too, and an editor settling a caret must not close the panel around it.
      const done = (event: AnimationEvent) => {
        if (event.target === node) {
          onClose()
        }
      }
      node.addEventListener("animationend", done)
      return () => node.removeEventListener("animationend", done)
    },
    [leaving, onClose],
  )

  return (
    <aside
      ref={leaveWhenDone}
      className={cn(
        "absolute inset-y-0 right-0 z-20 flex w-sheet flex-col border-l border-border bg-background",
        leaving ? "pointer-events-none animate-sheet-out" : "animate-sheet-in",
      )}
    >
      <div className="flex h-chrome items-center gap-2.5 border-b border-hair pr-4 pl-5.5 text-micro text-muted-foreground">
        {/* The name truncates; the controls do not. A long List name used to squeeze
            "Open full" until its two words stacked. */}
        <span className="min-w-0 flex-1 truncate text-secondary-foreground">
          {crumbs.join(" / ")}
        </span>

        <button
          type="button"
          onClick={onOpenFull}
          className="flex h-control-toolbar shrink-0 items-center gap-1 rounded-md bg-secondary px-2 text-micro whitespace-nowrap text-secondary-foreground transition-colors hover:text-foreground"
        >
          <Icon name="fullScreen" size="small" className="size-3" />
          <span>{t("note.openFull")}</span>
        </button>

        {/* An icon, not a word: the crumb already says where this is, and a second
            label beside "Open full" would read as a second destination. */}
        <IconButton
          name="close"
          label={t("note.close")}
          onClick={onLeave}
          className="ml-1 shrink-0"
        />
      </div>

      {/* The room at the foot is a spacer rather than padding: this scrolls and is a
          flex container, and a flex container drops the padding on its end edge. */}
      <div className="flex flex-1 flex-col gap-5 overflow-y-auto px-7.5 pt-7.5 after:block after:h-14 after:shrink-0 after:content-['']">
        {heldBy !== undefined && (
          <p className="rounded-md bg-secondary px-3 py-2 text-meta text-secondary-foreground">
            {t("note.heldBy", { name: heldBy })}
          </p>
        )}

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
