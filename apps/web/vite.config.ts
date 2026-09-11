import path from "node:path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
// vitest's defineConfig, which is Vite's plus the `test` block below.
import { defineConfig } from "vitest/config"

// The dev server proxies the API to the Go binary on :8081 so that the browser
// talks to one origin and cookies behave as they do in production.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: [
      { find: "@", replacement: path.resolve(import.meta.dirname, "./src") },
      // The generated shadcn components import "cn" directly and must not be edited,
      // so the name resolves to Nooks' configured merger — see src/lib/cn.ts.
      {
        find: /^cn$/,
        replacement: path.resolve(import.meta.dirname, "./src/lib/cn.ts"),
      },
    ],
  },
  test: {
    // A test that renders types one character at a time through a real event loop, and
    // CI runs the whole suite beside a Go build. Five seconds is the default and it is
    // not enough on a loaded machine; this is about the machine, not the code.
    testTimeout: 20_000,
    hookTimeout: 20_000,
    /*
     * Threads rather than forked processes, and two of them.
     *
     * Vitest's default pool forks a Node process per worker, each with its own heap and
     * its own jsdom. Threads share one heap instead, which is the difference between a
     * suite that runs beside a Go build and a Vite build and one the kernel kills —
     * three runs of ./scripts/preflight.sh in a row died here before this changed.
     *
     * One of them, because two still got killed. The suite is slower for it and that is
     * the right trade: a gate that fails half the time is not a gate, and this one is
     * what stands between a broken commit and main.
     *
     * Isolation is kept: each file still gets a fresh module registry, which the
     * module-level stores (the theme, the locale, the pending tick) depend on. Turning
     * that off is the other way to save memory and it would make tests share state.
     */
    pool: "threads",
    maxWorkers: 1,
  },
  server: {
    port: 3001,
    proxy: {
      "/nooks.api.v1": { target: "http://localhost:8081", changeOrigin: true },
      // The event stream must not be buffered by the dev proxy either.
      "/api/v1/events": { target: "http://localhost:8081", changeOrigin: true },
      "/api": { target: "http://localhost:8081", changeOrigin: true },
    },
  },
})
