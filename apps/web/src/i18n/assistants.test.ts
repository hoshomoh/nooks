import { describe, expect, it } from "vitest"
import { ASSISTANT_CLIENTS, BLOCK_KIND } from "@nooks/shared"

import en from "./locales/en.json"

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
