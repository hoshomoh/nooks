import { queryOptions } from "@tanstack/react-query"

import { publicClient } from "./api"

/** How often the page re-reads itself, in milliseconds. */
export const PUBLIC_REFRESH_MS = 30_000

/**
 * The one List an Instance has published.
 *
 * It refetches on its own, because the page is meant to stay open on the way to the
 * shop: somebody at home adding bread to the list is the whole point, and a Visitor
 * with no account has nothing to press to find out.
 */
export const publicListQuery = queryOptions({
  queryKey: ["public-list"],
  queryFn: () => publicClient.getPublicList({}),
  refetchInterval: PUBLIC_REFRESH_MS,
  refetchOnWindowFocus: true,
})
