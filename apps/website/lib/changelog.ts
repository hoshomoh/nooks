import { readFile } from "node:fs/promises"
import { resolve } from "node:path"

/** One line of a release, as the page groups them. */
export interface Change {
  /** New, Changed or Fixed — the heading release-please wrote it under. */
  kind: string
  text: string
}

/** One released version. */
export interface Release {
  version: string
  /** The day it was released, already written out. */
  date: string
  /** Whether anything in it changes the shape of stored data. */
  migration: boolean
  changes: Change[]
}

/**
 * Where release-please keeps the list.
 *
 * Resolved from the working directory rather than from `import.meta.url`, which points
 * into the compiled bundle once Turbopack has been over this file and not at the
 * repository at all — the page builds either way, and the wrong one quietly renders as
 * though nothing has ever been released.
 */
const CHANGELOG = resolve(process.cwd(), "../../CHANGELOG.md")

/** How release-please writes the head of an entry: `## [1.2.0](compare-url) (2026-09-12)`. */
const ENTRY = /^##+ \[?v?([\d.]+)]?(?:\([^)]*\))? \((\d{4}-\d{2}-\d{2})\)/
/** A section heading inside one entry. */
const SECTION = /^###+ (.+)$/
/** One line of change, with any link release-please appended to it removed. */
const CHANGE = /^\* (.+)$/

/**
 * The releases, read from the file release-please writes.
 *
 * Parsed rather than hand-kept. The words in each line were decided when the change was
 * committed, and a second list maintained by hand starts disagreeing with the first
 * inside a release or two.
 *
 * A missing file is a project that has not released yet, not a broken build: the page
 * says so and the site still deploys.
 */
export async function releases(): Promise<Release[]> {
  return readReleases(await readFile(CHANGELOG, "utf8").catch(() => ""))
}

/** readReleases is the reading itself, with no file in it. */
export function readReleases(markdown: string): Release[] {
  const found: Release[] = []
  let current: Release | null = null
  let kind = ""

  for (const line of markdown.split("\n")) {
    const entry = ENTRY.exec(line)
    if (entry) {
      current = {
        version: `v${entry[1]}`,
        date: written(entry[2] ?? ""),
        migration: false,
        changes: [],
      }
      kind = ""
      found.push(current)
      continue
    }
    if (!current) {
      continue
    }

    const section = SECTION.exec(line)
    if (section) {
      kind = plain(section[1] ?? "")
      // release-please marks these itself, and they are the ones worth a badge: a
      // change to stored data is the only kind somebody has to act on before upgrading.
      if (/breaking/i.test(kind)) {
        current.migration = true
        kind = "Changed"
      }
      continue
    }

    const change = CHANGE.exec(line)
    if (change && kind) {
      current.changes.push({ kind, text: plain(change[1] ?? "") })
    }
  }

  return found
}

/** written turns 2026-09-12 into 12 September 2026, the way the page reads dates. */
function written(iso: string): string {
  const day = new Date(`${iso}T00:00:00Z`)
  return Number.isNaN(day.getTime())
    ? iso
    : new Intl.DateTimeFormat("en-GB", {
        day: "numeric",
        month: "long",
        year: "numeric",
        timeZone: "UTC",
      }).format(day)
}

/**
 * plain strips the markdown a commit subject picks up on its way into the file.
 *
 * The scope release-please bolds, the commit link it appends, and the emphasis anybody
 * typed. The page draws its own hierarchy and has no room for a second one.
 */
function plain(text: string): string {
  return text
    .replace(/\s*\(\[[0-9a-f]+]\([^)]*\)\)\s*$/i, "")
    .replace(/\[([^\]]*)]\([^)]*\)/g, "$1")
    .replace(/\*\*([^*]+)\*\*/g, "$1")
    .replace(/[*_`]/g, "")
    .trim()
}
