package facebook_competitor

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	appLogger "github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func testFacebookCompetitorLogger() *zerolog.Logger {
	l := zerolog.New(io.Discard)
	return &l
}

type mockCompetitorRepo struct {
	getAccounts func(ctx context.Context, platformType string) ([]*mongomodels.Competitor, error)
}

func (m mockCompetitorRepo) GetAccounts(ctx context.Context, platformType string) ([]*mongomodels.Competitor, error) {
	if m.getAccounts != nil {
		return m.getAccounts(ctx, platformType)
	}
	return nil, nil
}

type mockRedisClient struct {
	do func(ctx context.Context, args ...interface{}) *redis.Cmd
}

func (m mockRedisClient) Do(ctx context.Context, args ...interface{}) *redis.Cmd {
	if m.do != nil {
		return m.do(ctx, args...)
	}
	cmd := redis.NewCmd(ctx)
	cmd.SetVal("")
	return cmd
}

func redisCmdWithText(ctx context.Context, val string) *redis.Cmd {
	cmd := redis.NewCmd(ctx)
	cmd.SetVal(val)
	return cmd
}

func redisCmdWithErr(ctx context.Context, err error) *redis.Cmd {
	cmd := redis.NewCmd(ctx)
	cmd.SetErr(err)
	return cmd
}

// processAccountJob now returns resolved assets/shared pics in AccountResult without
// writing to ClickHouse. The bulk write happens once in Run() after all workers finish.

func TestProcessAccountJob_Success(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := fbGetURLsFn
	origGetSharedPosts := chGetSharedPostsFn
	origGetSharedPics := fbGetSharedPicsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		fbGetURLsFn = origGetURLs
		chGetSharedPostsFn = origGetSharedPosts
		fbGetSharedPicsFn = origGetSharedPics
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, facebookID string, _, _ int) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
		if facebookID != "12345" {
			t.Fatalf("unexpected facebook id: %s", facebookID)
		}
		return []clickhousemodels.FacebookCompetitorMinimalMediaAsset{{PageID: "12345", PostID: "p1", MediaID: "mid_1"}}, nil
	}
	fbGetURLsFn = func(_ *social.FacebookClient, _ context.Context, facebookID string, accessToken string, assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
		if facebookID != "12345" || accessToken != "plain_token" || len(assets) != 1 {
			t.Fatalf("unexpected refresh inputs: id=%s token=%s assets=%d", facebookID, accessToken, len(assets))
		}
		return []clickhousemodels.FacebookCompetitorMinimalMediaAsset{{PageID: "12345", PostID: "p1", MediaID: "mid_1", Link: "https://example.com/p1.jpg"}}, nil
	}
	chGetSharedPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, facebookID string, _, _ int) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error) {
		if facebookID != "12345" {
			t.Fatalf("unexpected facebook id for shared posts: %s", facebookID)
		}
		return []clickhousemodels.FacebookCompetitorMinimalSharedPost{{FacebookID: "12345", PostID: "p2", SharedFromID: "source_1"}}, nil
	}
	fbGetSharedPicsFn = func(_ *social.FacebookClient, _ context.Context, facebookID string, accessToken string, posts []clickhousemodels.FacebookCompetitorMinimalSharedPost) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error) {
		if facebookID != "12345" || accessToken != "plain_token" || len(posts) != 1 {
			t.Fatalf("unexpected shared pic refresh inputs: id=%s token=%s posts=%d", facebookID, accessToken, len(posts))
		}
		return []clickhousemodels.FacebookCompetitorMinimalSharedPost{{FacebookID: "12345", PostID: "p2", SharedFromID: "source_1", SharedFromPic: "https://example.com/source.jpg"}}, nil
	}

	account := &mongomodels.Competitor{CompetitorID: "12345"}
	redisClient := mockRedisClient{
		do: func(ctx context.Context, args ...interface{}) *redis.Cmd {
			if len(args) != 2 || args[0] != "SRANDMEMBER" || args[1] != tokenQueueFacebook {
				t.Fatalf("unexpected redis args: %#v", args)
			}
			return redisCmdWithText(ctx, `{"token":"plain_token"}`)
		},
	}

	res := processAccountJob(context.Background(), &config.Config{DecryptionKey: "unused"}, redisClient, &social.FacebookClient{}, &conversions.ClickHouseSink{}, account, testFacebookCompetitorLogger())
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.FacebookID != "12345" {
		t.Fatalf("unexpected facebook_id: %s", res.FacebookID)
	}
	if res.PostsFound != 1 {
		t.Fatalf("expected PostsFound=1, got %d", res.PostsFound)
	}
	if res.SharedPostsFound != 1 {
		t.Fatalf("expected SharedPostsFound=1, got %d", res.SharedPostsFound)
	}
	if res.URLsResolved != 1 || len(res.ResolvedAssets) != 1 {
		t.Fatalf("expected 1 resolved asset, got URLsResolved=%d ResolvedAssets=%d", res.URLsResolved, len(res.ResolvedAssets))
	}
	if res.SharedPicsResolved != 1 || len(res.ResolvedSharedPics) != 1 {
		t.Fatalf("expected 1 resolved shared pic, got SharedPicsResolved=%d ResolvedSharedPics=%d", res.SharedPicsResolved, len(res.ResolvedSharedPics))
	}
	if res.ResolvedAssets[0].Link != "https://example.com/p1.jpg" {
		t.Fatalf("unexpected resolved link: %s", res.ResolvedAssets[0].Link)
	}
	if res.ResolvedSharedPics[0].SharedFromPic != "https://example.com/source.jpg" {
		t.Fatalf("unexpected resolved shared pic: %s", res.ResolvedSharedPics[0].SharedFromPic)
	}
}

func TestProcessAccountJob_GetPostsError(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetSharedPosts := chGetSharedPostsFn
	t.Cleanup(func() { chGetPostsFn = origGetPosts })
	t.Cleanup(func() { chGetSharedPostsFn = origGetSharedPosts })

	expectedErr := errors.New("clickhouse failed")
	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
		return nil, expectedErr
	}

	res := processAccountJob(context.Background(), &config.Config{}, mockRedisClient{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, &mongomodels.Competitor{CompetitorID: "12345"}, testFacebookCompetitorLogger())
	if !errors.Is(res.Err, expectedErr) {
		t.Fatalf("expected get posts error, got %+v", res)
	}
}

func TestProcessAccountJob_TokenError(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := fbGetURLsFn
	origGetSharedPosts := chGetSharedPostsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		fbGetURLsFn = origGetURLs
		chGetSharedPostsFn = origGetSharedPosts
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
		return []clickhousemodels.FacebookCompetitorMinimalMediaAsset{{PageID: "12345", PostID: "p1", MediaID: "mid_1"}}, nil
	}
	fbGetURLsFn = func(_ *social.FacebookClient, _ context.Context, _, _ string, _ []clickhousemodels.FacebookCompetitorMinimalMediaAsset) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
		t.Fatal("unexpected competitor refresh call")
		return nil, nil
	}
	chGetSharedPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error) {
		return []clickhousemodels.FacebookCompetitorMinimalSharedPost{}, nil
	}

	res := processAccountJob(context.Background(), &config.Config{}, mockRedisClient{
		do: func(ctx context.Context, args ...interface{}) *redis.Cmd {
			return redisCmdWithErr(ctx, errors.New("redis failed"))
		},
	}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, &mongomodels.Competitor{CompetitorID: "12345"}, testFacebookCompetitorLogger())
	if res.Err == nil {
		t.Fatalf("expected token error, got %+v", res)
	}
}

func TestProcessAccountJob_NoPosts(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetSharedPosts := chGetSharedPostsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		chGetSharedPostsFn = origGetSharedPosts
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
		return nil, nil
	}
	chGetSharedPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error) {
		return nil, nil
	}

	res := processAccountJob(context.Background(), &config.Config{}, mockRedisClient{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, &mongomodels.Competitor{CompetitorID: "12345"}, testFacebookCompetitorLogger())
	if res.Err != nil || res.PostsFound != 0 || res.SharedPostsFound != 0 || len(res.ResolvedAssets) != 0 || len(res.ResolvedSharedPics) != 0 {
		t.Fatalf("expected empty result for no posts, got %+v", res)
	}
}

func TestRun_BulkUpdateCalledOnce(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := fbGetURLsFn
	origGetSharedPosts := chGetSharedPostsFn
	origGetSharedPics := fbGetSharedPicsFn
	origBulkAssets := chBulkUpdateAssetsFn
	origBulkShared := chBulkUpdateSharedPicsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		fbGetURLsFn = origGetURLs
		chGetSharedPostsFn = origGetSharedPosts
		fbGetSharedPicsFn = origGetSharedPics
		chBulkUpdateAssetsFn = origBulkAssets
		chBulkUpdateSharedPicsFn = origBulkShared
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
		return []clickhousemodels.FacebookCompetitorMinimalMediaAsset{{PageID: "12345", PostID: "p1", MediaID: "mid_1"}}, nil
	}
	fbGetURLsFn = func(_ *social.FacebookClient, _ context.Context, _, _ string, assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
		return []clickhousemodels.FacebookCompetitorMinimalMediaAsset{{PageID: "12345", PostID: "p1", MediaID: "mid_1", Link: "https://example.com/p1.jpg"}}, nil
	}
	chGetSharedPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ string, _, _ int) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error) {
		return nil, nil
	}
	fbGetSharedPicsFn = func(_ *social.FacebookClient, _ context.Context, _, _ string, _ []clickhousemodels.FacebookCompetitorMinimalSharedPost) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error) {
		return nil, nil
	}

	bulkAssetsCalled := 0
	chBulkUpdateAssetsFn = func(_ *conversions.ClickHouseSink, _ context.Context, assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset) (int, error) {
		bulkAssetsCalled++
		return len(assets), nil
	}
	bulkSharedCalled := 0
	chBulkUpdateSharedPicsFn = func(_ *conversions.ClickHouseSink, _ context.Context, posts []clickhousemodels.FacebookCompetitorMinimalSharedPost) (int, error) {
		bulkSharedCalled++
		return len(posts), nil
	}

	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockCompetitorRepo{
		getAccounts: func(ctx context.Context, platformType string) ([]*mongomodels.Competitor, error) {
			return []*mongomodels.Competitor{
				{CompetitorID: "12345"},
				{CompetitorID: "67890"},
			}, nil
		},
	}, mockRedisClient{
		do: func(ctx context.Context, args ...interface{}) *redis.Cmd {
			return redisCmdWithText(ctx, `{"token":"plain_token"}`)
		},
	}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, 2)

	if bulkAssetsCalled != 1 {
		t.Fatalf("expected bulk asset update called once, got %d", bulkAssetsCalled)
	}
	if bulkSharedCalled != 0 {
		t.Fatalf("expected no bulk shared pic update (no shared posts), got %d", bulkSharedCalled)
	}
}

func TestRun_RepoError(t *testing.T) {
	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockCompetitorRepo{
		getAccounts: func(ctx context.Context, platformType string) ([]*mongomodels.Competitor, error) {
			return nil, errors.New("repo failed")
		},
	}, mockRedisClient{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, 1)
}

func TestRun_NoAccounts(t *testing.T) {
	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockCompetitorRepo{
		getAccounts: func(ctx context.Context, platformType string) ([]*mongomodels.Competitor, error) {
			if platformType != "facebook" {
				t.Fatalf("unexpected platform type: %s", platformType)
			}
			return nil, nil
		},
	}, mockRedisClient{}, &social.FacebookClient{}, &conversions.ClickHouseSink{}, 1)
}
