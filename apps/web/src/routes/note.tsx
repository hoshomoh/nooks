import { createRoute, lazyRouteComponent, notFound, redirect } from "@tanstack/react-router"

import { currentMemberQuery } from "@/lib/queries"
import { listQuery } from "@/lib/list-queries"
import { rootRoute } from "./root"

/** One Item's Note, at full width. */
export const noteRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/lists/$listUid/items/$itemUid",
  loader: async ({ context, params }) => {
    const member = await context.queryClient.ensureQueryData(currentMemberQuery)
    if (!member) {
      throw redirect({ to: "/sign-in" })
    }

    const list = await context.queryClient.ensureQueryData(listQuery(params.listUid))
    const item = list.items.find((candidate) => candidate.uid === params.itemUid)
    if (!list.list || !item) {
      throw notFound()
    }
    return { member, list: list.list, item }
  },
  component: lazyRouteComponent(() => import("@/screens/note"), "NoteScreen"),
})
