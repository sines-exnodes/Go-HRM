-- Reverse of 000011_create_system_config.
DROP TRIGGER IF EXISTS trg_system_config_set_updated_at ON system_config;
DROP INDEX  IF EXISTS idx_system_config_is_deleted;
DROP TABLE  IF EXISTS system_config;
