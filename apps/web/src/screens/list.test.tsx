import { afterEach, beforeAll, describe, expect, it, vi } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import {
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router"
import { render, screen, waitFor, within } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import type {
  GetInstanceResponse,
  GetListResponse,
  GetSidebarResponse,
  Item,
  Member,
  MoveItemResponse,
} from "@nooks/api"

import { readyForEnglish } from "@/test/i18n"
import { listClient } from "@/lib/api"
import { liveStore, type LiveState } from "@/lib/live-store"
import { listQuery, sidebarQuery } from "@/lib/list-queries"
import { currentMemberQuery, instanceQuery } from "@/lib/queries"
import { listRoute } from "@/routes/list"
import { ListScreen } from "./list"

beforeAll(readyForEnglish)

afterEach(() => {
  vi.restoreAllMocks()
})

/** NOBODY_ELSE is a stream with nobody else on this List. */
const NOBODY_ELSE: LiveState = { watchers: [], editing: [], version: 1 }

/** groceries is a List with one Item on it. */
function groceries(items: Item[] = [{ uid: "i1", label: "Milk" } as Item]): GetListResponse {
  return { list: { uid: "l1", name: "Groceries" }, items } as GetListResponse
}

/** tickedByGrace is Milk, ticked by somebody else at the moment given. */
function tickedByGrace(at: Date): Item {
  return { uid: "i1", label: "Milk", done: true, doneByUid: "m2", doneAt: at.toISOString() } as Item
}

/** openDone opens the section ticked Items move into, which starts closed. */
async function openDone(): Promise<void> {
  await userEvent.click(await screen.findByRole("button", { name: /done/i, expanded: false }))
}

/** show opens a List screen with everything it reads already in the cache. */
function show(live: LiveState, opened: GetListResponse = groceries()): QueryClient {
  vi.spyOn(liveStore, "getState").mockReturnValue(live)

  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  // The generated message types carry a $typeName the cache does not need for a read.
  queryClient.setQueryData(instanceQuery.queryKey, { name: "Brunnen Street" } as GetInstanceResponse)
  queryClient.setQueryData(currentMemberQuery.queryKey, { uid: "m1", name: "Ada" } as Member)
  queryClient.setQueryData(sidebarQuery.queryKey, {} as GetSidebarResponse)
  queryClient.setQueryData(listQuery("l1").queryKey, opened)

  // The List's own path and search, without the loader and the lazy import the app's
  // route adds: the cache already holds what the loader would have fetched.
  const root = createRootRoute()
  const list = createRoute({
    getParentRoute: () => root,
    path: "/lists/$listUid",
    validateSearch: listRoute.options.validateSearch,
    component: ListScreen,
  })
  const router = createRouter({
    routeTree: root.addChildren([list]),
    history: createMemoryHistory({ initialEntries: ["/lists/l1"] }),
  })

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
  return queryClient
}

describe("the list screen", () => {
  // Whoever has a Note open is read from the stream before anything else on the screen,
  // so the first List anybody opened while somebody else was writing is where it broke.
  it("opens while somebody else is writing in a Note", async () => {
    show({ ...NOBODY_ELSE, editing: [{ itemUid: "i1", name: "Grace" }] })

    expect(await screen.findAllByText("Milk")).not.toHaveLength(0)
  })

  // Coming back to a List paints the copy already in the cache, which was fetched when
  // the tick was new. Showing it as just landed again would be a tick from minutes ago.
  it("does not show a tick as just landed again when coming back to the List", async () => {
    show(NOBODY_ELSE, groceries([tickedByGrace(new Date())]))
    await openDone()

    // Once on the page and once on the printed sheet, so the row itself is drawn.
    await waitFor(() => expect(screen.getAllByText("Milk")).toHaveLength(2))
    expect(document.body.querySelector(".animate-settle")).toBeNull()
  })

  it("shows a tick somebody else makes while the List is open", async () => {
    const queryClient = show(NOBODY_ELSE)
    await screen.findAllByText("Milk")

    vi.spyOn(listClient, "getList").mockResolvedValue(groceries([tickedByGrace(new Date())]))
    await queryClient.refetchQueries({ queryKey: listQuery("l1").queryKey })
    await openDone()

    await waitFor(() => expect(document.body.querySelector(".animate-settle")).not.toBeNull())
  })
})

describe("moving an Item", () => {
  const milk = { uid: "i1", label: "Milk" } as Item
  const oats = { uid: "i2", label: "Oats" } as Item
  const eggs = { uid: "i3", label: "Eggs" } as Item

  /** onShelf is a List the Member may change, holding these Items in this order. */
  function onShelf(...items: Item[]): GetListResponse {
    return { list: { uid: "l1", name: "Groceries", canEdit: true }, items } as GetListResponse
  }

  /** rowOrder reads the open rows top to bottom, by the label each one carries. */
  function rowOrder(): string[] {
    return screen
      .getAllByRole("checkbox")
      .map((box) => box.closest<HTMLElement>(".group"))
      .map((row) => (row ? (within(row).getByRole("textbox") as HTMLInputElement).value : ""))
  }

  // The row lands where it was put the moment it is put there, not when the answer comes.
  it("moves the row before the Instance answers, and says where it went", async () => {
    let answer = () => {}
    const sent = vi.spyOn(listClient, "moveItem").mockReturnValue(
      new Promise((resolve) => {
        answer = () => resolve({} as MoveItemResponse)
      }),
    )
    vi.spyOn(listClient, "getList").mockResolvedValue(onShelf(oats, milk, eggs))
    show(NOBODY_ELSE, onShelf(milk, oats, eggs))
    await waitFor(() => expect(rowOrder()).toEqual(["Milk", "Oats", "Eggs"]))

    screen.getAllByRole("checkbox")[0]?.focus()
    await userEvent.keyboard("{Alt>}{ArrowDown}{/Alt}")

    expect(rowOrder()).toEqual(["Oats", "Milk", "Eggs"])
    expect(sent).toHaveBeenCalledWith({ itemUid: "i1", afterItemUid: "i2" })
    answer()
  })

  it("puts the row back when the Instance refuses the move", async () => {
    vi.spyOn(listClient, "moveItem").mockRejectedValue(new Error("refused"))
    // The refresh after it never lands, so only putting it back can restore the order.
    vi.spyOn(listClient, "getList").mockReturnValue(new Promise(() => {}))
    show(NOBODY_ELSE, onShelf(milk, oats, eggs))
    await waitFor(() => expect(rowOrder()).toEqual(["Milk", "Oats", "Eggs"]))

    screen.getAllByRole("checkbox")[1]?.focus()
    await userEvent.keyboard("{Alt>}{ArrowUp}{/Alt}")

    await waitFor(() => expect(rowOrder()).toEqual(["Milk", "Oats", "Eggs"]))
  })
})
