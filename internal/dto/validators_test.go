package dto

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sp(s string) *string { return &s }

// TestOptionalEmail is the regression guard for the bug where a client sending
// personal_email:"" to clear the field got a 400 "failed on the 'email' tag".
// The empty-string case is the one that was broken; the others must keep
// working so we don't trade the bug for lost validation.
func TestOptionalEmail(t *testing.T) {
	require.NoError(t, RegisterValidators())

	empty := ""
	cases := []struct {
		name  string
		email *string
		valid bool
	}{
		{"nil is skipped", nil, true},
		{"empty string clears the field", &empty, true},
		{"valid email passes", sp("a@b.com"), true},
		{"invalid email is rejected", sp("notanemail"), false},
		{"partial email is rejected", sp("a@"), false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// EmployeeUpdate — the struct in the reported bug.
			err := binding.Validator.ValidateStruct(&EmployeeUpdate{PersonalEmail: c.email})
			if c.valid {
				assert.NoError(t, err, "EmployeeUpdate")
			} else {
				assert.Error(t, err, "EmployeeUpdate")
			}

			// EmployeeSelfUpdate shares the tag and must behave identically.
			err = binding.Validator.ValidateStruct(&EmployeeSelfUpdate{PersonalEmail: c.email})
			if c.valid {
				assert.NoError(t, err, "EmployeeSelfUpdate")
			} else {
				assert.Error(t, err, "EmployeeSelfUpdate")
			}
		})
	}
}

// TestOptionalEmail_CreateStillValidates guards that swapping the tag on the
// create DTO did not silently drop email validation there. Create requires
// first/last name, so we assert specifically on the personal_email outcome by
// supplying valid names.
func TestOptionalEmail_CreateStillValidates(t *testing.T) {
	require.NoError(t, RegisterValidators())

	// Empty personal_email must not block a create.
	errEmpty := binding.Validator.ValidateStruct(&EmployeeCreate{
		FirstName:     "Vy",
		LastName:      "Nguyen",
		Email:         "login@example.com",
		PersonalEmail: sp(""),
	})
	assert.NoError(t, errEmpty)

	// A malformed personal_email must still be rejected on create.
	errBad := binding.Validator.ValidateStruct(&EmployeeCreate{
		FirstName:     "Vy",
		LastName:      "Nguyen",
		Email:         "login@example.com",
		PersonalEmail: sp("notanemail"),
	})
	assert.Error(t, errBad)
}
