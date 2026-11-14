import { beforeAll, describe, expect, it } from "vitest"
import { Code, ConnectError } from "@connectrpc/connect"
import { QueryClientProvider } from "@tanstack/react-query"
import {
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router"
import { render, screen } from "@testing-library/react"
import type { Member } from "@nooks/api"

import { currentMemberQuery } from "@/lib/queries"
import { createQueryClient } from "@/lib/query-client"
import { readyForEnglish } from "@/test/i18n"
import { RouteError } from "./route-error"

beforeAll(readyForEnglish)

/*
A session that ended while this browser still believed in it.

Who is signed in is kept with everything else so a List opens offline, so a reload put
the old answer back and the page let the Member in. Every read after that was refused,
and the screen said the server did not answer, which it had, with no way to sign in.
*/
describe("a session that ended under a remembered sign-in", () => {
  function signedInOnceThen(refusal: ConnectError) {
    const root = createRootRoute()
    const lists = createRoute({
      getParentRoute: () => root,
      path: "/",
      loader: async ({ context }) => {
        const { queryClient } = context as { queryClient: ReturnType<typeof createQueryClient> }
        const member = await queryClient.ensureQueryData(currentMemberQuery)
        if (!member) {
          throw new Error("the remembered sign-in should have let this page load")
        }
        await queryClient.ensureQueryData({ queryKey: ["sidebar"], queryFn: () => Promise.reject(refusal) })
      },
      component: () => <p>All lists</p>,
    })
    const signIn = createRoute({ getParentRoute: () => root, path: "/sign-in", component: () => <p>Sign in</p> })

    const queryClient = createQueryClient(() => void router.navigate({ to: "/sign-in" }))
    // What a reload restores from disk: somebody was signed in a day ago.
    queryClient.setQueryData(currentMemberQuery.queryKey, { uid: "m1", name: "Ada" } as Member)

    const router = createRouter({
      routeTree: root.addChildren([lists, signIn]),
      history: createMemoryHistory({ initialEntries: ["/"] }),
      context: { queryClient },
      defaultErrorComponent: RouteError,
    })
    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    )
    return queryClient
  }

  it("goes to sign in, rather than saying the server did not answer", async () => {
    const queryClient = signedInOnceThen(new ConnectError("sign in first", Code.Unauthenticated))

    expect(await screen.findByText("Sign in")).toBeInTheDocument()
    expect(screen.queryByText("That didn't load")).not.toBeInTheDocument()
    expect(queryClient.getQueryData(currentMemberQuery.queryKey)).toBeNull()
  })

  // A server that really did not answer is still said to have not answered.
  it("still says so when the server did not answer", async () => {
    signedInOnceThen(new ConnectError("down", Code.Unavailable))

    expect(await screen.findByText("That didn't load", {}, { timeout: 10_000 })).toBeInTheDocument()
  })
})
