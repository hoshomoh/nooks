import type { QueryClient } from "@tanstack/react-query"

import { liveStore, type LiveEvent } from "./live-store"
import { refreshLists } from "./refresh"

/**
 * Connects the live stream to the rest of the app, once, outside React.
 *
 * Two wires. Changes arriving invalidate what they affect, so a tick somebody else made
 * appears without a reload. Navigation points the stream at the List on screen, so
 * presence knows who is standing where.
 *
 * Neither belongs in a component: an event stream is a connection that outlives any
 * screen, and reacting to one from inside React would mean an effect whose only job is
 * to re-subscribe on every navigation.
 */
export interface LiveWiring {
  /** Tells the stream which List is on screen, or none. */
  navigated: (pathname: string) => void
  /** Stops listening. Called by a test; the app itself never stops. */
  stop: () => void
}

export function wireLiveUpdates(queryClient: QueryClient): LiveWiring {
  const stop = liveStore.onEvent((event) => void act(queryClient, event))
  return {
    navigated: (pathname) => liveStore.watch(listUidIn(pathname)),
    stop,
  }
}

/**
 * act re-reads whatever the change affected.
 *
 * The event says what changed rather than what it changed to, so this asks the same API
 * the screen always asks. A live update can never show somebody something the API would
 * not have given them.
 */
async function act(queryClient: QueryClient, event: LiveEvent): Promise<void> {
  switch (event.kind) {
    case "activity":
      await queryClient.invalidateQueries({ queryKey: ["activity"] })
      return
    case "presence":
      // Presence is read from the store itself; nothing on the server to re-read.
      return
    default:
      await refreshLists(queryClient)
  }
}

/**
 * listUidIn reads the List a path is showing, or an empty string.
 *
 * Pure, so which paths count as "on a List" can be read and tested in one place. An
 * Item's Note counts: a Member reading it is standing on that List.
 */
export function listUidIn(pathname: string): string {
  const match = /^\/lists\/([^/]+)/.exec(pathname)
  return match?.[1] ? decodeURIComponent(match[1]) : ""
}
