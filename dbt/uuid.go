package dbt

import (
	"database/sql/driver"
	"errors"
	"math/rand"
	"time"

	"github.com/nektro/go.etc/internal"
	"github.com/oklog/ulid"
)

// UUID is a custom sql/driver type to handle list columns
type UUID string

// NewUUID creates a new UUID
func NewUUID() UUID {
	t := time.Unix(0, time.Now().UnixNano()-internal.Epoch.UnixNano())
	var entropy = ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	return UUID(ulid.MustNew(ulid.Timestamp(t), entropy).String())
}

// IsUUID returns if id is a valid UUID, false if empty or invalid
func IsUUID(id UUID) bool {
	s := string(id)
	_, err := ulid.Parse(s)
	return err == nil
}

// Scan implements the database/sql/driver Scanner interface
func (t *UUID) Scan(value interface{}) error {
	if value == nil {
		*t = ""
		return nil
	}
	if bv, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := bv.(string); ok {
			if len(v) == 0 {
				*t = ""
			}
			if len(v) > 0 {
				*t = UUID(v)
			}
			return nil
		}
	}
	return errors.New("failed to scan UUID")
}

// V returns the underlying object
func (t UUID) V() ulid.ULID {
	v, _ := ulid.Parse(string(t))
	return v
}

// Value implements the database/sql/driver Valuer interface
func (t UUID) Value() (driver.Value, error) {
	return string(t), nil
}

// String implements the fmt Stringer interface
func (t UUID) String() string {
	return string(t)
}
