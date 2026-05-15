---
document_type: ANALYSIS
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
  - path: "./FLOWCHART.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
  - path: "../US-001-company-profile/ANALYSIS.md"
    relationship: sibling
  - path: "../../EP-004-attendance-management/US-001-attendance-list/ANALYSIS.md"
    relationship: cross-reference
revision_history: []
input_sources:
  - type: figma
    description: "Organization Settings — Attendance tab design"
    extraction_date: "2026-04-23"
---

# Analysis: Attendance Settings

**Epic:** EP-009 (Organization Settings)
**Story:** US-002-attendance-settings
**Status:** Draft

---

## Business Context

HR and administrators need to configure attendance-related policies that affect how the system evaluates employee check-ins. The primary setting is the **Late Arrival Threshold** — a time after which any employee check-in is marked as "Late" in the attendance system.

This configuration is organization-wide and integrates with:
- **MOBILE-APP EP-003**: Employee check-in captures the actual arrival time
- **WEB-APP EP-004**: Attendance matrix displays "Late" status based on this threshold

---

## Scope

### In Scope

- Late arrival threshold configuration (time picker)
- Organization-wide setting (applies to all employees)
- Immediate effect on future check-ins

### Out of Scope

- Per-department thresholds (future enhancement)
- Per-employee exceptions
- Auto-checkout time configuration
- GPS/location-based attendance rules
- Retroactive application to historical records

---

## Business Rules

1. **Late Determination**: Employee check-in time > Late Arrival Threshold → marked as "Late"
2. **Default Value**: 9:00 AM (if never configured)
3. **No Retroactive Changes**: Updating the threshold does NOT change historical attendance records
4. **Immediate Effect**: New threshold applies to all check-ins after save

---

## Open Questions

None — all questions resolved during DR creation.

---

## Notes

This story shares the Organization Settings page with US-001 (Company Profile). The page uses a tab-based layout:
- **Company Profile tab** (US-001)
- **Attendance tab** (this story)

**Related DR:** [DR-009-002-01-late-arrival-threshold-configuration.md](./details/DR-009-002-01-late-arrival-threshold-configuration.md)
