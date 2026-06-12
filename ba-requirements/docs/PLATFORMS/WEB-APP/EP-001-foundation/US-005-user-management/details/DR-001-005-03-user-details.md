---
document_type: DETAIL_REQUIREMENT
platform: WEB-APP
platform_display: "Exnodes HRM"
epic_id: EP-001
story_id: US-005
story_name: "User Management"
detail_id: DR-001-005-03
detail_name: "User Details"
parent_requirement: FR-US-005-03
status: draft
version: "1.2"
created_date: 2026-03-26
last_updated: 2026-05-20
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
---

# Detail Requirement: User Details

**Detail ID:** DR-001-005-03
**Parent Requirement:** FR-US-005-03
**Story:** US-005-user-management
**Epic:** EP-001 (Foundation)
**Status:** Draft
**Version:** 1.2

---

## 1. Use Case Description

As a **user with user management permission**, I want to **view a user's extended profile including personal information, account status, work profile, line manager, direct reports, ID cards, emergency contacts, salary, and banking information**, so that **I can review all relevant user data, see who reports to whom, and access management actions (edit, role change, email change, password reset, activate/deactivate, delete) from a single page**.

**Purpose:** Provide a centralized read-only view of all user data — personal, account, employee profile (including reporting line), identification documents, emergency contacts, compensation, and payment details — in one place. The User Details page also surfaces the inverse organizational view: a Direct Reports card listing all users whose Line Manager = current user, computed live from the inverse relationship. The page is the hub for all per-user management actions, accessed via a left-panel action menu. Sensitive data (Salary, Banking) is permission-gated to ensure only authorized roles can see it.

**Target Users:**
- Any role with **user view permission** — can see the profile data (Personal Information, Work Profile, ID Cards, Emergency Contact) and Overview tab
- Any role with **`user.salary.view` permission** — can additionally see the Salary card
- Any role with **`user.banking.view` permission** — can additionally see the Banking card (with bank account number masked to last 4 digits)
- Any role with **user management permission** — can additionally see and use the action buttons (Update Information, Change User Role, Change Email, Reset Password, Activate/Deactivate, Delete User)

**Key Functionality:**
- Read-only display of all user data (no inline editing)
- Static cover image + user avatar + name + status badge + role display
- Seven info cards: Personal Information (9 fields), Work Profile (7 fields including Line Manager), Direct Reports (list of subordinates), ID Cards (4 fields), Emergency Contact (list), Salary (permission-gated, 2 fields), Banking (permission-gated, 4 fields with masked account number)
- Line Manager displayed as clickable name (with `(Inactive)` suffix if deactivated) — navigates to manager's User Details
- Direct Reports card lists all users whose Line Manager = current user, computed live from inverse relationship
- Left action panel with 7 navigation buttons for management actions
- Back arrow to return to User List

---

## 2. User Workflow

**Entry Point:** User List → click on a user row (any column except gear icon)

**Preconditions:**
- User is signed in (US-001)
- User has user view permission (US-004)
- The target user exists in the system

**Main Flow:**
1. User clicks on a user row in the User List
2. System navigates to User Details page for the selected user
3. System loads user data from API
4. Page displays with "Overview" tab active by default
5. Left panel shows action buttons (visibility based on permissions)
6. Right panel shows cover image, avatar, name, status badge, role, and two info cards
7. User reviews the profile data

**Alternative Flows:**
- **Alt 1 — Click action button:** User clicks one of the 6 management action buttons → navigates to the corresponding action view (separate DRs)
- **Alt 2 — Back to list:** User clicks ← back arrow → returns to User List (preserving any active search/filters)
- **Alt 3 — User not found:** If user record has been deleted since the list was loaded → show "User not found" message with link to return to User List
- **Alt 4 — API error:** System fails to load user data → show error state with retry option; left panel remains visible

**Exit Points:**
- **Back arrow** → returns to User List (preserving search/filter state)
- **Action button** → navigates to the selected action view
- **Sidebar navigation** → navigates away from User Details entirely

---

## 3. Field Definitions

### Input Fields

No input fields — this page is entirely read-only.

### Interaction Elements

| Element Name | Type | Position | Visible To | Trigger Action | Description |
|--------------|------|----------|-----------|----------------|-------------|
| ← Back arrow | Icon button | Title bar, left of "User Details" | All with view permission | Return to User List (preserved state) | ArrowLeft icon (24×24px) |
| Overview | Button (active) | Left panel, position 1 | All with view permission | Shows Overview tab (default, current view) | Active/highlighted state |
| Update Information | Button | Left panel, position 2 | Management permission only | Navigates to edit user form | Planned — separate DR |
| Change User Role | Button | Left panel, position 3 | Management permission only | Navigates to role change view | Planned — separate DR |
| Change Email | Button | Left panel, position 4 | Management permission only | Navigates to email change view | Planned — separate DR |
| Reset Password | Button | Left panel, position 5 | Management permission only | Triggers password reset flow | Planned — separate DR |
| Activate/Deactivate | Button | Left panel, position 6 | Management permission only | Triggers status change flow | Context-sensitive label — separate DR |
| Delete User | Button (danger) | Left panel, position 7 | Management permission only | Triggers delete confirmation | Danger style — separate DR |
| Line Manager link | Clickable text | Work Profile card | All with `user.view` | Navigate to manager's User Details | Resolves to manager's full name; suffixed `(Inactive)` if deactivated |
| Direct Report row | Clickable list item | Direct Reports card | All with `user.view` | Navigate to that subordinate's User Details | Shows avatar, name, position, dept, status badge |

---

## 4. Data Display

### Profile Header

| Data Name | Data Type | Format | Empty State | Business Meaning |
|-----------|-----------|--------|-------------|------------------|
| Cover image | Static banner | 1445×150px, decorative | Always shown (system-provided) | Static image — same for all users, not configurable |
| Avatar | Circular image | 132×132px, overlapping cover by ~66px | Default placeholder (initials or silhouette) | User's profile photo |
| Full name | Text | Geist Semibold 30px (heading 2) | — (always populated) | First + Last name combined |
| Status badge | Colored badge | Next to name | — (always has a status) | "Activated" (green `#22c55e`) or "Deactivated" (gray) |
| Role | Text with icon | Below name, muted foreground `#737373` | — (always assigned) | User's system role (e.g., "Super Admin") |

### Personal Information Card

| Data Name | Icon | Format | Empty State | Business Meaning |
|-----------|------|--------|-------------|-----------------|
| Full name | User icon | Label (muted 12px) + value (regular 14px) | — | First + Last name combined |
| Email | Mail icon | Label + value | — | User's login email |
| Phone number | Phone icon | Label + value | "—" | Contact number |
| Date of birth | Calendar icon | Label + value (DD/MM/YYYY) | "—" | Birth date |
| Gender | Gender icon | Label + value | "—" | Male, Female, Other, etc. |
| Marital status | User icon | Label + value | "—" | Single, Married, Other |
| Nationality | Globe icon | Label + value | — (always populated — mandatory) | User's nationality (ISO country) |
| Permanent address | Location icon | Label + value (spans wider column) | "—" | Permanent address text |
| Temporary address | Location icon | Label + value (spans wider column) | "—" | Temporary/current address text |

### Work Profile Card

| Data Name | Icon | Format | Empty State | Business Meaning |
|-----------|------|--------|-------------|-----------------|
| Department | Building icon | Label + value | "—" (unassigned) | Organizational department |
| Position | Briefcase icon | Label + value | "—" (unassigned) | Job position |
| Experience from | Calendar icon | Label + value (4-digit year) | "—" | Career start year |
| Line Manager | User-Gear icon | Clickable name — links to manager's User Details page; `(Inactive)` suffix appended if manager is deactivated | "—" (no manager assigned) | Who this user reports to |
| Education level | Graduation icon | Label + value | — (always populated — mandatory) | Highest level: High school / College / Bachelor's / Master's / Doctorate |
| Skills | Star icon | Label + value (comma-separated) | "—" (no skills assigned) | Competency tags (read-only text) |
| CV/Resumé | File icon | Label + clickable link/filename | "—" (no file uploaded) | Click to download or open in new tab |

### Direct Reports Card

| Data Name | Icon | Format | Empty State | Business Meaning |
|-----------|------|--------|-------------|-----------------|
| Avatar | — | Circular image (small) per row | Default placeholder if no photo | Subordinate's profile photo |
| Full name | User icon | Clickable text per row — navigates to that subordinate's User Details | "No direct reports" (when list is empty) | Subordinate's full name |
| Position | Briefcase icon | Label + value per row | "—" (unassigned) | Subordinate's job position |
| Department | Building icon | Label + value per row | "—" (unassigned) | Subordinate's organizational department |
| Status badge | Colored badge | "Activated" (green) or "Deactivated" (gray) per row | — (always has a status) | Subordinate's current account status |

> **Note:** Direct Reports card is visible to any user with `user.view`. The list is computed live from the inverse Line Manager relationship at query time (no separately stored list). Rows are ordered alphabetically by full name. No pagination in v1 — all subordinates are shown in full. When the user has zero subordinates, the card shows "No direct reports" message. Card is read-only — no editing, no add/remove (changes happen via Update Information on individual user profiles).

### ID Cards Card

| Data Name | Icon | Format | Empty State | Business Meaning |
|-----------|------|--------|-------------|-----------------|
| Front image | Image icon | Clickable thumbnail (opens larger view in modal or new tab) | "—" (no image uploaded) | Front of user's ID card |
| Back image | Image icon | Clickable thumbnail (opens larger view in modal or new tab) | "—" (no image uploaded) | Back of user's ID card |
| ID number | Card icon | Label + value (plain text) | "—" | User's ID/citizen card number |
| Issue date | Calendar icon | Label + value (DD/MM/YYYY) | "—" | Date the ID was issued |

> **Note:** The entire ID Cards section is always shown. Empty fields display "—" individually — there is no full-section empty state.

### Emergency Contact Card

| Data Name | Icon | Format | Empty State | Business Meaning |
|-----------|------|--------|-------------|-----------------|
| Full name | User icon | Label + value per contact row | "No emergency contacts" (when list is empty) | Emergency contact's full name |
| Relationship | User icon | Label + value per contact row | — (omitted if list empty) | Relationship to the user (e.g., Spouse, Parent) |
| Phone number | Phone icon | Label + value per contact row | — (omitted if list empty) | Emergency contact's phone number |

> **Note:** All emergency contacts (0..N) are displayed in a list — no pagination. When the list is empty, the card shows the message "No emergency contacts" in place of contact rows.

### Salary Card (permission-gated)

| Data Name | Icon | Format | Empty State | Business Meaning |
|-----------|------|--------|-------------|-----------------|
| Base salary | Money icon | Label + value (VND, thousand separators + "₫" or "VND" suffix, e.g., "15,000,000 ₫") | "—" | User's base monthly salary |
| Insurance salary | Money icon | Label + value (VND, thousand separators + "₫" or "VND" suffix) | "—" | Salary used to calculate insurance contributions |

> **Note:** The entire Salary card is **hidden** (not rendered at all) for viewers who lack the `user.salary.view` permission — it does not appear as an empty card.

### Banking Card (permission-gated)

| Data Name | Icon | Format | Empty State | Business Meaning |
|-----------|------|--------|-------------|-----------------|
| Bank name | Bank icon | Label + value | "—" | Name of the user's bank (Vietnamese bank list) |
| Bank account number | Card icon | Label + masked value (last 4 digits only, e.g., "•••• 1234") | "—" | Bank account number — always masked, full number never displayed |
| Account name | User icon | Label + value | "—" | Name on the bank account |
| Transfer method | Money icon | Label + value | "—" | Bank transfer or Cash |

> **Note:** The entire Banking card is **hidden** (not rendered at all) for viewers who lack the `user.banking.view` permission — it does not appear as an empty card. The bank account number is **always masked** to last 4 digits for all viewers, regardless of permission level.

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page first opens, API fetching | Skeleton placeholders for avatar, name, cards |
| Populated | User data loaded successfully | Full profile with all visible cards displayed |
| Partial data | Some optional fields empty | "—" shown for empty optional fields; layout unchanged; new fields (Marital status, Permanent address, Temporary address, ID number, Issue date, salary fields, banking fields) render with "—" when empty |
| Error | API fails to load user data | Error message with retry button; left panel still visible |
| User not found | User deleted since list was loaded | "User not found" message with link to return to User List |
| View-only mode | User has view permission only | Overview button visible; 6 management action buttons hidden |
| Management mode | User has management permission | All 7 left-panel buttons visible |
| Salary card hidden | Viewer lacks `user.salary.view` permission | Salary card not rendered at all — layout adjusts; no empty placeholder card |
| Salary card visible | Viewer has `user.salary.view` permission | Salary card displayed; empty fields show "—"; populated fields show VND with thousand separators |
| Banking card hidden | Viewer lacks `user.banking.view` permission | Banking card not rendered at all — layout adjusts; no empty placeholder card |
| Banking card visible | Viewer has `user.banking.view` permission | Banking card displayed; bank account number always masked to last 4 digits (e.g., "•••• 1234"); empty fields show "—" |
| Emergency Contact empty | User has 0 emergency contacts | Card shows "No emergency contacts" message in place of contact rows |
| ID card images empty | No ID front/back uploaded | "—" shown in place of thumbnail; section still rendered |
| ID card thumbnail click | User clicks present front/back thumbnail | Opens larger image view in modal or new tab |
| Line manager assigned (active) | User has a manager whose account is active | Manager's full name shown as clickable link in Work Profile |
| Line manager assigned (inactive) | User has a manager whose account is deactivated | Manager's full name shown with "(Inactive)" suffix; still clickable |
| No line manager | User has no manager assigned (top of hierarchy) | "—" shown in Line Manager row |
| Direct reports populated | User has 1+ subordinates whose Line Manager = current user | Direct Reports card lists all subordinates (avatar, name, position, dept, status badge), ordered alphabetically |
| Direct reports empty | User has 0 subordinates | Direct Reports card shows "No direct reports" centered in card body |

### Page Layout (Design Reference)

```
┌──────────────────────────────────────────────────────────────────┐
│ Breadcrumb / Breadcrumb / Breadcrumb                  [Top Bar]  │
├──────────┬───────────────────────────────────────────────────────┤
│ [Sidebar]│ ← User Details > Henry Tran                          │
│  200px   │                                                       │
│          │ ┌─────────────┬──────────────────────────────────────┐│
│          │ │ [Overview]  │ ┌──────────────────────────────────┐ ││
│          │ │ Update Info │ │ ████████ Cover Image ████████████ │ ││
│          │ │ Change Role │ │ ┌────┐                           │ ││
│          │ │ Change Email│ │ │Avtr│ Henry Tran  [Activated]   │ ││
│          │ │ Reset Pass  │ │ └────┘ 🔑 Super Admin            │ ││
│          │ │ Act/Deact   │ │                                  │ ││
│          │ │ Delete User │ │ Personal Information              │ ││
│          │ │             │ │ Name | Email | Phone | DOB        │ ││
│          │ │  189px      │ │ Gender | Marital | Nationality    │ ││
│          │ │             │ │ Permanent address | Temp address  │ ││
│          │ │             │ │                                  │ ││
│          │ │             │ │ Work Profile                      │ ││
│          │ │             │ │ Dept | Position | Experience      │ ││
│          │ │             │ │ Line Manager (clickable link)     │ ││
│          │ │             │ │ Education level | Skills          │ ││
│          │ │             │ │ CV/Resumé                         │ ││
│          │ │             │ │                                  │ ││
│          │ │             │ │ Direct Reports                    │ ││
│          │ │             │ │ [Avatar] Name | Position | Dept   │ ││
│          │ │             │ │ [Avatar] Name | Position | Dept   │ ││
│          │ │             │ │ (or "No direct reports")          │ ││
│          │ │             │ │                                  │ ││
│          │ │             │ │ ID Cards                          │ ││
│          │ │             │ │ [Front img] [Back img]            │ ││
│          │ │             │ │ ID number | Issue date            │ ││
│          │ │             │ │                                  │ ││
│          │ │             │ │ Emergency Contact                 │ ││
│          │ │             │ │ Name | Relationship | Phone       │ ││
│          │ │             │ │ (or "No emergency contacts")      │ ││
│          │ │             │ │                                  │ ││
│          │ │             │ │ Salary (if user.salary.view)      │ ││
│          │ │             │ │ Base salary | Insurance salary    │ ││
│          │ │             │ │                                  │ ││
│          │ │             │ │ Banking (if user.banking.view)    │ ││
│          │ │             │ │ Bank | •••• 1234 | Account name   │ ││
│          │ │             │ │ Transfer method                   │ ││
│          │ │             │ └──────────────────────────────────┘ ││
│          │ └─────────────┴──────────────────────────────────────┘│
└──────────┴───────────────────────────────────────────────────────┘
```

> **Note:** Figma design available at node `3120:5348`. See ANALYSIS.md Design Context [ADD-ON] for full component inventory.

---

## 5. Acceptance Criteria

**Definition of Done — All criteria must be met:**

**Page Display:**
- **AC-01:** User Details page displays with title "User Details" + ← back arrow + "> [User full name]" breadcrumb
- **AC-02:** Page has two-panel layout: left action panel (189px) + right content panel (1445px)
- **AC-03:** Overview tab is active by default when page loads
- **AC-04:** Static cover image banner (1445×150px) is displayed at the top — same image for all users

**Profile Header:**
- **AC-05:** User's avatar displays as a circular image (132×132px) overlapping the cover image
- **AC-06:** If no avatar uploaded, a default placeholder is shown (initials or silhouette)
- **AC-07:** User's full name displays in Geist Semibold 30px next to the avatar
- **AC-08:** Status badge displays next to the name — "Activated" (green) or "Deactivated" (gray)
- **AC-09:** User's system role displays below the name with an icon prefix (e.g., "Super Admin")

**Personal Information Card:**
- **AC-10:** Personal Information card displays 9 fields: Full name, Email, Phone number, Date of birth, Gender, Marital status, Nationality, Permanent address, Temporary address
- **AC-11:** Each field displays with an icon prefix, label (muted text), and value (body text)
- **AC-12:** Permanent address and Temporary address each span a wider column for full address text
- **AC-13:** Nationality is always displayed with a value (mandatory field — never shows "—")
- **AC-14:** Empty optional fields (Phone number, Date of birth, Gender, Marital status, Permanent address, Temporary address) display "—" as placeholder value

**Work Profile Card:**
- **AC-15:** Work Profile card displays 7 fields: Department, Position, Experience from, Line Manager, Education level, Skills, CV/Resumé
- **AC-16:** Education level is always displayed with a value (mandatory field — never shows "—")
- **AC-17:** Skills displays as comma-separated text (e.g., "React JS, Wordpress")
- **AC-18:** CV/Resumé displays as a clickable link/filename — clicking downloads or opens the file
- **AC-19:** Empty optional fields (Skills, CV/Resumé) display "—" as placeholder value

**Left Action Panel:**
- **AC-20:** Left panel displays 7 stacked buttons: Overview, Update Information, Change User Role, Change Email, Reset Password, Activate/Deactivate, Delete User
- **AC-21:** Overview button shows active/highlighted state when on the Overview tab
- **AC-22:** Action buttons (positions 2–7) are visible only to users with management permission
- **AC-23:** Users with view-only permission see only the Overview button in the left panel
- **AC-24:** Each action button navigates to its respective action view (separate DRs — future)
- **AC-25:** Delete User button uses a distinct/danger style to signal destructive action

**Navigation:**
- **AC-26:** Clicking ← back arrow returns to the User List, preserving any active search and filter state
- **AC-27:** Breadcrumb in title bar shows "> [User full name]" to identify the current user
- **AC-28:** Page URL includes the user identifier so it can be bookmarked or shared

**Loading & Error States:**
- **AC-29:** Skeleton placeholders are shown while user data is loading
- **AC-30:** If API fails to load user data, an error message with retry button is displayed; left panel remains visible
- **AC-31:** If the user has been deleted since the list was loaded, "User not found" is displayed with a link to return to User List

**Access Control:**
- **AC-32:** User Details page is accessible only to users with user view permission
- **AC-33:** Direct URL access by unauthorized users redirects to an appropriate fallback page

**ID Cards Card:**
- **AC-34:** ID Cards card displays 4 fields: Front image, Back image, ID number, Issue date
- **AC-35:** Front image and Back image display as clickable thumbnails when present — clicking opens a larger view (modal or new tab)
- **AC-36:** When Front or Back image is not uploaded, "—" is displayed in place of the thumbnail
- **AC-37:** ID number displays as plain text (no format mask); empty value shows "—"
- **AC-38:** Issue date displays in DD/MM/YYYY format; empty value shows "—"
- **AC-39:** The entire ID Cards card is always rendered — empty fields display "—" individually, no full-section empty state

**Emergency Contact Card:**
- **AC-40:** Emergency Contact card displays a list of contacts; each row shows Full name, Relationship, and Phone number with icon prefixes
- **AC-41:** All emergency contacts are displayed (no pagination — list is shown in full)
- **AC-42:** When the user has zero emergency contacts, the card shows the message "No emergency contacts" in place of contact rows

**Salary Card (permission-gated):**
- **AC-43:** Salary card is rendered only when the viewer has the `user.salary.view` permission
- **AC-44:** Viewers without `user.salary.view` permission do not see the Salary card at all (not rendered as an empty card)
- **AC-45:** When visible, Salary card displays 2 fields: Base salary and Insurance salary
- **AC-46:** Salary values are formatted as VND numeric with thousand separators and "₫" or "VND" suffix (e.g., "15,000,000 ₫")
- **AC-47:** Empty salary fields display "—"

**Banking Card (permission-gated):**
- **AC-48:** Banking card is rendered only when the viewer has the `user.banking.view` permission
- **AC-49:** Viewers without `user.banking.view` permission do not see the Banking card at all (not rendered as an empty card)
- **AC-50:** When visible, Banking card displays 4 fields: Bank name, Bank account number, Account name, Transfer method
- **AC-51:** Bank account number is always masked to last 4 digits (e.g., "•••• 1234") for all viewers — the full account number is never displayed on this page
- **AC-52:** Empty banking fields display "—"

**Line Manager (Work Profile):**
- **AC-53:** Line Manager appears as a row in the Work Profile card
- **AC-54:** When a line manager is assigned and active, the manager's full name is displayed as a clickable link that navigates to that manager's User Details page
- **AC-55:** When the assigned line manager's account is deactivated, the manager's name is displayed with an `(Inactive)` suffix (e.g., "Sarah Le (Inactive)") and remains clickable
- **AC-56:** When the user has no line manager assigned (top of hierarchy), the Line Manager row displays "—"
- **AC-57:** Line Manager row is visible to any user with `user.view` permission (same as base profile)

**Direct Reports Card:**
- **AC-58:** A Direct Reports card is displayed stacked below the Work Profile card
- **AC-59:** Direct Reports card is visible to any user with `user.view` permission
- **AC-60:** Direct Reports card lists all users whose Line Manager = current user (computed live from the inverse relationship)
- **AC-61:** Each direct report row shows: Avatar, Full name, Position, Department, and an Activated/Deactivated status badge
- **AC-62:** Clicking a direct report row navigates to that subordinate's User Details page
- **AC-63:** Direct Reports list is ordered alphabetically by full name
- **AC-64:** When the user has zero direct reports, the card displays "No direct reports" centered in the card body
- **AC-65:** Direct Reports card is read-only — no add/remove/edit controls (changes happen via Update Information on individual user profiles)

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — full data | Click user with all fields populated (incl. ID card, emergency contacts, salary, banking) | All visible cards shown with values, status badge green | High |
| Partial data | Click user with no phone, addresses, marital status, skills, CV, ID number, ID images | All new and existing optional fields render with "—"; mandatory fields (Nationality, Education level) still show values | High |
| Active user | Click active user | "Activated" green badge, "Deactivate" button in panel | High |
| Inactive user | Click inactive user | "Deactivated" gray badge, "Activate" button in panel | High |
| Back to list | Click ← back arrow | Returns to User List with preserved filters | High |
| View-only permission | User with view-only role | Only Overview button visible; 6 action buttons hidden | High |
| Management permission | User with management role | All 7 action buttons visible | High |
| Loading state | Page first opens | Skeleton placeholders shown | Medium |
| API error | Backend unavailable | Error message with retry; left panel visible | Medium |
| User deleted | User removed since list loaded | "User not found" message with return link | Medium |
| No avatar | User has no profile photo | Default placeholder (initials or silhouette) | Medium |
| CV download | Click CV/Resumé link | File downloads or opens in new tab | Medium |
| Unauthorized access | User without view permission | Redirect / access denied | High |
| Salary hidden without permission | Viewer without `user.salary.view` opens any user's profile | Salary card not rendered at all — no empty placeholder card visible | High |
| Salary visible with permission | Viewer with `user.salary.view` opens user with populated salary | Salary card displays Base salary and Insurance salary in VND with thousand separators | High |
| Banking hidden without permission | Viewer without `user.banking.view` opens any user's profile | Banking card not rendered at all — no empty placeholder card visible | High |
| Banking account masking visible | Viewer with `user.banking.view` opens user with populated banking | Banking card visible; bank account number shown as masked (e.g., "•••• 1234") — full number never appears | High |
| Emergency Contact populated | User has 2 emergency contacts | Both contacts displayed in list with name, relationship, and phone | High |
| Emergency Contact empty | User has 0 emergency contacts | Card shows "No emergency contacts" message | High |
| ID card thumbnail click | Click present front or back ID image thumbnail | Opens larger image view in modal or new tab | Medium |
| ID card no images | User has not uploaded front/back images | "—" shown in place of each thumbnail; ID Cards card still rendered | Medium |
| Nationality always rendered | Any user (all have nationality — mandatory) | Nationality field always shows a value, never "—" | High |
| Education level always rendered | Any user (all have education level — mandatory) | Education level field always shows a value, never "—" | High |
| Partial data — new fields | User with empty Marital status, Temporary address, ID number, salary, banking values (where visible) | Each empty optional new field displays "—" | High |
| Line manager — active | User with an active line manager | Manager's name shown as clickable link; clicking navigates to manager's User Details | High |
| Line manager — inactive | User with a deactivated line manager | Manager name displayed with "(Inactive)" suffix; link remains clickable and navigates correctly | High |
| Line manager — none | Top-of-hierarchy user (no manager) | Line Manager row displays "—" | High |
| Direct reports — populated | User with 3 active direct reports | 3 rows displayed in Direct Reports card, alphabetically ordered, each with avatar + name + position + dept + green Activated badge | High |
| Direct reports — single inactive | User with 1 deactivated direct report | 1 row displayed with gray Deactivated badge | Medium |
| Direct reports — empty | User with 0 direct reports | "No direct reports" empty state shown in card body | High |
| Direct report row click | Click any direct report row | Navigates to that subordinate's User Details page | High |
| Direct reports — visible to all | Viewer with only base `user.view` (no management permission) | Direct Reports card is displayed | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **SR-01:** Access to the User Details page is controlled by view permission configured in US-004 — users without view permission cannot access this page
- **SR-02:** The 6 action buttons (Update Information, Change User Role, Change Email, Reset Password, Activate/Deactivate, Delete User) are only displayed to roles granted user management permission via US-004
- **SR-03:** Users with view-only permission see only the Overview button in the left panel — all management actions are hidden both visually and at the API level
- **SR-04:** The cover image is static and system-provided — not user-specific, not configurable
- **SR-05:** User data is fetched from the server on page load — data is not cached from the User List
- **SR-06:** If the user record no longer exists when the page loads (deleted between list click and page load), the system returns a "User not found" state
- **SR-07:** The status badge reflects the current account status at the time of page load: "Activated" (green) or "Deactivated" (gray)
- **SR-08:** The Activate/Deactivate button label is context-sensitive — shows "Deactivate" for active users and "Activate" for inactive users (consistent with gear icon behavior on User List)
- **SR-09:** CV/Resumé is displayed as a clickable link that triggers a file download or opens in a new tab — the file is served from the stored upload location
- **SR-10:** Skills are displayed as comma-separated plain text (not interactive chips — read-only on this page)
- **SR-11:** Empty optional fields (Phone number, Permanent address, Temporary address, Marital status, Skills, CV/Resumé, ID number, Issue date, salary fields, banking fields) display "—" — they are never hidden, only show placeholder value
- **SR-12:** The back arrow navigation preserves User List state (active search query, filter selections, current page, rows per page) — the list does not reload from scratch
- **SR-13:** The "Address" field has been renamed to "Permanent address"; a separate "Temporary address" field is displayed alongside it
- **SR-14:** Nationality is a mandatory field and is always displayed with a value (never "—")
- **SR-15:** Education level is a mandatory field and is always displayed with a value (never "—")
- **SR-16:** ID card front and back images are displayed as clickable thumbnails when present; clicking opens a larger image view (modal or new tab)
- **SR-17:** Emergency Contact card displays the message "No emergency contacts" when the user has zero contacts; otherwise displays the full list of contacts
- **SR-18:** Salary card visibility is gated by the `user.salary.view` permission — the entire card is not rendered when the viewer lacks this permission (not shown as an empty card)
- **SR-19:** Banking card visibility is gated by the `user.banking.view` permission — the entire card is not rendered when the viewer lacks this permission (not shown as an empty card)
- **SR-20:** Bank account number is always masked to last 4 digits (e.g., "•••• 1234") for all viewers with `user.banking.view` permission — the full account number is never displayed on this page, regardless of permission level
- **SR-21:** Salary values are formatted as VND numeric with thousand separators and "₫" or "VND" suffix (e.g., "15,000,000 ₫")
- **SR-22:** Permission gating for Salary and Banking is enforced at both the UI level (card not rendered) and the API level (data not returned) — UI hiding alone is not security
- **SR-23:** Line Manager is displayed as a clickable link resolving to the manager's full name at query time (not a snapshot) — name changes on the manager's profile reflect immediately
- **SR-24:** When the assigned line manager's account status is Deactivated, an `(Inactive)` suffix is appended to the manager's name; the link remains clickable
- **SR-25:** When the user has no line manager assigned, the Line Manager row displays "—" (top-of-hierarchy users)
- **SR-26:** The Direct Reports list is computed live from the inverse Line Manager relationship at query time — there is no separately stored list
- **SR-27:** The Direct Reports list is ordered alphabetically by full name
- **SR-28:** The Direct Reports card is read-only on this page — no editing, no add/remove controls (subordinate reassignment happens via Update Information on individual user profiles)
- **SR-29:** The Direct Reports card is visible to any user with `user.view` permission — no additional permission required

**State Transitions:**
```
[User List] → [Click user row] → [User Details: Loading]
[Loading] → [Data fetched] → [Overview displayed]
[Loading] → [API error] → [Error state with retry]
[Loading] → [User not found (404)] → [Not found state → link to User List]
[Overview] → [Click ← back] → [User List (preserved state)]
[Overview] → [Click action button] → [Action view (separate DR)]
[Overview] → [Click sidebar nav] → [Navigate away]
```

**Dependencies:**
- **US-001 (Authentication):** User must be signed in
- **US-004 (Role & Permission Management):** Controls view and management permissions including the new `user.salary.view` and `user.banking.view` fine-grained permissions; provides role display data
- **EP-008 US-001 (Department Management):** Source of department name displayed in Work Profile
- **EP-008 US-002 (Position Management):** Source of position name displayed in Work Profile
- **EP-008 US-003 (Skill Management):** Source of skill names displayed in Work Profile
- **DR-001-005-01 (User List):** Entry point — user row click navigates here
- **DR-001-005-02 (Create User):** Fields created there (including new ID Cards, Emergency Contact, Salary, Banking, Nationality, Marital status, Permanent/Temporary address, Education level, Line Manager) are displayed on this page
- **DR-001-005-04 (Update User Information):** Line Manager field is set/changed/cleared through Update Information — Direct Reports list updates accordingly
- **API endpoint to fetch a user's direct reports:** Inverse query (all users whose Line Manager = current user) — required to populate the Direct Reports card

---

## 7. UX Optimizations

**Usability Considerations:**

- **UX-01:** Skeleton placeholders shown during data load — matching the card layout to prevent layout shift
- **UX-02:** Avatar overlaps the cover image by ~66px — creating visual depth and a polished profile feel
- **UX-03:** Each data field uses icon + label (muted) + value (body) pattern — consistent visual rhythm across both cards, easy to scan
- **UX-04:** Personal Information uses 4-column layout (row 1) + 2-column (row 2); Work Profile uses 3-column + single rows — logical grouping by data density
- **UX-05:** Empty optional fields show "—" rather than being hidden — maintains consistent card layout and signals that the field exists but has no value
- **UX-06:** Status badge is color-coded (green = Activated, gray = Deactivated) and placed directly next to the name — immediately visible without scanning
- **UX-07:** Left action panel uses stacked full-width buttons with icon prefix — clear, scannable action menu; active tab (Overview) visually distinguished
- **UX-08:** Delete User button uses a distinct/danger style (red text or red icon) — visually separated from other actions to prevent accidental clicks
- **UX-09:** Back arrow (←) in the title bar provides a quick return path — no need to use browser back or sidebar navigation
- **UX-10:** CV/Resumé link opens in a new tab or triggers download — user stays on the profile page
- **UX-11:** Page URL includes user identifier — bookmarkable and shareable for quick access

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (≥1280px) | Full two-panel layout: left action panel (189px) + right content (1445px) |
| Below desktop | Out of scope for this release |

**Accessibility Requirements:**
- [x] Keyboard navigable — Tab through back arrow, left panel buttons, CV link
- [x] Screen reader compatible — labels for all data fields, status badge state announced
- [x] Avatar has alt text (user's full name)
- [x] Status badge communicated via text (not color-only — "Activated"/"Deactivated" text included)
- [x] Sufficient color contrast — meets WCAG 2.1 AA standards
- [x] Focus indicators visible on all interactive elements
- [x] CV/Resumé link has descriptive text (filename, not generic "Click here")

**Design References:**
- Figma: [User Details](https://www.figma.com/design/YEHeFgVZau7wmo9BZBVuZC?node-id=3120-5348) (node `3120:5348`)
- Design tokens: heading 2 (Geist Semibold 30px), paragraph small/regular (14px), muted foreground `#737373`, green/500 `#22c55e` (Activated badge), border `#e5e5e5`

---

## 8. Additional Information

### Out of Scope
- Inline editing of any field on this page (read-only display only)
- Update Information — separate DR (planned)
- Change User Role — separate DR (planned)
- Change Email — separate DR (planned)
- Reset Password — separate DR (planned)
- Activate/Deactivate — separate DR (planned)
- Delete User — separate DR (planned)
- User-specific cover image upload or customization (static/decorative only)
- Activity log or audit trail on the profile page
- User-to-user messaging or notes
- Mobile or tablet responsive layout
- Display of the full (unmasked) bank account number — the full number is never shown on this page for any viewer
- Editing salary fields on this page — handled by Update User Information DR with `user.salary.manage` permission
- Editing banking fields on this page — handled by Update User Information DR with `user.banking.manage` permission
- Editing emergency contacts on this page — handled by Update User Information DR
- Editing ID card fields on this page — handled by Update User Information DR
- Multi-currency salary display — VND only for this release
- Direct Reports pagination — current implementation shows all reports with no pagination (acceptable for v1 / small org)
- Direct Reports filtering or search within the card
- Multi-tier hierarchy visualization (skip-level / grand-subordinates)
- Org-chart visualization page
- Editing Direct Reports from this view — must be done from each individual subordinate's Update Information page

### Open Questions
- [ ] **Cover image source:** Is the static cover a single system-wide image, or one per role/department? — **Owner:** Design Team — **Status:** Pending
- [ ] **Avatar fallback:** Initials-based placeholder (e.g., "HT" for Henry Tran) or generic silhouette icon? — **Owner:** Design Team — **Status:** Pending
- [ ] **CV/Resumé interaction:** Click to download directly, or open in a new browser tab for preview? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Action button order:** Is the current order (Overview → Update → Role → Email → Password → Status → Delete) confirmed, or subject to change? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Future entry points:** Will other modules (e.g., Department List "view members", Role List "view assigned users") link directly to User Details? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Salary currency display format:** Suffix as "₫" symbol or "VND" text label? — **Owner:** Product Owner — **Status:** Pending
- [ ] **ID card image preview UX:** Open enlarged image in an in-page modal/lightbox, or in a new browser tab? — **Owner:** Design Team — **Status:** Pending
- [ ] **Banking account masking pattern:** Confirmed mask format — "•••• 1234" (bullet dots + last 4) or alternative (e.g., "XXXX-XXXX-XXXX-1234")? — **Owner:** Design Team — **Status:** Pending
- [ ] **ID Cards permission:** Should ID card images require a dedicated `user.idcards.view` permission like Salary/Banking, or remain under base `user.view`? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Direct Reports pagination threshold:** At what list size do we add pagination or search within the Direct Reports card? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Direct Reports count badge:** Should the Direct Reports card title show a count badge (e.g., "Direct Reports (5)") for at-a-glance visibility? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Manager link target:** Should the Line Manager clickable link open in the same tab (replacing current view) or a new tab (preserving context)? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-001-005-01: User List | Entry point — clicking a user row navigates here |
| DR-001-005-02: Create User | Fields created there are displayed on this page (including Line Manager) |
| DR-001-005-04: Update User Information | Line Manager field set/changed/cleared here — affects Direct Reports list |
| Update Information (planned) | Triggered from left panel button 2 |
| Change User Role (planned) | Triggered from left panel button 3 |
| Change Email (planned) | Triggered from left panel button 4 |
| Reset Password (planned) | Triggered from left panel button 5 |
| Activate/Deactivate (planned) | Triggered from left panel button 6 |
| Delete User (planned) | Triggered from left panel button 7 |
| US-001: Authentication | User must be signed in |
| US-004: Role & Permission Management | Controls view/management permissions; role displayed on profile |
| EP-008 US-001/002/003 | Department, Position, Skills data displayed in Work Profile |

### Notes
- This is the **first detail/profile page** in the HRM platform — all other screens are either list views or form pages. The two-panel layout (action menu + content) is a new pattern.
- The **left action panel** effectively replaces the gear icon dropdown from the User List — providing the same actions (plus additional ones like Change Email, Reset Password) in a more prominent, always-visible format.
- The **7 action buttons** represent 6 future DRs — each will be a separate detail requirement. The Overview button simply shows the current view (this DR).
- The page is **completely read-only** — there are zero input fields. All modifications happen via the action buttons which navigate to separate views/dialogs.
- The **cover image is static** and decorative — confirmed by user. No per-user customization, no upload flow.

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
| 1.0 | 2026-03-26 | BA Agent | Initial draft — Figma design context from node 3120:5348 |
| 1.1 | 2026-05-18 | BA Agent | Added 4 new read-only cards (ID Cards, Emergency Contact, Salary, Banking); added Nationality + Marital status + Temporary address + Education level fields; renamed Address → Permanent address; permission-gated visibility for Salary and Banking via new `user.salary.view` and `user.banking.view` permissions; banking account number always masked to last 4 digits |
| 1.2 | 2026-05-20 | BA Agent | Added Line Manager row to Work Profile card (clickable link to manager's profile, `(Inactive)` suffix when deactivated); added new "Direct Reports" sub-section listing all users whose Line Manager = current user (computed live from inverse relationship, alphabetical, click-to-navigate); two-variant `.team`/`.all` permission pattern documented in design spec for downstream stories |
