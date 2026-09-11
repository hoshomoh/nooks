import { createRoute, redirect } from "@tanstack/react-router"

import { SettingsAppearanceScreen } from "@/screens/settings-appearance"
import { currentMemberQuery } from "@/lib/queries"
import { rootRoute } from "./root"

/** How Nooks looks, and what language it speaks. */
export const settingsAppearanceRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/settings/appearance",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }
    return null
  },
  component: SettingsAppearanceScreen,
})
