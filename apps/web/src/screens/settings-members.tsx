import { useState } from "react"
import { useMutation, useQuery, useQueryClient, useSuspenseQuery } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { ConnectError, Code } from "@connectrpc/connect"
import { Role, type Group, type Member } from "@nooks/api"

import {
  AddMemberDialog,
  type NewMember,
  type TemporarySecret,
} from "@/components/ds/add-member-dialog"
import { Button } from "@/components/ds/button"
import { ConfirmDialog } from "@/components/ds/confirm-dialog"
import { Menu, MenuItem, MenuSeparator } from "@/components/ds/menu"
import { SettingsShell } from "@/components/ds/settings-shell"
import { memberClient } from "@/lib/api"
import { pendingRequestsQuery } from "@/lib/member-queries"
import { instanceQuery } from "@/lib/queries"
import { groupsQuery, membersQuery } from "@/lib/sharing-queries"
import type { Translate } from "@/lib/translate"
import { useMomentLabel } from "@/lib/use-moment-label"
import { useSignedInData } from "@/lib/use-signed-in-data"
import { JoinRequests } from "./join-requests"
import { useSettingsCounts } from "@/lib/use-settings-counts"
import { IconButton } from "@/components/ds/icon-button"

/**
 * The Members page: who is here, what they may do, and who is waiting.
 *
 * Requests sit under the table because a request is somebody who is not a Member yet,
 * and deciding on them belongs beside the people who already are.
 */
export function SettingsMembersScreen() {
  const { t } = useTranslation()
  const counts = useSettingsCounts()
  const queryClient = useQueryClient()
  const { member: signedIn } = useSignedInData()
  const instance = useSuspenseQuery(instanceQuery).data
  const members = useSuspenseQuery(membersQuery).data.members
  const groups = useSuspenseQuery(groupsQuery).data.groups
  const requests = useQuery(pendingRequestsQuery)

  const [adding, setAdding] = useState(false)
  const [secret, setSecret] = useState<TemporarySecret | null>(null)
  const [removing, setRemoving] = useState<Member | null>(null)

  const refresh = () => queryClient.invalidateQueries({ queryKey: ["members"] })

  const add = useMutation({
    mutationFn: (fresh: NewMember) => memberClient.addMember(fresh),
    onSuccess: async (res) => {
      await refresh()
      setSecret({
        memberName: res.member?.name ?? "",
        password: res.temporaryPassword,
      })
    },
  })

  const setRole = useMutation({
    mutationFn: ({ memberUid, role }: MemberRoleVariables) =>
      memberClient.setMemberRole({ memberUid, role }),
    onSuccess: refresh,
  })

  const remove = useMutation({
    mutationFn: (memberUid: string) => memberClient.removeMember({ memberUid }),
    onSuccess: refresh,
  })

  return (
    <SettingsShell
      active="/settings/members"
      crumb={t("settings.members")}
      counts={counts}
    >
      <header className="flex items-start gap-6">
        <div className="min-w-0">
          <h1 className="mb-1.5 text-page">{t("members.title")}</h1>
          <p className="max-w-135 text-chrome leading-[1.55] text-secondary-foreground">
            {t("members.blurb", { count: members.length, name: instance.name })}
          </p>
        </div>
        <Button className="ml-auto shrink-0" onClick={() => setAdding(true)}>
          {t("members.add")}
        </Button>
      </header>

      <div className="flex flex-col">
        <div className="grid grid-cols-[1fr_120px_110px_28px] items-center gap-4 border-b border-border pb-2.5 text-label text-muted-foreground uppercase">
          <span>{t("members.person")}</span>
          <span>{t("members.groupsColumn")}</span>
          <span>{t("members.role")}</span>
          <span />
        </div>

        {members.map((member) => (
          <MemberRow
            key={member.uid}
            member={member}
            groups={groups}
            isSignedIn={member.uid === signedIn?.uid}
            onSetRole={(role) => setRole.mutate({ memberUid: member.uid, role })}
            onRemove={() => setRemoving(member)}
          />
        ))}
      </div>

      <JoinRequests requests={requests.data?.joinRequests ?? []} />

      <AddMemberDialog
        open={adding}
        onOpenChange={(open) => {
          setAdding(open)
          if (!open) {
            add.reset()
          }
        }}
        onAdd={(fresh) => add.mutate(fresh)}
        error={addErrorOf(t, add.error)}
        secret={secret ?? undefined}
        onSecretRead={() => {
          setSecret(null)
          setAdding(false)
          add.reset()
        }}
      />

      <ConfirmDialog
        open={removing !== null}
        onOpenChange={(open) => !open && setRemoving(null)}
        title={t("members.removeTitle", { name: removing?.name ?? "" })}
        blurb={t("members.removeBlurb")}
        confirmLabel={t("members.removeConfirm")}
        destructive
        onConfirm={() => removing && remove.mutate(removing.uid)}
      />
    </SettingsShell>
  )
}

/** What the role mutation is told. */
interface MemberRoleVariables {
  memberUid: string
  role: Role
}

interface MemberRowProps {
  member: Member
  groups: Group[]
  /** An Admin cannot remove their own account, so the row does not offer it. */
  isSignedIn: boolean
  onSetRole: (role: Role) => void
  onRemove: () => void
}

/** One person, and what can be done about them. */
function MemberRow({ member, groups, isSignedIn, onSetRole, onRemove }: MemberRowProps) {
  const { t } = useTranslation()
  const timeOf = useMomentLabel()
  const arrived = member.lastSignedInAt !== ""

  return (
    <div className="grid min-h-14 grid-cols-[1fr_120px_110px_28px] items-center gap-4 border-b border-hair">
      <span className="flex items-center gap-3">
        <span className="grid size-7 shrink-0 place-items-center rounded-full bg-chip text-[11px] text-secondary-foreground">
          {initialsOf(member.name)}
        </span>
        <span className="flex min-w-0 flex-col">
          {/* Somebody who has not arrived is greyed by name and detail, never as a
              whole row: the account is real, and the row is not disabled. */}
          <span className={arrived ? "text-field" : "text-field text-muted-foreground"}>
            {member.name}
          </span>
          <span className="truncate text-micro text-muted-foreground">
            {member.email}
            {!arrived && ` · ${t("members.neverSignedIn")}`}
            {arrived && ` · ${t("members.addedOn", { date: timeOf(member.createdAt) })}`}
          </span>
        </span>
      </span>

      <span className="truncate text-small text-secondary-foreground">
        {groupsOf(member, groups) || t("members.noGroups")}
      </span>

      <span className="text-small text-secondary-foreground">
        {t(member.role === Role.ADMIN ? "members.admin" : "members.member")}
      </span>

      <Menu
        trigger={
          <IconButton
            name="more"
            scale="compact"
            label={t("listMenu.open")}
            className="justify-self-end"
          />
        }
      >
        <MenuItem
          onSelect={() => onSetRole(member.role === Role.ADMIN ? Role.MEMBER : Role.ADMIN)}
        >
          {t(member.role === Role.ADMIN ? "members.makeMember" : "members.makeAdmin")}
        </MenuItem>
        {!isSignedIn && (
          <>
            <MenuSeparator />
            <MenuItem destructive onSelect={onRemove}>
              {t("members.remove", { name: member.name })}
            </MenuItem>
          </>
        )}
      </Menu>
    </div>
  )
}

/** groupsOf names the Groups somebody is in, as the column reads them. */
function groupsOf(member: Member, groups: Group[]): string {
  return groups
    .filter((group) => group.members.some((each) => each.uid === member.uid))
    .map((group) => group.name)
    .join(", ")
}

/** initialsOf is the two letters an avatar carries. */
function initialsOf(name: string): string {
  return name.slice(0, 2).toUpperCase()
}

/**
 * addErrorOf turns a refused addition into something to read under the field.
 *
 * Only the one a person can fix is named. Anything else is a server fault, and telling
 * somebody their email is wrong when it is not would send them looking in the wrong
 * place.
 */
function addErrorOf(t: Translate, error: Error | null): string | undefined {
  if (error instanceof ConnectError && error.code === Code.AlreadyExists) {
    return t("error.emailTaken")
  }
  return error ? t("error.cannotReach") : undefined
}
