---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-003
story_id: US-001
story_name: "Request Tickets"
detail_id: DR-003-001-03
detail_name: "Update Request Ticket"
parent_requirement: FR-US-001-10
status: draft
version: "1.0"
created_date: 2026-04-03
last_updated: 2026-04-03
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
input_sources:
  - type: figma
    description: "Update Request Ticket screen"
    node_id: "3181:1976"
    extraction_date: "2026-04-03"
---

# Detail Requirement: Update Request Ticket

**Detail ID:** DR-003-001-03
**Parent Requirement:** FR-US-001-10
**Story:** US-001-request-tickets
**Epic:** EP-003 (Request Ticket Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **employee with request ticket access**, I want to **edit a request ticket that is still in Open status**, so that **I can correct or update the request details before it is processed**.

As a **manager or administrator with request ticket management permission**, I want to **edit a request ticket in Open or In Progress status**, so that **I can refine the ticket details as the request is being handled**.

**Purpose:** Allow authorized users to update an existing request ticket's content. The form mirrors the Create Request Ticket form with all fields pre-filled from the current ticket data. The submitting employee is locked — the ticket cannot be reassigned to a different person. Editing does not change the ticket's current status (no re-approval workflow applies to request tickets).

**Target Users:**
- **Employees** — can edit only their own tickets in **Open** status
- **Managers/Administrators** — can edit any ticket in **Open or In Progress** status

**Key Functionality:**
- Full-page edit form pre-filled with current ticket data
- Employee info card always visible as read-only context (locked — cannot change submitter)
- 3 mandatory fields + 1 optional (Attachment), same as Create form
- Editing does not change ticket status
- Concurrent edit protection — server rejects save if ticket status changed during editing
- Dual Save buttons: header Save + full-width bottom Save (both trigger same action)

---

## 2. User Workflow

**Entry Point:** Request Tickets List → Gear icon on a row → select "Edit" (only visible for editable-status tickets per role)

**Preconditions:**
- User is signed in (US-001)
- User has request ticket access permission (US-004)
- Ticket is in an editable status:
  - **Employees:** ticket is in **Open** status AND belongs to the logged-in user
  - **Managers/Admins:** ticket is in **Open** or **In Progress** status

**Gear Icon Edit Visibility Matrix:**

| Ticket Status | Employee (own ticket) | Manager/Admin |
|---|---|---|
| Open | Edit visible | Edit visible |
| In Progress | Edit not visible | Edit visible |
| On Hold | Edit not visible | Edit not visible |
| Resolved | Edit not visible | Edit not visible |
| Closed | Edit not visible | Edit not visible |
| Cancelled | Edit not visible | Edit not visible |

**Main Flow A -- Employee editing own Open ticket:**
1. Employee clicks gear icon on their Open ticket in the list
2. Employee selects "Edit" from the dropdown menu
3. System navigates to "Update Request Ticket" page
4. Employee section displays the submitter's read-only info card (the logged-in employee's own profile)
5. Request Info section displays pre-filled fields: Request Type, Subject, Description, Attachment
6. Employee modifies the desired fields
7. Employee clicks Save
8. System validates all mandatory fields
9. System updates the ticket; status remains **Open**
10. Success toast: "Request ticket has been updated"
11. System redirects to Request Tickets List

**Main Flow B -- Manager/Admin editing Open or In Progress ticket:**
1. Manager clicks gear icon on an Open or In Progress ticket
2. Manager selects "Edit" from the dropdown menu
3. System navigates to "Update Request Ticket" page
4. Employee section displays the submitter's read-only info card (locked employee field)
5. Request Info section displays pre-filled fields: Request Type, Subject, Description, Attachment
6. Manager modifies the desired fields
7. Manager clicks Save
8. System validates all mandatory fields
9. System checks the ticket is still in an editable status (concurrent edit protection)
10. System updates the ticket; status remains unchanged (Open or In Progress)
11. Success toast: "Request ticket has been updated"
12. System redirects to Request Tickets List

**Alternative Flows:**

| Action | Flow |
|--------|------|
| **Alt 1 -- Empty mandatory field** | Inline error "[Field name] is required". Form not submitted. |
| **Alt 2 -- Subject whitespace-only** | Inline error: "Subject is required". Form not submitted. |
| **Alt 3 -- Description whitespace-only** | Inline error: "Description is required". Form not submitted. |
| **Alt 4 -- Subject exceeds max length** | Inline error: "Subject must not exceed 200 characters". Form not submitted. |
| **Alt 5 -- Invalid attachment file type** | Inline error: "Only PDF, PNG, JPG, and DOCX files are accepted". File rejected. |
| **Alt 6 -- Attachment too large** | Inline error: "File size must not exceed 5MB". File rejected. |
| **Alt 7 -- Cancel (form modified)** | "Discard unsaved changes?" dialog. Confirm → redirect to list. Cancel → stay on form. |
| **Alt 8 -- Cancel (form untouched)** | Redirect to Request Tickets List immediately. |
| **Alt 9 -- Status changed during editing** | On Save: error toast "This ticket can no longer be edited" + redirect to list. |
| **Alt 10 -- Server error on save** | Error toast: "Something went wrong. Please try again." Form stays open, fields preserved. |

**Exit Points:**
- **Success:** Ticket updated → toast → redirect to Request Tickets List
- **Cancel:** Redirect to list (with or without confirmation)
- **Error:** Validation errors inline (user corrects and retries); server conflict (redirect to list)

---

## 3. Field Definitions

### Employee Section (always visible, read-only)

| Field Name | Field Type | Validation Rule | Mandatory | Default | Description |
|------------|------------|-----------------|-----------|---------|-------------|
| Employee | Read-only display (576px) | Not editable — locked to original submitter | N/A | Pre-filled with ticket submitter | Cannot be changed; shown as locked field |

### Employee Info Card (read-only, always shown)

| Field | Left Column | Right Column |
|-------|------------|-------------|
| Row 1 | Full name | Department |
| Row 2 | Email | Position |
| Row 3 | Phone number | -- |

**Note:** Info card is always visible for all roles in the edit form (unlike the Create form where employees see the employee section hidden). This provides context about whose ticket is being edited.

### Input Fields -- Request Info Section (pre-filled, editable)

| Field Name | Field Type | Validation Rule | Mandatory | Default | Description |
|------------|------------|-----------------|-----------|---------|-------------|
| Request Type | Dropdown (576px) | Must select a value | Yes (*) | Pre-filled from current ticket | IT Support, Facility, HR Inquiry, Office Supplies, Access Request, Travel & Expense, Training, Other |
| Subject | Text input (576px) | Not empty; trimmed; max 200 characters | Yes (*) | Pre-filled from current ticket | Short title for the request (displayed as "Request" column in list) |
| Description | Text area (576px) | Not empty; trimmed | Yes (*) | Pre-filled from current ticket | Detailed explanation of the request |
| Attachment | File input (576px) | PDF, PNG, JPG, DOCX; max 5MB | No | Shows current file if exists | Supporting document — can be kept, replaced, or removed |

### Interaction Elements

| Element | Type | Position | State/Condition | Trigger Action | Description |
|---------|------|----------|-----------------|----------------|-------------|
| Cancel | Button (secondary) | Header action bar | Always visible | If dirty → discard dialog; if clean → redirect | Returns to list |
| Save (header) | Button (primary) | Header action bar | Disabled + spinner while saving | Validates → updates ticket → toast → redirect | Submit changes |
| Save (bottom) | Button (primary, full-width) | Below form, 600x40px | Same behavior as header Save | Identical to header Save | Convenience duplicate |

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
| Employee field | Locked input (576px) — shows submitter name, grayed/disabled style | Always pre-filled | Cannot change submitter |
| Employee info card | 5 fields in 2-column layout | Always shown | Read-only context: who submitted this ticket |

**Request Info Section (pre-filled, editable):**

| Data Name | Format | Business Meaning |
|-----------|--------|-----------------|
| Request Type | Dropdown (576px) — pre-filled with current value | Category of the request |
| Subject | Text input (576px) — pre-filled with current value | Short title displayed in list "Request" column |
| Description | Text area (576px) — pre-filled with current value | Detailed explanation |
| Attachment | File input (576px) — shows current filename if exists; "Choose File / No file chosen" if none | Supporting document |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default (page load) | All fields pre-filled from current ticket | Employee info card visible; Request Info fields editable with current values |
| Attachment exists | Ticket has an existing attachment | Current filename shown next to "Choose File" |
| Attachment removed | User clears the attachment | "No file chosen" displayed |
| New file selected | User uploads a replacement file | New filename displayed; replaces current |
| Validation error | Save with invalid/empty fields | Inline red error(s) below affected fields |
| Saving | Save clicked, in progress | Both Save buttons spinner + disabled; all fields disabled |
| Success | Ticket updated | Toast "Request ticket has been updated" → redirect to list |
| Discard confirmation | Cancel with modified form | Modal: "Discard unsaved changes?" |
| Server conflict | Status changed by another user during edit | Error toast: "This ticket can no longer be edited" → redirect to list |

### Page Layout (Figma Reference: node `3181:1976`)

```
+---------------------------------------------------------------+
|  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]  |
+--------------+------------------------------------------------+
|  [Sidebar]   |  Update Request Ticket         [Cancel] [Save]  |
|  200px       |                                                 |
|              |     +------------------------------------+      |
|              |     | Employee                           |      |
|              |     | * Employee [locked, pre-filled]    |      |
|              |     |                                    |      |
|              |     | +-------------+-----------------+  |      |
|              |     | | Full name   | Department      |  |      |
|              |     | | Email       | Position        |  |      |
|              |     | | Phone       |                 |  |      |
|              |     | +-------------+-----------------+  |      |
|              |     +------------------------------------+      |
|              |     | Request Info                        |      |
|              |     | *Request Type [pre-filled]          |      |
|              |     | *Subject [pre-filled]               |      |
|              |     | *Description [pre-filled]           |      |
|              |     |  Attachment [current file or new]   |      |
|              |     +------------------------------------+      |
|              |     | [             Save               ]  |      |
|              |     +------------------------------------+      |
+--------------+------------------------------------------------+
```

**Design Gaps (from Figma extraction 2026-04-03):**

| Gap | Impact | Resolution |
|-----|--------|------------|
| Section 2 header reads "Service Ticket Info" | Old module name — should read "Request Info" | Rename to "Request Info" per business naming |
| Field reads "Service type" | Old field name — should read "Request Type" | Rename to "Request Type" per Create form |
| Only 2 input fields shown (Service type + Request details) | Subject field missing from design | Add Subject field between Request Type and Description; "Request details" = Description only |
| Employee shown in empty/unselected state | Edit form should pre-fill and lock the employee | Design locked pre-filled state for employee field |

---

## 5. Acceptance Criteria

**Definition of Done -- All criteria must be met:**

**Page Display:**
- **AC-01:** Page title displays "Update Request Ticket" in Geist Semibold 24px
- **AC-02:** Form has two sections: Employee (card 1, read-only) + Request Info (card 2, editable) — centered 600px card
- **AC-03:** Cancel + Save buttons in header action bar; Save button also at bottom (full-width 600px)
- **AC-04:** Both Save buttons perform identical submission logic

**Pre-fill Behavior:**
- **AC-05:** All Request Info fields pre-filled with current ticket values on page load: Request Type, Subject, Description
- **AC-06:** Attachment field shows current file (filename) if the ticket has an existing attachment
- **AC-07:** Employee info card always visible for all roles, showing the ticket submitter's read-only profile: Full Name, Email, Phone, Department, Position
- **AC-08:** Employee field is locked — cannot change the submitter

**Access Control:**
- **AC-09:** "Edit" option in gear icon visible only for eligible tickets (Open for employees owning the ticket; Open or In Progress for managers/admins)
- **AC-10:** "Edit" option hidden in gear icon when ticket is On Hold, Resolved, Closed, or Cancelled for all users
- **AC-11:** Employees can only edit their own tickets; they cannot access the edit page for other employees' tickets
- **AC-12:** Direct URL access to edit page by unauthorized users redirects to fallback page

**Editable Status Enforcement:**
- **AC-13:** Employee can edit own ticket only when status is Open
- **AC-14:** Manager/admin can edit ticket when status is Open or In Progress
- **AC-15:** Editing does not change the ticket's current status — an In Progress ticket remains In Progress after editing

**Request Info Fields:**
- **AC-16:** Request Type pre-filled with current value; same 8 options as Create (IT Support, Facility, HR Inquiry, Office Supplies, Access Request, Travel & Expense, Training, Other)
- **AC-17:** Subject pre-filled with current value; max 200 characters enforced
- **AC-18:** Description pre-filled with current value; mandatory
- **AC-19:** Attachment shows current filename if exists; "Choose File / No file chosen" if no attachment

**Attachment Behavior:**
- **AC-20:** User can keep current attachment by making no change to the file input
- **AC-21:** User can replace attachment by selecting a new file; new file replaces the old
- **AC-22:** User can remove attachment entirely (clear file input)
- **AC-23:** New attachment validated: PDF, PNG, JPG, DOCX; max 5MB
- **AC-24:** Invalid file type shows: "Only PDF, PNG, JPG, and DOCX files are accepted"
- **AC-25:** File exceeding 5MB shows: "File size must not exceed 5MB"

**Mandatory Field Validation:**
- **AC-26:** Mandatory fields: Request Type, Subject, Description
- **AC-27:** Empty mandatory fields show inline error: "[Field name] is required"
- **AC-28:** Subject and Description trimmed of whitespace; whitespace-only rejected as empty
- **AC-29:** Subject exceeding 200 characters shows: "Subject must not exceed 200 characters"

**Save Behavior:**
- **AC-30:** Save updates the request ticket; ticket status remains unchanged
- **AC-31:** Success toast: "Request ticket has been updated"
- **AC-32:** After save, redirects to Request Tickets List — updated ticket visible with unchanged status badge
- **AC-33:** Save buttons show spinner + disabled while processing; all fields disabled

**Cancel Behavior:**
- **AC-34:** Cancel on untouched form redirects without confirmation
- **AC-35:** Cancel on modified form shows "Discard unsaved changes?" dialog
- **AC-36:** "Form modified" includes any change to Request Type, Subject, Description, or Attachment

**Concurrent Edit Protection:**
- **AC-37:** If ticket status was changed by another user while the edit form was open, saving shows error toast: "This ticket can no longer be edited" and redirects to list
- **AC-38:** Server validates the ticket is still in an editable status before applying changes

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Employee edits own Open ticket | Modify Subject, Save | Ticket updated, status stays Open, toast, redirect | High |
| Manager edits In Progress ticket | Modify Description, Save | Ticket updated, status stays In Progress, toast, redirect | High |
| Fields pre-filled on load | Open edit form | All fields show current ticket values | High |
| Employee field locked | Open edit form (any role) | Employee field read-only; info card shown | High |
| Employee edits non-Open ticket | On Hold ticket (employee) | Edit not visible in gear icon | High |
| Manager edits On Hold ticket | On Hold ticket (manager) | Edit not visible in gear icon | High |
| Empty subject | Clear Subject, Save | "Subject is required" inline error | High |
| Whitespace-only description | Enter "   ", Save | "Description is required" | High |
| Subject max length | Enter 201 characters | "Subject must not exceed 200 characters" | Medium |
| Keep existing attachment | No file change, Save | Original attachment preserved | Medium |
| Replace attachment | Upload new PDF | New filename shown; replaces old | Medium |
| Remove attachment | Clear file input, Save | Ticket saved without attachment | Medium |
| Invalid attachment | Upload .exe | Error: accepted formats | Medium |
| Large attachment | Upload 6MB file | Error: max 5MB | Medium |
| Cancel dirty | Modify fields, Cancel | Discard dialog | Medium |
| Cancel clean | No changes, Cancel | Redirect immediately | Medium |
| Concurrent status change | Another user closes ticket while editing, then Save | Error toast + redirect to list | High |
| Server error | Server fails during save | Error toast; form preserved | Medium |
| Unauthorized access | No permission | Redirect to fallback page | High |
| Employee edits another employee's ticket | Direct URL access | Redirect to fallback page | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Edit page accessible only to users with request ticket access permission (US-004)
- **SR-02:** Employees can edit only their own tickets in Open status — enforced server-side
- **SR-03:** Managers/admins can edit any ticket in Open or In Progress status — enforced server-side
- **SR-04:** On Hold, Resolved, Closed, and Cancelled tickets are not editable by any role
- **SR-05:** Editing does not change the ticket's current status — no re-approval workflow applies to request tickets
- **SR-06:** The submitter (employee field) is locked — cannot be changed in the edit form
- **SR-07:** Employee info card is read-only — populated from employee profile (US-005); shows Full Name, Email, Phone, Department, Position
- **SR-08:** All validation rules identical to Create Request Ticket (DR-003-001-02): Request Type required, Subject required (max 200 chars, trimmed), Description required (trimmed), Attachment optional (PDF/PNG/JPG/DOCX, max 5MB)
- **SR-09:** Concurrent edit protection — server validates the ticket is still in an editable status at time of save; if not, returns conflict error
- **SR-10:** Both Save buttons (header + bottom) trigger identical save logic
- **SR-11:** "Form dirty" detection covers all editable fields: Request Type, Subject, Description, Attachment. Any change triggers discard confirmation on Cancel
- **SR-12:** After update, ticket appears in the list with unchanged status badge. Visible per data visibility rules (employee sees own; manager/admin sees all)
- **SR-13:** Attachment handling: no change = existing attachment preserved; new file uploaded = replaces existing; file cleared = attachment removed from ticket
- **SR-14:** Request Date is not updated on edit — it retains the original submission date

**State Transitions:**
```
[List] → gear → "Edit" → [Edit Form, pre-filled]
[Edit Form] → [Save, valid, status OK] → [ticket updated, status unchanged] → [toast] → [list]
[Edit Form] → [Save, invalid fields] → [inline errors] → [stay on form]
[Edit Form] → [Save, ticket no longer editable] → [error toast] → [list]
[Edit Form] → [Save, server error] → [error toast] → [stay on form]
[Edit Form] → [Cancel (clean)] → [list]
[Edit Form] → [Cancel (dirty)] → [discard dialog] → confirm → [list]
```

**Dependencies:**
- **US-001 (Authentication):** User identity for access control
- **US-004 (Role & Permission Management):** Controls access; determines editable scope per role
- **US-005 (User Management):** Submitter profile data for employee info card
- **EP-008 US-001/002 (Department/Position):** Department and position in info card
- **DR-003-001-01 (Request Tickets List):** Entry point via gear icon "Edit" action
- **DR-003-001-02 (Create Request Ticket):** Mirrors same form layout and field rules

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** All fields pre-filled on page load — user sees current values immediately without extra steps
- **UX-02:** Employee info card always visible — provides submitter context for all roles, consistent with Update Leave Request
- **UX-03:** Employee field displayed as locked/read-only with visual cue (grayed style) — prevents confusion about whether submitter can be changed
- **UX-04:** Attachment shows current filename if exists — user knows what is already attached before deciding to keep, replace, or remove
- **UX-05:** Both Save buttons perform the same action — save from anywhere on the page (header or after scrolling to bottom)
- **UX-06:** Save buttons show spinner while processing — prevents double submission
- **UX-07:** Concurrent conflict shown as toast with clear reason — user understands why save failed and is redirected to refreshed list
- **UX-08:** Success toast auto-dismisses after 5 seconds; message is clear and consistent with Create form pattern
- **UX-09:** Discard dialog defaults focus on Cancel (stay) — prevents accidental data loss
- **UX-10:** Form card centered (600px) — consistent with Create Request Ticket and Create Leave Request

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>=1280px) | Centered form card 600px |
| Below desktop | Out of scope for this release |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through all editable fields in logical order; locked fields skipped
- [x] Screen reader compatible — labels for all fields; info card announced as read-only
- [x] Locked employee field announced as "read-only" to screen readers
- [x] Request Type dropdown accessible via keyboard
- [x] Color contrast meets WCAG 2.1 AA
- [x] Focus indicators visible on all interactive elements
- [x] Error messages associated with fields via aria-describedby
- [x] Discard dialog traps keyboard focus

**Design References:**
- Pattern reference: Update Leave Request (DR-002-001-03) for edit form with locked employee
- Pattern reference: Create Request Ticket (DR-003-001-02) for form layout and field rules
- Figma reference: [Update Request Ticket](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3181-1976) (node `3181:1976`)
- Design tokens: primary `#171717`, background `#ffffff`, border `#e5e5e5`, muted foreground `#737373`, font: Geist

---

## 8. Additional Information

### Out of Scope
- Status management (Change to In Progress, On Hold, Resolve, etc.) — separate DR (DR-003-001-04)
- Cancel Request Ticket — separate DR (DR-003-001-05)
- Delete Request Ticket — separate DR (DR-003-001-06)
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
| Does editing change ticket status? | No — editing preserves current status (no re-approval workflow) | Knowledge Base (Edit Form Pattern 4.2 note) |
| Is employee info card visible for employees editing own tickets? | Yes — always visible as read-only for all roles | Knowledge Base (Edit Form Pattern 4.1), confirmed by Update Leave Request pattern |
| Can attachment be removed entirely? | Yes — clearing the file input removes the attachment | Standard UX for edit forms with optional file upload |
| Entry point | Gear icon → "Edit" in list view | Knowledge Base (Gear Icon 1.5), DR-003-001-01 |
| Success toast message | "Request ticket has been updated" | Knowledge Base (Edit Form Pattern 4.1) |

### Open Questions (Pending)

- [ ] **Figma design gap — Subject field missing:** The Figma design for Update Request Ticket shows only 2 input fields ("Service type" + "Request details") versus the 3 fields in the Create form (Request Type, Subject, Description). Is "Request details" a combined Subject+Description field, or is Subject accidentally missing? — **Owner:** Design Team — **Status:** Pending
- [ ] **Figma naming inconsistency:** Design uses old names "Service Ticket Info" and "Service type". Confirm final business names: "Request Info" and "Request Type" — **Owner:** Design Team — **Status:** Pending
- [ ] **Attachment removal UX:** Should removing an existing attachment show an explicit "Remove" button/link, or is clearing the file input sufficient? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-003-001-01: Request Tickets List | Entry point via gear icon "Edit" action |
| DR-003-001-02: Create Request Ticket | Same form structure and field rules; Edit mirrors Create with pre-filled data |
| DR-003-001-04: Manage Request Ticket Status (planned) | Status changes handled separately — edit does not change status |
| DR-003-001-05: Cancel Request Ticket (planned) | Separate action from edit |
| DR-003-001-06: Delete Request Ticket (planned) | Separate action from edit |
| US-001: Authentication | User identity for access control |
| US-004: Role & Permission Management | Determines editable scope per role |
| US-005: User Management | Submitter profile for employee info card |
| EP-008 US-001/002: Department & Position | Info card data |

### Notes
- **Mirrors Create Request Ticket (DR-003-001-02)** — same two-card layout, same fields, same validation, same dual Save buttons. Key differences: fields pre-filled, employee locked (not hidden), editing does not change status.
- **Employee section behavior differs from Create:** In Create, employees see the employee section hidden (auto-assigned). In Edit, the employee info card is always visible for all roles as read-only context — consistent with Update Leave Request (DR-002-001-03).
- **No status revert on edit:** Request Tickets have no approval lifecycle, so editing does not revert the status (unlike Leave Requests where editing an Approved request reverts it to Pending).
- **Request Date preserved:** The original submission date is not updated when the ticket is edited.
- **Concurrent edit protection matches Leave Request:** Same pattern — server checks editable status before saving, returns conflict error if status has changed.
- **Figma page title confirmed as "Update Request Ticket"** — used as the page heading (node `3181:2034`).

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | -- | -- | Pending |
| Product Owner | -- | -- | Pending |
| UX Designer | -- | -- | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-04-03 | BA Agent | Initial draft — Edit form mirrors Create; employee locked (not hidden); no status revert on edit; Figma extracted (node 3181:1976); 3 design gaps flagged (naming, missing Subject field, empty employee state) |
