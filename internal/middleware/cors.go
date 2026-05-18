package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS returns a CORS middleware whose allowed origins are gated by
// configuration.
//
//   - If allowedOrigins is non-empty, the request Origin is echoed back only
//     when it appears in the list (with `Vary: Origin` so caches key on it).
//   - If allowedOrigins is empty and appEnv == "development", a permissive
//     `Access-Control-Allow-Origin: *` is used to ease local work.
//
// An empty list in production is a misconfiguration and must be caught at
// startup (see cmd/server/main.go) — this middleware does not fail per-request.
func CORS(allowedOrigins []string, appEnv string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if o != "" {
			allowed[o] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if len(allowed) > 0 {
			if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
		} else if appEnv == "development" {
			c.Header("Access-Control-Allow-Origin", "*")
		}

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
