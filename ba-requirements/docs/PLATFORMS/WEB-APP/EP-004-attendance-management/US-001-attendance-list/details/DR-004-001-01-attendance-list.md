---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-004
story_id: US-001
story_name: "Attendance List"
detail_id: DR-004-001-01
detail_name: "Attendance List"
parent_requirement: FR-US-001-01
status: draft
version: "1.2"
created_date: 2026-04-21
last_updated: 2026-05-08
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../../../../MOBILE-APP/EP-003-attendance-management/US-001-daily-attendance/details/DR-003-001-01-check-in-out.md"
    relationship: data-source
input_sources:
  - type: text
    description: "Brainstorming session — calendar matrix design decisions"
    extraction_date: "2026-04-21"
---

# Detail Requirement: Attendance List

**Detail ID:** DR-004-001-01
**Parent Requirement:** FR-US-001-01
**Story:** US-001-attendance-list
**Epic:** EP-004 (Attendance Management)
**Status:** Draft
**Version:** 1.2

---

## 1. Use Case Description

As an **HR administrator**, I want to **view a monthly attendance matrix showing all employees' attendance status in a calendar-style layout**, so that **I can monitor attendance patterns, identify late arrivals and absences, and export data for reporting**.

As an **employee without management permission**, I want to **view my own attendance in the same matrix format**, so that **I can verify my check-in/out records and track my attendance history**.

**Purpose:** Provide a centralized monthly view of employee attendance data captured via the mobile app (MOBILE-APP EP-003). The calendar matrix format allows HR to quickly scan attendance patterns across the organization, with visual indicators for on-time arrivals, late arrivals, absences, and approved leave. This view integrates attendance data with leave records from EP-002 Leave Management.

**Target Users:**
- **HR Administrators** — Primary users for attendance monitoring, pattern analysis, and report generation
- **Administrators** — Full access for system administration and data export
- **CEO** — Overview access for company-wide attendance patterns
- **Leaders** — Department-level attendance visibility (with appropriate permissions)
- **Employees** — View own attendance only (without "Manage Data" permission)

**Key Functionality:**
- Monthly calendar matrix view (rows = employees, columns = dates 1-31)
- Visual status indicators per cell (✓ On-time, L Late, A Absent, — Weekend, H Holiday, AL/SL/½ Leave)
- **Combined cells:** when a date has half-day leave AND a check-in/out record, the cell renders as a diagonal split (leave color + `½` on one half, attendance color + status glyph on the other)
- Tooltip on hover showing check-in/out times, hours worked, or leave details (combined cells use a two-section tooltip — leave block + worked-half attendance block)
- **Monthly summary columns:** Total Late Time and Total Early Time per employee
- Filters: Month/Year picker, Department, Status, Employee search
- Export to Excel (bulk all employees + individual per-row export)
- Permission-based data visibility: employees see own row only unless they have "Manage Data" permission

---

## 2. User Workflow

**Entry Point:** Sidebar navigation → HRM > Attendance

**Preconditions:**
- User is signed in (EP-001 US-001)
- User's role has attendance view permission (EP-001 US-004)

**Main Flow:**
1. User clicks "Attendance" under HRM section in the sidebar
2. System loads the Attendance List page
3. System applies data visibility scoping based on permissions
4. System displays the monthly attendance matrix for the current month
5. System loads attendance data from check-in/out records (MOBILE-APP EP-003)
6. System loads approved leave data from Leave Management (EP-002)
7. User browses or takes one of the available actions

**Alternative Flows:**

| Action | Flow |
|--------|------|
| **Change Month** | User selects different month/year from date picker → matrix reloads with selected month's data |
| **Search** | User types employee name → list filters with debounce (300ms) → shows matching employees only |
| **Clear Search** | User clears search box → full list restored (respecting active filters and permissions) |
| **Filter: Department** | User clicks Department chip → selects one or more → matrix filters to show only employees in selected department(s) |
| **Filter: Status** | User clicks Status chip → selects On-time/Late/Absent/On Leave → shows employees with at least one day matching selected status(es) |
| **Reset** | User clicks Reset → all filters and search cleared → full matrix (within data visibility scope) |
| **Hover Cell** | User hovers over a cell → tooltip appears showing date, check-in/out times, hours, status |
| **Export All** | User clicks Export → downloads Excel file of all visible rows |
| **Export Individual** | User clicks gear icon → "Export" → downloads Excel file for that employee's month |

**Exit Points:**
- **Export** → File downloads; stays on page
- **Sidebar navigation** → Navigate to other pages

---

## 3. Field Definitions

### Filter Elements

| Filter Name | Type | Options Source | Multi-select | Default | Description |
|-------------|------|----------------|--------------|---------|-------------|
| Month/Year | Date picker | Calendar months | No | Current month | Select which month to view |
| Department | Dropdown chip | Departments from EP-008 US-001 | Yes | None (show all) | With in-dropdown search |
| Status | Dropdown chip | On-time, Late, Absent, On Leave | Yes | None (show all) | Shows employees with at least one day matching |
| Search | Text input | — | — | Empty | Employee name search with 300ms debounce |

### Interaction Elements

| Element | Type | Visible To | State/Condition | Trigger Action | Description |
|---------|------|------------|-----------------|----------------|-------------|
| Reset | Icon button | All with view permission | Visible when any filter/search active | Clears all filters and search | Refresh icon |
| Export | Button (primary) | All with view permission | Always visible | Downloads Excel of all visible rows | Top-right area |
| Gear Icon | Icon button | All with view permission | Per employee row | Opens action menu | Last column |
| Gear → Export | Menu item | All with view permission | Always | Downloads Excel for that employee | Individual export |

---

## 4. Data Display

### Matrix Layout

```
Filters: [April 2026 ▼] [Department ▼] [Status ▼] [Search...]                          [Export]

             | Mon | Tue | Wed | Thu | Fri | Sat | Sun | Mon | ... | Wed | Total Late | Total Early |
             |  1  |  2  |  3  |  4  |  5  |  6  |  7  |  8  | ... | 30  |    Time    |    Time     | ⚙ |
-------------|-----|-----|-----|-----|-----|-----|-----|-----|-----|-----|------------|-------------|---|
John Doe     |  ✓  |  L  |  ✓  |  ✓  |  ✓  |  —  |  —  | ½◣✓ | ... |  ✓  |   0h 18m   |    0h 00m   | ⚙ |
Jane Smith   |  ✓  |  ✓  |  A  |  ✓  |  ✓  |  —  |  —  |  ✓  | ... |  ✓  |   0h 00m   |    0h 25m   | ⚙ |
Bob Wilson   |  L  |  ✓  |  ✓  | SL  | SL  |  —  |  —  |  ✓  | ... |  ✓  |   1h 12m   |    0h 40m   | ⚙ |

(`½◣✓` denotes the combined-cell diagonal split: AM half on Annual Leave + PM half worked on-time. See "Combined Cells" subsection below.)
```

The two summary columns sit at the right edge of the date grid, **before** the actions (gear) column. They are sticky-right so the totals remain visible when the user horizontally scrolls through the date columns.

### Column Definitions

| Column | Width | Content | Sort | Description |
|--------|-------|---------|------|-------------|
| Employee Name | 180px | Full name | Alphabetical (default) | Fixed column (doesn't scroll horizontally) |
| Date 1-31 | 44px each | Status icon | — | One column per day of the month |
| Total Late Time | 110px | Duration `Xh Ym` | Yes (descending default when active) | Monthly sum of minutes the employee was late checking in. Sticky-right (does not scroll with date columns) |
| Total Early Time | 110px | Duration `Xh Ym` | Yes (descending default when active) | Monthly sum of minutes the employee left early (check-out before scheduled end-of-day). Sticky-right |
| Actions (⚙) | 48px | Gear icon | — | Last column, fixed |

### Monthly Summary Columns — Computation

The full computation rules live in **SR-011** (handles full-day-worked, AM-half-worked, and PM-half-worked cases). Quick reference:

| Day type | Total Late Time per-day contribution | Total Early Time per-day contribution |
|---|---|---|
| Full-day worked | `max(0, first_check_in − 09:00)` | `max(0, 18:00 − last_check_out)` |
| AM half worked + PM half on leave | `max(0, first_check_in − 09:00)` | `max(0, 12:00 − last_check_out)` |
| AM half on leave + PM half worked | `max(0, first_check_in − 13:15)` | `max(0, 18:00 − last_check_out)` |
| Full-day Leave (AL / SL / full-day ½) | 0 | 0 |
| Weekend, Holiday | 0 | 0 |
| Absent (no check-in / no check-out) | 0 | 0 |

**Display rules:**
- Format: `Xh Ym` (e.g., `0h 18m`, `1h 45m`); zero displays as `0h 00m`
- For employees with no late or early occurrences in the month: show `0h 00m` (do not hide or dash)
- Re-computed whenever the month/filter changes; respects the same data visibility scope as the matrix rows
- Aggregation is summed across all dates in the selected month
- See SR-002 for late thresholds, SR-011 for the full computation specification

### Cell Status Icons

| Icon | Meaning | Cell Background | Text Color |
|------|---------|-----------------|------------|
| ✓ | Present (on-time) | Green (#dcfce7) | Green (#166534) |
| L | Late (check-in > 9:00 AM for AM half, or > 1:15 PM for PM half) | Amber (#fef3c7) | Amber (#92400e) |
| A | Absent (no check-in on workday) | Red (#fee2e2) | Red (#991b1b) |
| — | Weekend (Sat/Sun) | Gray (#f3f4f6) | Gray (#6b7280) |
| H | Holiday (company-defined) | Blue (#dbeafe) | Blue (#1e40af) |
| AL | Annual Leave (full day) | Purple (#f3e8ff) | Purple (#7c3aed) |
| SL | Sick Leave (full day) | Orange (#ffedd5) | Orange (#c2410c) |
| ½ | Half-day Leave (no check-in on the worked half OR worked half not yet recorded) | Light blue (#e0f2fe) | Blue (#0369a1) |

### Combined Cells (half-day leave + check-in/out on the same day)

When a date has approved half-day leave AND a check-in/out record, the cell renders as a **diagonal split** rather than a single icon:

| Combination | Visual | Description |
|---|---|---|
| AM Half Leave + PM worked on-time | `½` (top-left, light blue) ╲ `✓` (bottom-right, green) | Diagonal split — leave half top-left, worked half bottom-right |
| AM Half Leave + PM worked late | `½` (top-left, light blue) ╲ `L` (bottom-right, amber) | Same split, worked half shows late glyph |
| AM Half Leave + PM no check-in | `½` (top-left, light blue) ╲ `A` (bottom-right, red) | Worked half is Absent for the second half |
| AM worked on-time + PM Half Leave | `✓` (top-left, green) ╲ `½` (bottom-right, light blue) | Diagonal mirrored — worked half top-left, leave half bottom-right |
| AM worked late + PM Half Leave | `L` (top-left, amber) ╲ `½` (bottom-right, light blue) | Same mirrored split, AM late glyph |
| AM no check-in + PM Half Leave | `A` (top-left, red) ╲ `½` (bottom-right, light blue) | AM half is Absent |

**Rendering rules:**
- Cell is split by a 135° diagonal (top-left ↔ bottom-right)
- Each half occupies the corner that mirrors its time-of-day: **AM = top-left**, **PM = bottom-right**
- Each corner uses its own background color (leave or attendance) and its own glyph in its own text color
- Glyph alignment: top-left corner glyph anchored top-left with small inset; bottom-right corner glyph anchored bottom-right with small inset
- A 1px subtle divider runs along the diagonal at WCAG-AA contrast against both halves
- The Sick-leave (`SL`) and Annual-leave (`AL`) labels are not used on combined cells — the leave half always shows `½` regardless of leave type. Specific leave type appears in the tooltip.

### Tooltip Content

**For attendance days (✓, L, A):**

```
┌─────────────────────────────┐
│ Monday, April 7, 2026       │
│ ─────────────────────────── │
│ Check-in:   08:47 AM        │
│ Check-out:  06:15 PM        │
│ Hours:      9h 28m          │
│ Status:     On-time ✓       │
└─────────────────────────────┘
```

**For absent days:**

```
┌─────────────────────────────┐
│ Wednesday, April 3, 2026    │
│ ─────────────────────────── │
│ No check-in recorded        │
│ Status:     Absent          │
└─────────────────────────────┘
```

**For leave days (AL, SL, ½):**

```
┌─────────────────────────────┐
│ Tuesday, April 8, 2026      │
│ ─────────────────────────── │
│ Annual Leave (Full Day)     │
│ Approved by: John Manager   │
└─────────────────────────────┘
```

**For weekends/holidays:**

```
┌─────────────────────────────┐
│ Saturday, April 6, 2026     │
│ ─────────────────────────── │
│ Weekend                     │
└─────────────────────────────┘
```

**For combined cells (half-day leave + worked half):**

```
┌─────────────────────────────┐
│ Tuesday, April 8, 2026      │
│ ─────────────────────────── │
│ Morning Half — Annual Leave │
│ Approved by: John Manager   │
│ ─────────────────────────── │
│ Afternoon — Worked          │
│ Check-in:   13:05 PM        │
│ Check-out:  18:00 PM        │
│ Hours:      4h 55m          │
│ Status:     On-time ✓       │
└─────────────────────────────┘
```

**For combined cells where the worked half has no check-in:**

```
┌─────────────────────────────┐
│ Tuesday, April 8, 2026      │
│ ─────────────────────────── │
│ Morning Half — Annual Leave │
│ Approved by: John Manager   │
│ ─────────────────────────── │
│ Afternoon — Absent          │
│ No check-in recorded        │
└─────────────────────────────┘
```

### Display States

| State | Condition | Display |
|-------|-----------|---------|
| Loading | Data fetching | Skeleton rows for employee names, skeleton cells for dates |
| Empty - No Employees | No employees match filters | "No employees found" message |
| Empty - No Data | Employees exist but no attendance data for month | Matrix shows with all cells as "—" or "A" depending on workday |
| Data Loaded | Normal state | Full matrix with status indicators |

---

## 5. Acceptance Criteria

### AC-001: View Monthly Attendance Matrix

**Given** I am logged in with "View Attendance" + "Manage Data" permissions
**When** I navigate to the Attendance List page
**Then** I see a monthly matrix with all employees and their attendance status for the current month

### AC-002: View Own Attendance Only

**Given** I am logged in with "View Attendance" permission only (no "Manage Data")
**When** I navigate to the Attendance List page
**Then** I see only my own attendance row in the matrix

### AC-003: Change Month

**Given** I am on the Attendance List page showing April 2026
**When** I select "March 2026" from the month picker
**Then** the matrix reloads showing March 2026 attendance data

### AC-004: Filter by Department

**Given** I am on the Attendance List page
**When** I select "Engineering" from the Department filter
**Then** only employees in the Engineering department are displayed

### AC-005: Filter by Status

**Given** I am on the Attendance List page
**When** I select "Late" from the Status filter
**Then** only employees with at least one late day in the month are displayed

### AC-006: Combine Filters

**Given** I have "Engineering" department and "Late" status filters active
**When** I view the matrix
**Then** I see only Engineering employees who have at least one late day

### AC-007: Search by Employee Name

**Given** I am on the Attendance List page
**When** I type "John" in the search box
**Then** only employees whose name contains "John" are displayed

### AC-008: View Tooltip - Attendance Day

**Given** I am viewing the attendance matrix
**When** I hover over a cell showing "✓" for John Doe on April 7
**Then** I see a tooltip with: "Monday, April 7, 2026", check-in time, check-out time, hours worked, "On-time ✓"

### AC-009: View Tooltip - Leave Day

**Given** John Doe has approved Annual Leave on April 8
**When** I hover over April 8 cell for John Doe
**Then** I see a tooltip with: "Tuesday, April 8, 2026", "Annual Leave (Full Day)", "Approved by: [manager name]"

### AC-010: View Tooltip - Weekend

**Given** April 6 is a Saturday
**When** I hover over April 6 cell for any employee
**Then** I see a tooltip with: "Saturday, April 6, 2026", "Weekend"

### AC-011: Export All Employees

**Given** I have filtered the list to show Engineering department (5 employees)
**When** I click the "Export" button
**Then** an Excel file downloads containing only the 5 Engineering employees' attendance for the month

### AC-012: Export Individual Employee

**Given** I am viewing the attendance matrix
**When** I click the gear icon on John Doe's row and select "Export"
**Then** an Excel file downloads containing only John Doe's attendance for the selected month

### AC-013: Late Status Display

**Given** an employee checked in at 9:15 AM on April 2
**When** I view the attendance matrix for April
**Then** April 2 cell shows "L" with amber background

### AC-014: Absent Status Display

**Given** an employee has no check-in record on April 3 (a workday)
**When** I view the attendance matrix for April
**Then** April 3 cell shows "A" with red background

### AC-015: Weekend Display

**Given** April 6 is a Saturday
**When** I view the attendance matrix for April
**Then** April 6 column shows "—" with gray muted background for all employees

### AC-016: Leave Integration

**Given** an employee has approved Sick Leave on April 4-5
**When** I view the April attendance matrix
**Then** April 4 and April 5 cells show "SL" with orange background for that employee

### AC-017: Reset Filters

**Given** I have Department and Status filters active
**When** I click the Reset button
**Then** all filters are cleared and the full list is displayed

### AC-018: Cell Click No Action

**Given** I am viewing the attendance matrix
**When** I click on a cell (not hover)
**Then** nothing happens (no navigation, no modal)

### AC-019: Total Late Time Column

**Given** John Doe checked in at 09:10 on April 2 (10 minutes late) and at 09:08 on April 8 (8 minutes late), and was on-time every other workday in April
**When** I view the April attendance matrix
**Then** John Doe's "Total Late Time" column shows `0h 18m`

### AC-020: Total Early Time Column

**Given** the scheduled end-of-day is 18:00 and Jane Smith checked out at 17:35 on April 4 (25 minutes early), and was not early on any other day in April
**When** I view the April attendance matrix
**Then** Jane Smith's "Total Early Time" column shows `0h 25m`

### AC-021: Summary Columns — Zero Case

**Given** an employee was on-time and never left early in the selected month
**When** I view the matrix
**Then** both "Total Late Time" and "Total Early Time" display `0h 00m` (not blank, not `—`)

### AC-022: Summary Columns Exclude Non-Working Days

**Given** an employee was on Annual Leave for the entire week of April 6-10, with weekends on April 4-5 and April 11-12
**When** I view the April attendance matrix
**Then** none of those Leave/Weekend days contribute to the Total Late Time or Total Early Time totals

### AC-023: Summary Columns Recompute on Month Change

**Given** I am viewing March 2026 with non-zero Total Late Time / Total Early Time values for John Doe
**When** I switch the month picker to April 2026
**Then** the two summary columns recompute and display April's totals (which may differ from March)

### AC-024: Summary Columns Sticky on Horizontal Scroll

**Given** I am viewing the matrix and the date columns overflow horizontally
**When** I scroll right through the date columns
**Then** the "Total Late Time" and "Total Early Time" columns remain visible at the right edge of the matrix (sticky-right), so the totals stay in view alongside the gear icon

### AC-025: Summary Columns in Excel Export

**Given** I trigger Export (bulk or individual)
**When** the Excel file is generated
**Then** the file includes "Total Late Time" and "Total Early Time" columns with the same `Xh Ym` formatted values shown in the UI

### AC-026: Combined Cell — AM leave + PM worked

**Given** John Doe has approved Morning Half Annual Leave on April 8 and checked in at 13:05 PM, checked out at 18:00 PM
**When** I view the April attendance matrix
**Then** April 8 cell for John Doe shows a diagonal-split cell with `½` in the top-left (light blue) and `✓` in the bottom-right (green)

### AC-027: Combined Cell — Tooltip Content

**Given** John Doe has approved Morning Half Annual Leave on April 8 (approved by John Manager) and worked the PM half (in 13:05, out 18:00, on-time)
**When** I hover over the April 8 combined cell for John Doe
**Then** I see a tooltip with two stacked sections: top section "Morning Half — Annual Leave" + "Approved by: John Manager"; bottom section "Afternoon — Worked" + check-in 13:05 + check-out 18:00 + hours 4h 55m + "Status: On-time ✓"

### AC-028: Combined Cell — PM Worked Late Math

**Given** Jane Smith has approved Morning Half Sick Leave on April 9 and checked in at 13:25 PM (10 minutes after the PM threshold of 13:15)
**When** I view the April attendance matrix
**Then** April 9 cell for Jane Smith shows `½` (top-left, light blue) and `L` (bottom-right, amber); and her "Total Late Time" column for April includes 10 minutes from this day

### AC-029: Combined Cell — AM Worked Early Math

**Given** Bob Wilson has approved Afternoon Half Annual Leave on April 10 and checked in at 09:00, checked out at 11:50 (10 minutes before the AM end of 12:00)
**When** I view the April attendance matrix
**Then** April 10 cell for Bob Wilson shows `✓` (top-left, green) and `½` (bottom-right, light blue); and his "Total Early Time" column for April includes 10 minutes from this day

### AC-030: Combined Cell — No Check-in on Worked Half

**Given** an employee has approved Morning Half Annual Leave on April 11 and never checked in for the PM half
**When** I view the April attendance matrix
**Then** April 11 cell shows `½` (top-left, light blue) and `A` (bottom-right, red); the tooltip's worked-half section reads "Afternoon — Absent / No check-in recorded"; and the day contributes 0 to both Total Late Time and Total Early Time (consistent with full-day Absent rule)

### AC-031: Combined Cell — Status Filter Multi-Match

**Given** an employee has approved Morning Half Annual Leave + worked PM half late on April 12
**When** I filter Status by "Late"
**Then** the employee row appears in results
**And When** I instead filter Status by "On Leave"
**Then** the same employee row appears in results
**Because** a combined-cell day matches multiple status values simultaneously

---

## 6. System Rules

### SR-001: Data Source

Attendance data is retrieved from MOBILE-APP EP-003 check-in/out records. This view is read-only — no attendance data can be modified from this page.

### SR-002: Late Thresholds

An employee is marked as "Late" if their first check-in of the worked half is after the late threshold for that half:

| Worked half | Late threshold |
|---|---|
| AM half (full day or AM-half-worked) | 09:00 |
| PM half (PM-half-worked, when AM half is on approved leave) | 13:15 |

Lunch break: **12:00 — 13:15**. Workday end: **18:00**. Thresholds and workday boundaries are configured in the system (MOBILE-APP EP-003).

### SR-003: Absent Logic

- **Full day absent:** An employee is marked as "Absent" for a full workday (Monday-Friday, excluding holidays) if they have no check-in record for that day AND no approved leave.
- **Worked half absent:** When the day has approved half-day leave AND the worked half has no check-in record, the worked half is rendered as `A` (Absent) in the combined-cell diagonal split. The day still counts as 0.5 days of approved leave.

### SR-004: Leave Integration

Leave data is retrieved from EP-002 Leave Management. Only approved leave requests are displayed in the attendance matrix. Leave types displayed: Annual Leave (AL), Sick Leave (SL), Half-day (½).

**Half-day leave + check-in coexistence (combined cells):**
- When an approved half-day leave (Morning Half or Afternoon Half) coexists with a check-in/out record for the same date, the cell renders as a diagonal-split combined cell (see §4 Combined Cells)
- Leave half occupies the diagonal corner mirroring the time-of-day (AM → top-left, PM → bottom-right)
- Worked half occupies the opposite corner with the appropriate attendance glyph
- The leave half always uses the generic `½` glyph in the combined cell regardless of underlying leave type (Annual, Sick, etc.); the specific leave type appears in the tooltip
- The worked half is evaluated against its own half's late threshold (per SR-002) and end-of-half boundary (per SR-011)

### SR-005: Weekend Identification

Saturdays and Sundays are marked as weekends ("—") regardless of check-in data. Weekend cells use gray muted styling.

### SR-006: Holiday Identification

Company-defined holidays are marked as "H" with blue muted styling. Holiday calendar is maintained in system configuration.

### SR-007: Permission-Based Visibility

| Permission | Data Visibility |
|------------|-----------------|
| View Attendance only | Own row only |
| View Attendance + Manage Data | All employees |

### SR-008: Filter Logic

- **Status filter:** Shows employees with at least one day matching selected status(es) in the month
- **Department filter:** OR logic within (shows employees in any selected department)
- **Combined filters:** AND logic across different filter types
- **Search:** AND with all active filters
- **Combined-cell multi-match:** A day with half-day leave + worked half matches multiple status values simultaneously. Filtering by "On Leave" matches because of the leave half; filtering by "Late"/"On-time"/"Absent" matches if the worked half meets that condition. The same day can therefore surface the row under several status filter selections.

### SR-009: Export Format

All exports are in Excel (.xlsx) format. Export includes:
- Employee name
- All dates in the selected month
- Status per date
- Check-in/out times (if applicable)
- Hours worked (if applicable)
- Total Late Time (monthly sum, formatted `Xh Ym`)
- Total Early Time (monthly sum, formatted `Xh Ym`)

### SR-010: Multiple Sessions Per Day

If an employee has multiple check-in/out sessions in a day (allowed by MOBILE-APP EP-003), the cell status is based on the **first check-in** of the day. The tooltip shows all sessions.

### SR-011: Monthly Summary Columns

The "Total Late Time" and "Total Early Time" columns are computed per employee per selected month, with **per-day contribution** depending on whether the day is full-worked or half-day-with-worked-half.

**Full-day worked (no half-day leave on the day):**

| Column | Per-day contribution |
|--------|---------------------|
| Total Late Time | `max(0, first_check_in − 09:00)` |
| Total Early Time | `max(0, 18:00 − last_check_out)` |

**Half-day worked (combined-cell day):** use the worked half's boundaries from SR-002, applied only to the worked half's check-in/out:

| Worked half | Total Late Time contribution | Total Early Time contribution |
|---|---|---|
| AM half worked + PM half on leave | `max(0, first_check_in − 09:00)` | `max(0, 12:00 − last_check_out)` |
| AM half on leave + PM half worked | `max(0, first_check_in − 13:15)` | `max(0, 18:00 − last_check_out)` |

The leave half always contributes 0 to both summaries.

**Aggregation:** Sum across all dates in the selected month.

**Rules:**
- For days with multiple sessions, "first check-in" and "last check-out" of the worked half are used (consistent with SR-010)
- Non-working days (Weekend, Holiday) contribute 0
- Full-day approved Leave days (AL, SL, full-day ½) contribute 0 — leave is not counted as late/early
- Combined-cell days with the worked half having no check-in (worked half = `A`) contribute 0 to both summaries (consistent with full-day Absent rule)
- Absent days contribute 0 to both summaries
- Display format: `Xh Ym` (zero shown as `0h 00m`); negative differences are clipped to 0
- Sticky-right column position so totals remain visible when the matrix scrolls horizontally
- Recomputed on every month change and on every filter change that affects the row's data visibility scope
- Excel export (SR-009) includes these two columns

**Schedule constants (confirmed):** workday 09:00 — 18:00; lunch break 12:00 — 13:15; AM late threshold 09:00; PM late threshold (when PM is the worked half) 13:15.

---

## 7. UX Optimizations

### Layout

| Item | Specification |
|------|--------------|
| Employee column | Fixed position (doesn't scroll with dates) |
| Date columns | Horizontally scrollable |
| Column header | Shows day of week (Mon, Tue...) + date number |
| Row height | 44px for comfortable touch targets |
| Cell width | 44px minimum for date columns |

### Loading States

| Element | Loading State |
|---------|--------------|
| Employee names | Skeleton text (180px width) |
| Date cells | Skeleton rectangles |
| Filters | Disabled during load |

### Responsiveness

- On smaller screens, the employee name column remains fixed while date columns scroll horizontally
- Month picker and filters stack vertically on mobile widths
- Minimum supported width: 768px

### Accessibility

| Feature | Implementation |
|---------|---------------|
| Status icons | aria-label with full status text (e.g., "On-time", "Late", "Absent") |
| Tooltip | Accessible via keyboard focus (Tab to cell, Enter to show tooltip) |
| Color contrast | All status colors meet WCAG AA contrast requirements |
| Screen reader | Row headers announce employee name; column headers announce date |

---

## 8. Additional Information

### Out of Scope

- Manual attendance entry or modification
- Attendance policy configuration (late threshold, etc.)
- Real-time/live attendance updates
- Weekly or daily drill-down views
- Attendance analytics or trend charts
- Notifications or alerts
- Payroll integration

### Dependencies

| Dependency | Source | Data |
|------------|--------|------|
| Check-in/out records | MOBILE-APP EP-003 | Timestamps, late flags |
| Leave records | WEB-APP EP-002 | Approved leave dates, types, approvers |
| Department list | WEB-APP EP-008 | Filter options |
| User accounts | WEB-APP EP-001 | Employee list, permissions |
| Holiday calendar | System configuration | Holiday dates |

### Open Questions

| Question | Owner | Status |
|----------|-------|--------|
| Export format (confirmed Excel) | Product Owner | Resolved |
| Status badge exact color tokens | Design Team | Pending |
| Holiday calendar source/configuration | Product Owner | Pending |
| Scheduled end-of-day time used to compute Total Early Time | Product Owner | Resolved — 18:00 (v1.2) |
| PM half late threshold for half-day combined cells | Product Owner | Resolved — 13:15 (v1.2) |
| Whether Total Early Time should also subtract approved early-leave permission requests (so an authorised early departure does not count) | Product Owner | Pending |

### Related Features

- **MOBILE-APP EP-003 US-001:** Check-In/Out — data source for attendance records
- **WEB-APP EP-002 US-001:** Leave Requests — leave data integration
- **WEB-APP EP-008 US-001:** Department Management — department filter options

---

## Document Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-04-21 | Claude | Initial draft — calendar matrix layout, filters, tooltip, export, permission scoping |
| 1.1 | 2026-05-07 | Claude | Stakeholder feedback — added two monthly summary columns to the matrix: **Total Late Time** (sum of late check-in minutes) and **Total Early Time** (sum of early check-out minutes). Updated §1 Key Functionality, §4 Matrix Layout diagram + Column Definitions table + new "Monthly Summary Columns — Computation" subsection, §5 added AC-019..AC-025, §6 added SR-011 (computation rules) and extended SR-009 (export now includes the two summary columns), §8 Open Questions added two items (scheduled end-of-day time, treatment of approved early-leave permission). Columns are sticky-right so totals stay visible during horizontal scrolling. |
| 1.2 | 2026-05-08 | Claude | Brainstormed handling of half-day leave + check-in/out coexistence (design captured in `docs/superpowers/specs/2026-05-08-half-day-leave-attendance-cell-design.md`). Added: §1 mention of combined cells; §4 Cell Status Icons clarified late thresholds and full-day applicability of AL/SL/½; §4 new "Combined Cells" subsection with diagonal-split rendering rules (AM=top-left, PM=bottom-right); §4 Matrix Layout diagram updated to include a combined cell (`½◣✓`) on April 8; §4 two new tooltip blocks (combined cell, combined cell with no check-in on worked half); §5 new AC-026..AC-031 covering combined-cell rendering + tooltip + late math (PM-worked) + early math (AM-worked) + no-check-in edge case + status filter multi-match; §6 SR-002 expanded with both AM (09:00) and PM (13:15) late thresholds and confirmed lunch break + workday end; §6 SR-003 extended to define worked-half Absent within a combined cell; §6 SR-004 extended with combined-cell coexistence rules; §6 SR-008 documented status filter multi-match for combined-cell rows; §6 SR-011 split into full-day-worked vs half-day-worked computations; §8 Open Questions resolved scheduled end-of-day = 18:00 and PM half late threshold = 13:15. |
