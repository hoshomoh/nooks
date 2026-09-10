/**
 * How a due date reads in a row, per the design: "Fri" this week, "Mon 1 Sep" beyond
 * it. Nothing is shown as a full date when a weekday would do.
 *
 * Pure, and takes today as an argument, so it can be tested without waiting for
 * Thursday.
 */

/** A day, as stored: YYYY-MM-DD. */
export type DueDate = string

const DAY_MS = 24 * 60 * 60 * 1000

/** parseDue turns a stored date into a Date at UTC midnight, or null. */
export function parseDue(due: DueDate): Date | null {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(due)) {
    return null
  }
  const parsed = new Date(`${due}T00:00:00Z`)
  return Number.isNaN(parsed.getTime()) ? null : parsed
}

/** daysUntil is whole days from today to the due date; negative when overdue. */
export function daysUntil(due: DueDate, today: Date): number | null {
  const parsed = parseDue(due)
  if (!parsed) {
    return null
  }
  const midnight = Date.UTC(today.getUTCFullYear(), today.getUTCMonth(), today.getUTCDate())
  return Math.round((parsed.getTime() - midnight) / DAY_MS)
}

/** isOverdue reports whether a due date has passed. Today is not overdue. */
export function isOverdue(due: DueDate, today: Date): boolean {
  const days = daysUntil(due, today)
  return days !== null && days < 0
}

/**
 * dueLabel is what the row shows: nothing for an undated Item, a weekday within the
 * coming week, and a day and month beyond it.
 */
export function dueLabel(due: DueDate, today: Date): string {
  const parsed = parseDue(due)
  const days = daysUntil(due, today)
  if (!parsed || days === null) {
    return ""
  }

  if (days === 0) {
    return "Today"
  }
  if (days === 1) {
    return "Tomorrow"
  }
  const weekday = WEEKDAYS[parsed.getUTCDay()]
  if (days > 1 && days < 7) {
    return weekday
  }
  return `${weekday} ${parsed.getUTCDate()} ${MONTHS[parsed.getUTCMonth()]}`
}

/**
 * The names are written out rather than taken from Intl.
 *
 * `toLocaleDateString("en-GB", { month: "short" })` renders September as "Sept", and
 * the design says "Sep". More to the point, ICU data changes between runtime versions,
 * so borrowing it would make the app's appearance depend on which Node built it.
 */
const WEEKDAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"]
const MONTHS = [
  "Jan", "Feb", "Mar", "Apr", "May", "Jun",
  "Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
]
