import { useState } from "react"
import { useTranslation } from "react-i18next"

import { Button, type ButtonProps } from "./button"
import { browserCopy } from "@/lib/copy-text"

export interface CopyButtonProps extends Omit<ButtonProps, "onClick" | "children"> {
  text: string
  /** What the button says before it has been pressed. */
  label: string
  /** How the text reaches the clipboard. The browser's own ways unless a test says otherwise. */
  copy?: (text: string) => Promise<boolean>
}

/** Where a press has got to. */
type CopyState = "ready" | "copied" | "failed"

/**
 * Copy, then Copied, and only when it was.
 *
 * It said Copied the moment it was pressed, before the copy had happened, so a copy that
 * failed looked like one that worked. On a plain HTTP page it never got as far as saying
 * anything. When nothing can copy, it says to select the text instead, and the text
 * beside it is styled so one tap selects all of it.
 *
 * It never changes back. A timer that returned it to Copy would be telling a Member
 * their copy had expired, which is not a thing that happens. Give it a key that
 * changes with the text if a fresh button is wanted for fresh content.
 */
export function CopyButton({ text, label, copy = browserCopy, ...props }: CopyButtonProps) {
  const { t } = useTranslation()
  const [state, setState] = useState<CopyState>("ready")
  const says: Record<CopyState, string> = {
    ready: label,
    copied: t("action.copied"),
    failed: t("action.copyFailed"),
  }

  return (
    <Button {...props} onClick={async () => setState((await copy(text)) ? "copied" : "failed")}>
      {says[state]}
    </Button>
  )
}
