import { describe, expect, it } from "vitest"
import type { List } from "@nooks/api"

import { isArchived } from "./list-state"

describe("whether a List has been put away", () => {
  it("is archived once there is a date on it", () => {
    expect(isArchived({ archivedAt: "2026-08-04T10:00:00Z" } as List)).toBe(true)
  })

  // The wire sends an empty string rather than nothing, so this is the ordinary case
  // and not an edge of one.
  it("is not archived while the date is empty", () => {
    expect(isArchived({ archivedAt: "" } as List)).toBe(false)
  })
})
