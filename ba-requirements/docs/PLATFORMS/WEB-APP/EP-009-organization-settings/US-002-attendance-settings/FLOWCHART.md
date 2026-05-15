---
document_type: FLOWCHART
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-009
story_id: US-002
story_name: "Attendance Settings"
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

# Business Process Flowcharts: Attendance Settings

**Epic:** EP-009 (Organization Settings)
**Story:** US-002-attendance-settings
**Last Updated:** 2026-04-23

---

## 1. Configure Late Arrival Threshold Flow

```mermaid
flowchart TD
    A[Admin navigates to Organization Settings] --> B[System displays settings page]
    B --> C[Admin selects Attendance tab]
    C --> D[System loads current threshold value]
    D --> E[Admin opens time picker]
    E --> F[Admin selects new threshold time]
    F --> G[Admin clicks Save]
    G --> H{Validation passes?}
    H -->|No| I[Show validation error]
    I --> E
    H -->|Yes| J[Save to database]
    J --> K[Show success toast]
    K --> L[Stay on page with updated value]
```

---

## 2. Late Arrival Calculation Flow

```mermaid
flowchart TD
    A[Employee checks in via Mobile App] --> B[System captures check-in time]
    B --> C[System retrieves late arrival threshold]
    C --> D{Check-in time > Threshold?}
    D -->|Yes| E[Mark attendance as Late]
    D -->|No| F[Mark attendance as On-Time]
    E --> G[Save attendance record]
    F --> G
    G --> H[Display in Attendance Matrix EP-004]
```

---

## 3. Threshold Change Impact

```mermaid
sequenceDiagram
    participant Admin
    participant Settings
    participant Database
    participant MobileApp
    participant AttendanceMatrix

    Admin->>Settings: Update threshold to 8:30 AM
    Settings->>Database: Save new threshold
    Database-->>Settings: Confirm saved
    Settings-->>Admin: Success toast
    
    Note over Database: Historical records unchanged
    
    MobileApp->>Database: Employee checks in at 8:45 AM
    Database-->>MobileApp: Threshold = 8:30 AM
    MobileApp->>Database: Save as "Late"
    
    AttendanceMatrix->>Database: Load attendance
    Database-->>AttendanceMatrix: Show "Late" status
```

---

## Notes & Assumptions

### Notes

- Threshold changes apply immediately to future check-ins
- Historical attendance records are not modified
- Mobile app reads threshold at check-in time

### Assumptions

- Single organization-wide threshold (no per-department)
- Time comparison uses same timezone for threshold and check-in
