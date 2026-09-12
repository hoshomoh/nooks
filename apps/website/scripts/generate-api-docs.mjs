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
const OUT = "content/docs/api"

/**
 * What each service is called in the sidebar.
 *
 * The generator names a folder after the proto service, which is the right name for a
 * thing a client is generated from and the wrong one for a reader looking for how to
 * share a list. A service missing from here keeps its generated name rather than
 * failing the build: a new endpoint should not be able to break the docs.
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

// Only what this script wrote. The overview beside it is prose somebody wrote, and
// clearing the whole directory would take it with them every build.
for (const [folder] of GROUPS) {
  await rm(`${OUT}/${folder}`, { recursive: true, force: true })
}

await generateFiles({
  input: createOpenAPI({ input: ["../../proto/gen/openapi.yaml"] }),
  output: OUT,
  // One page per operation, grouped by the service it belongs to: a single page of
  // thirty-seven endpoints is a page nobody reads to the end of.
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
 * The generator hands over "Token Service_ List Access Tokens", which is the operation
 * id with spaces in it. The service is already the folder, so only what comes after the
 * underscore is news, and it reads as a sentence rather than a Title Case Heading.
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
 * `root` is what gives it a sidebar of its own rather than a branch of the docs tree:
 * somebody reading the reference is doing a different job from somebody installing the
 * thing, and one sidebar holding both is a sidebar neither of them can scan.
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
