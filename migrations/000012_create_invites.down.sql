-- Reverse of 000012_create_invites.
DROP TRIGGER IF EXISTS trg_invites_set_updated_at ON invites;
DROP INDEX  IF EXISTS uq_invites_pending_email;
DROP INDEX  IF EXISTS uq_invites_token_live;
DROP INDEX  IF EXISTS idx_invites_invited_by;
DROP INDEX  IF EXISTS idx_invites_expires_at;
DROP INDEX  IF EXISTS idx_invites_is_deleted;
DROP INDEX  IF EXISTS idx_invites_email;
DROP TABLE  IF EXISTS invites;
