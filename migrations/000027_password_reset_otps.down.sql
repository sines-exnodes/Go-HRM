-- migrations/000027_password_reset_otps.down.sql

DROP TRIGGER IF EXISTS trg_password_reset_otps_set_updated_at ON password_reset_otps;
DROP TABLE IF EXISTS password_reset_otps;
