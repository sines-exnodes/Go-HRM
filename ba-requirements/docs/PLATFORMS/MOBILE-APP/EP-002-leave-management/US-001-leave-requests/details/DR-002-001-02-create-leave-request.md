---
document_type: DETAIL_REQUIREMENT
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-002
story_id: US-001
story_name: "Leave Requests"
detail_id: DR-002-001-02
detail_name: "Create Leave Request"
parent_requirement: FR-US-001-20
status: draft
version: "1.0"
created_date: 2026-04-17
last_updated: 2026-04-17
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "./DR-002-001-01-leave-requests-dashboard.md"
    relationship: sibling
  - path: "../../../../WEB-APP/EP-002-leave-management/US-001-leave-requests/details/DR-002-001-02-create-leave-request.md"
    relationship: reference
input_sources:
  - type: figma
    file_id: "YEHeFgVZau7wmo9BZBVuZC"
    node_id: "3281:1274"
    frame_name: "Leave Requests - Submit Leave Request"
    extraction_date: "2026-04-17"
---

# Detail Requirement: Create Leave Request

**Detail ID:** DR-002-001-02
**Parent Requirement:** FR-US-001-20
**Story:** US-001-leave-requests
**Epic:** EP-002 (Leave Management)
**Platform:** MOBILE-APP
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **employee**, I want to **submit a new leave request from my mobile device**, so that **I can request time off quickly without needing to access a desktop computer**.

**Purpose:** The Create Leave Request screen enables employees to submit leave requests on-the-go. The mobile form is streamlined for touch interaction with native date pickers, dropdown selections, and a simplified single-card layout. This feature supports the business goal of improving employee self-service adoption and reducing time-to-submission for leave requests.

**Target Users:**
- All authenticated employees with access to the mobile app
- Employees requesting time off while away from their desks
- Employees who prefer mobile-first interactions

**Key Functionality:**
- Select leave type from predefined options (Annual, Sick, Personal, Maternity, Unpaid)
- Select date range using native mobile date pickers
- Select leave period (Full Day, Morning Half, Afternoon Half)
- Provide reason for leave request (mandatory)
- Optionally attach supporting documentation
- Submit request with confirmation
- Balance validation before submission

---

## 2. User Workflow

**Entry Point:** Leave Requests Dashboard > Hero Card "Apply for Leave" button OR Dashboard > FAB "+" button

**Preconditions:**
- User is signed in to the mobile app (EP-001 US-001)
- User has employee role with leave request permission
- User has sufficient leave balance for the selected leave type (warning shown if insufficient)

**Main Flow:**
1. User taps "Apply for Leave" button on Leave Requests Dashboard
2. System navigates to Create Leave Request screen
3. System displays header with back arrow, "Submit Leave Request" title, and "Done" button (disabled)
4. System displays form card with all input fields
5. User taps "Leave Type" field and selects from dropdown options
6. User taps "From Date" field and selects start date using native date picker
7. User taps "To Date" field and selects end date using native date picker
8. User taps "Leave Period" field and selects period from dropdown options
9. User taps "Reason" field and enters reason text
10. User optionally taps "Attachment" field to upload supporting document
11. System enables "Done" button when all mandatory fields are completed
12. User taps "Done" button to submit request
13. System validates all fields and submits request to backend
14. System displays success feedback and navigates back to Dashboard

**Alternative Flows:**

| Action | Flow |
|--------|------|
| **Back without submission** | User taps back arrow > If form is dirty, show discard confirmation > If confirmed or clean, navigate back to Dashboard |
| **Insufficient balance** | User selects dates > System calculates total days > If balance insufficient, show warning banner but allow submission |
| **Invalid date range** | User selects To Date before From Date > System shows inline error "To Date must be after From Date" |
| **Same-day leave** | User selects same date for From and To > System allows (single day leave) |
| **Attachment upload** | User taps Attachment > System opens file picker > User selects file > System validates format/size > Shows filename |
| **Remove attachment** | User taps attached file > System shows remove option > User confirms > Attachment cleared |
| **Network error on submit** | User taps Done > API fails > System shows error toast "Unable to submit. Please try again." > Form remains open |

**Exit Points:**
- **Success:** Request submitted > Success toast "Leave request has been submitted" > Navigate to Dashboard (refreshed)
- **Cancel (clean form):** Back arrow > Navigate to Dashboard immediately
- **Cancel (dirty form):** Back arrow > Discard confirmation > If confirmed, navigate to Dashboard
- **Error:** Toast displayed > User can retry or cancel

---

## 3. Field Definitions

### Input Fields

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Leave Type | Dropdown | Must select one option | Yes | None (placeholder: "Select leave type") | Type of leave: Annual, Sick, Personal, Maternity, Unpaid |
| From Date | Date Picker | Any date (past dates allowed for retroactive requests) | Yes | None (placeholder: "Select date") | Start date of leave period |
| To Date | Date Picker | Must be >= From Date | Yes | None (placeholder: "Select date") | End date of leave period |
| Leave Period | Dropdown | Must select one option | Yes | None (placeholder: "Select period") | Full Day, Morning Half, Afternoon Half |
| Reason | Text Area | Min 1 character (after trim), max 500 characters | Yes | None (placeholder: "Enter reason") | Explanation for leave request (multi-line) |
| Attachment | File Upload | PDF, PNG, JPG, DOCX; max 5MB | No | None (placeholder: "Choose file") | Supporting documentation |

### Leave Type Options

| Option | Description |
|--------|-------------|
| Annual | Paid annual leave from employee quota |
| Sick | Sick leave for health-related absence |
| Personal | Personal leave for private matters |
| Maternity | Maternity/paternity leave |
| Unpaid | Unpaid leave (no quota deduction) |

### Leave Period Options

| Option | Day Calculation |
|--------|-----------------|
| Full Day | 1.0 day per date in range |
| Morning Half | 0.5 day per date in range |
| Afternoon Half | 0.5 day per date in range |

### Interaction Elements

| Element | Type | Location | State/Condition | Trigger Action | Description |
|---------|------|----------|-----------------|----------------|-------------|
| Back Arrow | Icon button (18x18) | Header left | Always visible | Navigate back (with discard check if dirty) | ArrowLeft icon |
| Page Title | Text | Header center-left | Always visible | None (display only) | "Submit Leave Request" |
| Done Button | Primary button (67x28) | Header right | Disabled until all mandatory fields filled | Submit form | Pill-shaped, white background, shadow |
| Leave Type Row | Touchable row | Form card | Placeholder or selected value | Open leave type picker | CaretDown icon indicator |
| From Date Row | Touchable row | Form card | Placeholder or selected date | Open native date picker | CalendarDots icon indicator |
| To Date Row | Touchable row | Form card | Placeholder or selected date | Open native date picker | CalendarDots icon indicator |
| Leave Period Row | Touchable row | Form card | Placeholder or selected value | Open period picker | CaretDown icon indicator |
| Reason Row | Text input row | Form card | Placeholder or entered text | Focus text input | No icon |
| Attachment Row | Touchable row | Form card | Placeholder or filename | Open file picker | CloudArrowUp icon indicator |

---

## 4. Data Display

### Header Section

| Element | Format | Description |
|---------|--------|-------------|
| Back Arrow | 18x18 ArrowLeft icon | Navigation back to Dashboard |
| Title | "Submit Leave Request" (14px medium, #020817) | Screen title |
| Done Button | Pill button (67x28), white bg, shadow | Primary submit action |

### Form Card Layout

| Section | Content | Separator |
|---------|---------|-----------|
| Leave Type | Label (100px) + Value/Placeholder + CaretDown icon | Line below |
| From Date | Label (100px) + Value/Placeholder + CalendarDots icon | Line below |
| To Date | Label (100px) + Value/Placeholder + CalendarDots icon | Line below |
| Leave Period | Label (100px) + Value/Placeholder + CaretDown icon | Line below |
| Reason | Label (100px) + Value/Placeholder (no icon) | Line below |
| Attachment | Label (100px) + Value/Placeholder + CloudArrowUp icon | None (last row) |

**Field Row Layout:**
- Label: 100px width, 14px medium, #020817 (text-foreground)
- Value: Flex-1, 14px regular, #64748b (text-muted-foreground) for placeholder, #020817 for actual value
- Icon: 18x18, aligned right
- Row height: 18px content + 20px gap = ~38px per row
- Separator: 1px line, #e2e8f0 (border-border)

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Initial | Screen first opens | All fields show placeholders, Done button disabled |
| Partial | Some fields filled | Completed fields show values, Done button disabled |
| Complete | All mandatory fields filled | All fields show values, Done button enabled |
| Submitting | Done button tapped | Done button shows loading spinner, form disabled |
| Error | Validation error | Inline error below affected field, red text |
| Success | Submission complete | Success toast, navigate to Dashboard |
| Balance Warning | Calculated days exceed remaining balance | Warning banner below form card (amber) |

### Page Layout (ASCII Diagram)

```
┌─────────────────────────────────────────┐
│ [<] Submit Leave Request        [Done] │  <- Header (70px)
├─────────────────────────────────────────┤
│                                         │
│ ┌─────────────────────────────────────┐ │
│ │ Leave Type *      Select leave type│ │  <- Form Card
│ │ ─────────────────────────────────── │ │     (rounded-xl, white bg)
│ │ From Date *       Select date    📅│ │
│ │ ─────────────────────────────────── │ │
│ │ To Date *         Select date    📅│ │
│ │ ─────────────────────────────────── │ │
│ │ Leave Period *    Select period   ▼│ │
│ │ ─────────────────────────────────── │ │
│ │ Reason *          Enter reason     │ │
│ │ ─────────────────────────────────── │ │
│ │ Attachment        Choose file    ☁️│ │
│ └─────────────────────────────────────┘ │
│                                         │
│ [Balance Warning Banner - if applicable]│
│                                         │
└─────────────────────────────────────────┘
```

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

**Navigation:**
- **AC-01:** Tapping "Apply for Leave" on Dashboard navigates to Create Leave Request screen
- **AC-02:** Header displays back arrow, "Submit Leave Request" title, and "Done" button
- **AC-03:** Tapping back arrow on clean form navigates immediately to Dashboard
- **AC-04:** Tapping back arrow on dirty form shows "Discard unsaved changes?" confirmation

**Form Display:**
- **AC-05:** Form card displays all 6 fields with labels and placeholder text
- **AC-06:** Mandatory fields are marked with asterisk (*): Leave Type, From Date, To Date, Leave Period, Reason
- **AC-07:** Attachment field is not marked with asterisk (optional)
- **AC-08:** Each field row shows appropriate icon: CaretDown for dropdowns, CalendarDots for dates, CloudArrowUp for attachment
- **AC-09:** Fields are separated by horizontal lines except the last field

**Field Interactions:**
- **AC-10:** Tapping Leave Type field opens leave type picker with 5 options
- **AC-11:** Tapping From Date field opens native mobile date picker
- **AC-12:** Tapping To Date field opens native mobile date picker
- **AC-13:** Tapping Leave Period field opens period picker with 3 options
- **AC-14:** Tapping Reason field focuses text input with keyboard
- **AC-15:** Tapping Attachment field opens device file picker

**Validation:**
- **AC-16:** From Date can be any date (past dates allowed for retroactive leave requests)
- **AC-17:** To Date must be greater than or equal to From Date
- **AC-18:** Selecting To Date before From Date shows inline error "To Date must be after From Date"
- **AC-19:** Reason field accepts 1-500 characters
- **AC-20:** Attachment accepts PDF, PNG, JPG, DOCX formats up to 5MB
- **AC-21:** Invalid file format shows error "File type not supported"
- **AC-22:** File exceeding 5MB shows error "File size must be under 5MB"

**Done Button:**
- **AC-23:** Done button is disabled until all 5 mandatory fields are completed
- **AC-24:** Done button becomes enabled when all mandatory fields have values
- **AC-25:** Tapping enabled Done button triggers form submission
- **AC-26:** Done button shows loading state during submission

**Submission:**
- **AC-27:** Successful submission shows toast "Leave request has been submitted"
- **AC-28:** Successful submission navigates to Dashboard with refreshed data
- **AC-29:** Failed submission shows toast "Unable to submit. Please try again."
- **AC-30:** Failed submission keeps form open with entered data preserved

**Balance Warning:**
- **AC-31:** If calculated leave days exceed remaining balance, warning banner is displayed
- **AC-32:** Warning does not block submission (user can still submit for manager approval)

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Navigate to form | Tap "Apply for Leave" on Dashboard | Create Leave Request screen displays | High |
| Complete all fields | Fill all 5 mandatory fields | Done button becomes enabled | High |
| Submit valid request | Tap Done with all fields valid | Success toast + navigate to Dashboard | High |
| Invalid date range | To Date < From Date | Inline error displayed | High |
| Cancel clean form | Tap back arrow with no changes | Navigate immediately to Dashboard | Medium |
| Cancel dirty form | Tap back arrow after entering data | Discard confirmation dialog | Medium |
| Upload valid attachment | Select PDF under 5MB | Filename displayed in Attachment row | Medium |
| Upload invalid file | Select .exe file | "File type not supported" error | Medium |
| Upload oversized file | Select 10MB file | "File size must be under 5MB" error | Medium |
| Insufficient balance | Request 10 days with 5 days remaining | Warning banner shown, can still submit | Medium |
| Network error | Submit with no connectivity | Error toast, form preserved | Medium |
| Same-day leave | From Date = To Date | Accepted as 1-day leave | Medium |
| Half-day leave | Select "Morning Half" period | 0.5 days calculated per date | Medium |
| Retroactive leave | Select From Date in the past | Accepted — allows post-facto submission | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Leave request is created with status "Pending" upon successful submission
- **SR-02:** Leave request is associated with the logged-in user's employee ID
- **SR-03:** Total leave days calculated: (End Date - Start Date + 1) * Period multiplier (1.0 for Full, 0.5 for Half)
- **SR-04:** Leave balance is NOT deducted at submission — only upon manager approval
- **SR-05:** If requested days exceed remaining balance, request proceeds but is flagged for manager attention
- **SR-06:** Attachment is uploaded to secure storage with reference stored in leave request record
- **SR-07:** Created timestamp and submitted_by user ID are recorded automatically
- **SR-08:** Leave type must match one of the configured types from WEB-APP (Annual, Sick, Personal, Maternity, Unpaid)
- **SR-09:** Only active employees can submit leave requests
- **SR-10:** Duplicate submission prevention: same user cannot submit overlapping date ranges for the same leave type within 1-minute window

**State Transitions:**

```
[Form Empty] → [User fills field] → [Form Partial]
[Form Partial] → [All mandatory fields filled] → [Form Complete]
[Form Complete] → [User taps Done] → [Submitting]
[Submitting] → [API Success] → [Success Toast] → [Dashboard]
[Submitting] → [API Error] → [Error Toast] → [Form Complete]
```

**Leave Days Calculation:**

```
Total Days = Number of Days in Range × Period Multiplier

Period Multipliers:
- Full Day: 1.0
- Morning Half: 0.5
- Afternoon Half: 0.5

Example: From Apr 24 to Apr 26 (3 days) with Morning Half = 3 × 0.5 = 1.5 days
```

**Dependencies:**
- **EP-001 US-001 (Authentication):** User must be signed in
- **EP-001 US-005 (User Profile):** Employee ID for request association
- **WEB-APP EP-002:** Leave management API, leave types configuration
- **DR-002-001-01 (Dashboard):** Entry point and post-submission destination

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Form loads with focus on first field (Leave Type) for quick entry
- **UX-02:** Native mobile date pickers provide familiar interaction patterns
- **UX-03:** Dropdown options use native iOS ActionSheet / Android BottomSheet
- **UX-04:** Keyboard automatically dismisses when tapping outside text fields
- **UX-05:** Done button remains visible in header (not scrolled away)
- **UX-06:** Form card has sufficient padding (16px) for comfortable touch targets
- **UX-07:** Field rows have adequate height (38px) meeting 44pt minimum touch target
- **UX-08:** Visual feedback on field selection (highlight/focus state)
- **UX-09:** Loading state on Done button prevents double-submission
- **UX-10:** Error messages appear inline near the relevant field
- **UX-11:** Success/error toasts auto-dismiss after 3 seconds
- **UX-12:** Haptic feedback on successful submission (device-dependent)

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Standard mobile (393px width) | Design as specified from Figma |
| Small mobile (< 360px) | Field labels may wrap; values truncate with ellipsis |
| Large mobile (> 414px) | Additional horizontal padding; form card expands |
| Tablet | Out of scope for this release |

**Accessibility Requirements:**
- [x] All interactive elements have minimum 44x44 touch target
- [x] Screen reader labels for all fields ("Leave Type, required, dropdown")
- [x] Field labels announce mandatory status ("Required" suffix)
- [x] Date picker values announced in readable format ("April 24, 2026")
- [x] Error messages announced when displayed
- [x] Done button state announced ("Done, button, disabled" vs "Done, button")
- [x] Form structure uses semantic grouping for screen readers
- [x] Sufficient color contrast for text (meets WCAG 2.1 AA)
- [x] Focus indicators visible on all interactive elements

**Design References:**
- Figma: [Submit Leave Request](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3281-1274) (node `3281:1274`)
- Design tokens:
  - Background: `#f0f1f3` (screen), `#ffffff` (card)
  - Text foreground: `#020817`
  - Text muted: `#64748b`
  - Border: `#e2e8f0`
  - Button shadow: `0px 2px 8.9px rgba(0,0,0,0.1)`
  - Radius: `rounded-xl` (8.8px) for card, `rounded-full` (9999px) for button
  - Font: Encode Sans (Regular, Medium, SemiBold)
  - Spacing: 12px (screen padding), 16px (card padding), 20px (field gap)

---

## 8. Additional Information

### Out of Scope
- Manager approval workflow (manager actions are out of scope for v1.0)
- Editing submitted leave request (separate future DR)
- Cancelling pending leave request (separate future DR)
- Leave calendar view for date selection
- Team leave visibility while selecting dates
- Recurring leave requests
- Multi-day period variation (e.g., Full Day for some dates, Half Day for others)
- Draft saving functionality
- Offline form submission with sync

### Open Questions

| Question | Answer | Owner |
|----------|--------|-------|
| Should Reason field support multi-line input? | **Yes** — textarea (multi-line) | Product Owner |
| Maximum attachment file size (currently assumed 5MB)? | **5MB** — consistent with WEB-APP | Product Owner |
| Should balance warning prevent submission or just warn? | **Warn only** — non-blocking, consistent with WEB-APP | Product Owner |
| Minimum reason length (currently 1 character)? | **1 character (after trim)** — consistent with WEB-APP | Product Owner |
| Should dates default to "tomorrow" for From Date? | **No default** — user must select | UX Team |

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-002-001-01: Leave Requests Dashboard | Entry point ("Apply for Leave" button) |
| DR-002-001-03: Leave Request Details (planned) | View submitted request details |
| DR-002-001-04: Leave Request List (planned) | View all submitted requests |
| EP-001 US-001: Authentication | User must be signed in |
| WEB-APP DR-002-001-02: Create Leave Request | Web equivalent form |

### Notes
- This form follows the **Create Form Pattern** (Section 2 of PROJECT_KNOWLEDGE.md)
- Unlike WEB-APP, mobile uses a **single-card layout** rather than multiple grouped cards
- The **Done button in header** pattern differs from WEB-APP's dual Save buttons (header + bottom)
- **Native pickers** are used for dates and dropdowns instead of custom components
- Balance validation is **advisory only** — employees can request leave exceeding balance for manager decision
- The **Attachment** field is optional per Figma design (no asterisk)
- **Retroactive leave requests allowed** — past dates can be selected for scenarios like sick leave submitted after recovery or emergency absences

### Differences from WEB-APP

| Aspect | WEB-APP | MOBILE-APP |
|--------|---------|------------|
| Layout | Full-page form with multiple cards | Single form card |
| Submit Button | Header Save + Bottom Save | Header "Done" button only |
| Date Picker | Custom calendar component | Native iOS/Android date picker |
| Dropdowns | Custom searchable dropdown | Native ActionSheet/BottomSheet |
| Employee Selection | Admin can select employee | Employees submit own requests only (v1.0) |
| Cancel | Header Cancel button | Back arrow with discard check |

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
| 1.0 | 2026-04-17 | BA Agent | Initial draft — Figma design context from node 3281:1274 |
