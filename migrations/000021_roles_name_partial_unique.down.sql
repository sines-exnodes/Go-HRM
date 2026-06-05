-- Restore the plain UNIQUE constraint.
DROP INDEX IF EXISTS uq_roles_name_active;
ALTER TABLE roles ADD CONSTRAINT roles_name_key UNIQUE (name);
