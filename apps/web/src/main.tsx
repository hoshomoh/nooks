import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { RouterProvider } from "@tanstack/react-router"

import "./i18n"

import { buildRouter } from "./router"
import { browserCacheDeps, persistCache, restoreCache } from "./lib/cache-store"
import { connectionStore } from "./lib/connection-store"
import { wireLiveUpdates } from "./lib/live-wiring"
import { wireOnline } from "./lib/online-wiring"
import "./index.css"

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // A home server can be slow or asleep; retrying twice is worth it, retrying
      // forever is not.
      retry: 2,
      refetchOnWindowFocus: false,
    },
  },
})

// What Nooks means by offline, rather than what the browser guesses. Wired before
// anything can be mutated, so a change made on a dead connection is paused and kept
// instead of failing and being lost.
wireOnline(connectionStore)

// Put back what this browser already knew, then keep it up to date. Restoring happens
// before the router is built: its loaders read with `ensureQueryData`, and a loader
// that finds an answer never reaches for the network at all.
restoreCache(queryClient, browserCacheDeps)
persistCache(queryClient, browserCacheDeps)

const router = buildRouter(queryClient)

// Live updates are wired outside React: the stream outlives any screen, and navigation
// is what tells it which List is being read.
const live = wireLiveUpdates(queryClient)
live.navigated(window.location.pathname)
router.subscribe("onResolved", ({ toLocation }) => live.navigated(toLocation.pathname))

const container = document.getElementById("root")
if (!container) {
  throw new Error("no #root element to mount into")
}

createRoot(container).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
)
