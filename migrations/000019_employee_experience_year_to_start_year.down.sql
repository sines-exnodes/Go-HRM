-- Down for 000019. The reverse (year -> count) is arithmetically approximate:
-- it assumes the apply-date year equals the up-migration's year. Documented
-- per the data-loss-guard convention.
DO $$
BEGIN
    RAISE NOTICE 'Reverting experience_year (year -> count) is approximate: it uses the current year, which may differ from the up-migration apply year.';
END$$;

UPDATE employees
SET experience_year = EXTRACT(YEAR FROM CURRENT_DATE)::int - experience_year
WHERE experience_year IS NOT NULL
  AND experience_year >= 1900;
