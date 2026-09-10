import { useTranslation } from "react-i18next"

import { dayHeading, dueLabel, type DateLabelOptions, type DueDate } from "./dates"
import { today } from "./dates"
import { useLocale } from "./use-locale"

export type UseDueLabel = {
  /** What a row shows beside an Item. */
  label: (due: DueDate) => string
  /** What a group of Items is headed with. */
  heading: (due: DueDate) => string
}

/**
 * Wires the date utility to the active language.
 *
 * The words and the date-fns locale come from one place, so a Member never sees English
 * weekdays under German headings.
 */
export function useDueLabel(now: Date = new Date()): UseDueLabel {
  const { t } = useTranslation()
  const { dateLocale } = useLocale()

  const options: DateLabelOptions = {
    from: today(now),
    locale: dateLocale,
    todayWord: t("date.today"),
    tomorrowWord: t("date.tomorrow"),
  }

  return {
    label: (due) => dueLabel(due, options),
    heading: (due) => dayHeading(due, options),
  }
}
