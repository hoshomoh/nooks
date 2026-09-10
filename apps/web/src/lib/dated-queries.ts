import { queryOptions } from "@tanstack/react-query"

import { listClient } from "./api"
import { rangeFrom, toStored, type DueDate } from "./dates"

/** How far ahead Upcoming looks. The design says the next two weeks. */
export const UPCOMING_DAYS = 14

/**
 * Everything overdue or due today.
 *
 * No lower bound: an Item that was due on Friday is still what a Member needs to see on
 * Tuesday, and burying it under a date range would be the app hiding its own backlog.
 */
export function todayQuery(from: Date) {
  const to = toStored(from)
  return queryOptions({
    queryKey: ["dated", "", to],
    queryFn: () => listClient.listDatedItems({ from: "", to }),
  })
}

/** Everything due in the next two weeks, starting tomorrow. */
export function upcomingQuery(from: Date) {
  const { start, end } = rangeFrom(from, UPCOMING_DAYS)
  return queryOptions({
    queryKey: ["dated", start, end],
    queryFn: () => listClient.listDatedItems({ from: start, to: end }),
  })
}

/** groupByDay buckets dated Items by their due date, keeping the order they arrived in. */
export function groupByDay<T extends { item?: { dueOn: string } }>(
  items: readonly T[],
): Array<{ day: DueDate; items: T[] }> {
  const order: DueDate[] = []
  const byDay = new Map<DueDate, T[]>()

  for (const entry of items) {
    const day = entry.item?.dueOn ?? ""
    if (!byDay.has(day)) {
      byDay.set(day, [])
      order.push(day)
    }
    byDay.get(day)?.push(entry)
  }
  return order.map((day) => ({ day, items: byDay.get(day) ?? [] }))
}
