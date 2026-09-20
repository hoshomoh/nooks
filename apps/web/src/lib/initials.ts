/**
 * initialsOf is the two letters an avatar carries, e.g. "Anna" → "AN".
 *
 * The first two characters rather than the first letter of each word: a household is
 * full of people entered as one name, and "Anna Schmidt" and "Anna" should not be drawn
 * differently for the same person depending on how they filled the field in.
 */
export function initialsOf(name: string): string {
  return name.slice(0, 2).toUpperCase()
}
