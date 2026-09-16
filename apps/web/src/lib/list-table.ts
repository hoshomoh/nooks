import type { List } from "@nooks/api"

import { isArchived, isCompleted } from "./list-groups"

/** Which Lists the table is showing. */
export type ListStatus = "all" | "active" | "completed" | "archived"

export const LIST_STATUSES: ListStatus[] = ["all", "active", "completed", "archived"]

/**
 * What the table is ordered by.
 *
 * Three, not five. "Oldest first" is "recently created" read upwards, and a household
 * with a dozen Lists does not need a fourth way to arrange them.
 */
export type ListSort = "updated" | "name" | "open"

export const LIST_SORTS: ListSort[] = ["updated", "name", "open"]

/** DEFAULT_SORT is what the table opens on: what changed last is what you came for. */
export const DEFAULT_SORT: ListSort = "updated"

/**
 * filterLists keeps the Lists the chosen status covers.
 *
 * Archived is its own answer rather than a flavour of the others: a List put away is
 * out of All, Active and Completed alike, because the point of archiving is that it
 * stops appearing where somebody is looking for what they are working on.
 */
export function filterLists(lists: readonly List[], status: ListStatus): List[] {
  if (status === "archived") {
    return lists.filter(isArchived)
  }

  const here = lists.filter((list) => !isArchived(list))
  switch (status) {
    case "active":
      return here.filter((list) => !isCompleted(list))
    case "completed":
      return here.filter(isCompleted)
    case "all":
      return here
  }
}

/**
 * sortLists orders a page of Lists.
 *
 * By name uses localeCompare rather than comparing the strings directly, so "avocados"
 * is not filed under the capitals and "Éclairs" lands where a reader expects. The
 * server cannot do the same without knowing whose language to sort in; the browser
 * knows, so this is the one place it can be right.
 */
export function sortLists(lists: readonly List[], sort: ListSort, locale?: string): List[] {
  const out = [...lists]
  switch (sort) {
    case "updated":
      return out.sort((a, b) => b.updatedAt.localeCompare(a.updatedAt))
    case "name":
      return out.sort((a, b) =>
        a.name.localeCompare(b.name, locale, { sensitivity: "base", numeric: true }),
      )
    case "open":
      // Most open first, and alphabetical between Lists with the same number, so the
      // order does not shuffle every time something is ticked.
      return out.sort(
        (a, b) =>
          b.openCount - a.openCount ||
          a.name.localeCompare(b.name, locale, { sensitivity: "base", numeric: true }),
      )
  }
}
