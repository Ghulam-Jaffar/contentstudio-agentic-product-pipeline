package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/pusher/pusher-http-go/v5"
	"github.com/rs/zerolog"
)

// Service handles email notifications
type Service struct {
	config     config.EmailConfig
	backendURL string
	logger     zerolog.Logger
	httpClient *http.Client
}

// NotificationPayload represents the payload sent to the backend API
type NotificationPayload struct {
	Title        string `json:"title"`
	Headline     string `json:"headline"`
	Description  string `json:"description"`
	UserID       string `json:"user_id"`
	WorkspaceID  string `json:"workspace_id"`
	Platform     string `json:"platform"`
	AccountID    string `json:"account_id"`
	Route        string `json:"route"`
	Channel      string `json:"channel"`
	Type         string `json:"type"`
	IsCompetitor bool   `json:"is_competitor,omitempty"`
}

// NewService creates a new notification service
func NewService(cfg config.EmailConfig, log zerolog.Logger, backendURL string) *Service {
	return &Service{
		config:     cfg,
		backendURL: backendURL,
		logger:     log.With().Str("component", "notification").Logger(),
		httpClient: &http.Client{},
	}
}

// zeroObjectID is the hex representation of a zero MongoDB ObjectID.
// It indicates a missing or unset user_id on a social integration document.
const zeroObjectID = "000000000000000000000000"

// SendAnalyticsNotification sends an analytics completion notification to the backend
// This is used for both social accounts and competitors when analytics are fetched
func (s *Service) SendAnalyticsNotification(
	userID string,
	workspaceID string,
	platform string,
	accountID string,
	accountName string,
	isCompetitor bool,
) error {
	if userID == "" || userID == zeroObjectID {
		s.logger.Warn().
			Str("platform", platform).
			Str("account_id", accountID).
			Str("workspace_id", workspaceID).
			Msg("SendAnalyticsNotification: skipping notification because user_id is missing or zero")
		return fmt.Errorf("SendAnalyticsNotification: user_id is missing or zero for account %s", accountID)
	}

	route := "social"
	if isCompetitor {
		route = "analyze/" + platform + "/" + accountID
	}

	payload := NotificationPayload{
		Title:        accountName + " - Data Fetched!",
		Headline:     "Analytics fetched for " + platform + " profile: " + accountName,
		Description:  "We have fetched and processed your data against your " + platform + " profile <span style='font-weight: 500;'>" + accountName + "</span> in workspace.",
		UserID:       userID,
		WorkspaceID:  workspaceID,
		Platform:     platform,
		AccountID:    accountID,
		Type:         "analytics_completed",
		Channel:      "analytics_completed",
		Route:        route,
		IsCompetitor: isCompetitor,
	}

	s.logger.Info().
		Interface("payload", payload).
		Msg("Prepared analytics notification payload")

	return s.SendNotificationToDB(payload)
}

// SendNotificationToDB sends a notification to the backend API for email delivery
func (s *Service) SendNotificationToDB(payload NotificationPayload) error {
	if s.backendURL == "" {
		s.logger.Warn().Msg("Backend URL not configured, skipping notification")
		return fmt.Errorf("Service.SendNotificationToDB: backend URL not configured")
	}

	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		s.logger.Warn().
			Err(err).
			Msg("Failed to marshal notification payload")
		return err
	}

	// Create request
	url := s.backendURL
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		s.logger.Warn().
			Err(err).
			Str("url", url).
			Msg("Failed to create HTTP request")
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("cache-control", "no-cache")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Warn().
			Err(err).
			Str("url", url).
			Str("user_id", payload.UserID).
			Msg("Failed to send notification to backend")
		return err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		s.logger.Warn().
			Int("status_code", resp.StatusCode).
			Str("url", url).
			Str("user_id", payload.UserID).
			Str("response", string(bodyBytes)).
			Msg("Backend returned error for notification")
		return fmt.Errorf("Service.SendNotificationToDB: backend returned status %d", resp.StatusCode)
	}

	s.logger.Info().
		Str("url", url).
		Str("user_id", payload.UserID).
		Str("platform", payload.Platform).
		Bool("is_competitor", payload.IsCompetitor).
		Msg("Notification sent to backend successfully")
	return nil
}

// PusherClient handles real-time notifications via Pusher
type PusherClient struct {
	config config.PusherConfig
	client *pusher.Client
	logger zerolog.Logger
}

// NewPusherClient creates a new Pusher client
func NewPusherClient(cfg config.PusherConfig, log zerolog.Logger) *PusherClient {
	logger := log.With().Str("component", "pusher").Logger()

	// Create Pusher client
	client := &pusher.Client{
		AppID:   cfg.AppID,
		Key:     cfg.Key,
		Secret:  cfg.Secret,
		Cluster: cfg.Cluster,
		Secure:  true,
	}

	return &PusherClient{
		config: cfg,
		client: client,
		logger: logger,
	}
}

// Trigger sends a real-time event via Pusher
func (p *PusherClient) Trigger(channel, event string, data interface{}) error {
	// Validate configuration
	if p.config.AppID == "" || p.config.Key == "" || p.config.Secret == "" || p.config.Cluster == "" {
		p.logger.Warn().
			Str("channel", channel).
			Str("event", event).
			Msg("Pusher configuration is incomplete")
		return fmt.Errorf("PusherClient.Trigger: pusher configuration is incomplete")
	}

	p.logger.Info().
		Str("channel", channel).
		Str("event", event).
		Msg("Triggering Pusher event")

	// Trigger the event
	err := p.client.Trigger(channel, event, data)
	if err != nil {
		p.logger.Warn().
			Err(err).
			Str("channel", channel).
			Str("event", event).
			Msg("Failed to trigger Pusher event")
		return err
	}

	p.logger.Info().
		Str("channel", channel).
		Str("event", event).
		Msg("Pusher event triggered successfully")
	return nil
}
