package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/spf13/cast"
)

// DefaultDateLayout specifies the default app date strings layout.
// We can edit this here depending on the locale of the system / user.
const DefaultDateLayout = "2006-01-02T15:04:05.000Z"

// NowDateTime returns new DateTime instance with the current local time.
func NowDateTime() DateTime {
	n := time.Now()
	return DateTime{t: &n}
}

// ParseDateTime creates a new DateTime from the provided value
// (could be [cast.ToTime] supported string, [time.Time], etc.).
func ParseDateTime(value any) (DateTime, error) {
	d := DateTime{}
	if err := d.Scan(value); err != nil {
		return d, err
	}
	return d, nil
}

// DateTime represents a [time.Time] instance in UTC that is wrapped
// and serialized using the app default date layout. It can be nil if
// the scanned value is zero or invalid.
type DateTime struct {
	t *time.Time
}

// Add adds the Duration to the DateTime and returns a new DateTime.
func (d *DateTime) Add(duration time.Duration) (DateTime, error) {
	t := d.Time()
	if t != nil {
		t2 := t.Add(duration)
		return DateTime{t: &t2}, nil
	}
	return DateTime{}, errors.New("cannot add duration to zero time")
}

// Time returns the internal [time.Time] instance.
func (d DateTime) Time() *time.Time {
	return d.t
}

// IsZero checks whether the current DateTime instance has zero time value.
func (d DateTime) IsZero() bool {
	return d.Time() == nil || d.Time().IsZero()
}

// String serializes the current DateTime instance into a formatted
// UTC date string.
//
// The zero value is serialized to an empty string.
func (d DateTime) String() string {
	t := d.Time()
	if t == nil || t.IsZero() {
		return ""
	}
	return t.UTC().Format(DefaultDateLayout)
}

func (d DateTime) Int() int64 {
	if d.t != nil {
		return d.t.UnixMicro()
	}
	return 0
}

// MarshalJSON implements the [json.Marshaler] interface.
func (d DateTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// UnmarshalJSON implements the [json.Unmarshaler] interface.
func (d *DateTime) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	return d.Scan(raw)
}

// Value implements the [driver.Valuer] interface.
func (d DateTime) Value() (driver.Value, error) {
	return d.Int(), nil
}

// Scan implements [sql.Scanner] interface to scan the provided value
// into the current DateTime instance.
func (d *DateTime) Scan(value any) error {
	switch v := value.(type) {
	case time.Time:
		d.t = &v
	case DateTime:
		d.t = v.Time()
	case string:
		if v == "" {
			d.t = nil
		} else {
			t, err := time.Parse(DefaultDateLayout, v)
			if err != nil {
				// check other common time formats
				t, err = cast.ToTimeE(v)
				if err != nil {
					d.t = nil
					break
				}
			}
			d.t = &t
		}
	case int64:
		if v == 0 {
			d.t = nil
		} else {
			time := time.UnixMicro(v)
			d.t = &time
		}
	case int, int32, uint, uint64, uint32:
		if v == 0 {
			d.t = nil
		} else {
			time := cast.ToTime(v)
			d.t = &time
		}
	default:
		d.t = nil
	}

	return nil
}
