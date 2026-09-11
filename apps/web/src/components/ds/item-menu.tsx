import { useState } from "react"
import { useTranslation } from "react-i18next"
import { cn } from "cn"

import { Button } from "./button"
import { IconButton } from "./icon-button"
import { Field } from "./field"
import { Icon } from "./icon"
import { DateCalendar } from "./date-calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { parseDue, toStored, type DueDate } from "@/lib/dates"
import { useLocale } from "@/lib/use-locale"

export interface ItemMenuActions {
  /** Opens the Item's sheet. */
  onView: () => void
  onSetDate: (dueOn: DueDate) => void
  onSetQuantity: (quantity: string) => void
  onDuplicate: () => void
  onDelete: () => void
}

export interface ItemMenuProps {
  /** What the Item carries now, so the menu opens on what is true. */
  quantity: string
  dueOn: DueDate
  actions: ItemMenuActions
}

/** Which of the menu's faces is showing. */
type Panel = "actions" | "date" | "quantity"

/**
 * The `···` on an Item's row.
 *
 * Setting a date or a quantity happens **inside the menu**: the entry that names the
 * field opens the field, rather than sending a Member somewhere else to find it. One
 * gesture, one place, and the row itself is left alone.
 *
 * A popover rather than a menu, because a menu is a list of actions and this holds a
 * calendar and a text field — controls that want the keyboard for themselves.
 *
 * Three of the entries the design lists are not here: moving between Lists, assigning
 * to somebody, and the note shortcut all wait for the milestones that build them. An
 * entry that does nothing teaches a Member that the menu is decoration.
 */
export function ItemMenu({ quantity, dueOn, actions }: ItemMenuProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [panel, setPanel] = useState<Panel>("actions")

  const close = () => {
    setOpen(false)
    setPanel("actions")
  }

  return (
    <Popover
      open={open}
      onOpenChange={(next) => {
        setOpen(next)
        if (!next) {
          setPanel("actions")
        }
      }}
    >
      <PopoverTrigger
        render={<IconButton name="more" scale="compact" label={t("itemMenu.open")} />}
      />

      <PopoverContent
        align="end"
        sideOffset={6}
        className="w-62 gap-0 rounded-menu border border-border p-1.5 shadow-[0_12px_32px_rgba(0,0,0,0.14)]"
      >
        {panel === "actions" && (
          <ActionsPanel
            onView={() => {
              close()
              actions.onView()
            }}
            onDate={() => setPanel("date")}
            onQuantity={() => setPanel("quantity")}
            onDuplicate={() => {
              close()
              actions.onDuplicate()
            }}
            onDelete={() => {
              close()
              actions.onDelete()
            }}
          />
        )}

        {panel === "date" && (
          <DatePanel
            dueOn={dueOn}
            onPick={(next) => {
              close()
              actions.onSetDate(next)
            }}
            onBack={() => setPanel("actions")}
          />
        )}

        {panel === "quantity" && (
          <QuantityPanel
            quantity={quantity}
            onSave={(next) => {
              close()
              actions.onSetQuantity(next)
            }}
            onBack={() => setPanel("actions")}
          />
        )}
      </PopoverContent>
    </Popover>
  )
}

interface ActionsPanelProps {
  onView: () => void
  onDate: () => void
  onQuantity: () => void
  onDuplicate: () => void
  onDelete: () => void
}

/** The menu's first face: what can be done to the Item. */
function ActionsPanel({ onView, onDate, onQuantity, onDuplicate, onDelete }: ActionsPanelProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col">
      <Entry label={t("itemMenu.view")} shortcut="↵" onSelect={onView} />
      {/* These two open a field rather than leaving: the entry names it, so the entry
          should be where it is. */}
      <Entry label={t("itemMenu.setDate")} opensPanel onSelect={onDate} />
      <Entry label={t("itemMenu.setQuantity")} opensPanel onSelect={onQuantity} />

      <span className="my-1.5 h-px bg-hair" />

      <Entry label={t("itemMenu.duplicate")} onSelect={onDuplicate} />
      <Entry label={t("itemMenu.delete")} shortcut="⌫" destructive onSelect={onDelete} />
    </div>
  )
}

interface EntryProps {
  label: string
  /** The key that does the same thing, e.g. "↵". Typed, because a keycap is a key. */
  shortcut?: string
  /** Whether choosing it turns the menu to another of its faces rather than acting. */
  opensPanel?: boolean
  destructive?: boolean
  onSelect: () => void
}

/** One row of the menu, at the size DESIGN.md §9 gives a menu item. */
function Entry({ label, shortcut, opensPanel, destructive, onSelect }: EntryProps) {
  return (
    <button
      type="button"
      onClick={onSelect}
      className={cn(
        "flex h-8 items-center rounded-md px-2.5 text-left text-chrome transition-colors",
        destructive ? "text-overdue hover:bg-destructive-bg" : "hover:bg-secondary",
      )}
    >
      {label}
      {shortcut && (
        <span className="ml-auto font-mono text-micro text-muted-foreground">{shortcut}</span>
      )}
      {opensPanel && <Icon name="forward" size="small" className="ml-auto text-muted-foreground" />}
    </button>
  )
}

interface DatePanelProps {
  dueOn: DueDate
  onPick: (dueOn: DueDate) => void
  onBack: () => void
}

/** The menu's date face: a calendar, and a way to take the date off. */
function DatePanel({ dueOn, onPick, onBack }: DatePanelProps) {
  const { t } = useTranslation()
  const { dateLocale } = useLocale()
  const selected = parseDue(dueOn) ?? undefined

  return (
    <div className="flex flex-col">
      <PanelHeading label={t("itemMenu.setDate")} onBack={onBack} />
      <DateCalendar
        selected={selected}
        onSelect={(day) => onPick(day ? toStored(day) : "")}
        locale={dateLocale}
      />
      {dueOn && (
        <button
          type="button"
          onClick={() => onPick("")}
          className="mt-1 border-t border-hair px-2.5 py-2 text-left text-small text-secondary-foreground hover:text-foreground"
        >
          {t("itemMenu.clearDate")}
        </button>
      )}
    </div>
  )
}

interface QuantityPanelProps {
  quantity: string
  onSave: (quantity: string) => void
  onBack: () => void
}

/** The menu's quantity face: one field, because a quantity is one word. */
function QuantityPanel({ quantity, onSave, onBack }: QuantityPanelProps) {
  const { t } = useTranslation()
  const [value, setValue] = useState(quantity)

  return (
    <form
      className="flex flex-col"
      onSubmit={(event) => {
        event.preventDefault()
        onSave(value.trim())
      }}
    >
      <PanelHeading label={t("itemMenu.setQuantity")} onBack={onBack} />
      <div className="px-1.5 pb-1.5">
        <Field
          label={t("note.quantity")}
          value={value}
          onChange={(event) => setValue(event.target.value)}
          autoFocus
        />
      </div>
      <div className="flex px-1.5 pb-1.5">
        <span className="flex-1" />
        <Button type="submit" scale="compact">
          {t("itemMenu.save")}
        </Button>
      </div>
    </form>
  )
}

interface PanelHeadingProps {
  label: string
  onBack: () => void
}

/** What the menu's second face is for, and the way back to the first. */
function PanelHeading({ label, onBack }: PanelHeadingProps) {
  const { t } = useTranslation()

  return (
    <div className="flex items-center gap-2 px-2.5 pt-1.5 pb-2">
      <IconButton
        name="back"
        scale="compact"
        label={t("itemMenu.back")}
        onClick={onBack}
        className="-ml-1"
      />
      <span className="text-label text-muted-foreground uppercase">{label}</span>
    </div>
  )
}
