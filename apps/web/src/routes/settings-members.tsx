import { createRoute, redirect } from "@tanstack/react-router"

import { SettingsMembersScreen } from "@/screens/settings-members"
import { currentMemberQuery, instanceQuery } from "@/lib/queries"
import { pendingRequestsQuery } from "@/lib/member-queries"
import { groupsQuery, membersQuery } from "@/lib/sharing-queries"
import { rootRoute } from "./root"

/** Who is here, and who is waiting. */
export const settingsMembersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings/members",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }

    await Promise.all([
      context.queryClient.ensureQueryData(instanceQuery),
      context.queryClient.ensureQueryData(membersQuery),
      context.queryClient.ensureQueryData(groupsQuery),
      // Only an Admin may read these, so a Member simply sees no requests section.
      context.queryClient.ensureQueryData(pendingRequestsQuery).catch(() => undefined),
    ])
    return null
  },
  component: SettingsMembersScreen,
})
