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
  /**
   * Which panel to draw, including while the palette is closing.
   *
   * A dialog takes a moment to animate out, and reading `mode` during that moment
   * would swap the panel for the search one on the way — so closing "Add a list" would
   * flash the search box at the Member as it went.
   */
  getPanel: () => OpenMode
  open: () => void
  openAddList: () => void
  close: () => void
}

/** The panels the palette has, leaving out the closed state. */
export type OpenMode = Exclude<PaletteMode, "closed">

export type CommandPaletteDeps = {
  /** Where the ⌘K listener is attached. */
  target: Pick<Window, "addEventListener" | "removeEventListener">
}

export function createCommandPaletteStore(deps: CommandPaletteDeps): CommandPaletteStore {
  let mode: PaletteMode = "closed"
  let panel: OpenMode = "search"
  const listeners = new Set<() => void>()

  const set = (next: PaletteMode) => {
    if (next === mode) {
      return
    }
    mode = next
    if (next !== "closed") {
      panel = next
    }
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
    getPanel: () => panel,
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
        getPanel: () => "search",
        open: () => {},
        openAddList: () => {},
        close: () => {},
      }
    : createCommandPaletteStore({ target: window })
