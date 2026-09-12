import type { BaseLayoutProps } from "fumadocs-ui/layouts/shared"

/**
 * What the docs layout is given.
 *
 * Almost nothing. The site has one header and one footer of its own, in the root
 * layout, so everything Fumadocs would otherwise put around the tree — a second bar
 * with the mark in it, a second GitHub link, a second theme control — is a second copy
 * of something the reader already has.
 */
export const baseOptions: BaseLayoutProps = {
  nav: { enabled: false },
}
