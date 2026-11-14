import type { RowPerson } from "@/components/ds/list-row"
import type { Translate } from "./translate"

/**
 * addedBy is the circle an open Item's row ends in: who put it on the List.
 *
 * A token is somebody's access narrowed, never an identity of its own, so the circle is
 * the Member's and the token is said after their name, with a spark on the circle.
 */
export function addedBy(t: Translate, name: string, tokenName: string): RowPerson | undefined {
  if (!name) {
    return undefined
  }
  if (!tokenName) {
    return { name, said: name }
  }
  return { name, said: t("list.addedByVia", { name, token: tokenName }), viaToken: true }
}

/**
 * tickedBy is the circle a completed Item's row ends in: who ticked it.
 *
 * `done_by_id` is ON DELETE SET NULL, so a tick made by somebody who has since been
 * removed comes back with nobody's name on it. Falling back to whoever added the Item
 * would put their name against a tick they did not make, so such a row has no circle and
 * the time alone says what can be said truthfully.
 */
export function tickedBy(name: string): RowPerson | undefined {
  return name ? { name, said: name } : undefined
}
