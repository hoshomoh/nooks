import { createRoute, redirect } from "@tanstack/react-router"

import { CalendarScreen } from "@/screens/calendar"
import { currentMemberQuery, instanceQuery } from "@/lib/queries"
import { listsQuery } from "@/lib/list-queries"
import { datedRangeQuery } from "@/lib/dated-queries"
import { monthWindow, today } from "@/lib/dates"
import { rootRoute } from "./root"

/** The month, as a lens on the same dated Items Upcoming lists. */
export const calendarRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/calendar",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }

    const { start, end } = monthWindow(today())

    const [instance, lists, dated] = await Promise.all([
      context.queryClient.ensureQueryData(instanceQuery),
      context.queryClient.ensureQueryData(listsQuery),
      context.queryClient.ensureQueryData(datedRangeQuery(start, end)),
    ])
    return { member, instance, lists, dated }
  },
  component: CalendarScreen,
})
