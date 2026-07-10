# Dashboard API - FE Integration Guide

**Date:** 2026-07-09  
**Backend:** Go HRM API, base path `/api/v1`. Endpoint requires `Authorization: Bearer <access_token>`.  
**Source of truth:** Swagger at `/swagger/index.html`. Drop this into the web repo as `api_info_go/dashboard.md`.

---

## What this endpoint does

`GET /api/v1/dashboard` returns the common web dashboard as a fixed-order list of widgets. There is **no dashboard-specific permission**. The endpoint is JWT-only, and each widget appears only when the caller has the relevant source-module permission and source data scope.

Widget order is stable:

1. `attendance_overview`
2. `pending_approvals`
3. `leave_summary`
4. `announcements`
5. `holidays_workdays`
6. `workforce_summary`

**Request Tickets is intentionally absent** in this release because the Request Tickets module is not implemented yet. Do not wait for a `request_tickets` widget from this endpoint until that module lands.

---

## Endpoint

| Method | Path | Permission | Returns |
|---|---|---|---|
| GET | `/dashboard` | authenticated | `DashboardRead` |

### Response envelope

```json
{
  "success": true,
  "data": {
    "greeting": {
      "name": "Jane Smith",
      "email": "jane.smith@example.com"
    },
    "widgets": [
      { "id": "attendance_overview" },
      { "id": "leave_summary" }
    ],
    "empty": false
  }
}
```

`greeting.name` is the employee full name when the user has an employee profile; otherwise it falls back to the auth email.

When no widget is visible:

```json
{
  "success": true,
  "data": {
    "greeting": {
      "name": "auth-only@example.com",
      "email": "auth-only@example.com"
    },
    "widgets": [],
    "empty": true,
    "empty_message": "No dashboard items are available yet."
  }
}
```

---

## Shared shapes

### `DashboardRead`

```json
{
  "greeting": { "name": "Jane Smith", "email": "jane.smith@example.com" },
  "widgets": [ /* DashboardWidgetRead[] */ ],
  "empty": false,
  "empty_message": "optional string"
}
```

### `DashboardWidgetRead`

```json
{
  "id": "leave_summary",
  "title": "Leave Summary",
  "scope": "own",
  "metrics": [
    { "key": "annual_remaining", "label": "Annual Remaining", "value": 11 }
  ],
  "items": [
    {
      "id": "uuid",
      "title": "Jane Smith - annual leave",
      "description": "Family trip",
      "status": "pending",
      "date": "2026-07-10T00:00:00Z",
      "url": "/leave-requests/uuid"
    }
  ],
  "actions": [
    { "key": "create_leave_request", "label": "Create Leave Request", "url": "/leave-requests/new" }
  ],
  "empty_message": "No upcoming leave."
}
```

Notes:
- `metrics[].value` can be a number or string. For example, `today_status` is a string; count/balance metrics are numbers.
- `items[]`, `actions[]`, and `empty_message` are optional UI helpers. The FE can render widgets from metrics alone when item lists are empty.
- `url` values are route hints, not hard requirements. Map them to the current web routing layer if paths differ.
- There is no dashboard timestamp by design.

---

## Widget visibility and payloads

### 1. Attendance Overview - `attendance_overview`

Appears when the caller has `attendance:read` or `attendance:manage_data`.

#### Own scope

Returned for callers with attendance read access but without manage-all scope, and only when the user has an employee profile.

```json
{
  "id": "attendance_overview",
  "title": "Attendance Overview",
  "scope": "own",
  "metrics": [
    { "key": "today_status", "label": "Today Status", "value": "checked_in" },
    { "key": "monthly_check_ins", "label": "Monthly Check-ins", "value": 7 }
  ],
  "items": [],
  "empty_message": "No attendance activity for today."
}
```

`today_status` values:
- `not_checked_in`
- `checked_in`
- `checked_out`

If the user's first check-in today was late, `items` includes one `late` item.

#### Organization scope

Returned when the caller has `attendance:manage_data`.

```json
{
  "id": "attendance_overview",
  "title": "Attendance Overview",
  "scope": "organization",
  "metrics": [
    { "key": "active_employees", "label": "Active Employees", "value": 42 },
    { "key": "present_today", "label": "Present Today", "value": 31 },
    { "key": "late_today", "label": "Late Today", "value": 3 },
    { "key": "absent_today", "label": "Absent Today", "value": 11 }
  ],
  "items": [
    { "id": "attendance_uuid", "title": "Late Employee", "status": "late", "date": "2026-07-09T00:00:00Z", "url": "/attendance" }
  ]
}
```

`items` contains up to 5 late attendance rows for today.

---

### 2. Pending Approvals - `pending_approvals`

Appears when the caller can approve leave:

| Permission | Scope |
|---|---|
| `leave_requests:approve_team` | `team` |
| `leave_requests:approve_all` | `organization` |
| legacy `leave_requests:approve` | `organization` |
| `leave_requests:manage` | `organization` |
| `*` | `organization` |

```json
{
  "id": "pending_approvals",
  "title": "Pending Approvals",
  "scope": "team",
  "metrics": [
    { "key": "pending_leave", "label": "Pending Leave", "value": 2 }
  ],
  "items": [
    {
      "id": "leave_uuid",
      "title": "Report Employee - personal leave",
      "description": "pending approval",
      "status": "pending",
      "date": "2099-01-02T00:00:00Z",
      "url": "/leave-requests/leave_uuid"
    }
  ],
  "empty_message": "No approvals are waiting."
}
```

Team scope uses the current employee's transitive reporting chain. If the approver has no employee profile, the team widget is omitted.

---

### 3. Leave Summary - `leave_summary`

Appears when the caller has an employee profile and any leave permission, including self-service permissions such as `leave_requests:read` or `leave_requests:create`.

```json
{
  "id": "leave_summary",
  "title": "Leave Summary",
  "scope": "own",
  "metrics": [
    { "key": "annual_remaining", "label": "Annual Remaining", "value": 12 },
    { "key": "sick_remaining", "label": "Sick Remaining", "value": 6 },
    { "key": "pending_requests", "label": "Pending Requests", "value": 1 },
    { "key": "upcoming_leave", "label": "Upcoming Leave", "value": 1 }
  ],
  "items": [
    {
      "id": "leave_uuid",
      "title": "Jane Smith - annual leave",
      "description": "planned leave",
      "status": "pending",
      "date": "2026-07-10T00:00:00Z",
      "url": "/leave-requests/leave_uuid"
    }
  ],
  "actions": [
    { "key": "create_leave_request", "label": "Create Leave Request", "url": "/leave-requests/new" }
  ],
  "empty_message": "No upcoming leave."
}
```

The `create_leave_request` quick action appears only when the caller has `leave_requests:create`.

Balance defaults match the leave module: annual quota defaults to 12 and sick quota defaults to 6 when no explicit quota row exists.

---

### 4. Announcements - `announcements`

Appears when the caller has `announcements:read` or `announcements:manage`.

```json
{
  "id": "announcements",
  "title": "Announcements",
  "scope": "targeted",
  "metrics": [
    { "key": "latest_sent", "label": "Latest Sent", "value": 4 }
  ],
  "items": [
    {
      "id": "announcement_uuid",
      "title": "Published update",
      "status": "published",
      "date": "2026-07-09T08:00:00Z",
      "url": "/announcements/announcement_uuid"
    }
  ],
  "empty_message": "No announcements yet."
}
```

Scope rules:
- `announcements:manage` or `*` -> `scope: "organization"` and can see all published announcements.
- `announcements:read` only -> `scope: "targeted"` and only sees published announcements visible to that employee by audience rules.
- If a read-only caller has no employee profile, the widget is omitted.

Backend permission note for FE role-management:
- New permission: `announcements:read`
- Existing `announcements:manage` effectively implies read server-side, so existing manage-only custom roles keep working.
- The standard seeded roles now receive `announcements:read` for Admin, HR Manager, Manager, and Employee. Super Admin keeps `*`.

Announcement read endpoints also now require `announcements:read`:

| Method | Path |
|---|---|
| GET | `/announcements` |
| GET | `/announcements/:id` |
| POST | `/announcements/:id/view` |
| GET | `/mobile/announcements` |
| GET | `/mobile/announcements/list` |
| GET | `/mobile/announcements/:id` |
| POST | `/mobile/announcements/:id/read` |

---

### 5. Holidays & Workdays - `holidays_workdays`

Appears when the caller has either holiday view/manage permission or workday view permission.

```json
{
  "id": "holidays_workdays",
  "title": "Holidays & Workdays",
  "scope": "organization",
  "metrics": [
    { "key": "upcoming_holidays", "label": "Upcoming Holidays", "value": 1 },
    { "key": "current_month_workdays", "label": "Current Month Workdays", "value": 23 }
  ],
  "items": [
    {
      "id": "holiday_uuid",
      "title": "Company Holiday",
      "date": "2026-07-10T00:00:00Z",
      "url": "/holidays"
    }
  ],
  "empty_message": "No holidays in the next 7 days."
}
```

Metric visibility:
- `upcoming_holidays` requires `organization:holidays_view` or `organization:holidays_manage`.
- `current_month_workdays` requires `organization:workdays_view`.

`current_month_workdays` counts weekdays in the current month. It does **not** subtract holidays, matching the Monthly Workdays API v1.1 behavior.

---

### 6. Workforce Summary - `workforce_summary`

Appears when the caller has `employees:read` or `users:read`.

```json
{
  "id": "workforce_summary",
  "title": "Workforce Summary",
  "scope": "organization",
  "metrics": [
    { "key": "active_employees", "label": "Active Employees", "value": 42 },
    { "key": "current_month_joiners", "label": "Current Month Joiners", "value": 3 }
  ],
  "items": [
    {
      "id": "employee_uuid",
      "title": "New Joiner",
      "date": "2026-07-01T00:00:00Z",
      "url": "/employees/employee_uuid"
    }
  ],
  "empty_message": "No new joiners this month."
}
```

`items` contains up to 5 employees whose `join_date` falls in the current month.

---

## Recommended FE rendering approach

Render widgets by `id`, not by array index. The backend keeps the array in product order, but `id` is the stable contract.

Suggested component map:

```ts
const dashboardWidgets = {
  attendance_overview: AttendanceOverviewWidget,
  pending_approvals: PendingApprovalsWidget,
  leave_summary: LeaveSummaryWidget,
  announcements: AnnouncementsWidget,
  holidays_workdays: HolidaysWorkdaysWidget,
  workforce_summary: WorkforceSummaryWidget,
} satisfies Record<string, React.ComponentType<{ widget: DashboardWidgetRead }>>;
```

Keep unknown widget IDs non-fatal:

```ts
for (const widget of data.widgets) {
  const Widget = dashboardWidgets[widget.id];
  if (!Widget) continue;
  render(<Widget widget={widget} />);
}
```

This lets FE tolerate the future `request_tickets` widget once that source module is implemented.

---

## TypeScript reference

```ts
type DashboardRead = {
  greeting: DashboardGreetingRead;
  widgets: DashboardWidgetRead[];
  empty: boolean;
  empty_message?: string;
};

type DashboardGreetingRead = {
  name: string;
  email: string;
};

type DashboardWidgetRead = {
  id: string;
  title: string;
  scope: "own" | "team" | "targeted" | "organization" | string;
  metrics: DashboardMetricRead[];
  items?: DashboardItemRead[];
  actions?: DashboardActionRead[];
  empty_message?: string;
};

type DashboardMetricRead = {
  key: string;
  label: string;
  value: string | number | boolean | null;
};

type DashboardItemRead = {
  id?: string;
  title: string;
  description?: string;
  status?: string;
  date?: string;
  url?: string;
};

type DashboardActionRead = {
  key: string;
  label: string;
  url: string;
};
```

---

## Error behavior

This endpoint is read-only.

| Status | Typical reason |
|---|---|
| 401 | Missing/invalid/expired access token |
| 500 | Unexpected source-module read error |

Lack of a source permission is **not** an error; the widget is simply omitted.

---

## Not in this release

- Request Tickets widget - blocked until the Request Tickets module exists.
- Per-user dashboard layout customization - fixed common layout only.
- Dashboard timestamp / "last updated" field - intentionally not present.
- Dashboard mutation endpoints - quick actions are navigation hints only.

---

## Quick FE checklist

- [ ] Add `GET /api/v1/dashboard` to the dashboard page data loader.
- [ ] Render widgets by `widget.id`; do not assume every widget is present.
- [ ] Support empty dashboard state via `data.empty` and `data.empty_message`.
- [ ] Add role-management display for `announcements:read`.
- [ ] Treat `announcements:manage` as higher privilege than read in the UI.
- [ ] Do not build a Request Tickets dashboard card from this endpoint yet.
- [ ] Keep metric formatting tolerant: `value` can be string or number.
