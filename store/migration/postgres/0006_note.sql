-- A Note is the optional document attached to one Item.
--
-- It is stored as markdown text rather than as rows of blocks: markdown is already what
-- "export as plain text" has to produce, it is what the conflict rule in DESIGN.md §11
-- compares, and the five block types the design names map onto it exactly. The editor
-- renders blocks; the storage is text.
ALTER TABLE item ADD COLUMN note TEXT NOT NULL DEFAULT '';
