import { useSyncExternalStore } from "react"

import { commandPaletteStore, type PaletteMode } from "./command-palette-store"

/** The palette, as a screen reads it. */
export interface CommandPalette {
  mode: PaletteMode
  open: () => void
  openAddList: () => void
  close: () => void
}

/** Reads the palette's state from its store. */
export function useCommandPalette(): CommandPalette {
  const mode = useSyncExternalStore(
    commandPaletteStore.subscribe,
    commandPaletteStore.getMode,
    commandPaletteStore.getMode,
  )
  return {
    mode,
    open: commandPaletteStore.open,
    openAddList: commandPaletteStore.openAddList,
    close: commandPaletteStore.close,
  }
}
