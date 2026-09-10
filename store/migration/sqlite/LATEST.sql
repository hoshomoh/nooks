-- Instance configuration is a handful of key/value rows rather than a one-row
-- table, so adding a setting never needs a migration.
CREATE TABLE setting (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
