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
//     `Access-Control-Allow-Credentials: true` is sent on the same response so
//     EventSource(withCredentials) and credentialed fetch() work.
//   - If allowedOrigins is empty and appEnv == "development", the request
//     Origin is echoed back as-is (effectively a per-request wildcard) +
//     credentials true. We deliberately do NOT use `Access-Control-Allow-Origin: *`
//     because browsers reject it on credentialed requests (cookies, basic auth,
//     EventSource(withCredentials)).
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
		originSet := false

		if origin != "" {
			if len(allowed) > 0 {
				if _, ok := allowed[origin]; ok {
					c.Header("Access-Control-Allow-Origin", origin)
					c.Header("Vary", "Origin")
					originSet = true
				}
			} else if appEnv == "development" {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				originSet = true
			}
		}

		if originSet {
			// Required for credentialed requests (cookies / EventSource
			// withCredentials). Browsers refuse cookies with wildcard
			// origin, so we always pair this with an echoed concrete
			// Origin — never `*`.
			c.Header("Access-Control-Allow-Credentials", "true")
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
