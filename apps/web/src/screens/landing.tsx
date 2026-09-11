import { getRouteApi } from "@tanstack/react-router"

import { Home } from "./home"
import { PublicListScreen } from "./public-list"

const route = getRouteApi("/")

/**
 * The front door.
 *
 * One address, two pages: a Member's Lists, or the public one for somebody who has no
 * account. Which it is was decided by the loader, so this only has to draw it.
 */
export function Landing() {
  const { signedIn } = route.useLoaderData()
  return signedIn ? <Home /> : <PublicListScreen />
}
