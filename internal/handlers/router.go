package handlers

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
)

// Handlers bundles every handler instance the router needs to wire. Phase 1
// will extend this struct with Auth, User, etc.
type Handlers struct {
	Health *HealthHandler
}

// NewHandlers constructs every handler in the bundle. Currently health-only;
// future phases add fields here in alphabetical order.
func NewHandlers() *Handlers {
	return &Handlers{
		Health: NewHealthHandler(),
	}
}

// RegisterRoutes wires every route onto r. If swaggerEnabled is true, the
// Swagger UI is mounted at /swagger/*any and Stoplight Elements (modern
// three-panel reference UI over the same OpenAPI spec) at /docs.
func RegisterRoutes(r *gin.Engine, h *Handlers, swaggerEnabled bool) {
	// Liveness probe — un-namespaced so load balancers can hit it without
	// knowing about /api/v1.
	r.GET("/health", h.Health.Health)

	if swaggerEnabled {
		r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
		r.GET("/docs", func(c *gin.Context) {
			c.Data(200, "text/html; charset=utf-8", []byte(stoplightDocsHTML))
		})
	}

	// /api/v1 is registered here so Phase 1+ has a place to plug routes in.
	// No routes are mounted on it yet.
	_ = r.Group("/api/v1")
}

// stoplightDocsHTML is the page served at /docs. It loads Stoplight Elements
// from unpkg (pinned) and points it at gin-swagger's /swagger/doc.json so the
// same OpenAPI spec drives both UIs. CDN dependency is acceptable here because
// /docs is gated by swaggerEnabled — i.e. dev/non-prod only.
const stoplightDocsHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>Exnodes HRM API — Reference</title>
  <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements@8.0.1/styles.min.css">
</head>
<body style="margin:0">
  <elements-api apiDescriptionUrl="/swagger/doc.json" router="hash" layout="sidebar"></elements-api>
  <script src="https://unpkg.com/@stoplight/elements@8.0.1/web-components.min.js"></script>
</body>
</html>
`
