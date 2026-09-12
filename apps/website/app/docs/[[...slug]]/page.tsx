import { notFound } from "next/navigation"
import { DocsBody, DocsDescription, DocsPage, DocsTitle } from "fumadocs-ui/page"

import type { GeneratedPageProps } from "fumadocs-openapi"

import { OpenAPIPage } from "@/components/api-page"
import { openapi } from "@/lib/openapi"
import { source } from "@/lib/source"

/** One docs page, rendered from its MDX file. */
export default async function Page(props: { params: Promise<{ slug?: string[] }> }) {
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

export async function generateMetadata(props: { params: Promise<{ slug?: string[] }> }) {
  const params = await props.params
  const page = source.getPage(params.slug)
  if (!page) {
    notFound()
  }
  return { title: page.data.title, description: page.data.description }
}
