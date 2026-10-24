package cryptodatadownload

import (
	"fmt"
	"strings"
	"time"
)

// requireString fails when a required string parameter is empty.
func requireString(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ArgumentError{Param: name, Msg: "value is required"}
	}
	return nil
}

// requireDate fails when a required date parameter is the zero time.
func requireDate(name string, value time.Time) error {
	if value.IsZero() {
		return &ArgumentError{Param: name, Msg: "date is required (UTC)"}
	}
	return nil
}

// dateString normalizes a date parameter to UTC YYYY-MM-DD.
func dateString(value time.Time) string {
	return value.UTC().Format("2006-01-02")
}

// checkEnum fails when value is non-empty and not one of allowed. Optional
// params pass "" / nil through unvalidated; the server applies its default.
func checkEnum(name, value string, allowed ...string) error {
	if value == "" {
		return nil
	}
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return &ArgumentError{Param: name, Msg: fmt.Sprintf("%q is not one of %s", value, strings.Join(allowed, ", "))}
}
