import { useState } from "react"
import { useTranslation } from "react-i18next"

import { Button, type ButtonProps } from "./button"

export interface CopyButtonProps extends Omit<ButtonProps, "onClick" | "children"> {
  text: string
  /** What the button says before it has been pressed. */
  label: string
}

/**
 * Copy, then Copied.
 *
 * It never changes back. A timer that returned it to Copy would be telling a Member
 * their copy had expired, which is not a thing that happens. Give it a key that
 * changes with the text if a fresh button is wanted for fresh content.
 */
export function CopyButton({ text, label, ...props }: CopyButtonProps) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  return (
    <Button
      {...props}
      onClick={() => {
        void navigator.clipboard.writeText(text)
        setCopied(true)
      }}
    >
      {copied ? t("action.copied") : label}
    </Button>
  )
}
