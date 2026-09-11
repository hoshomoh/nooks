import { queryOptions } from "@tanstack/react-query"

import { instanceClient } from "./api"

/**
 * Everything an Admin may change about the Instance.
 *
 * Separate from instanceQuery, which is the public profile the sign-in page reads: an
 * Admin's view includes which List is published, and that is not a fact anybody else
 * gets to read.
 */
export const instanceSettingsQuery = queryOptions({
  queryKey: ["instance-settings"],
  queryFn: () => instanceClient.getInstanceSettings({}),
})
