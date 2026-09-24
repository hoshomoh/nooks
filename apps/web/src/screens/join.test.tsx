import { beforeAll, describe, expect, it } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { RouterProvider, createRootRoute, createRouter } from "@tanstack/react-router"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { RequestStatus } from "@nooks/api"

import { readyForEnglish } from "@/test/i18n"
import { Join } from "./join"

beforeAll(readyForEnglish)

const REMEMBERED = "nooks.join-request"

/** waiting renders the screen as somebody who asked and has not been answered. */
function waiting(uid: string) {
  window.localStorage.clear()
  window.localStorage.setItem(REMEMBERED, uid)

  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  queryClient.setQueryData(["join-request", uid], { status: RequestStatus.PENDING })

  const router = createRouter({ routeTree: createRootRoute({ component: Join }) })
  return render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  )
}

/**
 * Asking again ends the request that was asked before.
 *
 * nooks sends no mail, so the identifier this browser remembers is the whole of how
 * somebody comes back to a request — and, once an Admin approves, the thing that turns
 * it into an account. It is kept on disk because an answer can take days.
 *
 * It used to be cleared from the screen and left on disk, so the next visit read it back
 * and returned somebody who had said they were starting over to the request they had
 * finished with. On a browser more than one person uses, it stayed live there too.
 */
describe("asking to join again", () => {
  it("shows the form rather than the old request", async () => {
    waiting("req_one")
    await userEvent.click(await screen.findByRole("button", { name: /ask again/i }))

    expect(await screen.findByLabelText(/email/i)).toBeInTheDocument()
  })

  it("forgets the request the browser was holding", async () => {
    waiting("req_one")
    await userEvent.click(await screen.findByRole("button", { name: /ask again/i }))

    expect(window.localStorage.getItem(REMEMBERED)).toBeNull()
  })
})
