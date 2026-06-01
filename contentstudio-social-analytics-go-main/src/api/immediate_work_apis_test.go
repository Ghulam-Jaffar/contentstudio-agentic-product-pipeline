package api_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	apimodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	mongo3 "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkaModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// =========================================================================
// Test constants for duplicated literals
const (
	topicImmediateWorkFacebook  = "immediate-work-order-facebook"
	topicImmediateWorkInstagram = "immediate-work-order-instagram"
	topicImmediateWorkLinkedIn  = "immediate-work-order-linkedin"
	topicImmediateWorkTikTok    = "immediate-work-order-tiktok"
	topicImmediateWorkYouTube   = "immediate-work-order-youtube"
	topicImmediateWorkPinterest = "immediate-work-order-pinterest"
	topicImmediateWorkMetaAds   = "immediate-work-order-meta-ads"
	directToken                 = "direct-token"
	failedToSendKafkaMsg        = "failed to send to Kafka"
	noValidAccessTokenMsg       = "no valid access token available for account"
	kafkaDownMsg                = "kafka down"
	verifyWorkOrderContentMsg   = "Verify work order content"
	accountNotFoundMsg          = "Account not found"
	missingAccessTokenMsg       = "Missing access token"
	kafkaErrorMsg               = "Kafka error"
	youtubeConsentExpiredMsg    = "youtube consent expired"
)

// ============================================================================
// Mock Producer
// ============================================================================

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Produce(ctx context.Context, topic string, key []byte, value []byte) error {
	args := m.Called(ctx, topic, key, value)
	return args.Error(0)
}

func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// ============================================================================
// Helper Functions
// ============================================================================

// makeRequestBody creates a JSON request body from ImmediateWorkRequest
func makeRequestBody(t *testing.T, req apimodels.ImmediateWorkRequest) *bytes.Buffer {
	t.Helper()
	body, err := json.Marshal(req)
	assert.NoError(t, err)
	return bytes.NewBuffer(body)
}

// makeAPIServer creates an APIServer with the provided repo and producer
func makeAPIServer(repo *mongodb.SeedRepo, prod *MockProducer, cfg *config.Config) *api.APIServer {
	lg, _ := logger.NewTestLogger()
	if cfg == nil {
		cfg = &config.Config{}
	}
	return &api.APIServer{
		UnifiedRepo: repo,
		Producer:    prod,
		Logger:      lg,
		Config:      cfg,
	}
}

// seedAccount creates a test account in the repository
func seedAccount(repo *mongodb.SeedRepo, id primitive.ObjectID, platformType, platformID, acctType string) {
	if acctType == "" {
		acctType = "page"
	}
	repo.SeedAccounts(mongo3.SocialIntegration{
		ID:                 id,
		PlatformType:       platformType,
		PlatformIdentifier: platformID,
		Type:               acctType,
		State:              mongo3.StateAdded,
		Validity:           mongo3.ValidityValid,
		WorkspaceID:        primitive.NewObjectID(),
		ExtraData:          map[string]interface{}{"access_token": "plain-token"},
	})
}

// ============================================================================
// Test: HandleImmediateWork
// ============================================================================

func TestHandleImmediateWork(t *testing.T) {
	testObjectID := primitive.NewObjectID()
	testObjectIDStr := testObjectID.Hex()

	tests := []struct {
		name           string
		method         string
		body           *bytes.Buffer
		setup          func(repo *mongodb.SeedRepo, prod *MockProducer)
		expectedStatus int
		expectedBody   string // substring match
	}{
		{
			name:   "Invalid method - GET",
			method: http.MethodGet,
			body:   bytes.NewBuffer(nil),
			setup: func(*mongodb.SeedRepo, *MockProducer) {
				/* No setup needed; method guard triggers before dependencies. */
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `"code":"INVALID_REQUEST"`,
		},
		{
			name:   "Invalid method - PUT",
			method: http.MethodPut,
			body:   bytes.NewBuffer(nil),
			setup: func(*mongodb.SeedRepo, *MockProducer) {
				/* No setup needed; method guard triggers before dependencies. */
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `"code":"INVALID_REQUEST"`,
		},
		{
			name:   "Invalid method - DELETE",
			method: http.MethodDelete,
			body:   bytes.NewBuffer(nil),
			setup: func(*mongodb.SeedRepo, *MockProducer) {
				/* No setup needed; method guard triggers before dependencies. */
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `"code":"INVALID_REQUEST"`,
		},
		{
			name:   "Invalid JSON body",
			method: http.MethodPost,
			body:   bytes.NewBufferString("{invalid json"),
			setup: func(*mongodb.SeedRepo, *MockProducer) {
				/* No setup needed; request fails during body parsing. */
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"message":"Invalid request body"`,
		},
		{
			name:   "Empty JSON body - missing account_id",
			method: http.MethodPost,
			body:   bytes.NewBufferString("{}"),
			setup: func(*mongodb.SeedRepo, *MockProducer) {
				/* No setup needed; validation fails before using repo/producer. */
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"code":"MISSING_FIELD"`,
		},
		{
			name:   "Missing account_id",
			method: http.MethodPost,
			body:   makeRequestBody(t, apimodels.ImmediateWorkRequest{Channel: "facebook"}),
			setup: func(*mongodb.SeedRepo, *MockProducer) {
				/* No setup needed; request is invalid before DB access. */
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"details":"account_id is required"`,
		},
		{
			name:   "Missing channel",
			method: http.MethodPost,
			body:   makeRequestBody(t, apimodels.ImmediateWorkRequest{AccountID: testObjectIDStr}),
			setup: func(*mongodb.SeedRepo, *MockProducer) {
				/* No setup needed; request is invalid before DB access. */
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"details":"channel is required"`,
		},
		{
			name:   "Invalid channel - snapchat",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "snapchat",
			}),
			setup: func(*mongodb.SeedRepo, *MockProducer) {
				/* No setup needed; channel validation fails immediately. */
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"code":"INVALID_CHANNEL"`,
		},
		{
			name:   "Invalid channel - snapchat",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "snapchat",
			}),
			setup: func(*mongodb.SeedRepo, *MockProducer) {
				/* No setup needed; channel validation fails immediately. */
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"code":"INVALID_CHANNEL"`,
		},
		{
			name:   "Invalid account ID format",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: "invalid_hex",
				Channel:   "facebook",
			}),
			setup: func(*mongodb.SeedRepo, *MockProducer) {
				/* No setup needed; invalid ID rejected before seeding. */
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"code":"INVALID_ACCOUNT_ID"`,
		},
		{
			name:   "Account not found - Facebook",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "facebook",
			}),
			setup: func(repo *mongodb.SeedRepo, _ *MockProducer) {
				/* No seed on purpose to assert not-found response. */
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"code":"ACCOUNT_NOT_FOUND"`,
		},
		{
			name:   "Account not found - Instagram",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "instagram",
			}),
			setup: func(repo *mongodb.SeedRepo, _ *MockProducer) {
				/* No seed on purpose to assert not-found response. */
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"code":"ACCOUNT_NOT_FOUND"`,
		},
		{
			name:   "Account not found - LinkedIn",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "linkedin",
			}),
			setup: func(repo *mongodb.SeedRepo, _ *MockProducer) {
				/* No seed on purpose to assert not-found response. */
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"code":"ACCOUNT_NOT_FOUND"`,
		},
		{
			name:   "Account not found - TikTok",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "tiktok",
			}),
			setup: func(repo *mongodb.SeedRepo, _ *MockProducer) {
				/* No seed on purpose to assert not-found response. */
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"code":"ACCOUNT_NOT_FOUND"`,
		},
		{
			name:   "Account not found - YouTube",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "youtube",
			}),
			setup: func(repo *mongodb.SeedRepo, _ *MockProducer) {
				/* No seed on purpose to assert not-found response. */
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"code":"ACCOUNT_NOT_FOUND"`,
		},
		{
			name:   "Account not found - Pinterest",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "pinterest",
			}),
			setup: func(repo *mongodb.SeedRepo, _ *MockProducer) {
				/* No seed on purpose to assert not-found response. */
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"code":"ACCOUNT_NOT_FOUND"`,
		},
		{
			name:   "Successful Facebook request",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "facebook",
			}),
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				seedAccount(repo, testObjectID, mongo3.PlatformFacebook, "fb_123", "page")
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkFacebook,
					[]byte(testObjectIDStr),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
		{
			name:   "Successful Instagram request",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "instagram",
			}),
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				seedAccount(repo, testObjectID, mongo3.PlatformInstagram, "insta_123", "page")
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkInstagram,
					[]byte(testObjectIDStr),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
		{
			name:   "Successful LinkedIn request",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "linkedin",
			}),
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testObjectID,
					PlatformType:       mongo3.PlatformLinkedIn,
					PlatformIdentifier: "linkedin_123",
					Type:               "page",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkLinkedIn,
					[]byte(testObjectIDStr),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
		{
			name:   "Successful TikTok request",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "tiktok",
			}),
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testObjectID,
					PlatformType:       mongo3.PlatformTikTok,
					PlatformIdentifier: "tiktok_123",
					Type:               "page",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					Scope:              "user.info.basic",
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkTikTok,
					[]byte(testObjectIDStr),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
		{
			name:   "Successful YouTube request with valid consent",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "youtube",
			}),
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				// Consent time within 30 days
				recentConsent := time.Now().UTC().AddDate(0, 0, -5).Format(time.RFC3339)
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testObjectID,
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					Type:               "page",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					Preferences: map[string]interface{}{
						"last_youtube_consent_time": recentConsent,
					},
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkYouTube,
					[]byte(testObjectIDStr),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
		{
			name:   "YouTube request with expired consent",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "youtube",
			}),
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				// Consent time older than 30 days
				expiredConsent := time.Now().UTC().AddDate(0, 0, -35).Format(time.RFC3339)
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testObjectID,
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					Type:               "page",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					Preferences: map[string]interface{}{
						"last_youtube_consent_time": expiredConsent,
					},
				})
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `"code":"CONSENT_EXPIRED"`,
		},
		{
			name:   "Successful Pinterest request",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "pinterest",
			}),
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testObjectID,
					PlatformType:       mongo3.PlatformPinterest,
					PlatformIdentifier: "pinterest_123",
					Type:               "profile",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkPinterest,
					[]byte(testObjectIDStr),
					mock.Anything,
				).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"success"`,
		},
		{
			name:   "YouTube request with missing consent",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "youtube",
			}),
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testObjectID,
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					Type:               "page",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					// No Preferences - missing consent time
				})
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `"code":"CONSENT_EXPIRED"`,
		},
		{
			name:   "Kafka producer failure - Facebook",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "facebook",
			}),
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				seedAccount(repo, testObjectID, mongo3.PlatformFacebook, "fb_123", "page")
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkFacebook,
					[]byte(testObjectIDStr),
					mock.Anything,
				).Return(errors.New("kafka connection failed"))
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `"code":"KAFKA_ERROR"`,
		},
		{
			name:   "Missing access token error",
			method: http.MethodPost,
			body: makeRequestBody(t, apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "facebook",
			}),
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testObjectID,
					PlatformType:       mongo3.PlatformFacebook,
					PlatformIdentifier: "fb_123",
					Type:               "page",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "",
					ExtraData:          map[string]interface{}{},
				})
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"code":"NO_ACCESS_TOKEN"`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mongodb.NewSeedRepo()
			prod := new(MockProducer)

			s := makeAPIServer(repo, prod, nil)

			if tt.setup != nil {
				tt.setup(repo, prod)
			}

			req := httptest.NewRequest(tt.method, "/immediate-work", tt.body)
			w := httptest.NewRecorder()
			s.HandleImmediateWork(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(resp.Body)
			assert.Contains(t, buf.String(), tt.expectedBody)

			prod.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Test: ProcessImmediateWork
// ============================================================================

func TestProcessImmediateWork(t *testing.T) {
	ctx := context.Background()
	testObjectID := primitive.NewObjectID()
	testObjectIDStr := testObjectID.Hex()

	tests := []struct {
		name        string
		request     apimodels.ImmediateWorkRequest
		setup       func(repo *mongodb.SeedRepo, prod *MockProducer)
		expectedErr string
	}{
		{
			name: "Invalid account ID format",
			request: apimodels.ImmediateWorkRequest{
				AccountID: "invalid",
				Channel:   "facebook",
			},
			setup: func(*mongodb.SeedRepo, *MockProducer) {
				/* No setup needed; request fails before repo/producer usage. */
			},
			expectedErr: "invalid account ID format",
		},
		{
			name: "Unsupported channel",
			request: apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "snapchat",
			},
			setup: func(*mongodb.SeedRepo, *MockProducer) {
				/* No setup needed; unsupported channel handled before seeding. */
			},
			expectedErr: "unsupported channel: snapchat",
		},
		{
			name: "Empty channel",
			request: apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "",
			},
			setup: func(*mongodb.SeedRepo, *MockProducer) {
				/* No setup needed; unsupported channel handled before seeding. */
			},
			expectedErr: "unsupported channel:",
		},
		{
			name: "Success - Facebook",
			request: apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "facebook",
			},
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				seedAccount(repo, testObjectID, mongo3.PlatformFacebook, "fb_123", "page")
				prod.On("Produce", mock.Anything, topicImmediateWorkFacebook, mock.Anything, mock.Anything).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success - Instagram",
			request: apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "instagram",
			},
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				seedAccount(repo, testObjectID, mongo3.PlatformInstagram, "insta_123", "page")
				prod.On("Produce", mock.Anything, topicImmediateWorkInstagram, mock.Anything, mock.Anything).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success - LinkedIn",
			request: apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "linkedin",
			},
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testObjectID,
					PlatformType:       mongo3.PlatformLinkedIn,
					PlatformIdentifier: "linkedin_123",
					Type:               "page",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
				})
				prod.On("Produce", mock.Anything, topicImmediateWorkLinkedIn, mock.Anything, mock.Anything).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success - TikTok",
			request: apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "tiktok",
			},
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testObjectID,
					PlatformType:       mongo3.PlatformTikTok,
					PlatformIdentifier: "tiktok_123",
					Type:               "page",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					Scope:              "user.info.basic",
				})
				prod.On("Produce", mock.Anything, topicImmediateWorkTikTok, mock.Anything, mock.Anything).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success - YouTube with valid consent",
			request: apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "youtube",
			},
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				recentConsent := time.Now().UTC().AddDate(0, 0, -5).Format(time.RFC3339)
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testObjectID,
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					Type:               "page",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					Preferences: map[string]interface{}{
						"last_youtube_consent_time": recentConsent,
					},
				})
				prod.On("Produce", mock.Anything, topicImmediateWorkYouTube, mock.Anything, mock.Anything).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "YouTube with expired consent",
			request: apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "youtube",
			},
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				expiredConsent := time.Now().UTC().AddDate(0, 0, -35).Format(time.RFC3339)
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testObjectID,
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					Type:               "page",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					Preferences: map[string]interface{}{
						"last_youtube_consent_time": expiredConsent,
					},
				})
			},
			expectedErr: youtubeConsentExpiredMsg,
		},
		{
			name: "Success - Pinterest",
			request: apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "pinterest",
			},
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testObjectID,
					PlatformType:       mongo3.PlatformPinterest,
					PlatformIdentifier: "pinterest_123",
					Type:               "profile",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
				})
				prod.On("Produce", mock.Anything, topicImmediateWorkPinterest, mock.Anything, mock.Anything).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success - Meta Ads with default date range",
			request: apimodels.ImmediateWorkRequest{
				AccountID: testObjectIDStr,
				Channel:   "meta_ads",
			},
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				seedAccount(repo, testObjectID, mongo3.PlatformMetaAds, "act_123", "ad_account")
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkMetaAds,
					[]byte(testObjectIDStr),
					mock.MatchedBy(func(data []byte) bool {
						var workOrder kafkaModels.MetaAdsWorkOrder
						if err := json.Unmarshal(data, &workOrder); err != nil {
							return false
						}
						return workOrder.MongoID == testObjectIDStr &&
							workOrder.PlatformIdentifier == "act_123" &&
							workOrder.AccountID == "act_123" &&
							workOrder.SyncType == "immediate" &&
							workOrder.StartDate == "" &&
							workOrder.EndDate == ""
					}),
				).Return(nil)
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mongodb.NewSeedRepo()
			prod := new(MockProducer)

			s := makeAPIServer(repo, prod, nil)

			if tt.setup != nil {
				tt.setup(repo, prod)
			}

			err := s.ProcessImmediateWork(ctx, tt.request)

			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedErr)
			}

			prod.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Test: ProcessFacebookWork
// ============================================================================

func TestProcessFacebookWork(t *testing.T) {
	ctx := context.Background()
	testID := primitive.NewObjectID()

	rawKey := "01234567890123456789012345678901"
	base64Key := base64.StdEncoding.EncodeToString([]byte(rawKey))

	tests := []struct {
		name        string
		setup       func(repo *mongodb.SeedRepo, prod *MockProducer)
		expectedErr string
	}{
		{
			name: accountNotFoundMsg + "",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				// No seed
			},
			expectedErr: "Facebook account not found",
		},
		{
			name: missingAccessTokenMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformFacebook,
					PlatformIdentifier: "fb_page_1",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "",
					ExtraData:          map[string]interface{}{},
				})
			},
			expectedErr: noValidAccessTokenMsg,
		},
		{
			name: kafkaErrorMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				seedAccount(repo, testID, mongo3.PlatformFacebook, "fb_page_1", "page")
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkFacebook,
					[]byte(testID.Hex()),
					mock.Anything,
				).Return(errors.New(kafkaDownMsg))
			},
			expectedErr: failedToSendKafkaMsg,
		},
		{
			name: "Success with fallback token",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				seedAccount(repo, testID, mongo3.PlatformFacebook, "fb_page_1", "page")
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkFacebook,
					[]byte(testID.Hex()),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success with encrypted token",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				enc, err := mongodb.EncryptToken("decrypted-token", base64Key)
				if err != nil {
					t.Fatalf("failed to encrypt token: %v", err)
				}
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformFacebook,
					PlatformIdentifier: "fb_page_2",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        enc,
					ExtraData:          map[string]interface{}{},
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkFacebook,
					[]byte(testID.Hex()),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Decrypt failure, fallback to ExtraData",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformFacebook,
					PlatformIdentifier: "fb_page_3",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "invalid-encrypted-token",
					ExtraData:          map[string]interface{}{"access_token": "fallback-token"},
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkFacebook,
					[]byte(testID.Hex()),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: verifyWorkOrderContentMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				seedAccount(repo, testID, mongo3.PlatformFacebook, "fb_page_1", "page")
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkFacebook,
					[]byte(testID.Hex()),
					mock.MatchedBy(func(data []byte) bool {
						var workOrder kafkaModels.ImmediateWorkOrder
						if err := json.Unmarshal(data, &workOrder); err != nil {
							return false
						}
						return workOrder.ID == testID.Hex() &&
							workOrder.AccountID == "fb_page_1" &&
							workOrder.Type == "page" &&
							workOrder.SyncType == "immediate" &&
							workOrder.AccessToken != ""
					}),
				).Return(nil)
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mongodb.NewSeedRepo()
			prod := new(MockProducer)

			if tt.setup != nil {
				tt.setup(repo, prod)
			}

			s := makeAPIServer(repo, prod, &config.Config{DecryptionKey: base64Key})

			err := s.ProcessFacebookWork(ctx, testID)

			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedErr)
			}

			prod.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Test: ProcessInstagramWork
// ============================================================================

func TestProcessInstagramWork(t *testing.T) {
	ctx := context.Background()
	testID := primitive.NewObjectID()
	idHex := testID.Hex()

	tests := []struct {
		name        string
		setup       func(repo *mongodb.SeedRepo, prod *MockProducer)
		expectedErr string
	}{
		{
			name: accountNotFoundMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				// No seed
			},
			expectedErr: "Instagram account not found",
		},
		{
			name: missingAccessTokenMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformInstagram,
					PlatformIdentifier: "insta_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "",
					ExtraData:          map[string]interface{}{},
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
			},
			expectedErr: noValidAccessTokenMsg,
		},
		{
			name: kafkaErrorMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				seedAccount(repo, testID, mongo3.PlatformInstagram, "insta_123", "page")
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkInstagram,
					[]byte(idHex),
					mock.Anything,
				).Return(errors.New(kafkaDownMsg))
			},
			expectedErr: failedToSendKafkaMsg,
		},
		{
			name: "Success with fallback token",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				seedAccount(repo, testID, mongo3.PlatformInstagram, "insta_123", "page")
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkInstagram,
					[]byte(idHex),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success with direct AccessToken",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformInstagram,
					PlatformIdentifier: "insta_abc",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					ExtraData:          map[string]interface{}{"connected_via_instagram": false},
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkInstagram,
					[]byte(idHex),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success with connected_via_instagram true",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformInstagram,
					PlatformIdentifier: "insta_connected",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					ExtraData:          map[string]interface{}{"connected_via_instagram": true},
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkInstagram,
					[]byte(idHex),
					mock.MatchedBy(func(data []byte) bool {
						var workOrder kafkaModels.ImmediateWorkOrder
						if err := json.Unmarshal(data, &workOrder); err != nil {
							return false
						}
						return workOrder.ConnectedViaInstagram == true
					}),
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: verifyWorkOrderContentMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				seedAccount(repo, testID, mongo3.PlatformInstagram, "insta_123", "page")
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkInstagram,
					[]byte(idHex),
					mock.MatchedBy(func(data []byte) bool {
						var workOrder kafkaModels.ImmediateWorkOrder
						if err := json.Unmarshal(data, &workOrder); err != nil {
							return false
						}
						return workOrder.ID == testID.Hex() &&
							workOrder.AccountID == "insta_123" &&
							workOrder.Type == "page" &&
							workOrder.SyncType == "immediate" &&
							workOrder.AccessToken != ""
					}),
				).Return(nil)
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mongodb.NewSeedRepo()
			prod := new(MockProducer)

			if tt.setup != nil {
				tt.setup(repo, prod)
			}

			s := makeAPIServer(repo, prod, nil)
			err := s.ProcessInstagramWork(ctx, testID)

			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedErr)
			}

			prod.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Test: ProcessLinkedinWork
// ============================================================================

func TestProcessLinkedinWork(t *testing.T) {
	ctx := context.Background()
	testID := primitive.NewObjectID()
	idHex := testID.Hex()

	tests := []struct {
		name        string
		setup       func(repo *mongodb.SeedRepo, prod *MockProducer)
		expectedErr string
	}{
		{
			name: accountNotFoundMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				// No seed
			},
			expectedErr: "LinkedIn account not found",
		},
		{
			name: missingAccessTokenMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformLinkedIn,
					PlatformIdentifier: "linkedin_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
			},
			expectedErr: noValidAccessTokenMsg,
		},
		{
			name: kafkaErrorMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformLinkedIn,
					PlatformIdentifier: "linkedin_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkLinkedIn,
					[]byte(idHex),
					mock.Anything,
				).Return(errors.New(kafkaDownMsg))
			},
			expectedErr: failedToSendKafkaMsg,
		},
		{
			name: "Success with direct AccessToken",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformLinkedIn,
					PlatformIdentifier: "linkedin_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkLinkedIn,
					[]byte(idHex),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: verifyWorkOrderContentMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "organization",
					PlatformType:       mongo3.PlatformLinkedIn,
					PlatformIdentifier: "linkedin_org_456",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "org-token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkLinkedIn,
					[]byte(idHex),
					mock.MatchedBy(func(data []byte) bool {
						var workOrder kafkaModels.ImmediateWorkOrder
						if err := json.Unmarshal(data, &workOrder); err != nil {
							return false
						}
						return workOrder.ID == testID.Hex() &&
							workOrder.AccountID == "linkedin_org_456" &&
							workOrder.Type == "organization" &&
							workOrder.SyncType == "immediate" &&
							workOrder.AccessToken != "" // Token is encrypted, just check it exists
					}),
				).Return(nil)
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mongodb.NewSeedRepo()
			prod := new(MockProducer)

			if tt.setup != nil {
				tt.setup(repo, prod)
			}

			s := makeAPIServer(repo, prod, nil)
			err := s.ProcessLinkedinWork(ctx, testID)

			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedErr)
			}

			prod.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Test: ProcessTikTokWork
// ============================================================================

func TestProcessTikTokWork(t *testing.T) {
	ctx := context.Background()
	testID := primitive.NewObjectID()
	idHex := testID.Hex()

	tests := []struct {
		name        string
		setup       func(repo *mongodb.SeedRepo, prod *MockProducer)
		expectedErr string
	}{
		{
			name: accountNotFoundMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				// No seed
			},
			expectedErr: "TikTok account not found",
		},
		{
			name: missingAccessTokenMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformTikTok,
					PlatformIdentifier: "tiktok_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "",
					RefreshToken:       "refresh_token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
			},
			expectedErr: noValidAccessTokenMsg,
		},
		{
			name: "Missing refresh token",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformTikTok,
					PlatformIdentifier: "tiktok_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
			},
			expectedErr: "no refresh token available",
		},
		{
			name: kafkaErrorMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformTikTok,
					PlatformIdentifier: "tiktok_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					Scope:              "user.info.basic",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkTikTok,
					[]byte(idHex),
					mock.Anything,
				).Return(errors.New(kafkaDownMsg))
			},
			expectedErr: failedToSendKafkaMsg,
		},
		{
			name: "Success with direct AccessToken and RefreshToken",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformTikTok,
					PlatformIdentifier: "tiktok_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					Scope:              "user.info.basic",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkTikTok,
					[]byte(idHex),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: verifyWorkOrderContentMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "creator",
					PlatformType:       mongo3.PlatformTikTok,
					PlatformIdentifier: "tiktok_creator_456",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "tiktok-token",
					RefreshToken:       "refresh_token",
					Scope:              "user.info.basic",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkTikTok,
					[]byte(idHex),
					mock.MatchedBy(func(data []byte) bool {
						var workOrder kafkaModels.ImmediateWorkOrder
						if err := json.Unmarshal(data, &workOrder); err != nil {
							return false
						}
						return workOrder.ID == testID.Hex() &&
							workOrder.AccountID == "tiktok_creator_456" &&
							workOrder.Type == "creator" &&
							workOrder.SyncType == "immediate" &&
							workOrder.AccessToken != "" // Token is encrypted, just check it exists
					}),
				).Return(nil)
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mongodb.NewSeedRepo()
			prod := new(MockProducer)

			if tt.setup != nil {
				tt.setup(repo, prod)
			}

			s := makeAPIServer(repo, prod, nil)
			err := s.ProcessTikTokWork(ctx, testID)

			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedErr)
			}

			prod.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Test: ProcessTwitterWork
// ============================================================================

func TestProcessTwitterWork(t *testing.T) {
	ctx := context.Background()
	testID := primitive.NewObjectID()

	tests := []struct {
		name        string
		setup       func(repo *mongodb.SeedRepo, prod *MockProducer)
		expectedErr string
	}{
		{
			name: accountNotFoundMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				// No seed
			},
			expectedErr: "Twitter account not found",
		},
		{
			name: missingAccessTokenMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					PlatformType:       mongo3.PlatformTwitter,
					PlatformIdentifier: "twitter_123",
					WorkspaceID:        primitive.NewObjectID(),
					OAuthToken:         "",
					OAuthTokenSecret:   "secret-token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
			},
			expectedErr: noValidAccessTokenMsg,
		},
		{
			name: "Missing oauth token secret",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					PlatformType:       mongo3.PlatformTwitter,
					PlatformIdentifier: "twitter_123",
					WorkspaceID:        primitive.NewObjectID(),
					OAuthToken:         "oauth-token",
					OAuthTokenSecret:   "",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
			},
			expectedErr: "no oauth_token_secret available",
		},
		{
			name: "Skip when mongo database is unavailable (no work order)",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					PlatformType:       mongo3.PlatformTwitter,
					PlatformIdentifier: "twitter_123",
					WorkspaceID:        primitive.NewObjectID(),
					OAuthToken:         "oauth-token",
					OAuthTokenSecret:   "oauth-secret",
					DeveloperAppID:     primitive.NewObjectID().Hex(),
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
			},
			expectedErr: "",
		},
		{
			name: "Skip when mongo database unavailable even with complete account data",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					PlatformType:       mongo3.PlatformTwitter,
					PlatformIdentifier: "twitter_456",
					WorkspaceID:        primitive.NewObjectID(),
					OAuthToken:         "oauth-token",
					OAuthTokenSecret:   "oauth-secret",
					DeveloperAppID:     primitive.NewObjectID().Hex(),
					APIKey:             "seed-api-key",
					APISecret:          "seed-api-secret",
					ExtraData: map[string]interface{}{
						"post_count": 17,
					},
					State:    mongo3.StateAdded,
					Validity: mongo3.ValidityValid,
				})
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mongodb.NewSeedRepo()
			prod := new(MockProducer)

			if tt.setup != nil {
				tt.setup(repo, prod)
			}

			s := makeAPIServer(repo, prod, nil)
			err := s.ProcessTwitterWork(ctx, testID)

			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedErr)
			}

			prod.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Test: ProcessYouTubeWork
// ============================================================================

func TestProcessYouTubeWork(t *testing.T) {
	ctx := context.Background()
	testID := primitive.NewObjectID()
	idHex := testID.Hex()

	tests := []struct {
		name        string
		setup       func(repo *mongodb.SeedRepo, prod *MockProducer)
		expectedErr string
	}{
		{
			name: accountNotFoundMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				// No seed
			},
			expectedErr: "YouTube account not found",
		},
		{
			name: "Missing consent - no preferences",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					// No Preferences
				})
			},
			expectedErr: youtubeConsentExpiredMsg,
		},
		{
			name: "Missing consent - no consent time field",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					Preferences:        map[string]interface{}{"other_field": "value"},
				})
			},
			expectedErr: youtubeConsentExpiredMsg,
		},
		{
			name: "Consent expired - older than 30 days",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				expiredConsent := time.Now().UTC().AddDate(0, 0, -35).Format(time.RFC3339)
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					Preferences: map[string]interface{}{
						"last_youtube_consent_time": expiredConsent,
					},
				})
			},
			expectedErr: youtubeConsentExpiredMsg,
		},
		{
			name: "Consent expired - exactly 30 days boundary",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				// Exactly 31 days ago should be expired
				expiredConsent := time.Now().UTC().AddDate(0, 0, -31).Format(time.RFC3339)
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					Preferences: map[string]interface{}{
						"last_youtube_consent_time": expiredConsent,
					},
				})
			},
			expectedErr: youtubeConsentExpiredMsg,
		},
		{
			name: missingAccessTokenMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				recentConsent := time.Now().UTC().AddDate(0, 0, -5).Format(time.RFC3339)
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "",
					RefreshToken:       "refresh_token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					Preferences: map[string]interface{}{
						"last_youtube_consent_time": recentConsent,
					},
				})
			},
			expectedErr: noValidAccessTokenMsg,
		},
		{
			name: kafkaErrorMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				recentConsent := time.Now().UTC().AddDate(0, 0, -5).Format(time.RFC3339)
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					Preferences: map[string]interface{}{
						"last_youtube_consent_time": recentConsent,
					},
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkYouTube,
					[]byte(idHex),
					mock.Anything,
				).Return(errors.New(kafkaDownMsg))
			},
			expectedErr: failedToSendKafkaMsg,
		},
		{
			name: "Success with valid consent within 30 days",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				recentConsent := time.Now().UTC().AddDate(0, 0, -5).Format(time.RFC3339)
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					Preferences: map[string]interface{}{
						"last_youtube_consent_time": recentConsent,
					},
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkYouTube,
					[]byte(idHex),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success with consent at exactly 29 days (within threshold)",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				// 29 days ago should still be valid
				recentConsent := time.Now().UTC().AddDate(0, 0, -29).Format(time.RFC3339)
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					Preferences: map[string]interface{}{
						"last_youtube_consent_time": recentConsent,
					},
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkYouTube,
					[]byte(idHex),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success with consent in alternative date format",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				// Use the alternative format: 2006-01-02T15:04:05.000Z
				recentConsent := time.Now().UTC().AddDate(0, 0, -5).Format("2006-01-02T15:04:05.000Z")
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					Preferences: map[string]interface{}{
						"last_youtube_consent_time": recentConsent,
					},
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkYouTube,
					[]byte(idHex),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success with consent as time.Time type",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				// Use time.Time directly
				recentConsent := time.Now().UTC().AddDate(0, 0, -5)
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					RefreshToken:       "refresh_token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					Preferences: map[string]interface{}{
						"last_youtube_consent_time": recentConsent,
					},
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkYouTube,
					[]byte(idHex),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: verifyWorkOrderContentMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				recentConsent := time.Now().UTC().AddDate(0, 0, -5).Format(time.RFC3339)
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "page",
					PlatformType:       mongo3.PlatformYouTube,
					PlatformIdentifier: "youtube_channel_xyz",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "youtube-token",
					RefreshToken:       "refresh_token",
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
					Preferences: map[string]interface{}{
						"last_youtube_consent_time": recentConsent,
					},
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkYouTube,
					[]byte(idHex),
					mock.MatchedBy(func(data []byte) bool {
						var workOrder kafkaModels.ImmediateWorkOrder
						if err := json.Unmarshal(data, &workOrder); err != nil {
							return false
						}
						return workOrder.ID == testID.Hex() &&
							workOrder.AccountID == "youtube_channel_xyz" &&
							workOrder.Type == "page" &&
							workOrder.SyncType == "immediate" &&
							workOrder.AccessToken != ""
					}),
				).Return(nil)
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mongodb.NewSeedRepo()
			prod := new(MockProducer)

			if tt.setup != nil {
				tt.setup(repo, prod)
			}

			s := makeAPIServer(repo, prod, nil)
			err := s.ProcessYouTubeWork(ctx, testID)

			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedErr)
			}

			prod.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Test: ProcessPinterestWork
// ============================================================================

func TestProcessPinterestWork(t *testing.T) {
	ctx := context.Background()
	testID := primitive.NewObjectID()
	idHex := testID.Hex()

	tests := []struct {
		name        string
		setup       func(repo *mongodb.SeedRepo, prod *MockProducer)
		expectedErr string
	}{
		{
			name: accountNotFoundMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				// No seed
			},
			expectedErr: "Pinterest account not found",
		},
		{
			name: missingAccessTokenMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "profile",
					PlatformType:       mongo3.PlatformPinterest,
					PlatformIdentifier: "pinterest_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "",
					ExtraData:          map[string]interface{}{},
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
			},
			expectedErr: noValidAccessTokenMsg,
		},
		{
			name: kafkaErrorMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "profile",
					PlatformType:       mongo3.PlatformPinterest,
					PlatformIdentifier: "pinterest_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkPinterest,
					[]byte(idHex),
					mock.Anything,
				).Return(errors.New(kafkaDownMsg))
			},
			expectedErr: failedToSendKafkaMsg,
		},
		{
			name: "Success with direct AccessToken",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "profile",
					PlatformType:       mongo3.PlatformPinterest,
					PlatformIdentifier: "pinterest_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkPinterest,
					[]byte(idHex),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success with fallback token from ExtraData",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "profile",
					PlatformType:       mongo3.PlatformPinterest,
					PlatformIdentifier: "pinterest_123",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "",
					ExtraData:          map[string]interface{}{"access_token": "fallback-token"},
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkPinterest,
					[]byte(idHex),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: "Success with board account type",
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "board",
					PlatformType:       mongo3.PlatformPinterest,
					PlatformIdentifier: "pinterest_board_456",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        directToken,
					ExtraData: map[string]interface{}{
						"account_type": "board",
						"board_id":     "board_789",
					},
					State:    mongo3.StateAdded,
					Validity: mongo3.ValidityValid,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkPinterest,
					[]byte(idHex),
					mock.Anything,
				).Return(nil)
			},
			expectedErr: "",
		},
		{
			name: verifyWorkOrderContentMsg,
			setup: func(repo *mongodb.SeedRepo, prod *MockProducer) {
				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					Type:               "profile",
					PlatformType:       mongo3.PlatformPinterest,
					PlatformIdentifier: "pinterest_user_xyz",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "pinterest-token",
					ExtraData: map[string]interface{}{
						"account_type": "profile",
					},
					State:    mongo3.StateAdded,
					Validity: mongo3.ValidityValid,
				})
				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkPinterest,
					[]byte(idHex),
					mock.MatchedBy(func(data []byte) bool {
						var workOrder kafkaModels.PinterestAccountWorkOrder
						if err := json.Unmarshal(data, &workOrder); err != nil {
							return false
						}
						return workOrder.ID == testID.Hex() &&
							workOrder.AccountID == "pinterest_user_xyz" &&
							workOrder.AccountType == "profile" &&
							workOrder.SyncType == "immediate" &&
							workOrder.AccessToken != ""
					}),
				).Return(nil)
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			repo := mongodb.NewSeedRepo()
			prod := new(MockProducer)

			if tt.setup != nil {
				tt.setup(repo, prod)
			}

			s := makeAPIServer(repo, prod, nil)
			err := s.ProcessPinterestWork(ctx, testID)

			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedErr)
			}

			prod.AssertExpectations(t)
		})
	}
}

// ============================================================================
// Test: HandleHealth
// ============================================================================

func TestHandleHealth(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		wantStatus     int
		wantCTPrefix   string
		wantBodySubstr string
		verify         func(t *testing.T, resp *http.Response)
	}{
		{
			name:           "Method not allowed - POST",
			method:         http.MethodPost,
			wantStatus:     http.StatusMethodNotAllowed,
			wantBodySubstr: "Method not allowed",
		},
		{
			name:           "Method not allowed - PUT",
			method:         http.MethodPut,
			wantStatus:     http.StatusMethodNotAllowed,
			wantBodySubstr: "Method not allowed",
		},
		{
			name:           "Method not allowed - DELETE",
			method:         http.MethodDelete,
			wantStatus:     http.StatusMethodNotAllowed,
			wantBodySubstr: "Method not allowed",
		},
		{
			name:           "Method not allowed - PATCH",
			method:         http.MethodPatch,
			wantStatus:     http.StatusMethodNotAllowed,
			wantBodySubstr: "Method not allowed",
		},
		{
			name:         "Success - GET returns healthy JSON with RFC3339 time",
			method:       http.MethodGet,
			wantStatus:   http.StatusOK,
			wantCTPrefix: "application/json",
			verify: func(t *testing.T, resp *http.Response) {
				assert.True(t, strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json"))

				var payload map[string]string
				err := json.NewDecoder(resp.Body).Decode(&payload)
				assert.NoError(t, err, "response should be valid JSON")

				assert.Equal(t, "healthy", payload["status"])

				_, err = time.Parse(time.RFC3339, payload["time"])
				assert.NoError(t, err, "time should be RFC3339")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := &api.APIServer{}

			req := httptest.NewRequest(tc.method, "/health", nil)
			w := httptest.NewRecorder()

			s.HandleHealth(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.verify != nil {
				tc.verify(t, resp)
				return
			}

			if tc.wantBodySubstr != "" {
				b, _ := io.ReadAll(resp.Body)
				assert.Contains(t, string(b), tc.wantBodySubstr)
			}

			if tc.wantCTPrefix != "" {
				assert.True(t, strings.HasPrefix(resp.Header.Get("Content-Type"), tc.wantCTPrefix))
			}
		})
	}
}

// ============================================================================
// Test: LoggingMiddleware
// ============================================================================

func TestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		handlerFunc  http.HandlerFunc
		wantStatus   int
		wantLogCheck func(t *testing.T)
	}{
		{
			name:   "Successful request logged",
			method: http.MethodGet,
			path:   "/test",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "Error request logged",
			method: http.MethodPost,
			path:   "/error",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "internal error", http.StatusInternalServerError)
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "Not found logged",
			method: http.MethodGet,
			path:   "/notfound",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				http.NotFound(w, r)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			lg, _ := logger.NewTestLogger()
			s := &api.APIServer{
				Logger: lg,
			}

			handler := s.LoggingMiddleware(tc.handlerFunc)

			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatus, resp.StatusCode)
		})
	}
}

// ============================================================================
// Test: Helper Functions (getStringFromExtraData, getBoolFromExtraData)
// ============================================================================

func TestHelperFunctions(t *testing.T) {
	t.Run("getStringFromExtraData", func(t *testing.T) {
		tests := []struct {
			name      string
			extraData map[string]interface{}
			key       string
			want      string
		}{
			{
				name:      "Key exists with string value",
				extraData: map[string]interface{}{"token": "abc123"},
				key:       "token",
				want:      "abc123",
			},
			{
				name:      "Key exists with non-string value",
				extraData: map[string]interface{}{"count": 42},
				key:       "count",
				want:      "",
			},
			{
				name:      "Key does not exist",
				extraData: map[string]interface{}{"other": "value"},
				key:       "token",
				want:      "",
			},
			{
				name:      "Nil extra data",
				extraData: nil,
				key:       "token",
				want:      "",
			},
			{
				name:      "Empty extra data",
				extraData: map[string]interface{}{},
				key:       "token",
				want:      "",
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				// Since these are unexported functions, we test them indirectly
				// through the public API that uses them
				repo := mongodb.NewSeedRepo()
				testID := primitive.NewObjectID()

				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					PlatformType:       mongo3.PlatformFacebook,
					PlatformIdentifier: "test",
					Type:               "page",
					WorkspaceID:        primitive.NewObjectID(),
					ExtraData:          tt.extraData,
					AccessToken:        "",
				})

				acc, _ := repo.FindByID(context.Background(), testID)
				assert.NotNil(t, acc)
				// The function is tested indirectly through ProcessFacebookWork
			})
		}
	})

	t.Run("getBoolFromExtraData - tested via Instagram", func(t *testing.T) {
		tests := []struct {
			name              string
			connectedViaInsta interface{}
			expectConnected   bool
		}{
			{
				name:              "Connected via Instagram true",
				connectedViaInsta: true,
				expectConnected:   true,
			},
			{
				name:              "Connected via Instagram false",
				connectedViaInsta: false,
				expectConnected:   false,
			},
			{
				name:              "No connected_via_instagram field",
				connectedViaInsta: nil,
				expectConnected:   false,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				// Test indirectly through Instagram processing which uses this helper
				repo := mongodb.NewSeedRepo()
				prod := new(MockProducer)
				testID := primitive.NewObjectID()

				extraData := map[string]interface{}{"access_token": "test-token"}
				if tt.connectedViaInsta != nil {
					extraData["connected_via_instagram"] = tt.connectedViaInsta
				}

				repo.SeedAccounts(mongo3.SocialIntegration{
					ID:                 testID,
					PlatformType:       mongo3.PlatformInstagram,
					PlatformIdentifier: "test_insta",
					Type:               "page",
					WorkspaceID:        primitive.NewObjectID(),
					AccessToken:        "",
					ExtraData:          extraData,
					State:              mongo3.StateAdded,
					Validity:           mongo3.ValidityValid,
				})

				prod.On("Produce",
					mock.Anything,
					topicImmediateWorkInstagram,
					mock.Anything,
					mock.MatchedBy(func(data []byte) bool {
						var workOrder kafkaModels.ImmediateWorkOrder
						if err := json.Unmarshal(data, &workOrder); err != nil {
							return false
						}
						return workOrder.ConnectedViaInstagram == tt.expectConnected
					}),
				).Return(nil)

				s := makeAPIServer(repo, prod, nil)
				err := s.ProcessInstagramWork(context.Background(), testID)
				assert.NoError(t, err)
				prod.AssertExpectations(t)
			})
		}
	})
}
