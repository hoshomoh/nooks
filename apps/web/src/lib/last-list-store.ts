/**
 * Which List a Member last added something to, as an external store.
 *
 * Today and Upcoming gather Items from every List, so an Item added there has to land
 * somewhere, and that is almost always the List they used last.
 *
 * In the browser rather than on the Instance: it is how one person is working this
 * week, not a fact about the household. Outside React because it is written from a
 * mutation and read from a screen.
 */
export interface LastListStore {
  subscribe: (listener: () => void) => () => void
  /** The List last added to, or an empty string when there is no memory of one. */
  getUid: () => string
  remember: (listUid: string) => void
}

export interface LastListStoreDeps {
  storage: Pick<Storage, "getItem" | "setItem">
}

/** LAST_LIST_STORAGE_KEY is where the memory is kept. */
export const LAST_LIST_STORAGE_KEY = "nooks.last-list"

export function createLastListStore(deps: LastListStoreDeps): LastListStore {
  let uid = read(deps.storage)
  const listeners = new Set<() => void>()

  return {
    subscribe(listener) {
      listeners.add(listener)
      return () => listeners.delete(listener)
    },
    getUid: () => uid,
    remember(next) {
      if (next === uid || !next) {
        return
      }
      uid = next
      try {
        deps.storage.setItem(LAST_LIST_STORAGE_KEY, next)
      } catch {
        // A Member in a private window still gets it for this session.
      }
      for (const listener of listeners) {
        listener()
      }
    },
  }
}

/** read returns the remembered List, or an empty string. */
function read(storage: Pick<Storage, "getItem">): string {
  try {
    return storage.getItem(LAST_LIST_STORAGE_KEY) ?? ""
  } catch {
    return ""
  }
}

export const lastListStore: LastListStore =
  typeof window === "undefined"
    ? { subscribe: () => () => {}, getUid: () => "", remember: () => {} }
    : createLastListStore({ storage: window.localStorage })
