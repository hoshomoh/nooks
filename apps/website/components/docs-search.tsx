"use client"

import { useSearchContext } from "fumadocs-ui/contexts/search"

/**
 * The way into the docs search, in the site's own header.
 *
 * Fumadocs normally puts this in a bar of its own. That bar is turned off — the site
 * already has one — so the trigger is borrowed here rather than lost. The dialog and
 * the ⌘K binding are still Fumadocs': this is a button, not a second search.
 */
export function DocsSearch() {
  const { setOpenSearch } = useSearchContext()

  return (
    <button
      type="button"
      onClick={() => setOpenSearch(true)}
      className="hidden h-8.5 flex-none items-center gap-2 rounded-md border border-border px-2.5 text-meta text-muted-foreground hover:text-foreground sm:flex"
    >
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" aria-hidden>
        <circle cx="10.5" cy="10.5" r="6.5" stroke="currentColor" strokeWidth="1.8" />
        <path d="M15.5 15.5L21 21" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
      </svg>
      Search
      <span className="flex gap-0.5">
        <Key>⌘</Key>
        <Key>K</Key>
      </span>
    </button>
  )
}

/** One key of the shortcut, drawn as a key rather than written as text. */
function Key({ children }: { children: React.ReactNode }) {
  return (
    <span className="rounded border border-border px-1 font-mono text-[11px] leading-4">
      {children}
    </span>
  )
}
