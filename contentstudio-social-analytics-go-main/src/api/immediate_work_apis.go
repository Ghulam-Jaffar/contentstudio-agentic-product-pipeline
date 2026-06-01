package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafakaModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
)

type APIServer struct {
	MongoClient            *mongo.Client
	UnifiedRepo            mongodb.UnifiedSocialRepository
	ListeningRepo          ListeningTopicGetter
	ListeningWorkspaceRepo ListeningWorkspaceQuotaChecker
	Producer               kafka.Producer
	Logger                 *logger.Logger
	Config                 *config.Config
	HttpServer             *http.Server
}

type APIError struct {
	Status    string `json:"status"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	Timestamp string `json:"timestamp"`
}

const (
	ErrCodeInvalidRequest       = "INVALID_REQUEST"
	ErrCodeMissingField         = "MISSING_FIELD"
	ErrCodeInvalidChannel       = "INVALID_CHANNEL"
	ErrCodeInvalidAccountID     = "INVALID_ACCOUNT_ID"
	ErrCodeAccountNotFound      = "ACCOUNT_NOT_FOUND"
	ErrCodeNoAccessToken        = "NO_ACCESS_TOKEN"
	ErrCodeNoRefreshToken       = "NO_REFRESH_TOKEN"
	ErrCodeConsentExpired       = "CONSENT_EXPIRED"
	ErrCodeKafkaError           = "KAFKA_ERROR"
	ErrCodeInternalError        = "INTERNAL_ERROR"
	ErrCodeForbidden            = "FORBIDDEN"
	ErrCodeTopicNotFound        = "TOPIC_NOT_FOUND"
	ErrCodeMentionsLimitReached = "MENTIONS_LIMIT_REACHED"

	// YouTube consent validation
	youtubeConsentMaxDays = 30
)

type immediateWorkContextKey struct{}

type immediateWorkOptions struct {
	StartDate string
	EndDate   string
	NTweets   int
}

func withImmediateWorkOptions(ctx context.Context, opts immediateWorkOptions) context.Context {
	return context.WithValue(ctx, immediateWorkContextKey{}, opts)
}

func getImmediateWorkOptions(ctx context.Context) immediateWorkOptions {
	opts, ok := ctx.Value(immediateWorkContextKey{}).(immediateWorkOptions)
	if !ok {
		return immediateWorkOptions{}
	}
	return opts
}

func (s *APIServer) sendErrorResponse(w http.ResponseWriter, statusCode int, code, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(APIError{
		Status:    "error",
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func (s *APIServer) sendSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// handleImmediateWork handles immediate work requests from PHP
func (s *APIServer) HandleImmediateWork(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendErrorResponse(w, http.StatusMethodNotAllowed, ErrCodeInvalidRequest, "Method not allowed", "Only POST method is supported")
		return
	}

	// Parse request body
	var req api.ImmediateWorkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to decode request body")
		s.sendErrorResponse(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}

	// Validate request
	if req.AccountID == "" {
		s.Logger.Error().Msg("Missing account_id in request")
		s.sendErrorResponse(w, http.StatusBadRequest, ErrCodeMissingField, "Missing required field", "account_id is required")
		return
	}
	if req.Channel == "" {
		s.Logger.Error().Msg("Missing channel in request")
		s.sendErrorResponse(w, http.StatusBadRequest, ErrCodeMissingField, "Missing required field", "channel is required")
		return
	}

	// Validate channel
	validChannels := map[string]bool{
		"facebook":  true,
		"instagram": true,
		"linkedin":  true,
		"tiktok":    true,
		"youtube":   true,
		"twitter":   true,
		"pinterest": true,
		"meta_ads":  true,
		"gmb":       true,
	}
	if !validChannels[req.Channel] {
		s.Logger.Error().Str("channel", req.Channel).Msg("Invalid channel")
		s.sendErrorResponse(w, http.StatusBadRequest, ErrCodeInvalidChannel, "Invalid channel", "Channel must be one of: facebook, instagram, linkedin, tiktok, youtube, twitter, pinterest, gmb")
		return
	}

	s.Logger.Info().
		Str("account_id", req.AccountID).
		Str("channel", req.Channel).
		Msg("Received immediate work request")

	// Process the request
	ctx := withImmediateWorkOptions(r.Context(), immediateWorkOptions{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		NTweets:   req.NTweets,
	})
	if err := s.ProcessImmediateWork(ctx, req); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to process immediate work request")
		s.handleProcessingError(w, err, req.AccountID, req.Channel)
		return
	}

	// Return success response
	s.sendSuccessResponse(w, map[string]interface{}{
		"status":     "success",
		"message":    "Work order dispatched successfully",
		"account_id": req.AccountID,
		"channel":    req.Channel,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// handleProcessingError maps internal errors to appropriate API error responses
func (s *APIServer) handleProcessingError(w http.ResponseWriter, err error, accountID, channel string) {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "invalid account ID format"):
		s.sendErrorResponse(w, http.StatusBadRequest, ErrCodeInvalidAccountID, "Invalid account ID format", errMsg)
	case strings.Contains(errMsg, "not found"):
		s.sendErrorResponse(w, http.StatusNotFound, ErrCodeAccountNotFound, "Account not found", fmt.Sprintf("No %s account found with ID: %s", channel, accountID))
	case strings.Contains(errMsg, "no valid access token"):
		s.sendErrorResponse(w, http.StatusBadRequest, ErrCodeNoAccessToken, "No access token available", errMsg)
	case strings.Contains(errMsg, "no refresh token"):
		s.sendErrorResponse(w, http.StatusBadRequest, ErrCodeNoRefreshToken, "No refresh token available", errMsg)
	case strings.Contains(errMsg, "youtube consent expired"):
		s.sendErrorResponse(w, http.StatusForbidden, ErrCodeConsentExpired, "YouTube consent has expired", errMsg)
	case strings.Contains(errMsg, "failed to send to Kafka"):
		s.sendErrorResponse(w, http.StatusServiceUnavailable, ErrCodeKafkaError, "Failed to dispatch work order", errMsg)
	default:
		s.sendErrorResponse(w, http.StatusInternalServerError, ErrCodeInternalError, "Internal server error", errMsg)
	}
}

// ProcessImmediateWork looks up the account and dispatches to Kafka
func (s *APIServer) ProcessImmediateWork(ctx context.Context, req api.ImmediateWorkRequest) error {
	accountID, err := primitive.ObjectIDFromHex(req.AccountID)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessImmediateWork: invalid account ID format: %w", err)
	}

	switch req.Channel {
	case "facebook":

		return s.ProcessFacebookWork(ctx, accountID)

	case "instagram":

		return s.ProcessInstagramWork(ctx, accountID)

	case "linkedin":

		return s.ProcessLinkedinWork(ctx, accountID)

	case "tiktok":

		return s.ProcessTikTokWork(ctx, accountID)
	case "twitter":

		return s.ProcessTwitterWork(ctx, accountID)
	case "youtube":
		return s.ProcessYouTubeWork(ctx, accountID)
	case "pinterest":
		return s.ProcessPinterestWork(ctx, accountID)
	case "gmb":
		return s.ProcessGMBWork(ctx, accountID)

	case "meta_ads":
		return s.ProcessMetaAdsWork(ctx, accountID)

	default:
		return fmt.Errorf("APIServer.ProcessImmediateWork: unsupported channel: %s", req.Channel)
	}
}

// ProcessFacebookWork handles Facebook account work orders
func (s *APIServer) ProcessFacebookWork(ctx context.Context, accountID primitive.ObjectID) error {
	opts := getImmediateWorkOptions(ctx)
	// Fetch account from MongoDB

	account, err := s.UnifiedRepo.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessFacebookWork: failed to fetch Facebook account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("APIServer.ProcessFacebookWork: Facebook account not found: %s", accountID.Hex())
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("platform_identifier", account.PlatformIdentifier).
		Str("type", account.Type).
		Msg("Found Facebook account")

	// Decrypt access token if needed
	accessToken := ""

	token := account.GetAccessToken()
	if token != "" {
		decrypted, err := crypto.DecryptToken(token, s.Config.DecryptionKey)
		if err != nil {
			s.Logger.Warn().Err(err).Msg("Failed to decrypt long access token, falling back to regular token")
			accessToken = getStringFromExtraData(account.ExtraData, "access_token")
		} else {
			accessToken = decrypted
		}
	} else {
		accessToken = getStringFromExtraData(account.ExtraData, "access_token")
	}

	if accessToken == "" {
		return fmt.Errorf("APIServer.ProcessFacebookWork: no valid access token available for account")
	}

	// Prepare work order
	workOrder := kafakaModels.ImmediateWorkOrder{
		ID:              accountID.Hex(),
		AccountID:       account.PlatformIdentifier,
		Type:            account.Type,
		AccessToken:     accessToken,
		WorkspaceID:     account.WorkspaceID.Hex(),
		LongAccessToken: token,
		SyncType:        "immediate",
		StartDate:       opts.StartDate,
		EndDate:         opts.EndDate,
	}

	// Marshal work order
	data, err := json.Marshal(workOrder)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessFacebookWork: failed to marshal work order: %w", err)
	}

	// Send to Kafka
	topic := "immediate-work-order-facebook"
	if err := s.Producer.Produce(ctx, topic, []byte(accountID.Hex()), data); err != nil {
		return fmt.Errorf("APIServer.ProcessFacebookWork: failed to send to Kafka: %w", err)
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("topic", topic).
		Msg("Successfully dispatched Facebook work order")

	return nil
}

// ProcessInstagramWork handles Instagram account work orders
func (s *APIServer) ProcessInstagramWork(ctx context.Context, accountID primitive.ObjectID) error {
	opts := getImmediateWorkOptions(ctx)
	// Fetch account from MongoDB
	account, err := s.UnifiedRepo.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessInstagramWork: failed to fetch Instagram account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("APIServer.ProcessInstagramWork: Instagram account not found: %s", accountID.Hex())
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("platform_identifier", account.PlatformIdentifier).
		Str("type", account.Type).
		Msg("Found Instagram account")

	// Get access token
	accessToken := account.GetAccessToken()
	if accessToken == "" {
		accessToken = getStringFromExtraData(account.ExtraData, "access_token")
	}
	if accessToken == "" {
		return fmt.Errorf("APIServer.ProcessInstagramWork: no valid access token available for account")
	}

	// Prepare work order
	workOrder := kafakaModels.ImmediateWorkOrder{
		ID:                    accountID.Hex(),
		AccountID:             account.PlatformIdentifier,
		Type:                  account.Type,
		AccessToken:           accessToken,
		WorkspaceID:           account.WorkspaceID.Hex(),
		SyncType:              "immediate",
		ConnectedViaInstagram: getBoolFromExtraData(account.ExtraData, "connected_via_instagram"),
		StartDate:             opts.StartDate,
		EndDate:               opts.EndDate,
	}

	// Marshal work order
	data, err := json.Marshal(workOrder)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessInstagramWork: failed to marshal work order: %w", err)
	}

	// Send to Kafka
	topic := "immediate-work-order-instagram"
	if err := s.Producer.Produce(ctx, topic, []byte(accountID.Hex()), data); err != nil {
		return fmt.Errorf("APIServer.ProcessInstagramWork: failed to send to Kafka: %w", err)
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("topic", topic).
		Msg("Successfully dispatched Instagram work order")

	return nil
}

// ProcessLinkedinWork handles LinkedIn account work orders
func (s *APIServer) ProcessLinkedinWork(ctx context.Context, accountID primitive.ObjectID) error {
	opts := getImmediateWorkOptions(ctx)
	// Fetch account from MongoDB

	account, err := s.UnifiedRepo.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessLinkedinWork: failed to fetch LinkedIn account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("APIServer.ProcessLinkedinWork: LinkedIn account not found: %s", accountID.Hex())
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("platform_identifier", account.PlatformIdentifier).
		Str("type", account.Type).
		Msg("Found LinkedIn account")

	// Get access token
	accessToken := account.GetAccessToken()
	if accessToken == "" {
		return fmt.Errorf("APIServer.ProcessLinkedinWork: no valid access token available for account")
	}

	// Prepare work order
	workOrder := kafakaModels.ImmediateWorkOrder{
		ID:          accountID.Hex(),
		AccountID:   account.PlatformIdentifier,
		Type:        account.Type,
		AccessToken: accessToken,
		WorkspaceID: account.WorkspaceID.Hex(),
		SyncType:    "immediate",
		StartDate:   opts.StartDate,
		EndDate:     opts.EndDate,
	}

	// Marshal work order
	data, err := json.Marshal(workOrder)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessLinkedinWork: failed to marshal work order: %w", err)
	}

	// Send to Kafka
	topic := "immediate-work-order-linkedin"
	if err := s.Producer.Produce(ctx, topic, []byte(accountID.Hex()), data); err != nil {
		return fmt.Errorf("APIServer.ProcessLinkedinWork: failed to send to Kafka: %w", err)
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("topic", topic).
		Msg("Successfully dispatched LinkedIn work order")

	return nil
}

// ProcessMetaAdsWork handles Meta Ads account work orders
func (s *APIServer) ProcessMetaAdsWork(ctx context.Context, accountID primitive.ObjectID) error {
	opts := getImmediateWorkOptions(ctx)

	// Fetch account from MongoDB
	account, err := s.UnifiedRepo.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessMetaAdsWork: failed to fetch Meta Ads account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("APIServer.ProcessMetaAdsWork: Meta Ads account not found: %s", accountID.Hex())
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("platform_identifier", account.PlatformIdentifier).
		Str("type", account.Type).
		Msg("Found Meta Ads account")

	// Resolve access token - try decrypt long token first
	token := account.GetAccessToken()
	accessToken := ""
	if token != "" {
		decrypted, derr := crypto.DecryptToken(token, s.Config.DecryptionKey)
		if derr != nil {
			s.Logger.Warn().Err(derr).Msg("Failed to decrypt long access token, falling back to extra_data access_token")
			accessToken = getStringFromExtraData(account.ExtraData, "access_token")
		} else {
			accessToken = decrypted
		}
	} else {
		accessToken = getStringFromExtraData(account.ExtraData, "access_token")
	}

	if accessToken == "" {
		return fmt.Errorf("APIServer.ProcessMetaAdsWork: no valid access token available for account")
	}

	// Prepare work order
	workOrder := kafakaModels.MetaAdsWorkOrder{
		MongoID:            accountID.Hex(),
		PlatformIdentifier: account.PlatformIdentifier,
		AccountID:          account.PlatformIdentifier, // processor expects act_XXXX here
		AccessToken:        accessToken,
		LongAccessToken:    token,
		WorkspaceID:        account.WorkspaceID.Hex(),
		UserID:             account.GetUserIDHex(),
		SyncType:           "immediate",
		StartDate:          opts.StartDate,
		EndDate:            opts.EndDate,
	}

	// Marshal work order
	data, err := json.Marshal(workOrder)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessMetaAdsWork: failed to marshal work order: %w", err)
	}

	// Send to Kafka
	topic := "immediate-work-order-meta-ads"
	if err := s.Producer.Produce(ctx, topic, []byte(accountID.Hex()), data); err != nil {
		return fmt.Errorf("APIServer.ProcessMetaAdsWork: failed to send to Kafka: %w", err)
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("topic", topic).
		Msg("Successfully dispatched Meta Ads work order")

	return nil
}

// ProcessTikTokWork handles TikTok account work orders
func (s *APIServer) ProcessTikTokWork(ctx context.Context, accountID primitive.ObjectID) error {
	opts := getImmediateWorkOptions(ctx)
	// Fetch account from MongoDB
	account, err := s.UnifiedRepo.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessTikTokWork: failed to fetch TikTok account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("APIServer.ProcessTikTokWork: TikTok account not found: %s", accountID.Hex())
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("platform_identifier", account.PlatformIdentifier).
		Str("type", account.Type).
		Msg("Found TikTok account")

	// Get access token
	accessToken := account.GetAccessToken()
	if accessToken == "" {
		accessToken = getStringFromExtraData(account.ExtraData, "access_token")
	}
	if accessToken == "" {
		return fmt.Errorf("APIServer.ProcessTikTokWork: no valid access token available for account")
	}

	// Get refresh token
	refreshToken := account.RefreshToken
	if refreshToken == "" {
		refreshToken = getStringFromExtraData(account.ExtraData, "refresh_token")
	}
	if refreshToken == "" {
		return fmt.Errorf("APIServer.ProcessTikTokWork: no refresh token available for account")
	}

	// Get scopes
	scope := account.Scope
	if scope == "" {
		scope = getStringFromExtraData(account.ExtraData, "scope")
	}
	if scope == "" {
		return fmt.Errorf("APIServer.ProcessTikTokWork: no scopes available for account")
	}

	// Prepare unified work order for immediate sync
	// Note: The unified immediate processor expects unified format, not platform-specific
	unifiedWorkOrder := map[string]interface{}{
		"id":            accountID.Hex(),
		"platform":      "tiktok",
		"account_id":    account.PlatformIdentifier, // TikTok creator ID
		"type":          account.Type,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"workspace_id":  account.WorkspaceID.Hex(),
		"sync_type":     "immediate",
		"start_date":    opts.StartDate,
		"end_date":      opts.EndDate,
	}

	// Marshal unified work order
	data, err := json.Marshal(unifiedWorkOrder)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessTikTokWork: failed to marshal work order: %w", err)
	}

	// Send to Kafka
	topic := "immediate-work-order-tiktok"
	if err := s.Producer.Produce(ctx, topic, []byte(accountID.Hex()), data); err != nil {
		return fmt.Errorf("APIServer.ProcessTikTokWork: failed to send to Kafka: %w", err)
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("platform_identifier", account.PlatformIdentifier).
		Str("topic", topic).
		Msg("Successfully dispatched TikTok work order")

	return nil
}

// ProcessTwitterWork handles Twitter account work orders
func (s *APIServer) ProcessTwitterWork(ctx context.Context, accountID primitive.ObjectID) error {
	opts := getImmediateWorkOptions(ctx)
	// Fetch account from MongoDB
	account, err := s.UnifiedRepo.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessTwitterWork: failed to fetch Twitter account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("APIServer.ProcessTwitterWork: Twitter account not found: %s", accountID.Hex())
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("platform_identifier", account.PlatformIdentifier).
		Str("type", account.Type).
		Msg("Found Twitter account")

	// Get OAuth token
	oauthToken := account.OAuthToken
	if oauthToken == "" {
		oauthToken = getStringFromExtraData(account.ExtraData, "oauth_token")
	}
	if oauthToken == "" {
		return fmt.Errorf("APIServer.ProcessTwitterWork: no valid access token available for account")
	}

	// Get OAuth token secret
	oauthTokenSecret := account.OAuthTokenSecret
	if oauthTokenSecret == "" {
		oauthTokenSecret = getStringFromExtraData(account.ExtraData, "oauth_token_secret")
	}
	if oauthTokenSecret == "" {
		return fmt.Errorf("APIServer.ProcessTwitterWork: no oauth_token_secret available for account")
	}

	if strings.TrimSpace(account.DeveloperAppID) == "" {
		s.Logger.Warn().
			Str("account_id", accountID.Hex()).
			Str("platform_identifier", account.PlatformIdentifier).
			Msg("Skipping immediate Twitter work order: missing developer_app_id")
		return nil
	}

	db := s.getMongoDatabase()
	if db == nil {
		s.Logger.Warn().
			Str("account_id", accountID.Hex()).
			Msg("Skipping immediate Twitter work order: mongo database unavailable for developer app lookup")
		return nil
	}
	twitterRepo := mongodb.NewTwitterRepository(db)
	app, err := twitterRepo.GetAnalyticsEnabledDeveloperAppByID(ctx, account.DeveloperAppID)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessTwitterWork: failed to fetch developer app: %w", err)
	}
	if app == nil {
		s.Logger.Warn().
			Str("account_id", accountID.Hex()).
			Str("platform_identifier", account.PlatformIdentifier).
			Str("developer_app_id", account.DeveloperAppID).
			Msg("Skipping immediate Twitter work order: non-null developer_app_id has no matching analytics-enabled developer app")
		return nil
	}

	// Prepare work order
	workOrder := kafakaModels.TwitterAccountWorkOrder{
		ID:               accountID.Hex(),
		WorkspaceID:      account.WorkspaceID.Hex(),
		TwitterID:        account.PlatformIdentifier,
		OAuthToken:       oauthToken,
		OAuthTokenSecret: oauthTokenSecret,
		NTweets:          opts.NTweets,
		APIKey:           app.APIKey,
		APISecret:        app.APISecret,
		AppName:          app.AppName,
		AppID:            app.ID.Hex(),
		ExecutedBy:       "internal",
		SyncType:         "full_sync",
	}

	// Marshal work order
	data, err := json.Marshal(workOrder)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessTwitterWork: failed to marshal work order: %w", err)
	}

	// Send to Kafka
	topic := "immediate-work-order-twitter"
	if err := s.Producer.Produce(ctx, topic, []byte(accountID.Hex()), data); err != nil {
		return fmt.Errorf("APIServer.ProcessTwitterWork: failed to send to Kafka: %w", err)
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("platform_identifier", account.PlatformIdentifier).
		Str("topic", topic).
		Msg("Successfully dispatched Twitter work order")

	return nil
}

func (s *APIServer) getMongoDatabase() *mongo.Database {
	if s.MongoClient == nil || s.Config == nil || s.Config.Mongo.Database == "" {
		return nil
	}
	return s.MongoClient.Database(s.Config.Mongo.Database)
}

// ProcessYouTubeWork handles YouTube account work orders
func (s *APIServer) ProcessYouTubeWork(ctx context.Context, accountID primitive.ObjectID) error {
	opts := getImmediateWorkOptions(ctx)
	// Fetch account from MongoDB
	account, err := s.UnifiedRepo.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessYouTubeWork: failed to fetch YouTube account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("APIServer.ProcessYouTubeWork: YouTube account not found: %s", accountID.Hex())
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("platform_identifier", account.PlatformIdentifier).
		Str("type", account.Type).
		Msg("Found YouTube account")

	// Check YouTube consent time - must be within the last 30 days
	if err := s.validateYouTubeConsent(account); err != nil {
		return err
	}

	// Get access token (may be encrypted)
	accessToken := account.GetAccessToken()
	if accessToken == "" {
		accessToken = getStringFromExtraData(account.ExtraData, "access_token")
	}
	if accessToken == "" {
		return fmt.Errorf("APIServer.ProcessYouTubeWork: no valid access token available for account")
	}

	// Get refresh token for YouTube OAuth (may be encrypted)
	// First check account.RefreshToken (extracted from embedded access_token document)
	// Then fall back to ExtraData fields
	refreshToken := account.RefreshToken
	if refreshToken == "" {
		refreshToken = getStringFromExtraData(account.ExtraData, "refresh_token")
	}
	if refreshToken == "" {
		refreshToken = getStringFromExtraData(account.ExtraData, "refreshToken")
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Bool("has_refresh_token", refreshToken != "").
		Msg("YouTube token info")

	// Prepare work order
	workOrder := kafakaModels.ImmediateWorkOrder{
		ID:           accountID.Hex(),
		AccountID:    account.PlatformIdentifier,
		Type:         account.Type,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		WorkspaceID:  account.WorkspaceID.Hex(),
		SyncType:     "immediate",
		StartDate:    opts.StartDate,
		EndDate:      opts.EndDate,
	}

	// Marshal work order
	data, err := json.Marshal(workOrder)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessYouTubeWork: failed to marshal work order: %w", err)
	}

	// Send to Kafka
	topic := "immediate-work-order-youtube"
	if err := s.Producer.Produce(ctx, topic, []byte(accountID.Hex()), data); err != nil {
		return fmt.Errorf("APIServer.ProcessYouTubeWork: failed to send to Kafka: %w", err)
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("topic", topic).
		Msg("Successfully dispatched YouTube work order")

	return nil
}

// ProcessPinterestWork handles Pinterest account work orders
func (s *APIServer) ProcessPinterestWork(ctx context.Context, accountID primitive.ObjectID) error {
	opts := getImmediateWorkOptions(ctx)
	// Fetch account from MongoDB
	account, err := s.UnifiedRepo.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessPinterestWork: failed to fetch Pinterest account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("APIServer.ProcessPinterestWork: Pinterest account not found: %s", accountID.Hex())
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("platform_identifier", account.PlatformIdentifier).
		Str("type", account.Type).
		Msg("Found Pinterest account")

	// Get access token - do NOT decrypt here, let the unified processor handle decryption
	// This is consistent with how Instagram and LinkedIn work
	accessToken := account.GetAccessToken()
	if accessToken == "" {
		accessToken = getStringFromExtraData(account.ExtraData, "access_token")
	}
	if accessToken == "" {
		return fmt.Errorf("APIServer.ProcessPinterestWork: no valid access token available for account")
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Int("token_length", len(accessToken)).
		Msg("Pinterest token retrieved (encrypted - will be decrypted by processor)")

	// Get account type and board ID if applicable
	accountType := getStringFromExtraData(account.ExtraData, "account_type")
	if accountType == "" {
		accountType = "profile" // default to profile
	}
	boardID := getStringFromExtraData(account.ExtraData, "board_id")

	// Prepare work order
	workOrder := kafakaModels.PinterestAccountWorkOrder{
		ID:          accountID.Hex(),
		AccountID:   account.PlatformIdentifier,
		AccessToken: accessToken,
		AccountType: accountType,
		BoardID:     boardID,
		WorkspaceID: account.WorkspaceID.Hex(),
		SyncType:    "immediate",
		StartDate:   opts.StartDate,
		EndDate:     opts.EndDate,
	}

	// Marshal work order
	data, err := json.Marshal(workOrder)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessPinterestWork: failed to marshal work order: %w", err)
	}

	// Send to Kafka
	topic := "immediate-work-order-pinterest"
	if err := s.Producer.Produce(ctx, topic, []byte(accountID.Hex()), data); err != nil {
		return fmt.Errorf("APIServer.ProcessPinterestWork: failed to send to Kafka: %w", err)
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("topic", topic).
		Msg("Successfully dispatched Pinterest work order")

	return nil
}

// ProcessGMBWork handles Google My Business account work orders
func (s *APIServer) ProcessGMBWork(ctx context.Context, accountID primitive.ObjectID) error {
	opts := getImmediateWorkOptions(ctx)
	// Fetch account from MongoDB
	account, err := s.UnifiedRepo.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessGMBWork: failed to fetch GMB account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("APIServer.ProcessGMBWork: GMB account not found: %s", accountID.Hex())
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("platform_identifier", account.PlatformIdentifier).
		Str("type", account.Type).
		Msg("Found GMB account")

	// Parse platform_identifier to extract GMB account ID and location ID
	// Format: "accounts/{accountID}/locations/{locationID}"
	gmbAccountID, locationID, err := parseGMBPlatformIdentifier(account.PlatformIdentifier)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessGMBWork: %w", err)
	}

	// Get access token - do NOT decrypt here, let the processor handle decryption
	accessToken := account.GetAccessToken()
	if accessToken == "" {
		accessToken = getStringFromExtraData(account.ExtraData, "access_token")
	}
	if accessToken == "" {
		return fmt.Errorf("APIServer.ProcessGMBWork: no valid access token available for account")
	}

	// Get refresh token
	refreshToken := account.RefreshToken
	if refreshToken == "" {
		refreshToken = getStringFromExtraData(account.ExtraData, "refresh_token")
	}
	if refreshToken == "" {
		return fmt.Errorf("APIServer.ProcessGMBWork: no refresh token available for account")
	}

	// Get account name and location name
	accountName := account.PlatformName
	if accountName == "" {
		accountName = getStringFromExtraData(account.ExtraData, "account_name")
	}
	locationName := getStringFromExtraData(account.ExtraData, "location_name")

	// Get language code (default to "en")
	languageCode := getStringFromExtraData(account.ExtraData, "language_code")
	if languageCode == "" {
		languageCode = "en"
	}

	// Prepare work order
	workOrder := kafakaModels.GMBAccountWorkOrder{
		ID:           accountID.Hex(),
		WorkspaceID:  account.WorkspaceID.Hex(),
		AccountID:    gmbAccountID,
		LocationID:   locationID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		AccountName:  accountName,
		LocationName: locationName,
		LanguageCode: languageCode,
		SyncType:     "immediate",
		StartDate:    opts.StartDate,
		EndDate:      opts.EndDate,
	}

	// Marshal work order
	data, err := json.Marshal(workOrder)
	if err != nil {
		return fmt.Errorf("APIServer.ProcessGMBWork: failed to marshal work order: %w", err)
	}

	// Send to Kafka
	topic := "immediate-work-order-gmb"
	if err := s.Producer.Produce(ctx, topic, []byte(accountID.Hex()), data); err != nil {
		return fmt.Errorf("APIServer.ProcessGMBWork: failed to send to Kafka: %w", err)
	}

	s.Logger.Info().
		Str("account_id", accountID.Hex()).
		Str("gmb_account_id", gmbAccountID).
		Str("location_id", locationID).
		Str("topic", topic).
		Msg("Successfully dispatched GMB work order")

	return nil
}

// parseGMBPlatformIdentifier parses a GMB platform identifier string.
// Format: "accounts/{accountID}/locations/{locationID}"
// Returns the account ID and location ID.
func parseGMBPlatformIdentifier(platformIdentifier string) (string, string, error) {
	parts := strings.Split(platformIdentifier, "/")
	if len(parts) != 4 || parts[0] != "accounts" || parts[2] != "locations" {
		return "", "", fmt.Errorf("invalid GMB platform_identifier format: %s (expected accounts/{id}/locations/{id})", platformIdentifier)
	}
	return parts[1], parts[3], nil
}

// Helper function to extract string from ExtraData
func getStringFromExtraData(extraData map[string]interface{}, key string) string {
	if val, ok := extraData[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

// getBoolFromExtraData safely extracts a boolean value from the ExtraData map.
// Returns false if the key is not found or the value is not a boolean.
func getBoolFromExtraData(extraData map[string]interface{}, key string) bool {
	if val, ok := extraData[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return false
}

func getIntFromExtraData(extraData map[string]interface{}, key string) int {
	if val, ok := extraData[key]; ok {
		return anyToInt(val)
	}
	return 0
}

func anyToInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return 0
}

// validateYouTubeConsent checks if the YouTube account has valid consent (within 30 days).
// Returns an error if consent is missing or expired.
func (s *APIServer) validateYouTubeConsent(account *mongomodels.SocialIntegration) error {
	if account.Preferences == nil {
		s.Logger.Warn().
			Str("account_id", account.ID.Hex()).
			Msg("YouTube account missing preferences, consent validation skipped")
		return fmt.Errorf("APIServer.validateYouTubeConsent: youtube consent expired: no consent time found")
	}

	consentTimeVal, exists := account.Preferences["last_youtube_consent_time"]
	if !exists {
		s.Logger.Warn().
			Str("account_id", account.ID.Hex()).
			Msg("YouTube account missing last_youtube_consent_time")
		return fmt.Errorf("APIServer.validateYouTubeConsent: youtube consent expired: no consent time found")
	}

	var consentTime time.Time
	switch v := consentTimeVal.(type) {
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			// Try other common formats
			parsed, err = time.Parse("2006-01-02T15:04:05.000Z", v)
			if err != nil {
				s.Logger.Error().
					Err(err).
					Str("account_id", account.ID.Hex()).
					Str("consent_time_raw", v).
					Msg("Failed to parse YouTube consent time")
				return fmt.Errorf("APIServer.validateYouTubeConsent: youtube consent expired: invalid consent time format")
			}
		}
		consentTime = parsed
	case time.Time:
		consentTime = v
	default:
		s.Logger.Error().
			Str("account_id", account.ID.Hex()).
			Interface("consent_time_type", consentTimeVal).
			Msg("YouTube consent time has unexpected type")
		return fmt.Errorf("APIServer.validateYouTubeConsent: youtube consent expired: invalid consent time type")
	}

	// Check if consent is within the allowed period
	consentCutoff := time.Now().UTC().AddDate(0, 0, -youtubeConsentMaxDays)
	if consentTime.Before(consentCutoff) {
		s.Logger.Warn().
			Str("account_id", account.ID.Hex()).
			Time("consent_time", consentTime).
			Time("cutoff_time", consentCutoff).
			Int("max_days", youtubeConsentMaxDays).
			Msg("YouTube consent has expired")
		return fmt.Errorf("APIServer.validateYouTubeConsent: youtube consent expired: consent is older than %d days", youtubeConsentMaxDays)
	}

	s.Logger.Debug().
		Str("account_id", account.ID.Hex()).
		Time("consent_time", consentTime).
		Msg("YouTube consent is valid")

	return nil
}

// loggingMiddleware logs all HTTP requests
func (s *APIServer) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(rw, r)

		// Log request details
		s.Logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Int("status", rw.statusCode).
			Dur("duration", time.Since(start)).
			Msg("HTTP request processed")
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// handleHealth handles health check requests
func (s *APIServer) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}
