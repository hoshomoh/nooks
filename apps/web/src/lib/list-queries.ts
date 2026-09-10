import { keepPreviousData, queryOptions } from "@tanstack/react-query"

import { listClient } from "./api"

/** Every List the signed-in Member can reach — what the sidebar renders. */
export const listsQuery = queryOptions({
  queryKey: ["lists"],
  queryFn: () => listClient.listLists({}),
})

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
