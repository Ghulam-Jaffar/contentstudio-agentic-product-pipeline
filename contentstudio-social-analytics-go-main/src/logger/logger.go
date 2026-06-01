package logger

import (
	"bytes"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Logger is a wrapper around zerolog.Logger
type Logger struct {
	zerolog.Logger
}

// Operation represents a named unit of work that logs start/completion with structured context.
type Operation struct {
	logger  *Logger
	name    string
	start   time.Time
	started bool
	fields  []func(*zerolog.Event)
	tags    map[string]string
	extras  map[string]interface{}
}

// New initializes a new zerolog.Logger instance.
// The `level` argument can be "debug", "info", "warn", "error", "fatal", "panic".
// It defaults to "info" if an invalid level is provided.
func New(level string) *Logger {
	enableSentryHookFromEnv()

	var l zerolog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = zerolog.DebugLevel
	case "info":
		l = zerolog.InfoLevel
	case "warn":
		l = zerolog.WarnLevel
	case "error":
		l = zerolog.ErrorLevel
	case "fatal":
		l = zerolog.FatalLevel
	case "panic":
		l = zerolog.PanicLevel
	default:
		l = zerolog.InfoLevel
	}

	// Use ConsoleWriter for pretty output during development
	// For production, you might want to switch to plain JSON output: zerolog.New(os.Stdout)
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	// Add caller information
	// For performance in production, you might remove .Caller()
	logger := zerolog.New(output).Level(l).With().Timestamp().Caller().Logger()

	if isSentryHookEnabled() {
		logger = logger.Hook(sentryHook{})
	}

	return &Logger{logger}
}

// NewNop returns a logger that discards all output. Useful for tests.
func NewNop() *Logger {
	return &Logger{zerolog.Nop()}
}

func NewTestLogger() (*Logger, *bytes.Buffer) {
	var buf bytes.Buffer

	// Use ConsoleWriter for human-readable output (optional: switch to JSON if needed)
	output := zerolog.ConsoleWriter{
		Out:        &buf,
		TimeFormat: time.RFC3339,
		NoColor:    true,
	}

	// Create the logger
	log := zerolog.New(output).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{log}, &buf
}

// NewTestLoggerWithHook creates a test logger with the sentry hook attached.
// Use this when tests need to verify that Error-level logs trigger the sentry hook.
func NewTestLoggerWithHook() (*Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	output := zerolog.ConsoleWriter{
		Out:        &buf,
		TimeFormat: time.RFC3339,
		NoColor:    true,
	}
	log := zerolog.New(output).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Caller().
		Logger().
		Hook(sentryHook{})
	return &Logger{log}, &buf
}

// WithModule returns a child logger with a "module" field attached.
// This helps keep logs consistent and searchable across components.
func (l *Logger) WithModule(module string) *Logger {
	ll := l.With().Str("module", module).Logger()
	return &Logger{ll}
}

// WithSubmodule returns a child logger with a "submodule" field attached.
// Useful for finer-grained breakdown inside a module.
func (l *Logger) WithSubmodule(sub string) *Logger {
	ll := l.With().Str("submodule", sub).Logger()
	return &Logger{ll}
}

// NewWith creates a new logger at the specified level and attaches a module field.
func NewWith(level, module string) *Logger {
	return New(level).WithModule(module)
}

// Operation creates a new structured operation tracker for long-running or important functions.
func (l *Logger) Operation(name string) *Operation {
	return &Operation{
		logger: l,
		name:   name,
		fields: make([]func(*zerolog.Event), 0, 4),
	}
}

// WithField attaches a custom field to every log emitted by the operation.
func (o *Operation) WithField(key string, value interface{}) *Operation {
	o.fields = append(o.fields, func(evt *zerolog.Event) {
		evt.Interface(key, value)
	})
	return o
}

// WithFields attaches multiple fields at once.
func (o *Operation) WithFields(fields map[string]interface{}) *Operation {
	for k, v := range fields {
		o.WithField(k, v)
	}
	return o
}

// WithSentryTag attaches a tag to the Sentry event if this operation logs an error.
func (o *Operation) WithSentryTag(key, value string) *Operation {
	if o.tags == nil {
		o.tags = make(map[string]string)
	}
	o.tags[key] = value
	return o
}

// WithSentryTags attaches multiple tags to the Sentry event if this operation logs an error.
func (o *Operation) WithSentryTags(tags map[string]string) *Operation {
	for k, v := range tags {
		o.WithSentryTag(k, v)
	}
	return o
}

// WithSentryExtra adds structured context to the Sentry event when this operation fails.
func (o *Operation) WithSentryExtra(key string, value interface{}) *Operation {
	if o.extras == nil {
		o.extras = make(map[string]interface{})
	}
	o.extras[key] = value
	return o
}

// WithSentryExtras attaches multiple extra context objects to the Sentry event.
func (o *Operation) WithSentryExtras(extras map[string]interface{}) *Operation {
	for k, v := range extras {
		o.WithSentryExtra(k, v)
	}
	return o
}

// WithEvent allows callers to provide a custom mutation for the underlying event.
func (o *Operation) WithEvent(mutator func(*zerolog.Event)) *Operation {
	if mutator != nil {
		o.fields = append(o.fields, mutator)
	}
	return o
}

// Start logs the beginning of the operation. If message is empty, a default is used.
func (o *Operation) Start(message string) *Operation {
	if o.started {
		return o
	}
	if message == "" {
		message = "operation started"
	}
	event := o.logger.Info().Str("operation", o.name)
	o.applyFields(event)
	event.Msg(message)
	o.start = time.Now()
	o.started = true
	return o
}

// Complete logs the completion of the operation with optional error information.
// If message is empty, it defaults to "operation completed" or "operation failed".
func (o *Operation) Complete(err error, message string) {
	if !o.started {
		o.start = time.Now()
	}
	if message == "" {
		if err != nil {
			message = "operation failed"
		} else {
			message = "operation completed"
		}
	}

	var event *zerolog.Event
	if err != nil {
		event = o.logger.Error().Err(err)
	} else {
		event = o.logger.Info()
	}

	event = event.Str("operation", o.name).Dur("duration", time.Since(o.start))
	o.applyFields(event)
	event.Msg(message)

	if err != nil {
		tags := map[string]string{"operation": o.name}
		for k, v := range o.tags {
			tags[k] = v
		}
		CaptureException(err, tags, o.extras)
	}
}

func (o *Operation) applyFields(evt *zerolog.Event) {
	for _, setter := range o.fields {
		setter(evt)
	}
}
