import { useTranslation } from "react-i18next"
import { cn } from "cn"

import type { SlashItem } from "@/lib/editor/slash-items"

export interface SlashMenuProps {
  items: SlashItem[]
  /** Which entry the keyboard is on. */
  selected: number
  onChoose: (item: SlashItem) => void
}

/**
 * The `/` menu, per DESIGN.md §10: 300px, radius 9, an uppercase heading, 34px rows.
 *
 * Every entry shows its markdown shorthand, so **the menu teaches the shortcut rather
 * than replacing it** — a Member who learns `### ` stops needing to open this.
 */
export function SlashMenu({ items, selected, onChoose }: SlashMenuProps) {
  const { t } = useTranslation()

  if (items.length === 0) {
    return null
  }

  return (
    <div className="w-75 rounded-menu border border-border bg-popover p-1.5 shadow-[0_12px_32px_rgba(0,0,0,0.14)]">
      <p className="px-2.5 pt-1.5 pb-2 text-label text-muted-foreground uppercase">
        {t("note.addToNote")}
      </p>

      {items.map((item, index) => (
        <button
          key={item.kind}
          type="button"
          onClick={() => onChoose(item)}
          className={cn(
            "flex h-8.5 w-full items-center gap-2.5 rounded-md px-2.5 text-left text-meta",
            index === selected && "bg-secondary",
          )}
        >
          {/* The glyph sits in a column of its own so every label starts at the same x. */}
          <span className="w-5 shrink-0 text-center text-micro text-secondary-foreground">
            {item.glyph}
          </span>
          <span>{item.label}</span>
          <span className="ml-auto font-mono text-keycap text-muted-foreground">
            {item.shorthand}
          </span>
        </button>
      ))}
    </div>
  )
}
