package types

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

// SQLULID is a wrapper for ulid.ULID that handles
// SQL scanning and valuing correctly with Postgres's uuid type.
type SQLULID struct {
	ulid.ULID
}

// Value implements the driver.Valuer interface.
// This is called when WE SEND data TO the database (e.g., in an INSERT).
func (su SQLULID) Value() (driver.Value, error) {
	// We cast to uuid.UUID, which the 'pq' driver understands perfectly.
	idAsUUID := uuid.UUID(su.ULID)
	return idAsUUID.String(), nil
}

// Scan implements the sql.Scanner interface.
// This is called when WE RECEIVE data FROM the database (e.g., in a SELECT).
func (su *SQLULID) Scan(src interface{}) error {
	var id uuid.UUID

	switch src := src.(type) {
	case []byte:

		if len(src) == 16 {
			copy(id[:], src)
		} else if len(src) == 36 {
			var err error
			id, err = uuid.Parse(string(src))
			if err != nil {
				return fmt.Errorf("failed to parse 36-byte UUID string-bytes: %w", err)
			}
		} else {
			return fmt.Errorf("bad data size for SQLULID scan: %d bytes", len(src))
		}

	case string:
		var err error
		id, err = uuid.Parse(src)
		if err != nil {
			return fmt.Errorf("failed to parse UUID string: %w", err)
		}

	default:
		return fmt.Errorf("cannot scan %T into SQLULID", src)
	}

	// Cast the [16]byte uuid.UUID back to our [16]byte ulid.ULID.
	su.ULID = ulid.ULID(id)
	return nil
}
