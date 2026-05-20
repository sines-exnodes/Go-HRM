package main

import (
	"github.com/google/uuid"

	"github.com/exnodes/hrm-api/internal/sse"
)

// sseHubAdapter bridges the concrete *sse.Hub to the service-layer
// HubBroadcaster interface. The interface uses a string event type and
// the concrete one uses sse.Event{Type, Data} — the adapter does the
// trivial translation so the service can be unit-tested with a tiny mock
// (services.captureHub) that has no transitive sse dependency.
type sseHubAdapter struct{ hub *sse.Hub }

// Broadcast satisfies services.HubBroadcaster.
func (a sseHubAdapter) Broadcast(eventType string, data any, filter func(uuid.UUID) bool) {
	a.hub.Broadcast(sse.Event{Type: eventType, Data: data}, sse.FilterFunc(filter))
}
