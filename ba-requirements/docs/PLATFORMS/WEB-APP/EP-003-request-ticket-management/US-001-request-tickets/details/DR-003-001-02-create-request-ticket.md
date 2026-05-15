---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-003
story_id: US-001
story_name: "Request Tickets"
detail_id: DR-003-001-02
detail_name: "Create Request Ticket"
parent_requirement: FR-US-001-09
status: draft
version: "1.0"
created_date: 2026-04-01
last_updated: 2026-04-01
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "./DR-003-001-01-request-tickets-list.md"
    relationship: sibling
---

# Detail Requirement: Create Request Ticket

**Detail ID:** DR-003-001-02
**Parent Requirement:** FR-US-001-09
**Story:** US-001-request-tickets
**Epic:** EP-003 (Request Ticket Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **employee with request ticket access**, I want to **submit a new request ticket by selecting a request type, entering a subject and description**, so that **my request is formally recorded and routed for handling by managers or administrators**.

As a **manager or administrator with request ticket management permission**, I want to **submit a request ticket on behalf of any employee by selecting the employee, specifying request details**, so that **requests can be recorded even when employees cannot submit them directly**.

**Purpose:** Allow authorized users to create request tickets. Employees create their own tickets (employee field hidden, auto-assigned to logged-in user); managers/administrators can create tickets on behalf of any employee by searching and selecting from the employee dropdown. The form displays the selected employee's profile summary to provide organizational context.

**Target Users:**
- **Employees** — create their own request tickets (Employee section hidden, auto-assigned to logged-in user)
- **Managers/Administrators** — create request tickets on behalf of any employee (Employee section visible with searchable dropdown)

**Key Functionality:**
- Dual-role form: simplified for employees, full for managers/administrators
- Two-section layout: Employee (with auto-populated info card) + Request Info
- Employee info card shows profile: Full Name, Email, Phone, Department, Position
- 4 mandatory fields + 1 optional (Attachment)
- All new tickets start with "Open" status
- Past dates not applicable (submission date auto-recorded by system)
- Dual Save buttons: header Save + full-width bottom Save (both trigger same action)

---

## 2. User Workflow

**Entry Point:** Request Tickets List -> click "+ Add Request" button

**Preconditions:**
- User is signed in (US-001)
- User has request ticket access permission (US-004)

**Main Flow A -- Employee creating own ticket:**
1. Employee clicks "+ Add Request" on Request Tickets List
2. System navigates to "Create A New Request Ticket" page
3. Employee section is **hidden** -- system auto-assigns to logged-in user
4. Request Info section displays with empty fields
5. Employee selects Request Type from dropdown
6. Employee enters Subject (short title for the request)
7. Employee enters Description (detailed explanation)
8. Employee optionally uploads an Attachment
9. Employee clicks Save
10. System validates all mandatory fields
11. System creates the request ticket with "Open" status
12. Success toast: "Request ticket has been submitted"
13. System redirects to Request Tickets List -- new ticket visible with Open badge

**Main Flow B -- Manager/Administrator creating on behalf of employee:**
1. Manager clicks "+ Add Request" on Request Tickets List
2. System navigates to "Create A New Request Ticket" page
3. Employee section is **visible** -- searchable dropdown
4. Manager searches and selects an employee
5. System displays employee info card: Full Name, Email, Phone, Department, Position
6. Manager fills Request Info (same as steps 5-8 in Flow A)
7. Manager clicks Save
8. System validates all fields
9. System creates the request ticket with "Open" status
10. Success toast: "Request ticket for '[employee name]' has been submitted"
11. System redirects to Request Tickets List

**Alternative Flows:**

| Action | Flow |
|--------|------|
| **Alt 1 -- Empty mandatory field** | Inline error "[Field name] is required". Form not submitted. |
| **Alt 2 -- Invalid attachment file type** | Inline error below Attachment: "Only PDF, PNG, JPG, and DOCX files are accepted". File not uploaded. |
| **Alt 3 -- Attachment too large** | Inline error below Attachment: "File size must not exceed 5MB". File not uploaded. |
| **Alt 4 -- Cancel (form modified)** | "Discard unsaved changes?" dialog. Confirm -> redirect to list. Cancel -> stay on form. |
| **Alt 5 -- Cancel (form untouched)** | Redirect to Request Tickets List immediately. |
| **Alt 6 -- Employee not found (manager flow)** | Dropdown search returns no match -> "No employees found" in dropdown. |
| **Alt 7 -- Server error on save** | Error toast: "Something went wrong. Please try again." Form stays open, fields preserved. |

**Exit Points:**
- **Success:** Ticket created (Open) -> toast -> redirect to Request Tickets List
- **Cancel:** Redirect to list (with or without confirmation)
- **Error:** Validation errors inline; user corrects and retries

---

## 3. Field Definitions

### Input Fields -- Employee Section (manager/admin only)

| Field Name | Field Type | Validation Rule | Mandatory | Default | Visible To | Description |
|------------|------------|-----------------|-----------|---------|-----------|-------------|
| Employee | Searchable dropdown (576px) | Must select a valid employee | Yes (*) | Empty -- "Select an employee" | Manager/Admin only (hidden for employees) | Search by name; on selection shows info card |

### Employee Info Card (read-only, appears after employee selection)

| Field | Left Column | Right Column |
|-------|------------|-------------|
| Row 1 | Full name | Department |
| Row 2 | Email | Position |
| Row 3 | Phone number | -- |

**Note:** Unlike the Leave Request info card, the Request Ticket info card does not include "Leave Days Remaining" as it is not relevant to request tickets. The card shows 5 fields in a 2-column layout (3 left, 2 right).

### Input Fields -- Request Info Section

| Field Name | Field Type | Validation Rule | Mandatory | Default | Description |
|------------|------------|-----------------|-----------|---------|-------------|
| Request Type | Dropdown (576px) | Must select a value | Yes (*) | "Select request type" | IT Support, Facility, HR Inquiry, Office Supplies, Access Request, Travel & Expense, Training, Other |
| Subject | Text input (576px) | Not empty; trimmed; max 200 characters | Yes (*) | Empty -- "Enter request subject" | Short title for the request (displayed as "Request" column in list) |
| Description | Text area (576px) | Not empty; trimmed | Yes (*) | Empty -- "Enter request description" | Detailed explanation of the request |
| Attachment | File input (576px) | PDF, PNG, JPG, DOCX; max 5MB | No | "Choose File / No file chosen" | Supporting document |

### Interaction Elements

| Element | Type | Position | State/Condition | Trigger Action | Description |
|---------|------|----------|-----------------|----------------|-------------|
| Cancel | Button (secondary) | Header action bar | Always visible | If dirty -> discard dialog; if clean -> redirect | Returns to list |
| Save (header) | Button (primary) | Header action bar | Disabled + spinner while saving | Validates -> creates Open ticket -> toast -> redirect | Submit form |
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

---

## 4. Data Display

### Information Shown to User

**Employee Section (manager/admin only):**

| Data Name | Format | Empty State | Business Meaning |
|-----------|--------|-------------|-----------------|
| Employee dropdown | Searchable dropdown (576px) | "Select an employee" | Target employee for the ticket |
| Info hint | Icon + text | "Select an employee to view their info" | Guidance before selection |
| Employee info card | 5 fields in 2-column layout | Hidden until selected | Read-only summary of selected employee |

**Request Info Section:**

| Data Name | Format | Empty State | Business Meaning |
|-----------|--------|-------------|-----------------|
| Request Type | Dropdown (576px) | "Select request type" | Category of the request |
| Subject | Text input (576px) | "Enter request subject" | Short title displayed in list |
| Description | Text area (576px) | "Enter request description" | Full details of the request |
| Attachment | File input (576px) | "Choose File / No file chosen" | Supporting document |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default (employee) | Employee creating own ticket | Employee section hidden; Request Info empty; Save enabled |
| Default (manager) | Manager creating on behalf | Employee section visible with dropdown + hint; Request Info empty |
| Employee selected | Manager selects employee | Info card appears (5 fields); hint replaced by card |
| Employee not found | Manager searches, no match | "No employees found" in dropdown |
| Validation error | Save with invalid/empty fields | Inline red error(s) below affected fields |
| Saving | Save clicked, in progress | Both Save buttons spinner + disabled; all fields disabled |
| Success | Ticket created | Toast -> redirect to Request Tickets List |
| Discard confirmation | Cancel with modified form | Modal: "Discard unsaved changes?" |
| File selected | User uploads attachment | Filename displayed next to input |

### Page Layout (Design Reference)

```
+---------------------------------------------------------------+
|  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]  |
+--------------+------------------------------------------------+
|  [Sidebar]   |  Create A New Request Ticket  [Cancel] [Save]  |
|  200px       |                                                 |
|              |     +------------------------------------+      |
|              |     | Employee                           |      |
|              |     | * Employee [searchable dropdown]   |      |
|              |     |                                    |      |
|              |     | +-------------+-----------------+  |      |
|              |     | | Full name   | Department      |  |      |
|              |     | | Email       | Position        |  |      |
|              |     | | Phone       |                 |  |      |
|              |     | +-------------+-----------------+  |      |
|              |     +------------------------------------+      |
|              |     | Request Info                        |      |
|              |     | *Request Type                       |      |
|              |     | *Subject                            |      |
|              |     | *Description                        |      |
|              |     |  Attachment [Choose File]            |      |
|              |     +------------------------------------+      |
|              |     | [             Save               ]  |      |
|              |     +------------------------------------+      |
+--------------+------------------------------------------------+
```

---

## 5. Acceptance Criteria

**Definition of Done -- All criteria must be met:**

**Page Display:**
- **AC-01:** Page title displays "Create A New Request Ticket" in Geist Semibold 24px
- **AC-02:** Form has two sections: Employee (card 1) + Request Info (card 2) -- centered 600px card
- **AC-03:** Cancel + Save buttons in header action bar; Save button also at bottom (full-width 600px)
- **AC-04:** Both Save buttons perform identical submission logic

**Employee Section -- Employee Flow:**
- **AC-05:** When an employee creates their own ticket, the Employee section is completely hidden
- **AC-06:** System auto-assigns the ticket to the logged-in employee

**Employee Section -- Manager/Admin Flow:**
- **AC-07:** When manager/admin creates a ticket, the Employee section is visible with a searchable dropdown
- **AC-08:** Employee dropdown searches by employee name with in-dropdown search
- **AC-09:** Before employee selection, hint text displays: "Select an employee to view their info"
- **AC-10:** After selecting an employee, info card appears with 5 read-only fields in 2-column layout
- **AC-11:** Info card displays: Full Name, Email, Phone (left) | Department, Position (right)
- **AC-12:** If employee search returns no match, "No employees found" displayed in dropdown

**Request Info Fields:**
- **AC-13:** Request Type displayed as a full-width (576px) dropdown
- **AC-14:** Request Type options: IT Support, Facility, HR Inquiry, Office Supplies, Access Request, Travel & Expense, Training, Other
- **AC-15:** Subject displayed as a full-width (576px) text input with max 200 characters
- **AC-16:** Description displayed as a full-width (576px) mandatory text area
- **AC-17:** Attachment is a full-width (576px) optional file input accepting PDF, PNG, JPG, DOCX (max 5MB)

**Mandatory Field Validation:**
- **AC-18:** All mandatory fields must be filled: Employee (manager flow only), Request Type, Subject, Description
- **AC-19:** Empty mandatory fields show inline error: "[Field name] is required"
- **AC-20:** Subject and Description trimmed of whitespace; whitespace-only rejected as empty
- **AC-21:** Subject exceeding 200 characters shows inline error: "Subject must not exceed 200 characters"

**Attachment:**
- **AC-22:** Attachment is optional -- ticket can be submitted without a file
- **AC-23:** After selecting a valid file, filename displayed next to input
- **AC-24:** Invalid file type shows: "Only PDF, PNG, JPG, and DOCX files are accepted"
- **AC-25:** File exceeding 5MB shows: "File size must not exceed 5MB"

**Save Behavior:**
- **AC-26:** Save creates the request ticket with "Open" status
- **AC-27:** Employee flow toast: "Request ticket has been submitted"
- **AC-28:** Manager flow toast: "Request ticket for '[employee name]' has been submitted"
- **AC-29:** After save, redirects to Request Tickets List -- new ticket visible with Open badge
- **AC-30:** Save buttons show spinner + disabled while processing; all fields disabled

**Cancel Behavior:**
- **AC-31:** Cancel on untouched form redirects without confirmation
- **AC-32:** Cancel on modified form shows "Discard unsaved changes?" dialog
- **AC-33:** "Form modified" includes any change to employee, request type, subject, description, or file

**Access Control:**
- **AC-34:** Page accessible to all authorized users with request ticket access permission (including employees)
- **AC-35:** Direct URL access by unauthorized users redirects to fallback page

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Employee own ticket | Fill Request Info, Save | Ticket created (Open), toast, redirect | High |
| Manager on behalf | Select employee, fill Request Info, Save | Ticket with employee name, toast, redirect | High |
| Employee section hidden | Employee logs in, + Add Request | No Employee section | High |
| Employee info card | Manager selects "Henry Tran" | Card: name, email, phone, dept, position | High |
| No employee match | Manager types "zzz" | "No employees found" | Medium |
| Empty subject | Leave Subject blank, Save | "Subject is required" | High |
| Whitespace-only subject | Enter "   ", Save | "Subject is required" | High |
| Subject max length | Enter 201 characters | "Subject must not exceed 200 characters" | Medium |
| Empty description | Leave Description blank, Save | "Description is required" | High |
| Empty request type | Leave Request Type unselected, Save | "Request Type is required" | High |
| Valid attachment | Upload 2MB PDF | Filename shown | Medium |
| Invalid attachment | Upload .exe | Error: accepted formats | Medium |
| Large attachment | Upload 6MB file | Error: max 5MB | Medium |
| Cancel dirty | Fill fields, Cancel | Discard dialog | Medium |
| Cancel clean | No changes, Cancel | Redirect immediately | Medium |
| Server error | Server fails during save | Error toast; form preserved | Medium |
| Unauthorized | No permission | Redirect to fallback | High |
| Request Type "Other" | Select "Other" | Ticket created with type "Other" | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** All authorized users with request ticket access permission can access the Create Request Ticket page -- this includes employees (not restricted to management permission only)
- **SR-02:** Permissions configured via US-004. No role names hardcoded.
- **SR-03:** When an **employee** creates a ticket, the Employee section is hidden. System auto-assigns to the logged-in user.
- **SR-04:** When a **manager/administrator** creates a ticket, the Employee section is visible with searchable dropdown.
- **SR-05:** Employee info card is read-only -- populated from employee profile (US-005). Shows 5 fields: Full Name, Email, Phone, Department, Position.
- **SR-06:** All newly created request tickets are assigned **"Open" status**.
- **SR-07:** Request Date is auto-recorded by the system at the time of submission -- not user-editable.
- **SR-08:** Request Type must be one of: IT Support, Facility, HR Inquiry, Office Supplies, Access Request, Travel & Expense, Training, Other.
- **SR-09:** Subject field trimmed of whitespace before saving. Whitespace-only treated as empty. Max 200 characters enforced.
- **SR-10:** Description field trimmed of whitespace before saving. Whitespace-only treated as empty.
- **SR-11:** Attachment: PDF, PNG, JPG, DOCX accepted. Max 5MB. Validated client-side on file selection.
- **SR-12:** Both Save buttons (header + bottom) trigger identical submission logic.
- **SR-13:** "Form dirty" detection covers all fields. Any change triggers discard confirmation on Cancel.
- **SR-14:** After creation, ticket appears in Request Tickets List with Open badge. Visible per data visibility rules (employee sees own; manager/admin sees all).
- **SR-15:** The employee dropdown in the manager flow loads active employees only -- inactive/deactivated accounts are excluded.
- **SR-16:** No duplicate ticket check -- employees can submit multiple tickets with the same subject and type.

**State Transitions:**
```
[Request Tickets List] -> "+ Add Request" -> [Create Form]
[Create Form: Employee flow] -> [Employee section hidden; auto-assigned]
[Create Form: Manager flow] -> [Select employee] -> [Info card with profile]
[Create Form] -> [Save, valid] -> [Open] -> [Toast] -> [List]
[Create Form] -> [Save, invalid] -> [Inline errors] -> [Stay on form]
[Create Form] -> [Cancel (clean)] -> [List]
[Create Form] -> [Cancel (dirty)] -> [Discard dialog]
[Create Form] -> [Save, server error] -> [Error toast] -> [Stay on form]
```

**Dependencies:**
- **US-001 (Authentication):** User identity for employee auto-assignment
- **US-004 (Role & Permission Management):** Controls access; determines employee vs manager flow
- **US-005 (User Management):** Employee profile data for info card
- **EP-008 US-001/002 (Department/Position):** Department and position in info card
- **DR-003-001-01 (Request Tickets List):** Entry point; new ticket appears with Open badge

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Employee dropdown is searchable -- manager can type to quickly find employees
- **UX-02:** Employee info card appears immediately on selection -- shows organizational context
- **UX-03:** Info hint guides manager before selection -- disappears after employee chosen
- **UX-04:** Employee section completely hidden for own tickets -- clean, simple form
- **UX-05:** Both Save buttons do the same thing -- save from anywhere on the page
- **UX-06:** Save buttons show spinner while processing -- prevents double submission
- **UX-07:** Request Type dropdown uses standard (non-searchable) dropdown -- 8 options is a manageable list
- **UX-08:** Subject field has visible max length indicator -- prevents overly long titles
- **UX-09:** Success toast auto-dismisses after 5 seconds; includes employee name for manager flow
- **UX-10:** Discard dialog defaults focus on Cancel (stay) -- prevents data loss
- **UX-11:** Form card centered (600px) -- consistent with Create Leave Request, Create User

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>=1280px) | Centered form card 600px |
| Below desktop | Out of scope for this release |

**Accessibility Requirements:**
- [x] Keyboard navigable -- Tab through all fields in logical order
- [x] Screen reader compatible -- labels for all fields; info card announced
- [x] Employee dropdown searchable via keyboard
- [x] Request Type dropdown accessible via keyboard
- [x] Color contrast meets WCAG 2.1 AA
- [x] Focus indicators visible on all interactive elements
- [x] Error messages associated with fields via aria-describedby

**Design References:**
- Pattern reference: Create Leave Request (DR-002-001-02) for dual-role form layout
- Pattern reference: Create User (DR-001-005-02) for centered form card
- Design tokens: primary `#171717`, background `#ffffff`, border `#e5e5e5`, muted foreground `#737373`, font: Geist

---

## 8. Additional Information

### Out of Scope
- Edit Request Ticket -- separate DR (DR-003-001-03)
- Status management -- separate DR (DR-003-001-04)
- Cancel Request Ticket -- separate DR (DR-003-001-05)
- Delete Request Ticket -- separate DR (DR-003-001-06)
- Request ticket templates per type (future enhancement)
- Auto-routing tickets to specific departments/handlers based on type (future enhancement)
- Priority/urgency field (future enhancement)
- Draft/save-as-draft (always submitted as Open)
- Email notification to managers on submission (future enhancement)
- Multiple file attachments (single file only for this release)
- Mobile or tablet layout

### Open Questions (Resolved)

| Question | Answer | Confirmed By |
|----------|--------|-------------|
| Employee field for own tickets | Hidden -- auto-assigned | Knowledge Base (Dual-Role Form Pattern) |
| Both Save buttons | Identical behavior | Knowledge Base (Create Form Pattern) |
| Initial status | Open (not Pending -- differs from Leave Requests) | Knowledge Base (Request Ticket Status 9.3) |
| Request Type values | IT Support, Facility, HR Inquiry, Office Supplies, Access Request, Travel & Expense, Training, Other | Knowledge Base (Entity Relationships 6.3) |
| Who can create tickets | All authorized users including employees | Knowledge Base (Section 1.7 exception) |
| Attachment formats | PDF, PNG, JPG, DOCX; max 5MB | Knowledge Base (Create Form Pattern 2.5) |
| Info card fields | Full Name, Email, Phone, Department, Position (no Leave Days Remaining) | Knowledge Base (Dual-Role Form Pattern 2.8) |
| Subject max length | 200 characters | User confirmed |

### Open Questions (Pending)

- [ ] **Figma design for Create Request Ticket:** No Figma design has been provided for this page. Layout is based on the established Create Form Pattern (DR-002-001-02). -- **Owner:** Design Team -- **Status:** Pending
- [ ] **Request Type dropdown searchable?** With 8 options, standard non-searchable dropdown is recommended. If more types are added in future, may need in-dropdown search. -- **Owner:** Product Owner -- **Status:** Pending
- [ ] **Description max length?** No max length currently defined for the description field. -- **Owner:** Product Owner -- **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-003-001-01: Request Tickets List | Entry point; new ticket appears with Open badge |
| DR-003-001-03: Edit Request Ticket (planned) | Mirrors this form with pre-filled data; Open status only (employee), Open/In Progress (manager) |
| DR-003-001-04: Manage Request Ticket Status (planned) | Processes tickets created here through lifecycle |
| DR-003-001-05: Cancel Request Ticket (planned) | Cancels tickets created here |
| DR-003-001-06: Delete Request Ticket (planned) | Deletes tickets created here |
| US-001: Authentication | User identity for auto-assignment |
| US-004: Role & Permission Management | Access control; employee vs manager flow |
| US-005: User Management | Employee profile for info card |
| EP-008 US-001/002: Department & Position | Info card data |

### Notes
- **Follows Dual-Role Form Pattern** -- same page behaves differently for employees vs managers/administrators, consistent with Create Leave Request (DR-002-001-02).
- **Employee info card has 5 fields** (not 6) -- unlike Leave Request which includes "Leave Days Remaining," the Request Ticket info card omits that field as it is not relevant.
- **Initial status is "Open" (not "Pending")** -- Request Tickets use a 6-status lifecycle starting with Open, unlike Leave Requests which start with Pending.
- **No warnings** (unlike Leave Request which has balance and overlap warnings) -- Request Ticket creation has no cross-record validation warnings.
- **No date fields** -- unlike Leave Request, the Request Ticket form does not have date inputs. The Request Date is auto-recorded on submission.
- **"+ Add Request" visible to employees** -- this is an exception to the standard pattern where only management permission sees the Add button, consistent with DR-003-001-01.
- **Subject field maps to "Request" column** in the list view -- the subject entered here is what gets displayed (and potentially truncated) in the Request column of DR-003-001-01.

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
| 1.0 | 2026-04-01 | BA Agent | Initial draft -- Dual-role form pattern; 4 mandatory fields + 1 optional; Open initial status; follows Create Leave Request pattern adapted for request tickets |
