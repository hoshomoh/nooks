import { createRoute } from "@tanstack/react-router"

import { ForgotPassword } from "@/screens/forgot-password"
import { rootRoute } from "./root"

export const forgotPasswordRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/forgot-password",
  component: ForgotPassword,
})
