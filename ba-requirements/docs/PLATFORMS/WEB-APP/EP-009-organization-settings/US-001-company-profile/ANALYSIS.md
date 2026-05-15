---
document_type: ANALYSIS
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
  - path: "./FLOWCHART.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
  - path: "../US-002-attendance-settings/ANALYSIS.md"
    relationship: sibling
revision_history: []
input_sources:
  - type: figma
    file_id: "Exn-HR-Design"
    node_id: "3312:17306"
    frame_name: "Organization Settings > Company Profile"
    extraction_date: "2026-04-23"
add_on_sections:
  - "Design Context [ADD-ON]"
---

# Analysis: Company Profile

**Epic:** EP-009 (Organization Settings)
**Story:** US-001-company-profile
**Status:** Draft

---

## Business Context

Organizations need to maintain their company profile information within the HR system. This information is used throughout the system for display purposes and may appear on official documents, exports, and employee-facing interfaces.

The Company Profile settings page allows administrators to manage core organization information from a centralized location.

---

## Scope

### In Scope (Current Phase)

- Company address management (view, edit, save)

### In Scope (Future Phases)

- Company name management
- Company logo upload
- Additional profile fields as needed

### Out of Scope

- Branch/office location management (separate feature)
- Multi-company/tenant configuration
- Legal entity information

---

## Open Questions

- [ ] What address fields are required? (Street, City, State, ZIP, Country?) — Owner: Product Owner
- [ ] Is company name editable or system-configured? — Owner: Product Owner

---

## Notes

This story shares the Organization Settings page with US-002 (Attendance Settings). The page uses a tab-based layout:
- **Company Profile tab** (this story)
- **Attendance tab** (US-002)

---

## Design Context [ADD-ON]

### Source Information

| Attribute | Value |
|-----------|-------|
| Figma File | Exn-HR-Design |
| Node ID | 3312:17306 |
| Frame Name | Organization Settings > Company Profile |
| Extraction Date | 2026-04-23 |
| Figma URL | https://www.figma.com/design/Exn-HR-Design?node-id=3312-17306 |

### Layout Overview

```
+------------------------------------------------------------------+
|  Sidebar (200px)  |  Main Content Area                           |
|                   |  +------------------------------------------+ |
|  [Logo]           |  | Breadcrumb                               | |
|  [User Profile]   |  +------------------------------------------+ |
|  ───────────────  |  | Organization Settings (H3)               | |
|  Users Management |  |                                          | |
|  Organization Data|  | [Company Profile] [Attendance]  (tabs)   | |
|  Menu Section     |  |                                          | |
|                   |  | +--------------------------------------+ | |
|                   |  | | Company Profile (Card 600px)        | | |
|                   |  | |                                      | | |
|                   |  | | * Company Address                    | | |
|                   |  | | [Search for address input + icon]    | | |
|                   |  | |                                      | | |
|                   |  | | [Map Preview (576x263px)]            | | |
|                   |  | |         [Pin Icon]                   | | |
|                   |  | +--------------------------------------+ | |
|                   |  |                                          | |
|                   |  | [         Save Button (600px)         ]  | |
|                   |  +------------------------------------------+ |
+------------------------------------------------------------------+
```

### Component Inventory

| Component | Node ID | Type | Purpose |
|-----------|---------|------|---------|
| Tab: Company Profile | 3312:17368 | Button | Active tab with accent background |
| Tab: Attendance | 3312:17369 | Button | Inactive tab (white background) |
| Company Address Input | 3312:17376 | Vertical Field | Address search input with location icon |
| Map Preview | 3312:17443 | Rounded Rectangle | Displays selected location on map |
| Map Pin | 3312:17483 | MapPin Icon | Marks selected location on map |
| Save Button | 3312:17380 | Button | Full-width (600px), black background (#010101) |

### Design Constraints

| Constraint | Value |
|------------|-------|
| Form Card Width | 600px |
| Tab Panel Width | 165px |
| Map Preview Size | 576px x 263px |
| Save Button Height | 40px |
| Input Field Height | 36px (min-height) |
| Border Radius (Card) | 6px |
| Border Radius (Button) | 6px |
| Primary Button Background | #010101 |
| Primary Button Text | #fff1f2 |

### Design Tokens Referenced

| Token | Value | Usage |
|-------|-------|-------|
| general/background | #ffffff | Card background |
| general/border | #e5e5e5 | Card border, input border |
| general/accent | #f5f5f5 | Active tab background |
| general/foreground | #0a0a0a | Text color |
| background/bg-primary | #010101 | Save button background |
| text/text-primary-foreground | #fff1f2 | Save button text |
| rounded-md | 6px | Border radius |
| spacing/5 | 20px | Card gap |

### User Flow from Design

1. User navigates to Organization Settings page
2. "Company Profile" tab is active by default (accent background)
3. User sees current company address in input field (or placeholder "Search for address")
4. User clicks input field to search/enter address
5. Map preview updates to show selected location with pin marker
6. User clicks "Save" button to persist changes
