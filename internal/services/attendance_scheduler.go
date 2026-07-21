package services

import (
	"context"
	"log"
	"time"
)

// nextAutoCheckout returns the next occurrence of hour:00:00 in loc that is
// strictly after now. When now is exactly hour:00 it returns tomorrow's run,
// so a fire that lands on time never immediately re-fires.
func nextAutoCheckout(now time.Time, loc *time.Location, hour int) time.Time {
	n := now.In(loc)
	run := time.Date(n.Year(), n.Month(), n.Day(), hour, 0, 0, 0, loc)
	if !run.After(n) {
		run = run.AddDate(0, 0, 1)
	}
	return run
}

// StartAutoCheckoutScheduler runs AutoCheckOut once a day at hour:00 in the
// company timezone, until ctx is cancelled. main.go launches it in a goroutine
// before serving.
//
// The process has no graceful-shutdown path today, so in practice ctx is
// process-lifetime and the goroutine dies with the process. The ctx plumbing
// exists so a shutdown hook can stop it cleanly if one is added later.
//
// Missed runs self-heal: if the server is down at hour:00, that day is skipped,
// but the next run's cutoff still closes every session left open before it —
// AutoCheckOut is OpenSessionsBefore(cutoff), not just "today".
func StartAutoCheckoutScheduler(ctx context.Context, svc *AttendanceService, tzName string, hour int) {
	if hour < 0 || hour > 23 {
		log.Printf("auto-checkout: invalid hour %d, falling back to 23", hour)
		hour = 23
	}
	loc := loadTZ(tzName)

	log.Printf("auto-checkout: scheduler started, first run at %s",
		nextAutoCheckout(time.Now(), loc, hour).Format(time.RFC3339))

	next := func(now time.Time) time.Time { return nextAutoCheckout(now, loc, hour) }
	fire := func(cutoff time.Time) {
		closed, err := svc.AutoCheckOut(ctx, cutoff)
		if err != nil {
			log.Printf("auto-checkout: run at %s failed: %v", cutoff.Format(time.RFC3339), err)
			return
		}
		log.Printf("auto-checkout: closed %d open session(s) at %s", closed, cutoff.Format(time.RFC3339))
	}
	runScheduleLoop(ctx, next, fire)
}

// runScheduleLoop sleeps until next(now), fires with that instant, and repeats
// until ctx is cancelled. Split out from StartAutoCheckoutScheduler so the
// timing math and loop mechanics are testable without waiting for the wall
// clock to reach 23:00.
func runScheduleLoop(ctx context.Context, next func(time.Time) time.Time, fire func(time.Time)) {
	for {
		now := time.Now()
		at := next(now)
		timer := time.NewTimer(at.Sub(now))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			fire(at)
		}
	}
}
