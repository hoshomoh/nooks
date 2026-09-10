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
  id                   BIGSERIAL PRIMARY KEY,
  uid                  TEXT      NOT NULL UNIQUE,
  name                 TEXT      NOT NULL,
  email                TEXT      NOT NULL UNIQUE,
  role                 TEXT      NOT NULL,
  password_hash        TEXT      NOT NULL,
  -- Set when an Admin creates the account with a temporary password. The Member must
  -- replace it before they can use Nooks.
  must_change_password BOOLEAN   NOT NULL DEFAULT FALSE,
  created_at           TEXT      NOT NULL,
  -- Empty until the Member has signed in at least once.
  last_signed_in_at    TEXT      NOT NULL DEFAULT ''
);

-- A signed-in browser. Only the hash of the token is stored, so a stolen database does
-- not hand over live sessions.
CREATE TABLE session (
  token_hash TEXT   NOT NULL PRIMARY KEY,
  member_id  BIGINT NOT NULL REFERENCES member (id) ON DELETE CASCADE,
  created_at TEXT   NOT NULL,
  expires_at TEXT   NOT NULL
);

CREATE INDEX idx_session_member_id ON session (member_id);
-- A Visitor asking for an account, and a Member asking to replace a forgotten
-- password. Nooks sends no email, so both wait here until an Admin acts on them.
CREATE TABLE join_request (
  id         BIGSERIAL PRIMARY KEY,
  uid        TEXT      NOT NULL UNIQUE,
  name       TEXT      NOT NULL,
  email      TEXT      NOT NULL,
  -- Anything the Visitor added, e.g. "It's Til, from upstairs". Optional.
  message    TEXT      NOT NULL DEFAULT '',
  -- PENDING, APPROVED or IGNORED. Ignoring is silent and never notifies the sender.
  status     TEXT      NOT NULL DEFAULT 'PENDING',
  created_at TEXT      NOT NULL,
  decided_at TEXT      NOT NULL DEFAULT ''
);

CREATE INDEX idx_join_request_status ON join_request (status);

CREATE TABLE reset_request (
  id         BIGSERIAL PRIMARY KEY,
  uid        TEXT      NOT NULL UNIQUE,
  member_id  BIGINT    NOT NULL REFERENCES member (id) ON DELETE CASCADE,
  status     TEXT      NOT NULL DEFAULT 'PENDING',
  created_at TEXT      NOT NULL,
  decided_at TEXT      NOT NULL DEFAULT '',
  -- An approval lets the Member set a new password, and expires in an hour so an
  -- approved-and-forgotten request does not stay usable.
  expires_at TEXT      NOT NULL DEFAULT ''
);

CREATE INDEX idx_reset_request_member_id ON reset_request (member_id);
