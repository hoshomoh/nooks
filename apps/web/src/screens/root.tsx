import { Suspense } from "react"
import { Outlet } from "@tanstack/react-router"

import { lazyNamed } from "@/components/ds/lazy"

/**
 * The palette, fetched when somebody presses the key that opens it.
 *
 * It is on every screen and reached from none of them until ⌘K, so a static import
 * here would put its list, its matching and its dialog in the chunk that loads before
 * anything is on screen.
 */
const CommandPalette = lazyNamed(
  () => import("@/components/ds/command-palette"),
  "CommandPalette",
)

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
      <Suspense fallback={null}>
        <CommandPalette />
      </Suspense>
    </>
  )
}
