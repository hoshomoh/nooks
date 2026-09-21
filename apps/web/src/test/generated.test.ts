import { globSync, readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

const GENERATED = "src/components/ui"

/**
 * The generated components stay what the generator produced.
 *
 * `src/components/ui/` is the output of `shadcn add`, and it has to stay that way or
 * the next add or update clobbers whatever was changed locally. STANDARDS §5 says so
 * and says what to do instead: wrap it in `ds/`, or move a token.
 *
 * Byte-for-byte cannot be checked without running the generator. What can be checked is
 * the shape an edit takes: somebody changing one of these reaches for something of
 * ours — a helper in `lib/`, a component in `ds/`, a string from the locale file — and
 * that import is the fingerprint. One of these files had been edited once already.
 */
describe("the generated components", () => {
  const files = globSync(`${GENERATED}/*.tsx`)

  it("are all there to be read", () => {
    expect(files.length).toBeGreaterThan(8)
  })

  it("reach for nothing of ours", () => {
    const reaching = files.flatMap((file) =>
      [...readFileSync(file, "utf8").matchAll(/from\s+"(@\/[^"]+)"/g)]
        .map((found) => found[1])
        .filter((specifier) => !specifier.startsWith("@/components/ui/"))
        .map((specifier) => `${file} imports ${specifier}`),
    )

    expect(reaching, `wrap it in ds/ instead — see STANDARDS §5`).toEqual([])
  })
})
