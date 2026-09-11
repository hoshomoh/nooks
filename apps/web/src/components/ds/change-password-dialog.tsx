import { useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "./button"
import { DIALOG_SURFACE } from "./dialog-surface"
import { Field } from "./field"
import { Dialog, DialogContent } from "@/components/ui/dialog"

/** What replacing a password needs: proof it is you, and the new one. */
export interface PasswordChange {
  currentPassword: string
  newPassword: string
}

export interface ChangePasswordDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onChange: (change: PasswordChange) => void
  /** What went wrong, e.g. a wrong current password. */
  error?: string
  pending?: boolean
}

/**
 * Replacing a password.
 *
 * A dialog rather than two fields on the page, because the settings page is a page of
 * things you change in passing and this is not one of them. The current password is
 * required so that an unattended browser cannot be used to take the account over.
 */
export function ChangePasswordDialog({
  open,
  onOpenChange,
  onChange,
  error,
  pending,
}: ChangePasswordDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent showCloseButton={false} className={DIALOG_SURFACE}>
        {/* Keyed on being open, so a second attempt starts from empty fields. */}
        <PasswordForm
          key={String(open)}
          error={error}
          pending={pending}
          onCancel={() => onOpenChange(false)}
          onChange={onChange}
        />
      </DialogContent>
    </Dialog>
  )
}

interface PasswordFormProps {
  error?: string
  pending?: boolean
  onCancel: () => void
  onChange: (change: PasswordChange) => void
}

/** The form itself, which owns what has been typed. */
function PasswordForm({ error, pending, onCancel, onChange }: PasswordFormProps) {
  const { t } = useTranslation()
  const [current, setCurrent] = useState("")
  const [next, setNext] = useState("")

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault()
        if (current && next) {
          onChange({ currentPassword: current, newPassword: next })
        }
      }}
    >
      <header className="flex flex-col gap-2 px-6.5 pt-6">
        <h2 className="text-dialog">{t("account.changePassword")}</h2>
        <p className="text-field text-secondary-foreground">{t("account.passwordBlurb")}</p>
      </header>

      <div className="flex flex-col gap-4.5 px-6.5 pt-5 pb-6">
        <Field
          label={t("account.currentPassword")}
          type="password"
          value={current}
          onChange={(event) => setCurrent(event.target.value)}
          error={error}
          autoFocus
        />
        <Field
          label={t("account.newPassword")}
          type="password"
          value={next}
          onChange={(event) => setNext(event.target.value)}
          hint={t("auth.passwordRule")}
        />
      </div>

      <footer className="flex items-center gap-3 border-t border-hair px-6.5 py-3.5">
        <span className="text-micro text-muted-foreground">{t("account.passwordNote")}</span>
        <span className="flex-1" />
        <Button tone="secondary" type="button" onClick={onCancel}>
          {t("action.cancel")}
        </Button>
        <Button type="submit" disabled={!current || !next || pending}>
          {t("account.changePassword")}
        </Button>
      </footer>
    </form>
  )
}
