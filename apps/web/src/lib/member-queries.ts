import { queryOptions } from "@tanstack/react-query"

import { requestClient } from "./api"

/**
 * Who is waiting to join.
 *
 * Read on the Members page, where the design puts them: a request is somebody who is
 * not a Member yet, and deciding on them belongs beside the people who already are.
 */
export const pendingRequestsQuery = queryOptions({
  queryKey: ["pending-requests"],
  queryFn: () => requestClient.listPendingRequests({}),
})
