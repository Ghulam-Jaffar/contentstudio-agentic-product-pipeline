package facebook

import (
	"context"
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

func testFacebookLogger() *zerolog.Logger {
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

func TestProcessAccountJob_Success(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetThumbs := fbGetThumbsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		fbGetThumbsFn = origGetThumbs
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, pageID string, _, _ int) ([]clickhousemodels.MinimalPost, error) {
		if pageID != "page_123" {
			t.Fatalf("unexpected pageID: %s", pageID)
		}
		return []clickhousemodels.MinimalPost{
			{PageID: "page_123", PostID: "post_1"},
			{PageID: "page_123", PostID: "post_2"},
		}, nil
	}
	fbGetThumbsFn = func(_ *social.FacebookClient, _ context.Context, pageID, accessToken, longAccessToken, decryptionKey string, posts []clickhousemodels.MinimalPost) ([]clickhousemodels.MinimalPost, error) {
		if pageID != "page_123" || accessToken != "extra_token" || longAccessToken != "long_token" || decryptionKey != "decrypt_key" {
			t.Fatalf("unexpected refresh inputs: pageID=%s access=%s long=%s key=%s", pageID, accessToken, longAccessToken, decryptionKey)
		}
		if len(posts) != 2 {
			t.Fatalf("unexpected posts length: %d", len(posts))
		}
		return []clickhousemodels.MinimalPost{
			{PageID: "page_123", PostID: "post_1", FullPicture: "https://example.com/post_1.jpg"},
		}, nil
	}

	cfg := &config.Config{DecryptionKey: "decrypt_key"}
	account := mongomodels.SocialIntegration{
		PlatformType:       mongomodels.PlatformFacebook,
		PlatformIdentifier: "page_123",
		AccessToken:        "access_token",
		LongAccessToken:    "long_token",
		ExtraData: map[string]interface{}{
			"access_token": "extra_token",
		},
	}

	res := processAccountJob(context.Background(), cfg, &social.FacebookClient{}, &conversions.ClickHouseSink{}, account, testFacebookLogger())

	if res.PageID != "page_123" || res.PostsFound != 2 || res.ThumbsResolved != 1 || res.Err != nil {
		t.Fatalf("unexpected result: %+v", res)
	}
	if len(res.Thumbs) != 1 || res.Thumbs[0].FullPicture == "" {
		t.Fatalf("expected thumbs to be returned: %+v", res.Thumbs)
	}
}

func TestProcessAccountJob_UsesExtraDataTokenFallback(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetThumbs := fbGetThumbsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		fbGetThumbsFn = origGetThumbs
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.MinimalPost, error) {
		return []clickhousemodels.MinimalPost{{PageID: "page_123", PostID: "post_1"}}, nil
	}
	fbGetThumbsFn = func(_ *social.FacebookClient, _ context.Context, _, accessToken, longAccessToken, _ string, _ []clickhousemodels.MinimalPost) ([]clickhousemodels.MinimalPost, error) {
		if accessToken != "extra_token" || longAccessToken != "plain_token" {
			t.Fatalf("unexpected token fallback: access=%s long=%s", accessToken, longAccessToken)
		}
		return []clickhousemodels.MinimalPost{{PageID: "page_123", PostID: "post_1", FullPicture: "https://example.com/post_1.jpg"}}, nil
	}

	account := mongomodels.SocialIntegration{
		PlatformIdentifier: "page_123",
		AccessToken:        "plain_token",
		ExtraData: map[string]interface{}{
			"access_token": "extra_token",
		},
	}

	res := processAccountJob(context.Background(), &config.Config{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, account, testFacebookLogger())
	if res.Err != nil || res.ThumbsResolved != 1 || len(res.Thumbs) != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestProcessAccountJob_GetPostsError(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetThumbs := fbGetThumbsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		fbGetThumbsFn = origGetThumbs
	})

	expectedErr := errors.New("clickhouse failed")
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.MinimalPost, error) {
		return nil, expectedErr
	}
	fbGetThumbsFn = func(_ *social.FacebookClient, _ context.Context, _, _, _, _ string, _ []clickhousemodels.MinimalPost) ([]clickhousemodels.MinimalPost, error) {
		t.Fatal("unexpected thumbnail refresh call")
		return nil, nil
	}

	cfg := &config.Config{}
	account := mongomodels.SocialIntegration{PlatformIdentifier: "page_123"}

	res := processAccountJob(context.Background(), cfg, &social.FacebookClient{}, &conversions.ClickHouseSink{}, account, testFacebookLogger())

	if res.Err == nil || res.Err.Error() != expectedErr.Error() {
		t.Fatalf("unexpected error result: %+v", res)
	}
	if res.PostsFound != 0 || res.ThumbsResolved != 0 || len(res.Thumbs) != 0 {
		t.Fatalf("unexpected counters after error: %+v", res)
	}
}

func TestProcessAccountJob_NoPosts(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetThumbs := fbGetThumbsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		fbGetThumbsFn = origGetThumbs
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.MinimalPost, error) {
		return []clickhousemodels.MinimalPost{}, nil
	}
	fbGetThumbsFn = func(_ *social.FacebookClient, _ context.Context, _, _, _, _ string, _ []clickhousemodels.MinimalPost) ([]clickhousemodels.MinimalPost, error) {
		t.Fatal("unexpected thumbnail refresh call")
		return nil, nil
	}

	res := processAccountJob(context.Background(), &config.Config{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, mongomodels.SocialIntegration{PlatformIdentifier: "page_1"}, testFacebookLogger())
	if res.PostsFound != 0 || res.ThumbsResolved != 0 || len(res.Thumbs) != 0 || res.Err != nil {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestProcessAccountJob_ThumbRefreshError(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetThumbs := fbGetThumbsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		fbGetThumbsFn = origGetThumbs
	})

	expectedErr := errors.New("facebook boom")
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.MinimalPost, error) {
		return []clickhousemodels.MinimalPost{{PageID: "page_1", PostID: "post_1"}}, nil
	}
	fbGetThumbsFn = func(_ *social.FacebookClient, _ context.Context, _, _, _, _ string, _ []clickhousemodels.MinimalPost) ([]clickhousemodels.MinimalPost, error) {
		return nil, expectedErr
	}

	res := processAccountJob(context.Background(), &config.Config{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, mongomodels.SocialIntegration{PlatformIdentifier: "page_1"}, testFacebookLogger())
	if !errors.Is(res.Err, expectedErr) || res.PostsFound != 1 || res.ThumbsResolved != 0 || len(res.Thumbs) != 0 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestProcessAccountJob_NoThumbsResolved(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetThumbs := fbGetThumbsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		fbGetThumbsFn = origGetThumbs
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.MinimalPost, error) {
		return []clickhousemodels.MinimalPost{{PageID: "page_1", PostID: "post_1"}}, nil
	}
	fbGetThumbsFn = func(_ *social.FacebookClient, _ context.Context, _, _, _, _ string, _ []clickhousemodels.MinimalPost) ([]clickhousemodels.MinimalPost, error) {
		return []clickhousemodels.MinimalPost{}, nil
	}

	res := processAccountJob(context.Background(), &config.Config{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, mongomodels.SocialIntegration{PlatformIdentifier: "page_1"}, testFacebookLogger())
	if res.PostsFound != 1 || res.ThumbsResolved != 0 || len(res.Thumbs) != 0 || res.Err != nil {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestRun_NoStalePostsIsNoop(t *testing.T) {
	origValid := mongoGetValidPageIDsFn
	origDistinct := chGetDistinctPageIDsFn
	t.Cleanup(func() { mongoGetValidPageIDsFn = origValid; chGetDistinctPageIDsFn = origDistinct })

	mongoGetValidPageIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"page_1"}, nil
	}
	chGetDistinctPageIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return nil, nil
	}

	Run(context.Background(), &config.Config{}, *appLogger.New("info"), mockRepo{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_DistinctQueryErrorStopsEarly(t *testing.T) {
	origValid := mongoGetValidPageIDsFn
	origDistinct := chGetDistinctPageIDsFn
	t.Cleanup(func() { mongoGetValidPageIDsFn = origValid; chGetDistinctPageIDsFn = origDistinct })

	mongoGetValidPageIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"page_1"}, nil
	}
	chGetDistinctPageIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return nil, errors.New("clickhouse boom")
	}

	Run(context.Background(), &config.Config{}, *appLogger.New("info"), mockRepo{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_SkipsAccountNotInMongo(t *testing.T) {
	origValid := mongoGetValidPageIDsFn
	origDistinct := chGetDistinctPageIDsFn
	origGetByIDs := mongoGetAccountsByIDsFn
	origMark := chMarkPageRefreshedFn
	t.Cleanup(func() {
		mongoGetValidPageIDsFn = origValid
		chGetDistinctPageIDsFn = origDistinct
		mongoGetAccountsByIDsFn = origGetByIDs
		chMarkPageRefreshedFn = origMark
	})

	mongoGetValidPageIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"page_missing"}, nil
	}
	chGetDistinctPageIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return []string{"page_missing"}, nil
	}
	mongoGetAccountsByIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context, _ string, _ []string) ([]mongomodels.SocialIntegration, error) {
		return nil, nil
	}
	chMarkPageRefreshedFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string) error {
		return nil
	}

	Run(context.Background(), &config.Config{}, *appLogger.New("info"), mockRepo{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_SkipsAccountWithNoToken(t *testing.T) {
	origValid := mongoGetValidPageIDsFn
	origDistinct := chGetDistinctPageIDsFn
	origGetByIDs := mongoGetAccountsByIDsFn
	origGetPosts := chGetPostsFn
	origMark := chMarkPageRefreshedFn
	t.Cleanup(func() {
		mongoGetValidPageIDsFn = origValid
		chGetDistinctPageIDsFn = origDistinct
		mongoGetAccountsByIDsFn = origGetByIDs
		chGetPostsFn = origGetPosts
		chMarkPageRefreshedFn = origMark
	})
	chMarkPageRefreshedFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string) error {
		return nil
	}

	mongoGetValidPageIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"page_1"}, nil
	}
	chGetDistinctPageIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return []string{"page_1"}, nil
	}
	mongoGetAccountsByIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context, _ string, _ []string) ([]mongomodels.SocialIntegration, error) {
		return []mongomodels.SocialIntegration{{PlatformIdentifier: "page_1"}}, nil
	}
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.MinimalPost, error) {
		t.Fatal("should not query ClickHouse for account with no token")
		return nil, nil
	}

	Run(context.Background(), &config.Config{}, *appLogger.New("info"), mockRepo{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_BulkUpdateError(t *testing.T) {
	origValid := mongoGetValidPageIDsFn
	origDistinct := chGetDistinctPageIDsFn
	origGetByIDs := mongoGetAccountsByIDsFn
	origGetPosts := chGetPostsFn
	origGetThumbs := fbGetThumbsFn
	origBulkUpdate := chBulkUpdateFn
	t.Cleanup(func() {
		mongoGetValidPageIDsFn = origValid
		chGetDistinctPageIDsFn = origDistinct
		mongoGetAccountsByIDsFn = origGetByIDs
		chGetPostsFn = origGetPosts
		fbGetThumbsFn = origGetThumbs
		chBulkUpdateFn = origBulkUpdate
	})

	mongoGetValidPageIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"page_1"}, nil
	}
	chGetDistinctPageIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return []string{"page_1"}, nil
	}
	mongoGetAccountsByIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context, _ string, _ []string) ([]mongomodels.SocialIntegration, error) {
		return []mongomodels.SocialIntegration{{PlatformIdentifier: "page_1", AccessToken: "token"}}, nil
	}
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.MinimalPost, error) {
		return []clickhousemodels.MinimalPost{{PageID: "page_1", PostID: "post_1"}}, nil
	}
	fbGetThumbsFn = func(_ *social.FacebookClient, _ context.Context, _, _, _, _ string, _ []clickhousemodels.MinimalPost) ([]clickhousemodels.MinimalPost, error) {
		return []clickhousemodels.MinimalPost{{PageID: "page_1", PostID: "post_1", FullPicture: "https://x"}}, nil
	}
	chBulkUpdateFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ []clickhousemodels.MinimalPost) (int, error) {
		return 0, errors.New("bulk update failed")
	}

	Run(context.Background(), &config.Config{}, *appLogger.New("info"), mockRepo{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_BulkUpdateFanout(t *testing.T) {
	origValid := mongoGetValidPageIDsFn
	origDistinct := chGetDistinctPageIDsFn
	origGetByIDs := mongoGetAccountsByIDsFn
	origGetPosts := chGetPostsFn
	origGetThumbs := fbGetThumbsFn
	origBulkUpdate := chBulkUpdateFn
	t.Cleanup(func() {
		mongoGetValidPageIDsFn = origValid
		chGetDistinctPageIDsFn = origDistinct
		mongoGetAccountsByIDsFn = origGetByIDs
		chGetPostsFn = origGetPosts
		fbGetThumbsFn = origGetThumbs
		chBulkUpdateFn = origBulkUpdate
	})

	mongoGetValidPageIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"page_1", "page_2"}, nil
	}
	chGetDistinctPageIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return []string{"page_1", "page_2"}, nil
	}
	mongoGetAccountsByIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context, _ string, ids []string) ([]mongomodels.SocialIntegration, error) {
		accs := make([]mongomodels.SocialIntegration, 0, len(ids))
		for _, id := range ids {
			accs = append(accs, mongomodels.SocialIntegration{PlatformIdentifier: id, AccessToken: "token"})
		}
		return accs, nil
	}
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, pageID string, _, _ int) ([]clickhousemodels.MinimalPost, error) {
		return []clickhousemodels.MinimalPost{{PageID: pageID, PostID: "post_1"}}, nil
	}
	fbGetThumbsFn = func(_ *social.FacebookClient, _ context.Context, pageID, _, _, _ string, _ []clickhousemodels.MinimalPost) ([]clickhousemodels.MinimalPost, error) {
		return []clickhousemodels.MinimalPost{{PageID: pageID, PostID: "post_1", FullPicture: "https://x/" + pageID}}, nil
	}
	bulkCalls := 0
	chBulkUpdateFn = func(_ *conversions.ClickHouseSink, _ context.Context, thumbs []clickhousemodels.MinimalPost) (int, error) {
		bulkCalls++
		if len(thumbs) != 2 {
			t.Fatalf("expected 2 thumbs from both accounts, got %d", len(thumbs))
		}
		return len(thumbs), nil
	}

	Run(context.Background(), &config.Config{}, *appLogger.New("info"), mockRepo{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, 10, "")

	if bulkCalls != 1 {
		t.Fatalf("expected 1 bulk update call, got %d", bulkCalls)
	}
}
