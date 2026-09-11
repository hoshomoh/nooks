/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { render, screen } from "@testing-library/react"
import { Sharing, type List } from "@nooks/api"

import "@/test/dom"
import { readyForEnglish } from "@/test/i18n"
import { DatedAddRow } from "./dated-add-row"
import { lastListStore } from "@/lib/last-list-store"
import { shift, today, toStored } from "@/lib/dates"

beforeAll(readyForEnglish)

/** list builds one of the Member's Lists. */
function list(uid: string, name: string): List {
  return {
    uid,
    name,
    sharing: Sharing.PRIVATE,
    canEdit: true,
    isOwner: true,
    isPinned: false,
    openCount: 0,
  } as List
}

/** show renders the row for a view whose Items are due on a given day. */
function show(lists: List[], defaultDue: string) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <DatedAddRow lists={lists} defaultDue={defaultDue} />
    </QueryClientProvider>,
  )
}

describe("adding from a view that gathers every List", () => {
  // A default that is not stated is a default nobody can trust.
  it("says where the Item will land and when it is due", () => {
    show([list("list_groceries", "Groceries")], toStored(today()))

    expect(screen.getByRole("textbox", { name: "Add to Groceries, due today" })).toBeInTheDocument()
  })

  it("lands on the List last added to", () => {
    lastListStore.remember("list_bike")
    show([list("list_groceries", "Groceries"), list("list_bike", "Bike")], toStored(today()))

    expect(screen.getByRole("textbox", { name: /Add to Bike/ })).toBeInTheDocument()
  })

  // A List that was deleted or unshared cannot be where the next Item goes.
  it("falls back to a List the Member still has", () => {
    lastListStore.remember("list_gone")
    show([list("list_groceries", "Groceries")], toStored(today()))

    expect(screen.getByRole("textbox", { name: /Add to Groceries/ })).toBeInTheDocument()
  })

  it("carries the view's date, not the day itself", () => {
    lastListStore.remember("list_groceries")
    show([list("list_groceries", "Groceries")], toStored(shift(today(), 1)))

    expect(screen.getByRole("textbox", { name: /due tomorrow/ })).toBeInTheDocument()
  })

  // A row that swallows an Item is worse than no row.
  it("does not appear when there is nowhere to put anything", () => {
    const { container } = show([], toStored(today()))

    expect(container).toBeEmptyDOMElement()
  })
})
