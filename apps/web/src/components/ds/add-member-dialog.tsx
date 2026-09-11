import { useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "./button"
import { Field } from "./field"
import { SecretOnce } from "./secret-once"
import { Dialog, DialogContent } from "@/components/ui/dialog"

/** What an Admin typed, to make an account from. */
export interface NewMember {
  name: string
  email: string
}

export interface AddMemberDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onAdd: (member: NewMember) => void
  /** What went wrong, e.g. an email somebody already uses. */
  error?: string
  /** The temporary password, once there is one. Shown, then never again. */
  secret?: TemporarySecret
  /** Clears the secret and closes, when the Admin has read it out. */
  onSecretRead: () => void
}

/** A password an Admin has to hand over, and who it belongs to. */
export interface TemporarySecret {
  memberName: string
  password: string
}

/**
 * Adding a Member: two steps in one dialog.
 *
 * The second step is the password, and it exists because Nooks has no mail server —
 * nothing will reach the new Member unless the Admin carries it. Showing it in the same
 * dialog is what makes that impossible to miss.
 */
export function AddMemberDialog({
  open,
  onOpenChange,
  onAdd,
  error,
  secret,
  onSecretRead,
}: AddMemberDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton={false}
        className="w-full max-w-dialog gap-0 rounded-2xl p-0 sm:max-w-dialog"
      >
        {secret ? (
          <SecretOnce
            secret={secret.password}
            title={t("members.secretTitle", { name: secret.memberName })}
            blurb={t("members.secretBlurb")}
            onDone={onSecretRead}
          />
        ) : (
          <AddMemberForm
            error={error}
            onCancel={() => onOpenChange(false)}
            onAdd={onAdd}
          />
        )}
      </DialogContent>
    </Dialog>
  )
}

interface AddMemberFormProps {
  error?: string
  onCancel: () => void
  onAdd: (member: NewMember) => void
}

/** The first step: who the account is for. */
function AddMemberForm({ error, onCancel, onAdd }: AddMemberFormProps) {
  const { t } = useTranslation()
  const [name, setName] = useState("")
  const [email, setEmail] = useState("")

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault()
        if (name.trim() && email.trim()) {
          onAdd({ name: name.trim(), email: email.trim() })
        }
      }}
    >
      <header className="flex flex-col gap-2 px-6.5 pt-6">
        <h2 className="text-dialog">{t("members.addTitle")}</h2>
        <p className="text-field text-secondary-foreground">{t("members.addBlurb")}</p>
      </header>

      <div className="flex flex-col gap-4.5 px-6.5 pt-5 pb-6">
        <Field
          label={t("members.name")}
          value={name}
          onChange={(event) => setName(event.target.value)}
          autoFocus
        />
        <Field
          label={t("members.email")}
          type="email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          hint={t("members.emailHint")}
          error={error}
        />
      </div>

      <footer className="flex items-center gap-3 border-t border-hair px-6.5 py-3.5">
        <span className="flex-1" />
        <Button tone="secondary" type="button" onClick={onCancel}>
          {t("action.cancel")}
        </Button>
        <Button type="submit">{t("members.addSubmit")}</Button>
      </footer>
    </form>
  )
}
