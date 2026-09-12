import { createRoute, lazyRouteComponent, redirect } from "@tanstack/react-router"

import { currentMemberQuery } from "@/lib/queries"
import { instanceSettingsQuery } from "@/lib/instance-queries"
import { rootRoute } from "./root"

/** You, and — for an Admin — the Instance. */
export const settingsGeneralRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings/general",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }
    // Only an Admin may read these, so a Member simply renders no Instance section.
    await context.queryClient.ensureQueryData(instanceSettingsQuery).catch(() => undefined)
    return null
  },
  component: lazyRouteComponent(() => import("@/screens/settings-general"), "SettingsGeneralScreen"),
})
