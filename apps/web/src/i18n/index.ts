import i18n from "i18next"
import { initReactI18next } from "react-i18next"
import type { BackendModule } from "i18next"

import { DEFAULT_LOCALE, localeCodes, matchLocale } from "./locales"

/**
 * Loads a locale's JSON on demand.
 *
 * Every language is one file in `locales/`, so adding a language is adding a file and
 * an entry in `locales.ts` — never a code change. Files are imported lazily so a
 * Member downloads only the language they read.
 */
const lazyLocale: BackendModule = {
  type: "backend",
  init: () => {},
  read: (language, _namespace, callback) => {
    const code = localeCodes.includes(language) ? language : DEFAULT_LOCALE
    import(`./locales/${code}.json`)
      .then((module) => callback(null, module.default ?? module))
      .catch((error: unknown) => callback(error as Error, false))
  },
}

/** browserLocale is the closest supported language to what the browser asks for. */
export function browserLocale(): string {
  if (typeof navigator === "undefined") {
    return DEFAULT_LOCALE
  }
  return matchLocale(navigator.languages ?? [navigator.language], localeCodes)
}

void i18n
  .use(lazyLocale)
  .use(initReactI18next)
  .init({
    lng: browserLocale(),
    fallbackLng: DEFAULT_LOCALE,
    supportedLngs: localeCodes,
    // React escapes everything it renders already.
    interpolation: { escapeValue: false },
  })

export default i18n
