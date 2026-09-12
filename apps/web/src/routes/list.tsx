import { createRoute, lazyRouteComponent, redirect } from "@tanstack/react-router"

import { currentMemberQuery, instanceQuery } from "@/lib/queries"
import { listQuery, listsQuery } from "@/lib/list-queries"
import { rootRoute } from "./root"

/** One List, with its Items. */
export const listRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/lists/$listUid",
  loader: async ({ context, params }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }

    // The three reads a signed-in screen always needs, fetched together rather than
    // one after another.
    const [instance, lists, list] = await Promise.all([
      context.queryClient.ensureQueryData(instanceQuery),
      context.queryClient.ensureQueryData(listsQuery),
      context.queryClient.ensureQueryData(listQuery(params.listUid)),
    ])
    return { member, instance, lists, list }
  },
  component: lazyRouteComponent(() => import("@/screens/list"), "ListScreen"),
})
