---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-002
story_name: "Position Management"
detail_id: DR-008-002-02
detail_name: "Create Position"
parent_requirement: FR-US-002-05
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
  - path: "./DR-008-002-01-position-list.md"
    relationship: sibling
---

# Detail Requirement: Create Position

**Detail ID:** DR-008-002-02
**Parent Requirement:** FR-US-002-05
**Story:** US-002-position-management
**Epic:** EP-008 (Organization Data)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with position management permission**, I want to **create a new position**, so that **it becomes available for selection across all HR modules (e.g., employee profiles)**.

**Purpose:** Allow authorized administrators to add new position records to the system. Positions are foundational reference data — once created, they are immediately available system-wide for use in employee assignments and other HR module selections.

**Target Users:** Any role with position management permission (configured via US-004). Users with view-only permission cannot access this feature.

**Key Functionality:**
- Single-field form (Position Name only — flat structure confirmed)
- Name uniqueness validation (case-insensitive)
- Returns to Position List on success or cancel

---

## 2. User Workflow

**Entry Point:** Position List → click "+ Add New" button (visible only to users with management permission)

**Preconditions:**
- User is signed in
- User's role has position management permission (configured via US-004)

**Main Flow:**
1. User clicks "+ Add New" on the Position List
2. System navigates to the Create Position full-page form
3. Form displays with "Position Information" card and empty Position Name field, auto-focused
4. User enters the position name
5. User clicks Save
6. System trims whitespace from the name
7. System validates: name is non-empty and unique (case-insensitive)
8. System creates the position
9. System returns user to the Position List — new position appears in the list

**Alternative Flows:**
- **Alt 1 — Empty Name:** At step 7, name is empty or whitespace-only → system shows inline error "Position name is required". User stays on form.
- **Alt 2 — Duplicate Name:** At step 7, name matches an existing position (case-insensitive) → system shows inline error "Position name already exists". User stays on form to correct.
- **Alt 3 — Name Too Long:** At step 7, name exceeds 100 characters → system shows inline error "Position name must be 100 characters or less". User stays on form.
- **Alt 4 — Cancel:** User clicks Cancel at any step → system returns to Position List. No position is created.
- **Alt 5 — Navigate Away:** User navigates away without saving → no position is created.

**Exit Points:**
- **Success:** Position created → user returned to Position List, new position visible
- **Cancel:** User clicks Cancel → returned to Position List, no changes
- **Error:** Validation error shown inline → user stays on form to correct

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Position Name | Text input | Non-empty, unique (case-insensitive), max 100 characters, trimmed before save | Yes | Empty | The name of the new position. Placeholder: "Enter position name" |

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Save | Primary button | Always visible; disabled while processing | Validates and submits the form | Dark background, white text, floppy disk icon. Placed top-right of page header. |
| Cancel | Secondary button | Always visible | Discards input and returns to Position List | Light background, dark text, X icon. Placed top-right of page header, left of Save. |

**Validation Error Messages:**

| Condition | Error Message | Display Location |
|-----------|--------------|------------------|
| Name is empty or whitespace-only on Save | "Position name is required" | Inline, below Position Name field |
| Name already exists (case-insensitive) | "Position name already exists" | Inline, below Position Name field |
| Name exceeds 100 characters | "Position name must be 100 characters or less" | Inline, below Position Name field |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Position Name field | Text input | Placeholder: "Enter position name" | Free text | The name to be assigned to the new position |
| Page title | Text | Always shown | "Create A New Position" | Identifies the current action |
| Breadcrumb | Navigation | Always shown | Organization Data / Positions / Create A New Position | Indicates location in the system |
| Section header | Text | Always shown | "Position Information" | Groups the form fields |
| Field label | Text | Always shown | "* Position name" (asterisk prefix = required) | Labels the input field |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Initial | Form first opens | Empty name field (auto-focused), Save and Cancel buttons enabled |
| Validation Error | Save clicked with invalid input | Inline error message below field; field highlighted; user stays on form |
| Saving | After Save clicked, system processing | Save button shows loading indicator and is disabled; form inputs disabled |
| Success | Position created | Redirected to Position List; new position visible in list |
| Cancel | User clicks Cancel | Redirected to Position List; no position created |

### Page Layout (Pending Design Delivery)

```
┌─────────────────────────────────────────────────────┐
│ [Sidebar]  │ Breadcrumb / Breadcrumb / Breadcrumb   │
│            ├─────────────────────────────────────────│
│            │ Create A New Position      [Cancel][Save]│
│            │                                         │
│            │   ┌──────────────────────────────────┐  │
│            │   │ Position Information              │  │
│            │   │                                  │  │
│            │   │  * Position name                 │  │
│            │   │  ┌────────────────────────────┐  │  │
│            │   │  │ Enter position name         │  │  │
│            │   │  └────────────────────────────┘  │  │
│            │   │  [error message if any]           │  │
│            │   └──────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

> **Note:** Figma design for Create Position is pending delivery from the Design Team. Layout above mirrors "Create A New Department" pattern (DR-008-001-02, Figma node `3059:1793`).

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

- **AC-01:** "+ Add New" button is visible only to users with position management permission
- **AC-02:** Clicking "+ Add New" navigates to the Create Position full-page form
- **AC-03:** Form displays a "Position Information" card with a single Position Name field, Save button, and Cancel button
- **AC-04:** Position Name field is auto-focused when the form opens — user can type immediately
- **AC-05:** User cannot submit the form without entering a position name
- **AC-06:** System shows "Position name is required" when Save is clicked with an empty field or whitespace-only input
- **AC-07:** System shows "Position name already exists" when the entered name matches an existing position (case-insensitive)
- **AC-08:** System shows "Position name must be 100 characters or less" when name exceeds 100 characters
- **AC-09:** Validation errors appear inline below the Position Name field (not as a toast or popup)
- **AC-10:** Save button is disabled while the system is processing the save request (prevents duplicate submissions)
- **AC-11:** Leading and trailing whitespace is trimmed from the position name before saving
- **AC-12:** Successfully created position appears in the Position List immediately after save
- **AC-13:** New position is available for selection in other HR modules immediately after creation
- **AC-14:** Clicking Cancel returns user to the Position List without creating a position
- **AC-15:** No position is created if the user navigates away without saving
- **AC-16:** Name uniqueness check is case-insensitive ("Engineer" and "engineer" are treated as duplicates)
- **AC-17:** Breadcrumb shows the correct path: Organization Data / Positions / Create A New Position

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — valid name | "Software Engineer" | Position created; returned to list; "Software Engineer" visible | High |
| Empty name | "" (empty) | Error: "Position name is required"; stays on form | High |
| Whitespace-only name | "   " | Error: "Position name is required"; stays on form | High |
| Duplicate name (exact) | "Manager" (already exists) | Error: "Position name already exists"; stays on form | High |
| Duplicate name (different case) | "manager" (when "Manager" exists) | Error: "Position name already exists"; stays on form | High |
| Name at max length | 100-character name | Position created successfully | Medium |
| Name over max length | 101-character name | Error: "Position name must be 100 characters or less" | Medium |
| Name with leading/trailing spaces | "  Analyst  " | Position created as "Analyst" (trimmed) | Medium |
| Cancel without input | Click Cancel | Return to list; no position created | High |
| Cancel after typing | Type name, then Cancel | Return to list; no position created | Medium |
| Unauthorized user | User without management permission | "+ Add New" button not visible | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Only users with position management permission can access the Create Position form. Unauthorized access attempts should redirect to the Position List.
- **SR-02:** Position name must be unique across all positions, checked case-insensitively (normalize to lowercase before comparison).
- **SR-03:** Position name is trimmed of leading and trailing whitespace before the uniqueness check and before saving.
- **SR-04:** A name consisting entirely of whitespace is treated as empty and rejected.
- **SR-05:** A successfully created position is immediately available system-wide — no additional activation step required.
- **SR-06:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.

**State Transitions:**
```
[Position List] → [Click Add New] → [Create Form: Initial]
[Create Form: Initial] → [Click Save, valid input] → [Saving] → [Position List]
[Create Form: Initial] → [Click Save, invalid input] → [Validation Error state]
[Create Form: Validation Error] → [Correct input, Click Save] → [Saving] → [Position List]
[Create Form: Any] → [Click Cancel] → [Position List]
[Create Form: Any] → [Navigate away] → [Position List]
```

**Dependencies:**
- **US-004 (Role & Permission Management):** Controls which roles can access the Create form and "+ Add New" button
- **EP-002 (Employee Management):** Positions created here appear in employee profile position selectors

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Position Name field is auto-focused on page load — user can begin typing without clicking the field
- **UX-02:** Save button transitions to a loading/spinner state on click and is disabled until the system responds — prevents double submission
- **UX-03:** Inline validation error appears directly below the Position Name field — not as a toast or popup, keeping context clear
- **UX-04:** Cancel (secondary style, X icon) and Save (primary style, floppy disk icon) are placed in the top-right page header — not inside the form card, consistent with HRM layout conventions
- **UX-05:** Form is a full page (not a modal) — consistent with Edit Position pattern
- **UX-06:** Pressing Enter while the Position Name field is focused triggers Save
- **UX-07:** Field label uses asterisk prefix format (`* Position name`) to indicate required field, consistent with the design system
- **UX-08:** Form content is grouped under a "Position Information" section card (white background, rounded border, max-width 600px, centered)

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Centered form card, max-width 600px |
| Tablet (768–1024px) | Full-width form with padding |
| Mobile (<768px) | Stacked layout, full-width input and buttons |

**Accessibility Requirements:**
- [ ] Keyboard navigable (Tab order: Position Name → Save → Cancel)
- [ ] Screen reader compatible (field label linked to input via ARIA)
- [ ] Error messages linked to field via ARIA for screen reader announcement
- [ ] Sufficient color contrast (primary button: dark bg / white text)
- [ ] Focus indicators visible on all interactive elements

**Design Reference:**
- Figma design for Create Position is **pending delivery** from the Design Team
- Expected to follow the same structure as "Create A New Department" (DR-008-001-02, Figma node `3059:1793`)

---

## 8. Additional Information

### Out of Scope
- Department association for the position (flat structure confirmed — position name field only in V1)
- Bulk position creation (create multiple at once)
- File or CSV import of positions
- Assigning a position level, grade, or salary band at creation time

### Open Questions
- [ ] **Design reference:** When will the Figma screen for Create Position be available? — **Owner:** Design Team — **Status:** Pending

### Related Features
- **DR-008-002-01** (Position List) — Create form is accessed via "+ Add New" button on the list; on success, returns here
- **DR-008-002-03** (Edit Position) — Shares the same full-page form pattern and layout
- **DR-008-002-04** (Delete Position) — Completes the CRUD lifecycle for position management
- **US-004** (Role & Permission Management) — Controls access to the Create form and "+ Add New" button
- **EP-002** (Employee Management) — Positions created here appear in employee profile selectors

### Notes
- Positions use hard delete (not soft deactivation). The Create form only adds; no status field exists.
- The "+ Add New" button on the Position List is only visible to users with management permission — users with view-only permission will never reach the Create form via normal navigation.

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
