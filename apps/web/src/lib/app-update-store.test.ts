import { describe, expect, it, vi } from "vitest"

import { createAppUpdateStore } from "./app-update-store"

/** A stand-in for the plugin's registerSW, so nothing here needs a service worker. */
function registrar() {
  let tell: (() => void) | undefined
  const update = vi.fn()
  const register = (options?: { onNeedRefresh?: () => void }) => {
    tell = options?.onNeedRefresh
    return update
  }
  return { register, update, arrive: () => tell?.() }
}

describe("a newer build", () => {
  it("is not waiting until the service worker says so", () => {
    const { register } = registrar()
    expect(createAppUpdateStore(register).getState()).toBe(false)
  })

  it("is waiting once it does", () => {
    const { register, arrive } = registrar()
    const store = createAppUpdateStore(register)

    arrive()

    expect(store.getState()).toBe(true)
  })

  it("tells whoever is listening", () => {
    const { register, arrive } = registrar()
    const store = createAppUpdateStore(register)
    const heard = vi.fn()
    store.subscribe(heard)

    arrive()

    expect(heard).toHaveBeenCalledOnce()
  })

  it("is never taken without being asked for", () => {
    // A Note is saved when the typing settles, so reloading early loses words.
    const { register, update, arrive } = registrar()
    createAppUpdateStore(register)

    arrive()

    expect(update).not.toHaveBeenCalled()
  })

  it("reloads into it when it is", () => {
    const { register, update } = registrar()
    createAppUpdateStore(register).apply()

    expect(update).toHaveBeenCalledWith(true)
  })
})
