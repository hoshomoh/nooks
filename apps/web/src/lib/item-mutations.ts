/**
 * What the Item mutations are told.
 *
 * Named here rather than written into each `mutationFn`, so the screens that tick the
 * same Item from three different views all agree on the shape.
 */

/** Ticking or unticking one Item. */
export interface SetDoneVariables {
  itemUid: string
  done: boolean
}

/** Saving an Item's Note. */
export interface SaveNoteVariables {
  itemUid: string
  note: string
}

/** Renaming an Item. */
export interface RenameItemVariables {
  itemUid: string
  label: string
}

/** Adding an Item to a List. */
export interface CreateItemVariables {
  listUid: string
  label: string
  quantity: string
  dueOn: string
}
