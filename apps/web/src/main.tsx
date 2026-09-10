import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { RouterProvider } from "@tanstack/react-router"

import { buildRouter } from "./router"
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
