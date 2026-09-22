import { createRoute, lazyRouteComponent, redirect } from "@tanstack/react-router"

import { instanceQuery } from "@/lib/queries"
import { rootRoute } from "./root"

/**
 * Asking for an account. Reachable signed out, which is the whole point.
 *
 * Unless the Instance has signup off, in which case the server refuses the request and
 * this is a form that cannot be submitted. Hiding the link on the sign-in screen is not
 * enough on its own: the address is typed, bookmarked and shared.
 */
export const joinRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/join",
  loader: async ({ context }) => {
    const instance = await context.queryClient.ensureQueryData(instanceQuery)
    if (!instance.publicSignup) {
      throw redirect({ to: "/sign-in" })
    }
  },
  component: lazyRouteComponent(() => import("@/screens/join"), "Join"),
})
