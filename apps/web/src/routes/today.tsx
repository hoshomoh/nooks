import { createRoute, redirect } from "@tanstack/react-router"

import { TodayScreen } from "@/screens/today"
import { currentMemberQuery, instanceQuery } from "@/lib/queries"
import { listsQuery } from "@/lib/list-queries"
import { todayQuery } from "@/lib/dated-queries"
import { today } from "@/lib/dates"
import { rootRoute } from "./root"

/** Everything overdue or due today, from every List the Member can reach. */
export const todayRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/today",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }

    const [instance, lists, dated] = await Promise.all([
      context.queryClient.ensureQueryData(instanceQuery),
      context.queryClient.ensureQueryData(listsQuery),
      context.queryClient.ensureQueryData(todayQuery(today())),
    ])
    return { member, instance, lists, dated }
  },
  component: TodayScreen,
})
