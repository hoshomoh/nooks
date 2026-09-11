import type { TextFile } from "./list-export"

/**
 * Hands a file to the Member.
 *
 * The only part of exporting that needs a browser, kept apart from the part that
 * decides what the file says — so the text and its name stay testable.
 */
export function download(file: TextFile): void {
  const url = URL.createObjectURL(new Blob([file.body], { type: "text/markdown" }))
  const link = document.createElement("a")
  link.href = url
  link.download = file.name
  link.click()
  URL.revokeObjectURL(url)
}
