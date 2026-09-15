import { afterEach, beforeAll, beforeEach, describe, expect, it } from "vitest"
import { render } from "@testing-library/react"
import type { Item } from "@nooks/api"

import { readyForEnglish } from "@/test/i18n"
import { PrintSheet } from "./print-sheet"

beforeAll(readyForEnglish)

/** The host the sheet portals into, standing in for the one in index.html. */
let printRoot: HTMLElement

beforeEach(() => {
  printRoot = document.createElement("div")
  printRoot.id = "print-root"
  document.body.append(printRoot)
})

afterEach(() => printRoot.remove())

/** someItems builds a List of a given length. */
function someItems(count: number): Item[] {
  return Array.from({ length: count }, (_, index) => ({
    uid: `item-${index}`,
    label: `Thing ${index}`,
    quantity: "",
    done: false,
    addedByName: "Anna",
  })) as Item[]
}

/**
 * show renders the sheet for a List, and answers with where it landed.
 *
 * Not the render container: the sheet portals itself out of the app's root, which is
 * the whole point of it — printing hides that root.
 */
function show(items: Item[]) {
  render(
    <PrintSheet
      instanceName="Brunnen Street"
      listName="Groceries"
      printedOn="Tuesday, 25 August"
      items={items}
    />,
  )
  return printRoot
}

describe("the printed sheet", () => {
  // Printing hides the app's root. A sheet inside it is hidden too, and the page comes
  // out blank.
  it("renders outside the app's root", () => {
    const { container } = render(
      <PrintSheet
        instanceName="Brunnen Street"
        listName="Groceries"
        printedOn="Tuesday, 25 August"
        items={someItems(1)}
      />,
    )

    expect(container.querySelector(".nooks-print-sheet")).toBeNull()
    expect(document.querySelector(".nooks-print-sheet")).not.toBeNull()
  })

  // A sheet is a snapshot of the List: somebody who ticked something on the way out
  // still wants to see that they did.
  it("prints every Item, ticked ones included", () => {
    const items = someItems(2)
    items[0] = { ...items[0], done: true } as Item

    const container = show(items)

    expect(container.querySelectorAll(".nooks-print-item")).toHaveLength(2 + 3)
    expect(container.querySelector(".nooks-print-box[data-done]")).not.toBeNull()
  })

  // For whatever gets remembered in the shop.
  it("ends with three blank rows", () => {
    expect(show(someItems(1)).querySelectorAll(".nooks-print-blank")).toHaveLength(3)
  })

  // DESIGN.md §15 overrules the frames here. Two columns fit more on a sheet and are
  // harder to read walking around a shop holding it, which is the only place this sheet
  // is ever used.
  it("stays one column however long the list is", () => {
    expect(show(someItems(40)).querySelector("[data-two-up]")).toBeNull()
  })

  it("names the Instance and the day it was printed", () => {
    const container = show(someItems(1))

    expect(container.textContent).toContain("Brunnen Street")
    expect(container.textContent).toContain("Tuesday, 25 August")
  })
})

describe("what the sheet says it is", () => {
  it("never calls a list something it is not", () => {
    // The eyebrow used to read "<instance> · shopping list" for every list, so a
    // packing list printed a header saying it was shopping. The design writes the kind
    // of list there, and Nooks has no field for one — so it says nothing rather than
    // guessing. The list's own name is the heading directly below it.
    render(
      <PrintSheet
        instanceName="Brunnen Street"
        listName="Packing — Norway"
        printedOn="Tue 25 August"
        items={[]}
      />,
    )

    expect(printRoot.textContent).toContain("Packing — Norway")
    expect(printRoot.textContent).not.toMatch(/shopping/i)
  })
})
