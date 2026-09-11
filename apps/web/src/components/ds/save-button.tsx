import { useTranslation } from "react-i18next"

import { Button } from "./button"

export interface SaveButtonProps {
  /**
   * Whether the form still matches what was last saved.
   *
   * It does both jobs: there is nothing to save, and the "Saved" it is showing is still
   * true. Tying the word to this rather than to a timer means it goes away exactly when
   * it stops being true, and never while it still is.
   */
  unchanged: boolean
  /** The save is in flight. */
  pending: boolean
  /** The last save succeeded. */
  succeeded: boolean
  /** What the button says when there is something to save. Defaults to Save. */
  label?: string
  /** Omit inside a form, where the submit handler does the work. */
  onClick?: () => void
  /** Anything else that has to be true before it can be pressed. */
  disabled?: boolean
}

/**
 * The button at the foot of a settings form, and the word that says it worked.
 *
 * One component because a settings page that says "Saved" forever is a settings page
 * that has stopped telling the truth, and that is not a bug worth finding twice.
 */
export function SaveButton({
  unchanged,
  pending,
  succeeded,
  label,
  onClick,
  disabled,
}: SaveButtonProps) {
  const { t } = useTranslation()

  return (
    <div className="flex items-center gap-3">
      <Button
        type={onClick ? "button" : "submit"}
        onClick={onClick}
        disabled={disabled || pending || unchanged}
      >
        {pending ? t("action.saving") : (label ?? t("action.save"))}
      </Button>
      {succeeded && unchanged && !pending && (
        <span className="text-micro text-muted-foreground">{t("action.saved")}</span>
      )}
    </div>
  )
}
