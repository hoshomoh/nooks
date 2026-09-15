/**
 * Where the project lives, for anything that has to link to it.
 *
 * A plain module rather than a constant exported from a component file: every export of
 * a "use client" module becomes a client reference when a server component imports it,
 * so a string borrowed from one arrives as an object and fails at prerender.
 */

/** SOURCE is the one outbound link on the site that is not our own. */
export const SOURCE = "https://github.com/hoshomoh/nooks"

/**
 * Where the documentation lives.
 *
 * Named here rather than written into the metadata, because the site links to it, the
 * app links to it from Settings, and the README points at it. A site that disagrees
 * with its own repository about where it is has told somebody the wrong thing.
 */
export const SITE = "https://usenooks.vercel.app"
