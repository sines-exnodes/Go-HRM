package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/exnodes/hrm-api/internal/errors"
)

// errorEnvelope is the JSON shape returned for any error response.
// It deliberately mirrors the success envelope's `success` field so the FE
// can detect failure by checking a single boolean.
type errorEnvelope struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Code    string         `json:"code"`
	Details map[string]any `json:"details,omitempty"`
}

// ErrorHandler converts any errors written via c.Error(err) into the standard
// JSON envelope. It must be registered before any handler that uses c.Error
// (i.e. at the very top of the middleware chain).
//
// Handlers should *not* call c.JSON themselves on the error path; they should
// `c.Error(apperrors.ErrXxx(...))` and `return`. This middleware writes the
// response after the handler chain finishes.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		// Use the last error written by the handler — that is the one the
		// developer most recently attached and is the most specific.
		last := c.Errors.Last().Err

		if ae, ok := apperrors.As(last); ok {
			c.AbortWithStatusJSON(ae.HTTP, errorEnvelope{
				Success: false,
				Message: ae.Message,
				Code:    ae.Code,
				Details: ae.Details,
			})
			return
		}

		// Unknown error → 500 internal_error. We deliberately do NOT leak the
		// raw error text to the FE; it is captured in Gin's error log already.
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorEnvelope{
			Success: false,
			Message: "internal server error",
			Code:    apperrors.CodeInternal,
		})
	}
}
