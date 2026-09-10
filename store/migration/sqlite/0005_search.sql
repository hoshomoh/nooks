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
