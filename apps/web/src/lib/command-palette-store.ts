/**
 * Whether the ⌘K palette is open, as an external store.
 *
 * It lives outside React because the keyboard shortcut is a document-level listener:
 * a store can own both the listener and the state, and React reads it with
 * useSyncExternalStore rather than synchronising with an effect.
 */
export type PaletteMode = "closed" | "search" | "add-list"

export type CommandPaletteStore = {
  subscribe: (listener: () => void) => () => void
  getMode: () => PaletteMode
  open: () => void
  openAddList: () => void
  close: () => void
}

export type CommandPaletteDeps = {
  /** Where the ⌘K listener is attached. */
  target: Pick<Window, "addEventListener" | "removeEventListener">
}

export function createCommandPaletteStore(deps: CommandPaletteDeps): CommandPaletteStore {
  let mode: PaletteMode = "closed"
  const listeners = new Set<() => void>()

  const set = (next: PaletteMode) => {
    if (next === mode) {
      return
    }
    mode = next
    for (const listener of listeners) {
      listener()
    }
  }

  deps.target.addEventListener("keydown", ((event: KeyboardEvent) => {
    if (event.key.toLowerCase() !== "k" || !(event.metaKey || event.ctrlKey)) {
      return
    }
    // The browser's own find-in-page is not what a Member means by ⌘K here.
    event.preventDefault()
    set(mode === "closed" ? "search" : "closed")
  }) as EventListener)

  return {
    subscribe(listener) {
      listeners.add(listener)
      return () => listeners.delete(listener)
    },
    getMode: () => mode,
    open: () => set("search"),
    openAddList: () => set("add-list"),
    close: () => set("closed"),
  }
}

/** The application's store, created once. */
export const commandPaletteStore: CommandPaletteStore =
  typeof window === "undefined"
    ? {
        subscribe: () => () => {},
        getMode: () => "closed",
        open: () => {},
        openAddList: () => {},
        close: () => {},
      }
    : createCommandPaletteStore({ target: window })
