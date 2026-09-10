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
  must_change_password INTEGER NOT NULL DEFAULT 0,
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
