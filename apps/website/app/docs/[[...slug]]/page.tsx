import { notFound } from "next/navigation"
import defaultMdxComponents from "fumadocs-ui/mdx"
import { Callout } from "fumadocs-ui/components/callout"
import { Card, Cards } from "fumadocs-ui/components/card"
import { Tab, Tabs } from "fumadocs-ui/components/tabs"
import { DocsBody, DocsDescription, DocsPage, DocsTitle } from "fumadocs-ui/page"

import type { GeneratedPageProps } from "fumadocs-openapi"

import { OpenAPIPage } from "@/components/api-page"
import { SetupBlock } from "@/components/setup-block"
import { openapi } from "@/lib/openapi"
import { source } from "@/lib/source"

/**
 * What Next hands a docs page: the path it was asked for, a promise because a dynamic
 * route is resolved per request.
 */
interface DocsPageProps {
  params: Promise<{ slug?: string[] }>
}

/** One docs page, rendered from its MDX file. */
export default async function Page(props: DocsPageProps) {
  const params = await props.params
  const page = source.getPage(params.slug)
  if (!page) {
    notFound()
  }

  const MDX = page.data.body
  // Reads the spec at build time so a reference page is static HTML like every other.
  const preloaded = await openapi.preloadOpenAPIPage(page)

  return (
    <DocsPage toc={page.data.toc} full={page.data.full}>
      <DocsTitle>{page.data.title}</DocsTitle>
      <DocsDescription>{page.data.description}</DocsDescription>
      <DocsBody>
        {/* The reference pages are generated from the spec and render themselves
            through this component; the hand-written pages never reach for it. */}
        <MDX
          components={{
            // Fumadocs' own first, then ours. Passing only ours replaces the whole set,
            // which quietly costs every code fence its copy button and its title bar —
            // the components a reader actually uses on a page full of commands.
            ...defaultMdxComponents,
            // The docs landing is a grid of cards rather than a list of links, so the
            // two components it needs are in scope for every page that wants them.
            Callout,
            Card,
            Cards,
            // Setting an assistant up is the same three facts written five ways, one
            // per client. Tabs put the reader's own client in front of them instead of
            // asking them to scroll past four they do not use.
            Tab,
            Tabs,
            // The client configuration blocks are written by @nooks/shared, the same
            // module the app's Assistants page renders, so the two cannot disagree.
            SetupBlock,
            OpenAPIPage: (props: GeneratedPageProps) => (
              <OpenAPIPage {...props} {...preloaded} />
            ),
          }}
        />
      </DocsBody>
    </DocsPage>
  )
}

/** Every page is known at build time, so the docs are static files. */
export function generateStaticParams() {
  return source.generateParams()
}

export async function generateMetadata(props: DocsPageProps) {
  const params = await props.params
  const page = source.getPage(params.slug)
  if (!page) {
    notFound()
  }
  return { title: page.data.title, description: page.data.description }
}
