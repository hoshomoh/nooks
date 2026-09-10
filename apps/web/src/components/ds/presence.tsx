import { useTranslation } from "react-i18next"

import { formatList } from "@/lib/format"
import { useLocale } from "@/lib/use-locale"

export interface PresenceProps {
  /** Who else is looking at this List. Empty when nobody is. */
  watchers: string[]
}

/**
 * Who else is here, per DESIGN.md §6.
 *
 * It replaces the sharing line rather than sitting beside it: while somebody else is
 * reading the same List, that is the more useful of the two facts, and the design gives
 * the line one job at a time.
 *
 * The colour is `--done`, the same green a tick lands in — it means "somebody else is
 * doing something", which is exactly what it is.
 */
export function Presence({ watchers }: PresenceProps) {
  const { t } = useTranslation()
  const { code } = useLocale()
  if (watchers.length === 0) {
    return null
  }

  return (
    <span className="text-done">
      {watchers.length === 1
        ? t("list.isHere", { name: watchers[0] })
        : t("list.areHere", { names: formatList(watchers, code) })}
    </span>
  )
}
