import { globSync } from "node:fs"
import {
  Visitor,
  type ArrowFunctionExpression,
  type Function,
  type ParamPattern,
  type TSTypeAnnotation,
} from "oxc-parser"
import { describe, expect, it } from "vitest"

import { parseFile } from "@/test/syntax"

/** GENERATED is upstream's, and its signatures are upstream's to write. */
const GENERATED = "src/components/ui/"

/**
 * A type in a signature has a name, in both the apps.
 *
 * STANDARDS §5: "A prop object written into a signature cannot be imported, extended, or
 * read at a glance." It is also the shape that spreads — the second component that needs
 * the same props copies the braces rather than importing them, and then the two drift.
 *
 * Parsed rather than matched: an object type in a signature spans lines, nests, and
 * shares its braces with every other use of `{` in TypeScript. There is no pattern for
 * it, only a syntax tree.
 *
 * The website is read too, and was not before. STANDARDS is the repository's rather than
 * the app's, and nothing said the website was outside it, which is how three signatures
 * ended up written the other way with nothing to say so. Read from here rather than
 * copied into a second test, because the detector is the thing that would drift.
 */
describe("a type in a signature", () => {
  const files = ours()

  it("is reading both apps", () => {
    expect(files.length).toBeGreaterThan(80)
    expect(files.some((file) => file.includes("website"))).toBe(true)
  })

  it("has a name", () => {
    const inline: string[] = []

    for (const file of files) {
      const { program, source, lineOf } = parseFile(file)

      const check = (node: Signature): void => {
        for (const parameter of node.params) {
          const typed = annotated(parameter)
          if (isWrittenIn(typed.typeAnnotation)) {
            const name = source.slice(typed.start, typed.typeAnnotation.start).trim()
            inline.push(`${file}:${lineOf(parameter.start)} ${name}`)
          }
        }
        if (isWrittenIn(node.returnType)) {
          inline.push(`${file}:${lineOf(node.start)} what it returns`)
        }
      }

      new Visitor({
        FunctionDeclaration: check,
        FunctionExpression: check,
        ArrowFunctionExpression: check,
        TSDeclareFunction: check,
        TSEmptyBodyFunctionExpression: check,
      }).visit(program)
    }

    expect(inline, "give these a name — see STANDARDS §5").toEqual([])
  })
})

type Signature = Function | ArrowFunctionExpression

/** AnnotatedParameter is the part of a parameter that carries its type. */
type AnnotatedParameter = Exclude<ParamPattern, { type: "TSParameterProperty" | "AssignmentPattern" }>

/**
 * annotated finds where a parameter's type is written: on the name inside a default
 * (`{ a }: Props = {}`) or a constructor's `private` one, and otherwise on the parameter.
 */
function annotated(parameter: ParamPattern): AnnotatedParameter {
  if (parameter.type === "TSParameterProperty") {
    return annotated(parameter.parameter)
  }
  if (parameter.type === "AssignmentPattern") {
    return annotated(parameter.left)
  }
  return parameter
}

/** isWrittenIn reports whether a type is an object type spelled out where it is used. */
function isWrittenIn(annotation: TSTypeAnnotation | null | undefined): annotation is TSTypeAnnotation {
  return annotation?.typeAnnotation.type === "TSTypeLiteral"
}

/**
 * ours is the TypeScript this repository writes: both apps, without their tests, which
 * describe rather than declare, and without the generated components, whose signatures
 * are upstream's to write.
 */
function ours(): string[] {
  const written = [...globSync("src/**/*.{ts,tsx}"), ...globSync("../website/**/*.{ts,tsx}")]
  return written
    .map((file) => String(file))
    .filter((file) => !file.includes(".test."))
    .filter((file) => !file.includes("node_modules"))
    .filter((file) => !file.endsWith(".d.ts"))
    .filter((file) => !file.replaceAll("\\", "/").includes(GENERATED))
}
