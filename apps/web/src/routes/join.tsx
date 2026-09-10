import { createRoute } from "@tanstack/react-router"

import { Join } from "@/screens/join"
import { rootRoute } from "./root"

/** Asking for an account. Reachable signed out, which is the whole point. */
export const joinRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/join",
  component: Join,
})
