import { useTranslation } from "react-i18next"

import { dayLabel, parseMoment, timeOfDay, type MomentLabelOptions } from "./dates"
import { useLocale } from "./use-locale"
import { useNow } from "./use-now"

export interface DoneDayLabels {
  /** Names a day the Member ticked things on: "Today", "Yesterday", "Friday", "3 Jul". */
  dayName: (at: Date) => string
  /** The clock time of one tick. The heading above it already says which day. */
  clock: (at: string) => string
}

/** Wires the two labels the completed section needs to the active language. */
export function useDoneDayLabels(now?: Date): DoneDayLabels {
  const ticking = useNow()
  const at = now ?? ticking
  const { t } = useTranslation()
  const { dateLocale } = useLocale()

  const options: MomentLabelOptions = {
    from: at,
    locale: dateLocale,
    yesterdayWord: t("date.yesterday"),
    todayWord: t("date.today"),
  }

  return {
    dayName: (at) => dayLabel(at, options),
    clock: (at) => {
      const parsed = parseMoment(at)
      return parsed ? timeOfDay(parsed, options) : ""
    },
  }
}
