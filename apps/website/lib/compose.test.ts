import { readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

const SHIPPED = "../../docker-compose.yml"
const PAGE = "content/docs/deploy/compose.mdx"

/**
 * The compose file on the page is the compose file in the repository.
 *
 * The page says "this file is the only thing you need", so somebody sets their household
 * up by copying it off the screen. Anybody who clones instead gets the one in the
 * repository. Two copies of the same file in two places drift, and the way it shows is
 * that one of them quietly stops having something — the health check was already missing
 * from the page while its own prose promised one.
 *
 * Compared without comments: the repository's copy explains itself to whoever opens it,
 * and the page has its prose around it for that.
 */
describe("the compose file", () => {
  it("is the same on the page as in the repository", () => {
    const shipped = withoutComments(readFileSync(SHIPPED, "utf8"))
    expect(shipped, "the repository's compose file is not where this expects").toContain("services:")

    expect(documentedYaml(), `keep ${PAGE} and docker-compose.yml the same file`).toEqual(shipped)
  })
})

/** documentedYaml is the first yaml block on the page, which is the whole file. */
function documentedYaml(): string {
  const page = readFileSync(PAGE, "utf8")
  const block = page.split("```yaml")[1]?.split("```")[0]
  if (!block) {
    throw new Error(`no yaml block in ${PAGE}`)
  }
  return withoutComments(block)
}

/** withoutComments drops commentary and blank lines, leaving what compose reads. */
function withoutComments(yaml: string): string {
  return yaml
    .split("\n")
    .filter((line) => line.trim() !== "" && !line.trim().startsWith("#"))
    .map((line) => line.trimEnd())
    .join("\n")
}
