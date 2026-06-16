---
document_type: ANALYSIS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-009
story_id: US-003
story_name: "Holiday Management"
status: draft
version: "1.0"
last_updated: "2026-06-15"
add_on_sections: []
approved_by: null
related_documents:
  - path: "./FLOWCHART.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
  - path: "../../EP-002-leave-management/US-001-leave-requests/ANALYSIS.md"
    relationship: cross-reference
  - path: "../../EP-004-attendance-management/US-001-attendance-list/ANALYSIS.md"
    relationship: cross-reference
revision_history: []
input_sources:
  - type: text
    description: "Brainstorming session — design spec approved 2026-06-15"
    extraction_date: "2026-06-15"
  - type: figma
    description: "Holidays list frame — nodeId 3537:3821 (list screen only; create/edit frame not yet designed)"
    extraction_date: "2026-06-15"
---

# Analysis: Holiday Management

**Epic:** EP-009 (Organization Settings)
**Story:** US-003-holiday-management
**Status:** Draft

---

## Business Context

HR teams currently have no centralized way to define company holidays within the Exnodes HRM system. Holidays are referenced as a business concept in attendance (streak calculation, matrix display) and leave management (leave day counting), but there is no admin UI to create or maintain them. This means:

- Leave request day counts cannot accurately exclude holidays
- The attendance matrix cannot display confirmed holiday names
- Attendance streak calculations rely on an unconfigured holiday calendar

This story provides the admin interface for creating and maintaining the company holiday calendar on a per-year basis, resolving these gaps across the system.

---

## Scope

### In Scope

- Holiday list page with year filter (dropdown), search by name, pagination
- Create Holiday — full-page form (Holiday Name, From Date, To Date, Total Days auto-calculated)
- Edit Holiday — same full-page form, pre-filled
- Delete Holiday — confirmation dialog with automatic recalculation of affected approved leave requests
- Import From Template — modal to import Vietnamese public holiday preset for a selected year, with row-level preview and selection
- Permission control via US-004 (view vs. manage)
- Auto-recalculation of approved leave request balances when holidays are created, edited, or deleted

### Out of Scope

- Payslip / payroll calculation (future DR)
- Recurring / auto-repeating holidays across years
- Per-department or per-employee holiday exceptions
- Mobile view (web admin only)
- Public holiday API integration (preset is system-maintained)
- Overlap validation between two holidays in the same year
- Status filter on the list

---

## Open Questions

- [ ] Which roles should receive `organization.holidays.manage` permission by default? — Owner: Product Owner
- [ ] When auto-recalculating leave after a holiday edit, should affected employees receive a notification? — Owner: Product Owner
- [ ] How many years of Vietnamese public holiday presets are available in the system? — Owner: Development Team
- [ ] Should HR be able to add custom (non-preset) holiday names alongside imported ones? (Yes — via Add New; confirmed in design) — Owner: N/A (resolved)

---

## Notes

- Design spec: `docs/superpowers/specs/2026-06-15-holiday-management-design.md`
- Figma list frame: nodeId `3537:3821` — confirms columns: Holiday Name, From Date, To Date, Total Days, Action; action bar: Search + "Import From Template" + "Add New"
- Create/Edit Figma frame not yet designed — form follows Create User pattern (DR-001-005-02)
- Leave day formula: `Leave Days Deducted = Calendar Days in range − Company Holiday Days in range`
- Half-day leave on a holiday: 0.5 days excluded proportionally
- Recalculation scope: Approved leave requests only; Pending/Rejected/Cancelled are not affected
