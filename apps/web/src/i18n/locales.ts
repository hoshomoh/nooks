import type { Locale as DateLocale } from "date-fns"
import { enGB } from "date-fns/locale"

/**
 * The languages Nooks speaks.
 *
 * Adding one is a JSON file in `locales/` and an entry here — no code changes. Where a
 * language needs different date or number conventions, it names the date-fns locale to
 * load with it; otherwise dates follow the language code.
 */
export type LocaleCode = string

export type LocaleDefinition = {
  /** The BCP 47 tag, e.g. "en", "en-GB", "pt-BR". */
  code: LocaleCode
  /** How the language names itself, for the picker. */
  nativeName: string
  /** Loads the date-fns locale, so date words match the language. */
  loadDateLocale: () => Promise<DateLocale>
}

/**
 * Every locale with a file in `locales/`.
 *
 * English is the source language: `en.json` is where the design's copy lives, and every
 * other file is a translation of it.
 */
export const LOCALES: LocaleDefinition[] = [
  {
    code: "en",
    nativeName: "English",
    loadDateLocale: async () => enGB,
  },
]

/** DEFAULT_LOCALE is the fallback when nothing else matches. */
export const DEFAULT_LOCALE: LocaleCode = "en"

/** FALLBACK_DATE_LOCALE is used until the real one has loaded. */
export const FALLBACK_DATE_LOCALE: DateLocale = enGB

/**
 * matchLocale picks the closest supported locale for a browser's preference.
 *
 * "pt-BR" prefers Brazilian Portuguese, falls back to any Portuguese, then to English.
 * Pure, so the matching rules can be tested without a browser.
 */
export function matchLocale(preferred: readonly string[], supported: readonly LocaleCode[]): LocaleCode {
  for (const want of preferred) {
    const exact = supported.find((code) => code.toLowerCase() === want.toLowerCase())
    if (exact) {
      return exact
    }
    const base = want.split("-")[0]?.toLowerCase()
    const partial = supported.find((code) => code.split("-")[0]?.toLowerCase() === base)
    if (partial) {
      return partial
    }
  }
  return DEFAULT_LOCALE
}

/** localeCodes is every supported code, for the picker and for matching. */
export const localeCodes: LocaleCode[] = LOCALES.map((locale) => locale.code)

/** definitionFor returns a locale's definition, or the default's. */
export function definitionFor(code: LocaleCode): LocaleDefinition {
  return LOCALES.find((locale) => locale.code === code) ?? LOCALES[0]
}
