package utils

import (
	"context"
	"testing"
)

func TestWithRequestID(t *testing.T) {
	cases := []struct {
		name      string
		requestID string
	}{
		{
			name:      "sets simple request ID",
			requestID: "abc123",
		},
		{
			name:      "sets UUID request ID",
			requestID: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:      "sets empty request ID",
			requestID: "",
		},
		{
			name:      "sets request ID with special characters",
			requestID: "req-123_456.789",
		},
		{
			name:      "sets long request ID",
			requestID: "very-long-request-id-that-exceeds-normal-length-for-testing-purposes",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx := WithRequestID(ctx, tc.requestID)

			if newCtx == nil {
				t.Fatal("expected non-nil context")
			}

			val := newCtx.Value(requestIDKey)
			if val == nil {
				t.Fatal("expected request ID in context, got nil")
			}

			if val.(string) != tc.requestID {
				t.Fatalf("expected request ID %q, got %q", tc.requestID, val.(string))
			}
		})
	}
}

func TestGetRequestID(t *testing.T) {
	cases := []struct {
		name      string
		requestID string
	}{
		{
			name:      "gets simple request ID",
			requestID: "abc123",
		},
		{
			name:      "gets UUID request ID",
			requestID: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:      "gets empty request ID",
			requestID: "",
		},
		{
			name:      "gets request ID with special characters",
			requestID: "req-123_456.789",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), requestIDKey, tc.requestID)
			got := GetRequestID(ctx)

			if got != tc.requestID {
				t.Fatalf("expected request ID %q, got %q", tc.requestID, got)
			}
		})
	}
}

func TestWithRequestID_PreservesContext(t *testing.T) {
	type customKey string
	const myKey customKey = "my-key"

	ctx := context.WithValue(context.Background(), myKey, "my-value")
	newCtx := WithRequestID(ctx, "test-request-id")

	if val := newCtx.Value(myKey); val != "my-value" {
		t.Fatalf("expected original context value preserved, got %v", val)
	}

	if val := newCtx.Value(requestIDKey); val != "test-request-id" {
		t.Fatalf("expected request ID %q, got %v", "test-request-id", val)
	}
}

func TestRequestIDKey_Type(t *testing.T) {
	if requestIDKey != CtxKey("request_id") {
		t.Fatalf("expected requestIDKey to be 'request_id', got %v", requestIDKey)
	}
}

func TestWithRequestID_Overwrite(t *testing.T) {
	ctx := WithRequestID(context.Background(), "first-id")
	newCtx := WithRequestID(ctx, "second-id")

	got := GetRequestID(newCtx)
	if got != "second-id" {
		t.Fatalf("expected overwritten request ID %q, got %q", "second-id", got)
	}
}
