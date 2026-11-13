import { afterEach, beforeAll, describe, expect, it, vi } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import {
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router"
import { render, screen } from "@testing-library/react"
import type { GetInstanceResponse, GetListResponse, GetSidebarResponse, Item, Member } from "@nooks/api"

import { readyForEnglish } from "@/test/i18n"
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
function groceries(): GetListResponse {
  return { list: { uid: "l1", name: "Groceries" }, items: [{ uid: "i1", label: "Milk" } as Item] } as GetListResponse
}

/** show opens a List screen with everything it reads already in the cache. */
function show(live: LiveState): void {
  vi.spyOn(liveStore, "getState").mockReturnValue(live)

  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  // The generated message types carry a $typeName the cache does not need for a read.
  queryClient.setQueryData(instanceQuery.queryKey, { name: "Brunnen Street" } as GetInstanceResponse)
  queryClient.setQueryData(currentMemberQuery.queryKey, { uid: "m1", name: "Ada" } as Member)
  queryClient.setQueryData(sidebarQuery.queryKey, {} as GetSidebarResponse)
  queryClient.setQueryData(listQuery("l1").queryKey, groceries())

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
}

describe("the list screen", () => {
  // Whoever has a Note open is read from the stream before anything else on the screen,
  // so the first List anybody opened while somebody else was writing is where it broke.
  it("opens while somebody else is writing in a Note", async () => {
    show({ ...NOBODY_ELSE, editing: [{ itemUid: "i1", name: "Grace" }] })

    expect(await screen.findAllByText("Milk")).not.toHaveLength(0)
  })
})
