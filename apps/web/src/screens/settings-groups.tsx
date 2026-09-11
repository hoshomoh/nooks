import { useState } from "react"
import { useMutation, useQueryClient, useSuspenseQuery } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import type { Group } from "@nooks/api"

import { Button } from "@/components/ds/button"
import { EmptyState } from "@/components/ds/empty-state"
import { PickPeople } from "@/components/ds/pick-people"
import { PromptDialog } from "@/components/ds/prompt-dialog"
import { SettingsShell } from "@/components/ds/settings-shell"
import { memberClient } from "@/lib/api"
import { formatList } from "@/lib/format"
import { groupsQuery, membersQuery } from "@/lib/sharing-queries"
import { useLocale } from "@/lib/use-locale"

/**
 * The Groups page: cards two-up, who is in each, and what each one reaches.
 *
 * A Group carries no permissions of its own, so a card says the only two things there
 * are to say about it — who is in it, and which Lists it reaches. Nothing else would be
 * true.
 */
export function SettingsGroupsScreen() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const groups = useSuspenseQuery(groupsQuery).data.groups
  const members = useSuspenseQuery(membersQuery).data.members

  const [adding, setAdding] = useState(false)
  const [editing, setEditing] = useState<Group | null>(null)

  const refresh = () => queryClient.invalidateQueries({ queryKey: ["groups"] })

  const add = useMutation({
    mutationFn: (name: string) => memberClient.createGroup({ name }),
    onSuccess: refresh,
  })

  const setMembers = useMutation({
    mutationFn: ({ groupUid, memberUids }: GroupMembersVariables) =>
      memberClient.setGroupMembers({ groupUid, memberUids }),
    // Who is in a Group decides which Lists they reach, so the sidebar changes too.
    onSuccess: () => queryClient.invalidateQueries(),
  })

  return (
    <SettingsShell
      active="/settings/groups"
      crumb={t("settings.groups")}
      counts={{ "/settings/members": members.length, "/settings/groups": groups.length }}
    >
      <header className="flex items-start gap-6">
        <div className="min-w-0">
          <h1 className="mb-1.5 text-page">{t("groups.title")}</h1>
          <p className="max-w-135 text-chrome leading-[1.55] text-secondary-foreground">
            {t("groups.blurb")}
          </p>
        </div>
        <Button className="ml-auto shrink-0" onClick={() => setAdding(true)}>
          {t("groups.add")}
        </Button>
      </header>

      {groups.length === 0 ? (
        <EmptyState title={t("groups.emptyTitle")} body={t("groups.emptyBody")} />
      ) : (
        <div className="grid grid-cols-2 gap-3">
          {groups.map((group) => (
            <GroupCard key={group.uid} group={group} onEdit={() => setEditing(group)} />
          ))}
        </div>
      )}

      <PromptDialog
        open={adding}
        onOpenChange={setAdding}
        title={t("groups.addTitle")}
        blurb={t("groups.addBlurb")}
        label={t("groups.name")}
        initialValue=""
        confirmLabel={t("groups.addSubmit")}
        onConfirm={(name) => add.mutate(name)}
      />

      <PickPeople
        open={editing !== null}
        onOpenChange={(open) => !open && setEditing(null)}
        title={t("groups.editTitle", { name: editing?.name ?? "" })}
        blurb={t("groups.editBlurb")}
        members={members}
        picked={(editing?.members ?? []).map((member) => member.uid)}
        confirmLabel={t("groups.save")}
        onConfirm={(memberUids) =>
          editing && setMembers.mutate({ groupUid: editing.uid, memberUids })
        }
      />
    </SettingsShell>
  )
}

/** What the membership mutation is told. */
interface GroupMembersVariables {
  groupUid: string
  memberUids: string[]
}

interface GroupCardProps {
  group: Group
  onEdit: () => void
}

/** One Group: who is in it, and what it reaches. */
function GroupCard({ group, onEdit }: GroupCardProps) {
  const { t } = useTranslation()
  const { code } = useLocale()

  return (
    <div className="flex flex-col gap-3 rounded-xl border border-border px-4 py-3.5">
      <div className="flex items-center gap-3">
        <span className="flex">
          {group.members.map((member, index) => (
            <span
              key={member.uid}
              className="grid size-5.5 place-items-center rounded-full border-[1.5px] border-background bg-chip text-[9px] text-secondary-foreground"
              style={index > 0 ? { marginLeft: "-6px" } : undefined}
            >
              {member.name.slice(0, 2).toUpperCase()}
            </span>
          ))}
        </span>
        <span className="min-w-0 truncate text-field font-medium">{group.name}</span>
        <span className="ml-auto shrink-0 text-micro text-muted-foreground">
          {t("groups.peopleCount", { count: group.members.length })}
        </span>
      </div>

      <p className="text-small text-secondary-foreground">
        {group.members.length > 0
          ? formatList(
              group.members.map((member) => member.name),
              code,
            )
          : t("groups.empty")}
      </p>

      <p className="border-t border-hair pt-3 text-micro text-muted-foreground">
        {group.listNames.length > 0
          ? t("groups.listsReached", { lists: formatList(group.listNames, code) })
          : t("groups.reachesNothing")}
      </p>

      <Button tone="secondary" scale="compact" className="self-start" onClick={onEdit}>
        {t("groups.editMembers")}
      </Button>
    </div>
  )
}
