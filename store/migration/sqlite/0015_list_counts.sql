-- What is left on a List, kept on the List.
--
-- Reading a page of Lists used to add up every Item in the Instance first: the counts
-- came from a GROUP BY over the whole item table, joined on, and the filters and the
-- order both depended on it. That is work proportional to everything there is, done to
-- draw twenty-five rows, and on a hundred thousand Lists it took nine seconds.
--
-- Held here instead, and rewritten by the store whenever an Item is added, ticked or
-- removed. The counts are what a Member reads beside a name, what Active and Completed
-- are decided by, and what "most open items" sorts on, so they are properties of the
-- List in every sense but where they were stored.
ALTER TABLE list ADD COLUMN open_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE list ADD COLUMN done_count INTEGER NOT NULL DEFAULT 0;

UPDATE list SET
  open_count = (SELECT COUNT(*) FROM item WHERE item.list_id = list.id AND item.deleted_at = '' AND item.done_at = ''),
  done_count = (SELECT COUNT(*) FROM item WHERE item.list_id = list.id AND item.deleted_at = '' AND item.done_at <> '');


-- The indexes a page of Lists is read through.
--
-- Every one of these is partial: "WHERE deleted_at = '' AND archived_at = ''" is true of
-- almost every List and is on every read that is not the Archived filter, so leaving the
-- rest out keeps the index the size of what is actually looked at.
--
-- The point of them is the ORDER BY, not the WHERE. Given an index whose order is the
-- order that was asked for, the database walks it and stops at twenty-five. Given only
-- an index that matches the filter, it has to find every match and sort the lot to know
-- which twenty-five come first: on a hundred thousand Lists that is the difference
-- between two milliseconds and a second.
CREATE INDEX idx_list_live_name ON list (lower(name))
  WHERE deleted_at = '' AND archived_at = '';
CREATE INDEX idx_list_live_updated ON list (updated_at DESC)
  WHERE deleted_at = '' AND archived_at = '';
CREATE INDEX idx_list_live_open ON list (open_count DESC, lower(name) ASC)
  WHERE deleted_at = '' AND archived_at = '';

-- My lists, which is one owner's, in name order.
CREATE INDEX idx_list_live_owner_name ON list (owner_id, lower(name))
  WHERE deleted_at = '' AND archived_at = '';

-- The Completed filter and the sidebar group of the same name. Its own index because
-- finished Lists can be a small fraction of the whole, and walking the name index past
-- every unfinished one to find twenty-five of them is the slow plan again.
CREATE INDEX idx_list_live_done_name ON list (lower(name))
  WHERE deleted_at = '' AND archived_at = '' AND open_count = 0 AND done_count > 0;

/*
idx_list_archived_at goes: archived_at is '' on all but a handful of Lists, so the index
is one enormous key. It cannot serve the Archived filter, which asks for the handful by
inequality, and it misleads the planner badly — sampled, it looks like the most selective
index there is, so the planner chose it, found every List a Member could reach through it,
and sorted the lot. Measured on a hundred thousand Lists: half a second with it, two
milliseconds without.
*/
DROP INDEX idx_list_archived_at;

-- What it should have been: the Lists that are archived, in the order they are read in.
CREATE INDEX idx_list_archived_name ON list (lower(name))
  WHERE deleted_at = '' AND archived_at <> '';

-- Shared with me: what somebody else made and shared with everybody, in name order.
CREATE INDEX idx_list_live_sharing_name ON list (sharing, lower(name))
  WHERE deleted_at = '' AND archived_at = '';
