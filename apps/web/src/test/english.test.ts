import { globSync, readFileSync } from "node:fs"
import ts from "typescript"
import { describe, expect, it } from "vitest"

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
 * syntax tree and asks the compiler what is markup and what is code.
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
      const source = ts.createSourceFile(
        file,
        readFileSync(file, "utf8"),
        ts.ScriptTarget.Latest,
        true,
        ts.ScriptKind.TSX,
      )

      const walk = (node: ts.Node): void => {
        if (ts.isJsxText(node) && readable(node.text) && !spoken(node)) {
          written.push(`${file}:${lineOf(source, node)} ${JSON.stringify(trim(node.text))}`)
        }
        if (ts.isJsxAttribute(node) && speaksToAMember(node.name.getText())) {
          const value = node.initializer
          if (value && ts.isStringLiteral(value) && readable(value.text)) {
            written.push(`${file}:${lineOf(source, node)} ${node.name.getText()}=${JSON.stringify(value.text)}`)
          }
        }
        ts.forEachChild(node, walk)
      }
      walk(source)
    }

    expect(written, "move these into src/i18n/locales/en.json").toEqual([])
  })
})

/**
 * spoken reports whether this text is a word rather than a label on a key or the name
 * of the product.
 */
function spoken(node: ts.JsxText): boolean {
  if (trim(node.text) === PRODUCT) {
    return true
  }
  const parent = node.parent
  return ts.isJsxElement(parent) && parent.openingElement.tagName.getText() === KEYCAP
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

function lineOf(source: ts.SourceFile, node: ts.Node): number {
  return source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1
}
