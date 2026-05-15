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
version: "1.0"
created_date: 2026-03-26
last_updated: 2026-03-26
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
**Version:** 1.0

---

## 1. Use Case Description

As a **user with user management permission**, I want to **view a user's full profile including personal information, account status, and work profile**, so that **I can review user details and access management actions (edit, role change, email change, password reset, activate/deactivate, delete) from a single page**.

**Purpose:** Provide a centralized read-only view of all user data — personal, account, and employee profile — in one place. The User Details page is the hub for all per-user management actions, accessed via a left-panel action menu. This replaces the need to navigate to separate pages for different user operations.

**Target Users:**
- Any role with **user view permission** — can see the profile data and Overview tab
- Any role with **user management permission** — can additionally see and use the action buttons (Update Information, Change User Role, Change Email, Reset Password, Activate/Deactivate, Delete User)

**Key Functionality:**
- Read-only display of all user data (no inline editing)
- Static cover image + user avatar + name + status badge + role display
- Two info cards: Personal Information (6 fields) + Work Profile (5 fields)
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
| Address | Location icon | Label + value (spans wider column) | "—" | Full address text |

### Work Profile Card

| Data Name | Icon | Format | Empty State | Business Meaning |
|-----------|------|--------|-------------|-----------------|
| Department | Building icon | Label + value | "—" (unassigned) | Organizational department |
| Position | Briefcase icon | Label + value | "—" (unassigned) | Job position |
| Experience from | Calendar icon | Label + value (4-digit year) | "—" | Career start year |
| Skills | Star icon | Label + value (comma-separated) | "—" (no skills assigned) | Competency tags (read-only text) |
| CV/Resumé | File icon | Label + clickable link/filename | "—" (no file uploaded) | Click to download or open in new tab |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Page first opens, API fetching | Skeleton placeholders for avatar, name, cards |
| Populated | User data loaded successfully | Full profile with all fields displayed |
| Partial data | Some optional fields empty | "—" shown for empty optional fields; layout unchanged |
| Error | API fails to load user data | Error message with retry button; left panel still visible |
| User not found | User deleted since list was loaded | "User not found" message with link to return to User List |
| View-only mode | User has view permission only | Overview button visible; 6 management action buttons hidden |
| Management mode | User has management permission | All 7 left-panel buttons visible |

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
│          │ │  189px      │ │ Gender | Address                  │ ││
│          │ │             │ │                                  │ ││
│          │ │             │ │ Work Profile                      │ ││
│          │ │             │ │ Dept | Position | Experience      │ ││
│          │ │             │ │ Skills                            │ ││
│          │ │             │ │ CV/Resumé                         │ ││
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
- **AC-10:** Personal Information card displays 6 fields: Full name, Email, Phone number, Date of birth, Gender, Address
- **AC-11:** Row 1 shows Full name, Email, Phone number, Date of birth in a 4-column layout
- **AC-12:** Row 2 shows Gender and Address in a 2-column layout (Address spans wider)
- **AC-13:** Each field displays with an icon prefix, label (muted text), and value (body text)
- **AC-14:** Empty optional fields display "—" as placeholder value

**Work Profile Card:**
- **AC-15:** Work Profile card displays 5 fields: Department, Position, Experience from, Skills, CV/Resumé
- **AC-16:** Row 1 shows Department, Position, Experience from in a 3-column layout
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

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Happy path — full data | Click user with all fields populated | All fields shown with values, status badge green | High |
| Partial data | Click user with no phone, address, skills, CV | "—" shown for empty optional fields | High |
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
- **SR-11:** Empty optional fields (Phone number, Address, Skills, CV/Resumé) display "—" — they are never hidden, only show placeholder value
- **SR-12:** The back arrow navigation preserves User List state (active search query, filter selections, current page, rows per page) — the list does not reload from scratch

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
- **US-004 (Role & Permission Management):** Controls view and management permissions; provides role display data
- **EP-008 US-001 (Department Management):** Source of department name displayed in Work Profile
- **EP-008 US-002 (Position Management):** Source of position name displayed in Work Profile
- **EP-008 US-003 (Skill Management):** Source of skill names displayed in Work Profile
- **DR-001-005-01 (User List):** Entry point — user row click navigates here
- **DR-001-005-02 (Create User):** Fields created there are displayed on this page

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

### Open Questions
- [ ] **Cover image source:** Is the static cover a single system-wide image, or one per role/department? — **Owner:** Design Team — **Status:** Pending
- [ ] **Avatar fallback:** Initials-based placeholder (e.g., "HT" for Henry Tran) or generic silhouette icon? — **Owner:** Design Team — **Status:** Pending
- [ ] **CV/Resumé interaction:** Click to download directly, or open in a new browser tab for preview? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Action button order:** Is the current order (Overview → Update → Role → Email → Password → Status → Delete) confirmed, or subject to change? — **Owner:** Product Owner — **Status:** Pending
- [ ] **Future entry points:** Will other modules (e.g., Department List "view members", Role List "view assigned users") link directly to User Details? — **Owner:** Product Owner — **Status:** Pending

### Related Features

| Feature | Relationship |
|---------|-------------|
| DR-001-005-01: User List | Entry point — clicking a user row navigates here |
| DR-001-005-02: Create User | Fields created there are displayed on this page |
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
