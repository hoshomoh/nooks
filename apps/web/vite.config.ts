import path from "node:path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
// vitest's defineConfig, which is Vite's plus the `test` block below.
import { defineConfig } from "vitest/config"

/**
 * Where the Go binary is listening.
 *
 * `:8081` is what `nooks` uses unless it is told otherwise, and `--addr` can tell it
 * otherwise. Anybody who moves it sets NOOKS_DEV_TARGET to match rather than editing a
 * tracked file, which would then be a change they have to remember not to commit.
 */
const INSTANCE = process.env.NOOKS_DEV_TARGET ?? "http://localhost:8081"

// The dev server proxies the API to the Go binary so that the browser
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
     * Two of them. One was tried when runs kept being killed, and kept being killed —
     * the machine had 11GB available at the time, so the pressure was never this suite's
     * to relieve.
     *
     * Isolation is kept: each file still gets a fresh module registry, which the
     * module-level stores (the theme, the locale, the pending tick) depend on. Turning
     * that off is the other way to save memory and it would make tests share state.
     */
    pool: "threads",
    maxWorkers: 2,
  },
  server: {
    port: 3001,
    /*
     * Everything the instance owns, forwarded to the binary.
     *
     * A path that is not here does not fall through to the instance — it falls through
     * to the SPA, which answers a GET with index.html and a POST with 404. That makes a
     * missing entry look like a working endpoint in a browser and a broken one to every
     * client that speaks the protocol, which is exactly how /mcp was lost.
     */
    proxy: {
      "/nooks.api.v1": { target: INSTANCE, changeOrigin: true },
      // The event stream must not be buffered by the dev proxy either.
      "/api/v1/events": { target: INSTANCE, changeOrigin: true },
      "/api": { target: INSTANCE, changeOrigin: true },
      "/mcp": { target: INSTANCE, changeOrigin: true },
    },
  },
})
