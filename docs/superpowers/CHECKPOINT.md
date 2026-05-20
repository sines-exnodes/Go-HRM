# Resume Checkpoint

**Last updated:** 2026-05-20
**Stopped at:** Phase 0–4 done, live-verified, and committed on `main`. Phase 4 (Skills + Labels) closed.
**Branch:** `main`
**HEAD:** ~115 commits (run `git log --oneline -1` for exact SHA)
**DB migration version:** **7** (main + test DB both at 7)

## How to resume next session

Tell Claude: *"Resume the Go migration — start Phase 5 per docs/superpowers/CHECKPOINT.md"*.

Plan: `docs/superpowers/plans/2026-05-15-phase-05-leave-requests.md`.

⚠️ Phase 5 plan was written pre-schema-split — re-audit before executing (leave_requests likely keys on `employee_id`, not `user_id`; check existing perm constants; next free migration number is **000008**; verify `employee_leave_quotas` Phase 2 table is compatible).

## Current state

### Phase 0 — DONE ✅
Phase 0 verification log: [docs/superpowers/verification/phase-00.md](verification/phase-00.md)

### Phase 1 — DONE ✅
Auth + RBAC. Phase 1 verification log: [docs/superpowers/verification/phase-01.md](verification/phase-01.md)

### Phase 2 — DONE ✅
Users + Employees + Dependents. Phase 2 verification log: [docs/superpowers/verification/phase-02.md](verification/phase-02.md). Code review applied (see review-fixes.md).

### Phase 3 — DONE ✅
Departments + Positions. Phase 3 verification log: [docs/superpowers/verification/phase-03.md](verification/phase-03.md).

### Phase 4 — DONE ✅ (2026-05-20)

Skills catalog (`skills`) + employee↔skill assignment (`employee_skills`) + announcement labels (`labels`). Live verification: [docs/superpowers/verification/phase-04.md](verification/phase-04.md) (22 steps, all green). Highlights:

- Migrations **000006** (skills + employee_skills join) and **000007** (labels) — round-trip verified (7 ↔ 5).
- Skill icon upload reuses Phase-2 `UploadService` + the review-fix #2 content-sniff (`http.DetectContentType`). Spoofed `Content-Type: image/png` on a PHP body returns 400 with NO upload attempted.
- Skill delete guard returns **409 + structured details** (`{employee_count, skill_id, skill_name}`) while any live `employee_skills` row references it. Mirror of the Phase 3 department/position guard.
- Employee↔skill assignment uses **PUT-replace** semantics; empty `skill_ids:[]` unassigns everything. Soft-deleted join rows are reactivated rather than re-inserted (avoids tripping `uq_employee_skills_pair_live`).
- Labels expose **only** `GET /api/v1/announcement-labels` and `POST /api/v1/announcement-labels` (get-or-create, case-insensitive). No update/delete by design — out of scope per Python source.
- **Seed gap closed:** `PermAnnounceManage` is now granted to **Admin** AND **HR Manager** in `seed_service.go`. Confirmed live via boot logs (`seed: merged permissions into role "Admin" / "HR Manager"`).

#### Phase 4 commits (in order)

| Commit | Task | Summary |
|---|---|---|
| `5b39672` | 1 | Migration 000006 — skills + employee_skills join |
| `1b63f4e` | 2 | Migration 000007 — announcement labels |
| `17c7459` | 3 | Models: Skill, EmployeeSkill, Label |
| `5961f6f` | 4 | DTOs: skill (dual form/json tags), label |
| `a67e7e1` | 5-7 | Repositories: skill, employee_skill (with reactivation), label |
| `67699dc` | 8+10 | Skill service + tests (incl content-spoof + delete-guard) |
| `4b09047` | 9+11 | Label service + tests (idempotent get-or-create) |
| `738a2c2` | 12+13 | Skill handler (multipart icon) + Label handler |
| `0387742` | 14 | Wire routes in main.go + seed PermAnnounceManage gap fix |
| `e318ab2` | 15+16 | Verification log + regenerated Swagger |

## TOOLING NOTE (2026-05-20)

Subagent dispatch (`Agent` with subagent_type) was **unavailable** in this session — every probed type (`team-lead`, `general-purpose`, prefix variants) returned "not found" with an empty available-agents list, despite the `ennam-dev-agent-team` plugin being enabled in `.claude/settings.json`. The plugin files exist on disk at `~/.claude/plugins/marketplaces/ennam-internal-plugins/stacks/ennam-dev-agent-team/agents/` but were not registered with the Agent tool.

Phase 4 was executed **inline by project-owner** under explicit user override of the "never write production code" rule (the user said `C` to that override and `tui muốn bạn spawn sao implement hết modules phase luôn` to authorise non-interactive completion). All 16 tasks landed atomically: 11 commits, all `make fmt && make vet && go test ./...` clean.

**Before Phase 5:** prefer restarting the session to restore subagent capability. With subagents back the recommended pattern is `superpowers:subagent-driven-development` — one subagent per task or per logical batch, with the project-owner only orchestrating.

## Code review status

Phase 0–3: review applied, fixes committed (`docs/superpowers/verification/review-fixes.md`). Two top security fixes live-re-verified.

Phase 4: **review not yet requested.** Recommendation — schedule a focused review of the skill multipart path (handler + service.uploadIcon) before Phase 5 starts, since that pattern will be reused for future upload endpoints (avatar already exists; org-settings logo will follow in Phase 8). The content-sniff is now tested twice (Phase 2 avatar + Phase 4 icon) but the duplication between the two readers (`readAvatar` and `readSkillIcon`) is a candidate for a small shared helper.

## Local environment notes

- Postgres: Docker container `ennam-ecom-postgres` at `localhost:5432`, user `ennam` / pass `ennam_dev_2026`, main DB `exnodes_hrm`, test DB `exnodes_hrm_test`. Both at migration version **7**.
- `.env` is git-ignored; the Phase 4 verification used `STORAGE_*` keys pointing at an ephemeral MinIO container (`phase04-minio`, `localhost:19000`, bucket `hrm-uploads`, user `minioadmin`/`minioadmin123`). The container was destroyed after verification — to re-run icon-upload tests, recreate it with the recipe in `docs/superpowers/verification/phase-04.md` §3.
- Storage is optional at server boot (`NewUploadService` does not ping S3); icon-upload endpoints return 500 when storage is unconfigured but skill CRUD without icon continues to work.
- Go toolchain: 1.25 per `go.mod`.

## Key design decisions (do NOT redo)

- **Schema split:** `users` (auth) ⟂ `employees` (HR) ⟂ `dependents` ⟂ `employee_skills`. Skill assignments live on `employee_id`, NOT `user_id` (Phase 4 REVISION NOTES item #3). Source of truth: [`migrations/000006_create_skills.up.sql`](../../migrations/000006_create_skills.up.sql).
- **Migrations:** versioned SQL only via `golang-migrate`. NEVER `AutoMigrate()`. See `[[feedback-migrations]]`.
- **Audit cols:** every entity has `created_at + updated_at + is_deleted + deleted_at`. See `[[feedback-audit-fields]]`.
- **DoD per phase:** must include real end-to-end verification (run server, curl flows, DB spot-check), commit verification log to `docs/superpowers/verification/phase-NN.md`. See `[[feedback-self-verify-each-phase]]`.
- **Skill icon upload reuses the Phase 2 avatar pattern:** `http.DetectContentType` sniff with allowlist; never trust the client-supplied `Content-Type`. The duplication in handler-side `readAvatar`/`readSkillIcon` is intentional (different size limits + ext sets) but could be extracted later.
- **Label scope is intentionally minimal:** list + get-or-create only, mirroring the Python source. Do NOT add update/delete in a future phase without explicit re-scoping (REVISION NOTES item #4).
- **PermAnnounceManage seeding:** Admin AND HR Manager carry it directly (not just via Super Admin's `*` wildcard). This is the load-bearing fix from REVISION NOTES item #5 — do NOT remove it without re-auditing the labels-API access path.

## Outstanding micro-items

- Untracked (intentional, IDE/project rules): `.claude/`, `AGENTS.md`, `CLAUDE.md`.
- Phase 4 plan file (`docs/superpowers/plans/2026-05-15-phase-04-skills-labels.md`) has unticked `- [ ]` checkboxes in the draft task bodies. They were superseded by REVISION NOTES; not worth a churn commit to tick them.
- Phase 2 carryover `EmployeeService.toRead` department/position nil-projection gap is still open — does NOT block Phase 5 but should be addressed when FE needs embedded objects on `GET /employees/me`.
