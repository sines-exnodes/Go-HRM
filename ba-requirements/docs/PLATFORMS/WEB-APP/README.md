# Exnodes HRM - Web Application

> **Platform:** Exnodes HRM (Web Application)
> **Target Users:** HR Administrators, Managers, Employees
> **Target Customer:** Small-Medium Businesses (SMBs)
> **Status:** Phase 1 - Foundation

---

## Platform Purpose

Exnodes HRM is a unified web application for managing human resource operations in small and medium businesses. It replaces spreadsheets and manual processes with a centralized platform covering:

- **Manage employees** with profiles, organizational structure, and employment records
- **Track attendance and leave** with configurable policies and approval workflows
- **Process payroll** with salary structures, deductions, and payslip generation
- **Handle recruitment** with job postings, applicant tracking, and onboarding
- **Manage performance** with review cycles, goals, and feedback
- **Administer training** with course management and skill tracking
- **Configure access** with role-based permissions for HR, managers, and employees

---

## User Roles

| Role | Description | Access Level |
|------|-------------|--------------|
| **HR Administrator** | HR team with full system access | All modules, settings, and reports |
| **Manager** | Department/team managers | Team-scoped: approvals, reviews, team dashboard |
| **Employee** | All staff members | Self-service: profile, requests, payslips, training |

---

## Core Modules

| # | Module | Description |
|---|--------|-------------|
| 1 | **Foundation** | Authentication, dashboard, navigation, settings, role management |
| 2 | **Employee Management** | Employee profiles, org structure, employment records, documents |
| 3 | **Attendance & Leave** | Attendance tracking, leave policies, requests, approvals, calendar |
| 4 | **Payroll** | Salary structures, deductions, payslip generation, disbursement |
| 5 | **Recruitment** | Job postings, applicant tracking, interviews, onboarding |
| 6 | **Performance Management** | Review cycles, goals, feedback, competency assessment |
| 7 | **Training & Development** | Course catalog, enrollment, skill tracking, certifications |

---

## Epic Structure

| Epic | Name | Status | Description |
|------|------|--------|-------------|
| EP-001 | Foundation | Draft | Authentication, Dashboard, Navigation, Settings |
| EP-002 | Employee Management | Planned | Profiles, Org Structure, Records, Documents |
| EP-003 | Attendance & Leave | Planned | Tracking, Policies, Requests, Approvals |
| EP-004 | Payroll | Planned | Salary, Deductions, Payslips, Disbursement |
| EP-005 | Recruitment | Planned | Job Posts, Applicants, Interviews, Onboarding |
| EP-006 | Performance Management | Planned | Reviews, Goals, Feedback, Assessment |
| EP-007 | Training & Development | Planned | Courses, Enrollment, Skills, Certifications |
| EP-008 | Organization Data | Approved | Department Management, Position Management |

---

## Epic Documents

| Epic | Path | Status |
|------|------|--------|
| EP-001: Foundation | [EP-001-foundation/EPIC.md](EP-001-foundation/EPIC.md) | Draft |
| EP-008: Organization Data | [EP-008-organization-data/EPIC.md](EP-008-organization-data/EPIC.md) | Approved |

---

## Story & Detail Requirement Inventory

### EP-001 — Foundation

| Story | Name | Documents | Detail Requirements |
|-------|------|-----------|---------------------|
| US-001 | Authentication | [ANALYSIS](EP-001-foundation/US-001-authentication/ANALYSIS.md), [REQUIREMENTS](EP-001-foundation/US-001-authentication/REQUIREMENTS.md), [FLOWCHART](EP-001-foundation/US-001-authentication/FLOWCHART.md), [TODO](EP-001-foundation/US-001-authentication/TODO.yaml) | DR-US-001-01 |
| US-004 | Role & Permission Management | [ANALYSIS](EP-001-foundation/US-004-role-permission-management/ANALYSIS.md), [REQUIREMENTS](EP-001-foundation/US-004-role-permission-management/REQUIREMENTS.md), [FLOWCHART](EP-001-foundation/US-004-role-permission-management/FLOWCHART.md), [TODO](EP-001-foundation/US-004-role-permission-management/TODO.yaml) | [DR-001-004-01 Role List](EP-001-foundation/US-004-role-permission-management/details/DR-001-004-01-role-list.md) |

### EP-008 — Organization Data

| Story | Name | Documents | Detail Requirements |
|-------|------|-----------|---------------------|
| US-001 | Department Management | [ANALYSIS](EP-008-organization-data/US-001-department-management/ANALYSIS.md), [REQUIREMENTS](EP-008-organization-data/US-001-department-management/REQUIREMENTS.md), [FLOWCHART](EP-008-organization-data/US-001-department-management/FLOWCHART.md), [TODO](EP-008-organization-data/US-001-department-management/TODO.yaml) | [DR-008-001-01 Department List](EP-008-organization-data/US-001-department-management/details/DR-008-001-01-department-list.md), [DR-008-001-02 Create](EP-008-organization-data/US-001-department-management/details/DR-008-001-02-create-department.md), [DR-008-001-03 Edit](EP-008-organization-data/US-001-department-management/details/DR-008-001-03-edit-department.md), [DR-008-001-04 Delete](EP-008-organization-data/US-001-department-management/details/DR-008-001-04-delete-department.md) |
| US-002 | Position Management | [ANALYSIS](EP-008-organization-data/US-002-position-management/ANALYSIS.md), [REQUIREMENTS](EP-008-organization-data/US-002-position-management/REQUIREMENTS.md), [FLOWCHART](EP-008-organization-data/US-002-position-management/FLOWCHART.md), [TODO](EP-008-organization-data/US-002-position-management/TODO.yaml) | [DR-008-002-01 Position List](EP-008-organization-data/US-002-position-management/details/DR-008-002-01-position-list.md), [DR-008-002-02 Create](EP-008-organization-data/US-002-position-management/details/DR-008-002-02-create-position.md), [DR-008-002-03 Edit](EP-008-organization-data/US-002-position-management/details/DR-008-002-03-edit-position.md), [DR-008-002-04 Delete](EP-008-organization-data/US-002-position-management/details/DR-008-002-04-delete-position.md) |

---

## Detail Requirements Index

### EP-001 — Foundation

#### US-001 Authentication

| ID | Name | Parent FR | Status | File |
|----|------|-----------|--------|------|
| DR-US-001-01 | Sign In | FR-US-001-01 | Draft | [DR-US-001-01.md](EP-001-foundation/US-001-authentication/details/DR-US-001-01.md) |

#### US-004 Role & Permission Management

| ID | Name | Parent FR | Status | File |
|----|------|-----------|--------|------|
| DR-001-004-01 | Role List | FR-US-004-01 | Draft | [DR-001-004-01-role-list.md](EP-001-foundation/US-004-role-permission-management/details/DR-001-004-01-role-list.md) |

### EP-008 — Organization Data

#### US-001 Department Management

| ID | Name | Parent FR | Status | File |
|----|------|-----------|--------|------|
| DR-008-001-01 | Department List | FR-US-001-01 | Draft | [DR-008-001-01-department-list.md](EP-008-organization-data/US-001-department-management/details/DR-008-001-01-department-list.md) |
| DR-008-001-02 | Create Department | FR-US-001-05 | Draft | [DR-008-001-02-create-department.md](EP-008-organization-data/US-001-department-management/details/DR-008-001-02-create-department.md) |
| DR-008-001-03 | Edit Department | FR-US-001-06 | Draft | [DR-008-001-03-edit-department.md](EP-008-organization-data/US-001-department-management/details/DR-008-001-03-edit-department.md) |
| DR-008-001-04 | Delete Department | FR-US-001-07 | Draft | [DR-008-001-04-delete-department.md](EP-008-organization-data/US-001-department-management/details/DR-008-001-04-delete-department.md) |

#### US-002 Position Management

| ID | Name | Parent FR | Status | File |
|----|------|-----------|--------|------|
| DR-008-002-01 | Position List | FR-US-002-01 | Draft | [DR-008-002-01-position-list.md](EP-008-organization-data/US-002-position-management/details/DR-008-002-01-position-list.md) |
| DR-008-002-02 | Create Position | FR-US-002-05 | Draft | [DR-008-002-02-create-position.md](EP-008-organization-data/US-002-position-management/details/DR-008-002-02-create-position.md) |
| DR-008-002-03 | Edit Position | FR-US-002-06 | Draft | [DR-008-002-03-edit-position.md](EP-008-organization-data/US-002-position-management/details/DR-008-002-03-edit-position.md) |
| DR-008-002-04 | Delete Position | FR-US-002-07 | Draft | [DR-008-002-04-delete-position.md](EP-008-organization-data/US-002-position-management/details/DR-008-002-04-delete-position.md) |

---

## Progress Summary

| Epic | Stories | Detail Reqs | Status |
|------|---------|-------------|--------|
| EP-001 Foundation | 2 (US-001, US-004) | 2 | In Progress |
| EP-008 Organization Data | 2 (US-001, US-002) | 8 | In Progress |
| **Total** | **4** | **10** | — |

---

## Naming Conventions

| Level | Format | Example |
|-------|--------|---------|
| Epic | EP-XXX | EP-001, EP-008 |
| User Story | US-XXX | US-001, US-004 |
| Detail Requirement ID | DR-{EPIC}-{US}-NN | DR-008-001-01 |
| Detail Requirement File | DR-{EPIC}-{US}-NN-{slug}.md | DR-008-001-01-department-list.md |

---

## Related Documentation

- [Project Overview](../../PROJECT_OVERVIEW.md) - Exnodes HRM project overview
- [Shared Design System](../../SHARED/DESIGN_SYSTEM.md) - Design tokens and patterns

---

**Last Updated:** 2026-03-23
**Status:** In Progress
