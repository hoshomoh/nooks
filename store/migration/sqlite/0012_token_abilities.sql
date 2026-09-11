-- What a token may do, as three abilities rather than one level.
--
-- "Read and write" was a guess at the shape of the question. The real one is what a
-- caller is for: a shortcut that adds shopping needs to add, a recipe importer needs to
-- read, and almost nothing needs to delete — which is why that one is off by default and
-- has to be asked for.
ALTER TABLE access_token ADD COLUMN can_read BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE access_token ADD COLUMN can_write BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE access_token ADD COLUMN can_delete BOOLEAN NOT NULL DEFAULT FALSE;

-- Existing tokens keep exactly what they had. A WRITE token could already delete, so it
-- keeps that too: a migration must not quietly take an ability away from a key that is
-- in use somewhere.
UPDATE access_token SET can_write = TRUE, can_delete = TRUE WHERE permission = 'WRITE';

ALTER TABLE access_token DROP COLUMN permission;
