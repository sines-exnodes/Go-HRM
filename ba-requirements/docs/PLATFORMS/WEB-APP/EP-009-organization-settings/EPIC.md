---
document_type: EPIC
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-009
epic_name: "Organization Settings"
status: approved
version: "1.0"
created_date: "2026-04-23"
last_updated: "2026-04-23"
approved_by: "Product Owner"
related_documents:
  - path: "../EP-001-foundation/EPIC.md"
    relationship: dependency
  - path: "../EP-004-attendance-management/EPIC.md"
    relationship: cross-reference
user_stories:
  - id: US-001
    name: "Company Profile"
    status: draft
    description: "Manage organization profile information including company name, address, and logo"
  - id: US-002
    name: "Attendance Settings"
    status: draft
    description: "Configure attendance policies including late arrival threshold"
  - id: US-003
    name: "Holiday Management"
    status: draft
    description: "Set up and manage company holidays per year; affects leave request calculations and future payslip generation"
  - id: US-004
    name: "Monthly Workdays"
    status: draft
    description: "Read-only auto-calculated view of workdays per month per year (excludes weekends and company holidays); feeds future payroll calculation"
---

# Epic: Organization Settings

**Epic ID:** EP-009
**Platform:** Exnodes HRM (WEB-APP)
**Status:** Approved
**Version:** 1.0

---

## Epic Overview

### Business Objective

Provide administrators with a centralized settings portal to configure organization-wide policies and company information. This epic consolidates all organization-level configurations that affect system behavior across modules.

### Scope

- Company profile management (name, address, logo)
- Attendance policy configuration (late arrival threshold)
- Future: Additional organization-wide settings as needed

### Out of Scope

- User-specific settings (handled per user profile)
- Module-specific configurations that don't affect org-wide behavior
- Role and permission management (EP-001)

### User Stories

| ID | Story Name | Description | Status |
|----|-----------|-------------|--------|
| US-001 | Company Profile | Manage organization profile: company name, address, logo | Draft |
| US-002 | Attendance Settings | Configure attendance policies: late arrival threshold | Draft |
| US-003 | Holiday Management | Set up company holidays per year; affects leave calculations and future payslip | Draft |
| US-004 | Monthly Workdays | Read-only auto-calculated workdays per month (excludes weekends + holidays); feeds payroll | Draft |

### Dependencies

- **Depends on:** EP-001 (Foundation) — Authentication and permissions required
- **Cross-reference:** EP-004 (Attendance Management) — Late threshold affects attendance calculations

---

## Stories Overview

### US-001: Company Profile

**Business Purpose:** Enable administrators to maintain the organization's profile information displayed throughout the system and on official documents.

**Key Deliverables:**
- Company profile settings page
- Company address management (current scope)
- Future: Company name, logo management

**User Value:** Organization can maintain accurate company information used in system displays and exports.

**Status:** Draft

**Documentation:** [US-001-company-profile/](./US-001-company-profile/)

---

### US-002: Attendance Settings

**Business Purpose:** Enable administrators to configure attendance-related policies that affect how employee check-ins are evaluated.

**Key Deliverables:**
- Attendance settings tab in Organization Settings
- Late arrival threshold configuration (time picker)
- Settings apply organization-wide to all employees

**User Value:** HR can define when employees are considered "late", ensuring consistent attendance tracking across the organization.

**Status:** Draft

**Documentation:** [US-002-attendance-settings/](./US-002-attendance-settings/)

---

### US-003: Holiday Management

**Business Purpose:** Enable administrators to maintain a company holiday calendar per year, used by the system to exclude non-working days from leave request calculations and future payslip generation.

**Key Deliverables:**
- Holiday list page with year filter, search, and pagination
- Create / Edit holiday form (full page, date range: From Date → To Date)
- Import from Vietnamese public holiday system preset
- Auto-recalculation of approved leave requests when holidays are added, edited, or deleted

**User Value:** HR can accurately track how many leave days employees use, with holidays automatically excluded from leave day counts.

**Status:** Draft

**Documentation:** [US-003-holiday-management/](./US-003-holiday-management/)

---

### US-004: Monthly Workdays

**Business Purpose:** Provide a read-only reference view of workdays per month for each year, automatically derived from the company holiday calendar and weekend exclusion. Serves as the foundation for future payroll calculation.

**Key Deliverables:**
- Monthly Workdays list page with year filter (dropdown)
- 12-row table (one per month) + pinned Total row
- Columns: Month, Total Days, Weekends, Holidays, Workdays
- Pure computed view — no stored values, no mutations

**User Value:** HR and payroll administrators can instantly see how many workdays exist in any month/year without manual calculation, with confidence that holidays are correctly excluded.

**Status:** Draft

**Documentation:** [US-004-monthly-workdays/](./US-004-monthly-workdays/)

---

## Success Criteria

### Business Acceptance Criteria

- [ ] **Company Address**: Admin can update company address
- [ ] **Late Threshold**: Admin can configure late arrival threshold
- [ ] **Immediate Effect**: Setting changes apply to future check-ins immediately
- [ ] **Access Control**: Only authorized roles can modify settings

---

## Change Log

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-04-23 | Initial epic creation | BA Agent |

---

**Document Version:** 1.0
**Last Updated:** 2026-04-23
**Author:** BA Agent
