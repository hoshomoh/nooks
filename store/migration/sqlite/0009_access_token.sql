-- A token is how anything that is not a browser reaches an Instance: a script, a
-- shortcut, an MCP client.
--
-- Only the hash is kept. The token itself is shown once, when it is made, and Nooks
-- cannot show it again — the same rule a password follows, for the same reason.
CREATE TABLE access_token (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  uid        TEXT    NOT NULL UNIQUE,
  -- Whose token it is. A token can never do more than the Member it belongs to, so an
  -- Admin cannot make one in somebody else's name.
  member_id  INTEGER NOT NULL REFERENCES member (id) ON DELETE CASCADE,
  -- What it is for, in the Member's words: "kitchen tablet", "shopping shortcut".
  name       TEXT    NOT NULL,
  token_hash TEXT    NOT NULL UNIQUE,
  -- READ or WRITE. Read means see and print; write adds ticking and adding.
  permission TEXT    NOT NULL,
  -- RFC 3339, or empty for a token that does not expire.
  expires_at TEXT    NOT NULL DEFAULT '',
  -- When it was last used, so a Member can tell which tokens are doing nothing.
  last_used_at TEXT  NOT NULL DEFAULT '',
  created_at TEXT    NOT NULL
);

CREATE INDEX idx_access_token_member_id ON access_token (member_id);

-- Which Lists a token may reach. A List that is not named here is invisible to it: not
-- forbidden, invisible — a token must not be able to discover what it cannot read.
CREATE TABLE access_token_list (
  token_id INTEGER NOT NULL REFERENCES access_token (id) ON DELETE CASCADE,
  list_id  INTEGER NOT NULL REFERENCES list (id) ON DELETE CASCADE,
  PRIMARY KEY (token_id, list_id)
);
