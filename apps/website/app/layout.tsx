import type { Metadata } from "next"

import "./globals.css"

export const metadata: Metadata = {
  title: "nooks — a household todo app",
  description:
    "A self-hosted todo app for a household. One list primitive, one way to add something, and a printed page that is a real deliverable.",
}

/**
 * The document.
 *
 * No font links: the faces are bundled by @nooks/design for the same reason the app
 * bundles them — an Instance may run on a machine with no internet, and a CDN call
 * would tell Google the site was visited.
 */
export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-background text-foreground font-sans antialiased">{children}</body>
    </html>
  )
}
