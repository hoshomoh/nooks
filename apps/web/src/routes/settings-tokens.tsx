import { createRoute, lazyRouteComponent, redirect } from "@tanstack/react-router"

import { currentMemberQuery } from "@/lib/queries"
import { tokensQuery } from "@/lib/token-queries"
import { rootRoute } from "./root"

/** The keys a Member has cut for things that are not browsers. */
export const settingsTokensRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings/tokens",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }

    await context.queryClient.ensureQueryData(tokensQuery)
    return null
  },
  component: lazyRouteComponent(() => import("@/screens/settings-tokens"), "SettingsTokensScreen"),
})
