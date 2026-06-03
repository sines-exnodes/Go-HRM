-- 000019_employee_experience_year_to_start_year
-- experience_year was stored as a COUNT of years (migration 000017 comment +
-- smoke fixture used 7). The BA + Python treat it as a 4-digit career-start
-- YEAR. Convert plausible counts (< 1900) to currentYear - count; leave any
-- value already year-shaped (>= 1900) untouched.
UPDATE employees
SET experience_year = EXTRACT(YEAR FROM CURRENT_DATE)::int - experience_year
WHERE experience_year IS NOT NULL
  AND experience_year < 1900;
