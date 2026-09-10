import type { LocaleCode } from "@/i18n/locales"

/**
 * Numbers and money, in whichever language a Member reads.
 *
 * Intl is right for these — unlike month names, number and currency conventions are
 * exactly what it is for, and there is no design-specified rendering to conflict with.
 * The locale is always passed in rather than read from the environment, so nothing
 * renders in one language while the rest of the page is in another.
 */

export type FormatNumberOptions = Intl.NumberFormatOptions

/** formatNumber renders a number in the Member's language. */
export function formatNumber(
  value: number,
  locale: LocaleCode,
  options: FormatNumberOptions = {},
): string {
  return new Intl.NumberFormat(locale, options).format(value)
}

/**
 * formatCurrency renders an amount with its currency, e.g. "€12.50" or "12,50 €"
 * depending on the language.
 */
export function formatCurrency(value: number, currency: string, locale: LocaleCode): string {
  return new Intl.NumberFormat(locale, { style: "currency", currency }).format(value)
}

/** formatList joins items the way the language does, e.g. "Anna, Jonas and Mira". */
export function formatList(items: readonly string[], locale: LocaleCode): string {
  return new Intl.ListFormat(locale, { style: "long", type: "conjunction" }).format(items)
}
