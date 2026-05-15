---
document_type: EPIC
platform: MOBILE-APP
platform_display: "Exnodes HRM Mobile"
epic_id: EP-003
epic_name: "Attendance Management"
status: approved
version: "1.0"
created_date: "2026-04-18"
last_updated: "2026-04-18"
approved_by: "Product Owner"
related_documents:
  - path: "../EP-001-foundation/EPIC.md"
    relationship: dependency
  - path: "../EP-002-leave-management/EPIC.md"
    relationship: sibling
user_stories:
  - id: US-001
    name: "Daily Attendance"
    status: in_progress
    description: "Check-in/out functionality with GPS verification, streak tracking, and attendance history"
---

# Epic: Attendance Management

**Epic ID:** EP-003
**Platform:** Exnodes HRM Mobile (MOBILE-APP)
**Status:** Approved
**Version:** 1.0

---

## Epic Overview

### Business Objective

Enable employees to record their daily attendance through the mobile application with location-based verification. The system tracks check-in/out times, calculates attendance streaks, and provides data for organizational reporting and compliance.

### Scope

- GPS-verified check-in/out functionality
- Attendance streak tracking (consecutive workdays)
- Monthly attendance summary
- Personal attendance history
- Push notifications (check-in reminders, streak milestones)
- Late arrival marking
- Admin-configurable settings (office location, late threshold)

### User Stories

| ID | Story Name | Description | Status |
|----|-----------|-------------|--------|
| US-001 | Daily Attendance | Check-in/out with GPS, streak tracking, history view | In Progress |

### Dependencies

- **EP-001 (Foundation):** Authentication (US-001) — user must be signed in
- **EP-001 (Foundation):** Navigation & Layout (US-002) — dashboard integration
- **System:** GPS/Location services availability
- **System:** Company holiday calendar integration
- **Admin Portal:** Office location and late threshold configuration

### Success Criteria

- Employees can check in/out only when within 50m of office location
- System accurately tracks attendance streaks (excluding weekends/holidays)
- Employees can view their attendance history and monthly summary
- Late arrivals (after 9:00 AM) are marked appropriately
- Auto-checkout occurs at 11 PM for forgotten check-outs
- Push notifications remind employees to check in/out

---

## Mobile-Specific Considerations

### Differences from WEB-APP

| Aspect | WEB-APP | MOBILE-APP |
|--------|---------|------------|
| Check-in Method | IP-based or manual | GPS-verified with manual tap |
| Location Verification | Office network detection | 50m radius GPS check |
| Primary View | Table/calendar view | Dashboard widget + history |
| Notifications | Email/in-app | Push notifications |
| Offline Behavior | Not applicable | Block check-in (requires GPS) |

### Mobile UX Patterns

- Prominent check-in/out button on dashboard
- Real-time location status indicator
- Pull-to-refresh for history updates
- Bottom sheet for attendance details
- Streak celebration animations
- Geofence-triggered reminders

---

## Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| GPS inaccuracy in buildings | Medium | High | Use 50m radius tolerance; consider Wi-Fi assist |
| Battery drain from location services | Medium | Medium | Use geofencing, not continuous GPS |
| Timezone issues for remote workers | Low | Medium | Use server timezone for all calculations |
| Holiday calendar sync failures | Low | Medium | Cache holiday data; manual override option |

---

## Change Log

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-04-18 | Initial EPIC created | BA Team |

---

**Document Status:** Approved
**Approval Required:** No - EPIC approved, stories can be created
**Last Updated:** 2026-04-18
