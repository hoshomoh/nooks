import { useState } from "react"
import { useTranslation } from "react-i18next"
import { cn } from "cn"

import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { parseDue, toStored, type DueDate } from "@/lib/dates"
import { useLocale } from "@/lib/use-locale"

export interface DateFieldProps {
  /** The date it carries, stored. Empty for none. */
  value: DueDate
  /** How that date reads, or the words for having none. */
  label: string
  /**
   * Whether this date was chosen, rather than supplied by the view.
   *
   * A chosen date is stated in the accent; one the view supplied is offered quietly,
   * because it applies until the Member says otherwise.
   */
  chosen: boolean
  onChange: (value: DueDate) => void
  /** Opened from outside, when something other than the pill asks for it. */
  open?: boolean
  onOpenChange?: (open: boolean) => void
}

/**
 * The date control, per DESIGN.md §6: a pill in the add row that is always present,
 * whether or not a date was typed.
 *
 * The calendar is the design system's own rather than the browser's: a native date
 * input looks different in every browser, and this is a control a Member sees on every
 * row they add.
 */
export function DateField({
  value,
  label,
  chosen,
  onChange,
  open: openFromOutside,
  onOpenChange,
}: DateFieldProps) {
  const { t } = useTranslation()
  const { dateLocale } = useLocale()
  const [openOnItsOwn, setOpenOnItsOwn] = useState(false)

  const open = openFromOutside ?? openOnItsOwn
  const setOpen = onOpenChange ?? setOpenOnItsOwn

  const selected = parseDue(value) ?? undefined

  const choose = (day: Date | undefined) => {
    onChange(day ? toStored(day) : "")
    setOpen(false)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <button
            type="button"
            aria-label={t("date.dueLabel")}
            className={cn(
              "flex h-6.5 shrink-0 items-center gap-1.5 rounded-md px-2.5 text-micro",
              "transition-colors",
              chosen
                ? "bg-shared-bg text-shared"
                : "border border-border text-secondary-foreground hover:bg-secondary",
            )}
          >
            <CalendarGlyph />
            <span className="whitespace-nowrap">{label}</span>
          </button>
        }
      />

      <PopoverContent align="end" className="w-auto gap-0 p-0">
        <Calendar
          mode="single"
          selected={selected}
          defaultMonth={selected}
          onSelect={choose}
          locale={dateLocale}
          autoFocus
        />
        {value && (
          <button
            type="button"
            onClick={() => choose(undefined)}
            className="border-t border-hair px-3 py-2 text-left text-small text-secondary-foreground hover:text-foreground"
          >
            {t("date.clearDate")}
          </button>
        )}
      </PopoverContent>
    </Popover>
  )
}

/** The calendar glyph, drawn at the size the row uses. */
function CalendarGlyph() {
  return (
    <svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden className="shrink-0">
      <rect x="1" y="2.2" width="10" height="9" rx="1.6" stroke="currentColor" strokeWidth="1.1" />
      <path
        d="M1 5h10M3.6 1v2.2M8.4 1v2.2"
        stroke="currentColor"
        strokeWidth="1.1"
        strokeLinecap="round"
      />
    </svg>
  )
}
