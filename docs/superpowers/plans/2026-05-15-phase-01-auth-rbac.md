# Phase 1: Auth + RBAC Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the Python FastAPI auth + RBAC stack to Go AND stand up the HR data shape (`employees`, `dependents`) that every subsequent phase will reference for `employee_id` ownership / HR fields. Stand up versioned SQL migrations for `roles`, `users`, `user_roles`, `employees`, `dependents`; ship JWT login/refresh/logout, password hashing, permission registry, JWT middleware, `RequirePerms` middleware, and an idempotent boot-time seed for system roles + super admin user + super admin employee row. End the phase with a `/api/v1/auth/login` + `/api/v1/roles/permissions` flow verified end-to-end against a running server.

**Architecture:** Layered (handlers → services → repositories → GORM/Postgres). The user/employee model is **split across 3 tables** (mirroring `exn-hr/backend`, NOT the Python source which uses one combined document):

- `users` — auth-only: `id`, `email`, `password_hash`, `is_active`, session-invalidation timestamps (`email_changed_at`, `password_reset_at`), audit cols. **No** `full_name`, **no** `department_id`, **no** `position_id`, **no** role-string column.
- `employees` — HR/personal info, 1-1 with `users` via `user_id` (UNIQUE). Holds `full_name`, `phone`, `personal_email`, `gender`, addresses, `dob`, nationality, ID card fields, `avatar_url`, education, marital status, emergency contact, work info (`department_id` / `position_id` / `manager_id`), contract & salary & banking. `department_id` / `position_id` are nullable UUIDs with **no FK constraint yet** — the constraints are added in Phase 3 when those tables exist. `manager_id` is a self-referential FK to `employees(id)`.
- `dependents` — 1-N with `employees` (child / parent / spouse / other) via `employee_id` with `ON DELETE CASCADE`.

Permissions are JSONB arrays on `roles`; `user_roles` is a join table with audit columns; permission gating is per-route via `middleware.RequirePerms(...)` inline at route declaration (mirroring FastAPI's `Depends(require_permissions(...))`). Wildcard `*` bypasses all checks. JWT is HS256 with `{sub, type, exp, iat}`; session-invalidation fields (`email_changed_at`, `password_reset_at`) are compared against `iat` on every request. Login responses embed the user's `employee` profile (full_name, avatar_url, department_id, position_id, manager_id) so the frontend never has to fetch `full_name` from `users`.

**Tech Stack:** Go 1.24, Gin, GORM (`gorm.io/driver/postgres`), Postgres (`citext`, `pgcrypto`, `uuid-ossp`), `golang-migrate/migrate/v4`, `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`, `swaggo/swag`, `testify`, real Postgres test DB.

**Module path:** `github.com/exnodes/hrm-api`

**Assumptions about Phase 0 (foundation already complete):**
- `go.mod` declares module `github.com/exnodes/hrm-api` and Go 1.24.
- `internal/config/config.go` loads env vars (DB DSN, `JWT_SECRET_KEY`, `JWT_ACCESS_TTL_MINUTES`, `JWT_REFRESH_TTL_DAYS`, `SUPER_ADMIN_EMAIL`, `SUPER_ADMIN_PASSWORD`, `SUPER_ADMIN_NAME`) via `godotenv`.
- `internal/config/db.go` exposes `Connect(cfg *Config) (*gorm.DB, error)`.
- `internal/models/base.go` provides `BaseModel` (UUID PK, audit cols) + `NotDeleted` scope.
- `internal/errors/errors.go` provides `AppError` + `ErrNotFound`, `ErrBadRequest`, `ErrConflict`, `ErrForbidden`, `ErrUnauthorized` constructors.
- `internal/middleware/error.go` converts `*AppError` to JSON `{success, message, code, details}`.
- `internal/dto/response.go` provides `Response[T]` envelope.
- Migration `000001_init_extensions.up.sql` creates `uuid-ossp`, `pgcrypto`, `citext` and a shared `set_updated_at()` PL/pgSQL function:
  ```sql
  CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
  BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
  $$ LANGUAGE plpgsql;
  ```
- `Makefile` has `migrate-up`, `migrate-down`, `migrate-new`, `migrate-force`, `run`, `test`, `swag`.
- `cmd/server/main.go` initializes config, DB, Gin engine, error middleware, Swagger handler, `/health`.
- `.env.example` exists; agent must add Phase 1 keys to it.

If any assumption is false, stop and reconcile with the spec (`docs/superpowers/specs/2026-05-15-go-migration-design.md`) before continuing.

**Non-negotiable rules (re-stated):**
1. Schema = versioned SQL only. No `db.AutoMigrate()`.
2. Audit columns + `set_updated_at` trigger + `is_deleted` index on every table.
3. Soft-delete via `NotDeleted` scope. Delete sets `is_deleted = TRUE, deleted_at = NOW()`.
4. UUID PKs via `gen_random_uuid()`.
5. Per-route permission via inline `middleware.RequirePerms(...)`. NOT group-level (the only group-level middleware is `JWT()`).
6. Swagger annotations on every new handler, with `@Security BearerAuth` on protected endpoints.
7. End-of-phase self verification (commands + responses logged to `docs/superpowers/verification/phase-01.md`).

**Commit cadence:** One commit per task; message style `feat(<scope>): <imperative summary>` (`feat(migration): add roles/users/user_roles schema`). Each commit must leave the tree compiling and the test suite green.

---

### Task 1: Migrations `000002_create_roles_users` + `000003_create_employees_dependents`

**Files:**
- Create: `migrations/000002_create_roles_users.up.sql`
- Create: `migrations/000002_create_roles_users.down.sql`
- Create: `migrations/000003_create_employees_dependents.up.sql`
- Create: `migrations/000003_create_employees_dependents.down.sql`

This task lands the full Phase 1 schema in two migrations (cleaner than one combined file): the auth tables first, then the HR-info tables that depend on `users`.

- [x] **Step 1.1: Write `000002_create_roles_users.up.sql`.**

  Create file `migrations/000002_create_roles_users.up.sql` with exactly this content:

  ```sql
  -- =========================================================================
  -- 000002_create_roles_users
  -- roles, users, user_roles + audit columns + triggers + indexes
  -- users is LEAN (auth only). HR fields live on the employees table created
  -- in 000003.
  -- =========================================================================

  -- ---------------- roles ----------------
  CREATE TABLE roles (
      id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      name            TEXT NOT NULL UNIQUE,
      description     TEXT NOT NULL DEFAULT '',
      is_system       BOOLEAN NOT NULL DEFAULT FALSE,
      permissions     JSONB NOT NULL DEFAULT '[]'::jsonb,
      created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      is_deleted      BOOLEAN NOT NULL DEFAULT FALSE,
      deleted_at      TIMESTAMPTZ NULL
  );
  CREATE INDEX idx_roles_is_deleted ON roles (is_deleted);
  CREATE INDEX idx_roles_name ON roles (name);
  CREATE TRIGGER trg_roles_set_updated_at
      BEFORE UPDATE ON roles
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();

  -- ---------------- users (AUTH ONLY) ----------------
  -- No full_name, no department_id, no position_id, no role string.
  -- HR / profile fields live on the employees table (000003).
  CREATE TABLE users (
      id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      email               CITEXT NOT NULL UNIQUE,
      password_hash       TEXT NOT NULL,
      is_active           BOOLEAN NOT NULL DEFAULT TRUE,
      email_changed_at    TIMESTAMPTZ NULL,
      password_reset_at   TIMESTAMPTZ NULL,
      created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      is_deleted          BOOLEAN NOT NULL DEFAULT FALSE,
      deleted_at          TIMESTAMPTZ NULL
  );
  CREATE INDEX idx_users_is_deleted ON users (is_deleted);
  CREATE TRIGGER trg_users_set_updated_at
      BEFORE UPDATE ON users
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();

  -- ---------------- user_roles ----------------
  CREATE TABLE user_roles (
      user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      role_id     UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      is_deleted  BOOLEAN NOT NULL DEFAULT FALSE,
      deleted_at  TIMESTAMPTZ NULL,
      PRIMARY KEY (user_id, role_id)
  );
  CREATE INDEX idx_user_roles_is_deleted ON user_roles (is_deleted);
  CREATE INDEX idx_user_roles_role_id ON user_roles (role_id);
  CREATE TRIGGER trg_user_roles_set_updated_at
      BEFORE UPDATE ON user_roles
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  ```

- [x] **Step 1.2: Write `000002_create_roles_users.down.sql`.**

  Create file `migrations/000002_create_roles_users.down.sql` with exactly this content:

  ```sql
  DROP TRIGGER IF EXISTS trg_user_roles_set_updated_at ON user_roles;
  DROP TABLE IF EXISTS user_roles;
  DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;
  DROP TABLE IF EXISTS users;
  DROP TRIGGER IF EXISTS trg_roles_set_updated_at ON roles;
  DROP TABLE IF EXISTS roles;
  ```

- [x] **Step 1.3: Write `000003_create_employees_dependents.up.sql`.**

  Create file `migrations/000003_create_employees_dependents.up.sql` with exactly this content:

  ```sql
  -- =========================================================================
  -- 000003_create_employees_dependents
  -- employees (1-1 with users), dependents (1-N with employees).
  --
  -- NOTE: department_id and position_id on employees are NULLABLE and
  -- intentionally have NO FK CONSTRAINT here. The departments / positions
  -- tables are introduced in Phase 3; the FK constraints are added in that
  -- phase (ALTER TABLE employees ADD CONSTRAINT ...).
  -- =========================================================================

  -- ---------------- employees ----------------
  CREATE TABLE employees (
      id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      user_id                     UUID NOT NULL UNIQUE REFERENCES users(id),

      -- Personal info
      full_name                   TEXT NOT NULL,
      phone                       TEXT NULL,
      personal_email              CITEXT NULL,
      gender                      TEXT NULL,                 -- male / female / other
      permanent_address           TEXT NULL,
      current_address             TEXT NULL,
      dob                         DATE NULL,
      nationality                 TEXT NULL,
      id_number                   TEXT NULL,
      id_issue_date               DATE NULL,
      id_front_image              TEXT NULL,
      id_back_image               TEXT NULL,
      avatar_url                  TEXT NULL,
      education                   TEXT NULL,                 -- high_school / college / university / master
      marital_status              TEXT NULL,                 -- single / married / other

      -- Emergency contact
      emergency_contact_name      TEXT NULL,
      emergency_contact_relation  TEXT NULL,
      emergency_contact_phone     TEXT NULL,

      -- Work info
      -- department_id / position_id: NO FK yet — added in Phase 3.
      department_id               UUID NULL,
      position_id                 UUID NULL,
      manager_id                  UUID NULL REFERENCES employees(id),
      join_date                   DATE NULL,
      contract_type               TEXT NOT NULL DEFAULT 'official',  -- probation / official
      contract_sign_date          DATE NULL,
      contract_end_date           DATE NULL,
      contract_renewal            INT  NOT NULL DEFAULT 1,

      -- Salary & insurance
      basic_salary                NUMERIC(18,2) NOT NULL DEFAULT 0,
      insurance_salary            NUMERIC(18,2) NOT NULL DEFAULT 0,

      -- Banking
      bank_account                TEXT NULL,
      bank_name                   TEXT NULL,
      bank_holder_name            TEXT NULL,
      payment_method              TEXT NOT NULL DEFAULT 'bank_transfer',  -- bank_transfer / cash

      created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      is_deleted                  BOOLEAN NOT NULL DEFAULT FALSE,
      deleted_at                  TIMESTAMPTZ NULL
  );
  CREATE INDEX idx_employees_is_deleted     ON employees (is_deleted);
  CREATE INDEX idx_employees_department_id  ON employees (department_id);
  CREATE INDEX idx_employees_position_id    ON employees (position_id);
  CREATE INDEX idx_employees_manager_id     ON employees (manager_id);
  -- (user_id already has a UNIQUE index by virtue of UNIQUE on the column)
  CREATE TRIGGER trg_employees_set_updated_at
      BEFORE UPDATE ON employees
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();

  -- ---------------- dependents ----------------
  CREATE TABLE dependents (
      id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
      full_name       TEXT NOT NULL,
      dob             DATE NULL,
      gender          TEXT NULL,           -- male / female / other
      relationship    TEXT NOT NULL,       -- child / parent / spouse / other
      created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      is_deleted      BOOLEAN NOT NULL DEFAULT FALSE,
      deleted_at      TIMESTAMPTZ NULL
  );
  CREATE INDEX idx_dependents_is_deleted  ON dependents (is_deleted);
  CREATE INDEX idx_dependents_employee_id ON dependents (employee_id);
  CREATE TRIGGER trg_dependents_set_updated_at
      BEFORE UPDATE ON dependents
      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  ```

- [x] **Step 1.4: Write `000003_create_employees_dependents.down.sql`.**

  Create file `migrations/000003_create_employees_dependents.down.sql` with exactly this content:

  ```sql
  DROP TRIGGER IF EXISTS trg_dependents_set_updated_at ON dependents;
  DROP TABLE IF EXISTS dependents;
  DROP TRIGGER IF EXISTS trg_employees_set_updated_at ON employees;
  DROP TABLE IF EXISTS employees;
  ```

- [x] **Step 1.5: Apply and roll back to confirm all four files work.**

  Run:
  ```
  make migrate-up
  make migrate-down
  make migrate-up
  ```

  Expected: each command prints the migration version it moved to and exits 0. After the final `migrate-up`, `psql $DATABASE_URL -c '\dt'` lists `roles`, `users`, `user_roles`, `employees`, `dependents`, plus the migrate metadata table.

  Spot-check column shape:
  ```
  psql $DATABASE_URL -c '\d employees' | head -40
  psql $DATABASE_URL -c '\d dependents'
  ```
  Expected: `employees` shows `full_name TEXT NOT NULL`, `department_id UUID NULL` (no FK), `manager_id UUID NULL` (FK to employees), `basic_salary NUMERIC(18,2) NOT NULL DEFAULT 0`, all 4 audit cols, and `trg_employees_set_updated_at` trigger. `dependents` shows `employee_id UUID NOT NULL` with `ON DELETE CASCADE` to employees.

- [x] **Step 1.6: Commit.**

  ```
  git add migrations/000002_create_roles_users.up.sql migrations/000002_create_roles_users.down.sql \
          migrations/000003_create_employees_dependents.up.sql migrations/000003_create_employees_dependents.down.sql
  git commit -m "feat(migration): add roles/users/user_roles and employees/dependents schema"
  ```

---

### Task 2: Models `role.go` + `user.go` + `employee.go` + `dependent.go`

**Files:**
- Create: `internal/models/role.go`
- Create: `internal/models/user.go`
- Create: `internal/models/employee.go`
- Create: `internal/models/dependent.go`

- [x] **Step 2.1: Write `internal/models/role.go`.**

  ```go
  package models

  import (
      "database/sql/driver"
      "encoding/json"
      "errors"
  )

  // StringSlice is a JSONB-backed []string for GORM.
  type StringSlice []string

  // Value implements driver.Valuer for GORM JSONB serialization.
  func (s StringSlice) Value() (driver.Value, error) {
      if s == nil {
          return []byte("[]"), nil
      }
      return json.Marshal(s)
  }

  // Scan implements sql.Scanner for GORM JSONB deserialization.
  func (s *StringSlice) Scan(src interface{}) error {
      if src == nil {
          *s = nil
          return nil
      }
      var bytes []byte
      switch v := src.(type) {
      case []byte:
          bytes = v
      case string:
          bytes = []byte(v)
      default:
          return errors.New("StringSlice: unsupported Scan source type")
      }
      if len(bytes) == 0 {
          *s = nil
          return nil
      }
      return json.Unmarshal(bytes, s)
  }

  // Role maps to the roles table.
  type Role struct {
      BaseModel
      Name        string      `gorm:"type:text;not null;uniqueIndex" json:"name"`
      Description string      `gorm:"type:text;not null;default:''" json:"description"`
      IsSystem    bool        `gorm:"not null;default:false" json:"is_system"`
      Permissions StringSlice `gorm:"type:jsonb;not null;default:'[]'::jsonb" json:"permissions"`
  }

  func (Role) TableName() string { return "roles" }
  ```

- [x] **Step 2.2: Write `internal/models/user.go` (AUTH ONLY — no full_name, no department_id, no position_id).**

  ```go
  package models

  import (
      "time"
  )

  // User maps to the users table. Auth-only — HR / profile fields live on
  // Employee (one-to-one via user_id).
  type User struct {
      BaseModel
      Email           string     `gorm:"type:citext;not null;uniqueIndex" json:"email"`
      PasswordHash    string     `gorm:"type:text;not null" json:"-"`
      IsActive        bool       `gorm:"not null;default:true" json:"is_active"`
      EmailChangedAt  *time.Time `json:"email_changed_at,omitempty"`
      PasswordResetAt *time.Time `json:"password_reset_at,omitempty"`

      // Many-to-many via user_roles join table.
      Roles []Role `gorm:"many2many:user_roles;joinForeignKey:user_id;joinReferences:role_id" json:"roles,omitempty"`

      // One-to-one with Employee — preloaded by handlers that need the HR
      // profile in the response shape (e.g., login). Pointer so an
      // un-preloaded User does not carry a zero-valued empty Employee.
      Employee *Employee `gorm:"foreignKey:UserID" json:"employee,omitempty"`
  }

  func (User) TableName() string { return "users" }
  ```

- [x] **Step 2.3: Write `internal/models/employee.go`.**

  ```go
  package models

  import (
      "time"

      "github.com/google/uuid"
  )

  // Employee maps to the employees table. 1-1 with User via UserID.
  type Employee struct {
      BaseModel
      UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`

      // Personal info
      FullName         string     `gorm:"type:text;not null" json:"full_name"`
      Phone            *string    `gorm:"type:text" json:"phone,omitempty"`
      PersonalEmail    *string    `gorm:"type:citext" json:"personal_email,omitempty"`
      Gender           *string    `gorm:"type:text" json:"gender,omitempty"`
      PermanentAddress *string    `gorm:"type:text" json:"permanent_address,omitempty"`
      CurrentAddress   *string    `gorm:"type:text" json:"current_address,omitempty"`
      DOB              *time.Time `gorm:"type:date" json:"dob,omitempty"`
      Nationality      *string    `gorm:"type:text" json:"nationality,omitempty"`
      IDNumber         *string    `gorm:"type:text" json:"id_number,omitempty"`
      IDIssueDate      *time.Time `gorm:"type:date" json:"id_issue_date,omitempty"`
      IDFrontImage     *string    `gorm:"type:text" json:"id_front_image,omitempty"`
      IDBackImage      *string    `gorm:"type:text" json:"id_back_image,omitempty"`
      AvatarURL        *string    `gorm:"type:text" json:"avatar_url,omitempty"`
      Education        *string    `gorm:"type:text" json:"education,omitempty"`
      MaritalStatus    *string    `gorm:"type:text" json:"marital_status,omitempty"`

      // Emergency contact
      EmergencyContactName     *string `gorm:"type:text" json:"emergency_contact_name,omitempty"`
      EmergencyContactRelation *string `gorm:"type:text" json:"emergency_contact_relation,omitempty"`
      EmergencyContactPhone    *string `gorm:"type:text" json:"emergency_contact_phone,omitempty"`

      // Work info — department_id / position_id have NO FK constraint until
      // Phase 3 introduces departments/positions tables.
      DepartmentID     *uuid.UUID `gorm:"type:uuid" json:"department_id,omitempty"`
      PositionID       *uuid.UUID `gorm:"type:uuid" json:"position_id,omitempty"`
      ManagerID        *uuid.UUID `gorm:"type:uuid" json:"manager_id,omitempty"`
      JoinDate         *time.Time `gorm:"type:date" json:"join_date,omitempty"`
      ContractType     string     `gorm:"type:text;not null;default:'official'" json:"contract_type"`
      ContractSignDate *time.Time `gorm:"type:date" json:"contract_sign_date,omitempty"`
      ContractEndDate  *time.Time `gorm:"type:date" json:"contract_end_date,omitempty"`
      ContractRenewal  int        `gorm:"not null;default:1" json:"contract_renewal"`

      // Salary & insurance
      BasicSalary     float64 `gorm:"type:numeric(18,2);not null;default:0" json:"basic_salary"`
      InsuranceSalary float64 `gorm:"type:numeric(18,2);not null;default:0" json:"insurance_salary"`

      // Banking
      BankAccount    *string `gorm:"type:text" json:"bank_account,omitempty"`
      BankName       *string `gorm:"type:text" json:"bank_name,omitempty"`
      BankHolderName *string `gorm:"type:text" json:"bank_holder_name,omitempty"`
      PaymentMethod  string  `gorm:"type:text;not null;default:'bank_transfer'" json:"payment_method"`

      // Relations
      User         *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
      Manager      *Employee    `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
      Subordinates []Employee   `gorm:"foreignKey:ManagerID" json:"subordinates,omitempty"`
      Dependents   []Dependent  `gorm:"foreignKey:EmployeeID" json:"dependents,omitempty"`
  }

  func (Employee) TableName() string { return "employees" }
  ```

- [x] **Step 2.4: Write `internal/models/dependent.go`.**

  ```go
  package models

  import (
      "time"

      "github.com/google/uuid"
  )

  // Dependent maps to the dependents table — people supported by an employee.
  type Dependent struct {
      BaseModel
      EmployeeID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"employee_id"`
      FullName     string     `gorm:"type:text;not null" json:"full_name"`
      DOB          *time.Time `gorm:"type:date" json:"dob,omitempty"`
      Gender       *string    `gorm:"type:text" json:"gender,omitempty"`        // male / female / other
      Relationship string     `gorm:"type:text;not null" json:"relationship"`   // child / parent / spouse / other
  }

  func (Dependent) TableName() string { return "dependents" }
  ```

- [x] **Step 2.5: Verify build.**

  ```
  go build ./...
  ```

  Expected: exit 0, no output.

- [x] **Step 2.6: Commit.**

  ```
  git add internal/models/role.go internal/models/user.go internal/models/employee.go internal/models/dependent.go
  git commit -m "feat(models): add Role, User (auth-only), Employee, Dependent GORM models"
  ```

---

### Task 3: Permission registry

**Files:**
- Create: `internal/permissions/registry.go`
- Create: `internal/permissions/registry_test.go`

- [x] **Step 3.1: Write `internal/permissions/registry.go`.**

  Mirror Python's `Permission` enum verbatim. Use a typed `Permission` (string alias) so callers get compile-time safety.

  ```go
  // Package permissions defines the centralized permission registry.
  // Mirrors app/core/permissions.py from the Python codebase.
  package permissions

  // Permission is a typed permission string.
  type Permission string

  const (
      PermAll Permission = "*"

      // Authentication
      PermAuthLogin Permission = "auth:login"

      // Users
      PermUsersRead         Permission = "users:read"
      PermUsersCreate       Permission = "users:create"
      PermUsersUpdate       Permission = "users:update"
      PermUsersDelete       Permission = "users:delete"
      PermUsersManageRoles  Permission = "users:manage_roles"
      PermUsersChangePwd    Permission = "users:change_password"

      // Roles
      PermRolesRead   Permission = "roles:read"
      PermRolesCreate Permission = "roles:create"
      PermRolesUpdate Permission = "roles:update"
      PermRolesDelete Permission = "roles:delete"

      // Departments
      PermDepartmentsRead   Permission = "departments:read"
      PermDepartmentsCreate Permission = "departments:create"
      PermDepartmentsUpdate Permission = "departments:update"
      PermDepartmentsDelete Permission = "departments:delete"

      // Positions
      PermPositionsRead   Permission = "positions:read"
      PermPositionsCreate Permission = "positions:create"
      PermPositionsUpdate Permission = "positions:update"
      PermPositionsDelete Permission = "positions:delete"

      // Skills
      PermSkillsRead   Permission = "skills:read"
      PermSkillsCreate Permission = "skills:create"
      PermSkillsUpdate Permission = "skills:update"
      PermSkillsDelete Permission = "skills:delete"

      // Leave Requests
      PermLeaveRead    Permission = "leave_requests:read"
      PermLeaveCreate  Permission = "leave_requests:create"
      PermLeaveUpdate  Permission = "leave_requests:update"
      PermLeaveDelete  Permission = "leave_requests:delete"
      PermLeaveApprove Permission = "leave_requests:approve"
      PermLeaveCancel  Permission = "leave_requests:cancel"
      PermLeaveManage  Permission = "leave_requests:manage"

      // Leave Quota
      PermLeaveQuotaManage Permission = "leave_quota:manage"

      // Attendance
      PermAttendanceRead   Permission = "attendance:read"
      PermAttendanceManage Permission = "attendance:manage_data"

      // Organization Settings
      PermOrgSettings Permission = "organization_settings:manage"

      // Announcements
      PermAnnounceManage Permission = "announcements:manage"
  )

  // AllPermissions returns the flat registry (for validation).
  func AllPermissions() []Permission {
      return []Permission{
          PermAuthLogin,
          PermUsersRead, PermUsersCreate, PermUsersUpdate, PermUsersDelete,
          PermUsersManageRoles, PermUsersChangePwd,
          PermRolesRead, PermRolesCreate, PermRolesUpdate, PermRolesDelete,
          PermDepartmentsRead, PermDepartmentsCreate, PermDepartmentsUpdate, PermDepartmentsDelete,
          PermPositionsRead, PermPositionsCreate, PermPositionsUpdate, PermPositionsDelete,
          PermSkillsRead, PermSkillsCreate, PermSkillsUpdate, PermSkillsDelete,
          PermLeaveRead, PermLeaveCreate, PermLeaveUpdate, PermLeaveDelete,
          PermLeaveApprove, PermLeaveCancel, PermLeaveManage,
          PermLeaveQuotaManage,
          PermAttendanceRead, PermAttendanceManage,
          PermOrgSettings,
          PermAnnounceManage,
      }
  }

  // IsValid returns true if p is a known permission (or the wildcard).
  func IsValid(p Permission) bool {
      if p == PermAll {
          return true
      }
      for _, known := range AllPermissions() {
          if known == p {
              return true
          }
      }
      return false
  }

  // PermissionItem describes a single permission in the picker.
  type PermissionItem struct {
      Key         Permission `json:"key"`
      Label       string     `json:"label"`
      Description string     `json:"description"`
  }

  // PermissionGroup is a category of related permissions, returned by
  // GET /api/v1/roles/permissions.
  type PermissionGroup struct {
      Resource    string           `json:"resource"`
      Label       string           `json:"label"`
      Permissions []PermissionItem `json:"permissions"`
  }

  // PermissionGroups is the structured catalog used by the FE permission picker.
  var PermissionGroups = []PermissionGroup{
      {
          Resource: "auth", Label: "Authentication",
          Permissions: []PermissionItem{
              {PermAuthLogin, "Login", "Sign in to the system"},
          },
      },
      {
          Resource: "users", Label: "Users",
          Permissions: []PermissionItem{
              {PermUsersRead, "View Users", "List and view user profiles"},
              {PermUsersCreate, "Create Users", "Create new user accounts"},
              {PermUsersUpdate, "Edit Users", "Update user profiles and settings"},
              {PermUsersDelete, "Activate / Deactivate Users", "Enable or disable user accounts"},
              {PermUsersManageRoles, "Manage User Roles", "Assign or remove roles from users"},
              {PermUsersChangePwd, "Change User Password", "Reset passwords for other users"},
          },
      },
      {
          Resource: "roles", Label: "Roles",
          Permissions: []PermissionItem{
              {PermRolesRead, "View Roles", "List and view role details"},
              {PermRolesCreate, "Create Roles", "Create new roles"},
              {PermRolesUpdate, "Edit Roles", "Update role name and permissions"},
              {PermRolesDelete, "Delete Roles", "Delete non-system roles"},
          },
      },
      {
          Resource: "departments", Label: "Departments",
          Permissions: []PermissionItem{
              {PermDepartmentsRead, "View Departments", "List and view departments"},
              {PermDepartmentsCreate, "Create Departments", "Create new departments"},
              {PermDepartmentsUpdate, "Edit Departments", "Rename departments"},
              {PermDepartmentsDelete, "Delete Departments", "Delete departments with no assigned employees"},
          },
      },
      {
          Resource: "positions", Label: "Positions",
          Permissions: []PermissionItem{
              {PermPositionsRead, "View Positions", "List and view positions"},
              {PermPositionsCreate, "Create Positions", "Create new positions"},
              {PermPositionsUpdate, "Edit Positions", "Rename positions"},
              {PermPositionsDelete, "Delete Positions", "Delete positions with no assigned employees"},
          },
      },
      {
          Resource: "skills", Label: "Skills",
          Permissions: []PermissionItem{
              {PermSkillsRead, "View Skills", "List and view skills"},
              {PermSkillsCreate, "Create Skills", "Create new skills"},
              {PermSkillsUpdate, "Edit Skills", "Update skill name, description, and icon"},
              {PermSkillsDelete, "Delete Skills", "Delete skills"},
          },
      },
      {
          Resource: "leave_requests", Label: "Leave Requests",
          Permissions: []PermissionItem{
              {PermLeaveRead, "View Leave Requests", "List and view leave requests"},
              {PermLeaveCreate, "Create Leave Requests", "Submit leave requests"},
              {PermLeaveUpdate, "Edit Leave Requests", "Update leave request details"},
              {PermLeaveDelete, "Delete Leave Requests", "Soft-delete leave requests"},
              {PermLeaveApprove, "Approve/Reject Leave Requests", "Approve or reject pending leave requests"},
              {PermLeaveCancel, "Cancel Leave Requests", "Cancel pending or approved leave requests"},
              {PermLeaveManage, "Manage Others' Leave Requests", "Create, edit, and view leave requests on behalf of other employees"},
          },
      },
      {
          Resource: "leave_quota", Label: "Leave Quota",
          Permissions: []PermissionItem{
              {PermLeaveQuotaManage, "Manage Leave Quota", "Change annual and sick leave quotas for employees"},
          },
      },
      {
          Resource: "attendance", Label: "Attendance",
          Permissions: []PermissionItem{
              {PermAttendanceRead, "View Attendance", "View the monthly attendance matrix"},
              {PermAttendanceManage, "Manage Attendance Data", "View all employees' attendance (without this, only own row is visible)"},
          },
      },
      {
          Resource: "organization_settings", Label: "Organization Settings",
          Permissions: []PermissionItem{
              {PermOrgSettings, "Manage Organization Settings", "View and update organization-wide settings such as late arrival threshold"},
          },
      },
      {
          Resource: "announcements", Label: "Announcements",
          Permissions: []PermissionItem{
              {PermAnnounceManage, "Manage Announcements", "Create, edit, send, and delete announcements"},
          },
      },
  }
  ```

- [x] **Step 3.2: Write `internal/permissions/registry_test.go` (TDD round-trip on the registry).**

  ```go
  package permissions

  import "testing"

  func TestIsValid(t *testing.T) {
      if !IsValid(PermAll) {
          t.Fatal("wildcard should be valid")
      }
      if !IsValid(PermUsersRead) {
          t.Fatal("users:read should be valid")
      }
      if IsValid(Permission("not:a:real:permission")) {
          t.Fatal("unknown permission should be invalid")
      }
  }

  func TestPermissionGroupsContainsAll(t *testing.T) {
      seen := map[Permission]bool{}
      for _, g := range PermissionGroups {
          for _, p := range g.Permissions {
              seen[p.Key] = true
          }
      }
      for _, p := range AllPermissions() {
          if !seen[p] {
              t.Errorf("permission %q is in AllPermissions but not in PermissionGroups", p)
          }
      }
  }
  ```

- [x] **Step 3.3: Run tests.**

  ```
  go test ./internal/permissions/...
  ```

  Expected: `ok  github.com/exnodes/hrm-api/internal/permissions ...`

- [x] **Step 3.4: Commit.**

  ```
  git add internal/permissions/
  git commit -m "feat(permissions): add permission constants and grouped catalog"
  ```

---

### Task 4: Password utility (bcrypt) — TDD

**Files:**
- Create: `pkg/utils/password.go`
- Create: `pkg/utils/password_test.go`

- [x] **Step 4.1: Write the test FIRST.**

  Create `pkg/utils/password_test.go`:

  ```go
  package utils

  import "testing"

  func TestHashAndVerifyPassword(t *testing.T) {
      hash, err := HashPassword("hunter2")
      if err != nil {
          t.Fatalf("hash err: %v", err)
      }
      if hash == "hunter2" {
          t.Fatal("hash should differ from plaintext")
      }
      if !CheckPassword("hunter2", hash) {
          t.Fatal("CheckPassword should accept the right password")
      }
      if CheckPassword("wrong", hash) {
          t.Fatal("CheckPassword should reject wrong password")
      }
  }

  func TestHashPasswordEmpty(t *testing.T) {
      if _, err := HashPassword(""); err == nil {
          t.Fatal("expected error for empty password")
      }
  }
  ```

- [x] **Step 4.2: Run the test — confirm it fails.**

  ```
  go test ./pkg/utils/...
  ```

  Expected: compile error / undefined `HashPassword`/`CheckPassword`.

- [x] **Step 4.3: Implement `pkg/utils/password.go`.**

  ```go
  package utils

  import (
      "errors"

      "golang.org/x/crypto/bcrypt"
  )

  // ErrEmptyPassword is returned when an empty password is provided.
  var ErrEmptyPassword = errors.New("password must not be empty")

  // HashPassword returns a bcrypt hash of the plaintext password using the
  // default cost (10).
  func HashPassword(plain string) (string, error) {
      if plain == "" {
          return "", ErrEmptyPassword
      }
      h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
      if err != nil {
          return "", err
      }
      return string(h), nil
  }

  // CheckPassword returns true when plain matches hashed.
  func CheckPassword(plain, hashed string) bool {
      if plain == "" || hashed == "" {
          return false
      }
      return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
  }
  ```

- [x] **Step 4.4: Confirm tests pass.**

  ```
  go test ./pkg/utils/...
  ```

  Expected: `ok  github.com/exnodes/hrm-api/pkg/utils ...`

- [x] **Step 4.5: Commit.**

  ```
  git add pkg/utils/password.go pkg/utils/password_test.go
  git commit -m "feat(utils): add bcrypt password hash and verify"
  ```

---

### Task 5: JWT utility — TDD

**Files:**
- Modify: `.env.example` (add `JWT_SECRET_KEY`, `JWT_ACCESS_TTL_MINUTES`, `JWT_REFRESH_TTL_DAYS`)
- Create: `pkg/utils/jwt.go`
- Create: `pkg/utils/jwt_test.go`

- [x] **Step 5.1: Add Phase 1 keys to `.env.example`.**

  Append:
  ```
  # Auth
  JWT_SECRET_KEY=change-me-please-32-bytes-minimum
  JWT_ACCESS_TTL_MINUTES=60
  JWT_REFRESH_TTL_DAYS=14

  # Seed
  SUPER_ADMIN_EMAIL=admin@exnodes.vn
  SUPER_ADMIN_PASSWORD=ChangeMe!2026
  SUPER_ADMIN_NAME=Super Admin
  ```

- [x] **Step 5.2: Write the test FIRST.**

  Create `pkg/utils/jwt_test.go`:

  ```go
  package utils

  import (
      "testing"
      "time"

      "github.com/google/uuid"
  )

  const testSecret = "test-secret-key-do-not-use-in-prod"

  func TestSignAndVerifyAccessToken(t *testing.T) {
      uid := uuid.New()
      tok, err := SignToken(uid.String(), TokenTypeAccess, testSecret, time.Minute)
      if err != nil {
          t.Fatalf("sign: %v", err)
      }
      claims, err := VerifyToken(tok, testSecret)
      if err != nil {
          t.Fatalf("verify: %v", err)
      }
      if claims.Subject != uid.String() {
          t.Errorf("sub mismatch: got %s", claims.Subject)
      }
      if claims.Type != TokenTypeAccess {
          t.Errorf("type mismatch: got %s", claims.Type)
      }
      if claims.IssuedAt == nil || claims.ExpiresAt == nil {
          t.Fatal("iat/exp must be set")
      }
  }

  func TestVerifyToken_ExpiredFails(t *testing.T) {
      tok, _ := SignToken("subj", TokenTypeAccess, testSecret, -time.Minute)
      if _, err := VerifyToken(tok, testSecret); err == nil {
          t.Fatal("expected error for expired token")
      }
  }

  func TestVerifyToken_BadSignatureFails(t *testing.T) {
      tok, _ := SignToken("subj", TokenTypeAccess, testSecret, time.Minute)
      if _, err := VerifyToken(tok, "wrong-secret"); err == nil {
          t.Fatal("expected error for bad signature")
      }
  }

  func TestVerifyToken_TamperedPayloadFails(t *testing.T) {
      tok, _ := SignToken("subj", TokenTypeAccess, testSecret, time.Minute)
      tampered := tok[:len(tok)-4] + "xxxx"
      if _, err := VerifyToken(tampered, testSecret); err == nil {
          t.Fatal("expected error for tampered token")
      }
  }
  ```

- [x] **Step 5.3: Run tests — confirm fail.**

  ```
  go test ./pkg/utils/...
  ```

  Expected: undefined `SignToken`, `VerifyToken`, `TokenTypeAccess`, `Claims`.

- [x] **Step 5.4: Implement `pkg/utils/jwt.go`.**

  ```go
  package utils

  import (
      "errors"
      "time"

      "github.com/golang-jwt/jwt/v5"
  )

  // TokenType is the value of the "type" claim.
  type TokenType string

  const (
      TokenTypeAccess  TokenType = "access"
      TokenTypeRefresh TokenType = "refresh"
  )

  // Claims is the JWT payload used by this service.
  type Claims struct {
      Type TokenType `json:"type"`
      jwt.RegisteredClaims
  }

  // SignToken issues a signed HS256 token with {sub, type, exp, iat}.
  // ttl may be negative — useful for tests of expired tokens.
  func SignToken(subject string, tokenType TokenType, secret string, ttl time.Duration) (string, error) {
      if subject == "" {
          return "", errors.New("subject must not be empty")
      }
      if secret == "" {
          return "", errors.New("secret must not be empty")
      }
      now := time.Now().UTC()
      claims := Claims{
          Type: tokenType,
          RegisteredClaims: jwt.RegisteredClaims{
              Subject:   subject,
              IssuedAt:  jwt.NewNumericDate(now),
              ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
          },
      }
      tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
      return tok.SignedString([]byte(secret))
  }

  // VerifyToken parses and validates a signed token. The token must be HS256
  // and have a non-expired exp claim.
  func VerifyToken(tokenString, secret string) (*Claims, error) {
      parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
      tok, err := parser.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
          return []byte(secret), nil
      })
      if err != nil {
          return nil, err
      }
      claims, ok := tok.Claims.(*Claims)
      if !ok || !tok.Valid {
          return nil, errors.New("invalid token")
      }
      return claims, nil
  }
  ```

- [x] **Step 5.5: Run tests.**

  ```
  go test ./pkg/utils/...
  ```

  Expected: all tests pass.

- [x] **Step 5.6: Commit.**

  ```
  git add pkg/utils/jwt.go pkg/utils/jwt_test.go .env.example
  git commit -m "feat(utils): add HS256 JWT sign/verify with access/refresh token types"
  ```

---

### Task 6: Role repository

**Files:**
- Create: `internal/repositories/role_repo.go`

- [x] **Step 6.1: Implement `internal/repositories/role_repo.go`.**

  ```go
  package repositories

  import (
      "context"
      "errors"

      "github.com/google/uuid"
      "gorm.io/gorm"

      "github.com/exnodes/hrm-api/internal/models"
  )

  // RoleRepository defines data access for roles.
  type RoleRepository interface {
      FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
      FindByName(ctx context.Context, name string) (*models.Role, error)
      FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Role, error)
      Create(ctx context.Context, role *models.Role) error
      Update(ctx context.Context, role *models.Role) error
  }

  type roleRepository struct{ db *gorm.DB }

  // NewRoleRepository constructs a Postgres-backed RoleRepository.
  func NewRoleRepository(db *gorm.DB) RoleRepository {
      return &roleRepository{db: db}
  }

  func notDeleted(db *gorm.DB) *gorm.DB {
      return db.Where("is_deleted = ?", false)
  }

  func (r *roleRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
      var role models.Role
      err := r.db.WithContext(ctx).Scopes(notDeleted).First(&role, "id = ?", id).Error
      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return nil, gorm.ErrRecordNotFound
          }
          return nil, err
      }
      return &role, nil
  }

  func (r *roleRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
      var role models.Role
      err := r.db.WithContext(ctx).Scopes(notDeleted).First(&role, "name = ?", name).Error
      if err != nil {
          return nil, err
      }
      return &role, nil
  }

  func (r *roleRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Role, error) {
      if len(ids) == 0 {
          return []models.Role{}, nil
      }
      var roles []models.Role
      err := r.db.WithContext(ctx).Scopes(notDeleted).Where("id IN ?", ids).Find(&roles).Error
      return roles, err
  }

  func (r *roleRepository) Create(ctx context.Context, role *models.Role) error {
      return r.db.WithContext(ctx).Create(role).Error
  }

  func (r *roleRepository) Update(ctx context.Context, role *models.Role) error {
      return r.db.WithContext(ctx).Save(role).Error
  }
  ```

- [x] **Step 6.2: Build.**

  ```
  go build ./...
  ```

  Expected: exit 0.

- [x] **Step 6.3: Commit.**

  ```
  git add internal/repositories/role_repo.go
  git commit -m "feat(repo): add role repository with NotDeleted scope"
  ```

---

### Task 7: User repository

**Files:**
- Create: `internal/repositories/user_repo.go`

- [x] **Step 7.1: Implement `internal/repositories/user_repo.go`.**

  ```go
  package repositories

  import (
      "context"
      "errors"
      "time"

      "github.com/google/uuid"
      "gorm.io/gorm"

      "github.com/exnodes/hrm-api/internal/models"
  )

  // UserRepository defines data access for users.
  type UserRepository interface {
      FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
      FindByIDWithRoles(ctx context.Context, id uuid.UUID) (*models.User, error)
      FindByIDWithRolesAndEmployee(ctx context.Context, id uuid.UUID) (*models.User, error)
      FindByEmail(ctx context.Context, email string) (*models.User, error)
      FindByEmailWithRoles(ctx context.Context, email string) (*models.User, error)
      FindByEmailWithRolesAndEmployee(ctx context.Context, email string) (*models.User, error)
      Create(ctx context.Context, user *models.User) error
      Update(ctx context.Context, user *models.User) error
      SoftDelete(ctx context.Context, id uuid.UUID) error
      ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
  }

  type userRepository struct{ db *gorm.DB }

  // NewUserRepository constructs a Postgres-backed UserRepository.
  func NewUserRepository(db *gorm.DB) UserRepository {
      return &userRepository{db: db}
  }

  func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
      var u models.User
      err := r.db.WithContext(ctx).Scopes(notDeleted).First(&u, "id = ?", id).Error
      if err != nil {
          return nil, err
      }
      return &u, nil
  }

  func (r *userRepository) FindByIDWithRoles(ctx context.Context, id uuid.UUID) (*models.User, error) {
      var u models.User
      err := r.db.WithContext(ctx).
          Scopes(notDeleted).
          Preload("Roles", notDeleted).
          First(&u, "id = ?", id).Error
      if err != nil {
          return nil, err
      }
      return &u, nil
  }

  func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
      var u models.User
      err := r.db.WithContext(ctx).Scopes(notDeleted).First(&u, "email = ?", email).Error
      if err != nil {
          return nil, err
      }
      return &u, nil
  }

  func (r *userRepository) FindByEmailWithRoles(ctx context.Context, email string) (*models.User, error) {
      var u models.User
      err := r.db.WithContext(ctx).
          Scopes(notDeleted).
          Preload("Roles", notDeleted).
          First(&u, "email = ?", email).Error
      if err != nil {
          return nil, err
      }
      return &u, nil
  }

  func (r *userRepository) FindByIDWithRolesAndEmployee(ctx context.Context, id uuid.UUID) (*models.User, error) {
      var u models.User
      err := r.db.WithContext(ctx).
          Scopes(notDeleted).
          Preload("Roles", notDeleted).
          Preload("Employee", notDeleted).
          First(&u, "id = ?", id).Error
      if err != nil {
          return nil, err
      }
      return &u, nil
  }

  func (r *userRepository) FindByEmailWithRolesAndEmployee(ctx context.Context, email string) (*models.User, error) {
      var u models.User
      err := r.db.WithContext(ctx).
          Scopes(notDeleted).
          Preload("Roles", notDeleted).
          Preload("Employee", notDeleted).
          First(&u, "email = ?", email).Error
      if err != nil {
          return nil, err
      }
      return &u, nil
  }

  func (r *userRepository) Create(ctx context.Context, user *models.User) error {
      return r.db.WithContext(ctx).Create(user).Error
  }

  func (r *userRepository) Update(ctx context.Context, user *models.User) error {
      return r.db.WithContext(ctx).Save(user).Error
  }

  func (r *userRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
      now := time.Now().UTC()
      res := r.db.WithContext(ctx).Model(&models.User{}).
          Where("id = ? AND is_deleted = false", id).
          Updates(map[string]interface{}{"is_deleted": true, "deleted_at": now})
      if res.Error != nil {
          return res.Error
      }
      if res.RowsAffected == 0 {
          return errors.New("user not found or already deleted")
      }
      return nil
  }

  // ReplaceRoles atomically replaces the user's role set.
  func (r *userRepository) ReplaceRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
      return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
          // Hard delete is acceptable on the join table — soft delete on a
          // many2many is rarely useful and complicates GORM associations.
          if err := tx.Exec("DELETE FROM user_roles WHERE user_id = ?", userID).Error; err != nil {
              return err
          }
          if len(roleIDs) == 0 {
              return nil
          }
          rows := make([]map[string]interface{}, 0, len(roleIDs))
          for _, rid := range roleIDs {
              rows = append(rows, map[string]interface{}{
                  "user_id": userID,
                  "role_id": rid,
              })
          }
          return tx.Table("user_roles").Create(&rows).Error
      })
  }
  ```

- [x] **Step 7.2: Build.**

  ```
  go build ./...
  ```

  Expected: exit 0.

- [x] **Step 7.3: Commit.**

  ```
  git add internal/repositories/user_repo.go
  git commit -m "feat(repo): add user repository with role preloads and soft delete"
  ```

---

### Task 7b: Employee + Dependent repositories

**Files:**
- Create: `internal/repositories/employee_repo.go`
- Create: `internal/repositories/dependent_repo.go`

- [x] **Step 7b.1: Implement `internal/repositories/employee_repo.go`.**

  ```go
  package repositories

  import (
      "context"
      "errors"
      "time"

      "github.com/google/uuid"
      "gorm.io/gorm"

      "github.com/exnodes/hrm-api/internal/models"
  )

  // EmployeeRepository defines data access for the HR profile of a user.
  type EmployeeRepository interface {
      Create(ctx context.Context, e *models.Employee) error
      Update(ctx context.Context, e *models.Employee) error
      SoftDelete(ctx context.Context, id uuid.UUID) error
      FindByID(ctx context.Context, id uuid.UUID) (*models.Employee, error)
      FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Employee, error)
      FindByIDWithUser(ctx context.Context, id uuid.UUID) (*models.Employee, error)
  }

  type employeeRepository struct{ db *gorm.DB }

  // NewEmployeeRepository constructs a Postgres-backed EmployeeRepository.
  func NewEmployeeRepository(db *gorm.DB) EmployeeRepository {
      return &employeeRepository{db: db}
  }

  func (r *employeeRepository) Create(ctx context.Context, e *models.Employee) error {
      return r.db.WithContext(ctx).Create(e).Error
  }

  func (r *employeeRepository) Update(ctx context.Context, e *models.Employee) error {
      return r.db.WithContext(ctx).Save(e).Error
  }

  func (r *employeeRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
      now := time.Now().UTC()
      res := r.db.WithContext(ctx).Model(&models.Employee{}).
          Where("id = ? AND is_deleted = false", id).
          Updates(map[string]interface{}{"is_deleted": true, "deleted_at": now})
      if res.Error != nil {
          return res.Error
      }
      if res.RowsAffected == 0 {
          return errors.New("employee not found or already deleted")
      }
      return nil
  }

  func (r *employeeRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
      var e models.Employee
      err := r.db.WithContext(ctx).Scopes(notDeleted).First(&e, "id = ?", id).Error
      if err != nil {
          return nil, err
      }
      return &e, nil
  }

  func (r *employeeRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.Employee, error) {
      var e models.Employee
      err := r.db.WithContext(ctx).Scopes(notDeleted).First(&e, "user_id = ?", userID).Error
      if err != nil {
          return nil, err
      }
      return &e, nil
  }

  func (r *employeeRepository) FindByIDWithUser(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
      var e models.Employee
      err := r.db.WithContext(ctx).
          Scopes(notDeleted).
          Preload("User", notDeleted).
          First(&e, "id = ?", id).Error
      if err != nil {
          return nil, err
      }
      return &e, nil
  }
  ```

- [x] **Step 7b.2: Implement `internal/repositories/dependent_repo.go`.**

  ```go
  package repositories

  import (
      "context"
      "errors"
      "time"

      "github.com/google/uuid"
      "gorm.io/gorm"

      "github.com/exnodes/hrm-api/internal/models"
  )

  // DependentRepository defines data access for the dependents of an employee.
  type DependentRepository interface {
      Create(ctx context.Context, d *models.Dependent) error
      Update(ctx context.Context, d *models.Dependent) error
      SoftDelete(ctx context.Context, id uuid.UUID) error
      FindByID(ctx context.Context, id uuid.UUID) (*models.Dependent, error)
      ListByEmployee(ctx context.Context, employeeID uuid.UUID) ([]models.Dependent, error)
  }

  type dependentRepository struct{ db *gorm.DB }

  // NewDependentRepository constructs a Postgres-backed DependentRepository.
  func NewDependentRepository(db *gorm.DB) DependentRepository {
      return &dependentRepository{db: db}
  }

  func (r *dependentRepository) Create(ctx context.Context, d *models.Dependent) error {
      return r.db.WithContext(ctx).Create(d).Error
  }

  func (r *dependentRepository) Update(ctx context.Context, d *models.Dependent) error {
      return r.db.WithContext(ctx).Save(d).Error
  }

  func (r *dependentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
      now := time.Now().UTC()
      res := r.db.WithContext(ctx).Model(&models.Dependent{}).
          Where("id = ? AND is_deleted = false", id).
          Updates(map[string]interface{}{"is_deleted": true, "deleted_at": now})
      if res.Error != nil {
          return res.Error
      }
      if res.RowsAffected == 0 {
          return errors.New("dependent not found or already deleted")
      }
      return nil
  }

  func (r *dependentRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Dependent, error) {
      var d models.Dependent
      err := r.db.WithContext(ctx).Scopes(notDeleted).First(&d, "id = ?", id).Error
      if err != nil {
          return nil, err
      }
      return &d, nil
  }

  func (r *dependentRepository) ListByEmployee(ctx context.Context, employeeID uuid.UUID) ([]models.Dependent, error) {
      var out []models.Dependent
      err := r.db.WithContext(ctx).
          Scopes(notDeleted).
          Where("employee_id = ?", employeeID).
          Order("created_at ASC").
          Find(&out).Error
      return out, err
  }
  ```

- [x] **Step 7b.3: Build.**

  ```
  go build ./...
  ```

  Expected: exit 0.

- [x] **Step 7b.4: Commit.**

  ```
  git add internal/repositories/employee_repo.go internal/repositories/dependent_repo.go
  git commit -m "feat(repo): add employee and dependent repositories with NotDeleted scope"
  ```

---

### Task 8: Service test helper (real Postgres test DB)

**Files:**
- Modify: `Makefile` (add `test-db-up` target if missing — only add, do not break existing targets)
- Create: `internal/services/testhelper_test.go`

- [x] **Step 8.1: Confirm the test DB DSN convention.**

  The test suite uses env var `TEST_DATABASE_URL` (e.g., `postgres://postgres:postgres@localhost:5432/exnodes_hrm_test?sslmode=disable`). If unset, the helper skips with a clear message so CI without a DB still passes vetting.

- [x] **Step 8.2: Write `internal/services/testhelper_test.go`.**

  ```go
  package services_test

  import (
      "context"
      "database/sql"
      "errors"
      "fmt"
      "os"
      "path/filepath"
      "runtime"
      "testing"
      "time"

      "github.com/golang-migrate/migrate/v4"
      _ "github.com/golang-migrate/migrate/v4/database/postgres"
      _ "github.com/golang-migrate/migrate/v4/source/file"
      "github.com/google/uuid"
      "gorm.io/driver/postgres"
      "gorm.io/gorm"

      "github.com/exnodes/hrm-api/internal/models"
      "github.com/exnodes/hrm-api/internal/permissions"
      "github.com/exnodes/hrm-api/internal/repositories"
      "github.com/exnodes/hrm-api/pkg/utils"
  )

  var (
      testDB           *gorm.DB
      testUserRepo     repositories.UserRepository
      testRoleRepo     repositories.RoleRepository
      testEmployeeRepo repositories.EmployeeRepository
  )

  // skipIfNoDB skips the test when TEST_DATABASE_URL is not set.
  func skipIfNoDB(t *testing.T) string {
      t.Helper()
      dsn := os.Getenv("TEST_DATABASE_URL")
      if dsn == "" {
          t.Skip("TEST_DATABASE_URL not set; skipping integration test")
      }
      return dsn
  }

  // TestMain bootstraps a real Postgres test DB, applies migrations, then
  // hands control to the test binary.
  func TestMain(m *testing.M) {
      dsn := os.Getenv("TEST_DATABASE_URL")
      if dsn == "" {
          // No DB: still let tests run (each test will skip itself).
          os.Exit(m.Run())
      }

      // Apply migrations from migrations/ relative to repo root.
      _, thisFile, _, _ := runtime.Caller(0)
      repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
      migDir := "file://" + filepath.Join(repoRoot, "migrations")

      sqlDB, err := sql.Open("postgres", dsn)
      if err != nil {
          fmt.Fprintf(os.Stderr, "sql.Open: %v\n", err)
          os.Exit(2)
      }
      sqlDB.SetConnMaxLifetime(time.Minute)

      mg, err := migrate.New(migDir, dsn)
      if err != nil {
          fmt.Fprintf(os.Stderr, "migrate.New: %v\n", err)
          os.Exit(2)
      }
      // Reset to a clean state.
      _ = mg.Drop()
      mg2, err := migrate.New(migDir, dsn)
      if err != nil {
          fmt.Fprintf(os.Stderr, "migrate.New(2): %v\n", err)
          os.Exit(2)
      }
      if err := mg2.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
          fmt.Fprintf(os.Stderr, "migrate.Up: %v\n", err)
          os.Exit(2)
      }

      gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
      if err != nil {
          fmt.Fprintf(os.Stderr, "gorm.Open: %v\n", err)
          os.Exit(2)
      }
      testDB = gdb
      testUserRepo = repositories.NewUserRepository(gdb)
      testRoleRepo = repositories.NewRoleRepository(gdb)
      testEmployeeRepo = repositories.NewEmployeeRepository(gdb)

      os.Exit(m.Run())
  }

  // truncateAll wipes the tables touched by Phase 1 tests.
  func truncateAll(t *testing.T) {
      t.Helper()
      if testDB == nil {
          return
      }
      // Order matters because of FK constraints; CASCADE covers the rest.
      if err := testDB.Exec(`TRUNCATE TABLE dependents, employees, user_roles, users, roles RESTART IDENTITY CASCADE`).Error; err != nil {
          t.Fatalf("truncate: %v", err)
      }
  }

  // makeRole inserts a role and returns it.
  func makeRole(t *testing.T, name string, perms []permissions.Permission, isSystem bool) *models.Role {
      t.Helper()
      ss := make(models.StringSlice, 0, len(perms))
      for _, p := range perms {
          ss = append(ss, string(p))
      }
      r := &models.Role{
          Name:        name,
          Description: name + " role",
          IsSystem:    isSystem,
          Permissions: ss,
      }
      if err := testRoleRepo.Create(context.Background(), r); err != nil {
          t.Fatalf("create role: %v", err)
      }
      return r
  }

  // makeUser inserts an auth-only user, optionally assigning roles, and returns it.
  func makeUser(t *testing.T, email, password string, roles ...*models.Role) *models.User {
      t.Helper()
      hash, err := utils.HashPassword(password)
      if err != nil {
          t.Fatalf("hash: %v", err)
      }
      u := &models.User{
          Email:        email,
          PasswordHash: hash,
          IsActive:     true,
      }
      if err := testUserRepo.Create(context.Background(), u); err != nil {
          t.Fatalf("create user: %v", err)
      }
      if len(roles) > 0 {
          ids := make([]uuid.UUID, 0, len(roles))
          for _, r := range roles {
              ids = append(ids, r.ID)
          }
          if err := testUserRepo.ReplaceRoles(context.Background(), u.ID, ids); err != nil {
              t.Fatalf("assign roles: %v", err)
          }
      }
      return u
  }

  // makeEmployee inserts an Employee row linked to the given user, with
  // sensible defaults. fullName falls back to the user's email when empty.
  func makeEmployee(t *testing.T, forUser *models.User, fullName string) *models.Employee {
      t.Helper()
      if fullName == "" {
          fullName = forUser.Email
      }
      e := &models.Employee{
          UserID:          forUser.ID,
          FullName:        fullName,
          ContractType:    "official",
          ContractRenewal: 1,
          PaymentMethod:   "bank_transfer",
      }
      if err := testEmployeeRepo.Create(context.Background(), e); err != nil {
          t.Fatalf("create employee: %v", err)
      }
      return e
  }
  ```

- [x] **Step 8.3: Validate the helper compiles.**

  ```
  go test -run x ./internal/services/...
  ```

  Expected: `ok` (no tests yet; -run x matches nothing).

- [x] **Step 8.4: Commit.**

  ```
  git add internal/services/testhelper_test.go
  git commit -m "test(services): add Postgres-backed test harness and factories"
  ```

---

### Task 9: Auth service — Login + Refresh + ResolveUserPermissions (TDD)

**Files:**
- Create: `internal/services/auth_service.go`
- Create: `internal/services/auth_service_test.go`

- [x] **Step 9.1: Write failing tests FIRST.**

  Create `internal/services/auth_service_test.go`:

  ```go
  package services_test

  import (
      "context"
      "testing"
      "time"

      "github.com/exnodes/hrm-api/internal/permissions"
      "github.com/exnodes/hrm-api/internal/services"
      "github.com/exnodes/hrm-api/pkg/utils"
  )

  const (
      jwtSecret  = "phase1-test-secret"
      accessTTL  = 15 * time.Minute
      refreshTTL = 7 * 24 * time.Hour
  )

  func newAuthSvc() *services.AuthService {
      return services.NewAuthService(testUserRepo, testRoleRepo, services.AuthConfig{
          JWTSecret:  jwtSecret,
          AccessTTL:  accessTTL,
          RefreshTTL: refreshTTL,
      })
  }

  func TestAuthService_Login_Success(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
      u := makeUser(t, "alice@test.com", "Secret123!", role)
      makeEmployee(t, u, "Alice Tester")

      svc := newAuthSvc()
      result, err := svc.Login(context.Background(), "alice@test.com", "Secret123!")
      if err != nil {
          t.Fatalf("login err: %v", err)
      }
      if result.Tokens.AccessToken == "" || result.Tokens.RefreshToken == "" {
          t.Fatal("expected non-empty tokens")
      }
      if result.User == nil || result.User.Employee == nil {
          t.Fatal("expected user with preloaded employee")
      }
      if result.User.Employee.FullName != "Alice Tester" {
          t.Errorf("employee.full_name: got %q", result.User.Employee.FullName)
      }
      claims, err := utils.VerifyToken(result.Tokens.AccessToken, jwtSecret)
      if err != nil {
          t.Fatalf("verify: %v", err)
      }
      if claims.Type != utils.TokenTypeAccess {
          t.Errorf("type: got %s", claims.Type)
      }
  }

  func TestAuthService_Login_WrongPassword(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
      makeUser(t, "alice@test.com", "Secret123!", role)

      svc := newAuthSvc()
      _, err := svc.Login(context.Background(), "alice@test.com", "wrong")
      if err == nil {
          t.Fatal("expected error")
      }
  }

  func TestAuthService_Login_UnknownEmail(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      svc := newAuthSvc()
      _, err := svc.Login(context.Background(), "ghost@test.com", "anything")
      if err == nil {
          t.Fatal("expected error")
      }
  }

  func TestAuthService_Login_InactiveAccount(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
      u := makeUser(t, "alice@test.com", "Secret123!", role)
      u.IsActive = false
      if err := testUserRepo.Update(context.Background(), u); err != nil {
          t.Fatalf("update: %v", err)
      }

      svc := newAuthSvc()
      _, err := svc.Login(context.Background(), "alice@test.com", "Secret123!")
      if err == nil {
          t.Fatal("expected error for inactive account")
      }
  }

  func TestAuthService_Login_MissingAuthLoginPermission(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      // Role with no permissions at all
      role := makeRole(t, "NoLogin", []permissions.Permission{}, false)
      makeUser(t, "alice@test.com", "Secret123!", role)

      svc := newAuthSvc()
      _, err := svc.Login(context.Background(), "alice@test.com", "Secret123!")
      if err == nil {
          t.Fatal("expected error when user lacks auth:login")
      }
  }

  func TestAuthService_Login_WildcardBypasses(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      role := makeRole(t, "Super Admin", []permissions.Permission{permissions.PermAll}, true)
      u := makeUser(t, "boss@test.com", "Secret123!", role)
      makeEmployee(t, u, "The Boss")

      svc := newAuthSvc()
      result, err := svc.Login(context.Background(), "boss@test.com", "Secret123!")
      if err != nil {
          t.Fatalf("login err: %v", err)
      }
      if result.Tokens.AccessToken == "" {
          t.Fatal("expected access token")
      }
  }

  func TestAuthService_Refresh_Success(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
      u := makeUser(t, "alice@test.com", "Secret123!", role)

      svc := newAuthSvc()
      refresh, err := utils.SignToken(u.ID.String(), utils.TokenTypeRefresh, jwtSecret, refreshTTL)
      if err != nil {
          t.Fatalf("sign refresh: %v", err)
      }
      result, err := svc.Refresh(context.Background(), refresh)
      if err != nil {
          t.Fatalf("refresh err: %v", err)
      }
      if result.Tokens.AccessToken == "" || result.Tokens.RefreshToken == "" {
          t.Fatal("expected token pair")
      }
  }

  func TestAuthService_Refresh_RejectsAccessToken(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      role := makeRole(t, "Employee", []permissions.Permission{permissions.PermAuthLogin}, true)
      u := makeUser(t, "alice@test.com", "Secret123!", role)

      svc := newAuthSvc()
      access, _ := utils.SignToken(u.ID.String(), utils.TokenTypeAccess, jwtSecret, accessTTL)
      if _, err := svc.Refresh(context.Background(), access); err == nil {
          t.Fatal("expected error: access token must not work as refresh")
      }
  }

  func TestAuthService_ResolveUserPermissions_UnionAcrossRoles(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      r1 := makeRole(t, "R1", []permissions.Permission{permissions.PermUsersRead}, false)
      r2 := makeRole(t, "R2", []permissions.Permission{permissions.PermRolesRead, permissions.PermUsersRead}, false)
      u := makeUser(t, "alice@test.com", "Secret123!", r1, r2)

      svc := newAuthSvc()
      perms, err := svc.ResolveUserPermissions(context.Background(), u.ID)
      if err != nil {
          t.Fatalf("resolve: %v", err)
      }
      if !perms[permissions.PermUsersRead] || !perms[permissions.PermRolesRead] {
          t.Fatalf("union missing: %v", perms)
      }
      if perms[permissions.PermAll] {
          t.Fatal("wildcard should not appear unless granted")
      }
  }

  func TestAuthService_ResolveUserPermissions_Wildcard(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      r := makeRole(t, "Super Admin", []permissions.Permission{permissions.PermAll}, true)
      u := makeUser(t, "boss@test.com", "Secret123!", r)

      svc := newAuthSvc()
      perms, _ := svc.ResolveUserPermissions(context.Background(), u.ID)
      if !perms[permissions.PermAll] {
          t.Fatal("expected wildcard permission")
      }
  }
  ```

- [x] **Step 9.2: Run tests — confirm fail.**

  ```
  go test ./internal/services/...
  ```

  Expected: compile errors for missing `services.NewAuthService`, `services.AuthConfig`, etc.

- [x] **Step 9.3: Implement `internal/services/auth_service.go`.**

  ```go
  package services

  import (
      "context"
      "errors"
      "time"

      "github.com/google/uuid"
      "gorm.io/gorm"

      apperr "github.com/exnodes/hrm-api/internal/errors"
      "github.com/exnodes/hrm-api/internal/models"
      "github.com/exnodes/hrm-api/internal/permissions"
      "github.com/exnodes/hrm-api/internal/repositories"
      "github.com/exnodes/hrm-api/pkg/utils"
  )

  // AuthConfig configures token TTLs and secret.
  type AuthConfig struct {
      JWTSecret  string
      AccessTTL  time.Duration
      RefreshTTL time.Duration
  }

  // TokenPair is the access+refresh result of Login/Refresh.
  type TokenPair struct {
      AccessToken  string
      RefreshToken string
      TokenType    string // always "Bearer"
  }

  // LoginResult bundles the token pair with the authenticated user (with
  // roles and employee profile preloaded) so handlers can render the full
  // login response shape without a second DB round-trip.
  type LoginResult struct {
      Tokens TokenPair
      User   *models.User // includes Roles and Employee
  }

  // AuthService handles login, refresh, and permission resolution.
  type AuthService struct {
      users repositories.UserRepository
      roles repositories.RoleRepository
      cfg   AuthConfig
  }

  // NewAuthService constructs an AuthService.
  func NewAuthService(users repositories.UserRepository, roles repositories.RoleRepository, cfg AuthConfig) *AuthService {
      return &AuthService{users: users, roles: roles, cfg: cfg}
  }

  // Login authenticates an email/password pair and returns a token pair plus
  // the authenticated user with Roles and Employee preloaded.
  func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
      user, err := s.users.FindByEmailWithRolesAndEmployee(ctx, email)
      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return nil, apperr.ErrUnauthorized("Invalid email or password")
          }
          return nil, err
      }
      if !user.IsActive {
          return nil, apperr.ErrUnauthorized("Your account has been deactivated. Contact your administrator.")
      }
      if user.PasswordHash == "" {
          return nil, apperr.ErrUnauthorized("Please set your password using the invite link sent to your email.")
      }
      if !utils.CheckPassword(password, user.PasswordHash) {
          return nil, apperr.ErrUnauthorized("Invalid email or password")
      }

      perms, err := s.resolvePermsFromUser(user.Roles)
      if err != nil {
          return nil, err
      }
      if !perms[permissions.PermAll] && !perms[permissions.PermAuthLogin] {
          return nil, apperr.ErrForbidden("You do not have permission to access this system.")
      }

      tokens, err := s.issueTokenPair(user.ID)
      if err != nil {
          return nil, err
      }
      return &LoginResult{Tokens: *tokens, User: user}, nil
  }

  // Refresh exchanges a refresh token for a new token pair (and returns the
  // refreshed User with Roles + Employee preloaded for the response shape).
  func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*LoginResult, error) {
      claims, err := utils.VerifyToken(refreshToken, s.cfg.JWTSecret)
      if err != nil {
          return nil, apperr.ErrUnauthorized("Invalid or expired refresh token")
      }
      if claims.Type != utils.TokenTypeRefresh {
          return nil, apperr.ErrBadRequest("Invalid token type")
      }
      uid, err := uuid.Parse(claims.Subject)
      if err != nil {
          return nil, apperr.ErrUnauthorized("Invalid token subject")
      }
      user, err := s.users.FindByIDWithRolesAndEmployee(ctx, uid)
      if err != nil {
          return nil, apperr.ErrUnauthorized("User not found or inactive")
      }
      if !user.IsActive {
          return nil, apperr.ErrUnauthorized("User not found or inactive")
      }
      tokens, err := s.issueTokenPair(user.ID)
      if err != nil {
          return nil, err
      }
      return &LoginResult{Tokens: *tokens, User: user}, nil
  }

  // Logout is stateless in Phase 1 — the client discards its tokens. Future
  // phases may add a server-side blacklist. The method exists so the handler
  // has a stable seam.
  func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
      _ = ctx
      _ = userID
      return nil
  }

  // ResolveUserPermissions returns the union of permissions across the user's
  // roles. Keys are the permission strings; presence == granted.
  func (s *AuthService) ResolveUserPermissions(ctx context.Context, userID uuid.UUID) (map[permissions.Permission]bool, error) {
      user, err := s.users.FindByIDWithRoles(ctx, userID)
      if err != nil {
          return nil, err
      }
      return s.resolvePermsFromUser(user.Roles)
  }

  func (s *AuthService) resolvePermsFromUser(roles []models.Role) (map[permissions.Permission]bool, error) {
      out := make(map[permissions.Permission]bool, 16)
      for _, r := range roles {
          for _, p := range r.Permissions {
              out[permissions.Permission(p)] = true
          }
      }
      return out, nil
  }

  func (s *AuthService) issueTokenPair(userID uuid.UUID) (*TokenPair, error) {
      access, err := utils.SignToken(userID.String(), utils.TokenTypeAccess, s.cfg.JWTSecret, s.cfg.AccessTTL)
      if err != nil {
          return nil, err
      }
      refresh, err := utils.SignToken(userID.String(), utils.TokenTypeRefresh, s.cfg.JWTSecret, s.cfg.RefreshTTL)
      if err != nil {
          return nil, err
      }
      return &TokenPair{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer"}, nil
  }
  ```

  Note: `errors` and `gorm.ErrRecordNotFound` are imported above only if referenced; the snippet uses `errors.Is(err, gorm.ErrRecordNotFound)` in `Login`, so both imports are required.

- [x] **Step 9.4: Run tests.**

  ```
  go test ./internal/services/...
  ```

  Expected: all 9 auth tests pass. If `TEST_DATABASE_URL` is unset they will all skip — confirm `go vet ./...` and `go build ./...` are clean as a fallback.

- [x] **Step 9.5: Commit.**

  ```
  git add internal/services/auth_service.go internal/services/auth_service_test.go
  git commit -m "feat(services): add auth service (login, refresh, resolve permissions)"
  ```

---

### Task 10: Seed service (idempotent) — TDD

**Files:**
- Create: `internal/services/seed_service.go`
- Create: `internal/services/seed_service_test.go`

- [x] **Step 10.1: Write failing tests FIRST.**

  Create `internal/services/seed_service_test.go`:

  ```go
  package services_test

  import (
      "context"
      "testing"

      "github.com/exnodes/hrm-api/internal/permissions"
      "github.com/exnodes/hrm-api/internal/services"
  )

  func newSeedSvc() *services.SeedService {
      return services.NewSeedService(testDB, testUserRepo, testRoleRepo, testEmployeeRepo, services.SeedConfig{
          SuperAdminEmail:    "admin@test.com",
          SuperAdminPassword: "ChangeMe!2026",
          SuperAdminName:     "Super Admin",
      })
  }

  func TestSeedService_FreshDatabase_CreatesSystemRolesAndAdmin(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      svc := newSeedSvc()

      if err := svc.Seed(context.Background()); err != nil {
          t.Fatalf("seed: %v", err)
      }

      // 5 system roles must exist.
      for _, name := range []string{"Super Admin", "Admin", "HR Manager", "Manager", "Employee"} {
          r, err := testRoleRepo.FindByName(context.Background(), name)
          if err != nil {
              t.Fatalf("expected role %s, got err %v", name, err)
          }
          if !r.IsSystem {
              t.Errorf("role %s should have is_system=true", name)
          }
      }

      // Super admin has wildcard.
      sa, err := testRoleRepo.FindByName(context.Background(), "Super Admin")
      if err != nil {
          t.Fatalf("super admin: %v", err)
      }
      var hasStar bool
      for _, p := range sa.Permissions {
          if permissions.Permission(p) == permissions.PermAll {
              hasStar = true
          }
      }
      if !hasStar {
          t.Fatal("Super Admin must have wildcard permission")
      }

      // Super admin user must exist and be linked.
      u, err := testUserRepo.FindByEmailWithRolesAndEmployee(context.Background(), "admin@test.com")
      if err != nil {
          t.Fatalf("admin user: %v", err)
      }
      foundSA := false
      for _, r := range u.Roles {
          if r.Name == "Super Admin" {
              foundSA = true
          }
      }
      if !foundSA {
          t.Fatal("admin user must be linked to Super Admin role")
      }
      if u.Employee == nil {
          t.Fatal("admin user must have a matching employee row")
      }
      if u.Employee.FullName != "Super Admin" {
          t.Errorf("employee.full_name: want %q, got %q", "Super Admin", u.Employee.FullName)
      }
      if u.Employee.ContractType != "official" {
          t.Errorf("employee.contract_type: want %q, got %q", "official", u.Employee.ContractType)
      }
  }

  func TestSeedService_RunTwice_Idempotent(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      svc := newSeedSvc()

      if err := svc.Seed(context.Background()); err != nil {
          t.Fatalf("seed 1: %v", err)
      }
      if err := svc.Seed(context.Background()); err != nil {
          t.Fatalf("seed 2: %v", err)
      }

      var roleCount int64
      testDB.Raw("SELECT COUNT(*) FROM roles WHERE is_deleted = false").Scan(&roleCount)
      if roleCount != 5 {
          t.Errorf("expected 5 roles after double-seed, got %d", roleCount)
      }
      var userCount int64
      testDB.Raw("SELECT COUNT(*) FROM users WHERE is_deleted = false").Scan(&userCount)
      if userCount != 1 {
          t.Errorf("expected 1 user after double-seed, got %d", userCount)
      }
      var empCount int64
      testDB.Raw("SELECT COUNT(*) FROM employees WHERE is_deleted = false").Scan(&empCount)
      if empCount != 1 {
          t.Errorf("expected 1 employee after double-seed, got %d", empCount)
      }
  }

  func TestSeedService_NoOverwriteOnExistingSuperAdminPassword(t *testing.T) {
      skipIfNoDB(t)
      truncateAll(t)
      svc := newSeedSvc()
      _ = svc.Seed(context.Background())

      // Capture hash, run seed again, hash must be unchanged.
      u1, _ := testUserRepo.FindByEmail(context.Background(), "admin@test.com")
      h1 := u1.PasswordHash
      _ = svc.Seed(context.Background())
      u2, _ := testUserRepo.FindByEmail(context.Background(), "admin@test.com")
      if u2.PasswordHash != h1 {
          t.Fatal("seed must not overwrite an existing super admin password")
      }
  }
  ```

- [x] **Step 10.2: Implement `internal/services/seed_service.go`.**

  ```go
  package services

  import (
      "context"
      "errors"
      "log"
      "time"

      "github.com/google/uuid"
      "gorm.io/gorm"

      "github.com/exnodes/hrm-api/internal/models"
      "github.com/exnodes/hrm-api/internal/permissions"
      "github.com/exnodes/hrm-api/internal/repositories"
      "github.com/exnodes/hrm-api/pkg/utils"
  )

  // SeedConfig configures the seed service.
  type SeedConfig struct {
      SuperAdminEmail    string
      SuperAdminPassword string
      SuperAdminName     string // default "Super Admin" if blank
  }

  // SeedService creates the 5 system roles, 1 super admin user, and the
  // matching employee row on boot. Safe to run repeatedly — operations are
  // merge/upsert and never overwrite manually-edited records.
  type SeedService struct {
      db        *gorm.DB
      users     repositories.UserRepository
      roles     repositories.RoleRepository
      employees repositories.EmployeeRepository
      cfg       SeedConfig
  }

  // NewSeedService constructs a SeedService.
  func NewSeedService(
      db *gorm.DB,
      users repositories.UserRepository,
      roles repositories.RoleRepository,
      employees repositories.EmployeeRepository,
      cfg SeedConfig,
  ) *SeedService {
      return &SeedService{db: db, users: users, roles: roles, employees: employees, cfg: cfg}
  }

  type roleSeed struct {
      Name        string
      Description string
      Permissions []permissions.Permission
  }

  func defaultRoles() []roleSeed {
      return []roleSeed{
          {
              Name:        "Super Admin",
              Description: "Full system access with all permissions",
              Permissions: []permissions.Permission{permissions.PermAll},
          },
          {
              Name:        "Admin",
              Description: "Administrative access for user and role management",
              Permissions: []permissions.Permission{
                  permissions.PermAuthLogin,
                  permissions.PermUsersRead, permissions.PermUsersCreate, permissions.PermUsersUpdate, permissions.PermUsersDelete,
                  permissions.PermUsersManageRoles, permissions.PermUsersChangePwd,
                  permissions.PermRolesRead, permissions.PermRolesCreate, permissions.PermRolesUpdate,
                  permissions.PermDepartmentsRead, permissions.PermDepartmentsCreate, permissions.PermDepartmentsUpdate, permissions.PermDepartmentsDelete,
                  permissions.PermPositionsRead, permissions.PermPositionsCreate, permissions.PermPositionsUpdate, permissions.PermPositionsDelete,
                  permissions.PermSkillsRead, permissions.PermSkillsCreate, permissions.PermSkillsUpdate, permissions.PermSkillsDelete,
                  permissions.PermLeaveRead, permissions.PermLeaveCreate, permissions.PermLeaveUpdate, permissions.PermLeaveDelete,
                  permissions.PermLeaveApprove, permissions.PermLeaveCancel, permissions.PermLeaveManage,
                  permissions.PermLeaveQuotaManage,
                  permissions.PermAttendanceRead, permissions.PermAttendanceManage,
                  permissions.PermOrgSettings,
              },
          },
          {
              Name:        "HR Manager",
              Description: "Human resources management access",
              Permissions: []permissions.Permission{
                  permissions.PermAuthLogin,
                  permissions.PermUsersRead, permissions.PermUsersCreate, permissions.PermUsersUpdate, permissions.PermUsersChangePwd,
                  permissions.PermRolesRead,
                  permissions.PermDepartmentsRead, permissions.PermDepartmentsCreate, permissions.PermDepartmentsUpdate,
                  permissions.PermPositionsRead, permissions.PermPositionsCreate, permissions.PermPositionsUpdate,
                  permissions.PermSkillsRead, permissions.PermSkillsCreate, permissions.PermSkillsUpdate,
                  permissions.PermLeaveRead, permissions.PermLeaveCreate, permissions.PermLeaveUpdate, permissions.PermLeaveDelete,
                  permissions.PermLeaveApprove, permissions.PermLeaveCancel, permissions.PermLeaveManage,
                  permissions.PermLeaveQuotaManage,
                  permissions.PermAttendanceRead, permissions.PermAttendanceManage,
                  permissions.PermOrgSettings,
              },
          },
          {
              Name:        "Manager",
              Description: "Team management access with user visibility",
              Permissions: []permissions.Permission{
                  permissions.PermAuthLogin,
                  permissions.PermUsersRead,
                  permissions.PermDepartmentsRead,
                  permissions.PermPositionsRead,
                  permissions.PermSkillsRead,
                  permissions.PermLeaveRead, permissions.PermLeaveCreate,
                  permissions.PermLeaveApprove, permissions.PermLeaveCancel, permissions.PermLeaveManage,
                  permissions.PermAttendanceRead, permissions.PermAttendanceManage,
              },
          },
          {
              Name:        "Employee",
              Description: "Basic employee access (own profile only)",
              Permissions: []permissions.Permission{
                  permissions.PermAuthLogin,
                  permissions.PermLeaveRead, permissions.PermLeaveCreate,
                  permissions.PermAttendanceRead,
              },
          },
      }
  }

  // Seed creates/updates the 5 system roles and the configured super admin.
  func (s *SeedService) Seed(ctx context.Context) error {
      if err := s.seedRoles(ctx); err != nil {
          return err
      }
      if err := s.seedSuperAdmin(ctx); err != nil {
          return err
      }
      return nil
  }

  func (s *SeedService) seedRoles(ctx context.Context) error {
      for _, rs := range defaultRoles() {
          existing, err := s.roles.FindByName(ctx, rs.Name)
          if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
              return err
          }
          desired := make(models.StringSlice, 0, len(rs.Permissions))
          for _, p := range rs.Permissions {
              desired = append(desired, string(p))
          }
          if existing == nil {
              r := &models.Role{
                  Name:        rs.Name,
                  Description: rs.Description,
                  IsSystem:    true,
                  Permissions: desired,
              }
              if err := s.roles.Create(ctx, r); err != nil {
                  return err
              }
              log.Printf("seed: created role %q", rs.Name)
              continue
          }
          // Merge: only add missing perms, never remove manually-added ones.
          if !existing.IsSystem {
              existing.IsSystem = true
          }
          present := map[string]bool{}
          for _, p := range existing.Permissions {
              present[p] = true
          }
          changed := false
          for _, p := range desired {
              if !present[p] {
                  existing.Permissions = append(existing.Permissions, p)
                  changed = true
              }
          }
          if changed {
              if err := s.roles.Update(ctx, existing); err != nil {
                  return err
              }
              log.Printf("seed: merged permissions into role %q", rs.Name)
          }
      }
      return nil
  }

  func (s *SeedService) seedSuperAdmin(ctx context.Context) error {
      if s.cfg.SuperAdminEmail == "" || s.cfg.SuperAdminPassword == "" {
          log.Printf("seed: SUPER_ADMIN_EMAIL/PASSWORD not set, skipping super admin user")
          return nil
      }
      saRole, err := s.roles.FindByName(ctx, "Super Admin")
      if err != nil {
          return err
      }

      adminName := s.cfg.SuperAdminName
      if adminName == "" {
          adminName = "Super Admin"
      }

      existing, err := s.users.FindByEmailWithRolesAndEmployee(ctx, s.cfg.SuperAdminEmail)
      if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
          return err
      }

      var userID uuid.UUID
      if existing != nil {
          userID = existing.ID

          // Ensure role linkage; never touch password.
          ids := []uuid.UUID{}
          hasSA := false
          for _, r := range existing.Roles {
              ids = append(ids, r.ID)
              if r.ID == saRole.ID {
                  hasSA = true
              }
          }
          if !hasSA {
              ids = append(ids, saRole.ID)
              if err := s.users.ReplaceRoles(ctx, existing.ID, ids); err != nil {
                  return err
              }
              log.Printf("seed: linked super admin role to existing user %q", existing.Email)
          }
      } else {
          hash, err := utils.HashPassword(s.cfg.SuperAdminPassword)
          if err != nil {
              return err
          }
          u := &models.User{
              Email:        s.cfg.SuperAdminEmail,
              PasswordHash: hash,
              IsActive:     true,
          }
          if err := s.users.Create(ctx, u); err != nil {
              return err
          }
          if err := s.users.ReplaceRoles(ctx, u.ID, []uuid.UUID{saRole.ID}); err != nil {
              return err
          }
          log.Printf("seed: created super admin user %q", u.Email)
          userID = u.ID
      }

      // Ensure the matching employee row exists (idempotent — never overwrite
      // a manually-edited employee record).
      _, err = s.employees.FindByUserID(ctx, userID)
      if err == nil {
          return nil
      }
      if !errors.Is(err, gorm.ErrRecordNotFound) {
          return err
      }
      today := time.Now().UTC()
      emp := &models.Employee{
          UserID:          userID,
          FullName:        adminName,
          ContractType:    "official",
          ContractRenewal: 1,
          PaymentMethod:   "bank_transfer",
          JoinDate:        &today,
      }
      if err := s.employees.Create(ctx, emp); err != nil {
          return err
      }
      log.Printf("seed: created super admin employee profile for %q", s.cfg.SuperAdminEmail)
      return nil
  }
  ```

- [x] **Step 10.3: Run tests.**

  ```
  go test ./internal/services/...
  ```

  Expected: all auth + seed tests pass (or skip if no DB).

- [x] **Step 10.4: Commit.**

  ```
  git add internal/services/seed_service.go internal/services/seed_service_test.go
  git commit -m "feat(services): add idempotent seed for system roles and super admin"
  ```

---

### Task 11: JWT middleware

**Files:**
- Create: `internal/middleware/auth.go`

- [x] **Step 11.1: Implement `internal/middleware/auth.go`.**

  ```go
  package middleware

  import (
      "context"
      "errors"
      "strings"
      "time"

      "github.com/gin-gonic/gin"
      "github.com/google/uuid"
      "gorm.io/gorm"

      apperr "github.com/exnodes/hrm-api/internal/errors"
      "github.com/exnodes/hrm-api/internal/models"
      "github.com/exnodes/hrm-api/internal/repositories"
      "github.com/exnodes/hrm-api/pkg/utils"
  )

  // Context keys for the authenticated user.
  const (
      ContextKeyUser    = "auth_user"
      ContextKeyUserID  = "auth_user_id"
      ContextKeyClaims  = "auth_claims"
  )

  // JWT returns Gin middleware that validates a Bearer access token and loads
  // the corresponding User from the database. Mirrors the Python
  // get_current_user dependency.
  func JWT(users repositories.UserRepository, jwtSecret string) gin.HandlerFunc {
      return func(c *gin.Context) {
          raw := extractToken(c)
          if raw == "" {
              _ = c.Error(apperr.ErrUnauthorized("Could not validate credentials"))
              c.Abort()
              return
          }
          claims, err := utils.VerifyToken(raw, jwtSecret)
          if err != nil {
              _ = c.Error(apperr.ErrUnauthorized("Could not validate credentials"))
              c.Abort()
              return
          }
          if claims.Type != utils.TokenTypeAccess {
              _ = c.Error(apperr.ErrUnauthorized("Invalid token type"))
              c.Abort()
              return
          }
          uid, err := uuid.Parse(claims.Subject)
          if err != nil {
              _ = c.Error(apperr.ErrUnauthorized("Invalid token payload"))
              c.Abort()
              return
          }

          user, err := users.FindByIDWithRoles(c.Request.Context(), uid)
          if err != nil {
              if errors.Is(err, gorm.ErrRecordNotFound) {
                  _ = c.Error(apperr.ErrUnauthorized("User not found"))
              } else {
                  _ = c.Error(apperr.ErrUnauthorized("Could not validate credentials"))
              }
              c.Abort()
              return
          }
          if !user.IsActive {
              _ = c.Error(apperr.ErrUnauthorized("User account is inactive"))
              c.Abort()
              return
          }

          // Session invalidation: reject tokens issued before credential changes.
          if claims.IssuedAt != nil {
              iat := claims.IssuedAt.Time
              if invalidatedBy(user.EmailChangedAt, iat) {
                  _ = c.Error(apperr.ErrUnauthorized("Session expired due to email change — please log in again"))
                  c.Abort()
                  return
              }
              if invalidatedBy(user.PasswordResetAt, iat) {
                  _ = c.Error(apperr.ErrUnauthorized("Session expired due to password reset — please log in again"))
                  c.Abort()
                  return
              }
          }

          c.Set(ContextKeyUser, user)
          c.Set(ContextKeyUserID, user.ID)
          c.Set(ContextKeyClaims, claims)
          c.Next()
      }
  }

  func extractToken(c *gin.Context) string {
      h := c.GetHeader("Authorization")
      if strings.HasPrefix(h, "Bearer ") {
          return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
      }
      return ""
  }

  func invalidatedBy(ts *time.Time, iat time.Time) bool {
      return ts != nil && iat.Before(ts.UTC())
  }

  // UserFromContext returns the authenticated user, or nil if not set.
  func UserFromContext(c *gin.Context) *models.User {
      v, ok := c.Get(ContextKeyUser)
      if !ok {
          return nil
      }
      u, _ := v.(*models.User)
      return u
  }

  // ContextWithUserID lifts the request context with the authenticated user ID.
  // Provided for services that don't take Gin context directly.
  func ContextWithUserID(c *gin.Context) context.Context {
      return c.Request.Context()
  }
  ```

- [x] **Step 11.2: Build.**

  ```
  go build ./...
  ```

  Expected: exit 0.

- [x] **Step 11.3: Commit.**

  ```
  git add internal/middleware/auth.go
  git commit -m "feat(middleware): add JWT middleware with session-invalidation checks"
  ```

---

### Task 12: `RequirePerms` middleware

**Files:**
- Create: `internal/middleware/permissions.go`

- [x] **Step 12.1: Implement `internal/middleware/permissions.go`.**

  ```go
  package middleware

  import (
      "github.com/gin-gonic/gin"

      apperr "github.com/exnodes/hrm-api/internal/errors"
      "github.com/exnodes/hrm-api/internal/permissions"
      "github.com/exnodes/hrm-api/internal/services"
  )

  // RequirePerms returns Gin middleware that gates the route on the user's
  // effective permission set (union of permissions across their roles).
  //
  // Semantics:
  //   - JWT middleware must run first (it sets the user on the context).
  //   - Wildcard "*" bypasses all checks.
  //   - ALL listed permissions must be present in the user's set.
  //   - On failure, returns ErrForbidden with details {required, missing}.
  func RequirePerms(authSvc *services.AuthService, required ...permissions.Permission) gin.HandlerFunc {
      return func(c *gin.Context) {
          user := UserFromContext(c)
          if user == nil {
              _ = c.Error(apperr.ErrUnauthorized("Could not validate credentials"))
              c.Abort()
              return
          }

          perms, err := authSvc.ResolveUserPermissions(c.Request.Context(), user.ID)
          if err != nil {
              _ = c.Error(apperr.ErrUnauthorized("Failed to resolve permissions"))
              c.Abort()
              return
          }

          if perms[permissions.PermAll] {
              c.Next()
              return
          }

          missing := make([]string, 0)
          for _, p := range required {
              if !perms[p] {
                  missing = append(missing, string(p))
              }
          }
          if len(missing) > 0 {
              reqList := make([]string, 0, len(required))
              for _, p := range required {
                  reqList = append(reqList, string(p))
              }
              err := apperr.ErrForbidden("Insufficient permissions")
              err.Details = map[string]any{
                  "required": reqList,
                  "missing":  missing,
              }
              _ = c.Error(err)
              c.Abort()
              return
          }
          c.Next()
      }
  }
  ```

- [x] **Step 12.2: Build.**

  ```
  go build ./...
  ```

  Expected: exit 0.

- [x] **Step 12.3: Commit.**

  ```
  git add internal/middleware/permissions.go
  git commit -m "feat(middleware): add RequirePerms with wildcard bypass and missing details"
  ```

---

### Task 13: Auth DTOs

**Files:**
- Create: `internal/dto/auth.go`

- [x] **Step 13.1: Implement `internal/dto/auth.go`.**

  ```go
  package dto

  import "github.com/google/uuid"

  // LoginRequest is the body for POST /api/v1/auth/login.
  type LoginRequest struct {
      Email    string `json:"email" binding:"required,email" example:"admin@exnodes.vn"`
      Password string `json:"password" binding:"required,min=1" example:"ChangeMe!2026"`
  }

  // RefreshRequest is the body for POST /api/v1/auth/refresh.
  type RefreshRequest struct {
      RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOi..."`
  }

  // EmployeeSummary is the slice of Employee fields embedded in auth responses.
  // The full Employee resource lives under /api/v1/employees.
  type EmployeeSummary struct {
      ID           uuid.UUID  `json:"id"`
      FullName     string     `json:"full_name"`
      AvatarURL    *string    `json:"avatar_url,omitempty"`
      DepartmentID *uuid.UUID `json:"department_id,omitempty"`
      PositionID   *uuid.UUID `json:"position_id,omitempty"`
      ManagerID    *uuid.UUID `json:"manager_id,omitempty"`
  }

  // RoleSummary is the slim role projection used in auth responses.
  type RoleSummary struct {
      ID          uuid.UUID `json:"id"`
      Name        string    `json:"name"`
      IsSystem    bool      `json:"is_system"`
      Permissions []string  `json:"permissions"`
  }

  // UserSummary is the user shape returned by the login/refresh endpoints. It
  // embeds the HR profile so the frontend never has to fetch full_name from
  // /users separately.
  type UserSummary struct {
      ID       uuid.UUID        `json:"id"`
      Email    string           `json:"email"`
      IsActive bool             `json:"is_active"`
      Employee *EmployeeSummary `json:"employee,omitempty"`
      Roles    []RoleSummary    `json:"roles"`
  }

  // LoginResponse is the body of a successful login or refresh.
  type LoginResponse struct {
      AccessToken  string      `json:"access_token"`
      RefreshToken string      `json:"refresh_token"`
      TokenType    string      `json:"token_type" example:"Bearer"`
      User         UserSummary `json:"user"`
  }

  // LogoutResponse is the body of a logout call (currently empty acknowledgement).
  type LogoutResponse struct {
      Message string `json:"message" example:"Logged out"`
  }
  ```

- [x] **Step 13.2: Commit.**

  ```
  git add internal/dto/auth.go
  git commit -m "feat(dto): add auth request/response shapes"
  ```

---

### Task 14: Auth handler

**Files:**
- Create: `internal/handlers/auth_handler.go`

- [x] **Step 14.1: Implement `internal/handlers/auth_handler.go`.**

  ```go
  package handlers

  import (
      "net/http"

      "github.com/gin-gonic/gin"

      apperr "github.com/exnodes/hrm-api/internal/errors"
      "github.com/exnodes/hrm-api/internal/dto"
      "github.com/exnodes/hrm-api/internal/middleware"
      "github.com/exnodes/hrm-api/internal/models"
      "github.com/exnodes/hrm-api/internal/services"
  )

  // AuthHandler handles authentication endpoints.
  type AuthHandler struct {
      auth *services.AuthService
  }

  // NewAuthHandler constructs an AuthHandler.
  func NewAuthHandler(auth *services.AuthService) *AuthHandler {
      return &AuthHandler{auth: auth}
  }

  // toUserSummary projects an auth-loaded User (with Roles + Employee preloaded)
  // into the auth response shape.
  func toUserSummary(u *models.User) dto.UserSummary {
      var emp *dto.EmployeeSummary
      if u.Employee != nil {
          emp = &dto.EmployeeSummary{
              ID:           u.Employee.ID,
              FullName:     u.Employee.FullName,
              AvatarURL:    u.Employee.AvatarURL,
              DepartmentID: u.Employee.DepartmentID,
              PositionID:   u.Employee.PositionID,
              ManagerID:    u.Employee.ManagerID,
          }
      }
      roles := make([]dto.RoleSummary, 0, len(u.Roles))
      for _, r := range u.Roles {
          perms := make([]string, 0, len(r.Permissions))
          for _, p := range r.Permissions {
              perms = append(perms, p)
          }
          roles = append(roles, dto.RoleSummary{
              ID:          r.ID,
              Name:        r.Name,
              IsSystem:    r.IsSystem,
              Permissions: perms,
          })
      }
      return dto.UserSummary{
          ID:       u.ID,
          Email:    u.Email,
          IsActive: u.IsActive,
          Employee: emp,
          Roles:    roles,
      }
  }

  // Login godoc
  // @Summary      Authenticate and receive access + refresh tokens
  // @Description  Exchanges an email + password for a token pair. Required permission: auth:login.
  // @Tags         Authentication
  // @Accept       json
  // @Produce      json
  // @Param        body  body      dto.LoginRequest  true  "Login credentials"
  // @Success      200   {object}  dto.Response[dto.LoginResponse]
  // @Failure      400   {object}  dto.Response[any]
  // @Failure      401   {object}  dto.Response[any]
  // @Failure      403   {object}  dto.Response[any]
  // @Router       /api/v1/auth/login [post]
  func (h *AuthHandler) Login(c *gin.Context) {
      var req dto.LoginRequest
      if err := c.ShouldBindJSON(&req); err != nil {
          _ = c.Error(apperr.ErrBadRequest(err.Error()))
          return
      }
      result, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[dto.LoginResponse]{
          Success: true,
          Message: "Login successful",
          Data: dto.LoginResponse{
              AccessToken:  result.Tokens.AccessToken,
              RefreshToken: result.Tokens.RefreshToken,
              TokenType:    "Bearer",
              User:         toUserSummary(result.User),
          },
      })
  }

  // Refresh godoc
  // @Summary      Exchange a refresh token for a new token pair
  // @Tags         Authentication
  // @Accept       json
  // @Produce      json
  // @Param        body  body      dto.RefreshRequest  true  "Refresh token"
  // @Success      200   {object}  dto.Response[dto.LoginResponse]
  // @Failure      400   {object}  dto.Response[any]
  // @Failure      401   {object}  dto.Response[any]
  // @Router       /api/v1/auth/refresh [post]
  func (h *AuthHandler) Refresh(c *gin.Context) {
      var req dto.RefreshRequest
      if err := c.ShouldBindJSON(&req); err != nil {
          _ = c.Error(apperr.ErrBadRequest(err.Error()))
          return
      }
      result, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
      if err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[dto.LoginResponse]{
          Success: true,
          Message: "Token refreshed",
          Data: dto.LoginResponse{
              AccessToken:  result.Tokens.AccessToken,
              RefreshToken: result.Tokens.RefreshToken,
              TokenType:    "Bearer",
              User:         toUserSummary(result.User),
          },
      })
  }

  // Logout godoc
  // @Summary      Acknowledge logout
  // @Description  Stateless logout — the client must discard its tokens.
  // @Tags         Authentication
  // @Produce      json
  // @Success      200   {object}  dto.Response[dto.LogoutResponse]
  // @Failure      401   {object}  dto.Response[any]
  // @Security     BearerAuth
  // @Router       /api/v1/auth/logout [post]
  func (h *AuthHandler) Logout(c *gin.Context) {
      user := middleware.UserFromContext(c)
      if user == nil {
          _ = c.Error(apperr.ErrUnauthorized("Could not validate credentials"))
          return
      }
      if err := h.auth.Logout(c.Request.Context(), user.ID); err != nil {
          _ = c.Error(err)
          return
      }
      c.JSON(http.StatusOK, dto.Response[dto.LogoutResponse]{
          Success: true,
          Message: "Logged out",
          Data:    dto.LogoutResponse{Message: "Logged out"},
      })
  }
  ```

- [x] **Step 14.2: Build.**

  ```
  go build ./...
  ```

- [x] **Step 14.3: Commit.**

  ```
  git add internal/handlers/auth_handler.go
  git commit -m "feat(handlers): add auth handler (login, refresh, logout) with swagger"
  ```

---

### Task 15: Role handler (permission catalog)

**Files:**
- Create: `internal/handlers/role_handler.go`

- [x] **Step 15.1: Implement `internal/handlers/role_handler.go`.**

  ```go
  package handlers

  import (
      "net/http"

      "github.com/gin-gonic/gin"

      "github.com/exnodes/hrm-api/internal/dto"
      "github.com/exnodes/hrm-api/internal/permissions"
  )

  // RoleHandler handles role-related endpoints. Phase 1 ships only the
  // permission catalog endpoint; full role CRUD comes in Phase 2.
  type RoleHandler struct{}

  // NewRoleHandler constructs a RoleHandler.
  func NewRoleHandler() *RoleHandler { return &RoleHandler{} }

  // ListPermissions godoc
  // @Summary      List grouped permissions (for the role-creation picker)
  // @Description  Returns the structured permission catalog. Requires authentication only (no specific permission).
  // @Tags         Roles
  // @Produce      json
  // @Success      200  {object}  dto.Response[[]permissions.PermissionGroup]
  // @Failure      401  {object}  dto.Response[any]
  // @Security     BearerAuth
  // @Router       /api/v1/roles/permissions [get]
  func (h *RoleHandler) ListPermissions(c *gin.Context) {
      c.JSON(http.StatusOK, dto.Response[[]permissions.PermissionGroup]{
          Success: true,
          Data:    permissions.PermissionGroups,
      })
  }
  ```

- [x] **Step 15.2: Commit.**

  ```
  git add internal/handlers/role_handler.go
  git commit -m "feat(handlers): add GET /api/v1/roles/permissions catalog endpoint"
  ```

---

### Task 16: Wire everything into `cmd/server/main.go`

**Files:**
- Modify: `cmd/server/main.go` (additive — preserve Phase 0 wiring)
- Modify: `internal/config/config.go` if it does not already expose Phase 1 fields

- [x] **Step 16.1: Ensure config has Phase 1 fields.**

  In `internal/config/config.go`, the `Config` struct must include:
  ```go
  JWTSecret              string
  JWTAccessTTLMinutes    int
  JWTRefreshTTLDays      int
  SuperAdminEmail        string
  SuperAdminPassword     string
  SuperAdminName         string
  ```
  Loaded via `os.Getenv` with the env keys `JWT_SECRET_KEY`, `JWT_ACCESS_TTL_MINUTES`, `JWT_REFRESH_TTL_DAYS`, `SUPER_ADMIN_EMAIL`, `SUPER_ADMIN_PASSWORD`, `SUPER_ADMIN_NAME`. Defaults: access 60 min, refresh 14 days, `SuperAdminName` defaults to `"Super Admin"` when blank. If `JWT_SECRET_KEY` is empty the app must `log.Fatal` on boot.

  Add these if missing (do not break existing keys).

- [x] **Step 16.2: Wire in `cmd/server/main.go`.**

  Inside `main()` after DB connect and before `router.Run(...)`, add (adapting variable names to match Phase 0 style):

  ```go
  // ---- repositories ----
  userRepo := repositories.NewUserRepository(db)
  roleRepo := repositories.NewRoleRepository(db)
  employeeRepo := repositories.NewEmployeeRepository(db)
  dependentRepo := repositories.NewDependentRepository(db)
  _ = dependentRepo // wired in later phases

  // ---- services ----
  authSvc := services.NewAuthService(userRepo, roleRepo, services.AuthConfig{
      JWTSecret:  cfg.JWTSecret,
      AccessTTL:  time.Duration(cfg.JWTAccessTTLMinutes) * time.Minute,
      RefreshTTL: time.Duration(cfg.JWTRefreshTTLDays) * 24 * time.Hour,
  })
  seedSvc := services.NewSeedService(db, userRepo, roleRepo, employeeRepo, services.SeedConfig{
      SuperAdminEmail:    cfg.SuperAdminEmail,
      SuperAdminPassword: cfg.SuperAdminPassword,
      SuperAdminName:     cfg.SuperAdminName,
  })

  // ---- run idempotent seed on boot ----
  if err := seedSvc.Seed(context.Background()); err != nil {
      log.Fatalf("seed: %v", err)
  }

  // ---- handlers ----
  authH := handlers.NewAuthHandler(authSvc)
  roleH := handlers.NewRoleHandler()

  // ---- routes ----
  v1 := router.Group("/api/v1")
  {
      // Public auth endpoints
      auth := v1.Group("/auth")
      auth.POST("/login", authH.Login)
      auth.POST("/refresh", authH.Refresh)

      // Protected endpoints
      authed := v1.Group("")
      authed.Use(middleware.JWT(userRepo, cfg.JWTSecret))

      authed.POST("/auth/logout", authH.Logout)
      authed.GET("/roles/permissions", roleH.ListPermissions)
  }
  ```

  Add the new imports:
  ```go
  "context"
  "time"

  "github.com/exnodes/hrm-api/internal/handlers"
  "github.com/exnodes/hrm-api/internal/middleware"
  "github.com/exnodes/hrm-api/internal/repositories"
  "github.com/exnodes/hrm-api/internal/services"
  ```

- [x] **Step 16.3: Regenerate Swagger.**

  ```
  make swag
  ```

  Expected: `docs/swagger.json`, `docs/swagger.yaml`, `docs/docs.go` are regenerated. Confirm `BearerAuth` `securityDefinitions` block is present (Phase 0 should have added it; if not, add to `cmd/server/main.go`):

  ```go
  // @securityDefinitions.apikey BearerAuth
  // @in header
  // @name Authorization
  // @description Type "Bearer {token}"
  ```

- [x] **Step 16.4: Verify build + tests.**

  ```
  go build ./...
  go vet ./...
  go test ./...
  ```

  Expected: all pass (service tests skip if `TEST_DATABASE_URL` is unset).

- [x] **Step 16.5: Commit.**

  ```
  git add cmd/server/main.go internal/config/config.go docs/
  git commit -m "feat(server): wire repos, services, middleware, routes, boot-time seed"
  ```

---

### Task 17: End-to-end verification + verification log

**Files:**
- Create: `docs/superpowers/verification/phase-01.md`

- [x] **Step 17.1: Create a clean database.**

  ```
  make migrate-down || true
  make migrate-up
  make migrate-version
  psql $DATABASE_URL -c '\dt'
  ```

  Expected: `migrate-version` prints `3` (000001 init, 000002 roles/users, 000003 employees/dependents). `\dt` shows `roles`, `users`, `user_roles`, `employees`, `dependents`. Spot-check:
  ```
  psql $DATABASE_URL -c '\d employees' | head -50
  ```
  Expected columns include `full_name`, `department_id`, `position_id`, `manager_id` (FK to employees), `basic_salary numeric(18,2)`, `contract_type`, plus 4 audit cols.

- [x] **Step 17.2: Run the server in the background.**

  Start the server (foreground in one terminal, or `make run &` in another). Confirm the seed log lines appear:
  - `seed: created role "Super Admin"`
  - `seed: created role "Admin"` ... etc
  - `seed: created super admin user "<SUPER_ADMIN_EMAIL>"`

  Then confirm:
  ```
  curl -s http://localhost:8080/health
  ```
  Expected: `{"success":true,"data":{"status":"ok"}}` (or whatever Phase 0 produced — non-empty 200).

- [x] **Step 17.3: Login as super admin.**

  ```
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
       -H 'Content-Type: application/json' \
       -d "{\"email\":\"${SUPER_ADMIN_EMAIL}\",\"password\":\"${SUPER_ADMIN_PASSWORD}\"}"
  ```

  Expected: `200` with body shape:
  ```json
  {
    "success": true,
    "message": "Login successful",
    "data": {
      "access_token": "eyJhbGciOi...",
      "refresh_token": "eyJhbGciOi...",
      "token_type": "Bearer",
      "user": {
        "id": "<uuid>",
        "email": "admin@exnodes.vn",
        "is_active": true,
        "employee": {
          "id": "<uuid>",
          "full_name": "Super Admin"
        },
        "roles": [
          { "id": "<uuid>", "name": "Super Admin", "is_system": true, "permissions": ["*"] }
        ]
      }
    }
  }
  ```

  Capture `access_token` as `ACCESS` and `refresh_token` as `REFRESH` env vars. Confirm `data.user.employee.full_name == "Super Admin"` (or whatever `SUPER_ADMIN_NAME` was set to).

- [x] **Step 17.4: Hit a protected endpoint with the token.**

  ```
  curl -s -H "Authorization: Bearer $ACCESS" http://localhost:8080/api/v1/roles/permissions | head -c 400
  ```

  Expected: `200` with `data` being an array of 11 `PermissionGroup` entries: `auth`, `users`, `roles`, `departments`, `positions`, `skills`, `leave_requests`, `leave_quota`, `attendance`, `organization_settings`, `announcements`. Count must match `len(permissions.PermissionGroups)`.

- [x] **Step 17.5: Hit the protected endpoint with NO token.**

  ```
  curl -s -o /dev/null -w '%{http_code}\n' http://localhost:8080/api/v1/roles/permissions
  ```

  Expected: `401`.

- [x] **Step 17.6: Hit with a tampered token.**

  ```
  curl -s -o /dev/null -w '%{http_code}\n' \
       -H "Authorization: Bearer ${ACCESS}xxxx" \
       http://localhost:8080/api/v1/roles/permissions
  ```

  Expected: `401`.

- [x] **Step 17.7: Wrong password.**

  ```
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
       -H 'Content-Type: application/json' \
       -d "{\"email\":\"${SUPER_ADMIN_EMAIL}\",\"password\":\"definitely-wrong\"}" \
       -o /dev/null -w '%{http_code}\n'
  ```

  Expected: `401`.

- [x] **Step 17.8: Login with a non-existent email.**

  ```
  curl -s -X POST http://localhost:8080/api/v1/auth/login \
       -H 'Content-Type: application/json' \
       -d '{"email":"ghost@nowhere.com","password":"anything"}' \
       -o /dev/null -w '%{http_code}\n'
  ```

  Expected: `401`.

- [x] **Step 17.9: Refresh flow.**

  ```
  curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
       -H 'Content-Type: application/json' \
       -d "{\"refresh_token\":\"${REFRESH}\"}"
  ```

  Expected: `200` with a brand-new pair of tokens.

  Then verify that an access token cannot be used as a refresh:
  ```
  curl -s -X POST http://localhost:8080/api/v1/auth/refresh \
       -H 'Content-Type: application/json' \
       -d "{\"refresh_token\":\"${ACCESS}\"}" \
       -o /dev/null -w '%{http_code}\n'
  ```
  Expected: `400` or `401`.

- [x] **Step 17.10: Logout.**

  ```
  curl -s -X POST http://localhost:8080/api/v1/auth/logout \
       -H "Authorization: Bearer $ACCESS" \
       -o /dev/null -w '%{http_code}\n'
  ```

  Expected: `200` (stateless — the token remains technically valid until expiry; documented as a Phase 1 limitation).

- [x] **Step 17.11: Confirm Swagger UI lists the new endpoints.**

  Visit `http://localhost:8080/swagger/index.html` in a browser (or `curl -s http://localhost:8080/swagger/doc.json | jq '.paths | keys'`). Expected paths include:
  - `/api/v1/auth/login`
  - `/api/v1/auth/refresh`
  - `/api/v1/auth/logout`
  - `/api/v1/roles/permissions`

- [x] **Step 17.12: Write the verification log.**

  Create `docs/superpowers/verification/phase-01.md` containing:
  - Date, phase name, agent name.
  - Output of `make migrate-version` (before and after).
  - The exact curl commands from steps 17.3–17.10 with the **redacted** tokens replaced by `<REDACTED>` and the response status codes / first ~10 lines of bodies.
  - The list of Swagger paths emitted in 17.11.
  - A short "Limitations" section noting that logout is currently stateless.

- [x] **Step 17.13: Final test pass.**

  ```
  go test ./...
  ```

  Expected: all green; any service tests that require a DB either pass or skip cleanly with no failures.

- [x] **Step 17.14: Commit the verification log.**

  ```
  git add docs/superpowers/verification/phase-01.md
  git commit -m "docs(verification): record Phase 1 end-to-end auth and RBAC walkthrough"
  ```

---

## Definition of Done (Phase 1)

All items below must be true before declaring the phase complete.

- [ ] `migrations/000002_create_roles_users.up.sql` + `.down.sql` AND `migrations/000003_create_employees_dependents.up.sql` + `.down.sql` exist, applied cleanly, and rollback restores a working DB.
- [ ] After `make migrate-up`, the DB contains tables `roles`, `users` (auth-only — no `full_name` / `department_id` / `position_id` column), `user_roles`, `employees` (full HR schema), `dependents`. All have 4 audit cols, `is_deleted` index, and a `set_updated_at` trigger. `employees.user_id` is UNIQUE; `employees.manager_id` is a self-ref FK; `dependents.employee_id` cascades on delete. `employees.department_id` and `employees.position_id` are nullable UUIDs with NO FK constraint (constraints added in Phase 3).
- [ ] `internal/models/role.go`, `internal/models/user.go` (auth-only), `internal/models/employee.go`, `internal/models/dependent.go` compile and embed `BaseModel`.
- [ ] `internal/models/user.go` does NOT declare a `FullName`, `DepartmentID`, or `PositionID` field — those live on `Employee`.
- [ ] `internal/permissions/registry.go` defines all 35 permissions from the Python source (1 auth + 6 users + 4 roles + 4 departments + 4 positions + 4 skills + 7 leave_requests + 1 leave_quota + 2 attendance + 1 organization_settings + 1 announcements) plus `PermAll` and the 11-group `PermissionGroups` catalog. Tests pass.
- [ ] `pkg/utils/password.go` and `pkg/utils/jwt.go` compile, tests pass (`go test ./pkg/utils/...`).
- [ ] `internal/repositories/role_repo.go`, `user_repo.go`, `employee_repo.go`, `dependent_repo.go` compile.
- [ ] `internal/services/auth_service.go` implements `Login`, `Refresh`, `Logout`, `ResolveUserPermissions`. `Login` and `Refresh` return `*LoginResult` (Tokens + User-with-Employee-preloaded). Tests pass (or skip if no DB).
- [ ] `internal/services/seed_service.go` is idempotent. Running twice does not duplicate rows. After seed, the DB contains exactly 5 roles, 1 user, AND 1 employee row whose `full_name` matches `SUPER_ADMIN_NAME` (default "Super Admin") and whose `user_id` equals the super admin user's ID. Tests pass.
- [ ] `internal/middleware/auth.go` validates Bearer tokens, loads user, checks `is_active` + `is_deleted` + session-invalidation timestamps.
- [ ] `internal/middleware/permissions.go` enforces `RequirePerms` with wildcard bypass and `{required, missing}` details on 403.
- [ ] `internal/dto/auth.go` defines `LoginRequest`, `LoginResponse` (with embedded `user.employee` summary), `RefreshRequest`, `LogoutResponse`, `UserSummary`, `EmployeeSummary`, `RoleSummary`.
- [ ] `internal/handlers/auth_handler.go` exposes login, refresh, logout — all with full Swagger annotations including `@Security BearerAuth` on logout. Login/refresh responses include `data.user.employee.full_name` populated from the seeded Employee row.
- [ ] `internal/handlers/role_handler.go` exposes `GET /api/v1/roles/permissions` — also annotated.
- [ ] `cmd/server/main.go` wires repos, services, middleware, routes, and runs the seed on boot.
- [ ] Swagger UI lists all 4 new endpoints.
- [ ] `go vet ./...` + `go test ./...` exit 0.
- [ ] `make migrate-up && make run` boots cleanly with no panic.
- [ ] All curl flows in Task 17 produce the expected status codes.
- [ ] `docs/superpowers/verification/phase-01.md` committed.
- [ ] One commit per task; every commit message starts with `feat(...)`/`test(...)`/`docs(...)`/`chore(...)` and describes the change.
