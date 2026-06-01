package fetcher

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mt "go.mongodb.org/mongo-driver/mongo/integration/mtest"

	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// mockProducer implements kafka.Producer for testing
type mockProducer struct {
	mu       sync.Mutex
	messages []producedMessage
	err      error
}

type producedMessage struct {
	topic string
	key   []byte
	value []byte
}

func (m *mockProducer) Produce(ctx context.Context, topic string, key, value []byte) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, producedMessage{topic: topic, key: key, value: value})
	return nil
}

func (m *mockProducer) Close() error {
	return nil
}

func (m *mockProducer) getMessages() []producedMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messages
}

// mockUnifiedSocialRepository implements mongodb.UnifiedSocialRepository for testing
type mockUnifiedSocialRepository struct {
	accounts     []mongomodels.SocialIntegration
	countErr     error
	paginatedErr error
	totalCount   int64
}

func (m *mockUnifiedSocialRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
	return nil, nil
}

func (m *mockUnifiedSocialRepository) GetByPlatformID(ctx context.Context, platformType, platformID string) (*mongomodels.SocialIntegration, error) {
	return nil, nil
}

func (m *mockUnifiedSocialRepository) GetValidAccounts(ctx context.Context, platformType string, accountTypes []string) ([]mongomodels.SocialIntegration, error) {
	return m.accounts, nil
}

func (m *mockUnifiedSocialRepository) GetAccountsByWorkspace(ctx context.Context, workspaceID primitive.ObjectID, platforms []string) ([]mongomodels.SocialIntegration, error) {
	return nil, nil
}

func (m *mockUnifiedSocialRepository) GetAccountsNeedingUpdate(ctx context.Context, platformType string, lastUpdateField string, hours int) ([]mongomodels.SocialIntegration, error) {
	return m.accounts, nil
}

func (m *mockUnifiedSocialRepository) GetAccountsNeedingUpdatePaginated(ctx context.Context, platformType string, accountTypes []string, hours int, skip, limit int64) ([]mongomodels.SocialIntegration, error) {
	if m.paginatedErr != nil {
		return nil, m.paginatedErr
	}
	if skip >= int64(len(m.accounts)) {
		return []mongomodels.SocialIntegration{}, nil
	}
	end := skip + limit
	if end > int64(len(m.accounts)) {
		end = int64(len(m.accounts))
	}
	return m.accounts[skip:end], nil
}

func (m *mockUnifiedSocialRepository) CountAccountsNeedingUpdate(ctx context.Context, platformType string, accountTypes []string, hours int) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	if m.totalCount > 0 {
		return m.totalCount, nil
	}
	return int64(len(m.accounts)), nil
}

func (m *mockUnifiedSocialRepository) GetAccountsNeedingUpdateByID(ctx context.Context, platformType string, accountTypes []string, hours int, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error) {
	if m.paginatedErr != nil {
		return nil, m.paginatedErr
	}
	// Filter by lastID
	var filtered []mongomodels.SocialIntegration
	for _, acc := range m.accounts {
		if lastID == primitive.NilObjectID || acc.ID.Hex() > lastID.Hex() {
			filtered = append(filtered, acc)
		}
	}
	// Apply limit
	if int64(len(filtered)) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}

func (m *mockUnifiedSocialRepository) Update(ctx context.Context, id primitive.ObjectID, updates primitive.M) error {
	return nil
}

func (m *mockUnifiedSocialRepository) UpdateAnalyticsTimestamp(ctx context.Context, id primitive.ObjectID, timestampType string, timestamp time.Time) error {
	return nil
}

func (m *mockUnifiedSocialRepository) UpdateTokens(ctx context.Context, id primitive.ObjectID, tokens map[string]string) error {
	return nil
}

func (m *mockUnifiedSocialRepository) UpdateState(ctx context.Context, id primitive.ObjectID, newState string) error {
	return nil
}

func (m *mockUnifiedSocialRepository) UpdateValidity(ctx context.Context, id primitive.ObjectID, newValidity string) error {
	return nil
}

func (m *mockUnifiedSocialRepository) Create(ctx context.Context, account *mongomodels.SocialIntegration) (primitive.ObjectID, error) {
	return primitive.NewObjectID(), nil
}

func (m *mockUnifiedSocialRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	return nil
}

func (m *mockUnifiedSocialRepository) GetYouTubeAccountsNeedingUpdatePaginated(ctx context.Context, hours int, consentDays int, skip, limit int64) ([]mongomodels.SocialIntegration, error) {
	if m.paginatedErr != nil {
		return nil, m.paginatedErr
	}
	if skip >= int64(len(m.accounts)) {
		return []mongomodels.SocialIntegration{}, nil
	}
	end := skip + limit
	if end > int64(len(m.accounts)) {
		end = int64(len(m.accounts))
	}
	return m.accounts[skip:end], nil
}

func (m *mockUnifiedSocialRepository) GetYouTubeAccountsNeedingUpdateByID(ctx context.Context, hours int, consentDays int, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error) {
	if m.paginatedErr != nil {
		return nil, m.paginatedErr
	}
	// Find starting index based on lastID
	startIdx := 0
	if lastID != primitive.NilObjectID {
		for i, acc := range m.accounts {
			if acc.ID == lastID {
				startIdx = i + 1
				break
			}
		}
	}
	if startIdx >= len(m.accounts) {
		return []mongomodels.SocialIntegration{}, nil
	}
	end := startIdx + int(limit)
	if end > len(m.accounts) {
		end = len(m.accounts)
	}
	return m.accounts[startIdx:end], nil
}

func (m *mockUnifiedSocialRepository) CountYouTubeAccountsNeedingUpdate(ctx context.Context, hours int, consentDays int) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	if m.totalCount > 0 {
		return m.totalCount, nil
	}
	return int64(len(m.accounts)), nil
}

func (m *mockUnifiedSocialRepository) GetValidAccountsByID(ctx context.Context, platformType string, accountTypes []string, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error) {
	return nil, nil
}

func (m *mockUnifiedSocialRepository) CountValidAccounts(ctx context.Context, platformType string, accountTypes []string) (int64, error) {
	return 0, nil
}

func (m *mockUnifiedSocialRepository) RecordProcessingError(ctx context.Context, id primitive.ObjectID, errorMessage string) error {
	return nil
}

func (m *mockUnifiedSocialRepository) ClearProcessingError(ctx context.Context, id primitive.ObjectID) error {
	return nil
}

func (m *mockUnifiedSocialRepository) InsertTwitterJobMetadata(ctx context.Context, payload mongodb.TwitterJobMetadataPayload) error {
	return nil
}

func (m *mockUnifiedSocialRepository) GetAccountsByPlatformIDs(ctx context.Context, platformType string, platformIDs []string) ([]mongomodels.SocialIntegration, error) {
	return nil, nil
}

func testLogger() zerolog.Logger {
	return zerolog.Nop()
}

// Test data helpers
func createTestFacebookAccount(id, fbID string, hasToken bool, hasWorkspace bool) mongomodels.SocialIntegration {
	acc := mongomodels.SocialIntegration{
		ID:                 primitive.NewObjectID(),
		PlatformType:       mongomodels.PlatformFacebook,
		PlatformIdentifier: fbID,
		Type:               "Page",
		State:              mongomodels.StateAdded,
		Validity:           mongomodels.ValidityValid,
	}
	if hasWorkspace {
		acc.WorkspaceID = primitive.NewObjectID()
	}
	if hasToken {
		acc.LongAccessToken = "test_long_token_" + fbID
		acc.ExtraData = map[string]interface{}{
			"access_token": "test_token_" + fbID,
		}
	}
	return acc
}

func createTestLinkedinAccount(id, linkedinID string, accountType string, hasToken bool, hasWorkspace bool) mongomodels.SocialIntegration {
	acc := mongomodels.SocialIntegration{
		ID:                 primitive.NewObjectID(),
		PlatformType:       mongomodels.PlatformLinkedIn,
		PlatformIdentifier: linkedinID,
		Type:               accountType,
		State:              mongomodels.StateAdded,
		Validity:           mongomodels.ValidityValid,
	}
	if hasWorkspace {
		acc.WorkspaceID = primitive.NewObjectID()
	}
	if hasToken {
		// LinkedIn uses AccessToken (not LongAccessToken)
		acc.AccessToken = "test_linkedin_token_" + linkedinID
	}
	return acc
}

func createTestInstagramAccount(id, igID string, connectedViaIG bool, hasToken bool, hasWorkspace bool) mongomodels.SocialIntegration {
	acc := mongomodels.SocialIntegration{
		ID:                 primitive.NewObjectID(),
		PlatformType:       mongomodels.PlatformInstagram,
		PlatformIdentifier: igID,
		Type:               "business",
		State:              mongomodels.StateAdded,
		Validity:           mongomodels.ValidityValid,
	}
	if hasWorkspace {
		acc.WorkspaceID = primitive.NewObjectID()
	}
	extraData := map[string]interface{}{
		"connected_via_instagram": connectedViaIG,
	}
	if hasToken {
		if connectedViaIG {
			acc.AccessToken = "ig_direct_token_" + igID
		} else {
			// FB-linked accounts use user_details or AccessToken
			acc.UserDetails = map[string]interface{}{
				"access_token": "fb_linked_token_" + igID,
			}
		}
		extraData["access_token"] = "extra_token_" + igID
	}
	acc.ExtraData = extraData
	return acc
}

// ============= Facebook Tests =============

func Test_buildFacebookAccountBatch_Table(t *testing.T) {
	cases := []struct {
		name            string
		accounts        []mongomodels.SocialIntegration
		syncType        string
		expectedBatch   int
		expectedSkipped int
	}{
		{
			name:            "empty accounts",
			accounts:        []mongomodels.SocialIntegration{},
			syncType:        "incremental",
			expectedBatch:   0,
			expectedSkipped: 0,
		},
		{
			name: "valid accounts with tokens",
			accounts: []mongomodels.SocialIntegration{
				createTestFacebookAccount("1", "fb_001", true, true),
				createTestFacebookAccount("2", "fb_002", true, true),
			},
			syncType:        "incremental",
			expectedBatch:   2,
			expectedSkipped: 0,
		},
		{
			name: "skip accounts without tokens",
			accounts: []mongomodels.SocialIntegration{
				createTestFacebookAccount("1", "fb_001", true, true),
				createTestFacebookAccount("2", "fb_002", false, true),
				createTestFacebookAccount("3", "fb_003", true, true),
			},
			syncType:        "full_sync",
			expectedBatch:   2,
			expectedSkipped: 1,
		},
		{
			name: "skip accounts without workspace",
			accounts: []mongomodels.SocialIntegration{
				createTestFacebookAccount("1", "fb_001", true, true),
				createTestFacebookAccount("2", "fb_002", true, false),
			},
			syncType:        "incremental",
			expectedBatch:   1,
			expectedSkipped: 1,
		},
		{
			name: "skip all invalid accounts",
			accounts: []mongomodels.SocialIntegration{
				createTestFacebookAccount("1", "fb_001", false, true),
				createTestFacebookAccount("2", "fb_002", true, false),
				createTestFacebookAccount("3", "fb_003", false, false),
			},
			syncType:        "incremental",
			expectedBatch:   0,
			expectedSkipped: 3,
		},
		{
			name: "mixed valid and invalid",
			accounts: []mongomodels.SocialIntegration{
				createTestFacebookAccount("1", "fb_001", true, true),
				createTestFacebookAccount("2", "fb_002", false, true),
				createTestFacebookAccount("3", "fb_003", true, false),
				createTestFacebookAccount("4", "fb_004", true, true),
			},
			syncType:        "incremental",
			expectedBatch:   2,
			expectedSkipped: 2,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			batch, skipped := buildFacebookAccountBatch(tc.accounts, tc.syncType, testLogger())

			if len(batch) != tc.expectedBatch {
				t.Fatalf("expected batch size %d, got %d", tc.expectedBatch, len(batch))
			}
			if skipped != tc.expectedSkipped {
				t.Fatalf("expected skipped %d, got %d", tc.expectedSkipped, skipped)
			}

			for _, item := range batch {
				if item.SyncType != tc.syncType {
					t.Fatalf("expected sync type %s, got %s", tc.syncType, item.SyncType)
				}
				if item.FacebookID == "" {
					t.Fatal("expected non-empty facebook_id")
				}
			}
		})
	}
}

func Test_processFacebookBatches_Table(t *testing.T) {
	cases := []struct {
		name             string
		accounts         []mongomodels.SocialIntegration
		countErr         error
		paginatedErr     error
		producerErr      error
		expectedMessages int
	}{
		{
			name:             "no accounts needing update",
			accounts:         []mongomodels.SocialIntegration{},
			expectedMessages: 0,
		},
		{
			name: "single batch of valid accounts",
			accounts: []mongomodels.SocialIntegration{
				createTestFacebookAccount("1", "fb_001", true, true),
				createTestFacebookAccount("2", "fb_002", true, true),
			},
			expectedMessages: 1,
		},
		{
			name: "multiple batches (simulated by setting limit)",
			accounts: func() []mongomodels.SocialIntegration {
				accs := make([]mongomodels.SocialIntegration, 250)
				for i := 0; i < 250; i++ {
					accs[i] = createTestFacebookAccount(string(rune(i)), "fb_"+string(rune(i)), true, true)
				}
				return accs
			}(),
			expectedMessages: 2,
		},
		{
			name:             "count error",
			accounts:         []mongomodels.SocialIntegration{},
			countErr:         errors.New("count failed"),
			expectedMessages: 0,
		},
		{
			name: "paginated fetch error",
			accounts: []mongomodels.SocialIntegration{
				createTestFacebookAccount("1", "fb_001", true, true),
			},
			paginatedErr:     errors.New("fetch failed"),
			expectedMessages: 0,
		},
		{
			name: "producer error",
			accounts: []mongomodels.SocialIntegration{
				createTestFacebookAccount("1", "fb_001", true, true),
			},
			producerErr:      errors.New("produce failed"),
			expectedMessages: 0,
		},
		{
			name: "all accounts skipped",
			accounts: []mongomodels.SocialIntegration{
				createTestFacebookAccount("1", "fb_001", false, true),
				createTestFacebookAccount("2", "fb_002", false, true),
			},
			expectedMessages: 0,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUnifiedSocialRepository{
				accounts:     tc.accounts,
				countErr:     tc.countErr,
				paginatedErr: tc.paginatedErr,
				totalCount:   int64(len(tc.accounts)),
			}
			producer := &mockProducer{err: tc.producerErr}

			processFacebookBatches(
				context.Background(),
				repo,
				producer,
				testLogger(),
				"incremental",
				[]string{"Page"},
			)

			messages := producer.getMessages()
			if len(messages) != tc.expectedMessages {
				t.Fatalf("expected %d messages, got %d", tc.expectedMessages, len(messages))
			}

			for _, msg := range messages {
				if msg.topic != topicFacebookBatch {
					t.Fatalf("expected topic %s, got %s", topicFacebookBatch, msg.topic)
				}

				var workOrder kafkamodels.FacebookBatchWorkOrder
				if err := json.Unmarshal(msg.value, &workOrder); err != nil {
					t.Fatalf("failed to unmarshal message: %v", err)
				}

				if workOrder.BatchID == "" {
					t.Fatal("expected non-empty batch_id")
				}
				if len(workOrder.Accounts) == 0 {
					t.Fatal("expected non-empty accounts in batch")
				}
			}
		})
	}
}

// ============= LinkedIn Tests =============

func Test_normalizeLinkedinAccountType_Table(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase page",
			input:    "page",
			expected: "Page",
		},
		{
			name:     "uppercase PAGE",
			input:    "PAGE",
			expected: "Page",
		},
		{
			name:     "lowercase profile",
			input:    "profile",
			expected: "Profile",
		},
		{
			name:     "uppercase PROFILE",
			input:    "PROFILE",
			expected: "Profile",
		},
		{
			name:     "mixed case Page",
			input:    "PaGe",
			expected: "Page",
		},
		{
			name:     "already correct Page",
			input:    "Page",
			expected: "Page",
		},
		{
			name:     "already correct Profile",
			input:    "Profile",
			expected: "Profile",
		},
		{
			name:     "unknown type passthrough",
			input:    "unknown",
			expected: "unknown",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeLinkedinAccountType(tc.input)
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func Test_buildAccountBatch_Table(t *testing.T) {
	cases := []struct {
		name            string
		accounts        []mongomodels.SocialIntegration
		syncType        string
		kafkaType       kafkamodels.LinkedinAccountType
		expectedBatch   int
		expectedSkipped int
	}{
		{
			name:            "empty accounts",
			accounts:        []mongomodels.SocialIntegration{},
			syncType:        "incremental",
			kafkaType:       kafkamodels.LinkedinAccountTypePage,
			expectedBatch:   0,
			expectedSkipped: 0,
		},
		{
			name: "valid page accounts",
			accounts: []mongomodels.SocialIntegration{
				createTestLinkedinAccount("1", "li_001", "Page", true, true),
				createTestLinkedinAccount("2", "li_002", "Page", true, true),
			},
			syncType:        "incremental",
			kafkaType:       kafkamodels.LinkedinAccountTypePage,
			expectedBatch:   2,
			expectedSkipped: 0,
		},
		{
			name: "valid profile accounts",
			accounts: []mongomodels.SocialIntegration{
				createTestLinkedinAccount("1", "li_001", "Profile", true, true),
			},
			syncType:        "full_sync",
			kafkaType:       kafkamodels.LinkedinAccountTypeProfile,
			expectedBatch:   1,
			expectedSkipped: 0,
		},
		{
			name: "skip accounts without tokens",
			accounts: []mongomodels.SocialIntegration{
				createTestLinkedinAccount("1", "li_001", "Page", true, true),
				createTestLinkedinAccount("2", "li_002", "Page", false, true),
			},
			syncType:        "incremental",
			kafkaType:       kafkamodels.LinkedinAccountTypePage,
			expectedBatch:   1,
			expectedSkipped: 1,
		},
		{
			name: "skip accounts without workspace",
			accounts: []mongomodels.SocialIntegration{
				createTestLinkedinAccount("1", "li_001", "Page", true, true),
				createTestLinkedinAccount("2", "li_002", "Page", true, false),
			},
			syncType:        "incremental",
			kafkaType:       kafkamodels.LinkedinAccountTypePage,
			expectedBatch:   1,
			expectedSkipped: 1,
		},
		{
			name: "mixed valid and invalid",
			accounts: []mongomodels.SocialIntegration{
				createTestLinkedinAccount("1", "li_001", "Page", true, true),
				createTestLinkedinAccount("2", "li_002", "Page", false, true),
				createTestLinkedinAccount("3", "li_003", "Page", true, false),
				createTestLinkedinAccount("4", "li_004", "Page", true, true),
			},
			syncType:        "incremental",
			kafkaType:       kafkamodels.LinkedinAccountTypePage,
			expectedBatch:   2,
			expectedSkipped: 2,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			batch, skipped := buildAccountBatch(tc.accounts, tc.syncType, tc.kafkaType, testLogger())

			if len(batch) != tc.expectedBatch {
				t.Fatalf("expected batch size %d, got %d", tc.expectedBatch, len(batch))
			}
			if skipped != tc.expectedSkipped {
				t.Fatalf("expected skipped %d, got %d", tc.expectedSkipped, skipped)
			}

			for _, item := range batch {
				if item.SyncType != tc.syncType {
					t.Fatalf("expected sync type %s, got %s", tc.syncType, item.SyncType)
				}
				if item.AccountType != tc.kafkaType {
					t.Fatalf("expected account type %s, got %s", tc.kafkaType, item.AccountType)
				}
				if item.LinkedinID == "" {
					t.Fatal("expected non-empty linkedin_id")
				}
			}
		})
	}
}

func Test_processLinkedinBatches_Table(t *testing.T) {
	cases := []struct {
		name             string
		accounts         []mongomodels.SocialIntegration
		countErr         error
		paginatedErr     error
		producerErr      error
		expectedMessages int
	}{
		{
			name:             "no accounts needing update",
			accounts:         []mongomodels.SocialIntegration{},
			expectedMessages: 0,
		},
		{
			name: "single batch of valid accounts",
			accounts: []mongomodels.SocialIntegration{
				createTestLinkedinAccount("1", "li_001", "Page", true, true),
				createTestLinkedinAccount("2", "li_002", "Page", true, true),
			},
			expectedMessages: 1,
		},
		{
			name:             "count error",
			accounts:         []mongomodels.SocialIntegration{},
			countErr:         errors.New("count failed"),
			expectedMessages: 0,
		},
		{
			name: "paginated fetch error",
			accounts: []mongomodels.SocialIntegration{
				createTestLinkedinAccount("1", "li_001", "Page", true, true),
			},
			paginatedErr:     errors.New("fetch failed"),
			expectedMessages: 0,
		},
		{
			name: "producer error",
			accounts: []mongomodels.SocialIntegration{
				createTestLinkedinAccount("1", "li_001", "Page", true, true),
			},
			producerErr:      errors.New("produce failed"),
			expectedMessages: 0,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUnifiedSocialRepository{
				accounts:     tc.accounts,
				countErr:     tc.countErr,
				paginatedErr: tc.paginatedErr,
				totalCount:   int64(len(tc.accounts)),
			}
			producer := &mockProducer{err: tc.producerErr}

			processLinkedinBatches(
				context.Background(),
				repo,
				producer,
				testLogger(),
				"incremental",
				"Page",
				topicLinkedinPageBatch,
				kafkamodels.LinkedinAccountTypePage,
			)

			messages := producer.getMessages()
			if len(messages) != tc.expectedMessages {
				t.Fatalf("expected %d messages, got %d", tc.expectedMessages, len(messages))
			}

			for _, msg := range messages {
				if msg.topic != topicLinkedinPageBatch {
					t.Fatalf("expected topic %s, got %s", topicLinkedinPageBatch, msg.topic)
				}

				var workOrder kafkamodels.LinkedinBatchWorkOrder
				if err := json.Unmarshal(msg.value, &workOrder); err != nil {
					t.Fatalf("failed to unmarshal message: %v", err)
				}

				if workOrder.BatchID == "" {
					t.Fatal("expected non-empty batch_id")
				}
			}
		})
	}
}

// ============= Instagram Tests =============

func Test_getInstagramAccessToken_Table(t *testing.T) {
	cases := []struct {
		name           string
		account        mongomodels.SocialIntegration
		connectedViaIG bool
		expectedToken  string
	}{
		{
			name: "direct IG connection with access_token",
			account: mongomodels.SocialIntegration{
				AccessToken: "direct_ig_token",
			},
			connectedViaIG: true,
			expectedToken:  "direct_ig_token",
		},
		{
			name: "direct IG connection with extra_data token",
			account: mongomodels.SocialIntegration{
				ExtraData: map[string]interface{}{
					"access_token": "extra_data_token",
				},
			},
			connectedViaIG: true,
			expectedToken:  "extra_data_token",
		},
		{
			name: "FB-linked with user_details token",
			account: mongomodels.SocialIntegration{
				UserDetails: map[string]interface{}{
					"access_token": "user_details_token",
				},
			},
			connectedViaIG: false,
			expectedToken:  "user_details_token",
		},
		{
			name: "FB-linked with access_token fallback",
			account: mongomodels.SocialIntegration{
				AccessToken: "access_token_fallback",
			},
			connectedViaIG: false,
			expectedToken:  "access_token_fallback",
		},
		{
			name:           "no token found",
			account:        mongomodels.SocialIntegration{},
			connectedViaIG: true,
			expectedToken:  "",
		},
		{
			name: "direct IG prefers access_token over extra_data",
			account: mongomodels.SocialIntegration{
				AccessToken: "primary_token",
				ExtraData: map[string]interface{}{
					"access_token": "extra_data_token",
				},
			},
			connectedViaIG: true,
			expectedToken:  "primary_token",
		},
		{
			name: "FB-linked prefers user_details over long_access_token",
			account: mongomodels.SocialIntegration{
				LongAccessToken: "long_token",
				UserDetails: map[string]interface{}{
					"access_token": "user_details_token",
				},
			},
			connectedViaIG: false,
			expectedToken:  "user_details_token",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := getInstagramAccessToken(tc.account, tc.connectedViaIG)
			if result != tc.expectedToken {
				t.Fatalf("expected token %q, got %q", tc.expectedToken, result)
			}
		})
	}
}

func Test_buildInstagramAccountBatch_Table(t *testing.T) {
	cases := []struct {
		name               string
		accounts           []mongomodels.SocialIntegration
		syncType           string
		seenIDs            map[string]bool
		expectedBatch      int
		expectedSkipped    int
		expectedDuplicates int
	}{
		{
			name:               "empty accounts",
			accounts:           []mongomodels.SocialIntegration{},
			syncType:           "incremental",
			seenIDs:            make(map[string]bool),
			expectedBatch:      0,
			expectedSkipped:    0,
			expectedDuplicates: 0,
		},
		{
			name: "valid direct IG accounts",
			accounts: []mongomodels.SocialIntegration{
				createTestInstagramAccount("1", "ig_001", true, true, true),
				createTestInstagramAccount("2", "ig_002", true, true, true),
			},
			syncType:           "incremental",
			seenIDs:            make(map[string]bool),
			expectedBatch:      2,
			expectedSkipped:    0,
			expectedDuplicates: 0,
		},
		{
			name: "valid FB-linked accounts",
			accounts: []mongomodels.SocialIntegration{
				createTestInstagramAccount("1", "ig_001", false, true, true),
			},
			syncType:           "full_sync",
			seenIDs:            make(map[string]bool),
			expectedBatch:      1,
			expectedSkipped:    0,
			expectedDuplicates: 0,
		},
		{
			name: "skip accounts without tokens",
			accounts: []mongomodels.SocialIntegration{
				createTestInstagramAccount("1", "ig_001", true, true, true),
				createTestInstagramAccount("2", "ig_002", true, false, true),
			},
			syncType:           "incremental",
			seenIDs:            make(map[string]bool),
			expectedBatch:      1,
			expectedSkipped:    1,
			expectedDuplicates: 0,
		},
		{
			name: "skip accounts without workspace",
			accounts: []mongomodels.SocialIntegration{
				createTestInstagramAccount("1", "ig_001", true, true, true),
				createTestInstagramAccount("2", "ig_002", true, true, false),
			},
			syncType:           "incremental",
			seenIDs:            make(map[string]bool),
			expectedBatch:      1,
			expectedSkipped:    1,
			expectedDuplicates: 0,
		},
		{
			name: "skip duplicate instagram IDs",
			accounts: []mongomodels.SocialIntegration{
				createTestInstagramAccount("1", "ig_001", true, true, true),
				createTestInstagramAccount("2", "ig_001", true, true, true), // same IG ID
			},
			syncType:           "incremental",
			seenIDs:            make(map[string]bool),
			expectedBatch:      1,
			expectedSkipped:    0,
			expectedDuplicates: 1,
		},
		{
			name: "skip already seen IDs",
			accounts: []mongomodels.SocialIntegration{
				createTestInstagramAccount("1", "ig_001", true, true, true),
			},
			syncType:           "incremental",
			seenIDs:            map[string]bool{"ig_001": true},
			expectedBatch:      0,
			expectedSkipped:    0,
			expectedDuplicates: 1,
		},
		{
			name: "mixed valid invalid and duplicates",
			accounts: []mongomodels.SocialIntegration{
				createTestInstagramAccount("1", "ig_001", true, true, true),
				createTestInstagramAccount("2", "ig_002", true, false, true), // no token
				createTestInstagramAccount("3", "ig_001", true, true, true),  // duplicate
				createTestInstagramAccount("4", "ig_003", true, true, false), // no workspace
				createTestInstagramAccount("5", "ig_004", true, true, true),
			},
			syncType:           "incremental",
			seenIDs:            make(map[string]bool),
			expectedBatch:      2,
			expectedSkipped:    2,
			expectedDuplicates: 1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			batch, skipped, duplicates := buildInstagramAccountBatch(tc.accounts, tc.syncType, tc.seenIDs, testLogger())

			if len(batch) != tc.expectedBatch {
				t.Fatalf("expected batch size %d, got %d", tc.expectedBatch, len(batch))
			}
			if skipped != tc.expectedSkipped {
				t.Fatalf("expected skipped %d, got %d", tc.expectedSkipped, skipped)
			}
			if duplicates != tc.expectedDuplicates {
				t.Fatalf("expected duplicates %d, got %d", tc.expectedDuplicates, duplicates)
			}

			for _, item := range batch {
				if item.SyncType != tc.syncType {
					t.Fatalf("expected sync type %s, got %s", tc.syncType, item.SyncType)
				}
				if item.InstagramID == "" {
					t.Fatal("expected non-empty instagram_id")
				}
			}
		})
	}
}

func Test_processInstagramBatches_Table(t *testing.T) {
	cases := []struct {
		name             string
		accounts         []mongomodels.SocialIntegration
		countErr         error
		paginatedErr     error
		producerErr      error
		expectedMessages int
	}{
		{
			name:             "no accounts needing update",
			accounts:         []mongomodels.SocialIntegration{},
			expectedMessages: 0,
		},
		{
			name: "single batch of valid accounts",
			accounts: []mongomodels.SocialIntegration{
				createTestInstagramAccount("1", "ig_001", true, true, true),
				createTestInstagramAccount("2", "ig_002", true, true, true),
			},
			expectedMessages: 1,
		},
		{
			name:             "count error",
			accounts:         []mongomodels.SocialIntegration{},
			countErr:         errors.New("count failed"),
			expectedMessages: 0,
		},
		{
			name: "paginated fetch error",
			accounts: []mongomodels.SocialIntegration{
				createTestInstagramAccount("1", "ig_001", true, true, true),
			},
			paginatedErr:     errors.New("fetch failed"),
			expectedMessages: 0,
		},
		{
			name: "producer error",
			accounts: []mongomodels.SocialIntegration{
				createTestInstagramAccount("1", "ig_001", true, true, true),
			},
			producerErr:      errors.New("produce failed"),
			expectedMessages: 0,
		},
		{
			name: "deduplication across batches",
			accounts: []mongomodels.SocialIntegration{
				createTestInstagramAccount("1", "ig_001", true, true, true),
				createTestInstagramAccount("2", "ig_001", true, true, true), // duplicate
				createTestInstagramAccount("3", "ig_002", true, true, true),
			},
			expectedMessages: 1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockUnifiedSocialRepository{
				accounts:     tc.accounts,
				countErr:     tc.countErr,
				paginatedErr: tc.paginatedErr,
				totalCount:   int64(len(tc.accounts)),
			}
			producer := &mockProducer{err: tc.producerErr}

			processInstagramBatches(
				context.Background(),
				repo,
				producer,
				testLogger(),
				"incremental",
			)

			messages := producer.getMessages()
			if len(messages) != tc.expectedMessages {
				t.Fatalf("expected %d messages, got %d", tc.expectedMessages, len(messages))
			}

			for _, msg := range messages {
				if msg.topic != topicInstagram {
					t.Fatalf("expected topic %s, got %s", topicInstagram, msg.topic)
				}

				var workOrder kafkamodels.InstagramBatchWorkOrder
				if err := json.Unmarshal(msg.value, &workOrder); err != nil {
					t.Fatalf("failed to unmarshal message: %v", err)
				}

				if workOrder.BatchID == "" {
					t.Fatal("expected non-empty batch_id")
				}
			}
		})
	}
}

// ============= Integration Tests with mtest =============

func Test_ProcessFacebookAccounts_Integration(t *testing.T) {
	h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))

	h.Run("process with no accounts", func(m *mt.T) {
		ns := mt.TestDb + ".social_integrations"
		m.AddMockResponses(
			mt.CreateCursorResponse(0, ns, mt.FirstBatch),
		)

		producer := &mockProducer{}
		ProcessFacebookAccounts(m.DB, producer, testLogger(), []string{"page"}, "incremental")

		if len(producer.getMessages()) != 0 {
			t.Fatal("expected no messages for empty result")
		}
	})

	h.Run("process with page type normalization", func(m *mt.T) {
		ns := mt.TestDb + ".social_integrations"
		m.AddMockResponses(
			mt.CreateCursorResponse(0, ns, mt.FirstBatch),
		)

		producer := &mockProducer{}
		ProcessFacebookAccounts(m.DB, producer, testLogger(), []string{"page", "Page"}, "full_sync")

		if len(producer.getMessages()) != 0 {
			t.Fatal("expected no messages for empty result")
		}
	})

	h.Run("process with valid accounts", func(m *mt.T) {
		ns := mt.TestDb + ".social_integrations"
		workspaceID := primitive.NewObjectID()
		doc := bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "platform_type", Value: "facebook"},
			{Key: "platform_identifier", Value: "fb_123"},
			{Key: "type", Value: "Page"},
			{Key: "workspace_id", Value: workspaceID},
			{Key: "state", Value: "added"},
			{Key: "validity", Value: "valid"},
			{Key: "long_access_token", Value: "test_token"},
		}

		m.AddMockResponses(
			mt.CreateCursorResponse(1, ns, mt.FirstBatch, bson.D{{Key: "n", Value: 1}}),
		)
		m.AddMockResponses(
			mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc),
			mt.CreateCursorResponse(0, ns, mt.NextBatch),
		)

		producer := &mockProducer{}
		ProcessFacebookAccounts(m.DB, producer, testLogger(), []string{"Page"}, "incremental")
	})
}

func Test_ProcessLinkedinAccounts_Integration(t *testing.T) {
	h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))

	h.Run("process with no accounts", func(m *mt.T) {
		ns := mt.TestDb + ".social_integrations"
		m.AddMockResponses(
			mt.CreateCursorResponse(0, ns, mt.FirstBatch),
		)
		m.AddMockResponses(
			mt.CreateCursorResponse(0, ns, mt.FirstBatch),
		)

		producer := &mockProducer{}
		ProcessLinkedinAccounts(context.Background(), m.DB, producer, testLogger(), []string{"Page"}, "incremental")

		if len(producer.getMessages()) != 0 {
			t.Fatal("expected no messages for empty result")
		}
	})

	h.Run("process with default account types", func(m *mt.T) {
		ns := mt.TestDb + ".social_integrations"
		m.AddMockResponses(
			mt.CreateCursorResponse(0, ns, mt.FirstBatch),
		)
		m.AddMockResponses(
			mt.CreateCursorResponse(0, ns, mt.FirstBatch),
		)

		producer := &mockProducer{}
		ProcessLinkedinAccounts(context.Background(), m.DB, producer, testLogger(), nil, "full_sync")

		if len(producer.getMessages()) != 0 {
			t.Fatal("expected no messages for empty result")
		}
	})

	h.Run("process with unknown account type", func(m *mt.T) {
		producer := &mockProducer{}
		ProcessLinkedinAccounts(context.Background(), m.DB, producer, testLogger(), []string{"unknown_type"}, "incremental")

		if len(producer.getMessages()) != 0 {
			t.Fatal("expected no messages for unknown account type")
		}
	})
}

func Test_ProcessInstagramAccounts_Integration(t *testing.T) {
	h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))

	h.Run("process with no accounts", func(m *mt.T) {
		ns := mt.TestDb + ".social_integrations"
		m.AddMockResponses(
			mt.CreateCursorResponse(0, ns, mt.FirstBatch),
		)

		producer := &mockProducer{}
		ProcessInstagramAccounts(context.Background(), m.DB, producer, testLogger(), nil, "incremental")

		if len(producer.getMessages()) != 0 {
			t.Fatal("expected no messages for empty result")
		}
	})
}
