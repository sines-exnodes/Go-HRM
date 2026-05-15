---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-001
story_name: "Department Management"
detail_id: DR-008-001-02
detail_name: "Create Department"
parent_requirement: FR-US-001-05
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
---

# Detail Requirement: Create Department

**Detail ID:** DR-008-001-02
**Parent Requirement:** FR-US-001-05
**Story:** US-001-department-management
**Epic:** EP-008 (Organization Data)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with department management permission**, I want to **create a new department**, so that **it becomes available for selection across all HR modules (e.g., employee profiles)**.

**Purpose:** Allow authorized administrators to add new department records to the system. Departments are foundational reference data — once created, they are immediately available system-wide for use in employee assignments and other HR module selections.

**Target Users:** Any role with department management permission (configured via US-004). Users with view-only permission cannot access this feature.

**Key Functionality:**
- Single-field form (Department Name only — flat structure confirmed)
- Name uniqueness validation (case-insensitive)
- Returns to Department List on success or cancel

---

## 2. User Workflow

**Entry Point:** Department List → click "Add New" button (visible only to users with management permission)

**Preconditions:**
- User is signed in
- User's role has department management permission (configured via US-004)

**Main Flow:**
1. User clicks "Add New" on the Department List
2. System navigates to the Create Department full-page form
3. Form displays with "Department Information" card and empty Department Name field, auto-focused
4. User enters the department name
5. User clicks Save
6. System trims whitespace from the name
7. System validates: name is non-empty and unique (case-insensitive)
8. System creates the department
9. System returns user to the Department List — new department appears in the list

**Alternative Flows:**
- **Alt 1 — Empty Name:** At step 7, name is empty or whitespace-only → system shows inline error "Department name is required". User stays on form.
- **Alt 2 — Duplicate Name:** At step 7, name matches an existing department (case-insensitive) → system shows inline error "Department name already exists". User stays on form to correct.
- **Alt 3 — Name Too Long:** At step 7, name exceeds 100 characters → system shows inline error "Department name must be 100 characters or less". User stays on form.
- **Alt 4 — Cancel:** User clicks Cancel at any step → system returns to Department List. No department is created.
- **Alt 5 — Navigate Away:** User navigates away without saving → no department is created.

**Exit Points:**
- **Success:** Department created → user returned to Department List, new department visible
- **Cancel:** User clicks Cancel → returned to Department List, no changes
- **Error:** Validation error shown inline → user stays on form to correct

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Department Name | Text input | Non-empty, unique (case-insensitive), max 100 characters, trimmed before save | Yes | Empty | The name of the new department. Placeholder: "Enter department name" |

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Save | Primary button | Always visible; disabled while processing | Validates and submits the form | Dark background, white text, floppy disk icon. Placed top-right of page header. |
| Cancel | Secondary button | Always visible | Discards input and returns to Department List | Light background, dark text, X icon. Placed top-right of page header, left of Save. |

**Validation Error Messages:**

| Condition | Error Message | Display Location |
|-----------|--------------|------------------|
| Name is empty or whitespace-only on Save | "Department name is required" | Inline, below Department Name field |
| Name already exists (case-insensitive) | "Department name already exists" | Inline, below Department Name field |
| Name exceeds 100 characters | "Department name must be 100 characters or less" | Inline, below Department Name field |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Department Name field | Text input | Placeholder: "Enter department name" | Free text | The name to be assigned to the new department |
| Page title | Text | Always shown | "Create A New Department" | Identifies the current action |
| Breadcrumb | Navigation | Always shown | Organization Data / Departments / Create A New Department | Indicates location in the system |
| Section header | Text | Always shown | "Department Information" | Groups the form fields |
| Field label | Text | Always shown | "* Department name" (asterisk prefix = required) | Labels the input field |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Initial | Form first opens | Empty name field (auto-focused), Save and Cancel buttons enabled |
| Validation Error | Save clicked with invalid input | Inline error message below field; field highlighted; user stays on form |
| Saving | After Save clicked, system processing | Save button shows loading indicator and is disabled; form inputs disabled |
| Success | Department created | Redirected to Department List; new department visible in list |
| Cancel | User clicks Cancel | Redirected to Department List; no department created |

### Page Layout (Design Reference)

```
┌─────────────────────────────────────────────────────┐
│ [Sidebar]  │ Breadcrumb / Breadcrumb / Breadcrumb   │
│            ├─────────────────────────────────────────│
│            │ Create A New Department    [Cancel][Save]│
│            │                                         │
│            │   ┌──────────────────────────────────┐  │
│            │   │ Department Information            │  │
│            │   │                                  │  │
│            │   │  * Department name               │  │
│            │   │  ┌────────────────────────────┐  │  │
│            │   │  │ Enter department name       │  │  │
│            │   │  └────────────────────────────┘  │  │
│            │   │  [error message if any]           │  │
│            │   └──────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

- **AC-01:** "Add New" button is visible only to users with department management permission
- **AC-02:** Clicking "Add New" navigates to the Create Department full-page form
- **AC-03:** Form displays a "Department Information" card with a single Department Name field, Save button, and Cancel button
- **AC-04:** Department Name field is auto-focused when the form opens — user can type immediately
- **AC-05:** User cannot submit the form without entering a department name
- **AC-06:** System shows "Department name is required" when Save is clicked with an empty field or whitespace-only input
- **AC-07:** System shows "Department name already exists" when the entered name matches an existing department (case-insensitive)
- **AC-08:** System shows "Department name must be 100 characters or less" when name exceeds 100 characters
- **AC-09:** Validation errors appear inline below the Department Name field (not as a toast or popup)
- **AC-10:** Save button is disabled while the system is processing the save request (prevents duplicate submissions)
- **AC-11:** Leading and trailing whitespace is trimmed from the department name before saving
- **AC-12:** Successfully created department appears in the Department List immediately after save
- **AC-13:** New department is available for selection in other HR modules immediately after creation
- **AC-14:** Clicking Cancel returns user to the Department List without creating a department
- **AC-15:** No department is created if the user navigates away without saving
- **AC-16:** Name uniqueness check is case-insensitive ("Engineering" and "engineering" are treated as duplicates)
- **AC-17:** Breadcrumb shows the correct path: Organization Data / Departments / Create A New Department

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — valid name | "Engineering" | Department created; returned to list; "Engineering" visible | High |
| Empty name | "" (empty) | Error: "Department name is required"; stays on form | High |
| Whitespace-only name | "   " | Error: "Department name is required"; stays on form | High |
| Duplicate name (exact) | "HR" (already exists) | Error: "Department name already exists"; stays on form | High |
| Duplicate name (different case) | "hr" (when "HR" exists) | Error: "Department name already exists"; stays on form | High |
| Name at max length | 100-character name | Department created successfully | Medium |
| Name over max length | 101-character name | Error: "Department name must be 100 characters or less" | Medium |
| Name with leading/trailing spaces | "  Finance  " | Department created as "Finance" (trimmed) | Medium |
| Cancel without input | Click Cancel | Return to list; no department created | High |
| Cancel after typing | Type name, then Cancel | Return to list; no department created | Medium |
| Unauthorized user | User without management permission | "Add New" button not visible | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Only users with department management permission can access the Create Department form. Unauthorized access attempts should redirect to the Department List.
- **SR-02:** Department name must be unique across all departments, checked case-insensitively (normalize to lowercase before comparison).
- **SR-03:** Department name is trimmed of leading and trailing whitespace before the uniqueness check and before saving.
- **SR-04:** A name consisting entirely of whitespace is treated as empty and rejected.
- **SR-05:** A successfully created department is immediately available system-wide — no additional activation step required.
- **SR-06:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.

**State Transitions:**
```
[Department List] → [Click Add New] → [Create Form: Initial]
[Create Form: Initial] → [Click Save, valid input] → [Saving] → [Department List]
[Create Form: Initial] → [Click Save, invalid input] → [Validation Error state]
[Create Form: Validation Error] → [Correct input, Click Save] → [Saving] → [Department List]
[Create Form: Any] → [Click Cancel] → [Department List]
[Create Form: Any] → [Navigate away] → [Department List]
```

**Dependencies:**
- **US-004 (Role & Permission Management):** Controls which roles can access the Create form and "Add New" button
- **EP-002 (Employee Management):** Departments created here appear in employee profile department selectors

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Department Name field is auto-focused on page load — user can begin typing without clicking the field
- **UX-02:** Save button transitions to a loading/spinner state on click and is disabled until the system responds — prevents double submission
- **UX-03:** Inline validation error appears directly below the Department Name field — not as a toast or popup, keeping context clear
- **UX-04:** Cancel (secondary style, X icon) and Save (primary style, floppy disk icon) are placed in the top-right page header — not inside the form card, consistent with HRM layout conventions
- **UX-05:** Form is a full page (not a modal) — consistent with Edit Department pattern
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
- Figma node `3059:1793` — "Create A New Department"

---

## 8. Additional Information

### Out of Scope
- Parent department assignment (flat structure confirmed — name field only)
- Bulk department creation (create multiple at once)
- File or CSV import of departments
- Assigning a department head or manager at creation time

### Open Questions
- None — all requirements confirmed by Product Owner and Design Team.

### Related Features
- **DR-008-001-01** (Department List) — Create form is accessed via "Add New" button on the list; on success, returns here
- **DR-008-001-03** (Edit Department) — Shares the same full-page form pattern and layout
- **DR-008-001-04** (Delete Department) — Completes the CRUD lifecycle for department management
- **US-004** (Role & Permission Management) — Controls access to the Create form and "Add New" button
- **EP-002** (Employee Management) — Departments created here appear in employee profile selectors

### Notes
- Departments use hard delete (not soft deactivation). The Create form only adds; no status field exists.
- The "Add New" button on the Department List is only visible to users with management permission — users with view-only permission will never reach the Create form via normal navigation.

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
