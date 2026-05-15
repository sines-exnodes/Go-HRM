---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-004
story_name: "Role & Permission Management"
detail_id: DR-001-004-02
detail_name: "Create Role"
parent_requirement: FR-US-004-04
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
input_sources:
  - type: figma
    description: "Create A New Role screen"
    node_id: "3083:1837"
    extraction_date: "2026-03-24"
---

# Detail Requirement: Create Role

**Detail ID:** DR-001-004-02
**Parent Requirement:** FR-US-004-04
**Story:** US-004-role-permission-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with role management permission**, I want to **create a new role by providing a name and selecting permissions** so that **new access patterns can be defined as the organization grows**.

**Purpose:** Enable administrators to define custom roles with specific permission sets, eliminating hardcoded roles and allowing flexible access control as business needs evolve. Without this feature, adding a new role requires a code change instead of an administrator action.

**Target Users:** Any user with role management permission (not limited to a specific role such as Super Admin).

**Key Functionality:**
- Enter a unique role name
- Select permissions from a dynamically loaded, module-grouped checkbox matrix
- Save the role to make it immediately available for user assignment
- Option to save and continue creating another role without leaving the page

---

## 2. User Workflow

**Entry Point:** "+ Add New" button on the Role List page (DR-001-004-01).

**Preconditions:**
- User is signed in (US-001 Authentication)
- User has role management permission (US-004)

**Main Flow:**
1. User clicks "+ Add New" on the Role List page
2. System navigates to the Create A New Role page
3. System loads available permissions from the API, grouped by module
4. System displays empty form: blank role name field, all permission checkboxes unchecked
5. User enters a role name in the Role Information section
6. User selects permissions by checking individual checkboxes or using "Select All" per module group
7. User clicks **Save** or **Save & Create Another**
8. System validates: role name is not empty, does not exceed 100 characters, is unique (case-insensitive)
9. If validation passes: system saves the role
10. System displays success toast: "Role '[role name]' has been created"
11. If **Save**: system redirects to Role List (new role visible in list)
12. If **Save & Create Another**: system clears the form, resets scroll to top, user remains on Create Role page

**Alternative Flows:**

- **Alt 1 — Validation fails (empty name):** System displays inline error below role name field: "Role name is required". Form is not submitted. User corrects and retries.
- **Alt 2 — Validation fails (duplicate name):** System displays inline error: "Role name already exists". Form is not submitted. User enters a different name.
- **Alt 3 — Validation fails (max length):** System displays inline error: "Role name must not exceed 100 characters". Form is not submitted.
- **Alt 4 — Cancel (form modified):** System shows confirmation dialog: "Discard unsaved changes?" User clicks Confirm → redirects to Role List. User clicks Cancel → stays on form.
- **Alt 5 — Cancel (form untouched):** System redirects to Role List immediately without confirmation.
- **Alt 6 — Permission API fails to load:** System displays error message in Permissions section with retry option. Role Information section remains functional.

**Exit Points:**
- **Success (Save):** Role created → toast → redirect to Role List
- **Success (Save & Create Another):** Role created → toast → form cleared, stay on page
- **Cancel:** Redirect to Role List (with or without confirmation depending on form state)
- **Error:** Validation errors shown inline; user corrects and retries

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Max Length | Description |
|------------|------------|-----------------|-----------|---------------|------------|-------------|
| Role name | Text input | Not empty; unique (case-insensitive); allowed characters: letters (a-z, A-Z), numbers (0-9), spaces, hyphens (-), ampersands (&); leading/trailing whitespace trimmed automatically; validated on Save | Yes (*) | Empty | 100 characters | The display name for the new role |
| Permission checkboxes | Checkbox Group (per permission) | None — role can be saved with zero permissions | No | All unchecked | N/A | Individual permission toggles, dynamically loaded from API, grouped by module |

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| Cancel | Button (secondary) | Left in action bar | Always visible | If form dirty → "Discard changes?" dialog; if clean → redirect to Role List | Discard and return to list |
| Save & Create Another | Button (secondary/outline) | Center-right in action bar | Always visible; disabled + spinner while saving | Validate → save → toast → clear form, stay on page | Save and continue creating |
| Save | Button (primary) | Right in action bar | Always visible; disabled + spinner while saving | Validate → save → toast → redirect to Role List | Save and return to list |
| Select All (per module) | Checkbox | Left of module group label row | Unchecked by default; indeterminate (—) if partial selection | Toggles all checkboxes in that module group on/off | Quick-select all permissions in a module |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Role name field | Text input | Placeholder: "Enter role name" | Free text | Identity of the role being created |
| Permission module groups | Dynamic list from API | No checkboxes shown if API returns empty | Module label + checkbox grid | Available access rights to assign |
| Permission checkbox labels | Text per checkbox | N/A — always populated from API | "View [Module]", "Create [Module]", etc. | Individual permission names |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default | Page loads successfully | Empty role name field with placeholder, all checkboxes unchecked, three buttons enabled |
| Loading permissions | API call in progress for permission list | Loading indicator in Permissions card |
| Permission load error | API call fails | Error message in Permissions card with retry option |
| Validation error — empty name | User clicks Save with blank role name | Inline error below field: "Role name is required" |
| Validation error — duplicate name | User clicks Save with existing name | Inline error below field: "Role name already exists" |
| Validation error — max length | User enters >100 characters | Inline error below field: "Role name must not exceed 100 characters" |
| Saving | Save/Save & Create Another clicked, request in progress | Clicked button shows loading state (disabled + spinner); other buttons disabled |
| Success (Save) | Role created | Toast: "Role '[role name]' has been created" → redirect to Role List |
| Success (Save & Create Another) | Role created | Toast: "Role '[role name]' has been created" → form clears, scroll resets to top |
| Discard confirmation | Cancel clicked with modified form | Modal dialog: "Discard unsaved changes?" with Confirm and Cancel buttons |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Form Display:**
- **AC-01:** Create Role page displays with page title "Create A New Role" and two sections: Role Information and Permissions
- **AC-02:** Role name field is marked as mandatory (*)
- **AC-03:** All permission checkboxes are unchecked by default
- **AC-04:** Three buttons are visible: Cancel, Save & Create Another, Save

**Validation:**
- **AC-05:** User cannot save a role with an empty role name — inline error "Role name is required" is shown
- **AC-06:** User cannot save a role with a name that already exists (case-insensitive) — inline error "Role name already exists" is shown
- **AC-07:** User cannot enter more than 100 characters for role name — inline error "Role name must not exceed 100 characters" is shown
- **AC-08:** Validation errors display inline below the role name field
- **AC-09:** Validation is triggered on Save / Save & Create Another click (not on blur)

**Save Behavior:**
- **AC-10:** Save creates the role, shows toast "Role '[name]' has been created", and redirects to Role List
- **AC-11:** Save & Create Another creates the role, shows toast "Role '[name]' has been created", clears the form, and stays on the page
- **AC-12:** Role can be saved with zero permissions selected

**Cancel Behavior:**
- **AC-13:** Cancel on an untouched form redirects to Role List without confirmation
- **AC-14:** Cancel on a modified form shows "Discard unsaved changes?" confirmation dialog — Confirm discards and redirects, Cancel stays on form

**Permissions:**
- **AC-15:** Permissions are grouped by module with a label header per group
- **AC-16:** Each permission is an individual checkbox that can be toggled independently
- **AC-17:** Permission checkboxes are arranged in a 4-column grid per module group
- **AC-18:** Each module group has a "Select All" checkbox that toggles all permissions in that group
- **AC-19:** If some permissions in a module are manually unchecked, Select All shows an indeterminate (—) state

**Access Control:**
- **AC-20:** Create Role page is accessible only to users with role management permission
- **AC-21:** Direct URL access by unauthorized users redirects to an appropriate fallback page

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — Save | Valid name "HR Manager" + Users permissions checked | Role created, toast "Role 'HR Manager' has been created", redirect to list | High |
| Happy path — Save & Create Another | Valid name "Viewer" + View permissions checked | Role created, toast shown, form cleared | High |
| Empty name | Click Save with blank name | Inline error "Role name is required" | High |
| Duplicate name | Enter "Admin" when "admin" exists, click Save | Inline error "Role name already exists" | High |
| Case-insensitive duplicate | "MANAGER" exists, enter "Manager" | Inline error "Role name already exists" | High |
| Max length exceeded | Enter 101 characters | Inline error "Role name must not exceed 100 characters" | Medium |
| Zero permissions | Name "Empty Role", no checkboxes | Role created successfully | Medium |
| Select All | Check Select All under Users module | All 4 Users permissions checked | Medium |
| Partial deselect after Select All | Select All Users, then uncheck "Delete Users" | Select All shows indeterminate (—) state | Medium |
| Cancel dirty form | Enter name, click Cancel | "Discard unsaved changes?" dialog appears | Medium |
| Cancel clean form | No changes, click Cancel | Redirect to Role List immediately | Low |
| Unauthorized access | User without permission visits Create Role URL | Redirect / access denied | High |

---

## 6. System Rules

**Business Logic:**
- **Rule 1:** Role name must be unique organization-wide (case-insensitive comparison — "Admin" = "admin" = "ADMIN")
- **Rule 2:** Role name is trimmed of leading/trailing whitespace before saving
- **Rule 3:** Allowed characters for role name: letters (a-z, A-Z), numbers (0-9), spaces, hyphens (-), ampersands (&)
- **Rule 4:** A role can be created with zero permissions — permissions can be assigned later via Edit Role
- **Rule 5:** Newly created role is immediately available for user assignment across the platform
- **Rule 6:** Only users with role management permission can access the Create Role page
- **Rule 7:** No limit on the number of roles an organization can create
- **Rule 8:** System logs role creation events — records the creating user, role name, assigned permissions, and timestamp
- **Rule 9:** The list of modules and their permissions is fetched dynamically from the API — the UI renders whatever the backend provides, with no hardcoded module or permission names

**State Transitions:**
```
[Role List] → "+ Add New" click → [Create Role Form (empty)]
[Create Role Form] → Save (valid) → [Role List + new role visible]
[Create Role Form] → Save & Create Another (valid) → [Create Role Form (cleared)]
[Create Role Form] → Cancel (clean) → [Role List]
[Create Role Form] → Cancel (dirty) → [Confirmation Dialog]
[Confirmation Dialog] → Confirm discard → [Role List]
[Confirmation Dialog] → Cancel → [Create Role Form]
```

**Dependencies:**
- **Depends on:** US-001 (Authentication) — user must be signed in
- **Depends on:** US-004 permission data — role management permission must exist to access this page
- **Depends on:** Permissions API — module list and permission names loaded dynamically
- **Consumed by:** User Management — newly created roles are available for assignment to users
- **Consumed by:** All modules — role-permission data used for access control enforcement

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Permission checkboxes grouped by module with clear visual separation (horizontal lines) — reduces cognitive load when assigning permissions
- **UX-02:** Each module group has a "Select All" checkbox — checking it selects all permissions in that module; unchecking deselects all; shows indeterminate (—) state when partially selected
- **UX-03:** Form retains state during the session — if user scrolls down to permissions then scrolls back up, role name input value is preserved
- **UX-04:** Save / Save & Create Another buttons show loading spinner while request is in progress, preventing double submission
- **UX-05:** Success toast auto-dismisses after 5 seconds (with manual close option)
- **UX-06:** After "Save & Create Another", scroll position resets to top of form for the next entry
- **UX-07:** Keyboard navigation supported — Tab through fields and checkboxes, Enter to submit

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Form card 600px centered, 4-column checkbox grid |
| Tablet (768-1024px) | Form card full-width with padding, 2-column checkbox grid |
| Mobile (<768px) | Form card full-width, 1-column checkbox grid, stacked buttons |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through all fields, checkboxes, and buttons
- [x] Screen reader compatible — labels associated with inputs, error messages announced
- [x] Sufficient color contrast — meets WCAG 2.1 AA standards
- [x] Focus indicators visible — clear focus ring on interactive elements
- [x] Form errors linked to fields via aria-describedby

**Design References:**
- Figma: [Create A New Role](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3083-1837) (node `3083:1837`)
- Design tokens: See ANALYSIS.md Section 7 — Design Context [ADD-ON]
- Pattern reference: Department Create form (EP-008 US-001) — same full-page form layout

---

## 8. Additional Information

### Out of Scope
- Edit Role — separate detail requirement (DR-001-004-03, planned)
- Delete Role — separate detail requirement (DR-001-004-04, planned)
- User-to-role assignment — managed in User Management story
- Creating new permission keys — permissions are system-defined by developers; administrators only assign them to roles
- Role duplication / cloning
- Role import/export

### Open Questions
- None remaining for Create Role. All questions resolved during requirement writing session.

### Related Features
- **DR-001-004-01:** Role List — entry point for Create Role via "+ Add New" button
- **DR-001-004-03:** Edit Role (planned) — expected to mirror Create Role layout with pre-filled data
- **DR-001-004-04:** Delete Role (planned) — modal confirmation, blocked if users are assigned
- **US-001:** Authentication — user must be signed in
- **All modules:** Consume role-permission data for access control enforcement

### Notes
- The permission matrix is **dynamic** — the module list and permission names are fetched from the API at runtime. The Figma design shows Users and Roles as confirmed modules with CRUD permissions, plus placeholder "Module" groups. The actual UI will render whatever the API returns.
- The Figma design shows 2 action buttons (Cancel + Save), but the confirmed requirement is **3 buttons**: Cancel, Save & Create Another, Save. The design should be updated to reflect this.
- Edit Role screen has not been extracted from Figma yet. It is expected to follow an identical layout with pre-filled role name and pre-checked permissions.

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
| 1.0 | 2026-03-24 | BA Agent | Initial draft — full 8-section detail requirement with Figma design context |
