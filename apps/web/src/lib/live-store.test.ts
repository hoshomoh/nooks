import { describe, expect, it, vi } from "vitest"

import { createLiveStore, type LiveConnection, type LiveEvent } from "./live-store"

/** fakeConnection is an EventSource that a test drives by hand. */
function fakeConnection() {
  let deliver: ((event: MessageEvent<string>) => void) | null = null
  const closed = vi.fn()
  const urls: string[] = []

  const connection: LiveConnection = {
    addEventListener: (_kind, listener) => {
      deliver = listener
    },
    close: closed,
  }

  return {
    urls,
    closed,
    connect: (url: string) => {
      urls.push(url)
      return connection
    },
    send: (event: LiveEvent) => {
      deliver?.({ data: JSON.stringify(event) } as MessageEvent<string>)
    },
    sendRaw: (data: string) => {
      deliver?.({ data } as MessageEvent<string>)
    },
  }
}

describe("the live stream", () => {
  it("follows the List being read", () => {
    const fake = fakeConnection()
    const store = createLiveStore({ connect: fake.connect, endpoint: "/events" })

    store.watch("list_groceries")

    expect(fake.urls).toEqual(["/events?list=list_groceries"])
  })

  it("does not reconnect when it is already where it should be", () => {
    const fake = fakeConnection()
    const store = createLiveStore({ connect: fake.connect, endpoint: "/events" })

    store.watch("list_groceries")
    store.watch("list_groceries")

    expect(fake.urls).toHaveLength(1)
  })

  it("holds who else is looking at that List", () => {
    const fake = fakeConnection()
    const store = createLiveStore({ connect: fake.connect, endpoint: "/events" })
    store.watch("list_groceries")

    fake.send({ kind: "presence", listUid: "list_groceries", watchers: ["Jonas"] })

    expect(store.getState().watchers).toEqual(["Jonas"])
  })

  // Presence is per List: somebody arriving on another one is not here.
  it("ignores presence for a List it is not reading", () => {
    const fake = fakeConnection()
    const store = createLiveStore({ connect: fake.connect, endpoint: "/events" })
    store.watch("list_groceries")

    fake.send({ kind: "presence", listUid: "list_bike", watchers: ["Jonas"] })

    expect(store.getState().watchers).toEqual([])
  })

  it("forgets who was there when it moves to another List", () => {
    const fake = fakeConnection()
    const store = createLiveStore({ connect: fake.connect, endpoint: "/events" })
    store.watch("list_groceries")
    fake.send({ kind: "presence", listUid: "list_groceries", watchers: ["Jonas"] })

    store.watch("list_bike")

    expect(store.getState().watchers).toEqual([])
    expect(fake.closed).toHaveBeenCalled()
  })

  it("tells whoever is listening what arrived", () => {
    const fake = fakeConnection()
    const store = createLiveStore({ connect: fake.connect, endpoint: "/events" })
    const heard = vi.fn()
    store.onEvent(heard)
    store.watch("list_groceries")

    fake.send({ kind: "list.changed", listUid: "list_groceries" })

    expect(heard).toHaveBeenCalledWith({ kind: "list.changed", listUid: "list_groceries" })
  })

  // The stream is a courtesy on top of data the app already has, and must never take
  // the app down.
  it("drops a message it cannot read", () => {
    const fake = fakeConnection()
    const store = createLiveStore({ connect: fake.connect, endpoint: "/events" })
    const heard = vi.fn()
    store.onEvent(heard)
    store.watch("list_groceries")

    fake.sendRaw("not json at all")

    expect(heard).not.toHaveBeenCalled()
  })
})
