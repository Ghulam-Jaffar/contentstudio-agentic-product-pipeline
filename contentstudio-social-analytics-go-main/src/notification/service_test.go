package notification

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/rs/zerolog"
)

// Test constants for duplicated literals
const (
	testExpectedTitle    = "acct - Data Fetched!"
	testErrorMismatchFmt = "error mismatch, got %v wantErr %v"
	testExpectedChannel  = "analytics_completed"
)

type roundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface for mock HTTP testing.
func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newDiscardLogger() zerolog.Logger {
	return zerolog.New(io.Discard)
}

func TestNewService(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		cfg        config.EmailConfig
		backendURL string
	}{
		{
			name:       "default client initialized",
			cfg:        config.EmailConfig{SMTPHost: "smtp", BackendURL: "http://example"},
			backendURL: "http://backend",
		},
		{
			name:       "empty backend still sets client",
			cfg:        config.EmailConfig{},
			backendURL: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc := NewService(tc.cfg, newDiscardLogger(), tc.backendURL)
			if svc == nil {
				t.Fatalf("service is nil")
			}
			if svc.httpClient == nil {
				t.Fatalf("http client not initialized")
			}
			if svc.backendURL != tc.backendURL {
				t.Fatalf("backendURL mismatch, got %s want %s", svc.backendURL, tc.backendURL)
			}
			if svc.config != tc.cfg {
				t.Fatalf("config mismatch, got %+v want %+v", svc.config, tc.cfg)
			}
		})
	}
}

func TestSendNotificationToDB(t *testing.T) {
	t.Parallel()

	basePayload := NotificationPayload{
		Title:       "Title",
		Headline:    "Headline",
		Description: "Desc",
		UserID:      "user",
		WorkspaceID: "workspace",
		Platform:    "platform",
		AccountID:   "account",
		Route:       "route",
		Channel:     "chan",
		Type:        "type",
	}

	cases := []struct {
		name    string
		setup   func(t *testing.T) (*Service, NotificationPayload)
		wantErr bool
	}{
		{
			name: "success 200",
			setup: func(t *testing.T) (*Service, NotificationPayload) {
				return setupServiceWithServer200(t, basePayload)
			},
			wantErr: false,
		},
		{
			name: "success 201",
			setup: func(t *testing.T) (*Service, NotificationPayload) {
				return setupServiceWithStatus(t, basePayload, http.StatusCreated)
			},
			wantErr: false,
		},
		{
			name: "backend url missing",
			setup: func(t *testing.T) (*Service, NotificationPayload) {
				return NewService(config.EmailConfig{}, newDiscardLogger(), ""), basePayload
			},
			wantErr: true,
		},
		{
			name: "http client error",
			setup: func(t *testing.T) (*Service, NotificationPayload) {
				svc := NewService(config.EmailConfig{}, newDiscardLogger(), "http://example")
				svc.httpClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("send error")
				})}
				return svc, basePayload
			},
			wantErr: true,
		},
		{
			name: "non ok status",
			setup: func(t *testing.T) (*Service, NotificationPayload) {
				return setupServiceWithStatus(t, basePayload, http.StatusBadRequest)
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			svc, payload := tc.setup(t)
			err := svc.SendNotificationToDB(payload)
			if (err != nil) != tc.wantErr {
				t.Fatalf(testErrorMismatchFmt, err, tc.wantErr)
			}
		})
	}
}

// setupServiceWithServer200 creates a service with mock server that validates full payload and returns 200.
func setupServiceWithServer200(t *testing.T, payload NotificationPayload) (*Service, NotificationPayload) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("unexpected content-type %s", ct)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		var received NotificationPayload
		if err := json.Unmarshal(body, &received); err != nil {
			t.Fatalf("failed to unmarshal payload: %v", err)
		}
		if received != payload {
			t.Fatalf("payload mismatch: got %+v want %+v", received, payload)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)
	svc := NewService(config.EmailConfig{}, newDiscardLogger(), server.URL)
	svc.httpClient = server.Client()
	return svc, payload
}

// setupServiceWithStatus creates a service with mock server that returns specified status code.
func setupServiceWithStatus(t *testing.T, payload NotificationPayload, status int) (*Service, NotificationPayload) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)
	svc := NewService(config.EmailConfig{}, newDiscardLogger(), server.URL)
	svc.httpClient = server.Client()
	return svc, payload
}

func TestSendAnalyticsNotification(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		isCompetitor  bool
		backendURL    string
		serverStatus  int
		expectRoute   string
		expectErr     bool
		expectTitle   string
		expectChannel string
	}{
		{
			name:          "non competitor success",
			isCompetitor:  false,
			backendURL:    "server",
			serverStatus:  http.StatusOK,
			expectRoute:   "social",
			expectErr:     false,
			expectTitle:   testExpectedTitle,
			expectChannel: testExpectedChannel,
		},
		{
			name:          "competitor success",
			isCompetitor:  true,
			backendURL:    "server",
			serverStatus:  http.StatusCreated,
			expectRoute:   "analyze/twitter/acc",
			expectErr:     false,
			expectTitle:   testExpectedTitle,
			expectChannel: testExpectedChannel,
		},
		{
			name:          "backend missing",
			isCompetitor:  false,
			backendURL:    "",
			serverStatus:  http.StatusOK,
			expectRoute:   "social",
			expectErr:     true,
			expectTitle:   testExpectedTitle,
			expectChannel: testExpectedChannel,
		},
		{
			name:          "backend returns error",
			isCompetitor:  false,
			backendURL:    "server",
			serverStatus:  http.StatusBadGateway,
			expectRoute:   "social",
			expectErr:     true,
			expectTitle:   testExpectedTitle,
			expectChannel: testExpectedChannel,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testAnalyticsNotificationCase(t, tc)
		})
	}
}

// testAnalyticsNotificationCase tests a single analytics notification scenario.
func testAnalyticsNotificationCase(t *testing.T, tc struct {
	name          string
	isCompetitor  bool
	backendURL    string
	serverStatus  int
	expectRoute   string
	expectErr     bool
	expectTitle   string
	expectChannel string
}) {
	var server *httptest.Server
	backendURL := tc.backendURL

	if tc.backendURL == "server" {
		server = setupAnalyticsServer(t, tc.expectRoute, tc.expectTitle, tc.expectChannel, tc.serverStatus)
		backendURL = server.URL
	}

	svc := NewService(config.EmailConfig{}, newDiscardLogger(), backendURL)
	if server != nil {
		svc.httpClient = server.Client()
	}

	err := svc.SendAnalyticsNotification("user", "workspace", "twitter", "acc", "acct", tc.isCompetitor)
	if (err != nil) != tc.expectErr {
		t.Fatalf(testErrorMismatchFmt, err, tc.expectErr)
	}
}

// setupAnalyticsServer creates a test server that validates analytics notification payload.
func setupAnalyticsServer(t *testing.T, expectRoute, expectTitle, expectChannel string, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		var payload NotificationPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("failed to unmarshal payload: %v", err)
		}
		if payload.Route != expectRoute {
			t.Fatalf("route mismatch, got %s want %s", payload.Route, expectRoute)
		}
		if payload.Channel != expectChannel {
			t.Fatalf("channel mismatch, got %s want %s", payload.Channel, expectChannel)
		}
		if payload.Title != expectTitle {
			t.Fatalf("title mismatch, got %s want %s", payload.Title, expectTitle)
		}
		w.WriteHeader(status)
	}))
}

func TestNewPusherClient(t *testing.T) {
	t.Parallel()

	cfg := config.PusherConfig{AppID: "app", Key: "key", Secret: "secret", Cluster: "cluster"}
	client := NewPusherClient(cfg, newDiscardLogger())
	if client == nil {
		t.Fatalf("pusher client nil")
	}
	if client.config != cfg {
		t.Fatalf("config mismatch: got %+v want %+v", client.config, cfg)
	}
	if client.client == nil {
		t.Fatalf("underlying client nil")
	}
}

func TestPusherTrigger(t *testing.T) {
	t.Parallel()

	cfg := config.PusherConfig{AppID: "app", Key: "key", Secret: "secret", Cluster: "cluster"}

	cases := []struct {
		name       string
		cfg        config.PusherConfig
		setup      func(p *PusherClient)
		wantErr    bool
		expectCall bool
	}{
		{
			name: "missing config returns error",
			cfg:  config.PusherConfig{},
			setup: func(p *PusherClient) {
				// No setup required: this test validates that Trigger() rejects an empty PusherConfig
				// during the configuration validation phase, before any HTTP client interaction
			},
			wantErr: true,
		},
		{
			name:    "http client error",
			cfg:     cfg,
			wantErr: true,
			setup: func(p *PusherClient) {
				p.client.HTTPClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					return nil, errors.New("network")
				})}
			},
		},
		{
			name:       "success triggers event",
			cfg:        cfg,
			expectCall: true,
			setup: func(p *PusherClient) {
				p.client.HTTPClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					if req == nil {
						return nil, errors.New("nil request")
					}
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(`{"ok":true}`)),
						Header:     make(http.Header),
						Request:    req,
					}, nil
				})}
			},
			wantErr: false,
		},
		{
			name:    "server returns error status",
			cfg:     cfg,
			wantErr: true,
			setup: func(p *PusherClient) {
				p.client.HTTPClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusBadRequest,
						Body:       io.NopCloser(bytes.NewBufferString(`{"error":"bad"}`)),
						Header:     make(http.Header),
						Request:    req,
					}, nil
				})}
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pusherClient := NewPusherClient(tc.cfg, newDiscardLogger())
			tc.setup(pusherClient)

			err := pusherClient.Trigger("channel", "event", map[string]string{"foo": "bar"})
			if (err != nil) != tc.wantErr {
				t.Fatalf(testErrorMismatchFmt, err, tc.wantErr)
			}
		})
	}
}

// ==================== Logging Contract Tests ====================

func TestLoggingContract_Notification_WarnLevelOnly(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	// Test SendNotificationToDB with missing backend URL (triggers Warn log)
	svc := NewService(config.EmailConfig{}, log.Logger, "")
	err := svc.SendNotificationToDB(NotificationPayload{
		Title:  "test",
		UserID: "user1",
	})
	if err == nil {
		t.Fatal("expected error for missing backend URL")
	}

	output := buf.String()
	if !strings.Contains(output, "WRN") {
		t.Fatalf("expected WRN in log output, got: %s", output)
	}
	if strings.Contains(output, "ERR") {
		t.Fatalf("unexpected ERR-level log in output: %s", output)
	}

	// Test SendNotificationToDB with HTTP client error
	buf.Reset()
	svc2 := NewService(config.EmailConfig{}, log.Logger, "http://example")
	svc2.httpClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	})}
	err = svc2.SendNotificationToDB(NotificationPayload{
		Title:  "test",
		UserID: "user1",
	})
	if err == nil {
		t.Fatal("expected error for HTTP client failure")
	}

	output = buf.String()
	if !strings.Contains(output, "WRN") {
		t.Fatalf("expected WRN in log output for HTTP error, got: %s", output)
	}
	if strings.Contains(output, "ERR") {
		t.Fatalf("unexpected ERR-level log for HTTP error: %s", output)
	}

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls, got %d", len(*captureRecords))
	}
}

func TestLoggingContract_Notification_NoCaptureException(t *testing.T) {
	log, _ := logger.NewTestLoggerWithHook()
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	// Trigger multiple error paths
	svc := NewService(config.EmailConfig{}, log.Logger, "")
	_ = svc.SendNotificationToDB(NotificationPayload{Title: "test", UserID: "u1"})

	svc2 := NewService(config.EmailConfig{}, log.Logger, "http://example")
	svc2.httpClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	})}
	_ = svc2.SendNotificationToDB(NotificationPayload{Title: "test", UserID: "u2"})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()
	svc3 := NewService(config.EmailConfig{}, log.Logger, server.URL)
	svc3.httpClient = server.Client()
	_ = svc3.SendNotificationToDB(NotificationPayload{Title: "test", UserID: "u3"})

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls across all error paths, got %d", len(*captureRecords))
	}
}
