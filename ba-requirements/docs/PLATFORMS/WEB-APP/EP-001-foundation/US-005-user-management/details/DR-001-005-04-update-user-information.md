---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-005
story_name: "User Management"
detail_id: DR-001-005-04
detail_name: "Update User Information"
parent_requirement: FR-US-005-11
status: draft
version: "1.0"
created_date: "2026-03-26"
last_updated: "2026-03-26"
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "./DR-001-005-01-user-list.md"
    relationship: sibling
  - path: "./DR-001-005-02-create-user.md"
    relationship: sibling
  - path: "./DR-001-005-03-user-details.md"
    relationship: sibling
input_sources:
  - type: figma
    description: "Update Information screen (User Details sub-page)"
    node_id: "3122:6323"
    extraction_date: "2026-03-26"
---

# Detail Requirement: Update User Information

**Detail ID:** DR-001-005-04
**Parent Requirement:** FR-US-005-11
**Story:** US-005-user-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with user management permission**, I want to **update a user's personal information and work profile**, so that **employee records stay accurate as personal details or organizational assignments change**.

**Purpose:** Enable administrators to keep user data current — updating names, contact details, department/position assignments, skills, and documents — without needing to recreate the user account. This is an edit form for existing users, accessed from the User Details page, and covers personal and work-related fields only. Email, role, and account status are managed via separate dedicated actions.

**Target Users:**
- Any user with user management permission — full edit access

**Key Functionality:**
- Pre-filled edit form with existing user data
- Personal Information editing (name, DOB, phone, gender, address, avatar)
- Work Profile editing (department, position, experience year, CV, skills)
- Single Save button — stays on page after saving
- Unsaved changes protection via confirmation dialog

---

## 2. User Workflow

**Entry Point:** User Details page → "Update Information" in the left action panel.

**Preconditions:**
- User is signed in (US-001 Authentication)
- User has user management permission (US-004)
- User is viewing a specific user's details (DR-001-005-03)

**Main Flow:**
1. User is on the User Details page (Overview tab)
2. User clicks "Update Information" in the left action panel
3. System displays the Update Information form pre-filled with the user's existing data
4. User modifies any fields in Personal Information and/or Work Profile
5. User clicks Save (full-width button at bottom of form)
6. System validates all mandatory fields
7. System saves changes
8. System shows success toast "User information updated successfully"
9. User stays on the Update Information page with updated data

**Alternative Flows:**

- **Alt 1 — Validation fails:** Inline errors shown below invalid fields. Save button re-enables. Form is not submitted.
- **Alt 2 — Navigate away (form dirty):** User clicks another action (Overview, Change Email, back arrow, etc.) while form has unsaved changes → "Discard unsaved changes?" confirmation dialog appears → Confirm: navigates to target page, changes lost / Cancel: stays on form, changes preserved.
- **Alt 3 — Navigate away (form clean):** User clicks another action with no unsaved changes → navigates immediately, no dialog.
- **Alt 4 — Save with no changes:** User clicks Save without modifying anything → saves silently, no error message.
- **Alt 5 — Server error:** Save fails due to server error → error toast with retry suggestion, form data preserved.

**Exit Points:**
- **Success:** Toast shown, stays on page with updated data
- **Navigate away (clean):** Direct navigation to target
- **Navigate away (dirty + confirm discard):** Navigation to target, changes lost
- **Navigate away (dirty + cancel discard):** Stay on form, changes preserved

---

## 3. Field Definitions

### Input Fields — Personal Information Card

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Placeholder | Description |
|------------|------------|-----------------|-----------|---------------|-------------|-------------|
| Upload Avatar | Image upload | .png/.jpeg/.jpg/.webp only; max 2MB; max 500×500px | No | Existing avatar or placeholder | — | User profile photo |
| First name | Text input (278px, left) | Not empty | Yes (*) | Pre-filled | "Enter first name" | User's given name |
| Last name | Text input (278px, right) | Not empty | Yes (*) | Pre-filled | "Enter last name" | User's family name |
| Date of birth | Date picker (278px, left) | Not empty; valid date | Yes (*) | Pre-filled | "Enter date of birth" | User's birth date |
| Phone number | Text input (278px, right) | Not empty; Vietnam format: 10 digits starting with 0 | Yes (*) | Pre-filled | "Enter phone number" | Contact phone number |
| Gender | Dropdown (576px, full width) | Not empty | Yes (*) | Pre-filled | "Select gender" | User's gender |
| Address | Text input (576px, full width) | None | No | Pre-filled or empty | "Enter address" | User's residential address |

### Input Fields — Work Profile Card

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Placeholder | Description |
|------------|------------|-----------------|-----------|---------------|-------------|-------------|
| Department | Dropdown (278px, left) | Not empty; active departments only | Yes (*) | Pre-filled | "Select department" | Organizational unit assignment |
| Position | Dropdown (278px, right) | Not empty; active positions only | Yes (*) | Pre-filled | "Select position" | Job title assignment |
| Experience from (year) | Text/year input (576px) | Not empty; 4-digit year > 1900 and ≤ current year | Yes (*) | Pre-filled | "Enter year" | Career start year |
| CV/Resumé | File input (576px) | Accepted formats TBD (aligned with Create User) | No | Existing file name or "No file chosen" | — | Resume document |
| Skills | "+ Add skills" button | None | No | Existing skills displayed | — | Employee competencies |

### Interaction Elements

| Element Name | Type | Position | State/Condition | Trigger Action | Description |
|--------------|------|----------|-----------------|----------------|-------------|
| Save | Button (primary, full-width 600px) | Bottom of form | Always visible | Validate → save → success toast → stay on page | Submit changes |
| Upload Avatar button | Button (secondary) | Within avatar upload area | Always visible | Opens file picker for image selection | Replace or set avatar |
| "+ Add skills" button | Button (secondary, 110×32px) | Below Skills label | Always visible | Opens skill selector | Add skills to user |
| Date picker calendar icon | Icon button | Inside Date of birth field | Always visible | Opens calendar date picker | Select date of birth |

**Notes:**
- No Cancel button in the design — users navigate away via left action panel or back arrow
- Email, User Role, and Account Status fields are NOT present on this form

---

## 4. Data Display

### Information Shown to User

| Data Element | Data Type | Display When Empty | Format | Business Meaning |
|-------------|-----------|-------------------|--------|------------------|
| Avatar | Image | Default placeholder silhouette | Thumbnail in upload area | User's profile photo |
| First name | Text | Input with placeholder | Geist Regular 14px in input field | Given name |
| Last name | Text | Input with placeholder | Geist Regular 14px in input field | Family name |
| Date of birth | Date | Input with placeholder | Date format in date picker | Birth date |
| Phone number | Text | Input with placeholder | Geist Regular 14px in input field | Contact number |
| Gender | Text | Dropdown with placeholder | Selected option in dropdown | Gender |
| Address | Text | Input with placeholder | Geist Regular 14px in input field | Residential address |
| Department | Text | Dropdown with placeholder | Selected option in dropdown | Assigned department |
| Position | Text | Dropdown with placeholder | Selected option in dropdown | Assigned position |
| Experience from (year) | Text | Input with placeholder | 4-digit year in input field | Career start year |
| CV/Resumé | File name | "No file chosen" | File name text next to Choose File button | Resume document |
| Skills | Tags/badges | "+ Add skills" button only | Skill tags with remove option | Assigned competencies |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default | Page loads | All fields pre-filled with existing user data |
| Loading | Data being fetched | Loading indicator / skeleton fields |
| Validation error (empty mandatory) | User clears a mandatory field and clicks Save | Inline error below field: "[Field name] is required" |
| Validation error (phone format) | Invalid phone number format | Inline error: "Please enter a valid Vietnam phone number (10 digits starting with 0)" |
| Validation error (year range) | Year ≤ 1900 or > current year | Inline error: "Please enter a valid year (after 1900)" |
| Validation error (avatar size) | File exceeds 2MB | Inline error: "File size must not exceed 2MB" |
| Validation error (avatar format) | Unsupported file type | Inline error: "Accepted formats: .png, .jpeg, .jpg, .webp" |
| Saving | User clicks Save, request in progress | Save button shows loading state (disabled + spinner) |
| Success | Changes saved successfully | Success toast: "User information updated successfully" |
| Discard confirmation | User navigates away with unsaved changes | Modal: "Discard unsaved changes?" with Confirm + Cancel options |
| Server error | Save fails | Error toast with retry suggestion, form data preserved |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Form Display:**
- **AC-01:** Update Information form displays within the User Details page when "Update Information" is clicked in the left action panel
- **AC-02:** Form uses a single-column layout (600px) with Personal Information and Work Profile cards stacked vertically
- **AC-03:** All fields are pre-filled with the user's existing data on load
- **AC-04:** Mandatory fields marked with asterisk (*): First name, Last name, Date of birth, Phone number, Gender, Department, Position, Experience from (year)
- **AC-05:** Email, User Role, and Account Status fields are NOT present on this form

**Validation:**
- **AC-06:** User cannot save with any mandatory field empty
- **AC-07:** Phone number must be 10 digits starting with 0 (Vietnam format)
- **AC-08:** Experience from (year) must be a valid 4-digit year, greater than 1900 and not exceeding the current year
- **AC-09:** Avatar upload accepts only .png/.jpeg/.jpg/.webp, max 2MB, max 500×500px
- **AC-10:** Validation errors display inline below the relevant field
- **AC-11:** Validation is triggered on Save (not on blur)

**Save Behavior:**
- **AC-12:** Save button saves changes, shows toast "User information updated successfully", and stays on the Update Information page with updated data
- **AC-13:** Save button shows loading state (disabled + spinner) while the request is in progress
- **AC-14:** Save with no changes is allowed silently (no error)

**Navigation & Unsaved Changes:**
- **AC-15:** No Cancel button on the form — users navigate via left action panel or back arrow
- **AC-16:** If form has unsaved changes and user navigates away, a "Discard unsaved changes?" confirmation dialog is shown
- **AC-17:** If form has no unsaved changes, navigation proceeds immediately without dialog
- **AC-18:** Smart dirty check — reverting a field to its original value makes the form "clean" again (no false positives)

**Avatar & Files:**
- **AC-19:** Existing avatar is displayed in the upload area on load; user can replace it
- **AC-20:** Existing CV/Resumé file name is shown; user can replace or remove it
- **AC-21:** Existing skills are displayed; user can add or remove skills via "+ Add skills"

**Access Control:**
- **AC-22:** Update Information action is visible only to users with user management permission
- **AC-23:** Direct URL access by unauthorized users redirects to an appropriate fallback

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path | Modify first name, click Save | Toast, stays on page, data updated | High |
| Multiple field changes | Modify name + department + phone | All changes saved, toast shown | High |
| Empty mandatory field | Clear first name, click Save | Inline error "First name is required" | High |
| Invalid phone | Enter "12345", click Save | Inline error about Vietnam format | High |
| Valid phone | Enter "0912345678" | Accepted, no error | High |
| Invalid year (too old) | Enter "1800" | Inline error "Please enter a valid year (after 1900)" | Medium |
| Invalid year (future) | Enter year > current year | Inline error "Year cannot exceed current year" | Medium |
| Avatar too large | Upload 5MB image | Inline error "File size must not exceed 2MB" | Medium |
| Avatar wrong format | Upload .gif file | Inline error about accepted formats | Medium |
| No changes + Save | Click Save without editing | Saves silently, no error | Medium |
| Navigate away (dirty) | Modify field, click Overview | "Discard unsaved changes?" dialog | High |
| Navigate away (clean) | No changes, click Overview | Navigates immediately | Medium |
| Confirm discard | Dirty form → dialog → Confirm | Navigates to Overview, changes lost | Medium |
| Cancel discard | Dirty form → dialog → Cancel | Stays on form, changes preserved | Medium |
| Smart dirty check | Change name, then revert to original | Form is clean, no dialog on navigate | Medium |
| Unauthorized access | User without permission visits URL | Redirect / access denied | High |

---

## 6. System Rules

**Business Logic:**
- **SR-01:** All fields are pre-filled with existing user data when the form loads — this is an edit form, not a create form
- **SR-02:** Email, User Role, and Account Status are NOT editable on this form — they are managed via separate dedicated actions (Change Email, Change User Role, Activate/Deactivate)
- **SR-03:** Phone number must follow Vietnam format: 10 digits, starting with 0 (e.g., 0912345678)
- **SR-04:** Experience from (year) must be > 1900 and ≤ current year
- **SR-05:** Avatar upload: .png/.jpeg/.jpg/.webp only, max 2MB, max 500×500px
- **SR-06:** Only users with user management permission can access the Update Information form
- **SR-07:** Saving with no changes is allowed — system processes it silently
- **SR-08:** Audit logging will be handled by a separate logging story — this DR does not define logging behavior
- **SR-09:** Department and Position dropdowns are populated from active departments/positions only (EP-008 data)
- **SR-10:** Skills are populated from the active skill catalog (EP-008 US-003)
- **SR-11:** Concurrent editing uses last-save-wins strategy — no conflict warning or optimistic locking
- **SR-12:** Smart dirty check compares current field values to originally loaded values — reverting to original values makes the form "clean"

**State Transitions:**
```
[User Details - Overview] → "Update Information" click → [Update Information form (pre-filled)]
[Update Information] → Save (valid) → [Update Information (updated data + success toast)]
[Update Information] → Save (invalid) → [Update Information (inline errors shown)]
[Update Information] → Navigate away (clean) → [Target page]
[Update Information] → Navigate away (dirty) → [Confirmation dialog]
[Confirmation dialog] → Confirm discard → [Target page]
[Confirmation dialog] → Cancel → [Update Information (changes preserved)]
```

**Dependencies:**
- **Depends on:** US-001 (Authentication) — user must be signed in
- **Depends on:** US-004 (Role & Permission Management) — access control enforcement
- **Depends on:** DR-001-005-03 (User Details) — parent page providing the action panel and entry point
- **Depends on:** EP-008 US-001 (Department Management) — department dropdown data
- **Depends on:** EP-008 US-002 (Position Management) — position dropdown data
- **Depends on:** EP-008 US-003 (Skill Management) — skill catalog data

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** All fields pre-filled on load — user immediately sees current data and can identify what needs changing
- **UX-02:** Save button shows loading spinner while request is in progress, preventing double submission
- **UX-03:** Success toast auto-dismisses after 5 seconds (with manual close option)
- **UX-04:** "Discard unsaved changes?" dialog only appears when form is actually dirty — prevents unnecessary interruptions
- **UX-05:** Smart dirty check: form tracks dirty state by comparing current values to original loaded values — reverting a field to its original value makes the form "clean" again
- **UX-06:** Keyboard navigation supported — Tab through fields, Enter to submit
- **UX-07:** Date of birth uses a date picker with calendar icon for easy selection
- **UX-08:** Department and Position dropdowns support type-ahead search for quick selection

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Action panel (189px) + form (600px), side-by-side field pairs |
| Tablet (768-1024px) | Action panel collapses, form full-width, side-by-side pairs preserved |
| Mobile (<768px) | Action panel becomes top menu, fields stack vertically (single column) |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through all fields and Save button
- [x] Screen reader compatible — field labels, error messages, toast announcements
- [x] Sufficient color contrast — meets WCAG 2.1 AA standards
- [x] Focus indicators visible — clear focus ring on interactive elements

**Design References:**
- Figma: [Update Information](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3122-6323) (node `3122:6323`)
- Design tokens: See ANALYSIS.md Section 7 — Design Context [ADD-ON]
- Pattern reference: Create User (DR-001-005-02) — shares field definitions and validation rules

---

## 8. Additional Information

### Out of Scope
- Change Email — separate action in User Details (planned)
- Change User Role — separate action in User Details (planned)
- Reset Password — separate action in User Details (planned)
- Activate/Deactivate — separate action in User Details (planned)
- Delete User — separate action in User Details (planned)
- User self-service profile editing (admin-only in this scope)
- Audit logging implementation (separate logging story)
- Concurrent editing conflict detection (last-save-wins strategy chosen)

### Open Questions
- None remaining. All questions resolved during requirement writing session.

### Related Features
- **DR-001-005-03:** User Details — parent page providing the action panel and entry point
- **DR-001-005-02:** Create User — shares most fields (Update Information is a subset without Email, Role, Account Status)
- **DR-001-005-01:** User List — navigation origin (User List → User Details → Update Information)
- **EP-008 US-001:** Department Management — department dropdown data source
- **EP-008 US-002:** Position Management — position dropdown data source
- **EP-008 US-003:** Skill Management — skill catalog data source
- **US-001:** Authentication — user must be signed in
- **US-004:** Role & Permission Management — access control enforcement

### Notes
- The Figma design does not include a Cancel button — this is intentional. Users navigate away via the left action panel or back arrow, with a confirmation dialog protecting unsaved changes.
- The Update Information form is a **subset** of Create User: it excludes Email (Change Email action), User Role (Change User Role action), and Account Status (Activate/Deactivate action). These are deliberately separated into individual actions for audit clarity and permission granularity.
- The form layout is **single-column** (600px) unlike Create User's two-column layout, because the User Account card (Role + Status) is absent.
- The Last name placeholder in Figma says "Enter first name" — this is a copy-paste error. The correct placeholder is "Enter last name".

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | BA Agent | 2026-03-26 | Draft |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |
| Tech Lead | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-03-26 | BA Agent | Initial draft — full 8-section detail requirement with Figma design context |
