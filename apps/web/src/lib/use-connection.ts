import { useSyncExternalStore } from "react"

import { connectionStore, type ConnectionState } from "./connection-store"

/**
 * Reads the connection from its store.
 *
 * useSyncExternalStore rather than an effect: whether the machine has a network is an
 * external mutable source, and this is the API React provides for exactly that.
 */
export function useConnection(): ConnectionState {
  return useSyncExternalStore(
    connectionStore.subscribe,
    connectionStore.getState,
    connectionStore.getState,
  )
}
