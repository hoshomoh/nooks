-- All lists can be sorted by when a List was made, newest or oldest first.
--
-- An index already in that order is walked and stopped at a page, like the ones beside
-- it; without one, every live List is read and sorted to find the first twenty-five.
-- One index serves both directions.
CREATE INDEX idx_list_live_created ON list (created_at DESC)
  WHERE deleted_at = '' AND archived_at = '';
