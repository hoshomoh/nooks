/**
 * initialOf is the one letter an avatar carries, e.g. "Anna" → "A".
 *
 * The first character rather than the first code unit, so a name that starts with an
 * accent or an emoji is not cut in half.
 */
export function initialOf(name: string): string {
  return (Array.from(name)[0] ?? "").toUpperCase()
}
