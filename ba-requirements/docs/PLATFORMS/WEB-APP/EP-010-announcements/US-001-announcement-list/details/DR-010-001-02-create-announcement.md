---
document_type: DETAIL_REQUIREMENT
# Platform Information
platform: web-app
platform_display: "WEB-APP"
# Epic & Story Identification
epic_id: EP-010
story_id: US-001
story_name: "Announcement List"
# Detail Requirement Identification
detail_id: DR-010-001-02
detail_name: "Create Announcement"
# Status & Version
status: draft
version: "1.2"
created_date: 2026-04-25
last_updated: 2026-04-25
# Document linking
related_documents:
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "./DR-010-001-01-announcement-list.md"
    relationship: sibling
  - path: "../../EPIC.md"
    relationship: parent
# Input sources
input_sources:
  - type: figma
    file_id: "exn-hr-design"
    node_id: "3345:4444"
    description: "Create A New Announcement screen design"
    extraction_date: 2026-04-25
---

# Detail Requirement: Create Announcement

**Detail ID:** DR-010-001-02
**Story:** US-001-announcement-list
**Epic:** EP-010 (Announcements)
**Status:** Draft
**Version:** 1.2

---

## 1. Use Case Description

As an **HR Manager or Admin**, I want to **create a new company announcement** so that **I can communicate important information to all employees or specific individuals**.

**Purpose:** Enable authorized personnel to compose and distribute company-wide or targeted announcements. The form supports two distinct save actions: saving as a draft for later review/editing, or immediately sending to the selected recipients.

**Target Users:**
- HR Managers with announcement management permission
- Admins with full system access

**Key Functionality:**
- Compose announcement with title and description
- Choose recipients: everyone or specific employees
- Save as draft for later editing and review
- Save and send immediately to distribute to recipients
- **Notification delivery** via two channels when sent:
  - **Email notification** to recipient's registered email address
  - **Push notification** to recipient's mobile device (via MOBILE-APP)

---

## 2. User Workflow

**Entry Point:** 
- Announcement List page > "+ Add Announcement" button
- Sidebar > Organization Settings > Announcements > "+ Add Announcement"

**Preconditions:**
- User is authenticated and logged in
- User has announcement management permission (via US-004 Permission Management)

**Main Flow:**
1. User clicks "+ Add Announcement" button from the Announcement List page
2. System navigates to the Create Announcement form page
3. System displays the page with title "Create A New Announcement" and empty form
4. User enters Title (mandatory) in the first card "Announcement Details"
5. User enters Description (mandatory) in the textarea field
6. User selects recipients in the second card "Receiver Info":
   - "Everyone" toggle is ON by default (all employees will receive)
   - OR user toggles "Specific one" and selects employee(s) from dropdown
7. User chooses one of three actions:
   - **Cancel**: Navigates back to list (with discard confirmation if form is dirty)
   - **Save As Draft**: Saves announcement with Draft status, redirects to list
   - **Save & Send**: Saves announcement with Sent status, immediately distributes to recipients via **email and push notification**, redirects to list

**Alternative Flows:**
- **Alt 1 - Validation Error:** User attempts to save without completing mandatory fields; inline error messages appear below each invalid field; save action is blocked
- **Alt 2 - Select Specific Recipients:** User toggles "Specific one" switch; Employee dropdown becomes enabled; user must select at least one employee before saving
- **Alt 3 - Cancel with Dirty Form:** User has entered data and clicks Cancel or back; system shows "Discard unsaved changes?" confirmation dialog

**Exit Points:**
- **Success (Draft):** Toast "Announcement has been saved as draft" + redirect to Announcement List
- **Success (Sent):** Toast "Announcement has been sent" + redirect to Announcement List; system triggers email and push notifications to recipients in background
- **Cancel (Clean):** Immediate redirect to Announcement List
- **Cancel (Dirty):** Confirmation dialog; if confirmed, redirect to list; if cancelled, stay on form
- **Error:** Inline validation errors displayed; user remains on form to correct

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Title | Text Input | Min 1 char, Max 100 chars, whitespace-only invalid | Yes | Empty | Announcement subject/headline |
| Description | Textarea | Min 1 char, Max 2000 chars, whitespace-only invalid | Yes | Empty | Full announcement content |
| Label | Single-select Dropdown (create-on-type) | Max 50 chars for new label name | No | Empty | Optional categorization tag for the announcement |
| Everyone | Toggle Switch | Mutually exclusive with "Specific one" | Yes (one must be selected) | ON | Send to all employees |
| Specific one | Toggle Switch | Mutually exclusive with "Everyone" | Yes (one must be selected) | OFF | Send to selected employees only |
| Employee | Dropdown (multi-select) | At least 1 employee required when "Specific one" is ON | Conditional | Empty | Select specific recipient(s) |

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Cancel (Header) | Secondary Button | Always enabled | Navigate back with discard check | Top-right header area, icon + "Cancel" text |
| Save As Draft (Header) | Secondary Button | Enabled when form is valid | Save as Draft status | Top-right header area, icon + "Save As Draft" text |
| Save & Send (Header) | Primary Button | Enabled when form is valid | Save as Sent status and distribute | Top-right header area, dark background, icon + "Save & Send" text |
| Save As Draft (Bottom) | Secondary Button | Enabled when form is valid | Same as header Save As Draft | Full-width 295px, white background with border |
| Save & Send (Bottom) | Primary Button | Enabled when form is valid | Same as header Save & Send | Full-width 295px, dark background (#010101) |
| Everyone Toggle | Toggle Switch | ON by default | Disable "Specific one" and Employee dropdown | Mutually exclusive selection |
| Specific one Toggle | Toggle Switch | OFF by default | Enable Employee dropdown, disable "Everyone" | Mutually exclusive selection |
| Employee Dropdown | Searchable Dropdown | Disabled when "Everyone" is ON; visible at 50% opacity | Open employee selection list | Multi-select with client-side search |
| Label Dropdown | Single-select Dropdown | Always enabled | Open label selection or create new | Select existing label OR type to create new |

### Receiver Selection Behavior

| State | Everyone Toggle | Specific one Toggle | Employee Dropdown |
|-------|-----------------|---------------------|-------------------|
| Default | ON | OFF | Disabled (50% opacity) |
| Specific Recipients | OFF | ON | Enabled (100% opacity, mandatory) |

### Label Dropdown Behavior (Create-on-Type)

| User Action | System Behavior | Result |
|-------------|-----------------|--------|
| Click dropdown | Opens dropdown showing all existing labels | User sees available labels |
| Type text matching existing label | Filters dropdown to matching labels | User can select from filtered list |
| Type text NOT matching any label | Shows filtered results (if partial match) + help text "Press Enter to add new" | User informed they can create |
| Press Enter with non-matching text | Creates new label with typed name, selects it | New label saved to system and selected |
| Select existing label | Label selected and dropdown closes | Label value set |
| Clear selection (X icon) | Removes selected label | Field returns to empty state |

---

## 4. Data Display

### Form Cards Layout

| Card | Title | Content |
|------|-------|---------|
| Card 1 | Announcement Details | Title field, Description textarea, Label dropdown |
| Card 2 | Receiver Info | Everyone toggle, Specific one toggle, Employee dropdown |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page initializing | Skeleton placeholders for form fields |
| Ready | Form loaded | Empty form with default values (Everyone toggle ON) |
| Dirty | User has entered any data | Form tracks changes for discard confirmation |
| Saving | Save button clicked | Button shows loading spinner, form disabled |
| Validation Error | Mandatory field empty or invalid | Inline error message below field in red text |
| Success | Save completed | Toast notification, redirect to list |

### Validation Error Messages

| Field | Error Condition | Error Message |
|-------|----------------|---------------|
| Title | Empty or whitespace-only | "Title is required" |
| Title | Exceeds 100 characters | "Title must not exceed 100 characters" |
| Description | Empty or whitespace-only | "Description is required" |
| Description | Exceeds 2000 characters | "Description must not exceed 2000 characters" |
| Employee | "Specific one" ON but no employee selected | "Please select at least one employee" |

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** Users with announcement management permission can access the Create Announcement form; users without permission cannot access
- **AC-02:** Title and Description fields are mandatory; form cannot be saved without completing both
- **AC-03:** "Everyone" toggle is ON by default; toggling to "Specific one" enables the Employee dropdown
- **AC-04:** "Everyone" and "Specific one" toggles are mutually exclusive (radio-button behavior)
- **AC-05:** When "Specific one" is selected, at least one employee must be chosen before saving
- **AC-06:** "Save As Draft" creates an announcement with Draft status and redirects to list with success toast
- **AC-07:** "Save & Send" creates an announcement with Sent status, distributes to recipients, and redirects to list with success toast
- **AC-13:** When "Save & Send" is clicked, system sends **email notification** to all recipients' registered email addresses
- **AC-14:** When "Save & Send" is clicked, system sends **push notification** to all recipients' mobile devices (MOBILE-APP)
- **AC-15:** Notifications are sent asynchronously in background; user does not wait for notification delivery
- **AC-08:** Cancel button with dirty form shows "Discard unsaved changes?" confirmation dialog
- **AC-09:** Cancel button with clean form navigates directly to Announcement List without confirmation
- **AC-10:** Validation errors display inline below the respective field
- **AC-11:** Both header and bottom button pairs trigger the same actions (dual save buttons)
- **AC-12:** Employee dropdown is disabled and shown at 50% opacity when "Everyone" is selected
- **AC-16:** Label field is optional; announcement can be saved with or without a label
- **AC-17:** Label dropdown shows all existing labels when opened; typing filters the list
- **AC-18:** When user types a label name that does not exist, system shows "Press Enter to add new" help text
- **AC-19:** Pressing Enter with a non-existing label name creates the new label and selects it
- **AC-20:** Newly created labels are immediately available for selection in future announcements

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Create and send to everyone | Fill title, description, leave Everyone ON, click Save & Send | Announcement sent to all employees, toast shown, redirect to list | High |
| Create and save as draft | Fill title, description, click Save As Draft | Announcement saved as Draft, toast shown, redirect to list | High |
| Create and send to specific employee | Fill fields, toggle Specific one, select employee, click Save & Send | Announcement sent to selected employee(s) only | High |
| Validation - missing title | Leave title empty, click Save | Inline error "Title is required", save blocked | High |
| Validation - missing description | Leave description empty, click Save | Inline error "Description is required", save blocked | High |
| Validation - no employee selected | Toggle Specific one, leave dropdown empty, click Save | Inline error "Please select at least one employee" | High |
| Cancel with dirty form | Enter title, click Cancel | Discard confirmation dialog appears | Medium |
| Cancel with clean form | Click Cancel immediately | Navigate directly to list | Medium |
| Toggle exclusivity | Toggle Specific one ON | Everyone toggles OFF, Employee dropdown enables | Medium |
| Character limit - title | Enter 101+ characters in title | Error "Title must not exceed 100 characters" | Medium |
| Notification - email sent | Save & Send to employee with email | Email received at employee's registered address | High |
| Notification - push sent | Save & Send to employee with MOBILE-APP | Push notification received on mobile device | High |
| Notification - no email | Save & Send to employee without registered email | No error; email skipped for that recipient | Medium |
| Label - select existing | Open Label dropdown, select existing label, save | Announcement saved with selected label | Medium |
| Label - create new | Type "New Category", press Enter, save | New label created and assigned to announcement | Medium |
| Label - optional | Fill required fields, leave Label empty, save | Announcement saved without label | Medium |
| Label - filter dropdown | Type partial text matching existing labels | Dropdown filters to show matching labels only | Low |
| Label - clear selection | Select a label, click X to clear, save | Announcement saved without label | Low |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **Rule 1:** Only users with announcement management permission (configured via US-004) can create announcements
- **Rule 2:** "Save As Draft" creates an announcement with status = "Draft"
- **Rule 3:** "Save & Send" creates an announcement with status = "Sent" and triggers distribution to recipients
- **Rule 4:** Once sent (status = Sent), announcements are immutable and cannot be edited
- **Rule 5:** Receiver selection is stored as:
  - "everyone = true" when Everyone toggle is ON
  - "everyone = false" + list of employee IDs when Specific one is selected
- **Rule 6:** Leading/trailing whitespace is trimmed from Title and Description before saving
- **Rule 7:** Whitespace-only input is treated as empty (validation fails)
- **Rule 8:** Created Date is auto-generated on save (UTC timestamp)
- **Rule 9:** Sent Date is populated only when "Save & Send" is used; null for drafts
- **Rule 10:** When "Save & Send" is executed, system triggers two notification channels:
  - **Email notification:** Sent to each recipient's registered email address with announcement title and content
  - **Push notification:** Sent to each recipient's mobile device via MOBILE-APP push service
- **Rule 11:** Notifications are processed asynchronously (background job); the UI does not block waiting for delivery
- **Rule 12:** If a recipient has no registered email, email notification is skipped for that recipient (no error)
- **Rule 13:** If a recipient has not enabled push notifications on MOBILE-APP, push is skipped for that recipient (no error)
- **Rule 14:** Email subject format: "[Company Name] Announcement: {Title}"
- **Rule 15:** Push notification title: "New Announcement"; body: "{Title}" (truncated to 100 chars if needed)
- **Rule 16:** Label is optional; announcements can be created without a label
- **Rule 17:** New labels are created on-the-fly when user types a non-existing name and presses Enter
- **Rule 18:** Label names are unique (case-insensitive); attempting to create a duplicate selects the existing label
- **Rule 19:** Label names are trimmed of leading/trailing whitespace before saving
- **Rule 20:** Label names have a maximum length of 50 characters
- **Rule 21:** Labels are shared across all announcements; once created, a label is available for all future announcements

**State Transitions:**

```
[New Form] → [Save As Draft] → [Draft Status]
[New Form] → [Save & Send] → [Sent Status]
```

**Permission Model:**

| Action | Required Permission |
|--------|---------------------|
| Access Create Form | Announcement Management |
| Save As Draft | Announcement Management |
| Save & Send | Announcement Management |

**Dependencies:**
- US-004 Permission Management - provides access control
- User/Employee data - for recipient selection dropdown
- Authentication system - user must be logged in
- **Email service** - for sending email notifications (SMTP/email provider)
- **Push notification service** - for sending push notifications to MOBILE-APP (Firebase/APNs)
- **Label data** - for label dropdown options and create-on-type functionality

---

## 7. UX Optimizations

**Usability Considerations:**

- **Optimization 1:** Dual save buttons (header + bottom) ensure quick access regardless of scroll position
- **Optimization 2:** Default "Everyone" toggle reduces clicks for the most common use case (company-wide announcements)
- **Optimization 3:** Visual feedback (50% opacity) clearly indicates Employee dropdown is disabled when Everyone is selected
- **Optimization 4:** Resizable Description textarea allows users to expand for longer content
- **Optimization 5:** Discard confirmation prevents accidental data loss
- **Optimization 6:** Auto-focus on Title field on page load for immediate typing

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Full layout with 600px centered form card |
| Below Desktop | Out of scope - web admin is desktop-only |

**Accessibility Requirements:**
- [x] Keyboard navigable - Tab through all fields and buttons in logical order
- [x] Screen reader compatible - Labels associated with inputs, toggle state announced
- [x] Sufficient color contrast - All text meets WCAG AA standards
- [x] Focus indicators visible - Clear focus ring on all interactive elements
- [x] Toggle state announced - Screen readers announce "Everyone: on/off" state

**Design References:**
- Figma Node: 3345:4444 (Create A New Announcement)
- Follows Create Form Pattern from Knowledge Base (Section 2)
- Button style consistent with other create forms (dual header + bottom buttons per DR-002-001-02, DR-003-001-02)

---

## 8. Additional Information

### Out of Scope
- Scheduled publishing (send at future date/time)
- File attachments (images, documents)
- Rich text editor (bold, italic, links)
- Mobile responsive layout
- Multi-language announcements
- Announcement categories or tags
- Preview before sending
- Notification delivery status tracking (read receipts)
- Retry failed notifications
- Notification preferences per user (opt-out)

### Open Questions
- None - all questions resolved

### Related Features
- DR-010-001-01: Announcement List (parent list view)
- DR-010-001-03: Edit Announcement (to be created - Draft only)
- DR-010-001-04: Send Announcement Confirmation (to be created)
- DR-010-001-05: Delete Announcement Confirmation (to be created)
- US-004: Permission Management (provides access control)

### Notes
- The design shows mutually exclusive toggle switches for receiver selection, functioning like radio buttons
- "Save & Send" is the primary action (dark background), while "Save As Draft" is secondary (light background)
- Employee dropdown uses searchable multi-select pattern consistent with other employee selection fields
- The form uses the established dual-button pattern: identical actions in header and bottom of form

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | | | Pending |
| Product Owner | | | Pending |
| UX Designer | | | Pending |
| Tech Lead | | | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-04-25 | Claude | Initial draft from Figma design extraction |
| 1.1 | 2026-04-25 | Claude | Added notification delivery details (email + push) for Save & Send action |
| 1.2 | 2026-04-25 | Claude | Added Label field with create-on-type capability (AC-16 to AC-20, Rules 16-21) |
