
/**
 * Making a whole row one target, without its contents eating the click.
 *
 * The row is a container with an element stretched over it and the contents on top. The
 * trap: a positioned child paints above that stretched element even with no z-index, so
 * the label swallows every click on it and the row works only in the gaps between its
 * own text.
 *
 * Found three times in three rows, so the parts are named here: what stretches, what is
 * inert, and what is raised above both.
 */

/** COVERING stretches an element over its container, behind the container's contents. */
export const COVERING = "absolute inset-0 rounded-[inherit]"

/**
 * INERT marks content that only shows something.
 *
 * Everything in a covered row is inert unless it has a job of its own. Text, counts,
 * dots, badges: none of them answer a click, and all of them are in the way of one.
 */
export const INERT = "pointer-events-none relative"

/** RAISED marks content that answers a click of its own — a checkbox, a `···`, a field. */
export const RAISED = "relative z-10"
