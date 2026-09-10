# Nook

Nook is a household todo app you run on your own machine. It is built around one primitive — a list of
items — and it treats the printed page as a real output rather than a fallback.

This file fixes the language. Every term below has one name; the `_Avoid_` line lists the names that
mean the same thing and must not be used in code, API, UI copy, or commit messages.

## Language

**Instance**:
One running copy of Nook and everything in it. It has a name, chosen at first run, that shows in the
sidebar and on printed pages.
_Avoid_: Workspace, tenant, organisation, server

**Instance name**:
The mutable display label of the Instance, e.g. "Brunnen Street".
_Avoid_: Site name, title, household name

**Member**:
A person with an account on the Instance.
_Avoid_: User, account holder, resident

**Admin**:
The Member role that can reach instance settings, approve requests, and manage Members and Groups.
The Member who completes first run is the first Admin.
_Avoid_: Owner, superuser, root

**Group**:
A named set of Members that exists only as a shortcut for sharing. Sharing a List with a Group shares
it with everyone in the Group, including Members added to it later. A Group carries no permissions of
its own.
_Avoid_: Team, role, circle

**List**:
A named, ordered collection of Items. The only container in Nook — there are no folders, projects, or
boards.
_Avoid_: Project, board, folder, collection

**List sharing**:
Who can reach a List: private, everyone on the Instance, or specific Members and Groups. Independent
of the Public list.
_Avoid_: Visibility, permission level, ACL

**Pinned list**:
A List a Member has pinned to the top of their own sidebar. Pinning is per-Member and never affects
anyone else's sidebar.
_Avoid_: Favourite, starred, bookmarked

**Item**:
One line on a List. It carries a label and, optionally, a quantity, a due date and a Note. Every Item
records the Member who added it.
_Avoid_: Task, todo, entry, row

**Item label**:
The text of an Item — the part the Member typed.
_Avoid_: Title, name, content

**Quantity**:
The optional free-text amount on an Item, e.g. "2" or "1 kg". It is text, never a number.
_Avoid_: Amount, count, qty field

**Tick**:
Marking an Item done, and the record that it was done. Ticks never conflict: a tick is a tick whoever
made it.
_Avoid_: Complete, check, toggle, finish

**Note**:
The optional document attached to one Item, made of blocks — paragraph, heading, checklist line,
quote, code. The same Note renders in the side sheet and full screen; only the type scale differs.
_Avoid_: Description, body, comment, details

**Note block**:
One line-level element of a Note, with a type. Markdown shorthand converts a block as the Member types
but is never displayed back to them.
_Avoid_: Node, element, paragraph (as a generic term)

**Side sheet**:
The 520px panel that slides over the list pane to show one Item. Opening an Item never replaces the
List with a page.
_Avoid_: Drawer, modal, detail page

**Today**:
The view gathering dated Items from every List the Member can reach. Undated Items never appear in it.
_Avoid_: Inbox, dashboard, home

**Upcoming**:
The view gathering dated Items for the next two weeks, grouped by day. Its calendar layout is a lens on
the same dated Items, not a second home for Lists.
_Avoid_: Schedule, agenda, planner

**Public list**:
The single List an Instance may expose read-only at a stable address, reachable with no account. There
is at most one at a time, there is no password on it, and search engines are not blocked from it.
_Avoid_: Shared link, published list, guest access

**Visitor**:
Someone reading the Public list without an account. A Visitor cannot tick, add, or print.
_Avoid_: Guest, anonymous user, public user

**Join request**:
A Visitor's request for an account, carrying a name, an email and an optional message. It waits in the
Instance until an Admin approves or ignores it. Ignoring is silent and never notifies the sender.
_Avoid_: Invitation, signup, application

**Reset request**:
A Member's request to replace a forgotten password. An Admin approves it out of band; approval expires
in an hour. The Member then sets the new password themselves.
_Avoid_: Password reset email, recovery link

**Temporary password**:
The password an Admin sets when adding a Member directly, read out once. The Member must replace it on
first sign-in.
_Avoid_: Invite code, initial password, default password

**Access token**:
A secret that lets something other than a browser reach Nook — a script, a shortcut, an assistant over
MCP. It is scoped to named Lists and to a set of permissions, belongs to the Member who made it, and
is shown exactly once. Nook stores a hash.
_Avoid_: API key, PAT, credential, secret (as a noun for this)

**Token scope**:
The named Lists an Access token can reach. Unpicked Lists are invisible to the token — it cannot see
that they exist.
_Avoid_: Permissions (that is the separate list of allowed operations)

**Activity**:
The in-app record of things that need a Member's attention or explain a change — join requests, reset
requests, shares, token use. Nook has no mail server, so Activity is the only place these surface.
_Avoid_: Notifications, feed, audit log, alerts

**Print sheet**:
The A4 rendering of one List: title, rule, items with real checkboxes, and blank lines to write on. A
deliverable, not a screenshot.
_Avoid_: PDF export, print preview, printout

## Principles

- One list primitive. Anything that feels like a second container is a view over Lists.
- Nook sends no email. Anything that would need one waits in Activity instead.
- Nothing is generated on a Member's behalf: a long Note previews as its own first line and a count of
  what is left, never a summary.
- A row is 44px whether it carries one piece of metadata or five.
- Four meaning colours, each with exactly one job: accent/shared, done/presence, offline, overdue/error.
