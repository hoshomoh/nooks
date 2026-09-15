import { ServerCodeBlock } from "fumadocs-ui/components/codeblock.rsc"
import { setupBlock, type AssistantClient } from "@nooks/shared"

/** EXAMPLE is the address every block on this page is written against. */
const EXAMPLE = "https://nooks.example"

export interface SetupBlockProps {
  client: AssistantClient
  /** What stands in for the secret. A documentation page has no token to put here. */
  token: string
}

/**
 * The block the app's own Assistants page writes, rendered into the docs.
 *
 * Not a second copy: a client that changes its format changes one file in
 * @nooks/shared, and both pages say the new thing.
 *
 * Highlighted on the server, like an ordinary fence, so this page stays static HTML and
 * nobody downloads a syntax highlighter to read five lines of JSON.
 */
export function SetupBlock({ client, token }: SetupBlockProps) {
  return (
    <ServerCodeBlock
      lang={client === "claudeCode" ? "bash" : "json"}
      themes={{ light: "github-light", dark: "github-dark" }}
      code={setupBlock({ client, address: EXAMPLE, token })}
    />
  )
}
