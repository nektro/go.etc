package dbt

import (
	"database/sql/driver"
	"errors"
	"strings"
)

// Array is a custom sql/driver type to handle list columns
type Array []string

// V returns the underlying object
func (t Array) V() []string {
	return []string(t)
}

// Scan implements the database/sql/driver Scanner interface
func (a *Array) Scan(value interface{}) error {
	if value == nil {
		*a = Array([]string{})
		return nil
	}
	if bv, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := bv.(string); ok {
			if len(v) == 0 {
				*a = Array([]string{})
			}
			if len(v) > 0 {
				*a = Array(strings.Split(v, ","))
			}
			return nil
		}
	}
	return errors.New("failed to scan Array")
}

// Value implements the database/sql/driver Valuer interface
func (a Array) Value() (driver.Value, error) {
	return strings.Join(a, ","), nil
}

// String implements the fmt Stringer interface
func (a Array) String() string {
	v, _ := a.Value()
	return v.(string)
}
