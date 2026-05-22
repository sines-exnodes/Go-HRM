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
	PermUsersRead        Permission = "users:read"
	PermUsersCreate      Permission = "users:create"
	PermUsersUpdate      Permission = "users:update"
	PermUsersDelete      Permission = "users:delete"
	PermUsersManageRoles Permission = "users:manage_roles"
	PermUsersChangePwd   Permission = "users:change_password"

	// Employees
	PermEmployeesRead   Permission = "employees:read"
	PermEmployeesCreate Permission = "employees:create"
	PermEmployeesUpdate Permission = "employees:update"
	PermEmployeesDelete Permission = "employees:delete"

	// Dependents
	PermDependentsManage Permission = "dependents:manage"

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

	// Invites (Phase 9)
	PermInviteManage Permission = "invites:manage"
)

// AllPermissions returns the flat registry (for validation).
func AllPermissions() []Permission {
	return []Permission{
		PermAuthLogin,
		PermUsersRead, PermUsersCreate, PermUsersUpdate, PermUsersDelete,
		PermUsersManageRoles, PermUsersChangePwd,
		PermEmployeesRead, PermEmployeesCreate, PermEmployeesUpdate, PermEmployeesDelete,
		PermDependentsManage,
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
		PermInviteManage,
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
		Resource: "employees", Label: "Employees",
		Permissions: []PermissionItem{
			{PermEmployeesRead, "View Employees", "List and view employee HR profiles"},
			{PermEmployeesCreate, "Create Employees", "Create new employee + user accounts"},
			{PermEmployeesUpdate, "Edit Employees", "Update employee HR fields (admin)"},
			{PermEmployeesDelete, "Delete Employees", "Soft-delete employee profiles"},
		},
	},
	{
		Resource: "dependents", Label: "Dependents",
		Permissions: []PermissionItem{
			{PermDependentsManage, "Manage Dependents", "Manage any employee's dependents (admin)"},
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
	{
		Resource: "invites", Label: "Invites",
		Permissions: []PermissionItem{
			{PermInviteManage, "Manage Invites", "Issue, resend, and revoke email invitations"},
		},
	},
}
