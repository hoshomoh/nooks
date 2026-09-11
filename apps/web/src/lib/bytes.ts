/** The units a size is read in, smallest first. */
const UNITS = ["B", "KB", "MB", "GB", "TB"]

/**
 * readableBytes writes a size the way somebody deciding whether to copy it would say it.
 *
 * One decimal place above a kilobyte and none below: "4.1 MB" is a useful answer and
 * "4,283,916 B" is a number nobody can hold in their head. Zero answers empty, because
 * a size we could not measure should say nothing rather than claim to be nothing.
 */
export function readableBytes(bytes: number): string {
  if (bytes <= 0) {
    return ""
  }

  let size = bytes
  let unit = 0
  while (size >= 1024 && unit < UNITS.length - 1) {
    size /= 1024
    unit += 1
  }
  return `${unit === 0 ? size : size.toFixed(1)} ${UNITS[unit]}`
}
