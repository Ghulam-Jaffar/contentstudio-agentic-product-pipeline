package mongo

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// MongoTime is a custom time type that can handle MongoDB timestamps
// stored as strings without timezone information (legacy data).
// It assumes UTC timezone when parsing strings without timezone.
type MongoTime struct {
	time.Time
}

// UnmarshalBSONValue implements the bson.ValueUnmarshaler interface
// to handle both BSON DateTime and string timestamp formats.
func (mt *MongoTime) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	switch t {
	case bsontype.DateTime:
		// Handle native BSON DateTime
		var dt time.Time
		err := bson.RawValue{Type: t, Value: data}.Unmarshal(&dt)
		if err != nil {
			return err
		}
		mt.Time = dt
		return nil

	case bsontype.String:
		// Handle string timestamps (legacy data)
		var s string
		err := bson.RawValue{Type: t, Value: data}.Unmarshal(&s)
		if err != nil {
			return err
		}

		// Try parsing fetcher MongoDB timestamp formats
		formats := []string{
			time.RFC3339,                    // "2006-01-02T15:04:05Z07:00"
			time.RFC3339Nano,                // "2006-01-02T15:04:05.999999999Z07:00"
			"2006-01-02T15:04:05",           // "2025-06-09T12:02:52" (assume UTC)
			"2006-01-02T15:04:05.999999",    // "2025-06-09T12:02:52.667850" (assume UTC)
			"2006-01-02T15:04:05.999999999", // nanosecond precision (assume UTC)
		}

		var parsed time.Time
		var parseErr error
		for _, format := range formats {
			parsed, parseErr = time.Parse(format, s)
			if parseErr == nil {
				// If no timezone info, assume UTC
				if parsed.Location() == time.UTC {
					mt.Time = parsed.UTC()
				} else {
					mt.Time = parsed
				}
				return nil
			}
		}

		return fmt.Errorf("MongoTime.UnmarshalBSONValue: unable to parse timestamp string '%s': %w", s, parseErr)

	case bsontype.Null:
		// Handle null values
		mt.Time = time.Time{}
		return nil

	default:
		return fmt.Errorf("MongoTime.UnmarshalBSONValue: cannot unmarshal %v into MongoTime", t)
	}
}

// MarshalBSONValue implements the bson.ValueMarshaler interface
// to serialize MongoTime as a BSON DateTime.
func (mt MongoTime) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(mt.Time)
}

// IsZero returns true if the time is zero (for omitempty support)
func (mt MongoTime) IsZero() bool {
	return mt.Time.IsZero()
}
