---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-002
story_id: US-001
story_name: "Leave Requests"
detail_id: DR-002-001-02
detail_name: "Create Leave Request"
parent_requirement: FR-US-001-09
status: draft
version: "1.0"
created_date: 2026-03-27
last_updated: 2026-03-27
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../FLOWCHART.md"
    relationship: sibling
  - path: "./DR-002-001-01-leave-requests-list.md"
    relationship: sibling
---

# Detail Requirement: Create Leave Request

**Detail ID:** DR-002-001-02
**Parent Requirement:** FR-US-001-09
**Story:** US-001-leave-requests
**Epic:** EP-002 (Leave Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As a **user with leave management permission**, I want to **submit a new leave request by selecting an employee, specifying dates, leave type, period, and reason**, so that **the request is formally recorded and routed for approval**.

**Purpose:** Allow authorized users to create leave requests. Employees create their own requests (employee field hidden, auto-assigned); admin/managers can create on behalf of any employee by searching and selecting from the employee dropdown. The form displays the selected employee's profile summary and remaining leave balance to help make informed requests.

**Target Users:**
- **Employees** — create their own requests (Employee section hidden, auto-assigned to logged-in user)
- **Admin/Managers** — create requests on behalf of any employee (Employee section visible with searchable dropdown)

**Key Functionality:**
- Dual-role form: simplified for employees, full for admin/managers
- Two-section layout: Employee (with auto-populated info card) + Leave Info
- Employee info card shows profile + Leave Days Remaining on selection
- 6 mandatory fields + 1 optional (Attachment)
- Leave balance warning (non-blocking) and overlapping dates warning (non-blocking)
- All new requests start with "Pending" status
- Past dates allowed for historical record-keeping

---

## 2. User Workflow

**Entry Point:** Leave Requests List → click "+ Add New" button

**Preconditions:**
- User is signed in (US-001)
- User has leave management permission (US-004)

**Main Flow A — Employee creating own request:**
1. Employee clicks "+ Add New" on Leave Requests List
2. System navigates to "Create A New Leave Request" page
3. Employee section is **hidden** — system auto-assigns to logged-in user
4. Leave Info section displays with empty fields
5. Employee enters From Date and To Date using date pickers
6. Employee selects Leave Period (Full Day / Morning Half / Afternoon Half)
7. Employee selects Leave Type (Annual / Sick / Personal / Maternity / Unpaid)
8. Employee enters Reason (mandatory, text area)
9. Employee optionally uploads an Attachment
10. Employee clicks Save
11. System validates all mandatory fields and date logic
12. System checks leave balance and overlapping dates — shows warnings if applicable (non-blocking)
13. System creates the leave request with "Pending" status
14. Success toast: "Leave request has been submitted"
15. System redirects to Leave Requests List — new request visible with Pending badge

**Main Flow B — Admin/Manager creating on behalf of employee:**
1. Admin clicks "+ Add New" on Leave Requests List
2. System navigates to "Create A New Leave Request" page
3. Employee section is **visible** — searchable dropdown
4. Admin searches and selects an employee
5. System displays employee info card: Full name, Email, Phone, Department, Position, **Leave Days Remaining**
6. Admin fills Leave Info (same as steps 5–9 in Flow A)
7. Admin clicks Save
8. System validates all fields + checks leave balance and overlapping dates (warnings only)
9. System creates the leave request with "Pending" status
10. Success toast: "Leave request for '[employee name]' has been submitted"
11. System redirects to Leave Requests List

**Alternative Flows:**
- **Alt 1 — Empty mandatory field:** Inline error "[Field name] is required". Form not submitted.
- **Alt 2 — To Date before From Date:** Inline error: "To Date must be on or after From Date". Form not submitted.
- **Alt 3 — Insufficient leave balance (warning):** Yellow inline warning + warning toast: "Insufficient leave balance. [X] days remaining, [Y] days requested." Form CAN still be submitted.
- **Alt 4 — Overlapping dates (warning):** Yellow inline warning + warning toast: "This employee already has a leave request for overlapping dates." Form CAN still be submitted.
- **Alt 5 — Invalid attachment file:** Inline error below Attachment: "Only PDF, PNG, JPG, and DOCX files are accepted" or "File size must not exceed 5MB". File not uploaded.
- **Alt 6 — Cancel (form modified):** "Discard unsaved changes?" dialog. Confirm → redirect to list. Cancel → stay on form.
- **Alt 7 — Cancel (form untouched):** Redirect to Leave Requests List immediately.
- **Alt 8 — Employee not found (admin flow):** Dropdown search returns no match → "No employees found" in dropdown.

**Exit Points:**
- **Success:** Request created (Pending) → toast → redirect to Leave Requests List
- **Cancel:** Redirect to list (with or without confirmation)
- **Error:** Validation errors inline; user corrects and retries

---

## 3. Field Definitions

### Input Fields — Employee Section (admin/manager only)

| Field Name | Field Type | Validation Rule | Mandatory | Default | Visible To | Description |
|------------|------------|-----------------|-----------|---------|-----------|-------------|
| Employee | Searchable dropdown (576px) | Must select a valid employee | Yes (*) | Empty — "Select an employee" | Admin/Manager only (hidden for employees) | Search by name; on selection shows info card |

### Employee Info Card (read-only, appears after employee selection)

| Field | Left Column | Right Column |
|-------|------------|-------------|
| Row 1 | Full name | Department |
| Row 2 | Email | Position |
| Row 3 | Phone number | Leave Days Remaining (decimal) |

### Input Fields — Leave Info Section

| Field Name | Field Type | Validation Rule | Mandatory | Default | Description |
|------------|------------|-----------------|-----------|---------|-------------|
| From Date | Date picker (278px) | Not empty; must be ≤ To Date; past dates allowed | Yes (*) | Empty — "Enter start of term" | Leave start date. Calendar icon. |
| To Date | Date picker (278px) | Not empty; must be ≥ From Date | Yes (*) | Empty — "Enter end of term" | Leave end date. Calendar icon. |
| Leave Period | Dropdown (278px) | Must select a value | Yes (*) | "Select period" | Full Day, Morning Half, Afternoon Half |
| Leave Type | Dropdown (278px) | Must select a value | Yes (*) | "Select leave type" | Annual, Sick, Personal, Maternity, Unpaid |
| Reason | Text area (576px) | Not empty; trimmed | Yes (*) | Empty — "Enter reason" | Explanation for the leave request |
| Attachment | File input (576px) | PDF, PNG, JPG, DOCX; max 5MB | No | "Choose File / No file chosen" | Supporting document |

### Interaction Elements

| Element | Type | Position | State/Condition | Trigger Action | Description |
|---------|------|----------|-----------------|----------------|-------------|
| Cancel | Button (secondary) | Header action bar | Always visible | If dirty → discard dialog; if clean → redirect | Returns to list |
| Save (header) | Button (primary) | Header action bar | Disabled + spinner while saving | Validates → creates Pending request → toast → redirect | Submit form |
| Save (bottom) | Button (primary, full-width) | Below form, 600×40px | Same behavior as header Save | Identical to header Save | Convenience duplicate |

### Validation Error Messages

| Condition | Error Message | Display | Blocks Submit? |
|-----------|--------------|---------|---------------|
| Mandatory field empty | "[Field name] is required" | Inline, below field | Yes |
| To Date before From Date | "To Date must be on or after From Date" | Inline, below To Date | Yes |
| Insufficient leave balance | "Insufficient leave balance. [X] days remaining, [Y] days requested." | Yellow inline + warning toast | **No** — warning only |
| Overlapping dates | "This employee already has a leave request for overlapping dates." | Yellow inline + warning toast | **No** — warning only |
| Invalid attachment type | "Only PDF, PNG, JPG, and DOCX files are accepted" | Inline, below Attachment | Yes (file rejected) |
| Attachment too large | "File size must not exceed 5MB" | Inline, below Attachment | Yes (file rejected) |

---

## 4. Data Display

### Information Shown to User

**Employee Section (admin/manager only):**

| Data Name | Format | Empty State | Business Meaning |
|-----------|--------|-------------|-----------------|
| Employee dropdown | Searchable dropdown (576px) | "Select an employee" | Target employee |
| Info hint | Icon + text | "Select an employee to view their info" | Guidance before selection |
| Employee info card | 6 fields in 2×3 grid | Hidden until selected | Read-only summary + leave balance |

**Leave Info Section:**

| Data Name | Format | Empty State | Business Meaning |
|-----------|--------|-------------|-----------------|
| From Date | Date picker (278px) | "Enter start of term" | Leave start |
| To Date | Date picker (278px) | "Enter end of term" | Leave end |
| Leave Period | Dropdown (278px) | "Select period" | Full Day / Morning Half / Afternoon Half |
| Leave Type | Dropdown (278px) | "Select leave type" | Annual / Sick / Personal / Maternity / Unpaid |
| Reason | Text area (576px) | "Enter reason" | Justification |
| Attachment | File input (576px) | "Choose File / No file chosen" | Supporting document |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Default (employee) | Employee creating own request | Employee section hidden; Leave Info empty; Save enabled |
| Default (admin) | Admin creating on behalf | Employee section visible with dropdown + hint; Leave Info empty |
| Employee selected | Admin selects employee | Info card appears (6 fields); hint replaced by card |
| Employee not found | Admin searches, no match | "No employees found" in dropdown |
| Leave balance warning | Requested days > remaining | Yellow inline warning + warning toast — still submittable |
| Overlapping dates warning | Dates overlap with existing request | Yellow inline warning + warning toast — still submittable |
| Validation error | Save with invalid/empty fields | Inline red error(s) below affected fields |
| Saving | Save clicked, in progress | Both Save buttons spinner + disabled; all fields disabled |
| Success | Request created | Toast → redirect to Leave Requests List |
| Discard confirmation | Cancel with modified form | Modal: "Discard unsaved changes?" |
| File selected | User uploads attachment | Filename displayed next to input |

### Page Layout (Design Reference)

```
┌─────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb               [Top Bar]   │
├──────────────┬──────────────────────────────────────────────────┤
│  [Sidebar]   │  Create A New Leave Request   [Cancel] [Save]   │
│  200px       │                                                  │
│              │     ┌────────────────────────────────────┐       │
│              │     │ Employee                           │       │
│              │     │ * Employee [searchable dropdown]   │       │
│              │     │                                    │       │
│              │     │ ┌─────────────┬─────────────────┐  │       │
│              │     │ │ Full name   │ Department      │  │       │
│              │     │ │ Email       │ Position        │  │       │
│              │     │ │ Phone       │ Leave Remaining │  │       │
│              │     │ └─────────────┴─────────────────┘  │       │
│              │     ├────────────────────────────────────┤       │
│              │     │ Leave Info                         │       │
│              │     │ *From Date      *To Date           │       │
│              │     │ *Leave Period   *Leave Type        │       │
│              │     │ *Reason                            │       │
│              │     │  Attachment [Choose File]          │       │
│              │     ├────────────────────────────────────┤       │
│              │     │ [            Save              ]   │       │
│              │     └────────────────────────────────────┘       │
└──────────────┴──────────────────────────────────────────────────┘
```

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Page Display:**
- **AC-01:** Page title displays "Create A New Leave Request" in Geist Semibold 24px
- **AC-02:** Form has two sections: Employee (card 1) + Leave Info (card 2) — centered 600px card
- **AC-03:** Cancel + Save buttons in header action bar; Save button also at bottom (full-width 600px)
- **AC-04:** Both Save buttons perform identical submission logic

**Employee Section — Employee Flow:**
- **AC-05:** When an employee creates their own request, the Employee section is completely hidden
- **AC-06:** System auto-assigns the request to the logged-in employee

**Employee Section — Admin/Manager Flow:**
- **AC-07:** When admin/manager creates a request, the Employee section is visible with a searchable dropdown
- **AC-08:** Employee dropdown searches by employee name with in-dropdown search
- **AC-09:** Before employee selection, hint text displays: "Select an employee to view their info"
- **AC-10:** After selecting an employee, info card appears with 6 read-only fields in 2-column layout
- **AC-11:** Info card displays: Full name, Email, Phone (left) | Department, Position, Leave Days Remaining (right)
- **AC-12:** Leave Days Remaining shows current balance as decimal (e.g., "8.5")
- **AC-13:** If employee search returns no match, "No employees found" displayed in dropdown

**Leave Info Fields:**
- **AC-14:** From Date and To Date displayed side-by-side (278px each) with calendar date pickers
- **AC-15:** Leave Period and Leave Type displayed side-by-side (278px each) as dropdowns
- **AC-16:** Leave Period options: Full Day, Morning Half, Afternoon Half
- **AC-17:** Leave Type options: Annual, Sick, Personal, Maternity, Unpaid
- **AC-18:** Reason is a full-width (576px) mandatory text area
- **AC-19:** Attachment is a full-width (576px) optional file input accepting PDF, PNG, JPG, DOCX (max 5MB)

**Date Validation:**
- **AC-20:** Past dates are allowed for historical record-keeping
- **AC-21:** To Date must be on or after From Date — inline error if violated
- **AC-22:** Total Days calculated automatically: inclusive counting with half-day support

**Leave Balance Warning:**
- **AC-23:** If requested days exceed Leave Days Remaining, yellow inline warning + warning toast shown
- **AC-24:** Warning message: "Insufficient leave balance. [X] days remaining, [Y] days requested."
- **AC-25:** Leave balance warning does NOT block submission — form can still be saved

**Overlapping Dates Warning:**
- **AC-26:** If employee already has a leave request for overlapping dates, yellow inline warning + warning toast shown
- **AC-27:** Warning message: "This employee already has a leave request for overlapping dates."
- **AC-28:** Overlapping dates warning does NOT block submission — form can still be saved

**Mandatory Field Validation:**
- **AC-29:** All mandatory fields must be filled: Employee (admin only), From Date, To Date, Leave Period, Leave Type, Reason
- **AC-30:** Empty mandatory fields show inline error: "[Field name] is required"
- **AC-31:** Reason trimmed of whitespace; whitespace-only rejected as empty

**Attachment:**
- **AC-32:** Attachment is optional — request can be submitted without a file
- **AC-33:** After selecting a valid file, filename displayed next to input
- **AC-34:** Invalid file type shows: "Only PDF, PNG, JPG, and DOCX files are accepted"
- **AC-35:** File exceeding 5MB shows: "File size must not exceed 5MB"

**Save Behavior:**
- **AC-36:** Save creates the leave request with "Pending" status
- **AC-37:** Employee flow toast: "Leave request has been submitted"
- **AC-38:** Admin flow toast: "Leave request for '[employee name]' has been submitted"
- **AC-39:** After save, redirects to Leave Requests List — new request visible with Pending badge
- **AC-40:** Save buttons show spinner + disabled while processing; all fields disabled

**Cancel Behavior:**
- **AC-41:** Cancel on untouched form redirects without confirmation
- **AC-42:** Cancel on modified form shows "Discard unsaved changes?" dialog
- **AC-43:** "Form modified" includes any change to employee, dates, dropdowns, reason, or file

**Access Control:**
- **AC-44:** Page accessible only to users with leave management permission
- **AC-45:** Direct URL access by unauthorized users redirects to fallback

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Employee own request | Fill Leave Info, Save | Request created (Pending), toast, redirect | High |
| Admin on behalf | Select employee, fill Leave Info, Save | Request with employee name, toast, redirect | High |
| Employee section hidden | Employee logs in, + Add New | No Employee section | High |
| Employee info card | Admin selects "Henry Tran" | Card: name, email, phone, dept, position, 8.5 days | High |
| No employee match | Admin types "zzz" | "No employees found" | Medium |
| To Date before From | From: Mar 30, To: Mar 25 | Inline error | High |
| Past dates OK | From: Mar 1 (past) | Submitted successfully | High |
| Balance warning | 10 days requested, 8.5 remaining | Yellow warning; still submittable | High |
| Balance override | Warning shown, Save anyway | Request created | High |
| Overlapping dates | Employee has leave Mar 25-27, request Mar 26-28 | Yellow warning; still submittable | High |
| Overlap override | Warning shown, Save anyway | Request created | High |
| Half-day calc | From=To=Mar 25, Morning Half | Total: 0.5 days | Medium |
| Valid attachment | Upload 2MB PDF | Filename shown | Medium |
| Invalid attachment | Upload .exe | Error: accepted formats | Medium |
| Empty reason | Leave Reason blank, Save | "Reason is required" | High |
| Cancel dirty | Fill fields, Cancel | Discard dialog | Medium |
| Unauthorized | No permission | Redirect | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Only users with leave management permission can access the Create Leave Request page
- **SR-02:** Permissions configured via US-004. No role names hardcoded.
- **SR-03:** When an **employee** creates a request, the Employee section is hidden. System auto-assigns to logged-in user.
- **SR-04:** When an **admin/manager** creates a request, the Employee section is visible with searchable dropdown.
- **SR-05:** Employee info card is read-only — populated from employee profile (US-005) and leave balance system.
- **SR-06:** **Leave Days Remaining** = Total Yearly Leave Allowance (configurable per user) − Sum of Approved Leave Days for the current year. Fetched real-time on employee selection.
- **SR-07:** All newly created leave requests are assigned **"Pending" status**.
- **SR-08:** **Past dates are allowed** for historical record-keeping. No "date in the past" validation.
- **SR-09:** **To Date must be ≥ From Date.** Requests with To Date before From Date are rejected.
- **SR-10:** **Total Days calculation:** Inclusive counting (From to To including both). Leave Period determines count per day: Full Day = 1.0, Morning Half = 0.5, Afternoon Half = 0.5.
- **SR-11:** **Leave balance check is a warning only.** If requested days > Leave Days Remaining → yellow warning shown. Submission is NOT blocked. This allows advance leave and special approvals.
- **SR-12:** **Overlapping dates check is a warning only.** If employee has an existing leave request (any status except Cancelled) for overlapping dates → yellow warning shown. Submission is NOT blocked.
- **SR-13:** Reason field trimmed of whitespace before saving. Whitespace-only treated as empty.
- **SR-14:** Attachment: PDF, PNG, JPG, DOCX accepted. Max 5MB. Validated client-side on file selection.
- **SR-15:** Both Save buttons (header + bottom) trigger identical submission logic.
- **SR-16:** "Form dirty" detection covers all fields. Any change triggers discard confirmation on Cancel.
- **SR-17:** After creation, request appears in Leave Requests List with Pending badge. Visible to any user with leave approval permission.
- **SR-18:** Overlapping dates check compares against all existing requests for the same employee where status is Pending or Approved (not Rejected or Cancelled).
- **SR-19:** Any user with leave approval permission can approve the submitted request — no specific manager routing.

**State Transitions:**
```
[Leave Requests List] → "+ Add New" → [Create Form]
[Create Form: Employee flow] → [Employee section hidden; auto-assigned]
[Create Form: Admin flow] → [Select employee] → [Info card + Leave Days Remaining]
[Create Form] → [Save, valid, no warnings] → [Pending] → [Toast] → [List]
[Create Form] → [Save, valid, balance warning] → [Yellow warning] → [User saves again] → [Pending] → [Toast] → [List]
[Create Form] → [Save, valid, overlap warning] → [Yellow warning] → [User saves again] → [Pending] → [Toast] → [List]
[Create Form] → [Save, invalid] → [Inline errors] → [Stay on form]
[Create Form] → [Cancel (clean)] → [List]
[Create Form] → [Cancel (dirty)] → [Discard dialog]
```

**Dependencies:**
- **US-001 (Authentication):** User identity for employee auto-assignment
- **US-004 (Role & Permission Management):** Controls access; determines employee vs admin flow; approval permission
- **US-005 (User Management):** Employee profile data for info card
- **EP-008 US-001/002 (Department/Position):** Department and position in info card
- **Leave Balance System:** Total Yearly Allowance (configurable per user) and approved days calculation
- **DR-002-001-01 (Leave Requests List):** Entry point; overlapping dates checked against existing requests

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Employee dropdown is searchable — admin can type to quickly find employees
- **UX-02:** Employee info card appears immediately on selection — shows full context including Leave Days Remaining
- **UX-03:** Leave Days Remaining prominently displayed — helps informed requests
- **UX-04:** Balance and overlap warnings use **yellow/amber styling** — communicates caution without hard error
- **UX-05:** Info hint guides admin before selection — disappears after employee chosen
- **UX-06:** Employee section completely hidden for own requests — clean, simple form
- **UX-07:** Both Save buttons do the same thing — save from anywhere on the page
- **UX-08:** Save buttons show spinner while processing — prevents double submission
- **UX-09:** Date pickers use calendar popups — no manual format needed
- **UX-10:** Success toast auto-dismisses after 5 seconds; includes employee name for admin flow
- **UX-11:** Discard dialog defaults focus on Cancel (stay) — prevents data loss
- **UX-12:** Form card centered (600px) — consistent with Create User, Create Skill

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Centered form card 600px |
| Below desktop | Out of scope |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through all fields in logical order
- [x] Screen reader compatible — labels for all fields; info card announced
- [x] Date pickers accessible via keyboard
- [x] Employee dropdown searchable via keyboard
- [x] Balance/overlap warnings announced (role="alert")
- [x] Color contrast meets WCAG 2.1 AA
- [x] Focus indicators visible

**Design References:**
- Figma empty: [node 3124:3344](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3124-3344)
- Figma filled: [node 3124:4291](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3124-4291)
- **Design gap:** Leave Type placeholder shows "Select department" — should be "Select leave type"

---

## 8. Additional Information

### Out of Scope
- Edit Leave Request — separate DR (DR-002-001-03)
- Approve/Reject — separate DR (DR-002-001-04)
- Cancel Leave Request — separate DR (DR-002-001-05)
- Leave balance configuration (yearly allowance setup — future epic)
- Leave type CRUD (future epic)
- Draft/save-as-draft (always submitted as Pending)
- Recurring leave requests
- Email notification to approver on submission
- Manager-specific approval routing (any approver can approve)
- Mobile or tablet layout

### Open Questions (Resolved)

| Question | Answer | Confirmed By |
|----------|--------|-------------|
| Employee field for own requests | Hidden — auto-assigned | Product Owner |
| Both Save buttons | Identical behavior | Product Owner |
| Past dates | Allowed for historical records | Product Owner |
| Leave balance | Warning only — non-blocking | Product Owner |
| Overlapping dates | Warning only — non-blocking | Product Owner |
| Attachment formats | PDF, PNG, JPG, DOCX; max 5MB | Product Owner |
| Approval routing | Any user with approval permission | Product Owner |
| Leave balance source | Total Yearly Allowance (configurable per user) − Approved days | Product Owner |

### Open Questions (Pending)
- [ ] **Leave Type placeholder bug:** Figma shows "Select department" — should be "Select leave type." — **Owner:** Design Team — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-002-001-01: Leave Requests List | Entry point; new request appears with Pending badge |
| DR-002-001-03: Edit Leave Request (planned) | Mirrors this form with pre-filled data; Pending only |
| DR-002-001-04: Approve/Reject (planned) | Processes requests created here |
| DR-002-001-05: Cancel Leave Request (planned) | Cancels requests created here |
| US-001: Authentication | User identity for auto-assignment |
| US-004: Role & Permission Management | Access control; employee vs admin flow |
| US-005: User Management | Employee profile for info card |
| EP-008 US-001/002: Dept & Position | Info card data |

### Notes
- **First dual-role form** — same page behaves differently for employees vs admin/managers.
- **Employee info card** with Leave Days Remaining is unique — no other create form shows related entity summary.
- **Two non-blocking warnings** (balance + overlap) are intentionally soft — allows legitimate overrides for advance leave, special circumstances.
- **Leave balance formula:** Leave Days Remaining = Total Yearly Allowance − Sum of Approved Leave Days (current year). The yearly allowance is configurable per user.
- **Overlapping check** compares against Pending and Approved requests only — Rejected and Cancelled are excluded.
- **Two Save buttons** (header + bottom) — usability pattern for forms longer than one viewport.

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
| 1.0 | 2026-03-27 | BA Agent | Initial draft — Figma nodes 3124:3344 (empty) + 3124:4291 (filled) |
