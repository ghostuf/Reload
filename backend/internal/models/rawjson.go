package models

import (
	"database/sql/driver"
	"fmt"
)

// RawJSON is a jsonb column stored as exactly the bytes it was given,
// rather than unmarshalled into a Go struct -- the point of keeping a
// provider's raw response around is to have the whole thing on hand for
// support and debugging, not just the few fields the rest of the app reads
// today.
type RawJSON []byte

// Value implements driver.Valuer, so GORM writes this straight into a
// jsonb column.
func (j RawJSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

// Scan implements sql.Scanner, accepting whatever the postgres driver hands
// back for a jsonb column.
func (j *RawJSON) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append(RawJSON(nil), v...)
	case string:
		*j = RawJSON(v)
	default:
		return fmt.Errorf("unsupported type %T for RawJSON", value)
	}
	return nil
}

// MarshalJSON embeds the stored bytes directly as a JSON value rather than
// base64-encoding them the way json.Marshal would for a bare []byte, so an
// API response carries this as a real nested object, not an opaque string.
func (j RawJSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON stores the raw bytes of whatever JSON value it is given.
func (j *RawJSON) UnmarshalJSON(data []byte) error {
	*j = append(RawJSON(nil), data...)
	return nil
}
