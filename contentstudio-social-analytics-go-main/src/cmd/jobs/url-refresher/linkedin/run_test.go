package linkedin

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

func testLinkedInLogger() *zerolog.Logger {
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
	origGetURLs := liGetURLsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		liGetURLsFn = origGetURLs
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, linkedinID string, _, _ int) ([]clickhousemodels.LinkedInMinimalPost, error) {
		if linkedinID != "li_123" {
			t.Fatalf("unexpected linkedinID: %s", linkedinID)
		}
		return []clickhousemodels.LinkedInMinimalPost{
			{LinkedinID: "li_123", PostID: "post_1", Activity: "urn:li:ugcPost:post_1"},
			{LinkedinID: "li_123", PostID: "post_2", Activity: "urn:li:ugcPost:post_2"},
		}, nil
	}
	liGetURLsFn = func(_ *social.LinkedInClient, _ context.Context, linkedinID, entityType, accessToken string, posts []clickhousemodels.LinkedInMinimalPost) ([]clickhousemodels.LinkedInMinimalPost, error) {
		if linkedinID != "li_123" || entityType != "organization" || accessToken != "plain_token" {
			t.Fatalf("unexpected refresh inputs: linkedinID=%s entityType=%s access=%s", linkedinID, entityType, accessToken)
		}
		if len(posts) != 2 {
			t.Fatalf("unexpected posts length: %d", len(posts))
		}
		return []clickhousemodels.LinkedInMinimalPost{
			{LinkedinID: "li_123", PostID: "post_1", Image: "https://example.com/post_1.jpg", Media: []string{"https://example.com/post_1.jpg"}},
		}, nil
	}

	cfg := &config.Config{DecryptionKey: "decrypt_key"}
	account := mongomodels.SocialIntegration{
		PlatformIdentifier: "li_123",
		AccessToken:        "plain_token",
		Type:               mongomodels.TypePage,
	}

	res := processAccountJob(context.Background(), cfg, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, account, testLinkedInLogger())
	if res.LinkedinID != "li_123" || res.EntityType != "organization" || res.PostsFound != 2 || res.URLsResolved != 1 || len(res.Refreshed) != 1 || res.Err != nil {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestResolveToken_UsesExtraDataFallback(t *testing.T) {
	got := resolveToken(mongomodels.SocialIntegration{
		AccessToken: "",
		ExtraData: map[string]interface{}{
			"access_token": "extra_token",
		},
	}, "any-key")
	if got != "extra_token" {
		t.Fatalf("expected extra_data token, got %q", got)
	}
}

func TestNormalizeAccountTypes_DefaultsToPage(t *testing.T) {
	got := normalizeAccountTypes("")
	if len(got) != 1 || got[0] != mongomodels.TypePage {
		t.Fatalf("expected default page account type, got %#v", got)
	}
}

func TestProcessAccountJob_GetPostsError(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := liGetURLsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		liGetURLsFn = origGetURLs
	})

	expectedErr := errors.New("clickhouse failed")
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.LinkedInMinimalPost, error) {
		return nil, expectedErr
	}
	liGetURLsFn = func(_ *social.LinkedInClient, _ context.Context, _, _, _ string, _ []clickhousemodels.LinkedInMinimalPost) ([]clickhousemodels.LinkedInMinimalPost, error) {
		t.Fatal("unexpected LinkedIn refresh call")
		return nil, nil
	}

	res := processAccountJob(context.Background(), &config.Config{}, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, mongomodels.SocialIntegration{PlatformIdentifier: "li_123", AccessToken: "token"}, testLinkedInLogger())
	if !errors.Is(res.Err, expectedErr) {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestProcessAccountJob_NoPosts(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := liGetURLsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		liGetURLsFn = origGetURLs
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.LinkedInMinimalPost, error) {
		return []clickhousemodels.LinkedInMinimalPost{}, nil
	}
	liGetURLsFn = func(_ *social.LinkedInClient, _ context.Context, _, _, _ string, _ []clickhousemodels.LinkedInMinimalPost) ([]clickhousemodels.LinkedInMinimalPost, error) {
		t.Fatal("unexpected LinkedIn refresh call")
		return nil, nil
	}

	res := processAccountJob(context.Background(), &config.Config{}, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, mongomodels.SocialIntegration{PlatformIdentifier: "li_123", AccessToken: "token"}, testLinkedInLogger())
	if res.PostsFound != 0 || res.URLsResolved != 0 || len(res.Refreshed) != 0 || res.Err != nil {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestProcessAccountJob_RefreshError(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := liGetURLsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		liGetURLsFn = origGetURLs
	})

	expectedErr := errors.New("linkedin boom")
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.LinkedInMinimalPost, error) {
		return []clickhousemodels.LinkedInMinimalPost{{LinkedinID: "li_123", PostID: "post_1"}}, nil
	}
	liGetURLsFn = func(_ *social.LinkedInClient, _ context.Context, _, _, _ string, _ []clickhousemodels.LinkedInMinimalPost) ([]clickhousemodels.LinkedInMinimalPost, error) {
		return nil, expectedErr
	}

	res := processAccountJob(context.Background(), &config.Config{}, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, mongomodels.SocialIntegration{PlatformIdentifier: "li_123", AccessToken: "token"}, testLinkedInLogger())
	if !errors.Is(res.Err, expectedErr) || res.PostsFound != 1 || res.URLsResolved != 0 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestProcessAccountJob_NoURLsResolved(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := liGetURLsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		liGetURLsFn = origGetURLs
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.LinkedInMinimalPost, error) {
		return []clickhousemodels.LinkedInMinimalPost{{LinkedinID: "li_123", PostID: "post_1"}}, nil
	}
	liGetURLsFn = func(_ *social.LinkedInClient, _ context.Context, _, _, _ string, _ []clickhousemodels.LinkedInMinimalPost) ([]clickhousemodels.LinkedInMinimalPost, error) {
		return []clickhousemodels.LinkedInMinimalPost{}, nil
	}

	res := processAccountJob(context.Background(), &config.Config{}, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, mongomodels.SocialIntegration{PlatformIdentifier: "li_123", AccessToken: "token"}, testLinkedInLogger())
	if res.PostsFound != 1 || res.URLsResolved != 0 || len(res.Refreshed) != 0 || res.Err != nil {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestRun_NoStalePostsIsNoop(t *testing.T) {
	origValid := mongoGetValidLinkedInIDsFn
	origDistinct := chGetDistinctLinkedInIDsFn
	t.Cleanup(func() {
		mongoGetValidLinkedInIDsFn = origValid
		chGetDistinctLinkedInIDsFn = origDistinct
	})

	mongoGetValidLinkedInIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"li_1"}, nil
	}
	chGetDistinctLinkedInIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return nil, nil
	}

	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockRepo{}, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_DistinctQueryErrorStopsEarly(t *testing.T) {
	origValid := mongoGetValidLinkedInIDsFn
	origDistinct := chGetDistinctLinkedInIDsFn
	t.Cleanup(func() {
		mongoGetValidLinkedInIDsFn = origValid
		chGetDistinctLinkedInIDsFn = origDistinct
	})

	mongoGetValidLinkedInIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"li_1"}, nil
	}
	chGetDistinctLinkedInIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return nil, errors.New("clickhouse boom")
	}

	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockRepo{}, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_SkipsAccountNotInMongo(t *testing.T) {
	origValid := mongoGetValidLinkedInIDsFn
	origDistinct := chGetDistinctLinkedInIDsFn
	origGetByIDs := mongoGetAccountsByIDsFn
	origMark := chMarkLinkedInRefreshedFn
	t.Cleanup(func() {
		mongoGetValidLinkedInIDsFn = origValid
		chGetDistinctLinkedInIDsFn = origDistinct
		mongoGetAccountsByIDsFn = origGetByIDs
		chMarkLinkedInRefreshedFn = origMark
	})

	mongoGetValidLinkedInIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"li_missing"}, nil
	}
	chGetDistinctLinkedInIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return []string{"li_missing"}, nil
	}
	mongoGetAccountsByIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context, _ string, _ []string) ([]mongomodels.SocialIntegration, error) {
		return nil, nil
	}
	chMarkLinkedInRefreshedFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string) error {
		return nil
	}

	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockRepo{}, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_SkipsAccountWithNoToken(t *testing.T) {
	origValid := mongoGetValidLinkedInIDsFn
	origDistinct := chGetDistinctLinkedInIDsFn
	origGetByIDs := mongoGetAccountsByIDsFn
	origGetPosts := chGetPostsFn
	origMark := chMarkLinkedInRefreshedFn
	t.Cleanup(func() {
		mongoGetValidLinkedInIDsFn = origValid
		chGetDistinctLinkedInIDsFn = origDistinct
		mongoGetAccountsByIDsFn = origGetByIDs
		chGetPostsFn = origGetPosts
		chMarkLinkedInRefreshedFn = origMark
	})
	chMarkLinkedInRefreshedFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string) error {
		return nil
	}

	mongoGetValidLinkedInIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"li_1"}, nil
	}
	chGetDistinctLinkedInIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return []string{"li_1"}, nil
	}
	mongoGetAccountsByIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context, _ string, _ []string) ([]mongomodels.SocialIntegration, error) {
		return []mongomodels.SocialIntegration{{PlatformIdentifier: "li_1"}}, nil
	}
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.LinkedInMinimalPost, error) {
		t.Fatal("should not query ClickHouse for account with no token")
		return nil, nil
	}

	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockRepo{}, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_BulkUpdateError(t *testing.T) {
	origValid := mongoGetValidLinkedInIDsFn
	origDistinct := chGetDistinctLinkedInIDsFn
	origGetByIDs := mongoGetAccountsByIDsFn
	origGetPosts := chGetPostsFn
	origGetURLs := liGetURLsFn
	origBulkUpdate := chBulkUpdateFn
	t.Cleanup(func() {
		mongoGetValidLinkedInIDsFn = origValid
		chGetDistinctLinkedInIDsFn = origDistinct
		mongoGetAccountsByIDsFn = origGetByIDs
		chGetPostsFn = origGetPosts
		liGetURLsFn = origGetURLs
		chBulkUpdateFn = origBulkUpdate
	})

	mongoGetValidLinkedInIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"li_123"}, nil
	}
	chGetDistinctLinkedInIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return []string{"li_123"}, nil
	}
	mongoGetAccountsByIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context, _ string, _ []string) ([]mongomodels.SocialIntegration, error) {
		return []mongomodels.SocialIntegration{{PlatformIdentifier: "li_123", AccessToken: "token"}}, nil
	}
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.LinkedInMinimalPost, error) {
		return []clickhousemodels.LinkedInMinimalPost{{LinkedinID: "li_123", PostID: "post_1"}}, nil
	}
	liGetURLsFn = func(_ *social.LinkedInClient, _ context.Context, _, _, _ string, _ []clickhousemodels.LinkedInMinimalPost) ([]clickhousemodels.LinkedInMinimalPost, error) {
		return []clickhousemodels.LinkedInMinimalPost{{LinkedinID: "li_123", PostID: "post_1", Image: "https://example.com/post_1.jpg"}}, nil
	}
	chBulkUpdateFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ []clickhousemodels.LinkedInMinimalPost) (int, error) {
		return 0, errors.New("bulk update failed")
	}

	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockRepo{}, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, 1, "page")
}

func TestRun_BulkUpdateLinkedInFanout(t *testing.T) {
	origValid := mongoGetValidLinkedInIDsFn
	origDistinct := chGetDistinctLinkedInIDsFn
	origGetByIDs := mongoGetAccountsByIDsFn
	origGetPosts := chGetPostsFn
	origGetURLs := liGetURLsFn
	origBulkUpdate := chBulkUpdateFn
	t.Cleanup(func() {
		mongoGetValidLinkedInIDsFn = origValid
		chGetDistinctLinkedInIDsFn = origDistinct
		mongoGetAccountsByIDsFn = origGetByIDs
		chGetPostsFn = origGetPosts
		liGetURLsFn = origGetURLs
		chBulkUpdateFn = origBulkUpdate
	})

	mongoGetValidLinkedInIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"li_aaa", "li_bbb"}, nil
	}
	chGetDistinctLinkedInIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return []string{"li_aaa", "li_bbb"}, nil
	}
	mongoGetAccountsByIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context, _ string, ids []string) ([]mongomodels.SocialIntegration, error) {
		accs := make([]mongomodels.SocialIntegration, 0, len(ids))
		for _, id := range ids {
			accs = append(accs, mongomodels.SocialIntegration{PlatformIdentifier: id, AccessToken: "token"})
		}
		return accs, nil
	}
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, linkedinID string, _, _ int) ([]clickhousemodels.LinkedInMinimalPost, error) {
		return []clickhousemodels.LinkedInMinimalPost{
			{LinkedinID: linkedinID, PostID: linkedinID + "_post_1"},
		}, nil
	}
	liGetURLsFn = func(_ *social.LinkedInClient, _ context.Context, linkedinID, _, _ string, posts []clickhousemodels.LinkedInMinimalPost) ([]clickhousemodels.LinkedInMinimalPost, error) {
		return []clickhousemodels.LinkedInMinimalPost{
			{LinkedinID: linkedinID, PostID: posts[0].PostID, Image: "https://example.com/img.jpg"},
		}, nil
	}

	bulkCalls := 0
	var bulkPosts []clickhousemodels.LinkedInMinimalPost
	chBulkUpdateFn = func(_ *conversions.ClickHouseSink, _ context.Context, posts []clickhousemodels.LinkedInMinimalPost) (int, error) {
		bulkCalls++
		bulkPosts = posts
		return len(posts), nil
	}

	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockRepo{}, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, 2, "page")

	if bulkCalls != 1 {
		t.Fatalf("expected exactly 1 bulk update call, got %d", bulkCalls)
	}
	if len(bulkPosts) != 2 {
		t.Fatalf("expected 2 posts in bulk update (one per account), got %d", len(bulkPosts))
	}
}

func TestRun_GetValidAccountsError(t *testing.T) {
	origValid := mongoGetValidLinkedInIDsFn
	origDistinct := chGetDistinctLinkedInIDsFn
	origGetByIDs := mongoGetAccountsByIDsFn
	t.Cleanup(func() {
		mongoGetValidLinkedInIDsFn = origValid
		chGetDistinctLinkedInIDsFn = origDistinct
		mongoGetAccountsByIDsFn = origGetByIDs
	})

	mongoGetValidLinkedInIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"li_1"}, nil
	}
	chGetDistinctLinkedInIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return []string{"li_1"}, nil
	}
	mongoGetAccountsByIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context, _ string, _ []string) ([]mongomodels.SocialIntegration, error) {
		return nil, errors.New("mongo failed")
	}

	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockRepo{}, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestRun_NoAccounts(t *testing.T) {
	origValid := mongoGetValidLinkedInIDsFn
	origDistinct := chGetDistinctLinkedInIDsFn
	t.Cleanup(func() {
		mongoGetValidLinkedInIDsFn = origValid
		chGetDistinctLinkedInIDsFn = origDistinct
	})

	mongoGetValidLinkedInIDsFn = func(_ mongodb.UnifiedSocialRepository, _ context.Context) ([]string, error) {
		return []string{"li_1"}, nil
	}
	chGetDistinctLinkedInIDsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _ []string) ([]string, error) {
		return nil, nil
	}

	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockRepo{}, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, 1, "")
}

func TestProcessAccountJob_ProfileAccountUsesProfileEntity(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := liGetURLsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		liGetURLsFn = origGetURLs
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.LinkedInMinimalPost, error) {
		return []clickhousemodels.LinkedInMinimalPost{{LinkedinID: "li_profile", PostID: "post_1"}}, nil
	}
	liGetURLsFn = func(_ *social.LinkedInClient, _ context.Context, linkedinID, entityType, accessToken string, _ []clickhousemodels.LinkedInMinimalPost) ([]clickhousemodels.LinkedInMinimalPost, error) {
		if linkedinID != "li_profile" || entityType != "profile" || accessToken != "profile_token" {
			t.Fatalf("unexpected refresh inputs: linkedinID=%s entityType=%s access=%s", linkedinID, entityType, accessToken)
		}
		return []clickhousemodels.LinkedInMinimalPost{}, nil
	}

	res := processAccountJob(context.Background(), &config.Config{}, &social.LinkedInClient{}, &conversions.ClickHouseSink{}, mongomodels.SocialIntegration{PlatformIdentifier: "li_profile", AccessToken: "profile_token", Type: mongomodels.TypeProfile}, testLinkedInLogger())
	if res.EntityType != "profile" || res.Err != nil {
		t.Fatalf("unexpected result: %+v", res)
	}
}

