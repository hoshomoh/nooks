import type { Item } from "@nooks/api"

/**
 * Changes made while the Instance was out of reach.
 *
 * A change made offline is paused rather than failed — see online-wiring.ts — which
 * means it is still going to be sent. The row it affects says so, because a tick that
 * looks identical to a saved one is the app quietly claiming something it has not done.
 */

/** The key a tick is queued under, so a paused one can be found again. */
export const TICK_MUTATION = ["item", "set-done"] as const

/** The key an added Item is queued under. */
export const ADD_MUTATION = ["item", "create"] as const

/** One paused change, as the mutation cache describes it. */
export interface QueuedMutation {
  isPaused: boolean
  variables: unknown
}

/** What is still waiting to be sent, by what it affects. */
export interface QueuedChanges {
  /** Items whose tick has not reached the Instance. */
  itemUids: ReadonlySet<string>
  /** Lists with an Item on them that has never been sent. */
  listUids: ReadonlySet<string>
}

/** NOTHING_QUEUED is the ordinary case: everything the Member did has landed. */
export const NOTHING_QUEUED: QueuedChanges = { itemUids: new Set(), listUids: new Set() }

/**
 * queuedChanges reads what is waiting out of the mutation cache's own record.
 *
 * Pure, and given plain objects rather than the cache: what counts as waiting is a
 * rule worth testing without standing up a QueryClient to do it.
 */
export function queuedChanges(mutations: readonly QueuedMutation[]): QueuedChanges {
  const itemUids = new Set<string>()
  const listUids = new Set<string>()

  for (const mutation of mutations) {
    if (!mutation.isPaused) {
      continue
    }
    const target = mutation.variables
    if (!isRecord(target)) {
      continue
    }
    if (typeof target.itemUid === "string") {
      itemUids.add(target.itemUid)
    }
    if (typeof target.listUid === "string") {
      listUids.add(target.listUid)
    }
  }

  return { itemUids, listUids }
}

/**
 * withTick flips one Item wherever a cached answer holds it.
 *
 * The same Item appears on its List, in Today, in Upcoming and in the calendar, and a
 * tick made offline has to show in all of them — otherwise the Member ticks it once and
 * watches it come back when they change screen.
 *
 * Both shapes are handled here because both are the same Item: a List answer holds
 * Items directly, a dated answer wraps each one with the List it came from.
 */
export function withTick<T>(data: T, itemUid: string, done: boolean): T {
  if (!isRecord(data) || !Array.isArray(data.items)) {
    return data
  }
  return { ...data, items: data.items.map((entry) => tickEntry(entry, itemUid, done)) } as T
}

/** tickEntry flips an Item, or the Item inside a dated entry, or leaves it alone. */
function tickEntry(entry: unknown, itemUid: string, done: boolean): unknown {
  if (!isRecord(entry)) {
    return entry
  }
  if (entry.uid === itemUid) {
    return { ...entry, done }
  }
  if (isRecord(entry.item) && entry.item.uid === itemUid) {
    return { ...entry, item: { ...entry.item, done } }
  }
  return entry
}

/**
 * provisionalItem is what an Item looks like before the Instance has seen it.
 *
 * It carries a uid of its own so React can key it and the Member can tick it, and that
 * uid is never sent anywhere: the refresh after the change lands replaces it with the
 * real Item.
 */
export function provisionalItem(fields: Pick<Item, "label" | "quantity" | "dueOn">): Item {
  return {
    $typeName: "nooks.api.v1.Item",
    uid: `pending-${crypto.randomUUID()}`,
    label: fields.label,
    quantity: fields.quantity,
    dueOn: fields.dueOn,
    done: false,
    addedByName: "",
    doneByName: "",
    doneAt: "",
    doneByUid: "",
    note: "",
    noteFirstLine: "",
    noteRemainingLines: 0,
    addedViaToken: "",
  }
}

/** PROVISIONAL_PREFIX marks a uid the Instance has never heard of. */
export const PROVISIONAL_PREFIX = "pending-"

/** isProvisional reports whether an Item is one this browser invented. */
export function isProvisional(itemUid: string): boolean {
  return itemUid.startsWith(PROVISIONAL_PREFIX)
}

/** withItem puts a new Item at the end of a cached List answer. */
export function withItem<T>(data: T, item: Item): T {
  if (!isRecord(data) || !Array.isArray(data.items)) {
    return data
  }
  return { ...data, items: [...data.items, item] } as T
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null
}
