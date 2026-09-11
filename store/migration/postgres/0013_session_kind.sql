-- What a session token is for.
--
-- A refresh token is long-lived and only ever travels in an HttpOnly cookie, so no
-- script can read it. An access token is short-lived and is handed to a caller in a
-- response body, which is the only way a REST client can hold a credential at all.
-- Splitting them means the thing that lasts a month is never the thing in a body.
ALTER TABLE session ADD COLUMN kind TEXT NOT NULL DEFAULT 'REFRESH';

-- Which refresh token minted an access token, so that signing out takes the access
-- tokens with it. Empty on a refresh token, which is nobody's child.
ALTER TABLE session ADD COLUMN parent_hash TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_session_parent_hash ON session (parent_hash);
