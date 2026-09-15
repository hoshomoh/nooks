import "@testing-library/jest-dom/vitest"
import { cleanup } from "@testing-library/react"
import { afterEach } from "vitest"

/**
 * What a test that renders needs.
 *
 * The setup file of the "dom" project in vite.config.ts, which is every *.test.tsx and
 * nothing else. A setup file rather than an import because workers are reused between
 * files: an import is evaluated once per worker, so the cleanup below would be
 * registered for the first file and for none of the ones after it.
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
