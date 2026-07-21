package services

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ict is Vietnam time (GMT+7) as a fixed zone, so the timing tests don't depend
// on tzdata being installed.
var ict = time.FixedZone("ICT", 7*3600)

func TestNextAutoCheckout(t *testing.T) {
	cases := []struct {
		name string
		now  time.Time
		hour int
		want time.Time
	}{
		{
			name: "before the hour same day",
			now:  time.Date(2026, 7, 21, 8, 30, 0, 0, ict),
			hour: 23,
			want: time.Date(2026, 7, 21, 23, 0, 0, 0, ict),
		},
		{
			name: "after the hour rolls to tomorrow",
			now:  time.Date(2026, 7, 21, 23, 30, 0, 0, ict),
			hour: 23,
			want: time.Date(2026, 7, 22, 23, 0, 0, 0, ict),
		},
		{
			name: "exactly at the hour rolls to tomorrow (strictly after)",
			now:  time.Date(2026, 7, 21, 23, 0, 0, 0, ict),
			hour: 23,
			want: time.Date(2026, 7, 22, 23, 0, 0, 0, ict),
		},
		{
			name: "month rollover",
			now:  time.Date(2026, 7, 31, 23, 30, 0, 0, ict),
			hour: 23,
			want: time.Date(2026, 8, 1, 23, 0, 0, 0, ict),
		},
		{
			name: "year rollover",
			now:  time.Date(2026, 12, 31, 23, 30, 0, 0, ict),
			hour: 23,
			want: time.Date(2027, 1, 1, 23, 0, 0, 0, ict),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := nextAutoCheckout(c.now, ict, c.hour)
			// Compare as absolute instants — the zone objects differ but the
			// moment must match, and the wall-clock time in ICT must be hour:00.
			assert.True(t, got.Equal(c.want), "want %s got %s", c.want, got)
			assert.Equal(t, c.hour, got.In(ict).Hour())
			assert.Equal(t, 0, got.In(ict).Minute())
			assert.True(t, got.After(c.now), "result must be strictly after now")
		})
	}
}

// The loop must fire when the timer elapses and then stop cleanly on cancel.
func TestRunScheduleLoop_FiresThenStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var fired int32
	fireCh := make(chan struct{}, 1)
	next := func(time.Time) time.Time { return time.Now().Add(40 * time.Millisecond) }
	fire := func(time.Time) {
		atomic.AddInt32(&fired, 1)
		select {
		case fireCh <- struct{}{}:
		default:
		}
	}

	done := make(chan struct{})
	go func() { runScheduleLoop(ctx, next, fire); close(done) }()

	select {
	case <-fireCh:
	case <-time.After(2 * time.Second):
		t.Fatal("loop never fired")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("loop did not stop after cancel")
	}
	assert.GreaterOrEqual(t, atomic.LoadInt32(&fired), int32(1))
}

// Cancelling before the timer elapses must return promptly without firing.
func TestRunScheduleLoop_CancelBeforeFire_DoesNotFire(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var fired int32
	next := func(time.Time) time.Time { return time.Now().Add(10 * time.Second) }
	fire := func(time.Time) { atomic.AddInt32(&fired, 1) }

	done := make(chan struct{})
	go func() { runScheduleLoop(ctx, next, fire); close(done) }()

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("loop did not stop promptly on cancel")
	}
	require.Equal(t, int32(0), atomic.LoadInt32(&fired), "must not fire before the timer elapses")
}
