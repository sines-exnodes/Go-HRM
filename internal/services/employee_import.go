package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/exnodes/hrm-api/internal/dto"
	apperrors "github.com/exnodes/hrm-api/internal/errors"
)

const (
	employeeImportMaxBytes = 2 * 1024 * 1024
	employeeImportMaxRows  = 500
)

// Canonical header set — any other header is a whole-file error.
var employeeImportAllowedHeaders = map[string]struct{}{
	"email": {}, "first_name": {}, "last_name": {}, "phone": {},
	"personal_email": {}, "gender": {}, "dob": {}, "nationality": {},
	"id_number": {}, "permanent_address": {}, "current_address": {},
	"education": {}, "marital_status": {}, "experience_year": {},
	"department": {}, "position": {}, "manager_email": {}, "join_date": {},
	"role": {}, "is_active": {}, "contract_type": {},
	"basic_salary": {}, "insurance_salary": {},
	"bank_account": {}, "bank_name": {}, "bank_holder_name": {}, "payment_method": {},
	"social_insurance_number": {}, "tax_identification_number": {},
}

var employeeImportRequiredHeaders = []string{"email", "first_name", "last_name"}

// ImportCSV parses a CSV body and creates one employee per data row via
// Create. Per-row failures never abort the file (partial success); only
// whole-file problems (empty, over limit, bad headers) return a top-level error.
// Salary/banking columns are gated by perms the same way the create handler does.
// Invite fan-out is owned by the handler — do not inject PasswordResetService here.
func (s *EmployeeService) ImportCSV(
	ctx context.Context,
	file []byte,
	perms EmployeeFieldPerms,
) (*dto.EmployeeImportResult, error) {
	if len(file) == 0 {
		return nil, apperrors.ErrBadRequest("CSV file is empty")
	}
	if len(file) > employeeImportMaxBytes {
		return nil, apperrors.ErrBadRequest("CSV file exceeds 2MB limit")
	}

	r := csv.NewReader(bytes.NewReader(file))
	r.LazyQuotes = true
	r.FieldsPerRecord = -1 // allow short rows; missing cells treated as empty
	records, err := r.ReadAll()
	if err != nil {
		return nil, apperrors.ErrBadRequest(fmt.Sprintf("invalid CSV: %v", err))
	}
	if len(records) < 2 {
		return nil, apperrors.ErrBadRequest("CSV must include a header row and at least one data row")
	}

	headers, err := normalizeImportHeaders(records[0])
	if err != nil {
		return nil, err
	}

	dataRows := records[1:]
	if len(dataRows) > employeeImportMaxRows {
		return nil, apperrors.ErrBadRequest(fmt.Sprintf(
			"CSV exceeds maximum of %d data rows", employeeImportMaxRows))
	}

	out := &dto.EmployeeImportResult{
		TotalRows: len(dataRows),
		Results:   make([]dto.EmployeeImportRowResult, 0, len(dataRows)),
	}

	for i, record := range dataRows {
		rowNum := i + 2 // 1-based file line (header = 1)
		cells := rowMap(headers, record)
		email := strings.TrimSpace(cells["email"])

		in, rowErr := s.buildImportCreate(ctx, cells, perms)
		if rowErr != nil {
			out.Failed++
			out.Results = append(out.Results, dto.EmployeeImportRowResult{
				Row:   rowNum,
				OK:    false,
				Email: email,
				Error: importRowErrMsg(rowErr),
			})
			continue
		}

		created, createErr := s.Create(ctx, in)
		if createErr != nil {
			out.Failed++
			out.Results = append(out.Results, dto.EmployeeImportRowResult{
				Row:   rowNum,
				OK:    false,
				Email: email,
				Error: importRowErrMsg(createErr),
			})
			continue
		}

		empID := created.ID
		userID := created.UserID
		out.Created++
		out.Results = append(out.Results, dto.EmployeeImportRowResult{
			Row:        rowNum,
			OK:         true,
			Email:      created.Email,
			EmployeeID: &empID,
			UserID:     &userID,
		})
	}

	return out, nil
}

func normalizeImportHeaders(raw []string) ([]string, error) {
	if len(raw) == 0 {
		return nil, apperrors.ErrBadRequest("CSV header row is empty")
	}
	headers := make([]string, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for i, h := range raw {
		name := strings.ToLower(strings.TrimSpace(h))
		if name == "" {
			return nil, apperrors.ErrBadRequest(fmt.Sprintf("empty header at column %d", i+1))
		}
		if _, ok := employeeImportAllowedHeaders[name]; !ok {
			return nil, apperrors.ErrBadRequest(fmt.Sprintf("unknown CSV column: %s", name))
		}
		if _, dup := seen[name]; dup {
			return nil, apperrors.ErrBadRequest(fmt.Sprintf("duplicate CSV column: %s", name))
		}
		seen[name] = struct{}{}
		headers[i] = name
	}
	for _, req := range employeeImportRequiredHeaders {
		if _, ok := seen[req]; !ok {
			return nil, apperrors.ErrBadRequest(fmt.Sprintf("missing required CSV column: %s", req))
		}
	}
	return headers, nil
}

func rowMap(headers []string, record []string) map[string]string {
	m := make(map[string]string, len(headers))
	for i, h := range headers {
		val := ""
		if i < len(record) {
			val = strings.TrimSpace(record[i])
		}
		m[h] = val
	}
	return m
}

func importRowErrMsg(err error) string {
	if ae, ok := apperrors.As(err); ok {
		return ae.Message
	}
	return err.Error()
}

func (s *EmployeeService) buildImportCreate(
	ctx context.Context,
	cells map[string]string,
	perms EmployeeFieldPerms,
) (dto.EmployeeCreate, error) {
	var in dto.EmployeeCreate

	email := cells["email"]
	if email == "" {
		return in, apperrors.ErrBadRequest("email is required")
	}
	first := cells["first_name"]
	if first == "" {
		return in, apperrors.ErrBadRequest("first_name is required")
	}
	last := cells["last_name"]
	if last == "" {
		return in, apperrors.ErrBadRequest("last_name is required")
	}
	in.Email = email
	in.FirstName = first
	in.LastName = last

	if v := cells["phone"]; v != "" {
		in.Phone = &v
	}
	if v := cells["personal_email"]; v != "" {
		in.PersonalEmail = &v
	}
	if v := cells["gender"]; v != "" {
		in.Gender = &v
	}
	if v := cells["nationality"]; v != "" {
		in.Nationality = &v
	}
	if v := cells["id_number"]; v != "" {
		in.IDNumber = &v
	}
	if v := cells["permanent_address"]; v != "" {
		in.PermanentAddress = &v
	}
	if v := cells["current_address"]; v != "" {
		in.CurrentAddress = &v
	}
	if v := cells["education"]; v != "" {
		in.Education = &v
	}
	if v := cells["marital_status"]; v != "" {
		in.MaritalStatus = &v
	}
	if v := cells["contract_type"]; v != "" {
		in.ContractType = &v
	}
	if v := cells["social_insurance_number"]; v != "" {
		in.SocialInsuranceNumber = &v
	}
	if v := cells["tax_identification_number"]; v != "" {
		in.TaxIdentificationNumber = &v
	}

	if v := cells["dob"]; v != "" {
		t, err := parseImportDate(v)
		if err != nil {
			return in, apperrors.ErrBadRequest("invalid dob: use YYYY-MM-DD")
		}
		in.DOB = &t
	}
	if v := cells["join_date"]; v != "" {
		t, err := parseImportDate(v)
		if err != nil {
			return in, apperrors.ErrBadRequest("invalid join_date: use YYYY-MM-DD")
		}
		in.JoinDate = &t
	}

	if v := cells["experience_year"]; v != "" {
		y, err := strconv.Atoi(v)
		if err != nil {
			return in, apperrors.ErrBadRequest("invalid experience_year: must be an integer year")
		}
		in.ExperienceYear = &y
	}

	if v := cells["is_active"]; v != "" {
		b, err := parseImportBool(v)
		if err != nil {
			return in, apperrors.ErrBadRequest("invalid is_active: use true/false/1/0")
		}
		in.IsActive = &b
	}
	// empty is_active → leave nil so Create defaults true

	// Salary columns: non-empty without perm → row error.
	salarySet := cells["basic_salary"] != "" || cells["insurance_salary"] != ""
	if err := GuardSalaryWrite(salarySet, perms); err != nil {
		return in, err
	}
	if v := cells["basic_salary"]; v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return in, apperrors.ErrBadRequest("invalid basic_salary")
		}
		in.BasicSalary = &f
	}
	if v := cells["insurance_salary"]; v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return in, apperrors.ErrBadRequest("invalid insurance_salary")
		}
		in.InsuranceSalary = &f
	}

	// Banking columns: non-empty without perm → row error.
	bankingSet := cells["bank_account"] != "" || cells["bank_name"] != "" ||
		cells["bank_holder_name"] != "" || cells["payment_method"] != ""
	if err := GuardBankingWrite(bankingSet, perms); err != nil {
		return in, err
	}
	if v := cells["bank_account"]; v != "" {
		in.BankAccount = &v
	}
	if v := cells["bank_name"]; v != "" {
		in.BankName = &v
	}
	if v := cells["bank_holder_name"]; v != "" {
		in.BankHolderName = &v
	}
	if v := cells["payment_method"]; v != "" {
		in.PaymentMethod = &v
	}

	if v := cells["department"]; v != "" {
		id, err := s.resolveImportDepartment(ctx, v)
		if err != nil {
			return in, err
		}
		in.DepartmentID = id
	}
	if v := cells["position"]; v != "" {
		id, err := s.resolveImportPosition(ctx, v)
		if err != nil {
			return in, err
		}
		in.PositionID = id
	}
	if v := cells["role"]; v != "" {
		id, err := s.resolveImportRole(ctx, v)
		if err != nil {
			return in, err
		}
		in.RoleIDs = []uuid.UUID{id}
	}
	if v := cells["manager_email"]; v != "" {
		id, err := s.resolveImportManager(ctx, v)
		if err != nil {
			return in, err
		}
		in.ManagerID = id
	}

	return in, nil
}

func parseImportDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func parseImportBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool: %s", s)
	}
}

func (s *EmployeeService) resolveImportDepartment(ctx context.Context, name string) (*uuid.UUID, error) {
	if s.depts == nil {
		return nil, apperrors.ErrInternal("department repository not configured")
	}
	d, err := s.depts.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, apperrors.ErrBadRequest(fmt.Sprintf("unknown department: %s", name))
	}
	id := d.ID
	return &id, nil
}

func (s *EmployeeService) resolveImportPosition(ctx context.Context, name string) (*uuid.UUID, error) {
	if s.positions == nil {
		return nil, apperrors.ErrInternal("position repository not configured")
	}
	p, err := s.positions.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, apperrors.ErrBadRequest(fmt.Sprintf("unknown position: %s", name))
	}
	id := p.ID
	return &id, nil
}

func (s *EmployeeService) resolveImportRole(ctx context.Context, name string) (uuid.UUID, error) {
	role, err := s.roles.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, apperrors.ErrBadRequest(fmt.Sprintf("unknown role: %s", name))
		}
		return uuid.Nil, err
	}
	return role.ID, nil
}

func (s *EmployeeService) resolveImportManager(ctx context.Context, email string) (*uuid.UUID, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrBadRequest(fmt.Sprintf("unknown manager email: %s", email))
		}
		return nil, err
	}
	if !u.IsActive {
		return nil, apperrors.ErrBadRequest(fmt.Sprintf("manager is inactive: %s", email))
	}
	emp, err := s.emps.FindByUserID(ctx, u.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrBadRequest(fmt.Sprintf("manager has no employee profile: %s", email))
		}
		return nil, err
	}
	id := emp.ID
	return &id, nil
}
