package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/exnodes/hrm-api/internal/middleware"
)

// newCORSRouter wires the CORS middleware into a throwaway router with a
// single OK handler. Callers send a request and assert on the response
// headers.
func newCORSRouter(allowed []string, env string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CORS(allowed, env))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func TestCORS_AllowlistedOrigin_EchoesAndAllowsCredentials(t *testing.T) {
	r := newCORSRouter([]string{"http://localhost:3000"}, "production")
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"),
		"credentialed CORS requires the credentials header alongside the echoed Origin")
	assert.Equal(t, "Origin", w.Header().Get("Vary"))
}

func TestCORS_NonAllowlistedOrigin_NoCORSHeaders(t *testing.T) {
	r := newCORSRouter([]string{"http://localhost:3000"}, "production")
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://evil.example")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"),
		"non-allowlisted origin must NOT be echoed")
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"),
		"credentials must NOT be granted to a non-allowlisted origin")
}

func TestCORS_DevEmptyAllowlist_EchoesRequestOrigin(t *testing.T) {
	r := newCORSRouter(nil, "development")
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://localhost:3001")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "http://localhost:3001", w.Header().Get("Access-Control-Allow-Origin"),
		"dev empty-allowlist mode must echo the requesting Origin (NOT '*'); wildcard would break credentialed SSE")
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "Origin", w.Header().Get("Vary"))
}

func TestCORS_NoOriginHeader_NoCORSHeaders(t *testing.T) {
	r := newCORSRouter(nil, "development")
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"),
		"requests without an Origin header are non-browser; no CORS reply needed")
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORS_OptionsPreflight_ShortCircuits(t *testing.T) {
	r := newCORSRouter([]string{"http://localhost:3000"}, "production")
	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code,
		"OPTIONS preflight must short-circuit with 204")
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}
