---
document_type: DETAIL_REQUIREMENT
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-003
story_id: US-001
story_name: "Daily Attendance"
detail_id: DR-003-001-01
detail_name: "Check-In/Out"
status: draft
version: "1.0"
created_date: "2026-04-18"
last_updated: "2026-04-18"
related_documents:
  - path: "../REQUIREMENTS.md"
    relationship: parent
  - path: "../EPIC.md"
    relationship: grandparent
  - path: "../../docs/SHARED/ATTENDANCE_REQUIREMENTS.md"
    relationship: reference
input_sources:
  - type: text
    description: "User confirmation of dashboard widget layout (Option B)"
    extraction_date: "2026-04-18"
  - type: document
    description: "ATTENDANCE_REQUIREMENTS.md business rules"
    extraction_date: "2026-04-18"
  - type: figma
    file_id: "YEHeFgVZau7wmo9BZBVuZC"
    node_id: "3296:1556"
    frame_name: "Checked In/Out"
    extraction_date: "2026-04-19"
---

# Detail Requirement: Check-In/Out

**Detail ID:** DR-003-001-01
**Parent Requirement:** FR-US-001-01 through FR-US-001-23
**Story:** US-001-daily-attendance
**Epic:** EP-003 (Attendance Management)
**Status:** Draft
**Version:** 1.0

---

## 1. Use Case Description

As an **employee**, I want to **check in and check out from work using my mobile phone** so that **my attendance is recorded accurately with location verification**.

**Purpose:** Enable employees to record their daily work attendance by tapping a button on the mobile app dashboard. The system verifies the employee is physically present at the office (within 50m radius) using GPS before allowing check-in/out. This ensures accurate attendance tracking while providing a simple, one-tap experience.

**Target Users:** All employees with access to the Exnodes HRM mobile application.

**Key Functionality:**
- GPS-verified check-in (location required)
- Check-out from anywhere (no location required)
- Dashboard widget with status display and action button
- Multiple sessions per day support
- Late arrival detection and marking
- Auto check-out for forgotten check-outs

---

## 2. User Workflow

**Entry Point:** Home Dashboard — the check-in/out widget is embedded as a card on the main dashboard screen. No separate navigation required.

**Preconditions:**
- Employee is authenticated and signed into the mobile app
- Device has GPS/Location services capability
- App has location permission granted

### Main Flow: Check-In

```
1. Employee opens app → lands on Home Dashboard
2. System displays Attendance Widget with current status
3. System silently acquires GPS location in background
4. IF within 50m of office:
   → Check-In button is ENABLED (green/primary style)
   → Location indicator shows "At Office" (green dot)
5. IF outside 50m of office:
   → Check-In button is DISABLED (muted style)
   → Location indicator shows "Not at Office" with distance
6. IF GPS unavailable:
   → Check-In button is DISABLED
   → Location indicator shows "Location unavailable" with prompt
7. Employee taps Check-In button
8. System records timestamp and location
9. System updates widget to show "Checked In" status with time
10. System triggers haptic feedback + success animation
11. IF check-in is after 9:00 AM → mark as "Late"
```

### Main Flow: Check-Out

```
1. Employee is already checked in (widget shows "Checked In" status)
2. System displays Check-Out button (replaces Check-In button)
3. Check-Out button is ALWAYS ENABLED (no location verification required)
4. Employee taps Check-Out button
5. System records check-out timestamp
6. System updates widget to show "Checked Out" status
7. System triggers haptic feedback + success animation
8. IF employee wants to check in again → flow returns to Check-In
```

**Note:** Unlike check-in, check-out does NOT require the employee to be at the office. Employees can check out from anywhere.

### Alternative Flows

- **Alt 1: GPS Permission Not Granted**
  1. System shows error banner: "Location access required"
  2. Tapping banner opens device settings
  3. After granting permission, widget refreshes

- **Alt 2: GPS Signal Weak/Unavailable**
  1. System shows "Acquiring location..." with spinner
  2. After 10 seconds timeout → show "Location unavailable"
  3. Button remains disabled until GPS acquired

- **Alt 3: Multiple Sessions (Re-check-in)**
  1. Employee has checked out earlier today
  2. Widget shows "Check In" button again
  3. Employee can check in for a new session
  4. Monthly count does NOT increment (already counted for today)

- **Alt 4: Auto Check-Out (11:00 PM)**
  1. Employee forgot to check out
  2. System automatically records check-out at 11:00 PM
  3. Record is flagged as "Auto Check-Out"
  4. Next app open shows "Checked Out" status

**Exit Points:**
- **Success (Check-In):** Widget updates to "Checked In" state, streak may increment
- **Success (Check-Out):** Widget updates to "Checked Out" state
- **Blocked:** Button remains disabled with clear reason displayed
- **Error:** Toast notification with retry guidance

---

## 3. Field Definitions

### Interaction Elements (from Figma)

| Element Name | Type | Size | State/Condition | Trigger Action | Description |
|--------------|------|------|-----------------|----------------|-------------|
| Check-In Button | Button | 114x36 | Enabled when within 50m radius + GPS available | Records check-in timestamp | Gray outlined button with CalendarDots icon, positioned in greeting header |
| Check-Out Button | Button | 126x36 | Shown after check-in, ALWAYS ENABLED (no GPS required) | Records check-out timestamp | Dark filled button with CalendarDots icon, replaces Check-In |
| Chat Button | Button | 88x36 | Always visible | Opens chat with manager | Secondary button in manager info section |

### Display Elements (Dashboard)

| Element Name | Type | State/Condition | Description |
|--------------|------|-----------------|-------------|
| Date Display | Label | Always visible | Format: "MMM DD, YY" (e.g., "Dec 09, 25") with CalendarDots icon |
| Greeting | Label | Always visible | "Hello [First Name]" — 29px height |
| Hero Card | Card | Changes based on check-in state | 369x78 card with title, subtitle, and right-side element |
| Hero Title | Label + Icon | Dynamic | "Get Started" (not checked in) or "Vibing..." (checked in) |
| Hero Subtitle | Label | Dynamic | Motivational text based on state |
| Work Timer | Label + Icon | Visible when checked in | Live counter "HH:MM" with ClockCountdown icon (blue) |
| Weather Icon | Icon | Visible when not checked in | CloudSun decorative icon (40x40) |

### GPS Location Data (System-Captured)

| Field Name | Type | Validation | Description |
|------------|------|------------|-------------|
| Latitude | Decimal | Valid coordinate range | Device GPS latitude at check-in |
| Longitude | Decimal | Valid coordinate range | Device GPS longitude at check-in |
| Accuracy | Decimal (meters) | Must be <= 50m | GPS accuracy at capture time |
| Timestamp | DateTime | Server-generated | UTC timestamp of action |

**Note:** GPS data is only captured for check-in. Check-out does not require or record location.

---

## 4. Data Display

### State: Not Checked In (Check Out State)

| Element | Content | Style |
|---------|---------|-------|
| Action Button | "Check In" with calendar icon | Gray outlined button (114x36) |
| Hero Card Title | "Get Started" | Green text with FlagBannerFold icon |
| Hero Card Subtitle | "Let's go to work and clock in in time" | Muted text |
| Hero Card Right | CloudSun icon (weather) | 40x40 decorative |

### State: Checked In

| Element | Content | Style |
|---------|---------|-------|
| Action Button | "Check Out" with calendar icon | Dark filled button (126x36) |
| Hero Card Title | "Vibing..." | Green text with Headset icon |
| Hero Card Subtitle | "Working hard for better life..." | Muted text |
| Hero Card Right | Timer "03:16" with ClockCountdown icon | Blue text, live counter |

### Information Shown to User

| Data Name | Data Type | Display When Empty | Format | Business Meaning |
|-----------|-----------|-------------------|--------|------------------|
| Check-In Status | Enum | Button shows "Check In" | Button label | Current attendance state |
| Work Timer | Duration | Hidden | "HH:MM" | Time elapsed since check-in (counts UP, live counter) |
| Late Indicator | Boolean | Hidden | Red "Late" badge | Whether today's check-in was after 9 AM |


### Display States

| State | Condition | What User Sees |
|-------|-----------|----------------|
| Loading | Widget initializing, acquiring GPS | Skeleton card with pulsing animation |
| Ready - At Office | GPS acquired, within 50m | Green "At Office" badge, enabled Check-In button |
| Ready - Not at Office | GPS acquired, outside 50m | Gray "Not at Office (Xm away)" badge, disabled button |
| Ready - No GPS | GPS unavailable or denied | Yellow "Location unavailable" badge, disabled button, settings prompt |
| Checked In | Employee has checked in today | "Checked In at [time]" text, Check-Out button shown |
| Checked In - Late | Checked in after 9:00 AM | Same as above + red "Late" badge |
| Checked Out | Employee has checked out | "Checked Out at [time]" text, Check-In button for new session |
| Processing | Check-in/out in progress | Button shows spinner, disabled |
| Error | Action failed | Toast message with error, button re-enabled |

---

## 5. Acceptance Criteria

**Definition of Done - All criteria must be met:**

- **AC-01:** Check-In button is enabled ONLY when employee's GPS location is within 50 meters of the configured office coordinates
- **AC-02:** Check-In button is disabled with clear visual indicator when GPS is unavailable or permission not granted
- **AC-03:** Check-In button is disabled with distance display when employee is outside 50m radius
- **AC-04:** Upon successful check-in, system records timestamp and displays confirmation with haptic feedback
- **AC-05:** Check-in after 9:00 AM (configurable threshold) is automatically marked as "Late" with visible badge
- **AC-06:** After check-in, the Check-In button is replaced with Check-Out button
- **AC-07:** Check-out is optional — employee is not required to check out
- **AC-08:** Check-out does NOT require GPS verification — employee can check out from anywhere
- **AC-09:** After check-out, employee can check in again for a new session (same day)
- **AC-10:** Multiple sessions per day are allowed, but monthly attendance count increments only once per calendar day
- **AC-11:** If employee does not check out by 11:00 PM, system performs auto check-out
- **AC-12:** Auto check-out records are marked distinctly from manual check-outs
- **AC-13:** Monthly count displays total unique days with at least one check-in this month
- **AC-14:** Streak count displays consecutive workdays (Mon-Fri) with at least one check-in
- **AC-15:** Streak calculation correctly excludes weekends (Sat-Sun do not break streak)
- **AC-16:** Streak calculation correctly excludes company holidays (holidays do not break streak)
- **AC-17:** Missing a workday (Mon-Fri, non-holiday) resets streak to zero
- **AC-18:** Widget displays location status in real-time (refreshes when location changes)
- **AC-19:** Manual refresh button allows employee to re-acquire GPS location
- **AC-20:** GPS location acquisition completes within 5 seconds under normal conditions

**Testing Scenarios:**

| Scenario | Input | Expected Output | Priority |
|----------|-------|-----------------|----------|
| Check-in at office | GPS within 50m, tap Check-In | Timestamp recorded, status updates, streak may increment | Critical |
| Check-in outside office | GPS outside 50m | Button disabled, "Not at Office (Xm away)" shown | Critical |
| Check-in no GPS | GPS unavailable | Button disabled, "Location unavailable" shown | Critical |
| Check-in late | Check-in at 9:15 AM | Record marked "Late", badge displayed | High |
| Check-out anywhere | Tap Check-Out from any location | Check-out recorded regardless of location | High |
| Multiple sessions | Check-out then check-in again | Second session recorded, monthly count unchanged | High |
| Auto check-out | No manual check-out by 11 PM | System records auto check-out, flagged | High |
| Streak continues | Check-in Mon-Fri, skip Sat-Sun, check-in Mon | Streak = 6 | High |
| Streak breaks | Miss a Wednesday | Streak resets to 0 | High |
| Holiday handling | Company holiday (no check-in) | Streak NOT reset | Medium |

---

## 6. System Rules

**Business Logic & Backend Behavior:**

- **Rule 1:** GPS verification uses Haversine formula to calculate distance between device coordinates and office coordinates; check-in allowed only if distance <= 50 meters
- **Rule 2:** GPS accuracy must be 50m or better; if device reports accuracy > 50m, treat as "GPS unavailable"
- **Rule 3:** All timestamps stored in UTC; displayed in company timezone
- **Rule 4:** Monthly attendance count = COUNT(DISTINCT date) WHERE employee has at least one check-in that day
- **Rule 5:** Auto check-out job runs at 11:00 PM company timezone; marks all open sessions as auto-closed
- **Rule 6:** Streak calculation:
  - Get all workdays (Mon-Fri minus company holidays) in reverse chronological order
  - Count consecutive days with at least one check-in
  - Stop counting at first workday without check-in
- **Rule 7:** Late threshold comparison uses company timezone (default 9:00 AM)
- **Rule 8:** Location data (lat/long) is recorded with check-in for audit trail; check-out does not require or record location

**State Transitions:**

```
[Not Checked In] → [Check-In Action] → [Checked In]
[Checked In] → [Check-Out Action] → [Checked Out]
[Checked Out] → [Check-In Action] → [Checked In] (new session)
[Checked In] → [11:00 PM Auto Job] → [Auto Checked Out]
```

**Calculations/Formulas:**

- **Distance:** `d = haversine(user_lat, user_lng, office_lat, office_lng)` in meters
- **Monthly Count:** `SELECT COUNT(DISTINCT DATE(check_in_time)) FROM attendance WHERE employee_id = ? AND MONTH(check_in_time) = CURRENT_MONTH`
- **Streak:** Server-calculated based on workday calendar and attendance records

**Dependencies:**

- EP-001 US-001: Authentication — user must be signed in
- EP-001 US-002: Navigation & Layout — dashboard integration
- Admin Portal: Office location coordinates and late threshold configuration
- System: Company holiday calendar integration
- Device: GPS/Location services availability

---

## 7. UX Optimizations

**Usability Considerations:**

- **Optimization 1:** Check-in button is minimum 56x56 points (larger than standard 44x44) for easy one-handed tap
- **Optimization 2:** Haptic feedback on successful check-in/out provides confirmation without requiring visual attention
- **Optimization 3:** Location status updates in real-time — if employee walks into office radius, button enables without manual refresh
- **Optimization 4:** Streak milestone celebrations (5, 10, 20, 30, 50, 100 days) trigger confetti animation and celebratory toast
- **Optimization 5:** Widget uses skeleton loading to prevent layout shift during GPS acquisition
- **Optimization 6:** Error messages are actionable (e.g., "Location unavailable — tap to open settings")

**Responsive Behavior:**

| Screen Size | Adaptation |
|-------------|------------|
| Standard iPhone (375px) | Widget spans full width with horizontal padding |
| Large iPhone / Android (414px+) | Same layout, larger touch targets |
| Small devices (320px) | Compact metrics row, button size maintained |

**Accessibility Requirements:**

- [x] All interactive elements have minimum 44x44 touch targets
- [x] Button states (enabled/disabled) announced to screen reader
- [x] Location status changes announced automatically
- [x] Color is not the only indicator of state (text labels accompany badges)
- [x] Haptic feedback supplements visual confirmation
- [x] High contrast mode supported

**Design References (from Figma):**

- **Figma File:** YEHeFgVZau7wmo9BZBVuZC, Node: 3296:1556
- **Check-In Button:** Gray outlined, 114x36, CalendarDots icon prefix
- **Check-Out Button:** Dark filled (#171717), 126x36, CalendarDots icon prefix
- **Hero Card:** 369x78, rounded corners, contains title + subtitle + right element
- **Metrics Cards:** 116x78 each, 3-column layout with icons and labels
- **Work Timer:** Blue text (#2563eb) with ClockCountdown icon
- **Hero Title (not checked in):** Green (#22c55e) "Get Started" with FlagBannerFold icon
- **Hero Title (checked in):** Green (#22c55e) "Vibing..." with Headset icon
- **Icons:** Phosphor Icons family (CalendarDots, Lightning, ClockCountdown, CloudSun, Headset, FlagBannerFold)

---

## 8. Additional Information

### Out of Scope

- Multiple office locations (v1 supports single office only)
- Remote work / work-from-home check-in (out of scope for v1)
- Manager approval workflow for late arrivals
- Offline check-in with later sync (requires GPS in real-time)
- Editing or correcting past attendance records (admin-only feature)
- Face recognition or biometric verification beyond device unlock
- Integration with payroll system (separate requirement)
- Geofence-triggered notifications (separate DR-003-001-04)

### Open Questions

- None — all requirements confirmed via ATTENDANCE_REQUIREMENTS.md

### Features from Figma — Scope Clarification

| Feature | Description | Status |
|---------|-------------|--------|
| **Work Timer** | Live "HH:MM" counter counting UP from check-in time (visible when checked in) | ✅ IN SCOPE |
| **Leaderboard Rank** | Shows employee's attendance ranking (e.g., "3rd place") with avatar stack | ❌ OUT OF SCOPE |
| **"In Time Streak"** | Streak counts consecutive ON-TIME check-ins (before 9 AM) | ❌ OUT OF SCOPE — use original "consecutive workdays" definition |
| **Hero Card States** | Motivational card with different content based on check-in state | ✅ IN SCOPE (design enhancement) |
| **Manager Info** | Shows direct manager with Chat button | ❌ OUT OF SCOPE for attendance (separate feature) |

### Related Features

- **DR-003-001-02:** Attendance Dashboard Widget (this DR covers the check-in/out functionality; DR-002 covers the broader dashboard display)
- **DR-003-001-03:** Attendance History (view past records)
- **DR-003-001-04:** Attendance Notifications (push notification reminders)
- **EP-001 US-001:** Authentication (dependency)
- **EP-001 US-002:** Navigation & Layout (dashboard integration)

### Notes

- GPS accuracy in buildings may be reduced; 50m radius provides tolerance for indoor positioning errors
- Battery optimization: Use geofencing for background location rather than continuous GPS polling
- Auto check-out at 11 PM is a server-side scheduled job, not dependent on app being open
- Streak calculations are performed server-side to ensure consistency across devices

---

## Approval & Sign-off

| Role | Name | Date | Status |
|------|------|------|--------|
| Business Analyst | BA Team | 2026-04-18 | Draft |
| Product Owner | — | — | Pending |
| UX Designer | — | — | Pending |
| Tech Lead | — | — | Pending |

---

**Document Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-04-18 | BA Team | Initial draft — dashboard widget layout confirmed (Option B) |
| 1.1 | 2026-04-19 | BA Team | Updated: Check-out does NOT require GPS/location — can be done from anywhere |
| 1.2 | 2026-04-19 | BA Team | Updated layout from Figma design: button in greeting header, hero card with state changes, work timer (counts up from check-in). Scope clarified: leaderboard and in-time streak OUT OF SCOPE |
