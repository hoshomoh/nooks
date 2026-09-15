import type { Locale as DateLocale } from "date-fns"

import {
  fold,
  monthNames,
  nextDayOfMonth,
  nextWeekday,
  shift,
  toStored,
  weekdayNames,
  type DueDate,
  type Weekday,
} from "./dates"

/**
 * The add row's parser: the sentence a Member typed, read as an Item.
 *
 * A quantity and a date are lifted out of the text so `Tomatoes 1kg sat` files an Item
 * called "Tomatoes" due on Saturday. Everything the parser is unsure of stays as the
 * name, because a wrong silent chip is worse than no chip — the Member has to notice it
 * before they can fix it.
 *
 * Pure, so the chips, the date control and the created Item all read one result, and so
 * the rules are testable without the UI.
 */

/** ChipKind says which of the two values a chip carries. */
export type ChipKind = "quantity" | "due"

/** A value lifted out of the sentence, as the row draws it. */
export interface AddRowChip {
  kind: ChipKind
  /** What the chip reads, already normalised. */
  label: string
  /** The text it was lifted from, which backspace puts back. */
  source: string
}

/** What a language contributes to parsing, beyond the words its calendar already has. */
export interface AddRowVocabulary {
  /** Words meaning the current day, such as "today" and "tonight". */
  todayWords: readonly string[]
  /** Words meaning the next day. */
  tomorrowWords: readonly string[]
  /**
   * Words that may stand before a weekday without moving it, such as "next".
   *
   * A bare weekday is always the one still to come, so there is no earlier Saturday for
   * "next saturday" to distinguish itself from. Giving the word a week of its own would
   * make two sentences a Member reads as the same differ by seven days. Anybody who
   * wants the later one writes "saturday in 2 weeks".
   */
  comingWords: readonly string[]
  /** Words introducing a distance, such as the "in" of "in 2 weeks". */
  inWords: readonly string[]
  /** What a day is called, singular and plural. */
  dayWords: readonly string[]
  /** What a week is called, singular and plural. */
  weekWords: readonly string[]
  /** Units a quantity may carry. A number needs none. */
  units: readonly string[]
}

/** What the parser needs besides the text. */
export interface AddRowParseOptions {
  /** The day relative words resolve against. */
  from: Date
  /** The date-fns locale, which supplies weekday and month names. */
  locale: DateLocale
  vocabulary: AddRowVocabulary
  /**
   * Kinds already lifted out of the sentence.
   *
   * The row parses what is left after earlier words became chips, so it has to be told
   * what those were: at most one quantity and one date per Item.
   */
  taken?: readonly ChipKind[]
}

/** The Item a line of text describes. */
export interface AddRowParse {
  /** What is left of the sentence, which is the Item's name. */
  name: string
  /** Normalised as `value unit`, or empty. */
  quantity: string
  /** The stored due date, or empty. */
  due: DueDate
  /** The chips to draw, in the order they were recognised. */
  chips: AddRowChip[]
}

/**
 * parseAddRow reads an Item out of a line of text.
 *
 * Only the trailing words are candidates, and the scan stops at the first word it does
 * not recognise. That is what keeps `Call 2 plumbers` a name: `plumbers` ends the scan
 * before the `2` is ever looked at.
 */
export function parseAddRow(text: string, options: AddRowParseOptions): AddRowParse {
  const words = text.split(/\s+/).filter(Boolean)
  const found = new Map<ChipKind, AddRowChip>()
  const taken = new Set<ChipKind>(options.taken ?? [])

  let end = words.length
  while (end > 1) {
    const value = takeTrailingValue(words, end, taken, options)
    if (!value) {
      break
    }
    taken.add(value.chip.kind)
    found.set(value.chip.kind, value.chip)
    end -= value.words
  }

  const quantity = found.get("quantity")
  const due = found.get("due")
  return {
    name: words.slice(0, end).join(" "),
    quantity: quantity?.label ?? "",
    due: due?.label ?? "",
    // Recognised last is written first: the chips read in the order they were typed.
    chips: [...found.values()].reverse(),
  }
}

/**
 * The most words one value can be written in: "saturday in 2 weeks".
 *
 * The window is what keeps the scan honest. Every word it widens by is another way for
 * a name to be mistaken for a date, so it is the length of the longest form the grammar
 * actually has rather than a number with room to spare.
 */
const LONGEST_VALUE = 4

/** One recognised value, and how many words it was written in. */
interface TakenValue {
  chip: AddRowChip
  words: number
}

/**
 * takeTrailingValue recognises the value at the end of the words, longest form first so
 * `250 g` is one quantity rather than a stray `g`, and `saturday in 2 weeks` is one date
 * rather than a Saturday with some words after it.
 *
 * Never the whole line: a value needs a name in front of it to belong to.
 */
function takeTrailingValue(
  words: string[],
  end: number,
  taken: ReadonlySet<ChipKind>,
  options: AddRowParseOptions,
): TakenValue | null {
  for (let length = Math.min(LONGEST_VALUE, end - 1); length >= 1; length -= 1) {
    const source = words.slice(end - length, end).join(" ")
    if (isProtected(source)) {
      continue
    }
    const chip = recognise(source, options)
    if (chip && !taken.has(chip.kind)) {
      return { chip, words: length }
    }
  }
  return null
}

/**
 * isProtected reports whether the Member has asked for the text to be left alone.
 *
 * Quoting is the escape hatch the design gives for a name that reads like a value:
 * `Buy "2kg" bag` keeps its words.
 */
function isProtected(source: string): boolean {
  return /["'\\]/.test(source)
}

/** recognise reads one candidate as a quantity or a date, or neither. */
function recognise(source: string, options: AddRowParseOptions): AddRowChip | null {
  const quantity = readQuantity(source, options.vocabulary.units)
  if (quantity) {
    return { kind: "quantity", label: quantity, source }
  }
  const due = readDue(source, options)
  if (due) {
    return { kind: "due", label: due, source }
  }
  return null
}

/**
 * readQuantity normalises a number and its unit to `value unit`, so `1kg` and `1 kg`
 * are stored the same way.
 */
function readQuantity(source: string, units: readonly string[]): string {
  const match = /^(\d+(?:[.,]\d+)?)\s?(\p{L}+)?$/u.exec(source)
  if (!match) {
    return ""
  }
  const [, value, unit] = match
  if (!unit) {
    return value ?? ""
  }
  return units.includes(fold(unit)) ? `${value} ${fold(unit)}` : ""
}

/** readDue resolves a date word, a weekday, or a written date, to a stored date. */
function readDue(source: string, options: AddRowParseOptions): DueDate {
  const { from, locale, vocabulary } = options
  const folded = fold(source)

  if (vocabulary.todayWords.some((word) => fold(word) === folded)) {
    return toStored(shift(from, 0))
  }
  if (vocabulary.tomorrowWords.some((word) => fold(word) === folded)) {
    return toStored(shift(from, 1))
  }

  const weekdays = weekdayNames(locale)
  const weekday = weekdays.get(folded)
  if (weekday !== undefined) {
    return toStored(nextWeekday(from, weekday))
  }

  const relative = readRelative(folded, weekdays, options)
  if (relative) {
    return toStored(relative)
  }

  const written = readWrittenDate(folded, options)
  return written ? toStored(written) : ""
}

/**
 * readRelative reads the forms that say how far away a day is rather than naming it.
 *
 * Three, and they are the ones people type: `next saturday`, `in 3 days`, and
 * `saturday in 2 weeks`.
 */
function readRelative(
  folded: string,
  weekdays: ReadonlyMap<string, Weekday>,
  options: AddRowParseOptions,
): Date | null {
  const { from, vocabulary } = options
  const words = folded.split(" ")

  if (words.length === 2 && says(vocabulary.comingWords, words[0])) {
    const weekday = weekdays.get(words[1] ?? "")
    return weekday === undefined ? null : nextWeekday(from, weekday)
  }

  const distance = readDistance(words, vocabulary)
  if (!distance) {
    return null
  }

  // "in 2 weeks", counted from today.
  if (distance.before === 0) {
    return shift(from, distance.days)
  }

  // "saturday in 2 weeks", counted from the Saturday that was coming anyway.
  if (distance.before === 1) {
    const weekday = weekdays.get(words[0] ?? "")
    return weekday === undefined ? null : shift(nextWeekday(from, weekday), distance.days)
  }
  return null
}

/** How far away a day was said to be, and how many words came before the saying. */
interface Distance {
  /** Words before the `in`. None for "in 2 weeks", one for "saturday in 2 weeks". */
  before: number
  days: number
}

/**
 * readDistance reads a trailing `in <number> <days|weeks>`.
 *
 * Whole numbers of at least one only. "in 0 days" is today said the hard way, and a
 * fractional week is not a day anybody means.
 */
function readDistance(words: string[], vocabulary: AddRowVocabulary): Distance | null {
  const before = words.length - 3
  if (before < 0 || !says(vocabulary.inWords, words[before])) {
    return null
  }

  const count = Number(words[before + 1])
  if (!Number.isInteger(count) || count < 1) {
    return null
  }

  const unit = words[before + 2]
  if (says(vocabulary.dayWords, unit)) {
    return { before, days: count }
  }
  if (says(vocabulary.weekWords, unit)) {
    return { before, days: count * 7 }
  }
  return null
}

/** says reports whether a word is one a language uses for something. */
function says(words: readonly string[], word: string | undefined): boolean {
  return word !== undefined && words.some((candidate) => fold(candidate) === fold(word))
}

/**
 * readWrittenDate reads the two forms the design accepts: `30/8` and `30 aug`, both
 * day first.
 */
function readWrittenDate(folded: string, options: AddRowParseOptions): Date | null {
  const numeric = /^(\d{1,2})[/.](\d{1,2})$/.exec(folded)
  if (numeric) {
    const day = Number(numeric[1])
    const month = Number(numeric[2])
    return month >= 1 && month <= 12 ? nextDayOfMonth(options.from, day, month - 1) : null
  }

  const named = /^(\d{1,2})\s+(\p{L}+)$/u.exec(folded)
  if (!named) {
    return null
  }
  const month = monthNames(options.locale).get(named[2] ?? "")
  return month === undefined ? null : nextDayOfMonth(options.from, Number(named[1]), month)
}
