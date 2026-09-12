import { createRoute, lazyRouteComponent } from "@tanstack/react-router"

import { rootRoute } from "./root"

/** Asking for an account. Reachable signed out, which is the whole point. */
export const joinRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/join",
  component: lazyRouteComponent(() => import("@/screens/join"), "Join"),
})
