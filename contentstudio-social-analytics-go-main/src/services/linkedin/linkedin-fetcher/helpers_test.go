package main

import (
	"errors"
	"testing"
	"time"
)

func TestIsTokenError(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"401 error", errors.New("got 401 Unauthorized"), true},
		{"403 error", errors.New("403 Forbidden"), true},
		{"unauthorized", errors.New("request unauthorized"), true},
		{"expired", errors.New("token expired"), true},
		{"invalid_token", errors.New("invalid_token"), true},
		{"invalid access token", errors.New("invalid access token"), true},
		{"access denied", errors.New("access denied"), true},
		{"permission error", errors.New("permission denied"), true},
		{"not authorized", errors.New("not authorized to access"), true},
		{"generic error", errors.New("network timeout"), false},
		{"500 error", errors.New("500 Internal Server Error"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := isTokenError(tc.err)
			if result != tc.expected {
				t.Fatalf("isTokenError(%v) = %v, expected %v", tc.err, result, tc.expected)
			}
		})
	}
}

func TestWrapTokenError(t *testing.T) {
	cases := []struct {
		name          string
		err           error
		expectNil     bool
		expectInvalid bool
		expectDenied  bool
	}{
		{"nil error", nil, true, false, false},
		{"401 error", errors.New("401 Unauthorized"), false, true, false},
		{"403 error", errors.New("403 Forbidden"), false, false, true},
		{"permission error", errors.New("permission denied"), false, false, true},
		{"generic error", errors.New("network timeout"), false, false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := wrapTokenError(tc.err)

			if tc.expectNil {
				if result != nil {
					t.Fatalf("expected nil, got %v", result)
				}
				return
			}

			if tc.expectInvalid {
				if !errors.Is(result, ErrTokenInvalid) {
					t.Fatalf("expected ErrTokenInvalid, got %v", result)
				}
			} else if tc.expectDenied {
				if !errors.Is(result, ErrTokenPermissionDenied) {
					t.Fatalf("expected ErrTokenPermissionDenied, got %v", result)
				}
			}
		})
	}
}

func TestChunk(t *testing.T) {
	cases := []struct {
		name     string
		input    []string
		size     int
		expected [][]string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			size:     3,
			expected: nil,
		},
		{
			name:     "nil slice",
			input:    nil,
			size:     3,
			expected: nil,
		},
		{
			name:     "size zero",
			input:    []string{"a", "b", "c"},
			size:     0,
			expected: nil,
		},
		{
			name:     "negative size",
			input:    []string{"a", "b", "c"},
			size:     -1,
			expected: nil,
		},
		{
			name:     "exact chunks",
			input:    []string{"a", "b", "c", "d", "e", "f"},
			size:     2,
			expected: [][]string{{"a", "b"}, {"c", "d"}, {"e", "f"}},
		},
		{
			name:     "partial last chunk",
			input:    []string{"a", "b", "c", "d", "e"},
			size:     2,
			expected: [][]string{{"a", "b"}, {"c", "d"}, {"e"}},
		},
		{
			name:     "size larger than input",
			input:    []string{"a", "b"},
			size:     5,
			expected: [][]string{{"a", "b"}},
		},
		{
			name:     "single element",
			input:    []string{"a"},
			size:     1,
			expected: [][]string{{"a"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := chunk(tc.input, tc.size)

			if tc.expected == nil {
				if result != nil {
					t.Fatalf("expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tc.expected) {
				t.Fatalf("expected %d chunks, got %d", len(tc.expected), len(result))
			}

			for i, chunk := range result {
				if len(chunk) != len(tc.expected[i]) {
					t.Fatalf("chunk %d: expected length %d, got %d", i, len(tc.expected[i]), len(chunk))
				}
				for j, v := range chunk {
					if v != tc.expected[i][j] {
						t.Fatalf("chunk[%d][%d]: expected %s, got %s", i, j, tc.expected[i][j], v)
					}
				}
			}
		})
	}
}

func TestChunk_Integers(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7}
	result := chunk(input, 3)

	expected := [][]int{{1, 2, 3}, {4, 5, 6}, {7}}
	if len(result) != len(expected) {
		t.Fatalf("expected %d chunks, got %d", len(expected), len(result))
	}

	for i, c := range result {
		if len(c) != len(expected[i]) {
			t.Fatalf("chunk %d: expected length %d, got %d", i, len(expected[i]), len(c))
		}
	}
}

func TestMapKeysToSlice(t *testing.T) {
	cases := []struct {
		name        string
		input       map[string]struct{}
		expectedLen int
	}{
		{
			name:        "empty map",
			input:       map[string]struct{}{},
			expectedLen: 0,
		},
		{
			name:        "nil map",
			input:       nil,
			expectedLen: 0,
		},
		{
			name: "single element",
			input: map[string]struct{}{
				"key1": {},
			},
			expectedLen: 1,
		},
		{
			name: "multiple elements",
			input: map[string]struct{}{
				"key1": {},
				"key2": {},
				"key3": {},
			},
			expectedLen: 3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapKeysToSlice(tc.input)

			if len(result) != tc.expectedLen {
				t.Fatalf("expected length %d, got %d", tc.expectedLen, len(result))
			}

			// Verify all keys are present
			for k := range tc.input {
				found := false
				for _, v := range result {
					if v == k {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("key %s not found in result", k)
				}
			}
		})
	}
}

func TestDecryptToken(t *testing.T) {
	cases := []struct {
		name          string
		token         string
		decryptionKey string
		expectEmpty   bool
	}{
		{
			name:          "empty token",
			token:         "",
			decryptionKey: "key",
			expectEmpty:   true,
		},
		{
			name:          "plain token returned as-is on decrypt failure",
			token:         "plain_token",
			decryptionKey: "invalid_key",
			expectEmpty:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := decryptToken(tc.token, tc.decryptionKey)

			if tc.expectEmpty {
				if result != "" {
					t.Fatalf("expected empty string, got %s", result)
				}
			} else {
				if result == "" {
					t.Fatal("expected non-empty string")
				}
			}
		})
	}
}

func TestCalculateDateRanges(t *testing.T) {
	cases := []struct {
		name     string
		syncType string
	}{
		{"incremental sync", "incremental"},
		{"full sync", "full"},
		{"full sync uppercase", "FULL"},
		{"empty sync type", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cutoff, start, end := calculateDateRanges(tc.syncType)

			// Basic sanity checks
			if cutoff.IsZero() {
				t.Fatal("cutoff time should not be zero")
			}
			if start.IsZero() {
				t.Fatal("start time should not be zero")
			}
			if end.IsZero() {
				t.Fatal("end time should not be zero")
			}

			// Start should be before end
			if !start.Before(end) {
				t.Fatalf("start (%v) should be before end (%v)", start, end)
			}

			// Cutoff should be around start time
			if cutoff.After(end) {
				t.Fatalf("cutoff (%v) should not be after end (%v)", cutoff, end)
			}

			// For incremental, date range should be ~10 days
			if tc.syncType == "incremental" {
				diff := end.Sub(start)
				if diff < 9*24*time.Hour || diff > 12*24*time.Hour {
					t.Fatalf("incremental sync should have ~10 day range, got %v", diff)
				}
			}
		})
	}
}

func TestParseStatsBatch(t *testing.T) {
	cases := []struct {
		name        string
		input       []byte
		expectedLen int
	}{
		{
			name:        "empty response",
			input:       []byte(`{"elements":[]}`),
			expectedLen: 0,
		},
		{
			name:        "invalid json",
			input:       []byte(`not json`),
			expectedLen: 0,
		},
		{
			name: "ugcPost elements",
			input: []byte(`{
				"elements": [
					{"ugcPost": "urn:li:ugcPost:123", "totalShareStatistics": {"likeCount": 10}},
					{"ugcPost": "urn:li:ugcPost:456", "totalShareStatistics": {"likeCount": 20}}
				]
			}`),
			expectedLen: 2,
		},
		{
			name: "share elements",
			input: []byte(`{
				"elements": [
					{"share": "urn:li:share:789", "totalShareStatistics": {"likeCount": 30}}
				]
			}`),
			expectedLen: 1,
		},
		{
			name: "mixed elements",
			input: []byte(`{
				"elements": [
					{"ugcPost": "urn:li:ugcPost:123", "totalShareStatistics": {"likeCount": 10}},
					{"share": "urn:li:share:456", "totalShareStatistics": {"likeCount": 20}}
				]
			}`),
			expectedLen: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseStatsBatch(tc.input)

			if len(result) != tc.expectedLen {
				t.Fatalf("expected %d elements, got %d", tc.expectedLen, len(result))
			}
		})
	}
}

func TestParseAssetBatch(t *testing.T) {
	cases := []struct {
		name        string
		input       []byte
		expectedLen int
	}{
		{
			name:        "empty response",
			input:       []byte(`{"results":{}}`),
			expectedLen: 0,
		},
		{
			name:        "invalid json",
			input:       []byte(`not json`),
			expectedLen: 0,
		},
		{
			name: "with id field",
			input: []byte(`{
				"results": {
					"key1": {"id": "asset123", "type": "image"},
					"key2": {"id": "asset456", "type": "video"}
				}
			}`),
			expectedLen: 2,
		},
		{
			name: "with asset field",
			input: []byte(`{
				"results": {
					"key1": {"asset": "asset789", "type": "document"}
				}
			}`),
			expectedLen: 1,
		},
		{
			name: "missing id",
			input: []byte(`{
				"results": {
					"key1": {"type": "image"}
				}
			}`),
			expectedLen: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseAssetBatch(tc.input)

			if len(result) != tc.expectedLen {
				t.Fatalf("expected %d elements, got %d", tc.expectedLen, len(result))
			}
		})
	}
}

func TestSemForAccount(t *testing.T) {
	// Test that same ID returns same semaphore
	sem1 := semForAccount("account123")
	sem2 := semForAccount("account123")

	if sem1 != sem2 {
		t.Fatal("expected same semaphore for same account ID")
	}

	// Test that different IDs return different semaphores
	sem3 := semForAccount("account456")
	if sem1 == sem3 {
		t.Fatal("expected different semaphores for different account IDs")
	}
}

func TestErrTokenInvalid(t *testing.T) {
	if ErrTokenInvalid == nil {
		t.Fatal("ErrTokenInvalid should not be nil")
	}
	if ErrTokenInvalid.Error() != "linkedin token invalid or expired" {
		t.Fatalf("unexpected error message: %s", ErrTokenInvalid.Error())
	}
}

func TestErrTokenPermissionDenied(t *testing.T) {
	if ErrTokenPermissionDenied == nil {
		t.Fatal("ErrTokenPermissionDenied should not be nil")
	}
	if ErrTokenPermissionDenied.Error() != "linkedin token permission denied" {
		t.Fatalf("unexpected error message: %s", ErrTokenPermissionDenied.Error())
	}
}

// ================== isExpectedError Tests ==================

func TestIsExpectedError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"EXPIRED_ACCESS_TOKEN", &testError{msg: "EXPIRED_ACCESS_TOKEN"}, true},
		{"INVALID_POST_FINDER_AUTHOR_ENTITY_TYPE", &testError{msg: "INVALID_POST_FINDER_AUTHOR_ENTITY_TYPE"}, true},
		{"status 401", &testError{msg: "linkedin api error (status 401): unauthorized"}, true},
		{"status 403", &testError{msg: "linkedin api error (status 403): forbidden"}, true},
		{"token expired lowercase", &testError{msg: "EXPIRED_ACCESS_TOKEN"}, true},
		{"permission denied", &testError{msg: "permission denied for resource"}, true},
		{"network error", &testError{msg: "connection timeout"}, false},
		{"parse error", &testError{msg: "json parse failed"}, false},
		{"status 500", &testError{msg: "internal server error"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExpectedError(tt.err)
			if got != tt.expected {
				t.Errorf("isExpectedError() = %v, want %v for error: %v", got, tt.expected, tt.err)
			}
		})
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
