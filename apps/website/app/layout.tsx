import type { Metadata } from "next"

import { RootProvider } from "fumadocs-ui/provider/next"

import { SiteFooter } from "@/components/site-footer"
import { SiteHeader } from "@/components/site-header"

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
  // next-themes writes the theme class onto <html> before React runs, so the markup
  // React sees does not match what it rendered. suppressHydrationWarning says so on
  // purpose rather than leaving a warning nobody can act on.
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="bg-background text-foreground flex min-h-dvh flex-col font-sans antialiased">
        {/* attribute: "class" is already the default, but it is stated because it is a
            contract rather than a preference: the app toggles a `.dark` class on the
            root element and the palette in @nooks/design hangs off that exact selector.
            A provider switched to data-theme would leave the site light forever. */}
        <RootProvider theme={{ attribute: "class", defaultTheme: "system" }}>
          <SiteHeader />
          {children}
          <SiteFooter />
        </RootProvider>
      </body>
    </html>
  )
}
