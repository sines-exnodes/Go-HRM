---
document_type: FLOWCHART
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-010
story_id: US-001
story_name: "Announcement List"
status: draft
version: "1.0"
last_updated: "2026-04-25"
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

# Business Process Flowcharts: Announcement List

**Epic:** EP-010 (Announcements)
**Story:** US-001-announcement-list
**Last Updated:** 2026-04-25

---

## 1. Announcement Lifecycle Flow

```mermaid
stateDiagram-v2
    [*] --> Draft: Create
    Draft --> Draft: Edit
    Draft --> Published: Publish
    Published --> Archived: Archive
    Draft --> [*]: Delete
    Archived --> [*]: Delete
```

---

## 2. Create Announcement Flow

```mermaid
flowchart TD
    A[User clicks + Add Announcement] --> B[System shows create form]
    B --> C[User fills in title, content]
    C --> D{Save as Draft or Publish?}
    D -->|Save Draft| E[Save with status=Draft]
    D -->|Publish| F[Save with status=Published]
    E --> G[Return to list, show success toast]
    F --> G
```

---

## 3. View Announcement List Flow

```mermaid
flowchart TD
    A[User navigates to Announcements] --> B[System loads announcement list]
    B --> C[Display table with announcements]
    C --> D{User action?}
    D -->|Search| E[Filter by keyword]
    D -->|Filter Status| F[Filter by Draft/Published/Archived]
    D -->|Click Row| G[Open announcement details]
    D -->|Click Edit| H[Open edit form]
    D -->|Click Delete| I[Show delete confirmation]
    E --> C
    F --> C
```

---

## Notes & Assumptions

### Notes

- Announcements follow a simple lifecycle: Draft → Published → Archived
- Only Draft announcements can be edited
- Delete is a hard delete (no soft delete/restore)

### Assumptions

- Only Admin/HR can create and manage announcements
- All employees can view published announcements
- No approval workflow for announcements
