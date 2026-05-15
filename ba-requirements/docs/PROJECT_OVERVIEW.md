---
document_type: PROJECT_OVERVIEW
project_name: "Exnodes HRM"
status: draft
created_date: "2026-02-25"
last_updated: "2026-02-25"
total_platforms: 1
total_epics: 8
version: "1.0"
input_sources:
  - type: text
    description: "Business requirements from stakeholder discussion"
    extraction_date: "2026-02-25"
  - type: figma
    file_id: null
    description: "Figma designs in progress - to be extracted when ready"
related_documents:
  - docs/PLATFORMS/WEB-APP/README.md
  - docs/SHARED/DESIGN_SYSTEM.md
add_on_sections: []
---

# Project Overview: Exnodes HRM

> This document provides a high-level overview of the Exnodes HRM project.
> It serves as the entry point for understanding the project scope, structure, and business objectives.

---

## Quick Reference

| Attribute | Value |
|-----------|-------|
| **Project Name** | Exnodes HRM |
| **Status** | Draft |
| **Total Platforms** | 1 (Unified Web Application) |
| **Total Epics** | 8 (suggested) |
| **Created Date** | 2026-02-25 |
| **Target Customer** | Small-Medium Businesses (SMBs) |
| **Target Completion** | TBD |

---

## 1. Business Context

### 1.1 Problem Statement

Small and medium businesses face significant challenges managing their human resources effectively:

- **Manual Processes**: HR operations handled through spreadsheets, paper forms, and disconnected tools
- **No Centralized System**: Employee data, attendance records, leave requests, and payroll information scattered across multiple files and systems
- **Compliance Risks**: Difficulty tracking labor law compliance, document expiry, and policy enforcement without a structured system
- **Limited Visibility**: Managers lack real-time insight into team attendance, leave balances, and performance
- **Talent Management Gaps**: No structured approach to recruitment, onboarding, performance reviews, or training
- **Time-Consuming Administration**: HR staff spend excessive time on repetitive tasks that could be automated

**Solution:**
Exnodes HRM provides a unified, affordable HR management platform purpose-built for SMBs — centralizing employee management, attendance, leave, payroll, recruitment, and performance into a single web application.

### 1.2 Desired Future State

**Unified HR Platform:**
- HR administrators manage all employee data, policies, and processes through one application
- Managers have visibility into their teams with self-service tools for approvals and reviews
- Employees can view their information, submit requests, and track their development
- Automated workflows reduce manual effort for common HR tasks
- Dashboards provide real-time insight into workforce metrics

---

## 2. Stakeholders

### 2.1 Primary Stakeholders

| Stakeholder | Role | Description |
|-------------|------|-------------|
| HR Administrators | System Administrator | Full access to all HR modules and system configuration |
| Managers | Team Leader | Department/team-scoped access for approvals, reviews, and team visibility |
| Employees | Self-Service User | Access to personal information, requests, and development tools |

### 2.2 User Types

| Role | Key Activities |
|------|----------------|
| **HR Administrator** | Manage employees, configure policies, process payroll, manage recruitment, generate reports |
| **Manager** | Approve leave/attendance, conduct performance reviews, view team dashboards |
| **Employee** | View profile, submit leave requests, clock attendance, view payslips, complete training |

---

## 3. Project Scope

### 3.1 Core HR Modules

| Module | Description |
|--------|-------------|
| **Employee Management** | Employee profiles, employment records, organizational structure, document management |
| **Attendance & Leave** | Attendance tracking, leave policies, leave requests, approval workflows, calendar |
| **Payroll** | Salary structures, deductions, allowances, payslip generation, disbursement tracking |

### 3.2 Talent Management Modules

| Module | Description |
|--------|-------------|
| **Recruitment** | Job postings, applicant tracking, interview scheduling, offer management, onboarding |
| **Performance Management** | Review cycles, goal setting, feedback, competency assessment, improvement plans |
| **Training & Development** | Course catalog, enrollment, skill tracking, certifications, learning paths |

### 3.3 Foundation Module

| Module | Description |
|--------|-------------|
| **Foundation** | Authentication, dashboard, navigation, role-based access, platform settings |

### 3.4 Explicitly Out of Scope

The following are **NOT** included in this project:

- Financial accounting or bookkeeping beyond payroll
- Tax filing or tax advisory automation
- Benefits administration (insurance, retirement plans)
- Time tracking for project/billing purposes
- Third-party HRIS data migration tools
- Mobile native application (web-responsive only)
- Multi-language / localization (initial release)
- AI-based recommendations or analytics

---

## 4. Platform Structure

### 4.1 Platform Overview

```
EXNODES HRM
│
└── Web Application (Unified Platform)
    ├── Users: HR Administrators, Managers, Employees
    ├── Purpose: End-to-end HR management for SMBs
    └── Status: Phase 1 - Foundation
```

---

## 5. Epic Structure

### 5.1 Suggested Epics

| Epic | Name | Key Deliverables |
|------|------|------------------|
| EP-001 | Foundation | Authentication, Dashboard, Navigation, Settings, Role Management |
| EP-002 | Employee Management | Employee profiles, Org structure, Employment records, Documents |
| EP-003 | Attendance & Leave | Attendance tracking, Leave policies, Requests, Approvals, Calendar |
| EP-004 | Payroll | Salary structures, Deductions, Payslip generation, Disbursement |
| EP-005 | Recruitment | Job postings, Applicant tracking, Interviews, Onboarding |
| EP-006 | Performance Management | Review cycles, Goals, Feedback, Competency assessment |
| EP-007 | Training & Development | Course catalog, Enrollment, Skill tracking, Certifications |
| EP-008 | Organization Data | Department management, Position management |

> **Note:** Epic structure will be refined as each module is analyzed in detail.

---

## 6. Success Criteria

### 6.1 Business Objectives

| ID | Objective | Success Metric |
|----|-----------|----------------|
| BO-1 | Centralize HR operations | Single platform replaces spreadsheets and manual processes |
| BO-2 | Reduce HR administrative time | Automated workflows for common tasks (leave, attendance, payroll) |
| BO-3 | Improve employee experience | Self-service portal for requests and information access |
| BO-4 | Enable data-driven decisions | Dashboard with real-time workforce metrics |
| BO-5 | Support compliance | Policy enforcement, document tracking, audit trails |
| BO-6 | Streamline talent management | Structured recruitment, performance, and training processes |

### 6.2 Documentation Quality Criteria

| Criterion | Requirement |
|-----------|-------------|
| Completeness | All required sections filled per template |
| Business Focus | Technical keywords < 30% of business keywords |
| Consistency | Metadata consistent across all documents |
| Traceability | All requirements trace to input sources |
| EPIC Approval | All EPICs approved before story creation |

---

## 7. Document Index

### 7.1 Platform Documents

| Document | Path | Status |
|----------|------|--------|
| Web Application Overview | [WEB-APP/README.md](PLATFORMS/WEB-APP/README.md) | Draft |
| EP-001 Foundation EPIC | [WEB-APP/EP-001-foundation/EPIC.md](PLATFORMS/WEB-APP/EP-001-foundation/EPIC.md) | Draft |
| EP-008 Organization Data EPIC | [WEB-APP/EP-008-organization-data/EPIC.md](PLATFORMS/WEB-APP/EP-008-organization-data/EPIC.md) | Draft |

### 7.2 Shared Documents

| Document | Path | Purpose |
|----------|------|---------|
| Design System | [SHARED/DESIGN_SYSTEM.md](SHARED/DESIGN_SYSTEM.md) | Design tokens and patterns (awaiting Figma) |

---

## 8. Next Steps

### Immediate Actions

1. [ ] Review and finalize this PROJECT_OVERVIEW.md
2. [ ] Review EP-001 Foundation EPIC
3. [ ] Approve EP-001 EPIC (set status to "approved")
4. [ ] Validate EPIC with `/validate-epic EP-001`
5. [ ] Create user stories with `/new-story US-001-xxx`

### Commands to Use

```bash
# Validate EPIC structure
/validate-epic EP-001

# Create user stories (after EPIC approved)
/new-story US-001-authentication

# Generate flowchart
/generate-flowchart US-001 "authentication flow"

# Extract Figma design context (when ready)
/figma-extract US-001-authentication
```

---

## 9. Change Log

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-02-25 | Initial project overview for Exnodes HRM | BA Team |

---

**Document Status:** Draft
**Last Updated:** 2026-02-25
**Next Review:** TBD
