---
document_type: ANALYSIS
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
  - path: "./FLOWCHART.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
  - path: "../US-003-holiday-management/ANALYSIS.md"
    relationship: dependency
revision_history: []
input_sources:
  - type: text
    description: "Brainstorming session — design spec approved 2026-06-16"
    extraction_date: "2026-06-16"
---

# Analysis: Monthly Workdays

**Epic:** EP-009 (Organization Settings)
**Story:** US-004-monthly-workdays
**Status:** Draft

---

## Business Context

HR and payroll administrators need to know how many workdays exist in each month of the year to support payroll calculation (e.g., computing daily salary rates, OT thresholds, and pro-rated salaries). Currently this must be calculated manually — error-prone and time-consuming, especially when company holidays shift the count.

This story provides a read-only auto-calculated reference page that shows the exact workday count per month for any selected year, derived from the company holiday calendar (EP-009 US-003) and weekend exclusion (Saturdays + Sundays).

---

## Scope

### In Scope

- Monthly Workdays page under Organization Settings (EP-009 US-004)
- Year filter dropdown (defaults to current year)
- 12-row table (one per month, January–December) + pinned Total row
- Columns: Month, Total Days, Weekends, Holidays, Workdays
- Pure computed view — calculated live on each page load from holiday calendar
- Info note when no holidays configured for selected year
- Permission: `organization.workdays.view` via US-004

### Out of Scope

- Mutations of any kind (read-only page, no Add/Edit/Delete)
- Workday overrides per month
- Half-day holiday handling (holidays counted as whole days)
- Export / download
- Payroll calculation integration (future DR)
- Mobile view
- Per-department or per-employee workday calendars

---

## Open Questions

- [ ] Should users without `organization.workdays.view` see the menu item at all, or is it hidden? — Owner: Product Owner
- [ ] When payroll is built, will it query this page's formula directly or store a snapshot? — Owner: Development Team (future)

---

## Notes

- Design spec: `docs/superpowers/specs/2026-06-16-monthly-workdays-design.md`
- No Figma frame exists for this page yet
- **Holiday-weekend overlap rule:** A holiday falling on a weekend still counts in the Holidays column and reduces Workdays — weekends and holidays are not mutually exclusive
- **Leap year:** February shows 29 Total Days automatically; no config needed
- **Cross-month holidays:** A holiday spanning two months is split — each month receives only the days that fall within it
- **Calculation trigger:** Live on every page load — no caching; holiday calendar changes reflect immediately on next visit
