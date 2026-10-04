import type { QueryClient } from "@tanstack/react-query"

import { refreshLists } from "./refresh"

/**
 * afterDeletingList takes somebody off a List that no longer exists, then re-reads what
 * the deletion changed.
 *
 * In that order. Re-reading first meant re-reading the List still on screen, which the
 * Instance answers with not found, and the page waited for that answer before it left.
 * Leaving first unmounts the List, so nothing is still asking for it, and dropping it
 * from the cache means the browser's saved copy cannot bring it back either.
 */
export async function afterDeletingList(
  queryClient: QueryClient,
  listUid: string,
  leave: () => Promise<void>,
): Promise<void> {
  await leave()
  queryClient.removeQueries({ queryKey: ["list", listUid], exact: true })
  await refreshLists(queryClient)
}
