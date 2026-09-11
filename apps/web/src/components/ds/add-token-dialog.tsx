import { useState } from "react"
import { useTranslation } from "react-i18next"
import { cn } from "cn"
import type { List, TokenAbilities } from "@nooks/api"

import { Button } from "./button"
import { DIALOG_SURFACE } from "./dialog-surface"
import { Field } from "./field"
import { SecretOnce } from "./secret-once"
import { TickBox } from "./tick-box"
import { Dialog, DialogContent } from "@/components/ui/dialog"
import { DateCalendar } from "./date-calendar"
import { atEndOf, momentIn, parseDue, toStored, type DueDate } from "@/lib/dates"
import { toggled } from "@/lib/toggle-uid"
import { useLocale } from "@/lib/use-locale"

/** What a Member asked for, to cut a token from. */
export interface NewToken {
  name: string
  abilities: TokenAbilities
  /** The Lists it may reach. Empty when allLists is set. */
  listUids: string[]
  /** Reach every List, including ones made later. */
  allLists: boolean
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
type Lifetime = "30" | "90" | "365" | "never" | "pick"

const LIFETIMES: Lifetime[] = ["30", "90", "365", "never", "pick"]

/** What a token is pointed at. */
type Scope = "all" | "some"

/** The three things a token may be allowed to do, in the order the design lists them. */
type Ability = "read" | "write" | "delete"

const ABILITIES: Ability[] = ["read", "write", "delete"]

/** What has been ticked so far. The wire type, without its generated marker. */
type Abilities = Record<Ability, boolean>

interface AddTokenFormProps {
  lists: List[]
  onCancel: () => void
  onAdd: (token: NewToken) => void
}

/** The form itself, which owns the answers so far. */
function AddTokenForm({ lists, onCancel, onAdd }: AddTokenFormProps) {
  const { t } = useTranslation()
  const { dateLocale } = useLocale()
  const [name, setName] = useState("")
  // Read and write on, delete off — the shape almost every caller wants, and the one
  // the design draws. Delete has to be asked for.
  const [abilities, setAbilities] = useState<Abilities>({ read: true, write: true, delete: false })
  const [scope, setScope] = useState<Scope>("all")
  const [picked, setPicked] = useState<string[]>([])
  const [lifetime, setLifetime] = useState<Lifetime>("90")
  const [chosenDay, setChosenDay] = useState<DueDate>("")

  // A token that names no List reaches none, so there is nothing to cut yet.
  const scopeReady = scope === "all" || picked.length > 0
  // A token that can do nothing wherever it reaches is a key that opens nothing.
  const abilitiesReady = abilities.read || abilities.write
  const expiryReady = lifetime !== "pick" || chosenDay !== ""
  const ready = name.trim().length > 0 && scopeReady && abilitiesReady && expiryReady

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault()
        if (!ready) {
          return
        }
        onAdd({
          name: name.trim(),
          abilities: abilities as TokenAbilities,
          listUids: scope === "all" ? [] : picked,
          allLists: scope === "all",
          expiresAt: expiryOf(lifetime, chosenDay),
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

        <fieldset className="flex flex-col gap-2.5">
          <legend className="pb-2 text-small font-medium text-secondary-foreground">
            {t("tokens.permission")}
          </legend>
          <div className="flex flex-col gap-1.5">
            {ABILITIES.map((ability) => (
              <button
                key={ability}
                type="button"
                role="checkbox"
                aria-checked={abilities[ability]}
                onClick={() =>
                  setAbilities({ ...abilities, [ability]: !abilities[ability] })
                }
                className={cn(
                  "flex items-start gap-3 rounded-xl border px-3.5 py-3 text-left transition-colors",
                  abilities[ability]
                    ? "border-[length:1.5px] border-shared bg-shared-bg"
                    : "border-border hover:bg-secondary",
                )}
              >
                <TickBox picked={abilities[ability]} className="mt-0.5" />
                <span className="flex flex-col gap-1">
                  <span className="text-field font-medium">{t(`tokens.${ability}Label`)}</span>
                  <span className="text-meta leading-[1.5] text-secondary-foreground">
                    {t(`tokens.${ability}Blurb`)}
                  </span>
                </span>
              </button>
            ))}
          </div>
        </fieldset>

        <ExplainedChoice
          label={t("tokens.scope")}
          options={[
            { value: "all", title: t("tokens.allLists"), blurb: t("tokens.allListsBlurb") },
            { value: "some", title: t("tokens.someLists"), blurb: t("tokens.someListsBlurb") },
          ]}
          chosen={scope}
          onChoose={setScope}
        />

        {scope === "some" && (
          <div className="flex flex-col gap-0.5">
            {lists.map((list) => (
              <button
                key={list.uid}
                type="button"
                role="checkbox"
                aria-checked={picked.includes(list.uid)}
                onClick={() => setPicked(toggled(picked, list.uid))}
                className={cn(
                  "flex min-h-row items-center gap-3 rounded-md px-2 py-1.5 text-left transition-colors",
                  picked.includes(list.uid) ? "bg-secondary" : "hover:bg-secondary",
                )}
              >
                <span className="text-field">{list.name}</span>
                <TickBox picked={picked.includes(list.uid)} className="ml-auto" />
              </button>
            ))}
          </div>
        )}

        <Choice
          label={t("tokens.expiry")}
          options={LIFETIMES.map((value) => ({
            value,
            label: t(`tokens.lifetime.${value}`),
          }))}
          chosen={lifetime}
          onChoose={setLifetime}
        />

        {/* The calendar itself, not a control that opens one. Asking for a day and then
            making a Member press a second thing to see the days is one step too many. */}
        {lifetime === "pick" && (
          <div className="rounded-xl border border-border p-1">
            <DateCalendar
              selected={parseDue(chosenDay) ?? undefined}
              onSelect={(day) => setChosenDay(day ? toStored(day) : "")}
              locale={dateLocale}
            />
          </div>
        )}
      </div>

      <footer className="flex items-center gap-3 border-t border-hair px-6.5 py-3.5">
        <span className="text-micro text-muted-foreground">{t("tokens.shownOnce")}</span>
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

/** expiryOf turns the answer into the moment the token stops working. */
function expiryOf(lifetime: Lifetime, chosenDay: DueDate): string {
  if (lifetime === "never") {
    return ""
  }
  if (lifetime === "pick") {
    return atEndOf(chosenDay)
  }
  return momentIn(Number(lifetime))
}

/** One of the answers an ExplainedChoice offers. */
interface ExplainedOption<T> {
  value: T
  title: string
  /** What picking it costs. Always shown: a choice nobody can read is not a choice. */
  blurb: string
}

interface ExplainedChoiceProps<T> {
  label: string
  options: ExplainedOption<T>[]
  chosen: T
  onChoose: (value: T) => void
}

/**
 * One answer out of a few, where each needs a sentence — DESIGN.md §7's bordered list.
 *
 * Every sentence is on the page at once. A control that explains only the answer
 * already chosen asks a Member to pick first and understand afterwards.
 */
function ExplainedChoice<T extends string | number>({
  label,
  options,
  chosen,
  onChoose,
}: ExplainedChoiceProps<T>) {
  return (
    <fieldset className="flex flex-col gap-2.5">
      <legend className="pb-2 text-small font-medium text-secondary-foreground">{label}</legend>
      <div className="flex flex-col gap-1.5">
        {options.map((option) => (
          <button
            key={String(option.value)}
            type="button"
            role="radio"
            aria-checked={option.value === chosen}
            onClick={() => onChoose(option.value)}
            className={cn(
              "flex flex-col gap-1 rounded-xl border px-3.5 py-3 text-left transition-colors",
              option.value === chosen
                ? "border-[length:1.5px] border-shared bg-shared-bg"
                : "border-border hover:bg-secondary",
            )}
          >
            <span className="text-field font-medium">{option.title}</span>
            <span className="text-meta leading-[1.5] text-secondary-foreground">
              {option.blurb}
            </span>
          </button>
        ))}
      </div>
    </fieldset>
  )
}
