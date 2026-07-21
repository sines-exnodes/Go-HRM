package dto

import (
	"fmt"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// emailValidator is a standalone instance used only to reuse go-playground's
// built-in email check inside optionalEmail. Kept package-private so it never
// gets tangled with gin's request-binding engine.
var emailValidator = validator.New()

// RegisterValidators installs the project's custom validation tags onto gin's
// default binding validator. Call once at startup, before any request is
// served. Re-registering the same tag simply overwrites it, so it is safe to
// call from tests too.
func RegisterValidators() error {
	eng, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return fmt.Errorf("gin binding validator engine is not *validator.Validate (got %T)", binding.Validator.Engine())
	}
	return eng.RegisterValidation("optemail", optionalEmail)
}

// optionalEmail passes an empty string and otherwise requires a valid email.
//
// This exists because `omitempty,email` on a *string does NOT skip a non-nil
// pointer to "". omitempty on a pointer tests nil-ness, not emptiness, so a
// personal_email sent as "" to CLEAR the field reaches the `email` tag and is
// rejected — even though the client legitimately means "unset this". Proven in
// TestOptionalEmail. Use `omitempty,optemail`: omitempty still skips a nil
// pointer, and optemail then accepts the empty-string-to-clear case.
func optionalEmail(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return true
	}
	return emailValidator.Var(s, "email") == nil
}
