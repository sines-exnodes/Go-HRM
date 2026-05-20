// Package sse provides an in-memory, single-instance pub/sub hub used to
// stream announcement (and future notification) events to connected
// clients over Server-Sent Events.
//
// SCALING LIMIT: This hub holds all subscriber channels in memory inside
// a single process. Horizontal scaling beyond one replica requires a
// shared backplane (Redis pub/sub, NATS, etc.) — explicitly out of scope
// for Phase 7. Document this constraint in deployment docs.
package sse

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// Event is the payload broadcast to subscribers. Type names the kind
// (e.g. "announcement_published"); Data is any JSON-marshalable payload.
type Event struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

// FilterFunc decides whether a given userID should receive an Event.
// A nil filter delivers to all subscribers.
type FilterFunc func(userID uuid.UUID) bool

// client is one connected SSE listener. Two browser tabs from the same
// user produce two distinct clients with the same userID but different
// ids — same-user multi-tab is intended.
type client struct {
	id     string
	userID uuid.UUID
	send   chan []byte
}

// Hub is the in-memory broadcast registry. Goroutine-safe.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*client
	closed  bool
}

// NewHub creates an empty hub.
func NewHub() *Hub {
	return &Hub{clients: make(map[string]*client)}
}

// Subscribe registers a new listener for userID. Returns:
//   - a receive-only channel that yields JSON-marshaled Event payloads
//     (callers wrap them in "event: ...\ndata: ...\n\n" for the wire);
//   - an idempotent unsubscribe function — safe to call multiple times.
//
// The send buffer is small (16). Slow consumers whose buffer fills
// silently drop events (see Broadcast) so a stuck client cannot block
// the publisher.
func (h *Hub) Subscribe(userID uuid.UUID) (<-chan []byte, func()) {
	c := &client{
		id:     uuid.NewString(),
		userID: userID,
		send:   make(chan []byte, 16),
	}
	h.mu.Lock()
	if h.closed {
		// Already stopped — return a closed channel and a no-op unsub.
		h.mu.Unlock()
		close(c.send)
		return c.send, func() {}
	}
	h.clients[c.id] = c
	h.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			if _, ok := h.clients[c.id]; ok {
				delete(h.clients, c.id)
				close(c.send)
			}
		})
	}
	return c.send, unsubscribe
}

// Broadcast sends event to every subscriber for whom filter(userID)
// returns true. A nil filter delivers to all subscribers. Slow consumers
// whose channel buffer is full are skipped — events are dropped per
// client, never blocking the publisher.
func (h *Hub) Broadcast(event Event, filter FilterFunc) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.closed {
		return
	}
	for _, c := range h.clients {
		if filter != nil && !filter(c.userID) {
			continue
		}
		select {
		case c.send <- payload:
		default:
			// Buffer full — drop for this slow client.
		}
	}
}

// ClientCount returns the current subscriber count. Intended for
// telemetry/tests; not a public API.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Stop closes all client channels and marks the hub as closed. Idempotent.
// Intended for tests and graceful shutdown.
func (h *Hub) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return
	}
	h.closed = true
	for id, c := range h.clients {
		close(c.send)
		delete(h.clients, id)
	}
}
