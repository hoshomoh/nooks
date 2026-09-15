import { useTranslation } from "react-i18next"

import { Button } from "./button"
import { CopyButton } from "./copy-button"

export interface SecretOnceProps {
  /** The secret itself, in clear. */
  secret: string
  /** Who or what it is for, e.g. "Til can sign in with this". */
  title: string
  /** Why it will not be shown again. */
  blurb: string
  onDone: () => void
}

/**
 * A secret shown once, per DESIGN.md §9.
 *
 * An accent panel, the secret in mono, a Copy, and a plain sentence saying it cannot be
 * shown again. **Never a warning triangle**: nothing has gone wrong, and dressing a
 * normal step as a hazard teaches a Member to ignore the ones that are.
 *
 * It takes the place of the form that asked for it, so it arrives the way a dialog
 * does rather than by the box changing under the Member's hand: the same short drop
 * the dialog itself came in on.
 */
export function SecretOnce({ secret, title, blurb, onDone }: SecretOnceProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-4 px-6.5 py-6 translate-y-0 transition-[opacity,translate] duration-160 ease-sheet starting:-translate-y-1.5 starting:opacity-0">
      <div className="flex flex-col gap-2">
        <h2 className="text-dialog">{title}</h2>
        <p className="text-field text-secondary-foreground">{blurb}</p>
      </div>

      <div className="flex items-center gap-3 rounded-xl border border-shared-line bg-shared-bg px-4 py-3.5">
        <code className="min-w-0 font-mono text-meta break-all text-foreground">{secret}</code>
        <CopyButton
          text={secret}
          label={t("action.copy")}
          tone="secondary"
          scale="compact"
          className="ml-auto shrink-0"
        />
      </div>

      <div className="flex items-center">
        <span className="flex-1" />
        <Button onClick={onDone}>{t("members.secretDone")}</Button>
      </div>
    </div>
  )
}
