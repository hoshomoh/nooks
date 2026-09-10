import path from "node:path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

// The dev server proxies the API to the Go binary on :8081 so that the browser
// talks to one origin and cookies behave as they do in production.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    port: 3001,
    proxy: {
      "/nooks.api.v1": { target: "http://localhost:8081", changeOrigin: true },
      "/api": { target: "http://localhost:8081", changeOrigin: true },
    },
  },
})
