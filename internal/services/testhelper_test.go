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

// skipIfNoDB skips the test when TEST_DATABASE_URL is not set so that
// developers (and CI without a DB) can still run `go test ./...` cleanly.
func skipIfNoDB(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	return dsn
}

// TestMain bootstraps a real Postgres test DB, applies migrations, then
// hands control to the test binary. When TEST_DATABASE_URL is unset we just
// run the suite — every test that needs the DB will call skipIfNoDB.
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
	if err := testDB.Exec(`TRUNCATE TABLE device_tokens, user_notification_settings, employee_leave_quotas, dependents, employees, user_roles, users, roles RESTART IDENTITY CASCADE`).Error; err != nil {
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
