import { createRoute, redirect } from "@tanstack/react-router"

import { ReplacePassword } from "@/screens/replace-password"
import { currentMemberQuery } from "@/lib/queries"
import { rootRoute } from "./root"

/**
 * Replacing a temporary password. An Admin had to pick the first one, so it cannot
 * stay: the Member chooses their own now and the Admin's stops working.
 */
export const replacePasswordRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/replace-password",
  loader: async ({ context }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }
    if (!member.mustChangePassword) {
      throw redirect({ to: "/" })
    }
    return { member }
  },
  component: ReplacePassword,
})
