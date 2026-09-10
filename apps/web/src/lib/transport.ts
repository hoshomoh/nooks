import { createConnectTransport } from "@connectrpc/connect-web"

/**
 * The transport every generated client uses.
 *
 * Same-origin: in prod one binary serves the API and the app, and in dev Vite proxies
 * to it. That keeps cookies working identically in both.
 */
export const transport = createConnectTransport({
  baseUrl: "/",
})
