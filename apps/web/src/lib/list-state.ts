import type { List } from "@nooks/api"

/**
 * isArchived reports whether a List has been put out of the sidebar.
 *
 * Archiving is not deleting: an archived List keeps its Items, is still searchable, and
 * is found under its own filter in All lists, which is where it is restored from.
 */
export function isArchived(list: List): boolean {
  return Boolean(list.archivedAt)
}
