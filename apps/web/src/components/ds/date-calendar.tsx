import { getDefaultClassNames, type Locale } from "react-day-picker"
import { cn } from "cn"

import { Calendar } from "@/components/ui/calendar"

export interface DateCalendarProps {
  /** The day already chosen, or undefined for none. */
  selected?: Date
  onSelect: (day: Date | undefined) => void
  /** The Member's locale, so the week starts and reads where they expect. */
  locale: Locale
}

/**
 * The app's calendar: one month, one day at a time, filling whatever holds it.
 *
 * The installed calendar sizes itself to its cells, which leaves a margin inside any
 * container wider than that and reads as a calendar that failed to load rather than one
 * that is finished. Only its root is overridden here — the cells already divide the
 * width once the root is allowed to take it.
 *
 * It wraps the installed component rather than changing it: that one is upstream's, and
 * a local edit is lost the next time it is pulled.
 */
export function DateCalendar({ selected, onSelect, locale }: DateCalendarProps) {
  const defaults = getDefaultClassNames()

  return (
    <Calendar
      mode="single"
      selected={selected}
      defaultMonth={selected}
      onSelect={onSelect}
      locale={locale}
      autoFocus
      classNames={{ root: cn("w-full", defaults.root) }}
    />
  )
}
