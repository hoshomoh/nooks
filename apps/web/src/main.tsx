import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { RouterProvider } from "@tanstack/react-router"

import "./i18n"

import { buildRouter } from "./router"
import { wireLiveUpdates } from "./lib/live-wiring"
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
