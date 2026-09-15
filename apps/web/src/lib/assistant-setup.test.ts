import { describe, expect, it } from "vitest"

import en from "@/i18n/locales/en.json"
import {
  ASSISTANT_CLIENTS,
  BLOCK_KIND,
  endpointOf,
  onlyThisMachine,
  setupBlock,
  TOKEN_PLACEHOLDER,
} from "./assistant-setup"

const ADDRESS = "https://nooks.example"
const SECRET = "nk_test_secret"

describe("the block an assistant is given", () => {
  it("points every client at /mcp", () => {
    for (const client of ASSISTANT_CLIENTS) {
      const block = setupBlock({ client, address: ADDRESS, token: SECRET })
      expect(block, client).toContain("https://nooks.example/mcp")
    }
  })

  it("carries the token in every client's block", () => {
    for (const client of ASSISTANT_CLIENTS) {
      const block = setupBlock({ client, address: ADDRESS, token: SECRET })
      expect(block, client).toContain(`Bearer ${SECRET}`)
    }
  })

  it("leaves a name to replace while there is no token", () => {
    const block = setupBlock({ client: "cursor", address: ADDRESS, token: TOKEN_PLACEHOLDER })
    expect(block).toContain("Bearer NOOKS_TOKEN")
  })

  it("writes JSON a client can actually parse", () => {
    for (const client of ASSISTANT_CLIENTS) {
      if (client === "claudeCode") {
        continue
      }
      const block = setupBlock({ client, address: ADDRESS, token: SECRET })
      expect(() => JSON.parse(block), client).not.toThrow()
    }
  })

  it("calls the map servers for VS Code and mcpServers for everyone else", () => {
    const vsCode: unknown = JSON.parse(setupBlock({ client: "vsCode", address: ADDRESS, token: SECRET }))
    expect(vsCode).toHaveProperty("servers.nooks")

    const cursor: unknown = JSON.parse(setupBlock({ client: "cursor", address: ADDRESS, token: SECRET }))
    expect(cursor).toHaveProperty("mcpServers.nooks")
  })

  it("bridges Claude Desktop rather than giving it a URL it cannot use", () => {
    const block = setupBlock({ client: "claudeDesktop", address: ADDRESS, token: SECRET })
    expect(block).toContain("mcp-remote")
    expect(block).not.toContain('"url"')
  })

  it("writes one command for Claude Code", () => {
    const block = setupBlock({ client: "claudeCode", address: ADDRESS, token: SECRET })
    expect(block).toContain("claude mcp add nooks --transport http")
    expect(block).toContain("--scope user")
  })
})

describe("where the instance answers", () => {
  it("does not double the slash", () => {
    expect(endpointOf("https://nooks.example/")).toBe("https://nooks.example/mcp")
  })
})

describe("whether an address is any use to somebody else", () => {
  it("knows the ones that are not", () => {
    expect(onlyThisMachine("http://localhost:8081")).toBe(true)
    expect(onlyThisMachine("http://127.0.0.1:8081")).toBe(true)
    expect(onlyThisMachine("http://[::1]:8081")).toBe(true)
    expect(onlyThisMachine("http://nooks.localhost")).toBe(true)
  })

  it("leaves a LAN address and a domain alone", () => {
    expect(onlyThisMachine("http://192.168.1.47:8081")).toBe(false)
    expect(onlyThisMachine("https://nooks.example")).toBe(false)
  })

  it("treats something that is not an address as somebody else's problem", () => {
    expect(onlyThisMachine("not an address")).toBe(false)
  })
})

describe("what the page says about each client", () => {
  // A client added to the list with no strings renders its own key at a Member, and
  // the page has five of everything, so it is the kind of miss nobody sees.
  it("has a name, a note, a block note and a file for every client", () => {
    for (const client of ASSISTANT_CLIENTS) {
      const suffix = client.charAt(0).toUpperCase() + client.slice(1)
      expect(en.assistants, client).toHaveProperty(`client${suffix}`)
      expect(en.assistants, client).toHaveProperty(`note${suffix}`)
      expect(en.assistants, client).toHaveProperty(`blockNote${suffix}`)
      if (BLOCK_KIND[client] === "file") {
        expect(en.assistants, client).toHaveProperty(`where${suffix}`)
      }
    }
  })
})
