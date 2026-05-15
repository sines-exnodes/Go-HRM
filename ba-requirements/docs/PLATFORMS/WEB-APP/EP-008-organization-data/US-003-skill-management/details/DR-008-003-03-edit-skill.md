---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-008
story_id: US-003
story_name: "Skill Management"
detail_id: DR-008-003-03
detail_name: "Edit Skill"
parent_requirement: FR-US-003-05
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
  - path: "./DR-008-003-02-create-skill.md"
    relationship: sibling
input_sources: []
---

# Detail Requirement: Edit Skill

**Detail ID:** DR-008-003-03
**Parent Requirement:** FR-US-003-05
**Story:** US-003-skill-management
**Epic:** EP-008 (Organization Data)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with organization data management permission**, I want to **edit an existing skill's name, description, and icon**, so that **skill definitions stay accurate as the organization's competency needs evolve**.

**Purpose:** Allow authorized administrators to update any aspect of an existing skill — its name, description, or icon image. Changes are reflected immediately across all HR modules that reference the skill. This keeps the skill catalog accurate without requiring skills to be deleted and recreated.

**Target Users:** Any user with org data management permission (configured via US-004). Users with view-only permission cannot access this feature.

**Key Functionality:**
- Full-page form pre-filled with the current skill name, description, and icon
- Skill name uniqueness validation (case-insensitive, excludes the current skill's own name)
- Duplicate name check is server-side on submit only — both inline error and error toast displayed
- File upload validation (format and size, client-side)
- Returns to Skill List on success or cancel

---

## 2. User Workflow

**Entry Point:** Skill List → gear icon on the target row → select "Edit"

**Preconditions:**
- User is signed in (US-001 Authentication)
- User has org data management permission (US-004)
- The skill to be edited exists in the list

**Main Flow:**
1. User locates the skill in the Skill List
2. User clicks the gear icon on the skill's row
3. System displays a dropdown with available actions (Edit, Delete)
4. User selects "Edit"
5. System navigates to the Edit A Skill full-page form
6. System loads current skill data from the API
7. Form displays pre-filled: skill name field populated with current name (auto-focused, cursor at end), description field populated with current description, icon file input showing current icon filename (or "No file chosen" if no icon was previously uploaded)
8. User modifies the skill name, description, and/or uploads a new icon
9. User clicks **Save**
10. System trims whitespace from name and description
11. System validates client-side: name is not empty, does not exceed 100 characters; file type and size are valid (if new file uploaded)
12. System submits to server; server validates name uniqueness excluding the current skill (case-insensitive)
13. System updates the skill record (name, description, icon)
14. System displays success toast: "Skill '[skill name]' has been updated"
15. System redirects to Skill List — updated skill visible with new name/description/icon

**Alternative Flows:**

- **Alt 1 — Validation fails (empty name):** System displays inline error below Skill name field: "Skill name is required". Form is not submitted. User corrects and retries.
- **Alt 2 — Validation fails (duplicate name — server-side):** Server returns duplicate error. System displays both an inline error: "Skill name already exists" AND an error toast: "Skill name already exists". User enters a different name.
- **Alt 3 — Validation fails (max length):** System displays inline error: "Skill name must not exceed 100 characters". Form is not submitted.
- **Alt 4 — Unchanged values:** All fields identical to current values (after trimming) — system saves successfully. No false duplicate error is raised.
- **Alt 5 — Invalid file type:** System displays inline error below Skill Icon: "Only PNG, JPG, JPEG, WEBP, and SVG files are accepted". File is not uploaded.
- **Alt 6 — File too large:** System displays inline error below Skill Icon: "File size must not exceed 2MB". File is not uploaded.
- **Alt 7 — Cancel (form modified):** System shows confirmation dialog: "Discard unsaved changes?" User clicks Confirm → redirects to Skill List. User clicks Cancel → stays on form.
- **Alt 8 — Cancel (form untouched):** System redirects to Skill List immediately without confirmation.

**Exit Points:**
- **Success:** Skill updated → toast "Skill '[name]' has been updated" → redirect to Skill List
- **Cancel:** Redirect to Skill List (with or without confirmation depending on form state)
- **Error:** Validation errors shown inline (+ error toast for server-side duplicate); user corrects and retries

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Max Length | Description |
|------------|------------|-----------------|-----------|---------------|------------|-------------|
| Skill name | Text input | Not empty; unique excluding self (case-insensitive, **server-side on submit**); allowed characters: letters (a-z, A-Z), numbers (0-9), spaces, hyphens (-), ampersands (&); leading/trailing whitespace trimmed; validated on Save | Yes (*) | Pre-filled with current name | 100 characters | The display name for the skill being edited. Placeholder: "Enter skill name" (shown only if field is cleared) |
| Description | Text input | Trimmed; no uniqueness check; no format restriction | No | Pre-filled with current description | 500 characters | Brief description of the skill. Placeholder: "Enter skill description" (shown only if field is cleared) |
| Skill Icon | File input | Accepted formats: PNG, JPG, JPEG, WEBP, SVG; max file size: 2MB; validated **client-side** on file selection | No | Current icon filename displayed (or "No file chosen" if none) | N/A | Custom icon image. Replaces existing icon on save. |

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| Cancel | Button (secondary) | Left in action bar | Always visible | If form dirty → "Discard changes?" dialog; if clean → redirect to Skill List | Discard and return to list |
| Save | Button (primary) | Right in action bar | Always visible; disabled + spinner while saving | Validate → save → toast → redirect to Skill List | Save changes and return to list |

**Key Difference from Create:** No "Save & Create Another" button — Edit always returns to the Skill List on save. Only 2 buttons (Cancel + Save).

**Validation Error Messages:**

| Condition | Error Message | Display Location |
|-----------|--------------|------------------|
| Name is empty or whitespace-only on Save | "Skill name is required" | Inline, below Skill name field |
| Name matches a different existing skill (case-insensitive) — server-side | "Skill name already exists" | **Both:** inline below Skill name field AND error toast |
| Name exceeds 100 characters | "Skill name must not exceed 100 characters" | Inline, below Skill name field |
| File type not accepted | "Only PNG, JPG, JPEG, WEBP, and SVG files are accepted" | Inline, below Skill Icon field |
| File exceeds 2MB | "File size must not exceed 2MB" | Inline, below Skill Icon field |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Skill name field | Text input | Placeholder: "Enter skill name" (if cleared) | Pre-filled with current name | Identity of the skill being edited |
| Description field | Text input | Placeholder: "Enter skill description" (if cleared) | Pre-filled with current description | Description of the skill |
| Skill Icon field | File input | "No file chosen" (if no icon uploaded previously) | Current icon filename displayed | Custom icon image for the skill |
| Icon preview | Image thumbnail | Hidden if no icon exists | Shows current icon; updates if new file selected | Visual confirmation of icon |
| Page title | Text | Always shown | "Edit A Skill" | Identifies the current action |
| Breadcrumb | Navigation | Always shown | Organization Data / Skills / Edit A Skill | Indicates location in the system |
| Section header | Text | Always shown | "Skill Information" | Groups the form fields |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page loads, API call in progress | Skeleton loader for all fields |
| Default | Page loaded, data populated | Pre-filled name, description, icon; Cancel and Save buttons enabled |
| File selected | User selects a new valid file | New filename displayed; preview updates |
| Validation error — empty name | Save clicked with cleared name | Inline error: "Skill name is required" |
| Validation error — duplicate name | Server returns duplicate on submit | Inline error + error toast: "Skill name already exists" |
| Validation error — max length | Name > 100 characters | Inline error: "Skill name must not exceed 100 characters" |
| Validation error — file type | Invalid file type selected | Inline error about accepted formats |
| Validation error — file size | File > 2MB selected | Inline error about max file size |
| Saving | Save clicked, request in progress | Save button shows spinner + disabled; Cancel button disabled |
| Success | Skill updated | Toast: "Skill '[name]' has been updated" → redirect to Skill List |
| Discard confirmation | Cancel clicked with modified form | Modal: "Discard unsaved changes?" with Confirm and Cancel buttons |

### Page Layout (Pending Design Delivery)

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Organization Data / Skills / Edit A Skill         │
├──────────────┬──────────────────────────────────────────────────┤
│  [Sidebar]   │  Edit A Skill                    [Cancel]  [Save] │
│              │                                                  │
│              │  ┌──────────────────────────────────────────┐    │
│              │  │ Skill Information                        │    │
│              │  │                                          │    │
│              │  │  * Skill name                            │    │
│              │  │  ┌──────────────────────────────────────┐│    │
│              │  │  │ [Current skill name]                 ││    │
│              │  │  └──────────────────────────────────────┘│    │
│              │  │  [error message if any]                   │    │
│              │  │                                          │    │
│              │  │  Description                             │    │
│              │  │  ┌──────────────────────────────────────┐│    │
│              │  │  │ [Current description]                ││    │
│              │  │  └──────────────────────────────────────┘│    │
│              │  │                                          │    │
│              │  │  Skill Icon                              │    │
│              │  │  ┌──────────────────────────────────────┐│    │
│              │  │  │ Choose File   [current-icon.png]     ││    │
│              │  │  └──────────────────────────────────────┘│    │
│              │  │  [error message if any]                   │    │
│              │  └──────────────────────────────────────────┘    │
└──────────────┴──────────────────────────────────────────────────┘
```

> **Note:** Figma design for Edit Skill is pending delivery from the Design Team. Layout mirrors "Create A Skill" (DR-008-003-02, Figma node `3089:1946`) with pre-filled data.

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Form Display:**
- **AC-01:** Edit Skill page displays with page title "Edit A Skill" and one section card: "Skill Information"
- **AC-02:** Skill name field is pre-filled with the current skill name and marked as mandatory (*)
- **AC-03:** Skill name field is auto-focused when the form opens, with cursor positioned at the end of the current name
- **AC-04:** Description field is pre-filled with the current description (or empty placeholder if no description exists)
- **AC-05:** Skill Icon field shows the current icon filename (or "No file chosen" if no icon was previously uploaded)
- **AC-06:** Two buttons are visible: Cancel and Save (no "Save & Create Another")

**Validation — Client-side (instant):**
- **AC-07:** User cannot save with an empty skill name — inline error "Skill name is required"
- **AC-08:** User cannot save with a name exceeding 100 characters — inline error "Skill name must not exceed 100 characters"
- **AC-09:** Uploading a file that is not PNG, JPG, JPEG, WEBP, or SVG shows inline error "Only PNG, JPG, JPEG, WEBP, and SVG files are accepted"
- **AC-10:** Uploading a file larger than 2MB shows inline error "File size must not exceed 2MB"

**Validation — Server-side (on submit):**
- **AC-11:** If skill name matches a different existing skill (case-insensitive), server returns error and system displays both an inline error "Skill name already exists" AND an error toast "Skill name already exists"
- **AC-12:** Saving with the same name as the current skill (unchanged) succeeds without a duplicate error
- **AC-13:** Duplicate name check cannot be performed before form submission — it is a server-side validation only

**Save Behavior:**
- **AC-14:** Save updates the skill, shows success toast "Skill '[name]' has been updated", and redirects to Skill List
- **AC-15:** Skill can be saved with an empty description (cleared description is valid)
- **AC-16:** Skill can be saved without changing the icon (existing icon is preserved)
- **AC-17:** Uploading a new icon replaces the existing icon
- **AC-18:** Leading and trailing whitespace is trimmed from skill name and description before saving
- **AC-19:** Updated skill name, description, and icon are reflected on the Skill List immediately after save

**Cancel Behavior:**
- **AC-20:** Cancel on an untouched form redirects to Skill List without confirmation
- **AC-21:** Cancel on a modified form shows "Discard unsaved changes?" dialog — Confirm discards and redirects, Cancel stays on form
- **AC-22:** "Form modified" is detected by comparing current name, description, and icon state against the original values loaded from API

**Access Control:**
- **AC-23:** Edit Skill page is accessible only to users with org data management permission
- **AC-24:** Direct URL access by unauthorized users redirects to an appropriate fallback page
- **AC-25:** The gear icon with Edit option is hidden for users without management permission

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — change all fields | Rename "Java" → "Java SE", update desc, new icon | Skill updated, success toast, redirect to list | High |
| Happy path — name only | Change "Python" → "Python 3" | Skill updated, success toast, redirect | High |
| Happy path — no changes | Open Edit, click Save immediately | Save succeeds, success toast, redirect | High |
| Happy path — new icon only | Keep name/desc, upload new PNG | Skill updated with new icon, toast, redirect | Medium |
| Happy path — clear description | Clear description, Save | Skill saved with empty description | Medium |
| Empty name | Clear name, click Save | Inline error "Skill name is required" | High |
| Duplicate name (other skill, server-side) | "React" when "react" exists as different skill | Inline error + error toast "Skill name already exists" | High |
| Same name — different case | Current: "Java"; Enter: "java" | Save succeeds (same skill, self-exclusion) | Medium |
| Max length exceeded | 101-character name | Inline error "Skill name must not exceed 100 characters" | Medium |
| Invalid file type | Upload .pdf file | Inline error about accepted formats | Medium |
| File too large | Upload 5MB PNG | Inline error "File size must not exceed 2MB" | Medium |
| Cancel dirty form | Change name, click Cancel | "Discard unsaved changes?" dialog | Medium |
| Cancel clean form | No changes, click Cancel | Redirect to Skill List immediately | Low |
| Unauthorized access | User without permission visits Edit Skill URL | Redirect / access denied | High |
| Whitespace trimming | Name "  React  " | Saved as "React" | Medium |

---

## 6. System Rules

**Business Logic:**
- **SR-01:** Skill name must be unique organization-wide (case-insensitive comparison), **excluding the current skill being edited**. Saving with the unchanged name is always valid.
- **SR-02:** Skill name is trimmed of leading/trailing whitespace before the uniqueness check and before saving.
- **SR-03:** Description is trimmed of leading/trailing whitespace before saving.
- **SR-04:** A skill can be saved with an empty description — description can be cleared.
- **SR-05:** If no new icon is uploaded, the existing icon is preserved unchanged.
- **SR-06:** If a new icon is uploaded, it replaces the existing icon. The old icon is no longer available.
- **SR-07:** Skill name uniqueness is checked **server-side on form submission only** — not on blur or while typing.
- **SR-08:** If server returns a duplicate name error, both an inline error below the Skill name field AND an error toast are displayed to the user.
- **SR-09:** Accepted icon file formats: PNG, JPG, JPEG, WEBP, SVG. Max file size: 2MB. File type and size are validated client-side on file selection.
- **SR-10:** Updated skill data is immediately reflected across all HR modules that reference this skill — no additional sync required.
- **SR-11:** Only users with org data management permission can access the Edit Skill page. The gear icon with Edit option is hidden for users without this permission.
- **SR-12:** Permissions are configured via US-004 (Role & Permission Management). No role names are hardcoded.
- **SR-13:** "Form dirty" detection compares current form state (name + description + icon) against the original values loaded from the API. Only actual changes trigger the discard confirmation on Cancel.

**State Transitions:**
```
[Skill List] → [Gear → Edit] → [Edit Skill Form (pre-filled)]
[Edit Skill Form] → Save (valid, server OK) → [Toast] → [Skill List (updated)]
[Edit Skill Form] → Save (valid, server: duplicate) → [Inline error + Error toast]
[Edit Skill Form] → Save (invalid, client-side) → [Inline error(s)]
[Edit Skill Form] → Cancel (clean) → [Skill List]
[Edit Skill Form] → Cancel (dirty) → [Confirmation Dialog]
[Confirmation Dialog] → Confirm discard → [Skill List]
[Confirmation Dialog] → Cancel → [Edit Skill Form]
```

**Dependencies:**
- **Depends on:** US-001 (Authentication) — user must be signed in
- **Depends on:** US-004 (Role & Permission Management) — org data management permission required
- **Consumed by:** EP-002 (Employee Management) — updated skill data reflected in employee profiles
- **Consumed by:** Skill List (DR-008-003-01) — updated skill appears with new name/description/icon

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Skill name field is auto-focused on page load with cursor at the end of the pre-filled name — user can edit immediately
- **UX-02:** Save button shows loading spinner while request is in progress, preventing double submission; Cancel button also disabled during save
- **UX-03:** Success toast auto-dismisses after 5 seconds (with manual close option)
- **UX-04:** Error toast (duplicate name from server) persists until user dismisses manually — does not auto-dismiss
- **UX-05:** File input shows current icon filename on load — user can see what's already uploaded without guessing
- **UX-06:** Selecting a new file immediately updates the displayed filename — user can confirm correct file before saving
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
- [x] Keyboard navigable — Tab through Skill name → Description → Skill Icon → Save → Cancel
- [x] Screen reader compatible — labels associated with inputs, error messages announced
- [x] Sufficient color contrast — meets WCAG 2.1 AA standards
- [x] Focus indicators visible — clear focus ring on interactive elements
- [x] Form errors linked to fields via aria-describedby
- [x] File input accessible via keyboard (Enter/Space to open file picker)

**Design References:**
- Figma design for Edit Skill is **pending delivery** from the Design Team
- Expected to follow the same structure as "Create A Skill" (DR-008-003-02, Figma node `3089:1946`)
- Same layout as Create — identical card and fields, with pre-filled data and 2 buttons (no "Save & Create Another")

---

## 8. Additional Information

### Out of Scope
- Creating a new skill from the Edit page (no "Save & Create Another" button)
- Removing an existing icon without replacing it (icon can be replaced but not deleted)
- Icon cropping or image editing within the upload flow
- Viewing or managing which employees are assigned to this skill from the Edit page
- Audit log or history of skill changes visible on the Edit page
- Bulk editing of multiple skills at once
- Skill duplication / cloning from the Edit page

### Open Questions
- [ ] **Edit Skill Figma screen:** When will the Design Team deliver the Edit Skill screen? Expected to mirror Create Skill with pre-filled data. — **Owner:** Design Team — **Status:** Pending
- [ ] **Icon removal:** Can an administrator remove an existing icon without uploading a replacement (revert to default placeholder)? Or must they always replace? — **Owner:** Product Owner — **Status:** Pending

### Related Features
- **DR-008-003-01:** Skill List — Edit is triggered from the gear icon; on success, returns here with updated data
- **DR-008-003-02:** Create Skill — Shares the same full-page form layout; Edit differs in pre-filled data, 2 buttons (no "Save & Create Another"), and self-exclusion uniqueness check
- **DR-008-003-04:** Delete Skill (planned) — Accessible from the same gear icon dropdown
- **US-001:** Authentication — user must be signed in
- **US-004:** Role & Permission Management — access control
- **EP-002:** Employee Management — updated skill data reflected in employee profiles

### Notes
- The Edit and Create forms are visually identical — same card, same 3 fields, same styling. The key differences are: (1) page title "Edit A Skill" vs "Create A Skill", (2) all fields pre-filled, (3) only 2 buttons — Cancel and Save (no "Save & Create Another"), (4) uniqueness check excludes the current skill.
- Duplicate name validation is **server-side only** — the system cannot check for existing names before submission. If server returns duplicate, both inline error AND error toast are shown.
- The icon file input on Edit shows the current icon's filename. If the user selects a new file, it replaces the existing icon on save. If no new file is selected, the existing icon is preserved.
- The "Form dirty" detection (SR-13) must compare all 3 fields — name, description, and icon selection — against the original API values to determine whether to show the discard confirmation dialog.

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
| 1.0 | 2026-03-24 | BA Agent | Initial draft — mirrors Create Skill (DR-008-003-02) with pre-filled data, self-exclusion uniqueness, and 2-button layout |
