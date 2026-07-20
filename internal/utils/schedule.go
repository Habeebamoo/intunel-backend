package utils

import (
    "fmt"
    "time"
)

func ParseScheduledAt(date, t, timezone string) (time.Time, error) {
	if timezone == "" {
		timezone = "UTC"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone: %s — use IANA format e.g. Africa/Lagos, Europe/London, America/New_York", timezone)
	}

	combined := fmt.Sprintf("%s %s", date, t)
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", combined, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date or time format, expected YYYY-MM-DD and HH:MM:SS")
	}

	utc := parsed.UTC()

	if utc.Before(time.Now().UTC()) {
		return time.Time{}, fmt.Errorf("scheduled time must be in the future")
	}

	return utc, nil
}