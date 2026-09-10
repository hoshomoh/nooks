import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"

import { App } from "./App"
import { ThemeProvider } from "./components/theme-provider"
import "./index.css"

const queryClient = new QueryClient()

const container = document.getElementById("root")
if (!container) {
  throw new Error("no #root element to mount into")
}

createRoot(container).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <App />
      </ThemeProvider>
    </QueryClientProvider>
  </StrictMode>,
)
