import { createRoute } from "@tanstack/react-router"

import { PublicListScreen } from "@/screens/public-list"
import { publicListQuery } from "@/lib/public-queries"
import { rootRoute } from "./root"

/**
 * The public list, at a stable address.
 *
 * No loader guard and no redirect: this is the one page that is meant to answer to
 * somebody with no account at all.
 */
export const publicListRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/public",
  loader: async ({ context }) => {
    await context.queryClient.ensureQueryData(publicListQuery)
    return null
  },
  component: PublicListScreen,
})
