import { gzipSync } from "node:zlib"
import { readFile, readdir } from "node:fs/promises"

/**
 * What a first visit costs, held to a budget.
 *
 * The app shipped as one 1.4MB chunk once, so every screen paid for the editor, the
 * calendar and the command menu whether or not it opened one. Splitting it fixed that;
 * this is what stops it growing back. A static import added to the wrong file is all it
 * takes, and nothing about that looks wrong in review.
 *
 * The entry chunk only. The lazy chunks are allowed to be large — that is the point of
 * their being lazy.
 */
/*
 * Where to look, because there are two places.
 *
 * `build` writes beside the app and `release` writes where the Go binary embeds from.
 * CI only ever runs the second — building twice to measure would be building twice —
 * so the directory is an argument rather than an assumption.
 */
const DIST = process.argv[2] ?? "dist/assets"

/*
 * BUDGET_BYTES is the ceiling for the entry chunk, gzipped.
 *
 * 120, against 112 today. Chosen by measuring the regression it exists to catch rather
 * than by picking a round number: putting the command palette back at the top of the
 * root screen takes the entry to 130, so anything looser would let exactly the mistake
 * this guards against through. Eight to grow into; more than that is a decision, and a
 * decision should be made here.
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
