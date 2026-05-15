# Attendance Management Requirements

**Document Type:** Business Requirements
**Status:** Draft
**Version:** 1.0
**Last Updated:** 2026-04-18

---

## 1. Overview

### Purpose
Enable employees to record their daily work attendance through a digital check-in/out system with location verification. The system tracks attendance patterns, encourages consistency through streak tracking, and provides data for organizational reporting.

### Goals
- Accurate attendance tracking with location verification
- Encourage punctuality and consistency through gamification (streaks)
- Provide employees visibility into their attendance records
- Generate data for HR reporting and compliance

### Target Users
- **Primary:** All employees
- **Secondary:** HR/Admin (configuration and reporting)

---

## 2. Core Features

### 2.1 Check-In

| Aspect | Requirement |
|--------|-------------|
| **Method** | Manual button tap with GPS verification |
| **Location Rule** | Must be within 50m of office location |
| **GPS Required** | Yes — block check-in if GPS unavailable |
| **Time Availability** | 24/7 (no time window restriction) |
| **Late Marking** | Mark as "late" if after 9:00 AM (configurable) |

**User Flow:**
1. Employee opens app
2. System verifies GPS location
3. If within radius → Check-in button enabled
4. If outside radius → Show error message with distance
5. If GPS unavailable → Show error, prompt to enable
6. Employee taps check-in → Record timestamp → Show confirmation

### 2.2 Check-Out

| Aspect | Requirement |
|--------|-------------|
| **Mandatory** | No — check-out is optional |
| **Location Rule** | NOT required — employee can check out from anywhere |
| **Auto Check-Out** | 11:00 PM if employee forgets (configurable) |
| **Auto Check-Out Flag** | Mark auto check-outs distinctly for reporting |

### 2.3 Multiple Sessions

| Rule | Description |
|------|-------------|
| **Sessions per day** | Multiple allowed (employee can check out and back in) |
| **Daily count** | Increments once per calendar day only |
| **Session tracking** | All sessions recorded for detailed reporting |

**Example:** Employee checks in at 8 AM, out at 12 PM for lunch, in at 1 PM, out at 6 PM = 1 attendance day, 2 sessions recorded.

---

## 3. Attendance Streak

### Definition
Consecutive workdays with at least one check-in.

### Rules

| Rule | Description |
|------|-------------|
| **Counts** | Monday through Friday (workdays only) |
| **Weekends** | Do NOT break streak |
| **Holidays** | Do NOT break streak (uses company holiday calendar) |
| **Missed workday** | Resets streak to zero |

**Example:**
- Mon-Fri checked in (streak = 5)
- Sat-Sun no check-in (streak stays 5)
- Mon checked in (streak = 6)
- Tue missed (streak resets to 0)
- Wed checked in (streak = 1)

### Display
- Current streak count shown on dashboard
- Milestone celebrations at 5, 10, 20, 30, 50, 100 days

---

## 4. Dashboard Display

The employee dashboard shows:

| Element | Description |
|---------|-------------|
| **Check-in status** | "Checked In" or "Not Checked In" |
| **Check-in button** | Prominent, state-aware (enabled/disabled based on location) |
| **Monthly count** | "X days checked in this month" |
| **Streak count** | "Y day streak" |
| **Today's time** | Check-in time if already checked in |
| **Late indicator** | Visual marker if today's check-in was late |

---

## 5. Attendance History

Employees can view their own attendance records:

| Feature | Description |
|---------|-------------|
| **List view** | Past attendance records with check-in/out times |
| **Late indicator** | Mark late arrivals |
| **Auto-checkout indicator** | Distinguish auto from manual check-out |
| **Monthly summary** | Total days attended per month |
| **Filter** | By month |
| **Refresh** | Pull-to-refresh on mobile |

---

## 6. Notifications

| Notification | Trigger | Priority |
|--------------|---------|----------|
| **Near office reminder** | Employee within office radius but not checked in | Medium |
| **Streak milestone** | Reaching 5, 10, 20, 30, 50, 100 day streak | Low |
| **Check-out reminder** | End of workday (e.g., 6 PM) if still checked in | Medium |

All notifications should be configurable (enable/disable in settings).

---

## 7. Admin Configuration

HR/Admin can configure:

| Setting | Default | Description |
|---------|---------|-------------|
| **Office location** | — | Latitude, longitude of office |
| **Check-in radius** | 50m | Acceptable distance from office |
| **Late threshold** | 9:00 AM | Time after which check-in is marked late |
| **Auto check-out time** | 11:00 PM | Time for automatic check-out |
| **Holiday calendar** | — | Company holidays (excluded from streak calculation) |

---

## 8. Business Rules Summary

| ID | Rule |
|----|------|
| BR-01 | Check-in requires GPS location within 50m of office |
| BR-02 | Check-in blocked if GPS unavailable |
| BR-03 | Check-out is optional and does NOT require location verification |
| BR-04 | Auto check-out at 11 PM if not manually done |
| BR-05 | Multiple sessions per day allowed |
| BR-06 | Daily attendance count = max 1 per calendar day |
| BR-07 | Streak = consecutive workdays (Mon-Fri) |
| BR-08 | Weekends and holidays do not break streak |
| BR-09 | Missed workday resets streak to zero |
| BR-10 | Late = check-in after 9:00 AM (configurable) |
| BR-11 | All timestamps in company timezone |

---

## 9. Data for Reporting

The system captures data to support:

| Report Type | Data Used |
|-------------|-----------|
| **Punctuality** | Check-in times, late flags |
| **Working hours** | Check-in/out timestamps, session durations |
| **Attendance rate** | Days attended vs. workdays in period |
| **Consistency** | Streak history, patterns |
| **Compliance** | Location verification, auto-checkout frequency |

Detailed reporting requirements to be documented separately.

---

## 10. Platform Implementation Notes

This requirement applies to:

| Platform | Notes |
|----------|-------|
| **Mobile App** | Primary platform — GPS via device, push notifications |
| **Web App** | May use IP-based location or manual with approval workflow |
| **Admin Portal** | Configuration interface for HR |

Platform-specific details documented in:
- `docs/PLATFORMS/MOBILE-APP/EP-003-attendance-management/`
- `docs/PLATFORMS/WEB-APP/EP-XXX-attendance-management/` (when created)

---

## 11. Open Questions

| Question | Status |
|----------|--------|
| Multiple office locations support in future? | Not in v1 |
| Remote work / work-from-home check-in? | Out of scope for v1 |
| Manager approval for late arrivals? | Out of scope for v1 |
| Integration with payroll system? | To be discussed |

---

## 12. Acceptance Criteria Summary

- [ ] Employee can check in only when within 50m of office
- [ ] Check-in blocked with clear message when GPS unavailable
- [ ] Check-in after 9 AM marked as late
- [ ] Check-out is optional, auto-checkout at 11 PM
- [ ] Multiple sessions per day supported
- [ ] Dashboard shows: status, monthly count, streak
- [ ] Streak correctly handles weekends and holidays
- [ ] Employee can view attendance history
- [ ] Notifications work for: near-office, streak milestone, checkout reminder
- [ ] Admin can configure: office location, radius, late threshold

---

**Document Version:** 1.0
**Author:** BA Team
