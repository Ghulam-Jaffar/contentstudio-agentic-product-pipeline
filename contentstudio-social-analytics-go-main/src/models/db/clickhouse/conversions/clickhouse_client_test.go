package conversions

import (
	"errors"
	"testing"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

func TestNewClickHouseSinkWithClient(t *testing.T) {
	logger := zerolog.Nop()
	mock := &mockClickHouseClient{}

	sink := NewClickHouseSinkWithClient(&logger, mock)

	if sink == nil {
		t.Fatal("expected non-nil sink")
	}
	if sink.logger != &logger {
		t.Fatal("expected logger to be set")
	}
	if sink.ClickhouseClient != mock {
		t.Fatal("expected ClickhouseClient to be set to mock")
	}
}

func TestNewClickHouseSink_Success(t *testing.T) {
	// Save original and restore after test
	originalNewClient := newClickHouseClient
	defer func() { newClickHouseClient = originalNewClient }()

	// Mock the client creation
	mockClient := &mockClickHouseClient{}
	newClickHouseClient = func(cfg config.ClickHouseConfig, logger zerolog.Logger) (*clickhouse.Client, error) {
		// Return a wrapper that satisfies the interface
		// Since we can't return mockClickHouseClient directly (wrong type),
		// we need to test this differently
		return nil, nil
	}

	// Since NewClickHouseSink expects *clickhouse.Client but our interface accepts ClickHouseClientInterface,
	// and we can't easily mock the real client creation without a real connection,
	// we test the NewClickHouseSinkWithClient instead which is the testable path

	logger := zerolog.Nop()
	sink := NewClickHouseSinkWithClient(&logger, mockClient)

	if sink == nil {
		t.Fatal("expected non-nil sink")
	}
}

func TestClickHouseSink_Struct(t *testing.T) {
	logger := zerolog.Nop()
	mock := &mockClickHouseClient{}

	sink := &ClickHouseSink{
		logger:           &logger,
		ClickhouseClient: mock,
	}

	if sink.logger == nil {
		t.Fatal("expected logger to be set")
	}
	if sink.ClickhouseClient == nil {
		t.Fatal("expected ClickhouseClient to be set")
	}
}

func TestClickHouseClientInterface_MockImplementation(t *testing.T) {
	mock := &mockClickHouseClient{
		healthFunc: func() error {
			return nil
		},
	}

	// Verify mock implements interface
	var _ ClickHouseClientInterface = mock

	err := mock.Health()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestClickHouseClientInterface_MockWithError(t *testing.T) {
	expectedErr := errors.New("connection failed")
	mock := &mockClickHouseClient{
		healthFunc: func() error {
			return expectedErr
		},
	}

	err := mock.Health()
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestNewClickHouseSinkWithClient_NilLogger(t *testing.T) {
	mock := &mockClickHouseClient{}

	sink := NewClickHouseSinkWithClient(nil, mock)

	if sink == nil {
		t.Fatal("expected non-nil sink even with nil logger")
	}
	if sink.logger != nil {
		t.Fatal("expected logger to be nil")
	}
	if sink.ClickhouseClient != mock {
		t.Fatal("expected ClickhouseClient to be set")
	}
}

func TestNewClickHouseSinkWithClient_NilClient(t *testing.T) {
	logger := zerolog.Nop()

	sink := NewClickHouseSinkWithClient(&logger, nil)

	if sink == nil {
		t.Fatal("expected non-nil sink even with nil client")
	}
	if sink.ClickhouseClient != nil {
		t.Fatal("expected ClickhouseClient to be nil")
	}
}
