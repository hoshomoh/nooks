import type { Locale as DateLocale } from "date-fns"

import i18n, { browserLocale } from "@/i18n"
import {
  DEFAULT_LOCALE,
  FALLBACK_DATE_LOCALE,
  definitionFor,
  type LocaleCode,
} from "@/i18n/locales"

/**
 * The active locale, as an external store.
 *
 * One place decides the language, and both halves of localisation follow it: i18next
 * for the words, and a date-fns locale for how dates read. Keeping them in one store is
 * what stops the two drifting — a Member reading German dates in an English interface.
 *
 * React reads this with useSyncExternalStore rather than synchronising with an effect.
 */
export type LocaleStore = {
  subscribe: (listener: () => void) => () => void
  getCode: () => LocaleCode
  getDateLocale: () => DateLocale
  setCode: (code: LocaleCode) => void
}

export type LocaleStoreDeps = {
  storage: Pick<Storage, "getItem" | "setItem">
  /** The language to start in when nothing is remembered. */
  initial: LocaleCode
  /** Applies the language to i18next. Injected so a test needs no i18next. */
  applyLanguage: (code: LocaleCode) => void
  /** Loads the date-fns locale that goes with a language. */
  loadDateLocale: (code: LocaleCode) => Promise<DateLocale>
}

/** LOCALE_STORAGE_KEY is where a Member's choice is remembered. */
export const LOCALE_STORAGE_KEY = "nooks.locale"

export function createLocaleStore(deps: LocaleStoreDeps): LocaleStore {
  let code = readStored(deps.storage) ?? deps.initial
  let dateLocale: DateLocale = FALLBACK_DATE_LOCALE
  const listeners = new Set<() => void>()

  const notify = () => {
    for (const listener of listeners) {
      listener()
    }
  }

  const apply = (next: LocaleCode) => {
    code = next
    deps.applyLanguage(next)
    notify()

    // The date words arrive a moment later; until then dates read in the fallback,
    // which is better than not rendering them at all.
    void deps.loadDateLocale(next).then((loaded) => {
      dateLocale = loaded
      notify()
    })
  }

  apply(code)

  return {
    subscribe(listener) {
      listeners.add(listener)
      return () => listeners.delete(listener)
    },
    getCode: () => code,
    getDateLocale: () => dateLocale,
    setCode(next) {
      if (next === code) {
        return
      }
      try {
        deps.storage.setItem(LOCALE_STORAGE_KEY, next)
      } catch {
        // A Member in a private window still gets the language for this session.
      }
      apply(next)
    },
  }
}

/** readStored returns a remembered language, or null. */
function readStored(storage: Pick<Storage, "getItem">): LocaleCode | null {
  try {
    return storage.getItem(LOCALE_STORAGE_KEY)
  } catch {
    return null
  }
}

/** The application's store. */
export const localeStore: LocaleStore =
  typeof window === "undefined"
    ? {
        subscribe: () => () => {},
        getCode: () => DEFAULT_LOCALE,
        getDateLocale: () => FALLBACK_DATE_LOCALE,
        setCode: () => {},
      }
    : createLocaleStore({
        storage: window.localStorage,
        initial: browserLocale(),
        applyLanguage: (next) => void i18n.changeLanguage(next),
        loadDateLocale: (next) => definitionFor(next).loadDateLocale(),
      })
