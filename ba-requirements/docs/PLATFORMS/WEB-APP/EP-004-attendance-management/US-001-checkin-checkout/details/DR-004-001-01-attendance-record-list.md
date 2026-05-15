---
document_type: DETAIL_REQUIREMENT
platform: web-app
platform_display: "Web App"
epic_id: EP-004
story_id: US-001
story_name: "Check-in/Check-out"
detail_id: DR-004-001-01
detail_name: "Attendance Record List"
status: draft
version: "1.0"
created_date: 2026-04-23
last_updated: 2026-04-23
related_documents:
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "../../EPIC.md"
    relationship: parent
input_sources:
  - type: figma
    description: "Attendance Record List screen design"
---

# Detail Requirement: Attendance Record List

**Detail ID:** DR-004-001-01
**Parent Requirement:** FR-004-001-01
**Story:** US-001-checkin-checkout
**Epic:** EP-004 (Attendance Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **Admin/HR**, I want to **view and filter all employee attendance records** so that **I can monitor attendance patterns, verify work hours, and identify issues requiring attention**.

As an **Employee**, I want to **view my own attendance records** so that **I can track my check-in/check-out history and verify my work hours are recorded correctly**.

**Purpose:** Provide a centralized view of attendance data with filtering capabilities to support attendance monitoring, payroll preparation, and compliance tracking. The scope includes both viewing attendance records AND the ability to create/edit manual attendance entries.

**Target Users:**
- Admin/HR: Full access to all employee records with management capabilities
- Employee: View-only access to their own records

**Key Functionality:**
- View attendance records in a paginated table
- Filter by date range, employee, and status
- Search by employee name or ID
- Export filtered results to Excel
- Create manual attendance entries (Admin/HR only)
- Edit existing attendance records (Admin/HR only)

---

## 2. User Workflow

**Entry Point:** Sidebar menu "Attendance" > "Attendance Records" (or direct navigation)

**Preconditions:**
- User is authenticated and has appropriate role permissions
- At least one employee exists in the system (for Admin/HR view)

**Main Flow:**
1. User navigates to Attendance Records from sidebar menu
2. System displays attendance record list with default filters (current month, all statuses)
3. System shows skeleton loading state while fetching data
4. Table populates with attendance records (paginated, 10 per page default)
5. User can adjust filters (date range, employee, status) to narrow results
6. User can search by employee name/ID using the search box
7. User can sort columns by clicking column headers
8. User can export filtered results using the Export button
9. Admin/HR can click "+ Add Attendance" to create manual entry
10. Admin/HR can click gear icon to access Edit/Delete actions

**Alternative Flows:**
- **Alt 1 - No Records Found:** System displays empty state with message "No attendance records found for the selected filters"
- **Alt 2 - Employee View:** Employee sees only their own records; "+ Add Attendance" button and gear icon are hidden
- **Alt 3 - Filter Change:** When filters change, table resets to page 1 and reloads data

**Exit Points:**
- **Success:** User views, filters, or exports attendance data as needed
- **Navigation:** User clicks another menu item to leave the page
- **Error:** System displays error toast and retry option if data fetch fails

---

## 3. Field Definitions

### Filter Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Date Range | Date Picker (From-To) | From date <= To date; max range 3 months | No | Current month (1st to today) | Filter records by date range |
| Employee | Dropdown (searchable) | Valid employee from list | No | All Employees | Filter by specific employee (Admin/HR only) |
| Status | Multi-select Dropdown | Valid status values | No | All Statuses | Filter by attendance status |
| Search | Text Input | Min 2 characters to trigger search | No | Empty | Search by employee name or ID |

### Time Entry Fields (for Create/Edit)

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Employee | Dropdown (searchable) | Must select valid employee | Yes | None | Employee for the attendance record |
| Date | Date Picker | Cannot be future date; within allowed edit window | Yes | Today | Date of attendance |
| Check-in Time | Time Picker | Valid time format (HH:mm); free-form selection | No | None | Time employee checked in |
| Check-out Time | Time Picker | Must be after check-in time if both provided; free-form selection | No | None | Time employee checked out |
| Note | Textarea | Max 500 characters | No | Empty | Reason for manual entry/edit |

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| + Add Attendance | Button (Primary) | Visible only for Admin/HR | Opens Create Attendance modal | Create manual attendance entry |
| Export | Button (Secondary) | Always visible; disabled during export | Exports filtered data to Excel | Download attendance report |
| Search Box | Text Input | Always visible | Triggers search after 300ms debounce | Search by employee name/ID |
| Filter Dropdowns | Dropdown | Always visible | Apply filter on selection | Filter table results |
| Column Headers | Clickable | Sortable columns indicated | Toggle sort order | Sort table by column |
| Gear Icon | Icon Button | Visible only for Admin/HR | Opens action dropdown (Edit, Delete) | Row-level actions |
| Pagination | Component | Visible when records > page size | Navigate between pages | Page through results |

---

## 4. Data Display

### Table Columns

| Column Name | Data Type | Display When Empty | Format | Business Meaning |
|-------------|-----------|-------------------|--------|------------------|
| Employee Name | Text | -- | "[Last Name], [First Name]" | Identifies the employee |
| Employee ID | Text | -- | "EMP-XXXXX" | Unique employee identifier |
| Date | Date | -- | "DD/MM/YYYY" | Date of attendance record |
| Check-in Time | Time | "--:--" | "HH:mm" | Time employee started work |
| Check-out Time | Time | "--:--" | "HH:mm" | Time employee ended work |
| Work Hours | Number | "0h 0m" | "Xh Ym" | Calculated total work duration |
| Status | Badge | -- | Color-coded badge | Current status of the record |
| Source | Text | -- | "System" / "Manual" | How the record was created |
| Actions | Icon | -- | Gear icon | Edit/Delete options (Admin/HR only) |

### Status Badge Colors

| Status | Badge Color | Description |
|--------|-------------|-------------|
| Present | Green | Normal attendance with both check-in and check-out |
| Late | Yellow/Orange | Check-in after scheduled start time |
| Early Leave | Yellow/Orange | Check-out before scheduled end time |
| Absent | Red | No attendance record for scheduled work day |
| Incomplete | Gray | Only check-in recorded (no check-out) |
| Manual Entry | Blue | Record created/edited manually by Admin/HR |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Initial load or filter change | Skeleton loading rows (5 rows) |
| Empty | No records match filters | Empty state: "No attendance records found" with icon |
| Error | API request fails | Error message with "Retry" button |
| Success | Data loaded successfully | Populated table with records |
| Exporting | Export in progress | Export button shows loading spinner, disabled |

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** Admin/HR can view attendance records for all employees; Employee can view only their own records
- **AC-02:** Table displays all required columns: Employee Name, Employee ID, Date, Check-in Time, Check-out Time, Work Hours, Status, Source, Actions
- **AC-03:** Date range filter defaults to current month and limits selection to maximum 3 months
- **AC-04:** Search triggers after 300ms debounce and matches employee name or ID (partial match)
- **AC-05:** Multi-select status filter shows chips for selected values and supports in-dropdown search
- **AC-06:** Pagination defaults to 10 records per page with options for 10/25/50
- **AC-07:** Export generates Excel file containing all filtered results (not just current page)
- **AC-08:** "+ Add Attendance" button is visible only for Admin/HR roles
- **AC-09:** Gear icon with Edit/Delete actions is visible only for Admin/HR roles
- **AC-10:** Work Hours is calculated automatically as (Check-out Time - Check-in Time)
- **AC-11:** Status badge displays correct color based on status value
- **AC-12:** Time pickers allow free-form time selection (any valid time, e.g., 9:15 AM)

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| View all records (Admin) | Login as Admin, navigate to Attendance Records | See all employee records | High |
| View own records (Employee) | Login as Employee, navigate to Attendance Records | See only own records, no Add button | High |
| Filter by date range | Select "01/04/2026 - 15/04/2026" | Only records within range shown | High |
| Filter by status | Select "Late" and "Absent" | Only records with those statuses shown | High |
| Search by name | Type "Nguyen" | Records with "Nguyen" in name shown | High |
| Export filtered data | Apply filters, click Export | Excel file downloads with filtered data | Medium |
| Empty state | Filter to date range with no records | Empty state message displayed | Medium |
| Pagination navigation | Click page 2 | Second page of results shown | Medium |
| Sort by column | Click "Date" header | Records sorted by date | Medium |
| Time picker selection | Select 9:15 AM for check-in | Time displays as "09:15" | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **Rule 1:** Employee role can only view records where employee_id matches their own user record
- **Rule 2:** Admin/HR roles can view all employee attendance records
- **Rule 3:** Work Hours calculation: If both check-in and check-out exist, calculate difference; otherwise show "0h 0m"
- **Rule 4:** Status is determined automatically based on: scheduled work hours vs actual check-in/check-out times
- **Rule 5:** Manual entries are flagged with Source = "Manual" and retain the editing user's ID for audit
- **Rule 6:** Export respects all applied filters and includes all matching records (not paginated)
- **Rule 7:** Maximum date range for filtering is 3 months to prevent performance issues
- **Rule 8:** Search is case-insensitive and uses partial matching (LIKE '%term%')

**Status Determination Logic:**
```
IF no record exists for scheduled work day → Status = "Absent"
IF check_in_time > scheduled_start_time → Status = "Late"
IF check_out_time < scheduled_end_time AND check_out_time exists → Status = "Early Leave"
IF check_in_time exists AND check_out_time is NULL → Status = "Incomplete"
IF check_in_time <= scheduled_start_time AND check_out_time >= scheduled_end_time → Status = "Present"
IF record was manually created/edited → Add "Manual Entry" indicator
```

**Calculations/Formulas:**
- Work Hours: `(check_out_time - check_in_time)` displayed as "Xh Ym"
- Late Duration: `(check_in_time - scheduled_start_time)` if positive

**Dependencies:**
- Employee Management (US-004): Employee list for dropdown and filtering
- User Authentication: Role-based access control
- Work Schedule Configuration: Scheduled times for status determination

---

## 7. UX Optimizations

**Usability Considerations:**

- **Optimization 1:** Skeleton loading (5 rows) while data loads to reduce perceived wait time
- **Optimization 2:** 300ms debounce on search input to prevent excessive API calls
- **Optimization 3:** Filter chips display selected values for quick visibility
- **Optimization 4:** Sticky table header when scrolling for easier column reference
- **Optimization 5:** Export button shows loading spinner and disables during export process
- **Optimization 6:** Toast notification confirms successful export with filename

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Full table with all columns visible |
| Below Desktop | Out of scope for this DR |

**Accessibility Requirements:**
- [x] Keyboard navigable (Tab through filters, Enter to apply)
- [x] Screen reader compatible (ARIA labels on interactive elements)
- [x] Sufficient color contrast (WCAG AA compliance)
- [x] Focus indicators visible on all interactive elements
- [x] Status badges have text labels, not color-only indicators

**Design References:**
- Follows established List View Pattern from PROJECT_KNOWLEDGE.md
- Consistent with other list screens in the system (Employee List, Leave Requests)

---

## 8. Additional Information

### Out of Scope
- Mobile/tablet responsive layouts (desktop only for web admin)
- Bulk attendance entry or modification
- Real-time attendance updates (manual refresh required)
- Attendance reports with charts/analytics (separate feature)
- GPS/WiFi verification display (mobile app feature)
- Integration with biometric devices

### Scope Expansion Note
This DR includes Create and Edit functionality for manual attendance entries. The detailed Create/Edit modal specifications will be documented in:
- DR-004-001-02: Create Manual Attendance Entry
- DR-004-001-03: Edit Attendance Record

### Open Questions
- None (all questions resolved)

### Related Features
- DR-004-001-02: Create Manual Attendance Entry (to be created)
- DR-004-001-03: Edit Attendance Record (to be created)
- US-004: Employee Management (employee data source)
- EP-004: Attendance Management (parent epic)

### Notes
- Time picker uses standard free-form selection allowing any valid time (e.g., 9:15, 14:37)
- Work schedule configuration is assumed to exist for status calculation
- Audit trail for manual entries should be maintained for compliance

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
| 1.0 | 2026-04-23 | Claude | Initial draft |
