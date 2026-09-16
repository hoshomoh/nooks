import { createRoute, redirect } from "@tanstack/react-router"

import { Landing } from "@/screens/landing"
import { currentMemberQuery, instanceQuery } from "@/lib/queries"
import { listPageQuery, sidebarQuery } from "@/lib/list-queries"
import {
  DEFAULT_SORT,
  LIST_SORTS,
  LIST_STATUSES,
  type ListSort,
  type ListStatus,
  type ListsSearch,
} from "@/lib/list-table"
import { publicListQuery } from "@/lib/public-queries"
import { rootRoute } from "./root"

/**
 * Where an arriving browser lands, whoever it belongs to.
 *
 * A Member gets their Lists. Somebody with no account gets the public list, if there is
 * one — that is the address an Instance hands out, so it has to be the one that answers.
 * Sending them to a sign-in page first would mean the link somebody shared opens a
 * request for credentials rather than the shopping.
 *
 * The decision happens in the loader, before anything renders, so no screen flashes on
 * the way through.
 */
export const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  validateSearch: (search: Record<string, unknown>): ListsSearch => {
    const out: ListsSearch = {}
    if (LIST_STATUSES.includes(search.status as ListStatus) && search.status !== "all") {
      out.status = search.status as ListStatus
    }
    if (LIST_SORTS.includes(search.sort as ListSort) && search.sort !== DEFAULT_SORT) {
      out.sort = search.sort as ListSort
    }
    const page = Number(search.page)
    if (Number.isFinite(page) && page > 1) {
      out.page = Math.floor(page)
    }
    return out
  },
  loaderDeps: ({ search }) => search,
  loader: async ({ context, deps }) => {
    const instance = await context.queryClient.ensureQueryData(instanceQuery)
    if (instance.needsSetup) {
      throw redirect({ to: "/setup" })
    }

    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      // Nothing published is the only case where a Visitor is asked to identify
      // themselves: there is genuinely nothing here for them otherwise.
      const page = await context.queryClient.ensureQueryData(publicListQuery)
      if (!page.published) {
        throw redirect({ to: "/sign-in" })
      }
      return { signedIn: false }
    }

    if (member.mustChangePassword) {
      throw redirect({ to: "/replace-password" })
    }

    await Promise.all([
      context.queryClient.ensureQueryData(sidebarQuery),
      context.queryClient.ensureQueryData(listPageQuery(deps)),
    ])
    return { signedIn: true }
  },
  component: Landing,
})
