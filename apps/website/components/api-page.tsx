"use client"

import { createOpenAPIPage } from "fumadocs-openapi/ui"

/**
 * The component the generated reference pages render themselves through.
 *
 * Its own client module because the factory builds a client component and carries the
 * Shiki bundle that highlights every request example. Built once, not per page.
 */
export const OpenAPIPage = createOpenAPIPage()
