---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-002
story_id: US-001
story_name: "Leave Requests"
detail_id: DR-002-001-06
detail_name: "Change Leave Quota"
parent_requirement: FR-US-001-14
status: draft
version: "1.0"
created_date: 2026-04-10
last_updated: 2026-04-10
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "./DR-002-001-01-leave-requests-list.md"
    relationship: sibling
  - path: "./DR-002-001-02-create-leave-request.md"
    relationship: sibling
  - path: "./DR-002-001-03-update-leave-request.md"
    relationship: sibling
  - path: "./DR-002-001-04-delete-leave-request.md"
    relationship: sibling
  - path: "./DR-002-001-05-change-leave-request-status.md"
    relationship: sibling
input_sources:
  - type: figma
    description: "Change Leave Quota screen"
    node_id: "3224:2523"
    extraction_date: "2026-04-10"
---

# Detail Requirement: Change Leave Quota

**Detail ID:** DR-002-001-06
**Parent Requirement:** FR-US-001-14
**Story:** US-001-leave-requests
**Epic:** EP-002 (Leave Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with leave quota management permission**, I want to **change an employee's leave quota (annual and sick leave)**, so that **the employee's available leave balance is updated to reflect company policy, tenure adjustments, or corrective actions**.

**Purpose:** This feature allows administrators and managers to adjust employee leave quotas directly from the User Details page. Leave quotas define the maximum number of days an employee can take for each leave type. Changes take effect immediately and impact all future leave calculations and requests.

**Target Users:**
- Users with **leave quota management permission** (typically HR administrators, managers)
- Permission controlled via US-004 (Role & Permission Management)

**Key Functionality:**
- Access via User Details page left action panel ("Leave Quota" button)
- Edit two quota fields: Annual Leave Quota and Sick Leave Quota
- Pre-filled with current quota values on page load
- Changes apply immediately upon save
- Single Save button to confirm changes

---

## 2. User Workflow

**Entry Point:** User Details page (EP-001 US-005) > Left action panel > "Leave Quota" button

**Preconditions:**
- User is signed in (US-001)
- User has leave quota management permission via US-004
- Target employee exists and is viewable by the user

**Main Flow:**
1. User navigates to User Details page for a specific employee
2. User clicks "Leave Quota" button in the left action panel
3. System displays Change Leave Quota form card with:
   - Card title: "Change Leave Quota"
   - Description text explaining the impact of changes
   - Two input fields pre-filled with current values: Annual Leave Quota, Sick Leave Quota
   - Save button (full-width, black background)
4. User modifies one or both quota values
5. User clicks "Save" button
6. System validates the input (must be non-negative numbers)
7. System updates the employee's leave quotas
8. System displays success toast: "Leave quota has been updated"
9. Form remains on the same page with updated values

**Alternative Flows:**
- **Alt 1 — Navigate away (dirty form):** User clicks another action button in left panel or back arrow while form has unsaved changes → "Discard unsaved changes?" confirmation dialog
- **Alt 2 — Navigate away (clean form):** No changes made → navigates immediately without confirmation
- **Alt 3 — Validation error:** Invalid input (negative number, non-numeric) → inline error below field
- **Alt 4 — API error:** Server fails to update → error toast: "Failed to update leave quota. Please try again."

**Exit Points:**
- **Success:** Quota updated → toast displayed → remains on same page
- **Navigate away:** User clicks another action button or back arrow → navigates to selected view
- **Error:** Error toast → form stays open for retry

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Annual Leave Quota | Number | Non-negative integer or decimal (0 or greater); max 365 | Yes | Current quota value | Total annual leave days allocated to the employee |
| Sick Leave Quota | Number | Non-negative integer or decimal (0 or greater); max 365 | Yes | Current quota value | Total sick leave days allocated to the employee |

**Validation Details:**
- Both fields accept whole numbers and decimals (e.g., 12, 15.5)
- Minimum value: 0
- Maximum value: 365 (no employee can have more than a year's worth of leave)
- Empty field treated as validation error: "[Field name] is required"
- Non-numeric input: "Please enter a valid number"
- Negative number: "Quota cannot be negative"

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Leave Quota (left panel) | Button | Highlighted when active | Shows Change Leave Quota form | Navigation button in User Details left action panel |
| Save | Primary button | Always enabled | Submits quota changes | Full-width black button at bottom of form card |
| Back arrow | Icon button | Always visible | Returns to User List | ArrowLeft icon in page title row |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Page title | Text | N/A | "User Details" | Identifies the page context |
| Employee name breadcrumb | Text | N/A | "> [Employee Name]" | Shows which employee's quota is being edited |
| Card title | Text | N/A | "Change Leave Quota" | Form heading |
| Description | Text | N/A | Multi-line explanatory text | Explains impact of changes |
| Annual Leave Quota | Number | "0" | Numeric input | Current annual leave allocation |
| Sick Leave Quota | Number | "0" | Numeric input | Current sick leave allocation |

### Layout Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb                                        │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ← User Details  > Henry Tran                                                │
│                                                                              │
│  ┌─────────────────────┐  ┌──────────────────────────────────────────────┐   │
│  │  Overview           │  │  Change Leave Quota                          │   │
│  │  Update Information │  │                                              │   │
│  │  Change User Role   │  │  Update the user's leave quota assigned to   │   │
│  │  Change Email       │  │  their account. The new quota will be        │   │
│  │  Reset Password     │  │  applied immediately and reflected in all    │   │
│  │  Activate/Deactivate│  │  leave-related calculations and requests.    │   │
│  │ [Leave Quota]       │  │                                              │   │
│  │  Delete User        │  │  Please ensure the leave quota is accurate   │   │
│  │                     │  │  before saving.                              │   │
│  │                     │  │                                              │   │
│  │                     │  │  * Annual Leave Quota                        │   │
│  │                     │  │  ┌──────────────────────────────────────┐    │   │
│  │                     │  │  │ 12                                   │    │   │
│  │                     │  │  └──────────────────────────────────────┘    │   │
│  │                     │  │                                              │   │
│  │                     │  │  * Sick Leave Quota                          │   │
│  │                     │  │  ┌──────────────────────────────────────┐    │   │
│  │                     │  │  │ 6                                    │    │   │
│  │                     │  │  └──────────────────────────────────────┘    │   │
│  │                     │  │                                              │   │
│  │                     │  │  ┌──────────────────────────────────────┐    │   │
│  │                     │  │  │              Save                    │    │   │
│  │                     │  │  └──────────────────────────────────────┘    │   │
│  └─────────────────────┘  └──────────────────────────────────────────────┘   │
│        189px                              600px                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page loading | Skeleton content in form card area |
| Default | Form displayed | Pre-filled quota values, Save button enabled |
| Dirty | Values changed from original | Same as default; dirty state tracked internally |
| Processing | Save clicked | Save button shows loading spinner, inputs disabled |
| Success | Save completed | Success toast; form shows updated values |
| Validation Error | Invalid input | Inline error below the invalid field |
| API Error | Server fails | Error toast; form stays open for retry |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Access & Visibility:**
- **AC-01:** "Leave Quota" button is visible in the User Details left action panel only to users with leave quota management permission
- **AC-02:** Users without permission do not see the "Leave Quota" button
- **AC-03:** Direct URL access without permission redirects to fallback page

**Form Display:**
- **AC-04:** Form card displays title "Change Leave Quota" and descriptive text
- **AC-05:** Annual Leave Quota field is pre-filled with the employee's current annual leave quota
- **AC-06:** Sick Leave Quota field is pre-filled with the employee's current sick leave quota
- **AC-07:** Both fields are marked as mandatory with asterisk (*)
- **AC-08:** Save button is full-width (600px) with black background (#010101)

**Validation:**
- **AC-09:** Empty field shows inline error: "[Field name] is required"
- **AC-10:** Non-numeric input shows inline error: "Please enter a valid number"
- **AC-11:** Negative number shows inline error: "Quota cannot be negative"
- **AC-12:** Value exceeding 365 shows inline error: "Maximum quota is 365 days"
- **AC-13:** Decimal values are accepted (e.g., 12.5)
- **AC-14:** Validation errors are shown inline below the respective field

**Save Flow:**
- **AC-15:** Clicking Save with valid values updates both quota fields
- **AC-16:** Save button shows loading spinner while processing
- **AC-17:** All inputs are disabled during processing
- **AC-18:** Success toast: "Leave quota has been updated"
- **AC-19:** After successful save, form remains on the same page with updated values
- **AC-20:** Changes take effect immediately — employee's available leave days are recalculated

**Navigation:**
- **AC-21:** "Leave Quota" button is highlighted when this form is active
- **AC-22:** Clicking another action button in left panel navigates to that view
- **AC-23:** If form has unsaved changes, "Discard unsaved changes?" confirmation appears before navigating
- **AC-24:** Back arrow returns to User List page
- **AC-25:** Back arrow with unsaved changes shows discard confirmation

**Error Handling:**
- **AC-26:** API error shows toast: "Failed to update leave quota. Please try again."
- **AC-27:** On API error, form stays open for retry
- **AC-28:** If employee record was deleted, error toast: "User not found" + redirect to User List

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Update annual quota | Change 12 to 15 | Quota updated; toast displayed | High |
| Update sick quota | Change 6 to 10 | Quota updated; toast displayed | High |
| Update both quotas | Change both values | Both quotas updated; single toast | High |
| Set quota to 0 | Enter 0 | Allowed; quota set to 0 | Medium |
| Enter decimal | Enter 12.5 | Allowed; quota set to 12.5 | Medium |
| Empty annual quota | Clear field, save | Inline error: "Annual Leave Quota is required" | High |
| Negative value | Enter -5 | Inline error: "Quota cannot be negative" | High |
| Value > 365 | Enter 400 | Inline error: "Maximum quota is 365 days" | Medium |
| Non-numeric input | Enter "abc" | Inline error: "Please enter a valid number" | High |
| Navigate with unsaved changes | Click Overview | Discard confirmation dialog | High |
| Navigate without changes | Click Overview | Immediate navigation, no dialog | Medium |
| No permission | User without perm | Leave Quota button not visible | High |
| API error | Server fails | Error toast; form stays open | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Leave quota management requires specific permission configured in US-004
- **SR-02:** Quota values are stored per employee per leave type (annual, sick)
- **SR-03:** Quota changes take effect immediately upon successful save
- **SR-04:** Changing quota does NOT affect already-approved leave requests
- **SR-05:** Leave Days Remaining is recalculated as: Quota - (Sum of Approved Leave Days for that type)
- **SR-06:** If new quota is less than already-used days, Leave Days Remaining becomes negative (allowed — admin must resolve)
- **SR-07:** Quota values support decimals for half-day precision (e.g., 12.5 days)
- **SR-08:** Maximum quota is 365 days per leave type (business constraint)
- **SR-09:** Audit trail: quota changes should be logged with timestamp, old value, new value, changed by

**State Transitions:**
```
[Page Load] → [Fetch employee quota] → [Pre-fill form]
[Edit field] → [Mark form dirty]
[Save clicked] → [Validate] → [Valid: Submit to API] → [Success: Update form, show toast]
[Save clicked] → [Validate] → [Invalid: Show inline errors]
[Navigate] → [Dirty: Show discard dialog] → [Confirm: Navigate] / [Cancel: Stay]
[Navigate] → [Clean: Navigate immediately]
```

**Calculations/Formulas:**
- Leave Days Remaining = Leave Quota - Sum of Approved Leave Days (for that leave type)
- If Leave Days Remaining < 0, employee has overused their quota

**Dependencies:**
- **US-001 (Authentication):** User must be signed in
- **US-004 (Role & Permission Management):** Controls leave quota management permission
- **US-005 (User Management):** Provides User Details page and employee data
- **EP-002 Leave Balance System:** Quota feeds into leave balance calculations
- **DR-002-001-02 (Create Leave Request):** Uses quota to validate available days
- **DR-002-001-05 (Change Leave Request Status):** Approve action deducts from quota-based balance

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Form is pre-filled with current values — user sees existing quota before making changes
- **UX-02:** Descriptive text explains the impact of changes before user edits
- **UX-03:** "Leave Quota" button highlighted in left panel indicates current active view
- **UX-04:** Full-width Save button is prominent and easy to target
- **UX-05:** Decimal support allows half-day precision (e.g., 15.5 annual days)
- **UX-06:** Inline validation errors appear immediately below the relevant field
- **UX-07:** Discard confirmation prevents accidental loss of changes
- **UX-08:** Form stays on same page after save — allows further adjustments without navigation
- **UX-09:** Success toast confirms the action without blocking workflow

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>=1280px) | Two-panel layout: 189px left panel + 600px form card |
| Below desktop | Out of scope |

**Accessibility Requirements:**
- [x] Keyboard navigable (Tab through fields, Enter to save)
- [x] Screen reader compatible (labels associated with inputs)
- [x] Sufficient color contrast (black button on white, error text in red)
- [x] Focus indicators visible on inputs and buttons
- [x] Error messages announced to screen readers

**Design References:**
- Figma: [Change Leave Quota](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3224-2523) (node `3224:2523`)
- Layout follows User Details two-panel pattern (consistent with Update Information, Change Email, etc.)
- Save button style: black (#010101) background, white text, full-width 600px

---

## 8. Additional Information

### Out of Scope
- Leave type configuration (adding new leave types beyond Annual/Sick) — future enhancement
- Bulk quota updates for multiple employees at once
- Quota history/audit log viewer (system logs changes, but no UI to view history)
- Automatic quota adjustments based on tenure or policy rules
- Leave accrual rules (monthly quota increases)
- Mobile or tablet layout
- Undo after save

### Open Questions
- [ ] **Additional leave types:** Should there be more leave types beyond Annual and Sick (e.g., Personal, Maternity, Unpaid)? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Quota audit log UI:** Should admins be able to view the history of quota changes for an employee? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| EP-001 US-005: User Details | Entry point — Leave Quota button in left action panel |
| DR-002-001-01: Leave Requests List | Displays leave data affected by quota |
| DR-002-001-02: Create Leave Request | Validates available days against quota |
| DR-002-001-05: Change Leave Request Status | Approve deducts from quota-based balance |
| US-001: Authentication | User must be signed in |
| US-004: Role & Permission Management | Controls leave quota management permission |

### Notes
- **Cross-epic feature:** Although this DR is placed in EP-002 (Leave Management), the entry point is in EP-001 (User Details page). This is intentional — the feature belongs to Leave Management conceptually, but is accessed from User Management for usability (all user-related actions in one place).
- **Negative balance allowed:** If admin sets quota below already-used days, the system allows negative Leave Days Remaining. This is a conscious design decision to give admins flexibility; the system does not block quota reduction.
- **Two leave types only:** Per Figma design, only Annual Leave Quota and Sick Leave Quota are supported. Additional leave types would require schema changes and UI updates.
- **Immediate effect:** Quota changes are immediate — no approval workflow for quota adjustments.

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
| 1.0 | 2026-04-10 | BA Agent | Initial draft — Change Leave Quota via User Details page with Annual and Sick leave quota fields |
