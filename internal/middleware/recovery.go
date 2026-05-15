package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	apperrors "github.com/exnodes/hrm-api/internal/errors"
)

// Recovery converts a panic in any downstream handler into a 500 JSON
// envelope and logs the stack trace. It must be registered before
// ErrorHandler so the panic is rewritten into the same envelope shape.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] %v\n%s", r, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "internal server error",
					"code":    apperrors.CodeInternal,
				})
			}
		}()
		c.Next()
	}
}
