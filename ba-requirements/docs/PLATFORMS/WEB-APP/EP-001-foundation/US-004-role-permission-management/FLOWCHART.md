---
document_type: FLOWCHART
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-004
story_name: "Role & Permission Management"
status: draft
version: "0.1"
last_updated: "2026-03-19"
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

# Business Process Flowcharts: Role & Permission Management

**Epic:** EP-001 (Foundation)
**Story:** US-004-role-permission-management
**Status:** 🔴 Draft — Stub (pending full flowchart elaboration)
**Last Updated:** 2026-03-19

---

> **Note:** This is a stub document. Full flowcharts will be created after Create/Edit Role designs are delivered and open questions are resolved.

---

## 1. Role List Flow (Preliminary)

```mermaid
flowchart TD
    A[Administrator opens Roles] --> B[System displays Role List]
    B --> C{User action?}
    C -->|Search| D[User types in search field]
    D --> E[List filters by role name]
    E --> C
    C -->|Paginate| F[User changes page or rows per page]
    F --> B
    C -->|Add New| G[Open Create Role flow]
    C -->|Gear icon| H[Open action menu]
    H --> I{Action selected?}
    I -->|Edit| J[Open Edit Role flow]
    I -->|Delete| K[Open Delete Role flow]
```

---

## 2. Delete Role Flow (Preliminary)

```mermaid
flowchart TD
    A[Administrator selects Delete from gear icon] --> B[System checks user count in real time]
    B --> C{User count?}
    C -->|count = 0| D[Show Confirmation Dialog]
    D --> E{Confirms?}
    E -->|Cancel| F[Dialog closes — no changes]
    E -->|Confirm| G[System permanently deletes role]
    G --> H[Role List refreshes — role removed]
    C -->|count >= 1| I[Show Blocked Dialog with user count]
    I --> J[User closes dialog — no changes]
```

---

**Document Control:**
- **Version:** 0.1
- **Status:** Draft — Stub
- **Last Updated:** 2026-03-19
