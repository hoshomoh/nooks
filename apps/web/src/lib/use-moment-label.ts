import { useTranslation } from "react-i18next"

import { momentLabel, parseMoment, type MomentLabelOptions } from "./dates"
import { useLocale } from "./use-locale"

/** Reads a stored timestamp as the Member would say it. */
export type FormatMoment = (at: string) => string

/** Wires the moment formatter to the active language. */
export function useMomentLabel(now: Date = new Date()): FormatMoment {
  const { t } = useTranslation()
  const { dateLocale } = useLocale()

  const options: MomentLabelOptions = {
    from: now,
    locale: dateLocale,
    yesterdayWord: t("date.yesterday"),
  }

  return (at) => {
    const parsed = parseMoment(at)
    return parsed ? momentLabel(parsed, options) : ""
  }
}
