import { createRootRouteWithContext } from "@tanstack/react-router"
import type { QueryClient } from "@tanstack/react-query"

import { Root } from "@/screens/root"

/**
 * The router's context. Loaders receive it, which is how a route can ensure its data is
 * present before it renders rather than fetching from inside a component.
 */
export type RouterContext = {
  queryClient: QueryClient
}

export const rootRoute = createRootRouteWithContext<RouterContext>()({
  component: Root,
})
