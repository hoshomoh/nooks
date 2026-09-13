import { useCallback, useMemo } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"

import { listClient } from "./api"
import { debounce } from "./debounce"
import { refreshLists } from "./refresh"
import type { SaveNoteVariables } from "./item-mutations"

/**
 * How long typing has to settle before a Note is saved.
 *
 * One number, because a Note is written in two places — the sheet beside the List and
 * the full-width page — and they are the same Note. Two copies of this drift the moment
 * somebody tunes one of them, and the Member finds that the same words save at two
 * different speeds depending on which screen they happened to open.
 */
const SETTLE_MS = 800

/** Saving a Note as it is written. */
export interface NoteAutosave {
  /** Schedule a save, replacing any that was waiting. */
  save: (variables: SaveNoteVariables) => void
  /**
   * Save what is waiting, now.
   *
   * Called on the way out of a Note. Without it the last sentence belongs to a timer
   * that never runs, and the Member watches their own words disappear.
   */
  flush: () => void
  /** Whether a save is in flight, for the word the screens show while it is. */
  isSaving: boolean
}

/**
 * The autosave both Note screens use.
 *
 * The mutation, the delay and the flush travelled together in each screen, which meant
 * the rule about what happens to unsaved words on the way out was written down twice.
 */
export function useNoteAutosave(): NoteAutosave {
  const queryClient = useQueryClient()

  const saveNote = useMutation({
    mutationFn: ({ itemUid, note }: SaveNoteVariables) => listClient.updateItem({ itemUid, note }),
    onSuccess: () => refreshLists(queryClient),
  })

  // `mutate` is referentially stable, so the debounce is built once.
  const autosave = useMemo(() => debounce(saveNote.mutate, SETTLE_MS), [saveNote.mutate])

  return {
    save: autosave.call,
    flush: useCallback(() => autosave.flush(), [autosave]),
    isSaving: saveNote.isPending,
  }
}
