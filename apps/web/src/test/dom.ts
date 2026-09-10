import "@testing-library/jest-dom/vitest"
import { cleanup } from "@testing-library/react"
import { afterEach } from "vitest"

/**
 * What a test that renders needs.
 *
 * Imported by the files that render components rather than set up globally, so the
 * tests that are pure functions keep starting in milliseconds. Pair it with the
 * `@vitest-environment jsdom` docblock at the top of such a file.
 */
afterEach(cleanup)

/**
 * jsdom has no layout, so it has no ResizeObserver either.
 *
 * Components that measure themselves — the command palette does — construct one on
 * mount and never get a callback. A stub is enough: there is nothing to observe.
 */
class NoLayoutResizeObserver implements ResizeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

globalThis.ResizeObserver ??= NoLayoutResizeObserver

// Scrolling an element into view is layout too, and there is none to scroll.
Element.prototype.scrollIntoView ??= () => {}
