import Link from "next/link"

import { SOURCE } from "@/lib/site"

/** One column of the footer. Every link goes somewhere that exists. */
interface FooterColumn {
  name: string
  links: { label: string; href: string }[]
}

const COLUMNS: FooterColumn[] = [
  {
    name: "Product",
    links: [
      { label: "Features", href: "/features" },
      { label: "Use cases", href: "/use-cases" },
    ],
  },
  {
    name: "Docs",
    links: [
      { label: "Install", href: "/docs" },
      { label: "Configure", href: "/docs/configure" },
      { label: "Back up", href: "/docs/back-up" },
      { label: "Reverse proxy", href: "/docs/reverse-proxy" },
    ],
  },
  {
    name: "Developers",
    links: [
      { label: "API reference", href: "/docs/api" },
      { label: "MCP server", href: "/docs/mcp" },
      { label: "Upgrading", href: "/docs/upgrade" },
    ],
  },
  {
    name: "Project",
    links: [
      { label: "Source", href: SOURCE },
      { label: "Licence", href: `${SOURCE}/blob/main/LICENSE` },
      { label: "Report an issue", href: `${SOURCE}/issues` },
    ],
  },
]

/**
 * The foot of every page.
 *
 * Four columns of links that all resolve. A footer full of placeholders is the first
 * thing that tells a reader the rest of the site might be placeholders too.
 */
export function SiteFooter() {
  return (
    <footer className="mt-24 border-t border-border">
      <div className="mx-auto grid max-w-site grid-cols-2 gap-10 px-6 py-14 sm:grid-cols-4">
        {COLUMNS.map((column) => (
          <div key={column.name} className="flex flex-col gap-3">
            <h2 className="text-label text-muted-foreground uppercase">{column.name}</h2>
            {column.links.map((link) => (
              <FooterLink key={link.label} {...link} />
            ))}
          </div>
        ))}
      </div>

      <div className="mx-auto flex max-w-site flex-wrap items-center gap-3 border-t border-hair px-6 py-5 text-meta text-muted-foreground">
        <span>AGPL-3.0. The source is yours, and so is anything you change.</span>
        <span className="flex-1" />
        <span>No telemetry, and nothing to turn off.</span>
      </div>
    </footer>
  )
}

/** A footer link, internal or out to the repository. */
function FooterLink({ label, href }: { label: string; href: string }) {
  const className = "text-meta text-secondary-foreground hover:text-foreground w-fit"
  if (href.startsWith("http")) {
    return (
      <a href={href} className={className}>
        {label}
      </a>
    )
  }
  return (
    <Link href={href} className={className}>
      {label}
    </Link>
  )
}
