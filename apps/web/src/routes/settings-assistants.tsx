import { createRoute, lazyRouteComponent, redirect } from "@tanstack/react-router"

import { currentMemberQuery } from "@/lib/queries"
import { rootRoute } from "./root"

/** How to point something that is not a browser at this Instance. */
export const settingsAssistantsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings/assistants",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }
    return null
  },
  component: lazyRouteComponent(
    () => import("@/screens/settings-assistants"),
    "SettingsAssistantsScreen",
  ),
})
