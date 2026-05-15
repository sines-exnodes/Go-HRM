---
document_type: DETAIL_REQUIREMENT
# Platform Information
platform: web-app
platform_display: "WEB-APP"
# Epic & Story Identification
epic_id: EP-010
story_id: US-001
story_name: "Announcement List"
# Detail Requirement Identification
detail_id: DR-010-001-01
detail_name: "Announcement List"
# Status & Version
status: draft
version: "1.0"
created_date: 2026-04-25
last_updated: 2026-04-25
# Document linking
related_documents:
  - path: "../ANALYSIS.md"
    relationship: sibling
  - path: "../../EPIC.md"
    relationship: parent
# Input sources
input_sources:
  - type: figma
    description: "Announcement List screen design"
    extracted_date: 2026-04-25
---

# Detail Requirement: Announcement List

**Detail ID:** DR-010-001-01
**Story:** US-001-announcement-list
**Epic:** EP-010 (Announcements)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **HR Manager or Admin**, I want to **view and manage all company announcements in a centralized list** so that **I can efficiently create, send, and track internal communications to employees**.

**Purpose:** Provide a single location for authorized personnel to manage all company announcements. The list view enables quick scanning of announcement status, content preview, and access to management actions (edit, send, delete) based on announcement status.

**Target Users:** 
- HR Managers with announcement management permission
- Admins with full system access

**Key Functionality:**
- View all announcements with status indicators (Draft/Sent)
- Search announcements by title
- Filter announcements by status
- Access gear icon actions based on announcement status
- Navigate to create new announcements

---

## 2. User Workflow

**Entry Point:** Sidebar menu > Organization Settings > Announcements

**Preconditions:**
- User is authenticated and logged in
- User has announcement management permission (via US-004 Permission Management)

**Main Flow:**
1. User clicks "Announcements" in the Organization Settings section of the sidebar
2. System displays the announcement list page with header "Announcements"
3. System loads and displays all announcements in a paginated table (default 10 per page)
4. Table shows: Title, Status (badge), Created Date, Sent Date (if applicable), Actions (gear icon)
5. User can search by typing in the search box (debounced 300ms, searches title field)
6. User can filter by status using the status dropdown filter
7. User can click gear icon to access available actions based on announcement status
8. User can click "+ Add Announcement" button to navigate to Create Announcement form

**Alternative Flows:**
- **Alt 1 - No Results:** If search/filter yields no results, display empty state with message "No announcements found"
- **Alt 2 - Loading:** While data loads, display skeleton loading state for table rows
- **Alt 3 - Error:** If data fetch fails, display error message with "Retry" button

**Exit Points:**
- **Success:** User views list, performs actions, or navigates to create/edit screens
- **Cancel:** User navigates away via sidebar or browser back
- **Error:** System displays error toast and offers retry option

---

## 3. Field Definitions

### Search & Filter Elements

| Field Name | Field Type | Validation Rule | Mandatory | Default Value | Description |
|------------|------------|-----------------|-----------|---------------|-------------|
| Search | Text Input | Max 100 characters | No | Empty | Search announcements by title |
| Status Filter | Dropdown | Single-select | No | "All" | Filter by Draft, Sent, or All |

### Table Columns

| Column Name | Data Type | Sort | Description |
|-------------|-----------|------|-------------|
| Title | Text | Yes (A-Z default) | Announcement title, truncated with ellipsis if > 50 chars |
| Status | Badge | Yes | Draft (gray) or Sent (green) badge |
| Created Date | Date | Yes | Format: DD/MM/YYYY |
| Sent Date | Date | Yes | Format: DD/MM/YYYY or "--" if not sent |
| Actions | Gear Icon | No | Context menu with status-based actions |

### Interaction Elements

| Element Name | Type | State/Condition | Trigger Action | Description |
|--------------|------|-----------------|----------------|-------------|
| + Add Announcement | Primary Button | Always visible for users with permission | Navigate to Create Announcement form | Top-right of page header |
| Search Box | Text Input | Always enabled | Debounced search (300ms) | Placeholder: "Search by title" |
| Status Filter | Dropdown | Always enabled | Filter table results | Options: All, Draft, Sent |
| Gear Icon | Icon Button | Visible per row | Opens action dropdown | Last column of each row |
| Pagination | Component | Visible when > 10 items | Navigate pages | Options: 10, 25, 50 per page |

---

## 4. Data Display

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Title | Text | N/A (required field) | Plain text, max 50 chars visible | Announcement subject line |
| Status | Badge | N/A (always has status) | Colored badge | Current lifecycle state |
| Created Date | Date | N/A (auto-generated) | DD/MM/YYYY | When announcement was created |
| Sent Date | Date | "--" | DD/MM/YYYY | When announcement was sent to employees |

### Status Badge Definitions

| Status | Badge Color | Description |
|--------|-------------|-------------|
| Draft | Gray | Announcement created but not yet sent |
| Sent | Green | Announcement has been sent to employees |

### Gear Icon Actions by Status

| Status | Available Actions |
|--------|-------------------|
| Draft | Edit, Send, Delete |
| Sent | Delete |

### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Initial data fetch | Skeleton rows (5 placeholder rows) |
| Empty - No Data | No announcements exist | "No announcements yet. Click '+ Add Announcement' to create your first announcement." |
| Empty - No Results | Search/filter yields nothing | "No announcements found matching your criteria." |
| Error | Data fetch fails | "Failed to load announcements. Please try again." with Retry button |
| Success | Data loaded | Populated table with announcements |

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** Users with announcement management permission can view the announcement list; users without permission cannot access the page
- **AC-02:** Table displays all announcements with Title, Status badge, Created Date, Sent Date, and Actions columns
- **AC-03:** Search filters announcements by title with 300ms debounce; matching is case-insensitive and partial (contains)
- **AC-04:** Status filter correctly filters by Draft, Sent, or shows All
- **AC-05:** Gear icon for Draft announcements shows Edit, Send, and Delete actions
- **AC-06:** Gear icon for Sent announcements shows only Delete action
- **AC-07:** "+ Add Announcement" button navigates to the Create Announcement form
- **AC-08:** Pagination defaults to 10 items with options for 25 and 50; resets to page 1 on search/filter change
- **AC-09:** Empty states display appropriate messages based on context (no data vs. no results)
- **AC-10:** Loading state shows skeleton placeholders while data is being fetched

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| View list with data | Navigate to Announcements | Table shows all announcements with correct columns | High |
| Search by title | Type "Holiday" in search | Only announcements containing "Holiday" in title shown | High |
| Filter by Draft | Select "Draft" from filter | Only Draft status announcements shown | High |
| Filter by Sent | Select "Sent" from filter | Only Sent status announcements shown | High |
| Gear icon - Draft | Click gear on Draft row | Shows Edit, Send, Delete options | High |
| Gear icon - Sent | Click gear on Sent row | Shows only Delete option | High |
| Empty state - no data | No announcements exist | Shows "No announcements yet" message | Medium |
| Empty state - no results | Search with no matches | Shows "No announcements found" message | Medium |
| Pagination | Have 15 announcements | Shows 10 on first page, 5 on second | Medium |
| Permission denied | User without permission | Cannot access page (redirect or 403) | High |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **Rule 1:** Only users with announcement management permission (configured via US-004) can access the announcement list
- **Rule 2:** Announcements have exactly 2 statuses: Draft and Sent
- **Rule 3:** Once an announcement is Sent, it cannot be edited (immutable after send)
- **Rule 4:** Deleting a Sent announcement removes it from admin view but does not retract from employees who received it
- **Rule 5:** Search is performed on the title field only; matching is case-insensitive partial match
- **Rule 6:** Default sort order is Created Date descending (newest first)

**State Transitions:**

```
[Draft] → [Send Action] → [Sent]
[Draft] → [Delete Action] → [Deleted]
[Sent] → [Delete Action] → [Deleted]
```

**Permission Model:**

| Action | Required Permission |
|--------|---------------------|
| View List | Announcement Management |
| Create | Announcement Management |
| Edit (Draft only) | Announcement Management |
| Send | Announcement Management |
| Delete | Announcement Management |

**Dependencies:**
- US-004 Permission Management - provides permission configuration
- Authentication system - user must be logged in
- Backend API for announcements CRUD operations

---

## 7. UX Optimizations

**Usability Considerations:**

- **Optimization 1:** Debounced search (300ms) prevents excessive API calls while typing
- **Optimization 2:** Skeleton loading provides visual feedback during data fetch
- **Optimization 3:** Status badges use distinct colors (gray/green) for quick visual scanning
- **Optimization 4:** Gear icon in last column follows established pattern from other list views
- **Optimization 5:** Contextual empty states guide users on next steps

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Desktop (>1024px) | Full table layout with all columns visible |
| Below Desktop | Out of scope - web admin is desktop-only |

**Accessibility Requirements:**
- [x] Keyboard navigable - all interactive elements accessible via Tab
- [x] Screen reader compatible - proper ARIA labels on buttons and status badges
- [x] Sufficient color contrast - badges meet WCAG AA standards
- [x] Focus indicators visible - clear focus ring on interactive elements

**Design References:**
- Follows established list view pattern from Leave Request List, OT Request List
- Status badges follow same color convention as other modules
- Gear icon behavior consistent with other admin list views

---

## 8. Additional Information

### Out of Scope
- Mobile responsive layout (web admin is desktop-only)
- Bulk actions (select multiple, bulk delete)
- Advanced search (by content, date range, etc.)
- Announcement scheduling (send at future date)
- Read receipts / tracking who viewed announcement
- Announcement categories or tags
- Rich text formatting preview in list view

### Open Questions
- None - all questions resolved

### Related Features
- DR-009-001-02: Create Announcement Form (to be created)
- DR-009-001-03: Edit Announcement Form (to be created)
- DR-009-001-04: Send Announcement Confirmation (to be created)
- DR-009-001-05: Delete Announcement Confirmation (to be created)
- US-004: Permission Management (provides access control)

### Notes
- Sent announcements are immutable by business decision to maintain audit integrity
- The "Sent Date" column distinguishes between draft and sent items at a glance
- Delete of sent announcement is administrative cleanup only; employees retain their copy

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | | | Pending |
| Product Owner | | | Pending |
| UX Designer | | | Pending |
| Tech Lead | | | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-04-25 | Claude | Initial draft |
