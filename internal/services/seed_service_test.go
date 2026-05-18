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
