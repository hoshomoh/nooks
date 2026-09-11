-- Whether a token reaches every List its Member can, rather than a named few.
--
-- Not the same as naming them all: a token scoped to everything keeps reaching a List
-- made next week, which is what somebody building a client actually wants. A token that
-- names Lists is still limited to exactly those.
ALTER TABLE access_token ADD COLUMN all_lists BOOLEAN NOT NULL DEFAULT FALSE;
