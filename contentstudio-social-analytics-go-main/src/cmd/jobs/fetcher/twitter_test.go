package fetcher

import (
	"context"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mt "go.mongodb.org/mongo-driver/mongo/integration/mtest"

	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

func TestShouldScheduleTwitterAccount(t *testing.T) {
	now := time.Date(2026, 2, 9, 12, 0, 0, 0, time.UTC) // Monday, 9th

	tests := []struct {
		name    string
		setting mongodb.TwitterJobSetting
		want    bool
	}{
		{name: "daily", setting: mongodb.TwitterJobSetting{JobType: "daily"}, want: true},
		{name: "weekly match", setting: mongodb.TwitterJobSetting{JobType: "weekly", TriggerDay: 1}, want: true},
		{name: "weekly no match", setting: mongodb.TwitterJobSetting{JobType: "weekly", TriggerDay: 3}, want: false},
		{name: "monthly match", setting: mongodb.TwitterJobSetting{JobType: "monthly", TriggerDay: 9}, want: true},
		{name: "monthly no match", setting: mongodb.TwitterJobSetting{JobType: "monthly", TriggerDay: 28}, want: false},
		{name: "never", setting: mongodb.TwitterJobSetting{JobType: "never"}, want: false},
		{name: "unknown", setting: mongodb.TwitterJobSetting{JobType: "hourly"}, want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := shouldScheduleTwitterAccount(tt.setting, now)
			if got != tt.want {
				t.Fatalf("shouldScheduleTwitterAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildTwitterAccountBatch_FiltersAndEnriches(t *testing.T) {
	h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))

	h.Run("filters by app and settings then enriches payload", func(m *mt.T) {
		ns := mt.TestDb + ".twitter_job_settings"
		m.AddMockResponses(
			mt.CreateCursorResponse(1, ns, mt.FirstBatch, bson.D{
				{Key: "platform_id", Value: "tw_001"},
				{Key: "job_type", Value: "daily"},
				{Key: "trigger_day", Value: 1},
				{Key: "post_count", Value: 33},
			}),
			mt.CreateCursorResponse(0, ns, mt.NextBatch),
		)

		validAppID := primitive.NewObjectID().Hex()
		accounts := []mongomodels.SocialIntegration{
			{
				ID:                 primitive.NewObjectID(),
				PlatformIdentifier: "tw_001",
				WorkspaceID:        primitive.NewObjectID(),
				OAuthToken:         "oauth_1",
				OAuthTokenSecret:   "secret_1",
				DeveloperAppID:     validAppID,
			},
			{
				ID:                 primitive.NewObjectID(),
				PlatformIdentifier: "tw_002",
				WorkspaceID:        primitive.NewObjectID(),
				OAuthToken:         "oauth_2",
				OAuthTokenSecret:   "secret_2",
				DeveloperAppID:     primitive.NewObjectID().Hex(),
			},
		}

		developerApps := map[string]mongodb.TwitterDeveloperApp{
			validAppID: {
				APIKey:    "app-key-1",
				APISecret: "app-secret-1",
			},
		}

		repo := mongodb.NewTwitterRepository(m.DB)
		batch, skipped, _, err := buildTwitterAccountBatch(
			context.Background(),
			repo,
			accounts,
			"incremental",
			zerolog.Nop(),
			developerApps,
			time.Date(2026, 2, 9, 9, 0, 0, 0, time.UTC),
		)
		if err != nil {
			t.Fatalf("buildTwitterAccountBatch() error = %v", err)
		}
		if skipped != 1 {
			t.Fatalf("skipped = %d, want %d", skipped, 1)
		}
		if len(batch) != 1 {
			t.Fatalf("len(batch) = %d, want 1", len(batch))
		}

		got := batch[0]
		if got.TwitterID != "tw_001" {
			t.Fatalf("twitter_id = %s, want tw_001", got.TwitterID)
		}
		if got.PostCount != 33 {
			t.Fatalf("post_count = %d, want 33", got.PostCount)
		}
		if got.APIKey != "app-key-1" {
			t.Fatalf("api_key = %s, want app-key-1", got.APIKey)
		}
		if got.APISecret != "app-secret-1" {
			t.Fatalf("api_secret = %s, want app-secret-1", got.APISecret)
		}
	})
}
