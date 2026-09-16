import type { GetSidebarResponse, List } from "@nooks/api"

/** ordered is the four groups in the order the column draws them. */
function ordered(groups: GetSidebarResponse | undefined) {
  return [groups?.pinned, groups?.mine, groups?.shared, groups?.completed]
}

/**
 * sidebarLists is the four groups read as one set.
 *
 * What is to hand rather than everything a Member can reach: the groups are capped by
 * the server, so anywhere this is the answer has a bounded number of rows to draw
 * whether the Instance holds ten Lists or ten thousand.
 */
export function sidebarLists(groups: GetSidebarResponse | undefined): List[] {
  return ordered(groups).flatMap((group) => group?.lists ?? [])
}

/**
 * reachableCount is the number beside All lists.
 *
 * The four groups are cut so that no List is in two of them, and archiving is what takes
 * one out of all four — which is exactly what All lists shows. So the totals add up.
 */
export function reachableCount(groups: GetSidebarResponse | undefined): number {
  return ordered(groups).reduce((count, group) => count + (group?.total ?? 0), 0)
}
