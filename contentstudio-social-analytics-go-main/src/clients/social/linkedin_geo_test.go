package social

import (
	"context"
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

func TestExtractGeoIDsFromJSON(t *testing.T) {
	cases := []struct {
		name        string
		input       []byte
		expectedLen int
	}{
		{
			name:        "empty input",
			input:       []byte{},
			expectedLen: 0,
		},
		{
			name:        "invalid json",
			input:       []byte(`not json`),
			expectedLen: 0,
		},
		{
			name:        "empty elements",
			input:       []byte(`{"elements":[]}`),
			expectedLen: 0,
		},
		{
			name: "with country geo IDs",
			input: []byte(`{
				"elements": [{
					"followerCountsByGeoCountry": [
						{"geo": "urn:li:geo:100"},
						{"geo": "urn:li:geo:101"}
					]
				}]
			}`),
			expectedLen: 2,
		},
		{
			name: "with city geo IDs",
			input: []byte(`{
				"elements": [{
					"followerCountsByGeo": [
						{"geo": "urn:li:geo:200"},
						{"geo": "urn:li:geo:201"},
						{"geo": "urn:li:geo:202"}
					]
				}]
			}`),
			expectedLen: 3,
		},
		{
			name: "with both country and city geo IDs",
			input: []byte(`{
				"elements": [{
					"followerCountsByGeoCountry": [
						{"geo": "urn:li:geo:100"}
					],
					"followerCountsByGeo": [
						{"geo": "urn:li:geo:200"}
					]
				}]
			}`),
			expectedLen: 2,
		},
		{
			name: "with duplicate geo IDs",
			input: []byte(`{
				"elements": [{
					"followerCountsByGeoCountry": [
						{"geo": "urn:li:geo:100"},
						{"geo": "urn:li:geo:100"}
					]
				}]
			}`),
			expectedLen: 1,
		},
		{
			name: "with missing colon in geo URN",
			input: []byte(`{
				"elements": [{
					"followerCountsByGeoCountry": [
						{"geo": "invalid-geo"}
					]
				}]
			}`),
			expectedLen: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ExtractGeoIDsFromJSON(tc.input)
			if len(result) != tc.expectedLen {
				t.Fatalf("expected %d geo IDs, got %d", tc.expectedLen, len(result))
			}
		})
	}
}

func TestExtractGeoIDsWithTypeFromJSON(t *testing.T) {
	cases := []struct {
		name         string
		input        []byte
		expectedLen  int
		checkType    bool
		expectedType string
	}{
		{
			name:        "empty input",
			input:       []byte{},
			expectedLen: 0,
		},
		{
			name:        "invalid json",
			input:       []byte(`not json`),
			expectedLen: 0,
		},
		{
			name: "country geo IDs have country type",
			input: []byte(`{
				"elements": [{
					"followerCountsByGeoCountry": [
						{"geo": "urn:li:geo:100"}
					]
				}]
			}`),
			expectedLen:  1,
			checkType:    true,
			expectedType: "country",
		},
		{
			name: "city geo IDs have city type",
			input: []byte(`{
				"elements": [{
					"followerCountsByGeo": [
						{"geo": "urn:li:geo:200"}
					]
				}]
			}`),
			expectedLen:  1,
			checkType:    true,
			expectedType: "city",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ExtractGeoIDsWithTypeFromJSON(tc.input)
			if len(result) != tc.expectedLen {
				t.Fatalf("expected %d geo IDs, got %d", tc.expectedLen, len(result))
			}
			if tc.checkType && len(result) > 0 {
				if result[0].Type != tc.expectedType {
					t.Fatalf("expected type '%s', got '%s'", tc.expectedType, result[0].Type)
				}
			}
		})
	}
}

func TestGeoIDWithType_Struct(t *testing.T) {
	geoID := GeoIDWithType{
		ID:   "100",
		Type: "country",
	}

	if geoID.ID != "100" {
		t.Fatalf("expected ID '100', got '%s'", geoID.ID)
	}
	if geoID.Type != "country" {
		t.Fatalf("expected Type 'country', got '%s'", geoID.Type)
	}
}

func TestNewGeoResolver(t *testing.T) {
	resolver := NewGeoResolver(nil, nil)

	if resolver == nil {
		t.Fatal("expected non-nil resolver")
	}
	if resolver.liClient != nil {
		t.Fatal("expected nil liClient")
	}
	if resolver.chClient != nil {
		t.Fatal("expected nil chClient")
	}
}

func TestGeoResolver_Struct(t *testing.T) {
	resolver := &GeoResolver{
		liClient: nil,
		chClient: nil,
	}

	if resolver.liClient != nil {
		t.Fatal("expected nil liClient")
	}
	if resolver.chClient != nil {
		t.Fatal("expected nil chClient")
	}
}

func TestExtractGeoIDsFromJSON_MultipleElements(t *testing.T) {
	input := []byte(`{
		"elements": [
			{
				"followerCountsByGeoCountry": [
					{"geo": "urn:li:geo:100"},
					{"geo": "urn:li:geo:101"}
				]
			},
			{
				"followerCountsByGeoCountry": [
					{"geo": "urn:li:geo:102"}
				]
			}
		]
	}`)

	result := ExtractGeoIDsFromJSON(input)
	if len(result) != 3 {
		t.Fatalf("expected 3 geo IDs, got %d", len(result))
	}
}

func TestExtractGeoIDsWithTypeFromJSON_Mixed(t *testing.T) {
	input := []byte(`{
		"elements": [{
			"followerCountsByGeoCountry": [
				{"geo": "urn:li:geo:100"}
			],
			"followerCountsByGeo": [
				{"geo": "urn:li:geo:200"}
			]
		}]
	}`)

	result := ExtractGeoIDsWithTypeFromJSON(input)
	if len(result) != 2 {
		t.Fatalf("expected 2 geo IDs, got %d", len(result))
	}

	typeMap := make(map[string]string)
	for _, g := range result {
		typeMap[g.ID] = g.Type
	}

	if typeMap["100"] != "country" {
		t.Fatal("expected ID 100 to be country type")
	}
	if typeMap["200"] != "city" {
		t.Fatal("expected ID 200 to be city type")
	}
}

// MockClickHouseClient implements the minimal interface needed for GeoResolver tests
type MockClickHouseClient struct {
	GetGeoMappingsFunc            func(ctx context.Context, geoIDs []string) (map[string]string, error)
	InsertGeoMappingsWithTypeFunc func(ctx context.Context, mappings []MockGeoMappingWithType) error
}

type MockGeoMappingWithType struct {
	ID   string
	Name string
	Type string
}

// MockLinkedInClientForGeo is a minimal mock for geo resolution
type MockLinkedInClientForGeo struct {
	ResolveGeoIDsFunc func(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error)
}

func (m *MockLinkedInClientForGeo) ResolveGeoIDs(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error) {
	if m.ResolveGeoIDsFunc != nil {
		return m.ResolveGeoIDsFunc(ctx, geoIDs, accessToken)
	}
	return map[string]string{}, nil
}

func TestGeoResolver_ResolveGeoIDsWithType_EmptyInput(t *testing.T) {
	resolver := NewGeoResolver(nil, nil)

	result, err := resolver.ResolveGeoIDsWithType(context.Background(), []GeoIDWithType{}, "token")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 results for empty input, got %d", len(result))
	}
}

func TestGeoResolver_ResolveGeoIDs_EmptyInput(t *testing.T) {
	resolver := NewGeoResolver(nil, nil)

	result, err := resolver.ResolveGeoIDs(context.Background(), []string{}, "token")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 results for empty input, got %d", len(result))
	}
}

func TestGeoResolver_ResolveGeoIDsFromFollowerData_EmptyGeoIDs(t *testing.T) {
	resolver := NewGeoResolver(nil, nil)

	// JSON with no geo IDs
	followerData := []byte(`{"elements": []}`)

	result, err := resolver.ResolveGeoIDsFromFollowerData(context.Background(), followerData, "token")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 results for empty geo IDs, got %d", len(result))
	}
}

func TestGeoResolver_ResolveGeoIDsFromFollowerData_InvalidJSON(t *testing.T) {
	resolver := NewGeoResolver(nil, nil)

	// Invalid JSON
	followerData := []byte(`{invalid json}`)

	result, err := resolver.ResolveGeoIDsFromFollowerData(context.Background(), followerData, "token")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return empty result for invalid JSON (ExtractGeoIDsFromJSON returns nil)
	if len(result) != 0 {
		t.Fatalf("expected 0 results for invalid JSON, got %d", len(result))
	}
}

// ==================== Logging Contract Tests ====================

// TestLoggingContract_GeoResolver_NoCaptureException verifies that the GeoResolver
// does not call CaptureException for any error paths. It logs warnings and returns
// errors to callers.
func TestLoggingContract_GeoResolver_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	resolver := NewGeoResolver(nil, nil)

	// Error path: empty input (early return, no errors)
	_, _ = resolver.ResolveGeoIDs(context.Background(), []string{}, "token")

	// Error path: invalid follower data JSON
	_, _ = resolver.ResolveGeoIDsFromFollowerData(context.Background(), []byte(`invalid`), "token")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls from GeoResolver, got %d", len(*captureRecords))
	}
}
