-- 000010_create_announcements.down.sql
DROP TRIGGER IF EXISTS trg_announcement_views_set_updated_at ON announcement_views;
DROP TRIGGER IF EXISTS trg_announcement_attachments_set_updated_at ON announcement_attachments;
DROP TRIGGER IF EXISTS trg_announcement_target_departments_set_updated_at ON announcement_target_departments;
DROP TRIGGER IF EXISTS trg_announcement_labels_set_updated_at ON announcement_labels;
DROP TRIGGER IF EXISTS trg_announcements_set_updated_at ON announcements;
DROP TABLE IF EXISTS announcement_views;
DROP TABLE IF EXISTS announcement_attachments;
DROP TABLE IF EXISTS announcement_target_departments;
DROP TABLE IF EXISTS announcement_labels;
DROP TABLE IF EXISTS announcements;
