---
document_type: EPIC
epic_id: EP-001
epic_name: "Foundation"
platform: WEB-APP
platform_display: "Exnodes HRM"
status: approved
approved_by: "BA Team"
priority: critical
created_date: "2026-02-25"
last_updated: "2026-02-26"
business_owner: TBD
related_documents:
  - docs/PROJECT_OVERVIEW.md
  - docs/PLATFORMS/WEB-APP/README.md
user_stories:
  - US-001-authentication
  - US-004-role-permission-management
  - US-005-user-management
add_on_sections: []
---

# EP-001: Foundation

> **Epic:** Foundation & Core Platform
> **Platform:** Exnodes HRM (Web Application)
> **Status:** Approved
> **Priority:** Critical

---

## 1. Epic Overview

### 1.1 Business Purpose

The Foundation epic establishes the core platform for Exnodes HRM, enabling users to securely access the system and navigate its features. This epic must be completed before any functional HR modules can be built.

### 1.2 Business Value

- **Secure Access**: HR administrators, managers, and employees can securely authenticate with appropriate access levels
- **Operational Overview**: Dashboard provides at-a-glance view of key HR metrics and pending actions
- **Consistent Navigation**: Unified layout and navigation structure across all future modules
- **Platform Configuration**: Organization settings, user management, and role-based access control

### 1.3 Success Criteria

| Criteria | Measurement |
|----------|-------------|
| Users can authenticate securely | Login/logout/password reset working for all roles |
| Dashboard displays relevant information | Role-specific metrics and quick actions visible |
| Navigation provides access to all modules | Menu structure supports all planned modules |
| Roles and permissions enforced | Users only see and access their authorized features |
| Organization settings configurable | Basic settings saved and applied correctly |

---

## 2. Scope

### 2.1 In Scope

- User authentication (login, logout, password reset, session management)
- Role-based access control (HR Admin, Manager, Employee)
- Dashboard with role-specific metrics and activity feed
- Main navigation structure and responsive layout
- Organization profile and settings
- User management (invite, deactivate, role assignment)

### 2.2 Out of Scope

- Specific HR module functionality (Employee, Attendance, Payroll, etc.)
- Advanced reporting and analytics
- Third-party integrations
- Notification system (will be added as cross-cutting concern later)
- Mobile native features (web-responsive only)

---

## 3. User Stories

### 3.1 Planned Stories

| Story ID | Name | Priority | Status |
|----------|------|----------|--------|
| US-001 | User Authentication | Critical | In Progress |
| US-002 | Dashboard | High | Planned |
| US-003 | Navigation & Layout | High | Planned |
| US-004 | Role & Permission Management | High | Planned |
| US-005 | Organization Settings | Medium | Planned |

### 3.2 Story Descriptions

**US-001: User Authentication**
- As a user, I want to securely log in to the platform so that I can access my authorized features
- Includes: Login, logout, password reset, session management, remember me

**US-002: Dashboard**
- As a user, I want to see a role-specific dashboard so that I can quickly understand key metrics and pending actions
- Includes: HR Admin dashboard (workforce overview), Manager dashboard (team summary), Employee dashboard (personal summary)

**US-003: Navigation & Layout**
- As a user, I want consistent navigation so that I can easily access different modules
- Includes: Sidebar navigation, header, breadcrumbs, responsive layout, user profile menu

**US-004: Role & Permission Management**
- As an HR administrator, I want to manage user roles and permissions so that users only access what they are authorized for
- Includes: Role definitions, permission matrix, user role assignment, access control enforcement

**US-005: Organization Settings**
- As an HR administrator, I want to configure organization settings so that the platform reflects our company structure
- Includes: Company profile, departments, locations, work schedules, general preferences

---

## 4. Dependencies

### 4.1 Prerequisites

- None (this is the foundation epic)

### 4.2 Dependent Epics

All other Exnodes HRM epics depend on EP-001:
- EP-002: Employee Management
- EP-003: Attendance & Leave
- EP-004: Payroll
- EP-005: Recruitment
- EP-006: Performance Management
- EP-007: Training & Development

### 4.3 External Dependencies

| Dependency | Type | Impact |
|------------|------|--------|
| Authentication service | Backend | Required for login functionality |
| User/Role data store | Backend | Required for access control |

---

## 5. Stakeholders

| Stakeholder | Role | Involvement |
|-------------|------|-------------|
| HR Administrators | Primary User | Requirements input, UAT, settings configuration |
| Managers | End User | Dashboard and navigation feedback |
| Employees | End User | Authentication and self-service feedback |
| Product Owner | Approver | Epic approval, priority decisions |

---

## 6. Acceptance Criteria

### 6.1 Epic Completion Criteria

- [ ] All user stories (US-001 through US-005) completed
- [ ] Authentication working for all three user roles
- [ ] Dashboard displaying role-specific content
- [ ] Navigation accessible from all pages
- [ ] Role-based access control enforced
- [ ] Organization settings saved and applied
- [ ] All user stories pass validation

### 6.2 Documentation Criteria

- [ ] All user stories have 4 required files (ANALYSIS, REQUIREMENTS, FLOWCHART, TODO)
- [ ] All validations pass with `/validate-story`
- [ ] Business focus maintained (technical keywords < 30%)

---

## 7. Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Authentication complexity for multi-role system | Medium | High | Clear role definitions early, simple permission model |
| Dashboard requirements unclear until modules built | Medium | Medium | Design extensible dashboard, placeholder widgets |
| Scope creep from module-specific features | Medium | Medium | Strict scope enforcement, defer to module epics |
| Navigation structure changes as modules added | Low | Low | Design for extensibility with menu configuration |

---

## 8. Timeline

| Milestone | Target | Status |
|-----------|--------|--------|
| EPIC Approved | 2026-02-26 | Complete |
| US-001 Complete | TBD | Planned |
| US-002 Complete | TBD | Planned |
| US-003 Complete | TBD | Planned |
| US-004 Complete | TBD | Planned |
| US-005 Complete | TBD | Planned |
| Epic Complete | TBD | Planned |

---

## 9. Change Log

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-02-25 | Initial EPIC created for Exnodes HRM | BA Team |
| 1.1 | 2026-02-26 | EPIC approved, US-001 story creation started | BA Team |

---

**Document Status:** Approved
**Approval Required:** No - EPIC approved, stories can be created
**Last Updated:** 2026-02-26
