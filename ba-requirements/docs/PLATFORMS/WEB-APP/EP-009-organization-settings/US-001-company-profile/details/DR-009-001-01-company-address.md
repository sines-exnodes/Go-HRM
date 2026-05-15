---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-009
story_id: US-001
story_name: "Company Profile"
detail_id: DR-009-001-01
detail_name: "Company Address"
status: draft
version: "1.1"
created_date: "2026-04-23"
last_updated: "2026-04-23"
related_documents:
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "../../EPIC.md"
    relationship: parent
  - path: "../../US-002-attendance-settings/details/DR-009-002-01-late-arrival-threshold-configuration.md"
    relationship: sibling
input_sources:
  - type: figma
    file_id: "Exn-HR-Design"
    node_id: "3312:17306"
    frame_name: "Organization Settings > Company Profile"
    extraction_date: "2026-04-23"
  - type: knowledge_base
    description: "Organization Settings pattern from DR-009-002-01 (Late Arrival Threshold)"
    extraction_date: "2026-04-23"
---

# Detail Requirement: Company Address

**Detail ID:** DR-009-001-01
**Parent Story:** US-001-company-profile
**Epic:** EP-009 (Organization Settings)
**Status:** Draft
**Version:** 1.1

---

## 1. Use Case Description

As an **Administrator**, I want to **update the company address** so that **the organization's location is accurately displayed throughout the system and on official documents/exports**.

**Purpose:** Enable administrators to maintain the organization's physical address using an address search field with map preview. The company address is used for display purposes in the HR system, employee-facing interfaces, and may appear on official documents or exports.

**Target Users:** 
- Administrators with organization settings management permission
- Super Admins
- HR Managers (if granted organization settings permission)

**Key Functionality:**
- Search and select company address using an address autocomplete field
- Preview selected location on an embedded map with pin marker
- Save the company address to persist changes immediately

---

## 2. User Workflow

**Entry Point:** Sidebar navigation > Organization Settings > Company Profile tab

**Preconditions:**
- User is authenticated and logged in (EP-001 US-001)
- User has "Organization Settings Management" permission (EP-001 US-004)
- Organization Settings page is accessible via sidebar menu

**Main Flow:**

1. User navigates to Organization Settings from the sidebar
2. System displays Organization Settings page with "Company Profile" tab active by default
3. User views the Company Profile card (600px width) containing the Company Address field
4. System pre-fills the address field with current saved address (or shows placeholder if empty)
5. User clicks on the address input field to begin editing
6. User types an address in the search field
7. System queries address service and displays autocomplete suggestions (debounce ~300ms)
8. User selects an address from the autocomplete dropdown
9. System updates the address field with selected address
10. System updates the map preview (576x263px) to show the selected location with a pin marker
11. User clicks the "Save" button (full-width, black background)
12. System validates and saves the address
13. System displays success toast: "Company address has been updated"
14. User remains on the same page with updated values

**Alternative Flows:**

| Action | Flow |
|--------|------|
| **Clear Address** | User clears the address field > Map preview shows no pin / default view > User clicks Save > Address is removed from system |
| **No Match Found** | User types address with no matches > System shows "No results found" in dropdown > User can type different address |
| **Navigate Away (Dirty)** | User modifies address but navigates away > System shows "Discard unsaved changes?" confirmation dialog |
| **Navigate Away (Clean)** | User navigates away without modifications > No confirmation, immediate navigation |
| **Tab Switch (Dirty)** | User switches to Attendance tab with unsaved changes > System shows "Discard unsaved changes?" confirmation dialog |

**Exit Points:**

- **Success:** Toast message "Company address has been updated", stay on page with saved values
- **Cancel/Discard:** Return to previous page or switch tabs (if dirty form discarded)
- **Error:** Error toast displayed, form remains open for correction

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Company Address | Text (Address Search) | Max 500 characters; address selected from autocomplete | Yes | Empty (placeholder: "Search for address") | Searchable address field with autocomplete from address service; location icon on right side |

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Company Profile Tab | Button (Tab) | Active by default (accent background #f5f5f5) | Click to view Company Profile section | Left panel tab navigation (165px width) |
| Attendance Tab | Button (Tab) | Inactive (white background) | Click to navigate to Attendance Settings (US-002) | Left panel tab navigation |
| Address Input | Input Field | Enabled when user has edit permission | On input: trigger address autocomplete search with ~300ms debounce | Text input (576px width, 36px min-height) with location icon on right side |
| Autocomplete Dropdown | Dropdown List | Visible when typing in address field | Click to select address | Shows matching address suggestions from address service |
| Map Preview | Display Area | Shows selected location | Read-only; updates when address changes | Embedded map (576x263px) with rounded corners (6px) showing selected location |
| Map Pin | Icon (MapPin) | Visible when address is selected | No action | 42x42px pin marker at selected location on map |
| Save Button | Primary Button | Always enabled | Click to save changes | Full-width (600px), 40px height, black background (#010101), white text (#fff1f2), "Save" label |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Company Address | Text | Placeholder: "Search for address" | Full address string from autocomplete selection | Current registered company location |
| Map Preview | Image/Map | Default map view or no pin shown | Embedded map (576x263px) with pin marker | Visual confirmation of selected location |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page loading | Skeleton content blocks for form card |
| Empty (New) | No address saved yet | Address field with placeholder "Search for address", map shows default view without pin |
| Populated | Address exists in system | Pre-filled address field, map showing location with pin marker |
| Searching | User typing in address field | Autocomplete dropdown appears below input with matching addresses |
| No Results | No matching addresses found | Dropdown shows "No results found" message |
| Selected | Address selected from autocomplete | Address text in input field, map updates with pin at selected location |
| Saving | Save button clicked | Save button shows loading spinner, form disabled |
| Success | Save completed | Success toast "Company address has been updated", form re-enabled with saved values |
| Error | Save failed | Error toast, form re-enabled for correction |

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** User with "Organization Settings Management" permission can access Organization Settings > Company Profile tab
- **AC-02:** Company Profile tab is active by default with accent background (#f5f5f5)
- **AC-03:** Company Address field displays current saved address or placeholder "Search for address" when empty
- **AC-04:** Address field has location icon on right side (as per Figma design)
- **AC-05:** Typing in address field triggers autocomplete suggestions from address service with ~300ms debounce
- **AC-06:** Selecting an address from autocomplete updates the input field and map preview
- **AC-07:** Map preview (576x263px) displays selected location with a MapPin marker (42x42px)
- **AC-08:** Save button is full-width (600px), 40px height, black background (#010101), white text
- **AC-09:** Clicking Save persists the address and displays success toast "Company address has been updated"
- **AC-10:** After successful save, user remains on the same page with updated values (no redirect)
- **AC-11:** Navigating away with unsaved changes shows "Discard unsaved changes?" confirmation dialog
- **AC-12:** Users without "Organization Settings Management" permission cannot access this page (redirect to fallback)
- **AC-13:** Address field validates maximum 500 characters

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path - Save new address | Search address, select from autocomplete, click Save | Address saved, map shows pin, toast shown, stay on page | High |
| Update existing address | Change address to new location, Save | New address saved, map updates with new pin location | High |
| Clear address | Clear field, Save | Address removed, map shows no pin | Medium |
| Address autocomplete | Type "Ho Chi Minh" | Autocomplete dropdown shows matching addresses | High |
| No autocomplete results | Type gibberish text | Dropdown shows "No results found" | Medium |
| Cancel with dirty form | Edit address, click back/navigate | "Discard unsaved changes?" dialog appears | High |
| Cancel with clean form | No changes, click back/navigate | Navigate immediately, no dialog | Medium |
| Tab switch with dirty form | Edit address, click Attendance tab | "Discard unsaved changes?" dialog appears | Medium |
| Invalid permission | User without permission accesses URL | Redirect to fallback page | High |
| Network error on save | Save with network failure | Error toast, form stays open for retry | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **Rule 1:** Only one company address is stored per organization (single-tenant system)
- **Rule 2:** Address is stored as a complete string from the selected autocomplete result
- **Rule 3:** Map coordinates (latitude/longitude) are derived from the selected address for map display
- **Rule 4:** Changes apply immediately upon save - no approval workflow required
- **Rule 5:** Address changes are logged for audit trail (timestamp, user who made change)
- **Rule 6:** Only users with "Organization Settings Management" permission can view and modify

**State Transitions:**

```
[No Address] --> [Type in Search] --> [Autocomplete Shown]
[Autocomplete Shown] --> [Select Address] --> [Address Selected + Map Updated]
[Address Selected] --> [Click Save] --> [Address Saved]
[Address Saved] --> [Edit Address] --> [Form Dirty]
[Form Dirty] --> [Navigate Away] --> [Discard Dialog]
[Discard Dialog] --> [Confirm Discard] --> [Navigate Away]
[Discard Dialog] --> [Cancel] --> [Stay on Form]
```

**Calculations/Formulas:**
- None applicable

**Dependencies:**
- Address autocomplete service (Google Places API or similar) for address search
- Map rendering service for preview (Google Maps, Mapbox, or similar)
- EP-001 (Foundation) - Authentication required
- EP-001 US-004 (Role & Permission Management) - Organization settings management permission

---

## 7. UX Optimizations

**Usability Considerations:**

- **Optimization 1:** Address autocomplete with debounce (~300ms) to reduce API calls while typing
- **Optimization 2:** Map preview updates immediately upon address selection for visual confirmation
- **Optimization 3:** Full-width Save button (600px) for easy click target
- **Optimization 4:** Stay on page after save - no redirect needed for single-field settings
- **Optimization 5:** Dirty form detection with discard confirmation prevents accidental data loss
- **Optimization 6:** Location icon in input field provides visual hint for field purpose

**Layout Specifications (from Figma):**

| Element | Specification |
|---------|--------------|
| Page Title | "Organization Settings" (24px, semibold, #09090b) |
| Tab Panel Width | 165px |
| Form Card Width | 600px |
| Form Card Background | White with border (#e4e4e7), rounded corners (6px) |
| Card Title | "Company Profile" (18px, semibold) |
| Address Input Height | 36px min-height |
| Map Preview Size | 576px x 263px |
| Map Border Radius | 6px |
| Save Button | Full-width 600px, 40px height, #010101 background, #fff1f2 text |
| Field Spacing | 20px gap (spacing/5 token) |

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Full layout as designed - sidebar + main content with 600px form card |
| Tablet (768-1024px) | Out of scope for WEB-APP (desktop only) |
| Mobile (<768px) | Out of scope for WEB-APP (desktop only) |

**Accessibility Requirements:**

- [x] Keyboard navigable - Tab through tabs, input, and Save button
- [x] Screen reader compatible - Labels and ARIA attributes for form fields
- [x] Sufficient color contrast - Text on backgrounds meets WCAG AA
- [x] Focus indicators visible - Focus ring on interactive elements
- [x] Autocomplete accessible - Arrow keys to navigate, Enter to select

**Design References:**

- Figma Frame: Organization Settings > Company Profile (Node ID: 3312:17306)
- Design Tokens: See ANALYSIS.md Design Context [ADD-ON] section

---

## 8. Additional Information

### Out of Scope

- Company name editing (future phase)
- Company logo upload (future phase)
- Multiple office/branch locations management
- Address field breakdown (separate street, city, state, ZIP fields)
- Manual coordinate entry (lat/long)
- Mobile/tablet responsive layout
- Address history/audit log UI
- Multi-company/tenant configuration

### Open Questions

- None - all questions resolved from Figma design and knowledge base patterns

### Related Features

- **US-002 Attendance Settings:** Shares the Organization Settings page via tab navigation
- **DR-009-002-01 Late Arrival Threshold:** Uses same page layout and UX patterns
- **EP-001 Foundation:** Provides authentication and permission framework
- **US-004 Role & Permission Management:** Defines who can access organization settings

### Notes

- This feature follows the **Organization Settings Form Pattern** similar to Leave Quota Management (Knowledge Base Section 11)
- Single-field form with address search + map preview
- Full-width Save button and stay-on-page behavior after save
- Tab-based navigation with Company Profile (active) and Attendance tabs

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | BA Agent | 2026-04-23 | Draft |
| Product Owner | | | Pending |
| UX Designer | | | Pending |
| Tech Lead | | | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-04-23 | BA Agent | Initial draft |
| 1.1 | 2026-04-23 | BA Agent | Updated with Figma design extraction - changed from multi-field to address search + map preview per design |
