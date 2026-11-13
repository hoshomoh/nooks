import { globSync } from "node:fs"
import { Visitor, type JSXAttribute, type JSXChild, type JSXElement, type JSXText } from "oxc-parser"
import { describe, expect, it } from "vitest"

import { parseFile } from "@/test/syntax"

/** GENERATED is upstream's, and its words are upstream's problem. */
const GENERATED = "src/components/ui/"

/**
 * KEYCAP draws a key, and a key is labelled the same in every language.
 *
 * Exempted by what encloses it rather than by which keys are named, so a new shortcut
 * needs no entry here — and a sentence that wandered inside one would still be caught.
 */
const KEYCAP = "Keycap"

/**
 * The product is not translated, and is lower case wherever it appears.
 *
 * `app.name` exists in the locale file and nothing asks for it, which is the right way
 * round: a name in there is a name somebody can translate.
 */
const PRODUCT = "nooks"

/**
 * No Member reads a word that is not in the locale file.
 *
 * STANDARDS §5: every string a Member reads comes from `src/i18n/locales/en.json`. A
 * word written into a component is a word nobody can translate and nobody can find —
 * and the app already carries a second language, so it is a word that stays English on
 * a screen where everything else did not.
 *
 * Parsed rather than matched. Two attempts at this with regular expressions produced
 * twenty-eight hits and then twenty-four, every one of them TypeScript: `useState<X>("y")`
 * hands a pattern looking for `>text<` exactly that, and `return null` is two words on a
 * line of its own. A guard that cries wolf is one somebody turns off, so this walks the
 * syntax tree and asks the parser what is markup and what is code.
 */
describe("the words in a component", () => {
  const files = globSync("src/**/*.tsx")
    .map((file) => String(file))
    .filter((file) => !file.includes(".test."))
    .filter((file) => !file.replaceAll("\\", "/").includes(GENERATED))

  it("is reading the components", () => {
    expect(files.length).toBeGreaterThan(40)
  })

  it("come from the locale file, not from the component", () => {
    const written: string[] = []

    for (const file of files) {
      const { program, lineOf } = parseFile(file)

      const wordsAmong = (children: JSXChild[], onAKey: boolean): void => {
        if (onAKey) {
          return
        }
        for (const child of children) {
          if (child.type !== "JSXText" || !readable(child.value) || trim(child.value) === PRODUCT) {
            continue
          }
          written.push(`${file}:${lineOf(firstWordOf(child))} ${JSON.stringify(trim(child.value))}`)
        }
      }

      new Visitor({
        JSXElement: (node) => wordsAmong(node.children, isKeycap(node)),
        JSXFragment: (node) => wordsAmong(node.children, false),
        JSXAttribute: (node) => {
          const name = attributeName(node)
          if (!speaksToAMember(name) || node.value?.type !== "Literal") {
            return
          }
          if (readable(node.value.value)) {
            written.push(`${file}:${lineOf(node.start)} ${name}=${JSON.stringify(node.value.value)}`)
          }
        },
      }).visit(program)
    }

    expect(written, "move these into src/i18n/locales/en.json").toEqual([])
  })
})

/** isKeycap reports whether this element draws a key, whose label is not a word. */
function isKeycap(node: JSXElement): boolean {
  const tag = node.openingElement.name
  return tag.type === "JSXIdentifier" && tag.name === KEYCAP
}

function attributeName(node: JSXAttribute): string {
  return node.name.type === "JSXIdentifier"
    ? node.name.name
    : `${node.name.namespace.name}:${node.name.name.name}`
}

/** firstWordOf is where the text starts once the indentation before it is skipped. */
function firstWordOf(node: JSXText): number {
  return node.start + node.value.length - node.value.trimStart().length
}

/** speaksToAMember names the attributes whose value a person reads or hears. */
function speaksToAMember(name: string): boolean {
  return ["aria-label", "aria-description", "placeholder", "title", "alt"].includes(name)
}

/** readable is text somebody would read, rather than whitespace or a bullet. */
function readable(text: string): boolean {
  return /[A-Za-z]{2,}/.test(trim(text))
}

function trim(text: string): string {
  return text.replaceAll(/\s+/g, " ").trim()
}
