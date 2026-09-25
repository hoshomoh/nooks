import path from "node:path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { VitePWA } from "vite-plugin-pwa"
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
  plugins: [
    react(),
    tailwindcss(),
    /*
     * Keeps the app itself on the machine that opened it.
     *
     * Reading, ticking and adding already survived losing the connection. Closing the
     * tab did not, because reopening it fetched the app from an Instance that was not
     * answering.
     *
     * A service worker needs a secure context, so this does nothing over plain HTTP to
     * a LAN address. That is the browser's rule, not ours.
     */
    VitePWA({
      registerType: "prompt",
      includeAssets: ["favicon.svg", "icon.svg"],
      manifest: {
        name: "nooks",
        short_name: "nooks",
        description: "A household todo app you run on your own machine.",
        start_url: "/",
        scope: "/",
        display: "standalone",
        background_color: "#F2F2F0",
        theme_color: "#F2F2F0",
        icons: [
          { src: "/icon-192.png", sizes: "192x192", type: "image/png" },
          { src: "/icon-512.png", sizes: "512x512", type: "image/png" },
          {
            src: "/icon-maskable-512.png",
            sizes: "512x512",
            type: "image/png",
            purpose: "maskable",
          },
        ],
      },
      workbox: {
        // The app, not the API. What the Instance knows is TanStack Query's to cache,
        // and a second copy here would be a second answer to the same question.
        globPatterns: ["**/*.{js,css,html,svg,png,woff2}"],
        // A client route is the app, not a missing file. Same rule as frontend.go.
        navigateFallback: "/index.html",
        navigateFallbackDenylist: [/^\/api/, /^\/mcp/, /^\/nooks\.api\./, /^\/healthz/],
      },
    }),
  ],
  resolve: {
    alias: [
      { find: "@", replacement: path.resolve(import.meta.dirname, "./src") },
      // The generated shadcn components import "cn" directly and must not be edited,
      // so the name resolves to nooks' configured merger — see src/lib/cn.ts.
      {
        find: /^cn$/,
        replacement: path.resolve(import.meta.dirname, "./src/lib/cn.ts"),
      },
    ],
  },
  test: {
    // Rendering a sentence one keystroke at a time through a real event loop takes
    // longer than vitest's five second default allows.
    testTimeout: 20_000,
    hookTimeout: 20_000,
    /*
     * Threads rather than forked processes, and two of them.
     *
     * The default pool forks a Node process per worker, each with its own heap and its
     * own jsdom. Threads share one heap, which is what keeps the suite inside the memory
     * of a machine that is also building Go and Vite.
     *
     * Two rather than one per core for the same reason: the extra speed is small and it
     * is spent when the suite runs beside everything else in ./scripts/ci.sh.
     */
    pool: "threads",
    maxWorkers: 2,
    /*
     * Workers are reused across files rather than replaced between them.
     *
     * Isolation re-evaluates the whole dependency graph once per test file, which was
     * two thirds of the run. What it was paying for is the cleanup between files, and
     * the two projects below buy that instead: a setup file runs per test file with or
     * without isolation, so a render left behind by one file cannot be found by the next.
     */
    isolate: false,
    projects: [
      {
        extends: true,
        test: {
          name: "unit",
          include: ["src/**/*.test.ts"],
          environment: "node",
        },
      },
      {
        extends: true,
        test: {
          name: "dom",
          include: ["src/**/*.test.tsx"],
          environment: "jsdom",
          setupFiles: ["./src/test/dom.ts"],
        },
      },
    ],
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
