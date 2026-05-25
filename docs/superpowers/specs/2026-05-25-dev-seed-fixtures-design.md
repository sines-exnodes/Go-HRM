# Dev Seed Fixtures — Design

**Date**: 2026-05-25
**Author**: brainstorming session (danny.tranhoang@exnodes.vn)
**Status**: Draft → awaiting user review

## 1. Goal

Provide a one-shot dev/demo fixture loader so FE and Mobile teams can pull the
repo, run a single command, and have a database populated with realistic
Vietnamese sample data covering every phase (employees, dependents, skills,
labels, leave requests, attendance, announcements).

Out of scope: production seeding (already handled by
[`internal/services/seed_service.go`](../../internal/services/seed_service.go)),
test fixtures for `make test` (lives inside test packages), and a re-seed /
truncate flow.

## 2. Constraints

- **Single file.** The user explicitly chose a SQL-only delivery — no Go
  service, no CLI binary. The file must be runnable directly with `psql -f`.
- **Idempotent.** Running the file twice must NOT create duplicates. The
  second run prints a notice and exits cleanly.
- **Vietnamese realism.** Names like `Nguyễn Văn An`, emails under
  `@exnodes.vn`, phone numbers `0xxx xxx xxx`, salary ranges that look
  plausible for a Vietnamese tech company.
- **Pre-requisite.** The base seed (5 system roles + super admin) MUST
  already have run. The dev fixtures depend on the seeded role IDs.
  Recommended flow: `make migrate-up` → `make run` once → stop → `make seed-dev`.
- **Layering note.** The file performs raw `INSERT`s. This bypasses the
  one-directional `handler → service → repository → GORM` rule deliberately:
  fixtures are not application logic. The trade-off was discussed and
  accepted in brainstorming — schema drift will surface immediately as a
  failing `psql` run rather than a silent type mismatch.

## 3. Deliverables

| Path | Purpose |
|---|---|
| `migrations/seeds/dev_fixtures.sql` | The fixture file itself |
| `Makefile` | New target `seed-dev` |
| `README.md` | New Quickstart step 8 + new row in "Make targets" table |
| `docs/superpowers/verification/devseed.md` | Verification log committed after live run |

No Go code, no new test file. The fixture file's own idempotency guard +
manual verification log are the proof of correctness.

## 4. File structure: `migrations/seeds/dev_fixtures.sql`

The file is organised top-to-bottom in FK dependency order. Each section is
demarcated by a Vietnamese `-- == Section X — ... ==` comment.

```
-- Header
--   * Title, date, purpose
--   * Login credentials block: "Tất cả user dev dùng password: Exnodes@2026"
--   * Bcrypt hash used (so reviewer can verify it matches Exnodes@2026)

-- == Idempotency guard ==
DO $$
DECLARE n INT;
BEGIN
  SELECT COUNT(*) INTO n FROM employees WHERE is_deleted = false;
  IF n > 1 THEN
    RAISE NOTICE 'dev_fixtures: detected % employees, skipping', n;
    -- We cannot RETURN from a top-level DO block to halt the script. Instead
    -- we set a session variable and gate every subsequent INSERT on it via a
    -- WHERE clause, OR we use a single transaction + raise an exception that
    -- the wrapper catches. Simpler: wrap each INSERT in a check.
  END IF;
END $$;
```

**Resolved approach for skip behaviour** (refining the sketch above):

The DO block above is informational only. Every subsequent INSERT uses
`ON CONFLICT DO NOTHING` against a UNIQUE constraint, so the second run
will be a no-op naturally without needing an explicit early exit:

- `users.email` is CITEXT UNIQUE → duplicate insert silently skipped
- `skills.name` is UNIQUE → same
- `labels.name` is UNIQUE → same
- `departments.name` (within parent scope) → enforced UNIQUE in
  migration 000005, so re-INSERT no-ops
- `employees.user_id` is UNIQUE → re-INSERT no-ops once the user is there
- `employee_skills (employee_id, skill_id)` is UNIQUE composite → no-op
- `employee_leave_quotas (employee_id, year, leave_type)` is UNIQUE
  composite → no-op
- For `leave_requests`, `attendance`, `announcements`, `dependents`: these
  do not have natural uniqueness, so we use **deterministic UUIDs in the
  INSERT** and rely on PK conflict. Each row's UUID is constructed from a
  predictable namespace (see §6) so the second run hits PK conflict and
  no-ops.

This means the explicit DO-block guard is **redundant** in normal cases but
still useful for the operator: it logs a single line telling the user "this
is a no-op, you already seeded" instead of producing a wall of
`INSERT 0 0` lines. We keep it for that reason but it does not control flow.

## 5. Volume & shape

Reproduced from the approved Section 2 of the brainstorm (no change):

| Entity | Count | Notes |
|---|---|---|
| Departments (new) | 2 | "Marketing", "Finance" — adds to the 4 base seed already creates |
| Positions (new) | 7 | Senior SE, Junior SE, Senior Mobile, Junior Mobile, Marketing Lead, Accountant, HR Specialist (extra) |
| Skills | 20 | Go, React, PostgreSQL, TypeScript, Docker, K8s, AWS, Figma, Excel, English, Photoshop, SQL, Python, Java, Vue, NextJS, Tailwind, GraphQL, gRPC, Redis |
| Users + Employees | 19 each | Tên tiếng Việt, email `<ten>.<ho>@exnodes.vn`, all `is_active=true` |
| `user_roles` links | 19 | 1 Admin, 2 HR Manager, 3 Manager, 13 Employee |
| Dependents | 15 | Distributed across 6 employees |
| `employee_skills` | ~50 | Each employee 2-4 skills, `level` ∈ {1..5} |
| Labels | 5 | "Thông báo chung", "Tuyển dụng", "Hoạt động công ty", "Khẩn", "Đào tạo" |
| Announcements | 10 | Mix of labels, posted across last 60 days |
| `employee_leave_quotas` | 19 × 3 = 57 | One row per employee per leave_type (annual=12, sick=5, personal=3) for year 2026 |
| `leave_requests` | 25 | Status mix: 10 approved, 8 pending, 5 rejected, 2 cancelled; dates Jan-May 2026 |
| `attendance` | ~200 | 5 employees × ~40 weekdays, a few late entries |

**Total rows inserted on first run**: ~410.

## 6. Deterministic UUID scheme

To make cross-section FK references readable and to enable PK-based
idempotency on tables without natural uniqueness, the file uses
hand-assigned UUIDs with a hex-prefix namespace:

| Prefix | Entity |
|---|---|
| `11111111-...` | departments |
| `22222222-...` | positions |
| `33333333-...` | skills |
| `44444444-...` | labels |
| `55555555-...` | users |
| `66666666-...` | employees |
| `77777777-...` | dependents |
| `88888888-...` | employee_skills |
| `99999999-...` | leave_quotas |
| `aaaaaaaa-...` | leave_requests |
| `bbbbbbbb-...` | attendance |
| `cccccccc-...` | announcements |

Within a prefix, the last 12 hex digits encode an index (e.g.
`66666666-0000-0000-0000-000000000007` is employee #7). This makes the SQL
file self-documenting — a reviewer reading a `manager_id` FK can
immediately tell which employee it points at.

## 7. Password handling

All 19 dev users share one bcrypt hash (cost 12) of `Exnodes@2026`. The
hash is generated once locally and pasted into the file as a top-of-file
comment plus the literal value used in the `users.password_hash` column.

A small reproducibility note at the top of the SQL file:
```
-- Login credentials for all dev users:
--   Email:    <firstname>.<lastname>@exnodes.vn  (e.g. nguyen.an@exnodes.vn)
--   Password: Exnodes@2026
-- Bcrypt hash below was generated with:
--   go run -e 'fmt.Println(string(bcrypt.GenerateFromPassword(...)))'
```

## 8. Safety

This is dev fixtures, not a destructive migration. Risk surface:
- Running against the wrong DB (e.g. someone's `DATABASE_URL` points at
  staging). Mitigation: the idempotency guard logs `"dev_fixtures:
  detected N employees, skipping"` so an accidental run against a
  populated environment is a no-op.
- We do NOT add an `APP_ENV=production` guard at the SQL layer (it would
  require a `DO $$ ... IF current_setting() ... $$` dance that is fragile).
  The Makefile target carries the burden:
  ```makefile
  seed-dev:
    @[ "$$APP_ENV" != "production" ] || { echo "Refusing to run in production"; exit 1; }
    psql "$$DATABASE_URL" -f migrations/seeds/dev_fixtures.sql
  ```

## 9. README changes

### New Quickstart step (after step 7 "Smoke-test")

```markdown
### 8. (Optional) Seed dev fixtures

If you want a database pre-populated with realistic sample data (19 sample
employees, departments, leave requests, attendance, announcements) for
FE/Mobile development:

\`\`\`bash
make seed-dev
\`\`\`

The script is idempotent — running it twice is a no-op. All sample users
share the password `Exnodes@2026`; emails follow `<firstname>.<lastname>@exnodes.vn`
(e.g. `nguyen.an@exnodes.vn`). The script refuses to run when `APP_ENV=production`.
```

### New row in "Make targets" table

```markdown
| `make seed-dev` | Load development sample data (idempotent; refuses in production) |
```

## 10. Verification plan

Committed log at `docs/superpowers/verification/devseed.md` after a live run:

1. `make migrate-up` on a fresh DB
2. `make run` once → ensure base seed (roles + super admin) created
3. Stop server, run `make seed-dev`
4. `psql` spot-checks:
   - `SELECT COUNT(*) FROM employees WHERE is_deleted=false;` → 20
   - `SELECT COUNT(*) FROM leave_requests;` → 25
   - `SELECT status, COUNT(*) FROM leave_requests GROUP BY status;` → mix
   - `SELECT COUNT(*) FROM attendance;` → ~200
5. `curl POST /api/v1/auth/login` with `nguyen.an@exnodes.vn / Exnodes@2026`
   → returns access token
6. `curl GET /api/v1/employees -H "Authorization: Bearer ..."` (Admin token)
   → returns 20 employees
7. Run `make seed-dev` again → `psql` output shows the NOTICE line and no
   duplicate rows
8. `APP_ENV=production make seed-dev` → exits 1

Verification log committed once all 8 checks pass.

## 11. Open items / non-goals

- **No re-seed flow** in this phase. If anyone wants to refresh dev data
  they manually `TRUNCATE ... CASCADE` the relevant tables first. Adding a
  `make seed-dev-reset` target is a future enhancement and explicitly out
  of scope here (avoids a tempting `--force` foot-gun).
- **No randomisation.** All data is hand-curated and deterministic. No
  faker library, no `gen_random_uuid()` in INSERT values (we use literals
  per §6). This matters because deterministic IDs are required for
  PK-based idempotency.
- **No FE/Mobile mock images.** Avatars, ID cards, attachments use
  placeholder URLs (`https://example.com/...`) — actual image hosting is
  out of scope.
