-- Down migration for 000016_add_announcement_target_users.
--
-- Refuses to proceed if any announcement currently has target_audience='custom',
-- since the constraint revert would orphan that audience mapping. Operator
-- must reassign such rows to 'all' or 'department' first.

DO $$
DECLARE
    n INT;
BEGIN
    SELECT COUNT(*) INTO n FROM announcements
    WHERE target_audience = 'custom' AND is_deleted = FALSE;
    IF n > 0 THEN
        RAISE EXCEPTION
            'Refusing to revert: % announcement(s) still use target_audience=custom. Reassign before downgrading.', n;
    END IF;
END$$;

ALTER TABLE announcements DROP CONSTRAINT announcements_target_audience_check;
ALTER TABLE announcements
    ADD CONSTRAINT announcements_target_audience_check
    CHECK (target_audience IN ('all','department'));

DROP TRIGGER IF EXISTS trg_announcement_target_users_set_updated_at ON announcement_target_users;
DROP TABLE IF EXISTS announcement_target_users;
