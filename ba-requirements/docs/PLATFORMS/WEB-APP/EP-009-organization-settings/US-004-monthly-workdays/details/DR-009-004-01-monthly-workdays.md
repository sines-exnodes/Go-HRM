---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
epic_id: EP-009
story_id: US-004
story_name: "Monthly Workdays"
detail_id: DR-009-004-01
detail_name: "Monthly Workdays"
status: draft
version: "1.0"
created_date: "2026-06-16"
last_updated: "2026-06-16"
related_documents:
  - path: "../ANALYSIS.md"
    relationship: parent
  - path: "../../EPIC.md"
    relationship: grandparent
  - path: "../../US-003-holiday-management/details/DR-009-003-01-holiday-management.md"
    relationship: dependency
input_sources:
  - type: text
    description: "Approved design spec — docs/superpowers/specs/2026-06-16-monthly-workdays-design.md"
    extraction_date: "2026-06-16"
  - type: text
    description: "ANALYSIS.md brainstorming session — 2026-06-16"
    extraction_date: "2026-06-16"
---

# Detail Requirement: Monthly Workdays

**Feature:** Monthly Workdays
**Epic:** EP-009 — Organization Settings
**Story:** US-004 — Monthly Workdays
**Detail ID:** DR-009-004-01
**Status:** Draft
**Version:** 1.0

---

## §1 Use Case Description

### User Story

As an HR or payroll administrator, I want to see the exact number of workdays in each month for any selected year, automatically calculated from the company holiday calendar and weekend exclusion, so that I can use accurate workday counts for payroll calculations, daily salary rates, OT thresholds, and pro-rated salary adjustments — without manual counting or error.

### Business Purpose

Currently, HR administrators must manually calculate workdays per month by subtracting weekends and company holidays from calendar days. This process is error-prone, time-consuming, and risks payroll errors when holiday calendars change. The Monthly Workdays page provides a single authoritative read-only reference derived live from the existing holiday calendar, eliminating manual calculation.

### Target Users

| Role | Access Level | Business Need |
|------|-------------|---------------|
| Admin | View | Reference workday counts for payroll configuration |
| HR | View | Use workday counts for payroll processing and salary calculations |
| CEO | View (if granted) | Review organizational workday summary |

All access controlled by the `organization.workdays.view` permission, assigned through EP-001 US-004 Role & Permission Management.

### Business Value

- Eliminates manual monthly workday calculation for HR and payroll administrators
- Ensures consistency with the company holiday calendar (US-003) — changes to holidays reflect automatically on next page visit
- Provides the foundational data for future payroll integration
- Reduces payroll errors caused by miscounted workdays

---

## §2 User Workflow

### Entry Point

The user navigates to Monthly Workdays via the left sidebar under "Organization Settings". The breadcrumb reads: **Organization Settings / Monthly Workdays**.

The user must hold the `organization.workdays.view` permission to access this page.

### Main Flow

1. User opens the Monthly Workdays page
2. The page loads showing a year dropdown (defaulting to the current calendar year) and a loading skeleton of 12 rows + Total row
3. On load completion, the table displays 12 rows (January through December) + 1 pinned Total row, with all 5 columns populated
4. User reads workday counts per month — no further interaction required for the primary use case
5. Optionally, user selects a different year from the year dropdown
6. The table reloads immediately for the selected year; all values update

### Year Filter Flow

1. User clicks the year dropdown (left side of action bar)
2. Dropdown shows available year options
3. User selects a year
4. Table reloads with computed values for the selected year
5. No other controls exist on the action bar (no search, no Add New, no export)

### No Holidays Configured Flow

1. User selects a year with no holidays configured in US-003
2. Table renders normally with Holidays = 0 for all months
3. An info note appears below the action bar:
   "No holidays configured for [year]. Set up holidays in Holiday Management."
   (the "Holiday Management" text links to the US-003 Holiday Management page)

### Error Flow

1. Page fails to load data
2. An error message is displayed in place of the table
3. A Retry button allows the user to attempt to reload

### Exit Points

- User navigates to another section via sidebar or breadcrumb
- User follows the "Holiday Management" deep link (info note) to configure holidays for the selected year

### Not Applicable Flows

This is a read-only page. There are no Create, Edit, Delete, or Export flows.

---

## §3 Field Definitions

This page contains no input form. The only interactive element is the year filter dropdown.

### Year Dropdown (Action Bar)

| Property | Value |
|----------|-------|
| Control type | Dropdown / Select |
| Position | Left side of action bar |
| Default value | Current calendar year (e.g., 2026 on load in 2026) |
| Available options | Agent suggestion: current year and a range of adjacent years (e.g., ±5 years), precise population rule to be confirmed — see Open Questions |
| On change | Table reloads for selected year; no other state is reset (no search, no pagination exist) |
| Required | Yes (always has a value; cannot be cleared) |

### Retry Button (Error State Only)

| Property | Value |
|----------|-------|
| Control type | Button |
| Label | "Retry" |
| Visible | Only when page data fails to load |
| Action | Triggers a fresh page data load attempt |

### "Holiday Management" Link (No-Holidays Info Note Only)

| Property | Value |
|----------|-------|
| Control type | Inline hyperlink |
| Label | "Holiday Management" |
| Visible | Only when the selected year has no holidays configured |
| Destination | Holiday Management page (US-003) |

---

## §4 Data Display

### Table Structure

| Property | Value |
|----------|-------|
| Row count | Always exactly 12 data rows (January through December) + 1 pinned Total row |
| Row order | Fixed: January (row 1) through December (row 12), Total (pinned last) |
| Sort | Not applicable — order is fixed by calendar month |
| Pagination | None — all 12 months always visible on one page |
| Search | None |
| Gear icon | None — no row-level actions exist |
| Export | None |

### Columns

| # | Column | Description | Total Row |
|---|--------|-------------|-----------|
| 1 | Month | Full month name (January, February, … December) | Label: "Total" |
| 2 | Total Days | Calendar days in that month for the selected year (accounts for leap year — February shows 28 or 29) | Sum of all 12 months |
| 3 | Weekends | Count of Saturdays + Sundays within the month | Sum of all 12 months |
| 4 | Holidays | Count of company holiday calendar days (from US-003) falling within the month for the selected year | Sum of all 12 months |
| 5 | Workdays | Total Days − Weekends − Holidays for that month | Sum of all 12 months (= annual workdays for selected year) |

All numeric columns (Total Days, Weekends, Holidays, Workdays) display whole numbers. No decimal values.

### Display States

| State | Condition | What the User Sees |
|-------|-----------|-------------------|
| Loading | Page data is being fetched | Skeleton rows: 12 data rows + Total row with placeholder bars; action bar and column headers are visible |
| Loaded — with holidays | Data loaded; holidays exist for selected year | Full table with all 5 columns populated; no additional notes |
| Loaded — no holidays | Data loaded; no holidays configured for selected year in US-003 | Full table with Holidays = 0 for all months and Workdays = Total Days − Weekends for all months; info note displayed: "No holidays configured for [year]. Set up holidays in Holiday Management." with a link to the Holiday Management page |
| Error | System fails to load page data | Error message in place of table; Retry button displayed |

### Info Note (No Holidays State)

Displayed below the action bar, above the table, when the selected year has no holidays in US-003.

> No holidays configured for [year]. Set up holidays in [Holiday Management].

Where `[year]` is the currently selected year and `[Holiday Management]` is a clickable link to the US-003 Holiday Management page.

---

## §5 Acceptance Criteria

### Minimum Conditions for Done

| # | Acceptance Criterion | Test Basis |
|---|---------------------|------------|
| AC-01 | Only users with `organization.workdays.view` permission can access the Monthly Workdays page; unauthorized users cannot reach this page | Permission model |
| AC-02 | The page defaults to the current calendar year on first load and displays 12 month rows plus a pinned Total row | Default state |
| AC-03 | Changing the year dropdown reloads the table with computed values for the selected year | Year filter behavior |
| AC-04 | The Workdays column for each month equals Total Days minus Weekends minus Holidays for that month | Core formula |
| AC-05 | A holiday falling on a weekend still appears in the Holidays column and reduces Workdays (holidays and weekends are not mutually exclusive) | Holiday-weekend overlap rule |
| AC-06 | A multi-day holiday that spans two months contributes only the days within each month to that month's Holidays count | Cross-month holiday split |
| AC-07 | February shows 29 Total Days in a leap year and 28 in a non-leap year without any manual configuration | Leap year handling |
| AC-08 | The Total row sums all numeric columns across the 12 months; the Workdays total equals annual workdays for the selected year | Total row calculation |
| AC-09 | When no holidays are configured for the selected year in US-003, all Holidays values show 0 and the info note with a link to Holiday Management is displayed | No-holidays state |
| AC-10 | A change to the holiday calendar (US-003) — adding, editing, or deleting a holiday — is reflected in the Monthly Workdays table on the next page load without any additional user action | Live recalculation dependency |
| AC-11 | During data loading, a skeleton of 12 rows + Total row is displayed; no layout shift occurs when data arrives | Loading state |
| AC-12 | When page data fails to load, an error message and a Retry button are displayed; clicking Retry re-attempts the load | Error state |

### Testing Scenarios

| # | Scenario | Input | Expected Result |
|---|----------|-------|----------------|
| TS-01 | Standard month — no holidays | Year: 2026, Month: March (31 days, 4 Saturdays, 4 Sundays = 8 weekends, 0 holidays) | Total Days = 31, Weekends = 8, Holidays = 0, Workdays = 23 |
| TS-02 | Month with holidays on weekdays | Year: 2026, Month: January (31 days, 4 Sat + 4 Sun = 8 weekends, 1 holiday on Monday Jan 1) | Total Days = 31, Weekends = 8, Holidays = 1, Workdays = 22 |
| TS-03 | Holiday-weekend overlap (from design spec) | Year: any, Month: January (31 days, 8 weekends, 1 holiday on a Saturday) | Weekends = 8, Holidays = 1, Workdays = 31 − 8 − 1 = 22; holiday is NOT excluded from Holidays because it falls on a weekend |
| TS-04 | Leap year | Year: 2024, Month: February | Total Days = 29 |
| TS-05 | Non-leap year | Year: 2025, Month: February | Total Days = 28 |
| TS-06 | Cross-month holiday split | Holiday: Jan 30 – Feb 2 (4 days total; 2 in January, 2 in February) | January Holidays += 2, February Holidays += 2; no other months affected |
| TS-07 | No holidays configured | Year: 2030 (no holidays in US-003 for this year) | All months: Holidays = 0, Workdays = Total Days − Weekends; info note with Holiday Management link shown |
| TS-08 | Total row correctness | Any year | Total row sums all numeric columns across 12 months; Workdays total = sum of 12 monthly Workdays values |
| TS-09 | Year change | User changes dropdown from 2026 to 2025 | Table reloads; all values update for 2025; no page refresh required |
| TS-10 | Holiday deletion reflects on next load | User deletes a holiday in US-003, then returns to Monthly Workdays | Deleted holiday no longer counted in Holidays column; Workdays increases accordingly |

---

## §6 System Rules

### Permission Model

| Permission | Access | Notes |
|-----------|--------|-------|
| `organization.workdays.view` | Can view the Monthly Workdays page | Read-only access; this is the only permission level — no manage counterpart exists |

No mutation actions exist on this page. There is no `organization.workdays.manage` permission.

Open question: Whether users lacking `organization.workdays.view` see the sidebar menu item (hidden vs. visible but disabled) is pending Product Owner decision (see Open Questions).

### Calculation Rules

#### Core Formula

```
Workdays (per month) = Total Days − Weekends − Holidays
```

Applied independently to each of the 12 months. Applied to the Total row as a sum of all monthly Workdays values.

#### Total Days

Calendar days in the month for the selected year. February = 28 in non-leap years, 29 in leap years. Calculated automatically by the system — no configuration required.

#### Weekends

Count of Saturdays and Sundays within the month. Fixed by calendar — cannot be configured.

#### Holidays

Count of individual calendar days covered by active holiday records in US-003 (EP-009 Holiday Management) that fall within the month for the selected year.

Rules:
- Multi-day holidays contribute each calendar day individually to its month
- A holiday spanning two months is split — only the days within each month count for that month
- A holiday falling on a weekend is NOT excluded; it counts in Holidays and reduces Workdays
- Soft-deleted holidays (removed from US-003) are excluded from the Holidays count
- Holidays from other years do not affect the current year's count

#### Recalculation Trigger

The page computes all values live on every page load. There are no stored or cached workday values. Any change to the holiday calendar (US-003) is automatically reflected on the next visit to this page — no manual refresh or synchronization step is required.

### State Transitions

This is a read-only computed view. There are no entity state transitions. The page transitions are:

| Page State | Trigger | Next State |
|-----------|---------|-----------|
| Initial / Year changed | User loads page or selects a new year | Loading |
| Loading | Data fetch completes successfully | Loaded (with or without holidays) |
| Loading | Data fetch fails | Error |
| Error | User clicks Retry | Loading |
| Loaded — no holidays | User navigates to Holiday Management and adds holidays, then returns | Loaded — with holidays |

### Dependencies

| Dependency | Story | Impact |
|-----------|-------|--------|
| Holiday Management | EP-009 US-003 | Provides the holiday calendar data used in the Holidays column. If US-003 has no holidays for the selected year, Holidays = 0 for all months. Monthly Workdays page is fully functional without holidays — it simply shows 0 for all Holidays cells. |
| Role & Permission Management | EP-001 US-004 | Provides the `organization.workdays.view` permission assignment. Without this permission, users cannot access the page. |

---

## §7 UX Optimizations

### Action Bar Layout

The action bar contains only the year dropdown, positioned on the left. The right side is intentionally empty — this is a read-only page with no creation, search, or export actions.

This intentionally deviates from the standard list page pattern (which includes search, filters, and pagination) because the Monthly Workdays page is a fixed-row computed summary, not a browsable dataset.

### Loading Experience

- On page load and on year change, a skeleton of 12 rows + Total row is displayed immediately
- Column headers are visible during skeleton loading
- No layout shift occurs when data arrives — skeleton rows match the final table dimensions
- Loading is silent (no spinner overlay) — skeleton rows communicate progress without blocking the interface

### Year Dropdown Behavior

- Defaults to the current calendar year on every fresh page load
- On change: table reloads immediately for the selected year
- No confirmation dialog required — selecting a year is a non-destructive, reversible action
- If no value is selected (edge case), the dropdown retains its current value (cannot be cleared)

### Info Note (No Holidays)

- Displayed below the action bar, above the table
- Uses a non-intrusive informational style (not an error)
- Includes a direct link to the Holiday Management page to reduce the steps needed to configure holidays
- The table still renders normally with Holidays = 0 — the note is supplementary, not a blocker

### Error State

- Error message replaces the table content
- Retry button allows the user to recover without a full page refresh
- Error message is concise — it does not expose technical details

### Responsive Behavior

| Viewport | Behavior |
|----------|---------|
| Desktop (≥ 1280px) | Full table layout with all 5 columns visible |
| Below desktop | Out of scope for this story — web admin is desktop-only |

### Accessibility Requirements

- [x] Year dropdown is keyboard-navigable (Tab to focus, Enter/Space to open, arrow keys to navigate options)
- [x] The info note link ("Holiday Management") is reachable and activatable via keyboard
- [x] Retry button is keyboard-accessible
- [x] Table has proper column headers (ARIA `<th>` or equivalent) for screen reader support
- [x] The pinned Total row is visually distinguishable from data rows (e.g., bold text or background differentiation)
- [x] Skeleton loading state does not trap keyboard focus
- [x] Color contrast meets WCAG AA for all text elements (month names, numeric values, info note)

### Design References

- Approved design spec: `docs/superpowers/specs/2026-06-16-monthly-workdays-design.md`
- No Figma frame exists for this page yet

---

## §8 Additional Information

### Out of Scope

The following are explicitly excluded from this detail requirement:

| Item | Reason |
|------|--------|
| Workday overrides per month | No mechanism to customize auto-calculated values |
| Half-day holiday handling | Holidays counted as whole calendar days only |
| Custom non-working days (ad-hoc closures) | Only company holidays from US-003 are supported |
| Export / download of the workday table | Read-only reference page; no export feature |
| Payroll calculation integration | Future DR — this page is the data foundation only |
| Mobile view | Web admin platform only |
| Per-department or per-employee workday calendars | Org-wide only; no department/employee segmentation |
| Partial month calculations | Always full calendar months, January through December |
| Bulk mutations or data management | Page is entirely read-only |

### Open Questions

| # | Question | Owner | Status |
|---|----------|-------|--------|
| OQ-01 | Should users without `organization.workdays.view` see the "Monthly Workdays" menu item in the sidebar at all, or should it be hidden? | Product Owner | Pending |
| OQ-02 | What years should the year dropdown list? (e.g., current year only, current ± 5 years, all years with holiday data, or a fixed range?) | Product Owner | Pending |
| OQ-03 | When payroll is built in a future DR, will it query this page's live formula directly, or store a snapshot of workday counts at payroll run time? | Development Team | Pending (future DR) |

### Related Features

| Feature | Relationship |
|---------|-------------|
| Holiday Management (EP-009 US-003) | Data source for the Holidays column — holidays added/edited/deleted in US-003 are reflected immediately on next Monthly Workdays load |
| Role & Permission Management (EP-001 US-004) | Source of the `organization.workdays.view` permission assignment |
| Future Payroll DR | Monthly Workdays provides the authoritative workday count foundation for payroll calculation |

### Notes

- This is the first Detail Requirement for a pure read-only computed summary table in this project. No stored values exist — all calculations are performed live on each page load.
- The holiday-weekend overlap rule (a holiday on a weekend still reduces Workdays) is explicitly per the design spec and should not be "corrected" during development — it is the intended business behavior.
- The cross-month holiday split rule is inherited from the holiday calendar data model in US-003 (DR-009-003-01).

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Product Owner | — | — | Pending |
| BA Lead | — | — | Pending |
| Tech Lead | — | — | Pending |

---

## Document Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-06-16 | Initial draft | BA Agent (DR-009-004-01) |
