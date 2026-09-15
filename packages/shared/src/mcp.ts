/**
 * What to paste where, to point an assistant at an Instance.
 *
 * Every client wants the same three facts: the address, the HTTP transport, and one
 * header. Each one wants them in its own shape.
 *
 * Shared because those shapes are stated twice: on the Assistants page in the app, and
 * on the MCP page of the documentation site. They were two hand-written sets of fences
 * that agreed only as long as somebody remembered both. Now a client that changes its
 * format is one edit here and one test.
 */

/** The clients the Assistants page knows how to write a block for. */
export type AssistantClient = "claudeCode" | "cursor" | "vsCode" | "claudeDesktop" | "other"

/** The clients in the order the page offers them. */
export const ASSISTANT_CLIENTS: AssistantClient[] = [
  "claudeCode",
  "cursor",
  "vsCode",
  "claudeDesktop",
  "other",
]

/** Whether the block is a command to run or a file to save. */
export type BlockKind = "terminal" | "file"

/** Which block each client takes, so the header says the right word. */
export const BLOCK_KIND: Record<AssistantClient, BlockKind> = {
  claudeCode: "terminal",
  cursor: "file",
  vsCode: "file",
  claudeDesktop: "file",
  other: "file",
}

/**
 * TOKEN_PLACEHOLDER stands in until the token exists.
 *
 * A name rather than dashes or angle brackets: somebody who copies the block early
 * still has something they can search for and replace by hand.
 */
export const TOKEN_PLACEHOLDER = "NOOKS_TOKEN"

export interface SetupBlockParams {
  client: AssistantClient
  /** Where the Instance answers, as an origin, with no path and no trailing slash. */
  address: string
  /** The secret, or TOKEN_PLACEHOLDER while there is not one yet. */
  token: string
}

/** setupBlock writes the command or file one client needs. */
export function setupBlock({ client, address, token }: SetupBlockParams): string {
  const url = endpointOf(address)

  switch (client) {
    case "claudeCode":
      return [
        `claude mcp add nooks --transport http ${url} \\`,
        `  --header "Authorization: Bearer ${token}" \\`,
        "  --scope user",
      ].join("\n")

    case "cursor":
      return remote({ key: "mcpServers", url, token })

    case "vsCode":
      // VS Code is the odd one out: "servers", not "mcpServers".
      return remote({ key: "servers", url, token, typed: true })

    case "claudeDesktop":
      // Claude Desktop only speaks the local protocol, so this starts a bridge that
      // makes the HTTP request on its behalf.
      return [
        "{",
        '  "mcpServers": {',
        '    "nooks": {',
        '      "command": "npx",',
        '      "args": [',
        '        "-y",',
        '        "mcp-remote",',
        `        "${url}",`,
        '        "--header",',
        `        "Authorization: Bearer ${token}"`,
        "      ]",
        "    }",
        "  }",
        "}",
      ].join("\n")

    case "other":
      return remote({ key: "mcpServers", url, token, typed: true })
  }
}

interface RemoteBlockParams {
  /** What the client calls the map of servers. */
  key: "mcpServers" | "servers"
  url: string
  token: string
  /** Whether the client wants the transport named. */
  typed?: boolean
}

/** The shape almost every client takes: a URL and one header, under a name. */
function remote({ key, url, token, typed }: RemoteBlockParams): string {
  return [
    "{",
    `  "${key}": {`,
    '    "nooks": {',
    ...(typed ? ['      "type": "http",'] : []),
    `      "url": "${url}",`,
    `      "headers": { "Authorization": "Bearer ${token}" }`,
    "    }",
    "  }",
    "}",
  ].join("\n")
}

/** endpointOf is where the Instance serves MCP, from where it serves the app. */
export function endpointOf(address: string): string {
  return `${address.replace(/\/+$/, "")}/mcp`
}

/**
 * onlyThisMachine reports whether an address is one nothing else can reach.
 *
 * A Member reading the page on localhost has an address that works perfectly for them
 * and for nobody else, including an assistant on the laptop next to them. Knowing that
 * is what lets the page ask for a better one before they copy a block that cannot work.
 */
export function onlyThisMachine(address: string): boolean {
  const host = hostOf(address)
  return (
    host === "localhost" ||
    host.endsWith(".localhost") ||
    // URL keeps the brackets on an IPv6 hostname.
    host === "[::1]" ||
    /^127\./.test(host)
  )
}

/** hostOf reads the hostname, or "" from something that is not an address. */
function hostOf(address: string): string {
  try {
    return new URL(address).hostname.toLowerCase()
  } catch {
    return ""
  }
}
