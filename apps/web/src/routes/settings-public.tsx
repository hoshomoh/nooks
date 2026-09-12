import { createRoute, lazyRouteComponent, redirect } from "@tanstack/react-router"

import { currentMemberQuery } from "@/lib/queries"
import { instanceSettingsQuery } from "@/lib/instance-queries"
import { listsQuery } from "@/lib/list-queries"
import { rootRoute } from "./root"

/** Which List an Instance publishes, and how much of it a Visitor sees. */
export const settingsPublicRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings/public",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }

    await Promise.all([
      context.queryClient.ensureQueryData(instanceSettingsQuery),
      context.queryClient.ensureQueryData(listsQuery),
    ])
    return null
  },
  component: lazyRouteComponent(() => import("@/screens/settings-public"), "SettingsPublicScreen"),
})
