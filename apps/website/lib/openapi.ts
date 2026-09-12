import { createOpenAPI } from "fumadocs-openapi/server"

/**
 * The spec the reference pages render, read from where the protos generate it.
 *
 * Not a copy in the website: the file is written by `buf generate`, so a reference page
 * describing an endpoint that no longer exists would require somebody to have changed
 * the proto without regenerating — which CI refuses.
 */
export const openapi = createOpenAPI({
  input: ["../../proto/gen/openapi.yaml"],
})
