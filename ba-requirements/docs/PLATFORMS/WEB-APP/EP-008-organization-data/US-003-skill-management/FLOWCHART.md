---
document_type: FLOWCHART
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-003
story_name: "Skill Management"
status: draft
version: "1.0"
last_updated: "2026-03-24"
related_documents:
  - path: "./ANALYSIS.md"
    relationship: sibling
  - path: "./REQUIREMENTS.md"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
---

# Flowchart: Skill Management

**Epic:** EP-008 (Organization Data)
**Story:** US-003-skill-management
**Status:** Draft

---

## 1. Skill CRUD Flow

```mermaid
flowchart TD
    A[Administrator opens Skill List] --> B{Action?}
    B -->|View| C[Browse / Search / Paginate]
    B -->|Create| D[Click '+ Add New']
    B -->|Edit| E[Click Gear → Edit]
    B -->|Delete| F[Click Gear → Delete]

    D --> G[Navigate to Create Skill page]
    G --> H[Enter skill name]
    H --> I[Click Save / Save & Create Another]
    I --> J{Valid?}
    J -->|Yes| K[Skill created → Toast → Return to list / Clear form]
    J -->|No| L[Show inline error]
    L --> H

    E --> M[Navigate to Edit Skill page]
    M --> N[Modify skill name]
    N --> O[Click Save]
    O --> P{Valid?}
    P -->|Yes| Q[Skill updated → Toast → Return to list]
    P -->|No| R[Show inline error]
    R --> N

    F --> S{Employees assigned?}
    S -->|Yes ≥1| T[Show blocked message with employee count]
    S -->|No = 0| U[Show confirmation dialog]
    U --> V{Confirm?}
    V -->|Yes| W[Skill deleted → Toast → Refresh list]
    V -->|No| X[Cancel → Stay on list]
```

---

**Document Version:** 1.0
**Last Updated:** 2026-03-24
