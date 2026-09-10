import { createRoute, redirect } from "@tanstack/react-router"

import { Home } from "@/screens/home"
import { currentMemberQuery, instanceQuery } from "@/lib/queries"
import { rootRoute } from "./root"

/**
 * Where an arriving Member lands.
 *
 * The decision happens in the loader, before anything renders: an Instance with no
 * Admin goes to first run, a signed-out browser goes to sign in, and a Member who was
 * given a temporary password goes to replace it. Doing this in a loader rather than an
 * effect means no screen flashes on the way through.
 */
export const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  loader: async ({ context }) => {
    const instance = await context.queryClient.ensureQueryData(instanceQuery)
    if (instance.needsSetup) {
      throw redirect({ to: "/setup" })
    }

    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }
    if (member.mustChangePassword) {
      throw redirect({ to: "/replace-password" })
    }
    return { instance, member }
  },
  component: Home,
})
