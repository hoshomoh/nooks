import { gzipSync } from "node:zlib"
import { readFile, readdir } from "node:fs/promises"

/**
 * What a first visit costs, held to a budget.
 *
 * Without this the app drifts back to one chunk, where every screen pays for the editor,
 * the calendar and the command menu whether or not it opens one. One static import in
 * the wrong file is all it takes, and nothing about that looks wrong in review.
 *
 * The entry chunk only. A lazy chunk is allowed to be large.
 */
// `build` writes beside the app and `release` writes where the Go binary embeds from,
// so the directory to measure is an argument rather than an assumption.
const DIST = process.argv[2] ?? "dist/assets"

/*
 * BUDGET_BYTES is the ceiling for the entry chunk, gzipped.
 *
 * Measured against the regression it exists to catch, not picked round: importing the
 * command palette eagerly from the root screen takes the entry to 130, so anything
 * looser would let that through. Raising it is a decision, and it is made here.
 */
const BUDGET_BYTES = 120 * 1024

const files = await readdir(DIST)
const entry = files.find((name) => name.startsWith("index-") && name.endsWith(".js"))
if (!entry) {
  console.error(`no entry chunk in ${DIST} — run the build first`)
  process.exit(1)
}

const bytes = gzipSync(await readFile(`${DIST}/${entry}`)).length
const asKb = (n) => `${(n / 1024).toFixed(1)} kB`

if (bytes > BUDGET_BYTES) {
  console.error(
    `the entry chunk is ${asKb(bytes)} gzipped, over the ${asKb(BUDGET_BYTES)} budget.\n` +
      `Something a screen only sometimes needs is being imported at the top of a file ` +
      `that always loads. Import it with lazyNamed, or lazyRouteComponent for a route.`,
  )
  process.exit(1)
}

console.log(`entry chunk ${asKb(bytes)} gzipped, within the ${asKb(BUDGET_BYTES)} budget`)
