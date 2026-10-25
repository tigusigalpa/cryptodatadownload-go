package cryptodatadownload

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// ParseDate parses an exact YYYY-MM-DD date in UTC for date-partitioned
// query fields. It rejects invalid, zero and out-of-range dates locally.
func ParseDate(value string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, &ArgumentError{Param: "date", Msg: "expected a valid YYYY-MM-DD date (UTC)"}
	}
	if err := requireDate("date", date); err != nil {
		return time.Time{}, err
	}
	return date, nil
}

// requireString fails when a required string parameter is empty.
func requireString(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ArgumentError{Param: name, Msg: "value is required"}
	}
	return nil
}

// requireDate validates a supplied date after normalizing it to UTC.
func requireDate(name string, value time.Time) error {
	if value.IsZero() {
		return &ArgumentError{Param: name, Msg: "date is required (UTC)"}
	}
	if year := value.UTC().Year(); year < 1 || year > 9999 {
		return &ArgumentError{Param: name, Msg: "date must have a UTC year between 0001 and 9999"}
	}
	return nil
}

func requireFinite(name string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return &ArgumentError{Param: name, Msg: "value must be finite"}
	}
	return nil
}

func checkPagination(name string, value, maximum int) error {
	if value < 0 {
		return &ArgumentError{Param: name, Msg: "value must not be negative"}
	}
	if maximum > 0 && value > maximum {
		return &ArgumentError{Param: name, Msg: fmt.Sprintf("value must not exceed %d", maximum)}
	}
	return nil
}

func validateRequestsPerSecond(rps float64) error {
	if math.IsNaN(rps) || math.IsInf(rps, 0) || rps < 0 {
		return fmt.Errorf("cryptodatadownload: requests per second must be finite and non-negative: %v", rps)
	}
	if rps > 0 {
		interval := float64(time.Second) / rps
		if interval < 1 || interval >= float64(math.MaxInt64) {
			return fmt.Errorf("cryptodatadownload: requests per second produces an unsupported interval: %v", rps)
		}
	}
	return nil
}

// dateString normalizes a date parameter to UTC YYYY-MM-DD.
func dateString(value time.Time) string {
	return value.UTC().Format("2006-01-02")
}

// checkEnum fails when an explicitly supplied value is not one of allowed.
// Optional parameters skip this check only when their pointer is nil.
func checkEnum(name, value string, allowed ...string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return &ArgumentError{Param: name, Msg: fmt.Sprintf("%q is not one of %s", value, strings.Join(allowed, ", "))}
}
