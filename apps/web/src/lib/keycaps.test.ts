import { describe, expect, it } from "vitest"

import { pressed, readKeycap, type Keycap } from "./keycaps"

/** press builds the key event a Member's keyboard would send. */
function press(key: string, held: Partial<KeyboardEvent> = {}): KeyboardEvent {
  return {
    key,
    metaKey: false,
    ctrlKey: false,
    shiftKey: false,
    altKey: false,
    ...held,
  } as KeyboardEvent
}

describe("reading a keycap", () => {
  it("reads a plain letter", () => {
    expect(readKeycap("D")).toEqual({ command: false, shift: false, key: "d" })
  })

  it("reads a function key", () => {
    expect(readKeycap("F2")).toEqual({ command: false, shift: false, key: "f2" })
  })

  it("reads the modifiers in front of it", () => {
    expect(readKeycap("⌘⇧A")).toEqual({ command: true, shift: true, key: "a" })
  })

  it("reads the keys that are drawn rather than spelled", () => {
    expect(readKeycap("⌫")?.key).toBe("backspace")
    expect(readKeycap("↵")?.key).toBe("enter")
  })

  /*
   * The slot a keycap sits in holds other things.
   *
   * A sort menu puts a tick against the order it is in, in the same place. Reading that
   * as a key would bind a menu entry to a character no keyboard has.
   */
  it("is not a key at all when it is a mark", () => {
    expect(readKeycap("✓")).toBeNull()
    expect(readKeycap("▸")).toBeNull()
  })
})

describe("whether a press is the key a menu named", () => {
  const pin = readKeycap("⌘D") as Keycap
  const archive = readKeycap("⌘⇧A") as Keycap

  // One keycap, two keyboards: ⌘ is what a Mac calls the key Ctrl is everywhere else.
  it("takes command or control for ⌘", () => {
    expect(pressed(pin, press("d", { metaKey: true }))).toBe(true)
    expect(pressed(pin, press("d", { ctrlKey: true }))).toBe(true)
  })

  it("is not the same key without the modifier", () => {
    expect(pressed(pin, press("d"))).toBe(false)
  })

  it("wants shift only when the keycap drew it", () => {
    expect(pressed(pin, press("d", { metaKey: true, shiftKey: true }))).toBe(false)
    expect(pressed(archive, press("a", { metaKey: true, shiftKey: true }))).toBe(true)
    expect(pressed(archive, press("a", { metaKey: true }))).toBe(false)
  })

  // Alt makes a different character on most layouts, so it is a different press.
  it("is not the same key with alt held", () => {
    expect(pressed(pin, press("d", { metaKey: true, altKey: true }))).toBe(false)
  })

  it("reads a capital as the letter it is", () => {
    expect(pressed(archive, press("A", { metaKey: true, shiftKey: true }))).toBe(true)
  })
})
