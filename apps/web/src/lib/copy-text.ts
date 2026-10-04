/** The two ways a browser will put text on the clipboard. */
export interface CopyWays {
  /** The clipboard API. A browser offers it only on a secure page: HTTPS, or localhost. */
  clipboard: Pick<Clipboard, "writeText"> | undefined
  /** The older way: select the text and ask the browser to copy the selection. */
  bySelecting: (text: string) => boolean
}

/**
 * copyText puts text on the clipboard, and says whether it did.
 *
 * The clipboard API is the right way and is missing on a plain HTTP page at a LAN
 * address, which is how a lot of home servers are reached. Calling it there threw before
 * anything could say so, and a Copy button that does nothing looks like it worked. The
 * older way still works on those pages, so it is what is tried next.
 */
export async function copyText(text: string, ways: CopyWays): Promise<boolean> {
  if (ways.clipboard) {
    try {
      await ways.clipboard.writeText(text)
      return true
    } catch {
      // Refused, usually because the page was not focused. The other way may still work.
    }
  }
  return ways.bySelecting(text)
}

/**
 * copyBySelecting copies through a hidden field, which every browser still supports on
 * any page. execCommand is marked deprecated, and nothing else copies on an insecure one.
 */
function copyBySelecting(text: string): boolean {
  const field = document.createElement("textarea")
  field.value = text
  field.readOnly = true
  field.style.position = "fixed"
  field.style.opacity = "0"
  document.body.append(field)
  field.select()
  try {
    return document.execCommand("copy")
  } catch {
    return false
  } finally {
    field.remove()
  }
}

/** browserCopy is copyText with the ways this browser has. */
export function browserCopy(text: string): Promise<boolean> {
  return copyText(text, {
    clipboard: "clipboard" in navigator ? navigator.clipboard : undefined,
    bySelecting: copyBySelecting,
  })
}
