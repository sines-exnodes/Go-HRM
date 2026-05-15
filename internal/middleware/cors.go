package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS returns a permissive CORS middleware suitable for early development
// (`Access-Control-Allow-Origin: *`). Production deployments should swap this
// for a stricter allow-list — tracked as a Phase 2+ hardening item.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Max-Age", "600")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
