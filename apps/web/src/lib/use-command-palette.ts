import { useSyncExternalStore } from "react"

import { commandPaletteStore, type OpenMode, type PaletteMode } from "./command-palette-store"

/** The palette, as a screen reads it. */
export interface CommandPalette {
  mode: PaletteMode
  /** Which panel to draw, including while the palette is closing. */
  panel: OpenMode
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
  const panel = useSyncExternalStore(
    commandPaletteStore.subscribe,
    commandPaletteStore.getPanel,
    commandPaletteStore.getPanel,
  )

  return {
    mode,
    panel,
    open: commandPaletteStore.open,
    openAddList: commandPaletteStore.openAddList,
    close: commandPaletteStore.close,
  }
}
