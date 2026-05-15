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
