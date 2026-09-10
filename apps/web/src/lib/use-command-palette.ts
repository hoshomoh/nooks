import { useSyncExternalStore } from "react"

import { commandPaletteStore, type PaletteMode } from "./command-palette-store"

/** Reads the palette's state from its store. */
export function useCommandPalette(): {
  mode: PaletteMode
  open: () => void
  openAddList: () => void
  close: () => void
} {
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
