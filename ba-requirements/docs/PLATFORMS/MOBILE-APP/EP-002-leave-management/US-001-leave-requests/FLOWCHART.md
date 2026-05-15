---
document_type: FLOWCHART
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-002
story_id: US-001
story_name: "Leave Requests"
status: draft
version: "1.0"
last_updated: "2026-04-16"
---

# Flowcharts: Leave Requests

**Epic:** EP-002 (Leave Management)
**Story:** US-001-leave-requests
**Platform:** MOBILE-APP

---

## 1. Dashboard Load Flow

```mermaid
flowchart TD
    A[User taps Leave tab] --> B[Show loading skeleton]
    B --> C{Cached data exists?}
    C -->|Yes| D[Display cached data]
    C -->|No| E[Show loading indicator]
    
    D --> F[Fetch fresh data in background]
    E --> F
    
    F --> G{API success?}
    G -->|Yes| H[Update dashboard with fresh data]
    G -->|No| I{Cached data shown?}
    
    I -->|Yes| J[Show subtle error toast]
    I -->|No| K[Show error state with retry]
    
    H --> L[Dashboard ready]
    J --> L
```

---

## 2. Create Leave Request Flow

```mermaid
flowchart TD
    A[User taps FAB +] --> B[Open create request screen]
    B --> C[Select leave type]
    C --> D[Select start date]
    D --> E[Select end date]
    E --> F[Select leave period]
    F --> G[Enter reason - optional]
    G --> H[Attach document - optional]
    
    H --> I[Tap Submit]
    I --> J{Client validation}
    
    J -->|Invalid| K[Show inline errors]
    K --> C
    
    J -->|Valid| L{Check leave balance}
    L -->|Insufficient| M[Show balance warning]
    M --> N{User confirms?}
    N -->|No| C
    N -->|Yes| O[Submit to API]
    
    L -->|Sufficient| O
    
    O --> P{API success?}
    P -->|Yes| Q[Show success toast]
    Q --> R[Navigate to dashboard]
    R --> S[Dashboard shows new pending request]
    
    P -->|No| T[Show error message]
    T --> U[User can retry or cancel]
```

---

## 3. View Request Details Flow

```mermaid
flowchart TD
    A[User taps request in list] --> B[Navigate to details screen]
    B --> C[Fetch request details]
    C --> D[Display full information]
    
    D --> E{Request status?}
    E -->|Pending| F[Show Cancel button]
    E -->|Approved| G[Show status only]
    E -->|Rejected| G
    E -->|Cancelled| G
    
    F --> H{User taps Cancel?}
    H -->|Yes| I[Show confirmation dialog]
    I --> J{User confirms?}
    J -->|Yes| K[Call cancel API]
    J -->|No| D
    
    K --> L{API success?}
    L -->|Yes| M[Show success toast]
    M --> N[Navigate back to list/dashboard]
    L -->|No| O[Show error toast]
    O --> D
```

---

## 4. Pull-to-Refresh Flow

```mermaid
flowchart TD
    A[User pulls down on dashboard/list] --> B[Show refresh indicator]
    B --> C[Fetch latest data from API]
    
    C --> D{API success?}
    D -->|Yes| E[Update displayed data]
    E --> F[Hide refresh indicator]
    
    D -->|No| G[Show error toast]
    G --> F
```

---

## 5. Filter Request List Flow

```mermaid
flowchart TD
    A[User taps filter icon] --> B[Open filter bottom sheet]
    B --> C[Select status filters]
    C --> D[Select date range - optional]
    D --> E[Tap Apply]
    
    E --> F[Close bottom sheet]
    F --> G[Show loading in list]
    G --> H[Fetch filtered results]
    
    H --> I{Results found?}
    I -->|Yes| J[Display filtered list]
    I -->|No| K[Show empty state]
    
    J --> L[Show active filter indicator]
    K --> L
    
    L --> M{User taps Clear Filters?}
    M -->|Yes| N[Reset to default view]
    N --> G
```

---

**Document Version:** 1.0
**Last Updated:** 2026-04-16
**Author:** BA Team
