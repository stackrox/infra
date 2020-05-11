package config

import (
	"encoding/json"
	"time"
)

var _ json.Marshaler = (*JSONDuration)(nil)
var _ json.Unmarshaler = (*JSONDuration)(nil)

// JSONDuration represents a time.Duration that is able to be (JSON) marshaled
// into or unmarshaled from the more human-friendly duration format.
type JSONDuration time.Duration

// Duration returns the internal time.Duration value.
func (d JSONDuration) Duration() time.Duration {
	return time.Duration(d)
}

// String returns a string representing the duration in the form "72h3m0.5s".
func (d JSONDuration) String() string {
	return d.Duration().String()
}

// MarshalJSON implements json.Marshaler.
func (d JSONDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *JSONDuration) UnmarshalJSON(b []byte) error {
	var i int64
	if err := json.Unmarshal(b, &i); err == nil {
		*d = JSONDuration(i)
		return nil
	}

	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = JSONDuration(duration)
	return nil
}
