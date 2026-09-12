import { createRoute, lazyRouteComponent, redirect } from "@tanstack/react-router"

import { aboutQuery } from "@/lib/about-queries"
import { currentMemberQuery } from "@/lib/queries"
import { rootRoute } from "./root"

/** What this copy of Nooks is, and how much it holds. */
export const settingsAboutRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings/about",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }
    await context.queryClient.ensureQueryData(aboutQuery)
    return null
  },
  component: lazyRouteComponent(() => import("@/screens/settings-about"), "SettingsAboutScreen"),
})
