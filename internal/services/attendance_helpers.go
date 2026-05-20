package services

import (
	"math"
	"time"
)

// haversineMeters returns the great-circle distance in meters between two
// (lat, lon) points on Earth. Used for GPS-based check-in proximity when
// OFFICE_GPS_ENABLED=true.
func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusMeters = 6371000.0
	rad := math.Pi / 180.0
	phi1, phi2 := lat1*rad, lat2*rad
	dPhi := (lat2 - lat1) * rad
	dLam := (lon2 - lon1) * rad
	a := math.Sin(dPhi/2)*math.Sin(dPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(dLam/2)*math.Sin(dLam/2)
	return earthRadiusMeters * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// loadTZ resolves a tzdata zone name. Falls back to UTC on parse failure
// — a misconfigured COMPANY_TIMEZONE shouldn't take down the server.
func loadTZ(name string) *time.Location {
	if name == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

// todayInTZ returns (now-local, today-midnight-local) in the given zone.
func todayInTZ(loc *time.Location) (time.Time, time.Time) {
	now := time.Now().In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	return now, today
}

// isWorkday returns true for Mon-Fri. Saturday/Sunday are weekend.
func isWorkday(d time.Time) bool {
	wd := d.Weekday()
	return wd != time.Saturday && wd != time.Sunday
}

// hoursBetween returns hours between in and out, rounded to two decimals.
// Returns nil when the session is still open (out == nil).
func hoursBetween(in time.Time, out *time.Time) *float64 {
	if out == nil {
		return nil
	}
	h := math.Round(out.Sub(in).Hours()*100) / 100
	return &h
}

// thresholdAt builds an hh:mm timestamp on the same calendar day as ref,
// preserving ref's Location. Used to compare a check-in time against the
// configurable late threshold without losing the timezone.
func thresholdAt(ref time.Time, hour, minute int) time.Time {
	return time.Date(ref.Year(), ref.Month(), ref.Day(), hour, minute, 0, 0, ref.Location())
}

// parseDateYMD parses a YYYY-MM-DD string in the given location. Used for
// the admin Create's date field and the matrix/list date filters.
func parseDateYMD(s string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, loc)
}

// truncateToDateInTZ strips the time component, pinning to midnight in the
// given zone. Postgres DATE columns are date-only — passing a non-midnight
// timestamp can introduce timezone-rounding bugs.
func truncateToDateInTZ(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}
