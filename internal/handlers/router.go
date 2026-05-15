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
// Swagger UI is mounted at /swagger/*any.
func RegisterRoutes(r *gin.Engine, h *Handlers, swaggerEnabled bool) {
	// Liveness probe — un-namespaced so load balancers can hit it without
	// knowing about /api/v1.
	r.GET("/health", h.Health.Health)

	if swaggerEnabled {
		r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
	}

	// /api/v1 is registered here so Phase 1+ has a place to plug routes in.
	// No routes are mounted on it yet.
	_ = r.Group("/api/v1")
}
