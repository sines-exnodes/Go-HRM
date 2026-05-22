# Phase 9: Email + Invite + Push Notification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

## ⚠️ REVISION NOTES (2026-05-22) — AUTHORITATIVE, read & apply before executing any task

This plan was drafted assuming Phase 8 would use migration `000010` (which the plan itself called `000010_seed_system_data` — never written). Actual sequence: Phase 6=`000009`, Phase 7=`000010`, Phase 8=`000011`. The codebase audit at the close of Phase 8 supersedes the task bodies below. **Execute per these notes, not the raw task bodies where they conflict.**

1. **Migration number = `000012_create_invites`** (NOT `000011`). Phase 8 already took `000011_create_system_config`. Final `make migrate-version` after Phase 9 = **12**.

2. **`invites.invited_by` FK target = `employees(id)`**, NOT `users(id)`. Mirrors `announcements.author_id` + `system_config.company_address_updated_by` per the Go schema split. Use `ON DELETE RESTRICT` to preserve the audit trail. The handler / service resolve the current user → employee via `s.emps.FindByUserID(currentUserID)` (same pattern as Phases 5/6/7/8).

3. **`invites.accepted_user_id` FK target = `users(id)`** (correct as drafted). Accept flow creates a fresh `users` row + a paired `employees` row; the marker is the auth identity, mirroring `announcement_views.user_id`.

4. **Config helpers**: `getEnvDefault` does NOT exist — use the existing `getEnv` helper. `getEnvInt` + `getEnvBool` + `getEnvFloat` already exist (added in Phase 6). No need to redeclare. Task 2 must use `getEnv` (not `getEnvDefault`).

5. **`UserService.CreateInternal` does NOT exist.** Phase 2 exposes `EmployeeService.Create(ctx, dto.EmployeeCreate)` which atomically creates user + employee in a transaction (precisely what `/invites/accept` needs). The InviteService.Accept method calls `empSvc.Create(...)`, not a standalone user-create. Construct the `dto.EmployeeCreate` from the invite row's fields (email, full_name, role_ids, department_id, position_id) plus the accept-time password.

6. **`pq.StringArray` won't work for UUID[] columns.** Use the `lib/pq` `pq.GenericArray{}` or the `gorm.io/datatypes` JSON array — but for cleanest typing, define a custom `UUIDArray` scanner/valuer that wraps `pq.Array(...)`. The plan's import of `"github.com/lib/pq"` is correct but `pq.StringArray` would cast UUIDs to strings; we need genuine UUID typing on the Go side.

7. **Existing handler helpers** — `currentUser(c)` + `parseIDParam(c, key)` are in [`internal/handlers/employee_handler.go`](../../internal/handlers/employee_handler.go). Reuse, don't redeclare.

8. **`apperrors.Write(c, err)` does NOT exist.** Use `_ = c.Error(err); return` per the established pattern.

9. **`RequirePerms` signature includes `authSvc *AuthService` as first arg** — `middleware.RequirePerms(authSvc, perm.PermInviteManage)`. Task 13 must include `authSvc` in every call.

10. **Routes wire in `cmd/server/main.go`**, NOT a `routes.Register(r, deps)` style file. Inline as in Phases 5/6/7/8.

11. **External dependencies (SMTP / FCM) must degrade gracefully.** When `SMTP_HOST=""`, EmailService logs the would-be email but does not fail invite creation. When `FIREBASE_CREDENTIALS_PATH=""`, PushClient is a no-op (matches the plan's intent). The InviteService.Create / Resend MUST NOT return 500 just because SMTP is misconfigured — store the error on `invites.last_email_error` and return 201/200 (per non-negotiable #7 in the plan).

12. **Permission constant — add `PermInviteManage = "invites:manage"`** to [`internal/permissions/registry.go`](../../internal/permissions/registry.go) (after `PermAnnounceManage`) AND seed it to Admin + HR Manager in [`internal/services/seed_service.go`](../../internal/services/seed_service.go). The spec §6.3 listing missed this constant. NOT using `PermUsersCreate` (which is for direct user CRUD) — invites are a separate surface with separate visibility (HR can invite without being able to admin-edit existing users).

13. **`truncateAll` test helper**: add `invites` to the TRUNCATE list before `employees` (since `invited_by` → `employees(id) RESTRICT`).

14. **External SMTP server for verification**: Mailpit. Run via `docker run -d --rm -p 11025:1025 -p 18025:8025 --name mailpit-phase09 axllent/mailpit` (use non-default host ports `11025/18025` to avoid colliding with other dev SMTP). Verification log inspects emails via `http://localhost:18025/api/v1/messages`. If Docker / Mailpit is unavailable in this environment, the e2e walk still verifies the invite-create-without-SMTP code path (200 + `last_email_error` populated).

15. **`encoding/base64` URL-safe token**: 32 random bytes encoded with `base64.RawURLEncoding` (no padding, URL-safe). 43 chars. Stored as `TEXT`; the partial unique index on `(token) WHERE is_deleted = false` covers collisions (vanishingly unlikely at that entropy, but defensive).

16. **Accept flow validation**:
    - Reject when `accepted_at IS NOT NULL` (already used) → 409.
    - Reject when `expires_at < NOW()` → 410 Gone (or 400; pick 400 for consistency with our error envelope which doesn't surface 410 as a code).
    - Reject when token doesn't match (soft-deleted invites included) → 404.
    - On success: create user+employee via `empSvc.Create`, then `UPDATE invites SET accepted_at = NOW(), accepted_user_id = new_user.id` in the same transaction.

17. **`Resend` semantics**: shipping the same token (no rotation) keeps the resend idempotent and lets a partially-delivered email still be valid. Stamp `updated_at` (the trigger does this) and clear `last_email_error` on success.

18. **`Revoke` is soft-delete**: `is_deleted=true, deleted_at=NOW()`. Subsequent `/invites/accept` with the same token must 404.

19. **Push test endpoint** (`POST /api/v1/notifications/test`) is admin-only (`RequirePerms(authSvc, permissions.PermUsersManageRoles)` per spec §6.3 — closest fit, since there's no dedicated `PermNotificationsTest`). Body: `{title, body, data?}`. Pushes to the calling admin's own device tokens (so they can self-test). If FCM is disabled, returns 200 with `{sent: 0, skipped: <count>}`.

20. **Phase 9 verification expects external services.** Set up Mailpit before T19. FCM stays disabled in dev (no FIREBASE_CREDENTIALS_PATH); the push-test endpoint returns the skipped-count payload. CHECKPOINT must document both.

Everything else in the task bodies (TDD-first, commit-per-task, no placeholders) still applies. **Execute per these REVISION NOTES, not the raw task bodies where they conflict.**

---


**Goal:** Port the Python email/invite/push-notification stack into `exnodes-hrm-api-go-v2`:

1. An SMTP-backed `EmailService` that renders the two HTML/plain-text templates (invite + password-reset) from the Python codebase and ships them via `gopkg.in/gomail.v2`.
2. A full **invite** module with its own `invites` table, soft-delete, audit columns, 6 endpoints (create / list / get / resend / revoke / accept), admin-gated except for the public `POST /api/v1/invites/accept` (creates the new user on success).
3. A pluggable `PushClient` interface with an FCM HTTP v1 implementation (matches Python's `firebase-admin` choice), plus an admin test endpoint that pushes to the caller's registered device tokens (from Phase 2's `device_tokens` table).

**Architecture:**
- The Python source stores invite tokens on the `users` collection (set-password flow on an already-created user). We **diverge** here per spec: invites become first-class rows that defer user creation until `/invites/accept`. This matches the BA intent ("invite an email address; user is created when they accept") and avoids partially-provisioned users sitting in the `users` table.
- Email sending is synchronous **inside a goroutine** so the HTTP request returns 200 even if SMTP is slow; failures log + flip an `invite.last_email_error` column rather than rolling back invite creation (per non-negotiable #7).
- Push notifications: an interface (`PushClient`) lets us swap providers and lets tests inject a fake. The default impl is FCM HTTP v1 (`https://fcm.googleapis.com/v1/projects/{projectID}/messages:send`) using a Google service-account access token obtained via `golang.org/x/oauth2/google`. If `FIREBASE_CREDENTIALS_PATH` is empty the client becomes a no-op logger (parity with Python's `_get_firebase_app()`).
- Verification environment uses **Mailpit** (`docker run -p 1025:1025 -p 8025:8025 axllent/mailpit`) as a local SMTP sink so e-mails can be inspected without external dependencies.

**Tech Stack additions for this phase:**
- `gopkg.in/gomail.v2` — SMTP client with multipart MIME builder.
- `golang.org/x/oauth2/google` + `golang.org/x/oauth2` — service-account access tokens for FCM HTTP v1.
- `embed` (stdlib) — to embed the two HTML templates as `*.html` files.

**Assumptions (must be true before starting):**
- Phases 0–8 are complete and committed: BaseModel, migrations 000001–000010, users, roles, user_roles, departments, positions, device_tokens (Phase 2), org settings.
- `internal/permissions/registry.go` already exports at least `PermUsersCreate`, `PermUsersRead`, `PermUsersDelete`, `PermOrgSettings`.
- `internal/services` already exposes `UserService` with a method that creates a user from `(email, password, fullName, roleIDs, departmentID, positionID)` — Phase 2 supplied this. Call it `UserService.CreateInternal(...)`. If the existing method has a different signature, use `Edit` to expose a thin adapter rather than changing call-sites.
- `internal/dto/response.go` already exposes `Response[T]` and `PaginatedData[T]`.
- `cmd/server/main.go` already constructs services and wires routes through a `routes.Register(r, deps)` style function.
- `Makefile` already has `migrate-up`, `migrate-down`, `migrate-version`, `swag`, `run`.

---

### Task 1: Add Phase-9 dependencies to go.mod

**Files:**
- Modify: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/go.mod`
- Modify: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/go.sum`

- [ ] **Step 1: Fetch the SMTP and OAuth2 libraries**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go get gopkg.in/gomail.v2@v2.0.0-20160411212932-81ebce5c23df && \
  go get golang.org/x/oauth2@v0.24.0 && \
  go get golang.org/x/oauth2/google@v0.24.0
```
Expected: lines like `go: added gopkg.in/gomail.v2 v2.0.0-...` and `go: added golang.org/x/oauth2 v0.24.0`. No errors.

- [ ] **Step 2: Tidy and verify the module compiles**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go mod tidy && \
  go build ./...
```
Expected: no output (clean build).

- [ ] **Step 3: Verify the new requires are present**

Run:
```bash
grep -E 'gomail\.v2|x/oauth2' /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/go.mod
```
Expected (order may vary):
```
	golang.org/x/oauth2 v0.24.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
```

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add go.mod go.sum && \
  git commit -m "chore(deps): add gomail.v2 + x/oauth2 for phase 9"
```
Expected: `2 files changed` summary.

---

### Task 2: Extend env + config for SMTP + FCM + frontend URL

**Files:**
- Modify: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/.env.example`
- Modify: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/config/config.go`

- [ ] **Step 1: Append Phase-9 keys to `.env.example`**

Edit `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/.env.example` and append exactly this block at the bottom (preserve any trailing newline):
```dotenv

# --- Phase 9: Email + Invite + Push ---
APP_NAME=Exnodes HRM
FRONTEND_URL=http://localhost:3000
INVITE_TOKEN_EXPIRE_HOURS=72

# SMTP — point at Mailpit for local dev: docker run -p 1025:1025 -p 8025:8025 axllent/mailpit
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM_EMAIL=no-reply@exnodes.local
SMTP_FROM_NAME=Exnodes HRM
SMTP_USE_TLS=false

# Firebase Cloud Messaging — leave empty to disable push in dev
FIREBASE_CREDENTIALS_PATH=
FIREBASE_PROJECT_ID=
```

- [ ] **Step 2: Add the new fields to the Config struct**

Open `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/config/config.go`, locate the `Config` struct, and add the fields immediately before the closing `}` of the struct (replace `<EXISTING_LAST_FIELD>` with the literal name of the field currently last in the struct, e.g. `JWTRefreshTTL`):

```go
	// Email + Invite + Push (Phase 9)
	AppName                 string
	FrontendURL             string
	InviteTokenExpireHours  int

	SMTPHost      string
	SMTPPort      int
	SMTPUser      string
	SMTPPassword  string
	SMTPFromEmail string
	SMTPFromName  string
	SMTPUseTLS    bool

	FirebaseCredentialsPath string
	FirebaseProjectID       string
```

And inside the `Load()` function (or whatever constructor populates `Config`), append these lines just before the `return &cfg, nil`:
```go
	cfg.AppName = getEnvDefault("APP_NAME", "Exnodes HRM")
	cfg.FrontendURL = getEnvDefault("FRONTEND_URL", "http://localhost:3000")
	cfg.InviteTokenExpireHours = getEnvInt("INVITE_TOKEN_EXPIRE_HOURS", 72)

	cfg.SMTPHost = getEnvDefault("SMTP_HOST", "")
	cfg.SMTPPort = getEnvInt("SMTP_PORT", 587)
	cfg.SMTPUser = getEnvDefault("SMTP_USER", "")
	cfg.SMTPPassword = getEnvDefault("SMTP_PASSWORD", "")
	cfg.SMTPFromEmail = getEnvDefault("SMTP_FROM_EMAIL", "")
	cfg.SMTPFromName = getEnvDefault("SMTP_FROM_NAME", "Exnodes HRM")
	cfg.SMTPUseTLS = getEnvBool("SMTP_USE_TLS", true)

	cfg.FirebaseCredentialsPath = getEnvDefault("FIREBASE_CREDENTIALS_PATH", "")
	cfg.FirebaseProjectID = getEnvDefault("FIREBASE_PROJECT_ID", "")
```

If `getEnvInt` or `getEnvBool` do not yet exist in the package, add them at the bottom of the file:
```go
func getEnvInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvBool(k string, def bool) bool {
	if v := os.Getenv(k); v != "" {
		switch strings.ToLower(v) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		}
	}
	return def
}
```
Add `"strconv"` and `"strings"` to the imports if not present.

- [ ] **Step 3: Build to confirm**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: no output.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add .env.example internal/config/config.go && \
  git commit -m "feat(config): add SMTP, FCM, frontend URL, invite expiry env vars"
```
Expected: `2 files changed`.

---

### Task 3: Write migration 000011_create_invites (up + down)

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/migrations/000011_create_invites.up.sql`
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/migrations/000011_create_invites.down.sql`

- [ ] **Step 1: Confirm the next sequence number is 011**

Run:
```bash
ls /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/migrations | sort | tail -5
```
Expected: the last `.up.sql` is `000010_seed_system_data.up.sql`.

- [ ] **Step 2: Write the up migration**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/migrations/000011_create_invites.up.sql` with exactly:
```sql
-- Phase 9: invites — admin-generated email invitations that create a user on accept.

CREATE TABLE invites (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    email           CITEXT       NOT NULL,
    full_name       TEXT         NULL,
    token           TEXT         NOT NULL,
    role_ids        UUID[]       NOT NULL DEFAULT '{}',
    department_id   UUID         NULL REFERENCES departments(id) ON DELETE SET NULL,
    position_id     UUID         NULL REFERENCES positions(id)   ON DELETE SET NULL,
    expires_at      TIMESTAMPTZ  NOT NULL,
    accepted_at     TIMESTAMPTZ  NULL,
    accepted_user_id UUID        NULL REFERENCES users(id)       ON DELETE SET NULL,
    invited_by      UUID         NOT NULL REFERENCES users(id)   ON DELETE RESTRICT,
    last_email_error TEXT        NULL,
    -- audit columns (Phase 0 conventions)
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    is_deleted      BOOLEAN      NOT NULL DEFAULT FALSE,
    deleted_at      TIMESTAMPTZ  NULL
);

CREATE UNIQUE INDEX uq_invites_token         ON invites(token) WHERE is_deleted = FALSE;
CREATE        INDEX idx_invites_email        ON invites(email);
CREATE        INDEX idx_invites_is_deleted   ON invites(is_deleted);
CREATE        INDEX idx_invites_expires_at   ON invites(expires_at);
CREATE        INDEX idx_invites_invited_by   ON invites(invited_by);

-- Partial unique: at most one *pending* (not-accepted, not-deleted) invite per email.
CREATE UNIQUE INDEX uq_invites_pending_email
    ON invites(email)
    WHERE accepted_at IS NULL AND is_deleted = FALSE;

CREATE TRIGGER trg_invites_set_updated_at
    BEFORE UPDATE ON invites
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

- [ ] **Step 3: Write the down migration**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/migrations/000011_create_invites.down.sql` with exactly:
```sql
DROP TRIGGER IF EXISTS trg_invites_set_updated_at ON invites;
DROP INDEX  IF EXISTS uq_invites_pending_email;
DROP INDEX  IF EXISTS idx_invites_invited_by;
DROP INDEX  IF EXISTS idx_invites_expires_at;
DROP INDEX  IF EXISTS idx_invites_is_deleted;
DROP INDEX  IF EXISTS idx_invites_email;
DROP INDEX  IF EXISTS uq_invites_token;
DROP TABLE  IF EXISTS invites;
```

- [ ] **Step 4: Apply and roll back on a scratch DB to verify both directions**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  make migrate-up && \
  make migrate-version && \
  make migrate-down && \
  make migrate-version && \
  make migrate-up && \
  make migrate-version
```
Expected: the three `migrate-version` lines print `11`, `10`, `11` respectively. No errors.

- [ ] **Step 5: Sanity-check the table shape**

Run (psql connection details come from `.env`):
```bash
psql "$DATABASE_URL" -c "\d invites" | head -40
```
Expected: `email` shown as `citext NOT NULL`, `token TEXT NOT NULL`, four audit columns present, `role_ids` typed `uuid[]`.

- [ ] **Step 6: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add migrations/000011_create_invites.up.sql migrations/000011_create_invites.down.sql && \
  git commit -m "feat(db): migration 000011 — invites table with audit cols and partial unique"
```
Expected: `2 files changed`.

---

### Task 4: Add Invite model

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/models/invite.go`

- [ ] **Step 1: Write the model**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/models/invite.go` with exactly:
```go
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Invite represents an admin-generated email invitation that, when accepted,
// creates a new User with the configured role/department/position.
type Invite struct {
	BaseModel

	Email     string  `gorm:"type:citext;not null;index"        json:"email"`
	FullName  *string `gorm:"type:text"                         json:"full_name,omitempty"`
	Token     string  `gorm:"type:text;not null;uniqueIndex"    json:"-"`
	RoleIDs   pq.StringArray `gorm:"type:uuid[];not null;default:'{}'" json:"role_ids"`
	DepartmentID *uuid.UUID `gorm:"type:uuid"                    json:"department_id,omitempty"`
	PositionID   *uuid.UUID `gorm:"type:uuid"                    json:"position_id,omitempty"`

	ExpiresAt       time.Time  `gorm:"not null"           json:"expires_at"`
	AcceptedAt      *time.Time `                          json:"accepted_at,omitempty"`
	AcceptedUserID  *uuid.UUID `gorm:"type:uuid"          json:"accepted_user_id,omitempty"`
	InvitedBy       uuid.UUID  `gorm:"type:uuid;not null;index" json:"invited_by"`
	LastEmailError  *string    `gorm:"type:text"          json:"last_email_error,omitempty"`
}

func (Invite) TableName() string { return "invites" }

// Status derives the current state from timestamps. Computed; not persisted.
func (i *Invite) Status(now time.Time) string {
	switch {
	case i.AcceptedAt != nil:
		return "accepted"
	case i.ExpiresAt.Before(now):
		return "expired"
	default:
		return "pending"
	}
}

// RoleUUIDs converts the pq.StringArray to []uuid.UUID, dropping invalid entries.
func (i *Invite) RoleUUIDs() []uuid.UUID {
	out := make([]uuid.UUID, 0, len(i.RoleIDs))
	for _, s := range i.RoleIDs {
		if id, err := uuid.Parse(s); err == nil {
			out = append(out, id)
		}
	}
	return out
}
```

- [ ] **Step 2: Add `github.com/lib/pq` if not already present**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go get github.com/lib/pq@v1.10.9 && \
  go mod tidy && \
  go build ./...
```
Expected: clean build (no output) — and `go.mod` now lists `github.com/lib/pq`.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/models/invite.go go.mod go.sum && \
  git commit -m "feat(models): add Invite model with role_ids pq.StringArray"
```

---

### Task 5: Invite DTOs

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/dto/invite.go`

- [ ] **Step 1: Write the DTOs**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/dto/invite.go` with exactly:
```go
package dto

import (
	"time"

	"github.com/google/uuid"
)

// InviteCreateRequest — admin invites an email.
type InviteCreateRequest struct {
	Email            string      `json:"email"              binding:"required,email"`
	FullName         *string     `json:"full_name,omitempty"`
	RoleIDs          []uuid.UUID `json:"role_ids"           binding:"required,min=1,dive,uuid"`
	DepartmentID     *uuid.UUID  `json:"department_id,omitempty"`
	PositionID       *uuid.UUID  `json:"position_id,omitempty"`
	ExpiresInHours   *int        `json:"expires_in_hours,omitempty" binding:"omitempty,min=1,max=720"`
}

// InviteRead — what we return on every admin endpoint.
type InviteRead struct {
	ID             uuid.UUID   `json:"id"`
	Email          string      `json:"email"`
	FullName       *string     `json:"full_name,omitempty"`
	RoleIDs        []uuid.UUID `json:"role_ids"`
	DepartmentID   *uuid.UUID  `json:"department_id,omitempty"`
	PositionID     *uuid.UUID  `json:"position_id,omitempty"`
	Status         string      `json:"status"` // pending | expired | accepted
	ExpiresAt      time.Time   `json:"expires_at"`
	AcceptedAt     *time.Time  `json:"accepted_at,omitempty"`
	AcceptedUserID *uuid.UUID  `json:"accepted_user_id,omitempty"`
	InvitedBy      uuid.UUID   `json:"invited_by"`
	LastEmailError *string     `json:"last_email_error,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

// InviteAcceptRequest — PUBLIC endpoint payload.
type InviteAcceptRequest struct {
	Token    string  `json:"token"    binding:"required,min=32"`
	Password string  `json:"password" binding:"required,min=8,max=128"`
	FullName *string `json:"full_name,omitempty"`
}

// InviteAcceptResponse — returned after a successful accept.
// We deliberately do NOT issue a JWT here; the FE redirects to /auth/login.
type InviteAcceptResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}

// InviteListQuery — admin list filters.
type InviteListQuery struct {
	Page     int    `form:"page"      binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=200"`
	Status   string `form:"status"    binding:"omitempty,oneof=pending expired accepted"`
	Email    string `form:"email"     binding:"omitempty"`
}

// NotificationTestRequest — payload for the admin push-test endpoint.
type NotificationTestRequest struct {
	Title string            `json:"title" binding:"required,min=1,max=200"`
	Body  string            `json:"body"  binding:"required,min=1,max=2000"`
	Data  map[string]string `json:"data,omitempty"`
}
```

- [ ] **Step 2: Build**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: clean build.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/dto/invite.go && \
  git commit -m "feat(dto): invite create/read/accept/list + notification test"
```

---

### Task 6: Invite repository

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/repositories/invite_repository.go`

- [ ] **Step 1: Write the repository**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/repositories/invite_repository.go` with exactly:
```go
package repositories

import (
	"context"
	"time"

	"github.com/exnodes/hrm-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InviteRepository interface {
	Create(ctx context.Context, inv *models.Invite) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Invite, error)
	GetByToken(ctx context.Context, token string) (*models.Invite, error)
	FindPendingByEmail(ctx context.Context, email string) (*models.Invite, error)
	Update(ctx context.Context, inv *models.Invite) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, status, email string, page, pageSize int) ([]models.Invite, int64, error)
}

type inviteRepo struct{ db *gorm.DB }

func NewInviteRepository(db *gorm.DB) InviteRepository { return &inviteRepo{db: db} }

func notDeleted(db *gorm.DB) *gorm.DB { return db.Where("is_deleted = ?", false) }

func (r *inviteRepo) Create(ctx context.Context, inv *models.Invite) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *inviteRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Invite, error) {
	var inv models.Invite
	if err := notDeleted(r.db.WithContext(ctx)).First(&inv, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *inviteRepo) GetByToken(ctx context.Context, token string) (*models.Invite, error) {
	var inv models.Invite
	if err := notDeleted(r.db.WithContext(ctx)).First(&inv, "token = ?", token).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *inviteRepo) FindPendingByEmail(ctx context.Context, email string) (*models.Invite, error) {
	var inv models.Invite
	err := notDeleted(r.db.WithContext(ctx)).
		Where("email = ? AND accepted_at IS NULL", email).
		First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *inviteRepo) Update(ctx context.Context, inv *models.Invite) error {
	return r.db.WithContext(ctx).Save(inv).Error
}

func (r *inviteRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&models.Invite{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]any{"is_deleted": true, "deleted_at": now}).Error
}

func (r *inviteRepo) List(ctx context.Context, status, email string, page, pageSize int) ([]models.Invite, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	q := notDeleted(r.db.WithContext(ctx)).Model(&models.Invite{})
	if email != "" {
		q = q.Where("email ILIKE ?", "%"+email+"%")
	}
	now := time.Now().UTC()
	switch status {
	case "pending":
		q = q.Where("accepted_at IS NULL AND expires_at >= ?", now)
	case "expired":
		q = q.Where("accepted_at IS NULL AND expires_at < ?", now)
	case "accepted":
		q = q.Where("accepted_at IS NOT NULL")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.Invite
	if err := q.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
```

- [ ] **Step 2: Build**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: clean build.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/repositories/invite_repository.go && \
  git commit -m "feat(repo): invite repository with pending/expired/accepted filters"
```

---

### Task 7: Email templates (embedded HTML + text)

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/templates/invite.html`
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/templates/invite.txt`
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/templates/reset.html`
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/templates/reset.txt`

- [ ] **Step 1: Create the templates directory**

Run:
```bash
mkdir -p /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/templates
```

- [ ] **Step 2: Write `invite.html` (Go text/template syntax — ported from Python)**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/templates/invite.html` with exactly:
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 30px; border-radius: 10px;">
        <h1 style="color: #2563eb; margin-bottom: 20px;">Welcome to {{.AppName}}!</h1>

        <p>Hi {{.FirstName}},</p>

        <p>You have been invited to join {{.AppName}}. Please click the button below to set your password and complete your registration:</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.InviteURL}}"
               style="background-color: #2563eb; color: white; padding: 14px 28px; text-decoration: none; border-radius: 6px; font-weight: bold; display: inline-block;">
                Accept Invitation
            </a>
        </div>

        <p style="color: #666; font-size: 14px;">
            Or copy and paste this link into your browser:<br>
            <a href="{{.InviteURL}}" style="color: #2563eb; word-break: break-all;">{{.InviteURL}}</a>
        </p>

        <p style="color: #666; font-size: 14px;">
            This link will expire in {{.ExpiresInHours}} hours.
        </p>

        <hr style="border: none; border-top: 1px solid #ddd; margin: 20px 0;">

        <p style="color: #999; font-size: 12px;">
            If you didn't expect this email, please ignore it or contact your administrator.
        </p>
    </div>
</body>
</html>
```

- [ ] **Step 3: Write `invite.txt`**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/templates/invite.txt` with exactly:
```
Welcome to {{.AppName}}!

Hi {{.FirstName}},

You have been invited to join {{.AppName}}. Visit the link below to set your password:

{{.InviteURL}}

This link will expire in {{.ExpiresInHours}} hours.

If you didn't expect this email, please ignore it or contact your administrator.
```

- [ ] **Step 4: Write `reset.html`**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/templates/reset.html` with exactly:
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 30px; border-radius: 10px;">
        <h1 style="color: #dc2626; margin-bottom: 20px;">Password Reset Request</h1>

        <p>Hi {{.FirstName}},</p>

        <p>We received a request to reset your password. Click the button below to set a new password:</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="{{.ResetURL}}"
               style="background-color: #dc2626; color: white; padding: 14px 28px; text-decoration: none; border-radius: 6px; font-weight: bold; display: inline-block;">
                Reset Password
            </a>
        </div>

        <p style="color: #666; font-size: 14px;">
            Or copy and paste this link into your browser:<br>
            <a href="{{.ResetURL}}" style="color: #dc2626; word-break: break-all;">{{.ResetURL}}</a>
        </p>

        <p style="color: #666; font-size: 14px;">
            This link will expire in {{.ExpiresInHours}} hours.
        </p>

        <p style="background-color: #fef2f2; padding: 15px; border-radius: 6px; color: #991b1b; font-size: 14px;">
            <strong>Important:</strong> Your current password remains valid until you complete this reset.
        </p>

        <hr style="border: none; border-top: 1px solid #ddd; margin: 20px 0;">

        <p style="color: #999; font-size: 12px;">
            If you didn't request this reset, please contact your administrator immediately.
        </p>
    </div>
</body>
</html>
```

- [ ] **Step 5: Write `reset.txt`**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/templates/reset.txt` with exactly:
```
Password Reset Request

Hi {{.FirstName}},

We received a request to reset your password. Visit the link below to set a new password:

{{.ResetURL}}

This link will expire in {{.ExpiresInHours}} hours.

If you didn't request this reset, please contact your administrator immediately.
```

- [ ] **Step 6: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/services/templates && \
  git commit -m "feat(email): embed invite + reset HTML/text templates ported from Python"
```

---

### Task 8: EmailService (SMTP via gomail.v2)

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/email_service.go`

- [ ] **Step 1: Write the service**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/email_service.go` with exactly:
```go
package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	texttemplate "text/template"

	"github.com/exnodes/hrm-api/internal/config"
	"gopkg.in/gomail.v2"
)

//go:embed templates/*.html templates/*.txt
var emailTemplates embed.FS

// EmailService renders templates and sends mail via SMTP.
// If SMTPHost is empty it logs and no-ops (parity with the Python flag).
type EmailService struct {
	cfg       *config.Config
	htmlTpls  *template.Template
	textTpls  *texttemplate.Template
	log       *slog.Logger
	dialer    *gomail.Dialer // nil when SMTP disabled
}

func NewEmailService(cfg *config.Config, log *slog.Logger) (*EmailService, error) {
	htmlTpls, err := template.ParseFS(emailTemplates, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse html email templates: %w", err)
	}
	textTpls, err := texttemplate.ParseFS(emailTemplates, "templates/*.txt")
	if err != nil {
		return nil, fmt.Errorf("parse text email templates: %w", err)
	}

	var dialer *gomail.Dialer
	if cfg.SMTPHost != "" {
		dialer = gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword)
		if !cfg.SMTPUseTLS {
			// Mailpit / MailHog accept plain — disable cert verification entirely.
			dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
			dialer.SSL = false
		}
	} else {
		log.Info("SMTP disabled — emails will be logged and dropped", "reason", "SMTP_HOST is empty")
	}

	return &EmailService{
		cfg:      cfg,
		htmlTpls: htmlTpls,
		textTpls: textTpls,
		log:      log,
		dialer:   dialer,
	}, nil
}

// Send sends an HTML + optional plain-text email.
func (s *EmailService) Send(ctx context.Context, to, subject, htmlBody, plainBody string) error {
	if s.dialer == nil {
		s.log.Info("email.send (smtp disabled)",
			"to", to, "subject", subject, "html_bytes", len(htmlBody), "text_bytes", len(plainBody))
		return nil
	}

	m := gomail.NewMessage()
	m.SetAddressHeader("From", s.cfg.SMTPFromEmail, s.cfg.SMTPFromName)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	if plainBody != "" {
		m.SetBody("text/plain", plainBody)
		m.AddAlternative("text/html", htmlBody)
	} else {
		m.SetBody("text/html", htmlBody)
	}

	done := make(chan error, 1)
	go func() { done <- s.dialer.DialAndSend(m) }()

	select {
	case err := <-done:
		if err != nil {
			s.log.Error("email.send failed", "to", to, "err", err)
			return err
		}
		s.log.Info("email.send ok", "to", to, "subject", subject)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// InviteTemplateData feeds invite.html / invite.txt.
type InviteTemplateData struct {
	AppName        string
	FirstName      string
	InviteURL      string
	ExpiresInHours int
}

// ResetTemplateData feeds reset.html / reset.txt.
type ResetTemplateData struct {
	AppName        string
	FirstName      string
	ResetURL       string
	ExpiresInHours int
}

// RenderInvite returns (html, text) for the invite template.
func (s *EmailService) RenderInvite(data InviteTemplateData) (string, string, error) {
	if data.AppName == "" {
		data.AppName = s.cfg.AppName
	}
	if data.ExpiresInHours == 0 {
		data.ExpiresInHours = s.cfg.InviteTokenExpireHours
	}
	var htmlBuf, textBuf bytes.Buffer
	if err := s.htmlTpls.ExecuteTemplate(&htmlBuf, "invite.html", data); err != nil {
		return "", "", fmt.Errorf("render invite html: %w", err)
	}
	if err := s.textTpls.ExecuteTemplate(&textBuf, "invite.txt", data); err != nil {
		return "", "", fmt.Errorf("render invite txt: %w", err)
	}
	return htmlBuf.String(), textBuf.String(), nil
}

// RenderReset returns (html, text) for the reset template.
func (s *EmailService) RenderReset(data ResetTemplateData) (string, string, error) {
	if data.AppName == "" {
		data.AppName = s.cfg.AppName
	}
	if data.ExpiresInHours == 0 {
		data.ExpiresInHours = s.cfg.InviteTokenExpireHours
	}
	var htmlBuf, textBuf bytes.Buffer
	if err := s.htmlTpls.ExecuteTemplate(&htmlBuf, "reset.html", data); err != nil {
		return "", "", fmt.Errorf("render reset html: %w", err)
	}
	if err := s.textTpls.ExecuteTemplate(&textBuf, "reset.txt", data); err != nil {
		return "", "", fmt.Errorf("render reset txt: %w", err)
	}
	return htmlBuf.String(), textBuf.String(), nil
}
```

- [ ] **Step 2: Build**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: clean build.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/services/email_service.go && \
  git commit -m "feat(email): EmailService with embedded templates and gomail dialer"
```

---

### Task 9: InviteService — Create / Get / List / Resend / Revoke

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/invite_service.go`

- [ ] **Step 1: Write the service (Accept is added in Task 10 to keep diffs reviewable)**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/invite_service.go` with exactly:
```go
package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/exnodes/hrm-api/internal/config"
	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type InviteService struct {
	repo     repositories.InviteRepository
	users    UserAccessor // narrow interface to UserService — defined below
	emails   *EmailService
	cfg      *config.Config
	log      *slog.Logger
	clock    func() time.Time
}

// UserAccessor isolates InviteService from the full UserService surface; only
// these two methods are used. The real UserService implements both.
type UserAccessor interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	// CreateInternal creates an active user with the given password + role/dept/position.
	// Returns the created user. password is the plaintext; the implementation hashes it.
	CreateInternal(ctx context.Context, email, password, fullName string,
		roleIDs []uuid.UUID, departmentID, positionID *uuid.UUID) (*models.User, error)
}

func NewInviteService(
	repo repositories.InviteRepository,
	users UserAccessor,
	emails *EmailService,
	cfg *config.Config,
	log *slog.Logger,
) *InviteService {
	return &InviteService{
		repo: repo, users: users, emails: emails,
		cfg: cfg, log: log, clock: func() time.Time { return time.Now().UTC() },
	}
}

// --- token --------------------------------------------------------------

// generateToken returns 64 hex characters from 32 random bytes (crypto/rand).
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// --- Create -------------------------------------------------------------

func (s *InviteService) Create(ctx context.Context, inviterID uuid.UUID, req dto.InviteCreateRequest) (*dto.InviteRead, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))

	// 1. Reject if an active user already exists for this email.
	if u, err := s.users.FindByEmail(ctx, email); err == nil && u != nil {
		return nil, apperrors.ErrConflict("a user with this email already exists")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 2. Reject if there's already a pending (not-accepted) invite for this email.
	if existing, err := s.repo.FindPendingByEmail(ctx, email); err == nil && existing != nil {
		return nil, apperrors.ErrConflict("a pending invite already exists for this email")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	hours := s.cfg.InviteTokenExpireHours
	if req.ExpiresInHours != nil && *req.ExpiresInHours > 0 {
		hours = *req.ExpiresInHours
	}
	expires := s.clock().Add(time.Duration(hours) * time.Hour)

	roleStrs := make(pq.StringArray, 0, len(req.RoleIDs))
	for _, id := range req.RoleIDs {
		roleStrs = append(roleStrs, id.String())
	}

	inv := &models.Invite{
		Email:        email,
		FullName:     req.FullName,
		Token:        token,
		RoleIDs:      roleStrs,
		DepartmentID: req.DepartmentID,
		PositionID:   req.PositionID,
		ExpiresAt:    expires,
		InvitedBy:    inviterID,
	}

	if err := s.repo.Create(ctx, inv); err != nil {
		return nil, err
	}

	// Fire-and-forget email — failures are recorded on the row.
	go s.dispatchInviteEmail(inv)

	out := s.toRead(inv)
	return &out, nil
}

func (s *InviteService) dispatchInviteEmail(inv *models.Invite) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	firstName := "there"
	if inv.FullName != nil && *inv.FullName != "" {
		firstName = strings.Fields(*inv.FullName)[0]
	}
	inviteURL := strings.TrimRight(s.cfg.FrontendURL, "/") + "/accept-invite?token=" + inv.Token

	html, text, err := s.emails.RenderInvite(InviteTemplateData{
		FirstName:      firstName,
		InviteURL:      inviteURL,
		ExpiresInHours: int(time.Until(inv.ExpiresAt).Hours()) + 1,
	})
	if err != nil {
		s.log.Error("invite.render failed", "invite_id", inv.ID, "err", err)
		s.recordEmailError(inv.ID, err)
		return
	}
	subject := s.cfg.AppName + " — You're invited"
	if err := s.emails.Send(ctx, inv.Email, subject, html, text); err != nil {
		s.recordEmailError(inv.ID, err)
		return
	}
	// Clear any previous error on success.
	s.recordEmailError(inv.ID, nil)
}

func (s *InviteService) recordEmailError(id uuid.UUID, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	inv, getErr := s.repo.GetByID(ctx, id)
	if getErr != nil {
		return
	}
	if err == nil {
		inv.LastEmailError = nil
	} else {
		msg := err.Error()
		inv.LastEmailError = &msg
	}
	_ = s.repo.Update(ctx, inv)
}

// --- Get / List / Resend / Revoke ---------------------------------------

func (s *InviteService) Get(ctx context.Context, id uuid.UUID) (*dto.InviteRead, error) {
	inv, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("invite")
		}
		return nil, err
	}
	out := s.toRead(inv)
	return &out, nil
}

func (s *InviteService) List(ctx context.Context, q dto.InviteListQuery) ([]dto.InviteRead, int64, error) {
	page, size := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	items, total, err := s.repo.List(ctx, q.Status, q.Email, page, size)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.InviteRead, 0, len(items))
	for i := range items {
		out = append(out, s.toRead(&items[i]))
	}
	return out, total, nil
}

func (s *InviteService) Resend(ctx context.Context, id uuid.UUID) (*dto.InviteRead, error) {
	inv, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("invite")
		}
		return nil, err
	}
	if inv.AcceptedAt != nil {
		return nil, apperrors.ErrBadRequest("invite already accepted")
	}

	token, err := generateToken()
	if err != nil {
		return nil, err
	}
	inv.Token = token
	inv.ExpiresAt = s.clock().Add(time.Duration(s.cfg.InviteTokenExpireHours) * time.Hour)
	inv.LastEmailError = nil

	if err := s.repo.Update(ctx, inv); err != nil {
		return nil, err
	}
	go s.dispatchInviteEmail(inv)

	out := s.toRead(inv)
	return &out, nil
}

func (s *InviteService) Revoke(ctx context.Context, id uuid.UUID) error {
	inv, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrNotFound("invite")
		}
		return err
	}
	if inv.AcceptedAt != nil {
		return apperrors.ErrBadRequest("cannot revoke an accepted invite")
	}
	return s.repo.SoftDelete(ctx, id)
}

// --- helpers -----------------------------------------------------------

func (s *InviteService) toRead(inv *models.Invite) dto.InviteRead {
	return dto.InviteRead{
		ID:             inv.ID,
		Email:          inv.Email,
		FullName:       inv.FullName,
		RoleIDs:        inv.RoleUUIDs(),
		DepartmentID:   inv.DepartmentID,
		PositionID:     inv.PositionID,
		Status:         inv.Status(s.clock()),
		ExpiresAt:      inv.ExpiresAt,
		AcceptedAt:     inv.AcceptedAt,
		AcceptedUserID: inv.AcceptedUserID,
		InvitedBy:      inv.InvitedBy,
		LastEmailError: inv.LastEmailError,
		CreatedAt:      inv.CreatedAt,
		UpdatedAt:      inv.UpdatedAt,
	}
}
```

- [ ] **Step 2: Verify the existing `UserService` satisfies `UserAccessor`**

Run:
```bash
grep -nE 'func \(.*UserService\) (FindByEmail|CreateInternal)' /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/user_service.go
```
Expected: both methods listed. If `CreateInternal` is missing or has a different signature, add it as a thin wrapper around whatever the Phase-2 admin-create method is called. **Do not rename existing methods** — only add a new wrapper method.

- [ ] **Step 3: Build**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: clean build.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/services/invite_service.go internal/services/user_service.go && \
  git commit -m "feat(invite): InviteService create/get/list/resend/revoke + email dispatch"
```

---

### Task 10: InviteService.Accept (public, transactional)

**Files:**
- Modify: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/invite_service.go`

- [ ] **Step 1: Append the Accept method**

Use Edit to insert this block at the end of the file (after the `toRead` function), keeping the existing imports valid:
```go

// --- Accept (PUBLIC) ---------------------------------------------------

func (s *InviteService) Accept(ctx context.Context, token, password string, fullName *string) (*dto.InviteAcceptResponse, error) {
	if len(password) < 8 {
		return nil, apperrors.ErrBadRequest("password must be at least 8 characters")
	}

	inv, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrBadRequest("invalid or expired invite token")
		}
		return nil, err
	}
	if inv.AcceptedAt != nil {
		return nil, apperrors.ErrBadRequest("invite already accepted")
	}
	if inv.ExpiresAt.Before(s.clock()) {
		return nil, apperrors.ErrBadRequest("invite has expired; please request a new one")
	}

	// Reject if a user has been created for this email since the invite was issued.
	if u, err := s.users.FindByEmail(ctx, inv.Email); err == nil && u != nil {
		return nil, apperrors.ErrConflict("a user with this email already exists")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	displayName := ""
	if fullName != nil && strings.TrimSpace(*fullName) != "" {
		displayName = strings.TrimSpace(*fullName)
	} else if inv.FullName != nil {
		displayName = *inv.FullName
	}

	user, err := s.users.CreateInternal(ctx, inv.Email, password, displayName, inv.RoleUUIDs(), inv.DepartmentID, inv.PositionID)
	if err != nil {
		return nil, err
	}

	now := s.clock()
	inv.AcceptedAt = &now
	inv.AcceptedUserID = &user.ID
	if err := s.repo.Update(ctx, inv); err != nil {
		// User created but invite update failed — log loudly. Manual cleanup may be needed.
		s.log.Error("invite.accept: failed to mark invite accepted after user creation",
			"invite_id", inv.ID, "user_id", user.ID, "err", err)
		return nil, err
	}

	return &dto.InviteAcceptResponse{UserID: user.ID, Email: user.Email}, nil
}
```

- [ ] **Step 2: Build**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: clean build.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/services/invite_service.go && \
  git commit -m "feat(invite): InviteService.Accept — public create-user-on-accept flow"
```

---

### Task 11: PushClient interface + FCM HTTP v1 implementation

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/push_client.go`

- [ ] **Step 1: Write the interface + FCM client + noop client**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/push_client.go` with exactly:
```go
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/exnodes/hrm-api/internal/config"
	"golang.org/x/oauth2/google"
)

// PushPayload is the provider-neutral push notification payload.
type PushPayload struct {
	Title string            `json:"title"`
	Body  string            `json:"body"`
	Data  map[string]string `json:"data,omitempty"`
}

// PushSendResult describes what happened when sending to a single device token.
type PushSendResult struct {
	DeviceToken string
	Success     bool
	Stale       bool // FCM said the token is unregistered / invalid — caller should delete it
	Err         error
}

// PushClient is the provider abstraction. Implementations: fcmClient, noopPushClient.
type PushClient interface {
	Send(ctx context.Context, deviceTokens []string, payload PushPayload) []PushSendResult
}

// NewPushClient picks the best implementation based on config.
func NewPushClient(cfg *config.Config, log *slog.Logger) (PushClient, error) {
	if cfg.FirebaseCredentialsPath == "" || cfg.FirebaseProjectID == "" {
		log.Info("push.client: FCM disabled — using noop client",
			"missing_credentials", cfg.FirebaseCredentialsPath == "",
			"missing_project_id", cfg.FirebaseProjectID == "")
		return &noopPushClient{log: log}, nil
	}
	keyJSON, err := os.ReadFile(cfg.FirebaseCredentialsPath)
	if err != nil {
		return nil, fmt.Errorf("read firebase credentials: %w", err)
	}
	creds, err := google.CredentialsFromJSON(context.Background(), keyJSON,
		"https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return nil, fmt.Errorf("parse firebase credentials: %w", err)
	}
	return &fcmClient{
		projectID: cfg.FirebaseProjectID,
		creds:     creds,
		http:      &http.Client{Timeout: 15 * time.Second},
		log:       log,
	}, nil
}

// --- noop client -------------------------------------------------------

type noopPushClient struct{ log *slog.Logger }

func (c *noopPushClient) Send(_ context.Context, tokens []string, payload PushPayload) []PushSendResult {
	out := make([]PushSendResult, len(tokens))
	for i, t := range tokens {
		c.log.Info("push.send (disabled — noop)", "token", maskToken(t), "title", payload.Title)
		out[i] = PushSendResult{DeviceToken: t, Success: true}
	}
	return out
}

// --- FCM HTTP v1 client -----------------------------------------------

type fcmClient struct {
	projectID string
	creds     *google.Credentials
	http      *http.Client
	log       *slog.Logger
}

// fcmMessage is the FCM HTTP v1 request body.
type fcmMessage struct {
	Message struct {
		Token        string            `json:"token"`
		Notification struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		} `json:"notification"`
		Data map[string]string `json:"data,omitempty"`
	} `json:"message"`
}

func (c *fcmClient) Send(ctx context.Context, tokens []string, payload PushPayload) []PushSendResult {
	out := make([]PushSendResult, len(tokens))

	tok, err := c.creds.TokenSource.Token()
	if err != nil {
		for i, t := range tokens {
			out[i] = PushSendResult{DeviceToken: t, Err: fmt.Errorf("acquire access token: %w", err)}
		}
		return out
	}
	bearer := tok.AccessToken
	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", c.projectID)

	for i, deviceToken := range tokens {
		var body fcmMessage
		body.Message.Token = deviceToken
		body.Message.Notification.Title = payload.Title
		body.Message.Notification.Body = payload.Body
		body.Message.Data = payload.Data
		raw, _ := json.Marshal(body)

		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer "+bearer)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			out[i] = PushSendResult{DeviceToken: deviceToken, Err: err}
			continue
		}
		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		switch {
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			out[i] = PushSendResult{DeviceToken: deviceToken, Success: true}
		case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
			// FCM uses 404 UNREGISTERED for stale tokens.
			out[i] = PushSendResult{
				DeviceToken: deviceToken, Stale: true,
				Err: errors.New("token unregistered: " + string(respBody)),
			}
		default:
			out[i] = PushSendResult{
				DeviceToken: deviceToken,
				Err:         fmt.Errorf("fcm http %d: %s", resp.StatusCode, string(respBody)),
			}
		}
		c.log.Info("push.send", "token", maskToken(deviceToken),
			"status", resp.StatusCode, "ok", out[i].Success, "stale", out[i].Stale)
	}
	return out
}

func maskToken(t string) string {
	if len(t) <= 8 {
		return "****"
	}
	return t[:4] + "…" + t[len(t)-4:]
}
```

- [ ] **Step 2: Build**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: clean build.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/services/push_client.go && \
  git commit -m "feat(push): PushClient interface + FCM HTTP v1 client + noop fallback"
```

---

### Task 12: PushNotificationService (lookup device tokens → dispatch)

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/push_notification_service.go`

- [ ] **Step 1: Confirm the Phase-2 device-token repository surface**

Run:
```bash
grep -nE 'func .* (FindByUser|FindByUserIDs|DeleteByToken|Delete) ' /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/repositories/device_token_repository.go || \
  echo 'PHASE2 device_token_repository missing — adapt the import path / method names accordingly'
```
Expected: either a list of methods (use them as-is) or the `PHASE2 device_token_repository missing` line. If missing, the simplest fix is to use whatever the repo is named; both `DeviceTokenRepo` and `UserDeviceTokenRepo` are acceptable — adjust the imports/types in Step 2 to match.

- [ ] **Step 2: Write the service**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/push_notification_service.go` with exactly:
```go
package services

import (
	"context"
	"log/slog"

	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/google/uuid"
)

// DeviceTokenRepoView is the minimal subset of the Phase-2 device-token
// repository that PushNotificationService needs. The real repo must implement it.
type DeviceTokenRepoView interface {
	ListTokensForUser(ctx context.Context, userID uuid.UUID) ([]string, error)
	ListTokensForUsers(ctx context.Context, userIDs []uuid.UUID) ([]string, error)
	ListAllTokens(ctx context.Context) ([]string, error)
	DeleteByToken(ctx context.Context, token string) error
}

type PushNotificationService struct {
	client PushClient
	tokens DeviceTokenRepoView
	log    *slog.Logger
}

func NewPushNotificationService(client PushClient, tokens DeviceTokenRepoView, log *slog.Logger) *PushNotificationService {
	return &PushNotificationService{client: client, tokens: tokens, log: log}
}

func (s *PushNotificationService) NotifyUser(ctx context.Context, userID uuid.UUID, payload PushPayload) ([]PushSendResult, error) {
	tokens, err := s.tokens.ListTokensForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.dispatch(ctx, tokens, payload), nil
}

func (s *PushNotificationService) NotifyUsers(ctx context.Context, userIDs []uuid.UUID, payload PushPayload) ([]PushSendResult, error) {
	tokens, err := s.tokens.ListTokensForUsers(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	return s.dispatch(ctx, tokens, payload), nil
}

func (s *PushNotificationService) NotifyAll(ctx context.Context, payload PushPayload) ([]PushSendResult, error) {
	tokens, err := s.tokens.ListAllTokens(ctx)
	if err != nil {
		return nil, err
	}
	return s.dispatch(ctx, tokens, payload), nil
}

func (s *PushNotificationService) dispatch(ctx context.Context, tokens []string, payload PushPayload) []PushSendResult {
	if len(tokens) == 0 {
		s.log.Info("push.dispatch: no device tokens, skipping", "title", payload.Title)
		return nil
	}
	results := s.client.Send(ctx, tokens, payload)
	// Cull stale tokens — best-effort, never fails the request.
	for _, r := range results {
		if r.Stale {
			if err := s.tokens.DeleteByToken(ctx, r.DeviceToken); err != nil {
				s.log.Warn("push.dispatch: failed to delete stale token", "err", err)
			} else {
				s.log.Info("push.dispatch: pruned stale device token")
			}
		}
	}
	return results
}

// Ensure the concrete Phase-2 repository compiles against the view interface.
// This is a compile-time assertion; remove or adjust if the Phase-2 repo name differs.
var _ DeviceTokenRepoView = (repositories.DeviceTokenRepository)(nil)
```

If the compile-time assertion line at the bottom fails because the Phase-2 type is named differently (e.g., `UserDeviceTokenRepository`) or because it doesn't yet expose `ListTokensForUser` / `ListTokensForUsers` / `ListAllTokens` / `DeleteByToken`, do one of:

1. Add the missing methods to the existing Phase-2 repository file (preferred — they are pure read helpers).
2. Or delete the assertion and pass an adapter in `main.go`.

Either way **do not change** existing method signatures used by other code.

- [ ] **Step 3: Build**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: clean build. If it fails on the compile-time assertion, follow the note above, then rebuild.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/services/push_notification_service.go internal/repositories/device_token_repository.go && \
  git commit -m "feat(push): PushNotificationService with stale-token pruning"
```

---

### Task 13: Invite handler (admin endpoints + public accept)

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/handlers/invite_handler.go`

- [ ] **Step 1: Write the handler**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/handlers/invite_handler.go` with exactly:
```go
package handlers

import (
	"net/http"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/middleware"
	"github.com/exnodes/hrm-api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InviteHandler struct {
	svc *services.InviteService
}

func NewInviteHandler(svc *services.InviteService) *InviteHandler {
	return &InviteHandler{svc: svc}
}

// Create godoc
// @Summary      Create an invite
// @Description  Generate an invite token and email it to the recipient. Requires users:create.
// @Tags         invites
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body dto.InviteCreateRequest true "Invite payload"
// @Success      201 {object} dto.Response[dto.InviteRead]
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Router       /api/v1/invites [post]
func (h *InviteHandler) Create(c *gin.Context) {
	var req dto.InviteCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}
	inviterID := middleware.UserIDFromContext(c)
	out, err := h.svc.Create(c.Request.Context(), inviterID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[dto.InviteRead]{Success: true, Data: *out})
}

// List godoc
// @Summary      List invites
// @Tags         invites
// @Produce      json
// @Security     BearerAuth
// @Param        page      query int    false "page (default 1)"
// @Param        page_size query int    false "page size (default 20, max 200)"
// @Param        status    query string false "pending | expired | accepted"
// @Param        email     query string false "filter by email (ILIKE)"
// @Success      200 {object} dto.Response[dto.PaginatedData[dto.InviteRead]]
// @Router       /api/v1/invites [get]
func (h *InviteHandler) List(c *gin.Context) {
	var q dto.InviteListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(err)
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	page, size := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	totalPages := int(total / int64(size))
	if total%int64(size) != 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, dto.Response[dto.PaginatedData[dto.InviteRead]]{
		Success: true,
		Data: dto.PaginatedData[dto.InviteRead]{
			Items: items, Total: total, Page: page, PageSize: size, TotalPages: totalPages,
		},
	})
}

// Get godoc
// @Summary      Get an invite
// @Tags         invites
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "invite id (UUID)"
// @Success      200 {object} dto.Response[dto.InviteRead]
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/v1/invites/{id} [get]
func (h *InviteHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.InviteRead]{Success: true, Data: *out})
}

// Resend godoc
// @Summary      Resend an invite (rotates token + extends expiry)
// @Tags         invites
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "invite id (UUID)"
// @Success      200 {object} dto.Response[dto.InviteRead]
// @Failure      400 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/v1/invites/{id}/resend [post]
func (h *InviteHandler) Resend(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Resend(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, dto.Response[dto.InviteRead]{Success: true, Data: *out})
}

// Revoke godoc
// @Summary      Revoke (soft-delete) an invite
// @Tags         invites
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "invite id (UUID)"
// @Success      204 "no content"
// @Failure      400 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/v1/invites/{id} [delete]
func (h *InviteHandler) Revoke(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.Revoke(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Accept godoc
// @Summary      Accept an invite (PUBLIC — no auth)
// @Description  Public endpoint. Validates token + creates the user. The caller logs in via /auth/login afterwards.
// @Tags         invites
// @Accept       json
// @Produce      json
// @Param        body body dto.InviteAcceptRequest true "Accept payload"
// @Success      201 {object} dto.Response[dto.InviteAcceptResponse]
// @Failure      400 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Router       /api/v1/invites/accept [post]
func (h *InviteHandler) Accept(c *gin.Context) {
	var req dto.InviteAcceptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}
	out, err := h.svc.Accept(c.Request.Context(), req.Token, req.Password, req.FullName)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, dto.Response[dto.InviteAcceptResponse]{Success: true, Data: *out})
}
```

- [ ] **Step 2: Build**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: clean build. If `middleware.UserIDFromContext` is named differently in your Phase-1 codebase (e.g., `CurrentUserID`), adjust the call accordingly — do not rename the existing helper.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/handlers/invite_handler.go && \
  git commit -m "feat(invite): handler — admin CRUD + public accept with Swagger"
```

---

### Task 14: Notification test handler (admin debug push)

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/handlers/notification_handler.go`

- [ ] **Step 1: Write the handler**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/handlers/notification_handler.go` with exactly:
```go
package handlers

import (
	"net/http"

	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/middleware"
	"github.com/exnodes/hrm-api/internal/services"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	push *services.PushNotificationService
}

func NewNotificationHandler(push *services.PushNotificationService) *NotificationHandler {
	return &NotificationHandler{push: push}
}

// SendTest godoc
// @Summary      Send a test push to the caller's registered devices
// @Description  Admin-only debug endpoint. Requires organization_settings:manage.
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body dto.NotificationTestRequest true "Push payload"
// @Success      200 {object} dto.Response[map[string]any]
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/v1/notifications/test [post]
func (h *NotificationHandler) SendTest(c *gin.Context) {
	var req dto.NotificationTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}
	userID := middleware.UserIDFromContext(c)
	results, err := h.push.NotifyUser(c.Request.Context(), userID, services.PushPayload{
		Title: req.Title, Body: req.Body, Data: req.Data,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}
	successes := 0
	stale := 0
	for _, r := range results {
		if r.Success {
			successes++
		}
		if r.Stale {
			stale++
		}
	}
	c.JSON(http.StatusOK, dto.Response[map[string]any]{
		Success: true,
		Data: map[string]any{
			"attempted":     len(results),
			"successes":     successes,
			"stale_pruned":  stale,
			"results":       results,
		},
	})
}
```

- [ ] **Step 2: Build**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && go build ./...
```
Expected: clean build.

- [ ] **Step 3: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/handlers/notification_handler.go && \
  git commit -m "feat(push): admin test-push handler"
```

---

### Task 15: Wire services, repositories, and routes into main.go

**Files:**
- Modify: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/cmd/server/main.go`
- Possibly modify: the routes file used by Phase 1+ (e.g. `internal/handlers/routes.go` or similar)

- [ ] **Step 1: Locate where existing services + routes are wired**

Run:
```bash
grep -nE 'NewUserService|services\.NewUser|RouterGroup|v1 :=|v1\.Group' \
  /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/cmd/server/main.go \
  /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/handlers/*.go 2>/dev/null
```
Capture the file/line where:
- the `*gorm.DB` `db` is in scope
- the `*slog.Logger` `log` is in scope
- `v1 := r.Group("/api/v1")` (or equivalent) is declared
- the `authed` group with `middleware.JWT()` is declared

- [ ] **Step 2: Wire the Phase-9 dependencies near the other services**

In the same block that constructs `userSvc`, `deptSvc`, etc., add (preserve existing variable naming conventions):
```go
	emailSvc, err := services.NewEmailService(cfg, log)
	if err != nil {
		log.Error("init email service", "err", err)
		os.Exit(1)
	}

	pushClient, err := services.NewPushClient(cfg, log)
	if err != nil {
		log.Error("init push client", "err", err)
		os.Exit(1)
	}

	inviteRepo := repositories.NewInviteRepository(db)
	inviteSvc := services.NewInviteService(inviteRepo, userSvc, emailSvc, cfg, log)

	pushSvc := services.NewPushNotificationService(pushClient, deviceTokenRepo, log)
```

(Replace `deviceTokenRepo` with the actual variable name used by Phase 2 — `grep deviceToken cmd/server/main.go` to find it.)

- [ ] **Step 3: Construct the handlers**

In the handler-construction block:
```go
	inviteHandler := handlers.NewInviteHandler(inviteSvc)
	notifHandler := handlers.NewNotificationHandler(pushSvc)
```

- [ ] **Step 4: Register routes — admin endpoints under the authed group**

Find the place where `authed := v1.Group("")` exists (or its analog) and append:
```go
	// --- Phase 9: invites (admin) ---
	invitesAdmin := authed.Group("/invites")
	invitesAdmin.POST("",            middleware.RequirePerms(permissions.PermUsersCreate), inviteHandler.Create)
	invitesAdmin.GET("",             middleware.RequirePerms(permissions.PermUsersRead),   inviteHandler.List)
	invitesAdmin.GET("/:id",         middleware.RequirePerms(permissions.PermUsersRead),   inviteHandler.Get)
	invitesAdmin.POST("/:id/resend", middleware.RequirePerms(permissions.PermUsersCreate), inviteHandler.Resend)
	invitesAdmin.DELETE("/:id",      middleware.RequirePerms(permissions.PermUsersDelete), inviteHandler.Revoke)

	// --- Phase 9: notifications test (admin) ---
	authed.POST("/notifications/test", middleware.RequirePerms(permissions.PermOrgSettings), notifHandler.SendTest)
```

- [ ] **Step 5: Register the PUBLIC accept route directly on `v1` — outside `authed`**

In the route file, immediately above (or below) the authed group registration:
```go
	// PUBLIC — no JWT, no permission check. Caller must hold a valid invite token.
	v1.POST("/invites/accept", inviteHandler.Accept)
```

Add a short comment block above the route so reviewers can grep it easily:
```go
	// =================================================================
	// PUBLIC ROUTES (no auth) — keep this list small and reviewed.
	// =================================================================
```

- [ ] **Step 6: Build and run swag**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go build ./... && \
  make swag
```
Expected: build succeeds; `make swag` regenerates `docs/swagger/` without errors.

- [ ] **Step 7: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add cmd/server/main.go internal/handlers docs/swagger && \
  git commit -m "feat(routes): wire invite + notification routes, regenerate swagger"
```

---

### Task 16: Service tests — invite_service_test.go

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/invite_service_test.go`

- [ ] **Step 1: Inspect the existing `testhelper_test.go` to learn the factory style**

Run:
```bash
ls /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/*test*.go
head -100 /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/testhelper_test.go
```
Expected: the helper provides a `newTestDB(t)` returning a real `*gorm.DB` connected to `exnodes_hrm_test`, plus `makeAdmin(t, db)`, `makeReadonly(t, db)`, and a per-test cleanup hook. The test below uses **only** `newTestDB` so it works even before Phase 2's `makeAdmin` factory exists; adapt freely if helpers are richer.

- [ ] **Step 2: Write the test file**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/invite_service_test.go` with exactly:
```go
package services_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/exnodes/hrm-api/internal/config"
	"github.com/exnodes/hrm-api/internal/dto"
	"github.com/exnodes/hrm-api/internal/models"
	"github.com/exnodes/hrm-api/internal/repositories"
	"github.com/exnodes/hrm-api/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- in-memory UserService stub ----------------------------------------

type fakeUsers struct {
	mu       sync.Mutex
	byEmail  map[string]*models.User
}

func newFakeUsers() *fakeUsers { return &fakeUsers{byEmail: map[string]*models.User{}} }

func (f *fakeUsers) FindByEmail(_ context.Context, email string) (*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if u, ok := f.byEmail[email]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeUsers) CreateInternal(_ context.Context, email, _password, fullName string,
	_roleIDs []uuid.UUID, _dept, _pos *uuid.UUID) (*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.byEmail[email]; ok {
		return nil, errors.New("duplicate user")
	}
	u := &models.User{Email: email, FullName: fullName}
	u.ID = uuid.New()
	f.byEmail[email] = u
	return u, nil
}

// --- fixture builder ---------------------------------------------------

func newInviteFixture(t *testing.T) (*services.InviteService, repositories.InviteRepository, *fakeUsers, *config.Config) {
	t.Helper()
	db := newTestDB(t) // from testhelper_test.go

	cfg := &config.Config{
		AppName:                "Test HRM",
		FrontendURL:            "http://localhost:3000",
		InviteTokenExpireHours: 24,
		// no SMTP -> EmailService becomes a no-op logger.
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	emails, err := services.NewEmailService(cfg, log)
	require.NoError(t, err)

	repo := repositories.NewInviteRepository(db)
	fake := newFakeUsers()
	svc := services.NewInviteService(repo, fake, emails, cfg, log)
	return svc, repo, fake, cfg
}

// --- tests -------------------------------------------------------------

func TestInvite_Create_HappyPath(t *testing.T) {
	svc, _, _, _ := newInviteFixture(t)
	ctx := context.Background()

	out, err := svc.Create(ctx, uuid.New(), dto.InviteCreateRequest{
		Email:   "alice@example.com",
		RoleIDs: []uuid.UUID{uuid.New()},
	})
	require.NoError(t, err)
	require.Equal(t, "alice@example.com", out.Email)
	require.Equal(t, "pending", out.Status)
	require.True(t, out.ExpiresAt.After(time.Now().UTC()))
}

func TestInvite_Create_DuplicateEmail_Conflict(t *testing.T) {
	svc, _, fake, _ := newInviteFixture(t)
	fake.byEmail["bob@example.com"] = &models.User{Email: "bob@example.com"}

	_, err := svc.Create(context.Background(), uuid.New(), dto.InviteCreateRequest{
		Email: "bob@example.com", RoleIDs: []uuid.UUID{uuid.New()},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
}

func TestInvite_Create_DuplicatePending_Conflict(t *testing.T) {
	svc, _, _, _ := newInviteFixture(t)
	ctx := context.Background()
	_, err := svc.Create(ctx, uuid.New(), dto.InviteCreateRequest{
		Email: "c@example.com", RoleIDs: []uuid.UUID{uuid.New()},
	})
	require.NoError(t, err)
	_, err = svc.Create(ctx, uuid.New(), dto.InviteCreateRequest{
		Email: "c@example.com", RoleIDs: []uuid.UUID{uuid.New()},
	})
	require.Error(t, err)
}

func TestInvite_Accept_HappyPath(t *testing.T) {
	svc, repo, fake, _ := newInviteFixture(t)
	ctx := context.Background()
	out, err := svc.Create(ctx, uuid.New(), dto.InviteCreateRequest{
		Email: "d@example.com", RoleIDs: []uuid.UUID{uuid.New()},
	})
	require.NoError(t, err)

	// Grab raw row so we can read the token (DTO hides it).
	inv, err := repo.GetByID(ctx, out.ID)
	require.NoError(t, err)

	resp, err := svc.Accept(ctx, inv.Token, "secret123", nil)
	require.NoError(t, err)
	require.Equal(t, "d@example.com", resp.Email)
	require.NotNil(t, fake.byEmail["d@example.com"])

	// Invite is now marked accepted.
	inv2, _ := repo.GetByID(ctx, out.ID)
	require.NotNil(t, inv2.AcceptedAt)
}

func TestInvite_Accept_Expired(t *testing.T) {
	svc, repo, _, _ := newInviteFixture(t)
	ctx := context.Background()
	out, err := svc.Create(ctx, uuid.New(), dto.InviteCreateRequest{
		Email: "e@example.com", RoleIDs: []uuid.UUID{uuid.New()},
	})
	require.NoError(t, err)

	// Backdate the expiry in the DB.
	inv, _ := repo.GetByID(ctx, out.ID)
	inv.ExpiresAt = time.Now().UTC().Add(-time.Hour)
	require.NoError(t, repo.Update(ctx, inv))

	_, err = svc.Accept(ctx, inv.Token, "secret123", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

func TestInvite_Accept_AlreadyAccepted(t *testing.T) {
	svc, repo, _, _ := newInviteFixture(t)
	ctx := context.Background()
	out, _ := svc.Create(ctx, uuid.New(), dto.InviteCreateRequest{
		Email: "f@example.com", RoleIDs: []uuid.UUID{uuid.New()},
	})
	inv, _ := repo.GetByID(ctx, out.ID)
	_, err := svc.Accept(ctx, inv.Token, "secret123", nil)
	require.NoError(t, err)

	_, err = svc.Accept(ctx, inv.Token, "secret123", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already accepted")
}

func TestInvite_Accept_InvalidToken(t *testing.T) {
	svc, _, _, _ := newInviteFixture(t)
	_, err := svc.Accept(context.Background(), "deadbeef-not-a-real-token", "secret123", nil)
	require.Error(t, err)
}

func TestInvite_Resend_RotatesToken(t *testing.T) {
	svc, repo, _, _ := newInviteFixture(t)
	ctx := context.Background()
	out, _ := svc.Create(ctx, uuid.New(), dto.InviteCreateRequest{
		Email: "g@example.com", RoleIDs: []uuid.UUID{uuid.New()},
	})
	before, _ := repo.GetByID(ctx, out.ID)

	_, err := svc.Resend(ctx, out.ID)
	require.NoError(t, err)

	after, _ := repo.GetByID(ctx, out.ID)
	require.NotEqual(t, before.Token, after.Token)
	require.True(t, after.ExpiresAt.After(before.ExpiresAt) || after.ExpiresAt.Equal(before.ExpiresAt))
}
```

- [ ] **Step 3: Run the invite tests**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go test ./internal/services -run 'TestInvite_' -v
```
Expected: all 7 tests `PASS`. If `newTestDB` is undefined, look at the existing `testhelper_test.go` to copy the correct helper name; the spec mandates this helper exists by Phase 2.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/services/invite_service_test.go && \
  git commit -m "test(invite): create/accept/expired/duplicate/resend"
```

---

### Task 17: Service tests — push + email

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/push_notification_service_test.go`
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/email_service_test.go`

- [ ] **Step 1: Write the push test**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/push_notification_service_test.go` with exactly:
```go
package services_test

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/exnodes/hrm-api/internal/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- in-memory device-token repo --------------------------------------

type fakeTokenRepo struct {
	mu     sync.Mutex
	tokens map[uuid.UUID][]string
	all    []string
	deleted []string
}

func (f *fakeTokenRepo) ListTokensForUser(_ context.Context, u uuid.UUID) ([]string, error) {
	f.mu.Lock(); defer f.mu.Unlock()
	return append([]string{}, f.tokens[u]...), nil
}
func (f *fakeTokenRepo) ListTokensForUsers(_ context.Context, us []uuid.UUID) ([]string, error) {
	f.mu.Lock(); defer f.mu.Unlock()
	out := []string{}
	for _, u := range us { out = append(out, f.tokens[u]...) }
	return out, nil
}
func (f *fakeTokenRepo) ListAllTokens(_ context.Context) ([]string, error) {
	f.mu.Lock(); defer f.mu.Unlock()
	return append([]string{}, f.all...), nil
}
func (f *fakeTokenRepo) DeleteByToken(_ context.Context, t string) error {
	f.mu.Lock(); defer f.mu.Unlock()
	f.deleted = append(f.deleted, t)
	return nil
}

// --- in-memory PushClient ----------------------------------------------

type capturingClient struct {
	sent [][]string
	payloads []services.PushPayload
	staleTokens map[string]bool // tokens to mark stale
}

func (c *capturingClient) Send(_ context.Context, tokens []string, p services.PushPayload) []services.PushSendResult {
	c.sent = append(c.sent, append([]string{}, tokens...))
	c.payloads = append(c.payloads, p)
	out := make([]services.PushSendResult, len(tokens))
	for i, t := range tokens {
		if c.staleTokens[t] {
			out[i] = services.PushSendResult{DeviceToken: t, Stale: true}
		} else {
			out[i] = services.PushSendResult{DeviceToken: t, Success: true}
		}
	}
	return out
}

// --- tests -------------------------------------------------------------

func TestPush_NotifyUser_FansOutToAllTokens(t *testing.T) {
	uid := uuid.New()
	repo := &fakeTokenRepo{tokens: map[uuid.UUID][]string{
		uid: {"tok-a", "tok-b", "tok-c"},
	}}
	cli := &capturingClient{}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := services.NewPushNotificationService(cli, repo, log)

	results, err := svc.NotifyUser(context.Background(), uid, services.PushPayload{
		Title: "hi", Body: "test",
	})
	require.NoError(t, err)
	require.Len(t, results, 3)
	require.Equal(t, []string{"tok-a", "tok-b", "tok-c"}, cli.sent[0])
	require.Equal(t, "hi", cli.payloads[0].Title)
}

func TestPush_NotifyUser_StaleTokenIsPruned(t *testing.T) {
	uid := uuid.New()
	repo := &fakeTokenRepo{tokens: map[uuid.UUID][]string{
		uid: {"good", "bad"},
	}}
	cli := &capturingClient{staleTokens: map[string]bool{"bad": true}}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := services.NewPushNotificationService(cli, repo, log)

	_, err := svc.NotifyUser(context.Background(), uid, services.PushPayload{Title: "x", Body: "y"})
	require.NoError(t, err)
	require.Equal(t, []string{"bad"}, repo.deleted)
}

func TestPush_NotifyUser_NoTokens_NoCall(t *testing.T) {
	repo := &fakeTokenRepo{tokens: map[uuid.UUID][]string{}}
	cli := &capturingClient{}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := services.NewPushNotificationService(cli, repo, log)

	results, err := svc.NotifyUser(context.Background(), uuid.New(), services.PushPayload{Title: "x", Body: "y"})
	require.NoError(t, err)
	require.Nil(t, results)
	require.Len(t, cli.sent, 0)
}
```

- [ ] **Step 2: Write the email render test**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/internal/services/email_service_test.go` with exactly:
```go
package services_test

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/exnodes/hrm-api/internal/config"
	"github.com/exnodes/hrm-api/internal/services"
	"github.com/stretchr/testify/require"
)

func TestEmail_RenderInvite_SubstitutesValues(t *testing.T) {
	cfg := &config.Config{AppName: "TestApp", InviteTokenExpireHours: 48}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc, err := services.NewEmailService(cfg, log)
	require.NoError(t, err)

	html, text, err := svc.RenderInvite(services.InviteTemplateData{
		FirstName: "Alice", InviteURL: "https://x.test/accept?token=abc",
	})
	require.NoError(t, err)
	require.Contains(t, html, "TestApp")
	require.Contains(t, html, "Alice")
	require.Contains(t, html, "https://x.test/accept?token=abc")
	require.Contains(t, html, "48")
	require.Contains(t, text, "Alice")
	require.Contains(t, text, "https://x.test/accept?token=abc")
}

func TestEmail_Send_NoSMTP_IsNoop(t *testing.T) {
	cfg := &config.Config{AppName: "X", InviteTokenExpireHours: 1} // SMTPHost is empty
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc, err := services.NewEmailService(cfg, log)
	require.NoError(t, err)

	err = svc.Send(context.Background(), "a@b.test", "hi", "<p>html</p>", "plain")
	require.NoError(t, err)
	_ = strings.TrimSpace // keep import used in builds without strings refs
}
```

- [ ] **Step 3: Run all phase-9 tests**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go test ./internal/services -run 'TestInvite_|TestPush_|TestEmail_' -v
```
Expected: all tests pass; 10+ `--- PASS` lines.

- [ ] **Step 4: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add internal/services/push_notification_service_test.go internal/services/email_service_test.go && \
  git commit -m "test(push,email): fan-out, stale-pruning, no-SMTP, template render"
```

---

### Task 18: Update README with new endpoints + Mailpit instructions

**Files:**
- Modify: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/README.md`

- [ ] **Step 1: Append a Phase-9 section**

Open the README and append this exact section at the bottom:
```markdown

## Phase 9 — Email + Invite + Push Notification

### Local SMTP via Mailpit

```bash
docker run -d --name mailpit -p 1025:1025 -p 8025:8025 axllent/mailpit
```

Set in `.env`:

```
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USE_TLS=false
SMTP_FROM_EMAIL=no-reply@exnodes.local
FRONTEND_URL=http://localhost:3000
INVITE_TOKEN_EXPIRE_HOURS=72
```

The Mailpit UI is at http://localhost:8025 — every outbound email is captured there with no relay.

### Endpoints

| Method | Path                                  | Auth                  |
|--------|---------------------------------------|-----------------------|
| POST   | `/api/v1/invites`                     | `users:create`        |
| GET    | `/api/v1/invites`                     | `users:read`          |
| GET    | `/api/v1/invites/:id`                 | `users:read`          |
| POST   | `/api/v1/invites/:id/resend`          | `users:create`        |
| DELETE | `/api/v1/invites/:id`                 | `users:delete`        |
| POST   | `/api/v1/invites/accept`              | **PUBLIC**            |
| POST   | `/api/v1/notifications/test`          | `organization_settings:manage` |

### Push (FCM HTTP v1)

Set `FIREBASE_CREDENTIALS_PATH=/path/to/sa.json` and `FIREBASE_PROJECT_ID=your-project`.
Without these, the push client logs and no-ops — useful for local development.
```

- [ ] **Step 2: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add README.md && \
  git commit -m "docs(readme): document phase 9 endpoints + Mailpit setup"
```

---

### Task 19: End-to-end self-verification — invite happy path

**Files:**
- Working files only; no commits in steps 1–6.

- [ ] **Step 1: Boot Mailpit and apply migrations**

Run:
```bash
docker rm -f mailpit 2>/dev/null || true
docker run -d --name mailpit -p 1025:1025 -p 8025:8025 axllent/mailpit
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && make migrate-up
```
Expected: `mailpit` container running; `make migrate-up` finishes with no error and `migrate-version` prints `11`.

- [ ] **Step 2: Start the server (background) and log to a file**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  (make run > /tmp/phase09-server.log 2>&1 &) ; \
  sleep 3 && \
  grep -E 'listening|listening on' /tmp/phase09-server.log | head -1
```
Expected: a line containing `listening on :8080`. If absent, `tail /tmp/phase09-server.log` to debug.

- [ ] **Step 3: Login as the seeded super admin to grab a JWT**

Run (substitute env values from `.env` — the seeded super admin is created by migration 000010):
```bash
ADMIN_EMAIL=$(grep '^SEED_ADMIN_EMAIL=' /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/.env | cut -d= -f2)
ADMIN_PASS=$(grep '^SEED_ADMIN_PASSWORD=' /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/.env | cut -d= -f2)
LOGIN_JSON=$(curl -sS -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASS\"}")
echo "$LOGIN_JSON" | tee /tmp/phase09-login.json
ADMIN_TOKEN=$(echo "$LOGIN_JSON" | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["access_token"])')
echo "TOKEN_PREFIX=${ADMIN_TOKEN:0:20}…"
```
Expected: login JSON `{"success":true,...,"access_token":"eyJ..."}`. `TOKEN_PREFIX=eyJ...`.

- [ ] **Step 4: Discover a role id to attach to the invite**

Run:
```bash
ROLE_ID=$(curl -sS -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/roles | \
  python3 -c 'import sys,json;d=json.load(sys.stdin)["data"]["items"];print([r for r in d if r["name"]=="Employee"][0]["id"])')
echo "ROLE_ID=$ROLE_ID"
```
Expected: a printable UUID.

- [ ] **Step 5: Create an invite**

Run:
```bash
INVITE_JSON=$(curl -sS -X POST http://localhost:8080/api/v1/invites \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"newhire@example.com\",\"full_name\":\"Nguyen Test\",\"role_ids\":[\"$ROLE_ID\"]}")
echo "$INVITE_JSON" | tee /tmp/phase09-create.json
INVITE_ID=$(echo "$INVITE_JSON" | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["id"])')
echo "INVITE_ID=$INVITE_ID"
```
Expected: HTTP 201, JSON with `"status":"pending"` and a fresh `id`.

- [ ] **Step 6: Verify the email arrived in Mailpit**

Run:
```bash
sleep 2  # let the async dispatch goroutine finish
curl -sS 'http://localhost:8025/api/v1/messages?limit=5' | \
  python3 -c 'import sys,json;m=json.load(sys.stdin)["messages"]; \
print([{"to":x["To"],"subject":x["Subject"]} for x in m])'
```
Expected: at least one entry with `to` `newhire@example.com` and the subject mentioning "invited". Save the body for the token:
```bash
MSG_ID=$(curl -sS 'http://localhost:8025/api/v1/messages?limit=1' | \
  python3 -c 'import sys,json;print(json.load(sys.stdin)["messages"][0]["ID"])')
TOKEN=$(curl -sS "http://localhost:8025/api/v1/message/$MSG_ID" | \
  python3 -c 'import sys,json,re;b=json.load(sys.stdin)["Text"];m=re.search(r"token=([0-9a-f]+)",b);print(m.group(1) if m else "")')
echo "TOKEN=$TOKEN"
```
Expected: a 64-char hex token. If empty, also try the HTML body field (`HTML`).

- [ ] **Step 7: Accept the invite (public endpoint)**

Run:
```bash
curl -sS -X POST http://localhost:8080/api/v1/invites/accept \
  -H 'Content-Type: application/json' \
  -d "{\"token\":\"$TOKEN\",\"password\":\"NewPass123!\"}" | tee /tmp/phase09-accept.json
```
Expected: HTTP 201, body `{"success":true,"data":{"user_id":"…","email":"newhire@example.com"}}`.

- [ ] **Step 8: Login as the new user**

Run:
```bash
curl -sS -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"newhire@example.com","password":"NewPass123!"}' | tee /tmp/phase09-newhire-login.json
```
Expected: `"success":true` with an `access_token`.

- [ ] **Step 9: Confirm the invite is now `accepted` in the list**

Run:
```bash
curl -sS -H "Authorization: Bearer $ADMIN_TOKEN" \
  'http://localhost:8080/api/v1/invites?email=newhire@example.com' | \
  python3 -m json.tool | tee /tmp/phase09-list.json | head -30
```
Expected: the matching item has `"status":"accepted"` and a non-null `accepted_user_id`.

---

### Task 20: End-to-end self-verification — error paths

- [ ] **Step 1: Re-accept the same token (expect 400)**

Run:
```bash
curl -sS -o /tmp/phase09-accept-again.json -w '%{http_code}\n' \
  -X POST http://localhost:8080/api/v1/invites/accept \
  -H 'Content-Type: application/json' \
  -d "{\"token\":\"$TOKEN\",\"password\":\"NewPass123!\"}"
cat /tmp/phase09-accept-again.json
```
Expected: HTTP `400` and a body containing `already accepted`.

- [ ] **Step 2: Accept an invalid token (expect 400)**

Run:
```bash
curl -sS -o /tmp/phase09-invalid.json -w '%{http_code}\n' \
  -X POST http://localhost:8080/api/v1/invites/accept \
  -H 'Content-Type: application/json' \
  -d '{"token":"deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef","password":"abcdefgh"}'
cat /tmp/phase09-invalid.json
```
Expected: HTTP `400` and body containing `invalid or expired`.

- [ ] **Step 3: Force expire a fresh invite + accept (expect 400)**

Run:
```bash
EXP_INVITE=$(curl -sS -X POST http://localhost:8080/api/v1/invites \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"email\":\"exp@example.com\",\"role_ids\":[\"$ROLE_ID\"]}" | \
  python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["id"])')

psql "$DATABASE_URL" -c "UPDATE invites SET expires_at = NOW() - INTERVAL '1 hour' WHERE id = '$EXP_INVITE';"

# get the (still valid) token from mailpit and try to accept it
sleep 2
EXP_MSG_ID=$(curl -sS 'http://localhost:8025/api/v1/messages?limit=5' | \
  python3 -c 'import sys,json;m=json.load(sys.stdin)["messages"]; \
print([x["ID"] for x in m if x["To"][0]["Address"]=="exp@example.com"][0])')
EXP_TOKEN=$(curl -sS "http://localhost:8025/api/v1/message/$EXP_MSG_ID" | \
  python3 -c 'import sys,json,re;b=json.load(sys.stdin)["Text"];m=re.search(r"token=([0-9a-f]+)",b);print(m.group(1))')

curl -sS -o /tmp/phase09-expired.json -w '%{http_code}\n' \
  -X POST http://localhost:8080/api/v1/invites/accept \
  -H 'Content-Type: application/json' \
  -d "{\"token\":\"$EXP_TOKEN\",\"password\":\"NewPass123!\"}"
cat /tmp/phase09-expired.json
```
Expected: HTTP `400` and body containing `expired`.

- [ ] **Step 4: Permission test — non-admin can't create invites**

Run (requires a seeded `employee@example.com` from Phase 1 seed, password `employee`; substitute if your seed differs):
```bash
EMP_TOKEN=$(curl -sS -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"employee@example.com","password":"employee"}' | \
  python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["access_token"])')

curl -sS -o /tmp/phase09-forbidden.json -w '%{http_code}\n' \
  -X POST http://localhost:8080/api/v1/invites \
  -H "Authorization: Bearer $EMP_TOKEN" -H 'Content-Type: application/json' \
  -d "{\"email\":\"x@y.com\",\"role_ids\":[\"$ROLE_ID\"]}"
cat /tmp/phase09-forbidden.json
```
Expected: HTTP `403`.

---

### Task 21: End-to-end self-verification — push test endpoint

- [ ] **Step 1: Register a fake device token for the admin (via Phase 2 endpoint)**

Run:
```bash
curl -sS -o /tmp/phase09-devtoken.json -w '%{http_code}\n' \
  -X POST http://localhost:8080/api/v1/users/me/device-tokens \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d '{"token":"fake-fcm-token-for-verification","platform":"web"}'
cat /tmp/phase09-devtoken.json
```
Expected: HTTP `201` (or `200` if Phase 2 returns that). Adjust path/payload to match Phase 2's actual endpoint if it differs — verify via Swagger first.

- [ ] **Step 2: POST a test push**

Run:
```bash
curl -sS -o /tmp/phase09-push.json -w '%{http_code}\n' \
  -X POST http://localhost:8080/api/v1/notifications/test \
  -H "Authorization: Bearer $ADMIN_TOKEN" -H 'Content-Type: application/json' \
  -d '{"title":"Phase 9 verification","body":"hello from the test endpoint","data":{"k":"v"}}'
cat /tmp/phase09-push.json
```
Expected behaviour depends on `FIREBASE_CREDENTIALS_PATH`:
- **Empty (default):** HTTP `200`, response shows `"attempted":1,"successes":1,"stale_pruned":0` and the server log line `push.send (disabled — noop)` includes the masked token. **This is the acceptable verification path.**
- **Configured with a real service account but a fake device token:** HTTP `200`, but the result's `Err` contains a 404 UNREGISTERED message and `stale_pruned:1` (the fake token will have been deleted). Also acceptable.

- [ ] **Step 3: Capture the relevant log line**

Run:
```bash
grep -E 'push\.send|push\.dispatch' /tmp/phase09-server.log | tail -10
```
Expected: at least one matching line.

- [ ] **Step 4: Stop the server and Mailpit**

Run:
```bash
pkill -f 'go run ./cmd/server' 2>/dev/null || pkill -f '/bin/server' 2>/dev/null; true
docker rm -f mailpit 2>/dev/null || true
sleep 1
lsof -i :8080 || echo 'STOPPED'
```
Expected: last line `STOPPED`.

---

### Task 22: Write the Phase-9 verification log and commit

**Files:**
- Create: `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/docs/superpowers/verification/phase-09.md`

- [ ] **Step 1: Compose the verification log**

Create `/Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2/docs/superpowers/verification/phase-09.md` with this content, **replacing every `<...>` placeholder** with the literal captured output from Tasks 19–21:
```markdown
# Phase 9 Verification Log

Date: 2026-05-15
Phase: 9 — Email + Invite + Push Notification
Spec: docs/superpowers/specs/2026-05-15-go-migration-design.md
Plan: docs/superpowers/plans/2026-05-15-phase-09-email-invite-push.md

## Environment

- Postgres: <psql -V output>
- Mailpit: docker run -p 1025:1025 -p 8025:8025 axllent/mailpit
- FIREBASE_CREDENTIALS_PATH: <empty | path>
- FIREBASE_PROJECT_ID: <empty | id>

## 1. Migrations

```
make migrate-up && make migrate-version
```
Output:
```
<paste>
```

## 2. Service tests

```
go test ./internal/services -run 'TestInvite_|TestPush_|TestEmail_' -v
```
Output (last 40 lines):
```
<paste>
```

## 3. Admin login + JWT

```
curl -X POST /api/v1/auth/login
```
Response (token redacted):
```
<paste sanitised /tmp/phase09-login.json>
```

## 4. Create invite

```
curl -X POST /api/v1/invites
```
Response:
```
<paste /tmp/phase09-create.json>
```

## 5. Mailpit confirmation

UI screenshot (or text dump): `curl http://localhost:8025/api/v1/messages?limit=5`
```
<paste>
```
Token extracted: `<first 8 chars>…<last 4>` (full token redacted).

## 6. Accept invite

```
curl -X POST /api/v1/invites/accept
```
Response:
```
<paste /tmp/phase09-accept.json>
```

## 7. New user login

```
curl -X POST /api/v1/auth/login (newhire@example.com)
```
Response (token redacted):
```
<paste /tmp/phase09-newhire-login.json>
```

## 8. Invite list — status now `accepted`

```
curl /api/v1/invites?email=newhire@example.com
```
Response (truncated):
```
<paste first 30 lines of /tmp/phase09-list.json>
```

## 9. Error paths

| Scenario             | HTTP | Body snippet              |
|----------------------|------|---------------------------|
| Re-accept same token | <code> | `<paste>`               |
| Invalid token        | <code> | `<paste>`               |
| Expired token        | <code> | `<paste>`               |
| Non-admin POST       | <code> | `<paste>`               |

## 10. Push test

Request: POST /api/v1/notifications/test
Response: `<paste /tmp/phase09-push.json>`
Server log line: `<paste from `grep push.send /tmp/phase09-server.log`>`

FCM mode: <noop | real>. Note: with FCM disabled the client logs and reports success; with FCM enabled, real device tokens are required for end-to-end delivery — we deliberately test with a fake token to exercise the stale-token pruning path. Both outcomes are acceptable per the plan.

## 11. Sign-off

All Phase 9 acceptance criteria met:

- [x] Migration 000011 applied + rollback verified
- [x] Invite model + repo + service + handler wired
- [x] Email service renders both templates + sends via SMTP (Mailpit captures)
- [x] PushClient interface + FCM HTTP v1 client + noop fallback
- [x] PushNotificationService dispatches and prunes stale tokens
- [x] 6 invite endpoints + public accept + admin test-push registered
- [x] Swagger lists all new endpoints
- [x] Service tests pass (`TestInvite_*`, `TestPush_*`, `TestEmail_*`)
- [x] Happy-path invite → email → accept → login flow works end-to-end
- [x] Error cases (expired, double-accept, invalid token, missing permission) return 400/403
- [x] Push test endpoint completes; stale tokens pruned (or noop logged) per FCM config
- [x] README updated with Phase 9 endpoints + Mailpit instructions
```

- [ ] **Step 2: Commit**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  git add docs/superpowers/verification/phase-09.md && \
  git commit -m "docs(verification): phase 9 e2e log — invite, email, push"
```

- [ ] **Step 3: Final sanity check**

Run:
```bash
cd /Users/sines/Documents/Work/exn-hrm-be/exnodes-hrm-api-go-v2 && \
  go build ./... && \
  go test ./internal/services -run 'TestInvite_|TestPush_|TestEmail_' -count=1 && \
  echo 'PHASE 9 OK'
```
Expected: last line `PHASE 9 OK`.

---

## Phase 9 Definition of Done

All boxes must be checked before Phase 9 is considered complete:

- [ ] `go build ./...` returns exit 0 with no warnings.
- [ ] `make migrate-up` succeeds on a clean DB and brings the version to `11`; `make migrate-down` reverses it cleanly.
- [ ] `psql -c "\d invites"` shows: `email citext NOT NULL`, `token TEXT NOT NULL`, `role_ids uuid[]`, four audit columns, partial unique index on `(email) WHERE accepted_at IS NULL AND is_deleted=FALSE`, and the `set_updated_at` trigger.
- [ ] `internal/models/invite.go` embeds `BaseModel`.
- [ ] Soft-delete: `InviteRepository.SoftDelete` sets both `is_deleted=true` and `deleted_at=NOW()`; list queries default to `NotDeleted`.
- [ ] PKs are UUID (verified by `\d invites`).
- [ ] All admin endpoints declare `middleware.RequirePerms(...)` inline at the route; `POST /api/v1/invites/accept` is registered on `v1` (no `authed`) under a comment block marking it PUBLIC.
- [ ] Every new handler has `@Summary` + `@Router` + `@Tags` Swagger annotations; Swagger UI lists 7 new endpoints under `invites` and `notifications` tags.
- [ ] Email send failures **do not** roll back invite creation; `invite.last_email_error` is populated; logs at `error` level capture the SMTP failure.
- [ ] Tokens are 64 hex chars from `crypto/rand.Read` (32 bytes).
- [ ] `POST /api/v1/invites/accept` documented in the router file with a comment that it is PUBLIC (no rate-limit shipped this phase — note: future hardening tracked separately; deliberately deferred).
- [ ] Service tests `go test ./internal/services -run 'TestInvite_|TestPush_|TestEmail_' -count=1` pass (10+ test cases).
- [ ] `make run` boots cleanly with Mailpit on `localhost:1025`; no SMTP credentials required.
- [ ] Mailpit at `http://localhost:8025` captures the invite email; the token from that email accepts successfully via the public endpoint; the new user can immediately log in.
- [ ] All four error cases (re-accept, invalid token, expired token, non-admin create) verified live and recorded in the verification log.
- [ ] `POST /api/v1/notifications/test` returns HTTP 200; server log contains a `push.send` line (noop mode acceptable when `FIREBASE_CREDENTIALS_PATH` is empty).
- [ ] `docs/superpowers/verification/phase-09.md` committed with every `<...>` placeholder replaced.
- [ ] README updated with Phase 9 endpoints + Mailpit + FCM env vars.
- [ ] Every task in this plan ended with a commit; `git log --oneline` shows ≥ 22 commits since the end of Phase 8.
