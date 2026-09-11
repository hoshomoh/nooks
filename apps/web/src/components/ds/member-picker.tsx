import { useTranslation } from "react-i18next"
import { cn } from "cn"
import type { Member } from "@nooks/api"

import { TickBox } from "./tick-box"

export interface MemberPickerProps {
  /** Everybody who could be picked. */
  members: Member[]
  /** Who is picked now, by identifier. */
  chosen: string[]
  onToggle: (memberUid: string) => void
}

/**
 * The list of people a dialog picks from.
 *
 * Its own component because two dialogs pick people — making a Group and changing who
 * is in one — and a row that is clickable in one of them and not the other is the kind
 * of difference nobody notices until they are using it.
 */
export function MemberPicker({ members, chosen, onToggle }: MemberPickerProps) {
  const { t } = useTranslation()

  if (members.length === 0) {
    return <p className="py-2 text-field text-muted-foreground">{t("groups.nobodyYet")}</p>
  }

  return (
    <div className="flex max-h-[38vh] flex-col gap-0.5 overflow-y-auto">
      {members.map((member) => (
        <button
          key={member.uid}
          type="button"
          role="checkbox"
          aria-checked={chosen.includes(member.uid)}
          onClick={() => onToggle(member.uid)}
          className={cn(
            "flex min-h-row items-center gap-3 rounded-md px-2 py-1.5 text-left transition-colors",
            chosen.includes(member.uid) ? "bg-secondary" : "hover:bg-secondary",
          )}
        >
          <span className="grid size-6 shrink-0 place-items-center rounded-full bg-chip font-mono text-[10px] text-secondary-foreground">
            {member.name.slice(0, 2).toUpperCase()}
          </span>
          <span className="text-field">{member.name}</span>
          <span className="truncate text-micro text-muted-foreground">{member.email}</span>
          <TickBox picked={chosen.includes(member.uid)} className="ml-auto" />
        </button>
      ))}
    </div>
  )
}
