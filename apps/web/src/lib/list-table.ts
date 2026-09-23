import { ListOrder, ListStatus as StatusWire, type ListListsResponse } from "@nooks/api"

/** Which Lists the table is showing. */
export type ListStatus = "all" | "active" | "completed" | "archived" | "deleted"

export const LIST_STATUSES: ListStatus[] = ["all", "active", "completed", "archived", "deleted"]

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

/** PAGE_SIZE is how many rows a page holds. */
export const PAGE_SIZE = 25

/**
 * How All lists is being read.
 *
 * In the address rather than in the screen, so the sidebar can point at the finished
 * Lists, a refresh keeps the view, and Back undoes a filter rather than leaving.
 */
export interface ListsSearch {
  status?: ListStatus
  sort?: ListSort
  /** 1-based. Absent is the first page, so the commonest address stays clean. */
  page?: number
}

/*
The words in the address, and the enums on the wire.

Two vocabularies because they answer to different people: one is typed by whoever is
reading a URL, the other is generated. Mapping them in one table here is what keeps the
screen from knowing about either.
*/
const STATUS_ON_THE_WIRE: Record<ListStatus, StatusWire> = {
  all: StatusWire.UNSPECIFIED,
  active: StatusWire.ACTIVE,
  completed: StatusWire.COMPLETED,
  archived: StatusWire.ARCHIVED,
  deleted: StatusWire.DELETED,
}

const SORT_ON_THE_WIRE: Record<ListSort, ListOrder> = {
  updated: ListOrder.UPDATED,
  name: ListOrder.NAME,
  open: ListOrder.OPEN,
}

export function statusOnTheWire(status: ListStatus): StatusWire {
  return STATUS_ON_THE_WIRE[status]
}

export function sortOnTheWire(sort: ListSort): ListOrder {
  return SORT_ON_THE_WIRE[sort]
}

/** Where a page sits in the set, for "1–25 of 60" and for the two arrows. */
export interface PageBounds {
  /** Whether there is a page after this one. */
  more: boolean
  /** The 1-based range this page covers. Both zero when it holds nothing. */
  from: number
  to: number
}

/**
 * boundsOf says where a page sits, read from the answer itself.
 *
 * From the answer rather than from what was asked for: the server has the last word on
 * how big a page is, and a range worked out from a size it did not agree to would count
 * rows that are not there.
 *
 * "Is there another page" rather than "how many pages", because past a point the server
 * stops counting and answers "at least a thousand". A page that came back short is the
 * last one whatever the total says.
 */
export function boundsOf(answer: ListListsResponse): PageBounds {
  const size = answer.pageSize || PAGE_SIZE
  const shown = answer.lists.length
  const start = (Math.max(answer.page, 1) - 1) * size

  return {
    more: shown === size && (answer.atLeast || start + shown < answer.total),
    from: shown === 0 ? 0 : start + 1,
    to: start + shown,
  }
}
