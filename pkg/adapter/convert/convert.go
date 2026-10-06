// Package convert holds the proto→domain conversion helpers shared by the
// gRPC client adapters under pkg/adapter/<domain>.
package convert

import "time"

// TimePtr parses a timestamp string into *time.Time. It accepts the
// "2006-01-02 15:04:05" layout first, then falls back to RFC3339, and returns
// nil when the input is empty or cannot be parsed.
func TimePtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return nil
		}
	}
	return &t
}

// NullableString returns nil for an empty string, otherwise a pointer to it.
func NullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// NullableInt32 returns a pointer to i.
func NullableInt32(i int32) *int32 {
	return &i
}
