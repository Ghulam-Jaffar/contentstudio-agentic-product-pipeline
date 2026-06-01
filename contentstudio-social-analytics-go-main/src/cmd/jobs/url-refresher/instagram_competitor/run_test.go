package instagram_competitor

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

func testInstagramCompetitorLogger() *zerolog.Logger {
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

// processAccountJob now returns resolved posts/profilePictureURL in AccountResult without
// writing to ClickHouse. The bulk write happens once in Run() after all workers finish.

func TestProcessAccountJob_Success(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := igGetURLsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		igGetURLsFn = origGetURLs
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, instagramID int64, _, _ int) ([]clickhousemodels.InstagramCompetitorMinimalPost, error) {
		if instagramID != 98765 {
			t.Fatalf("unexpected instagram id: %d", instagramID)
		}
		return []clickhousemodels.InstagramCompetitorMinimalPost{{InstagramID: 98765, PostID: "p1"}}, nil
	}
	igGetURLsFn = func(_ *social.InstagramClient, _ context.Context, username string, posts []clickhousemodels.InstagramCompetitorMinimalPost, accessToken string, businessAccountID string) ([]clickhousemodels.InstagramCompetitorMinimalPost, string, error) {
		if username != "brand_slug" || accessToken != "plain_token" || businessAccountID != "biz_1" || len(posts) != 1 {
			t.Fatalf("unexpected refresh inputs: user=%s token=%s business=%s posts=%d", username, accessToken, businessAccountID, len(posts))
		}
		return []clickhousemodels.InstagramCompetitorMinimalPost{{InstagramID: 98765, PostID: "p1", MediaURL: "https://example.com/p1.jpg"}}, "https://example.com/profile.jpg", nil
	}

	res := processAccountJob(context.Background(), &config.Config{DecryptionKey: "unused"}, mockRedisClient{
		do: func(ctx context.Context, args ...interface{}) *redis.Cmd {
			return redisCmdWithText(ctx, `{"token":"plain_token","platform_id":"biz_1"}`)
		},
	}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, &mongomodels.Competitor{CompetitorID: int64(98765), Slug: "brand_slug"}, testInstagramCompetitorLogger())

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.InstagramID != 98765 {
		t.Fatalf("unexpected instagram_id: %d", res.InstagramID)
	}
	if res.PostsFound != 1 {
		t.Fatalf("expected PostsFound=1, got %d", res.PostsFound)
	}
	if res.URLsResolved != 1 || len(res.ResolvedPosts) != 1 {
		t.Fatalf("expected 1 resolved post, got URLsResolved=%d ResolvedPosts=%d", res.URLsResolved, len(res.ResolvedPosts))
	}
	if res.ProfilePictureURL != "https://example.com/profile.jpg" {
		t.Fatalf("unexpected profile picture url: %s", res.ProfilePictureURL)
	}
	if res.ResolvedPosts[0].MediaURL != "https://example.com/p1.jpg" {
		t.Fatalf("unexpected resolved media url: %s", res.ResolvedPosts[0].MediaURL)
	}
}

func TestProcessAccountJob_InvalidCompetitorID(t *testing.T) {
	res := processAccountJob(context.Background(), &config.Config{}, mockRedisClient{}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, &mongomodels.Competitor{CompetitorID: "not-a-number"}, testInstagramCompetitorLogger())
	if res.Err == nil {
		t.Fatalf("expected invalid id error, got %+v", res)
	}
}

func TestProcessAccountJob_TokenError(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := igGetURLsFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		igGetURLsFn = origGetURLs
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ int64, _, _ int) ([]clickhousemodels.InstagramCompetitorMinimalPost, error) {
		return []clickhousemodels.InstagramCompetitorMinimalPost{{InstagramID: 98765, PostID: "p1"}}, nil
	}
	igGetURLsFn = func(_ *social.InstagramClient, _ context.Context, _ string, _ []clickhousemodels.InstagramCompetitorMinimalPost, _, _ string) ([]clickhousemodels.InstagramCompetitorMinimalPost, string, error) {
		t.Fatal("unexpected competitor refresh call")
		return nil, "", nil
	}

	res := processAccountJob(context.Background(), &config.Config{}, mockRedisClient{
		do: func(ctx context.Context, args ...interface{}) *redis.Cmd {
			return redisCmdWithErr(ctx, errors.New("redis failed"))
		},
	}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, &mongomodels.Competitor{CompetitorID: int64(98765), Slug: "brand_slug"}, testInstagramCompetitorLogger())
	if res.Err == nil {
		t.Fatalf("expected token error, got %+v", res)
	}
}

func TestProcessAccountJob_NoPosts(t *testing.T) {
	origGetPosts := chGetPostsFn
	t.Cleanup(func() { chGetPostsFn = origGetPosts })

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ int64, _, _ int) ([]clickhousemodels.InstagramCompetitorMinimalPost, error) {
		return nil, nil
	}

	res := processAccountJob(context.Background(), &config.Config{}, mockRedisClient{}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, &mongomodels.Competitor{CompetitorID: int64(98765), Slug: "brand_slug"}, testInstagramCompetitorLogger())
	if res.Err != nil || res.PostsFound != 0 || len(res.ResolvedPosts) != 0 {
		t.Fatalf("expected empty result for no posts, got %+v", res)
	}
}

func TestRun_BulkUpdateCalledOnce(t *testing.T) {
	origGetPosts := chGetPostsFn
	origGetURLs := igGetURLsFn
	origBulkUpdate := chBulkUpdateFn
	t.Cleanup(func() {
		chGetPostsFn = origGetPosts
		igGetURLsFn = origGetURLs
		chBulkUpdateFn = origBulkUpdate
	})

	chGetPostsFn = func(_ *conversions.ClickHouseSink, _ context.Context, _ int64, _, _ int) ([]clickhousemodels.InstagramCompetitorMinimalPost, error) {
		return []clickhousemodels.InstagramCompetitorMinimalPost{{InstagramID: 98765, PostID: "p1"}}, nil
	}
	igGetURLsFn = func(_ *social.InstagramClient, _ context.Context, _ string, posts []clickhousemodels.InstagramCompetitorMinimalPost, _, _ string) ([]clickhousemodels.InstagramCompetitorMinimalPost, string, error) {
		return []clickhousemodels.InstagramCompetitorMinimalPost{{InstagramID: 98765, PostID: "p1", MediaURL: "https://example.com/p1.jpg"}}, "https://example.com/profile.jpg", nil
	}

	bulkCalled := 0
	chBulkUpdateFn = func(_ *conversions.ClickHouseSink, _ context.Context, posts []clickhousemodels.InstagramCompetitorMinimalPost, profilePics map[int64]string) (int, error) {
		bulkCalled++
		return len(posts), nil
	}

	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockCompetitorRepo{
		getAccounts: func(ctx context.Context, platformType string) ([]*mongomodels.Competitor, error) {
			return []*mongomodels.Competitor{
				{CompetitorID: int64(98765), Slug: "brand_slug"},
				{CompetitorID: int64(11111), Slug: "other_slug"},
			}, nil
		},
	}, mockRedisClient{
		do: func(ctx context.Context, args ...interface{}) *redis.Cmd {
			return redisCmdWithText(ctx, `{"token":"plain_token","platform_id":"biz_1"}`)
		},
	}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, 2)

	if bulkCalled != 1 {
		t.Fatalf("expected bulk update called once, got %d", bulkCalled)
	}
}

func TestRun_RepoError(t *testing.T) {
	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockCompetitorRepo{
		getAccounts: func(ctx context.Context, platformType string) ([]*mongomodels.Competitor, error) {
			return nil, errors.New("repo failed")
		},
	}, mockRedisClient{}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, 1)
}

func TestRun_NoAccounts(t *testing.T) {
	log := appLogger.New("info")
	Run(context.Background(), &config.Config{}, *log, mockCompetitorRepo{
		getAccounts: func(ctx context.Context, platformType string) ([]*mongomodels.Competitor, error) {
			if platformType != "instagram" {
				t.Fatalf("unexpected platform type: %s", platformType)
			}
			return nil, nil
		},
	}, mockRedisClient{}, &social.InstagramClient{}, &conversions.ClickHouseSink{}, 1)
}
