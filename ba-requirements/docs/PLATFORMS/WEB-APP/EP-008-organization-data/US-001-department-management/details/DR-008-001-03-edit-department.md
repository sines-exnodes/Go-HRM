---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-001
story_name: "Department Management"
detail_id: DR-008-001-03
detail_name: "Edit Department"
parent_requirement: FR-US-001-06
status: draft
version: "1.0"
created_date: 2026-03-05
last_updated: 2026-03-05
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "./DR-008-001-01-department-list.md"
    relationship: sibling
  - path: "./DR-008-001-02-create-department.md"
    relationship: sibling
---

# Detail Requirement: Edit Department

**Detail ID:** DR-008-001-03
**Parent Requirement:** FR-US-001-06
**Story:** US-001-department-management
**Epic:** EP-008 (Organization Data)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with department management permission**, I want to **edit an existing department's name**, so that **the department information stays accurate as the business changes**.

**Purpose:** Allow authorized administrators to correct or update the name of an existing department. When a department is renamed, the change is reflected immediately across all HR modules that reference that department.

**Target Users:** Any role with department management permission (configured via US-004). Users with view-only permission cannot access this feature.

**Key Functionality:**
- Full-page form pre-filled with the current department name
- Name uniqueness validation (case-insensitive, excludes the current department's own name)
- Returns to Department List on save or cancel

---

## 2. User Workflow

**Entry Point:** Department List → gear icon on the target row → select "Edit"

**Preconditions:**
- User is signed in
- User's role has department management permission (configured via US-004)
- The department to be edited exists in the list

**Main Flow:**
1. User locates the department in the Department List
2. User clicks the gear icon on the department's row
3. System displays a dropdown with available actions (Edit, Delete)
4. User selects "Edit"
5. System navigates to the Edit Department full-page form
6. Form displays pre-filled with the current department name; field is auto-focused (cursor at end of name)
7. User modifies (or keeps) the department name
8. User clicks Save
9. System trims whitespace from the name
10. System validates: name is non-empty and unique excluding the current department (case-insensitive)
11. System updates the department record
12. System returns user to the Department List — updated department name is visible

**Alternative Flows:**
- **Alt 1 — Empty Name:** At step 10, name is empty or whitespace-only → system shows inline error "Department name is required". User stays on form.
- **Alt 2 — Duplicate Name (Other Department):** At step 10, name matches a different existing department (case-insensitive) → system shows inline error "Department name already exists". User stays on form to correct.
- **Alt 3 — Name Too Long:** At step 10, name exceeds 100 characters → system shows inline error "Department name must be 100 characters or less". User stays on form.
- **Alt 4 — Unchanged Name:** At step 10, name is identical to the current department name (after trimming) → system saves successfully. No false duplicate error is raised.
- **Alt 5 — Cancel:** User clicks Cancel at any step → system returns to Department List. No changes are made.
- **Alt 6 — Navigate Away:** User navigates away without saving → no changes are made.

**Exit Points:**
- **Success:** Department updated → user returned to Department List, updated name visible
- **Cancel:** User clicks Cancel → returned to Department List, no changes
- **Error:** Validation error shown inline → user stays on form to correct

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Department Name | Text input | Non-empty, unique across departments excluding self (case-insensitive), max 100 characters, trimmed before save | Yes | Pre-filled with current name | The name of the department being edited. Placeholder: "Enter department name" (shown only if field is cleared) |

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Save | Primary button | Always visible; disabled while processing | Validates and submits the edit | Dark background, white text, floppy disk icon. Placed top-right of page header. |
| Cancel | Secondary button | Always visible | Discards changes and returns to Department List | Light background, dark text, X icon. Placed top-right of page header, left of Save. |

**Validation Error Messages:**

| Condition | Error Message | Display Location |
|-----------|--------------|------------------|
| Name is empty or whitespace-only on Save | "Department name is required" | Inline, below Department Name field |
| Name matches a different existing department (case-insensitive) | "Department name already exists" | Inline, below Department Name field |
| Name exceeds 100 characters | "Department name must be 100 characters or less" | Inline, below Department Name field |

**Key Difference from Create:** The uniqueness check excludes the current department — saving with the same name (unchanged) is always valid.

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Department Name field | Text input | Placeholder: "Enter department name" (if cleared) | Pre-filled with current name | The name currently assigned to this department |
| Page title | Text | Always shown | "Edit A Department" | Identifies the current action |
| Breadcrumb | Navigation | Always shown | Organization Data / Departments / Edit A Department | Indicates location in the system |
| Section header | Text | Always shown | "Department Information" | Groups the form fields |
| Field label | Text | Always shown | "* Department name" (asterisk prefix = required) | Labels the input field |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Initial | Form first opens | Department Name field pre-filled with current name (auto-focused, cursor at end); Save and Cancel buttons enabled |
| Validation Error | Save clicked with invalid input | Inline error message below field; field highlighted; user stays on form |
| Saving | After Save clicked, system processing | Save button shows loading indicator and is disabled; form inputs disabled |
| Success | Department updated | Redirected to Department List; updated department name visible in list |
| Cancel | User clicks Cancel | Redirected to Department List; no changes made |

### Page Layout (Design Reference)

```
┌─────────────────────────────────────────────────────┐
│ [Sidebar]  │ Breadcrumb / Breadcrumb / Breadcrumb   │
│            ├─────────────────────────────────────────│
│            │ Edit A Department          [Cancel][Save]│
│            │                                         │
│            │   ┌──────────────────────────────────┐  │
│            │   │ Department Information            │  │
│            │   │                                  │  │
│            │   │  * Department name               │  │
│            │   │  ┌────────────────────────────┐  │  │
│            │   │  │ [Current department name]   │  │  │
│            │   │  └────────────────────────────┘  │  │
│            │   │  [error message if any]           │  │
│            │   └──────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

- **AC-01:** Edit option is available in the gear icon dropdown for each department row
- **AC-02:** Edit option is visible only to users with department management permission
- **AC-03:** Selecting "Edit" navigates to the Edit Department full-page form
- **AC-04:** Form displays pre-filled with the current department name
- **AC-05:** Department Name field is auto-focused when the form opens, with cursor positioned at the end of the current name
- **AC-06:** User cannot submit the form without entering a department name
- **AC-07:** System shows "Department name is required" when Save is clicked with an empty field or whitespace-only input
- **AC-08:** System shows "Department name already exists" when the entered name matches a different existing department (case-insensitive)
- **AC-09:** Saving with the same name as the current department (unchanged) succeeds without a duplicate error
- **AC-10:** System shows "Department name must be 100 characters or less" when name exceeds 100 characters
- **AC-11:** Validation errors appear inline below the Department Name field (not as a toast or popup)
- **AC-12:** Save button is disabled while the system is processing the save request (prevents duplicate submissions)
- **AC-13:** Leading and trailing whitespace is trimmed from the department name before saving
- **AC-14:** Successfully updated department name appears in the Department List immediately after save
- **AC-15:** Updated name is reflected across all HR modules that reference this department
- **AC-16:** Clicking Cancel returns user to the Department List without saving any changes
- **AC-17:** Breadcrumb shows the correct path: Organization Data / Departments / Edit A Department

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — change name | Current: "HR"; Enter: "Human Resources" | Department renamed; returned to list; "Human Resources" visible | High |
| Happy path — no change | Current: "HR"; Enter: "HR" (unchanged) | Save succeeds; returned to list; "HR" still visible | High |
| Happy path — whitespace trimmed | Current: "HR"; Enter: "  Finance  " | Department renamed to "Finance"; returned to list | Medium |
| Empty name | Clear field, click Save | Error: "Department name is required"; stays on form | High |
| Whitespace-only name | Enter "   ", click Save | Error: "Department name is required"; stays on form | High |
| Duplicate name (other dept) | "IT" (when "IT" already exists as different dept) | Error: "Department name already exists"; stays on form | High |
| Duplicate — different case | "it" (when "IT" exists as different dept) | Error: "Department name already exists"; stays on form | High |
| Same name — different case | Current: "HR"; Enter: "hr" | Save succeeds (same department, case-insensitive match excluded) | Medium |
| Name at max length | 100-character name | Department updated successfully | Medium |
| Name over max length | 101-character name | Error: "Department name must be 100 characters or less" | Medium |
| Cancel after editing | Modify name, then Cancel | Return to list; original name unchanged | High |
| Unauthorized user | User without management permission | Gear icon not visible (no access to Edit) | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Only users with department management permission can access the Edit Department form. The gear icon with Edit option is hidden for users without this permission.
- **SR-02:** Department name uniqueness is checked case-insensitively across all departments, **excluding the current department being edited**. This allows saving with an unchanged name.
- **SR-03:** Department name is trimmed of leading and trailing whitespace before the uniqueness check and before saving.
- **SR-04:** A name consisting entirely of whitespace is treated as empty and rejected.
- **SR-05:** The updated department name is immediately reflected across all HR modules — no additional refresh or sync step required.
- **SR-06:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.

**State Transitions:**
```
[Department List] → [Gear icon → Select Edit] → [Edit Form: Initial (pre-filled)]
[Edit Form: Initial] → [Click Save, valid input] → [Saving] → [Department List]
[Edit Form: Initial] → [Click Save, invalid input] → [Validation Error state]
[Edit Form: Validation Error] → [Correct input, Click Save] → [Saving] → [Department List]
[Edit Form: Any] → [Click Cancel] → [Department List]
[Edit Form: Any] → [Navigate away] → [Department List (no changes)]
```

**Dependencies:**
- **US-004 (Role & Permission Management):** Controls which roles can access the Edit action and gear icon
- **EP-002 (Employee Management):** Employee profile records referencing this department will reflect the updated name

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Department Name field is auto-focused on page load with cursor positioned at the end of the pre-filled name — user can edit immediately without clicking the field
- **UX-02:** Save button transitions to a loading/spinner state on click and is disabled until the system responds — prevents double submission
- **UX-03:** Inline validation error appears directly below the Department Name field — not as a toast or popup, keeping context clear
- **UX-04:** Cancel (secondary style, X icon) and Save (primary style, floppy disk icon) are placed in the top-right page header — not inside the form card, consistent with HRM layout conventions
- **UX-05:** Form is a full page (not a modal) — consistent with Create Department pattern
- **UX-06:** Pressing Enter while the Department Name field is focused triggers Save
- **UX-07:** Field label uses asterisk prefix format (`* Department name`) to indicate required field, consistent with the design system
- **UX-08:** Form content is grouped under a "Department Information" section card (white background, rounded border, max-width 600px, centered)

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Centered form card, max-width 600px |
| Tablet (768–1024px) | Full-width form with padding |
| Mobile (<768px) | Stacked layout, full-width input and buttons |

**Accessibility Requirements:**
- [ ] Keyboard navigable (Tab order: Department Name → Save → Cancel)
- [ ] Screen reader compatible (field label linked to input via ARIA)
- [ ] Error messages linked to field via ARIA for screen reader announcement
- [ ] Sufficient color contrast (primary button: dark bg / white text)
- [ ] Focus indicators visible on all interactive elements

**Design Reference:**
- Figma node `3060:22578` — "Edit A Department"
- Same layout as Create Department (`3059:1793`) — identical structure, different page title and pre-filled field

---

## 8. Additional Information

### Out of Scope
- Editing department code or internal ID (name field only)
- Bulk editing of multiple departments at once
- Edit history or audit log for name changes
- Reassigning employees to a different department via this form
- Parent department assignment (flat structure confirmed — name field only)

### Open Questions
- None — all requirements confirmed by Product Owner and Design Team.

### Related Features
- **DR-008-001-01** (Department List) — Edit form is accessed via gear icon on the list; on success, returns here with updated name
- **DR-008-001-02** (Create Department) — Shares the same full-page form layout; Edit differs only in pre-filled field and page title
- **DR-008-001-04** (Delete Department) — Completes the CRUD lifecycle; accessible from the same gear icon dropdown
- **US-004** (Role & Permission Management) — Controls access to the Edit action and gear icon visibility
- **EP-002** (Employee Management) — Employee records referencing this department reflect the updated name immediately

### Notes
- The Edit and Create forms are visually identical — same card, same field, same button placement. The only differences are: (1) page title "Edit A Department" vs "Create A New Department", and (2) the Department Name field is pre-filled in Edit vs empty in Create.
- The uniqueness check on Edit explicitly excludes the current department. This is the critical difference from Create's uniqueness check and must be implemented correctly to avoid blocking unchanged saves.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | — | — | Pending |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-05 | BA Agent | Initial draft |
