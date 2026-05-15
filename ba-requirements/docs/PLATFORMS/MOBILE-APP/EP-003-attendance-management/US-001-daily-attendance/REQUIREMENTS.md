---
document_type: REQUIREMENTS
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-003
story_id: US-001
story_name: "Daily Attendance"
status: draft
version: "1.0"
last_updated: "2026-04-18"
approved_by: null
related_documents:
  - path: "./ANALYSIS.md"
    relationship: sibling
  - path: "./FLOWCHART.md"
    relationship: sibling
  - path: "../EPIC.md"
    relationship: parent
---

# Requirements: Daily Attendance

**Epic:** EP-003 (Attendance Management)
**Story:** US-001-daily-attendance
**Platform:** MOBILE-APP
**Status:** Draft

---

## 1. Overview

This document specifies the functional and non-functional requirements for the daily attendance check-in/out feature in the Exnodes HRM mobile application. The module enables employees to record their attendance with GPS verification, track their attendance streaks, and view their attendance history.

---

## 2. Functional Requirements

### 2.1 Check-In Functionality

| ID | Requirement | Priority | DR Reference |
|----|-------------|----------|--------------|
| FR-US-001-01 | Check-in button visible on employee dashboard | Critical | TBD |
| FR-US-001-02 | GPS location verification before allowing check-in | Critical | TBD |
| FR-US-001-03 | Check-in allowed only within 50m radius of office location | Critical | TBD |
| FR-US-001-04 | Block check-in if GPS is unavailable or disabled | Critical | TBD |
| FR-US-001-05 | Display clear error message when outside office radius | High | TBD |
| FR-US-001-06 | Display clear error message when GPS unavailable | High | TBD |
| FR-US-001-07 | Record check-in timestamp upon successful check-in | Critical | TBD |
| FR-US-001-08 | Mark check-in as "late" if after configured threshold (default 9:00 AM) | High | TBD |
| FR-US-001-09 | Visual confirmation upon successful check-in | High | TBD |
| FR-US-001-10 | Check-in available 24/7 (no time window restriction) | Medium | TBD |

### 2.2 Check-Out Functionality

| ID | Requirement | Priority | DR Reference |
|----|-------------|----------|--------------|
| FR-US-001-11 | Check-out button available after check-in | High | TBD |
| FR-US-001-12 | Check-out is optional (not mandatory) | High | TBD |
| FR-US-001-13 | GPS verification for check-out (same as check-in) | Medium | TBD |
| FR-US-001-14 | Record check-out timestamp upon successful check-out | High | TBD |
| FR-US-001-15 | Auto check-out at 11:00 PM if employee forgets | High | TBD |
| FR-US-001-16 | Mark auto check-out records distinctly from manual | Medium | TBD |
| FR-US-001-17 | Visual confirmation upon successful check-out | High | TBD |

### 2.3 Multiple Sessions

| ID | Requirement | Priority | DR Reference |
|----|-------------|----------|--------------|
| FR-US-001-20 | Allow multiple check-in/out sessions per day | High | TBD |
| FR-US-001-21 | After check-out, employee can check-in again | High | TBD |
| FR-US-001-22 | Monthly attendance count increments once per day only | High | TBD |
| FR-US-001-23 | Track all sessions for reporting purposes | Medium | TBD |

### 2.4 Attendance Dashboard Display

| ID | Requirement | Priority | DR Reference |
|----|-------------|----------|--------------|
| FR-US-001-30 | Display current check-in status (checked in / not checked in) | Critical | TBD |
| FR-US-001-31 | Display number of checked-in days this month | High | TBD |
| FR-US-001-32 | Display current attendance streak (consecutive workdays) | High | TBD |
| FR-US-001-33 | Streak calculation excludes weekends | High | TBD |
| FR-US-001-34 | Streak calculation excludes company holidays | High | TBD |
| FR-US-001-35 | Streak resets when a workday is missed | Medium | TBD |
| FR-US-001-36 | Display today's check-in time if checked in | Medium | TBD |
| FR-US-001-37 | Display late indicator if checked in late | Medium | TBD |

### 2.5 Attendance History

| ID | Requirement | Priority | DR Reference |
|----|-------------|----------|--------------|
| FR-US-001-40 | View list of past attendance records | High | TBD |
| FR-US-001-41 | Display check-in and check-out timestamps for each day | High | TBD |
| FR-US-001-42 | Display late status for each record | Medium | TBD |
| FR-US-001-43 | Display auto-checkout indicator where applicable | Low | TBD |
| FR-US-001-44 | Monthly summary view with total days attended | Medium | TBD |
| FR-US-001-45 | Filter history by month | Medium | TBD |
| FR-US-001-46 | Pull-to-refresh to update history | Medium | TBD |

### 2.6 Notifications

| ID | Requirement | Priority | DR Reference |
|----|-------------|----------|--------------|
| FR-US-001-50 | Push notification when near office but not checked in | Medium | TBD |
| FR-US-001-51 | Push notification for streak milestone celebrations | Low | TBD |
| FR-US-001-52 | Push notification reminder to check out at end of day | Medium | TBD |
| FR-US-001-53 | Configurable notification preferences | Low | TBD |

### 2.7 Admin Configuration (Backend/API)

| ID | Requirement | Priority | DR Reference |
|----|-------------|----------|--------------|
| FR-US-001-60 | Office location (latitude, longitude) configurable by admin | High | TBD |
| FR-US-001-61 | Check-in radius (default 50m) configurable by admin | Medium | TBD |
| FR-US-001-62 | Late threshold time (default 9:00 AM) configurable by admin | High | TBD |
| FR-US-001-63 | Auto check-out time (default 11:00 PM) configurable by admin | Medium | TBD |
| FR-US-001-64 | Company holiday calendar integration | High | TBD |

---

## 3. Non-Functional Requirements

| Category | Requirement | Target |
|----------|-------------|--------|
| Performance | GPS location acquisition | < 5 seconds |
| Performance | Check-in confirmation response | < 2 seconds |
| Performance | History list load time | < 2 seconds |
| Reliability | GPS accuracy | 50m tolerance |
| Usability | Check-in button | Prominent, minimum 48x48 points |
| Usability | Location status | Clear visual indicator |
| Battery | Location services | Use geofencing, not continuous GPS |
| Offline | Check-in behavior | Block until GPS available |
| Accessibility | Screen reader | All elements labeled |
| Security | Location data | Encrypted transmission |

---

## 4. Business Rules

| Rule ID | Rule Description |
|---------|-----------------|
| BR-001 | Employee must be within 50m of configured office location to check in |
| BR-002 | Check-in without valid GPS location is not permitted |
| BR-003 | Multiple check-in/out sessions per day are allowed |
| BR-004 | Daily attendance count increments maximum once per calendar day |
| BR-005 | Attendance streak counts consecutive workdays only (Mon-Fri) |
| BR-006 | Weekends do not break attendance streak |
| BR-007 | Company holidays do not break attendance streak |
| BR-008 | Missing a workday resets streak to zero |
| BR-009 | Check-in after configured threshold (default 9:00 AM) is marked as late |
| BR-010 | If no check-out by 11:00 PM, system performs auto check-out |
| BR-011 | Auto check-out records are marked distinctly for reporting |

---

## 5. Detail Requirements Index

| DR ID | Feature | Status | File |
|-------|---------|--------|------|
| DR-003-001-01 | Check-In/Out Screen | TBD | `details/DR-003-001-01-check-in-out.md` |
| DR-003-001-02 | Attendance Dashboard Widget | TBD | `details/DR-003-001-02-attendance-dashboard.md` |
| DR-003-001-03 | Attendance History | TBD | `details/DR-003-001-03-attendance-history.md` |
| DR-003-001-04 | Attendance Notifications | TBD | `details/DR-003-001-04-attendance-notifications.md` |

---

## 6. Acceptance Criteria

### Check-In/Out
- [ ] Employee sees check-in button on dashboard when not checked in
- [ ] Check-in button is disabled when outside 50m office radius
- [ ] Check-in button is disabled when GPS is unavailable
- [ ] Employee sees clear error message explaining why check-in is blocked
- [ ] Employee receives visual confirmation upon successful check-in
- [ ] Check-in after 9:00 AM is marked as "late"
- [ ] Employee can check out at any time (optional)
- [ ] Auto check-out occurs at 11:00 PM if not manually checked out
- [ ] Employee can check in again after checking out (multiple sessions)

### Dashboard Display
- [ ] Dashboard shows current check-in status
- [ ] Dashboard shows number of checked-in days this month
- [ ] Dashboard shows current attendance streak
- [ ] Streak correctly excludes weekends and holidays
- [ ] Streak resets to zero when a workday is missed

### History
- [ ] Employee can view list of past attendance records
- [ ] Each record shows check-in and check-out timestamps
- [ ] Late arrivals are clearly indicated
- [ ] Auto check-outs are distinguishable from manual
- [ ] Employee can filter history by month
- [ ] Pull-to-refresh updates the history list

### Notifications
- [ ] Employee receives notification when near office but not checked in
- [ ] Employee receives streak milestone celebrations (e.g., 10-day streak)
- [ ] Employee receives end-of-day reminder to check out
- [ ] Notifications can be configured/disabled in settings

### Admin Configuration
- [ ] Admin can configure office location coordinates
- [ ] Admin can configure check-in radius
- [ ] Admin can configure late threshold time
- [ ] System respects company holiday calendar for streak calculation

---

## 7. Data for Reporting

The attendance data captured will support the following reports (to be detailed in separate reporting requirements):

| Data Point | Description | Usage |
|------------|-------------|-------|
| Check-in timestamp | Time employee checked in | Punctuality reports |
| Check-out timestamp | Time employee checked out | Working hours calculation |
| Late flag | Whether check-in was after threshold | Punctuality metrics |
| Auto-checkout flag | Whether checkout was automatic | Compliance tracking |
| Location coordinates | GPS position at check-in | Audit trail |
| Session count | Number of sessions per day | Activity tracking |

---

**Document Version:** 1.0
**Last Updated:** 2026-04-18
**Author:** BA Team
