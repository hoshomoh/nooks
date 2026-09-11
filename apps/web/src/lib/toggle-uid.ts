/**
 * toggled adds or removes one identifier from a picked set.
 *
 * Every picker in the app does this and nothing else to its list, so it is written once
 * and the pickers stay about what they are picking.
 */
export function toggled(chosen: string[], uid: string): string[] {
  return chosen.includes(uid) ? chosen.filter((each) => each !== uid) : [...chosen, uid]
}
