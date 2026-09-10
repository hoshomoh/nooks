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
