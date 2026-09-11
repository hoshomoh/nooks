/** @vitest-environment jsdom */
import { afterEach, beforeAll, beforeEach, describe, expect, it } from "vitest"
import { render } from "@testing-library/react"
import type { Item } from "@nooks/api"

import "@/test/dom"
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

  it("stays one column while it fits on one", () => {
    expect(show(someItems(4)).querySelector("[data-two-up]")).toBeNull()
  })

  // Past a page, two-up rather than spilling onto a second sheet to carry as well.
  it("goes two-up once it would not", () => {
    expect(show(someItems(30)).querySelector("[data-two-up]")).not.toBeNull()
  })

  it("names the Instance and the day it was printed", () => {
    const container = show(someItems(1))

    expect(container.textContent).toContain("Brunnen Street")
    expect(container.textContent).toContain("Tuesday, 25 August")
  })
})
