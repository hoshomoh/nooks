import { afterEach, beforeAll, describe, expect, it, vi } from "vitest"
import { act, renderHook } from "@testing-library/react"

import { readyForEnglish } from "@/test/i18n"
import { TICK_MS } from "./clock-store"
import { useDueLabel } from "./use-due-label"

beforeAll(readyForEnglish)

afterEach(() => {
  vi.useRealTimers()
})

describe("a due label", () => {
  it("moves on at midnight without anything else rendering", () => {
    vi.useFakeTimers({ now: new Date("2026-11-13T23:59:30"), toFake: ["Date", "setInterval", "clearInterval"] })
    const { result } = renderHook(() => useDueLabel())

    expect(result.current.label("2026-11-14")).toBe("Tomorrow")

    act(() => {
      vi.advanceTimersByTime(TICK_MS)
    })

    expect(result.current.label("2026-11-14")).toBe("Today")
  })
})
