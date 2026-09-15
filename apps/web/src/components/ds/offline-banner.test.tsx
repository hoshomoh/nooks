import { afterEach, beforeAll, describe, expect, it } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import { readyForEnglish } from "@/test/i18n"
import { OfflineBanner } from "./offline-banner"
import { connectionStore } from "@/lib/connection-store"

beforeAll(readyForEnglish)

// The store is the application's own, so a test that puts it offline must put it back.
afterEach(() => connectionStore.markSeen())

function show() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <OfflineBanner />
    </QueryClientProvider>,
  )
}

describe("the offline banner", () => {
  it("says nothing while the Instance is answering", () => {
    connectionStore.markSeen()
    show()

    // The strip is in the page so that it can open by pushing rather than by arriving,
    // and so that the live region is there to change. Closed, it says nothing and the
    // keyboard cannot reach into it.
    expect(screen.getByRole("status")).toHaveTextContent("")
    expect(screen.queryByRole("button", { name: /retry/i })?.closest("[inert]")).not.toBeNull()
  })

  it("appears when a request does not arrive, with a way to ask again", async () => {
    connectionStore.markUnreachable()
    show()

    expect(screen.getByRole("status")).toHaveTextContent(/offline/i)
    expect(screen.getByRole("button", { name: /retry/i })).toBeInTheDocument()
  })

  it("opens by pushing the screen down rather than by appearing in front of it", () => {
    connectionStore.markUnreachable()
    const { container } = show()

    // The height is what travels. A strip that is simply put there moves everything
    // below it by its own height between two frames, under whatever the Member was
    // reading or typing at the time.
    const strip = container.firstElementChild
    expect(strip).toHaveClass("transition-[grid-template-rows]")
    expect(strip).toHaveClass("grid-rows-[1fr]")
  })

  it("says when contact was last good", () => {
    connectionStore.markSeen()
    connectionStore.markUnreachable()
    show()

    // Last seen today, so the moment reads as a clock time.
    expect(screen.getByRole("status")).toHaveTextContent(/\d\d:\d\d/)
  })

  it("goes away by itself once the Instance answers again", async () => {
    connectionStore.markUnreachable()
    show()
    expect(screen.getByRole("status")).toBeInTheDocument()

    connectionStore.markSeen()

    await expect.poll(() => screen.getByRole("status").textContent).toBe("")
  })

  it("is a strip rather than a barrier, so the screen underneath still works", async () => {
    connectionStore.markUnreachable()
    show()

    // Retry asks again; it does not block, confirm, or take the Member anywhere.
    await userEvent.click(screen.getByRole("button", { name: /retry/i }))

    expect(screen.getByRole("status")).toHaveTextContent(/offline/i)
  })
})
