import { createRoute, lazyRouteComponent, redirect } from "@tanstack/react-router"

import { currentMemberQuery } from "@/lib/queries"
import { groupsQuery, membersQuery } from "@/lib/sharing-queries"
import { rootRoute } from "./root"

/** The Groups, and who is in them. */
export const settingsGroupsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings/groups",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }

    await Promise.all([
      context.queryClient.ensureQueryData(groupsQuery),
      context.queryClient.ensureQueryData(membersQuery),
    ])
    return null
  },
  component: lazyRouteComponent(() => import("@/screens/settings-groups"), "SettingsGroupsScreen"),
})
