import {
  eachDayOfInterval,
  endOfMonth,
  endOfWeek,
  isSameMonth,
  startOfMonth,
  startOfWeek,
} from "date-fns"

import { toStored, type DueDate } from "./dates"

/** One cell in the month grid. */
export type CalendarDay = {
  /** The day, in stored form, so it can be matched against an Item's due date. */
  date: DueDate
  /** The number to print in the corner. */
  dayOfMonth: number
  /** False for the days either side that fill out the first and last weeks. */
  inMonth: boolean
}

/** One row of the grid. */
export type CalendarWeek = {
  /** The Monday of the week, for a stable key. */
  key: DueDate
  days: CalendarDay[]
}

/** The grid a month is drawn on. */
export type CalendarMonth = {
  /** The first of the month, for headings. */
  month: Date
  weeks: CalendarWeek[]
}

/**
 * buildMonth lays out the weeks a month is drawn on, padded to whole weeks.
 *
 * Weeks start on Monday: the design's calendar is headed Mon–Sun, and a household's
 * week does not begin on Sunday. Pure, so the awkward months — one starting on a
 * Sunday, one ending on a Monday — can be tested directly.
 */
export function buildMonth(month: Date): CalendarMonth {
  const first = startOfMonth(month)
  const start = startOfWeek(first, { weekStartsOn: 1 })
  const end = endOfWeek(endOfMonth(first), { weekStartsOn: 1 })

  const days = eachDayOfInterval({ start, end }).map((date) => ({
    date: toStored(date),
    dayOfMonth: date.getDate(),
    inMonth: isSameMonth(date, first),
  }))

  const weeks: CalendarWeek[] = []
  for (let i = 0; i < days.length; i += 7) {
    const week = days.slice(i, i + 7)
    weeks.push({ key: week[0].date, days: week })
  }
  return { month: first, weeks }
}

/** byDay buckets anything with a due date, for looking up a cell's contents. */
export function byDay<T>(entries: readonly T[], dueOf: (entry: T) => string): Map<DueDate, T[]> {
  const map = new Map<DueDate, T[]>()
  for (const entry of entries) {
    const due = dueOf(entry)
    if (!due) {
      continue
    }
    const existing = map.get(due)
    if (existing) {
      existing.push(entry)
    } else {
      map.set(due, [entry])
    }
  }
  return map
}
