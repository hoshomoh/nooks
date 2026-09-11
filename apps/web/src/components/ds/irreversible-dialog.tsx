import { useState, type ReactNode } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "./button"
import { DIALOG_SURFACE } from "./dialog-surface"
import { Field } from "./field"
import { Dialog, DialogContent } from "@/components/ui/dialog"

/** What the two fields carry once they are filled in. */
export interface IrreversibleConfirmation {
  /** The name, typed out, so a Member has read what they are losing. */
  typedName: string
  password: string
}

export interface IrreversibleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  /** What goes and what is kept, as plain sentences rather than a warning. */
  blurb: ReactNode
  /** The exact name that has to be typed back. */
  name: string
  /** How the name field is labelled, e.g. "Type Brunnen Street to confirm". */
  nameLabel: string
  confirmLabel: string
  onConfirm: (confirmation: IrreversibleConfirmation) => void
  error?: string
  pending?: boolean
}

/**
 * The confirmation for something that cannot be undone, per DESIGN.md §9.
 *
 * Two fields, doing different jobs. Typing the name makes a Member read what they are
 * about to lose. The password is the one that matters: an unattended browser is how
 * this realistically happens by accident, and a name can be copied off the screen in
 * front of you.
 *
 * The confirm button is the outlined destructive, never filled. A filled red button is
 * a thing to click; this is a thing to decide.
 */
export function IrreversibleDialog(props: IrreversibleDialogProps) {
  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent showCloseButton={false} className={DIALOG_SURFACE}>
        {/* Keyed on being open, so it never reopens holding what was typed last time. */}
        <ConfirmForm key={String(props.open)} {...props} />
      </DialogContent>
    </Dialog>
  )
}

/** The form itself, which owns what has been typed. */
function ConfirmForm({
  title,
  blurb,
  name,
  nameLabel,
  confirmLabel,
  onConfirm,
  onOpenChange,
  error,
  pending,
}: IrreversibleDialogProps) {
  const { t } = useTranslation()
  const [typedName, setTypedName] = useState("")
  const [password, setPassword] = useState("")

  const ready = typedName.trim().toLowerCase() === name.toLowerCase() && password !== ""

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault()
        if (ready) {
          onConfirm({ typedName: typedName.trim(), password })
        }
      }}
    >
      <header className="flex flex-col gap-2 px-6.5 pt-6">
        <h2 className="text-dialog">{title}</h2>
        <div className="flex flex-col gap-2 text-field leading-[1.6] text-secondary-foreground">
          {blurb}
        </div>
      </header>

      <div className="flex flex-col gap-4.5 px-6.5 pt-5 pb-6">
        <Field
          label={nameLabel}
          value={typedName}
          onChange={(event) => setTypedName(event.target.value)}
          autoFocus
        />
        <Field
          label={t("danger.yourPassword")}
          type="password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          hint={t("danger.passwordHint")}
          error={error}
        />
      </div>

      <footer className="flex items-center gap-3 border-t border-hair px-6.5 py-3.5">
        <span className="text-micro text-muted-foreground">{t("danger.noUndo")}</span>
        <span className="flex-1" />
        <Button tone="secondary" type="button" onClick={() => onOpenChange(false)}>
          {t("action.cancel")}
        </Button>
        <Button tone="destructive" type="submit" disabled={!ready || pending}>
          {confirmLabel}
        </Button>
      </footer>
    </form>
  )
}
