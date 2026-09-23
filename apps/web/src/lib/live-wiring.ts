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
  /**
   * Tells the stream which List is on screen and which Note is open on it.
   *
   * The Note is a search parameter rather than part of the path, because opening one is
   * a state of the List screen rather than a place of its own. Both are read here, so
   * the claim on a Note is made and given up by navigating, which is what opening and
   * closing the sheet does.
   */
  navigated: (pathname: string, search?: string) => void
  /** Stops listening. Called by a test; the app itself never stops. */
  stop: () => void
}

export function wireLiveUpdates(queryClient: QueryClient): LiveWiring {
  const stop = liveStore.onEvent((event) => void actOn(queryClient, event))
  return {
    navigated: (pathname, search) =>
      liveStore.watch(listUidIn(pathname), openItemIn(search ?? "")),
    stop,
  }
}

/**
 * actOn re-reads whatever the change affected.
 *
 * The event says what changed rather than what it changed to, so this asks the same API
 * the screen always asks. A live update can never show somebody something the API would
 * not have given them.
 *
 * Exported for the same reason listUidIn is: which keys a kind invalidates is the whole
 * of this wiring, and it is worth being able to read it back.
 */
export async function actOn(queryClient: QueryClient, event: LiveEvent): Promise<void> {
  switch (event.kind) {
    case "activity":
      await queryClient.invalidateQueries({ queryKey: ["activity"] })
      return
    case "presence":
      // Presence is read from the store itself; nothing on the server to re-read.
      return
    case "member.changed":
      // Who we are signed in as is read once and held for the life of the tab, on
      // purpose: it does not change because somebody ticked the milk. When it does
      // change, this is the only thing that says so.
      await queryClient.invalidateQueries({ queryKey: ["current-member"] })
      return
    case "list.changed":
    case "lists.changed":
      await refreshLists(queryClient)
      return
    default:
      // A kind this tab has never heard of, from an Instance newer than it. Re-reading
      // the Lists is the safe guess and beats throwing; naming the known kinds above
      // is what keeps this branch meaning only that.
      await refreshLists(queryClient)
  }
}

/**
 * listUidIn reads the List a path is showing, or an empty string.
 *
 * Pure, so which paths count as "on a List" can be read and tested in one place. An
 * Item's Note counts: a Member reading it is standing on that List.
 */
/**
 * openItemIn is the Item whose Note is open, read from the search.
 *
 * Exported for the same reason listUidIn is: what the stream is told about a screen is
 * the whole of this wiring, and it is worth being able to read it back.
 */
export function openItemIn(search: string): string {
  return new URLSearchParams(search).get("item") ?? ""
}

export function listUidIn(pathname: string): string {
  const match = /^\/lists\/([^/]+)/.exec(pathname)
  return match?.[1] ? decodeURIComponent(match[1]) : ""
}
