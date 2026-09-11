-- Which token put an Item on a List, when it was not a browser.
--
-- The Member is still recorded beside it: a token is somebody's access narrowed, never
-- an identity of its own, so a row says who and — only when it matters — what through.
--
-- ON DELETE SET NULL rather than CASCADE: revoking a key must not take what it added
-- off the List. The row simply stops saying which key it came through.
ALTER TABLE item ADD COLUMN added_by_token_id INTEGER
  REFERENCES access_token (id) ON DELETE SET NULL;
