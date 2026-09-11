/** Where a settings page lives. Only routes that exist are listed. */
export type SettingsRoute =
  | "/settings/general"
  | "/settings/tokens"
  | "/settings/members"
  | "/settings/groups"
  | "/settings/public"
  | "/settings/about"

/** What a nav entry counts, so the number beside it comes from the right query. */
export type SettingsCount = "tokens" | "members" | "groups"

/** One entry in the settings column. */
export interface SettingsSection {
  to: SettingsRoute
  labelKey: string
  /** Hidden from a Member: the server refuses them, so offering it would be a lie. */
  adminOnly?: boolean
  /** Which count sits at the right of the entry, when there is one. */
  counts?: SettingsCount
}

/**
 * The settings pages, in the order the design lists them.
 *
 * The settings column and the command palette both read this, so a page added here is
 * reachable from both without anybody remembering to add it twice.
 *
 * A Member sees three of them. The other three are an Admin's, and the server refuses
 * them either way — the filter is so the column does not offer a door that will not
 * open.
 */
export const SETTINGS_SECTIONS: SettingsSection[] = [
  { to: "/settings/general", labelKey: "settings.general" },
  { to: "/settings/tokens", labelKey: "settings.tokens", counts: "tokens" },
  { to: "/settings/members", labelKey: "settings.members", adminOnly: true, counts: "members" },
  { to: "/settings/groups", labelKey: "settings.groups", adminOnly: true, counts: "groups" },
  { to: "/settings/public", labelKey: "settings.publicList", adminOnly: true },
  { to: "/settings/about", labelKey: "settings.about" },
]
