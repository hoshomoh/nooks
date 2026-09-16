-- Archiving a List: out of the sidebar, still there.
--
-- A household finishes with a List long before it wants to lose one. Deleting takes the
-- Items with it and cannot be undone; archiving only takes it out of the way, and it
-- stays searchable and restorable.
--
-- A property of the List rather than of one Member, like sharing and unlike pinning:
-- everybody who can reach a shared List sees the same thing when somebody finishes
-- with it. That is also what lets a row say who archived it.
ALTER TABLE list ADD COLUMN archived_at TEXT NOT NULL DEFAULT '';
ALTER TABLE list ADD COLUMN archived_by_id INTEGER NOT NULL DEFAULT 0;

CREATE INDEX idx_list_archived_at ON list (archived_at);
