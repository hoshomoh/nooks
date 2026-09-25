import { readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

import { LIMITS } from "@/lib/limits"

/**
 * The protos where the Instance declares what it will refuse, by the field each limit
 * is named after.
 */
const DECLARED_IN: Record<keyof typeof LIMITS, [string, string, string]> = {
  itemLabel: ["list_service", "CreateItemRequest", "label"],
  itemQuantity: ["list_service", "CreateItemRequest", "quantity"],
  itemNote: ["list_service", "UpdateItemRequest", "note"],
  listName: ["list_service", "CreateListRequest", "name"],
  memberName: ["member_service", "AddMemberRequest", "name"],
  memberEmail: ["member_service", "AddMemberRequest", "email"],
  groupName: ["member_service", "CreateGroupRequest", "name"],
  tokenName: ["token_service", "CreateAccessTokenRequest", "name"],
  instanceName: ["instance_service", "InstanceSettings", "name"],
  joinMessage: ["auth_service", "RequestJoinRequest", "message"],
}

/**
 * What the app says a field holds is what the Instance will accept.
 *
 * The count under a field is a promise: type this much and it will save. The Instance is
 * the only thing that actually decides, so the number in the app is a copy, and a copy
 * of a number is a number that drifts. A field that says twenty characters are left when
 * the Instance stopped accepting them twenty ago is worse than no count at all, because
 * somebody trusted it.
 *
 * Read out of the protos, which is where the limit is now declared. It used to be read
 * out of the Go, which held the numbers itself; the constraint moved into the proto so
 * that the generated clients and the published API reference carry it too.
 */
describe("the field limits", () => {
  it("are the Instance's own numbers", () => {
    for (const [field, limit] of Object.entries(LIMITS)) {
      const where = DECLARED_IN[field as keyof typeof LIMITS]
      expect(where, `${field} is not named in DECLARED_IN`).toBeDefined()
      expect(declaredIn(where), `${field} in ${where[0]}.proto`).toBe(limit)
    }
  })

  // A rename or a move would otherwise leave every number above agreeing with nothing.
  it("is reading fields that exist", () => {
    for (const field of Object.keys(LIMITS)) {
      expect(Object.keys(DECLARED_IN)).toContain(field)
    }
  })
})

/** declaredIn is the max_len one proto field carries. */
function declaredIn([file, message, field]: [string, string, string]): number {
  const source = readFileSync(`../../proto/nooks/api/v1/${file}.proto`, "utf8")

  const start = source.indexOf(`message ${message} {`)
  expect(start, `message ${message} in ${file}.proto`).toBeGreaterThan(-1)
  const body = source.slice(start, source.indexOf("\n}", start))

  const found = body.match(
    new RegExp(`\\b${field} = \\d+ \\[\\(buf\\.validate\\.field\\)\\.string\\.max_len = (\\d+)\\]`),
  )
  expect(found, `${message}.${field} declares no max_len`).not.toBeNull()
  return Number(found?.[1])
}
