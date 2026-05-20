-- 000009_create_attendance.down.sql
DROP TRIGGER IF EXISTS trg_attendance_sessions_set_updated_at ON attendance_sessions;
DROP TRIGGER IF EXISTS trg_attendance_set_updated_at ON attendance;
DROP TABLE IF EXISTS attendance_sessions;
DROP TABLE IF EXISTS attendance;
