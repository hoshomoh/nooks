import { readFileSync } from "node:fs"
import { parseSync, type Program } from "oxc-parser"

/** ParsedFile is a source file as a syntax tree, with the way back to its lines. */
export type ParsedFile = {
  program: Program
  source: string
  /** lineOf is the line, counted from one, that an offset into the source falls on. */
  lineOf: (offset: number) => number
}

/**
 * parseFile reads a file into its syntax tree, for the guards that need to tell markup
 * from code or a type from a value.
 *
 * A file that does not parse fails the guard rather than passing it, since a guard that
 * read nothing has checked nothing.
 */
export function parseFile(file: string): ParsedFile {
  const source = readFileSync(file, "utf8")
  const { program, errors } = parseSync(file, source)
  if (errors.length > 0) {
    throw new Error(`${file} does not parse: ${errors[0].message}`)
  }
  return {
    program,
    source,
    lineOf: (offset) => source.slice(0, offset).split("\n").length,
  }
}
