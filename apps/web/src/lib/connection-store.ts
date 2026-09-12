/**
 * Whether the Instance is reachable, as an external store.
 *
 * The machine's network state is an external mutable source, so React reads it through
 * useSyncExternalStore rather than synchronising with an effect. The store also carries
 * what the browser cannot tell us — a request that failed while the machine still
 * believed it was online — because a Member on hotel wifi is offline in every sense
 * that matters to them.
 */

/** What a screen reads about the connection. */
export interface ConnectionState {
  /** False while the machine reports no network, or the Instance has stopped answering. */
  online: boolean
  /** When contact was last good, as a stored moment, or null if it never has been. */
  lastSeenAt: string | null
}

export interface ConnectionStore {
  subscribe: (listener: () => void) => () => void
  getState: () => ConnectionState
  /** Records that the Instance answered. Called wherever a request succeeds. */
  markSeen: () => void
  /** Records that it did not. The banner does not wait for the machine to agree. */
  markUnreachable: () => void
}

/** Only the two events this store listens for, so a test needs no window. */
export interface ConnectionEvents {
  addEventListener: (type: "online" | "offline", listener: () => void) => void
  removeEventListener: (type: "online" | "offline", listener: () => void) => void
}

export interface ConnectionStoreDeps {
  /** The machine's own answer, read fresh each time rather than captured once. */
  isOnline: () => boolean
  events: ConnectionEvents
  now: () => Date
}

export function createConnectionStore(deps: ConnectionStoreDeps): ConnectionStore {
  // Held as one object and replaced rather than mutated: useSyncExternalStore compares
  // snapshots by identity, so a fresh object per read would re-render forever.
  let state: ConnectionState = {
    online: deps.isOnline(),
    lastSeenAt: deps.isOnline() ? deps.now().toISOString() : null,
  }
  const listeners = new Set<() => void>()

  const set = (next: ConnectionState) => {
    if (next.online === state.online && next.lastSeenAt === state.lastSeenAt) {
      return
    }
    state = next
    for (const listener of listeners) {
      listener()
    }
  }

  const seen = () => set({ online: true, lastSeenAt: deps.now().toISOString() })
  // Going offline keeps the last good moment: it is the only thing the banner has to
  // say about how stale what the Member is reading might be.
  const lost = () => set({ online: false, lastSeenAt: state.lastSeenAt })

  return {
    subscribe(listener) {
      // The window listeners live only while something is watching, which is the shape
      // useSyncExternalStore asks for.
      if (listeners.size === 0) {
        deps.events.addEventListener("online", seen)
        deps.events.addEventListener("offline", lost)
      }
      listeners.add(listener)

      return () => {
        listeners.delete(listener)
        if (listeners.size === 0) {
          deps.events.removeEventListener("online", seen)
          deps.events.removeEventListener("offline", lost)
        }
      }
    },
    getState: () => state,
    markSeen: seen,
    markUnreachable: lost,
  }
}

/** OFFLINE is the state a machine with no window reports: nothing has been seen. */
const OFFLINE: ConnectionState = { online: true, lastSeenAt: null }

/** The application's store. */
export const connectionStore: ConnectionStore =
  typeof window === "undefined"
    ? {
        subscribe: () => () => {},
        getState: () => OFFLINE,
        markSeen: () => {},
        markUnreachable: () => {},
      }
    : createConnectionStore({
        isOnline: () => window.navigator.onLine,
        events: window,
        now: () => new Date(),
      })
