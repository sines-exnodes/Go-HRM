-- migrations/000023_holidays.up.sql

CREATE TABLE holidays (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    year        INTEGER     NOT NULL,
    name        TEXT        NOT NULL,
    from_date   DATE        NOT NULL,
    to_date     DATE        NOT NULL,
    is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Unique name per year among non-deleted rows (AC-03)
CREATE UNIQUE INDEX uq_holidays_name_year
    ON holidays (year, LOWER(name))
    WHERE is_deleted = FALSE;

CREATE INDEX idx_holidays_year
    ON holidays (year, from_date)
    WHERE is_deleted = FALSE;

CREATE TRIGGER trg_holidays_set_updated_at
    BEFORE UPDATE ON holidays
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE holiday_templates (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    year        INTEGER     NOT NULL,
    name        TEXT        NOT NULL,
    from_date   DATE        NOT NULL,
    to_date     DATE        NOT NULL,
    is_deleted  BOOLEAN     NOT NULL DEFAULT FALSE,
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_holiday_templates_year ON holiday_templates (year);

CREATE TRIGGER trg_holiday_templates_set_updated_at
    BEFORE UPDATE ON holiday_templates
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Vietnamese public holiday presets (approximate statutory dates; users adjust via the UI)
INSERT INTO holiday_templates (year, name, from_date, to_date) VALUES
    -- 2025 (Ất Tỵ — Snake Year, Lunar New Year Jan 29)
    (2025, 'Tết Dương Lịch',               '2025-01-01', '2025-01-01'),
    (2025, 'Tết Nguyên Đán',               '2025-01-27', '2025-02-02'),
    (2025, 'Ngày Giỗ Tổ Hùng Vương',       '2025-04-07', '2025-04-07'),
    (2025, 'Ngày Giải Phóng & Lao Động',   '2025-04-30', '2025-05-01'),
    (2025, 'Ngày Quốc Khánh',              '2025-09-01', '2025-09-02'),
    -- 2026 (Bính Ngọ — Horse Year, Lunar New Year Feb 17)
    (2026, 'Tết Dương Lịch',               '2026-01-01', '2026-01-01'),
    (2026, 'Tết Nguyên Đán',               '2026-02-15', '2026-02-21'),
    (2026, 'Ngày Giỗ Tổ Hùng Vương',       '2026-04-26', '2026-04-26'),
    (2026, 'Ngày Giải Phóng & Lao Động',   '2026-04-30', '2026-05-01'),
    (2026, 'Ngày Quốc Khánh',              '2026-09-01', '2026-09-02'),
    -- 2027 (Đinh Mùi — Goat Year, Lunar New Year Feb 6)
    (2027, 'Tết Dương Lịch',               '2027-01-01', '2027-01-01'),
    (2027, 'Tết Nguyên Đán',               '2027-02-04', '2027-02-10'),
    (2027, 'Ngày Giỗ Tổ Hùng Vương',       '2027-04-15', '2027-04-15'),
    (2027, 'Ngày Giải Phóng & Lao Động',   '2027-04-30', '2027-05-01'),
    (2027, 'Ngày Quốc Khánh',              '2027-09-01', '2027-09-02');
