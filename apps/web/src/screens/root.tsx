import { Outlet } from "@tanstack/react-router"

import { CommandPalette } from "@/components/ds/command-palette"

/**
 * Every screen, with ⌘K over it.
 *
 * The palette lives inside the router rather than beside it, because jumping somewhere
 * is the whole point of it: rendered as a sibling of the RouterProvider it has no
 * router to navigate with, and every result silently does nothing.
 */
export function Root() {
  return (
    <>
      <Outlet />
      <CommandPalette />
    </>
  )
}
