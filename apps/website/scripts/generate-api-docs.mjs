import { rm } from "node:fs/promises"

import { generateFiles } from "fumadocs-openapi"
import { createOpenAPI } from "fumadocs-openapi/server"

/**
 * Turns the generated OpenAPI spec into MDX the docs pipeline already knows how to
 * render.
 *
 * The spec itself comes from the protos, so this is the third thing derived from one
 * description of the API: the app's client, the REST routes, and now the reference.
 * Nobody writes down an endpoint twice, which is the only way the docs and the server
 * stay in step.
 */
const OUT = "content/docs/reference"

await rm(OUT, { recursive: true, force: true })

await generateFiles({
  input: createOpenAPI({ input: ["../../proto/gen/openapi.yaml"] }),
  output: OUT,
  // One page per operation, grouped by the service it belongs to: a single page of
  // thirty-seven endpoints is a page nobody reads to the end of.
  per: "operation",
  groupBy: "tag",
})
