import { createRoute, redirect } from "@tanstack/react-router"

import { SettingsAccountScreen } from "@/screens/settings-account"
import { currentMemberQuery } from "@/lib/queries"
import { rootRoute } from "./root"

/** A Member's own account: their name, their email, their password. */
export const settingsAccountRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings/account",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }
    return null
  },
  component: SettingsAccountScreen,
})
