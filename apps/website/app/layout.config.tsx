import type { BaseLayoutProps } from "fumadocs-ui/layouts/shared"

import { Mark } from "@nooks/design/mark"

/** What every layout shares: the mark, the name, and the way back to the site. */
export const baseOptions: BaseLayoutProps = {
  nav: {
    title: (
      <span className="flex items-center gap-2">
        <Mark size={20} />
        <span className="font-medium">nooks</span>
      </span>
    ),
  },
  githubUrl: "https://github.com/hoshomoh/nooks",
}
