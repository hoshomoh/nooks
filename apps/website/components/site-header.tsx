"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"

import { DocsSearch } from "./docs-search"
import { ThemeToggle } from "./theme-toggle"
import { SOURCE } from "@/lib/site"

/** Where the header points, in the order it reads. */
const NAV = [
  { label: "Features", href: "/features" },
  { label: "Use cases", href: "/use-cases" },
  { label: "Docs", href: "/docs" },
  { label: "API", href: "/docs/reference" },
] as const

/**
 * The bar at the top of every page.
 *
 * It sticks, because the docs are long and the way back out should not depend on
 * scrolling. The blur behind it is the only decorative effect on the site: the content
 * under it has to stay legible while it passes.
 */
export function SiteHeader() {
  const pathname = usePathname()

  return (
    <header className="sticky top-0 z-30 border-b border-border bg-background/90 backdrop-blur-md">
      <div className="mx-auto flex h-14 max-w-site items-center gap-3.5 px-6">
        <Link href="/" className="flex flex-none items-center gap-2.5">
          <Mark />
          <span className="text-section tracking-tight">nooks</span>
        </Link>

        <nav className="flex flex-none items-center gap-0.5">
          {NAV.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              aria-current={isHere(pathname, item.href) ? "page" : undefined}
              className={`flex h-8.5 items-center rounded-md px-2.5 text-chrome font-medium transition-colors ${
                isHere(pathname, item.href)
                  ? "bg-secondary text-foreground"
                  : "text-secondary-foreground hover:text-foreground"
              }`}
            >
              {item.label}
            </Link>
          ))}
        </nav>

        <span className="flex-1" />

        {/* Only where there is something to search. On the marketing pages it would be a
            control that opens a dialog listing pages the reader is already looking at. */}
        {pathname.startsWith("/docs") && <DocsSearch />}
        <ThemeToggle />
        <a href={SOURCE} className="text-meta text-muted-foreground hover:text-foreground">
          GitHub
        </a>
        <Link
          href="/docs"
          className="flex h-8.5 flex-none items-center rounded-md bg-primary px-3.5 text-chrome font-medium text-primary-foreground"
        >
          Install
        </Link>
      </div>
    </header>
  )
}

/**
 * isHere reports whether a nav entry covers the page being read.
 *
 * The reference lives under the docs, so a prefix match alone would light both. The
 * longest matching entry wins, which is what a reader expects of a breadcrumb.
 */
function isHere(pathname: string, href: string): boolean {
  const candidates = NAV.filter((item) => pathname === item.href || pathname.startsWith(`${item.href}/`))
  const longest = candidates.reduce((best, item) => (item.href.length > best.length ? item.href : best), "")
  return longest === href
}

/** The mark, at the size the header uses it. */
function Mark() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden className="flex-none">
      <path d="M3 6h9" stroke="currentColor" strokeWidth="2.2" strokeLinecap="square" />
      <path d="M3 11h18" stroke="currentColor" strokeWidth="1" />
      <path
        d="M3 15.5h14M3 19.5h9"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="square"
        className="text-muted-foreground"
      />
    </svg>
  )
}
