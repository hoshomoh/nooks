import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Link, useNavigate } from "@tanstack/react-router"
import { RequestStatus } from "@nooks/api"

import { AuthShell } from "@/components/ds/auth-shell"
import { Button } from "@/components/ds/button"
import { Field } from "@/components/ds/field"
import { authClient } from "@/lib/api"
import { messageFrom } from "@/lib/errors"
import {
  forgetPendingRequest,
  readPendingRequest,
  rememberPendingRequest,
} from "@/lib/pending-request"

/**
 * A forgotten password. Nooks sends no email: an Admin checks it is really you however
 * they like, approves, and you set the new password yourself.
 */
export function ForgotPassword() {
  const [requestUid, setRequestUid] = useState(() => readPendingRequest("reset"))

  if (!requestUid) {
    return <AskForReset onSent={setRequestUid} />
  }
  return <CheckResetRequest requestUid={requestUid} onStartOver={() => setRequestUid(null)} />
}

/** The quiet link back, used by both states of this screen. */
function BackToSignIn() {
  const { t } = useTranslation()
  return (
    <Link to="/sign-in" className="text-shared">
      {t("auth.reset.backToSignIn")}
    </Link>
  )
}

type AskForResetProps = {
  onSent: (requestUid: string) => void
}

function AskForReset({ onSent }: AskForResetProps) {
  const { t } = useTranslation()
  const [emailOrName, setEmailOrName] = useState("")

  const requestReset = useMutation({
    mutationFn: () => authClient.requestPasswordReset({ emailOrName }),
    onSuccess: (res) => {
      rememberPendingRequest("reset", res.requestUid)
      onSent(res.requestUid)
    },
  })

  return (
    <AuthShell
      title={t("auth.reset.title")}
      blurb={t("auth.reset.blurb")}
      footerLeft={<BackToSignIn />}
      footerRight={t("auth.reset.footer")}
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          requestReset.mutate()
        }}
      >
        <Field
          label={t("auth.reset.emailOrName")}
          value={emailOrName}
          onChange={(event) => setEmailOrName(event.target.value)}
          hint={t("auth.reset.emailOrNameHint")}
          autoComplete="email"
          required
        />

        {requestReset.isError && (
          <p className="text-destructive text-secondary">{messageFrom(requestReset.error)}</p>
        )}

        <Button type="submit" disabled={requestReset.isPending} className="self-start">
          {requestReset.isPending ? t("auth.join.submitting") : t("auth.join.submit")}
        </Button>
      </form>
    </AuthShell>
  )
}

type CheckResetRequestProps = {
  requestUid: string
  onStartOver: () => void
}

function CheckResetRequest({ requestUid, onStartOver }: CheckResetRequestProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [newPassword, setNewPassword] = useState("")

  const request = useQuery({
    queryKey: ["reset-request", requestUid],
    queryFn: () => authClient.getResetRequest({ requestUid }),
  })

  const completeReset = useMutation({
    mutationFn: () => authClient.completePasswordReset({ requestUid, newPassword }),
    onSuccess: async () => {
      forgetPendingRequest("reset")
      await queryClient.invalidateQueries()
      await navigate({ to: "/" })
    },
  })

  if (request.data?.status !== RequestStatus.APPROVED) {
    return (
      <AuthShell
        eyebrow={t("auth.join.sentEyebrow")}
        title={t("auth.reset.sentTitle")}
        blurb={t("auth.reset.sentBlurb")}
        footerLeft={<BackToSignIn />}
        footerRight={t("auth.reset.sentFooter")}
      >
        <div className="flex items-center gap-3">
          <Button
            tone="secondary"
            onClick={() => request.refetch()}
            disabled={request.isFetching}
          >
            {request.isFetching ? t("auth.join.checking") : t("auth.join.checkAgain")}
          </Button>
          <button type="button" onClick={onStartOver} className="text-muted-foreground text-small">
            {t("auth.join.askAgain")}
          </button>
        </div>
      </AuthShell>
    )
  }

  return (
    <AuthShell
      eyebrow={t("auth.reset.approvedEyebrow")}
      title={t("auth.reset.approvedTitle")}
      blurb={t("auth.reset.approvedBlurb")}
      footerRight=""
    >
      <form
        className="flex flex-col gap-5"
        onSubmit={(event) => {
          event.preventDefault()
          completeReset.mutate()
        }}
      >
        <Field
          label={t("auth.reset.newPassword")}
          type="password"
          value={newPassword}
          onChange={(event) => setNewPassword(event.target.value)}
          hint={t("auth.setup.passwordHint")}
          autoComplete="new-password"
          required
        />

        {completeReset.isError && (
          <p className="text-destructive text-secondary">{messageFrom(completeReset.error)}</p>
        )}

        <Button type="submit" disabled={completeReset.isPending} className="self-start">
          {completeReset.isPending ? t("auth.replacePassword.submitting") : t("auth.reset.submit")}
        </Button>
      </form>
    </AuthShell>
  )
}
