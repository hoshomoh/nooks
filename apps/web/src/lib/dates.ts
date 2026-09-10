import {
  addDays,
  addYears,
  differenceInCalendarDays,
  endOfMonth,
  format,
  getDay,
  isBefore,
  isValid,
  parseISO,
  startOfDay,
  startOfMonth,
} from "date-fns"
import type { Locale as DateLocale } from "date-fns"

/**
 * Every date operation in Nooks. Nothing else parses, formats or compares a date.
 *
 * date-fns rather than Intl or hand-rolled arithmetic: its month names are stable
 * across runtimes, where `Intl.DateTimeFormat("en-GB")` renders September as "Sept" and
 * its data changes with the runtime's ICU version. The design says "Sep".
 *
 * Every function takes the day it should measure against, so none of them read the
 * clock and all of them are testable without waiting for Thursday.
 */

/** A due date as stored: a day, not a time. */
export type DueDate = string

/** STORED_DATE is the format an Item's due date is written in. */
export const STORED_DATE = "yyyy-MM-dd"

/** parseDue turns a stored date into a Date, or null when it is not one. */
export function parseDue(due: DueDate): Date | null {
  if (!due) {
    return null
  }
  const parsed = parseISO(due)
  return isValid(parsed) ? parsed : null
}

/** toStored writes a Date back in the stored format. */
export function toStored(date: Date): DueDate {
  return format(date, STORED_DATE)
}

/** today is the current day with the time discarded. */
export function today(now: Date = new Date()): Date {
  return startOfDay(now)
}

/** daysUntil is whole calendar days from today to the due date; negative when overdue. */
export function daysUntil(due: DueDate, from: Date): number | null {
  const parsed = parseDue(due)
  return parsed ? differenceInCalendarDays(parsed, from) : null
}

/** isOverdue reports whether a due date has passed. Something due today is not late. */
export function isOverdue(due: DueDate, from: Date): boolean {
  const days = daysUntil(due, from)
  return days !== null && days < 0
}

/** isDueToday reports whether a due date is today. */
export function isDueToday(due: DueDate, from: Date): boolean {
  return daysUntil(due, from) === 0
}

/**
 * What a date needs in order to read in a Member's own language.
 *
 * The near-day words are passed in rather than looked up here, so this file stays pure
 * and knows nothing about i18next: the caller translates, this decides which words to
 * use and how the rest is spelled.
 */
export type DateLabelOptions = {
  /** The day to measure against. */
  from: Date
  /** The date-fns locale, so weekday and month names match the language. */
  locale: DateLocale
  /** The word for the current day. */
  todayWord: string
  /** The word for the next day. */
  tomorrowWord: string
}

/**
 * dueLabel is what a row shows: nothing for an undated Item, a name for the near days,
 * a weekday within the coming week, and a day and month beyond it.
 */
export function dueLabel(due: DueDate, options: DateLabelOptions): string {
  const parsed = parseDue(due)
  const days = daysUntil(due, options.from)
  if (!parsed || days === null) {
    return ""
  }

  if (days === 0) {
    return options.todayWord
  }
  if (days === 1) {
    return options.tomorrowWord
  }
  if (days > 1 && days < 7) {
    return format(parsed, "EEE", { locale: options.locale })
  }
  return format(parsed, "EEE d MMM", { locale: options.locale })
}

/** dayHeading is how a day reads above a group of Items, e.g. "Friday". */
export function dayHeading(due: DueDate, options: DateLabelOptions): string {
  const parsed = parseDue(due)
  const days = daysUntil(due, options.from)
  if (!parsed || days === null) {
    return ""
  }
  if (days === 0) {
    return options.todayWord
  }
  if (days === 1) {
    return options.tomorrowWord
  }
  return format(parsed, "EEEE", { locale: options.locale })
}

/** A window of days, as the two stored dates that bound it. */
export interface DueRange {
  start: DueDate
  end: DueDate
}

/** rangeFrom returns the stored dates bounding a window of days starting at from. */
export function rangeFrom(from: Date, days: number): DueRange {
  return { start: toStored(from), end: toStored(addDays(from, days)) }
}

/** monthWindow is the stored dates bounding the month a day falls in. */
export function monthWindow(from: Date): DueRange {
  return { start: toStored(startOfMonth(from)), end: toStored(endOfMonth(from)) }
}

/** monthHeading is a month's name in a language, e.g. "August". */
export function monthHeading(month: Date, locale: DateLocale): string {
  return format(month, "LLLL", { locale })
}

/** dayFullHeading is a day written out, e.g. "Tuesday, 25 August". */
export function dayFullHeading(day: Date, locale: DateLocale): string {
  return format(day, "EEEE, d MMMM", { locale })
}

/** shift returns the day a whole number of days away from another. */
export function shift(from: Date, days: number): Date {
  return addDays(startOfDay(from), days)
}

/**
 * nextWeekday returns the coming occurrence of a weekday, today excluded.
 *
 * Typing "sat" on a Saturday means the Saturday after this one: the day being named is
 * the one still to come, otherwise it would have been called "today".
 */
export function nextWeekday(from: Date, weekday: Weekday): Date {
  const start = startOfDay(from)
  const ahead = (weekday - getDay(start) + 7) % 7
  return addDays(start, ahead === 0 ? 7 : ahead)
}

/** Weekday is a day of the week as date-fns numbers them: 0 is Sunday. */
export type Weekday = number

/**
 * nextDayOfMonth returns the coming occurrence of a day and month, rolling into next
 * year when the date has already passed. Returns null for a day that month never has.
 */
export function nextDayOfMonth(from: Date, day: number, month: Month): Date | null {
  const start = startOfDay(from)
  const candidate = startOfDay(new Date(start.getFullYear(), month, day))
  if (candidate.getDate() !== day || candidate.getMonth() !== month) {
    return null
  }
  return isBefore(candidate, start) ? addYears(candidate, 1) : candidate
}

/** Month is a month as date-fns numbers them: 0 is January. */
export type Month = number

/**
 * weekdayNames maps every spelling of a weekday in a language to its number.
 *
 * Taken from the date-fns locale rather than a table in each locale file: a language
 * already names its own days, and asking translators to repeat them is how the two
 * drift apart.
 */
export function weekdayNames(locale: DateLocale): Map<string, Weekday> {
  const names = new Map<string, Weekday>()
  for (let weekday = 0; weekday < 7; weekday += 1) {
    const day = new Date(2024, 0, 7 + weekday)
    addSpellings(names, weekday, [
      format(day, "EEEE", { locale }),
      format(day, "EEE", { locale }),
    ])
  }
  return names
}

/** monthNames maps every spelling of a month in a language to its number. */
export function monthNames(locale: DateLocale): Map<string, Month> {
  const names = new Map<string, Month>()
  for (let month = 0; month < 12; month += 1) {
    const day = new Date(2024, month, 1)
    addSpellings(names, month, [
      format(day, "MMMM", { locale }),
      format(day, "MMM", { locale }),
    ])
  }
  return names
}

/**
 * addSpellings records each way a name is written, folded for comparison.
 *
 * The first spelling wins a collision, so a full name is never shadowed by another
 * month's abbreviation.
 */
function addSpellings(names: Map<string, number>, value: number, spellings: string[]): void {
  for (const spelling of spellings) {
    const folded = fold(spelling)
    if (folded && !names.has(folded)) {
      names.set(folded, value)
    }
  }
}

/**
 * fold reduces a word to what it is compared by: lower case, without the trailing stop
 * some languages put on an abbreviation.
 */
export function fold(word: string): string {
  return word.toLocaleLowerCase().replace(/\.$/, "")
}
