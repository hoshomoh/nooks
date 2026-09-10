import { createRouter } from "@tanstack/react-router"
import type { QueryClient } from "@tanstack/react-query"

import { forgotPasswordRoute } from "./routes/forgot-password"
import { indexRoute } from "./routes/index"
import { joinRoute } from "./routes/join"
import { replacePasswordRoute } from "./routes/replace-password"
import { rootRoute } from "./routes/root"
import { setupRoute } from "./routes/setup"
import { signInRoute } from "./routes/sign-in"

const routeTree = rootRoute.addChildren([
  indexRoute,
  setupRoute,
  signInRoute,
  replacePasswordRoute,
  joinRoute,
  forgotPasswordRoute,
])

/** buildRouter takes the query client so loaders can prime the cache before rendering. */
export function buildRouter(queryClient: QueryClient) {
  return createRouter({
    routeTree,
    context: { queryClient },
    defaultPreload: "intent",
  })
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof buildRouter>
  }
}
