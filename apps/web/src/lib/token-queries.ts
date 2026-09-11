import { queryOptions } from "@tanstack/react-query"

import { tokenClient } from "./api"

/**
 * The signed-in Member's own Access tokens.
 *
 * Never anybody else's: the server answers with what the caller may see, and a Member
 * who is also an Admin sees other people's tokens listed as existing, not readable.
 */
export const tokensQuery = queryOptions({
  queryKey: ["access-tokens"],
  queryFn: () => tokenClient.listAccessTokens({}),
})
