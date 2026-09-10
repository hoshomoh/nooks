import {
  addDays,
  differenceInCalendarDays,
  format,
  isValid,
  parseISO,
  startOfDay,
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

/** rangeFrom returns the stored dates bounding a window of days starting at from. */
export function rangeFrom(from: Date, days: number): { start: DueDate; end: DueDate } {
  return { start: toStored(from), end: toStored(addDays(from, days)) }
}
