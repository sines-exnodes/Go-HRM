package sse

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestHub_SubscribeAndBroadcastAll verifies the baseline pub/sub path:
// one subscriber receives an unfiltered broadcast.
func TestHub_SubscribeAndBroadcastAll(t *testing.T) {
	hub := NewHub()
	defer hub.Stop()

	uid := uuid.New()
	ch, unsubscribe := hub.Subscribe(uid)
	defer unsubscribe()

	hub.Broadcast(Event{Type: "announcement_published", Data: map[string]any{"x": 1}}, nil)

	select {
	case b := <-ch:
		var ev Event
		if err := json.Unmarshal(b, &ev); err != nil {
			t.Fatalf("unmarshal event: %v", err)
		}
		if ev.Type != "announcement_published" {
			t.Fatalf("want type announcement_published, got %s", ev.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("did not receive broadcast within 500ms")
	}
}

// TestHub_BroadcastWithFilter verifies the FilterFunc gate: only matched
// userIDs receive the event.
func TestHub_BroadcastWithFilter(t *testing.T) {
	hub := NewHub()
	defer hub.Stop()

	u1 := uuid.New()
	u2 := uuid.New()
	ch1, un1 := hub.Subscribe(u1)
	defer un1()
	ch2, un2 := hub.Subscribe(u2)
	defer un2()

	filter := func(userID uuid.UUID) bool { return userID == u1 }
	hub.Broadcast(Event{Type: "announcement_published", Data: map[string]any{}}, filter)

	gotU1 := false
	gotU2 := false
	deadline := time.After(300 * time.Millisecond)
	for {
		select {
		case <-ch1:
			gotU1 = true
		case <-ch2:
			gotU2 = true
		case <-deadline:
			if !gotU1 {
				t.Fatal("u1 did not receive filtered broadcast")
			}
			if gotU2 {
				t.Fatal("u2 unexpectedly received filtered broadcast")
			}
			return
		}
	}
}

// TestHub_UnsubscribeStopsDelivery verifies an unsubscribed client's
// channel is closed and no further events arrive.
func TestHub_UnsubscribeStopsDelivery(t *testing.T) {
	hub := NewHub()
	defer hub.Stop()

	uid := uuid.New()
	ch, unsubscribe := hub.Subscribe(uid)
	unsubscribe()

	hub.Broadcast(Event{Type: "x"}, nil)

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("received event after unsubscribe")
		}
		// Closed channel — expected.
	case <-time.After(150 * time.Millisecond):
		// Also acceptable.
	}
}

// TestHub_UnsubscribeIdempotent verifies multiple unsubscribe calls do
// not panic (sync.Once gate).
func TestHub_UnsubscribeIdempotent(t *testing.T) {
	hub := NewHub()
	defer hub.Stop()

	_, unsubscribe := hub.Subscribe(uuid.New())
	unsubscribe()
	unsubscribe()
	unsubscribe()
}

// TestHub_ConcurrentSubscribers fans out to N concurrent subscribers and
// verifies each receives the broadcast within the deadline.
func TestHub_ConcurrentSubscribers(t *testing.T) {
	hub := NewHub()
	defer hub.Stop()

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n)
	chs := make([]<-chan []byte, n)
	uns := make([]func(), n)
	for i := 0; i < n; i++ {
		ch, un := hub.Subscribe(uuid.New())
		chs[i] = ch
		uns[i] = un
		go func(idx int) {
			defer wg.Done()
			select {
			case <-chs[idx]:
			case <-time.After(500 * time.Millisecond):
				t.Errorf("client %d timed out", idx)
			}
		}(i)
	}
	hub.Broadcast(Event{Type: "ping"}, nil)
	wg.Wait()
	for _, un := range uns {
		un()
	}
}

// TestHub_StopClosesAllClients verifies Stop() drains every registered
// client and is idempotent.
func TestHub_StopClosesAllClients(t *testing.T) {
	hub := NewHub()

	_, un1 := hub.Subscribe(uuid.New())
	_, un2 := hub.Subscribe(uuid.New())
	if hub.ClientCount() != 2 {
		t.Fatalf("want 2 clients, got %d", hub.ClientCount())
	}

	hub.Stop()
	if hub.ClientCount() != 0 {
		t.Fatalf("Stop() did not drain clients, got %d", hub.ClientCount())
	}
	// Idempotent.
	hub.Stop()
	// Unsub after stop must not panic.
	un1()
	un2()
}

// TestHub_SubscribeAfterStopReturnsClosed verifies post-Stop Subscribe
// returns a closed channel + no-op unsubscribe (no leak).
func TestHub_SubscribeAfterStopReturnsClosed(t *testing.T) {
	hub := NewHub()
	hub.Stop()

	ch, un := hub.Subscribe(uuid.New())
	un() // no-op
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("channel should be closed after Stop")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("channel should be closed (instant), but blocked")
	}
}
