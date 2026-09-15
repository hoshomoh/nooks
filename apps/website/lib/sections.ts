import type { LayoutTab } from "fumadocs-ui/layouts/shared"

/** Where the reference lives. Everything not under it is the documentation. */
export const API_PREFIX = "/docs/api"

/** inApi reports whether a page belongs to the reference rather than the guide. */
export function inApi(url: string): boolean {
  return url === API_PREFIX || url.startsWith(`${API_PREFIX}/`)
}

/**
 * The two halves of the docs, for the switcher above the tree.
 *
 * Written out rather than derived: Fumadocs finds them by looking for folders marked
 * `root` inside the tree, and the documentation is the tree itself. Marking it `root`
 * reads as though it should work and does nothing.
 *
 * Each carries the pages it covers so the right one is marked current. Matching on the
 * URL alone would have `/docs` claim every reference page, since all of them sit under
 * it.
 */
export function sections(urls: readonly string[]): LayoutTab[] {
  return [
    {
      title: "Documentation",
      url: "/docs",
      urls: new Set(urls.filter((url) => !inApi(url))),
    },
    {
      title: "API",
      url: API_PREFIX,
      urls: new Set(urls.filter(inApi)),
    },
  ]
}
