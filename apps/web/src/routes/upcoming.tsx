import { createRoute, redirect } from "@tanstack/react-router"

import { UpcomingScreen } from "@/screens/upcoming"
import { currentMemberQuery, instanceQuery } from "@/lib/queries"
import { listsQuery } from "@/lib/list-queries"
import { upcomingQuery } from "@/lib/dated-queries"
import { today } from "@/lib/dates"
import { rootRoute } from "./root"

/** The next two weeks, grouped by day. */
export const upcomingRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/upcoming",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }

    const [instance, lists, dated] = await Promise.all([
      context.queryClient.ensureQueryData(instanceQuery),
      context.queryClient.ensureQueryData(listsQuery),
      context.queryClient.ensureQueryData(upcomingQuery(today())),
    ])
    return { member, instance, lists, dated }
  },
  component: UpcomingScreen,
})
