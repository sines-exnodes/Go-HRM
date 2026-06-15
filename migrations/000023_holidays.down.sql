-- migrations/000023_holidays.down.sql
DROP TRIGGER IF EXISTS trg_holidays_set_updated_at ON holidays;
DROP TABLE IF EXISTS holidays;

DROP TRIGGER IF EXISTS trg_holiday_templates_set_updated_at ON holiday_templates;
DROP TABLE IF EXISTS holiday_templates;
