-- Instance configuration is a handful of key/value rows rather than a one-row
-- table, so adding a setting never needs a migration.
CREATE TABLE setting (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
-- A Member is a person with an account on this Instance. See CONTEXT.md.
--
-- Timestamps are RFC3339 text and booleans are 0/1 integers in both drivers. That is
-- not either engine's idiom, but it keeps one Go code path reading both, and the store
-- is the only thing that ever reads these columns.
CREATE TABLE member (
  id                   INTEGER PRIMARY KEY AUTOINCREMENT,
  uid                  TEXT    NOT NULL UNIQUE,
  name                 TEXT    NOT NULL,
  email                TEXT    NOT NULL UNIQUE,
  role                 TEXT    NOT NULL,
  password_hash        TEXT    NOT NULL,
  -- Set when an Admin creates the account with a temporary password. The Member must
  -- replace it before they can use Nooks.
  must_change_password BOOLEAN NOT NULL DEFAULT FALSE,
  created_at           TEXT    NOT NULL,
  -- Empty until the Member has signed in at least once.
  last_signed_in_at    TEXT    NOT NULL DEFAULT ''
);

-- A signed-in browser. Only the hash of the token is stored, so a stolen database does
-- not hand over live sessions.
CREATE TABLE session (
  token_hash TEXT    NOT NULL PRIMARY KEY,
  member_id  INTEGER NOT NULL REFERENCES member (id) ON DELETE CASCADE,
  created_at TEXT    NOT NULL,
  expires_at TEXT    NOT NULL
);

CREATE INDEX idx_session_member_id ON session (member_id);
-- A Visitor asking for an account, and a Member asking to replace a forgotten
-- password. Nooks sends no email, so both wait here until an Admin acts on them.
CREATE TABLE join_request (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  uid        TEXT    NOT NULL UNIQUE,
  name       TEXT    NOT NULL,
  email      TEXT    NOT NULL,
  -- Anything the Visitor added, e.g. "It's Til, from upstairs". Optional.
  message    TEXT    NOT NULL DEFAULT '',
  -- PENDING, APPROVED or IGNORED. Ignoring is silent and never notifies the sender.
  status     TEXT    NOT NULL DEFAULT 'PENDING',
  created_at TEXT    NOT NULL,
  decided_at TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX idx_join_request_status ON join_request (status);

CREATE TABLE reset_request (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  uid        TEXT    NOT NULL UNIQUE,
  member_id  INTEGER NOT NULL REFERENCES member (id) ON DELETE CASCADE,
  status     TEXT    NOT NULL DEFAULT 'PENDING',
  created_at TEXT    NOT NULL,
  decided_at TEXT    NOT NULL DEFAULT '',
  -- An approval lets the Member set a new password, and expires in an hour so an
  -- approved-and-forgotten request does not stay usable.
  expires_at TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX idx_reset_request_member_id ON reset_request (member_id);
-- A List is a named, ordered collection of Items — the only container in Nooks.
CREATE TABLE list (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  uid        TEXT    NOT NULL UNIQUE,
  name       TEXT    NOT NULL,
  owner_id   INTEGER NOT NULL REFERENCES member (id) ON DELETE CASCADE,
  -- PRIVATE, INSTANCE (everyone here) or SPECIFIC (named Members and Groups).
  sharing    TEXT    NOT NULL DEFAULT 'PRIVATE',
  -- When false, whoever it is shared with may see and print but not tick or add.
  can_edit   BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TEXT    NOT NULL,
  updated_at TEXT    NOT NULL,
  -- Deleting is soft: a List removed by mistake is recoverable until it is purged.
  deleted_at TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX idx_list_owner_id ON list (owner_id);

-- One line on a List.
CREATE TABLE item (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  uid         TEXT    NOT NULL UNIQUE,
  list_id     INTEGER NOT NULL REFERENCES list (id) ON DELETE CASCADE,
  label       TEXT    NOT NULL,
  -- Free text, never a number: "2", "1 kg", "the good beans".
  quantity    TEXT    NOT NULL DEFAULT '',
  -- A date, not a time: YYYY-MM-DD, or empty for an Item with no due date.
  due_on      TEXT    NOT NULL DEFAULT '',
  -- Manual order. A float so an Item can be dropped between two others without
  -- renumbering the rest.
  position    REAL    NOT NULL,
  -- Empty until the Item is ticked. A tick is a tick whoever made it.
  done_at     TEXT    NOT NULL DEFAULT '',
  done_by_id  INTEGER REFERENCES member (id) ON DELETE SET NULL,
  added_by_id INTEGER NOT NULL REFERENCES member (id) ON DELETE CASCADE,
  created_at  TEXT    NOT NULL,
  updated_at  TEXT    NOT NULL,
  deleted_at  TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX idx_item_list_id ON item (list_id, position);
CREATE INDEX idx_item_due_on ON item (due_on);

-- Pinning is per-Member and never affects anyone else's sidebar.
CREATE TABLE list_pin (
  member_id INTEGER NOT NULL REFERENCES member (id) ON DELETE CASCADE,
  list_id   INTEGER NOT NULL REFERENCES list (id) ON DELETE CASCADE,
  PRIMARY KEY (member_id, list_id)
);
-- One index for everything a Member wrote: List names now, Item labels and quantities
-- now, Note blocks when Notes land. Rows are maintained by the store on write rather
-- than by triggers, so the logic lives in one testable place and reads the same on
-- both drivers.
--
-- FTS5 is SQLite's own full-text engine: it ranks with bm25 and matches prefixes.
CREATE VIRTUAL TABLE search_index USING fts5 (
  kind UNINDEXED,
  uid UNINDEXED,
  list_id UNINDEXED,
  text
);
