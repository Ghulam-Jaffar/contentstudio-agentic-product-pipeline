package httputil

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func TestWriteJSON(t *testing.T) {
	type floatData struct {
		Value float64
	}

	tests := []struct {
		name           string
		statusCode     int
		body           interface{}
		expectedStatus int
		checkBody      func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:           "status 200 with map body",
			statusCode:     http.StatusOK,
			body:           map[string]string{"key": "value"},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result map[string]string
				json.NewDecoder(w.Body).Decode(&result)
				if result["key"] != "value" {
					t.Fatalf("expected 'value', got %q", result["key"])
				}
			},
		},
		{
			name:           "status 201",
			statusCode:     http.StatusCreated,
			body:           map[string]int{"count": 5},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "nil body",
			statusCode:     http.StatusOK,
			body:           nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "sanitizes NaN to 0",
			statusCode:     http.StatusOK,
			body:           &floatData{Value: math.NaN()},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result map[string]float64
				json.NewDecoder(w.Body).Decode(&result)
				if result["Value"] != 0 {
					t.Fatalf("expected 0, got %v", result["Value"])
				}
			},
		},
		{
			name:           "sanitizes +Inf to 0",
			statusCode:     http.StatusOK,
			body:           &floatData{Value: math.Inf(1)},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result map[string]float64
				json.NewDecoder(w.Body).Decode(&result)
				if result["Value"] != 0 {
					t.Fatalf("expected 0, got %v", result["Value"])
				}
			},
		},
		{
			name:           "sanitizes -Inf to 0",
			statusCode:     http.StatusOK,
			body:           &floatData{Value: math.Inf(-1)},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result map[string]float64
				json.NewDecoder(w.Body).Decode(&result)
				if result["Value"] != 0 {
					t.Fatalf("expected 0, got %v", result["Value"])
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteJSON(w, tc.statusCode, tc.body)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("expected application/json, got %q", ct)
			}
			if tc.checkBody != nil {
				tc.checkBody(t, w)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	logger := zerolog.New(io.Discard)

	tests := []struct {
		name            string
		err             error
		expectedStatus  int
		expectedMessage string
	}{
		{
			name:            "validation error returns 400",
			err:             NewValidationError("invalid field"),
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "invalid field",
		},
		{
			name:            "generic error returns 500",
			err:             errors.New("db connection failed"),
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "internal server error",
		},
		{
			name:            "unauthorized http error returns 401",
			err:             NewUnauthorizedError("invalid token"),
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "invalid token",
		},
		{
			name:            "forbidden http error returns 403",
			err:             NewForbiddenError("access denied"),
			expectedStatus:  http.StatusForbidden,
			expectedMessage: "access denied",
		},
		{
			name:            "bad request http error returns 400",
			err:             NewBadRequestError("missing required param"),
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "missing required param",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteError(w, logger, tc.err)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
			var resp ErrorResponse
			json.NewDecoder(w.Body).Decode(&resp)
			if resp.Status != false {
				t.Fatal("expected status false")
			}
			if resp.Message != tc.expectedMessage {
				t.Fatalf("expected %q, got %q", tc.expectedMessage, resp.Message)
			}
		})
	}
}

func TestWriteStatusError(t *testing.T) {
	w := httptest.NewRecorder()

	WriteStatusError(w, http.StatusServiceUnavailable, "service unavailable")

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if resp.Status {
		t.Fatal("expected status false")
	}
	if resp.Message != "service unavailable" {
		t.Fatalf("expected message %q, got %q", "service unavailable", resp.Message)
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{name: "simple message", message: "field required"},
		{name: "empty message", message: ""},
		{name: "detailed message", message: "start_date must be in YYYY-MM-DD format"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := NewValidationError(tc.message)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if err.Error() != tc.message {
				t.Fatalf("expected %q, got %q", tc.message, err.Error())
			}
			if err.Message != tc.message {
				t.Fatalf("expected Message %q, got %q", tc.message, err.Message)
			}
		})
	}
}

func TestSanitizeFloats(t *testing.T) {
	type floatStruct struct {
		Val float64
	}
	type float32Struct struct {
		Val float32
	}
	type nestedStruct struct {
		Inner *floatStruct
	}
	type sliceStruct struct {
		Items []floatStruct
	}
	type interfaceStruct struct {
		Val interface{}
	}

	tests := []struct {
		name  string
		input interface{}
		check func(t *testing.T)
	}{
		{
			name:  "nil input",
			input: nil,
			check: func(t *testing.T) { /* should not panic */ },
		},
		{
			name:  "pointer to struct with NaN float64",
			input: &floatStruct{Val: math.NaN()},
			check: func(t *testing.T) {
				// checked via sanitizeFloats modifying in-place
			},
		},
		{
			name:  "pointer to struct with NaN float32",
			input: &float32Struct{Val: float32(math.NaN())},
			check: func(t *testing.T) {},
		},
		{
			name:  "normal float left unchanged",
			input: &floatStruct{Val: 42.5},
			check: func(t *testing.T) {},
		},
		{
			name:  "nested pointer with -Inf",
			input: &nestedStruct{Inner: &floatStruct{Val: math.Inf(-1)}},
			check: func(t *testing.T) {},
		},
		{
			name:  "nil nested pointer",
			input: &nestedStruct{Inner: nil},
			check: func(t *testing.T) { /* should not panic */ },
		},
		{
			name:  "slice of structs with NaN",
			input: &sliceStruct{Items: []floatStruct{{Val: math.NaN()}, {Val: 5.0}}},
			check: func(t *testing.T) {},
		},
		{
			name: "map with interface pointer",
			input: &map[string]interface{}{
				"a": &floatStruct{Val: math.NaN()},
			},
			check: func(t *testing.T) {},
		},
		{
			name: "map with direct float values",
			input: &map[string]interface{}{
				"a": math.NaN(),
				"b": math.Inf(1),
			},
			check: func(t *testing.T) {},
		},
		{
			name:  "interface field with struct",
			input: &interfaceStruct{Val: &floatStruct{Val: math.NaN()}},
			check: func(t *testing.T) {},
		},
		{
			name:  "nil interface field",
			input: &interfaceStruct{Val: nil},
			check: func(t *testing.T) { /* should not panic */ },
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			sanitizeFloats(tc.input) // must not panic
			tc.check(t)
		})
	}

	// Verify in-place mutation for key cases
	t.Run("verify NaN float64 becomes 0", func(t *testing.T) {
		d := &floatStruct{Val: math.NaN()}
		sanitizeFloats(d)
		if d.Val != 0 {
			t.Fatalf("expected 0, got %v", d.Val)
		}
	})
	t.Run("verify NaN float32 becomes 0", func(t *testing.T) {
		d := &float32Struct{Val: float32(math.NaN())}
		sanitizeFloats(d)
		if d.Val != 0 {
			t.Fatalf("expected 0, got %v", d.Val)
		}
	})
	t.Run("verify nested -Inf becomes 0", func(t *testing.T) {
		d := &nestedStruct{Inner: &floatStruct{Val: math.Inf(-1)}}
		sanitizeFloats(d)
		if d.Inner.Val != 0 {
			t.Fatalf("expected 0, got %v", d.Inner.Val)
		}
	})
	t.Run("verify slice NaN becomes 0", func(t *testing.T) {
		d := &sliceStruct{Items: []floatStruct{{Val: math.NaN()}, {Val: 5.0}}}
		sanitizeFloats(d)
		if d.Items[0].Val != 0 {
			t.Fatalf("expected 0, got %v", d.Items[0].Val)
		}
		if d.Items[1].Val != 5.0 {
			t.Fatalf("expected 5.0, got %v", d.Items[1].Val)
		}
	})
	t.Run("verify normal float unchanged", func(t *testing.T) {
		d := &floatStruct{Val: 42.5}
		sanitizeFloats(d)
		if d.Val != 42.5 {
			t.Fatalf("expected 42.5, got %v", d.Val)
		}
	})
	t.Run("verify map float values become 0", func(t *testing.T) {
		d := &map[string]interface{}{
			"a": math.NaN(),
			"b": math.Inf(-1),
			"c": 3.5,
		}
		sanitizeFloats(d)
		if got := (*d)["a"].(float64); got != 0 {
			t.Fatalf("expected 0, got %v", got)
		}
		if got := (*d)["b"].(float64); got != 0 {
			t.Fatalf("expected 0, got %v", got)
		}
		if got := (*d)["c"].(float64); got != 3.5 {
			t.Fatalf("expected 3.5, got %v", got)
		}
	})
}
