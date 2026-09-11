import { createRouter } from "@tanstack/react-router"

import { RoutePending } from "./components/ds/loading-rows"
import { RouteError } from "./components/ds/route-error"
import type { QueryClient } from "@tanstack/react-query"

import { forgotPasswordRoute } from "./routes/forgot-password"
import { indexRoute } from "./routes/index"
import { joinRoute } from "./routes/join"
import { listRoute } from "./routes/list"
import { noteRoute } from "./routes/note"
import { publicListRoute } from "./routes/public-list"
import { calendarRoute } from "./routes/calendar"
import { todayRoute } from "./routes/today"
import { upcomingRoute } from "./routes/upcoming"
import { replacePasswordRoute } from "./routes/replace-password"
import { rootRoute } from "./routes/root"
import { settingsGroupsRoute } from "./routes/settings-groups"
import { settingsMembersRoute } from "./routes/settings-members"
import { settingsPublicRoute } from "./routes/settings-public"
import { setupRoute } from "./routes/setup"
import { signInRoute } from "./routes/sign-in"

const routeTree = rootRoute.addChildren([
  indexRoute,
  setupRoute,
  signInRoute,
  replacePasswordRoute,
  joinRoute,
  forgotPasswordRoute,
  listRoute,
  todayRoute,
  upcomingRoute,
  calendarRoute,
  noteRoute,
  settingsMembersRoute,
  settingsGroupsRoute,
  publicListRoute,
  settingsPublicRoute,
])

/** buildRouter takes the query client so loaders can prime the cache before rendering. */
export function buildRouter(queryClient: QueryClient) {
  return createRouter({
    routeTree,
    context: { queryClient },
    defaultPreload: "intent",
    // Also the Suspense boundary every screen reads its data behind.
    defaultPendingComponent: RoutePending,
    // And the error boundary. A screen reads its data with useSuspenseQuery, which
    // throws when a refetch fails, and a blank page is not an answer.
    defaultErrorComponent: RouteError,
  })
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof buildRouter>
  }
}
