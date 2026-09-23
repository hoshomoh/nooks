/**
 * The stream of changes other people are making, as an external store.
 *
 * It lives outside React because an EventSource is a connection, not a value: one is
 * opened per List being read, and React reads what it has produced with
 * useSyncExternalStore rather than synchronising with an effect.
 */

export type LiveKind =
  | "list.changed"
  | "lists.changed"
  | "activity"
  | "presence"
  | "member.changed"

/** One change, as the Instance describes it. */
export interface LiveEvent {
  kind: LiveKind
  /** The List it concerns, when it concerns one. */
  listUid?: string
  /** Who is looking at that List, for a presence event. */
  watchers?: string[]
  /** Who has a Note open, by Item. Never includes the reader. */
  editing?: NoteEditor[]
}

/** One person with one Note open. */
export interface NoteEditor {
  itemUid: string
  name: string
}

/** What a screen reads from the stream. */
export interface LiveState {
  /** Who else is looking at the List being read. Never includes the reader. */
  watchers: string[]
  /**
   * Who else has a Note open on it, by Item.
   *
   * A Note somebody else is in is read rather than written, which is how two people are
   * stopped from losing each other's words. Never includes the reader, so a Note this
   * Member has open never appears here and never reads as locked to them.
   */
  editing: NoteEditor[]
  /** Bumped on every change, so a reader can tell that something arrived. */
  version: number
}

export interface LiveStore {
  subscribe: (listener: () => void) => () => void
  getState: () => LiveState
  /**
   * Points the connection at a List, or at nothing.
   *
   * Called as a Member moves between screens. Pointing it where it already points does
   * nothing, so this is safe to call on every render.
   */
  watch: (listUid: string, editingItemUid?: string) => void
  /** Called with every change, so a reader can act on it. */
  onEvent: (listener: (event: LiveEvent) => void) => () => void
}

export interface LiveStoreDeps {
  /** Opens the stream. Injected so a test needs no EventSource. */
  connect: (url: string) => LiveConnection
  /** Where the stream lives, without the query string. */
  endpoint: string
}

/** The part of an EventSource this store uses. */
export interface LiveConnection {
  addEventListener: (kind: "message", listener: (event: MessageEvent<string>) => void) => void
  close: () => void
}

/** NOBODY is the state before anything has arrived. */
const NOBODY: LiveState = { watchers: [], editing: [], version: 0 }

export function createLiveStore(deps: LiveStoreDeps): LiveStore {
  let state = NOBODY
  let watching = ""
  let editing = ""
  let connection: LiveConnection | null = null
  const listeners = new Set<() => void>()
  const eventListeners = new Set<(event: LiveEvent) => void>()

  const notify = () => {
    for (const listener of listeners) {
      listener()
    }
  }

  const receive = (event: LiveEvent) => {
    if (event.kind === "presence" && event.listUid === watching) {
      state = {
        watchers: event.watchers ?? [],
        editing: event.editing ?? [],
        version: state.version + 1,
      }
    } else {
      state = { ...state, version: state.version + 1 }
    }
    notify()

    for (const listener of eventListeners) {
      listener(event)
    }
  }

  const reconnect = () => {
    connection?.close()
    state = NOBODY
    connection = deps.connect(urlFor(deps.endpoint, watching, editing))
    connection.addEventListener("message", (message) => {
      const event = parseEvent(message.data)
      if (event) {
        receive(event)
      }
    })
  }

  return {
    subscribe(listener) {
      listeners.add(listener)
      return () => listeners.delete(listener)
    },
    getState: () => state,
    /*
     * Points the connection at a List, and at the Note being edited on it.
     *
     * Opening a Note reconnects, which is how the Instance is told it is being held:
     * the claim lives on the connection, so a laptop closed mid-edit releases it when
     * the stream drops rather than holding a Note for ever.
     */
    watch(listUid, editingItemUid = "") {
      if (listUid === watching && editingItemUid === editing && connection) {
        return
      }
      watching = listUid
      editing = editingItemUid
      reconnect()
      notify()
    },
    onEvent(listener) {
      eventListeners.add(listener)
      return () => eventListeners.delete(listener)
    },
  }
}

/** urlFor names the List the stream should follow, and the Note held open on it. */
function urlFor(endpoint: string, listUid: string, editing: string): string {
  if (!listUid) {
    return endpoint
  }
  const where = `${endpoint}?list=${encodeURIComponent(listUid)}`
  return editing ? `${where}&editing=${encodeURIComponent(editing)}` : where
}

/**
 * parseEvent reads one message, or nothing.
 *
 * A message that cannot be read is dropped rather than thrown: the stream is a
 * courtesy on top of data the app already has, and it must never take the app down.
 */
function parseEvent(data: string): LiveEvent | null {
  try {
    const parsed: unknown = JSON.parse(data)
    if (typeof parsed !== "object" || parsed === null || !("kind" in parsed)) {
      return null
    }
    return parsed as LiveEvent
  } catch {
    return null
  }
}

/** LIVE_ENDPOINT is where the Instance streams changes from. */
export const LIVE_ENDPOINT = "/api/v1/events"

export const liveStore: LiveStore =
  typeof window === "undefined"
    ? {
        subscribe: () => () => {},
        getState: () => NOBODY,
        watch: () => {},
        onEvent: () => () => {},
      }
    : createLiveStore({
        endpoint: LIVE_ENDPOINT,
        connect: (url) => new EventSource(url, { withCredentials: true }),
      })
