---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-003
story_name: "Skill Management"
detail_id: DR-008-003-02
detail_name: "Create Skill"
parent_requirement: FR-US-003-04
status: draft
version: "1.0"
created_date: "2026-03-24"
last_updated: "2026-03-24"
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "../TODO.yaml"
    relationship: sibling
  - path: "./DR-008-003-01-skill-list.md"
    relationship: sibling
input_sources:
  - type: figma
    description: "Create A Skill screen"
    node_id: "3089:1946"
    extraction_date: "2026-03-24"
---

# Detail Requirement: Create Skill

**Detail ID:** DR-008-003-02
**Parent Requirement:** FR-US-003-04
**Story:** US-003-skill-management
**Epic:** EP-008 (Organization Data)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with organization data management permission**, I want to **create a new skill by providing a name, optional description, and optional icon**, so that **the skill becomes available for employee competency tracking across all HR modules**.

**Purpose:** Allow administrators to add skills to the organization's centralized skill catalog. Skills are reference data — once created, they are immediately available for assignment to employees in EP-002 and for use in training, performance, and recruitment modules. Without this feature, the skill catalog cannot grow as the organization's competency needs evolve.

**Target Users:** Any user with org data management permission (configured via US-004). Users with view-only permission cannot access this feature.

**Key Functionality:**
- 3-field form: Skill name (required), Description (optional), Skill Icon (optional file upload)
- Name uniqueness validation (case-insensitive, server-side on submit)
- File upload validation (format and size, client-side)
- Option to save and continue creating another skill without leaving the page
- Returns to Skill List on success (Save) or cancel

---

## 2. User Workflow

**Entry Point:** Skill List → click "+ Add New" button (visible only to users with management permission)

**Preconditions:**
- User is signed in (US-001 Authentication)
- User has org data management permission (US-004)

**Main Flow:**
1. User clicks "+ Add New" on the Skill List page
2. System navigates to the Create A Skill page
3. Form displays with "Skill Information" card: empty Skill name field (auto-focused), empty Description field, Skill Icon file input ("Choose File / No file chosen")
4. User enters a skill name
5. User optionally enters a description
6. User optionally uploads a skill icon (image file)
7. User clicks **Save** or **Save & Create Another**
8. System trims whitespace from name and description
9. System validates client-side: name is not empty, does not exceed 100 characters; file type and size are valid (if uploaded)
10. System submits to server; server validates name uniqueness (case-insensitive)
11. If validation passes: system saves the skill
12. System displays success toast: "Skill '[skill name]' has been created"
13. If **Save**: system redirects to Skill List (new skill visible in list)
14. If **Save & Create Another**: system clears the form, resets scroll to top, auto-focuses Skill name field, user remains on Create Skill page

**Alternative Flows:**

- **Alt 1 — Validation fails (empty name):** System displays inline error below Skill name field: "Skill name is required". Form is not submitted. User corrects and retries.
- **Alt 2 — Validation fails (duplicate name — server-side):** Server returns duplicate error. System displays both an inline error below Skill name field: "Skill name already exists" AND an error toast: "Skill name already exists". User enters a different name.
- **Alt 3 — Validation fails (max length):** System displays inline error: "Skill name must not exceed 100 characters". Form is not submitted.
- **Alt 4 — Invalid file type:** System displays inline error below Skill Icon: "Only PNG, JPG, JPEG, WEBP, and SVG files are accepted". File is not uploaded.
- **Alt 5 — File too large:** System displays inline error below Skill Icon: "File size must not exceed 2MB". File is not uploaded.
- **Alt 6 — Cancel (form modified):** System shows confirmation dialog: "Discard unsaved changes?" User clicks Confirm → redirects to Skill List. User clicks Cancel → stays on form.
- **Alt 7 — Cancel (form untouched):** System redirects to Skill List immediately without confirmation.

**Exit Points:**
- **Success (Save):** Skill created → toast → redirect to Skill List
- **Success (Save & Create Another):** Skill created → toast → form cleared, stay on page
- **Cancel:** Redirect to Skill List (with or without confirmation depending on form state)
- **Error:** Validation errors shown inline (+ error toast for server-side duplicate); user corrects and retries

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Max Length | Description |
|------------|------------|-----------------|-----------|---------------|------------|-------------|
| Skill name | Text input | Not empty; unique (case-insensitive, **server-side on submit**); allowed characters: letters (a-z, A-Z), numbers (0-9), spaces, hyphens (-), ampersands (&); leading/trailing whitespace trimmed; validated on Save | Yes (*) | Empty | 100 characters | The display name for the new skill. Placeholder: "Enter skill name" |
| Description | Text input | Trimmed; no uniqueness check; no format restriction | No | Empty | 500 characters | Brief description of the skill. Placeholder: "Enter skill description" |
| Skill Icon | File input | Accepted formats: PNG, JPG, JPEG, WEBP, SVG; max file size: 2MB; validated **client-side** on file selection | No | No file chosen | N/A | Custom icon image for the skill. Displayed in the Icon column on the Skill List. |

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| Cancel | Button (secondary) | Left in action bar | Always visible | If form dirty → "Discard changes?" dialog; if clean → redirect to Skill List | Discard and return to list |
| Save & Create Another | Button (secondary/outline) | Center-right in action bar | Always visible; disabled + spinner while saving | Validate → save → toast → clear form, stay on page | Save and continue creating |
| Save | Button (primary) | Right in action bar | Always visible; disabled + spinner while saving | Validate → save → toast → redirect to Skill List | Save and return to list |

**Validation Error Messages:**

| Condition | Error Message | Display Location |
|-----------|--------------|------------------|
| Name is empty or whitespace-only on Save | "Skill name is required" | Inline, below Skill name field |
| Name already exists (case-insensitive) — server-side | "Skill name already exists" | **Both:** inline below Skill name field AND error toast |
| Name exceeds 100 characters | "Skill name must not exceed 100 characters" | Inline, below Skill name field |
| File type not accepted | "Only PNG, JPG, JPEG, WEBP, and SVG files are accepted" | Inline, below Skill Icon field |
| File exceeds 2MB | "File size must not exceed 2MB" | Inline, below Skill Icon field |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Skill name field | Text input | Placeholder: "Enter skill name" | Free text | Identity of the skill being created |
| Description field | Text input | Placeholder: "Enter skill description" | Free text | Brief description of the skill's purpose |
| Skill Icon field | File input | "Choose File / No file chosen" | Browser native file input | Custom icon image for the skill |
| Icon preview | Image thumbnail | Hidden (no preview until file selected) | Small image preview next to file input after selection | Confirms correct file was selected |
| Page title | Text | Always shown | "Create A Skill" | Identifies the current action |
| Breadcrumb | Navigation | Always shown | Organization Data / Skills / Create A Skill | Indicates location in the system |
| Section header | Text | Always shown | "Skill Information" | Groups the form fields |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default | Page loads successfully | Empty name field (auto-focused), empty description, file input "No file chosen", 3 buttons enabled |
| File selected | User selects a valid file | Filename displayed next to "Choose File"; optional preview thumbnail |
| Validation error — empty name | Save clicked with blank name | Inline error below field: "Skill name is required" |
| Validation error — duplicate name | Server returns duplicate on submit | Inline error below field: "Skill name already exists" + error toast |
| Validation error — max length | Name > 100 characters | Inline error below field: "Skill name must not exceed 100 characters" |
| Validation error — file type | Invalid file type selected | Inline error below Skill Icon: "Only PNG, JPG, JPEG, WEBP, and SVG files are accepted" |
| Validation error — file size | File > 2MB selected | Inline error below Skill Icon: "File size must not exceed 2MB" |
| Saving | Save/Save & Create Another clicked, request in progress | Clicked button shows spinner + disabled; other buttons disabled |
| Success (Save) | Skill created | Toast: "Skill '[name]' has been created" → redirect to Skill List |
| Success (Save & Create Another) | Skill created | Toast: "Skill '[name]' has been created" → form clears, scroll resets to top |
| Discard confirmation | Cancel clicked with modified form | Modal dialog: "Discard unsaved changes?" with Confirm and Cancel buttons |

### Page Layout (from Figma)

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Organization Data / Skills / Create A Skill       │
├──────────────┬──────────────────────────────────────────────────┤
│  [Sidebar]   │  Create A Skill    [Cancel] [Save&Create] [Save] │
│              │                                                  │
│              │  ┌──────────────────────────────────────────┐    │
│              │  │ Skill Information                        │    │
│              │  │                                          │    │
│              │  │  * Skill name                            │    │
│              │  │  ┌──────────────────────────────────────┐│    │
│              │  │  │ Enter skill name                     ││    │
│              │  │  └──────────────────────────────────────┘│    │
│              │  │  [error message if any]                   │    │
│              │  │                                          │    │
│              │  │  Description                             │    │
│              │  │  ┌──────────────────────────────────────┐│    │
│              │  │  │ Enter skill description              ││    │
│              │  │  └──────────────────────────────────────┘│    │
│              │  │                                          │    │
│              │  │  Skill Icon                              │    │
│              │  │  ┌──────────────────────────────────────┐│    │
│              │  │  │ Choose File   No file chosen         ││    │
│              │  │  └──────────────────────────────────────┘│    │
│              │  │  [error message if any]                   │    │
│              │  └──────────────────────────────────────────┘    │
└──────────────┴──────────────────────────────────────────────────┘
```

> **Note:** Figma shows 2 buttons (Cancel + Save). Confirmed requirement is 3 buttons (Cancel, Save & Create Another, Save) following the Role create pattern. Design should be updated.

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Form Display:**
- **AC-01:** Create Skill page displays with page title "Create A Skill" and one section card: "Skill Information"
- **AC-02:** Skill name field is marked as mandatory (*)
- **AC-03:** Description field is optional (no asterisk)
- **AC-04:** Skill Icon field shows "Choose File / No file chosen" by default
- **AC-05:** Three buttons are visible: Cancel, Save & Create Another, Save

**Validation — Client-side (instant):**
- **AC-06:** User cannot save with an empty skill name — inline error "Skill name is required"
- **AC-07:** User cannot save with a name exceeding 100 characters — inline error "Skill name must not exceed 100 characters"
- **AC-08:** Uploading a file that is not PNG, JPG, JPEG, WEBP, or SVG shows inline error "Only PNG, JPG, JPEG, WEBP, and SVG files are accepted"
- **AC-09:** Uploading a file larger than 2MB shows inline error "File size must not exceed 2MB"

**Validation — Server-side (on submit):**
- **AC-10:** If skill name matches an existing skill (case-insensitive), server returns error and system displays both an inline error "Skill name already exists" AND an error toast "Skill name already exists"
- **AC-11:** Duplicate name check cannot be performed before form submission — it is a server-side validation only

**Save Behavior:**
- **AC-12:** Save creates the skill, shows success toast "Skill '[name]' has been created", and redirects to Skill List
- **AC-13:** Save & Create Another creates the skill, shows success toast "Skill '[name]' has been created", clears the form, and stays on the page
- **AC-14:** After Save & Create Another, scroll position resets to top and Skill name field is auto-focused
- **AC-15:** Skill can be saved without a description (empty description is valid)
- **AC-16:** Skill can be saved without an icon (default placeholder used on Skill List)
- **AC-17:** Leading and trailing whitespace is trimmed from skill name and description before saving

**Cancel Behavior:**
- **AC-18:** Cancel on an untouched form redirects to Skill List without confirmation
- **AC-19:** Cancel on a modified form shows "Discard unsaved changes?" confirmation dialog — Confirm discards and redirects, Cancel stays on form
- **AC-20:** "Form modified" includes any change to name, description, or icon file selection

**Icon Upload:**
- **AC-21:** After selecting a valid file, the filename is displayed next to the file input
- **AC-22:** Uploaded icon is displayed in the Icon column on the Skill List after saving
- **AC-23:** If no icon is uploaded, a default placeholder is shown on the Skill List

**Access Control:**
- **AC-24:** Create Skill page is accessible only to users with org data management permission
- **AC-25:** Direct URL access by unauthorized users redirects to an appropriate fallback page

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — Save with all fields | Name "JavaScript" + description + PNG icon | Skill created, success toast, redirect to list | High |
| Happy path — Save minimal | Name "Python" only, no desc, no icon | Skill created, success toast, redirect to list | High |
| Happy path — Save & Create Another | Name "React" + description | Skill created, success toast, form cleared | High |
| Empty name | Click Save with blank name | Inline error "Skill name is required" | High |
| Duplicate name (server-side) | "Java" when "java" exists | Inline error + error toast "Skill name already exists" | High |
| Case-insensitive duplicate | "PYTHON" when "Python" exists | Inline error + error toast "Skill name already exists" | High |
| Max length exceeded | 101-character name | Inline error "Skill name must not exceed 100 characters" | Medium |
| Invalid file type | Upload .pdf file | Inline error about accepted formats | Medium |
| File too large | Upload 5MB PNG | Inline error "File size must not exceed 2MB" | Medium |
| Valid icon upload | Upload 500KB PNG | Filename displayed, saved with skill | Medium |
| Cancel dirty form | Enter name, click Cancel | "Discard unsaved changes?" dialog | Medium |
| Cancel clean form | No changes, click Cancel | Redirect to Skill List immediately | Low |
| Unauthorized access | User without permission visits URL | Redirect / access denied | High |
| Whitespace trimming | Name "  React  " | Saved as "React" | Medium |

---

## 6. System Rules

**Business Logic:**
- **SR-01:** Skill name must be unique organization-wide (case-insensitive comparison — "JavaScript" = "javascript" = "JAVASCRIPT")
- **SR-02:** Skill name is trimmed of leading/trailing whitespace before saving
- **SR-03:** Description is trimmed of leading/trailing whitespace before saving
- **SR-04:** A skill can be created without a description — description defaults to empty
- **SR-05:** A skill can be created without an icon — system uses a default placeholder icon on the Skill List
- **SR-06:** Skill name uniqueness is checked **server-side on form submission only** — not on blur or while typing
- **SR-07:** If server returns a duplicate name error, both an inline error below the Skill name field AND an error toast are displayed to the user
- **SR-08:** Accepted icon file formats: PNG, JPG, JPEG, WEBP, SVG. Max file size: 2MB. File type and size are validated client-side on file selection.
- **SR-09:** Newly created skill is immediately available for employee assignment across the platform — no additional activation step required
- **SR-10:** Only users with org data management permission can access the Create Skill page
- **SR-11:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.
- **SR-12:** The uploaded icon is stored server-side and served as a URL for display in the Skill List Icon column

**State Transitions:**
```
[Skill List] → "+ Add New" click → [Create Skill Form (empty)]
[Create Skill Form] → Save (valid, server OK) → [Toast] → [Skill List + new skill visible]
[Create Skill Form] → Save (valid, server: duplicate) → [Inline error + Error toast]
[Create Skill Form] → Save & Create Another (valid) → [Toast] → [Create Skill Form (cleared)]
[Create Skill Form] → Save (invalid, client-side) → [Inline error(s)]
[Create Skill Form] → Cancel (clean) → [Skill List]
[Create Skill Form] → Cancel (dirty) → [Confirmation Dialog]
[Confirmation Dialog] → Confirm discard → [Skill List]
[Confirmation Dialog] → Cancel → [Create Skill Form]
```

**Dependencies:**
- **Depends on:** US-001 (Authentication) — user must be signed in
- **Depends on:** US-004 (Role & Permission Management) — org data management permission required
- **Consumed by:** EP-002 (Employee Management) — skills available for employee profile assignment
- **Consumed by:** Skill List (DR-008-003-01) — new skill appears in list after creation

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Skill name field is auto-focused on page load — user can begin typing immediately
- **UX-02:** Save / Save & Create Another buttons show loading spinner while request is in progress, preventing double submission; all buttons disabled during save
- **UX-03:** Success toast auto-dismisses after 5 seconds (with manual close option)
- **UX-04:** Error toast (duplicate name from server) persists until user dismisses manually — does not auto-dismiss
- **UX-05:** After "Save & Create Another", scroll resets to top and Skill name field is re-focused for next entry
- **UX-06:** File input shows selected filename after choosing a file — user can confirm correct file before saving
- **UX-07:** "Discard unsaved changes?" dialog uses default focus on Cancel (stay on form) — prevents accidental data loss
- **UX-08:** Pressing Enter while Skill name or Description field is focused triggers Save
- **UX-09:** Inline validation errors appear directly below the relevant field — not as popups, keeping context clear

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Centered form card, 600px width |
| Tablet (768–1024px) | Full-width form with padding |
| Mobile (<768px) | Stacked layout, full-width inputs and buttons |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through Skill name → Description → Skill Icon → Save & Create Another → Save → Cancel
- [x] Screen reader compatible — labels associated with inputs, error messages announced
- [x] Sufficient color contrast — meets WCAG 2.1 AA standards
- [x] Focus indicators visible — clear focus ring on interactive elements
- [x] Form errors linked to fields via aria-describedby
- [x] File input accessible via keyboard (Enter/Space to open file picker)

**Design References:**
- Figma: [Create A Skill](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3089-1946) (node `3089:1946`)
- Design tokens: See ANALYSIS.md Section 7 — Design Context [ADD-ON]
- Pattern reference: Create Role (DR-001-004-02) — same 3-button layout pattern

---

## 8. Additional Information

### Out of Scope
- Edit Skill — separate detail requirement (DR-008-003-03, planned)
- Delete Skill — separate detail requirement (DR-008-003-04, planned)
- Employee-to-skill assignment — managed in Employee Management (EP-002)
- Skill proficiency levels (beginner, intermediate, expert) — future enhancement
- Skill categories or grouping — future enhancement
- Bulk skill creation or CSV import
- Icon cropping or image editing within the upload flow
- Skill duplication / cloning

### Open Questions
- [ ] **Description max length:** 500 characters assumed — confirm with Product Owner. — **Owner:** Product Owner — **Status:** Pending
- [ ] **Icon dimensions:** Should uploaded icons be resized/cropped to a standard size (e.g., 48×48px)? Or stored at original dimensions? — **Owner:** Product Owner — **Status:** Pending
- [ ] **"Save & Create Another" button:** Not in Figma design (only Cancel + Save shown). Confirmed as requirement following Role pattern — Design Team should update Figma. — **Owner:** Design Team — **Status:** Pending

### Related Features
- **DR-008-003-01:** Skill List — entry point for Create Skill via "+ Add New" button; new skill appears here after creation
- **DR-008-003-03:** Edit Skill (planned) — expected to mirror Create Skill layout with pre-filled data
- **DR-008-003-04:** Delete Skill (planned) — modal confirmation, blocked if employees assigned
- **US-001:** Authentication — user must be signed in
- **US-004:** Role & Permission Management — access control
- **EP-002:** Employee Management — skills created here are available for employee profile assignment

### Notes
- The Figma design shows **2 buttons** (Cancel + Save) but the confirmed requirement is **3 buttons** (Cancel, Save & Create Another, Save) — following the Role create pattern. Design should be updated to reflect this.
- The Skill create form has **3 fields** (name, description, icon) — richer than Department/Position (name only) but simpler than Role (name + permission matrix).
- The icon file upload uses the browser's native file input ("Choose File / No file chosen"). No custom drag-and-drop uploader is designed.
- Skills are NOT listed in the sidebar navigation under Organization Data in the current Figma design — this is a known navigation gap that should be resolved by the Design Team.
- Duplicate name validation is **server-side only** — the system cannot check for existing names before submission. This differs from client-side validations (empty, max length, file type/size) which are instant.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | BA Agent | 2026-03-24 | Draft |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |
| Tech Lead | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-24 | BA Agent | Initial draft — full 8-section detail requirement with Figma design context |
