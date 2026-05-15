# Half-day Leave + Check-in/Out — Attendance List Cell Design

**Date:** 2026-05-08
**Target DR:** [DR-004-001-01 Attendance List](../../PLATFORMS/WEB-APP/EP-004-attendance-management/US-001-attendance-list/details/DR-004-001-01-attendance-list.md) (will be revised to v1.2)
**Status:** Approved (brainstorming → DR revision)

---

## Problem

DR-004-001-01 v1.1 lists `½ Half-day Leave` as a single cell-status icon, but does not define what happens when a date has **both** an approved half-day leave **and** a check-in/out record from the worked half. Specifically, none of the following are specified:

- Cell symbol when half-day + check-in coexist
- Cell background when half-day + check-in coexist
- Tooltip content
- Late-time computation (which scheduled boundary applies?)
- Early-time computation (which end-of-half applies?)
- Status filter matching (does this row match "On Leave"? "Late"? Both?)
- Edge case where the worked half has no check-in

This is a real-world combination since `EP-002 DR-002-001-02 Create Leave Request §SR-10` allows employees to take Morning-Half or Afternoon-Half leave (counted as 0.5 days) and continue to work the remaining half.

---

## Decisions

### Cell rendering — diagonal split

When a date has approved half-day leave AND a check-in/out record, the cell is rendered as a **diagonal split** rather than a single icon:

| Position | Content |
|---|---|
| Leave-half corner (top-left for AM, bottom-right for PM) | Light-blue background (`#e0f2fe`) + `½` glyph in leave color (`#0369a1`) |
| Worked-half corner | Attendance-status background + status glyph (`✓` on-time, `L` late, `A` absent-no-checkin) |

The diagonal corner that holds the leave half mirrors the half-of-day: AM leave occupies the top-left, PM leave occupies the bottom-right. This visual metaphor lets a viewer read the combined day at a glance.

### Tooltip — two stacked sections

The hover tooltip is split into two labeled sections separated by a divider:

```
Tuesday, April 8, 2026
─────────────────────────────
Morning Half — Annual Leave
Approved by: John Manager
─────────────────────────────
Afternoon — Worked
Check-in:   13:05 PM
Check-out:  18:00 PM
Status:     On-time ✓
```

If the worked half has no check-in, the second section reads:
```
Afternoon — Absent
No check-in recorded
```

### Late / Early time math — worked-half boundaries (Approach 1)

The Total Late Time and Total Early Time monthly summary columns use the **worked half's** schedule boundaries. The leave half always contributes 0 to both.

| Worked half | Late threshold (used for Total Late Time) | End-of-half (used for Total Early Time) |
|---|---|---|
| AM half worked, PM half on leave | 09:00 | 12:00 |
| AM half on leave, PM half worked | 13:15 | 18:00 |

Per-day contributions:
- `Total Late Time += max(0, first_check_in − late_threshold_of_worked_half)`
- `Total Early Time += max(0, end_of_worked_half − last_check_out)`
- Days where the worked half has no check-in/check-out contribute 0 (the day shows `A` on the worked half but does not penalise totals; aligns with full-day Absent rule)

### Schedule constants — confirmed

| Constant | Value |
|---|---|
| Workday start (AM late threshold) | 09:00 |
| AM half end / lunch start | 12:00 |
| Lunch end / PM late threshold | 13:15 |
| PM half end / workday end | 18:00 |

These resolve two open questions from DR-004-001-01 v1.1:
- "Scheduled end-of-day time used to compute Total Early Time" → **18:00**
- (New) "PM half late threshold" → **13:15** (one-hour-fifteen lunch break, 12:00–13:15)

### Status filter — multi-match on combined-cell rows

A row that has at least one half-day-leave-with-attendance day matches multiple status filter values simultaneously:

- Filter "On Leave" — matches because the day has approved leave
- Filter "Late" — matches if the worked half's check-in exceeded its threshold
- Filter "On-time" — matches if the worked half's check-in was at or before its threshold
- Filter "Absent" — matches if the worked half has no check-in

Multi-select status filters use OR logic across the values selected (existing SR-008 behavior); a combined-cell day participates in any of the values it matches.

### Edge case — no check-in on the worked half

Cell still uses diagonal split. Leave-half corner shows `½` in leave color/background; worked-half corner shows `A` in absent color/background (`#fee2e2` background, `#991b1b` glyph). Tooltip shows "Afternoon — Absent / No check-in recorded" in the worked-half section.

---

## DR-004-001-01 Sections to Revise (v1.2)

| Section | Change |
|---|---|
| §1 Key Functionality | Add bullet noting half-day leave + worked half is rendered as a diagonal split cell |
| §4 Cell Status Icons | Add a "Combined cells (half-day leave + worked half)" subsection with the four diagonal-split combinations: `½ + ✓`, `½ + L`, `½ + A` (and AM/PM positioning) |
| §4 Matrix Layout diagram | Update one example date in the ASCII matrix to show a combined cell |
| §4 Tooltip Content | Add a fifth tooltip block: "For combined half-day cells" with the two-stacked-section layout |
| §5 New ACs | AC-026..AC-030: combined-cell rendering, combined-cell tooltip, AM-worked late math, PM-worked late math, no-check-in-on-worked-half rendering |
| §6 SR-002 | Clarify both thresholds (AM late = 09:00, PM late = 13:15) |
| §6 SR-003 | Refine Absent logic to cover the worked-half-of-half-day case (the worked half is Absent if no check-in, even though the day is partially on leave) |
| §6 SR-004 | Extend Leave Integration to describe coexistence of half-day leave + attendance, including the diagonal split cell |
| §6 SR-008 | Document multi-match behavior for combined-cell rows in the Status filter |
| §6 SR-011 | Refine Total Late Time / Total Early Time computation: when the day has half-day leave, use the worked half's late threshold and end-of-half boundary |
| §8 Open Questions | Resolve two existing entries: scheduled end-of-day → 18:00, and document the new PM half threshold = 13:15 |
| Version history | Add v1.2 entry summarising the half-day-combined-cell handling |

---

## Out of Scope

- Quarter-day leave (only Full Day / Morning Half / Afternoon Half are supported in EP-002)
- Two separate half-day leaves on the same day from different leave types (e.g., AM Annual + PM Sick) — not supported by the leave model
- Manual override of the diagonal split presentation per role
- Configuration of the workday boundaries via UI (workday/lunch times come from system configuration)
