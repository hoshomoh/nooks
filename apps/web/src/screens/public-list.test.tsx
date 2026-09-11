/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { RouterProvider, createRootRoute, createRouter } from "@tanstack/react-router"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import type { GetPublicListResponse, PublicItem } from "@nooks/api"

import "@/test/dom"
import { readyForEnglish } from "@/test/i18n"
import { PublicListScreen } from "./public-list"
import { publicListQuery } from "@/lib/public-queries"

beforeAll(readyForEnglish)

/** show renders the page with an answer already in the cache. */
function show(page: Partial<GetPublicListResponse>) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  // The generated message type carries a $typeName the cache does not need for a
  // read; the page only ever reads fields.
  queryClient.setQueryData(publicListQuery.queryKey, {
    published: false,
    instanceName: "",
    listName: "",
    items: [],
    allowJoin: false,
    ...page,
  } as GetPublicListResponse)

  const router = createRouter({
    routeTree: createRootRoute({ component: PublicListScreen }),
  })

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

/** groceries is a published List with one thing on it. */
const groceries: Partial<GetPublicListResponse> = {
  published: true,
  instanceName: "Brunnen Street",
  listName: "Groceries",
  allowJoin: true,
  items: [{ label: "Milk", done: false, quantity: "2", dueOn: "", addedByName: "" }] as PublicItem[],
}

describe("the public list", () => {
  it("shows the List it was sent for", async () => {
    show(groceries)

    expect(await screen.findByRole("heading", { name: "Groceries" })).toBeInTheDocument()
    expect(screen.getByText("Milk")).toBeInTheDocument()
  })

  // A Visitor is not a Member with fewer buttons.
  it("offers no way to reach anything else on the Instance", async () => {
    show(groceries)

    await screen.findByRole("heading", { name: "Groceries" })
    expect(screen.queryByRole("button", { name: "Search" })).not.toBeInTheDocument()
    expect(screen.queryByText("All lists")).not.toBeInTheDocument()
  })

  // Under the row they touched, not as a wall: it answers the thing they just tried,
  // and it names the thing rather than talking about ticking in general.
  it("asks nothing until a Visitor reaches for a row, then names what they reached for", async () => {
    show(groceries)

    await screen.findByRole("heading", { name: "Groceries" })
    expect(screen.queryByText(/Sign in to tick/)).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole("button", { name: "Milk" }))
    expect(screen.getByText("Sign in to tick “Milk” off")).toBeInTheDocument()
  })

  // Reaching for a row files the tick, so the sentence is a statement about what has
  // already happened rather than a promise about signing in.
  it("says the tick is kept", async () => {
    show(groceries)

    await screen.findByRole("heading", { name: "Groceries" })
    await userEvent.click(screen.getByRole("button", { name: "Milk" }))

    expect(screen.getByText(/remembered and applied/)).toBeInTheDocument()
  })

  it("offers to ask for an account only when the Instance allows it", async () => {
    show({ ...groceries, allowJoin: false })

    await screen.findByRole("heading", { name: "Groceries" })
    await userEvent.click(screen.getByRole("button", { name: "Milk" }))

    // Signing in is always offered; asking for an account is the Instance's decision.
    expect(screen.getAllByRole("link", { name: "Sign in" }).length).toBeGreaterThan(0)
    expect(screen.queryByRole("link", { name: /Ask to join/ })).not.toBeInTheDocument()
  })

  it("leads with how much is left, and how fresh the page is", async () => {
    show({ ...groceries, openCount: 6, updatedAt: new Date().toISOString() })

    expect(await screen.findByText("6 open")).toBeInTheDocument()
    expect(screen.getByText(/^updated /)).toBeInTheDocument()
    expect(screen.getByText("read-only for visitors")).toBeInTheDocument()
  })

  // An Instance with nothing published must not hint at what it holds.
  it("says nothing is published, and names nothing", async () => {
    show({ published: false })

    expect(await screen.findByText("Nothing is published here")).toBeInTheDocument()
  })

  it("explains the page rather than asking a Visitor to do work", async () => {
    show({ ...groceries, items: [] })

    expect(await screen.findByText("Nothing on the list")).toBeInTheDocument()
  })
})
