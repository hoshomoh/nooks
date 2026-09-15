import { rm } from "node:fs/promises"

import { generateFiles } from "fumadocs-openapi"
import { createOpenAPI } from "fumadocs-openapi/server"

/**
 * Turns the generated OpenAPI spec into MDX the docs pipeline already renders.
 *
 * The spec comes from the protos, so the app's client, the REST routes and the
 * reference are all derived from one description. Nobody writes an endpoint down twice.
 */
const OUT = "content/docs/api"

/**
 * What each service is called in the sidebar.
 *
 * The generated folder name is the proto service, which is right for generating a
 * client and wrong for a reader looking up how to share a list. A service missing here
 * keeps its generated name rather than failing the build.
 */
const GROUPS = [
  ["listservice", "Lists and items"],
  ["tokenservice", "Access tokens"],
  ["memberservice", "People and groups"],
  ["activityservice", "Activity"],
  ["publicservice", "Public lists"],
  ["instanceservice", "Instance"],
  ["authservice", "Sessions and passwords"],
  ["requestservice", "Joining and resets"],
]

// Only what this script wrote. The overview beside it is hand-written prose.
for (const [folder] of GROUPS) {
  await rm(`${OUT}/${folder}`, { recursive: true, force: true })
}

await generateFiles({
  input: createOpenAPI({ input: ["../../proto/gen/openapi.yaml"] }),
  output: OUT,
  // One page per operation: a single page of every endpoint is one nobody finishes.
  per: "operation",
  groupBy: "tag",
  frontmatter: (title, description) => ({ title: readable(title), description }),
  beforeWrite(files) {
    files.push(...groupMeta(), referenceMeta())
  },
})

/**
 * readable turns an operation id into something a person would write.
 *
 * The generator hands over "Token Service_ List Access Tokens". The service is already
 * the folder, so only what follows the underscore is news.
 */
function readable(title) {
  const afterService = title.includes("_") ? title.slice(title.indexOf("_") + 1) : title
  const words = afterService.trim().split(/\s+/)
  if (words.length === 0) {
    return title
  }
  return [words[0], ...words.slice(1).map((word) => word.toLowerCase())].join(" ")
}

/** groupMeta names each service folder, in the order a reader wants them. */
function groupMeta() {
  return GROUPS.map(([folder, title], index) => ({
    path: `${folder}/meta.json`,
    content: JSON.stringify({ title, pages: ["..."], defaultOpen: index === 0 }, null, 2),
  }))
}

/**
 * referenceMeta makes the API its own section, and orders the groups within it.
 *
 * `root` gives it a sidebar of its own. Reading the reference and installing the thing
 * are different jobs, and one sidebar holding both is scannable for neither.
 */
function referenceMeta() {
  return {
    path: "meta.json",
    content: JSON.stringify(
      {
        title: "API",
        root: true,
        pages: ["index", ...GROUPS.map(([folder]) => folder)],
      },
      null,
      2,
    ),
  }
}
