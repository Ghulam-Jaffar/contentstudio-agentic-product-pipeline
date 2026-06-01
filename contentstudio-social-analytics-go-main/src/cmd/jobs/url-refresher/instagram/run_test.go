package instagram

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	appLogger "github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type testEncryptedPayload struct {
	IV    string `json:"iv"`
	Value string `json:"value"`
}

func encryptTokenForTest(t *testing.T, plaintext, base64Key string) string {
	t.Helper()

	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		t.Fatalf("failed to decode test key: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32-byte test key, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	padded := pkcs7PadForTest([]byte(plaintext), aes.BlockSize)
	iv := bytes.Repeat([]byte{0x11}, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	payload, err := json.Marshal(testEncryptedPayload{
		IV:    base64.StdEncoding.EncodeToString(iv),
		Value: base64.StdEncoding.EncodeToString(ciphertext),
	})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	return base64.StdEncoding.EncodeToString(payload)
}

func pkcs7PadForTest(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	if padding == 0 {
		padding = blockSize
	}
	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

func testInstagramLogger() *zerolog.Logger {
	l := zerolog.New(io.Discard)
	return &l
}

type mockRepo struct {
	getValidAccounts             func(ctx context.Context, platformType string, accountTypes []string) ([]mongomodels.SocialIntegration, error)
	countValidAccountsFunc       func(ctx context.Context, platformType string, accountTypes []string) (int64, error)
	getValidAccountsByIDFunc     func(ctx context.Context, platformType string, accountTypes []string, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error)
	getAccountsByPlatformIDsFunc func(ctx context.Context, platformType string, platformIDs []string) ([]mongomodels.SocialIntegration, error)
}

func (m mockRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
	panic("unexpected FindByID call")
}
func (m mockRepo) GetByPlatformID(ctx context.Context, platformType, platformID string) (*mongomodels.SocialIntegration, error) {
	panic("unexpected GetByPlatformID call")
}
func (m mockRepo) GetByWorkspaceAndPlatformID(ctx context.Context, workspaceID string, platformType, platformID string) (*mongomodels.SocialIntegration, error) {
	panic("unexpected GetByWorkspaceAndPlatformID call")
}
func (m mockRepo) GetValidAccounts(ctx context.Context, platformType string, accountTypes []string) ([]mongomodels.SocialIntegration, error) {
	if m.getValidAccounts != nil {
		return m.getValidAccounts(ctx, platformType, accountTypes)
	}
	return nil, nil
}
func (m mockRepo) GetAccountsByWorkspace(ctx context.Context, workspaceID primitive.ObjectID, platforms []string) ([]mongomodels.SocialIntegration, error) {
	panic("unexpected GetAccountsByWorkspace call")
}
func (m mockRepo) GetAccountsNeedingUpdate(ctx context.Context, platformType string, lastUpdateField string, hours int) ([]mongomodels.SocialIntegration, error) {
	panic("unexpected GetAccountsNeedingUpdate call")
}
func (m mockRepo) GetAccountsNeedingUpdatePaginated(ctx context.Context, platformType string, accountTypes []string, hours int, skip, limit int64) ([]mongomodels.SocialIntegration, error) {
	panic("unexpected GetAccountsNeedingUpdatePaginated call")
}
func (m mockRepo) CountAccountsNeedingUpdate(ctx context.Context, platformType string, accountTypes []string, hours int) (int64, error) {
	panic("unexpected CountAccountsNeedingUpdate call")
}
func (m mockRepo) GetAccountsNeedingUpdateByID(ctx context.Context, platformType string, accountTypes []string, hours int, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error) {
	panic("unexpected GetAccountsNeedingUpdateByID call")
}
func (m mockRepo) GetYouTubeAccountsNeedingUpdatePaginated(ctx context.Context, hours int, consentDays int, skip, limit int64) ([]mongomodels.SocialIntegration, error) {
	panic("unexpected GetYouTubeAccountsNeedingUpdatePaginated call")
}
func (m mockRepo) GetYouTubeAccountsNeedingUpdateByID(ctx context.Context, hours int, consentDays int, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error) {
	panic("unexpected GetYouTubeAccountsNeedingUpdateByID call")
}
func (m mockRepo) CountYouTubeAccountsNeedingUpdate(ctx context.Context, hours int, consentDays int) (int64, error) {
	panic("unexpected CountYouTubeAccountsNeedingUpdate call")
}
func (m mockRepo) Update(ctx context.Context, id primitive.ObjectID, updates primitive.M) error {
	panic("unexpected Update call")
}
func (m mockRepo) UpdateAnalyticsTimestamp(ctx context.Context, id primitive.ObjectID, timestampType string, timestamp time.Time) error {
	panic("unexpected UpdateAnalyticsTimestamp call")
}
func (m mockRepo) UpdateTokens(ctx context.Context, id primitive.ObjectID, tokens map[string]string) error {
	panic("unexpected UpdateTokens call")
}
func (m mockRepo) UpdateState(ctx context.Context, id primitive.ObjectID, newState string) error {
	panic("unexpected UpdateState call")
}
func (m mockRepo) UpdateValidity(ctx context.Context, id primitive.ObjectID, newValidity string) error {
	panic("unexpected UpdateValidity call")
}
func (m mockRepo) RecordProcessingError(ctx context.Context, id primitive.ObjectID, errorMessage string) error {
	panic("unexpected RecordProcessingError call")
}
func (m mockRepo) ClearProcessingError(ctx context.Context, id primitive.ObjectID) error {
	panic("unexpected ClearProcessingError call")
}
func (m mockRepo) Create(ctx context.Context, account *mongomodels.SocialIntegration) (primitive.ObjectID, error) {
	panic("unexpected Create call")
}
func (m mockRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	panic("unexpected Delete call")
}
func (m mockRepo) InsertTwitterJobMetadata(ctx context.Context, payload mongodb.TwitterJobMetadataPayload) error {
	panic("unexpected InsertTwitterJobMetadata call")
}
func (m mockRepo) CountValidAccounts(ctx context.Context, platformType string, accountTypes []string) (int64, error) {
	if m.countValidAccountsFunc != nil {
		return m.countValidAccountsFunc(ctx, platformType, accountTypes)
	}
	return 0, nil
}
func (m mockRepo) GetValidAccountsByID(ctx context.Context, platformType string, accountTypes []string, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error) {
	if m.getValidAccountsByIDFunc != nil {
		return m.getValidAccountsByIDFunc(ctx, platformType, accountTypes, lastID, limit)
	}
	return nil, nil
}
func (m mockRepo) GetAccountsByPlatformIDs(ctx context.Context, platformType string, platformIDs []string) ([]mongomodels.SocialIntegration, error) {
	if m.getAccountsByPlatformIDsFunc != nil {
		return m.getAccountsByPlatformIDsFunc(ctx, platformType, platformIDs)
	}
	return nil, nil
}

func TestResolveToken(t *testing.T) {
	t.Run("prefers direct Instagram access token", func(t *testing.T) {
		got := resolveToken(mongomodels.SocialIntegration{
			PlatformType: mongomodels.PlatformInstagram,
			AccessToken:  "direct_token",
			ExtraData: map[string]interface{}{
				"connected_via_instagram": true,
			},
		}, "any-key")
		if got != "direct_token" {
			t.Fatalf("expected direct token, got %q", got)
		}
	})

	t.Run("decrypts connected Instagram access token", func(t *testing.T) {
		key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x42}, 32))
		encrypted := encryptTokenForTest(t, "decrypted_ig_token", key)

		got := resolveToken(mongomodels.SocialIntegration{
			PlatformType: mongomodels.PlatformInstagram,
			AccessToken:  encrypted,
			ExtraData: map[string]interface{}{
				"connected_via_instagram": true,
			},
		}, key)
		if got != "decrypted_ig_token" {
			t.Fatalf("expected decrypted token, got %q", got)
		}
	})

	t.Run("falls back to user_details access token", func(t *testing.T) {
		got := resolveToken(mongomodels.SocialIntegration{
			PlatformType: mongomodels.PlatformInstagram,
			UserDetails: map[string]interface{}{
				"access_token": "user_details_token",
			},
		}, "any-key")
		if got != "user_details_token" {
			t.Fatalf("expected user_details token, got %q", got)
		}
	})

	t.Run("decrypts long access token before falling back", func(t *testing.T) {
		key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x24}, 32))
		encrypted := encryptTokenForTest(t, "decrypted_long_token", key)

		got := resolveToken(mongomodels.SocialIntegration{
			PlatformType:    mongomodels.PlatformInstagram,
			AccessToken:     "short_access_token",
			LongAccessToken: encrypted,
		}, key)
		if got != "decrypted_long_token" {
			t.Fatalf("expected long token, got %q", got)
		}
	})
}

func TestProcessAccountJob_Success(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := igGetURLsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		igGetURLsFn = origGetURLs
	})

	var gotAccountID string
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, instagramID string, _, _ int) ([]clickhousemodels.InstagramMinimalPost, error) {
		gotAccountID = instagramID
		return []clickhousemodels.InstagramMinimalPost{
			{InstagramID: "ig_legacy", MediaID: "media_1"},
			{InstagramID: "ig_legacy", MediaID: "media_2"},
		}, nil
	}
	igGetURLsFn = func(_ *social.InstagramClient, _ context.Context, instagramID string, account mongomodels.SocialIntegration, decryptionKey string, posts []clickhousemodels.InstagramMinimalPost) ([]clickhousemodels.InstagramMinimalPost, error) {
		if instagramID != "ig_legacy" || account.AccessToken != "access_token" || account.LongAccessToken != "long_token" || decryptionKey != "decrypt_key" {
			t.Fatalf("unexpected refresh inputs: instagramID=%s access=%s long=%s key=%s", instagramID, account.AccessToken, account.LongAccessToken, decryptionKey)
		}
		if len(posts) != 2 {
			t.Fatalf("unexpected posts length: %d", len(posts))
		}
		return []clickhousemodels.InstagramMinimalPost{
			{InstagramID: "ig_legacy", MediaID: "media_1", MediaURL: []string{"https://example.com/media_1.jpg"}},
		}, nil
	}

	cfg := &config.Config{DecryptionKey: "decrypt_key"}
	account := mongomodels.SocialIntegration{
		PlatformIdentifier: "",
		InstagramID:        "ig_legacy",
		AccessToken:        "access_token",
		LongAccessToken:    "long_token",
	}

	res := processAccountJob(context.Background(), cfg, &social.InstagramClient{}, &conversions.ClickHouseSink{}, account, testInstagramLogger())

	if gotAccountID != "ig_legacy" {
		t.Fatalf("expected fallback instagram id, got %q", gotAccountID)
	}
	if res.InstagramID != "ig_legacy" || res.PostsFound != 2 || res.URLsResolved != 1 || len(res.Refreshed) != 1 || res.Err != nil {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestProcessAccountJob_GetPostsError(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := igGetURLsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		igGetURLsFn = origGetURLs
	})

	expectedErr := errors.New("clickhouse failed")
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.InstagramMinimalPost, error) {
		return nil, expectedErr
	}
	igGetURLsFn = func(_ *social.InstagramClient, _ context.Context, _ string, _ mongomodels.SocialIntegration, _ string, _ []clickhousemodels.InstagramMinimalPost) ([]clickhousemodels.InstagramMinimalPost, error) {
		t.Fatal("unexpected media refresh call")
		return nil, nil
	}

	cfg := &config.Config{}
	account := mongomodels.SocialIntegration{
		PlatformIdentifier: "ig_123",
	}

	res := processAccountJob(context.Background(), cfg, &social.InstagramClient{}, &conversions.ClickHouseSink{}, account, testInstagramLogger())

	if res.Err == nil || res.Err.Error() != expectedErr.Error() {
		t.Fatalf("unexpected error result: %+v", res)
	}
	if res.PostsFound != 0 || res.URLsResolved != 0 || len(res.Refreshed) != 0 {
		t.Fatalf("unexpected counters after error: %+v", res)
	}
}

func TestProcessAccountJob_NoPosts(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := igGetURLsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		igGetURLsFn = origGetURLs
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.InstagramMinimalPost, error) {
		return []clickhousemodels.InstagramMinimalPost{}, nil
	}
	igGetURLsFn = func(_ *social.InstagramClient, _ context.Context, _ string, _ mongomodels.SocialIntegration, _ string, _ []clickhousemodels.InstagramMinimalPost) ([]clickhousemodels.InstagramMinimalPost, error) {
		t.Fatal("unexpected refresh call")
		return nil, nil
	}

	res := processAccountJob(context.Background(), &config.Config{}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, mongomodels.SocialIntegration{PlatformIdentifier: "ig_1"}, testInstagramLogger())
	if res.PostsFound != 0 || res.URLsResolved != 0 || len(res.Refreshed) != 0 || res.Err != nil {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestProcessAccountJob_URLRefreshError(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := igGetURLsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		igGetURLsFn = origGetURLs
	})

	expectedErr := errors.New("instagram boom")
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.InstagramMinimalPost, error) {
		return []clickhousemodels.InstagramMinimalPost{{InstagramID: "ig_1", MediaID: "m1"}}, nil
	}
	igGetURLsFn = func(_ *social.InstagramClient, _ context.Context, _ string, _ mongomodels.SocialIntegration, _ string, _ []clickhousemodels.InstagramMinimalPost) ([]clickhousemodels.InstagramMinimalPost, error) {
		return nil, expectedErr
	}

	res := processAccountJob(context.Background(), &config.Config{}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, mongomodels.SocialIntegration{PlatformIdentifier: "ig_1"}, testInstagramLogger())
	if !errors.Is(res.Err, expectedErr) || res.PostsFound != 1 || res.URLsResolved != 0 || len(res.Refreshed) != 0 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestProcessAccountJob_NoURLsResolved(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := igGetURLsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		igGetURLsFn = origGetURLs
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.InstagramMinimalPost, error) {
		return []clickhousemodels.InstagramMinimalPost{{InstagramID: "ig_1", MediaID: "m1"}}, nil
	}
	igGetURLsFn = func(_ *social.InstagramClient, _ context.Context, _ string, _ mongomodels.SocialIntegration, _ string, _ []clickhousemodels.InstagramMinimalPost) ([]clickhousemodels.InstagramMinimalPost, error) {
		return []clickhousemodels.InstagramMinimalPost{}, nil
	}

	res := processAccountJob(context.Background(), &config.Config{}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, mongomodels.SocialIntegration{PlatformIdentifier: "ig_1"}, testInstagramLogger())
	if res.PostsFound != 1 || res.URLsResolved != 0 || len(res.Refreshed) != 0 || res.Err != nil {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestRun_NoStalePostsIsNoop(t *testing.T) {
	origValid := mongoGetValidInstagramIDsFn
	origDistinct := chGetDistinctInstagramIDsFn
	t.Cleanup(func() {
		mongoGetValidInstagramIDsFn = origValid
		chGetDistinctInstagramIDsFn = origDistinct
	})

	mongoGetValidInstagramIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"ig_1"}, nil
	}
	chGetDistinctInstagramIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return nil, nil
	}

	Run(context.Background(), &config.Config{}, *appLogger.New("info"), mockRepo{}, &social.InstagramClient{}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_DistinctQueryErrorStopsEarly(t *testing.T) {
	origValid := mongoGetValidInstagramIDsFn
	origDistinct := chGetDistinctInstagramIDsFn
	t.Cleanup(func() {
		mongoGetValidInstagramIDsFn = origValid
		chGetDistinctInstagramIDsFn = origDistinct
	})

	mongoGetValidInstagramIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"ig_1"}, nil
	}
	chGetDistinctInstagramIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return nil, errors.New("clickhouse boom")
	}

	Run(context.Background(), &config.Config{}, *appLogger.New("info"), mockRepo{}, &social.InstagramClient{}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_SkipsAccountNotInMongo(t *testing.T) {
	origValid := mongoGetValidInstagramIDsFn
	origDistinct := chGetDistinctInstagramIDsFn
	origGetByIDs := mongoGetAccountsByIDsFn
	origMark := chMarkInstagramRefreshedFn
	t.Cleanup(func() {
		mongoGetValidInstagramIDsFn = origValid
		chGetDistinctInstagramIDsFn = origDistinct
		mongoGetAccountsByIDsFn = origGetByIDs
		chMarkInstagramRefreshedFn = origMark
	})

	mongoGetValidInstagramIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"ig_missing"}, nil
	}
	chGetDistinctInstagramIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return []string{"ig_missing"}, nil
	}
	mongoGetAccountsByIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context, _ string, _ []string) ([]mongomodels.SocialIntegration, error) {
		return nil, nil
	}
	chMarkInstagramRefreshedFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string) error {
		return nil
	}

	Run(context.Background(), &config.Config{}, *appLogger.New("info"), mockRepo{}, &social.InstagramClient{}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_SkipsAccountWithNoToken(t *testing.T) {
	origValid := mongoGetValidInstagramIDsFn
	origDistinct := chGetDistinctInstagramIDsFn
	origGetByIDs := mongoGetAccountsByIDsFn
	origGetPosts := chGetPostsFn
	origMark := chMarkInstagramRefreshedFn
	t.Cleanup(func() {
		mongoGetValidInstagramIDsFn = origValid
		chGetDistinctInstagramIDsFn = origDistinct
		mongoGetAccountsByIDsFn = origGetByIDs
		chGetPostsFn = origGetPosts
		chMarkInstagramRefreshedFn = origMark
	})
	chMarkInstagramRefreshedFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string) error {
		return nil
	}

	mongoGetValidInstagramIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"ig_1"}, nil
	}
	chGetDistinctInstagramIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return []string{"ig_1"}, nil
	}
	mongoGetAccountsByIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context, _ string, _ []string) ([]mongomodels.SocialIntegration, error) {
		return []mongomodels.SocialIntegration{{PlatformIdentifier: "ig_1"}}, nil
	}
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.InstagramMinimalPost, error) {
		t.Fatal("should not query ClickHouse for account with no token")
		return nil, nil
	}

	Run(context.Background(), &config.Config{}, *appLogger.New("info"), mockRepo{}, &social.InstagramClient{}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_BulkUpdateError(t *testing.T) {
	origValid := mongoGetValidInstagramIDsFn
	origDistinct := chGetDistinctInstagramIDsFn
	origGetByIDs := mongoGetAccountsByIDsFn
	origGetPosts := chGetPostsFn
	origGetURLs := igGetURLsFn
	origBulkUpdate := chBulkUpdateFn
	origBulkMark := chBulkMarkInstagramRefreshedFn
	t.Cleanup(func() {
		mongoGetValidInstagramIDsFn = origValid
		chGetDistinctInstagramIDsFn = origDistinct
		mongoGetAccountsByIDsFn = origGetByIDs
		chGetPostsFn = origGetPosts
		igGetURLsFn = origGetURLs
		chBulkUpdateFn = origBulkUpdate
		chBulkMarkInstagramRefreshedFn = origBulkMark
	})
	chBulkMarkInstagramRefreshedFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ []string) error { return nil }

	mongoGetValidInstagramIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"ig_1"}, nil
	}
	chGetDistinctInstagramIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return []string{"ig_1"}, nil
	}
	mongoGetAccountsByIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context, _ string, _ []string) ([]mongomodels.SocialIntegration, error) {
		return []mongomodels.SocialIntegration{{PlatformIdentifier: "ig_1", AccessToken: "token"}}, nil
	}
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.InstagramMinimalPost, error) {
		return []clickhousemodels.InstagramMinimalPost{{InstagramID: "ig_1", MediaID: "m1"}}, nil
	}
	igGetURLsFn = func(_ *social.InstagramClient, _ context.Context, instagramID string, _ mongomodels.SocialIntegration, _ string, _ []clickhousemodels.InstagramMinimalPost) ([]clickhousemodels.InstagramMinimalPost, error) {
		return []clickhousemodels.InstagramMinimalPost{{InstagramID: instagramID, MediaID: "m1", MediaURL: []string{"https://x"}}}, nil
	}
	chBulkUpdateFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ []clickhousemodels.InstagramMinimalPost) (int, error) {
		return 0, errors.New("bulk update failed")
	}

	Run(context.Background(), &config.Config{}, *appLogger.New("info"), mockRepo{}, &social.InstagramClient{}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_BulkUpdateFanout(t *testing.T) {
	origValid := mongoGetValidInstagramIDsFn
	origDistinct := chGetDistinctInstagramIDsFn
	origGetByIDs := mongoGetAccountsByIDsFn
	origGetPosts := chGetPostsFn
	origGetURLs := igGetURLsFn
	origBulkUpdate := chBulkUpdateFn
	origBulkMark := chBulkMarkInstagramRefreshedFn
	t.Cleanup(func() {
		mongoGetValidInstagramIDsFn = origValid
		chGetDistinctInstagramIDsFn = origDistinct
		mongoGetAccountsByIDsFn = origGetByIDs
		chGetPostsFn = origGetPosts
		igGetURLsFn = origGetURLs
		chBulkUpdateFn = origBulkUpdate
		chBulkMarkInstagramRefreshedFn = origBulkMark
	})
	chBulkMarkInstagramRefreshedFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ []string) error { return nil }

	mongoGetValidInstagramIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"ig_1", "ig_2"}, nil
	}
	chGetDistinctInstagramIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return []string{"ig_1", "ig_2"}, nil
	}
	mongoGetAccountsByIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context, _ string, ids []string) ([]mongomodels.SocialIntegration, error) {
		accs := make([]mongomodels.SocialIntegration, 0, len(ids))
		for _, id := range ids {
			accs = append(accs, mongomodels.SocialIntegration{PlatformIdentifier: id, AccessToken: "token"})
		}
		return accs, nil
	}
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, instagramID string, _, _ int) ([]clickhousemodels.InstagramMinimalPost, error) {
		return []clickhousemodels.InstagramMinimalPost{{InstagramID: instagramID, MediaID: "m1"}}, nil
	}
	igGetURLsFn = func(_ *social.InstagramClient, _ context.Context, instagramID string, _ mongomodels.SocialIntegration, _ string, _ []clickhousemodels.InstagramMinimalPost) ([]clickhousemodels.InstagramMinimalPost, error) {
		return []clickhousemodels.InstagramMinimalPost{{InstagramID: instagramID, MediaID: "m1", MediaURL: []string{"https://x/" + instagramID}}}, nil
	}
	bulkCalls := 0
	chBulkUpdateFn = func(_ *conversions.ClickHouseSink, _ context.Context, posts []clickhousemodels.InstagramMinimalPost) (int, error) {
		bulkCalls++
		if len(posts) != 2 {
			t.Fatalf("expected 2 posts from both accounts, got %d", len(posts))
		}
		return len(posts), nil
	}

	Run(context.Background(), &config.Config{}, *appLogger.New("info"), mockRepo{}, &social.InstagramClient{}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, 10, "")

	if bulkCalls != 1 {
		t.Fatalf("expected 1 bulk update call, got %d", bulkCalls)
	}
}
