---
document_type: FLOWCHART
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-009
story_id: US-003
story_name: "Holiday Management"
status: draft
version: "1.0"
last_updated: "2026-06-15"
add_on_sections: []
approved_by: null
related_documents:
  - path: "./ANALYSIS.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
revision_history: []
---

# Business Process Flowcharts: Holiday Management

**Epic:** EP-009 (Organization Settings)
**Story:** US-003-holiday-management
**Last Updated:** 2026-06-15

---

## Table of Contents

1. [Primary Process Flow — Holiday CRUD](#1-primary-process-flow--holiday-crud)
2. [Import From Template Flow](#2-import-from-template-flow)
3. [Delete & Leave Recalculation Flow](#3-delete--leave-recalculation-flow)
4. [Leave Day Calculation with Holidays](#4-leave-day-calculation-with-holidays)
5. [State Diagram](#5-state-diagram)
6. [Notes & Assumptions](#6-notes--assumptions)

---

## 1. Primary Process Flow — Holiday CRUD

```mermaid
flowchart TD
    A[HR opens Holidays page] --> B[Select year from dropdown]
    B --> C[View holiday list for year]
    C --> D{Action?}
    D -->|Add New| E[Navigate to Create Holiday page]
    D -->|Edit| F[Navigate to Update Holiday page]
    D -->|Delete| G[Show delete confirmation dialog]
    D -->|Search| H[Filter list by holiday name]
    E --> I[Fill Holiday Name, From Date, To Date]
    I --> J{Form valid?}
    J -->|No| K[Show inline validation errors]
    K --> I
    J -->|Yes| L[Save holiday]
    L --> M[Toast: Holiday has been created]
    M --> C
    F --> N[Pre-fill form with existing data]
    N --> O{Form changed and valid?}
    O -->|No| P[Show errors or discard dialog]
    P --> N
    O -->|Yes| Q[Save changes]
    Q --> R[Recalculate affected approved leave requests]
    R --> S[Toast: Holiday has been updated]
    S --> C
```

---

## 2. Import From Template Flow

```mermaid
flowchart TD
    A[HR clicks Import From Template] --> B[Open Import modal]
    B --> C[Select target year]
    C --> D{Year has existing holidays?}
    D -->|Yes| E[Show warning: N holidays exist, duplicates will be skipped]
    D -->|No| F[Load Vietnamese public holiday preset for year]
    E --> F
    F --> G{Preset available?}
    G -->|No| H[Show: No template available for year]
    G -->|Yes| I[Display preview table with all rows pre-checked]
    I --> J[HR reviews and unchecks any rows to exclude]
    J --> K[HR clicks Import N Holidays]
    K --> L[System imports checked holidays, skips duplicates]
    L --> M[Recalculate affected approved leave requests]
    M --> N[Toast: X holidays imported, Y skipped]
    N --> O[List refreshes for imported year]
```

---

## 3. Delete & Leave Recalculation Flow

```mermaid
flowchart TD
    A[HR clicks Delete from gear menu] --> B[Show confirmation dialog]
    B --> C{HR confirms?}
    C -->|Cancel| D[Dialog closes, no change]
    C -->|Delete| E[Soft-delete holiday record]
    E --> F[Find all Approved leave requests overlapping deleted holiday dates]
    F --> G{Any affected requests?}
    G -->|No| H[Toast: Holiday has been deleted]
    G -->|Yes| I[Recalculate leave days for each affected request]
    I --> J[Update leave balance: holiday days re-added to consumed count]
    J --> K[Toast: Holiday deleted. N leave request(s) recalculated]
    H --> L[List refreshes]
    K --> L
```

---

## 4. Leave Day Calculation with Holidays

```mermaid
flowchart TD
    A[Employee submits leave request: From → To] --> B[System counts calendar days in range]
    B --> C[Query company holiday calendar for same year]
    C --> D[Count holiday days that fall within leave range]
    D --> E{Holiday days > 0?}
    E -->|No| F[Leave days = calendar days]
    E -->|Yes| G[Leave days = calendar days minus holiday days]
    F --> H[Deduct from employee leave quota]
    G --> H
    H --> I[Leave request saved with correct day count]
```

---

## 5. State Diagram

```mermaid
stateDiagram-v2
    [*] --> Active: Holiday created / imported
    Active --> Active: Edited (dates/name changed → recalculate leave)
    Active --> Deleted: Soft-deleted (→ recalculate leave)
    Deleted --> [*]
```

---

## 6. Notes & Assumptions

### Notes

- Holiday recalculation is triggered by: create (if overlaps existing approved leave), edit (date changes), delete
- Only `Approved` status leave requests are recalculated; Pending/Rejected/Cancelled are not affected
- Total Days is always computed: (To Date − From Date) + 1 calendar days (inclusive)
- The year filter on the list shows all years with at least one holiday + current year

### Assumptions

- Vietnamese public holiday preset is maintained within the system (no external API)
- A holiday spanning year boundaries (e.g., Dec 31 – Jan 2) belongs to the year of its From Date
- Half-day leave on a holiday date results in 0.5 holiday days excluded

### Open Questions

- [ ] Should employees receive a notification when their leave balance is auto-recalculated? — Owner: Product Owner
- [ ] Which roles have `organization.holidays.manage` permission by default? — Owner: Product Owner
