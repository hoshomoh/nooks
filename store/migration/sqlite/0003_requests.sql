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
