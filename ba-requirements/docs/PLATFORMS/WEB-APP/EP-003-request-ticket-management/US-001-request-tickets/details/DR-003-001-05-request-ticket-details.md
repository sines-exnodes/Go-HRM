---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-003
story_id: US-001
story_name: "Request Tickets"
detail_id: DR-003-001-05
detail_name: "Request Ticket Details"
parent_requirement: FR-US-001-11
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
  - path: "./DR-003-001-03-update-request-ticket.md"
    relationship: sibling
  - path: "./DR-003-001-04-delete-request-ticket.md"
    relationship: sibling
---

# Detail Requirement: Request Ticket Details

**Detail ID:** DR-003-001-05
**Parent Requirement:** FR-US-001-11
**Story:** US-001-request-tickets
**Epic:** EP-003 (Request Ticket Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **employee with request ticket access**, I want to **view the full details of one of my submitted request tickets**, so that **I can review its content, track its current status, and take actions available to me as the submitter (Close or Reopen)**.

As a **manager or administrator with request ticket management permission**, I want to **view the full details of any request ticket and perform status transition actions from a dedicated page**, so that **I can review the complete ticket information and advance the ticket through its lifecycle without navigating away from the detail context**.

**Purpose:** Provide a dedicated read-only view of a single request ticket's full information, combined with status management actions accessible directly from the page. This page serves two purposes: (1) displaying all ticket data in a structured, readable format that cannot be fully shown in the list view (e.g., full description, attachment), and (2) offering contextual action buttons so that managers/admins and the submitting employee can take the next appropriate lifecycle step without returning to the list.

**Target Users:**
- **Employees** — can view their own tickets; can perform submitter-exclusive actions (Close on Resolved, Reopen on Resolved) from this page
- **Managers/Administrators** — can view any ticket; can perform all applicable status transition actions (In Progress, On Hold, Resume, Resolve, Cancel) and access Edit/Delete from this page

**Key Functionality:**
- Two-panel layout: left action menu with contextual action buttons + right content panel with full ticket data
- Read-only display of all ticket fields grouped into logical sections
- Status badge prominently displayed alongside ticket subject
- Action buttons in left panel change based on ticket status and user role (same logic as gear icon in list)
- Navigation from Request Tickets List (via gear icon "View Details" or row click) and back
- No inline editing — all modifications go through the Update Request Ticket page (DR-003-001-03)

---

## 2. User Workflow

**Entry Point:** Request Tickets List → click anywhere on the ticket row OR gear icon → "View Details"

**Preconditions:**
- User is signed in (EP-001 US-001)
- User has request ticket view permission (EP-001 US-004)
- For employees: the ticket belongs to the logged-in user
- For managers/admins: any ticket in the system

**Main Flow:**
1. User is on the Request Tickets List page
2. User clicks on a ticket row or selects "View Details" from the gear icon dropdown
3. System navigates to the Request Ticket Details page for the selected ticket
4. System loads and displays all ticket data in the right content panel (skeleton loading while fetching)
5. System renders contextual action buttons in the left action menu based on ticket status and user role
6. User reviews the ticket information (subject, description, type, dates, employee info, status, attachment)
7. User optionally clicks an action button (e.g., "In Progress", "Resolve", "Close", "Edit", "Delete")
8. For status actions: a confirmation dialog appears (see Section 3); user confirms or cancels
9. On confirmation: system performs the action, updates the ticket status, and refreshes the detail page to reflect the new state
10. User clicks "Back to List" or navigates away when done

**Alternative Flows:**
- **Alt 1 — Back navigation:** User clicks "Back to List" → returns to Request Tickets List with previously active search/filters preserved
- **Alt 2 — Edit action:** User clicks "Edit" → navigates to Update Request Ticket page (DR-003-001-03)
- **Alt 3 — Delete action:** User clicks "Delete" → opens deletion confirmation dialog (DR-003-001-04 pattern); on confirm, soft-deletes ticket and redirects to list
- **Alt 4 — Status action with confirmation:** All status transitions (e.g., In Progress, On Hold, Resume, Resolve, Close, Reopen, Cancel) trigger a confirmation dialog before proceeding
- **Alt 5 — Status changed by another user:** If another user changes the ticket's status between page load and action attempt, the server rejects the action; an error toast is shown: "This ticket's status has changed. The page will now refresh." The page reloads to show the current state
- **Alt 6 — Direct URL access without permission:** User without view permission navigates directly to the URL → redirected to fallback/unauthorized page
- **Alt 7 — Ticket not found:** Ticket has been deleted or does not exist → system shows an error state: "Request ticket not found." with a "Back to List" link

**Exit Points:**
- **Back to List:** Returns to Request Tickets List (search/filter state preserved)
- **Edit:** Navigates to Update Request Ticket form
- **After Delete:** Redirects to Request Tickets List; success toast shown
- **After status action:** Page refreshes in place to show updated status and revised action buttons

---

## 3. Field Definitions

### Input Fields

No direct data-entry fields — this is a read-only detail view. Status actions and deletion are triggered via action buttons that open confirmation dialogs; the dialogs themselves have no free-text input.

### Interaction Elements

**Left Action Menu:**

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| Back to List | Link/Button | Always visible | Returns to Request Tickets List | Navigation back; preserves list state |
| Edit | Secondary button | Visible when: Employee (own Open ticket) OR Manager/Admin (Open or In Progress ticket) | Navigates to Update Request Ticket page | Redirects to DR-003-001-03 |
| In Progress | Action button | Visible to Manager/Admin when status = Open | Opens status change confirmation dialog | Advances ticket from Open to In Progress |
| On Hold | Action button | Visible to Manager/Admin when status = In Progress | Opens status change confirmation dialog | Pauses work on the ticket |
| Resume | Action button | Visible to Manager/Admin when status = On Hold | Opens status change confirmation dialog | Resumes work; moves back to In Progress |
| Resolve | Action button | Visible to Manager/Admin when status = In Progress | Opens status change confirmation dialog | Marks ticket as resolved pending submitter confirmation |
| Close | Action button | Visible to Employee (own ticket) when status = Resolved | Opens status change confirmation dialog | Confirms resolution; moves to Closed (terminal) |
| Reopen | Action button | Visible to Employee (own ticket) when status = Resolved | Opens status change confirmation dialog | Reopens ticket back to Open (resolution unsatisfactory) |
| Cancel | Danger action button | Visible to Employee (own non-terminal ticket: Open, In Progress, On Hold) and Manager/Admin (any non-terminal ticket) | Opens status change confirmation dialog | Cancels the ticket; moves to Cancelled (terminal) |
| Delete | Danger action button | Visible per DR-003-001-04 delete visibility rules | Opens deletion confirmation dialog (DR-003-001-04 pattern) | Soft-deletes the ticket |

**Status Change Confirmation Dialogs (per action):**

| Action | Dialog Title | Confirmation Message |
|--------|--------------|----------------------|
| In Progress | "Mark as In Progress" | "Mark this request ticket as In Progress? This indicates work has started." |
| On Hold | "Put On Hold" | "Put this request ticket on hold? Work will be paused until resumed." |
| Resume | "Resume Ticket" | "Resume this request ticket? It will be moved back to In Progress." |
| Resolve | "Mark as Resolved" | "Mark this request ticket as Resolved? The submitter will need to confirm the resolution." |
| Close | "Close Ticket" | "Close this request ticket? This confirms the resolution is satisfactory. This action cannot be undone." |
| Reopen | "Reopen Ticket" | "Reopen this request ticket? It will be moved back to Open status." |
| Cancel | "Cancel Ticket" | "Cancel this request ticket? This action cannot be undone." |

---

## 4. Data Display

### Ticket Information Displayed

**Employee Section:**

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Employee Full Name | Text | "—" | Full name string | Identifies who submitted the ticket |
| Department | Text | "—" | Department name | Employee's organizational unit |
| Position | Text | "—" | Position title | Employee's role in the organization |
| Email | Text | "—" | email@domain.com | Employee contact |
| Phone | Text | "—" | Phone number string | Employee contact |

**Request Info Section:**

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Subject | Text | "—" | Full subject text (no truncation on detail page) | Short title identifying the request |
| Status | Badge | N/A (always has a status) | Colored badge per §9.3 knowledge base | Current lifecycle stage of the ticket |
| Request Type | Text | "—" | Type label (e.g., "IT Support", "Facility") | Category of the request |
| Request Date | Date | "—" | DD/MM/YYYY or system date format | Date the ticket was submitted |
| Description | Text (multiline) | "—" | Full text, multi-line, no truncation | Complete request details as entered by the submitter |
| Attachment | File link | "No attachment" | Clickable filename linking to file download | Supporting document provided with the ticket |

**Status Badge Colors (from §9.3):**

| Status | Badge Color | Meaning |
|--------|-------------|---------|
| Open | Blue | New ticket, awaiting action |
| In Progress | Amber | Actively being worked on |
| On Hold | Orange | Work paused temporarily |
| Resolved | Green | Solution provided, awaiting submitter confirmation |
| Closed | Gray | Submitter confirmed resolution; finalized |
| Cancelled | Red | Ticket cancelled |

### Layout Diagram

```
+---------------------------------------------------------------+
|  Breadcrumb / Request Tickets / [Subject]          [Top Bar]  |
+--------------+------------------------------------------------+
|  [Sidebar]   |  Request Ticket Details                        |
|  200px       |                                                |
|              +------------------+-----------------------------+
|              | Left Panel       | Right Content Panel         |
|              | (Action Menu)    |                             |
|              |                  |  +-------------------------+|
|              | [Back to List]   |  | Employee                ||
|              |                  |  | Full Name               ||
|              | [Edit]           |  | Department | Position   ||
|              |                  |  | Email | Phone           ||
|              | --- Status ---   |  +-------------------------+|
|              | [In Progress]    |  | Request Info            ||
|              | [On Hold]        |  | Subject   [Status Badge]||
|              | [Resume]         |  | Request Type            ||
|              | [Resolve]        |  | Request Date            ||
|              | [Close]          |  | Description (full text) ||
|              | [Reopen]         |  | Attachment [link]       ||
|              |                  |  +-------------------------+|
|              | --- Danger ---   |                             |
|              | [Cancel]         |                             |
|              | [Delete]         |                             |
+--------------+------------------+-----------------------------+
```

**Notes on layout:**
- Left panel action buttons render only the actions applicable to the current user role and ticket status
- Action buttons are grouped: navigation (Back, Edit), status actions, danger actions (Cancel, Delete) — visually separated
- Right panel uses the icon + label + value format consistent with User Details (§3.2 knowledge base)
- The page title in the header is "Request Ticket Details"

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page first loads; fetching ticket data | Skeleton content blocks in right panel; action buttons render after data loads |
| Loaded | Ticket data fetched successfully | Full ticket information with applicable action buttons |
| Status action processing | User confirmed a status change; processing | Confirmed action button shows loading spinner; all action buttons disabled |
| Status updated | Status change succeeded | Page refreshes in place; status badge updates; action buttons re-render for new status; success toast shown |
| Delete processing | User confirmed deletion; processing | Handled by deletion dialog per DR-003-001-04 |
| Not found | Ticket deleted or does not exist | Error message: "Request ticket not found." with "Back to List" link |
| Unauthorized | User without view permission accesses URL | Redirect to system fallback/unauthorized page |
| Error — status conflict | Another user changed status before this action | Error toast: "This ticket's status has changed. The page will now refresh." Page reloads |

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Page Access & Loading:**
- **AC-01:** Clicking a ticket row on the Request Tickets List navigates the user to the Request Ticket Details page for that ticket
- **AC-02:** Selecting "View Details" from the gear icon on the list also navigates to the Request Ticket Details page
- **AC-03:** The detail page displays skeleton loading content while ticket data is being fetched
- **AC-04:** Employees can only access the detail page for their own tickets; accessing another employee's ticket URL redirects to the unauthorized page
- **AC-05:** Admin/managers can access the detail page for any ticket in the system

**Data Display:**
- **AC-06:** All ticket fields are displayed in read-only format: Subject, Status badge, Request Type, Request Date, Description, and Attachment
- **AC-07:** Employee section displays: Full Name, Department, Position, Email, and Phone of the submitter
- **AC-08:** The Subject field is displayed in full (no truncation) on the detail page
- **AC-09:** The Description field is displayed in full multi-line format (no truncation)
- **AC-10:** When no attachment is present, the attachment area displays "No attachment"
- **AC-11:** When an attachment is present, the attachment displays as a clickable filename that downloads or opens the file
- **AC-12:** Fields with no value display "—" (em dash) as a placeholder
- **AC-13:** The status badge uses the correct color for each of the 6 statuses per the established badge color scheme

**Action Buttons — Visibility:**
- **AC-14:** The Edit button is visible when: the user is an employee viewing their own Open ticket, OR the user is a manager/admin and the ticket is in Open or In Progress status
- **AC-15:** "In Progress" button is visible to managers/admins only when the ticket status is Open
- **AC-16:** "On Hold" button is visible to managers/admins only when the ticket status is In Progress
- **AC-17:** "Resume" button is visible to managers/admins only when the ticket status is On Hold
- **AC-18:** "Resolve" button is visible to managers/admins only when the ticket status is In Progress
- **AC-19:** "Close" button is visible to the submitting employee only when the ticket status is Resolved
- **AC-20:** "Reopen" button is visible to the submitting employee only when the ticket status is Resolved
- **AC-21:** "Cancel" button is visible to employees (own tickets in Open, In Progress, or On Hold status) and to managers/admins (any non-terminal ticket: Open, In Progress, On Hold, Resolved)
- **AC-22:** "Delete" button visibility follows the same rules as the gear icon Delete in DR-003-001-04
- **AC-23:** No action buttons are shown on the left panel for Closed or Cancelled tickets (except Back to List and Delete for admin on Cancelled)
- **AC-24:** View-only users (no management permission) see no action buttons other than "Back to List"

**Status Actions:**
- **AC-25:** Clicking any status action button opens a confirmation dialog with the appropriate title and message per the dialog table in Section 3
- **AC-26:** Confirming a status action updates the ticket status; the page refreshes in place and displays the new status badge
- **AC-27:** After a successful status change, a success toast is displayed (e.g., "Request ticket has been marked as In Progress")
- **AC-28:** Cancelling or closing a status action dialog returns the user to the detail page with no changes
- **AC-29:** If the ticket's status has been changed by another user between page load and action, an error toast informs the user and the page reloads to show the current state

**Navigation:**
- **AC-30:** "Back to List" returns the user to the Request Tickets List and preserves any active search/filter state
- **AC-31:** Clicking "Edit" navigates to the Update Request Ticket form (DR-003-001-03) pre-filled with the ticket's current data
- **AC-32:** After a successful deletion from the detail page, the user is redirected to the Request Tickets List with a success toast

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Employee views own Open ticket | Click on own Open ticket row | Detail page loads; Edit + Cancel shown; no status transition buttons | High |
| Employee views own Resolved ticket | Click on own Resolved ticket row | Detail page loads; Close + Reopen shown | High |
| Manager views any Open ticket | Click on any Open ticket | Detail page loads; Edit + In Progress + Cancel shown | High |
| Manager views In Progress ticket | Click on In Progress ticket | Detail page loads; On Hold + Resolve + Edit + Cancel shown | High |
| Manager views On Hold ticket | Click on On Hold ticket | Detail page loads; Resume + Cancel shown | High |
| Manager views Resolved ticket | Click on Resolved ticket | Detail page loads; Cancel shown (for manager); no Close/Reopen (submitter-only) | High |
| Employee views own Closed ticket | Click on own Closed ticket | Detail page loads; no action buttons except Back to List | Medium |
| Employee views own Cancelled ticket | Click on own Cancelled ticket | Detail page loads; no action buttons except Back to List | Medium |
| Status change — In Progress | Manager clicks "In Progress"; confirms | Status badge updates to Amber "In Progress"; success toast | High |
| Status change — Resolve | Manager clicks "Resolve"; confirms | Status badge updates to Green "Resolved"; success toast | High |
| Status change — Close | Employee clicks "Close"; confirms | Status badge updates to Gray "Closed"; Close + Reopen disappear | High |
| Status change — Reopen | Employee clicks "Reopen"; confirms | Status badge updates to Blue "Open"; Edit + Cancel shown | High |
| Status change — Cancel | Employee clicks "Cancel" on own Open ticket; confirms | Status badge updates to Red "Cancelled"; all action buttons disappear | High |
| Status conflict | Another user changes status; current user clicks action | Error toast; page reloads with current status | Medium |
| Ticket not found | Access deleted ticket URL | "Request ticket not found." error state with Back to List link | Medium |
| Attachment present | Ticket with attachment | Clickable filename shown; file accessible on click | Medium |
| No attachment | Ticket without attachment | "No attachment" placeholder shown | Medium |
| Back to list | Click "Back to List" | Returns to list; active search/filter preserved | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Only users with request ticket view permission can access the Request Ticket Details page. Permissions are configured via EP-001 US-004. No role names are hardcoded.
- **SR-02:** **Data visibility:** Employees can only view the detail page for their own submitted tickets. Accessing another employee's ticket via direct URL is blocked server-side with a redirect to the unauthorized page. Managers/admins can view any ticket.
- **SR-03:** **Read-only display:** The detail page presents ticket data in read-only format. No inline editing is permitted. Modifications are handled exclusively through the Update Request Ticket form (DR-003-001-03).
- **SR-04:** **Status transition rules:** All status transitions from this page follow the same business rules as status transitions from the list gear icon. The server enforces valid transitions; invalid transition requests are rejected.
- **SR-05:** **Submitter-exclusive actions:** "Close" and "Reopen" are only available to the submitting employee on their own Resolved ticket. Managers/admins cannot close or reopen on behalf of the submitter from this page.
- **SR-06:** **Concurrent status change protection:** Before processing any status action, the server verifies the ticket is still in the expected status. If the status has changed since the page loaded, the action is rejected and the client is instructed to refresh.
- **SR-07:** **Ticket not found:** If the ticket ID in the URL no longer exists (soft-deleted or invalid), the server returns a not-found response and the page displays an error state with a link back to the list.
- **SR-08:** **Back navigation state preservation:** The previous list state (search query, active filters, current page) is preserved when navigating back to the list from the detail page.
- **SR-09:** **Attachment access:** Clicking an attachment link opens or downloads the file. The file access is permission-controlled — only users who can view the ticket can access the attachment.
- **SR-10:** **Delete from detail page:** The delete action on this page follows the exact same rules and behavior as delete from the list gear icon (DR-003-001-04) — confirmation dialog, soft delete, redirect to list on success.

**Status Transition Rules (from §9.3 knowledge base):**
```
Open          → In Progress   (Manager/Admin)
In Progress   → On Hold       (Manager/Admin)
On Hold       → In Progress   (Manager/Admin — Resume)
In Progress   → Resolved      (Manager/Admin)
Resolved      → Closed        (Submitting Employee — Close)
Resolved      → Open          (Submitting Employee — Reopen)
Open          → Cancelled     (Employee own ticket OR Manager/Admin)
In Progress   → Cancelled     (Employee own ticket OR Manager/Admin)
On Hold       → Cancelled     (Employee own ticket OR Manager/Admin)
Resolved      → Cancelled     (Manager/Admin only)
Closed        → [terminal — no further transitions]
Cancelled     → [terminal — no further transitions]
```

**Dependencies:**
- **EP-001 US-001 (Authentication):** User must be signed in
- **EP-001 US-004 (Role & Permission Management):** Controls page access, action button visibility, and status transition authorization
- **EP-001 US-005 (User Management):** Employee profile data (name, department, position, email, phone) displayed in Employee section
- **DR-003-001-01 (Request Tickets List):** Entry point to the detail page; receives user back after navigation
- **DR-003-001-03 (Update Request Ticket):** Target of the Edit action
- **DR-003-001-04 (Delete Request Ticket):** Delete logic reused on this page

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Skeleton loading blocks in the right content panel prevent layout shift and provide visual feedback while ticket data loads — pattern consistent with Detail/Profile View standard (§8.1 knowledge base)
- **UX-02:** Action buttons in the left panel are rendered only after data loads — prevents the user from clicking an action before the page knows the current ticket status
- **UX-03:** Action buttons are visually grouped into three tiers: navigation (Back, Edit), status transitions (In Progress, On Hold, Resume, Resolve, Close, Reopen), and danger actions (Cancel, Delete) — separated by a visual divider to reduce risk of accidental destructive actions
- **UX-04:** The Cancel and Delete buttons use danger/red styling to signal destructive intent — consistent with the established danger button pattern across the HRM platform
- **UX-05:** Status action confirmation dialogs include specific, action-appropriate messaging (not a generic "Are you sure?") so the user understands the consequence before confirming
- **UX-06:** After a successful status change, the page refreshes in place (no full navigation) — the user remains in the detail context and sees the updated status badge and revised action buttons immediately
- **UX-07:** The status badge is prominently displayed near the Subject/title — users can assess the current lifecycle stage at a glance without scanning the full page
- **UX-08:** "Back to List" preserves active search/filter state — users working through a filtered list can navigate back to exactly where they were without losing context
- **UX-09:** The full Description and Subject are displayed without truncation on this page — in contrast to the list where Subject is truncated. Users come to this page specifically to read the full content
- **UX-10:** Attachment shows as a clickable filename — user knows a file is present and can download it directly from this page without navigating elsewhere
- **UX-11:** Success toasts auto-dismiss after 5 seconds; error toasts persist until dismissed

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Two-panel layout: left action menu (~220px) + right content panel (remaining width), consistent with User Details pattern |
| Below desktop | Out of scope |

**Accessibility Requirements:**
- [x] Keyboard navigable — all action buttons reachable via Tab; Enter/Space activates buttons
- [x] Screen reader compatible — page title announced; action buttons labeled with their action name; status badge includes text label (not color only)
- [x] Sufficient color contrast — status badge colors and text meet WCAG AA standard
- [x] Focus indicators visible on all interactive elements
- [x] Confirmation dialogs are modal with focus trap; Escape key closes dialog (same as Cancel); default focus on the less-destructive button (Cancel)
- [x] Attachment link includes accessible label indicating file name and type

**Design References:**
- No dedicated Figma frame extracted for Request Ticket Details — layout pattern follows User Details (DR-001-005-03) two-panel structure adapted for request ticket data and status management actions
- Status badge colors follow §9.3 of the knowledge base

---

## 8. Additional Information

### Out of Scope
- Inline editing of any field directly on the detail page — all edits go through the Update Request Ticket form
- Comment thread or activity log on the ticket detail page (future enhancement)
- Status change history/audit trail visible to the end user on this page (future enhancement)
- Automated email or in-app notifications triggered from status changes on this page (future enhancement)
- Bulk status actions — this page handles one ticket at a time
- Mobile or tablet layout
- Assigning tickets to specific staff members from this page (future enhancement)
- SLA or due date display (future enhancement)

### Open Questions
- [ ] **Row click vs. gear icon:** Should clicking anywhere on a ticket row navigate to the detail page, or should the user only access details via the gear icon "View Details" option? — **Owner:** Product Owner — **Status:** Pending
- [ ] **"View Details" in gear icon:** Does the gear icon on the list include an explicit "View Details" option, or is row click the only way in? Existing DRs (DR-003-001-01) did not explicitly define a "View Details" gear action — this needs confirmation. — **Owner:** Product Owner — **Status:** Pending
- [ ] **Status change success toast wording:** Confirm exact toast messages for each status transition (e.g., "Request ticket has been marked as In Progress" vs. "Ticket status updated to In Progress"). — **Owner:** Product Owner — **Status:** Pending
- [ ] **Page title format:** Confirm whether the page title should be "Request Ticket Details" (generic) or display the ticket subject (e.g., "IT Support — Laptop not working"). — **Owner:** UX Designer — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-003-001-01: Request Tickets List | Entry point to this page; receives user back on Back navigation |
| DR-003-001-03: Update Request Ticket | Target of the Edit action from this page |
| DR-003-001-04: Delete Request Ticket | Delete logic applied from this page |
| EP-001 US-001: Authentication | User must be signed in |
| EP-001 US-004: Role & Permission Management | Controls access and action visibility |
| EP-001 US-005: User Management | Employee data displayed in the Employee section |

### Notes
- **Detail view vs. status management page:** This page serves the dual purpose of displaying ticket data AND acting as the primary location for status transitions. This is consistent with the request ticket lifecycle where statuses need to be advanced in context, with full ticket information visible. The alternative (all status actions only from the list gear icon) was not chosen because the list shows truncated data and does not provide full context for decision-making.
- **"Close" and "Reopen" are submitter-exclusive:** These actions confirm the submitter's satisfaction with the resolution. Managers/admins intentionally cannot close on behalf of the submitter — the submitter must confirm. Managers/admins can instead cancel a resolved ticket if needed.
- **Consistent left-panel pattern:** The two-panel layout with a left action menu is established in the User Details pattern (DR-001-005-03). This DR extends that pattern to include status-specific action buttons, not just navigation/admin buttons.
- **No Figma design available for this screen** — layout decisions are based on the established User Details two-panel pattern and the request ticket status lifecycle rules confirmed in DR-003-001-01.

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
| 1.0 | 2026-04-03 | BA Agent | Initial draft — read-only detail view with two-panel layout and contextual status management actions |
