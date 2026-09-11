import { useState } from "react"
import { useTranslation } from "react-i18next"
import { cn } from "cn"
import type { Member } from "@nooks/api"

import { Button } from "./button"
import { DIALOG_SURFACE } from "./dialog-surface"
import { Dialog, DialogContent } from "@/components/ui/dialog"

export interface PickPeopleProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  /** What picking these people will mean. */
  blurb: string
  /** Everybody who could be picked. */
  members: Member[]
  /** Who is picked when the dialog opens. */
  picked: string[]
  confirmLabel: string
  onConfirm: (memberUids: string[]) => void
}

/**
 * Choosing people from everyone here.
 *
 * Membership is one decision, like sharing: what the dialog sends replaces what was
 * there, rather than being a list of additions and removals to apply in order.
 */
export function PickPeople({
  open,
  onOpenChange,
  title,
  blurb,
  members,
  picked,
  confirmLabel,
  onConfirm,
}: PickPeopleProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton={false}
        className={DIALOG_SURFACE}
      >
        {/* Keyed on who is picked, so opening it again starts from what is true now. */}
        <PickPeopleForm
          key={picked.join(",")}
          title={title}
          blurb={blurb}
          members={members}
          picked={picked}
          confirmLabel={confirmLabel}
          onCancel={() => onOpenChange(false)}
          onConfirm={(chosen) => {
            onConfirm(chosen)
            onOpenChange(false)
          }}
        />
      </DialogContent>
    </Dialog>
  )
}

interface PickPeopleFormProps {
  title: string
  blurb: string
  members: Member[]
  picked: string[]
  confirmLabel: string
  onCancel: () => void
  onConfirm: (memberUids: string[]) => void
}

/** The form itself, which owns who is picked so far. */
function PickPeopleForm({
  title,
  blurb,
  members,
  picked,
  confirmLabel,
  onCancel,
  onConfirm,
}: PickPeopleFormProps) {
  const { t } = useTranslation()
  const [chosen, setChosen] = useState<string[]>(picked)

  const toggle = (uid: string) =>
    setChosen(chosen.includes(uid) ? chosen.filter((each) => each !== uid) : [...chosen, uid])

  return (
    <div className="flex flex-col">
      <header className="flex flex-col gap-2 px-6.5 pt-6">
        <h2 className="text-dialog">{title}</h2>
        <p className="text-field text-secondary-foreground">{blurb}</p>
      </header>

      <div className="flex max-h-[46vh] flex-col gap-0.5 overflow-y-auto px-6.5 pt-5 pb-1">
        {members.map((member) => (
          <button
            key={member.uid}
            type="button"
            role="checkbox"
            aria-checked={chosen.includes(member.uid)}
            onClick={() => toggle(member.uid)}
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
            <span
              className={cn(
                "ml-auto grid size-4 shrink-0 place-items-center rounded-sm text-[10px] text-background",
                chosen.includes(member.uid)
                  ? "border-[1.5px] border-shared bg-shared"
                  : "border-[1.5px] border-control",
              )}
            >
              {chosen.includes(member.uid) ? "✓" : ""}
            </span>
          </button>
        ))}
      </div>

      <footer className="mt-4 flex items-center gap-3 border-t border-hair px-6.5 py-3.5">
        <span className="flex-1" />
        <Button tone="secondary" onClick={onCancel}>
          {t("action.cancel")}
        </Button>
        <Button onClick={() => onConfirm(chosen)}>{confirmLabel}</Button>
      </footer>
    </div>
  )
}
