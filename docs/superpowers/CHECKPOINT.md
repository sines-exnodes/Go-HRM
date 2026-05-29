# Resume Checkpoint — MIGRATION COMPLETE 🎉 · post-migration API parity in flight

**Last updated:** 2026-05-29
**Stopped at:** Phases 0–9 complete on `main`. Post-migration **Python ↔ Go API parity** work ongoing: announcements (merged), employees (two PRs pushed, awaiting open).
**Branch:** `main` — parity work on `feat/employees-parity` → `feat/employees-salary-banking`.
**DB migration version:** **16** on `main`; **17** on `feat/employees-parity` (000017_employee_parity).
**See:** [Post-migration parity work](#post-migration-parity-work-python--go-api-parity) below for current state.

## How to resume next session

The implementation pipeline is closed. Suggested next priorities (in
descending value):

1. **Bundled code review** — Phases 4-9 have not been formally reviewed. See "Code review status" below for the recommended bundle.
2. **Phase 6 follow-up** — switch attendance thresholds from env vars to `system_config` lookup (Phase 8 produced the row; Phase 6 still reads env). One small commit.
3. **Phase 7 attachment-upload HTTP handler** — model + repo are in place; the route is the one missing piece.
4. **Manager-role completeness** — Phase 5 manager lacks `Update/Delete` on leave_requests (flagged for next BA pass).
5. **Production env wiring** — set `FIREBASE_CREDENTIALS_PATH` + real SMTP host for invite emails.
6. **Phase 9 password-reset flow** — reuse `EmailService` + add a `password-reset.html` template if BA confirms scope.

If a Phase 10 is added later: REVISION NOTES pattern remains the way to draft it. Latest taken migration = **000017**; next is **000018**.

### Resume entry points

1. **`docs/superpowers/CHECKPOINT.md`** (this file).
2. **`.serena/memories/project_overview.md`** — code-map + boot protocol.
3. **`CLAUDE.md`** — auto-loaded into every Claude session.

## Post-migration parity work (Python ↔ Go API parity)

The migration is done; ongoing work reconciles Go's API shape with the Python
source, audited module-by-module: audit → locked decisions → PR(s) →
verification log → FE doc → handoff for deferred items.

### Announcements — DONE (merged)
PR #5 (`body`→`description`, `send_now`, brief mobile widget) + PR #6 (hybrid
per-user targeting `target_audience:"custom"` + `recipient_ids[]`, CORS fix).
Migrations 000013–000016, both on `main`. 13-decision audit.

### Employees — PUSHED (PRs not yet opened)
19-decision audit (Python `users` module ↔ Go `users`⟂`employees`⟂`dependents`).
Two stacked branches off `main`@`bcb4c0b`:

- **`feat/employees-parity`** (PR A, `a377e44`) — emergency-contact list,
  leave-quota/skills/cv on read, widened self-edit (name/gender/dob), self &
  destructive guards, admin change-email, marital/education enum fixes.
  **Migration 000017_employee_parity.**
- **`feat/employees-salary-banking`** (PR B, stacked on A) — salary/banking
  field-level perms (`users:salary_view/manage`, `users:banking_view/manage`)
  + account-number masking + write-gate (decision #6). No migration.

Verified: [`verification/employees-parity-pr-a.md`](verification/employees-parity-pr-a.md)
+ [`verification/employees-parity-pr-b.md`](verification/employees-parity-pr-b.md)
(build/vet, full integration suite, migration up/down round-trip, live HTTP e2e).

**3 decisions deferred** → [`handoff-2026-05-29-employee-parity.md`](handoff-2026-05-29-employee-parity.md):
#10 line-manager suite (picker + direct-reports + cycle validation), #11 cv/
id-card upload endpoints, #15 role-level assignment authority (N/A — Go RBAC
has no role levels).

FE docs (on disk, uncommitted in the web repo `api_info_go/`): `employee.md`
(new) + `me.md` (refreshed for the new `/employees/me` shape).

**Resume order:** open PR A → `main`, then rebase/merge PR B (it's stacked);
commit the FE docs on the web side; then pick up the deferred items from the
handoff. PR B currently contains A's commits via a merge of A into B.

## Phase Summary (final)

| # | Module | Migration | Verification | Commits |
|---|---|---|---|---|
| 0 | Foundation | 000001 | [phase-00.md](verification/phase-00.md) | — |
| 1 | Auth + RBAC | 000002 | [phase-01.md](verification/phase-01.md) | — |
| 2 | Users + Employees + Dependents | 000003 + 000004 | [phase-02.md](verification/phase-02.md) | — |
| 3 | Departments + Positions | 000005 | [phase-03.md](verification/phase-03.md) | — |
| 4 | Skills + Labels | 000006 + 000007 | [phase-04.md](verification/phase-04.md) | — |
| 5 | Leave Requests + Quota | 000008 | [phase-05.md](verification/phase-05.md) | ~10 |
| 6 | Attendance + Matrix | 000009 | [phase-06.md](verification/phase-06.md) | 14 |
| 7 | Announcements + Mobile + SSE | 000010 | [phase-07.md](verification/phase-07.md) | 15 |
| 8 | Organization Settings | 000011 | [phase-08.md](verification/phase-08.md) | 8 |
| 9 | Email + Invite + Push | 000012 | [phase-09.md](verification/phase-09.md) | 12 |

### Phase 9 — DONE ✅ (2026-05-22)

Email + Invite + Push notifications. 7 new endpoints (5 admin invite +
1 public accept + 1 push test), invite table (migration **000012**),
embedded HTML/plain email templates, SMTP via gomail, FCM HTTP v1 push
client with no-op fallback. Live verification:
[docs/superpowers/verification/phase-09.md](verification/phase-09.md)
(23 e2e steps + DB spot-check, all green, real email delivery
verified via Mailpit). Highlights:

- **Migration 000012** — `invites` table with `invited_by →
  employees(id) RESTRICT` (schema split) + `accepted_user_id → users(id)
  SET NULL` (auth-level marker populated on accept). Two partial unique
  indexes: `(token) WHERE is_deleted=FALSE` and `(email) WHERE
  accepted_at IS NULL AND is_deleted=FALSE`. `role_ids UUID[]` column
  for role-on-accept assignment (custom `UUIDArray` scanner/valuer).
- **Diverges from Python source** intentionally: invites are first-class
  rows; user creation is deferred to `/invites/accept`. Avoids partially-
  provisioned users sitting in the `users` table.
- **`POST /invites/accept` is genuinely public** — registered outside
  the JWT-protected group. Token in body is the credential. Atomically
  creates user + employee via `empSvc.Create` + assigns roles via
  `users.ReplaceRoles` + stamps `accepted_at` + `accepted_user_id`.
  Verified live (user logs in immediately after accept).
- **SMTP graceful degradation** — when `SMTP_HOST=""`, the EmailService
  returns `ErrEmailDisabled` and the InviteService records the message
  on `invites.last_email_error`. Invite creation never returns 500 on
  SMTP misconfig (REVISION NOTES #11). Verified via service test.
  Mailpit pipeline proves the real-delivery path end-to-end.
- **FCM HTTP v1 push client** with JWT-based service-account auth
  (`google.JWTConfigFromJSON`). When `FIREBASE_CREDENTIALS_PATH=""` or
  the file is unreadable, returns a no-op logger client that satisfies
  the interface but reports `IsConfigured()=false`. Boot never fails
  on FCM misconfig. `/notifications/test` returns the diagnostic
  `{sent, skipped, errors}` envelope.
- **`PermInviteManage` constant added** to registry + seeded to Admin
  and HR Manager via merge-seed (the seed service appends missing
  perms to existing system roles on every boot — idempotent).
- **Token format**: 32 random bytes → URL-safe base64 (no padding) →
  43 chars. Verified at the Mailpit boundary.
- **Composite signal — accept flow validation**: replay (409),
  unknown token (404), short password (400), revoked token (404),
  duplicate-pending (409), existing-user-email (409). All verified
  live in steps 12-17.

#### Phase 9 commits (in order)

| Commit | Task | Summary |
|---|---|---|
| `4a388f6` | – | Plan REVISION NOTES at the top |
| `2b520f8` | T1+T2 | Dependencies (gomail/oauth2) + env keys |
| `3bf9f51` | T3 | Migration 000012 — invites table |
| `8e70793` | T4+T5 | Invite model + UUIDArray scanner + DTOs |
| `0153a53` | T6 | InviteRepository |
| `7923375` | T7+T8 | EmailService + embedded HTML/text templates |
| `7d9dc35` | T9+T10 | InviteService (CRUD + Accept) |
| `0ede002` | T11+T12 | PushClient interface + FCM impl + PushNotificationService |
| `ae0a871` | – | PermInviteManage + seed |
| `2214f14` | T13+T14 | InviteHandler + NotificationHandler |
| `f470410` | T15 | Wire in main.go (public /accept + 5 admin /invites + /notifications/test) |
| `6ccd8e9` | T16+T17 | 17 service tests + truncateAll extension + Swagger regen |
| `14e872e` | T19-T22 | E2E verification log (Mailpit pipeline + DB spot-check) |
| _this_ | T18 | README + CHECKPOINT close (migration complete) |

## TOOLING NOTE (unchanged)

Subagent dispatch (`Agent` with `subagent_type`) is **structurally unavailable** in the VSCode-extension SDK runtime. Inline by project-owner (commit-per-task) carried all 10 phases.

## Code review status

Phases 0–3: review applied, fixes committed.

Phases 4–9: **review not yet requested.** Recommendation — one final bundled review covering:

- **Multipart upload pattern** (avatar P2, skill icon P4, leave attachment P5; will add announcement attachments + logo at follow-up). The `http.DetectContentType` + MIME allowlist sniff is now at four sites — extract into a shared helper.
- **Two-layer access control** (`RequirePerms` + `asAdmin bool` ownership branch — Phases 5/6/7/8/9). Pattern works; document it in `.serena/memories`.
- **Composite-PK reactivation pattern** (`AnnouncementLabel` + `AnnouncementTargetDepartment` + `EmployeeSkill`). Three sites share this; extract a small generic helper.
- **Singleton repo pattern** (`SystemConfigRepository`) — 4-method interface, DB-level CHECK as the last line of defense.
- **SSE hub design** (single-process, drop-on-full-buffer) — fine for a single replica; document the scaling boundary clearly.
- **Public endpoints surface** — `/auth/login`, `/auth/refresh`, `/invites/accept`. Worth a security pass before production exposure (rate-limiting in particular).

## Local environment notes

- **Postgres**: Docker container `ennam-ecom-postgres` at `localhost:5432`, user `ennam` / pass `ennam_dev_2026`, main DB `exnodes_hrm`, test DB `exnodes_hrm_test`. Both at migration version **12**.
- **Port 8080 conflict**: `ennam-kg-server` container holds host port 8080 in this dev environment. Phases 7-9 live verifications ran on `PORT=8082`. CI default stays 8080.
- **Mailpit for SMTP verification**: `docker run -d --rm -p 11025:1025 -p 18025:8025 --name mailpit-phase09 axllent/mailpit`. UI at `http://localhost:18025`.
- **FCM disabled in dev** — `FIREBASE_CREDENTIALS_PATH=""`. PushClient is the no-op logger. Production rollout: set the env var to a service-account JSON + `FIREBASE_PROJECT_ID`.
- `.env` is git-ignored. `.env.example` has every key the project reads.
- Go toolchain: 1.25 per `go.mod`.

## Key design decisions (do NOT redo)

- **Schema split:** every cross-aggregate FK from Phase 2 onward targets `employees(id)`, NOT `users(id)`. Exceptions: `users.id` is the FK target for auth-level surfaces (`device_tokens`, `user_notification_settings`, `announcement_views`, `invites.accepted_user_id`).
- **Migrations:** versioned SQL only via `golang-migrate`. NEVER `AutoMigrate()`.
- **Audit cols:** every entity has `created_at + updated_at + is_deleted + deleted_at` + `BEFORE UPDATE` trigger. Singletons keep the columns for schema parity even when soft-delete is meaningless.
- **Singleton tables:** PK is a fixed sentinel UUID + `CHECK (id = '…')` constraint at the DB level. Repo exposes only `Get / EnsureExists / UpdateFields(map)`.
- **Composite-PK join models:** declare an explicit Go model; replace-set logic uses snapshot-diff-reactivate.
- **Repo joins MUST qualify `is_deleted`** — `models.NotDeleted` becomes ambiguous after a JOIN to a table that carries `is_deleted`.
- **SSE broadcast is a "refresh hint"** — FE refetches via GET on receipt; visibility is enforced on the read path.
- **Idempotent markers** — `Clauses(clause.OnConflict{DoNothing:true})` preserves first-occurrence semantics.
- **Partial PATCH semantics:** pointer-typed DTO fields → only write when non-nil. Empty PATCH = no-op success. Stamp `updated_by/at` only when "real" content fields change.
- **Public endpoints with token-as-credential** (Phase 9 `/invites/accept`) live OUTSIDE the JWT group. Token validated in the service, not the middleware.
- **External integrations degrade gracefully** — empty SMTP_HOST / FIREBASE_CREDENTIALS_PATH disable the respective integration without crashing the boot. Record errors on the relevant row (e.g. `invites.last_email_error`) rather than rolling back the request.

## Outstanding micro-items

- Untracked (intentional): `.claude/`, `AGENTS.md`, `CLAUDE.md`.
- Phase 5/6/7/8/9 plan files have unticked `- [ ]` checkboxes in draft task bodies — superseded by REVISION NOTES blocks; not worth churn commits.
- Phase 5 manager-role completeness gap (lacks `Update/Delete` on leave_requests).
- Phase 6 attendance service still reads thresholds from env vars — should switch to `system_config` lookup now that the row exists.
- Phase 7 attachment-upload HTTP handler is deferred.
- Phase 7 `target_audience='custom'` deferred (no backing table).
- Phase 7 scheduled-publish cron deferred.
- Phase 8 logo upload deferred.
- Phase 9 password-reset email flow deferred (needs BA confirmation).
- Phase 9 `/invites/accept` rate-limiting (consider IP-based at the reverse proxy).
- Phase 9 FCM topic-based fanout deferred (per-device only at present).
