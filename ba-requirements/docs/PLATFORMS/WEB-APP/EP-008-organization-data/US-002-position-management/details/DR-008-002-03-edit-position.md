---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-002
story_name: "Position Management"
detail_id: DR-008-002-03
detail_name: "Edit Position"
parent_requirement: FR-US-002-06
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
  - path: "./DR-008-002-02-create-position.md"
    relationship: sibling
---

# Detail Requirement: Edit Position

**Detail ID:** DR-008-002-03
**Parent Requirement:** FR-US-002-06
**Story:** US-002-position-management
**Epic:** EP-008 (Organization Data)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with position management permission**, I want to **edit an existing position's name**, so that **the position information stays accurate as the business changes**.

**Purpose:** Allow authorized administrators to correct or update the name of an existing position. When a position is renamed, the change is reflected immediately across all HR modules that reference that position.

**Target Users:** Any role with position management permission (configured via US-004). Users with view-only permission cannot access this feature.

**Key Functionality:**
- Full-page form pre-filled with the current position name
- Name uniqueness validation (case-insensitive, excludes the current position's own name)
- Returns to Position List on save or cancel

---

## 2. User Workflow

**Entry Point:** Position List → gear icon on the target row → select "Edit"

**Preconditions:**
- User is signed in
- User's role has position management permission (configured via US-004)
- The position to be edited exists in the list

**Main Flow:**
1. User locates the position in the Position List
2. User clicks the gear icon on the position's row
3. System displays a dropdown with available actions (Edit, Delete)
4. User selects "Edit"
5. System navigates to the Edit Position full-page form
6. Form displays pre-filled with the current position name; field is auto-focused (cursor at end of name)
7. User modifies (or keeps) the position name
8. User clicks Save
9. System trims whitespace from the name
10. System validates: name is non-empty and unique excluding the current position (case-insensitive)
11. System updates the position record
12. System returns user to the Position List — updated position name is visible

**Alternative Flows:**
- **Alt 1 — Empty Name:** At step 10, name is empty or whitespace-only → system shows inline error "Position name is required". User stays on form.
- **Alt 2 — Duplicate Name (Other Position):** At step 10, name matches a different existing position (case-insensitive) → system shows inline error "Position name already exists". User stays on form to correct.
- **Alt 3 — Name Too Long:** At step 10, name exceeds 100 characters → system shows inline error "Position name must be 100 characters or less". User stays on form.
- **Alt 4 — Unchanged Name:** At step 10, name is identical to the current position name (after trimming) → system saves successfully. No false duplicate error is raised.
- **Alt 5 — Cancel:** User clicks Cancel at any step → system returns to Position List. No changes are made.
- **Alt 6 — Navigate Away:** User navigates away without saving → no changes are made.

**Exit Points:**
- **Success:** Position updated → user returned to Position List, updated name visible
- **Cancel:** User clicks Cancel → returned to Position List, no changes
- **Error:** Validation error shown inline → user stays on form to correct

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Position Name | Text input | Non-empty, unique across positions excluding self (case-insensitive), max 100 characters, trimmed before save | Yes | Pre-filled with current name | The name of the position being edited. Placeholder: "Enter position name" (shown only if field is cleared) |

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Save | Primary button | Always visible; disabled while processing | Validates and submits the edit | Dark background, white text, floppy disk icon. Placed top-right of page header. |
| Cancel | Secondary button | Always visible | Discards changes and returns to Position List | Light background, dark text, X icon. Placed top-right of page header, left of Save. |

**Validation Error Messages:**

| Condition | Error Message | Display Location |
|-----------|--------------|------------------|
| Name is empty or whitespace-only on Save | "Position name is required" | Inline, below Position Name field |
| Name matches a different existing position (case-insensitive) | "Position name already exists" | Inline, below Position Name field |
| Name exceeds 100 characters | "Position name must be 100 characters or less" | Inline, below Position Name field |

**Key Difference from Create:** The uniqueness check excludes the current position — saving with the same name (unchanged) is always valid.

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Position Name field | Text input | Placeholder: "Enter position name" (if cleared) | Pre-filled with current name | The name currently assigned to this position |
| Page title | Text | Always shown | "Edit A Position" | Identifies the current action |
| Breadcrumb | Navigation | Always shown | Organization Data / Positions / Edit A Position | Indicates location in the system |
| Section header | Text | Always shown | "Position Information" | Groups the form fields |
| Field label | Text | Always shown | "* Position name" (asterisk prefix = required) | Labels the input field |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Initial | Form first opens | Position Name field pre-filled with current name (auto-focused, cursor at end); Save and Cancel buttons enabled |
| Validation Error | Save clicked with invalid input | Inline error message below field; field highlighted; user stays on form |
| Saving | After Save clicked, system processing | Save button shows loading indicator and is disabled; form inputs disabled |
| Success | Position updated | Redirected to Position List; updated position name visible in list |
| Cancel | User clicks Cancel | Redirected to Position List; no changes made |

### Page Layout (Pending Design Delivery)

```
┌─────────────────────────────────────────────────────┐
│ [Sidebar]  │ Breadcrumb / Breadcrumb / Breadcrumb   │
│            ├─────────────────────────────────────────│
│            │ Edit A Position            [Cancel][Save]│
│            │                                         │
│            │   ┌──────────────────────────────────┐  │
│            │   │ Position Information              │  │
│            │   │                                  │  │
│            │   │  * Position name                 │  │
│            │   │  ┌────────────────────────────┐  │  │
│            │   │  │ [Current position name]     │  │  │
│            │   │  └────────────────────────────┘  │  │
│            │   │  [error message if any]           │  │
│            │   └──────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

> **Note:** Figma design for Edit Position is pending delivery from the Design Team. Layout above mirrors "Edit A Department" pattern (DR-008-001-03, Figma node `3060:22578`).

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

- **AC-01:** Edit option is available in the gear icon dropdown for each position row
- **AC-02:** Edit option is visible only to users with position management permission
- **AC-03:** Selecting "Edit" navigates to the Edit Position full-page form
- **AC-04:** Form displays pre-filled with the current position name
- **AC-05:** Position Name field is auto-focused when the form opens, with cursor positioned at the end of the current name
- **AC-06:** User cannot submit the form without entering a position name
- **AC-07:** System shows "Position name is required" when Save is clicked with an empty field or whitespace-only input
- **AC-08:** System shows "Position name already exists" when the entered name matches a different existing position (case-insensitive)
- **AC-09:** Saving with the same name as the current position (unchanged) succeeds without a duplicate error
- **AC-10:** System shows "Position name must be 100 characters or less" when name exceeds 100 characters
- **AC-11:** Validation errors appear inline below the Position Name field (not as a toast or popup)
- **AC-12:** Save button is disabled while the system is processing the save request (prevents duplicate submissions)
- **AC-13:** Leading and trailing whitespace is trimmed from the position name before saving
- **AC-14:** Successfully updated position name appears in the Position List immediately after save
- **AC-15:** Updated name is reflected across all HR modules that reference this position
- **AC-16:** Clicking Cancel returns user to the Position List without saving any changes
- **AC-17:** Breadcrumb shows the correct path: Organization Data / Positions / Edit A Position

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — change name | Current: "Analyst"; Enter: "Senior Analyst" | Position renamed; returned to list; "Senior Analyst" visible | High |
| Happy path — no change | Current: "Manager"; Enter: "Manager" (unchanged) | Save succeeds; returned to list; "Manager" still visible | High |
| Happy path — whitespace trimmed | Current: "Manager"; Enter: "  Director  " | Position renamed to "Director"; returned to list | Medium |
| Empty name | Clear field, click Save | Error: "Position name is required"; stays on form | High |
| Whitespace-only name | Enter "   ", click Save | Error: "Position name is required"; stays on form | High |
| Duplicate name (other position) | "Engineer" (when "Engineer" already exists as different position) | Error: "Position name already exists"; stays on form | High |
| Duplicate — different case | "engineer" (when "Engineer" exists as different position) | Error: "Position name already exists"; stays on form | High |
| Same name — different case | Current: "Manager"; Enter: "manager" | Save succeeds (same position, case-insensitive match excluded) | Medium |
| Name at max length | 100-character name | Position updated successfully | Medium |
| Name over max length | 101-character name | Error: "Position name must be 100 characters or less" | Medium |
| Cancel after editing | Modify name, then Cancel | Return to list; original name unchanged | High |
| Unauthorized user | User without management permission | Gear icon not visible (no access to Edit) | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Only users with position management permission can access the Edit Position form. The gear icon with Edit option is hidden for users without this permission.
- **SR-02:** Position name uniqueness is checked case-insensitively across all positions, **excluding the current position being edited**. This allows saving with an unchanged name.
- **SR-03:** Position name is trimmed of leading and trailing whitespace before the uniqueness check and before saving.
- **SR-04:** A name consisting entirely of whitespace is treated as empty and rejected.
- **SR-05:** The updated position name is immediately reflected across all HR modules — no additional refresh or sync step required.
- **SR-06:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.

**State Transitions:**
```
[Position List] → [Gear icon → Select Edit] → [Edit Form: Initial (pre-filled)]
[Edit Form: Initial] → [Click Save, valid input] → [Saving] → [Position List]
[Edit Form: Initial] → [Click Save, invalid input] → [Validation Error state]
[Edit Form: Validation Error] → [Correct input, Click Save] → [Saving] → [Position List]
[Edit Form: Any] → [Click Cancel] → [Position List]
[Edit Form: Any] → [Navigate away] → [Position List (no changes)]
```

**Dependencies:**
- **US-004 (Role & Permission Management):** Controls which roles can access the Edit action and gear icon
- **EP-002 (Employee Management):** Employee profile records referencing this position will reflect the updated name

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Position Name field is auto-focused on page load with cursor positioned at the end of the pre-filled name — user can edit immediately without clicking the field
- **UX-02:** Save button transitions to a loading/spinner state on click and is disabled until the system responds — prevents double submission
- **UX-03:** Inline validation error appears directly below the Position Name field — not as a toast or popup, keeping context clear
- **UX-04:** Cancel (secondary style, X icon) and Save (primary style, floppy disk icon) are placed in the top-right page header — not inside the form card, consistent with HRM layout conventions
- **UX-05:** Form is a full page (not a modal) — consistent with Create Position pattern
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
- Figma design for Edit Position is **pending delivery** from the Design Team
- Expected to follow the same structure as "Edit A Department" (DR-008-001-03, Figma node `3060:22578`)
- Same layout as Create Position — identical structure, different page title and pre-filled field

---

## 8. Additional Information

### Out of Scope
- Editing position code or internal ID (name field only)
- Bulk editing of multiple positions at once
- Edit history or audit log for name changes
- Reassigning employees to a different position via this form
- Department association for the position (out of scope for V1)

### Open Questions
- [ ] **Design reference:** When will the Figma screen for Edit Position be available? — **Owner:** Design Team — **Status:** Pending

### Related Features
- **DR-008-002-01** (Position List) — Edit form is accessed via gear icon on the list; on success, returns here with updated name
- **DR-008-002-02** (Create Position) — Shares the same full-page form layout; Edit differs only in pre-filled field and page title
- **DR-008-002-04** (Delete Position) — Completes the CRUD lifecycle; accessible from the same gear icon dropdown
- **US-004** (Role & Permission Management) — Controls access to the Edit action and gear icon visibility
- **EP-002** (Employee Management) — Employee records referencing this position reflect the updated name immediately

### Notes
- The Edit and Create forms are visually identical — same card, same field, same button placement. The only differences are: (1) page title "Edit A Position" vs "Create A New Position", and (2) the Position Name field is pre-filled in Edit vs empty in Create.
- The uniqueness check on Edit explicitly excludes the current position. This is the critical difference from Create's uniqueness check and must be implemented correctly to avoid blocking unchanged saves.

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
