/** Where a settings page lives. Only routes that exist are listed. */
export type SettingsRoute =
  | "/settings/members"
  | "/settings/groups"
  | "/settings/tokens"
  | "/settings/public"

/** One entry in the settings column. */
export interface SettingsSection {
  to: SettingsRoute
  labelKey: string
}

/**
 * The settings pages, in the order the design lists them.
 *
 * The settings column and the command palette both read this, so a page added here is
 * reachable from both without anybody remembering to add it twice.
 *
 * General and About arrive with the milestones that build them. A nav entry that goes
 * nowhere is worse than one that is not there yet.
 */
export const SETTINGS_SECTIONS: SettingsSection[] = [
  { to: "/settings/members", labelKey: "settings.members" },
  { to: "/settings/groups", labelKey: "settings.groups" },
  { to: "/settings/tokens", labelKey: "settings.tokens" },
  { to: "/settings/public", labelKey: "settings.publicList" },
]
