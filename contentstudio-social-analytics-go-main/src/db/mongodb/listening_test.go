package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

func newTestListeningRepo(mt *mtest.T) *ListeningRepository {
	log, _ := logger.NewTestLogger()
	return &ListeningRepository{
		collection: mt.Coll,
		log:        log,
	}
}

func newTestListeningWorkspaceRepo(mt *mtest.T) *ListeningWorkspaceRepository {
	log, _ := logger.NewTestLogger()
	return &ListeningWorkspaceRepository{
		workspaceCollection: mt.DB.Collection("workspace"),
		userCollection:      mt.DB.Collection("users"),
		planCollection:      mt.DB.Collection("subscription_plans"),
		log:                 log,
	}
}

func TestListeningRepository_GetTopicByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("found", func(mt *mtest.T) {
		validID := primitive.NewObjectID().Hex()
		repo := newTestListeningRepo(mt)
		expected := mongoModels.ListeningTopic{
			ID:               primitive.NewObjectID(),
			TopicID:          validID,
			WorkspaceID:      "ws-1",
			Name:             "Test Topic",
			IncludeKeywords:  []string{"go", "rust"},
			ExcludeKeywords:  []string{"spam"},
			EnabledPlatforms: []string{"twitter", "reddit"},
			MentionsLimit:    1000,
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.listening_topics", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: expected.ID},
			{Key: "topic_id", Value: expected.TopicID},
			{Key: "workspace_id", Value: expected.WorkspaceID},
			{Key: "name", Value: expected.Name},
			{Key: "include_keywords", Value: bson.A{"go", "rust"}},
			{Key: "exclude_keywords", Value: bson.A{"spam"}},
			{Key: "enabled_platforms", Value: bson.A{"twitter", "reddit"}},
			{Key: "mentions_limit", Value: expected.MentionsLimit},
			{Key: "mentions_limit_reached", Value: false},
		}))

		topic, err := repo.GetTopicByID(context.Background(), validID)
		require.NoError(t, err)
		require.NotNil(t, topic)
		assert.Equal(t, validID, topic.TopicID)
		assert.Equal(t, "Test Topic", topic.Name)
		assert.Equal(t, []string{"go", "rust"}, topic.IncludeKeywords)
		assert.Equal(t, 1000, topic.MentionsLimit)
		assert.False(t, topic.MentionsLimitReached)
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := newTestListeningRepo(mt)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.listening_topics", mtest.FirstBatch))

		topic, err := repo.GetTopicByID(context.Background(), primitive.NewObjectID().Hex())
		require.NoError(t, err)
		assert.Nil(t, topic)
	})

}

func TestListeningRepository_GetMentionsCount(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("found", func(mt *mtest.T) {
		validID := primitive.NewObjectID().Hex()
		repo := newTestListeningRepo(mt)
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.listening_topics", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "topic_id", Value: validID},
			{Key: "workspace_id", Value: "ws-1"},
			{Key: "usage", Value: bson.D{
				{Key: "mentions_count", Value: 42},
			}},
		}))

		count, err := repo.GetMentionsCount(context.Background(), validID)
		require.NoError(t, err)
		assert.Equal(t, 42, count)
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := newTestListeningRepo(mt)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.listening_topics", mtest.FirstBatch))

		_, err := repo.GetMentionsCount(context.Background(), primitive.NewObjectID().Hex())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "topic not found")
	})
}

func TestListeningRepository_GetActiveTopics(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("returns active topics", func(mt *mtest.T) {
		repo := newTestListeningRepo(mt)

		first := mtest.CreateCursorResponse(1, "db.listening_topics", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "topic_id", Value: "topic-1"},
			{Key: "mentions_limit_reached", Value: false},
		})
		second := mtest.CreateCursorResponse(1, "db.listening_topics", mtest.NextBatch, bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "topic_id", Value: "topic-2"},
			{Key: "mentions_limit_reached", Value: false},
		})
		killCursors := mtest.CreateCursorResponse(0, "db.listening_topics", mtest.NextBatch)
		mt.AddMockResponses(first, second, killCursors)

		topics, err := repo.GetActiveTopics(context.Background())
		require.NoError(t, err)
		assert.Len(t, topics, 2)
		assert.Equal(t, "topic-1", topics[0].TopicID)
		assert.Equal(t, "topic-2", topics[1].TopicID)

		started := mt.GetStartedEvent()
		require.NotNil(t, started)
		require.Equal(t, "find", started.CommandName)

		commandJSON, err := bson.MarshalExtJSON(started.Command, false, false)
		require.NoError(t, err)
		assert.Contains(t, string(commandJSON), "\"filter\":{\"deleted_at\":null}")
	})
}

func TestListeningRepository_IncrementMentionsCount(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		validID := primitive.NewObjectID().Hex()
		repo := newTestListeningRepo(mt)
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "nModified", Value: 1},
			{Key: "n", Value: 1},
		})

		err := repo.IncrementMentionsCount(context.Background(), validID, 50)
		require.NoError(t, err)
	})

	mt.Run("topic not found", func(mt *mtest.T) {
		repo := newTestListeningRepo(mt)
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "nModified", Value: 0},
			{Key: "n", Value: 0},
		})

		err := repo.IncrementMentionsCount(context.Background(), primitive.NewObjectID().Hex(), 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "topic not found")
	})
}

func TestListeningRepository_SetMentionsLimitReached(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		validID := primitive.NewObjectID().Hex()
		repo := newTestListeningRepo(mt)
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "nModified", Value: 1},
			{Key: "n", Value: 1},
		})

		err := repo.SetMentionsLimitReached(context.Background(), validID)
		require.NoError(t, err)
	})
}

func TestListeningRepository_MarkInitialSyncDone(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	tests := []struct {
		name        string
		topicID     string
		workspaceID string
		nModified   int32
		nMatched    int32
		wantApplied bool
		wantErrSub  string
	}{
		{
			name:        "applies when no prior event_timestamp",
			topicID:     primitive.NewObjectID().Hex(),
			workspaceID: "ws-1",
			nModified:   1,
			nMatched:    1,
			wantApplied: true,
		},
		{
			name:        "stale event is skipped",
			topicID:     primitive.NewObjectID().Hex(),
			workspaceID: "ws-1",
			nModified:   0,
			nMatched:    0,
			wantApplied: false,
		},
		{
			name:        "workspace mismatch is skipped",
			topicID:     primitive.NewObjectID().Hex(),
			workspaceID: "ws-other",
			nModified:   0,
			nMatched:    0,
			wantApplied: false,
		},
		{
			name:        "invalid topic ID returns error",
			topicID:     "not-an-object-id",
			workspaceID: "ws-1",
			wantErrSub:  "invalid topic ID",
		},
	}

	for _, tc := range tests {
		tc := tc
		mt.Run(tc.name, func(mt *mtest.T) {
			repo := newTestListeningRepo(mt)

			if tc.wantErrSub == "" {
				mt.AddMockResponses(bson.D{
					{Key: "ok", Value: 1},
					{Key: "n", Value: tc.nMatched},
					{Key: "nModified", Value: tc.nModified},
				})
			}

			applied, err := repo.MarkInitialSyncDone(
				context.Background(),
				tc.topicID,
				tc.workspaceID,
				time.Date(2026, 5, 3, 18, 30, 0, 0, time.UTC),
			)

			if tc.wantErrSub != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrSub)
				assert.False(t, applied)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantApplied, applied)
		})
	}

	mt.Run("filter scopes by workspace and last_event_timestamp", func(mt *mtest.T) {
		repo := newTestListeningRepo(mt)
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: int32(1)},
			{Key: "nModified", Value: int32(1)},
		})

		eventAt := time.Date(2026, 5, 3, 18, 30, 0, 0, time.UTC)
		_, err := repo.MarkInitialSyncDone(
			context.Background(),
			primitive.NewObjectID().Hex(),
			"ws-42",
			eventAt,
		)
		require.NoError(t, err)

		started := mt.GetStartedEvent()
		require.NotNil(t, started)
		require.Equal(t, "update", started.CommandName)

		commandJSON, err := bson.MarshalExtJSON(started.Command, false, false)
		require.NoError(t, err)
		assert.Contains(t, string(commandJSON), "\"workspace_id\":\"ws-42\"")
		assert.Contains(t, string(commandJSON), "last_event_timestamp")
		assert.Contains(t, string(commandJSON), "is_initial_sync_done")
	})
}

func TestListeningRepository_MarkFirstMentionsReceived(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	tests := []struct {
		name       string
		topicID    string
		nModified  int32
		nMatched   int32
		mongoErr   *mtest.WriteError
		wantWon    bool
		wantErrSub string
	}{
		{
			name:      "first call wins when first_mentions_received_at is null",
			topicID:   primitive.NewObjectID().Hex(),
			nModified: 1,
			nMatched:  1,
			wantWon:   true,
		},
		{
			name:      "second call loses race because field is already set",
			topicID:   primitive.NewObjectID().Hex(),
			nModified: 0,
			nMatched:  0,
			wantWon:   false,
		},
		{
			name:      "matched but unmodified still loses race",
			topicID:   primitive.NewObjectID().Hex(),
			nModified: 0,
			nMatched:  1,
			wantWon:   false,
		},
		{
			name:       "invalid topic ID returns error",
			topicID:    "not-an-object-id",
			wantErrSub: "invalid topic ID",
		},
		{
			name:       "mongo update error is wrapped",
			topicID:    primitive.NewObjectID().Hex(),
			mongoErr:   &mtest.WriteError{Index: 0, Code: 11000, Message: "duplicate"},
			wantErrSub: "failed to update",
		},
	}

	for _, tc := range tests {
		tc := tc
		mt.Run(tc.name, func(mt *mtest.T) {
			repo := newTestListeningRepo(mt)

			if tc.wantErrSub == "" {
				mt.AddMockResponses(bson.D{
					{Key: "ok", Value: 1},
					{Key: "n", Value: tc.nMatched},
					{Key: "nModified", Value: tc.nModified},
				})
			} else if tc.mongoErr != nil {
				mt.AddMockResponses(mtest.CreateWriteErrorsResponse(*tc.mongoErr))
			}

			won, err := repo.MarkFirstMentionsReceived(context.Background(), tc.topicID)

			if tc.wantErrSub != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrSub)
				assert.False(t, won)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantWon, won)
		})
	}

	mt.Run("filter scopes by null or missing first_mentions_received_at", func(mt *mtest.T) {
		repo := newTestListeningRepo(mt)
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: int32(1)},
			{Key: "nModified", Value: int32(1)},
		})

		_, err := repo.MarkFirstMentionsReceived(context.Background(), primitive.NewObjectID().Hex())
		require.NoError(t, err)

		started := mt.GetStartedEvent()
		require.NotNil(t, started)
		require.Equal(t, "update", started.CommandName)

		commandJSON, err := bson.MarshalExtJSON(started.Command, false, false)
		require.NoError(t, err)
		assert.Contains(t, string(commandJSON), "first_mentions_received_at")
		assert.Contains(t, string(commandJSON), "$exists")
	})
}

func TestListeningRepository_UpdateLastFetched(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	tests := []struct {
		name                 string
		cursors              map[string]string
		wantCursorField      bool
		wantCursorFieldValue string
	}{
		{
			name: "stores last_fetched_cursors when provided",
			cursors: map[string]string{
				"twitter:golang": "cursor-abc",
				"reddit:golang":  "cursor-xyz",
			},
			wantCursorField:      true,
			wantCursorFieldValue: `"twitter:golang":"cursor-abc"`,
		},
		{
			name:            "omits last_fetched_cursors when nil",
			cursors:         nil,
			wantCursorField: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		mt.Run(tc.name, func(mt *mtest.T) {
			validID := primitive.NewObjectID().Hex()
			repo := newTestListeningRepo(mt)
			mt.AddMockResponses(bson.D{
				{Key: "ok", Value: 1},
				{Key: "nModified", Value: 1},
				{Key: "n", Value: 1},
			})

			err := repo.UpdateLastFetched(context.Background(), validID, time.Now(), tc.cursors)
			require.NoError(t, err)

			started := mt.GetStartedEvent()
			require.NotNil(t, started)
			require.Equal(t, "update", started.CommandName)

			commandJSON, err := bson.MarshalExtJSON(started.Command, false, false)
			require.NoError(t, err)
			if tc.wantCursorField {
				assert.Contains(t, string(commandJSON), `"last_fetched_cursors"`)
				assert.Contains(t, string(commandJSON), tc.wantCursorFieldValue)
			} else {
				assert.NotContains(t, string(commandJSON), `"last_fetched_cursors"`)
			}
		})
	}
}

func TestListeningRepository_ResetUsage(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		validID := primitive.NewObjectID().Hex()
		repo := newTestListeningRepo(mt)
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "nModified", Value: 1},
			{Key: "n", Value: 1},
		})

		err := repo.ResetUsage(context.Background(), validID, time.Now())
		require.NoError(t, err)
	})
}

func TestListeningRepository_GetAIContext(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("invalid topic id returns error", func(mt *mtest.T) {
		repo := newTestListeningRepo(mt)

		snap, err := repo.GetAIContext(context.Background(), "not-a-hex-id")
		require.Error(t, err)
		assert.Equal(t, mongoModels.TopicContextSnapshot{}, snap)
	})

	mt.Run("not found returns empty snapshot without error", func(mt *mtest.T) {
		repo := newTestListeningRepo(mt)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.listening_topics", mtest.FirstBatch))

		snap, err := repo.GetAIContext(context.Background(), primitive.NewObjectID().Hex())
		require.NoError(t, err)
		assert.Equal(t, mongoModels.TopicContextSnapshot{}, snap)
	})

	mt.Run("top-level fields win when populated", func(mt *mtest.T) {
		validID := primitive.NewObjectID().Hex()
		repo := newTestListeningRepo(mt)

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.listening_topics", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "name", Value: "Acme Brand"},
			{Key: "type", Value: "own_brand"},
			{Key: "ai_context", Value: bson.D{
				{Key: "brand_name", Value: "Acme"},
				{Key: "brand_keywords", Value: bson.A{"acme"}},
				{Key: "industry", Value: "SaaS"},
				{Key: "competitors", Value: bson.A{
					bson.D{{Key: "name", Value: "Foo"}, {Key: "keywords", Value: bson.A{"foo"}}},
				}},
			}},
			{Key: "ai_context_hint", Value: "B2B only"},
			{Key: "include_keywords", Value: bson.A{"acme", "saas"}},
			{Key: "query", Value: bson.D{
				{Key: "ai_context", Value: bson.D{{Key: "brand_name", Value: "SHOULD_NOT_USE"}}},
				{Key: "ai_context_hint", Value: "should-not-use"},
				{Key: "include_keywords", Value: bson.A{"should-not-use"}},
			}},
		}))

		snap, err := repo.GetAIContext(context.Background(), validID)
		require.NoError(t, err)
		assert.Equal(t, "Acme Brand", snap.TopicName)
		assert.Equal(t, "own_brand", snap.TopicType)
		assert.Equal(t, "Acme", snap.AIContext.BrandName)
		assert.Equal(t, "SaaS", snap.AIContext.Industry)
		assert.Equal(t, []string{"acme"}, snap.AIContext.BrandKeywords)
		assert.Equal(t, "B2B only", snap.Hint)
		assert.Equal(t, []string{"acme", "saas"}, snap.TopicKeywords)
		require.Len(t, snap.AIContext.Competitors, 1)
		assert.Equal(t, "Foo", snap.AIContext.Competitors[0].Name)
	})

	mt.Run("query.* fields used when top-level empty", func(mt *mtest.T) {
		validID := primitive.NewObjectID().Hex()
		repo := newTestListeningRepo(mt)

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.listening_topics", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "name", Value: "Laravel-Created"},
			{Key: "type", Value: "industry"},
			{Key: "query", Value: bson.D{
				{Key: "ai_context", Value: bson.D{
					{Key: "brand_name", Value: "Laravel-Brand"},
					{Key: "industry", Value: "FinTech"},
				}},
				{Key: "ai_context_hint", Value: "from-query"},
				{Key: "include_keywords", Value: bson.A{"loan", "credit"}},
			}},
		}))

		snap, err := repo.GetAIContext(context.Background(), validID)
		require.NoError(t, err)
		assert.Equal(t, "Laravel-Brand", snap.AIContext.BrandName)
		assert.Equal(t, "FinTech", snap.AIContext.Industry)
		assert.Equal(t, "from-query", snap.Hint)
		assert.Equal(t, []string{"loan", "credit"}, snap.TopicKeywords)
	})

	mt.Run("partial fallback: top-level hint set, query ai_context used", func(mt *mtest.T) {
		validID := primitive.NewObjectID().Hex()
		repo := newTestListeningRepo(mt)

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.listening_topics", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: primitive.NewObjectID()},
			{Key: "ai_context_hint", Value: "top-level-hint"},
			{Key: "query", Value: bson.D{
				{Key: "ai_context", Value: bson.D{{Key: "brand_name", Value: "FromQuery"}}},
				{Key: "ai_context_hint", Value: "should-not-overwrite"},
			}},
		}))

		snap, err := repo.GetAIContext(context.Background(), validID)
		require.NoError(t, err)
		assert.Equal(t, "FromQuery", snap.AIContext.BrandName)
		assert.Equal(t, "top-level-hint", snap.Hint)
	})
}

func TestToInt(t *testing.T) {
	dec, err := primitive.ParseDecimal128("5000")
	require.NoError(t, err)

	cases := []struct {
		name string
		in   interface{}
		want int
	}{
		{"nil", nil, 0},
		{"int", 5000, 5000},
		{"int32", int32(5000), 5000},
		{"int64", int64(5000), 5000},
		{"float64", float64(5000), 5000},
		{"bool true", true, 1},
		{"bool false", false, 0},
		{"string numeric", "5000", 5000},
		{"string padded", "  5000 ", 5000},
		{"string float", "5000.7", 5000},
		{"string empty", "", 0},
		{"string garbage", "abc", 0},
		{"decimal128", dec, 5000},
		{"unsupported", []int{1, 2}, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, toInt(tc.in))
		})
	}
}

func TestListeningWorkspaceRepository_resolveMentionLimit(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	dec5000, err := primitive.ParseDecimal128("5000")
	require.NoError(t, err)

	type planResp struct {
		present bool
		doc     bson.D
	}

	cases := []struct {
		name        string
		user        bson.D   // nil means user not found
		plan        planResp // present=false simulates ErrNoDocuments
		want        int
		wantHasSub  bool
		wantErr     bool
	}{
		{
			name: "user missing returns 0 with no plan lookup",
			user: nil,
			want: 0,
		},
		{
			name: "no subscription_id returns addon only",
			user: bson.D{
				{Key: "_id", Value: "user-1"},
				{Key: "addons", Value: bson.D{{Key: "listening_mentions", Value: 750}}},
				{Key: "has_social_listening_subscription", Value: true},
			},
			want:       750,
			wantHasSub: true,
		},
		{
			name: "plan int limit",
			user: bson.D{
				{Key: "_id", Value: "user-1"},
				{Key: "subscription_id", Value: "plan-1"},
				{Key: "stackable", Value: 1},
				{Key: "addons", Value: bson.D{}},
				{Key: "has_social_listening_subscription", Value: true},
			},
			plan: planResp{present: true, doc: bson.D{
				{Key: "_id", Value: "plan-1"},
				{Key: "limits", Value: bson.D{{Key: "listening_mentions", Value: 10000}}},
			}},
			want:       10000,
			wantHasSub: true,
		},
		{
			name: "plan limit stored as string still parses (the actual bug)",
			user: bson.D{
				{Key: "_id", Value: "user-1"},
				{Key: "subscription_id", Value: "plan-1"},
				{Key: "stackable", Value: 1},
				{Key: "addons", Value: bson.D{}},
				{Key: "has_social_listening_subscription", Value: true},
			},
			plan: planResp{present: true, doc: bson.D{
				{Key: "_id", Value: "plan-1"},
				{Key: "limits", Value: bson.D{{Key: "listening_mentions", Value: "5000"}}},
			}},
			want:       5000,
			wantHasSub: true,
		},
		{
			name: "plan limit stored as Decimal128 still parses",
			user: bson.D{
				{Key: "_id", Value: "user-1"},
				{Key: "subscription_id", Value: "plan-1"},
				{Key: "stackable", Value: 1},
				{Key: "addons", Value: bson.D{}},
				{Key: "has_social_listening_subscription", Value: true},
			},
			plan: planResp{present: true, doc: bson.D{
				{Key: "_id", Value: "plan-1"},
				{Key: "limits", Value: bson.D{{Key: "listening_mentions", Value: dec5000}}},
			}},
			want:       5000,
			wantHasSub: true,
		},
		{
			name: "stackable greater than 1 multiplies plan limit",
			user: bson.D{
				{Key: "_id", Value: "user-1"},
				{Key: "subscription_id", Value: "plan-1"},
				{Key: "stackable", Value: 3},
				{Key: "addons", Value: bson.D{}},
				{Key: "has_social_listening_subscription", Value: true},
			},
			plan: planResp{present: true, doc: bson.D{
				{Key: "_id", Value: "plan-1"},
				{Key: "limits", Value: bson.D{{Key: "listening_mentions", Value: 1000}}},
			}},
			want:       3000,
			wantHasSub: true,
		},
		{
			name: "addon string adds onto plan",
			user: bson.D{
				{Key: "_id", Value: "user-1"},
				{Key: "subscription_id", Value: "plan-1"},
				{Key: "stackable", Value: 1},
				{Key: "addons", Value: bson.D{{Key: "listening_mentions", Value: "2500"}}},
				{Key: "has_social_listening_subscription", Value: true},
			},
			plan: planResp{present: true, doc: bson.D{
				{Key: "_id", Value: "plan-1"},
				{Key: "limits", Value: bson.D{{Key: "listening_mentions", Value: 1000}}},
			}},
			want:       3500,
			wantHasSub: true,
		},
		{
			name: "plan not found and no addon resolves to zero with no subscription flag",
			user: bson.D{
				{Key: "_id", Value: "user-1"},
				{Key: "subscription_id", Value: "plan-missing"},
				{Key: "stackable", Value: 1},
				{Key: "addons", Value: bson.D{}},
				{Key: "has_social_listening_subscription", Value: false},
			},
			plan:       planResp{present: false},
			want:       0,
			wantHasSub: false,
		},
	}

	for _, tc := range cases {
		mt.Run(tc.name, func(mt *mtest.T) {
			repo := newTestListeningWorkspaceRepo(mt)

			if tc.user == nil {
				mt.AddMockResponses(
					mtest.CreateCursorResponse(0, "test.users", mtest.FirstBatch),
				)
			} else {
				mt.AddMockResponses(
					mtest.CreateCursorResponse(1, "test.users", mtest.FirstBatch, tc.user),
					mtest.CreateCursorResponse(0, "test.users", mtest.NextBatch),
				)
				if subFound := bsonHasField(tc.user, "subscription_id"); subFound {
					if tc.plan.present {
						mt.AddMockResponses(
							mtest.CreateCursorResponse(1, "test.subscription_plans", mtest.FirstBatch, tc.plan.doc),
							mtest.CreateCursorResponse(0, "test.subscription_plans", mtest.NextBatch),
						)
					} else {
						mt.AddMockResponses(
							mtest.CreateCursorResponse(0, "test.subscription_plans", mtest.FirstBatch),
						)
					}
				}
			}

			got, hasSub, err := repo.resolveMentionLimit(context.Background(), "user-1")
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantHasSub, hasSub)
		})
	}
}

func bsonHasField(d bson.D, key string) bool {
	for _, e := range d {
		if e.Key == key {
			return true
		}
	}
	return false
}

func TestListeningWorkspaceRepository_IncrementWorkspaceMentionsCount(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("owner has no workspaces returns silently", func(mt *mtest.T) {
		repo := newTestListeningWorkspaceRepo(mt)

		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "test.workspace", mtest.FirstBatch),
			mtest.CreateCursorResponse(0, "test.workspace", mtest.FirstBatch),
		)

		newCount, mentionLimit, err := repo.IncrementWorkspaceMentionsCount(context.Background(), "orphan-id", 1)
		require.NoError(t, err)
		assert.Equal(t, 0, newCount)
		assert.Equal(t, 0, mentionLimit)
	})

	mt.Run("propagation failure is logged but does not fail increment", func(mt *mtest.T) {
		repo := newTestListeningWorkspaceRepo(mt)

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.workspace", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "workspace-1"},
				{Key: "user_id", Value: "user-1"},
				{Key: "used_mention_credits", Value: 5},
			}),
			mtest.CreateCursorResponse(0, "test.workspace", mtest.NextBatch),
			mtest.CreateCursorResponse(1, "test.workspace", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "workspace-1"},
				{Key: "user_id", Value: "user-1"},
				{Key: "used_mention_credits", Value: 5},
			}),
			mtest.CreateCursorResponse(0, "test.workspace", mtest.NextBatch),
			mtest.CreateCursorResponse(1, "test.users", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "user-1"},
				{Key: "subscription_id", Value: "plan-1"},
				{Key: "stackable", Value: 1},
				{Key: "addons", Value: bson.D{}},
				{Key: "has_social_listening_subscription", Value: true},
			}),
			mtest.CreateCursorResponse(0, "test.users", mtest.NextBatch),
			mtest.CreateCursorResponse(1, "test.subscription_plans", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "plan-1"},
				{Key: "limits", Value: bson.D{{Key: "listening_mentions", Value: 10000}}},
			}),
			mtest.CreateCursorResponse(0, "test.subscription_plans", mtest.NextBatch),
			mtest.CreateSuccessResponse(
				bson.E{Key: "value", Value: bson.D{
					{Key: "_id", Value: "workspace-1"},
					{Key: "user_id", Value: "user-1"},
					{Key: "used_mention_credits", Value: 6},
				}},
			),
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Message: "mirror fail",
				Code:    123,
			}),
		)

		newCount, mentionLimit, err := repo.IncrementWorkspaceMentionsCount(context.Background(), "workspace-1", 1)
		require.NoError(t, err)
		assert.Equal(t, 6, newCount)
		assert.Equal(t, 10000, mentionLimit)
	})

	mt.Run("increments usage even when social listening subscription is inactive", func(mt *mtest.T) {
		repo := newTestListeningWorkspaceRepo(mt)

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.workspace", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "workspace-1"},
				{Key: "user_id", Value: "user-1"},
				{Key: "used_mention_credits", Value: 5},
			}),
			mtest.CreateCursorResponse(0, "test.workspace", mtest.NextBatch),
			mtest.CreateCursorResponse(1, "test.workspace", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "workspace-1"},
				{Key: "user_id", Value: "user-1"},
				{Key: "used_mention_credits", Value: 5},
			}),
			mtest.CreateCursorResponse(0, "test.workspace", mtest.NextBatch),
			mtest.CreateCursorResponse(1, "test.users", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "user-1"},
				{Key: "subscription_id", Value: "plan-1"},
				{Key: "stackable", Value: 1},
				{Key: "addons", Value: bson.D{}},
				{Key: "has_social_listening_subscription", Value: false},
			}),
			mtest.CreateCursorResponse(0, "test.users", mtest.NextBatch),
			mtest.CreateCursorResponse(1, "test.subscription_plans", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "plan-1"},
				{Key: "limits", Value: bson.D{{Key: "listening_mentions", Value: 10000}}},
			}),
			mtest.CreateCursorResponse(0, "test.subscription_plans", mtest.NextBatch),
			mtest.CreateSuccessResponse(
				bson.E{Key: "value", Value: bson.D{
					{Key: "_id", Value: "workspace-1"},
					{Key: "user_id", Value: "user-1"},
					{Key: "used_mention_credits", Value: 6},
				}},
			),
			mtest.CreateSuccessResponse(),
		)

		newCount, mentionLimit, err := repo.IncrementWorkspaceMentionsCount(context.Background(), "workspace-1", 1)
		require.NoError(t, err)
		assert.Equal(t, 6, newCount)
		assert.Equal(t, 10000, mentionLimit)
	})
}

func TestListeningWorkspaceRepository_GetWorkspaceUsage(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("missing social listening subscription flag blocks usage", func(mt *mtest.T) {
		repo := newTestListeningWorkspaceRepo(mt)

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.workspace", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "workspace-1"},
				{Key: "user_id", Value: "user-1"},
				{Key: "used_mention_credits", Value: 5},
			}),
			mtest.CreateCursorResponse(0, "test.workspace", mtest.NextBatch),
			mtest.CreateCursorResponse(1, "test.workspace", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "workspace-1"},
				{Key: "user_id", Value: "user-1"},
				{Key: "used_mention_credits", Value: 5},
			}),
			mtest.CreateCursorResponse(0, "test.workspace", mtest.NextBatch),
			mtest.CreateCursorResponse(1, "test.users", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "user-1"},
				{Key: "subscription_id", Value: "plan-1"},
				{Key: "stackable", Value: 1},
				{Key: "addons", Value: bson.D{{Key: "listening_mentions", Value: 500}}},
				{Key: "has_social_listening_subscription", Value: false},
			}),
			mtest.CreateCursorResponse(0, "test.users", mtest.NextBatch),
			mtest.CreateCursorResponse(1, "test.subscription_plans", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "plan-1"},
				{Key: "limits", Value: bson.D{{Key: "listening_mentions", Value: 10000}}},
			}),
			mtest.CreateCursorResponse(0, "test.subscription_plans", mtest.NextBatch),
		)

		mentionsCount, mentionLimit, exists, err := repo.GetWorkspaceUsage(context.Background(), "workspace-1")
		require.NoError(t, err)
		assert.Equal(t, 0, mentionsCount)
		assert.Equal(t, 0, mentionLimit)
		assert.False(t, exists)
	})
}
