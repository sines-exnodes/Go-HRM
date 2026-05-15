---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-009
story_id: US-002
story_name: "Attendance Settings"
detail_id: DR-009-002-01
detail_name: "Late Arrival Threshold Configuration"
parent_requirement: FR-US-002-01
status: draft
version: "1.0"
created_date: 2026-04-23
last_updated: 2026-04-23
related_documents:
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../../EPIC.md"
    relationship: parent
  - path: "../../../EP-004-attendance-management/US-001-attendance-list/details/DR-004-001-01-attendance-list.md"
    relationship: data-consumer
  - path: "../../../../MOBILE-APP/EP-003-attendance-management/US-001-daily-attendance/details/DR-003-001-01-check-in-out.md"
    relationship: data-consumer
input_sources:
  - type: figma
    description: "Organization Settings — Attendance tab with Late Arrival Threshold time picker"
    extraction_date: "2026-04-23"
migration_note: "Relocated from EP-004/US-001 (DR-004-001-02) to EP-009/US-002 (DR-009-002-01)"
---

# Detail Requirement: Late Arrival Threshold Configuration

**Detail ID:** DR-009-002-01
**Parent Requirement:** FR-US-002-01
**Story:** US-002-attendance-settings
**Epic:** EP-009 (Organization Settings)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **Admin or HR manager**, I want to **configure the late arrival threshold time** so that **employees who check in after this time are automatically marked as late in the attendance system**.

**Purpose:** This feature allows authorized users to define the organization's late arrival policy by setting a specific time threshold. When employees check in via the mobile app (MOBILE-APP EP-003) after this configured time, their attendance is flagged as "Late" (L) in the attendance matrix (EP-004 DR-004-001-01). This configuration is organization-wide and affects all employees.

**Target Users:**
- **Admin** — Primary user for organization-wide settings configuration
- **HR Manager** — May configure attendance policies based on company requirements

**Key Functionality:**
- View current late arrival threshold value
- Modify threshold to any valid time (e.g., 9:00 AM, 9:15 AM, 8:30 AM)
- Save changes with immediate effect on future attendance calculations
- Help text explaining the impact of this setting

---

## 2. User Workflow

**Entry Point:** Sidebar navigation > Organization Settings > Attendance tab (left panel)

**Preconditions:**
- User is signed in (EP-001 US-001)
- User's role has "Organization Settings Management" permission (EP-001 US-004)

**Main Flow:**
1. User navigates to Organization Settings from the sidebar or system menu
2. System loads the Organization Settings page
3. User clicks the "Attendance" tab in the left panel
4. System displays the Attendance settings with the Late Arrival Threshold field
5. System pre-fills the time picker with the current threshold value (default: 9:00 AM)
6. User modifies the time value using the time picker
7. User clicks the Save button
8. System validates the input (valid time format)
9. System saves the new threshold value
10. System displays success toast: "Late arrival threshold has been updated"
11. User remains on the same page with updated value

**Alternative Flows:**

| Action | Flow |
|--------|------|
| **Invalid Time** | User enters invalid time format > System shows inline validation error > Save button remains disabled until corrected |
| **No Changes** | User clicks Save without modifying value > System saves (no error), toast shown, no actual change |
| **Navigate Away (Dirty)** | User modifies value but navigates away without saving > System shows "Discard unsaved changes?" confirmation dialog |
| **Navigate Away (Clean)** | User navigates away without modifications > No confirmation, immediate navigation |

**Exit Points:**
- **Success:** Toast message displayed, stay on page with saved value
- **Cancel/Discard:** Return to previous page (if dirty form discarded)
- **Tab Switch:** Navigate to another settings tab (with dirty form check if applicable)

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Late Arrival Threshold | Time Picker | Valid time format (HH:MM AM/PM); must be between 12:00 AM and 11:59 PM | Yes | 9:00 AM | The time after which employee check-ins are marked as "Late" |

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Attendance Tab | Tab button | Left panel tab list | Displays Attendance settings | Active state shows filled/highlighted background |
| Time Picker | Input + dropdown | Always enabled | Opens time selection UI | Allows free-form time entry or selection from common times |
| Save | Button (primary) | Full-width, bottom of form | Saves threshold value | Black background (#010101), white text |
| Help Text | Static text | Always visible below field | None | "Employees who check in after this time will be marked as late." |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Current Threshold | Time | N/A (always has value) | "9:00 AM" or "9:15 AM" | The currently configured late arrival cutoff time |
| Help Text | Static text | N/A | Sentence | Explains what the setting controls |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page/tab loading | Skeleton for time picker field |
| Data Loaded | Normal state | Time picker pre-filled with current value, help text visible, Save button enabled |
| Validation Error | Invalid time entered | Inline error below field: "Please enter a valid time" |
| Saving | After Save clicked | Save button shows loading spinner, form disabled |
| Success | After successful save | Toast: "Late arrival threshold has been updated" |

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** Admin/HR can navigate to Organization Settings > Attendance tab and see the Late Arrival Threshold field
- **AC-02:** The time picker displays the current threshold value (default 9:00 AM if never configured)
- **AC-03:** User can select or type any valid time (e.g., 9:15 AM, 8:45 AM, 10:00 AM)
- **AC-04:** Help text "Employees who check in after this time will be marked as late." is always visible below the field
- **AC-05:** Save button is full-width with black background (#010101)
- **AC-06:** After clicking Save, success toast "Late arrival threshold has been updated" is displayed
- **AC-07:** After saving, user remains on the same page with the updated value
- **AC-08:** If user modifies the value and navigates away without saving, "Discard unsaved changes?" confirmation appears
- **AC-09:** If user has not modified the value, navigation away does not show confirmation dialog
- **AC-10:** The saved threshold value is used by the attendance system to determine late arrivals for all future check-ins
- **AC-11:** Users without "Organization Settings Management" permission cannot access this page (redirected to fallback)

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| View current threshold | Navigate to Attendance tab | Time picker shows current value (e.g., 9:00 AM) | High |
| Change threshold | Select 9:15 AM, click Save | Toast shown, value persists on page refresh | High |
| Invalid time format | Enter "25:00" or "abc" | Inline error, Save disabled | High |
| Dirty form navigation | Change value, click sidebar link | "Discard unsaved changes?" dialog appears | Medium |
| Clean form navigation | No changes, click sidebar link | Immediate navigation, no dialog | Medium |
| Permission denied | User without permission accesses URL directly | Redirected to fallback page | High |
| Attendance calculation | Set threshold to 9:15 AM, employee checks in at 9:10 AM | Employee marked as On-time (not Late) | High |
| Attendance calculation | Set threshold to 9:15 AM, employee checks in at 9:20 AM | Employee marked as Late | High |

---

## 6. System Rules

### SR-001: Permission Required

Only users with "Organization Settings Management" permission can access and modify this setting. Users without permission are redirected to a fallback page when attempting direct URL access.

### SR-002: Organization-Wide Setting

The late arrival threshold is a single, organization-wide value. There are no per-department or per-employee thresholds in this version.

### SR-003: Immediate Effect

Changes to the threshold take effect immediately upon save. All check-ins recorded after the save time are evaluated against the new threshold.

### SR-004: No Retroactive Changes

Changing the threshold does NOT retroactively update existing attendance records. Historical attendance (already marked as On-time or Late) remains unchanged.

### SR-005: Default Value

If the threshold has never been configured, the system uses 9:00 AM as the default value. This default is pre-populated in the time picker on first access.

### SR-006: Time Zone

The threshold time is interpreted in the organization's configured time zone (system-wide setting). All check-in comparisons use the same time zone.

### SR-007: Late Calculation Logic

An employee's first check-in of the day is compared against the threshold:
- **First check-in <= threshold** = On-time (displayed as "checkmark" in attendance matrix)
- **First check-in > threshold** = Late (displayed as "L" in attendance matrix)

### SR-008: Integration with Attendance Matrix

The threshold configured here directly affects the attendance status displayed in EP-004 DR-004-001-01 (Attendance List). When the threshold changes, future attendance entries reflect the new policy.

---

## 7. UX Optimizations

### Layout

| Item | Specification |
|------|--------------|
| Page Structure | Organization Settings page with left panel tab navigation |
| Attendance Tab | One of multiple tabs in the left panel (alongside Company Profile) |
| Form Card | Single card containing the Late Arrival Threshold field |
| Card Width | 600px (consistent with other settings forms per knowledge base §11.1) |
| Field Layout | Label above field, help text below field |
| Save Button | Full-width (matches card width), positioned at bottom of form |

### Button Styling

| Element | Style |
|---------|-------|
| Save Button | Background: #010101 (black), Text: white, Full-width |
| Save Button (loading) | Shows spinner, button disabled |
| Save Button (disabled) | Reduced opacity when validation fails |

### Loading States

| Element | Loading State |
|---------|--------------|
| Time Picker | Skeleton rectangle |
| Save Button | Disabled during page load |
| Form Submission | Button shows loading spinner, form inputs disabled |

### Dirty Form Behavior

| Scenario | Behavior |
|----------|----------|
| User modifies field | Form marked as "dirty" |
| Navigate away (dirty) | Confirmation dialog: "Discard unsaved changes?" with Cancel and Discard buttons |
| Navigate away (clean) | Immediate navigation, no confirmation |
| Save successful | Form marked as "clean", value becomes new baseline |

### Accessibility

| Feature | Implementation |
|---------|---------------|
| Time Picker | Keyboard accessible (Tab to focus, type or arrow keys to change) |
| Help Text | Associated with field via aria-describedby |
| Error Messages | Announced to screen readers via aria-live region |
| Save Button | Clear focus indicator |
| Form Labels | Properly associated with inputs via for/id attributes |

---

## 8. Additional Information

### Out of Scope

- Per-department or per-employee threshold configuration
- Multiple thresholds for different days of the week
- Grace period configuration (e.g., 5-minute buffer before marking late)
- Automatic notifications when employees are late
- Threshold change history/audit log
- Bulk threshold updates via import

### Dependencies

| Dependency | Source | Relationship |
|------------|--------|--------------|
| User authentication | EP-001 US-001 | Required for page access |
| Permission system | EP-001 US-004 | Controls who can modify settings |
| Attendance data | MOBILE-APP EP-003 | Check-in times compared against threshold |
| Attendance List | EP-004 DR-004-001-01 | Displays late status based on threshold |

### Open Questions

None - all questions resolved from Figma context and knowledge base.

### Related Features

- **EP-004 DR-004-001-01 (Attendance List):** Displays attendance matrix with Late (L) status based on this threshold
- **MOBILE-APP EP-003 US-001 (Check-In/Out):** Records check-in times that are compared against threshold
- **EP-009 US-001 (Company Profile):** Sibling story in Organization Settings epic

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | [Name] | [Date] | Pending |
| Product Owner | [Name] | [Date] | Pending |
| UX Designer | [Name] | [Date] | Pending |
| Tech Lead | [Name] | [Date] | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-04-23 | BA Team | Initial draft (created as DR-004-001-02) |
| 1.1 | 2026-04-23 | BA Team | Relocated to EP-009/US-002 as DR-009-002-01 |
