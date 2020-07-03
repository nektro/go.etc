package dbt

import (
	"database/sql/driver"
	"errors"
	"strconv"
	"strings"
	"time"
)

// Duration is a custom sql/driver type to handle 3-state permissions
type Duration [2]int

// duration units
const (
	DurS = iota // 0: second
	DurM        // 1: minute
	DurH        // 2: hour
	DurD        // 3: day
	DurW        // 4: week
	DurO        // 5: month
	DurY        // 6: year
)

var (
	durationUnits  = map[int]time.Duration{}
	durationUnitsS = [...]string{"s", "m", "h", "d", "w", "o", "y"}
)

func init() {
	durationUnits[0] = time.Second
	durationUnits[1] = time.Minute
	durationUnits[2] = time.Hour
	durationUnits[3] = durationUnits[2] * 24
	durationUnits[4] = durationUnits[3] * 7
	durationUnits[5] = durationUnits[4] * 4
	durationUnits[6] = durationUnits[5] * 12
}

// DurationZero is the zero value for Duration
var DurationZero = Duration([...]int{0, DurS})

// Scan - Implement the database/sql Scanner interface
func (t *Duration) Scan(value interface{}) error {
	if value == nil {
		*t = DurationZero
		return nil
	}
	if bv, err := driver.String.ConvertValue(value); err == nil {
		if v, ok := bv.(string); ok {
			vv := strings.SplitN(v, ":", 2)
			if len(vv) == 2 {
				a, _ := strconv.Atoi(vv[0])
				b, _ := strconv.Atoi(vv[1])
				*t = [...]int{a, b}
				return nil
			}
			*t = DurationZero
			return nil
		}
	}
	return errors.New("failed to scan Duration")
}

// V returns the underlying object
func (t Duration) V() time.Duration {
	return time.Duration(t[0]) * durationUnits[t[1]]
}

// Value - Implement the database/sql Valuer interface
func (t Duration) Value() (driver.Value, error) {
	return strconv.Itoa(t[0]) + ":" + strconv.Itoa(t[1]), nil
}

// String implements the fmt Stringer interface
func (t Duration) String() string {
	return t.V().String()
}
