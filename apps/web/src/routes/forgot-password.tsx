import { createRoute, lazyRouteComponent } from "@tanstack/react-router"

import { rootRoute } from "./root"

export const forgotPasswordRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/forgot-password",
  component: lazyRouteComponent(() => import("@/screens/forgot-password"), "ForgotPassword"),
})
