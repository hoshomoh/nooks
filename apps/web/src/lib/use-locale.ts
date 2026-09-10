import { useSyncExternalStore } from "react"
import type { Locale as DateLocale } from "date-fns"

import { localeStore } from "./locale-store"
import type { LocaleCode } from "@/i18n/locales"

export type UseLocale = {
  code: LocaleCode
  /** The date-fns locale that goes with the language. */
  dateLocale: DateLocale
  setCode: (code: LocaleCode) => void
}

/** Reads the active locale from its store. */
export function useLocale(): UseLocale {
  const code = useSyncExternalStore(localeStore.subscribe, localeStore.getCode, localeStore.getCode)
  const dateLocale = useSyncExternalStore(
    localeStore.subscribe,
    localeStore.getDateLocale,
    localeStore.getDateLocale,
  )
  return { code, dateLocale, setCode: localeStore.setCode }
}
