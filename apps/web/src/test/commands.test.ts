import { readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

const DOCS = ["../../AGENTS.md", "../../CONTRIBUTING.md", "../../README.md"]

/**
 * A command the documentation tells somebody to run is a command that runs.
 *
 * `AGENTS.md` said `cd proto && buf generate`, and buf is a workspace dependency rather
 * than something installed globally — CONTRIBUTING.md says so in as many words — so that
 * line answers `buf: command not found`. It is the first thing an agent asked to change
 * a proto would try. The repository already wraps all three in root scripts for exactly
 * this reason, and the documentation named the unwrapped ones.
 *
 * Two questions, because the fault had two halves. A tool has to be one a clean checkout
 * can actually reach, and a pnpm script has to exist in the package.json for whatever
 * directory the command says to run it in.
 */
describe("the commands the documentation gives", () => {
  const commands = DOCS.flatMap((doc) => commandsIn(doc).map((command) => ({ doc, ...command })))

  // A floor, so a parser that stopped finding commands does not agree with everything.
  it("are found at all", () => {
    expect(commands.length).toBeGreaterThan(10)
  })

  it("name a tool a clean checkout has", () => {
    // Everything else lives in node_modules and is only on PATH through pnpm.
    const onPath = new Set(["go", "gofmt", "docker", "git", "pnpm", "npx"])
    const wrong = commands
      .filter(({ tool }) => !onPath.has(tool) && !tool.startsWith("./"))
      .map(({ doc, line }) => `${doc}: ${line}`)

    expect(wrong, "not on PATH in a clean checkout; go through pnpm").toEqual([])
  })

  it("name a pnpm script that exists where it is run", () => {
    const wrong = commands
      .filter(({ tool }) => tool === "pnpm")
      .filter(({ script }) => script !== "" && !BUILT_IN.has(script))
      .filter(({ cwd, script }) => !scriptsIn(cwd).has(script))
      .map(({ doc, line, cwd }) => `${doc}: ${line} (no such script in ${cwd}/package.json)`)

    expect(wrong, "a documented pnpm script that is not there").toEqual([])
  })
})

/** BUILT_IN is what pnpm answers itself, with no script of that name. */
const BUILT_IN = new Set(["install", "add", "remove", "dlx", "exec", "why"])

/** One command a reader would type, and where the page says to type it. */
interface Command {
  line: string
  /** Directory it runs in, relative to the repository root. */
  cwd: string
  /** The binary it invokes. */
  tool: string
  /** The pnpm script it asks for, or "" when it is not a pnpm command. */
  script: string
}

/**
 * commandsIn reads the shell blocks of a page.
 *
 * Three shapes had to be understood before this said anything true. A command split
 * over lines with a trailing backslash is one command, not a second one starting at
 * whatever word the next line begins with. `FOO=bar go test` runs go, not FOO. And
 * `pnpm --filter @nooks/web dev` runs the web package's script, which is the other way
 * the README writes what AGENTS.md writes as `cd apps/web && pnpm dev`.
 *
 * The first version of this understood none of them and reported five faults that were
 * not there.
 */
function commandsIn(doc: string): Command[] {
  const text = readFileSync(doc, "utf8")
  const blocks = [...text.matchAll(/```bash\n([\s\S]*?)```/g)].map((m) => m[1])

  const found: Command[] = []
  for (const block of blocks) {
    for (const line of joinContinuations(block)) {
      const moved = /^cd\s+(\S+)\s*&&\s*(.*)$/.exec(line)
      let cwd = moved ? moved[1] : "."
      const rest = moved ? moved[2] : line

      // Drop any NAME=value prefixes: they set the environment, they are not the tool.
      const words = rest.split(/\s+/).filter((word) => !/^[A-Z_][A-Z0-9_]*=/.test(word))
      if (words.length === 0 || words[0] === "cd") {
        continue
      }

      let script = ""
      if (words[0] === "pnpm") {
        if (words[1] === "--filter" && words[2] !== undefined) {
          cwd = directoryOf(words[2]) ?? cwd
          script = words[3] ?? ""
        } else {
          script = words[1] ?? ""
        }
      }
      found.push({ line, cwd, tool: words[0], script })
    }
  }
  return found
}

/** joinContinuations makes one line of a command written across several. */
function joinContinuations(block: string): string[] {
  const lines: string[] = []
  let carried = ""
  for (const raw of block.split("\n")) {
    const line = raw.replace(/#.*$/, "").trim()
    if (line === "") {
      continue
    }
    if (line.endsWith("\\")) {
      carried += `${line.slice(0, -1).trim()} `
      continue
    }
    lines.push((carried + line).trim())
    carried = ""
  }
  if (carried.trim() !== "") {
    lines.push(carried.trim())
  }
  return lines
}

/** directoryOf is where a workspace package lives, by the name pnpm filters on. */
function directoryOf(name: string): string | null {
  for (const dir of ["apps/web", "apps/website", "packages/api", "packages/shared", "packages/design"]) {
    try {
      const parsed = JSON.parse(readFileSync(`../../${dir}/package.json`, "utf8")) as { name?: string }
      if (parsed.name === name) {
        return dir
      }
    } catch {
      continue
    }
  }
  return null
}

/** scriptsIn is what a package.json in that directory offers. */
function scriptsIn(cwd: string): Set<string> {
  const at = cwd === "." ? "../../package.json" : `../../${cwd}/package.json`
  try {
    const parsed = JSON.parse(readFileSync(at, "utf8")) as { scripts?: Record<string, string> }
    return new Set(Object.keys(parsed.scripts ?? {}))
  } catch {
    return new Set()
  }
}
