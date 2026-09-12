import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useQuery } from "@tanstack/react-query"
import { cn } from "cn"
import { Sharing, type Group, type List, type Member } from "@nooks/api"

import { Button } from "./button"
import { COVERING, INERT, RAISED } from "./covering"
import { DIALOG_BODY, DIALOG_SURFACE } from "./dialog-surface"
import { Dialog, DialogContent } from "@/components/ui/dialog"
import { groupsQuery, listSharesQuery, membersQuery } from "@/lib/sharing-queries"
import { TickBox } from "./tick-box"

/** What the dialog sends back when the Member saves. */
export interface ShareDecision {
  sharing: Sharing
  canEdit: boolean
  /** Who it reaches by name. Empty for any other kind of sharing. */
  memberUids: string[]
  groupUids: string[]
}

export interface ShareDialogProps {
  list: List
  /** What this Instance calls itself, for "Everyone at Brunnen Street". */
  instanceName: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave: (decision: ShareDecision) => void
}

/**
 * The share dialog, per DESIGN.md §9 — and the second step behind it.
 *
 * Three choices, one toggle, one Save: sharing is a decision a Member makes once, not a
 * list they maintain. Picking "specific people" opens the second step rather than
 * growing this one, because a household with twelve people would make the first step
 * unreadable.
 */
export function ShareDialog({
  list,
  instanceName,
  open,
  onOpenChange,
  onSave,
}: ShareDialogProps) {
  // The dialog holds the decision until it is saved, so closing it changes nothing.
  const [sharing, setSharing] = useState<Sharing>(list.sharing)
  const [canEdit, setCanEdit] = useState(list.canEdit)
  const [picking, setPicking] = useState(false)

  const shares = useQuery(listSharesQuery(list.uid, open))
  const [named, setNamed] = useState<NamedShares | null>(null)
  const chosen = named ?? {
    memberUids: shares.data?.memberUids ?? [],
    groupUids: shares.data?.groupUids ?? [],
  }

  const save = () => {
    onSave({
      sharing,
      canEdit,
      memberUids: sharing === Sharing.SPECIFIC ? chosen.memberUids : [],
      groupUids: sharing === Sharing.SPECIFIC ? chosen.groupUids : [],
    })
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton={false}
        className={DIALOG_SURFACE}
      >
        {picking ? (
          <PeopleStep
            chosen={chosen}
            onChange={setNamed}
            onBack={() => setPicking(false)}
          />
        ) : (
          <ChoiceStep
            listName={list.name}
            instanceName={instanceName}
            sharing={sharing}
            canEdit={canEdit}
            chosenCount={chosen.memberUids.length + chosen.groupUids.length}
            onSharing={(next) => {
              setSharing(next)
              if (next === Sharing.SPECIFIC) {
                setPicking(true)
              }
            }}
            onCanEdit={setCanEdit}
            onChoosePeople={() => setPicking(true)}
            onCancel={() => onOpenChange(false)}
            onSave={save}
          />
        )}
      </DialogContent>
    </Dialog>
  )
}

/** Who a List reaches by name, while the dialog is open. */
interface NamedShares {
  memberUids: string[]
  groupUids: string[]
}

interface ChoiceStepProps {
  listName: string
  instanceName: string
  sharing: Sharing
  canEdit: boolean
  /** How many people and Groups are picked, for the "specific" option's summary. */
  chosenCount: number
  onSharing: (sharing: Sharing) => void
  onCanEdit: (canEdit: boolean) => void
  onChoosePeople: () => void
  onCancel: () => void
  onSave: () => void
}

/** The first step: who can reach this List at all. */
function ChoiceStep({
  listName,
  instanceName,
  sharing,
  canEdit,
  chosenCount,
  onSharing,
  onCanEdit,
  onChoosePeople,
  onCancel,
  onSave,
}: ChoiceStepProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col">
      <header className="flex flex-col gap-2 px-6.5 pt-6">
        <h2 className="text-dialog">{t("share.title", { name: listName })}</h2>
        <p className="text-field text-secondary-foreground">{t("share.blurb")}</p>
      </header>

      <div className="flex flex-col gap-1.5 px-6.5 pt-5">
        <Choice
          title={t("share.private")}
          blurb={t("share.privateBlurb")}
          selected={sharing === Sharing.PRIVATE}
          onSelect={() => onSharing(Sharing.PRIVATE)}
        />
        <Choice
          title={t("share.instance", { name: instanceName })}
          blurb={t("share.instanceBlurb")}
          selected={sharing === Sharing.INSTANCE}
          onSelect={() => onSharing(Sharing.INSTANCE)}
        />
        <Choice
          title={t("share.specific")}
          blurb={chosenCount > 0 ? t("share.shareWith", { count: chosenCount }) : t("share.specificBlurb")}
          selected={sharing === Sharing.SPECIFIC}
          onSelect={() => onSharing(Sharing.SPECIFIC)}
          action={
            sharing === Sharing.SPECIFIC
              ? { label: t("share.choose"), onClick: onChoosePeople }
              : undefined
          }
        />
      </div>

      <div className="mx-6.5 mt-4.5 flex items-center gap-3 border-t border-hair pt-4.5">
        <div className="flex flex-col gap-0.5">
          <span className="text-field">{t("share.canEdit")}</span>
          <span className="text-micro text-muted-foreground">{t("share.canEditBlurb")}</span>
        </div>
        <Toggle on={canEdit} onChange={onCanEdit} label={t("share.canEdit")} />
      </div>

      <footer className="mt-5 flex items-center gap-3 border-t border-hair px-6.5 py-3.5">
        <CopyAddress />
        <span className="flex-1" />
        <Button tone="secondary" onClick={onCancel}>
          {t("action.cancel")}
        </Button>
        <Button onClick={onSave}>{t("share.save")}</Button>
      </footer>
    </div>
  )
}

interface ChoiceAction {
  label: string
  onClick: () => void
}

interface ChoiceProps {
  title: string
  blurb: string
  selected: boolean
  onSelect: () => void
  /** A way into the second step, shown only on the option that has one. */
  action?: ChoiceAction
}

/**
 * One of the three ways a List can be shared.
 *
 * The whole card is the control, not the radio and the title within it. A card that
 * looks like one target but only answers to two small parts of itself is a card a
 * Member will click and think is broken.
 */
function Choice({ title, blurb, selected, onSelect, action }: ChoiceProps) {
  return (
    <div
      className={cn(
        "relative grid grid-cols-[18px_1fr] gap-3 rounded-xl p-3.5 transition-colors",
        selected ? "border-[length:1.5px] border-shared bg-shared-bg" : "border border-border",
      )}
    >
      <button
        type="button"
        role="radio"
        aria-checked={selected}
        aria-label={title}
        onClick={onSelect}
        className={COVERING}
      />

      {/* Inert, all of it. A positioned child paints above the button behind it and
          would swallow the click that the whole card is supposed to answer. */}
      <span
        aria-hidden
        className={cn(
          INERT,
          "mt-0.5 size-[15px] rounded-full transition-colors",
          selected ? "border-[4.5px] border-shared" : "border-[1.5px] border-control",
        )}
      />

      <div className={cn(INERT, "flex flex-col gap-0.5")}>
        <span className="text-field font-medium">{title}</span>
        <span className="text-small text-secondary-foreground">{blurb}</span>
        {action && (
          <button
            type="button"
            onClick={action.onClick}
            className={cn(RAISED, "pointer-events-auto self-start pt-1 text-small text-shared hover:underline")}
          >
            {action.label}
          </button>
        )}
      </div>
    </div>
  )
}

interface ToggleProps {
  on: boolean
  onChange: (on: boolean) => void
  label: string
}

/** The can-edit toggle, per DESIGN.md §7: a 34×20 track with a 16px knob. */
function Toggle({ on, onChange, label }: ToggleProps) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={on}
      aria-label={label}
      onClick={() => onChange(!on)}
      className={cn(
        "ml-auto flex h-5 w-8.5 shrink-0 items-center rounded-full p-0.5 transition-colors",
        on ? "justify-end bg-shared" : "justify-start bg-toggle-off",
      )}
    >
      <span className="size-4 rounded-full bg-knob" />
    </button>
  )
}

/** Copies the List's address, so it can be pasted where a person already talks. */
function CopyAddress() {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  const copy = async () => {
    await navigator.clipboard.writeText(window.location.href)
    setCopied(true)
  }

  return (
    <button
      type="button"
      onClick={() => void copy()}
      className="text-field text-shared hover:underline"
    >
      {copied ? t("share.copied") : t("share.copyAddress")}
    </button>
  )
}

interface PeopleStepProps {
  chosen: NamedShares
  onChange: (chosen: NamedShares) => void
  onBack: () => void
}

/** The second step: Groups first, then individuals. */
function PeopleStep({ chosen, onChange, onBack }: PeopleStepProps) {
  const { t } = useTranslation()
  const groups = useQuery(groupsQuery)
  const members = useQuery(membersQuery)

  const toggleGroup = (uid: string) =>
    onChange({ ...chosen, groupUids: toggled(chosen.groupUids, uid) })

  const toggleMember = (uid: string) =>
    onChange({ ...chosen, memberUids: toggled(chosen.memberUids, uid) })

  const total = chosen.memberUids.length + chosen.groupUids.length

  return (
    <div className="flex flex-col">
      <header className="flex flex-col gap-2 px-6.5 pt-6">
        <h2 className="text-dialog">{t("share.peopleTitle")}</h2>
        <p className="text-field text-secondary-foreground">{t("share.peopleBlurb")}</p>
      </header>

      <div className={`flex flex-col gap-0.5 px-6.5 pt-5 pb-1 ${DIALOG_BODY}`}>
        {(groups.data?.groups ?? []).map((group) => (
          <PickRow
            key={group.uid}
            badge={String(group.members.length)}
            name={group.name}
            detail={namesIn(group)}
            isGroup
            picked={chosen.groupUids.includes(group.uid)}
            onToggle={() => toggleGroup(group.uid)}
          />
        ))}
        {(members.data?.members ?? []).map((member) => (
          <PickRow
            key={member.uid}
            badge={initialsOf(member.name)}
            name={member.name}
            detail=""
            picked={chosen.memberUids.includes(member.uid)}
            onToggle={() => toggleMember(member.uid)}
          />
        ))}
      </div>

      <footer className="mt-4 flex items-center gap-3 border-t border-hair px-6.5 py-3.5">
        <span className="text-micro text-muted-foreground">{t("share.peopleFooter")}</span>
        <span className="flex-1" />
        <Button tone="secondary" onClick={onBack}>
          {t("share.back")}
        </Button>
        <Button onClick={onBack}>{t("share.shareWith", { count: total })}</Button>
      </footer>
    </div>
  )
}

interface PickRowProps {
  /** Two initials for a person, a count for a Group. */
  badge: string
  name: string
  /** Who is in a Group, or how a person was shared with. */
  detail: string
  isGroup?: boolean
  picked: boolean
  onToggle: () => void
}

/** One person or Group to share with. */
function PickRow({ badge, name, detail, isGroup, picked, onToggle }: PickRowProps) {
  return (
    <button
      type="button"
      role="checkbox"
      aria-checked={picked}
      onClick={onToggle}
      className={cn(
        "flex min-h-row items-center gap-3 rounded-md px-2 py-1.5 text-left transition-colors",
        picked ? "bg-secondary" : "hover:bg-secondary",
      )}
    >
      <span
        className={cn(
          "grid size-6 shrink-0 place-items-center bg-chip font-mono text-[10px] text-secondary-foreground",
          isGroup ? "rounded-md" : "rounded-full",
        )}
      >
        {badge}
      </span>
      <span className="text-field">{name}</span>
      {detail && <span className="truncate text-micro text-muted-foreground">{detail}</span>}
      <TickBox picked={picked} className="ml-auto" />
    </button>
  )
}

/** toggled adds a value to a set of identifiers, or takes it out. */
function toggled(uids: string[], uid: string): string[] {
  return uids.includes(uid) ? uids.filter((each) => each !== uid) : [...uids, uid]
}

/** namesIn reads a Group's membership as a sentence. */
function namesIn(group: Group): string {
  return group.members.map((member: Member) => member.name).join(", ")
}

/** initialsOf is the two letters a person's avatar carries. */
function initialsOf(name: string): string {
  return name.slice(0, 2).toUpperCase()
}
