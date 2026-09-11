import { describe, expect, it } from "vitest"
import type { Item, List } from "@nooks/api"

import { exportFileFor, listAsText } from "./list-export"

const groceries = { uid: "l", name: "Groceries" } as List

/** item builds one line of a List. */
function item(label: string, extra: Partial<Item> = {}): Item {
  return { uid: label, label, quantity: "", done: false, ...extra } as Item
}

describe("a List as plain text", () => {
  // Export is not a format of its own: it is the markdown a Note is written in.
  it("writes a checklist anything else can read", () => {
    expect(listAsText(groceries, [item("Milk"), item("Oats", { done: true })])).toBe(
      "# Groceries\n\n- [ ] Milk\n- [x] Oats\n",
    )
  })

  it("keeps a quantity with the thing it counts", () => {
    expect(listAsText(groceries, [item("Tomatoes", { quantity: "1 kg" })])).toContain(
      "- [ ] Tomatoes — 1 kg",
    )
  })

  it("adds nothing that was not on the List", () => {
    expect(listAsText(groceries, [])).toBe("# Groceries\n\n")
  })
})

describe("the file it is saved as", () => {
  it("is named after the List", () => {
    expect(exportFileFor(groceries, []).name).toBe("groceries.md")
  })

  it("makes a name with punctuation in it safe to be a filename", () => {
    expect(exportFileFor({ name: "Packing — Norway" } as List, []).name).toBe(
      "packing-norway.md",
    )
  })

  it("keeps a name that is not written in Latin letters", () => {
    expect(exportFileFor({ name: "買い物" } as List, []).name).toBe("買い物.md")
  })

  it("falls back rather than making a file with no name", () => {
    expect(exportFileFor({ name: "···" } as List, []).name).toBe("list.md")
  })
})
