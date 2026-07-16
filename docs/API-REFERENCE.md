# Exnodes HRM API — Reference (FE + Mobile)

**Base URL:** `http://localhost:8080/api/v1` (đổi theo deploy).
**Swagger UI:** `GET /swagger/index.html`.
**Total endpoints:** ~80 across 9 phases. Mọi response liệt kê đầy đủ field — không cần lookup type khác.

---

## 0. Conventions

### Authentication

- JWT (HS256). Sau khi login, gửi header:
  ```http
  Authorization: Bearer <access_token>
  ```
- Access token TTL 60 phút; refresh token 14 ngày.
- Đổi password/email → toàn bộ token cũ bị invalidate.
- **EventSource (SSE)** dùng query `?token=<jwt>` (không gắn header được).

### Wrapper response

Mọi response thành công:
```json
{
  "success": true,
  "message": "optional human-readable",
  "data": <object | array | paginated>
}
```

Pagination wrapper:
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

Error wrapper:
```json
{
  "success": false,
  "code": "bad_request|unauthorized|forbidden|not_found|conflict|internal_error",
  "message": "human-readable",
  "details": { "missing": ["users:read"] }
}
```

### Common query params cho list endpoints

| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | ≥1 |
| page_size | int | (varies per module: 10/20/50) | max 100 hoặc 200 |
| search | string | — | ILIKE substring match (module-specific column) |

### Date / time

- Datetime: ISO-8601 với timezone (`2026-05-22T08:30:00+07:00` hoặc `Z`).
- Date: `YYYY-MM-DD` cho query param; ISO-8601 cho body.
- Server timezone: `Asia/Ho_Chi_Minh`.

### Field annotations

- ✓ = required
- ✗ = optional (có thể null/omit)
- "always present" = field luôn xuất hiện trong response (có thể là `null`)
- "omitempty" = field bị bỏ khỏi JSON khi null/zero

---

## Permission roles tóm tắt

| Role | Perm count | Description |
|---|---|---|
| Super Admin | 1 (`*`) | Wildcard, bypass mọi gate |
| Admin | 40 | Full CRUD trên mọi resource (trừ Delete cho HR-only on Departments/Positions/Skills) |
| HR Manager | 32 | Như Admin trừ Delete trên Users/Employees/Departments/Positions/Skills |
| Manager | 12 | Read-only + manage leave/attendance |
| Employee | 7 | Self-service only |

Permission matrix đầy đủ ở [Phụ lục A](#phụ-lục-a--permission-matrix-đầy-đủ).

---

## 1. Authentication

### `POST /auth/login`

**Auth:** Public.
**Mô tả:** Đăng nhập, trả token + user profile + embedded employee summary + roles.

**Request body:**

| Field | Type | Required | Validation |
|---|---|---|---|
| email | string | ✓ | RFC 5322 email |
| password | string | ✓ | min 1 char |

**Response 200 `data`:**

| Field | Type | Always present | Description |
|---|---|---|---|
| access_token | string | ✓ | JWT |
| refresh_token | string | ✓ | JWT |
| token_type | string | ✓ | luôn `"Bearer"` |
| user.id | uuid | ✓ | |
| user.email | string | ✓ | |
| user.is_active | bool | ✓ | |
| user.employee | object \| null | ✓ | nullable nếu user chưa có HR profile |
| user.employee.id | uuid | ✓ | |
| user.employee.full_name | string | ✓ | |
| user.employee.avatar_url | string \| null | omitempty | |
| user.employee.department_id | uuid \| null | omitempty | |
| user.employee.position_id | uuid \| null | omitempty | |
| user.employee.manager_id | uuid \| null | omitempty | |
| user.roles | array | ✓ | có thể rỗng |
| user.roles[].id | uuid | ✓ | |
| user.roles[].name | string | ✓ | |
| user.roles[].is_system | bool | ✓ | |
| user.roles[].permissions | array of string | ✓ | e.g. `["users:read", "*"]` |

**Errors:** `401 unauthorized` (sai email/password / user inactive).

---

### `POST /auth/refresh`

**Auth:** Public.

**Request body:**

| Field | Type | Required |
|---|---|---|
| refresh_token | string | ✓ |

**Response 200 `data`:** Cùng shape như `POST /auth/login`.

---

### `POST /auth/logout`

**Auth:** JWT.
**Request body:** none.
**Response 200:** `{"success": true, "message": "Logged out"}`.

---

### `GET /roles/permissions`

**Auth:** JWT.
**Mô tả:** Trả toàn bộ permission catalog cho FE permission-picker.

**Response 200 `data`:** array of:

| Field | Type | Description |
|---|---|---|
| resource | string | e.g. `"users"`, `"employees"` |
| label | string | Display label (e.g. `"Users"`) |
| permissions | array | |
| permissions[].key | string | Permission constant (e.g. `"users:read"`) |
| permissions[].label | string | Display label |
| permissions[].description | string | Tooltip/description |

14 groups, 41 perms total.

---

## 2. Users

### `GET /users/me`

**Auth:** JWT.

**Response 200 `data`:**

| Field | Type | Always present | Description |
|---|---|---|---|
| id | uuid | ✓ | |
| email | string | ✓ | |
| is_active | bool | ✓ | |
| notifications_enabled | bool | ✓ | |
| roles | array | ✓ | |
| roles[].id | uuid | ✓ | |
| roles[].name | string | ✓ | |
| roles[].description | string | omitempty | |
| roles[].is_system | bool | ✓ | |
| roles[].permissions | array of string | ✓ | |
| employee | object \| null | omitempty | EmployeeSummary (nếu user có HR profile) |
| employee.id | uuid | ✓ | |
| employee.full_name | string | ✓ | |
| employee.avatar_url | string \| null | omitempty | |
| employee.department_id | uuid \| null | omitempty | |
| employee.position_id | uuid \| null | omitempty | |
| employee.manager_id | uuid \| null | omitempty | |
| created_at | datetime | ✓ | |
| updated_at | datetime | ✓ | |

---

### `POST /users/me/change-password`

**Auth:** JWT.
**Side-effect:** Invalidate toàn bộ token cũ.

**Request body:**

| Field | Type | Required | Validation |
|---|---|---|---|
| current_password | string | ✓ | min 1 |
| new_password | string | ✓ | min 8 |

**Response 200:** `{"success": true, "message": "Password changed successfully"}`.

---

### `POST /users/me/change-email`

**Auth:** JWT.
**Side-effect:** Invalidate toàn bộ token cũ.

**Request body:**

| Field | Type | Required | Validation |
|---|---|---|---|
| new_email | string | ✓ | email format, không trùng user khác |
| current_password | string | ✓ | min 1 |

**Response 200 `data`:**

| Field | Type |
|---|---|
| email | string (new email) |

---

### `POST /users/me/device-tokens`

**Auth:** JWT.

**Request body:**

| Field | Type | Required | Validation |
|---|---|---|---|
| device_id | string | ✓ | Mobile gen, unique per device |
| token | string | ✓ | FCM token |
| platform | string | ✗ | `android` \| `ios` \| `web` \| `unknown` |

**Response 200:** `{"success": true, "message": "Device token registered"}`.

---

### `DELETE /users/me/device-tokens/:token`

**Auth:** JWT.
**Path param:** `token` (FCM token string, URL-encoded nếu chứa ký tự đặc biệt).
**Response 200:** `{"success": true}`.

---

### `PATCH /users/me/notification-settings`

**Auth:** JWT.

**Request body:**

| Field | Type | Required |
|---|---|---|
| notifications_enabled | bool | ✓ |

**Response 200 `data`:**

| Field | Type |
|---|---|
| notifications_enabled | bool |

---

### `GET /users` — Admin list

**Auth:** JWT + `users:read`.

**Query params:**

| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 10 | max 100 |
| search | string | — | ILIKE on email |
| is_active | bool | — | filter |

**Response 200 `data` (PaginatedData):**

| Field | Type | Description |
|---|---|---|
| items | array | xem dưới |
| total | int64 | |
| page | int | |
| page_size | int | |
| total_pages | int | |

**`items[]` shape:**

| Field | Type | Always present | Description |
|---|---|---|---|
| id | uuid | ✓ | |
| email | string | ✓ | |
| is_active | bool | ✓ | |
| roles | array | ✓ | same shape as `/me` |
| employee | object \| null | omitempty | EmployeeSummary (id, full_name, avatar_url, dept/pos/manager IDs) |
| created_at | datetime | ✓ | |
| updated_at | datetime | ✓ | |

---

### `GET /users/:id`

**Auth:** JWT + `users:read`.
**Path param:** `id` (uuid).
**Response 200 `data`:** Cùng shape `items[]` ở trên.

---

### `PATCH /users/:id`

**Auth:** JWT + `users:update`.

**Request body:**

| Field | Type | Required |
|---|---|---|
| is_active | bool | ✗ |

**Response 200 `data`:** UserAdminRead (giống `items[]`).

---

### `DELETE /users/:id`

**Auth:** JWT + `users:delete`.
**Request body:** none.
**Response 200:** `{"success": true, "message": "User deactivated"}`.

---

### `PATCH /users/:id/change-password`

**Auth:** JWT + `users:change_password`.

**Request body:**

| Field | Type | Required | Validation |
|---|---|---|---|
| new_password | string | ✓ | min 8 |

**Response 200:** `{"success": true, "message": "Password changed"}`.

---

### `PUT /users/:id/roles`

**Auth:** JWT + `users:manage_roles`.

**Request body:**

| Field | Type | Required |
|---|---|---|
| role_ids | array of uuid | ✓ |

PUT-replace: array thay thế toàn bộ role hiện tại. Empty array = clear all.

**Response 200:** `{"success": true}`.

---

## 3. Employees

### `GET /employees/me`

**Auth:** JWT.

**Response 200 `data`:**

| Field | Type | Always present | Description |
|---|---|---|---|
| id | uuid | ✓ | Employee ID |
| user_id | uuid | ✓ | |
| email | string | ✓ | |
| full_name | string | ✓ | |
| phone | string \| null | omitempty | |
| personal_email | string \| null | omitempty | |
| gender | string \| null | omitempty | `male`\|`female`\|`other` |
| dob | datetime \| null | omitempty | |
| nationality | string \| null | omitempty | |
| id_number | string \| null | omitempty | |
| id_issue_date | datetime \| null | omitempty | |
| id_front_image | string \| null | omitempty | URL |
| id_back_image | string \| null | omitempty | URL |
| permanent_address | string \| null | omitempty | |
| current_address | string \| null | omitempty | |
| education | string \| null | omitempty | |
| marital_status | string \| null | omitempty | `single`\|`married`\|`divorced`\|`widowed` |
| emergency_contact_name | string \| null | omitempty | |
| emergency_contact_relation | string \| null | omitempty | |
| emergency_contact_phone | string \| null | omitempty | |
| avatar_url | string \| null | omitempty | |
| department | object \| null | omitempty | `{id: uuid, name: string}` |
| position | object \| null | omitempty | `{id: uuid, name: string}` |
| manager | object \| null | omitempty | `{id: uuid, name: string}` |
| join_date | datetime \| null | omitempty | |
| contract_type | string \| null | omitempty | |
| contract_sign_date | datetime \| null | omitempty | |
| contract_end_date | datetime \| null | omitempty | |
| contract_renewal | bool \| null | omitempty | |
| basic_salary | float \| null | omitempty | |
| insurance_salary | float \| null | omitempty | |
| bank_account | string \| null | omitempty | |
| bank_name | string \| null | omitempty | |
| bank_holder_name | string \| null | omitempty | |
| payment_method | string \| null | omitempty | |
| is_active | bool | ✓ | |
| roles | array | ✓ | (xem shape ở `/users/me`) |
| dependents | array | omitempty | Embedded — xem section Dependents |
| created_at | datetime | ✓ | |
| updated_at | datetime | ✓ | |

---

### `PATCH /employees/me`

**Auth:** JWT.
**Mô tả:** Self-update — server enforce whitelist (silently bỏ qua field ngoài whitelist).

**Request body (whitelist — chỉ các field này được apply):**

| Field | Type | Validation |
|---|---|---|
| phone | string | |
| personal_email | string | email format |
| permanent_address | string | |
| current_address | string | |
| marital_status | string | `single`\|`married`\|`divorced`\|`widowed` |
| emergency_contact_name | string | |
| emergency_contact_relation | string | |
| emergency_contact_phone | string | |

**Response 200 `data`:** Same shape as `GET /employees/me`.

---

### `PATCH /employees/me/avatar`

**Auth:** JWT.
**Content-Type:** `multipart/form-data`.

**Form fields:**

| Field | Type | Required | Constraints |
|---|---|---|---|
| avatar | file | ✓ | image/* (jpeg/png/gif/webp), max 5MB |

**Response 200 `data`:** EmployeeRead (giống `/employees/me`).

---

### `GET /employees` — Admin list

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

**Response 200 `data` (PaginatedData):**

| Field | Type |
|---|---|
| items | array of EmployeeRead (xem shape ở `/employees/me` — đầy đủ field) |
| total | int64 |
| page | int |
| page_size | int |
| total_pages | int |

---

### `POST /employees` — Admin create

**Auth:** JWT + `employees:create`.

**Request body:**

| Field | Type | Required | Validation |
|---|---|---|---|
| email | string | ✓ | email format, không trùng |
| password | string | ✓ | min 8 |
| is_active | bool | ✗ | default true |
| full_name | string | ✓ | min 1, max 200 |
| phone | string | ✗ | |
| personal_email | string | ✗ | email format |
| gender | string | ✗ | `male`\|`female`\|`other` |
| dob | datetime | ✗ | |
| nationality | string | ✗ | |
| id_number | string | ✗ | |
| id_issue_date | datetime | ✗ | |
| id_front_image | string | ✗ | URL |
| id_back_image | string | ✗ | URL |
| permanent_address | string | ✗ | |
| current_address | string | ✗ | |
| education | string | ✗ | |
| marital_status | string | ✗ | `single`\|`married`\|`divorced`\|`widowed` |
| emergency_contact_name | string | ✗ | |
| emergency_contact_relation | string | ✗ | |
| emergency_contact_phone | string | ✗ | |
| department_id | uuid | ✗ | |
| position_id | uuid | ✗ | |
| manager_id | uuid | ✗ | |
| join_date | datetime | ✗ | |
| contract_type | string | ✗ | |
| contract_sign_date | datetime | ✗ | |
| contract_end_date | datetime | ✗ | |
| contract_renewal | bool | ✗ | |
| basic_salary | float | ✗ | ≥0 |
| insurance_salary | float | ✗ | ≥0 |
| bank_account | string | ✗ | |
| bank_name | string | ✗ | |
| bank_holder_name | string | ✗ | |
| payment_method | string | ✗ | |
| role_ids | array of uuid | ✗ | mặc định gán "Employee" role |

**Response 201 `data`:** EmployeeRead (shape giống `/employees/me`).
**Errors:** `409 conflict` nếu email đã tồn tại.

---

### `GET /employees/:id`

**Auth:** JWT + `employees:read`.
**Response 200 `data`:** EmployeeRead.

---

### `PATCH /employees/:id`

**Auth:** JWT + `employees:update`.

**Request body:** Same fields as POST (trừ email/password), tất cả optional. Plus:

| Field | Type | Description |
|---|---|---|
| clear_department | bool | set `true` → xóa department_id |
| clear_position | bool | set `true` → xóa position_id |
| clear_manager | bool | set `true` → xóa manager_id |
| is_active | bool | toggle auth user.is_active |

**Response 200 `data`:** EmployeeRead.

---

### `DELETE /employees/:id`

**Auth:** JWT + `employees:delete`.
**Response 200:** `{"success": true, "message": "Employee deactivated"}`.

---

### `PATCH /employees/:id/avatar`

**Auth:** JWT + `employees:update`.
**Content-Type:** `multipart/form-data` với `avatar` (file, image/*, max 5MB).
**Response 200 `data`:** EmployeeRead.

---

### `PATCH /employees/:id/leave-quota`

**Auth:** JWT + `leave_quota:manage`.

**Request body:**

| Field | Type | Required | Validation |
|---|---|---|---|
| annual_leave_quota | float | ✓ | ≥0, ≤365 |
| sick_leave_quota | float | ✓ | ≥0, ≤365 |

**Response 200 `data`:** EmployeeRead.

---

## 4. Dependents

Nested under `/employees/:id/dependents`. Owner-OR-`dependents:manage` gate.

### `GET /employees/:id/dependents`

**Auth:** JWT + (owner OR `dependents:manage`).

**Query params:**

| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 50 | max 200 |

**Response 200 `data` (PaginatedData):**

`items[]` shape:

| Field | Type | Always present | Description |
|---|---|---|---|
| id | uuid | ✓ | |
| employee_id | uuid | ✓ | |
| full_name | string | ✓ | |
| dob | datetime \| null | omitempty | |
| gender | string \| null | omitempty | `male`\|`female`\|`other` |
| relationship | string | ✓ | `child`\|`parent`\|`spouse`\|`sibling`\|`other` |
| created_at | datetime | ✓ | |
| updated_at | datetime | ✓ | |

---

### `POST /employees/:id/dependents`

**Auth:** JWT + (owner OR `dependents:manage`).

**Request body:**

| Field | Type | Required | Validation |
|---|---|---|---|
| full_name | string | ✓ | min 1, max 200 |
| dob | datetime | ✗ | |
| gender | string | ✗ | `male`\|`female`\|`other` |
| relationship | string | ✓ | `child`\|`parent`\|`spouse`\|`sibling`\|`other` |

**Response 201 `data`:** Dependent (shape như `items[]` ở GET).

---

### `PATCH /employees/:id/dependents/:dependentID`

**Auth:** JWT + (owner OR `dependents:manage`).

**Request body** (tất cả optional):

| Field | Type | Validation |
|---|---|---|
| full_name | string | |
| dob | datetime | |
| gender | string | enum |
| relationship | string | enum |

**Response 200 `data`:** Dependent.

---

### `DELETE /employees/:id/dependents/:dependentID`

**Auth:** JWT + (owner OR `dependents:manage`).
**Response 200:** `{"success": true}`.

---

## 5. Departments

### `GET /departments`

**Auth:** JWT + `departments:read`.

**Query params:**

| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 10 | max 100 |
| search | string | — | name ILIKE |
| parent_id | string | — | uuid HOẶC literal `"root"`/`"null"` → top-level only |

**Response 200 `data` (PaginatedData):**

`items[]` shape:

| Field | Type | Always present | Description |
|---|---|---|---|
| id | uuid | ✓ | |
| name | string | ✓ | |
| description | string | ✓ | empty string nếu không set |
| parent_id | uuid \| null | omitempty | |
| parent | object \| null | omitempty | Recursive Department shape (id, name, description, parent_id, ...) |
| created_at | datetime | ✓ | |
| updated_at | datetime | ✓ | |

---

### `POST /departments`

**Auth:** JWT + `departments:create`.

**Request body:**

| Field | Type | Required | Validation |
|---|---|---|---|
| name | string | ✓ | min 1, max 100 |
| description | string | ✗ | max 1000 |
| parent_id | uuid | ✗ | null = root |

**Response 201 `data`:** Department (shape như list `items[]`).
**Errors:** `409` nếu name trùng.

---

### `GET /departments/:id`

**Auth:** JWT + `departments:read`.
**Response 200 `data`:** Department.

---

### `PATCH /departments/:id`

**Auth:** JWT + `departments:update`.

**Request body** (tất cả optional):

| Field | Type | Validation | Notes |
|---|---|---|---|
| name | string | min 1, max 100 | |
| description | string | max 1000 | |
| parent_id | uuid | | |
| clear_parent | bool | | set `true` → biến thành root |

**Response 200 `data`:** Department.

---

### `DELETE /departments/:id`

**Auth:** JWT + `departments:delete`.
**Response 200:** `{"success": true}`.
**Errors:** `409` nếu còn child departments / positions / employees assigned.

---

## 6. Positions

### `GET /positions`

**Auth:** JWT + `positions:read`.

**Query params:**

| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 10 | max 100 |
| search | string | — | name ILIKE |
| department_id | string (uuid) | — | |

**Response 200 `data` (PaginatedData):**

`items[]` shape:

| Field | Type | Always present | Description |
|---|---|---|---|
| id | uuid | ✓ | |
| name | string | ✓ | |
| description | string | ✓ | empty string nếu không set |
| department_id | uuid | ✓ | |
| department | object \| null | omitempty | Department shape (id, name, description, parent_id, parent, timestamps) |
| created_at | datetime | ✓ | |
| updated_at | datetime | ✓ | |

---

### `POST /positions`

**Auth:** JWT + `positions:create`.

**Request body:**

| Field | Type | Required | Validation |
|---|---|---|---|
| name | string | ✓ | min 1, max 100 |
| description | string | ✗ | max 1000 |
| department_id | uuid | ✓ | |

**Response 201 `data`:** Position.

---

### `GET /positions/:id`

**Auth:** JWT + `positions:read`.
**Response 200 `data`:** Position.

---

### `PATCH /positions/:id`

**Auth:** JWT + `positions:update`.

**Request body** (tất cả optional):

| Field | Type | Validation |
|---|---|---|
| name | string | min 1, max 100 |
| description | string | max 1000 |
| department_id | uuid | |

**Response 200 `data`:** Position.

---

### `DELETE /positions/:id`

**Auth:** JWT + `positions:delete`.
**Response 200:** `{"success": true}`.
**Errors:** `409` nếu còn employees assigned.

---

## 7. Skills

### `GET /skills`

**Auth:** JWT + `skills:read`.

**Query params:**

| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 20 | max 100 |
| search | string | — | name ILIKE |

Sort cố định: `name ASC`.

**Response 200 `data` (PaginatedData):**

`items[]` shape:

| Field | Type | Always present | Description |
|---|---|---|---|
| id | uuid | ✓ | |
| name | string | ✓ | |
| description | string | ✓ | empty string nếu không set |
| icon_url | string \| null | omitempty | |
| created_at | datetime | ✓ | |
| updated_at | datetime | ✓ | |

---

### `POST /skills`

**Auth:** JWT + `skills:create`.
**Content-Type:** `multipart/form-data` (có thể đính kèm icon).

**Form fields:**

| Field | Type | Required | Validation |
|---|---|---|---|
| name | string | ✓ | unique case-insensitive |
| description | string | ✗ | |
| icon | file | ✗ | image/*, max 2MB |

**Response 201 `data`:** Skill.
**Errors:** `409` nếu name đã tồn tại.

---

### `GET /skills/:id` / `PATCH /skills/:id`

PATCH cũng multipart, fields giống POST nhưng all optional.
**Response 200 `data`:** Skill.

---

### `DELETE /skills/:id`

**Auth:** JWT + `skills:delete`.
**Response 200:** `{"success": true}`.
**Errors:** `409` nếu còn employees assigned. Response body:

| Field | Type |
|---|---|
| skill_id | uuid |
| skill_name | string |
| employee_count | int64 |

---

### `GET /employees/:id/skills`

**Auth:** JWT + `employees:read`.
**Response 200 `data`:** array of Skill (shape giống `/skills items[]`).

---

### `PUT /employees/:id/skills`

**Auth:** JWT + `employees:update`.

**Request body:**

| Field | Type | Required |
|---|---|---|
| skill_ids | array of uuid | ✓ |

PUT-replace: empty array = clear all assignments.

**Response 200:** `{"success": true}`.

---

## 8. Announcement Labels

### `GET /announcement-labels`

**Auth:** JWT + `announcements:manage`.
**Mô tả:** Trả toàn bộ labels (không pagination), sort name ASC.

**Response 200 `data`:** array of:

| Field | Type | Always present | Description |
|---|---|---|---|
| id | uuid | ✓ | |
| name | string | ✓ | |
| created_at | datetime | ✓ | |
| updated_at | datetime | ✓ | |

---

### `POST /announcement-labels`

**Auth:** JWT + `announcements:manage`.
**Mô tả:** Get-or-create. Case-insensitive unique trên `name`.

**Request body:**

| Field | Type | Required | Validation |
|---|---|---|---|
| name | string | ✓ | min 1, max 50 |

**Response codes:**
- `201 Created` nếu tạo mới
- `200 OK` nếu trả về label đã tồn tại

**Response `data`:** Label (giống `items[]` ở GET).

---

## 9. Leave Requests

### `GET /leave-requests` — List

**Auth:** JWT + `leave_requests:read`.
**Scope:** Non-admin (Manager/Employee) tự động filter to self (service-side).

**Query params:**

| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 10 | max 100 |
| search | string | — | employee name/email |
| status | array of string | — | Repeat param: `?status=pending&status=approved`. Values: `pending`\|`approved`\|`rejected`\|`cancelled` |
| department_id | uuid | — | |
| position_id | uuid | — | |

**Response 200 `data` (PaginatedData):**

`items[]` shape — **LeaveRequest**:

| Field | Type | Always present | Description |
|---|---|---|---|
| id | uuid string | ✓ | |
| employee | object \| null | omitempty | `{id: uuid string, name: string}` |
| department | object \| null | omitempty | `{id: uuid string, name: string}` |
| position | object \| null | omitempty | `{id: uuid string, name: string}` |
| from_date | datetime | ✓ | |
| to_date | datetime | ✓ | |
| leave_period | string | ✓ | `full_day`\|`morning_half`\|`afternoon_half` |
| leave_type | string | ✓ | `annual`\|`sick`\|`personal`\|`maternity`\|`unpaid` |
| total_days | float | ✓ | e.g. 3.0, 0.5 (half day) |
| reason | string | ✓ | |
| attachment_url | string \| null | omitempty | |
| status | string | ✓ | `pending`\|`approved`\|`rejected`\|`cancelled` |
| created_by | uuid string | ✓ | Employee ID của người tạo (= subject when self-created; admin khi proxy-create) |
| created_at | datetime | ✓ | |
| updated_at | datetime | ✓ | |

---

### `GET /leave-requests/dashboard/me`

**Auth:** JWT.
**Mô tả:** Balance summary + 5 upcoming + 5 recent history.

**Response 200 `data`:**

| Field | Type | Description |
|---|---|---|
| balance.year | int | |
| balance.annual_quota | float | |
| balance.annual_used | float | |
| balance.annual_remaining | float | |
| balance.sick_quota | float | |
| balance.sick_used | float | |
| balance.sick_remaining | float | |
| balance.leaves_this_year | int | |
| upcoming | array of LeaveRequest | (shape giống `items[]` ở list) |
| history | array of LeaveRequest | |

---

### `GET /leave-requests/history/me` — Paginated history

**Auth:** JWT.

**Query params:**

| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 10 | max 100 |
| status | array of string | — | Repeat param |
| start_date | date | — | `YYYY-MM-DD` |
| end_date | date | — | `YYYY-MM-DD` |

**Response 200 `data` (PaginatedData):** items = LeaveRequest.

---

### `GET /leave-requests/balance/:employee_id`

**Auth:** JWT + `leave_requests:read`.

**Query params:** `year` (int, default = năm hiện tại).

**Response 200 `data`:** Cùng shape `balance` ở `/dashboard/me`.

---

### `POST /leave-requests` — Create

**Auth:** JWT + `leave_requests:create`.
**Content-Type:** `application/json` HOẶC `multipart/form-data` (đính kèm file).

**Body JSON / multipart field `data` (JSON-encoded):**

| Field | Type | Required | Description |
|---|---|---|---|
| employee_id | uuid | ✗ | Admin proxy-create (cần `leave_requests:manage`); nếu omit = self |
| from_date | datetime | ✓ | ISO-8601 |
| to_date | datetime | ✓ | ≥ from_date |
| leave_period | string | ✓ | `full_day`\|`morning_half`\|`afternoon_half` |
| leave_type | string | ✓ | `annual`\|`sick`\|`personal`\|`maternity`\|`unpaid` |
| reason | string | ✓ | min 1 |

**Multipart form fields:**

| Field | Type | Required | Validation |
|---|---|---|---|
| data | string | ✓ (multipart) | JSON-encoded request body |
| attachment | file | ✗ | image/* + application/pdf, max 10MB |

**Response 201 `data`:**

| Field | Type | Description |
|---|---|---|
| request | object | LeaveRequest (shape giống `items[]` ở list) |
| warnings | array of string | Non-blocking warnings, e.g. `["Insufficient annual leave balance: requested 3.0 day(s), remaining 0.0 day(s)", "Date range overlaps existing leave request(s): 2026-05-10..2026-05-12 (approved)"]`. Empty array = không có warning. **Request VẪN được tạo dù có warning.** |

---

### `GET /leave-requests/:id`

**Auth:** JWT + `leave_requests:read` (owner OR admin).
**Response 200 `data`:** LeaveRequest.

---

### `PATCH /leave-requests/:id`

**Auth:** JWT + `leave_requests:update` (owner-of-pending OR admin).
**Content-Type:** JSON HOẶC multipart.

**Request body (tất cả optional):**

| Field | Type | Validation |
|---|---|---|
| from_date | datetime | |
| to_date | datetime | |
| leave_period | string | enum |
| leave_type | string | enum |
| reason | string | |

**Multipart `attachment`** (file, optional) → replace existing.

**State machine:**
- `rejected`/`cancelled` → 409 (terminal)
- `approved` + admin patch → revert về `pending`
- `approved` + non-admin → 403
- `pending` → patch bình thường

**Response 200 `data`:** Same as Create response — `{request, warnings}`.

---

### `POST /leave-requests/:id/approve`

**Auth:** JWT + `leave_requests:approve`.
**Response 200 `data`:** LeaveRequest (status="approved").

---

### `POST /leave-requests/:id/reject`

**Auth:** JWT + `leave_requests:approve`.
**Response 200 `data`:** LeaveRequest (status="rejected").

---

### `POST /leave-requests/:id/cancel`

**Auth:** JWT + `leave_requests:cancel` (owner OR admin).
**Response 200 `data`:** LeaveRequest (status="cancelled").

---

### `POST /leave-requests/:id/delete`

**Auth:** JWT + `leave_requests:delete`.
**Lưu ý:** **POST** chứ không phải DELETE (matches Python source).
- Admin: xóa bất kỳ status nào.
- Non-admin owner: chỉ pending của chính mình.

**Response 200:** `{"success": true, "message": "Leave request deleted"}`.

---

## 10. Attendance

### `POST /attendance/check-in`

**Auth:** JWT.

**Request body (tất cả optional):**

| Field | Type | Required | Description |
|---|---|---|---|
| check_in | datetime | ✗ | Override; default = now (company TZ) |
| work_location | string | ✗ | `office`\|`remote`\|`hybrid`\|`field` |
| notes | string | ✗ | |
| latitude | float | ✗ | Chỉ check khi `OFFICE_GPS_ENABLED=true` |
| longitude | float | ✗ | |
| accuracy | float | ✗ | |

**Response 200 `data`** — **Attendance**:

| Field | Type | Always present | Description |
|---|---|---|---|
| id | uuid | ✓ | |
| employee_id | uuid | ✓ | |
| employee | object \| null | omitempty | (xem dưới) |
| date | string | ✓ | `YYYY-MM-DD` (company TZ) |
| is_late | bool | ✓ | Computed từ FIRST check-in vs threshold; KHÔNG re-evaluate |
| is_half_day | bool | ✓ | Set khi total hours-worked < threshold tại check-out |
| work_location | string \| null | omitempty | |
| notes | string \| null | omitempty | |
| sessions | array | ✓ | Mỗi session là 1 cặp check-in/check-out |
| sessions[].id | uuid | ✓ | |
| sessions[].check_in | datetime | ✓ | UTC |
| sessions[].check_out | datetime \| null | omitempty | UTC |
| sessions[].is_auto_checkout | bool | ✓ | |
| sessions[].hours_worked | float \| null | omitempty | null khi session đang mở |
| check_in | datetime \| null | omitempty | First session's check_in (convenience) |
| check_out | datetime \| null | omitempty | Last session's check_out |
| hours_worked | float \| null | omitempty | Total across sessions |
| created_at | datetime | ✓ | |
| updated_at | datetime | ✓ | |

**`employee` embedded shape:**

| Field | Type |
|---|---|
| id | uuid |
| full_name | string |
| avatar_url | string \| null (omitempty) |
| department | `{id: uuid, name: string}` \| null (omitempty) |
| position | `{id: uuid, name: string}` \| null (omitempty) |

**Errors:** `409` khi đã có open session.

---

### `POST /attendance/check-out`

**Auth:** JWT.

**Request body (tất cả optional):**

| Field | Type | Description |
|---|---|---|
| check_out | datetime | Default = now |
| notes | string | |

**Response 200 `data`:** Attendance (shape ở trên).
**Errors:** `409` nếu không có open session, `400` nếu chưa check-in hôm nay.

---

### `GET /attendance/today`

**Auth:** JWT.

**Response 200 `data`:**

| Field | Type | Always present | Description |
|---|---|---|---|
| status | string | ✓ | `not_checked_in`\|`checked_in`\|`checked_out` |
| is_late | bool | ✓ | |
| sessions | array | ✓ | Same session shape ở Attendance |
| current_check_in | datetime \| null | omitempty | Set khi status=checked_in |
| monthly_count | int | ✓ | Số ngày có check-in trong tháng |
| streak | int | ✓ | Số ngày làm việc liên tiếp (Mon-Fri) có check-in |

---

### `GET /attendance/me`

**Auth:** JWT.

**Query params:**

| Param | Type | Default |
|---|---|---|
| page | int | 1 |
| page_size | int | 20 |
| start_date | date | — |
| end_date | date | — |

**Response 200 `data` (PaginatedData):** items = Attendance.

---

### `GET /attendance` — Admin list

**Auth:** JWT + `attendance:read`.
**Scope:** Non-admin tự động scope to self (service-side).

**Query params:**

| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 20 | max 100 |
| employee_id | uuid | — | |
| department_id | uuid | — | |
| start_date | date | — | `YYYY-MM-DD` |
| end_date | date | — | `YYYY-MM-DD` |
| status | string | — | `on_time`\|`late` |

**Response 200 `data` (PaginatedData):** items = Attendance.

---

### `GET /attendance/matrix` — Monthly grid view

**Auth:** JWT + `attendance:read`.

**Query params:**

| Param | Type | Default | Description |
|---|---|---|---|
| month | int | now | 1-12 |
| year | int | now | ≥2000 |
| page | int | 1 | |
| page_size | int | 20 | per-employee pagination |
| search | string | — | filter employee name (admin only) |
| department_id | uuid | — | |
| status | string | — | CSV: `on_time,late,absent,weekend,no_data` |

**Response 200 `data`:**

| Field | Type | Always present | Description |
|---|---|---|---|
| year | int | ✓ | |
| month | int | ✓ | 1-12 |
| days_in_month | int | ✓ | 28/29/30/31 |
| items | array of EmployeeRow | ✓ | |
| total | int | ✓ | Total employees after filter |
| page | int | ✓ | |
| page_size | int | ✓ | |
| total_pages | int | ✓ | |

**EmployeeRow shape:**

| Field | Type | Always present | Description |
|---|---|---|---|
| employee_id | uuid | ✓ | |
| employee_name | string | ✓ | |
| avatar_url | string \| null | omitempty | |
| department_name | string \| null | omitempty | |
| cells | map[int → Cell] | ✓ | Keyed by day-of-month (1..days_in_month) |
| total_late_minutes | int | ✓ | |
| total_early_minutes | int | ✓ | |

**Cell shape:**

| Field | Type | Always present | Description |
|---|---|---|---|
| date | string | ✓ | `YYYY-MM-DD` |
| day | int | ✓ | 1..days_in_month |
| status | string | ✓ | `on_time`\|`late`\|`absent`\|`weekend`\|`no_data` |
| check_in | datetime \| null | omitempty | First session of the day |
| check_out | datetime \| null | omitempty | Last session of the day |
| hours_worked | float \| null | omitempty | Total |
| is_late | bool | ✓ | |
| sessions | array | omitempty | Optional detail |

---

### `GET /attendance/:id`

**Auth:** JWT + `attendance:read` (owner OR admin).
**Response 200 `data`:** Attendance.

---

### `POST /attendance` — Admin manual create

**Auth:** JWT + `attendance:manage_data`.

**Request body:**

| Field | Type | Required | Description |
|---|---|---|---|
| employee_id | uuid | ✓ | |
| date | string | ✓ | `YYYY-MM-DD` |
| check_in | datetime | ✗ | Auto-derives is_late if provided + is_late not set |
| check_out | datetime | ✗ | |
| is_late | bool | ✗ | Override |
| is_half_day | bool | ✗ | |
| work_location | string | ✗ | enum |
| notes | string | ✗ | |

**Response 201 `data`:** Attendance.
**Errors:** `409` nếu đã có row cho (employee_id, date).

---

### `PATCH /attendance/:id`

**Auth:** JWT + `attendance:manage_data`.

**Request body (tất cả optional):**

| Field | Type | Description |
|---|---|---|
| is_late | bool | |
| is_half_day | bool | |
| work_location | string | enum |
| notes | string | |
| check_in | datetime | Adjust first session |
| check_out | datetime | Adjust first session |

**Response 200 `data`:** Attendance.

---

### `DELETE /attendance/:id`

**Auth:** JWT + `attendance:manage_data`.
**Response 200:** `{"success": true, "message": "Deleted"}`.

---

## 11. Announcements (web)

### `GET /announcements` — List

**Auth:** JWT.
**Scope:** Non-admin chỉ thấy rows họ "see" theo visibility predicate. Admin (`announcements:manage`) thấy tất.

**Query params:**

| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 20 | max 100 |
| search | string | — | title/body ILIKE |
| status | string | — | `draft`\|`scheduled`\|`published`\|`archived` |
| label_id | uuid | — | |
| pinned | bool | — | |
| scope | string | `all` | `all` (visibility-filtered) \| `mine` (authored by me, includes drafts) \| `targeted-at-me` (audience match only) |
| department_id | uuid | — | Admin filter |

**Response 200 `data` (PaginatedData):**

`items[]` shape — **Announcement**:

| Field | Type | Always present | Description |
|---|---|---|---|
| id | uuid | ✓ | |
| title | string | ✓ | |
| body | string | ✓ | HTML allowed |
| summary | string \| null | omitempty | |
| status | string | ✓ | `draft`\|`scheduled`\|`published`\|`archived` |
| scheduled_at | datetime \| null | omitempty | |
| published_at | datetime \| null | omitempty | |
| target_audience | string | ✓ | `all`\|`department` |
| pinned | bool | ✓ | |
| cover_image_url | string \| null | omitempty | |
| author | object \| null | omitempty | (xem dưới) |
| labels | array | ✓ | Empty array nếu không có |
| labels[].id | uuid | ✓ | |
| labels[].name | string | ✓ | |
| target_departments | array | ✓ | Empty nếu target_audience=all |
| target_departments[].id | uuid | ✓ | |
| target_departments[].name | string | ✓ | |
| attachments | array | ✓ | Empty nếu không có |
| attachments[].id | uuid | ✓ | |
| attachments[].url | string | ✓ | |
| attachments[].filename | string | ✓ | |
| attachments[].content_type | string | ✓ | |
| attachments[].size_bytes | int64 | ✓ | |
| attachments[].created_at | datetime | ✓ | |
| has_viewed | bool | ✓ | True nếu caller đã mark viewed |
| created_at | datetime | ✓ | |
| updated_at | datetime | ✓ | |

**`author` embedded shape:**

| Field | Type | Always present |
|---|---|---|
| id | uuid | ✓ |
| full_name | string | ✓ |
| avatar_url | string \| null | omitempty |

---

### `GET /announcements/:id`

**Auth:** JWT.
**Response 200 `data`:** Announcement (full shape ở trên).
**Errors:** `403` nếu không thấy được, `404` nếu không tồn tại.

---

### `POST /announcements/:id/view`

**Auth:** JWT.
**Mô tả:** Idempotent — gọi lại không thay đổi viewed_at.
**Request body:** none.
**Response 200:** `{"success": true, "message": "Marked as viewed"}`.

---

### `POST /announcements`

**Auth:** JWT + `announcements:manage`.

**Request body:**

| Field | Type | Required | Validation/Notes |
|---|---|---|---|
| title | string | ✓ | min 1 |
| body | string | ✓ | min 1; HTML allowed |
| summary | string | ✗ | |
| status | string | ✗ | default `draft`. Set `published` để publish ngay (broadcast SSE) |
| scheduled_at | datetime | ✗ | |
| target_audience | string | ✗ | default `all`. `department` cần department_ids |
| pinned | bool | ✗ | |
| cover_image_url | string | ✗ | |
| label_ids | array of uuid | ✗ | |
| department_ids | array of uuid | ✗ | Required khi target_audience=department |

**Response 201 `data`:** Announcement.

---

### `PATCH /announcements/:id`

**Auth:** JWT + `announcements:manage` (owner OR admin).

**Request body (tất cả optional):**

| Field | Type | Notes |
|---|---|---|
| title | string | min 1 |
| body | string | min 1 |
| summary | string | |
| status | string | enum |
| scheduled_at | datetime | |
| target_audience | string | enum |
| pinned | bool | |
| cover_image_url | string | |
| label_ids | array of uuid \| null | null = leave unchanged, `[]` = clear all |
| department_ids | array of uuid \| null | null = leave unchanged, `[]` = clear all |

**Response 200 `data`:** Announcement.

---

### `DELETE /announcements/:id`

**Auth:** JWT + `announcements:manage` (owner OR admin).
**Response 200:** `{"success": true, "message": "Deleted"}`.

---

### `POST /announcements/:id/publish`

**Auth:** JWT + `announcements:manage`.
**Mô tả:** Set status='published', stamp published_at, broadcast SSE. No-op nếu đã published.
**Response 200 `data`:** Announcement.

---

## 12. Mobile Announcements

### `GET /mobile/announcements`

**Auth:** JWT.

**Query params:** `page` (default 1), `page_size` (default 20, max 100).

**Response 200 `data` (PaginatedData):**

`items[]` shape — **MobileAnnouncementBrief**:

| Field | Type | Always present | Description |
|---|---|---|---|
| id | uuid | ✓ | |
| title | string | ✓ | |
| summary | string \| null | omitempty | |
| cover_image_url | string \| null | omitempty | |
| status | string | ✓ | Always `published` (filtered) |
| pinned | bool | ✓ | |
| published_at | datetime \| null | omitempty | |
| labels | array | ✓ | `[{id, name}]` |
| has_viewed | bool | ✓ | |

**Note:** Body bị bỏ — fetch chi tiết qua `/mobile/announcements/:id`.

---

### `GET /mobile/announcements/:id`

**Auth:** JWT.
**Response 200 `data`:** Announcement (full shape giống web — có body + attachments).

---

## 13. Server-Sent Events

### `GET /sse/announcements`

**Auth:** JWT — qua header HOẶC `?token=<jwt>`.
**Content-Type:** `text/event-stream`.

**Mô tả:** Long-lived stream. Server gửi:

1. **`event: connected`** — ngay khi connect.
   ```
   event: connected
   data: {"connection_id":"<uuid>"}
   ```

2. **`event: announcement_published`** — mỗi khi có announcement publish.
   ```
   event: announcement_published
   data: {"type":"announcement_published","data":{...}}
   ```
   `data` shape:

   | Field | Type | Always present | Description |
   |---|---|---|---|
   | id | uuid | ✓ | |
   | title | string | ✓ | |
   | summary | string \| null | omitempty | |
   | target_audience | string | ✓ | |
   | department_ids | array of uuid | omitempty | |
   | pinned | bool | ✓ | |
   | published_at | datetime | ✓ | |

3. **`: keepalive`** comment mỗi 30s — FE ignore.

**Note FE/MB:** Khi nhận `announcement_published`, **refetch** list qua `GET /announcements` hoặc `/mobile/announcements` (server đã apply visibility filter ở GET).

**Errors:** `401` nếu token thiếu/invalid.

---

## 14. Organization Settings

### `GET /organization-settings/attendance`

**Auth:** JWT + `organization_settings:manage`.

**Response 200 `data`:**

| Field | Type | Always present |
|---|---|---|
| late_threshold_hour | int | ✓ |
| late_threshold_minute | int | ✓ |
| checkout_threshold_hour | int | ✓ |
| checkout_threshold_minute | int | ✓ |

---

### `PATCH /organization-settings/attendance`

**Auth:** JWT + `organization_settings:manage`.

**Request body (tất cả optional, partial PATCH):**

| Field | Type | Validation |
|---|---|---|
| late_threshold_hour | int | 0-23 |
| late_threshold_minute | int | 0-59 |
| checkout_threshold_hour | int | 0-23 |
| checkout_threshold_minute | int | 0-59 |

**Response 200 `data`:** Same as GET.

---

### `GET /organization-settings/company-profile`

**Auth:** JWT (open read).

**Response 200 `data`:**

| Field | Type | Always present | Description |
|---|---|---|---|
| company_address | string \| null | omitempty | |
| company_latitude | float \| null | omitempty | |
| company_longitude | float \| null | omitempty | |
| company_address_updated_at | datetime \| null | omitempty | |
| company_address_updated_by | uuid \| null | omitempty | Employee ID |
| updated_by_name | string \| null | omitempty | Resolved từ Employee.FullName |

Empty `{}` nếu chưa set.

---

### `PATCH /organization-settings/company-profile`

**Auth:** JWT + `organization_settings:manage`.

**Request body (tất cả optional):**

| Field | Type | Validation |
|---|---|---|
| company_address | string | |
| company_latitude | float | -90 ≤ x ≤ 90 |
| company_longitude | float | -180 ≤ x ≤ 180 |

Khi có ANY address field thay đổi → server stamp `company_address_updated_at` + `company_address_updated_by` = current employee.

**Response 200 `data`:** Same shape as GET.

---

## 15. Invites

### `GET /invites`

**Auth:** JWT + `invites:manage`.

**Query params:**

| Param | Type | Default | Description |
|---|---|---|---|
| page | int | 1 | |
| page_size | int | 20 | max 100 |
| email | string | — | ILIKE match |
| status | string | — | `pending`\|`accepted`\|`expired` |

**Response 200 `data` (PaginatedData):**

`items[]` shape — **Invite**:

| Field | Type | Always present | Description |
|---|---|---|---|
| id | uuid | ✓ | |
| email | string | ✓ | |
| full_name | string \| null | omitempty | |
| role_ids | array of uuid | ✓ | Empty array nếu không set |
| department_id | uuid \| null | omitempty | |
| position_id | uuid \| null | omitempty | |
| expires_at | datetime | ✓ | |
| accepted_at | datetime \| null | omitempty | |
| accepted_user_id | uuid \| null | omitempty | Populated khi accepted |
| status | string | ✓ | Derived: `pending`\|`accepted`\|`expired`\|`revoked` |
| invited_by | uuid | ✓ | Employee ID |
| inviter | object \| null | omitempty | `{id: uuid, full_name: string}` |
| last_email_error | string \| null | omitempty | Populated nếu SMTP send fail |
| created_at | datetime | ✓ | |
| updated_at | datetime | ✓ | |

---

### `POST /invites` — Issue invite

**Auth:** JWT + `invites:manage`.

**Request body:**

| Field | Type | Required | Validation |
|---|---|---|---|
| email | string | ✓ | RFC 5322, không trùng user/pending-invite |
| full_name | string | ✗ | Pre-fill, invitee có thể override khi accept |
| role_ids | array of uuid | ✗ | Gán roles khi accept; mặc định "Employee" |
| department_id | uuid | ✗ | |
| position_id | uuid | ✗ | |

**Response 201 `data`:** Invite (shape giống list `items[]`).

**Behavior:**
- Server generate 32-byte token (base64-url, 43 chars), gửi email qua SMTP.
- SMTP fail → `last_email_error` được populate; request vẫn `201`.
- FE nên warning khi `last_email_error != null` + cho phép Resend.

**Errors:**
- `409` — email đã có user.
- `409` — email đã có pending invite.

---

### `GET /invites/:id`

**Auth:** JWT + `invites:manage`.
**Response 200 `data`:** Invite.

---

### `POST /invites/:id/resend`

**Auth:** JWT + `invites:manage`.
**Mô tả:** Gửi lại email với CÙNG token (không rotate). Clear `last_email_error` nếu success.
**Request body:** none.
**Response 200 `data`:** Invite.
**Errors:** `409` nếu invite đã accepted hoặc expired.

---

### `DELETE /invites/:id`

**Auth:** JWT + `invites:manage`.
**Mô tả:** Soft-delete (revoke). Token sau đó 404 nếu cố accept.
**Response 200:** `{"success": true, "message": "Invite revoked"}`.
**Errors:** `409` nếu đã accepted.

---

### `POST /invites/accept` — **PUBLIC**

**Auth:** Public (token in body là credential).
**Mô tả:** Tạo user + employee từ invite, gán roles. User có thể login ngay sau đó.

**Request body:**

| Field | Type | Required | Validation |
|---|---|---|---|
| token | string | ✓ | Từ email link |
| password | string | ✓ | min 8 |
| full_name | string | ✗ | Override invite.full_name |

**Response 200 `data`:**

| Field | Type | Always present |
|---|---|---|
| user_id | uuid | ✓ |
| email | string | ✓ |
| full_name | string | ✓ |
| message | string | ✓ |

**Errors:**
- `400` — token expired, password sai validation, missing field.
- `404` — token không tồn tại / đã revoke.
- `409` — token đã dùng (`"Invite has already been used"`).

---

## 16. Notifications

### `POST /notifications/test`

**Auth:** JWT + `users:manage_roles`.
**Mô tả:** Push test tới các device tokens của caller (self-test). Khi FCM disabled, tất cả tokens count là `skipped`.

**Request body:**

| Field | Type | Required |
|---|---|---|
| title | string | ✓ |
| body | string | ✓ |
| data | object (key→any) | ✗ |

**Response 200 `data`:**

| Field | Type | Always present | Description |
|---|---|---|---|
| sent | int | ✓ | Số tokens push thành công |
| skipped | int | ✓ | Số tokens fail / skip (FCM disabled = total) |
| errors | array of string | ✓ | Per-token error messages (empty array nếu không có) |

---

## Phụ lục A — Permission matrix đầy đủ

| Permission key | Super Admin | Admin | HR Manager | Manager | Employee |
|---|:-:|:-:|:-:|:-:|:-:|
| `*` (wildcard) | ✓ | | | | |
| `auth:login` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `users:read` | ✓ | ✓ | ✓ | ✓ | |
| `users:create` | ✓ | ✓ | ✓ | | |
| `users:update` | ✓ | ✓ | ✓ | | |
| `users:delete` | ✓ | ✓ | | | |
| `users:manage_roles` | ✓ | ✓ | | | |
| `users:change_password` | ✓ | ✓ | ✓ | | |
| `roles:read` | ✓ | ✓ | ✓ | | |
| `roles:create` | ✓ | ✓ | | | |
| `roles:update` | ✓ | ✓ | | | |
| `roles:delete` | ✓ | | | | |
| `employees:read` | ✓ | ✓ | ✓ | | |
| `employees:create` | ✓ | ✓ | ✓ | | |
| `employees:update` | ✓ | ✓ | ✓ | | |
| `employees:delete` | ✓ | ✓ | | | |
| `dependents:manage` | ✓ | ✓ | ✓ | | |
| `departments:read` | ✓ | ✓ | ✓ | ✓ | |
| `departments:create` | ✓ | ✓ | ✓ | | |
| `departments:update` | ✓ | ✓ | ✓ | | |
| `departments:delete` | ✓ | ✓ | | | |
| `positions:read` | ✓ | ✓ | ✓ | ✓ | |
| `positions:create` | ✓ | ✓ | ✓ | | |
| `positions:update` | ✓ | ✓ | ✓ | | |
| `positions:delete` | ✓ | ✓ | | | |
| `skills:read` | ✓ | ✓ | ✓ | ✓ | |
| `skills:create` | ✓ | ✓ | ✓ | | |
| `skills:update` | ✓ | ✓ | ✓ | | |
| `skills:delete` | ✓ | ✓ | | | |
| `leave_requests:read` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `leave_requests:create` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `leave_requests:update` | ✓ | ✓ | ✓ | | ✓ (own only) |
| `leave_requests:delete` | ✓ | ✓ | ✓ | | ✓ (own pending) |
| `leave_requests:approve` | ✓ | ✓ | ✓ | ✓ | |
| `leave_requests:cancel` | ✓ | ✓ | ✓ | ✓ | ✓ (own) |
| `leave_requests:manage` | ✓ | ✓ | ✓ | ✓ | |
| `leave_quota:manage` | ✓ | ✓ | ✓ | | |
| `attendance:read` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `attendance:manage_data` | ✓ | ✓ | ✓ | ✓ | |
| `organization_settings:manage` | ✓ | ✓ | ✓ | | |
| `announcements:manage` | ✓ | ✓ | ✓ | | |
| `invites:manage` | ✓ | ✓ | ✓ | | |

**Service-level scoping:** Manager + Employee có `leave_requests:read` + `attendance:read` nhưng service tự động scope to self/team — không thấy toàn bộ.

---

## Phụ lục B — Error codes

| HTTP | code | Khi nào |
|---|---|---|
| 400 | `bad_request` | Validation fail (binding tag), query param sai format, body sai schema |
| 401 | `unauthorized` | Thiếu/sai token, password sai, JWT expired hoặc invalidated by password/email change |
| 403 | `forbidden` | Thiếu permission (response chứa `details.missing: ["perm:key"]`) hoặc service-level ownership check fail |
| 404 | `not_found` | Resource không tồn tại / đã soft-delete |
| 409 | `conflict` | Duplicate (email/name), state machine violation (cancel đã cancel, accept token đã used) |
| 500 | `internal_error` | Bug — log server-side, không expose chi tiết |

---

## Phụ lục C — File upload conventions

- Content-Type: `multipart/form-data`.
- Server sniff content type bằng `http.DetectContentType` (không tin client header).

| Upload | Field | Max size | MIME allowlist |
|---|---|---|---|
| Avatar (employee) | `avatar` | 5MB | image/jpeg, image/png, image/webp |
| Skill icon | `icon` | 2MB | image/* |
| Leave attachment | `attachment` | 10MB | image/* + application/pdf |

Server trả URL public sau upload từ bucket AWS S3 đã cấu hình.

---

## Phụ lục D — Real-time (SSE) FE pattern

```js
const es = new EventSource(`${BASE}/sse/announcements?token=${jwt}`);

es.addEventListener('connected', (ev) => {
  const { connection_id } = JSON.parse(ev.data);
  console.log('SSE connected:', connection_id);
});

es.addEventListener('announcement_published', (ev) => {
  const envelope = JSON.parse(ev.data);
  // envelope.type === "announcement_published"
  // envelope.data === { id, title, summary, target_audience, department_ids, pinned, published_at }

  // Refetch full list để pickup row mới (visibility filter ở GET)
  refetchAnnouncements();
});

es.onerror = (e) => {
  // EventSource tự reconnect; chỉ cần log
};
```

---

## Phụ lục E — Mobile-specific

- Mobile dùng JWT giống web (cùng `/auth/login` flow).
- Device token đăng ký qua `POST /users/me/device-tokens` (mobile gen `device_id` unique).
- Push debug: `POST /notifications/test`.
- Mobile có 2 route riêng (`/mobile/announcements`, `/mobile/announcements/:id`) trả payload nhỏ hơn (no body trong list).
- SSE hoạt động trên mobile qua EventSource polyfill hoặc native (e.g. react-native-event-source).

---

## Phụ lục F — Versioning + Deferred

- API hiện tại: v1 (`/api/v1`).
- Migration version: **12** (`make migrate-version`).
- Swagger JSON: `GET /swagger/doc.json`.

**Roadmap (chưa ship):**
- Phase 7: announcement attachment-upload endpoint (schema + repo ready).
- Phase 7: `target_audience='custom'` (per-user targeting).
- Phase 9: password-reset email flow.
- Phase 9: production push triggers (announcement / leave / attendance events).
- Phase 6: attendance reads `system_config` thay vì env vars.

---

**Last updated:** 2026-05-22 — feature-complete sau Phase 9 + permission seed fix.
**Contact:** danny.tranhoang@exnodes.vn.
