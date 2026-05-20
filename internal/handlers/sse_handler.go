package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/sse"
)

// SSEHandler owns the long-lived /sse/announcements stream. One handler
// instance per server boot — it depends on a shared *sse.Hub singleton.
type SSEHandler struct {
	hub *sse.Hub
}

// NewSSEHandler constructs the handler. hub must be the same singleton
// passed to the announcement service so broadcasts reach connected
// subscribers.
func NewSSEHandler(hub *sse.Hub) *SSEHandler {
	return &SSEHandler{hub: hub}
}

// Stream godoc
// @Summary      Subscribe to announcement events via Server-Sent Events
// @Description  Long-lived SSE stream. Emits an `announcement_published` event whenever a new announcement is published. Clients should reconnect on disconnect. Auth token may be passed as `Authorization: Bearer ...` OR `?token=` query param (EventSource cannot set headers — REVISION NOTES #9). Sends a `: keepalive` comment every 30s to keep proxies from closing idle connections.
// @Tags         sse
// @Produce      text/event-stream
// @Param        token  query     string  false  "JWT access token (alternative to Authorization header)"
// @Success      200    {string}  string  "event stream"
// @Failure      401    {object}  map[string]interface{}
// @Router       /api/v1/sse/announcements [get]
// @Security     BearerAuth
func (h *SSEHandler) Stream(c *gin.Context) {
	u, ok := currentUser(c)
	if !ok {
		// currentUser writes the 401 error envelope and returns false.
		return
	}

	ch, unsubscribe := h.hub.Subscribe(u.ID)
	defer unsubscribe()

	// SSE response headers.
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	// Disable nginx buffering (the proxy header). Without this, events
	// can sit in the proxy's buffer for many seconds.
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		// Should never happen with Gin's default ResponseWriter, but
		// fail loud if some middleware swapped it out.
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	flusher.Flush()

	// Initial hello frame — lets the client confirm the stream is alive
	// before any real events arrive.
	connID := uuid.NewString()
	_, _ = fmt.Fprintf(c.Writer, "event: connected\ndata: {\"connection_id\":\"%s\"}\n\n", connID)
	flusher.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	gone := c.Request.Context().Done()
	for {
		select {
		case <-gone:
			return
		case payload, open := <-ch:
			if !open {
				return
			}
			if _, err := io.WriteString(
				c.Writer,
				fmt.Sprintf("event: announcement_published\ndata: %s\n\n", string(payload)),
			); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			// SSE comment line — invisible to consumers but keeps the
			// TCP connection from going idle.
			if _, err := io.WriteString(c.Writer, ": keepalive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
