import { queryOptions } from "@tanstack/react-query"

import { activityClient } from "./api"

/**
 * What is waiting for the signed-in Member.
 *
 * Nooks has no mail server, so this is the only place a join request, a reset request
 * or a share surfaces. It is read on every screen, because the dot beside Activity has
 * to be right wherever the Member happens to be.
 */
export const activityQuery = queryOptions({
  queryKey: ["activity"],
  queryFn: () => activityClient.listActivity({}),
})
