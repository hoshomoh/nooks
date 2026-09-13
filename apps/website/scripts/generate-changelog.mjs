import { mkdir, readFile, writeFile } from "node:fs/promises"

/**
 * Puts the changelog in the docs, where somebody choosing a version is already looking.
 *
 * CHANGELOG.md is written by release-please from the Conventional Commits on main, so
 * this copies rather than composes: the words in it were decided when each change was
 * committed, and a second hand-kept list would start disagreeing with the first within
 * a release or two.
 *
 * Written as `.md` and not `.mdx`. A commit subject is prose somebody typed, and prose
 * containing a brace or an angle bracket is an MDX expression — which would turn a
 * routine commit message into a failed docs build weeks later.
 */
const OUT = "content/docs/changelog.md"
const SOURCE = "../../CHANGELOG.md"

/** What the page says before there is anything to say. */
const NOTHING_YET = `Nooks has not had a release yet.

Until it does, run it from a checkout — see [Docker](/docs/deploy/docker). The first
release will list every version here, with what changed in each.`

/**
 * The changelog as the docs want it.
 *
 * release-please opens the file with a heading and a note about how it is generated.
 * The page has a title of its own, and a reader who has arrived at a changelog does not
 * need telling that changelogs are kept.
 */
function body(markdown) {
  return markdown
    .replace(/^# Changelog\s*/m, "")
    .replace(/^The format is based on.*$/gm, "")
    .replace(/^and this project adheres to.*$/gm, "")
    .trim()
}

const markdown = await readFile(SOURCE, "utf8").catch(() => "")

await mkdir("content/docs", { recursive: true })
await writeFile(
  OUT,
  `---
title: Changelog
description: Every released version, and what changed in it.
---

${markdown ? body(markdown) : NOTHING_YET}
`,
)

console.log(`changelog: ${markdown ? "written from CHANGELOG.md" : "no releases yet"}`)
