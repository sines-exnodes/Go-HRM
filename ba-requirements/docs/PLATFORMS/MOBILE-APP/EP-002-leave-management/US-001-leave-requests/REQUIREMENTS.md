---
document_type: REQUIREMENTS
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-002
story_id: US-001
story_name: "Leave Requests"
status: draft
version: "1.0"
last_updated: "2026-04-16"
approved_by: null
related_documents:
  - path: "./ANALYSIS.md"
    relationship: sibling
  - path: "./FLOWCHART.md"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
---

# Requirements: Leave Requests

**Epic:** EP-002 (Leave Management)
**Story:** US-001-leave-requests
**Platform:** MOBILE-APP
**Status:** Draft

---

## 1. Overview

This document specifies the functional and non-functional requirements for mobile leave request management in the Exnodes HRM application. The module enables employees to view their leave balances, submit requests, and track request status.

---

## 2. Functional Requirements

### 2.1 Leave Requests Dashboard

| ID | Requirement | Priority | DR Reference |
|----|-------------|----------|--------------|
| FR-US-001-01 | Dashboard screen with leave summary | Critical | DR-002-001-01 |
| FR-US-001-02 | Display leave balance by type (Annual, Sick) | Critical | DR-002-001-01 |
| FR-US-001-03 | Display pending requests count | High | DR-002-001-01 |
| FR-US-001-04 | Display recent requests (last 5) | High | DR-002-001-01 |
| FR-US-001-05 | Quick action button for new request | Critical | DR-002-001-01 |
| FR-US-001-06 | Navigation to full request list | High | DR-002-001-01 |
| FR-US-001-07 | Pull-to-refresh functionality | Medium | DR-002-001-01 |

### 2.2 Leave Request List

| ID | Requirement | Priority | DR Reference |
|----|-------------|----------|--------------|
| FR-US-001-10 | List view of all user's leave requests | High | TBD |
| FR-US-001-11 | Filter by status (Pending, Approved, Rejected, Cancelled) | Medium | TBD |
| FR-US-001-12 | Filter by date range | Medium | TBD |
| FR-US-001-13 | Sort by date (newest first default) | Medium | TBD |
| FR-US-001-14 | Tap to view request details | High | TBD |

### 2.3 Create Leave Request

| ID | Requirement | Priority | DR Reference |
|----|-------------|----------|--------------|
| FR-US-001-20 | Leave type selection | Critical | TBD |
| FR-US-001-21 | Date range picker (from/to) | Critical | TBD |
| FR-US-001-22 | Leave period selection (Full day, Half day AM/PM) | High | TBD |
| FR-US-001-23 | Reason/notes field (optional) | Medium | TBD |
| FR-US-001-24 | Attachment upload (optional) | Low | TBD |
| FR-US-001-25 | Balance validation before submission | High | TBD |
| FR-US-001-26 | Submit and confirmation | Critical | TBD |

### 2.4 Leave Request Details

| ID | Requirement | Priority | DR Reference |
|----|-------------|----------|--------------|
| FR-US-001-30 | View full request details | High | TBD |
| FR-US-001-31 | Display status with visual indicator | High | TBD |
| FR-US-001-32 | Cancel option for pending requests | High | TBD |
| FR-US-001-33 | View attached documents | Medium | TBD |

---

## 3. Non-Functional Requirements

| Category | Requirement | Target |
|----------|-------------|--------|
| Performance | Dashboard load time | < 2 seconds |
| Performance | List pagination | 20 items per page |
| Usability | Touch targets | Minimum 44x44 points |
| Usability | Pull-to-refresh | Standard iOS/Android gesture |
| Offline | Dashboard data | Cache last-loaded data |
| Accessibility | Screen reader | All elements labeled |

---

## 4. Detail Requirements Index

| DR ID | Feature | Status | File |
|-------|---------|--------|------|
| DR-002-001-01 | Leave Requests Dashboard | Draft | `details/DR-002-001-01-leave-requests-dashboard.md` |
| DR-002-001-02 | Create Leave Request | Draft | `details/DR-002-001-02-create-leave-request.md` |

---

## 5. Acceptance Criteria

### Dashboard
- [ ] User sees leave balance summary on dashboard
- [ ] User sees count of pending requests
- [ ] User can tap to create new request
- [ ] User can pull-to-refresh to update data
- [ ] User can navigate to full request list

### Create Request
- [ ] User can select leave type
- [ ] User can select date range
- [ ] User sees balance warning if insufficient
- [ ] User receives confirmation on successful submission

### Request Details
- [ ] User can view full details of any request
- [ ] User can cancel pending requests
- [ ] Status is clearly indicated visually

---

**Document Version:** 1.0
**Last Updated:** 2026-04-16
**Author:** BA Team
