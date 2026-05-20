# Session Boot / Resume Protocol

**Goal:** orient a fresh session in <2 minutes without scanning the source tree. Follow this order — do NOT skip steps and do NOT scan `internal/` before steps 1–3.

## The order

1. **Read `docs/superpowers/CHECKPOINT.md`.** Single source of truth for: current phase, what's verified, what's next, known follow-ups, blockers, local-env notes. Replace in place — do not append siblings.

2. **Read the relevant phase plan.** Located at `docs/superpowers/plans/2026-05-15-phase-NN-*.md`. **Always look for the `## ⚠️ REVISION NOTES` block at the top of the plan first** — every phase so far (4, 5) has had load-bearing corrections to the draft task bodies. Execute per REVISION NOTES, not the raw task bodies where they conflict.

3. **Read prior verification logs** as needed: `docs/superpowers/verification/phase-NN.md`. They document end-to-end live proof + load-bearing fixes surfaced during verification. Each phase's log embeds the commit list for that phase.

4. **Read agent-memory files** for per-agent context: `.claude/agent-memory/<agent-name>/MEMORY.md` (index) + linked files. Local only; not committed.

5. **Only now** read targeted source files for the specific question. Prefer Serena's symbol-level tools (`find_symbol`, `get_references`, `find_referencing_symbols`) over `grep` / `cat` of full files — they're cheaper and more precise.

## What to update before ending a session

Update `docs/superpowers/CHECKPOINT.md` at the end of every significant session:

- What was done
- What is verified (link to the verification log if you ran one)
- What is next
- Blockers

Keep it concise. If a session was interrupted or failed, still update it noting the failure. This is AGENTS.md Rule 10 — applies to every agent.

## Anti-patterns

- ❌ `find /Users/.../exnodes-hrm-api-go-v2 -name "*.go"` — wastes tokens, ignores existing knowledge
- ❌ `ls -R internal/` — same
- ❌ Reading every test file to figure out the data model — read CHECKPOINT.md and the schema in migrations/ instead
- ❌ Re-discovering REVISION NOTES content by reading the plan task bodies — start at the REVISION NOTES block
- ❌ Trusting plan task bodies blindly — every phase has had a revision-notes correction
- ❌ Trusting REVISION NOTES blindly — Phase 5 verification surfaced one wrong claim ("no seed gap to close") that turned out to need a fix

## Tooling notes

- **Subagent dispatch** has been unavailable across the last two sessions (Phase 4 + Phase 5) despite the `ennam-dev-agent-team` plugin being enabled. Every probed `subagent_type` returns "not found" with an empty available-agents list. When subagents come back online, the recommended pattern is `superpowers:subagent-driven-development` — one subagent per task or per logical batch, with `project-owner` only orchestrating.
- **Serena MCP**: this very directory (`.serena/memories/`) is read on activation. Prefer Serena tools over generic `grep`/`cat` for symbol lookup.
