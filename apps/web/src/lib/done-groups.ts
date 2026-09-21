import type { Item } from "@nooks/api"

import { parseMoment } from "./dates"

/**
 * The completed Items of a List, gathered by the day they were ticked.
 *
 * A count on its own cannot answer the question a Member actually has, which is which
 * three of the twenty were today. Grouping by day answers it, and says who ticked what
 * where that can still be said: a tick is somebody doing something, not a row changing
 * colour, and one made by somebody since removed carries the time alone rather than
 * somebody else's name.
 *
 * The most recent days are what a List opens with. A List a household has used for a
 * year has a year of ticks in it, and a section that grows without end is one nobody
 * opens twice — but the line saying how many are behind it opens them, so this takes
 * the number of days as an argument rather than deciding it.
 */
export interface DoneDay {
  /** The local calendar day, as yyyy-mm-dd. Stable to sort and to key on. */
  key: string
  /** Any moment within that day, for naming it. */
  at: Date
  /** What was ticked on it, most recent first. */
  items: Item[]
}

export interface DoneGrouping {
  days: DoneDay[]
  /** How many are behind the last line rather than in a day above it. */
  earlier: number
}

/** SHOWN_DAYS is how many days keep their own heading. */
export const SHOWN_DAYS = 2

/**
 * groupDone gathers ticked Items by day, newest first.
 *
 * An Item nobody can date counts towards `earlier` rather than inventing a day for it:
 * the line says how many are not shown, which is true either way.
 *
 * shownDays is an argument rather than a constant because the line that says how many
 * are left has to be able to reach them. Pass enough days and `earlier` is nothing.
 */
export function groupDone(items: readonly Item[], shownDays: number = SHOWN_DAYS): DoneGrouping {
  const byDay = new Map<string, { at: Date; items: Item[] }>()
  let undated = 0

  for (const item of items) {
    const at = parseMoment(item.doneAt)
    if (!at) {
      undated += 1
      continue
    }
    const key = dayKey(at)
    const day = byDay.get(key)
    if (day) {
      day.items.push(item)
    } else {
      byDay.set(key, { at, items: [item] })
    }
  }

  const days = [...byDay.entries()]
    .map(([key, day]) => ({ key, at: day.at, items: sortByTicked(day.items) }))
    .sort((a, b) => b.key.localeCompare(a.key))

  const shown = days.slice(0, shownDays)
  const hidden = days.slice(shownDays)
  const earlier = hidden.reduce((total, day) => total + day.items.length, undated)

  return { days: shown, earlier }
}

/** sortByTicked puts the most recent tick at the top of its day. */
function sortByTicked(items: Item[]): Item[] {
  return [...items].sort((a, b) => b.doneAt.localeCompare(a.doneAt))
}

/**
 * dayKey is the local calendar day.
 *
 * Built from the parts rather than sliced off an ISO string, because the stored moment
 * is UTC and a tick at 23:30 in Berlin belongs to the day the Member was having.
 */
function dayKey(at: Date): string {
  const month = String(at.getMonth() + 1).padStart(2, "0")
  const day = String(at.getDate()).padStart(2, "0")
  return `${at.getFullYear()}-${month}-${day}`
}

/**
 * doneAttribution is the key and values for the line under a ticked Item.
 *
 * `done_by_id` is ON DELETE SET NULL, so a tick made by somebody who has since been
 * removed comes back with nobody's name on it. Falling back to whoever added the Item
 * put their name against a tick they did not make — on a shared List, in front of the
 * household, and wrong. The time on its own is the most that can be said truthfully.
 */
export interface DoneAttribution {
  key: "list.doneBy" | "list.doneWhen"
  values: { name?: string; when: string }
}

export function doneAttribution(doneByName: string, when: string): DoneAttribution {
  if (!doneByName) {
    return { key: "list.doneWhen", values: { when } }
  }
  return { key: "list.doneBy", values: { name: doneByName, when } }
}
