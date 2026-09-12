import { readFile } from "node:fs/promises"
import { glob } from "node:fs/promises"

/**
 * Every internal link on the built site has to reach a page that was built.
 *
 * This exists because one did not. The API reference was moved into a folder and the
 * header went on pointing at the folder, which is not a page — so the one link most
 * likely to be clicked led nowhere, and nothing failed. A build that succeeds while
 * the navigation is broken is a build that cannot be trusted to say anything.
 *
 * Checked against the output rather than the sources: what matters is whether the page
 * a reader arrives at exists, not whether an MDX file with a similar name does.
 */
const OUT = ".next/server/app"

const built = new Set()
for await (const file of glob(`${OUT}/**/*.html`)) {
  built.add(routeOf(file))
}

/**
 * routeOf is the path a reader types to reach a built file.
 *
 * `index.html` is the directory it sits in, not a page called "index" — at the root
 * that is `/`, which is the single most linked-to path on the site.
 */
function routeOf(file) {
  const path = file.slice(OUT.length + 1, -".html".length)
  if (path === "index") {
    return "/"
  }
  return "/" + (path.endsWith("/index") ? path.slice(0, -"/index".length) : path)
}

if (built.size === 0) {
  console.error("no built pages found — run the build first")
  process.exit(1)
}

const broken = new Map()
for await (const file of glob(`${OUT}/**/*.html`)) {
  const html = await readFile(file, "utf8")
  for (const [, href] of html.matchAll(/href="(\/[^"#]*)"/g)) {
    const target = href.length > 1 ? href.replace(/\/$/, "") : href
    if (built.has(target) || isAsset(target)) {
      continue
    }
    broken.set(target, (broken.get(target) ?? 0) + 1)
  }
}

if (broken.size > 0) {
  console.error("links with no page behind them:\n")
  for (const [href, count] of [...broken].sort((a, b) => b[1] - a[1])) {
    console.error(`  ${href}  — linked from ${count} page(s)`)
  }
  process.exit(1)
}

console.log(`every internal link resolves — ${built.size} pages`)

/** isAsset reports whether a path is a file the build emits rather than a page. */
function isAsset(href) {
  return href.startsWith("/_next/") || /\.[a-z0-9]+$/i.test(href)
}
