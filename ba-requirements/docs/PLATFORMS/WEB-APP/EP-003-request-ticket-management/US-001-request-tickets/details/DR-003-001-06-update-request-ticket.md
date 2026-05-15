---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-003
story_id: US-001
story_name: "Request Tickets"
detail_id: DR-003-001-06
detail_name: "Update Request Ticket"
parent_requirement: FR-US-001-10
status: draft
version: "1.0"
created_date: 2026-04-03
last_updated: 2026-04-03
supersedes: DR-003-001-03
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "./DR-003-001-01-request-tickets-list.md"
    relationship: sibling
  - path: "./DR-003-001-02-create-request-ticket.md"
    relationship: sibling
  - path: "./DR-003-001-03-update-request-ticket.md"
    relationship: superseded-by-this
  - path: "./DR-003-001-05-request-ticket-details.md"
    relationship: sibling
input_sources:
  - type: figma
    description: "Update Request Ticket — initial state (employee not yet selected)"
    node_id: "3213:4286"
    extraction_date: "2026-04-03"
  - type: figma
    description: "Update Request Ticket — populated state (employee selected, info card visible)"
    node_id: "3213:4524"
    extraction_date: "2026-04-03"
---

# Detail Requirement: Update Request Ticket

**Detail ID:** DR-003-001-06
**Supersedes:** DR-003-001-03 (written without updated Figma context)
**Parent Requirement:** FR-US-001-10
**Story:** US-001-request-tickets
**Epic:** EP-003 (Request Ticket Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **employee with request ticket access**, I want to **edit a request ticket that is still in Open status**, so that **I can correct or update the request details before it is processed**.

As a **manager or administrator with request ticket management permission**, I want to **edit a request ticket in Open or In Progress status**, so that **I can refine the ticket details as the request is being handled**.

**Purpose:** Allow authorized users to update an existing request ticket's content. The edit form is presented within a two-panel layout — the same contextual layout as the ticket detail page — giving the user visibility of all available ticket actions (Request Details, Update Request, Delete Request) without losing their place in the workflow. All editable fields are pre-filled with the current ticket values. The submitting employee is locked and cannot be changed. Editing does not affect the ticket's current status.

**Target Users:**
- **Employees** — can edit only their own tickets in **Open** status
- **Managers/Administrators** — can edit any ticket in **Open** or **In Progress** status

**Key Functionality:**
- Two-panel page: action buttons (left, 162px) + edit form (right, 600px)
- Edit form pre-filled with current ticket data
- Employee info card visible as read-only context (locked — cannot change submitter)
- 3 mandatory fields + 1 optional (Attachment), mirroring the Create form
- Editing does not change ticket status
- Concurrent edit protection — server rejects save if ticket status changed during editing
- Single Save button: full-width at bottom of form

---

## 2. User Workflow

**Entry Point:** Request Tickets List → Gear icon on a row → "Edit" option (visible only for editable-status tickets per role)

**Preconditions:**
- User is signed in (US-001)
- User has request ticket access permission (US-004)
- Ticket is in an editable status:
  - **Employees:** ticket status is **Open** AND belongs to the logged-in user
  - **Managers/Admins:** ticket status is **Open** or **In Progress**

**Gear Icon Edit Visibility Matrix:**

| Ticket Status | Employee (own ticket) | Manager/Admin |
|---|---|---|
| Open | Edit visible | Edit visible |
| In Progress | Edit not visible | Edit visible |
| On Hold | Edit not visible | Edit not visible |
| Resolved | Edit not visible | Edit not visible |
| Closed | Edit not visible | Edit not visible |
| Cancelled | Edit not visible | Edit not visible |

**Left Panel Action Button Visibility:**

| Button | Employee (own ticket) | Manager/Admin |
|--------|----------------------|---------------|
| Request Details | Always visible | Always visible |
| Update Request | Visible when ticket is editable per role (Open for employee; Open or In Progress for manager/admin) | Same |
| Delete Request | Visible per delete rules (DR-003-001-04) | Always visible |

**Main Flow A — Employee editing own Open ticket:**
1. Employee clicks gear icon on their Open ticket in the list
2. Employee selects "Edit" from the dropdown menu
3. System navigates to the Update Request Ticket page
4. Left panel displays: Request Details, Update Request (active/highlighted), Delete Request buttons
5. Right panel displays the edit form: Employee card (read-only info card) + Request Info card (pre-filled, editable)
6. Employee section shows the logged-in employee's own profile as a locked, read-only info card
7. Employee modifies the desired fields in the Request Info section
8. Employee clicks Save
9. System validates all mandatory fields
10. System updates the ticket; status remains **Open**
11. Success toast: "Request ticket has been updated"
12. System redirects to Request Tickets List

**Main Flow B — Manager/Admin editing Open or In Progress ticket:**
1. Manager clicks gear icon on an Open or In Progress ticket
2. Manager selects "Edit" from the dropdown menu
3. System navigates to the Update Request Ticket page
4. Left panel displays: Request Details, Update Request (active/highlighted), Delete Request buttons
5. Right panel displays the edit form: Employee card (read-only locked submitter info) + Request Info card (pre-filled, editable)
6. Manager modifies the desired fields in the Request Info section
7. Manager clicks Save
8. System validates all mandatory fields
9. System checks the ticket is still in an editable status (concurrent edit protection)
10. System updates the ticket; status remains unchanged (Open or In Progress)
11. Success toast: "Request ticket has been updated"
12. System redirects to Request Tickets List

**Alternative Flows:**

| Action | Flow |
|--------|------|
| **Alt 1 — Empty mandatory field** | Inline error "[Field name] is required". Form not submitted. |
| **Alt 2 — Subject whitespace-only** | Inline error: "Subject is required". Form not submitted. |
| **Alt 3 — Description whitespace-only** | Inline error: "Description is required". Form not submitted. |
| **Alt 4 — Subject exceeds max length** | Inline error: "Subject must not exceed 200 characters". Form not submitted. |
| **Alt 5 — Invalid attachment file type** | Inline error: "Only PDF, PNG, JPG, and DOCX files are accepted". File rejected. |
| **Alt 6 — Attachment too large** | Inline error: "File size must not exceed 5MB". File rejected. |
| **Alt 7 — Back arrow clicked (form modified)** | "Discard unsaved changes?" dialog. Confirm → redirect to list. Cancel → stay on form. |
| **Alt 8 — Back arrow clicked (form untouched)** | Redirect to Request Tickets List immediately. |
| **Alt 9 — Status changed during editing** | On Save: error toast "This ticket can no longer be edited" + redirect to list. |
| **Alt 10 — Server error on save** | Error toast: "Something went wrong. Please try again." Form stays open, fields preserved. |

**Exit Points:**
- **Success:** Ticket updated → toast → redirect to Request Tickets List
- **Back/Cancel:** Redirect to list (with or without discard confirmation)
- **Error:** Validation errors inline (user corrects and retries); server conflict (redirect to list)

---

## 3. Field Definitions

### Employee Section (always visible, read-only)

| Field Name | Field Type | Validation Rule | Mandatory | Default | Description |
|------------|------------|-----------------|-----------|---------|-------------|
| Employee | Read-only display (576px) | Not editable — locked to original submitter | N/A | Pre-filled with ticket submitter | Cannot be changed; shown as a locked, pre-filled dropdown-style field |

### Employee Info Card (read-only, always shown after employee pre-fill)

| Field | Left Column | Right Column |
|-------|-------------|--------------|
| Row 1 | Full name (icon + label + value) | Department (icon + label + value) |
| Row 2 | Email (icon + label + value) | Position (icon + label + value) |
| Row 3 | Phone number (icon + label + value) | — |

**Note:** Unlike the Create form (which shows the hint "Select an employee to view their info" in the empty state), the Edit form always pre-fills and shows the info card on load. The hint banner is never shown in the edit context — the submitter is always known.

### Input Fields — Request Info Section (pre-filled, editable)

| Field Name | Field Type | Validation Rule | Mandatory | Default | Description |
|------------|------------|-----------------|-----------|---------|-------------|
| Request Type | Dropdown (576px) | Must select a value | Yes (*) | Pre-filled from current ticket | IT Support, Facility, HR Inquiry, Office Supplies, Access Request, Travel & Expense, Training, Other |
| Subject | Text input (576px) | Not empty; trimmed; max 200 characters | Yes (*) | Pre-filled from current ticket | Short title for the request (displayed as "Request" column in list) |
| Description | Text area (576px) | Not empty; trimmed | Yes (*) | Pre-filled from current ticket | Detailed explanation of the request (labeled "Request details" in Figma — see design gap) |
| Attachment | File input (576px) | PDF, PNG, JPG, DOCX; max 5MB | No | Shows current filename if attachment exists | Supporting document — can be kept, replaced, or removed |

### Interaction Elements

| Element | Type | Position | State/Condition | Trigger Action | Description |
|---------|------|----------|-----------------|----------------|-------------|
| Back arrow | Icon button | Left of page title (top of content area) | Always visible | If dirty → discard dialog; if clean → redirect to list | Navigate back to list |
| Request Details | Button (left panel) | Left action panel, position 1 | Always visible | Navigate to ticket detail view | View read-only ticket details |
| Update Request | Button (left panel, active) | Left action panel, position 2 | Visible when ticket is editable per role | Current page (active/highlighted state) | Currently active — the edit form |
| Delete Request | Button (left panel, danger) | Left action panel, position 3 | Visible per delete permission rules | Opens delete confirmation dialog | Triggers soft delete of the ticket |
| Save | Button (primary, full-width 600px, bg #010101) | Bottom of form | Disabled + spinner while saving | Validates → updates ticket → toast → redirect | Submit changes |

### Validation Error Messages

| Condition | Error Message | Display | Blocks Submit? |
|-----------|--------------|---------|---------------|
| Mandatory field empty | "[Field name] is required" | Inline, below field | Yes |
| Subject exceeds max length | "Subject must not exceed 200 characters" | Inline, below Subject | Yes |
| Subject whitespace-only | "Subject is required" | Inline, below Subject | Yes |
| Description whitespace-only | "Description is required" | Inline, below Description | Yes |
| Invalid attachment type | "Only PDF, PNG, JPG, and DOCX files are accepted" | Inline, below Attachment | Yes (file rejected) |
| Attachment too large | "File size must not exceed 5MB" | Inline, below Attachment | Yes (file rejected) |
| Status changed during edit | "This ticket can no longer be edited" | Error toast | Yes (redirect to list) |

---

## 4. Data Display

### Information Shown to User

**Employee Section (always visible, read-only):**

| Data Name | Format | State | Business Meaning |
|-----------|--------|-------|-----------------|
| Employee field | Locked dropdown-style input (576px) — shows submitter name, grayed/disabled style | Always pre-filled | Cannot change submitter |
| Employee info card | 5 fields in 2-column layout (Left: Full name, Email, Phone; Right: Department, Position) — each field uses icon + label + value pattern | Always visible | Read-only context: who submitted this ticket |

**Request Info Section (pre-filled, editable):**

| Data Name | Format | Business Meaning |
|-----------|--------|-----------------|
| Request Type | Dropdown (576px) — pre-filled with current value | Category of the request |
| Subject | Text input (576px) — pre-filled with current value | Short title displayed in list "Request" column |
| Description | Text area (576px) — pre-filled with current value | Detailed explanation of the request |
| Attachment | File input (576px) — shows current filename if exists; "Choose File / No file chosen" if none | Supporting document |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default (page load) | All fields pre-filled from current ticket | Employee info card visible; Request Info fields editable with current values |
| Attachment exists | Ticket has an existing attachment | Current filename shown next to "Choose File" |
| Attachment removed | User clears the attachment input | "No file chosen" displayed |
| New file selected | User uploads a replacement file | New filename displayed; replaces current |
| Validation error | Save with invalid/empty fields | Inline red error(s) below affected fields |
| Saving | Save clicked, in progress | Save button shows spinner + disabled; all fields disabled |
| Success | Ticket updated | Toast "Request ticket has been updated" → redirect to list |
| Discard confirmation | Back arrow clicked with modified form | Modal: "Discard unsaved changes?" |
| Server conflict | Status changed by another user while editing | Error toast: "This ticket can no longer be edited" → redirect to list |

### Page Layout

**Figma References:** node `3213:4286` (initial state) and `3213:4524` (employee populated)

```
+---------------------------------------------------------------+
|  [Logo]  [Nav: SidebarSimple]  Breadcrumb / ... / Breadcrumb  |
+----------------+----------------------------------------------+
|  [Sidebar]     |  <- Update Request Ticket  > #22052000 - Training |
|  200px         |                                              |
|                |  [Request Details  ]   +------------------+ |
|                |  [Update Request   ]   | Employee         | |
|                |  [Delete Request   ]   | * Employee       | |
|                |  162px (left panel)    |   [locked input] | |
|                |                        |                  | |
|                |                        | Full name  Dept  | |
|                |                        | Email      Pos   | |
|                |                        | Phone            | |
|                |                        +------------------+ |
|                |                        | Request Info     | |
|                |                        | *Request type    | |
|                |                        |  [dropdown]      | |
|                |                        | *Request details | |
|                |                        |  [textarea]      | |
|                |                        |  Attachment      | |
|                |                        |  [Choose File]   | |
|                |                        +------------------+ |
|                |                        | [     Save     ] | |
+----------------+----------------------------------------------+
```

**Key layout differences from DR-003-001-03 (superseded):**
- No Cancel/Save buttons in the header action bar — the new design omits them
- Left action panel (162px) added with 3 buttons: Request Details, Update Request, Delete Request
- The page title breadcrumb now shows the ticket reference: `#[ticketId] - [RequestType]`
- Navigation back to list is via the back arrow (ArrowLeft icon) in the title row

### Design Gaps (from Figma extraction 2026-04-03)

| Gap | Business Impact | Resolution |
|-----|-----------------|------------|
| Section header reads "Service Ticket Info" | Old module name — inconsistent with "Request Ticket" branding | Rename to "Request Info" to match Create form and business naming |
| Field label reads "Request details" | Unclear scope — subject/title not shown as a separate field in Figma | Business requirement: keep separate Subject (title) + Description fields as defined in Create form; "Request details" maps to Description only |
| Subject field missing from Figma | Create form has 3 input fields; Update form shows only 2 | Add Subject field between Request Type and Description — consistent with Create form and list display |
| Hint banner shown in both states | Initial state shows hint; populated state shows info card | In edit form context, the employee is always pre-filled — hint banner should never show on edit page load |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Page Display:**
- **AC-01:** Page title displays "Update Request Ticket" (Geist Semibold 24px) with a back arrow (ArrowLeft icon) to the left
- **AC-02:** Page title row includes a breadcrumb segment showing the ticket reference: `> #[ticketId] - [RequestType]`
- **AC-03:** Left action panel (162px) displays three buttons: Request Details, Update Request (active/highlighted), Delete Request
- **AC-04:** Edit form (600px) is positioned to the right of the action panel, not full-width
- **AC-05:** Edit form has two cards: Employee card (read-only) + Request Info card (editable)
- **AC-06:** Single Save button at the bottom of the form — full-width 600px, background #010101

**Pre-fill Behavior:**
- **AC-07:** All Request Info fields pre-filled with current ticket values on page load: Request Type, Subject, Description
- **AC-08:** Attachment field shows current filename if the ticket has an existing attachment
- **AC-09:** Employee field is pre-filled with the submitter's name and displayed as locked (read-only, grayed style) — info card immediately visible without requiring selection
- **AC-10:** Employee info card always visible on page load, showing the ticket submitter's profile: Full Name, Email, Phone, Department, Position (no "Select an employee" hint shown)

**Access Control:**
- **AC-11:** "Edit" option in gear icon visible only for eligible tickets: Open for employees owning the ticket; Open or In Progress for managers/admins
- **AC-12:** "Edit" option hidden when ticket is On Hold, Resolved, Closed, or Cancelled for all users
- **AC-13:** Employees can only edit their own tickets; direct URL access to another employee's edit page redirects to fallback
- **AC-14:** Direct URL access by users without request ticket permission redirects to fallback page

**Editable Status Enforcement:**
- **AC-15:** Employee can edit own ticket only when status is Open
- **AC-16:** Manager/admin can edit ticket when status is Open or In Progress
- **AC-17:** Editing does not change the ticket's current status — an In Progress ticket remains In Progress after save

**Left Panel Button Rules:**
- **AC-18:** "Update Request" button in left panel is only visible when the ticket is in an editable status per the user's role (same rules as gear icon Edit visibility)
- **AC-19:** "Delete Request" button in left panel follows delete permission rules (DR-003-001-04)
- **AC-20:** "Request Details" button navigates to the ticket detail view (DR-003-001-05)

**Request Info Fields:**
- **AC-21:** Request Type pre-filled with current value; same 8 options: IT Support, Facility, HR Inquiry, Office Supplies, Access Request, Travel & Expense, Training, Other
- **AC-22:** Subject pre-filled with current value; max 200 characters enforced
- **AC-23:** Description pre-filled with current value; mandatory
- **AC-24:** Attachment shows current filename if exists; "Choose File / No file chosen" if no attachment

**Attachment Behavior:**
- **AC-25:** User can keep current attachment by making no change to the file input
- **AC-26:** User can replace attachment by selecting a new file; new file replaces the old
- **AC-27:** User can remove attachment entirely by clearing the file input
- **AC-28:** New attachment validated: PDF, PNG, JPG, DOCX; max 5MB
- **AC-29:** Invalid file type shows: "Only PDF, PNG, JPG, and DOCX files are accepted"
- **AC-30:** File exceeding 5MB shows: "File size must not exceed 5MB"

**Mandatory Field Validation:**
- **AC-31:** Mandatory fields: Request Type, Subject, Description
- **AC-32:** Empty mandatory fields show inline error: "[Field name] is required"
- **AC-33:** Subject and Description trimmed of whitespace; whitespace-only rejected as empty
- **AC-34:** Subject exceeding 200 characters shows: "Subject must not exceed 200 characters"

**Save Behavior:**
- **AC-35:** Save updates the request ticket; ticket status remains unchanged
- **AC-36:** Success toast: "Request ticket has been updated"
- **AC-37:** After save, redirects to Request Tickets List — updated ticket visible with unchanged status badge
- **AC-38:** Save button shows spinner + disabled while processing; all fields disabled

**Navigation/Cancel Behavior:**
- **AC-39:** Back arrow on untouched form redirects to list without confirmation
- **AC-40:** Back arrow on modified form shows "Discard unsaved changes?" dialog
- **AC-41:** "Form modified" includes any change to Request Type, Subject, Description, or Attachment

**Concurrent Edit Protection:**
- **AC-42:** If ticket status was changed by another user while the edit form was open, saving shows error toast: "This ticket can no longer be edited" and redirects to list
- **AC-43:** Server validates the ticket is still in an editable status before applying changes

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Employee edits own Open ticket | Modify Subject, Save | Ticket updated, status stays Open, toast, redirect | High |
| Manager edits In Progress ticket | Modify Description, Save | Ticket updated, status stays In Progress, toast, redirect | High |
| Fields pre-filled on load | Open edit page | All fields show current ticket values | High |
| Employee field locked and info card shown | Open edit page (any role) | Employee field read-only; info card visible immediately | High |
| No hint banner on edit page | Open edit page | "Select an employee to view their info" banner NOT shown | High |
| Employee edits non-Open ticket | On Hold ticket (employee) | Edit not visible in gear icon; Update Request not in left panel | High |
| Manager edits On Hold ticket | On Hold ticket (manager) | Edit not visible in gear icon; Update Request not in left panel | High |
| Empty Subject | Clear Subject, Save | "Subject is required" inline error | High |
| Whitespace-only Description | Enter "   ", Save | "Description is required" | High |
| Subject max length | Enter 201 characters | "Subject must not exceed 200 characters" | Medium |
| Keep existing attachment | No file change, Save | Original attachment preserved | Medium |
| Replace attachment | Upload new PDF | New filename shown; replaces old | Medium |
| Remove attachment | Clear file input, Save | Ticket saved without attachment | Medium |
| Invalid attachment | Upload .exe | Error: accepted formats message | Medium |
| Large attachment | Upload 6MB file | Error: max 5MB message | Medium |
| Back arrow dirty | Modify fields, click back | Discard dialog | Medium |
| Back arrow clean | No changes, click back | Redirect immediately | Medium |
| Breadcrumb in title | View edit page | Title shows `> #[ticketId] - [RequestType]` | Medium |
| Left panel buttons visible | Open In Progress ticket as manager | Request Details + Update Request + Delete Request in left panel | High |
| Update Request highlighted | View edit page | Update Request button in left panel shows active/highlighted state | Medium |
| Concurrent status change | Another user closes ticket while editing, then Save | Error toast "This ticket can no longer be edited" + redirect | High |
| Server error on save | Server fails during save | Error toast; form preserved | Medium |
| Unauthorized access | No permission | Redirect to fallback page | High |
| Employee edits another user's ticket | Direct URL access | Redirect to fallback page | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Edit page accessible only to users with request ticket access permission (US-004)
- **SR-02:** Employees can edit only their own tickets in Open status — enforced server-side
- **SR-03:** Managers/admins can edit any ticket in Open or In Progress status — enforced server-side
- **SR-04:** On Hold, Resolved, Closed, and Cancelled tickets are not editable by any role
- **SR-05:** Editing does not change the ticket's current status — no re-approval workflow applies to request tickets
- **SR-06:** The submitter (employee field) is locked — cannot be changed in the edit form
- **SR-07:** Employee info card is read-only — populated from employee profile (US-005); shows Full Name, Email, Phone, Department, Position; always pre-filled on edit page load (no selection step)
- **SR-08:** All validation rules identical to Create Request Ticket (DR-003-001-02): Request Type required, Subject required (max 200 chars, trimmed), Description required (trimmed), Attachment optional (PDF/PNG/JPG/DOCX, max 5MB)
- **SR-09:** Concurrent edit protection — server validates the ticket is still in an editable status at time of save; if not, returns conflict error
- **SR-10:** "Form dirty" detection covers all editable fields: Request Type, Subject, Description, Attachment. Any change triggers discard confirmation on back navigation
- **SR-11:** After update, ticket appears in the list with unchanged status badge. Visible per data visibility rules (employee sees own; manager/admin sees all)
- **SR-12:** Attachment handling: no change = existing attachment preserved; new file uploaded = replaces existing; file cleared = attachment removed from ticket
- **SR-13:** Request Date is not updated on edit — it retains the original submission date
- **SR-14:** Left panel "Update Request" and "Delete Request" buttons render conditionally — "Update Request" shown only when ticket is in an editable status per the user's role; "Delete Request" shown per delete permission rules (DR-003-001-04)

**State Transitions:**
```
[List] → gear → "Edit" → [Edit Form, pre-filled, two-panel layout]
[Edit Form] → [Save, valid, status OK] → [ticket updated, status unchanged] → [toast] → [list]
[Edit Form] → [Save, invalid fields] → [inline errors] → [stay on form]
[Edit Form] → [Save, ticket no longer editable] → [error toast] → [list]
[Edit Form] → [Save, server error] → [error toast] → [stay on form]
[Edit Form] → [Back arrow (clean)] → [list]
[Edit Form] → [Back arrow (dirty)] → [discard dialog] → confirm → [list]
[Edit Form] → [Request Details (left panel)] → [ticket detail view]
[Edit Form] → [Delete Request (left panel)] → [delete confirmation dialog]
```

**Dependencies:**
- **US-001 (Authentication):** User identity for access control
- **US-004 (Role & Permission Management):** Controls access; determines editable scope per role
- **US-005 (User Management):** Submitter profile data for employee info card
- **EP-008 US-001/002 (Department/Position):** Department and position in info card
- **DR-003-001-01 (Request Tickets List):** Entry point via gear icon "Edit" action
- **DR-003-001-02 (Create Request Ticket):** Mirrors same form field rules and validation
- **DR-003-001-04 (Delete Request Ticket):** Delete Request button in left panel follows these rules
- **DR-003-001-05 (Request Ticket Details):** Request Details button in left panel navigates here

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** All fields pre-filled on page load — user sees current values immediately without extra steps
- **UX-02:** Employee info card always visible and pre-filled — no selection step needed in edit context; provides submitter context for all roles at a glance
- **UX-03:** Employee field displayed as locked with visual cue (grayed/disabled style) — prevents confusion about whether the submitter can be changed
- **UX-04:** Attachment shows current filename if exists — user knows what is attached before deciding to keep, replace, or remove
- **UX-05:** Two-panel layout with left action panel — user can navigate to Request Details or trigger Delete from the same page, reducing round-trips between pages
- **UX-06:** "Update Request" button in left panel shown in active/highlighted state — visual feedback that the user is currently on the edit view
- **UX-07:** Breadcrumb `#[ticketId] - [RequestType]` in title row — user always knows which ticket is being edited
- **UX-08:** Save button shows spinner while processing — prevents double submission
- **UX-09:** Concurrent conflict shown as toast with clear reason — user understands why save failed and is redirected to refreshed list
- **UX-10:** Success toast auto-dismisses after 5 seconds; message is clear and consistent with Create form pattern
- **UX-11:** Discard dialog defaults focus on Cancel (stay) — prevents accidental data loss
- **UX-12:** Form card centered (600px) — consistent with Create Request Ticket

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>=1280px) | Two-panel layout: left action panel (162px) + right form card (600px) |
| Below desktop | Out of scope for this release |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through all editable fields in logical order; locked employee field skipped
- [x] Screen reader compatible — all field labels announced; employee info card announced as read-only
- [x] Locked employee field announced as "read-only" to screen readers
- [x] Request Type dropdown accessible via keyboard
- [x] Left panel action buttons keyboard accessible with logical tab order
- [x] Color contrast meets WCAG 2.1 AA
- [x] Focus indicators visible on all interactive elements
- [x] Error messages associated with fields via aria-describedby
- [x] Discard dialog traps keyboard focus

**Design References:**
- Figma (initial state): [Update Request Ticket](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3213-4286) (node `3213:4286`)
- Figma (populated state): [Update Request Ticket](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3213-4524) (node `3213:4524`)
- Pattern reference: Request Ticket Details (DR-003-001-05) for two-panel layout with left action panel
- Pattern reference: Create Request Ticket (DR-003-001-02) for form field rules and validation
- Design tokens: primary `#010101`, background `#ffffff`, border `#e5e5e5`, muted foreground `#737373`, font: Geist

---

## 8. Additional Information

### Out of Scope
- Status management (Change to In Progress, On Hold, Resolve, etc.) — handled via detail page (DR-003-001-05)
- Delete Request Ticket — separate DR (DR-003-001-04)
- Editing On Hold, Resolved, Closed, or Cancelled tickets — not supported in this release
- Changing the submitter (employee field) in the edit form — submitter is locked
- Changing the Request Date — auto-recorded at creation, not editable
- Edit history / audit trail of changes — future enhancement
- Multiple file attachments — single file only for this release
- Mobile or tablet layout
- Bulk editing of multiple tickets
- Auto-save / draft preservation on session timeout

### Open Questions (Resolved)

| Question | Answer | Confirmed By |
|----------|--------|-------------|
| Does editing change ticket status? | No — editing preserves current status (no re-approval workflow) | Knowledge Base §4.2 |
| Is employee info card always visible? | Yes — always pre-filled and visible as read-only for all roles | Knowledge Base §4.1 |
| Can attachment be removed entirely? | Yes — clearing the file input removes the attachment | Knowledge Base §4.4 |
| Entry point | Gear icon → "Edit" in list view | Knowledge Base §1.5, DR-003-001-01 |
| Success toast message | "Request ticket has been updated" | Knowledge Base §4.1 |
| Should header Cancel/Save be present? | No — updated Figma design (node 3213:4286/4524) omits header buttons; back arrow + single bottom Save only | Figma extraction 2026-04-03 |
| Is there a left action panel? | Yes — two-panel layout confirmed: left panel (162px) with Request Details, Update Request, Delete Request | Figma extraction 2026-04-03 |

### Open Questions (Pending)

- [ ] **Subject field missing from Figma:** The updated Figma shows only "Request type" + "Request details" in the info section (2 fields). The Create form has 3 fields (Request Type, Subject, Description). Is "Request details" a combined Subject+Description field, or should a separate Subject field be added above it? — **Owner:** Design Team — **Status:** Pending
- [ ] **Section header naming:** Figma still reads "Service Ticket Info". Business name should be "Request Info" — confirm final label. — **Owner:** Design Team — **Status:** Pending
- [ ] **Attachment removal UX:** Should removing an existing attachment show an explicit "Remove" button/link, or is clearing the native file input sufficient? — **Owner:** Product Owner — **Status:** Pending
- [ ] **"Update Request" button label:** Left panel button reads "Update Request" in Figma. Confirm this is the final label (vs. "Edit Request"). — **Owner:** Design Team — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-003-001-01: Request Tickets List | Entry point via gear icon "Edit" action |
| DR-003-001-02: Create Request Ticket | Same form field rules; Edit mirrors Create with pre-filled data |
| DR-003-001-03: Update Request Ticket (v1) | Superseded by this DR — written without updated Figma design |
| DR-003-001-04: Delete Request Ticket | Delete Request button in left panel follows these rules |
| DR-003-001-05: Request Ticket Details | Request Details button in left panel navigates here; same two-panel layout origin |
| US-001: Authentication | User identity for access control |
| US-004: Role & Permission Management | Determines editable scope per role |
| US-005: User Management | Submitter profile for employee info card |
| EP-008 US-001/002: Department & Position | Info card data |

### Notes
- **Supersedes DR-003-001-03** — that DR was written based on an earlier Figma extraction (node `3181:1976`). The updated design (node `3213:4286` and `3213:4524`) introduces a two-panel layout with a left action panel, removes the header Cancel/Save buttons, and adds a ticket reference breadcrumb in the title row. DR-003-001-06 reflects the current design state.
- **Two-panel layout matches Detail page pattern (§3.1):** The edit form adopts the same layout as the Request Ticket Details page — left panel for navigation/actions, right panel for content. This is consistent with the detail page and provides a unified context for all ticket-level actions.
- **No header Cancel/Save:** Unlike the Create form pattern (§2.1) which places Cancel + Save in the header, the updated edit form uses only a back arrow for navigation and a single bottom Save. This is a confirmed design deviation for the edit context.
- **Employee info card pre-filled without selection:** Unlike the Create form where the employee section starts empty (with hint banner), the edit form always opens with the submitter pre-filled and the info card visible. No selection interaction occurs.
- **No status revert on edit:** Request Tickets have no approval lifecycle, so editing does not revert the status (unlike Leave Requests where editing an Approved request reverts it to Pending).
- **Request Date preserved:** The original submission date is not updated when the ticket is edited.

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | — | — | Pending |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-04-03 | BA Agent | Initial draft — supersedes DR-003-001-03; reflects updated Figma design (nodes 3213:4286, 3213:4524); two-panel layout confirmed (left action panel: Request Details, Update Request, Delete Request); header Cancel/Save removed; back arrow + single bottom Save; breadcrumb with ticket reference; 4 design gaps documented |
