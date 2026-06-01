package logger

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	cases := []struct {
		name          string
		level         string
		expectedLevel zerolog.Level
	}{
		{"debug level", "debug", zerolog.DebugLevel},
		{"info level", "info", zerolog.InfoLevel},
		{"warn level", "warn", zerolog.WarnLevel},
		{"error level", "error", zerolog.ErrorLevel},
		{"fatal level", "fatal", zerolog.FatalLevel},
		{"panic level", "panic", zerolog.PanicLevel},
		{"uppercase DEBUG", "DEBUG", zerolog.DebugLevel},
		{"uppercase INFO", "INFO", zerolog.InfoLevel},
		{"mixed case Info", "Info", zerolog.InfoLevel},
		{"invalid level defaults to info", "invalid", zerolog.InfoLevel},
		{"empty level defaults to info", "", zerolog.InfoLevel},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			log := New(tc.level)
			if log == nil {
				t.Fatal("expected non-nil logger")
			}
		})
	}
}

func TestNewTestLogger(t *testing.T) {
	log, buf := NewTestLogger()

	if log == nil {
		t.Fatal("expected non-nil logger")
	}
	if buf == nil {
		t.Fatal("expected non-nil buffer")
	}

	log.Info().Msg("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Fatalf("expected buffer to contain 'test message', got: %s", output)
	}
}

func TestLogger_WithModule(t *testing.T) {
	log, buf := NewTestLogger()

	moduleLog := log.WithModule("test-module")
	if moduleLog == nil {
		t.Fatal("expected non-nil logger")
	}

	moduleLog.Info().Msg("module message")

	output := buf.String()
	if !strings.Contains(output, "test-module") {
		t.Fatalf("expected buffer to contain 'test-module', got: %s", output)
	}
}

func TestLogger_WithSubmodule(t *testing.T) {
	log, buf := NewTestLogger()

	subLog := log.WithSubmodule("test-submodule")
	if subLog == nil {
		t.Fatal("expected non-nil logger")
	}

	subLog.Info().Msg("submodule message")

	output := buf.String()
	if !strings.Contains(output, "test-submodule") {
		t.Fatalf("expected buffer to contain 'test-submodule', got: %s", output)
	}
}

func TestNewWith(t *testing.T) {
	log := NewWith("debug", "my-module")
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestOperation_Basic(t *testing.T) {
	log, buf := NewTestLogger()

	op := log.Operation("test-operation")
	if op == nil {
		t.Fatal("expected non-nil operation")
	}
	if op.name != "test-operation" {
		t.Fatalf("expected name 'test-operation', got %s", op.name)
	}
	if op.logger != log {
		t.Fatal("expected operation logger to match")
	}

	op.Start("starting test")
	op.Complete(nil, "completed test")

	output := buf.String()
	if !strings.Contains(output, "test-operation") {
		t.Fatalf("expected output to contain 'test-operation', got: %s", output)
	}
}

func TestOperation_WithField(t *testing.T) {
	log, buf := NewTestLogger()

	op := log.Operation("field-test").
		WithField("key1", "value1").
		WithField("key2", 123)

	op.Start("")
	op.Complete(nil, "")

	output := buf.String()
	if !strings.Contains(output, "field-test") {
		t.Fatalf("expected output to contain 'field-test', got: %s", output)
	}
}

func TestOperation_WithFields(t *testing.T) {
	log, buf := NewTestLogger()

	fields := map[string]interface{}{
		"field1": "value1",
		"field2": 42,
		"field3": true,
	}

	op := log.Operation("fields-test").WithFields(fields)
	op.Start("test")
	op.Complete(nil, "done")

	output := buf.String()
	if !strings.Contains(output, "fields-test") {
		t.Fatalf("expected output to contain 'fields-test', got: %s", output)
	}
}

func TestOperation_WithSentryTag(t *testing.T) {
	log, _ := NewTestLogger()

	op := log.Operation("sentry-tag-test").
		WithSentryTag("env", "test").
		WithSentryTag("service", "logger")

	if op.tags == nil {
		t.Fatal("expected tags map to be initialized")
	}
	if op.tags["env"] != "test" {
		t.Fatalf("expected tag 'env' to be 'test', got %s", op.tags["env"])
	}
	if op.tags["service"] != "logger" {
		t.Fatalf("expected tag 'service' to be 'logger', got %s", op.tags["service"])
	}
}

func TestOperation_WithSentryTags(t *testing.T) {
	log, _ := NewTestLogger()

	tags := map[string]string{
		"tag1": "value1",
		"tag2": "value2",
	}

	op := log.Operation("sentry-tags-test").WithSentryTags(tags)

	if len(op.tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(op.tags))
	}
}

func TestOperation_WithSentryExtra(t *testing.T) {
	log, _ := NewTestLogger()

	op := log.Operation("sentry-extra-test").
		WithSentryExtra("data", map[string]int{"count": 5}).
		WithSentryExtra("items", []string{"a", "b"})

	if op.extras == nil {
		t.Fatal("expected extras map to be initialized")
	}
	if len(op.extras) != 2 {
		t.Fatalf("expected 2 extras, got %d", len(op.extras))
	}
}

func TestOperation_WithSentryExtras(t *testing.T) {
	log, _ := NewTestLogger()

	extras := map[string]interface{}{
		"extra1": "value1",
		"extra2": 123,
	}

	op := log.Operation("sentry-extras-test").WithSentryExtras(extras)

	if len(op.extras) != 2 {
		t.Fatalf("expected 2 extras, got %d", len(op.extras))
	}
}

func TestOperation_WithEvent(t *testing.T) {
	log, buf := NewTestLogger()

	mutator := func(evt *zerolog.Event) {
		evt.Str("custom", "field")
	}

	op := log.Operation("event-test").WithEvent(mutator)
	op.Start("test")

	output := buf.String()
	if !strings.Contains(output, "event-test") {
		t.Fatalf("expected output to contain 'event-test', got: %s", output)
	}
}

func TestOperation_WithEvent_Nil(t *testing.T) {
	log, _ := NewTestLogger()

	op := log.Operation("nil-event-test").WithEvent(nil)

	initialFieldsCount := len(op.fields)
	if initialFieldsCount != 0 {
		t.Fatalf("expected 0 fields after nil mutator, got %d", initialFieldsCount)
	}
}

func TestOperation_Start_DefaultMessage(t *testing.T) {
	log, buf := NewTestLogger()

	op := log.Operation("default-start-test")
	op.Start("")

	output := buf.String()
	if !strings.Contains(output, "operation started") {
		t.Fatalf("expected default message 'operation started', got: %s", output)
	}
}

func TestOperation_Start_DoubleStart(t *testing.T) {
	log, buf := NewTestLogger()

	op := log.Operation("double-start-test")
	op.Start("first start")
	buf.Reset()
	op.Start("second start")

	output := buf.String()
	if strings.Contains(output, "second start") {
		t.Fatal("expected second start to be ignored")
	}
}

func TestOperation_Complete_WithError(t *testing.T) {
	log, buf := NewTestLogger()

	op := log.Operation("error-test")
	op.Start("starting")

	testErr := &testError{msg: "test error"}
	op.Complete(testErr, "")

	output := buf.String()
	if !strings.Contains(output, "operation failed") {
		t.Fatalf("expected default error message 'operation failed', got: %s", output)
	}
}

func TestOperation_Complete_WithoutStart(t *testing.T) {
	log, buf := NewTestLogger()

	op := log.Operation("no-start-test")
	op.Complete(nil, "completed without start")

	output := buf.String()
	if !strings.Contains(output, "completed without start") {
		t.Fatalf("expected message 'completed without start', got: %s", output)
	}
}

func TestOperation_Complete_Duration(t *testing.T) {
	log, buf := NewTestLogger()

	op := log.Operation("duration-test")
	op.Start("starting")
	time.Sleep(10 * time.Millisecond)
	op.Complete(nil, "done")

	output := buf.String()
	if !strings.Contains(output, "duration") {
		t.Fatalf("expected output to contain 'duration', got: %s", output)
	}
}

func TestOperation_Chaining(t *testing.T) {
	log, buf := NewTestLogger()

	op := log.Operation("chained-test").
		WithField("key", "value").
		WithSentryTag("tag", "value").
		WithSentryExtra("extra", "data").
		Start("chained start")

	op.Complete(nil, "chained complete")

	output := buf.String()
	if !strings.Contains(output, "chained-test") {
		t.Fatalf("expected output to contain 'chained-test', got: %s", output)
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestOperation_ApplyFields(t *testing.T) {
	var buf bytes.Buffer
	log := zerolog.New(&buf)
	logger := &Logger{log}

	op := logger.Operation("apply-fields-test").
		WithField("field1", "value1").
		WithField("field2", 42)

	evt := log.Info()
	op.applyFields(evt)
	evt.Msg("test")

	output := buf.String()
	if output == "" {
		t.Fatal("expected some output")
	}
}
