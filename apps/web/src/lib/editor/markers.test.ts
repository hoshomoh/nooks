import { describe, expect, it } from "vitest"

import { continueList, markerOf, toggleTodo } from "./markers"

describe("markerOf", () => {
  it("reads each shorthand", () => {
    expect(markerOf("### Where")).toEqual({ kind: "heading", length: 4 })
    expect(markerOf("# Big")).toEqual({ kind: "heading", length: 2 })
    expect(markerOf("> They pack up at two")).toEqual({ kind: "quote", length: 2 })
    expect(markerOf("- [ ] Whole bean")).toEqual({ kind: "todo", length: 6 })
    expect(markerOf("- [x] Got it")).toEqual({ kind: "todo-done", length: 6 })
  })

  it("treats an unmarked line as a paragraph", () => {
    expect(markerOf("Saturday market")).toEqual({ kind: "paragraph", length: 0 })
  })

  // "- [ ] " must win over "- ", or a checklist line keeps its box.
  it("prefers the longer marker", () => {
    expect(markerOf("- [ ] Whole bean").kind).toBe("todo")
  })

  // A hash inside a sentence is not a heading.
  it("requires the marker at the start, with its space", () => {
    expect(markerOf("Ask for #4").kind).toBe("paragraph")
    expect(markerOf("###Where").kind).toBe("paragraph")
  })
})

describe("toggleTodo", () => {
  it("ticks and unticks a checklist line", () => {
    expect(toggleTodo("- [ ] Whole bean")).toBe("- [x] Whole bean")
    expect(toggleTodo("- [x] Whole bean")).toBe("- [ ] Whole bean")
  })

  it("leaves anything else alone", () => {
    expect(toggleTodo("### Where")).toBe("### Where")
    expect(toggleTodo("Saturday market")).toBe("Saturday market")
  })
})

describe("continueList", () => {
  // A checklist should keep going without retyping the marker.
  it("carries a checklist onto the next line", () => {
    expect(continueList("- [ ] Whole bean")).toBe("- [ ] ")
    expect(continueList("- [x] Got it")).toBe("- [ ] ")
  })

  it("carries a quote onto the next line", () => {
    expect(continueList("> They pack up")).toBe("> ")
  })

  // Enter on an empty marked line means they are finished with the list.
  it("stops on an empty marked line", () => {
    expect(continueList("- [ ] ")).toBe("")
    expect(continueList("> ")).toBe("")
  })

  it("does nothing for a paragraph or a heading", () => {
    expect(continueList("Saturday market")).toBeNull()
    expect(continueList("### Where")).toBeNull()
  })
})
