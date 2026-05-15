---
document_type: ANALYSIS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-004
story_id: US-001
story_name: "Attendance List"
status: draft
version: "1.0"
last_updated: "2026-04-21"
add_on_sections: []
approved_by: null
related_documents:
  - path: "./REQUIREMENTS.md"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
  - path: "../../EP-002-leave-management/US-001-leave-requests/ANALYSIS.md"
    relationship: reference
  - path: "../../../MOBILE-APP/EP-003-attendance-management/US-001-daily-attendance/details/DR-003-001-01-check-in-out.md"
    relationship: data-source
revision_history: []
input_sources:
  - type: text
    description: "Brainstorming session for attendance matrix design"
    extraction_date: "2026-04-21"
---

# Analysis: Attendance List

**Epic:** EP-004 (Attendance Management)
**Story:** US-001-attendance-list
**Status:** Draft

---

## 1. Business Context

### Problem Statement

HR and management need visibility into employee attendance patterns across the organization. With attendance data captured via mobile check-in/out (MOBILE-APP EP-003), there is no web-based interface for HR to view, analyze, and export this data. Without a centralized attendance view, identifying late arrivals, absences, and attendance trends requires manual data extraction.

### Stakeholders

- **Primary Users:** HR staff who monitor and report on attendance
- **Secondary Users:** Administrators, CEO, and team leaders who need attendance visibility
- **Data Source:** Employees using the mobile app for check-in/out (MOBILE-APP EP-003)
- **Business Owner:** HR department leadership

### Business Goals

- Provide HR with a clear monthly view of all employee attendance
- Enable quick identification of late arrivals and absences
- Integrate leave data to show complete employee availability picture
- Support attendance data export for reporting and payroll purposes

---

## 2. Scope Definition

### In Scope

- Monthly attendance matrix view (calendar-style: rows = employees, columns = dates)
- Attendance status display (on-time, late, absent, weekends, holidays, leave types)
- Tooltip details on hover (check-in/out times, hours worked, leave info)
- Filters: month/year, department, status, employee search
- Export to Excel (bulk all employees, individual employee)
- Permission-based data visibility (own row vs. all employees)

### Out of Scope

- Attendance data entry or modification (data comes from mobile check-in/out)
- Attendance policy configuration (late threshold, etc.)
- Real-time attendance monitoring / live updates
- Attendance summary reports or analytics dashboards
- Payroll integration
- Attendance alerts or notifications

---

## 3. User Personas

### HR Administrator

- **Goal:** Monitor company-wide attendance, identify patterns, generate reports
- **Access Level:** View all employees' attendance, export data
- **Key Actions:** Filter by department, identify late/absent employees, export monthly reports

### Team Leader

- **Goal:** Track team attendance for scheduling and resource planning
- **Access Level:** View department employees' attendance (with "Manage Data" permission)
- **Key Actions:** Filter by status to find absences, review team availability

### Employee

- **Goal:** View own attendance history
- **Access Level:** View only own attendance row (without "Manage Data" permission)
- **Key Actions:** Check personal attendance record, verify check-in times

---

## 4. Business Rules

### Data Source

- Attendance data originates from MOBILE-APP EP-003 (Check-In/Out)
- Each check-in records: timestamp, GPS coordinates, late flag
- Each check-out records: timestamp, auto-checkout flag
- Multiple sessions per day are supported

### Attendance Status Logic

| Status | Condition | Display |
|--------|-----------|---------|
| On-time (✓) | First check-in ≤ 9:00 AM on a workday | Green cell |
| Late (L) | First check-in > 9:00 AM on a workday | Amber cell |
| Absent (A) | No check-in on a workday | Red cell |
| Weekend (—) | Saturday or Sunday | Gray muted cell |
| Holiday (H) | Company-defined holiday | Blue muted cell |
| Annual Leave (AL) | Approved annual leave | Purple cell |
| Sick Leave (SL) | Approved sick leave | Orange cell |
| Half-day (½) | Approved half-day leave | Light blue cell |

### Permission-Based Visibility

- **View Attendance only:** User sees only their own attendance row
- **View Attendance + Manage Data:** User sees all employees' attendance

### Filter Logic

- Status filter shows employees with **at least one day** matching selected status(es) in the month
- Multiple filters combine with AND logic
- Search applies to employee name (case-insensitive, partial match)

---

## 5. Integration Points

### Data Dependencies

| Source | Data | Usage |
|--------|------|-------|
| MOBILE-APP EP-003 | Check-in/out records | Primary attendance data |
| EP-002 Leave Management | Approved leave requests | Display leave days in matrix |
| EP-008 Organization Data | Department list | Department filter options |
| EP-001 Foundation | User accounts | Employee list, permissions |

### Cross-Platform Relationship

```
MOBILE-APP EP-003                    WEB-APP EP-004
┌─────────────────────┐              ┌─────────────────────┐
│ Check-In/Out        │──── data ───▶│ Attendance List     │
│ (Employee action)   │              │ (HR/Admin view)     │
└─────────────────────┘              └─────────────────────┘
```

---

## 6. Assumptions & Constraints

### Assumptions

- Late threshold is fixed at 9:00 AM (from MOBILE-APP configuration)
- Company holidays are pre-configured in the system
- Leave data is available from EP-002 with approval status
- All employees have user accounts in the system

### Constraints

- Attendance data is read-only in this module (no manual edits)
- Export format is Excel (.xlsx) only
- Monthly view only (no weekly or daily drill-down in v1.0)
