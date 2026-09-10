-- One index for everything a Member wrote: List names now, Item labels and quantities
-- now, Note blocks when Notes land. Rows are maintained by the store on write rather
-- than by triggers, so the logic lives in one testable place and reads the same on
-- both drivers.
--
-- The tsvector is generated from the text, so it can never fall out of step with it.
CREATE TABLE search_index (
  kind    TEXT   NOT NULL,
  uid     TEXT   NOT NULL,
  list_id BIGINT NOT NULL DEFAULT 0,
  text    TEXT   NOT NULL,
  search  TSVECTOR GENERATED ALWAYS AS (to_tsvector('simple', text)) STORED,
  PRIMARY KEY (kind, uid)
);

CREATE INDEX idx_search_index_search ON search_index USING GIN (search);
