# Task Completion — Pre-Commit / Pre-PR Checklist

Run these in order before claiming a task done. "Tests pass" is false if any were skipped (AGENTS.md Rule 12).

## Always

```bash
make fmt    # gofmt -s -w .
make vet    # go vet ./...
make test   # go test ./...
```

`make test` requires a reachable Postgres test DB. If absent: `make test-db-up` first (idempotent).

## When handler annotations / Swagger comments changed

```bash
make swag   # regenerate docs/swagger/ — NEVER hand-edit the output
```

Then `git add docs/swagger/` so the regenerated artifacts ship with the code change.

## When schema changed (new/modified migration)

1. `make migrate-up` against the local dev DB to confirm it applies cleanly.
2. `make migrate-down` then `make migrate-up` again to confirm the down migration is reversible (catch missing-DROP / order bugs).
3. Check `make migrate-version` matches the new head.
4. Update the boot-time required-version assert in `internal/config` if applicable.

## Verification (end-to-end, per phase)

Unit tests alone are insufficient. A phase is not done until:

1. Server runs (`make run`) without boot errors.
2. The new API flow is exercised against a real DB (curl / integration tests) — golden path + at least one failure path.
3. DB state is spot-checked (psql) for the audit columns + soft-delete behaviour.
4. A verification log is committed at `docs/superpowers/verification/phase-NN.md` with the actual requests/responses + commit list.

## Always update the checkpoint at session end

Even on failure/interruption — write what was done, what's verified, what's next, blockers, into `docs/superpowers/CHECKPOINT.md`. Replace in place; do not append siblings (AGENTS.md Rule 10).

## Never

- `--no-verify` on git commit unless the user explicitly asks.
- `db.AutoMigrate()` to "fix" a schema gap — always a new SQL migration.
- Hand-edit `docs/swagger/` — regenerate via `make swag`.
- Mark complete with skipped tests.
