/** @vitest-environment jsdom */
import { beforeAll, describe, expect, it, vi } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { RouterProvider, createRootRoute, createRouter } from "@tanstack/react-router"
import { render, screen } from "@testing-library/react"
import { Sharing, type List } from "@nooks/api"

import "@/test/dom"
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

/** show renders the sidebar inside a router, since every row is a link. */
function show(lists: List[] = [groceries]) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const router = createRouter({
    routeTree: createRootRoute({
      component: () => (
        <Sidebar
          instanceName="Brunnen Street"
          memberName="Anna"
          lists={lists}
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
