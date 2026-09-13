import { describe, expect, it } from "vitest"

import { inApi, sections } from "./sections"

/*
The switcher has to offer both halves, from either half.

It offered one for a release: Fumadocs builds the list from folders marked `root` inside
the tree, and the documentation is the tree — so marking it `root` looked right, did
nothing, and left the reference with no way back.
*/
const PAGES = [
  "/docs",
  "/docs/install",
  "/docs/deploy/docker",
  "/docs/using/sharing",
  "/docs/api",
  "/docs/api/listservice/ListService_GetList",
]

describe("the docs switcher", () => {
  it("offers both halves", () => {
    expect(sections(PAGES).map((tab) => tab.title)).toEqual(["Documentation", "API"])
  })

  it("points each half at its own landing page", () => {
    expect(sections(PAGES).map((tab) => tab.url)).toEqual(["/docs", "/docs/api"])
  })

  it("gives every page to exactly one half", () => {
    // Which is what decides the one shown as current. Both halves claiming a page, or
    // neither claiming it, is a switcher that cannot say where the reader is.
    const [docs, api] = sections(PAGES)
    for (const url of PAGES) {
      const claims = [docs?.urls?.has(url), api?.urls?.has(url)].filter(Boolean)
      expect(claims, `${url} is claimed by ${claims.length} halves`).toHaveLength(1)
    }
  })

  it("does not let the documentation claim the reference", () => {
    // Every reference page sits under /docs, so a check on the URL alone would hand
    // them all to the documentation and leave the reference never looking current.
    const [docs] = sections(PAGES)

    expect(docs?.urls?.has("/docs/api")).toBe(false)
    expect(docs?.urls?.has("/docs/api/listservice/ListService_GetList")).toBe(false)
  })
})

describe("what counts as the reference", () => {
  it("takes the landing page and everything under it", () => {
    expect(inApi("/docs/api")).toBe(true)
    expect(inApi("/docs/api/listservice/ListService_GetList")).toBe(true)
  })

  it("leaves the documentation alone", () => {
    expect(inApi("/docs")).toBe(false)
    expect(inApi("/docs/using/sharing")).toBe(false)
  })

  it("does not match a page that merely starts with the same letters", () => {
    expect(inApi("/docs/apiary")).toBe(false)
  })
})
