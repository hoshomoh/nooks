import { createRoute, redirect } from "@tanstack/react-router"

import { SignIn } from "@/screens/sign-in"
import { instanceQuery } from "@/lib/queries"
import { rootRoute } from "./root"

export const signInRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/sign-in",
  loader: async ({ context }) => {
    const instance = await context.queryClient.ensureQueryData(instanceQuery)
    // An Instance with no Admin has nobody to sign in as.
    if (instance.needsSetup) {
      throw redirect({ to: "/setup" })
    }
    return { instance }
  },
  component: SignIn,
})
