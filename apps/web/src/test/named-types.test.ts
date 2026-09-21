import { globSync, readFileSync } from "node:fs"
import ts from "typescript"
import { describe, expect, it } from "vitest"

/** GENERATED is upstream's, and its signatures are upstream's to write. */
const GENERATED = "src/components/ui/"

/**
 * A type in a signature has a name.
 *
 * STANDARDS §5: "A prop object written into a signature cannot be imported, extended, or
 * read at a glance." It is also the shape that spreads — the second component that needs
 * the same props copies the braces rather than importing them, and then the two drift.
 *
 * Parsed rather than matched: an object type in a signature spans lines, nests, and
 * shares its braces with every other use of `{` in TypeScript. There is no pattern for
 * it, only a syntax tree.
 */
describe("a type in a signature", () => {
  const files = ours()

  it("is reading the app", () => {
    expect(files.length).toBeGreaterThan(80)
  })

  it("has a name", () => {
    const inline: string[] = []

    for (const file of files) {
      const source = ts.createSourceFile(
        file,
        readFileSync(file, "utf8"),
        ts.ScriptTarget.Latest,
        true,
        file.endsWith(".tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS,
      )

      const walk = (node: ts.Node): void => {
        if (takesArguments(node)) {
          for (const parameter of node.parameters) {
            if (parameter.type && ts.isTypeLiteralNode(parameter.type)) {
              inline.push(`${file}:${lineOf(source, parameter)} ${parameter.name.getText()}`)
            }
          }
          if (node.type && ts.isTypeLiteralNode(node.type)) {
            inline.push(`${file}:${lineOf(source, node)} what it returns`)
          }
        }
        ts.forEachChild(node, walk)
      }
      walk(source)
    }

    expect(inline, "give these a name — see STANDARDS §5").toEqual([])
  })
})

type Signature =
  | ts.FunctionDeclaration
  | ts.MethodDeclaration
  | ts.ArrowFunction
  | ts.FunctionExpression

function takesArguments(node: ts.Node): node is Signature {
  return (
    ts.isFunctionDeclaration(node) ||
    ts.isMethodDeclaration(node) ||
    ts.isArrowFunction(node) ||
    ts.isFunctionExpression(node)
  )
}

/** ours is the app's own source: tests describe, and the generated components are not ours. */
function ours(): string[] {
  return globSync("src/**/*.{ts,tsx}")
    .map((file) => String(file))
    .filter((file) => !file.includes(".test."))
    .filter((file) => !file.replaceAll("\\", "/").includes(GENERATED))
}

function lineOf(source: ts.SourceFile, node: ts.Node): number {
  return source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1
}
