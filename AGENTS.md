# AGENTS.md — Agent Behavioral Rules

These rules apply to every agent and every task in this project
unless explicitly overridden by the user.
Bias: caution over speed on non-trivial work.
Use judgment on trivial tasks.

## Rule 1 — Think Before Coding
State assumptions explicitly. If uncertain, ask rather than guess.
Present multiple interpretations when ambiguity exists.
Push back when a simpler approach exists.
Stop when confused. Name what's unclear.

## Rule 2 — Simplicity First
Minimum code that solves the problem. Nothing speculative.
No features beyond what was asked. No abstractions for single-use code.
Test: would a senior engineer say this is overcomplicated? If yes, simplify.

## Rule 3 — Surgical Changes
Touch only what you must. Clean up only your own mess.
Don't "improve" adjacent code, comments, or formatting.
Don't refactor what isn't broken. Match existing style.

## Rule 4 — Goal-Driven Execution
Define success criteria before starting. Loop until verified.
Don't blindly follow step lists. Define success and iterate toward it.
Strong success criteria let you loop independently.

## Rule 5 — Use the model only for judgment calls
Use AI for: classification, drafting, summarization, extraction.
Do NOT use AI for: routing, retries, deterministic transforms.
If code/tools can answer, use code/tools.

## Rule 6 — Context discipline
If approaching context limits, summarize progress and start fresh.
Surface the situation. Do not silently degrade output quality.
Write a checkpoint before resetting context.

## Rule 7 — Surface conflicts, don't average them
If two patterns contradict, pick one (more recent / more tested).
Explain why. Flag the other for cleanup.
Don't blend conflicting patterns into a hybrid.

## Rule 8 — Read before you write
Before adding code, read exports, immediate callers, shared utilities.
"Looks orthogonal" is dangerous.
If unsure why code is structured a certain way, ask or check git blame.

## Rule 9 — Tests verify intent, not just behavior
Tests must encode WHY behavior matters, not just WHAT it does.
A test that can't fail when business logic changes is wrong.

## Rule 10 — Checkpoint after every significant step
Summarize what was done, what's verified, what's left.
Don't continue from a state you can't describe back.
If you lose track, stop and restate.

## Rule 11 — Match the codebase's conventions, even if you disagree
Conformance > taste inside the codebase.
If you genuinely think a convention is harmful, surface it.
Don't fork silently.

## Rule 12 — Fail loud
"Completed" is wrong if anything was skipped silently.
"Tests pass" is wrong if any were skipped.
Default to surfacing uncertainty, not hiding it.

<!-- BEGIN:go-agent-rules -->
# This is a Go / Gin / GORM service — stack-specific hard rules

This repo is **Go 1.25 + Gin + GORM + PostgreSQL** (`github.com/exnodes/hrm-api`).
APIs and conventions may differ from your training data — verify against
`go.mod`, the `Makefile`, and existing code before writing. Heed these
non-negotiable rules:

- **Migrations are versioned SQL only.** Create up/down pairs in `migrations/`
  via `make migrate-new name=<snake>`. `db.AutoMigrate()` is **prohibited** —
  the server asserts the applied migration version on boot and refuses to
  start if behind or dirty (`internal/config`).
- **Every entity** carries the four audit columns `created_at`,
  `updated_at`, `is_deleted BOOLEAN`, `deleted_at TIMESTAMPTZ`, plus a
  per-table `BEFORE UPDATE` trigger calling `set_updated_at()`. PKs are
  UUIDs via `gen_random_uuid()` (pgcrypto).
- **Soft delete** uses the custom `NotDeleted` GORM scope — **never**
  GORM's built-in `gorm.DeletedAt`.
- **Layering is one-directional:** `handler → service → repository → GORM`.
  Handlers never touch the DB directly; services never import `gin`;
  repositories expose interfaces.
- **Error model:** services return `*errors.AppError`; the `ErrorHandler`
  middleware renders the JSON response envelope. Don't write ad-hoc
  `c.JSON(...)` error bodies.
- **Before claiming done:** run `make fmt && make vet && make test`.
  When handler/Swagger annotations change, regenerate docs with
  `make swag` — never hand-edit `docs/swagger/`.
<!-- END:go-agent-rules -->
