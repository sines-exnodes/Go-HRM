---
document_type: EPIC
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-004
epic_name: "Attendance Management"
status: approved
version: "1.0"
created_date: "2026-04-21"
last_updated: "2026-04-21"
approved_by: "Product Owner"
related_documents:
  - path: "../EP-001-foundation/EPIC.md"
    relationship: dependency
  - path: "../../MOBILE-APP/EP-003-attendance-management/EPIC.md"
    relationship: cross-platform
user_stories:
  - id: US-001
    name: "Attendance List"
    status: draft
    description: "Monthly attendance matrix view with calendar-style layout for HR/Admin to view all employee attendance"
---

# Epic: Attendance Management

**Epic ID:** EP-004
**Platform:** Exnodes HRM (WEB-APP)
**Status:** Approved
**Version:** 1.0

---

## Epic Overview

### Business Objective

Provide HR, Admin, CEO, and Leaders with a web-based portal to view and manage employee attendance records captured via the mobile app. The system displays attendance data in an intuitive monthly calendar matrix format, allowing quick identification of attendance patterns, late arrivals, absences, and leave integration.

### Scope

- Monthly attendance matrix view (rows = employees, columns = dates)
- Attendance status visualization (on-time, late, absent, leave types)
- Attendance data filtering and search
- Data export capabilities (Excel)
- Integration with Leave Management (EP-002) for leave day display

### Out of Scope

- Employee check-in/out functionality (handled by MOBILE-APP EP-003)
- Attendance policy configuration
- Automated attendance alerts/notifications
- Payroll integration

### User Stories

| ID | Story Name | Description | Status |
|----|-----------|-------------|--------|
| US-001 | Attendance List | Monthly calendar-style attendance matrix showing all employees' attendance status with filters, search, and export | Draft |

### Dependencies

- **EP-001 (Foundation):** Authentication (US-001), Role & Permission Management (US-004), User Management (US-005)
- **EP-002 (Leave Management):** Leave Requests (US-001) — leave data displayed in attendance matrix
- **EP-008 (Organization Data):** Department Management (US-001) — department filter
- **MOBILE-APP EP-003 (Attendance Management):** Check-In/Out (US-001) — source of attendance data

### Success Criteria

- HR/Admin can view monthly attendance for all employees in a single matrix view
- Users can filter by department, status, and search by employee name
- Attendance integrates with approved leave data (AL, SL, half-day)
- Users can export attendance data to Excel (bulk and individual)
- Permission-based access: users without "Manage Data" permission see only their own attendance

### Target Users

- **HR:** Primary users for attendance monitoring and reporting
- **Admin:** Full access for system administration
- **CEO:** Overview access for company-wide attendance patterns
- **Leaders:** Department-level attendance visibility (with appropriate permissions)
- **Employees:** View own attendance only (without "Manage Data" permission)

---

## Cross-Platform Relationship

This epic is the **WEB-APP counterpart** to MOBILE-APP EP-003 (Attendance Management):

| Platform | Epic | Purpose |
|----------|------|---------|
| MOBILE-APP | EP-003 | Employee-facing: Check-in/out with GPS verification |
| WEB-APP | EP-004 | Admin-facing: View and manage attendance records |

Data flows from mobile check-ins to the web attendance list for HR review.
