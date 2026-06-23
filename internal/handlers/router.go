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
// Swagger UI is mounted at /swagger/*any and Scalar (modern API reference UI
// over the same OpenAPI spec) at /docs.
func RegisterRoutes(r *gin.Engine, h *Handlers, swaggerEnabled bool) {
	// Liveness probe — un-namespaced so load balancers can hit it without
	// knowing about /api/v1.
	r.GET("/health", h.Health.Health)

	if swaggerEnabled {
		r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
		r.GET("/docs", func(c *gin.Context) {
			c.Data(200, "text/html; charset=utf-8", []byte(scalarDocsHTML))
		})
	}

	// /api/v1 is registered here so Phase 1+ has a place to plug routes in.
	// No routes are mounted on it yet.
	_ = r.Group("/api/v1")
}

// scalarDocsHTML is the page served at /docs. It loads Scalar from jsDelivr
// and points it at gin-swagger's /swagger/doc.json. CDN dependency is
// acceptable here because /docs is gated by swaggerEnabled — dev/non-prod only.
const scalarDocsHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Exnodes HRM API — Reference</title>
</head>
<body>
  <script
    id="api-reference"
    data-url="/swagger/doc.json"
    data-configuration='{"authentication":{"preferredSecurityScheme":"BearerAuth"}}'
  ></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>
`
