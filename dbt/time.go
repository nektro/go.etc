package dbt

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// TimeFormat is the underlying format of Time.String()[:19]
const TimeFormat = "2006-01-02 15:04:05"

// TimeZero is empty time
const TimeZero = "0000-01-01 00:00:00"

// Time is a custom sql/driver type to handle UTC times
type Time time.Time

// NewTime accepts a string in TimeFormat and returns a Time. TimeZero on error.
func NewTime(s string) Time {
	r, err := time.Parse(TimeFormat, s)
	if err != nil {
		return NewTime(TimeZero)
	}
	return Time(r)
}

// Scan implements the database/sql/driver Scanner interface
func (t *Time) Scan(value interface{}) error {
	if value == nil {
		// *t = Time([]string{})
		return nil
	}
	if bv, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := bv.(string); ok {
			if len(v) == 0 {
				v = TimeZero
			}
			*t = NewTime(v)
			return nil
		}
	}
	return errors.New("failed to scan Time")
}

// V returns the underlying object
func (t Time) V() time.Time {
	return time.Time(t)
}

// Value implements the database/sql/driver Valuer interface
func (t Time) Value() (driver.Value, error) {
	s := t.V().String()[:19]
	if s == TimeZero {
		return "", nil
	}
	return s, nil
}

// String implements the fmt Stringer interface
func (t Time) String() string {
	v, _ := t.Value()
	s := v.(string)
	return strings.Replace(s, " ", "T", 1) + "Z"
}

// MarshalJSON implements the json Marshal interface
func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
