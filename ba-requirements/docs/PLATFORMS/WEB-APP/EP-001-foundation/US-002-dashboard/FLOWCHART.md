---
document_type: FLOWCHART
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-002
story_name: "Dashboard"
status: draft
version: "1.0"
last_updated: "2026-07-09"
related_documents:
  - path: "./ANALYSIS.md"
    relationship: sibling
  - path: "./REQUIREMENTS.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
---

# Business Process Flowcharts: Dashboard

**Epic:** EP-001 (Foundation)
**Story:** US-002-dashboard

---

## 1. Dashboard Load After Sign In

```mermaid
flowchart TD
    A[User signs in successfully] --> B[System routes user to Dashboard]
    B --> C[System checks user permissions and data scope]
    C --> D[Build visible widget list]
    D --> E[Load each widget summary]
    E --> F{Any widget data available?}
    F -->|Yes| G[Display dashboard with permitted widgets]
    F -->|No permitted widgets| H[Display welcome state with profile and navigation guidance]
    E --> I{Widget load failure?}
    I -->|One widget fails| J[Show partial error inside that widget]
    I -->|All widgets fail| K[Show full dashboard error with Retry]
```

---

## 2. Permission-Controlled Dashboard Composition

```mermaid
flowchart TD
    A[Authenticated user opens Dashboard] --> B{User has organization-wide permissions?}
    B -->|Yes| C[Show organization-wide data in permitted widgets]
    B -->|No| D{User has team-scope permissions?}
    D -->|Yes| E[Show team-scoped data in permitted widgets]
    D -->|No| F[Show personal data in permitted widgets]
    C --> G[Use same dashboard layout and widget order]
    E --> G
    F --> G
    G --> H[Hide widgets, rows, links, and actions without permission]
    H --> I[Display remaining widgets with counts and short actionable lists]
```

---

## 3. Widget Navigation

```mermaid
flowchart TD
    A[User reviews dashboard widget] --> B{User selects widget action}
    B -->|View All| C[Navigate to source module list]
    B -->|Open row| D[Navigate to source record detail]
    B -->|Quick action| E[Navigate to existing create or workflow page]
    C --> F[Source module applies its normal permissions and filters]
    D --> F
    E --> F
```

---

## Notes

- Dashboard does not perform create, edit, delete, approve, reject, send, or close actions inline.
- Source module pages remain authoritative for record actions and detail views.
- Dashboard uses one common layout; permissions determine which widgets and actions are visible.
