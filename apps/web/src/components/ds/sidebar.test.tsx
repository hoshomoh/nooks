import { beforeAll, describe, expect, it, vi } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { RouterProvider, createRootRoute, createRouter } from "@tanstack/react-router"
import { render, screen } from "@testing-library/react"
import { Sharing, type GetSidebarResponse, type List } from "@nooks/api"

import { readyForEnglish } from "@/test/i18n"
import { Sidebar } from "./sidebar"

beforeAll(readyForEnglish)

/** groceries is a List with something open on it. */
const groceries = {
  uid: "list_groceries",
  name: "Groceries",
  sharing: Sharing.PRIVATE,
  canEdit: true,
  isOwner: true,
  isPinned: false,
  openCount: 7,
} as List

/** mine puts Lists in My lists, saying how many there are altogether. */
function mine(lists: List[], total = lists.length): GetSidebarResponse {
  return { mine: { lists, total } } as GetSidebarResponse
}

/** show renders the sidebar inside a router, since every row is a link. */
function show(groups: GetSidebarResponse = mine([groceries])) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const router = createRouter({
    routeTree: createRootRoute({
      component: () => (
        <Sidebar
          instanceName="Brunnen Street"
          memberName="Anna"
          groups={groups}
          onSearch={vi.fn()}
          onAddList={vi.fn()}
        />
      ),
    }),
  })

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

describe("a List in the sidebar", () => {
  // The `···` sits above the row's link, and everything else in the row has to let the
  // click through — otherwise the name itself swallows it.
  it("is a link a Member can follow", async () => {
    show()

    const link = await screen.findByRole("link", { name: "Groceries" })
    expect(link).toHaveAttribute("href", "/lists/list_groceries")
  })

  it("keeps its count on screen", async () => {
    show()
    expect(await screen.findByText("7")).toBeInTheDocument()
  })

  // A trigger that is only mounted on hover takes with it the element its menu is
  // positioned against, and the menu jumps to the corner of the page.
  it("keeps the menu's trigger mounted whether or not the pointer is over it", async () => {
    show()
    expect(await screen.findByRole("button", { name: "More" })).toBeInTheDocument()
  })
})

describe("a group the server had to cut short", () => {
  // The sidebar draws a handful whatever a Member has. The row is what says the rest
  // are somewhere, and where.
  it("offers the rest in All lists", async () => {
    show(mine([groceries], 240))

    const seeAll = await screen.findByRole("link", { name: "See all 240" })
    expect(seeAll).toHaveAttribute("href", "/")
  })

  it("says nothing when the group is all there is", async () => {
    show(mine([groceries]))

    await screen.findByRole("link", { name: "Groceries" })
    expect(screen.queryByText(/See all/)).not.toBeInTheDocument()
  })
})
