package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

func TestMongoTime_UnmarshalBSONValue_DateTime(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	
	data, err := bson.Marshal(bson.M{"time": now})
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result struct {
		Time MongoTime `bson:"time"`
	}
	err = bson.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !result.Time.Time.Equal(now) {
		t.Fatalf("expected %v, got %v", now, result.Time.Time)
	}
}

func TestMongoTime_UnmarshalBSONValue_String_RFC3339(t *testing.T) {
	timeStr := "2025-06-09T12:02:52Z"
	expected, _ := time.Parse(time.RFC3339, timeStr)

	raw := bson.RawValue{
		Type:  bsontype.String,
		Value: mustMarshalString(timeStr),
	}

	var mt MongoTime
	err := mt.UnmarshalBSONValue(raw.Type, raw.Value)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !mt.Time.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, mt.Time)
	}
}

func TestMongoTime_UnmarshalBSONValue_String_NoTimezone(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			name:     "basic timestamp without timezone",
			input:    "2025-06-09T12:02:52",
			expected: time.Date(2025, 6, 9, 12, 2, 52, 0, time.UTC),
		},
		{
			name:     "timestamp with microseconds",
			input:    "2025-06-09T12:02:52.667850",
			expected: time.Date(2025, 6, 9, 12, 2, 52, 667850000, time.UTC),
		},
		{
			name:     "RFC3339 format",
			input:    "2025-06-09T12:02:52+00:00",
			expected: time.Date(2025, 6, 9, 12, 2, 52, 0, time.UTC),
		},
		{
			name:     "RFC3339Nano format",
			input:    "2025-06-09T12:02:52.123456789Z",
			expected: time.Date(2025, 6, 9, 12, 2, 52, 123456789, time.UTC),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			raw := bson.RawValue{
				Type:  bsontype.String,
				Value: mustMarshalString(tc.input),
			}

			var mt MongoTime
			err := mt.UnmarshalBSONValue(raw.Type, raw.Value)
			if err != nil {
				t.Fatalf("failed to unmarshal %q: %v", tc.input, err)
			}

			if !mt.Time.Equal(tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, mt.Time)
			}
		})
	}
}

func TestMongoTime_UnmarshalBSONValue_Null(t *testing.T) {
	raw := bson.RawValue{
		Type:  bsontype.Null,
		Value: nil,
	}

	var mt MongoTime
	err := mt.UnmarshalBSONValue(raw.Type, raw.Value)
	if err != nil {
		t.Fatalf("failed to unmarshal null: %v", err)
	}

	if !mt.Time.IsZero() {
		t.Fatalf("expected zero time for null, got %v", mt.Time)
	}
}

func TestMongoTime_UnmarshalBSONValue_InvalidType(t *testing.T) {
	raw := bson.RawValue{
		Type:  bsontype.Int32,
		Value: []byte{1, 0, 0, 0},
	}

	var mt MongoTime
	err := mt.UnmarshalBSONValue(raw.Type, raw.Value)
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestMongoTime_UnmarshalBSONValue_InvalidString(t *testing.T) {
	raw := bson.RawValue{
		Type:  bsontype.String,
		Value: mustMarshalString("not-a-valid-timestamp"),
	}

	var mt MongoTime
	err := mt.UnmarshalBSONValue(raw.Type, raw.Value)
	if err == nil {
		t.Fatal("expected error for invalid timestamp string")
	}
}

func TestMongoTime_MarshalBSONValue(t *testing.T) {
	now := time.Now().UTC()
	mt := MongoTime{Time: now}

	btype, data, err := mt.MarshalBSONValue()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if btype != bsontype.DateTime {
		t.Fatalf("expected DateTime type, got %v", btype)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty data")
	}
}

func TestMongoTime_IsZero(t *testing.T) {
	cases := []struct {
		name     string
		mt       MongoTime
		expected bool
	}{
		{
			name:     "zero time",
			mt:       MongoTime{},
			expected: true,
		},
		{
			name:     "non-zero time",
			mt:       MongoTime{Time: time.Now()},
			expected: false,
		},
		{
			name:     "explicit zero",
			mt:       MongoTime{Time: time.Time{}},
			expected: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := tc.mt.IsZero()
			if result != tc.expected {
				t.Fatalf("expected IsZero() = %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestMongoTime_RoundTrip(t *testing.T) {
	original := MongoTime{Time: time.Date(2025, 6, 15, 10, 30, 45, 0, time.UTC)}

	data, err := bson.Marshal(bson.M{"time": original})
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result struct {
		Time MongoTime `bson:"time"`
	}
	err = bson.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !result.Time.Time.Equal(original.Time) {
		t.Fatalf("round trip failed: expected %v, got %v", original.Time, result.Time.Time)
	}
}

func mustMarshalString(s string) []byte {
	strLen := int32(len(s) + 1)
	result := make([]byte, 4+len(s)+1)
	result[0] = byte(strLen)
	result[1] = byte(strLen >> 8)
	result[2] = byte(strLen >> 16)
	result[3] = byte(strLen >> 24)
	copy(result[4:], s)
	result[len(result)-1] = 0
	return result
}
