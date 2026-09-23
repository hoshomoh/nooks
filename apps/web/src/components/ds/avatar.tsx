import { cn } from "cn"

/**
 * The sizes an avatar comes in.
 *
 * Three, named rather than numbered, so that reaching for a fourth is a decision about
 * the design rather than a number typed into one file. The same reason icons have
 * three.
 */
export type AvatarSize = "small" | "medium" | "large"

interface SizeStyle {
  box: string
  text: string
}

/**
 * The text keeps its ratio to the circle, near enough two fifths, so two initials sit
 * inside it at every size rather than filling the small one and rattling in the large.
 */
const SIZES: Record<AvatarSize, SizeStyle> = {
  small: { box: "size-5.5", text: "text-[9px]" },
  medium: { box: "size-6", text: "text-[10px]" },
  large: { box: "size-7", text: "text-[11px]" },
}

export interface AvatarProps {
  /** What it draws: two initials for a person, a count for a Group. */
  badge: string
  size?: AvatarSize
  /** A Group is a square. A Group is not a person and does not read as one. */
  isGroup?: boolean
  /** A ring in the page's own ground, for avatars drawn overlapping each other. */
  ringed?: boolean
  className?: string
}

/**
 * The chip a person is drawn as, per DESIGN.md §4.
 *
 * It was written by hand in six places at three sizes, four of them in the UI font and
 * two in mono, which is one thing drawn six ways. The font is the UI font everywhere
 * now: the design gives this chip a fill and a radius and says nothing about type, and
 * mono for two letters was the odd one out rather than a decision.
 */
export function Avatar({ badge, size = "medium", isGroup, ringed, className }: AvatarProps) {
  // Not hidden from a screen reader, which is what it was not before. Five of the six
  // places draw the name beside it, so the initials are read twice, and the sixth is a
  // Group's member count that is said nowhere else. Which of those is right is a
  // DESIGN §12 question rather than something to settle while moving markup.
  const scale = SIZES[size]

  return (
    <span
      className={cn(
        "grid shrink-0 place-items-center bg-chip text-secondary-foreground",
        scale.box,
        scale.text,
        isGroup ? "rounded-md" : "rounded-full",
        ringed && "border-[1.5px] border-background",
        className,
      )}
    >
      {badge}
    </span>
  )
}
