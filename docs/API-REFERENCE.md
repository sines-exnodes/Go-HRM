# Exnodes HRM API — Reference (FE + Mobile)

**Base URL:** `http://localhost:8080/api/v1` (production: thay đổi theo deploy).
**Swagger UI:** `GET /swagger/index.html` — same data, interactive.
**Total endpoints:** ~80 across 9 phases. Built on Go 1.25 + Gin + GORM + Postgres.

---

## 0. Conventions

### Authentication

- JWT (HS256). Sau khi `POST /auth/login`, gửi mọi request kèm header:
  ```http
  Authorization: Bearer <access_token>
  ```
- Access token TTL: 60 phút. Refresh token TTL: 14 ngày. Refresh qua
  `POST /auth/refresh`.
- Khi user đổi password / email, tất cả access token cũ bị invalidate.
- **EventSource** (SSE) không gắn header được → dùng query `?token=<jwt>`.

### Response envelope

Mọi response thành công:
```json
{
  "success": true,
  "message": "optional human-readable",
  "data": <T or PaginatedData<T>>
}
```

### Pagination envelope

Mọi list endpoint trả về:
```json
{
  "success": true,
  "data": {
    "items": [ ... ],
    "total": 123,
    "page": 1,
    "page_size": 20,
    "total_pages": 7
  }
}
```

Query params chung: `page` (mặc định 1), `page_size` (mặc định tùy module: 10/20/50, tối đa 100/200).

### Error envelope

```json
{
  "success": false,
  "code": "bad_request | unauthorized | forbidden | not_found | conflict | internal_error",
  "message": "human-readable",
  "details": { ... }     // optional, e.g. {"missing": ["permission:key"]} cho 403
}
```

HTTP status: 400 / 401 / 403 / 404 / 409 / 500 tương ứng.

### Permission system

5 system roles được seed:

| Role | Perm count | Phạm vi |
|---|---|---|
| **Super Admin** | 1 (`*`) | Wildcard — bypass mọi permission gate |
| **Admin** | 40 | Full CRUD trên Users/Employees/Departments/Positions/Skills/Leave/Attendance/Announcements/OrgSettings/Invites |
| **HR Manager** | 32 | Như Admin trừ delete trên Users/Employees + delete trên Departments/Positions/Skills |
| **Manager** | 12 | Read-only trên hầu hết tài nguyên + approve/cancel leave + manage attendance |
| **Employee** | 7 | Self-service only (own leave, own attendance, own profile) |

Permission catalog đầy đủ ở `GET /roles/permissions` (14 groups / 41 perms).

### Date / time

- Datetimes: **ISO-8601** với timezone (`2026-05-22T08:30:00+07:00` hoặc `Z`).
- Dates (DOB, leave from/to): `YYYY-MM-DD` HOẶC full ISO-8601 — repo store UTC.
- Server timezone mặc định `Asia/Ho_Chi_Minh` (env `COMPANY_TIMEZONE`).

### UUID

Mọi `id` là UUID v4 (`gen_random_uuid()`). Format chuẩn:
`a1b2c3d4-e5f6-7890-abcd-ef1234567890`.

---

## 1. Authentication (Phase 1)

### `POST /auth/login` — Đăng nhập

**Auth:** Public.
**Mô tả:** Đăng nhập với email + password, trả về JWT access/refresh + user profile.

**Request body:**
| Field | Type | Required | Validation | Description |
|---|---|---|---|---|
| email | string | ✓ | RFC 5322 email | |
| password | string | ✓ | min 1 char | Plain text — server hash bcrypt |

**Response 200 `data`:**
```json
{
  "access_token": "eyJhbGc…",
  "refresh_token": "eyJhbGc…",
  "token_type": "Bearer",
  "user": {
    "id": "uuid",
    "email": "admin@exnodes.vn",
    "is_active": true,
    "employee": {
      "id": "uuid",
      "full_name": "Super Admin",
      "avatar_url": "https://…",
      "department_id": "uuid",
      "position_id": "uuid",
      "manager_id": null
    },
    "roles": [{"id":"uuid","name":"Super Admin","is_system":true,"permissions":["*"]}]
  }
}
```

**Errors:** `401 unauthorized` (sai email/password / user inactive).

---

### `POST /auth/refresh` — Lấy access token mới

**Auth:** Public (refresh token in body).

**Request body:**
| Field | Type | Required | Description |
|---|---|---|---|
| refresh_token | string | ✓ | Refresh token đã cấp lúc login |

**Response 200 `data`:** Cùng shape với `LoginResponse`.

---

### `POST /auth/logout` — Đăng xuất

**Auth:** JWT.
**Mô tả:** Invalidate access token (stateless — tokens vẫn hết hạn theo TTL; phiên ghi log).

**Request body:** (empty)
**Response 200:** `{"success":true,"message":"Logged out"}`

---

### `GET /roles/permissions` — Liệt kê toàn bộ permission catalog

**Auth:** JWT.
**Mô tả:** Trả về 14 nhóm permission cho FE permission-picker.

**Response 200 `data`:** array of `PermissionGroup`:
```json
[
  {
    "resource": "users",
    "label": "Users",
    "permissions": [
      {"key": "users:read", "label": "View Users", "description": "List and view user profiles"},
      ...
    ]
  },
  ...
]
```

---

## 2. Users (Phase 2)

### `GET /users/me` — Profile của user đang đăng nhập

**Auth:** JWT.

**Response 200 `data`:**
| Field | Type | Description |
|---|---|---|
| id | uuid | |
| email | string | |
| is_active | bool | |
| roles | array of `RoleRead` | |
| notifications_enabled | bool | |
| employee | object/null | `EmployeeSummary` (id, full_name, avatar_url, dept/pos/manager IDs) |
| created_at / updated_at | datetime | |

---

### `POST /users/me/change-password` — Self-service đổi mật khẩu

**Auth:** JWT.

**Request body:**
| Field | Type | Required | Validation |
|---|---|---|---|
| current_password | string | ✓ | min 1 |
| new_password | string | ✓ | min 8 |

**Response 200:** `{"success":true,"message":"Password changed successfully"}`

**Side-effect:** Tất cả access token cũ của user này bị invalidate (phải login lại).

---

### `POST /users/me/change-email` — Self-service đổi email

**Auth:** JWT.

**Request body:**
| Field | Type | Required | Validation |
|---|---|---|---|
| new_email | string | ✓ | email format, không trùng |
| current_password | string | ✓ | min 1 |

**Response 200:** `{"success":true,"data":{"email":"new@example.com"}}`.

**Side-effect:** Tokens cũ invalidate.

---

### `POST /users/me/device-tokens` — Đăng ký device token (mobile push)

**Auth:** JWT.

**Request body:**
| Field | Type | Required | Description |
|---|---|---|---|
| device_id | string | ✓ | Unique device identifier (mobile gen) |
| token | string | ✓ | FCM token |
| platform | string | ✗ | `android` \| `ios` \| `web` \| `unknown` |

**Response 200:** `{"success":true,"message":"Device token registered"}`

---

### `DELETE /users/me/device-tokens/:token` — Hủy đăng ký device

**Auth:** JWT.
**Path param:** `token` (FCM token string).
**Response 200:** `{"success":true}`

---

### `PATCH /users/me/notification-settings` — Bật/tắt notifications

**Auth:** JWT.

**Request body:**
| Field | Type | Required |
|---|---|---|
| notifications_enabled | bool | ✓ |

**Response 200 `data`:** `{"notifications_enabled": true}`

---

### `GET /users` — List users (admin)

**Auth:** JWT + `users:read`.

**Query params:**
| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 10 | max 100 |
| search | string | — | ILIKE match trên email |
| is_active | bool | — | filter |

**Response 200 `data`:** `PaginatedData<UserAdminRead>` (id, email, is_active, roles[], employee summary, timestamps).

---

### `GET /users/:id` — Get user theo id (admin)

**Auth:** JWT + `users:read`.
**Path param:** `id` (uuid).
**Response 200 `data`:** `UserAdminRead`.

---

### `PATCH /users/:id` — Toggle is_active (admin)

**Auth:** JWT + `users:update`.

**Request body:**
| Field | Type | Required |
|---|---|---|
| is_active | bool | ✗ |

**Response 200 `data`:** `UserAdminRead`.

---

### `DELETE /users/:id` — Soft delete user (admin)

**Auth:** JWT + `users:delete`.

**Request body:** (empty hoặc include current_password nếu yêu cầu confirm)
**Response 200:** `{"success":true,"message":"User deactivated"}`

---

### `PATCH /users/:id/change-password` — Admin reset password

**Auth:** JWT + `users:change_password`.

**Request body:**
| Field | Type | Required | Validation |
|---|---|---|---|
| new_password | string | ✓ | min 8 |

**Response 200:** `{"success":true,"message":"Password changed"}`

---

### `PUT /users/:id/roles` — Gán roles cho user

**Auth:** JWT + `users:manage_roles`.

**Request body:**
| Field | Type | Required |
|---|---|---|
| role_ids | array of uuid | ✓ |

PUT-replace semantics — array thay thế toàn bộ role hiện tại.
**Response 200:** `{"success":true}`.

---

## 3. Employees (Phase 2)

### `GET /employees/me` — Profile HR của chính mình

**Auth:** JWT.
**Response 200 `data`:** `EmployeeRead` (full HR profile + dependents embedded). Khi gọi `/me`, các field nhạy cảm (salary, bank) cũng trả nếu user là chính mình. Quản lý (Admin/HR) gọi `/employees/:id` để xem người khác.

---

### `PATCH /employees/me` — Self-service edit profile

**Auth:** JWT.
**Mô tả:** Chỉ một số field nhất định — server enforce whitelist.

**Request body — allowed fields only:**
| Field | Type | Required | Description |
|---|---|---|---|
| phone | string | ✗ | |
| personal_email | string | ✗ | email format |
| permanent_address | string | ✗ | |
| current_address | string | ✗ | |
| marital_status | string | ✗ | `single`/`married`/`divorced`/`widowed` |
| emergency_contact_name | string | ✗ | |
| emergency_contact_relation | string | ✗ | |
| emergency_contact_phone | string | ✗ | |

**KHÔNG chấp nhận:** salary, dept_id, position_id, manager_id, contract, role_ids, is_active — bất kỳ field nào ngoài whitelist sẽ bị bỏ qua silently.

**Response 200 `data`:** updated `EmployeeRead`.

---

### `PATCH /employees/me/avatar` — Upload avatar

**Auth:** JWT.
**Content-Type:** `multipart/form-data`.

**Form fields:**
| Field | Type | Required | Description |
|---|---|---|---|
| avatar | file | ✓ | image/* (jpeg/png/gif/webp), max 5MB |

**Response 200 `data`:** `EmployeeRead` (với `avatar_url` mới).

---

### `GET /employees` — List employees (admin)

**Auth:** JWT + `employees:read`.

**Query params:**
| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 20 | max 100 |
| search | string | — | full_name ILIKE |
| department_id | uuid | — | |
| position_id | uuid | — | |
| manager_id | uuid | — | |
| role_id | uuid | — | |
| is_active | bool | — | |

**Response 200 `data`:** `PaginatedData<EmployeeRead>`.

---

### `POST /employees` — Create employee + user (admin)

**Auth:** JWT + `employees:create`.

**Request body (rút gọn — full schema xem [employee.go](../internal/dto/employee.go)):**

| Field | Type | Required | Description |
|---|---|---|---|
| **email** | string | ✓ | Login email |
| **password** | string | ✓ | min 8 |
| **full_name** | string | ✓ | max 200 |
| is_active | bool | ✗ | default true |
| phone, personal_email, gender, dob, nationality, id_number, … | various | ✗ | HR personal info |
| permanent_address, current_address, education, marital_status | string | ✗ | |
| emergency_contact_name/relation/phone | string | ✗ | |
| department_id, position_id, manager_id | uuid | ✗ | |
| join_date | date | ✗ | |
| contract_type, contract_sign_date, contract_end_date, contract_renewal | various | ✗ | |
| basic_salary, insurance_salary | float | ✗ | ≥0 |
| bank_account, bank_name, bank_holder_name, payment_method | string | ✗ | |
| role_ids | array of uuid | ✗ | mặc định "Employee" role |

**Response 201 `data`:** `EmployeeRead`.
**Errors:** `409 conflict` (email đã tồn tại).

---

### `GET /employees/:id` — Get employee detail (admin)

**Auth:** JWT + `employees:read`.
**Path param:** `id` (employee uuid).
**Response 200 `data`:** `EmployeeRead`.

---

### `PATCH /employees/:id` — Update employee (admin)

**Auth:** JWT + `employees:update`.

**Request body:** Same fields as Create (trừ email/password) — tất cả optional. Đặc biệt:
| Field | Description |
|---|---|
| clear_department | bool, set true để xóa department_id |
| clear_position | bool, set true để xóa position_id |
| clear_manager | bool, set true để xóa manager_id |
| is_active | bool, toggle auth user.is_active |

**Response 200 `data`:** updated `EmployeeRead`.

---

### `DELETE /employees/:id` — Soft delete (admin)

**Auth:** JWT + `employees:delete`.
**Response 200:** `{"success":true,"message":"Employee deactivated"}`

---

### `PATCH /employees/:id/avatar` — Admin upload avatar cho người khác

**Auth:** JWT + `employees:update`.
**Content-Type:** `multipart/form-data` với field `avatar` (file).
**Response 200 `data`:** `EmployeeRead`.

---

### `PATCH /employees/:id/leave-quota` — Set leave quota cho employee

**Auth:** JWT + `leave_quota:manage`.

**Request body:**
| Field | Type | Required | Validation |
|---|---|---|---|
| annual_leave_quota | float | ✓ | ≥0, ≤365 |
| sick_leave_quota | float | ✓ | ≥0, ≤365 |

**Response 200 `data`:** `EmployeeRead` (with updated quota).

---

## 4. Dependents (Phase 2)

Nested under `/employees/:id/dependents`. Owner-or-admin gate.

### `GET /employees/:id/dependents` — List dependents

**Auth:** JWT + (owner OR `dependents:manage`).

**Query params:** `page` (default 1), `page_size` (default 50, max 200).
**Response 200 `data`:** `PaginatedData<DependentRead>`.

---

### `POST /employees/:id/dependents` — Add dependent

**Auth:** JWT + (owner OR `dependents:manage`).

**Request body:**
| Field | Type | Required | Validation |
|---|---|---|---|
| full_name | string | ✓ | min 1, max 200 |
| dob | datetime | ✗ | ISO-8601 |
| gender | string | ✗ | `male`/`female`/`other` |
| relationship | string | ✓ | `child`/`parent`/`spouse`/`sibling`/`other` |

**Response 201 `data`:** `DependentRead` (id, employee_id, full_name, dob, gender, relationship, timestamps).

---

### `PATCH /employees/:id/dependents/:dependentID` — Update dependent

**Auth:** JWT + (owner OR `dependents:manage`).
**Path params:** `id` (employee), `dependentID` (dependent).
**Request body:** Same fields as Create, all optional.
**Response 200 `data`:** `DependentRead`.

---

### `DELETE /employees/:id/dependents/:dependentID` — Soft delete

**Auth:** JWT + (owner OR `dependents:manage`).
**Response 200:** `{"success":true}`.

---

## 5. Departments (Phase 3)

### `GET /departments` — List departments (tree-friendly)

**Auth:** JWT + `departments:read`.

**Query params:**
| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 10 | max 100 |
| search | string | — | name ILIKE |
| parent_id | string | — | uuid HOẶC literal `"root"`/`"null"` → top-level only |

**Response 200 `data`:** `PaginatedData<DepartmentRead>` (id, name, description, parent_id, embedded parent, timestamps).

---

### `POST /departments` — Create

**Auth:** JWT + `departments:create`.

**Request body:**
| Field | Type | Required | Validation |
|---|---|---|---|
| name | string | ✓ | min 1, max 100 |
| description | string | ✗ | max 1000 |
| parent_id | uuid | ✗ | self-ref, null = root |

**Response 201 `data`:** `DepartmentRead`.
**Errors:** `409` nếu name trùng.

---

### `GET /departments/:id` — Get

**Auth:** JWT + `departments:read`.
**Response 200 `data`:** `DepartmentRead`.

---

### `PATCH /departments/:id` — Update

**Auth:** JWT + `departments:update`.

**Request body:**
| Field | Type | Required | Notes |
|---|---|---|---|
| name | string | ✗ | min 1, max 100 |
| description | string | ✗ | max 1000 |
| parent_id | uuid | ✗ | |
| clear_parent | bool | ✗ | set true → biến thành root (phân biệt với "không đổi") |

**Response 200 `data`:** `DepartmentRead`.

---

### `DELETE /departments/:id` — Delete

**Auth:** JWT + `departments:delete`.
**Errors:** `409` nếu còn child departments / positions / employees assigned.

---

## 6. Positions (Phase 3)

### `GET /positions` — List

**Auth:** JWT + `positions:read`.

**Query params:** `page`, `page_size` (max 100), `search`, `department_id` (uuid string).
**Response 200 `data`:** `PaginatedData<PositionRead>`.

---

### `POST /positions` — Create

**Auth:** JWT + `positions:create`.

**Request body:**
| Field | Type | Required | Validation |
|---|---|---|---|
| name | string | ✓ | min 1, max 100 |
| description | string | ✗ | max 1000 |
| department_id | uuid | ✓ | |

**Response 201 `data`:** `PositionRead` (id, name, description, department_id, embedded department, timestamps).

---

### `GET /positions/:id` / `PATCH /positions/:id` / `DELETE /positions/:id`

Cùng shape Departments. PATCH body có `name`, `description`, `department_id` (all optional). DELETE 409 nếu còn employees assigned.

---

## 7. Skills (Phase 4)

### `GET /skills` — List skill catalog

**Auth:** JWT + `skills:read`.
**Query params:** `page`, `page_size` (max 100), `search`. Sort cố định name ASC.
**Response 200 `data`:** `PaginatedData<SkillRead>` (id, name, description, icon_url, timestamps).

---

### `POST /skills` — Create skill

**Auth:** JWT + `skills:create`.
**Content-Type:** `multipart/form-data` (vì có thể upload icon).

**Form fields:**
| Field | Type | Required | Notes |
|---|---|---|---|
| name | string | ✓ | unique (case-insensitive) |
| description | string | ✗ | |
| icon | file | ✗ | image/*, max 2MB |

**Response 201 `data`:** `SkillRead`.

---

### `GET /skills/:id` / `PATCH /skills/:id` / `DELETE /skills/:id`

PATCH cũng multipart (name/description/icon optional). DELETE 409 nếu skill còn assigned cho employees (response body chứa `{skill_id, skill_name, employee_count}`).

---

### `GET /employees/:id/skills` — Skills của employee

**Auth:** JWT + `employees:read`.
**Response 200 `data`:** array of `SkillRead`.

---

### `PUT /employees/:id/skills` — Replace skill set

**Auth:** JWT + `employees:update`.

**Request body:**
| Field | Type | Required | Notes |
|---|---|---|---|
| skill_ids | array of uuid | ✓ | PUT-replace semantics; rỗng = clear all |

**Response 200:** `{"success":true}`.

---

## 8. Announcement Labels (Phase 4)

Get-or-create catalog cho announcement tagging.

### `GET /announcement-labels` — List

**Auth:** JWT + `announcements:manage`.
**Response 200 `data`:** array of `LabelRead` (id, name, timestamps). Sort name ASC, không phân trang.

---

### `POST /announcement-labels` — Get-or-create

**Auth:** JWT + `announcements:manage`.

**Request body:**
| Field | Type | Required | Validation |
|---|---|---|---|
| name | string | ✓ | min 1, max 50; case-insensitive unique |

**Response codes:**
- `201` nếu tạo mới
- `200` nếu trả về label cũ

**Response `data`:** `LabelRead`.

---

## 9. Leave Requests (Phase 5)

### `GET /leave-requests` — List (admin or scoped to self)

**Auth:** JWT + `leave_requests:read`.
**Note:** Manager/Employee thấy own only (service scope).

**Query params:**
| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 10 | max 100 |
| search | string | — | match employee name/email |
| status | string[] | — | repeat param: `?status=pending&status=approved` (values: `pending`/`approved`/`rejected`/`cancelled`) |
| department_id | uuid | — | |
| position_id | uuid | — | |

**Response 200 `data`:** `PaginatedData<LeaveRequestRead>`.

---

### `GET /leave-requests/dashboard/me` — Dashboard summary

**Auth:** JWT.
**Mô tả:** Balance + 5 upcoming + 5 recent history của user đang đăng nhập.

**Response 200 `data`:**
```json
{
  "balance": {
    "year": 2026,
    "annual_quota": 12, "annual_used": 3, "annual_remaining": 9,
    "sick_quota": 6, "sick_used": 0, "sick_remaining": 6,
    "leaves_this_year": 1
  },
  "upcoming": [ ...LeaveRequestRead ],
  "history": [ ...LeaveRequestRead ]
}
```

---

### `GET /leave-requests/history/me` — Paginated history (self)

**Auth:** JWT.

**Query params:**
| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 10 | max 100 |
| status | string[] | — | repeat param |
| start_date | date | — | YYYY-MM-DD |
| end_date | date | — | YYYY-MM-DD |

**Response 200 `data`:** `PaginatedData<LeaveRequestRead>`.

---

### `GET /leave-requests/balance/:employee_id` — Quota của employee

**Auth:** JWT + `leave_requests:read`.

**Query params:** `year` (int, default = year hiện tại).
**Response 200 `data`:** `LeaveBalanceSummary` (xem dashboard).

---

### `POST /leave-requests` — Create

**Auth:** JWT + `leave_requests:create`.
**Content-Type:** `application/json` HOẶC `multipart/form-data` (đính kèm file).

**Body / form field `data` (JSON):**
| Field | Type | Required | Description |
|---|---|---|---|
| employee_id | uuid | ✗ | Admin only — đặt invite cho employee khác (cần `leave_requests:manage`) |
| from_date | datetime | ✓ | ISO-8601 |
| to_date | datetime | ✓ | ISO-8601 ≥ from_date |
| leave_period | string | ✓ | `full_day` \| `morning_half` \| `afternoon_half` |
| leave_type | string | ✓ | `annual` \| `sick` \| `personal` \| `maternity` \| `unpaid` |
| reason | string | ✓ | min 1 |

**Multipart field:**
| Field | Type | Required | Notes |
|---|---|---|---|
| attachment | file | ✗ | image/* + pdf, max 10MB |

**Response 201 `data`:**
```json
{
  "request": { ...LeaveRequestRead },
  "warnings": [
    "Insufficient annual leave balance: requested 3.0 day(s), remaining 0.0 day(s)",
    "Date range overlaps existing leave request(s): …"
  ]
}
```
**Note:** `warnings` không block — request vẫn được tạo, FE chỉ hiển thị warning.

---

### `GET /leave-requests/:id` — Get

**Auth:** JWT + `leave_requests:read` (owner OR admin).
**Response 200 `data`:** `LeaveRequestRead` (id, employee/department/position briefs, from/to_date, total_days, reason, attachment_url, status, created_by, timestamps).

---

### `PATCH /leave-requests/:id` — Update

**Auth:** JWT + `leave_requests:update` (owner-of-pending OR admin).

**Body / form field `data`:**
| Field | Type | Required |
|---|---|---|
| from_date, to_date | datetime | ✗ |
| leave_period, leave_type | string | ✗ |
| reason | string | ✗ |

Multipart `attachment` (file, optional) replaces existing.

**State machine:**
- Rejected/cancelled → 409 (terminal, không edit)
- Approved + admin patch → reverts về pending
- Approved + non-admin → 403
- Pending → cứ patch

**Response 200 `data`:** `LeaveRequestWriteResult` (request + warnings).

---

### `POST /leave-requests/:id/approve` — Approve

**Auth:** JWT + `leave_requests:approve`.
**Response 200 `data`:** `LeaveRequestRead` (status="approved").

---

### `POST /leave-requests/:id/reject` — Reject

**Auth:** JWT + `leave_requests:approve` (same perm).
**Response 200 `data`:** `LeaveRequestRead` (status="rejected").

---

### `POST /leave-requests/:id/cancel` — Cancel

**Auth:** JWT + `leave_requests:cancel` (owner OR admin).
**Pre-conditions:** status pending or approved.
**Response 200 `data`:** `LeaveRequestRead` (status="cancelled").

---

### `POST /leave-requests/:id/delete` — Soft delete

**Auth:** JWT + `leave_requests:delete`.
**Note:** **POST** chứ không phải DELETE (matches Python source).
- Admin: xóa bất kỳ status nào.
- Owner: chỉ xóa được pending của chính mình.

**Response 200:** `{"success":true,"message":"Leave request deleted"}`.

---

## 10. Attendance (Phase 6)

### `POST /attendance/check-in` — Check-in

**Auth:** JWT.

**Request body (optional):**
| Field | Type | Required | Description |
|---|---|---|---|
| check_in | datetime | ✗ | Override; default "now" trong company TZ |
| work_location | string | ✗ | `office`/`remote`/`hybrid`/`field` |
| notes | string | ✗ | |
| latitude, longitude, accuracy | float | ✗ | Chỉ check khi `OFFICE_GPS_ENABLED=true` |

**Response 200 `data`:** `AttendanceRead` (sessions array with check_in/check_out, is_late, is_half_day, hours_worked, employee brief).

**Errors:** `409` nếu đang có open session.

---

### `POST /attendance/check-out` — Check-out

**Auth:** JWT.

**Request body (optional):**
| Field | Type | Required |
|---|---|---|
| check_out | datetime | ✗ |
| notes | string | ✗ |

**Response 200 `data`:** `AttendanceRead`.
**Errors:** `409` nếu không có open session, `400` nếu chưa check-in hôm nay.

---

### `GET /attendance/today` — Status hôm nay

**Auth:** JWT.

**Response 200 `data`:**
```json
{
  "status": "checked_in | checked_out | not_checked_in",
  "is_late": false,
  "sessions": [ ... ],
  "current_check_in": "2026-05-22T08:30:00+07:00",
  "monthly_count": 15,
  "streak": 5
}
```

---

### `GET /attendance/me` — List attendance của chính mình

**Auth:** JWT.

**Query params:** `page` (default 1), `page_size` (default 20), `start_date`, `end_date` (YYYY-MM-DD).
**Response 200 `data`:** `PaginatedData<AttendanceRead>`.

---

### `GET /attendance` — Admin list (hoặc scoped to self)

**Auth:** JWT + `attendance:read`.
**Note:** Non-admin tự động scope to self (service-side).

**Query params:**
| Param | Type | Default | Description |
|---|---|---|---|
| page, page_size | int | 1 / 20 | max 100 |
| employee_id | uuid | — | |
| department_id | uuid | — | |
| start_date, end_date | date | — | YYYY-MM-DD |
| status | string | — | `on_time` \| `late` |

---

### `GET /attendance/matrix` — Monthly matrix view

**Auth:** JWT + `attendance:read`.

**Query params:**
| Param | Type | Default | Description |
|---|---|---|---|
| month | int | now | 1-12 |
| year | int | now | ≥2000 |
| page, page_size | int | 1 / 20 | per-employee pagination |
| search | string | — | filter employee name (admin only) |
| department_id | uuid | — | |
| status | string | — | CSV: `on_time,late,absent,weekend,no_data` |

**Response 200 `data`:**
```json
{
  "year": 2026, "month": 5, "days_in_month": 31,
  "items": [
    {
      "employee_id": "uuid", "employee_name": "Alice",
      "avatar_url": null, "department_name": "Engineering",
      "cells": {
        "1": {"date":"2026-05-01","day":1,"status":"on_time","check_in":"…","check_out":"…","hours_worked":8.0,"is_late":false,"sessions":[...]},
        "2": {"date":"2026-05-02","day":2,"status":"weekend",...},
        ...
      },
      "total_late_minutes": 12,
      "total_early_minutes": 30
    }
  ],
  "total": 8, "page": 1, "page_size": 20, "total_pages": 1
}
```

---

### `GET /attendance/:id` — Get one row

**Auth:** JWT + `attendance:read` (owner OR admin).
**Response 200 `data`:** `AttendanceRead`.

---

### `POST /attendance` — Admin manual create

**Auth:** JWT + `attendance:manage_data`.

**Request body:**
| Field | Type | Required | Description |
|---|---|---|---|
| employee_id | uuid | ✓ | Target employee |
| date | string | ✓ | YYYY-MM-DD |
| check_in | datetime | ✗ | Auto-derives is_late if provided |
| check_out | datetime | ✗ | |
| is_late | bool | ✗ | Override |
| is_half_day | bool | ✗ | |
| work_location | string | ✗ | enum |
| notes | string | ✗ | |

**Response 201 `data`:** `AttendanceRead`.
**Errors:** `409` nếu đã có row (employee, date).

---

### `PATCH /attendance/:id` — Admin update

**Auth:** JWT + `attendance:manage_data`.

**Request body:** Same fields as create, all optional. `check_in`/`check_out` adjust first session.

**Response 200 `data`:** `AttendanceRead`.

---

### `DELETE /attendance/:id` — Admin soft delete

**Auth:** JWT + `attendance:manage_data`.
**Response 200:** `{"success":true,"message":"Deleted"}`.

---

## 11. Announcements (Phase 7)

### `GET /announcements` — List (web)

**Auth:** JWT.
**Note:** Non-admin chỉ thấy rows họ được "see" theo visibility predicate (published + (author OR audience match)). Admin (`announcements:manage`) thấy tất cả.

**Query params:**
| Param | Type | Default | Description |
|---|---|---|---|
| page, page_size | int | 1 / 20 | max 100 |
| search | string | — | title/body ILIKE |
| status | string | — | `draft`/`scheduled`/`published`/`archived` |
| label_id | uuid | — | |
| pinned | bool | — | |
| scope | string | `all` | `all` (visibility-filtered) / `mine` (authored by me, includes drafts) / `targeted-at-me` (audience match only) |
| department_id | uuid | — | Admin filter |

**Response 200 `data`:** `PaginatedData<AnnouncementRead>`.

---

### `GET /announcements/:id` — Get

**Auth:** JWT.
**Response 200 `data`:** `AnnouncementRead`:
```json
{
  "id":"uuid","title":"…","body":"<p>HTML</p>","summary":"…",
  "status":"published","scheduled_at":null,"published_at":"…",
  "target_audience":"all|department",
  "pinned":false,"cover_image_url":"https://…",
  "author":{"id":"uuid","full_name":"…","avatar_url":"…"},
  "labels":[{"id":"uuid","name":"…"}],
  "target_departments":[{"id":"uuid","name":"…"}],
  "attachments":[{"id":"uuid","url":"…","filename":"…","content_type":"…","size_bytes":12345,"created_at":"…"}],
  "has_viewed":true,
  "created_at":"…","updated_at":"…"
}
```
**Errors:** `403` nếu không thấy được, `404` nếu không tồn tại.

---

### `POST /announcements/:id/view` — Mark viewed

**Auth:** JWT.
**Mô tả:** Idempotent — gọi lại không thay đổi viewed_at (giữ thời điểm đầu tiên).
**Response 200:** `{"success":true,"message":"Marked as viewed"}`.

---

### `POST /announcements` — Create

**Auth:** JWT + `announcements:manage`.

**Request body:**
| Field | Type | Required | Notes |
|---|---|---|---|
| title | string | ✓ | min 1 |
| body | string | ✓ | min 1, HTML allowed |
| summary | string | ✗ | |
| status | string | ✗ | default `draft`. Pass `published` để publish ngay (broadcasts SSE) |
| scheduled_at | datetime | ✗ | |
| target_audience | string | ✗ | `all` (default) \| `department` |
| pinned | bool | ✗ | |
| cover_image_url | string | ✗ | |
| label_ids | array of uuid | ✗ | |
| department_ids | array of uuid | ✗ | Required khi target_audience=department |

**Response 201 `data`:** `AnnouncementRead`. Nếu `status=published`, server broadcasts `announcement_published` event qua SSE hub.

---

### `PATCH /announcements/:id` — Update

**Auth:** JWT + `announcements:manage` (owner OR admin).
**Request body:** Tất cả field từ Create đều optional (pointer). `label_ids`/`department_ids`:
- `null` → giữ nguyên
- `[]` → clear all

**Response 200 `data`:** `AnnouncementRead`.

---

### `DELETE /announcements/:id` — Soft delete

**Auth:** JWT + `announcements:manage` (owner OR admin).
**Response 200:** `{"success":true,"message":"Deleted"}`.

---

### `POST /announcements/:id/publish` — Publish

**Auth:** JWT + `announcements:manage`.
**Mô tả:** Set status='published' + stamp published_at + broadcast SSE event. No-op nếu đã published.
**Response 200 `data`:** `AnnouncementRead`.

---

## 12. Mobile Announcements (Phase 7)

Subset for mobile clients — payload nhỏ hơn, luôn visibility-filtered.

### `GET /mobile/announcements` — Mobile list

**Auth:** JWT.
**Query params:** `page` (default 1), `page_size` (default 20, max 100).

**Response 200 `data`:** `PaginatedData<MobileAnnouncementBrief>`:
```json
{
  "id":"uuid","title":"…","summary":"…","cover_image_url":"…",
  "status":"published","pinned":false,"published_at":"…",
  "labels":[{"id":"uuid","name":"…"}],
  "has_viewed":false
}
```
**Body field bị bỏ** — fetch chi tiết qua `/mobile/announcements/:id`.

---

### `GET /mobile/announcements/:id` — Mobile detail

**Auth:** JWT.
**Response 200 `data`:** `AnnouncementRead` (full payload với body + attachments).

---

## 13. Server-Sent Events (Phase 7)

### `GET /sse/announcements` — Realtime announcement stream

**Auth:** JWT (qua header HOẶC `?token=<jwt>`).
**Content-Type:** `text/event-stream`.

**Mô tả:** Long-lived stream. Server gửi:
- 1 frame `event: connected` ngay khi connect (chứa `connection_id`).
- `event: announcement_published` mỗi khi có announcement mới được publish.
- `: keepalive` comment mỗi 30s (FE ignore).

**Event format:**
```
event: announcement_published
data: {"type":"announcement_published","data":{"id":"uuid","title":"…","summary":"…","target_audience":"all","department_ids":[],"pinned":false,"published_at":"…"}}

```

**Note FE/MB:** Khi nhận event, **refetch** list qua `GET /announcements` hoặc `/mobile/announcements` — server đã apply visibility filter trên GET, không phải trên broadcast.

**Errors:** `401` nếu token thiếu / invalid.

---

## 14. Organization Settings (Phase 8)

### `GET /organization-settings/attendance` — Đọc threshold attendance

**Auth:** JWT + `organization_settings:manage`.

**Response 200 `data`:**
```json
{
  "late_threshold_hour": 9, "late_threshold_minute": 0,
  "checkout_threshold_hour": 18, "checkout_threshold_minute": 0
}
```

---

### `PATCH /organization-settings/attendance` — Update threshold

**Auth:** JWT + `organization_settings:manage`.

**Request body (all optional, partial PATCH):**
| Field | Type | Validation |
|---|---|---|
| late_threshold_hour | int | 0-23 |
| late_threshold_minute | int | 0-59 |
| checkout_threshold_hour | int | 0-23 |
| checkout_threshold_minute | int | 0-59 |

**Response 200 `data`:** Same shape as GET (updated).

---

### `GET /organization-settings/company-profile` — Company info

**Auth:** JWT (open read — bất kỳ user nào đã login).

**Response 200 `data`:**
```json
{
  "company_address": "123 HRM Lane, Hanoi",
  "company_latitude": 21.0285,
  "company_longitude": 105.8542,
  "company_address_updated_at": "2026-05-22T08:48:29Z",
  "company_address_updated_by": "uuid",
  "updated_by_name": "Super Admin"
}
```
Các field nullable — empty `{}` khi chưa set.

---

### `PATCH /organization-settings/company-profile` — Update

**Auth:** JWT + `organization_settings:manage`.

**Request body (all optional):**
| Field | Type | Validation |
|---|---|---|
| company_address | string | |
| company_latitude | float | -90 ≤ x ≤ 90 |
| company_longitude | float | -180 ≤ x ≤ 180 |

Bất kỳ field address nào được set → server stamp `company_address_updated_at` + `company_address_updated_by` (= current employee).

**Response 200 `data`:** Same shape as GET (updated).

---

## 15. Invites (Phase 9)

### `GET /invites` — List

**Auth:** JWT + `invites:manage`.

**Query params:**
| Param | Type | Default | Description |
|---|---|---|---|
| page, page_size | int | 1 / 20 | max 100 |
| email | string | — | ILIKE match |
| status | string | — | `pending` \| `accepted` \| `expired` |

**Response 200 `data`:** `PaginatedData<InviteRead>`.

---

### `POST /invites` — Issue invite

**Auth:** JWT + `invites:manage`.

**Request body:**
| Field | Type | Required | Description |
|---|---|---|---|
| email | string | ✓ | RFC 5322 |
| full_name | string | ✗ | Pre-fill, invitee có thể override |
| role_ids | array of uuid | ✗ | Gán roles khi accept |
| department_id | uuid | ✗ | |
| position_id | uuid | ✗ | |

**Response 201 `data`:** `InviteRead`:
```json
{
  "id":"uuid","email":"newbie@example.com","full_name":"New User",
  "role_ids":["uuid"],"department_id":"uuid","position_id":"uuid",
  "expires_at":"2026-05-25T08:00:00Z","accepted_at":null,
  "accepted_user_id":null,"status":"pending",
  "invited_by":"uuid","inviter":{"id":"uuid","full_name":"Super Admin"},
  "last_email_error":null,
  "created_at":"…","updated_at":"…"
}
```

**Behavior:**
- Server generate token, gửi email qua SMTP.
- Nếu SMTP misconfig → `last_email_error` được populate, request vẫn 201 success.
- FE nên hiển thị warning khi `last_email_error != null` và cho phép Resend.

**Errors:**
- `409` — email đã có user.
- `409` — email đã có pending invite (resend qua endpoint khác).

---

### `GET /invites/:id` — Get

**Auth:** JWT + `invites:manage`.
**Response 200 `data`:** `InviteRead`.

---

### `POST /invites/:id/resend` — Resend email

**Auth:** JWT + `invites:manage`.
**Mô tả:** Gửi lại email với CÙNG token (không rotate). Clear `last_email_error` nếu success.
**Response 200 `data`:** `InviteRead`.
**Errors:** `409` nếu invite đã accepted hoặc expired.

---

### `DELETE /invites/:id` — Revoke (soft delete)

**Auth:** JWT + `invites:manage`.
**Mô tả:** Soft-delete invite. Token sau đó sẽ 404 nếu cố accept.
**Response 200:** `{"success":true,"message":"Invite revoked"}`.
**Errors:** `409` nếu đã accepted.

---

### `POST /invites/accept` — **PUBLIC** Accept invite

**Auth:** Public (token in body là credential).
**Mô tả:** Tạo user + employee từ invite, gán roles. Sau đó user có thể login.

**Request body:**
| Field | Type | Required | Validation |
|---|---|---|---|
| token | string | ✓ | Token từ email link (43 chars base64-url) |
| password | string | ✓ | min 8 |
| full_name | string | ✗ | Override invite.full_name |

**Response 200 `data`:**
```json
{
  "user_id": "uuid",
  "email": "newbie@example.com",
  "full_name": "New User",
  "message": "Account created — you can now log in"
}
```

**Errors:**
- `400` — token expired, password quá ngắn, validation fail.
- `404` — token không tồn tại / đã revoke.
- `409` — token đã dùng (`"Invite has already been used"`).

---

## 16. Notifications (Phase 9)

### `POST /notifications/test` — Push notification test (admin debug)

**Auth:** JWT + `users:manage_roles`.
**Mô tả:** Gửi push notification tới các device tokens đã đăng ký của chính caller (self-test). Khi FCM chưa config, tất cả devices count là `skipped`.

**Request body:**
| Field | Type | Required | Description |
|---|---|---|---|
| title | string | ✓ | |
| body | string | ✓ | |
| data | object | ✗ | Custom payload (FCM data fields) |

**Response 200 `data`:**
```json
{
  "sent": 2,
  "skipped": 0,
  "errors": []
}
```
Khi FCM disabled (no `FIREBASE_CREDENTIALS_PATH`): `sent=0`, `skipped=<token_count>`.

---

## Phụ lục A — Permission matrix tóm tắt

| Permission | Super Admin | Admin | HR Manager | Manager | Employee |
|---|:-:|:-:|:-:|:-:|:-:|
| `*` (wildcard) | ✓ | | | | |
| `auth:login` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `users:read` | ✓ | ✓ | ✓ | ✓ | |
| `users:create/update` | ✓ | ✓ | ✓ | | |
| `users:delete` | ✓ | ✓ | | | |
| `users:manage_roles` | ✓ | ✓ | | | |
| `users:change_password` | ✓ | ✓ | ✓ | | |
| `roles:read` | ✓ | ✓ | ✓ | | |
| `roles:create/update` | ✓ | ✓ | | | |
| `employees:read/create/update` | ✓ | ✓ | ✓ | | |
| `employees:delete` | ✓ | ✓ | | | |
| `dependents:manage` | ✓ | ✓ | ✓ | | |
| `departments:*` | ✓ | ✓ | ✓ (Read/Create/Update) | Read | |
| `positions:*` | ✓ | ✓ | ✓ (Read/Create/Update) | Read | |
| `skills:*` | ✓ | ✓ | ✓ (Read/Create/Update) | Read | |
| `leave_requests:read` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `leave_requests:create` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `leave_requests:update` | ✓ | ✓ | ✓ | | ✓ (own) |
| `leave_requests:cancel` | ✓ | ✓ | ✓ | ✓ | ✓ (own) |
| `leave_requests:delete` | ✓ | ✓ | ✓ | | ✓ (own) |
| `leave_requests:approve` | ✓ | ✓ | ✓ | ✓ | |
| `leave_requests:manage` | ✓ | ✓ | ✓ | ✓ | |
| `leave_quota:manage` | ✓ | ✓ | ✓ | | |
| `attendance:read` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `attendance:manage_data` | ✓ | ✓ | ✓ | ✓ | |
| `announcements:manage` | ✓ | ✓ | ✓ | | |
| `organization_settings:manage` | ✓ | ✓ | ✓ | | |
| `invites:manage` | ✓ | ✓ | ✓ | | |

**Lưu ý:** Manager + Employee có `leave_requests:read` nhưng service-side
**scope to self** — Manager chỉ thấy team của mình, Employee chỉ thấy chính mình.
Tương tự `attendance:read` — non-admin chỉ thấy own rows.

---

## Phụ lục B — Common error codes

| HTTP | code | Khi nào |
|---|---|---|
| 400 | `bad_request` | Validation fail, query param sai format, body sai schema |
| 401 | `unauthorized` | Thiếu/sai token, password sai, JWT expired hoặc invalidated |
| 403 | `forbidden` | Thiếu permission (response chứa `details.missing: ["perm:key"]`), hoặc service-level ownership check fail |
| 404 | `not_found` | Resource không tồn tại / đã soft-delete |
| 409 | `conflict` | Duplicate (email/name), state machine violation (cancel đã cancel, accept đã used) |
| 500 | `internal_error` | Bug — log ra server, không expose chi tiết |

---

## Phụ lục C — File upload conventions

- Content-Type: `multipart/form-data`.
- Size limits:
  - Avatar (employee): 5MB, image/* (jpeg/png/gif/webp).
  - Skill icon: 2MB, image/*.
  - Leave attachment: 10MB, image/* + application/pdf.
- Server sniff content type bằng `http.DetectContentType` (không tin client header).
- Trả URL public sau upload (S3-compatible storage, Supabase-backed trong dev).

---

## Phụ lục D — Real-time (SSE)

- 1 connection / 1 tab. EventSource auto-reconnect khi disconnect.
- Server hub in-process — single-replica scaling limit. Production multi-replica cần Redis pub/sub backplane.
- FE pattern:
  ```js
  const es = new EventSource(`${BASE}/sse/announcements?token=${jwt}`);
  es.addEventListener('announcement_published', (ev) => {
    const payload = JSON.parse(ev.data).data;
    // Refetch GET /announcements để pickup row mới (visibility filter ở GET)
  });
  ```

---

## Phụ lục E — Mobile-specific notes

- Mobile dùng JWT giống web (cùng `/auth/login` flow).
- Device token đăng ký qua `POST /users/me/device-tokens` (mobile gen `device_id` unique).
- Push notification: `POST /notifications/test` dùng để debug.
  Production push tự động trigger khi có announcement / leave / attendance event (deferred to roadmap).
- Mobile có 2 route riêng (`/mobile/announcements`, `/mobile/announcements/:id`) trả payload nhỏ hơn.
- SSE hoạt động trên mobile (qua EventSource polyfill hoặc native).

---

## Phụ lục F — Versioning + Roadmap

- API hiện tại: v1 (`/api/v1`).
- Migration version: **12** (xem `make migrate-version`).
- Swagger JSON: `GET /swagger/doc.json` — machine-readable.
- Deferred items (roadmap):
  - Phase 7: announcement attachment-upload endpoint (schema + repo ready).
  - Phase 7: `target_audience='custom'` (per-user targeting).
  - Phase 9: password-reset email flow.
  - Phase 9: FCM topic-based fanout + production push triggers (currently only `/test` debug).

---

**Last updated:** 2026-05-22 — feature-complete sau Phase 9.
**Contact:** danny.tranhoang@exnodes.vn.
