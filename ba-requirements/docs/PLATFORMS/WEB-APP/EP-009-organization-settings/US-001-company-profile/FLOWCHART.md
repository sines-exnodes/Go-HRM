---
document_type: FLOWCHART
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-009
story_id: US-001
story_name: "Company Profile"
status: draft
version: "1.0"
last_updated: "2026-04-23"
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

# Business Process Flowcharts: Company Profile

**Epic:** EP-009 (Organization Settings)
**Story:** US-001-company-profile
**Last Updated:** 2026-04-23

---

## 1. Update Company Address Flow

```mermaid
flowchart TD
    A[Admin navigates to Organization Settings] --> B[System displays settings page]
    B --> C[Admin selects Company Profile tab]
    C --> D[System loads current company data]
    D --> E[Admin edits address fields]
    E --> F{Changes made?}
    F -->|No| G[Admin navigates away]
    F -->|Yes| H[Admin clicks Save]
    H --> I{Validation passes?}
    I -->|No| J[Show validation errors]
    J --> E
    I -->|Yes| K[Save to database]
    K --> L[Show success toast]
    L --> M[Stay on page with updated values]
```

---

## 2. Dirty Form Navigation Flow

```mermaid
flowchart TD
    A[Admin has unsaved changes] --> B[Admin attempts to navigate away]
    B --> C{Show discard confirmation}
    C -->|Discard| D[Navigate away without saving]
    C -->|Cancel| E[Stay on page]
    C -->|Save| F[Save changes first]
    F --> G{Save successful?}
    G -->|Yes| H[Navigate to target]
    G -->|No| I[Show error, stay on page]
```

---

## Notes & Assumptions

### Notes

- Company Profile tab is part of the Organization Settings page
- Address changes are saved immediately upon clicking Save
- No approval workflow for profile changes

### Assumptions

- Only Admin/HR with organization settings permission can access
- Single company profile (no multi-tenant support)
