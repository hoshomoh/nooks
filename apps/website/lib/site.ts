/**
 * Facts about the project that more than one component needs.
 *
 * A plain module rather than a constant exported from a component file: every export of
 * a "use client" module becomes a client reference when a server component imports it,
 * so a string borrowed from one arrives as an object and fails at prerender.
 */

/** SOURCE is the one outbound link on the site that is not our own. */
export const SOURCE = "https://github.com/hoshomoh/nooks"
