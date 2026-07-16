# Verification — Mobile forgot-password (OTP)

**Feature:** `DR-001-001-02-forgot-password.md` (MOBILE-APP / EP-001 / US-001)
**Branch:** `feat/mobile-forgot-password-otp`
**Migration:** 000027 (`password_reset_otps`)
**Date:** 2026-07-16

## What this is

The mobile OTP forgot-password flow. It sits alongside — and does not touch —
the existing **web** flow (`POST /auth/forgot-password`, one-click link,
migration 000024). Two endpoints are new; Screen 3 finishes on the existing
shared `POST /auth/reset-password`.

| Endpoint | Purpose |
|---|---|
| `POST /api/v1/auth/mobile/forgot-password` | Mail a 6-digit code. Also backs "Resend Code" (Alt-4). |
| `POST /api/v1/auth/mobile/verify-otp` | Consume the code, return a 10-min single-use `reset_token`. |
| `POST /api/v1/auth/reset-password` (existing) | Set the new password using that token. |

## Locked decisions (DR conflicts resolved with the project owner)

The DR contradicts itself and the codebase in three places. Resolutions:

1. **SR-05 over Alt-3** — the code is consumed at verify. Alt-3's "OTP remains
   valid if the user backs out of Screen 3" is **not** implemented; the app must
   hold the `reset_token`. Pinned by `TestOTPVerify_ReplayAfterConsume_Rejected`.
2. **AC-05 / SR-08 over the web flow's enumerate guard** — an unknown or
   deactivated email returns **404**, so the app can show the inline "No account
   found" error. This is deliberately an account-existence oracle and diverges
   from `POST /auth/forgot-password`, which still always returns 200. The DR
   flags SR-08 as unconfirmed with security — **still worth a security sign-off.**
3. **SR-09 deferred** — sessions are NOT invalidated across devices. `UpdatePassword`
   stamps `users.password_reset_at`, but nothing enforces it, so access tokens
   issued before a reset stay valid until their TTL runs out. Needs a JWT
   middleware change; logged as a follow-up.

Thresholds are the DR's stated numbers, all env-tunable (`OTP_*` in `.env.example`)
because the DR lists them as pending PO/security confirmation: 10-min code expiry,
3 requests / 15 min, 60s resend cooldown, 5 verify attempts, 10-min reset token.

## Added beyond the DR

**`OTP_MAX_VERIFY_ATTEMPTS` (default 5).** The DR rate-limits OTP *requests*
but never bounds *verify* attempts. A 6-digit code is only 10^6 wide, so one
issued code plus unlimited guesses is walkable inside the 10-minute window.
After 5 wrong submissions the code is burned.

## Evidence

### Build / vet / format

```
go build ./...   clean
go vet ./...     clean
gofmt -s -l .    no files from this change
make swag        regenerated; both /auth/mobile/* paths present in swagger.json
```

### Migration

`TestMain` does `Drop()` + `Up()`, so 000027 is applied from scratch on every
test run. Post-run: `schema_migrations` = **27**, `dirty` = f. Table shape
confirmed via `\d password_reset_otps` (11 columns, PK, 2 indexes, updated_at
trigger).

### Unit / integration — 14/14 PASS

`go test ./internal/services/ -run TestOTP` — 14 passed, 0 skipped, 14.5s.

| Test | Pins |
|---|---|
| `TestOTPRequest_UnknownEmail_404` | AC-05 / SR-08 |
| `TestOTPRequest_InactiveUser_404` | SR-01 — deactivated ≡ missing |
| `TestOTPRequest_CreatesHashedCodeAndSwallowsEmailFailure` | code hashed at rest; SMTP failure recorded, not raised |
| `TestOTPRequest_SupersedesPreviousCode` | SR-03 |
| `TestOTPRequest_ResendCooldown_429` | Alt-4 / UX-05 |
| `TestOTPRequest_RateLimitCountsSupersededCodes_429` | AC-06 / SR-04 + the soft-delete counting regression |
| `TestOTPRequest_RateLimitWindowRolls` | window releases |
| `TestOTPVerify_HappyPath_TokenResetsPassword` | AC-10 + AC-20 end to end |
| `TestOTPVerify_ReplayAfterConsume_Rejected` | SR-05 |
| `TestOTPVerify_WrongCode_ReportsRemainingAttempts` | AC-11 |
| `TestOTPVerify_BurnsCodeAfterMaxAttempts` | verify-attempt bound |
| `TestOTPVerify_ExpiredCode_400` | AC-12 |
| `TestOTPVerify_UnknownEmail_404` | AC-05 |
| `TestOTPVerify_CodeIsScopedToItsAccount` | a code cannot cross accounts |

### Live HTTP smoke — 14/14 as expected

Server on `:8082` against `exnodes_hrm_test` (migration 27), SMTP → Mailpit
(`:11025`). **Real email delivered and rendered** — the OTP template only
renders when SMTP is configured, so an SMTP-disabled smoke would not have
exercised it.

| # | Step | Result |
|---|---|---|
| 1 | unknown email | 404 `No account found with this email address` |
| 2 | malformed email | 400 |
| 3 | registered email | 200 + `expires_at` / `resend_available_at` |
| 4 | immediate resend | 429 `Please wait 60 seconds…` + `retry_after_seconds:60` |
| 5 | Mailpit received | subject `438350 is your Exnodes HRM (dev) verification code` |
| 6 | template render | code + name + 10-min expiry all interpolated |
| 7 | wrong code | 400 `Incorrect code…` + `remaining_attempts:4` |
| 8 | 5-digit code | 400 (binding `len=6`) |
| 9 | correct code | 200 + `reset_token` |
| 10 | replay same code | 400 `This code has expired…` (SR-05) |
| 11 | reset with token | 200 |
| 12 | login, new password | 200 |
| 13 | login, old password | 401 |
| 14 | reuse reset_token | 400 `Reset token has already been used` |

### DB spot-check (post-flow)

```
password_reset_otps:   code_hash=$2a$12$… (bcrypt cost 12), attempt_count=1,
                       consumed=t, is_deleted=f, last_email_error='' (no error)
password_reset_tokens: used_at set, ttl=00:09:59 (~10 min)
users:                 password_reset_at stamped
```

## Known gaps / follow-ups

- **SR-09 not implemented** — no cross-device session invalidation (decision 3).
- **SR-08 enumeration** — 404 on unknown email is an account-existence oracle,
  accepted deliberately; the DR itself wants security to confirm.
- **Pre-existing suite failures, NOT from this change** — `go test ./internal/services/`
  fails 4 leave-request tests on a clean `main` checkout too (verified by
  stashing this branch and re-running): `TestApprove_TeamScope_CanApproveSubordinate`,
  `TestApprove_TeamScope_RejectsNonSubordinate`, `TestReject_TeamScope_CanRejectSubordinate`,
  `TestUpdate_EmptyPatch_DoesNotRevertApprovedStatus`. All fail with
  `forbidden: Only an admin can edit an approved leave request`. Needs its own fix.
- **Dev DB `exnodes_hrm` is at migration 19** — 8 behind disk. The smoke ran
  against the test DB for that reason. `make migrate-up` before deploying.
- Password policy stays `min=8` (existing `ResetPasswordRequest` binding). The
  DR's policy questions are still open with the PO; the strength bar is a
  client-side concern.
