-- =========================================================================
-- 000021_roles_name_partial_unique
-- Replace the plain UNIQUE constraint on roles.name with a partial unique
-- index that only covers non-deleted rows. This allows the same name to be
-- reused after a soft-delete, which the soft-delete semantics require.
-- =========================================================================

-- Remove the plain UNIQUE constraint (it was defined inline in 000002).
ALTER TABLE roles DROP CONSTRAINT roles_name_key;

-- Partial unique index: uniqueness enforced only among live (non-deleted) rows.
CREATE UNIQUE INDEX idx_roles_name_active
    ON roles (name)
    WHERE is_deleted = FALSE;
