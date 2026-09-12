import { createRoute, lazyRouteComponent, redirect } from "@tanstack/react-router"

import { instanceQuery } from "@/lib/queries"
import { rootRoute } from "./root"

/**
 * First run: make yourself an account and name the Instance, in one form.
 *
 * The loader turns this away once the Instance has an Admin, so a bookmarked URL cannot
 * reopen a step that is closed.
 */
export const setupRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/setup",
  loader: async ({ context }) => {
    const instance = await context.queryClient.ensureQueryData(instanceQuery)
    if (!instance.needsSetup) {
      throw redirect({ to: "/" })
    }
  },
  component: lazyRouteComponent(() => import("@/screens/setup"), "Setup"),
})
