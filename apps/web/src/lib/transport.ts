import { createConnectTransport } from "@connectrpc/connect-web"

import { connectionStore } from "./connection-store"
import { reportReachability } from "./reachability"

/**
 * The transport every generated client uses.
 *
 * Same-origin: in prod one binary serves the API and the app, and in dev Vite proxies
 * to it. That keeps cookies working identically in both.
 *
 * Every request reports what it found out to the connection store on its way back, so
 * the offline banner is answering from what actually happened rather than from what
 * the browser believes about the network.
 */
export const transport = createConnectTransport({
  baseUrl: "/",
  interceptors: [reportReachability(connectionStore)],
})
