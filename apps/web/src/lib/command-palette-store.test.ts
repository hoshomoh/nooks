import { describe, expect, it } from "vitest"

import { createCommandPaletteStore } from "./command-palette-store"

/** A fake window that records its keydown listener so a test can fire one. */
function fakeTarget() {
  let handler: ((event: KeyboardEvent) => void) | null = null
  return {
    target: {
      addEventListener: (_type: string, listener: EventListenerOrEventListenerObject) => {
        handler = listener as (event: KeyboardEvent) => void
      },
      removeEventListener: () => {},
    } as unknown as Window,
    press(key: string, modifiers: { metaKey?: boolean; ctrlKey?: boolean } = {}) {
      let defaultPrevented = false
      handler?.({
        key,
        metaKey: modifiers.metaKey ?? false,
        ctrlKey: modifiers.ctrlKey ?? false,
        preventDefault: () => {
          defaultPrevented = true
        },
      } as KeyboardEvent)
      return defaultPrevented
    },
  }
}

describe("createCommandPaletteStore", () => {
  it("starts closed", () => {
    const { target } = fakeTarget()
    expect(createCommandPaletteStore({ target }).getMode()).toBe("closed")
  })

  it("opens on ⌘K and closes on a second press", () => {
    const fake = fakeTarget()
    const store = createCommandPaletteStore({ target: fake.target })

    fake.press("k", { metaKey: true })
    expect(store.getMode()).toBe("search")

    fake.press("k", { metaKey: true })
    expect(store.getMode()).toBe("closed")
  })

  it("accepts Ctrl+K, for a Member not on a Mac", () => {
    const fake = fakeTarget()
    const store = createCommandPaletteStore({ target: fake.target })

    fake.press("K", { ctrlKey: true })
    expect(store.getMode()).toBe("search")
  })

  // The browser's own find-in-page is not what ⌘K means here.
  it("prevents the browser's default", () => {
    const fake = fakeTarget()
    createCommandPaletteStore({ target: fake.target })
    expect(fake.press("k", { metaKey: true })).toBe(true)
  })

  it("ignores k without a modifier, so typing an item works", () => {
    const fake = fakeTarget()
    const store = createCommandPaletteStore({ target: fake.target })

    fake.press("k")
    expect(store.getMode()).toBe("closed")
  })

  it("notifies subscribers once per real change", () => {
    const fake = fakeTarget()
    const store = createCommandPaletteStore({ target: fake.target })

    let notified = 0
    const unsubscribe = store.subscribe(() => void notified++)

    store.open()
    store.open() // already open
    expect(notified).toBe(1)

    unsubscribe()
    store.close()
    expect(notified).toBe(1)
  })
})
