---
document_type: REQUIREMENTS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-004
story_id: US-001
story_name: "Attendance List"
status: draft
version: "1.0"
last_updated: "2026-04-21"
related_documents:
  - path: "./ANALYSIS.md"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
---

# Requirements: Attendance List

**Epic:** EP-004 (Attendance Management)
**Story:** US-001-attendance-list
**Status:** Draft

---

## User Story

**As an** HR administrator,
**I want to** view a monthly attendance matrix showing all employees' attendance status,
**So that** I can monitor attendance patterns, identify late arrivals and absences, and export data for reporting.

---

## Functional Requirements

### FR-US-001-01: Monthly Calendar Matrix Layout

The system shall display attendance data in a calendar-style matrix where:
- Rows represent employees (employee name as row header)
- Columns represent dates (1-31 for the selected month)
- Each cell shows the attendance status for that employee on that date

### FR-US-001-02: Attendance Status Display

Each cell shall display a status indicator:

| Status | Icon | Cell Color | Meaning |
|--------|------|------------|---------|
| On-time | ✓ | Green | Check-in ≤ 9:00 AM |
| Late | L | Amber | Check-in > 9:00 AM |
| Absent | A | Red | No check-in on workday |
| Weekend | — | Gray (muted) | Saturday/Sunday |
| Holiday | H | Blue (muted) | Company holiday |
| Annual Leave | AL | Purple | Approved annual leave |
| Sick Leave | SL | Orange | Approved sick leave |
| Half-day | ½ | Light blue | Approved half-day leave |

### FR-US-001-03: Tooltip on Hover

When hovering over a cell, the system shall display a tooltip containing:

**For attendance days:**
- Full date (e.g., "Monday, April 7, 2026")
- Check-in time
- Check-out time
- Hours worked
- Status label

**For leave days:**
- Full date
- Leave type (e.g., "Annual Leave (Full Day)")
- Approved by (manager name)

### FR-US-001-04: Month/Year Filter

The system shall provide a date picker to select the month and year to view.
- Default: current month
- Changing the month reloads the attendance data for the selected period

### FR-US-001-05: Department Filter

The system shall provide a multi-select dropdown to filter by department(s).
- Options populated from EP-008 department list
- Multiple departments can be selected (OR logic within filter)
- Chip display shows count when multiple selected

### FR-US-001-06: Status Filter

The system shall provide a multi-select dropdown to filter by attendance status:
- On-time
- Late
- Absent
- On Leave

Filter shows employees with **at least one day** matching selected status(es) in the month.

### FR-US-001-07: Employee Search

The system shall provide a text search field for employee name.
- 300ms debounce before search triggers
- Case-insensitive, partial match
- Searches employee full name

### FR-US-001-08: Bulk Export

The system shall provide an "Export" button that exports all visible rows (respecting active filters) to Excel (.xlsx) format.

### FR-US-001-09: Individual Export

Each employee row shall have a gear icon with an "Export" action that exports that employee's monthly attendance to Excel (.xlsx).

### FR-US-001-10: Weekend Display

Weekend columns (Saturday, Sunday) shall be displayed with:
- Gray/muted background color
- Same column width as workdays
- "—" status indicator

### FR-US-001-11: Permission-Based Data Visibility

- Users with "View Attendance" permission only: see their own attendance row
- Users with "View Attendance" + "Manage Data" permission: see all employees

### FR-US-001-12: No Cell Click Action

Clicking on a cell shall have no action. Interaction is hover-only (tooltip display).

---

## Non-Functional Requirements

### NFR-US-001-01: Performance

The attendance matrix shall load within 3 seconds for up to 100 employees.

### NFR-US-001-02: Responsiveness

The table shall be horizontally scrollable on smaller screens while keeping the employee name column fixed.

### NFR-US-001-03: Accessibility

- Tooltip content shall be accessible via keyboard focus
- Status colors shall have sufficient contrast (WCAG AA)
- Status icons shall have aria-labels for screen readers

---

## Acceptance Criteria

### AC-001: View Monthly Attendance

**Given** I am logged in with "View Attendance" + "Manage Data" permissions
**When** I navigate to the Attendance List page
**Then** I see a monthly matrix with all employees and their attendance status for the current month

### AC-002: Filter by Department

**Given** I am on the Attendance List page
**When** I select "Engineering" from the Department filter
**Then** only employees in the Engineering department are displayed

### AC-003: Filter by Status

**Given** I am on the Attendance List page
**When** I select "Late" from the Status filter
**Then** only employees with at least one late day in the month are displayed

### AC-004: View Tooltip

**Given** I am viewing the attendance matrix
**When** I hover over a cell showing "✓"
**Then** I see a tooltip with the date, check-in time, check-out time, hours worked, and status

### AC-005: Export All

**Given** I have filtered the list to show Engineering department
**When** I click the "Export" button
**Then** an Excel file downloads containing only Engineering employees' attendance

### AC-006: Export Individual

**Given** I am viewing the attendance matrix
**When** I click the gear icon on John Doe's row and select "Export"
**Then** an Excel file downloads containing only John Doe's attendance for the month

### AC-007: Own Attendance Only

**Given** I am logged in with "View Attendance" permission only (no "Manage Data")
**When** I navigate to the Attendance List page
**Then** I see only my own attendance row

### AC-008: Leave Integration

**Given** an employee has approved Annual Leave on April 8
**When** I view the April attendance matrix
**Then** April 8 shows "AL" with purple background for that employee
