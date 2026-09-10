-- A Group is a named set of Members that exists only as a shortcut for sharing. It
-- carries no permissions of its own.
CREATE TABLE member_group (
  id         BIGSERIAL PRIMARY KEY,
  uid        TEXT    NOT NULL UNIQUE,
  name       TEXT    NOT NULL,
  created_at TEXT    NOT NULL
);

CREATE TABLE group_member (
  group_id  BIGINT NOT NULL REFERENCES member_group (id) ON DELETE CASCADE,
  member_id BIGINT NOT NULL REFERENCES member (id) ON DELETE CASCADE,
  PRIMARY KEY (group_id, member_id)
);

-- Who a List is shared with by name. A row here is one Member or one Group; sharing
-- with a Group reaches everyone in it, including people added to it later.
CREATE TABLE list_share (
  id        BIGSERIAL PRIMARY KEY,
  list_id   BIGINT NOT NULL REFERENCES list (id) ON DELETE CASCADE,
  member_id BIGINT REFERENCES member (id) ON DELETE CASCADE,
  group_id  BIGINT REFERENCES member_group (id) ON DELETE CASCADE
);

CREATE INDEX idx_list_share_list_id ON list_share (list_id);
CREATE INDEX idx_list_share_member_id ON list_share (member_id);

-- Activity is where everything that would be an email elsewhere waits. Nooks has no
-- mail server, so this is the only place these surface.
CREATE TABLE activity (
  id         BIGSERIAL PRIMARY KEY,
  uid        TEXT    NOT NULL UNIQUE,
  -- Who should see it. Join and reset requests go to every Admin, so this is the
  -- Member the entry is *for*, one row each.
  member_id  BIGINT NOT NULL REFERENCES member (id) ON DELETE CASCADE,
  -- JOIN_REQUEST, RESET_REQUEST, LIST_SHARED or CONFLICT.
  kind       TEXT    NOT NULL,
  -- What it says, already written by whoever raised it.
  text       TEXT    NOT NULL,
  -- What it points at: a request uid, a list uid, or empty.
  target_uid TEXT    NOT NULL DEFAULT '',
  read_at    TEXT    NOT NULL DEFAULT '',
  created_at TEXT    NOT NULL
);

CREATE INDEX idx_activity_member_id ON activity (member_id, created_at);
