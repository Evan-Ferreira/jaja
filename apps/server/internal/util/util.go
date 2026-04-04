package util

import (
	"time"
)

// ParseISODate parses an ISO 8601 / RFC3339 date string into a *time.Time.
// Returns nil if s is nil or cannot be parsed.
func ParseISODate(s *string) *time.Time {
	if s == nil {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil
	}
	return &t
}
