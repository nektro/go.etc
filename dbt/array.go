package dbt

import (
	"database/sql/driver"
	"errors"
	"strings"
)

// Array is a custom sql/driver type to handle list columns
type Array []string

// Scan implements the database/sql/driver Scanner interface
func (t *Array) Scan(value interface{}) error {
	if value == nil {
		*t = Array([]string{})
		return nil
	}
	if bv, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := bv.(string); ok {
			if len(v) == 0 {
				*t = Array([]string{})
			}
			if len(v) > 0 {
				*t = Array(strings.Split(v, "\n"))
			}
			return nil
		}
	}
	return errors.New("failed to scan Array")
}

// V returns the underlying object
func (t Array) V() []string {
	return []string(t)
}

// Value implements the database/sql/driver Valuer interface
func (t Array) Value() (driver.Value, error) {
	return strings.Join(t, "\n"), nil
}

// String implements the fmt Stringer interface
func (t Array) String() string {
	v, _ := t.Value()
	return v.(string)
}
