-- The Items a count has to walk, and nothing else in them.
--
-- The counts beside a List are written onto it, and recount rewrites them by counting
-- rather than by adding and subtracting: a counter kept by arithmetic drifts the first
-- time a path forgets to adjust it, and the number a Member reads is then wrong with
-- nothing anywhere to notice. That trade is the right one and these keep it cheap.
--
-- idx_item_list_id is (list_id, position), so a count could find a List's Items but not
-- tell a ticked one from an open one without fetching every row to look. Counting was
-- therefore work proportional to everything on the List, done twice, on every add, tick
-- and delete — and ticking is the thing a household does most. It showed as a tick that
-- grew with the List: on one machine, a few milliseconds at a hundred Items and most of
-- a second at a hundred thousand, rising in step with what was on it.
--
-- One index per count, holding only the rows that count belongs to, so each is a range
-- scan that never opens a row. The same ticks come back to a few milliseconds and stay
-- there for as long as a household is likely to go. BenchmarkTickOnACrowdedList is how
-- to see it rather than take it on trust.
CREATE INDEX idx_item_open_on_list ON item (list_id) WHERE deleted_at = '' AND done_at = '';
CREATE INDEX idx_item_done_on_list ON item (list_id) WHERE deleted_at = '' AND done_at <> '';
