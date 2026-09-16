import { keepPreviousData, queryOptions } from "@tanstack/react-query"

import { listClient } from "./api"
import {
  DEFAULT_SORT,
  PAGE_SIZE,
  sortOnTheWire,
  statusOnTheWire,
  type ListsSearch,
} from "./list-table"

/**
 * The four groups the sidebar draws, each capped and each saying how many there are.
 *
 * Not a page of every List filtered here. A Member with a great many of them would
 * otherwise have all of them sent to draw a column that shows twenty.
 */
export const sidebarQuery = queryOptions({
  queryKey: ["sidebar"],
  queryFn: () => listClient.getSidebar({}),
})

/**
 * One page of All lists, filtered and ordered by the server.
 *
 * keepPreviousData because paging should move the table, not empty it: without it the
 * rows vanish for as long as the next page takes and the page jumps to the top.
 */
export function listPageQuery(search: ListsSearch) {
  const { status = "all", sort = DEFAULT_SORT, page = 1 } = search

  return queryOptions({
    queryKey: ["lists", status, sort, page],
    queryFn: () =>
      listClient.listLists({
        status: statusOnTheWire(status),
        order: sortOnTheWire(sort),
        page,
        pageSize: PAGE_SIZE,
      }),
    placeholderData: keepPreviousData,
  })
}

/** One List and its Items. */
export function listQuery(listUid: string) {
  return queryOptions({
    queryKey: ["list", listUid],
    queryFn: () => listClient.getList({ listUid }),
  })
}

/**
 * Search results for what a Member typed. Empty queries are not sent.
 *
 * Every keystroke is a different query, so without keepPreviousData the panel would
 * empty and refill on each letter — the results flickering while the Member is still
 * deciding what to look for.
 */
export function searchQuery(query: string) {
  return queryOptions({
    queryKey: ["search", query],
    queryFn: () => listClient.search({ query }),
    enabled: query.trim().length > 0,
    placeholderData: keepPreviousData,
  })
}
