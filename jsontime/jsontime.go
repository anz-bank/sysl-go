package jsontime

import (
	"encoding/json"
	"errors"
	"time"
)

// Time is an alias of time.Time.
//
// Can marshal to and from json and yaml.
type Time time.Time

// Time returns the equivalent time.Time struct.
func (t Time) Time() time.Time {
	return time.Time(t)
}

// Time returns the equivalent time.Time struct.
func (t Time) String() string {
	return time.Time(t).String()
}

// MarshalJSON implements json.Marshaller.
func (t Time) MarshalJSON() ([]byte, error) {
	return time.Time(t).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaller.
func (t *Time) UnmarshalJSON(data []byte) error {
	var tt time.Time
	if err := json.Unmarshal(data, &tt); err != nil {
		return err
	}
	*t = Time(tt)
	return nil
}

// MarshalYAML implements yaml.Marshaller.
func (t Time) MarshalYAML() (interface{}, error) {
	if time.Time(t).Year() < 0 || time.Time(t).Year() > 10000 {
		// As in Time.MarshalJSON
		return nil, errors.New("Time.MarshalYAML: year outside range [0,9999]")
	}
	return time.Time(t).Format(time.RFC3339Nano), nil
}

// UnmarshalYAML implements yaml.Unmarshaller.
func (t *Time) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	if s == "null" {
		return nil
	}
	tim, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return err
	}
	*t = Time(tim)
	return nil
}
