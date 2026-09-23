import { readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

import { LIMITS } from "@/lib/limits"

/** Where the Instance declares what it will refuse. */
const SERVER = "../../server/router/api/v1/auth.go"

/**
 * What the app says a field holds is what the Instance will accept.
 *
 * The count under a field is a promise: type this much and it will save. The Instance is
 * the only thing that actually decides, so the number in the app is a copy, and a copy
 * of a number is a number that drifts. A field that says twenty characters are left when
 * the Instance stopped accepting them twenty ago is worse than no count at all, because
 * somebody trusted it.
 *
 * Read out of the Go rather than shared through the protos, which carry no limits. If
 * they ever do, this goes and the generated client carries them instead.
 */
describe("the field limits", () => {
  const source = readFileSync(SERVER, "utf8")

  // A file that stopped holding these would let every number below agree with nothing.
  it("is reading the constants", () => {
    expect(source).toContain("limitItemLabel")
  })

  it("are the Instance's own numbers", () => {
    for (const [field, limit] of Object.entries(LIMITS)) {
      expect(declaredIn(source, field), `${field} in ${SERVER}`).toBe(limit)
    }
  })
})

/**
 * declaredIn is the value of the Go constant for one field.
 *
 * `itemLabel` is `limitItemLabel`, and the Go is written with underscores in the long
 * ones, which is a Go spelling of the same number.
 */
function declaredIn(source: string, field: string): number {
  const name = "limit" + field[0].toUpperCase() + field.slice(1)
  const found = source.match(new RegExp(`\\b${name}\\s*=\\s*([\\d_]+)`))
  expect(found, `${name} is not declared`).not.toBeNull()
  return Number(found?.[1].replaceAll("_", ""))
}
