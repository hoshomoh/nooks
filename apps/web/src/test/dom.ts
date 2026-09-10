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
