---
document_type: REQUIREMENTS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-005
story_name: "User Management"
status: draft
version: "0.1"
last_updated: "2026-03-24"
add_on_sections: ["UI Specifications"]
approved_by: null
related_documents:
  - path: "./ANALYSIS.md"
    relationship: sibling
  - path: "./FLOWCHART.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
revision_history: []
---

# Requirements: User Management

**Epic:** EP-001 (Foundation)
**Story:** US-005-user-management
**Status:** Draft — Stub (pending full requirements elaboration)

---

## 1. User Stories

| Story ID | As a... | I want to... | So that... | Priority |
|----------|---------|-------------|------------|----------|
| US-005-01 | User with user management permission | View all users in a searchable, filterable list | I can browse and manage all user accounts in the system | Critical |
| US-005-02 | User with user management permission | Create a new user account with profile details and role assignment | New staff can access the HRM system with appropriate permissions | Critical |
| US-005-03 | User with user management permission | Edit an existing user's profile and role | User information stays accurate as roles and assignments change | High |
| US-005-04 | User with user management permission | Deactivate or reactivate a user account | Departing staff lose access and returning staff regain it without data loss | High |

---

## 2. Functional Requirements (Preliminary)

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-US-005-01 | User list view with 9 columns (First Name, Last Name, Department, Position, System Role, Email, Phone Number, Status, Action) | Critical | From Figma |
| FR-US-005-02 | Search by name, email, and phone number | High | From Figma |
| FR-US-005-03 | Filter by Department | High | From Figma |
| FR-US-005-04 | Filter by Position | High | From Figma |
| FR-US-005-05 | Filter by System Role | High | From Figma |
| FR-US-005-06 | Filter by Status | High | From Figma |
| FR-US-005-07 | Reset all filters | Medium | From Figma |
| FR-US-005-08 | Pagination with configurable rows per page | Medium | From Figma |
| FR-US-005-09 | Export user list | Medium | From Figma |
| FR-US-005-10 | Create user (profile + role + dept/position assignment) | Critical | Pending design |
| FR-US-005-11 | Edit user (profile and/or role and/or dept/position) | High | Pending design |
| FR-US-005-12 | Deactivate/Reactivate user | High | Pending design |
| FR-US-005-13 | Gear icon action menu (Edit / Deactivate or Activate) | High | From Figma |

---

## 3. Business Rules (Preliminary)

| ID | Rule |
|----|------|
| BR-US-005-01 | User email addresses must be unique within the organization |
| BR-US-005-02 | A user must be assigned exactly one system role |
| BR-US-005-03 | A user must be assigned to a department and position |
| BR-US-005-04 | Deactivated users cannot log in but their data is preserved |
| BR-US-005-05 | Access to user management is controlled by permission (US-004) |

---

## 5. UI Specifications [ADD-ON]

> Extracted from Figma design via `/figma-extract` on 2026-03-24.
> Source: [User List](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3048-6011) (node `3048:6011`)

### 5.1 User List — Page Structure

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Page title | "User List" — Geist Semibold 24px, line-height 28.8px, letter-spacing -1, color `#0a0a0a` | `3048:6058` |
| Breadcrumb | Placeholder (expected: Users Management / Users) | `3048:6051` |
| Sidebar toggle | SidebarSimple icon, 16×16, top-left of content area | `3048:6050` |

### 5.2 User List — Action Bar

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Search input | Width: 320px, left-aligned. Placeholder: "Search by name, email, phone number..." Border: `#e4e4e7`, bg: `#ffffff`, radius: 6px | `3048:6062` |
| Filter: Department | Chip/button with filter icon + "Department" label. After search input. | `3048:6063` |
| Filter: Position | Chip/button with filter icon + "Position" label | `3048:6064` |
| Filter: System Role | Chip/button with filter icon + "System Role" label | `3048:6065` |
| Filter: Status | Chip/button with filter icon + "Status (2)" label. Count badge shows active filter values. | `3048:6066` |
| Reset | Button with refresh icon. Clears all active filters. | `3048:6067` |
| Export button | Right-aligned. Secondary style. Label: "Export". | `3048:6069` |
| Add New button | Right-aligned. Primary style: bg `#171717`, text `#fafafa`, radius: 6px. Label: "+ Add New". | `3048:6070` |

### 5.3 User List — Table

| Column | Width | Content | Figma Node |
|--------|-------|---------|------------|
| First Name | ~184px | User first name | `3048:6072` |
| Last Name | ~184px | User last name | `3048:6073` |
| Department | ~184px | Assigned department name | `3048:6074` |
| Position | ~184px | Assigned position name | `3048:6075` |
| System Role | ~184px | Assigned role name | `3048:6076` |
| Email | ~184px | User email address | `3048:6077` |
| Phone Number | ~184px | User phone number | `3048:6078` |
| Status | ~184px | Status badge (colored, "TBD" placeholder) | `3048:6079` |
| Action | ~184px | Gear icon for row actions | `3048:6080` |

### 5.4 User List — Pagination

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Rows per page | Dropdown, default: 10. Options: 10, 25, 50. | `3048:6081` |
| Page indicator | "Page 1 of 10" | `3048:6081` |
| Page navigation | Numbered buttons with prev/next arrows | `3048:6081` |

### 5.5 Acceptance Criteria from Design

- **AC-UI-01:** User List page displays with page title "User List" in Geist Semibold 24px
- **AC-UI-02:** Search input is 320px wide with placeholder "Search by name, email, phone number..."
- **AC-UI-03:** 4 filter chips displayed inline: Department, Position, System Role, Status
- **AC-UI-04:** Reset button with refresh icon clears all filters
- **AC-UI-05:** Export and "+ Add New" buttons are right-aligned
- **AC-UI-06:** Table has 9 columns: First Name, Last Name, Department, Position, System Role, Email, Phone Number, Status, Action
- **AC-UI-07:** Each data row has a gear icon in the Action column
- **AC-UI-08:** Status column shows colored status badges
- **AC-UI-09:** Pagination displays below the table
- **AC-UI-10:** "+ Add New" button and gear icon are hidden for users without user management permission

### 5.6 Design Gaps Requiring Resolution

| # | Gap | Impact | Action Required |
|---|-----|--------|-----------------|
| 1 | Status badge values unknown ("TBD") | Cannot define status states | PO: confirm Active/Inactive/Suspended and colors |
| 2 | Filter chip behavior unknown | Cannot define filter UX | PO: dropdown, multi-select, or single-select? |
| 3 | Status filter count "(2)" meaning | Unclear UX | PO: count of active filter values? |
| 4 | Gear icon actions undefined | Cannot define per-row actions | PO: Edit + Deactivate/Activate? |
| 5 | Search multi-field behavior | Name, email, phone — how? | PO: single search across all 3? Or separate? |
| 6 | Export format | Cannot implement | PO: CSV or Excel? |
| ~~7~~ | ~~Create/Edit User screens not designed~~ | ~~Cannot define forms~~ | ✅ Create User extracted 2026-03-24 — see Section 5.8. Edit User still pending. |
| 8 | Empty state not designed | No empty guidance | Design Team: provide empty state |

### 5.8 Create A New User — Page Structure

> Extracted from Figma on 2026-03-24.
> Source: [Create A New User](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3050-3782) (node `3050:3782`)

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Page title | "Create A New User" — Geist Semibold 24px | `3050:3829` |
| Breadcrumb | Placeholder (expected: Users Management / Users / Create A New User) | `3050:3822` |
| Layout | Full-page form, **two-column** — left: Personal Information (600px), right: User Account + Employee Profile (600px) | `3055:1482` |
| Action bar | Cancel + Save right-aligned (2 buttons, no "Save & Create Another") | `3057:1627` |

### 5.9 Create A New User — Personal Information Card (Left Column)

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Card | 600×615px, left column (x=217) | `3050:4203` |
| Section title | "Personal Information" — Geist Medium 14px | `3050:4205` |
| Upload Avatar | Image upload area. Accepts: .png/.jpeg/.jpg/.webp. Upload button + format hint. | `3050:4206` |
| * First name | Vertical Field (278px, left). Mandatory. Placeholder: "Enter first name" | `3050:4399` |
| * Last name | Vertical Field (278px, right). Mandatory. Placeholder: "Enter first name" (likely should be "Enter last name") | `3050:4428` |
| * Email | Vertical Field (278px, left). Mandatory. Placeholder: "Enter email" | `3050:4476` |
| * Phone number | Vertical Field (278px, right). Mandatory. Placeholder: "Enter phone number" | `3050:4477` |
| * Date of birth | Vertical Field with calendar icon (278px, left). Mandatory. Placeholder: "Enter date of birth" | `3050:4504` |
| * Gender | Vertical Field dropdown (278px, right). Mandatory. Placeholder: "Select gender" | `3050:4505` |
| Address | Form Field (576px, full width). Optional. Placeholder: "Enter address" | `3050:4221` |

### 5.10 Create A New User — User Account Card (Right Column Top)

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Card | 600×200px, right column (x=837) | `3055:1143` |
| Section title | "User Account" — Geist Medium 14px | `3055:1145` |
| * User Role | Vertical Field dropdown (576px). Mandatory. Placeholder: "Select user role" | `3055:1327` |
| * Account status | Toggle switch (33×18px) with label "Account status" and sub-label "Activate/Deactivate". Mandatory. | `3055:1320` |

### 5.11 Create A New User — Employee Profile Card (Right Column Bottom)

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Card | 600×380px, right column below User Account | `3055:1353` |
| Section title | "Employee Profile" — Geist Medium 14px | `3055:1355` |
| * Department | Vertical Field dropdown (278px, left). Mandatory. Placeholder: "Select department" | `3055:1373` |
| * Position | Vertical Field dropdown (278px, right). Mandatory. Placeholder: "Select position" | `3055:1483` |
| * Experience from (year) | Vertical Field (576px). Mandatory. Placeholder: "Enter year" | `3055:1499` |
| CV/Resumé | File input (576px). Optional. "Choose File / No file chosen" | `3089:2545` |
| Skills | Label "Skills" + button "+ Add skills" (110×32px). Optional. | `3055:1545` |

### 5.12 Acceptance Criteria from Design — Create User

- **AC-UI-11:** Create User page displays with page title "Create A New User" in Geist Semibold 24px
- **AC-UI-12:** Form uses a two-column layout: Personal Information (left), User Account + Employee Profile (right)
- **AC-UI-13:** Cancel and Save buttons are right-aligned in the title action bar
- **AC-UI-14:** Personal Information card includes: Upload Avatar, First name, Last name, Email, Phone number, Date of birth, Gender, Address
- **AC-UI-15:** User Account card includes: User Role dropdown, Account status toggle
- **AC-UI-16:** Employee Profile card includes: Department dropdown, Position dropdown, Experience from (year), CV/Resumé file input, Skills "+ Add skills" button
- **AC-UI-17:** Mandatory fields marked with asterisk (*): First name, Last name, Email, Phone number, Date of birth, Gender, User Role, Account status, Department, Position, Experience from
- **AC-UI-18:** Side-by-side field pairs use 278px width each with 20px gap
- **AC-UI-19:** Create User page is accessible only to users with user management permission

### 5.13 Design Gaps — Create User

| # | Gap | Impact | Action Required |
|---|-----|--------|-----------------|
| 9 | Avatar upload constraints | Cannot validate | PO: max file size, dimensions |
| 10 | CV/Resumé file constraints | Cannot validate | PO: accepted formats (PDF, DOCX?), max size |
| 11 | Skills selector mechanism | Cannot define UX | PO: modal, dropdown, or multi-select? |
| 12 | Gender dropdown options | Cannot populate | PO: Male, Female, Other, Prefer not to say? |
| 13 | Experience from (year) format | Cannot validate | PO: free text? Year picker? |
| 14 | Account status default | Cannot set initial state | PO: default Active on create? |
| 15 | Email uniqueness validation | Cannot define error | PO: server-side on submit? |
| 16 | Phone number format | Cannot validate | PO: international? digits only? |
| 17 | Last name placeholder copy-paste error | Says "Enter first name" | Design fix: "Enter last name" |
| ~~18~~ | ~~Edit User screen not designed~~ | ~~Cannot define edit form~~ | ✅ Update Information extracted 2026-03-26 — see Section 5.14 |

### 5.14 Update Information — Page Structure

> Extracted from Figma on 2026-03-26.
> Source: [Update Information](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3122-6323) (node `3122:6323`)

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Page context | "User Details" — Geist Semibold 24px, with back arrow + user name breadcrumb (e.g., "> Henry Tran") | `3122:6383` |
| Layout | Single-column form (600px), accessed from User Details left action panel | `3122:6804` |
| Action panel (left) | 7 buttons: Overview, Update Information (active), Change User Role, Change Email, Reset Password, Activate/Deactivate, Delete User | `3122:6609` |
| Save button | Full-width (600px), primary style, bottom of form. Single button (no Cancel). | `3122:6805` |

### 5.15 Update Information — Personal Information Card

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Card | 600px wide, single-column | `3122:6641` |
| Section title | "Personal Information" — Geist Medium 14px | `3122:6643` |
| Upload Avatar | Image upload area. Accepts: .png/.jpeg/.jpg/.webp. | `3122:6644` |
| * First name | Vertical Field (278px, left). Mandatory. Placeholder: "Enter first name" | `3122:6650` |
| * Last name | Vertical Field (278px, right). Mandatory. Placeholder: "Enter first name" (design copy error — should be "Enter last name") | `3122:6651` |
| * Date of birth | Vertical Field with calendar icon (278px, left). Mandatory. Placeholder: "Enter date of birth" | `3122:6745` |
| * Phone number | Vertical Field (278px, right). Mandatory. Placeholder: "Enter phone number" | `3122:6654` |
| * Gender | Vertical Field dropdown (576px, full width). Mandatory. Placeholder: "Select gender" | `3122:6657` |
| Address | Form Field (576px, full width). Optional. Placeholder: "Enter address" | `3122:6658` |

### 5.16 Update Information — Work Profile Card

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Card | 600px wide, below Personal Information | `3122:6428` |
| Section title | "Work Profile" — Geist Medium 14px | `3122:6446` |
| * Department | Vertical Field dropdown (278px, left). Mandatory. Placeholder: "Select department" | `3122:6449` |
| * Position | Vertical Field dropdown (278px, right). Mandatory. Placeholder: "Select position" | `3122:6450` |
| * Experience from (year) | Vertical Field (576px). Mandatory. Placeholder: "Enter year" | `3122:6451` |
| CV/Resumé | File input (576px). Optional. "Choose File / No file chosen" | `3122:6454` |
| Skills | Label "Skills" + button "+ Add skills" (110×32px). Optional. | `3122:6457` |

### 5.17 Key Differences: Update Information vs Create User

| Aspect | Create User | Update Information |
|--------|-------------|-------------------|
| Layout | Two-column (3 cards) | **Single-column** (2 cards stacked) |
| Email field | Present (mandatory) | **Absent** — managed via "Change Email" |
| User Role | Present (mandatory) | **Absent** — managed via "Change User Role" |
| Account Status | Present (toggle) | **Absent** — managed via "Activate/Deactivate" |
| Buttons | Cancel + Save (header) | **Save only** (full-width, bottom) |
| Context | Standalone page | **Sub-page** of User Details |
| Data | Empty form | **Pre-filled** with existing user data |

### 5.18 Acceptance Criteria from Design — Update Information

- **AC-UI-20:** Update Information is accessed from the User Details left action panel
- **AC-UI-21:** Form uses a single-column layout (600px) with Personal Information and Work Profile cards stacked vertically
- **AC-UI-22:** Personal Information card includes: Upload Avatar, First name, Last name, Date of birth, Phone number, Gender, Address
- **AC-UI-23:** Work Profile card includes: Department, Position, Experience from (year), CV/Resumé, Skills
- **AC-UI-24:** Email field is NOT present on this form (managed separately via Change Email action)
- **AC-UI-25:** User Role and Account Status are NOT present (managed via separate actions)
- **AC-UI-26:** Single Save button, full-width (600px), primary style, positioned at the bottom of the form
- **AC-UI-27:** All fields are pre-filled with the user's existing data on load
- **AC-UI-28:** Mandatory fields marked with asterisk (*): First name, Last name, Date of birth, Phone number, Gender, Department, Position, Experience from
- **AC-UI-29:** Update Information page is accessible only to users with user management permission

### 5.19 Design Gaps — Update Information

| # | Gap | Impact | Action Required |
|---|-----|--------|-----------------|
| 19 | No Cancel button on form | User cannot discard changes via button | PO: intentional? Use back arrow / action panel to leave? |
| 20 | No "unsaved changes" warning | User may lose edits by clicking another action | PO: show confirmation dialog when navigating away with changes? |
| 21 | Avatar constraints same as Create User | See gaps #9 | Same resolution needed |
| 22 | CV/Resumé constraints same as Create User | See gaps #10 | Same resolution needed |
| 23 | Skills selector same as Create User | See gaps #11 | Same resolution needed |

### 5.20 Change User Role — Page Structure

> Extracted from Figma on 2026-03-26.
> Source: [Change User Role](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-6810) (node `3123:6810`)

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Page context | "User Details" — same header as Update Information, with back arrow + user name breadcrumb | `3123:6870` |
| Layout | Single card (600px), accessed from User Details left action panel | `3123:6883` |
| Card title | "User Account" — Geist Medium 14px | `3123:7105` |
| Descriptive text | Two paragraphs explaining role change impact and immediacy | `3123:7520` |
| * User Role | Vertical Field dropdown (576px). Mandatory. Placeholder: "Select user role". Pre-filled with current role. | `3123:7107` |
| Save button | Full-width (600px), primary style, below the card. Single button (no Cancel). | `3123:6915` |

### 5.21 Acceptance Criteria from Design — Change User Role

- **AC-UI-30:** Change User Role is accessed from the User Details left action panel
- **AC-UI-31:** "User Account" card displays with descriptive text explaining role change impact
- **AC-UI-32:** Single mandatory field: User Role dropdown, pre-filled with current role
- **AC-UI-33:** Save button is full-width (600px), primary style, positioned below the card
- **AC-UI-34:** No Cancel button — navigation via action panel or back arrow
- **AC-UI-35:** Change User Role page is accessible only to users with user management permission

### 5.22 Change Email — Page Structure

> Extracted from Figma on 2026-03-26.
> Source: [Change Email](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-7139) (node `3123:7139`)

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Page context | "User Details" — same header pattern, back arrow + user name breadcrumb | `3123:7199` |
| Layout | Single card (600px), accessed from User Details left action panel | `3123:7212` |
| Card title | "Change Email" — Geist Medium 14px | `3123:7215` |
| Descriptive text | **⚠️ DESIGN ERROR:** Currently shows Reset Password text. Needs replacement with email change context. | `3123:7518` |
| * New Email | Vertical Field text input (576px). Mandatory. Placeholder: "Enter email" | `3123:7310` |
| Save button | Full-width (600px), primary style, below the card. Single button (no Cancel). | `3123:7218` |

### 5.23 Acceptance Criteria from Design — Change Email

- **AC-UI-36:** Change Email is accessed from the User Details left action panel
- **AC-UI-37:** "Change Email" card displays with descriptive text (pending design fix)
- **AC-UI-38:** Single mandatory field: New Email text input, placeholder "Enter email"
- **AC-UI-39:** Save button is full-width (600px), primary style, positioned below the card
- **AC-UI-40:** No Cancel button — navigation via action panel or back arrow
- **AC-UI-41:** Change Email page is accessible only to users with user management permission

### 5.24 Design Gaps — Change Email

| # | Gap | Impact | Action Required |
|---|-----|--------|-----------------|
| 24 | Descriptive text is copy-paste error from Reset Password | Incorrect user guidance | **Design Team: replace with email change context text** |
| 25 | Current email not displayed | User cannot see what email they are changing from | PO: show current email as read-only label above the input? |
| 26 | No email format validation shown in design | Cannot define error states | Assume: inline error for invalid email format |

### 5.25 Reset Password — Page Structure

> Extracted from Figma on 2026-03-26.
> Source: [Reset Password](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-7334) (node `3123:7334`)

| Element | Specification | Figma Node |
|---------|--------------|------------|
| Page context | "User Details" — same header pattern, back arrow + user name breadcrumb | `3123:7394` |
| Layout | Single card (600px), accessed from User Details left action panel | `3123:7407` |
| Card title | "Reset Password" — Geist Medium 14px | `3123:7410` |
| Descriptive text | Two paragraphs: reset link flow (time-limited, single-use, sent to email) | `3123:7504` |
| Email display | Read-only: mail icon + "Email" label + user's email address (e.g., "henry@exnodes.vn") | `3123:7511` |
| Reset Password button | Full-width (600px), primary style, label "Reset Password" (not "Save"). No other buttons. | `3123:7417` |

### 5.26 Acceptance Criteria from Design — Reset Password

- **AC-UI-42:** Reset Password is accessed from the User Details left action panel
- **AC-UI-43:** "Reset Password" card displays with descriptive text explaining the reset link flow
- **AC-UI-44:** User's current email is displayed read-only with mail icon — admin verifies before triggering
- **AC-UI-45:** No input fields — this is a pure action screen
- **AC-UI-46:** Single button labeled "Reset Password" (not "Save"), full-width, primary style
- **AC-UI-47:** Reset Password page is accessible only to users with user management permission

---

## 6. Next Steps

- [ ] Resolve open questions with Product Owner (status values, filter behavior, gear actions)
- [ ] Obtain Create/Edit User Figma screens from Design Team
- [ ] Elaborate full requirements with acceptance criteria and NFRs
- [ ] Write detail requirements (DRs) for each feature

---

**Document Version:** 0.1
**Last Updated:** 2026-03-24
**Author:** BA Agent
**Reviewer:** Pending
