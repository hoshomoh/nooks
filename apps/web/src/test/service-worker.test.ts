import { readFileSync } from "node:fs"
import { describe, expect, it } from "vitest"

const SERVER = "../../server/server.go"
const CONFIG = "vite.config.ts"

/**
 * Every path the Instance answers is kept away from the app's offline fallback.
 *
 * The service worker answers a navigation it has never seen with index.html, because a
 * client route is the app rather than a missing file. Anything the server answers itself
 * has to be excluded, or an installed app serves the shell in its place and the request
 * never reaches the Instance. Nothing fails loudly: it breaks for whoever installed the
 * app and works for everybody else.
 *
 * Read off the server rather than listed here, so adding a route is what makes this ask
 * the question.
 */
describe("the offline fallback", () => {
  const server = readFileSync(SERVER, "utf8")
  const config = readFileSync(CONFIG, "utf8")
  const mounted = mountedPaths(server)

  // A regex that matched nothing would agree that nothing needs excluding.
  it("is reading the routes", () => {
    expect(mounted.length).toBeGreaterThan(3)
    expect(mounted).toContain("/healthz")
  })

  it("lets nothing the server answers fall through to the app", () => {
    const denied = denylist(config)
    const swallowed = mounted.filter((path) => !denied.some((rule) => rule.test(path)))

    expect(swallowed, `add these to navigateFallbackDenylist in ${CONFIG}`).toEqual([])
  })

  /*
   * The Connect handlers mount themselves, under a prefix taken from the proto package,
   * so no path for them is written in the server at all and the check above cannot see
   * them. They are the bulk of the API.
   */
  it("excludes the handlers that mount themselves", () => {
    expect(server).toMatch(/NewListServiceHandler/)
    expect(denylist(config).some((rule) => rule.test("/nooks.api.v1.ListService/GetList"))).toBe(
      true,
    )
  })
})

/** mountedPaths is every path written into the server's own router. */
function mountedPaths(server: string): string[] {
  const written = server.matchAll(/mux\.Handle(?:Func)?\(\s*"(?:[A-Z]+ )?(\/[^"]*)"/g)
  return [...written].map((found) => found[1]).filter((path) => path !== "/")
}

/** denylist is what the service worker is told never to answer with the app. */
function denylist(config: string): RegExp[] {
  const block = config.match(/navigateFallbackDenylist:\s*\[([^\]]*)\]/)
  expect(block, `navigateFallbackDenylist is not in ${CONFIG}`).not.toBeNull()
  const rules = block?.[1].matchAll(/\/((?:\\.|[^/\\])+)\//g) ?? []
  return [...rules].map((found) => new RegExp(found[1]))
}
