import { queryOptions } from "@tanstack/react-query"

import { listClient, memberClient } from "./api"

/** Everyone on the Instance — who a List can be shared with. */
export const membersQuery = queryOptions({
  queryKey: ["members"],
  queryFn: () => memberClient.listMembers({}),
})

/** Every Group, with who is in it. */
export const groupsQuery = queryOptions({
  queryKey: ["groups"],
  queryFn: () => memberClient.listGroups({}),
})

/**
 * Who one List reaches by name.
 *
 * Only its owner may read this, so it is fetched when the share dialog opens rather
 * than alongside the List.
 */
export function listSharesQuery(listUid: string, enabled: boolean) {
  return queryOptions({
    queryKey: ["list-shares", listUid],
    queryFn: () => listClient.getListShares({ listUid }),
    enabled,
  })
}
