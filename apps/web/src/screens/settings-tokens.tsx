import { useState } from "react"
import { useMutation, useQueryClient, useSuspenseQuery } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { Permission, type AccessToken } from "@nooks/api"

import { AddTokenDialog, type NewToken, type TokenSecret } from "@/components/ds/add-token-dialog"
import { Button } from "@/components/ds/button"
import { ConfirmDialog } from "@/components/ds/confirm-dialog"
import { EmptyState } from "@/components/ds/empty-state"
import { IconButton } from "@/components/ds/icon-button"
import { SettingsShell } from "@/components/ds/settings-shell"
import { tokenClient } from "@/lib/api"
import { tokensQuery } from "@/lib/token-queries"
import type { Translate } from "@/lib/translate"
import { useMomentLabel, type FormatMoment } from "@/lib/use-moment-label"
import { useSignedInData } from "@/lib/use-signed-in-data"

/**
 * The Access tokens page: the keys a Member has cut, and a way to stop one working.
 *
 * A token is never readable again, so this page describes rather than shows: what it is
 * for, what it reaches, and when it was last used. Last used is the useful column —
 * it is how a Member recognises the one they have forgotten about.
 */
export function SettingsTokensScreen() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { lists, member } = useSignedInData()
  const tokens = useSuspenseQuery(tokensQuery).data.tokens

  const [adding, setAdding] = useState(false)
  const [secret, setSecret] = useState<TokenSecret | null>(null)
  const [revoking, setRevoking] = useState<AccessToken | null>(null)

  const refresh = () => queryClient.invalidateQueries({ queryKey: tokensQuery.queryKey })

  const add = useMutation({
    mutationFn: (fresh: NewToken) => tokenClient.createAccessToken(fresh),
    onSuccess: async (res) => {
      await refresh()
      setSecret({ tokenName: res.token?.name ?? "", secret: res.secret })
    },
  })

  const revoke = useMutation({
    mutationFn: (tokenUid: string) => tokenClient.revokeAccessToken({ tokenUid }),
    onSuccess: refresh,
  })

  return (
    <SettingsShell
      active="/settings/tokens"
      crumb={t("settings.tokens")}
      counts={{ "/settings/tokens": tokens.length }}
    >
      <header className="flex items-start gap-6">
        <div className="min-w-0">
          <h1 className="mb-1.5 text-page">{t("tokens.title")}</h1>
          <p className="max-w-135 text-chrome leading-[1.55] text-secondary-foreground">
            {t("tokens.blurb")}
          </p>
        </div>
        <Button className="ml-auto shrink-0" onClick={() => setAdding(true)}>
          {t("tokens.add")}
        </Button>
      </header>

      {tokens.length === 0 ? (
        <EmptyState title={t("tokens.emptyTitle")} body={t("tokens.emptyBody")} />
      ) : (
        <TokenTable tokens={tokens} ownName={member?.name ?? ""} onRevoke={setRevoking} />
      )}

      <AddTokenDialog
        open={adding}
        onOpenChange={(open) => {
          setAdding(open)
          if (!open) {
            add.reset()
          }
        }}
        lists={lists}
        onAdd={(fresh) => add.mutate(fresh)}
        secret={secret ?? undefined}
        onSecretRead={() => {
          setSecret(null)
          setAdding(false)
          add.reset()
        }}
      />

      <ConfirmDialog
        open={revoking !== null}
        onOpenChange={(open) => !open && setRevoking(null)}
        title={t("tokens.revokeTitle", { name: revoking?.name ?? "" })}
        blurb={t("tokens.revokeBlurb")}
        confirmLabel={t("tokens.revokeConfirm")}
        destructive
        onConfirm={() => revoking && revoke.mutate(revoking.uid)}
      />
    </SettingsShell>
  )
}

interface TokenTableProps {
  tokens: AccessToken[]
  /** The reader's own name, so somebody else's token can say whose it is. */
  ownName: string
  onRevoke: (token: AccessToken) => void
}

/** The tokens, as a table of what each one is and reaches. */
function TokenTable({ tokens, ownName, onRevoke }: TokenTableProps) {
  const { t } = useTranslation()
  const moment = useMomentLabel()

  return (
    <div className="flex flex-col">
      <div className="grid grid-cols-[1fr_120px_140px_24px] items-center gap-4 border-b border-border pb-2.5 text-label text-muted-foreground uppercase">
        <span>{t("tokens.tokenColumn")}</span>
        <span>{t("tokens.permission")}</span>
        <span>{t("tokens.lastUsed")}</span>
        <span />
      </div>

      {tokens.map((token) => (
        <TokenRow
          key={token.uid}
          token={token}
          ownName={ownName}
          moment={moment}
          onRevoke={() => onRevoke(token)}
        />
      ))}
    </div>
  )
}

interface TokenRowProps {
  token: AccessToken
  ownName: string
  moment: FormatMoment
  onRevoke: () => void
}

/** One token: what it is for, what it reaches, and when it was last used. */
function TokenRow({ token, ownName, moment, onRevoke }: TokenRowProps) {
  const { t } = useTranslation()
  const mine = token.memberName === ownName

  return (
    <div className="grid min-h-13 grid-cols-[1fr_120px_140px_24px] items-center gap-4 border-b border-hair">
      <div className="flex min-w-0 flex-col gap-0.5">
        <span className="truncate text-field">{token.name}</span>
        <span className="truncate text-micro text-muted-foreground">
          {mine
            ? describeScope(token, t, moment)
            : t("tokens.someoneElses", { name: token.memberName })}
        </span>
      </div>

      <span className="text-meta text-secondary-foreground">
        {t(token.permission === Permission.READ ? "tokens.read" : "tokens.write")}
      </span>

      <span className="text-meta text-secondary-foreground">
        {token.lastUsedAt ? moment(token.lastUsedAt) : t("tokens.neverUsed")}
      </span>

      <IconButton
        name="close"
        scale="compact"
        label={t("tokens.revoke", { name: token.name })}
        onClick={onRevoke}
      />
    </div>
  )
}

/**
 * What the line under a Member's own token says: its Lists, then when it runs out.
 *
 * Only ever their own. The server does not send another Member's List names, and a page
 * that had somewhere to put them would be a page waiting for the server to slip.
 */
function describeScope(token: AccessToken, t: Translate, moment: FormatMoment): string {
  const lists = token.listNames.join(", ") || t("tokens.noLists")
  if (!token.expiresAt) {
    return `${lists} · ${t("tokens.neverExpires")}`
  }
  return `${lists} · ${t("tokens.expires", { date: moment(token.expiresAt) })}`
}
