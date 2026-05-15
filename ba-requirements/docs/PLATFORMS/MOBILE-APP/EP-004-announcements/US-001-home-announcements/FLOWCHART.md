---
document_type: FLOWCHART
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-004
story_id: US-001
story_name: "Home Announcements"
status: draft
version: "1.0"
last_updated: "2026-04-28"
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

# Business Process Flowcharts: Home Announcements

**Epic:** EP-004 (Announcements)
**Story:** US-001-home-announcements
**Last Updated:** 2026-04-28

---

## 1. View Announcements on Home Screen

```mermaid
flowchart TD
    A[User opens app / navigates to Home] --> B[System fetches top 5 announcements]
    B --> C{Announcements exist?}
    C -->|Yes| D[Display announcement cards]
    C -->|No| E[Show empty state message]
    D --> F{User taps announcement?}
    F -->|Yes| G[Navigate to Announcement Detail]
    F -->|No| H[User continues with other actions]
    G --> I[Mark announcement as read]
    I --> J[Display full announcement content]
```

---

## 2. Pull-to-Refresh Flow

```mermaid
flowchart TD
    A[User pulls down on home screen] --> B[Show refresh indicator]
    B --> C[Fetch latest announcements from API]
    C --> D{Fetch successful?}
    D -->|Yes| E[Update announcement list]
    D -->|No| F[Show error toast]
    E --> G[Hide refresh indicator]
    F --> G
```

---

## 3. Push Notification Flow

```mermaid
sequenceDiagram
    participant Admin as HR/Admin (WEB-APP)
    participant Server as Backend
    participant Push as Push Service
    participant Mobile as Employee Mobile

    Admin->>Server: Save & Send Announcement
    Server->>Push: Trigger push notification
    Push->>Mobile: Deliver notification
    Mobile->>Mobile: Show notification banner
    Note over Mobile: User taps notification
    Mobile->>Mobile: Open app to Announcement Detail
```

---

## Notes & Assumptions

### Notes

- Home screen shows maximum 5 announcements
- Only "Sent" status announcements are displayed (not Draft)
- Announcements ordered by Sent Date descending (newest first)

### Assumptions

- User must be authenticated to see announcements
- Announcements targeted to "Everyone" or specifically to the user are shown
- Read status is tracked per user per announcement
