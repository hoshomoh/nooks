import { queryOptions } from "@tanstack/react-query"

import { instanceClient } from "./api"

/** What this copy of Nooks is and how much it holds. */
export const aboutQuery = queryOptions({
  queryKey: ["instance-about"],
  queryFn: () => instanceClient.getInstanceAbout({}),
})
