---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-004
story_name: "Role & Permission Management"
detail_id: DR-001-004-03
detail_name: "Edit Role"
parent_requirement: FR-US-004-05
status: draft
version: "1.0"
created_date: "2026-03-24"
last_updated: "2026-03-24"
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "../TODO.yaml"
    relationship: sibling
  - path: "./DR-001-004-01-role-list.md"
    relationship: sibling
  - path: "./DR-001-004-02-create-role.md"
    relationship: sibling
input_sources: []
---

# Detail Requirement: Edit Role

**Detail ID:** DR-001-004-03
**Parent Requirement:** FR-US-004-05
**Story:** US-004-role-permission-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with role management permission**, I want to **edit an existing role's name and permissions** so that **role definitions stay accurate as business needs change**.

**Purpose:** Allow authorized administrators to update a role's name and reassign its permissions. When a role is updated, the changes take effect immediately for all users assigned to that role — no additional activation step required. This keeps access control aligned with evolving business structures without requiring new roles to be created.

**Target Users:** Any user with role management permission (not limited to a specific role such as Super Admin).

**Key Functionality:**
- Full-page form pre-filled with the current role name and pre-checked permissions
- Role name uniqueness validation (case-insensitive, excludes the current role's own name)
- Permission checkboxes pre-checked based on current assignments, with Select All per module
- Permission changes take effect immediately for all assigned users upon save
- Returns to Role List on success or cancel

---

## 2. User Workflow

**Entry Point:** Role List → gear icon on the target row → select "Edit"

**Preconditions:**
- User is signed in (US-001 Authentication)
- User has role management permission (US-004)
- The role to be edited exists in the list

**Main Flow:**
1. User locates the role in the Role List
2. User clicks the gear icon on the role's row
3. System displays a dropdown with available actions (Edit, Delete)
4. User selects "Edit"
5. System navigates to the Edit A Role full-page form
6. System loads current role data and available permissions from the API
7. Form displays pre-filled: role name field populated with current name (auto-focused, cursor at end), permission checkboxes checked to match current assignments
8. User modifies the role name and/or toggles permission checkboxes
9. User clicks **Save**
10. System trims whitespace from the role name
11. System validates: name is not empty, does not exceed 100 characters, is unique excluding the current role (case-insensitive)
12. System updates the role record (name and permissions)
13. System displays success toast: "Role '[role name]' has been updated"
14. System redirects to Role List — updated role visible with new name/permissions

**Alternative Flows:**

- **Alt 1 — Validation fails (empty name):** System displays inline error below role name field: "Role name is required". Form is not submitted. User corrects and retries.
- **Alt 2 — Validation fails (duplicate name):** System displays inline error: "Role name already exists". Form is not submitted. User enters a different name.
- **Alt 3 — Validation fails (max length):** System displays inline error: "Role name must not exceed 100 characters". Form is not submitted.
- **Alt 4 — Unchanged name:** Name is identical to the current role name (after trimming) — system saves successfully. No false duplicate error is raised.
- **Alt 5 — Cancel (form modified):** System shows confirmation dialog: "Discard unsaved changes?" User clicks Confirm → redirects to Role List. User clicks Cancel → stays on form.
- **Alt 6 — Cancel (form untouched):** System redirects to Role List immediately without confirmation.
- **Alt 7 — Permission API fails to load:** System displays error message in Permissions section with retry option. Role Information section remains functional with pre-filled name.

**Exit Points:**
- **Success:** Role updated → toast "Role '[name]' has been updated" → redirect to Role List
- **Cancel:** Redirect to Role List (with or without confirmation depending on form state)
- **Error:** Validation errors shown inline; user corrects and retries

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Max Length | Description |
|------------|------------|-----------------|-----------|---------------|------------|-------------|
| Role name | Text input | Not empty; unique excluding self (case-insensitive); allowed characters: letters (a-z, A-Z), numbers (0-9), spaces, hyphens (-), ampersands (&); leading/trailing whitespace trimmed automatically; validated on Save | Yes (*) | Pre-filled with current name | 100 characters | The display name for the role being edited |
| Permission checkboxes | Checkbox Group (per permission) | None — role can be saved with zero permissions | No | Pre-checked based on current assignments | N/A | Individual permission toggles, dynamically loaded from API, grouped by module |

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| Cancel | Button (secondary) | Left in action bar | Always visible | If form dirty → "Discard changes?" dialog; if clean → redirect to Role List | Discard and return to list |
| Save | Button (primary) | Right in action bar | Always visible; disabled + spinner while saving | Validate → save → toast → redirect to Role List | Save changes and return to list |
| Select All (per module) | Checkbox | Left of module group label row | Checked if all permissions in group are assigned; unchecked if none; indeterminate (—) if partial | Toggles all checkboxes in that module group on/off | Quick-select/deselect all permissions in a module |

**Key Difference from Create:** No "Save & Create Another" button — Edit always returns to the Role List on save.

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Role name field | Text input | Placeholder: "Enter role name" (if cleared) | Pre-filled with current name | Identity of the role being edited |
| Permission module groups | Dynamic list from API | No checkboxes shown if API returns empty | Module label + checkbox grid | Available access rights to assign |
| Permission checkbox labels | Text per checkbox | N/A — always populated from API | "View [Module]", "Create [Module]", etc. | Individual permission names |
| Permission checkbox states | Checked/unchecked | N/A | Pre-checked based on current role-permission assignments | Which permissions are currently assigned |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page loads, API calls in progress | Skeleton loader for role name, loading indicator in Permissions card |
| Default | Page loaded, data populated | Pre-filled role name field, pre-checked permission checkboxes, Cancel and Save buttons enabled |
| Permission load error | Permission API call fails | Error message in Permissions card with retry option. Role name field still functional. |
| Validation error — empty name | User clears name and clicks Save | Inline error below field: "Role name is required" |
| Validation error — duplicate name | Name matches a different existing role (case-insensitive) | Inline error below field: "Role name already exists" |
| Validation error — max length | User enters >100 characters | Inline error below field: "Role name must not exceed 100 characters" |
| Saving | Save clicked, request in progress | Save button shows loading state (disabled + spinner); Cancel button disabled |
| Success | Role updated | Toast: "Role '[role name]' has been updated" → redirect to Role List |
| Discard confirmation | Cancel clicked with modified form | Modal dialog: "Discard unsaved changes?" with Confirm and Cancel buttons |

### Page Layout (Pending Design Delivery)

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Users Management / Roles / Edit A Role  [Top Bar] │
├──────────────┬──────────────────────────────────────────────────┤
│  [Sidebar]   │  Edit A Role                    [Cancel]  [Save] │
│              │                                                  │
│              │  ┌──────────────────────────────────────────┐    │
│              │  │ Role Information                         │    │
│              │  │ * Role name                              │    │
│              │  │ ┌──────────────────────────────────────┐ │    │
│              │  │ │ [Current role name]                   │ │    │
│              │  │ └──────────────────────────────────────┘ │    │
│              │  └──────────────────────────────────────────┘    │
│              │                                                  │
│              │  ┌──────────────────────────────────────────┐    │
│              │  │ Permissions                              │    │
│              │  │ ─────────────────────────────────────    │    │
│              │  │ Users                                    │    │
│              │  │ ☑ View Users  ☑ Create  ☐ Edit  ☐ Delete│    │
│              │  │ ─────────────────────────────────────    │    │
│              │  │ Roles                                    │    │
│              │  │ ☑ View Roles  ☑ Create  ☑ Edit  ☑ Delete│    │
│              │  │ ─────────────────────────────────────    │    │
│              │  │ [Additional module groups...]            │    │
│              │  └──────────────────────────────────────────┘    │
└──────────────┴──────────────────────────────────────────────────┘
```

> **Note:** Figma design for Edit Role is pending delivery from the Design Team. Layout above mirrors "Create A New Role" pattern (DR-001-004-02, Figma node `3083:1837`) with pre-filled data.

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Form Display:**
- **AC-01:** Edit Role page displays with page title "Edit A Role" and two sections: Role Information and Permissions
- **AC-02:** Role name field is pre-filled with the current role name and marked as mandatory (*)
- **AC-03:** Role name field is auto-focused when the form opens, with cursor positioned at the end of the current name
- **AC-04:** Permission checkboxes are pre-checked to match the role's current permission assignments
- **AC-05:** Two buttons are visible: Cancel and Save (no "Save & Create Another")

**Validation:**
- **AC-06:** User cannot save a role with an empty role name — inline error "Role name is required" is shown
- **AC-07:** User cannot save a role with a name that matches a different existing role (case-insensitive) — inline error "Role name already exists" is shown
- **AC-08:** Saving with the same name as the current role (unchanged) succeeds without a duplicate error
- **AC-09:** User cannot enter more than 100 characters for role name — inline error "Role name must not exceed 100 characters" is shown
- **AC-10:** Validation errors display inline below the role name field
- **AC-11:** Validation is triggered on Save click (not on blur)

**Save Behavior:**
- **AC-12:** Save updates the role, shows toast "Role '[name]' has been updated", and redirects to Role List
- **AC-13:** Role can be saved with zero permissions selected (all unchecked)
- **AC-14:** Permission changes take effect immediately for all users currently assigned to the role — no additional sync or refresh required

**Cancel Behavior:**
- **AC-15:** Cancel on an untouched form redirects to Role List without confirmation
- **AC-16:** Cancel on a modified form shows "Discard unsaved changes?" confirmation dialog — Confirm discards and redirects, Cancel stays on form
- **AC-17:** "Form modified" is detected by comparing current name and permission state against the original values loaded from API

**Permissions:**
- **AC-18:** Permissions are grouped by module with a label header per group
- **AC-19:** Each permission is an individual checkbox that can be toggled independently
- **AC-20:** Permission checkboxes are arranged in a 4-column grid per module group
- **AC-21:** Each module group has a "Select All" checkbox — checked if all in group are assigned, indeterminate (—) if partial, unchecked if none
- **AC-22:** Select All state on page load reflects the pre-checked permissions for each module

**Access Control:**
- **AC-23:** Edit Role page is accessible only to users with role management permission
- **AC-24:** Direct URL access by unauthorized users redirects to an appropriate fallback page

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — rename + change permissions | Current: "HR Manager"; Change to "HR Lead" + add View Reports | Role updated, toast "Role 'HR Lead' has been updated", redirect to list | High |
| Happy path — no changes | Open Edit, click Save immediately | Save succeeds, toast shown, redirect to list | High |
| Happy path — permissions only | Keep name, uncheck "Delete Users" | Role updated with reduced permissions, toast shown | High |
| Empty name | Clear name, click Save | Inline error "Role name is required" | High |
| Duplicate name (other role) | Enter "Admin" when "admin" exists as different role | Inline error "Role name already exists" | High |
| Case-insensitive duplicate | "MANAGER" exists as different role, enter "Manager" | Inline error "Role name already exists" | High |
| Same name — different case | Current: "HR"; Enter: "hr" | Save succeeds (same role, self-exclusion) | Medium |
| Max length exceeded | Enter 101 characters | Inline error "Role name must not exceed 100 characters" | Medium |
| Zero permissions | Uncheck all, click Save | Role updated with 0 permissions, toast shown | Medium |
| Select All → partial deselect | Select All Users, then uncheck "Delete Users" | Select All shows indeterminate (—) state | Medium |
| Cancel dirty form | Change name, click Cancel | "Discard unsaved changes?" dialog appears | Medium |
| Cancel clean form | No changes, click Cancel | Redirect to Role List immediately | Low |
| Unauthorized access | User without permission visits Edit Role URL | Redirect / access denied | High |
| Permission API fails | API returns error on load | Error in Permissions card with retry, name field still pre-filled | Medium |
| Permission changes immediate | Save with changed permissions | All assigned users' access updated immediately | High |

---

## 6. System Rules

**Business Logic:**
- **SR-01:** Role name must be unique organization-wide (case-insensitive comparison), **excluding the current role being edited**. Saving with the unchanged name is always valid.
- **SR-02:** Role name is trimmed of leading/trailing whitespace before the uniqueness check and before saving.
- **SR-03:** Allowed characters for role name: letters (a-z, A-Z), numbers (0-9), spaces, hyphens (-), ampersands (&).
- **SR-04:** A role can be saved with zero permissions — all permissions can be revoked.
- **SR-05:** Permission changes take effect immediately for all users assigned to the role. No additional activation, refresh, or sync step required (BR-US-004-04).
- **SR-06:** Only users with role management permission can access the Edit Role page. The gear icon with Edit option is hidden for users without this permission.
- **SR-07:** System logs role update events — records the editing user, previous and new role name, previous and new permissions, and timestamp.
- **SR-08:** The list of modules and their permissions is fetched dynamically from the API — the UI renders whatever the backend provides, with no hardcoded module or permission names.
- **SR-09:** "Form dirty" detection compares current form state (name + permission selections) against the original values loaded from the API. Only actual changes trigger the discard confirmation on Cancel.
- **SR-10:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.

**State Transitions:**
```
[Role List] → [Gear → Edit] → [Edit Role Form (pre-filled)]
[Edit Role Form] → Save (valid) → [Saving] → [Toast] → [Role List (updated)]
[Edit Role Form] → Save (invalid) → [Validation Error state]
[Edit Role Form: Validation Error] → [Correct input, Save] → [Saving] → [Toast] → [Role List]
[Edit Role Form] → Cancel (clean) → [Role List]
[Edit Role Form] → Cancel (dirty) → [Confirmation Dialog]
[Confirmation Dialog] → Confirm discard → [Role List]
[Confirmation Dialog] → Cancel → [Edit Role Form]
```

**Dependencies:**
- **Depends on:** US-001 (Authentication) — user must be signed in
- **Depends on:** US-004 permission data — role management permission must exist to access this page
- **Depends on:** Permissions API — module list and permission names loaded dynamically
- **Consumed by:** User Management — updated role permissions affect all assigned users immediately
- **Consumed by:** All modules — role-permission data used for access control enforcement

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Role name field is auto-focused on page load with cursor at the end of the pre-filled name — user can edit immediately without clicking the field
- **UX-02:** Permission checkboxes grouped by module with clear visual separation (horizontal lines) — reduces cognitive load when reviewing and adjusting permissions
- **UX-03:** Each module group has a "Select All" checkbox — on page load, its state (checked, unchecked, indeterminate) reflects the current permissions; toggling it selects/deselects all in the group
- **UX-04:** Save button shows loading spinner while request is in progress, preventing double submission; Cancel button also disabled during save
- **UX-05:** Success toast auto-dismisses after 5 seconds (with manual close option)
- **UX-06:** Keyboard navigation supported — Tab through fields and checkboxes, Enter to submit
- **UX-07:** "Discard unsaved changes?" dialog uses default focus on "Cancel" (stay on form) — prevents accidental data loss if user presses Enter

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Form card 600px centered, 4-column checkbox grid |
| Tablet (768–1024px) | Form card full-width with padding, 2-column checkbox grid |
| Mobile (<768px) | Form card full-width, 1-column checkbox grid, stacked buttons |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through all fields, checkboxes, and buttons
- [x] Screen reader compatible — labels associated with inputs, error messages announced
- [x] Sufficient color contrast — meets WCAG 2.1 AA standards
- [x] Focus indicators visible — clear focus ring on interactive elements
- [x] Form errors linked to fields via aria-describedby

**Design References:**
- Figma design for Edit Role is **pending delivery** from the Design Team
- Expected to follow the same structure as "Create A New Role" (DR-001-004-02, Figma node `3083:1837`)
- Same layout as Create Role — identical structure, with pre-filled name and pre-checked permissions
- Design tokens: See ANALYSIS.md Section 7 — Design Context [ADD-ON]

---

## 8. Additional Information

### Out of Scope
- Creating a new role from the Edit page (no "Save & Create Another" button)
- Role duplication / cloning from the Edit page
- Editing the role's internal ID or system key (name and permissions only)
- Viewing or managing which users are assigned to this role from the Edit page
- Audit log or history of permission changes visible on the Edit page
- Bulk editing of multiple roles at once

### Open Questions
- [ ] **Edit Role Figma screen:** When will the Design Team deliver the Edit Role screen? Expected to mirror Create Role with pre-filled data. — **Owner:** Design Team — **Status:** Pending

### Related Features
- **DR-001-004-01:** Role List — Edit is triggered from the gear icon; on success, returns here with updated name/permissions
- **DR-001-004-02:** Create Role — Shares the same full-page form layout; Edit differs in pre-filled data, "Save" only (no "Save & Create Another"), and self-exclusion uniqueness check
- **DR-001-004-04:** Delete Role (planned) — Accessible from the same gear icon dropdown
- **US-001:** Authentication — user must be signed in
- **All modules:** Updated role-permission data takes effect immediately for access control enforcement

### Notes
- The Edit and Create forms are visually identical — same cards, same fields, same button placement (except Create has "Save & Create Another"). The key differences are: (1) page title "Edit A Role" vs "Create A New Role", (2) role name field is pre-filled, (3) permission checkboxes are pre-checked, (4) uniqueness check excludes the current role.
- **Permission changes are immediate** (BR-US-004-04). When an administrator unchecks a permission and saves, all users assigned to that role lose that access right immediately — no refresh or re-login required. This is a critical business rule that must be clearly communicated to administrators.
- The "Select All" checkbox state on page load may be indeterminate (—) if the role has only some permissions in a module — this differs from Create where all Select All checkboxes start unchecked.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | BA Agent | 2026-03-24 | Draft |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |
| Tech Lead | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-24 | BA Agent | Initial draft — mirrors Create Role (DR-001-004-02) with pre-filled data and self-exclusion uniqueness |
