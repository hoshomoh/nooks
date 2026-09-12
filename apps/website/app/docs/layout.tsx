import { DocsLayout } from "fumadocs-ui/layouts/docs"
import type { ReactNode } from "react"

import { baseOptions } from "@/app/layout.config"
import { source } from "@/lib/source"

/**
 * The docs chrome: a tree on the left, the page in the middle, its headings right.
 *
 * Its own header is turned off because the site already has one in the root layout. Two
 * bars with the same mark in them, one above the other, is what a docs section bolted
 * onto a site looks like — and the way back to the site should not change depending on
 * which page a reader is on.
 */
export default function Layout({ children }: { children: ReactNode }) {
  return (
    <DocsLayout
      tree={source.pageTree}
      {...baseOptions}
      // The tree is the navigation here. A control that folds it away is a control
      // whose only outcome is a reader who cannot find the page they came for, and the
      // footer it would carry is a theme control and a GitHub link the header already
      // has.
      sidebar={{ collapsible: false, footer: null }}
    >
      {children}
    </DocsLayout>
  )
}
