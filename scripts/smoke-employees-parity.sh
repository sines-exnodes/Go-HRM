#!/usr/bin/env bash
#
# Smoke test for the employees Python-shape parity work (PR A + PR B).
# Boots the server against a test DB and exercises every audited behavior with
# real HTTP requests, asserting the expected status / shape. Mirrors the
# verification logs (docs/superpowers/verification/employees-parity-pr-{a,b}.md).
#
# Requires: a reachable Postgres test DB, `jq`, `go` (run from the repo root).
# The test DB is migrated automatically by the server on boot ONLY if already
# at the latest version; otherwise run `make migrate-up` against it first, or
# let `go test ./...` (TestMain) provision it once.
#
# Usage:   ./scripts/smoke-employees-parity.sh
# Env:     TEST_DB_URL  PORT  ADMIN_EMAIL  ADMIN_PASSWORD  SERVER_BIN
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"; cd "$ROOT"

TEST_DB_URL="${TEST_DB_URL:-postgres://postgres:devpassword@127.0.0.1:5432/exnodes_hrm_test?sslmode=disable}"
PORT="${PORT:-8082}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@local.dev}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-ChangeMe!2026}"
SERVER_BIN="${SERVER_BIN:-/tmp/hrm-smoke-server}"
BASE="http://127.0.0.1:${PORT}/api/v1"

PASS=0; FAIL=0
pass() { echo "  ✓ $1"; PASS=$((PASS+1)); }
fail() { echo "  ✗ $1   ${2:-}"; FAIL=$((FAIL+1)); }
# eq EXPECTED ACTUAL LABEL
eq()   { [ "$1" = "$2" ] && pass "$3" || fail "$3" "(expected '$1', got '$2')"; }

# req METHOD URL [json] [token] -> sets CODE and RESP
req() {
  local m="$1" url="$2" data="${3:-}" tok="${4:-$TOK}" out
  local args=(-s -w $'\n%{http_code}' -X "$m" -H "Authorization: Bearer $tok")
  [ -n "$data" ] && args+=(-H 'Content-Type: application/json' -d "$data")
  out=$(curl "${args[@]}" "$url"); CODE="${out##*$'\n'}"; RESP="${out%$'\n'*}"
}
jqv() { echo "$RESP" | jq -r "$1"; }

echo "==> build"
go build -o "$SERVER_BIN" ./cmd/server || { echo "build failed"; exit 1; }

echo "==> boot on :$PORT (test DB)"
DATABASE_URL="$TEST_DB_URL" DB_HOST=127.0.0.1 DB_PORT=5432 DB_USER=postgres \
  DB_PASSWORD=devpassword DB_NAME=exnodes_hrm_test DB_SSLMODE=disable \
  JWT_SECRET_KEY="smoke-secret" PORT="$PORT" APP_ENV=development GIN_MODE=release \
  CORS_ALLOWED_ORIGINS="" SUPER_ADMIN_EMAIL="$ADMIN_EMAIL" \
  SUPER_ADMIN_PASSWORD="$ADMIN_PASSWORD" SUPER_ADMIN_NAME="Smoke Admin" \
  "$SERVER_BIN" >/tmp/hrm-smoke-run.log 2>&1 &
SERVER_PID=$!
trap 'kill "$SERVER_PID" 2>/dev/null' EXIT

for _ in $(seq 1 40); do curl -sf "http://127.0.0.1:${PORT}/health" >/dev/null 2>&1 && break; sleep 0.5; done
curl -sf "http://127.0.0.1:${PORT}/health" >/dev/null || { echo "server did not start — see /tmp/hrm-smoke-run.log"; tail -20 /tmp/hrm-smoke-run.log; exit 1; }

SFX="$RANDOM$RANDOM"
TOK=$(curl -s -X POST "$BASE/auth/login" -H 'Content-Type: application/json' \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\"}" | jq -r '.data.access_token')
[ -n "$TOK" ] && [ "$TOK" != "null" ] && pass "admin login" || { fail "admin login"; exit 1; }

echo "== #4/#8 create employee with emergency-contact list + experience_year + cv_url =="
req POST "$BASE/employees" "{\"email\":\"smk-$SFX@x.com\",\"password\":\"Pass12345\",\"first_name\":\"Smoke\",\"last_name\":\"$SFX\",\"experience_year\":2018,\"cv_url\":\"https://x/cv.pdf\",\"basic_salary\":5000,\"bank_account\":\"190233445566\",\"bank_name\":\"ACB\",\"emergency_contacts\":[{\"full_name\":\"Mom\",\"relationship\":\"parent\",\"phone_number\":\"0900\"},{\"full_name\":\"Dad\"}]}"
EMP=$(jqv '.data.id'); EMP_UID=$(jqv '.data.user_id')
eq 201 "$CODE" "#4 create (201)"
eq 2 "$(jqv '.data.emergency_contacts | length')" "#4 echo has 2 emergency contacts"
eq 2018 "$(jqv '.data.experience_year')" "#8 experience_year echoed as a year"
eq "https://x/cv.pdf" "$(jqv '.data.cv_url')" "#8 cv_url echoed"
eq "190233445566" "$(jqv '.data.bank_account')" "#6 write echo bank_account UNMASKED"

echo "== #5 leave quota present on read (defaults 12/6) =="
req GET "$BASE/employees/$EMP"
eq 12 "$(jqv '.data.annual_leave_quota')" "#5 annual_leave_quota=12 default"
eq 6  "$(jqv '.data.sick_leave_quota')"   "#5 sick_leave_quota=6 default"
eq "•••• 5566" "$(jqv '.data.bank_account')" "#6 read MASKS bank_account"

echo "== #4 PATCH emergency contacts replace -> clear =="
req PATCH "$BASE/employees/$EMP" '{"emergency_contacts":[{"full_name":"Spouse","relationship":"spouse"}]}'
eq 1 "$(jqv '.data.emergency_contacts | length')" "#4 replace -> 1 contact"
req PATCH "$BASE/employees/$EMP" '{"emergency_contacts":[]}'
eq 0 "$(jqv '.data.emergency_contacts | length')" "#4 clear -> 0 contacts"

echo "== #17 marital_status enum tightened =="
req PATCH "$BASE/employees/$EMP" '{"marital_status":"divorced"}'
eq 400 "$CODE" "#17 marital_status=divorced rejected (400)"

echo "== #13 admin change-email (NEW endpoint) =="
req POST "$BASE/users/$EMP_UID/change-email" "{\"new_email\":\"smk-renamed-$SFX@x.com\"}"
eq 200 "$CODE" "#13 admin change-email (200)"

echo "== #12 self-guards (on the admin's own records) =="
req GET "$BASE/employees/me"; ADMIN_EMP=$(jqv '.data.id'); ADMIN_UID=$(jqv '.data.user_id')
eq 0 "$(jqv '.data.emergency_contacts | length')" "#7 /employees/me returns the new shape"
req DELETE "$BASE/employees/$ADMIN_EMP"
eq 400 "$CODE" "#12 cannot delete own employee (400)"
req PATCH "$BASE/users/$ADMIN_UID" '{"is_active":false}'
eq 400 "$CODE" "#12 cannot deactivate own user (400)"

echo "== #7 self-edit widened: name/gender/dob via /employees/me =="
req PATCH "$BASE/employees/me" '{"first_name":"Renamed","last_name":"Admin","gender":"female"}'
eq "Renamed" "$(jqv '.data.first_name')" "#7 self can edit first_name"
eq "female" "$(jqv '.data.gender')" "#7 self can edit gender"

echo "== #6 limited-role read-strip + write-gate (needs a no-salary role) =="
RGO="$(mktemp -d)/role.go"
cat > "$RGO" <<'GOEOF'
package main
import ("database/sql";"fmt";"os";_ "github.com/lib/pq")
func main(){db,e:=sql.Open("postgres",os.Args[1]);if e!=nil{os.Exit(1)};defer db.Close()
 var id string
 if e:=db.QueryRow(`INSERT INTO roles (name,description,is_system,permissions) VALUES ('SmokeFieldViewer','smoke',false,'["auth:login","employees:read","employees:create"]'::jsonb) ON CONFLICT (name) DO UPDATE SET permissions=EXCLUDED.permissions RETURNING id`).Scan(&id);e!=nil{os.Exit(1)}
 fmt.Print(id)}
GOEOF
RID=$(go run "$RGO" "$TEST_DB_URL" 2>/dev/null || true)
if [ -n "$RID" ]; then
  req POST "$BASE/employees" "{\"email\":\"smk-view-$SFX@x.com\",\"password\":\"Pass12345\",\"first_name\":\"Viewer\",\"last_name\":\"$SFX\"}"
  VUID=$(jqv '.data.user_id')
  req PUT "$BASE/users/$VUID/roles" "{\"role_ids\":[\"$RID\"]}"
  VTOK=$(curl -s -X POST "$BASE/auth/login" -H 'Content-Type: application/json' -d "{\"email\":\"smk-view-$SFX@x.com\",\"password\":\"Pass12345\"}" | jq -r '.data.access_token')
  req GET "$BASE/employees/$EMP" "" "$VTOK"
  eq "null" "$(jqv '.data.basic_salary')" "#6 viewer: salary STRIPPED"
  eq "null" "$(jqv '.data.bank_account')" "#6 viewer: banking STRIPPED"
  [ "$(jqv '.data.first_name')" != "null" ] && pass "#6 viewer: non-salary fields still visible" || fail "#6 viewer non-salary fields"
  req POST "$BASE/employees" "{\"email\":\"smk-z-$SFX@x.com\",\"password\":\"Pass12345\",\"first_name\":\"Z\",\"last_name\":\"Test\",\"basic_salary\":1}" "$VTOK"
  eq 403 "$CODE" "#6 write-gate: salary without salary_manage (403)"
  [ "$(jqv '.message')" = "You do not have permission to set salary fields" ] && pass "#6 write-gate message is the field guard (not route gate)" || fail "#6 write-gate message" "(got '$(jqv '.message')')"
else
  echo "  ! skipped #6 limited-role checks (could not seed SmokeFieldViewer role via go)"
fi

echo "== #10 line-manager suite (validation + picker + direct-reports + rich brief) =="
req POST "$BASE/employees" "{\"email\":\"smk-mgr-$SFX@x.com\",\"password\":\"Pass12345\",\"first_name\":\"Mgr\",\"last_name\":\"$SFX\"}"
MGR=$(jqv '.data.id')
req POST "$BASE/employees" "{\"email\":\"smk-rep-$SFX@x.com\",\"password\":\"Pass12345\",\"first_name\":\"Rep\",\"last_name\":\"$SFX\",\"manager_id\":\"$MGR\"}"
REP=$(jqv '.data.id')
eq 201 "$CODE" "#10 create employee with a valid manager (201)"
req GET "$BASE/employees/$REP"
eq "$MGR" "$(jqv '.data.manager.id')" "#10 read embeds rich manager brief"
eq "true" "$(jqv '.data.manager.is_active')" "#10 manager brief carries is_active"
req PATCH "$BASE/employees/$MGR" "{\"manager_id\":\"$MGR\"}"
eq 400 "$CODE" "#10 self-as-manager rejected (400)"
req PATCH "$BASE/employees/$MGR" "{\"manager_id\":\"$REP\"}"
eq 400 "$CODE" "#10 cycle rejected (400)"
req POST "$BASE/employees" "{\"email\":\"smk-bad-$SFX@x.com\",\"password\":\"Pass12345\",\"first_name\":\"Bad\",\"last_name\":\"Test\",\"manager_id\":\"00000000-0000-0000-0000-000000000000\"}"
eq 400 "$CODE" "#10 nonexistent manager rejected (400)"
req GET "$BASE/employees/manager-candidates?for_employee_id=$MGR"
eq 200 "$CODE" "#10 manager-candidates (200)"
eq "false" "$(echo "$RESP" | jq --arg id "$MGR" '[.data[].id] | index($id) != null')" "#10 candidates exclude self"
eq "false" "$(echo "$RESP" | jq --arg id "$REP" '[.data[].id] | index($id) != null')" "#10 candidates exclude subordinate"
req GET "$BASE/employees/$MGR/direct-reports"
eq 200 "$CODE" "#10 direct-reports (200)"
eq "true" "$(echo "$RESP" | jq --arg id "$REP" '[.data[].id] | index($id) != null')" "#10 direct-reports includes the report"

echo "== parity-2: inline skill_ids on create + multi-select department filter =="
# POST /skills is multipart/form-data — use curl -F directly (req helper sends JSON).
SKILL_RESP=$(curl -s -w $'\n%{http_code}' -X POST "$BASE/skills" \
  -H "Authorization: Bearer $TOK" \
  -F "name=SmokeSkill$SFX" -F "description=smoke")
SKILL_CODE="${SKILL_RESP##*$'\n'}"; SKILL_BODY="${SKILL_RESP%$'\n'*}"
SKILL_ID=$(echo "$SKILL_BODY" | jq -r '.data.id')
eq 201 "$SKILL_CODE" "#parity2 create skill (201)"

# Create employee with inline skill_ids; assert the skill is echoed back.
req POST "$BASE/employees" "{\"email\":\"smk-sk-$SFX@x.com\",\"password\":\"Pass12345\",\"first_name\":\"Skilled\",\"last_name\":\"$SFX\",\"skill_ids\":[\"$SKILL_ID\"]}"
EMP_SK=$(jqv '.data.id')
eq 201 "$CODE" "#parity2 create employee with skill_ids (201)"
eq 1 "$(jqv '.data.skills | length')" "#parity2 skill echoed on create (length=1)"

# Multi-select department filter: GET /employees?department_id=<A>&department_id=<B>
# Use EMP (from the main smoke employee) and EMP_SK dept IDs — both have no dept assigned,
# so filter by two nil-safe UUIDs just proves the endpoint returns 200 with array params.
DEPT_A="00000000-0000-0000-0000-000000000001"
DEPT_B="00000000-0000-0000-0000-000000000002"
req GET "$BASE/employees?department_id=$DEPT_A&department_id=$DEPT_B"
eq 200 "$CODE" "#parity2 multi-select department_id filter returns 200"

echo
echo "==================== SMOKE SUMMARY ===================="
echo "  PASS: $PASS    FAIL: $FAIL"
echo "======================================================="
[ "$FAIL" -eq 0 ]
