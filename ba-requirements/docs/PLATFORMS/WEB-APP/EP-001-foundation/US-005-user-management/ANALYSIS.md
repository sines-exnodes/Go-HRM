---
document_type: ANALYSIS
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-005
story_name: "User Management"
status: draft
version: "1.0"
last_updated: "2026-03-24"
add_on_sections: ["Design Context"]
approved_by: null
related_documents:
  - path: "./REQUIREMENTS.md"
    relationship: sibling
  - path: "./FLOWCHART.md"
    relationship: sibling
  - path: "./TODO.yaml"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
revision_history: []
input_sources:
  - type: figma
    description: "User List screen"
    node_id: "3048:6011"
    extraction_date: "2026-03-24"
  - type: figma
    description: "Create A New User screen"
    node_id: "3050:3782"
    extraction_date: "2026-03-24"
  - type: figma
    description: "Update Information screen (User Details sub-page)"
    node_id: "3122:6323"
    extraction_date: "2026-03-26"
---

# Analysis: User Management

**Epic:** EP-001 (Foundation)
**Story:** US-005-user-management
**Status:** Draft

---

## 1. Business Context

### Problem Statement

Organizations need a centralized way to manage user accounts — who can access the HRM system, with what role, and in what department/position. Without a user management module, onboarding new staff, updating assignments, and deactivating departing employees requires direct database or admin panel intervention.

### Stakeholders

- **Primary Users**: Administrators with user management permission
- **Secondary Users**: All HRM users (managed by this module)
- **Business Owner**: HR department leadership

### Business Goals

- Goal 1: Provide a centralized interface for managing all user accounts in the HRM platform
- Goal 2: Enable administrators to create, edit, and deactivate user accounts with proper role and department/position assignment

---

## 2. Scope Definition

### In Scope

- User list view (search, filter, table, pagination, export)
- Create new user
- Edit existing user
- Deactivate/activate user
- Assign department, position, and system role to users
- Access controlled by role permissions (US-004)

### Out of Scope

- Self-service user registration (admin-only creation)
- User authentication flow (covered by US-001)
- Role/permission definition (covered by US-004)
- Department/position definition (covered by EP-008)
- Employee HR profile (covered by EP-002)

### Dependencies

- **Internal**: US-001 (Authentication), US-004 (Role & Permission), EP-008 US-001 (Department), EP-008 US-002 (Position)
- **Downstream**: All modules — users created here are the actors across the platform

---

## 7. Design Context [ADD-ON]

> Extracted from Figma design via `/figma-extract` on 2026-03-24.
> One screen extracted: User List.

### Source Information

| Attribute | Value |
|-----------|-------|
| **Figma File** | [Exnodes HRM](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC) |
| **Frame** | [User List](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3048-6011) |
| **Node ID** | `3048:6011` |
| **Dimensions** | 1920 × 1080 |
| **Extraction Date** | 2026-03-24 |

---

### Screen: User List

#### Layout Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb                            [Top Bar] │
├──────────────┬──────────────────────────────────────────────────────────────┤
│  [Logo]      │  User List                                                  │
│              │                                                             │
│  Richard R.  │  [Search 320px] [Department] [Position] [System Role]       │
│  Super Admin │                 [Status (2)] [Reset ↻]    [Export] [+AddNew]│
│  ──────────  │                                                             │
│  Users Mgmt  │  ┌──────┬──────┬──────┬──────┬──────┬──────┬──────┬──────┬──┐│
│  > Users ◄   │  │First │Last  │Dept  │Posit │System│Email │Phone │Status│⚙ ││
│  > Roles     │  │Name  │Name  │      │ion   │Role  │      │Number│      │  ││
│  > Perms     │  ├──────┼──────┼──────┼──────┼──────┼──────┼──────┼──────┼──┤│
│  ──────────  │  │Text  │Text  │Text  │Text  │Text  │Text  │Text  │[TBD] │⚙ ││
│  Menu Sect.  │  │Text  │Text  │Text  │Text  │Text  │Text  │Text  │[TBD] │⚙ ││
│  > Menu 1    │  │Text  │Text  │Text  │Text  │Text  │Text  │Text  │[TBD] │⚙ ││
│  > Menu 2    │  │Text  │Text  │Text  │Text  │Text  │Text  │Text  │[TBD] │⚙ ││
│  > Menu 3    │  │Text  │Text  │Text  │Text  │Text  │Text  │Text  │[TBD] │⚙ ││
│              │  │Text  │Text  │Text  │Text  │Text  │Text  │Text  │[TBD] │⚙ ││
│              │  └──────┴──────┴──────┴──────┴──────┴──────┴──────┴──────┴──┘│
│              │                                                             │
│              │  Rows per page [10▼]          Page 1 of 10  1…2 [3] 4…5 >  │
└──────────────┴──────────────────────────────────────────────────────────────┘
Sidebar: 200px │ Content: 1694px
```

#### Component Inventory

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Page title "User List" | `3048:6058` | Page heading (Geist Semibold 24px) | Designed |
| Search input | `3048:6062` | Search by name, email, phone number — 320px | Designed |
| Filter: Department | `3048:6063` | Filter by department (dropdown/chip) | Designed |
| Filter: Position | `3048:6064` | Filter by position (dropdown/chip) | Designed |
| Filter: System Role | `3048:6065` | Filter by system role (dropdown/chip) | Designed |
| Filter: Status | `3048:6066` | Filter by user status — shows count "(2)" | Designed |
| Reset button | `3048:6067` | Clear all filters | Designed |
| Export button | `3048:6069` | Export current filtered list | Designed |
| + Add New button | `3048:6070` | Create new user — primary button | Designed |
| Table: First Name | `3048:6072` | User first name — ~184px | Placeholder data |
| Table: Last Name | `3048:6073` | User last name — ~184px | Placeholder data |
| Table: Department | `3048:6074` | Assigned department — ~184px | Placeholder data |
| Table: Position | `3048:6075` | Assigned position — ~184px | Placeholder data |
| Table: System Role | `3048:6076` | Assigned role — ~184px | Placeholder data |
| Table: Email | `3048:6077` | User email address — ~184px | Placeholder data |
| Table: Phone Number | `3048:6078` | User phone number — ~184px | Placeholder data |
| Table: Status | `3048:6079` | Active/inactive status badge ("TBD") — ~184px | Placeholder |
| Table: Action | `3048:6080` | Per-row gear icon — ~184px | Designed |
| Pagination | `3048:6081` | Rows per page + page navigation | Designed |

#### Design Constraints

- **Sidebar width:** 200px fixed left
- **Content area width:** 1694px
- **Table columns:** 9 columns, each ~184px (equal distribution)
- **Search input width:** 320px, left-aligned
- **Filter chips:** 4 filter buttons inline after search (Department, Position, System Role, Status)
- **Reset button:** After filter chips, with refresh icon
- **Action bar right:** Export + Add New buttons right-aligned
- **Status badge:** Colored badge ("TBD" placeholder) — actual status values unknown
- **Pagination:** Rows per page selector (default 10) + numbered page navigation

#### Key Differences from Other List Screens

| Aspect | Department/Position/Skill | User List |
|--------|--------------------------|-----------|
| Table columns | 2–4 columns | **9 columns** |
| Search scope | Name only | **Name, email, phone number** |
| Filters | None | **4 filter chips** (Department, Position, System Role, Status) |
| Filter count badge | N/A | **Status shows count "(2)"** |
| Reset button | N/A | **Present** — clears all filters |
| Export button | Dept/Position: yes, Role/Skill: no | **Yes** |
| Status column | None (or no active/inactive state) | **Status badge per row** |

#### Design Tokens Referenced

| Token | Value | Used For |
|-------|-------|----------|
| `general/primary` | `#171717` | Add New button background |
| `general/primary foreground` | `#fafafa` | Add New button text |
| `general/background` | `#ffffff` | Page + table background |
| `general/border` | `#e5e5e5` | Table borders, dividers |
| `general/muted` | `#f5f5f5` | Table header background |
| `general/muted foreground` | `#737373` | Placeholder text, secondary labels |
| `general/foreground` | `#0a0a0a` | Primary text |
| `text/text-foreground` | `#09090b` | Table cell text |
| `heading 3` | Geist Semibold 24px / 28.8px LH / -1 LS | Page title |
| `rounded-md` | 6px | Button, input border-radius |
| Font family | Geist | All text |

#### Gaps Identified from Design — User List

| Gap | Business Impact | Recommendation |
|-----|-----------------|----------------|
| Status badge values unknown ("TBD") | Cannot define status states | PO: confirm status values (Active, Inactive, Suspended?) and colors |
| Filter chip behavior unknown | Cannot define filter UX | PO: are filters dropdowns, multi-select, or single-select? |
| Status filter shows "(2)" | Unclear what count represents | PO: is this the number of active filter values or matching results? |
| Gear icon actions undefined | Cannot define per-row actions | PO: confirm Edit + Deactivate/Activate? Or Edit + Delete? |
| Search scope unclear | "Search by name, email, phone number..." — is this all 3 fields? | PO: confirm multi-field search behavior |
| Export format undefined | Cannot implement export | PO: CSV or Excel? What columns included? |
| Breadcrumb placeholder | Navigation path unclear | Expected: Users Management / Users |
| ~~Create/Edit User screens not designed~~ | ~~Cannot define form layout~~ | ✅ Create User extracted 2026-03-24 (node `3050:3782`). Edit User still pending. |
| Empty state not designed | No guidance when no users exist | Design Team: provide empty state |

### Screen: Create A New User

> Extracted from Figma on 2026-03-24.
> Source: [Create A New User](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3050-3782) (node `3050:3782`)

#### Layout Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb                   [Cancel]    [Save] │
├──────────────┬──────────────────────────────────────────────────────────────┤
│  [Sidebar]   │  Create A New User                                          │
│              │                                                             │
│              │  ┌─────────────────────────┐  ┌─────────────────────────┐   │
│              │  │ Personal Information    │  │ User Account            │   │
│              │  │                         │  │                         │   │
│              │  │  ┌───────────────────┐  │  │ * User Role       [▾]  │   │
│              │  │  │  Upload Avatar    │  │  │ * Account status  [⊙]  │   │
│              │  │  │  .png/.jpeg/...   │  │  │   Activate/Deactivate  │   │
│              │  │  └───────────────────┘  │  └─────────────────────────┘   │
│              │  │                         │                                │
│              │  │ *First name  *Last name │  ┌─────────────────────────┐   │
│              │  │ *Email     *Phone number│  │ Employee Profile        │   │
│              │  │ *Date of birth *Gender  │  │                         │   │
│              │  │  Address                │  │ *Department  *Position  │   │
│              │  └─────────────────────────┘  │ *Experience from (year) │   │
│              │                               │  CV/Resumé [Choose File]│   │
│              │                               │  Skills [+ Add skills]  │   │
│              │                               └─────────────────────────┘   │
└──────────────┴──────────────────────────────────────────────────────────────┘
```

#### Component Inventory — Create User Form

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Page title "Create A New User" | `3050:3829` | Context heading (Geist Semibold 24px) | Designed |
| Cancel button | `3057:1591` | Discard and return to User List | Designed |
| Save button (primary) | `3057:1607` | Submit form | Designed |
| **Personal Information card** | `3050:4203` | Left column — personal data (600×615px) | Designed |
| Upload Avatar area | `3050:4206` | Image upload — accepts .png/.jpeg/.jpg/.webp | Designed |
| * First name field | `3050:4399` | Mandatory text input (278px, left half) | Designed |
| * Last name field | `3050:4428` | Mandatory text input (278px, right half) | Designed |
| * Email field | `3050:4476` | Mandatory text input (278px, left half) | Designed |
| * Phone number field | `3050:4477` | Mandatory text input (278px, right half) | Designed |
| * Date of birth field | `3050:4504` | Mandatory date picker (278px, left half) | Designed |
| * Gender field | `3050:4505` | Mandatory dropdown (278px, right half) | Designed |
| Address field | `3050:4221` | Optional text input (576px, full width) | Designed |
| **User Account card** | `3055:1143` | Right column top — account settings (600×200px) | Designed |
| * User Role dropdown | `3055:1327` | Mandatory — select from available roles (576px) | Designed |
| * Account status toggle | `3055:1320` | Mandatory switch — Activate/Deactivate | Designed |
| **Employee Profile card** | `3055:1353` | Right column bottom — work assignment (600×380px) | Designed |
| * Department dropdown | `3055:1373` | Mandatory — select department (278px, left half) | Designed |
| * Position dropdown | `3055:1483` | Mandatory — select position (278px, right half) | Designed |
| * Experience from (year) | `3055:1499` | Mandatory — year input (576px) | Designed |
| CV/Resumé file input | `3089:2545` | Optional file upload (576px) | Designed |
| Skills "+ Add skills" button | `3055:1545` | Optional — opens skill selector | Designed |

#### Design Constraints — Create User Form

- **Form style:** Full page (not modal), two-column layout
- **Left column:** Personal Information card (600×615px) at x=217
- **Right column:** User Account (600×200px) + Employee Profile (600×380px) at x=837
- **Gap between columns:** 20px
- **Field layout:** Side-by-side pairs (278px each with 20px gap) or full-width (576px)
- **Action bar:** Cancel + Save right-aligned in page header (no "Save & Create Another")
- **Avatar upload:** Centered in card, accepts .png/.jpeg/.jpg/.webp
- **Account status:** Toggle switch with "Activate/Deactivate" label
- **Skills:** "+ Add skills" button — opens a selector (mechanism TBD)

#### Key Differences from Other Create Forms

| Aspect | Dept/Position (1 field) | Skill (3 fields) | Role (name + permissions) | **User (15 fields)** |
|--------|------------------------|-------------------|---------------------------|---------------------|
| Layout | Single card, centered | Single card, centered | Single card + permissions card | **Two-column, 3 cards** |
| Fields | 1 mandatory | 1 mandatory + 2 optional | 1 mandatory + permission matrix | **11 mandatory + 4 optional** |
| Dropdowns | None | None | None (checkboxes) | **4** (Gender, Role, Dept, Position) |
| File uploads | None | 1 (icon) | None | **2** (Avatar + CV) |
| Toggle | None | None | None | **1** (Account status) |
| Multi-select | None | None | Permission checkboxes | **Skills** (+ Add skills) |

#### Gaps Identified from Design — Create User

| Gap | Business Impact | Recommendation |
|-----|-----------------|----------------|
| Avatar upload constraints unknown | Cannot validate | PO: max file size, dimensions, accepted formats beyond .png/.jpeg/.jpg/.webp? |
| CV/Resumé file constraints unknown | Cannot validate | PO: accepted formats (PDF, DOCX?), max file size |
| Skills selector mechanism unknown | Cannot define UX | PO: modal? dropdown? multi-select with search? |
| Gender dropdown options unknown | Cannot populate | PO: Male, Female, Other, Prefer not to say? |
| Experience from (year) format unknown | Cannot validate | PO: free text year? Year picker? Range allowed? |
| Account status default unknown | Cannot set initial state | PO: default Active on create? Or admin chooses? |
| Email uniqueness validation | Cannot define error | PO: server-side on submit? Real-time check? |
| Phone number format unknown | Cannot validate | PO: international format? Country code? Digits only? |
| No validation error states shown | Cannot define errors | Assume: inline below fields (consistent with other forms) |
| ~~Edit User screen not designed~~ | ~~Cannot define edit form~~ | ✅ Update Information extracted 2026-03-26 (node `3122:6323`) |

### Screen: Update Information (User Details Sub-Page)

> Extracted from Figma on 2026-03-26.
> Source: [Update Information](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3122-6323) (node `3122:6323`)

#### Layout Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb                            [Top Bar] │
├──────────────┬──────────────────────────────────────────────────────────────┤
│  [Sidebar]   │  ← User Details  > Henry Tran                              │
│              │                                                             │
│              │  ┌──────────────┐  ┌────────────────────────────────────┐   │
│              │  │ Overview     │  │ Personal Information               │   │
│              │  │ ✏ Update     │  │                                    │   │
│              │  │   Info  ◄   │  │  ┌──────────────────────────┐      │   │
│              │  │ 🔑 Change   │  │  │    Upload Avatar         │      │   │
│              │  │   User Role │  │  │    .png/.jpeg/.jpg/.webp  │      │   │
│              │  │ ✉ Change    │  │  └──────────────────────────┘      │   │
│              │  │   Email     │  │                                    │   │
│              │  │ 🔒 Reset    │  │  *First name      *Last name      │   │
│              │  │   Password  │  │  *Date of birth    *Phone number   │   │
│              │  │ 🔓 Activate │  │  *Gender                          │   │
│              │  │  /Deactivate│  │   Address                         │   │
│              │  │ ❌ Delete    │  └────────────────────────────────────┘   │
│              │  │   User      │                                           │
│              │  └──────────────┘  ┌────────────────────────────────────┐   │
│              │                    │ Work Profile                       │   │
│              │                    │                                    │   │
│              │                    │  *Department       *Position       │   │
│              │                    │  *Experience from (year)           │   │
│              │                    │   CV/Resumé [Choose File]          │   │
│              │                    │   Skills [+ Add skills]            │   │
│              │                    └────────────────────────────────────┘   │
│              │                                                             │
│              │  [          Save (full-width, primary)          ]           │
└──────────────┴──────────────────────────────────────────────────────────────┘
```

#### Component Inventory — Update Information

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Page title "User Details" | `3122:6383` | Context heading with back arrow + user name breadcrumb | Designed |
| Back arrow (ArrowLeft) | `3122:6382` | Return to User List | Designed |
| User name breadcrumb | `3122:6798` | Shows which user is being edited (e.g., "Henry Tran") | Designed |
| **Action panel (left)** | `3122:6609` | 7 navigation buttons for user management actions | Designed |
| Overview button | `3122:6610` | View read-only user profile | Designed |
| Update Information button (active) | `3122:6611` | Edit personal info + work profile (current screen) | Designed |
| Change User Role button | `3122:6789` | Change assigned role | Designed |
| Change Email button | `3122:6612` | Change user email | Designed |
| Reset Password button | `3122:6613` | Reset user password | Designed |
| Activate/Deactivate button | `3122:6614` | Toggle account status | Designed |
| Delete User button | `3122:6615` | Delete user account | Designed |
| **Personal Information card** | `3122:6641` | Single-column card (600px) — personal data | Designed |
| Upload Avatar area | `3122:6644` | Image upload — accepts .png/.jpeg/.jpg/.webp | Designed |
| * First name field | `3122:6650` | Mandatory text input (278px, left half) | Designed |
| * Last name field | `3122:6651` | Mandatory text input (278px, right half) | Designed |
| * Date of birth field | `3122:6745` | Mandatory date picker (278px, left half) | Designed |
| * Phone number field | `3122:6654` | Mandatory text input (278px, right half) | Designed |
| * Gender field | `3122:6657` | Mandatory dropdown (576px, full width) | Designed |
| Address field | `3122:6658` | Optional text input (576px, full width) | Designed |
| **Work Profile card** | `3122:6428` | Single-column card (600px) — work assignment | Designed |
| * Department dropdown | `3122:6449` | Mandatory — select department (278px, left half) | Designed |
| * Position dropdown | `3122:6450` | Mandatory — select position (278px, right half) | Designed |
| * Experience from (year) | `3122:6451` | Mandatory — year input (576px) | Designed |
| CV/Resumé file input | `3122:6454` | Optional file upload (576px) | Designed |
| Skills "+ Add skills" button | `3122:6457` | Optional — opens skill selector | Designed |
| Save button (primary, full-width) | `3122:6805` | Submit changes — 600px wide | Designed |

#### Key Differences: Update Information vs Create User

| Aspect | Create User | Update Information |
|--------|-------------|-------------------|
| Layout | Two-column (Personal + Account/Profile) | **Single-column** (Personal + Work Profile stacked) |
| Email field | Present (mandatory) | **Absent** (managed via "Change Email" action) |
| User Role | Present (mandatory dropdown) | **Absent** (managed via "Change User Role" action) |
| Account Status | Present (toggle) | **Absent** (managed via "Activate/Deactivate" action) |
| Buttons | Cancel + Save (right-aligned, header) | **Save only** (full-width, bottom of form) |
| Context | Standalone page | **Sub-page of User Details** (accessed via left action panel) |
| Data | Empty form | **Pre-filled** with existing user data |
| Page title | "Create A New User" | "User Details" with user name breadcrumb |

### Screen: Change User Role (User Details Sub-Page)

> Extracted from Figma on 2026-03-26.
> Source: [Change User Role](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-6810) (node `3123:6810`)

#### Layout Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb                            [Top Bar] │
├──────────────┬──────────────────────────────────────────────────────────────┤
│  [Sidebar]   │  ← User Details  > Henry Tran                              │
│              │                                                             │
│              │  ┌──────────────┐  ┌────────────────────────────────────┐   │
│              │  │ Overview     │  │ User Account                       │   │
│              │  │ ✏ Update     │  │                                    │   │
│              │  │   Info       │  │  Use this option to update the     │   │
│              │  │ 🔑 Change   │  │  user's role and permissions...     │   │
│              │  │   User Role◄│  │                                    │   │
│              │  │ ✉ Change    │  │  Please ensure the correct role    │   │
│              │  │   Email     │  │  is assigned...                     │   │
│              │  │ 🔒 Reset    │  │                                    │   │
│              │  │   Password  │  │  * User Role  [Select user role ▾] │   │
│              │  │ 🔓 Activate │  │                                    │   │
│              │  │  /Deactivate│  └────────────────────────────────────┘   │
│              │  │ ❌ Delete    │                                           │
│              │  │   User      │  [          Save (full-width)          ]   │
│              │  └──────────────┘                                           │
└──────────────┴──────────────────────────────────────────────────────────────┘
```

#### Component Inventory — Change User Role

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Page title "User Details" | `3123:6870` | Context heading with back arrow + user name breadcrumb | Designed |
| Back arrow (ArrowLeft) | `3123:6869` | Return to User List | Designed |
| User name breadcrumb | `3123:6873` | Shows which user (e.g., "Henry Tran") | Designed |
| **Action panel (left)** | `3123:6875` | 7 navigation buttons — "Change User Role" is active | Designed |
| **User Account card** | `3123:7103` | Single card (600×275px) — role change context | Designed |
| Descriptive text | `3123:7520` | Two paragraphs explaining role change impact and immediacy | Designed |
| * User Role dropdown | `3123:7107` | Mandatory dropdown (576px). Placeholder: "Select user role" | Designed |
| Save button (primary, full-width) | `3123:6915` | Submit role change — 600px wide | Designed |

#### Design Details

- **Descriptive text (paragraph 1):** "Use this option to update the user's role and permissions within the system. Selecting a new role will immediately apply the corresponding access rights and restrictions"
- **Descriptive text (paragraph 2):** "Please ensure the correct role is assigned, as this determines what the user can view and manage in the system. Changes take effect instantly and do not require any further action."
- **Single field only** — User Role dropdown, pre-filled with current role
- **Save button** — full-width, primary style, below the card
- **No Cancel button** — consistent with Update Information pattern

### Screen: Change Email (User Details Sub-Page)

> Extracted from Figma on 2026-03-26.
> Source: [Change Email](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-7139) (node `3123:7139`)

#### Layout Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb                            [Top Bar] │
├──────────────┬──────────────────────────────────────────────────────────────┤
│  [Sidebar]   │  ← User Details  > Henry Tran                              │
│              │                                                             │
│              │  ┌──────────────┐  ┌────────────────────────────────────┐   │
│              │  │ Overview     │  │ Change Email                       │   │
│              │  │ ✏ Update     │  │                                    │   │
│              │  │   Info       │  │  [Descriptive text — see note]     │   │
│              │  │ 🔑 Change   │  │                                    │   │
│              │  │   User Role │  │  * New Email  [Enter email     ]   │   │
│              │  │ ✉ Change    │  │                                    │   │
│              │  │   Email  ◄  │  └────────────────────────────────────┘   │
│              │  │ 🔒 Reset    │                                           │
│              │  │   Password  │  [          Save (full-width)          ]   │
│              │  │ 🔓 Activate │                                           │
│              │  │  /Deactivate│                                           │
│              │  │ ❌ Delete    │                                           │
│              │  │   User      │                                           │
│              │  └──────────────┘                                           │
└──────────────┴──────────────────────────────────────────────────────────────┘
```

#### Component Inventory — Change Email

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Card title "Change Email" | `3123:7215` | Section heading | Designed |
| Descriptive text | `3123:7518` | Context about email change — **⚠️ COPY ERROR: currently shows Reset Password text** | Design Error |
| * New Email field | `3123:7310` | Mandatory text input (576px). Placeholder: "Enter email" | Designed |
| Save button (primary, full-width) | `3123:7218` | Submit email change — 600px wide | Designed |

#### Design Error — Descriptive Text

The descriptive text in the Figma design is **incorrectly copied from the Reset Password screen**. It currently reads:
> "Use this option to reset a user's password. When triggered, the system will send a secure password reset link..."

This should be replaced with text describing email change behavior. **Design Team action required.**

### Screen: Reset Password (User Details Sub-Page)

> Extracted from Figma on 2026-03-26.
> Source: [Reset Password](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3123-7334) (node `3123:7334`)

#### Layout Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Breadcrumb / Breadcrumb / Breadcrumb                            [Top Bar] │
├──────────────┬──────────────────────────────────────────────────────────────┤
│  [Sidebar]   │  ← User Details  > Henry Tran                              │
│              │                                                             │
│              │  ┌──────────────┐  ┌────────────────────────────────────┐   │
│              │  │ Overview     │  │ Reset Password                     │   │
│              │  │ ✏ Update     │  │                                    │   │
│              │  │   Info       │  │  Use this option to reset a        │   │
│              │  │ 🔑 Change   │  │  user's password...                │   │
│              │  │   User Role │  │                                    │   │
│              │  │ ✉ Change    │  │  Please ensure the user's email    │   │
│              │  │   Email     │  │  address is correct...             │   │
│              │  │ 🔒 Reset    │  │                                    │   │
│              │  │   Password◄ │  │  📧 Email                          │   │
│              │  │ 🔓 Activate │  │     henry@exnodes.vn               │   │
│              │  │  /Deactivate│  └────────────────────────────────────┘   │
│              │  │ ❌ Delete    │                                           │
│              │  │   User      │  [       Reset Password (full-width)   ]  │
│              │  └──────────────┘                                           │
└──────────────┴──────────────────────────────────────────────────────────────┘
```

#### Component Inventory — Reset Password

| Component | Node ID | Business Purpose | Design Status |
|-----------|---------|-----------------|---------------|
| Card title "Reset Password" | `3123:7410` | Section heading | Designed |
| Descriptive text | `3123:7504` | Two paragraphs explaining reset link flow: time-limited, single-use, sent to registered email | Designed |
| Email display (read-only) | `3123:7511` | Shows user's current email with icon — admin verifies before triggering | Designed |
| Email label | `3123:7514` | "Email" label above the email value | Designed |
| Email value | `3123:7515` | "henry@exnodes.vn" — user's registered email | Designed |
| Reset Password button (primary, full-width) | `3123:7417` | Trigger password reset — 600px wide. Label: "Reset Password" (not "Save") | Designed |

#### Design Details

- **No input fields** — this is a pure action screen, unlike Change Role (dropdown) or Change Email (text input)
- **Button label:** "Reset Password" (not "Save" — action-specific label)
- **Descriptive text (paragraph 1):** "Use this option to reset a user's password. When triggered, the system will send a secure password reset link to the user's registered email address. The email will include a login URL that allows the user to access the reset page and create a new password."
- **Descriptive text (paragraph 2):** "Please ensure the user's email address is correct before proceeding. The reset link is time-limited for security purposes and can only be used once."
- **Email display:** Read-only with mail icon — shows the email where the reset link will be sent

---

**Document Version:** 1.0
**Last Updated:** 2026-03-26
**Author:** BA Agent
**Reviewer:** Pending
