---
document_type: FLOWCHART
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-001
story_name: "Department Management"
status: draft
version: "1.0"
last_updated: "2026-03-03"
add_on_sections: []
approved_by: null
related_documents:
  - path: "./ANALYSIS.md"
    relationship: sibling
  - path: "./REQUIREMENTS.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
revision_history: []
---

# Business Process Flowcharts: Department Management

**Epic:** EP-008 (Organization Data)
**Story:** US-001-department-management
**Last Updated:** 2026-03-03

---

## Table of Contents

1. [Department List Flow](#1-department-list-flow)
2. [Create Department Flow](#2-create-department-flow)
3. [Edit Department Flow](#3-edit-department-flow)
4. [Deactivate Department Flow](#4-deactivate-department-flow)
5. [Actor Interactions](#5-actor-interactions)
6. [State Diagram](#6-state-diagram)
7. [Notes & Assumptions](#7-notes--assumptions)

---

## 1. Department List Flow

```mermaid
flowchart TD
    A[Administrator opens Departments] --> B[System displays Department List]
    B --> C{User action?}
    C -->|Search| D[User types in search field]
    D --> E[List filters by department name]
    E --> C
    C -->|Paginate| F[User changes page or rows per page]
    F --> B
    C -->|Export| G[System exports department list]
    G --> H[File downloaded]
    C -->|Add New| I[Open Create Department flow]
    C -->|Gear icon| J[Open action menu]
    J --> K{Action selected?}
    K -->|Edit| L[Open Edit Department flow]
    K -->|Deactivate| M[Open Deactivate Department flow]
```

---

## 2. Create Department Flow

```mermaid
flowchart TD
    A[Administrator clicks Add New] --> B[System opens Create Department form]
    B --> C[Administrator enters department name]
    C --> D[Administrator clicks Save]
    D --> E{Name is unique?}
    E -->|No| F[Show: Department name already exists]
    F --> C
    E -->|Yes| G[System creates department with Active status]
    G --> H[Return to Department List]
    H --> I[New department appears in list]
```

### Key Steps

1. **Trigger** — Administrator clicks "Add New" button on Department List
2. **Name Entry** — Required field; validated for uniqueness on save
3. **Validation** — System checks name uniqueness
4. **Creation** — Department created as Active; list refreshes

---

## 3. Edit Department Flow

```mermaid
flowchart TD
    A[Administrator clicks gear icon] --> B[Action menu appears]
    B --> C[Administrator selects Edit]
    C --> D[System opens Edit form pre-filled with current name]
    D --> E[Administrator modifies department name]
    E --> F[Administrator clicks Save]
    F --> G{Name is unique?}
    G -->|No| H[Show: Department name already exists]
    H --> E
    G -->|Yes| I[System updates department record]
    I --> J[Return to Department List]
    J --> K[Updated name reflected in list]
```

---

## 4. Deactivate Department Flow

```mermaid
flowchart TD
    A[Administrator clicks gear icon] --> B[Action menu appears]
    B --> C[Administrator selects Deactivate]
    C --> D[System shows confirmation dialog]
    D --> E[Dialog shows: department name + employee count assigned]
    E --> F{Administrator confirms?}
    F -->|Cancel| G[No changes made, dialog closes]
    F -->|Confirm| H[System marks department as Inactive]
    H --> I[Department removed from active selection dropdowns]
    I --> J[Historical employee records preserved]
    J --> K[Department List updated - dept shows as Inactive]
```

### Key Steps

1. **Trigger** — Administrator selects Deactivate from gear menu
2. **Confirmation** — Dialog shows affected employee count to prevent accidental deactivation
3. **Deactivation** — Status set to Inactive; no data is deleted
4. **Cascade** — Department removed from all active dropdowns across modules
5. **Preservation** — Historical records unchanged

---

## 5. Actor Interactions

### Create Department Sequence

```mermaid
sequenceDiagram
    participant Admin as Administrator
    participant DeptList as Department List
    participant Form as Create Form
    participant System

    Admin->>DeptList: Click Add New
    DeptList-->>Form: Open create form

    Admin->>Form: Enter department name
    Admin->>Form: Click Save

    Form->>System: Validate name uniqueness
    System-->>Form: Validation passed

    System->>System: Create department (Active)
    System-->>DeptList: Refresh list
    DeptList-->>Admin: New department visible in list
```

---

## 6. State Diagram

### Department Lifecycle States

```mermaid
stateDiagram-v2
    [*] --> Active: Administrator creates department

    Active --> Active: Administrator edits name or parent
    Active --> Inactive: Administrator deactivates

    Inactive --> Active: Administrator reactivates
    Inactive --> [*]: (Future) Permanent deletion if no historical references
```

**States:**
- **Active:** Department is available for employee assignment and visible in all active selection dropdowns
- **Inactive:** Department is deactivated; excluded from new assignments but preserved in historical records

---

## 7. Notes & Assumptions

### Assumptions

1. All department management actions require appropriate role permission (via US-004)
2. Departments are a flat list — no parent-child hierarchy (confirmed by Product Owner)
3. Search filters in real-time as user types (no submit required)
4. Export downloads the currently visible/filtered list

### Open Questions Affecting Flows

- Gear icon actions need confirmation — Edit + Deactivate confirmed, View TBD
- Reactivation flow not yet confirmed — is there a "reactivate" option for inactive departments?

---

**Document Control:**
- **Version:** 1.0
- **Status:** Draft
- **Last Updated:** 2026-03-03
