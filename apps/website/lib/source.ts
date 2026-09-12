import { loader } from "fumadocs-core/source"

import { docs } from "@/.source/server"

/** Every docs page, read from content/docs and served under /docs. */
export const source = loader({
  baseUrl: "/docs",
  source: docs.toFumadocsSource(),
})
