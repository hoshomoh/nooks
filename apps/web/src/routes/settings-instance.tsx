import { createRoute, redirect } from "@tanstack/react-router"

import { SettingsInstanceScreen } from "@/screens/settings-instance"
import { currentMemberQuery } from "@/lib/queries"
import { instanceSettingsQuery } from "@/lib/instance-queries"
import { rootRoute } from "./root"

/** What the Instance is called, who may join, and the language it falls back to. */
export const settingsInstanceRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings/instance",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }
    await context.queryClient.ensureQueryData(instanceSettingsQuery)
    return null
  },
  component: SettingsInstanceScreen,
})
