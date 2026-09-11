-- What became of a request, once an Admin decided.
--
-- An entry that has been acted on stops offering to act: whichever Admin got there
-- first, the others should see what happened rather than a button that now does
-- nothing. Empty means nobody has decided yet.
ALTER TABLE activity ADD COLUMN outcome TEXT NOT NULL DEFAULT '';
