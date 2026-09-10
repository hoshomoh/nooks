import { useTranslation } from "react-i18next"

import { parseAddRow, type AddRowParse, type AddRowVocabulary, type ChipKind } from "./add-row"
import { today } from "./dates"
import { useLocale } from "./use-locale"

/** Reads one line of the add row, in the Member's language. */
export type ParseAddRowLine = (text: string, taken?: readonly ChipKind[]) => AddRowParse

/**
 * Wires the add row's parser to the active language.
 *
 * Weekday and month names come from the date-fns locale, so a new language needs only
 * the handful of words below in its JSON file — not a calendar of its own.
 */
export function useAddRowParse(now: Date = new Date()): ParseAddRowLine {
  const { t } = useTranslation()
  const { dateLocale } = useLocale()

  const vocabulary: AddRowVocabulary = {
    todayWords: wordList(t("addRow.todayWords", { returnObjects: true })),
    tomorrowWords: wordList(t("addRow.tomorrowWords", { returnObjects: true })),
    units: wordList(t("addRow.units", { returnObjects: true })),
  }

  return (text, taken) =>
    parseAddRow(text, { from: today(now), locale: dateLocale, vocabulary, taken })
}

/**
 * wordList reads a list of words from the locale file.
 *
 * A missing or mistyped key is a translation that is still being written, not a reason
 * to stop parsing: the row keeps working with one fewer kind of word.
 */
function wordList(value: unknown): readonly string[] {
  if (!Array.isArray(value)) {
    return []
  }
  return value.filter((word): word is string => typeof word === "string")
}
