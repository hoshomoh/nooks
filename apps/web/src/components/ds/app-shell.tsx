import type { ReactNode } from "react"

import { Sidebar } from "./sidebar"
import type { List } from "@nooks/api"

/**
 * The signed-in layout, per DESIGN.md §5: a 258px sidebar beside everything else.
 */
export type AppShellProps = {
  instanceName: string
  memberName: string
  lists: List[]
  /** The List currently open, so the sidebar can mark it. */
  activeListUid?: string
  onSearch: () => void
  onAddList: () => void
  children: ReactNode
}

export function AppShell({
  instanceName,
  memberName,
  lists,
  activeListUid,
  onSearch,
  onAddList,
  children,
}: AppShellProps) {
  return (
    <div className="flex min-h-dvh bg-background">
      <Sidebar
        instanceName={instanceName}
        memberName={memberName}
        lists={lists}
        activeListUid={activeListUid}
        onSearch={onSearch}
        onAddList={onAddList}
      />
      <main className="flex min-w-0 flex-1 flex-col">{children}</main>
    </div>
  )
}
