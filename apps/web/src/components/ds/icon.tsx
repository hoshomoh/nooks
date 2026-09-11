import {
  CalendarDays,
  Check,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Ellipsis,
  Maximize2,
  Plus,
  Search,
  X,
} from "lucide-react"
import { cn } from "cn"

/**
 * Every icon the app draws, named by what it means rather than what it looks like.
 *
 * Nothing outside this file imports from the icon library. An icon reached by meaning
 * survives being redrawn; one reached by shape has to be found and renamed everywhere
 * the day the shape changes.
 */
const GLYPHS = {
  back: ChevronLeft,
  forward: ChevronRight,
  collapse: ChevronDown,
  expand: ChevronRight,
  close: X,
  check: Check,
  date: CalendarDays,
  fullScreen: Maximize2,
  more: Ellipsis,
  search: Search,
  add: Plus,
} as const

export type IconName = keyof typeof GLYPHS

/**
 * The sizes an icon comes in.
 *
 * Three, because a fourth would be a decision made per-use, which is how a set of icons
 * stops looking like a set.
 */
export type IconSize = "small" | "medium" | "large"

const SIZES: Record<IconSize, string> = {
  small: "size-3.5",
  medium: "size-4",
  large: "size-[18px]",
}

export interface IconProps {
  name: IconName
  size?: IconSize
  className?: string
}

/**
 * One icon, drawn at the app's weight.
 *
 * The stroke is set here and only here: icons at mixed weights read as icons from
 * different families, which is what typed punctuation standing in for an icon looks
 * like too.
 *
 * Always decorative. An icon that carries meaning on its own belongs in an IconButton,
 * which takes the label that says what it does.
 */
export function Icon({ name, size = "medium", className }: IconProps) {
  const Glyph = GLYPHS[name]
  return <Glyph aria-hidden strokeWidth={1.75} className={cn("shrink-0", SIZES[size], className)} />
}
