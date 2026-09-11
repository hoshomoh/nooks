import { useState } from "react"
import { useTranslation } from "react-i18next"
import type { Member } from "@nooks/api"

import { Button } from "./button"
import { DIALOG_SURFACE } from "./dialog-surface"
import { Field } from "./field"
import { MemberPicker } from "./member-picker"
import { Dialog, DialogContent } from "@/components/ui/dialog"
import { toggled } from "@/lib/toggle-uid"

/** A Group as it is being made: a name, and who is in it. */
export interface NewGroup {
  name: string
  memberUids: string[]
}

export interface AddGroupDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  /** Everybody on the Instance. */
  members: Member[]
  onAdd: (group: NewGroup) => void
}

/**
 * Making a Group: what it is called, and who is in it, in one dialog.
 *
 * A Group exists in order to share with the people in it, so naming it and filling it
 * is one decision. Asking for the name, then making a Member find the Group they just
 * made in order to put anybody in it, is the same decision asked twice.
 *
 * Nobody has to be picked. A Group made empty is a perfectly good Group to fill later.
 */
export function AddGroupDialog({ open, onOpenChange, members, onAdd }: AddGroupDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent showCloseButton={false} className={DIALOG_SURFACE}>
        {/* Keyed on being open, so a second Group does not start from the first's answers. */}
        <AddGroupForm
          key={String(open)}
          members={members}
          onCancel={() => onOpenChange(false)}
          onAdd={(group) => {
            onAdd(group)
            onOpenChange(false)
          }}
        />
      </DialogContent>
    </Dialog>
  )
}

interface AddGroupFormProps {
  members: Member[]
  onCancel: () => void
  onAdd: (group: NewGroup) => void
}

/** The form itself, which owns the answers so far. */
function AddGroupForm({ members, onCancel, onAdd }: AddGroupFormProps) {
  const { t } = useTranslation()
  const [name, setName] = useState("")
  const [chosen, setChosen] = useState<string[]>([])

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault()
        if (name.trim()) {
          onAdd({ name: name.trim(), memberUids: chosen })
        }
      }}
    >
      <header className="flex flex-col gap-2 px-6.5 pt-6">
        <h2 className="text-dialog">{t("groups.addTitle")}</h2>
        <p className="text-field text-secondary-foreground">{t("groups.addBlurb")}</p>
      </header>

      <div className="flex flex-col gap-5 px-6.5 pt-5 pb-6">
        <Field
          label={t("groups.name")}
          value={name}
          onChange={(event) => setName(event.target.value)}
          autoFocus
        />

        <div className="flex flex-col gap-2">
          <span className="text-label text-muted-foreground uppercase">{t("groups.who")}</span>
          <MemberPicker
            members={members}
            chosen={chosen}
            onToggle={(uid) => setChosen(toggled(chosen, uid))}
          />
        </div>
      </div>

      <footer className="flex items-center gap-3 border-t border-hair px-6.5 py-3.5">
        <span className="text-micro text-muted-foreground">
          {t("groups.chosenCount", { count: chosen.length })}
        </span>
        <span className="flex-1" />
        <Button tone="secondary" type="button" onClick={onCancel}>
          {t("action.cancel")}
        </Button>
        <Button type="submit" disabled={!name.trim()}>
          {t("groups.addSubmit")}
        </Button>
      </footer>
    </form>
  )
}
