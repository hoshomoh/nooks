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
 * It replaces the sharing line rather than sitting beside it: the design gives the line
 * one job at a time, and while somebody else is reading the List that is the more
 * useful fact. The colour is `--done`, the green a tick lands in.
 *
 * It fades up rather than swapping between frames. This is the one line that changes
 * without the Member having done anything, so it is worth a moment to notice.
 */
export function Presence({ watchers }: PresenceProps) {
  const { t } = useTranslation()
  const { code } = useLocale()
  if (watchers.length === 0) {
    return null
  }

  return (
    <span className="text-done transition-opacity duration-120 ease-out starting:opacity-0">
      {watchers.length === 1
        ? t("list.isHere", { name: watchers[0] })
        : t("list.areHere", { names: formatList(watchers, code) })}
    </span>
  )
}
