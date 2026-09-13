import {
  addDays,
  addYears,
  differenceInCalendarDays,
  eachDayOfInterval,
  endOfDay,
  endOfMonth,
  endOfWeek,
  format,
  formatISO,
  getDay,
  isBefore,
  isSameMonth,
  isValid,
  parseISO,
  startOfDay,
  startOfMonth,
  startOfWeek,
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

/** What a moment needs in order to read in a Member's own language. */
export interface MomentLabelOptions {
  /** The moment to measure against, usually now. */
  from: Date
  /** The date-fns locale, so weekday and month names match the language. */
  locale: DateLocale
  /** The word for the day before this one. */
  yesterdayWord: string
}

/**
 * momentLabel is when something happened, at the resolution that is still useful.
 *
 * Today it is a clock time, because "18:44" is how a Member remembers this morning.
 * Further back the clock stops mattering and the day is what is left.
 */
export function momentLabel(at: Date, options: MomentLabelOptions): string {
  const days = differenceInCalendarDays(startOfDay(options.from), startOfDay(at))
  if (days <= 0) {
    return format(at, "HH:mm", { locale: options.locale })
  }
  if (days === 1) {
    return options.yesterdayWord
  }
  if (days < 7) {
    return format(at, "EEEE", { locale: options.locale })
  }
  return format(at, "d MMM", { locale: options.locale })
}

/**
 * SETTLE_MS is how long a tick somebody else made is held before it settles.
 *
 * DESIGN.md §6 says a second. A little longer here, so a tick that took a moment to
 * arrive is still shown as having just landed.
 */
export const SETTLE_MS = 3000

/** happenedToday reports whether a moment falls on the day being measured against. */
export function happenedToday(at: string, from: Date): boolean {
  const parsed = parseMoment(at)
  return parsed !== null && differenceInCalendarDays(startOfDay(from), startOfDay(parsed)) === 0
}

/** justHappened reports whether a moment is recent enough to still be news. */
export function justHappened(at: string, now: Date): boolean {
  const parsed = parseMoment(at)
  return parsed !== null && now.getTime() - parsed.getTime() < SETTLE_MS
}

/** parseMoment reads an RFC 3339 timestamp, or null when it is not one. */
export function parseMoment(value: string): Date | null {
  if (!value) {
    return null
  }
  const parsed = parseISO(value)
  return isValid(parsed) ? parsed : null
}

/**
 * momentIn is an RFC 3339 timestamp a number of days from now.
 *
 * For an expiry, which is a moment rather than a day: a token stops working at a time,
 * and the server is told when in the format it reads back.
 */
export function momentIn(days: number, from: Date = new Date()): string {
  return formatISO(addDays(from, days))
}

/**
 * atEndOf is the last moment of a chosen day, as an RFC 3339 timestamp.
 *
 * A Member picking a day for an expiry means the end of that day, not midnight at the
 * start of it: a token that stops working the moment the day begins would have expired
 * the day before, as far as anybody using it is concerned.
 */
export function atEndOf(day: DueDate): string {
  const parsed = parseDue(day)
  return parsed ? formatISO(endOfDay(parsed)) : ""
}

/** rangeFrom returns the stored dates bounding a window of days starting at from. */
export function rangeFrom(from: Date, days: number): DueRange {
  return { start: toStored(from), end: toStored(addDays(from, days)) }
}

/**
 * WEEK_STARTS_ON_MONDAY is the only week Nooks draws.
 *
 * DESIGN.md §8: a week does not begin on Sunday. date-fns defaults to Sunday, so every
 * call that cares has to say otherwise — which is why they all live in this file.
 */
const WEEK_STARTS_ON_MONDAY = { weekStartsOn: 1 } as const

/** The first and last day a month's grid covers: whole weeks, Monday first. */
export interface DaySpan {
  start: Date
  end: Date
}

/**
 * monthGrid is the span a month's grid covers.
 *
 * Wider than the month: the grid is whole weeks, so it reaches back into the previous
 * month and forward into the next to fill the first and last rows.
 */
export function monthGrid(month: Date): DaySpan {
  const first = startOfMonth(month)
  return {
    start: startOfWeek(first, WEEK_STARTS_ON_MONDAY),
    end: endOfWeek(endOfMonth(first), WEEK_STARTS_ON_MONDAY),
  }
}

/** daysIn is every day the span covers, in order. */
export function daysIn(span: DaySpan): Date[] {
  return eachDayOfInterval(span)
}

/** firstOfMonth is the month a day belongs to, as its first day. */
export function firstOfMonth(date: Date): Date {
  return startOfMonth(date)
}

/** sameMonth reports whether a day belongs to the month being drawn. */
export function sameMonth(date: Date, month: Date): boolean {
  return isSameMonth(date, month)
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
 * The spellings already worked out for a language.
 *
 * The add row parses on every keystroke and asks for these several times a line, and
 * answering means formatting twenty-six dates. A date-fns locale is a module object
 * that never changes, so each answer is worked out once and kept against it.
 */
const weekdaysByLocale = new WeakMap<DateLocale, Map<string, Weekday>>()
const monthsByLocale = new WeakMap<DateLocale, Map<string, Month>>()

/**
 * weekdayNames maps every spelling of a weekday in a language to its number.
 *
 * Taken from the date-fns locale rather than a table in each locale file: a language
 * already names its own days, and asking translators to repeat them is how the two
 * drift apart.
 */
export function weekdayNames(locale: DateLocale): ReadonlyMap<string, Weekday> {
  const known = weekdaysByLocale.get(locale)
  if (known) {
    return known
  }

  const names = new Map<string, Weekday>()
  for (let weekday = 0; weekday < 7; weekday += 1) {
    const day = new Date(2024, 0, 7 + weekday)
    addSpellings(names, weekday, [
      format(day, "EEEE", { locale }),
      format(day, "EEE", { locale }),
    ])
  }
  weekdaysByLocale.set(locale, names)
  return names
}

/** monthNames maps every spelling of a month in a language to its number. */
export function monthNames(locale: DateLocale): ReadonlyMap<string, Month> {
  const known = monthsByLocale.get(locale)
  if (known) {
    return known
  }

  const names = new Map<string, Month>()
  for (let month = 0; month < 12; month += 1) {
    const day = new Date(2024, month, 1)
    addSpellings(names, month, [
      format(day, "MMMM", { locale }),
      format(day, "MMM", { locale }),
    ])
  }
  monthsByLocale.set(locale, names)
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
