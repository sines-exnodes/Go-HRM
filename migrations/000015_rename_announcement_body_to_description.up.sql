-- =========================================================================
-- 000015_rename_announcement_body_to_description
-- Rename announcements.body → announcements.description to match the Python
-- repo's wire shape (parity audit decision #2 on the announcements API).
--
-- BREAKING CHANGE for API clients: the JSON field `body` is renamed to
-- `description` on Create / Update / Read. Existing rows keep their data;
-- only the column name changes. The down migration reverses cleanly.
--
-- Note: the `summary` column is intentionally left in place — the team
-- decided to keep it as a Go-only short blurb separate from the full
-- description (parity audit decision #9a).
-- =========================================================================

ALTER TABLE announcements RENAME COLUMN body TO description;
