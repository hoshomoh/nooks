/** @vitest-environment jsdom */
import { afterEach, beforeAll, describe, expect, it } from "vitest"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"

import "@/test/dom"
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

    expect(screen.queryByRole("status")).not.toBeInTheDocument()
  })

  it("appears when a request does not arrive, with a way to ask again", async () => {
    connectionStore.markUnreachable()
    show()

    expect(screen.getByRole("status")).toHaveTextContent(/offline/i)
    expect(screen.getByRole("button", { name: /retry/i })).toBeInTheDocument()
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

    await expect.poll(() => screen.queryByRole("status")).toBeNull()
  })

  it("is a strip rather than a barrier, so the screen underneath still works", async () => {
    connectionStore.markUnreachable()
    show()

    // Retry asks again; it does not block, confirm, or take the Member anywhere.
    await userEvent.click(screen.getByRole("button", { name: /retry/i }))

    expect(screen.getByRole("status")).toBeInTheDocument()
  })
})
