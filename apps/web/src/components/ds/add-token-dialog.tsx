import { useState } from "react"
import { useTranslation } from "react-i18next"
import { cn } from "cn"
import { Permission, type List } from "@nooks/api"

import { Button } from "./button"
import { DIALOG_SURFACE } from "./dialog-surface"
import { Field } from "./field"
import { SecretOnce } from "./secret-once"
import { TickBox } from "./tick-box"
import { Dialog, DialogContent } from "@/components/ui/dialog"
import { momentIn } from "@/lib/dates"

/** What a Member asked for, to cut a token from. */
export interface NewToken {
  name: string
  permission: Permission
  listUids: string[]
  /** RFC 3339, or empty for a token that does not expire. */
  expiresAt: string
}

/** A token's secret, and what it is for. Shown once. */
export interface TokenSecret {
  tokenName: string
  secret: string
}

export interface AddTokenDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  /** Every List the Member can reach — a token can never be pointed past these. */
  lists: List[]
  onAdd: (token: NewToken) => void
  /** The secret, once there is one. Shown, then never again. */
  secret?: TokenSecret
  /** Clears the secret and closes, when the Member has copied it. */
  onSecretRead: () => void
}

/**
 * Cutting an Access token: what it is for, what it may reach, and for how long.
 *
 * Two steps in one dialog, like adding a Member: the secret is the second, and it
 * exists in exactly one moment. Putting it anywhere else would suggest it can be
 * fetched again.
 */
export function AddTokenDialog({
  open,
  onOpenChange,
  lists,
  onAdd,
  secret,
  onSecretRead,
}: AddTokenDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent showCloseButton={false} className={DIALOG_SURFACE}>
        {secret ? (
          <SecretOnce
            secret={secret.secret}
            title={t("tokens.secretTitle", { name: secret.tokenName })}
            blurb={t("tokens.secretBlurb")}
            onDone={onSecretRead}
          />
        ) : (
          // Keyed on being open, so a second token does not start from the first's answers.
          <AddTokenForm
            key={String(open)}
            lists={lists}
            onCancel={() => onOpenChange(false)}
            onAdd={onAdd}
          />
        )}
      </DialogContent>
    </Dialog>
  )
}

/** How long a token lasts, as a Member thinks about it rather than as a date. */
type Lifetime = "30" | "90" | "365" | "never"

const LIFETIMES: Lifetime[] = ["30", "90", "365", "never"]

interface AddTokenFormProps {
  lists: List[]
  onCancel: () => void
  onAdd: (token: NewToken) => void
}

/** The form itself, which owns the answers so far. */
function AddTokenForm({ lists, onCancel, onAdd }: AddTokenFormProps) {
  const { t } = useTranslation()
  const [name, setName] = useState("")
  const [permission, setPermission] = useState<Permission>(Permission.WRITE)
  const [scope, setScope] = useState<string[]>([])
  const [lifetime, setLifetime] = useState<Lifetime>("90")

  const toggle = (uid: string) =>
    setScope(scope.includes(uid) ? scope.filter((each) => each !== uid) : [...scope, uid])

  // A token that names no List reaches none, so there is nothing to cut yet.
  const ready = name.trim().length > 0 && scope.length > 0

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault()
        if (!ready) {
          return
        }
        onAdd({
          name: name.trim(),
          permission,
          listUids: scope,
          expiresAt: expiryOf(lifetime),
        })
      }}
    >
      <header className="flex flex-col gap-2 px-6.5 pt-6">
        <h2 className="text-dialog">{t("tokens.addTitle")}</h2>
        <p className="text-field text-secondary-foreground">{t("tokens.addBlurb")}</p>
      </header>

      <div className="flex max-h-[56vh] flex-col gap-5 overflow-y-auto px-6.5 pt-5 pb-6">
        <Field
          label={t("tokens.name")}
          value={name}
          onChange={(event) => setName(event.target.value)}
          hint={t("tokens.nameHint")}
          autoFocus
        />

        <Choice
          label={t("tokens.permission")}
          options={[
            { value: Permission.READ, label: t("tokens.read"), blurb: t("tokens.readBlurb") },
            { value: Permission.WRITE, label: t("tokens.write"), blurb: t("tokens.writeBlurb") },
          ]}
          chosen={permission}
          onChoose={setPermission}
        />

        <fieldset className="flex flex-col gap-2">
          <legend className="pb-2 text-label text-muted-foreground uppercase">
            {t("tokens.scope")}
          </legend>
          <p className="pb-1 text-micro leading-[1.5] text-muted-foreground">
            {t("tokens.scopeBlurb")}
          </p>
          {lists.map((list) => (
            <button
              key={list.uid}
              type="button"
              role="checkbox"
              aria-checked={scope.includes(list.uid)}
              onClick={() => toggle(list.uid)}
              className={cn(
                "flex min-h-row items-center gap-3 rounded-md px-2 py-1.5 text-left transition-colors",
                scope.includes(list.uid) ? "bg-secondary" : "hover:bg-secondary",
              )}
            >
              <span className="text-field">{list.name}</span>
              <TickBox picked={scope.includes(list.uid)} className="ml-auto" />
            </button>
          ))}
        </fieldset>

        <Choice
          label={t("tokens.expiry")}
          options={LIFETIMES.map((value) => ({
            value,
            label: t(`tokens.lifetime.${value}`),
          }))}
          chosen={lifetime}
          onChoose={setLifetime}
        />
      </div>

      <footer className="flex items-center gap-3 border-t border-hair px-6.5 py-3.5">
        <span className="flex-1" />
        <Button tone="secondary" type="button" onClick={onCancel}>
          {t("action.cancel")}
        </Button>
        <Button type="submit" disabled={!ready}>
          {t("tokens.addSubmit")}
        </Button>
      </footer>
    </form>
  )
}

/** One of the answers a Choice offers. */
interface ChoiceOption<T> {
  value: T
  label: string
  /** What picking it costs, for the choices that need a sentence. */
  blurb?: string
}

interface ChoiceProps<T> {
  label: string
  options: ChoiceOption<T>[]
  chosen: T
  onChoose: (value: T) => void
}

/**
 * One answer out of a few, per DESIGN.md §7's chips.
 *
 * A row of chips rather than a select: three or four answers are quicker to read side
 * by side than behind a control that has to be opened to see what is in it.
 */
function Choice<T extends string | number>({ label, options, chosen, onChoose }: ChoiceProps<T>) {
  return (
    <fieldset className="flex flex-col gap-2.5">
      <legend className="pb-2 text-label text-muted-foreground uppercase">{label}</legend>
      <div className="flex flex-wrap gap-2">
        {options.map((option) => (
          <button
            key={String(option.value)}
            type="button"
            role="radio"
            aria-checked={option.value === chosen}
            onClick={() => onChoose(option.value)}
            className={cn(
              "flex h-7.5 items-center rounded-full px-3.5 text-meta transition-colors",
              option.value === chosen
                ? "bg-secondary font-medium text-foreground"
                : "border border-border text-secondary-foreground hover:bg-secondary",
            )}
          >
            {option.label}
          </button>
        ))}
      </div>
      {options.find((option) => option.value === chosen)?.blurb && (
        <p className="text-micro leading-[1.5] text-muted-foreground">
          {options.find((option) => option.value === chosen)?.blurb}
        </p>
      )}
    </fieldset>
  )
}

/** expiryOf turns a lifetime into the moment the token stops working. */
function expiryOf(lifetime: Lifetime): string {
  return lifetime === "never" ? "" : momentIn(Number(lifetime))
}
