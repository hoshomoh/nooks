import type { PublicListSettings } from "@nooks/api"

import { instanceClient, listClient } from "./api"
import { shift, today, toStored } from "./dates"
import type { Translate } from "./translate"

/**
 * The List a new Instance opens onto.
 *
 * An empty Instance is a fair description of the truth and a poor way to meet a
 * product: a Member who has never seen Nooks cannot tell that an Item can carry a
 * quantity, a date or a Note, because nothing on screen has one. So the first List
 * shows each of them once, on things somebody might actually buy.
 *
 * Written from the browser rather than the server, because the words have to be in the
 * Member's language and only the browser knows which that is. It is one round of calls,
 * once, on the day an Instance is created.
 */
export async function createStarterList(t: Translate, now: Date = new Date()): Promise<void> {
  const created = await listClient.createList({ name: t("starter.listName") })
  const listUid = created.list?.uid
  if (!listUid) {
    return
  }

  for (const item of starterItems(t, now)) {
    await listClient.createItem({ listUid, ...item })
  }

  await publish(listUid)
}

/**
 * Publishes the first List, so the Instance's address answers with something.
 *
 * An Instance whose front door asks for credentials is one nobody can be shown. The
 * first List is also the least private thing on it — it was made by Nooks, not by
 * anybody — so it is the one that can be public before an Admin has decided anything.
 *
 * Contributor names stay off, as they do everywhere by default. Settings → Public list
 * is where that is changed, or the page taken down.
 */
async function publish(listUid: string): Promise<void> {
  const current = await instanceClient.getInstanceSettings({})
  if (!current.settings) {
    return
  }

  const publicList: PublicListSettings = {
    ...(current.settings.publicList as PublicListSettings),
    listUid,
    showNames: false,
    showMeta: true,
    allowJoin: true,
  }

  await instanceClient.updateInstanceSettings({
    settings: { ...current.settings, publicList },
  })
}

/** One line of the starter List. */
interface StarterItem {
  label: string
  quantity: string
  dueOn: string
  note: string
}

/**
 * starterItems is what the first List holds.
 *
 * Between them they show a quantity, a date, and a Note with every block type in it —
 * each feature once, and nothing that looks like filler. Pure, so what a new Instance
 * opens onto can be read here rather than run.
 */
export function starterItems(t: Translate, now: Date): StarterItem[] {
  const plain = { quantity: "", dueOn: "", note: "" }
  return [
    { ...plain, label: t("starter.washingUp") },
    { ...plain, label: t("starter.milk"), quantity: t("starter.milkQuantity") },
    {
      ...plain,
      label: t("starter.tomatoes"),
      quantity: t("starter.tomatoesQuantity"),
      dueOn: toStored(shift(today(now), 2)),
    },
    { ...plain, label: t("starter.coffee"), note: t("starter.coffeeNote") },
    { ...plain, label: t("starter.bakingPaper") },
  ]
}
