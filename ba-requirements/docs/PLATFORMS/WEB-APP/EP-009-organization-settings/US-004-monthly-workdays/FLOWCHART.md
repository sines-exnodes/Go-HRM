---
document_type: FLOWCHART
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-009
story_id: US-004
story_name: "Monthly Workdays"
status: draft
version: "1.0"
last_updated: "2026-06-16"
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

# Business Process Flowcharts: Monthly Workdays

**Epic:** EP-009 (Organization Settings)
**Story:** US-004-monthly-workdays
**Last Updated:** 2026-06-16

---

## 1. Primary Process Flow — Page Load & Display

```mermaid
flowchart TD
    A[User navigates to Monthly Workdays] --> B{Has organization.workdays.view?}
    B -->|No| C[Access denied / hidden from nav]
    B -->|Yes| D[Page loads with current year selected]
    D --> E[System fetches holiday calendar for selected year]
    E --> F[For each month Jan–Dec: compute Total Days, Weekends, Holidays, Workdays]
    F --> G[Sum all months for Total row]
    G --> H{Holidays found for year?}
    H -->|Yes| I[Render full table with all columns populated]
    H -->|No| J[Render table with Holidays = 0 + info note]
    I --> K[User can change year via dropdown]
    J --> K
    K --> E
```

---

## 2. Workday Calculation Flow (per month)

```mermaid
flowchart TD
    A[Input: Year + Month] --> B[Count calendar days in month → Total Days]
    B --> C[Count Saturdays + Sundays in month → Weekends]
    C --> D[Query holiday records for selected year]
    D --> E[For each holiday: expand From Date → To Date into individual dates]
    E --> F[Filter dates that fall within this month]
    F --> G[Count filtered dates → Holidays]
    G --> H[Workdays = Total Days − Weekends − Holidays]
    H --> I[Return: Total Days, Weekends, Holidays, Workdays]
```

---

## 3. Edge Case — Cross-Month Holiday

```mermaid
flowchart TD
    A[Holiday: Dec 30 → Jan 2] --> B[Expand to: Dec 30, Dec 31, Jan 1, Jan 2]
    B --> C{Which month?}
    C -->|December| D[Dec 30, Dec 31 → count toward December Holidays]
    C -->|January| E[Jan 1, Jan 2 → count toward January Holidays]
```

---

## 4. Notes & Assumptions

### Notes
- Page is purely read-only — no user mutations
- Calculation runs live on every page load; holiday calendar changes reflect on next visit
- A holiday falling on a weekend still counts in Holidays (not discarded)
- Leap years handled automatically by the calendar day count

### Open Questions
- [ ] Should users without `organization.workdays.view` see the nav item (hidden) or see it greyed out? — Owner: Product Owner
