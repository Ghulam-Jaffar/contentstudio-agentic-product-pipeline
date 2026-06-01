package mongodb

import (
	"context"
	"strings"
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

func existingListeningViewDoc(name, viewType string) bson.D {
	return bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "workspace_id", Value: "ws-1"},
		{Key: "name", Value: name},
		{Key: "type", Value: viewType},
	}
}

func mustMarshalCommandJSON(t *testing.T, command bson.Raw) string {
	t.Helper()

	data, err := bson.MarshalExtJSON(command, false, false)
	require.NoError(t, err)
	return string(data)
}

func newTestListeningViewsRepo(mt *mtest.T) *ListeningViewsRepository {
	log, _ := logger.NewTestLogger()
	return &ListeningViewsRepository{
		collection: mt.Coll,
		log:        log,
	}
}

func TestListeningViewsRepository_GetViewByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("found", func(mt *mtest.T) {
		repo := newTestListeningViewsRepo(mt)
		id := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.listening_views", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: id},
			{Key: "workspace_id", Value: "ws-1"},
			{Key: "name", Value: "All Mentions"},
			{Key: "icon", Value: "Inbox"},
			{Key: "type", Value: "system"},
		}))

		view, err := repo.GetViewByID(context.Background(), id.Hex())
		require.NoError(t, err)
		require.NotNil(t, view)
		assert.Equal(t, "ws-1", view.WorkspaceID)
		assert.Equal(t, "All Mentions", view.Name)
		assert.Equal(t, "system", view.Type)
	})

	mt.Run("not found returns nil without error", func(mt *mtest.T) {
		repo := newTestListeningViewsRepo(mt)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.listening_views", mtest.FirstBatch))

		view, err := repo.GetViewByID(context.Background(), primitive.NewObjectID().Hex())
		require.NoError(t, err)
		assert.Nil(t, view)
	})

	mt.Run("invalid id returns error", func(mt *mtest.T) {
		repo := newTestListeningViewsRepo(mt)

		_, err := repo.GetViewByID(context.Background(), "not-a-valid-hex-id")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ID")
	})
}

func TestListeningViewsRepository_ListViews(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("returns views for workspace", func(mt *mtest.T) {
		repo := newTestListeningViewsRepo(mt)

		first := mtest.CreateCursorResponse(1, "db.listening_views", mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: primitive.NewObjectID()},
				{Key: "workspace_id", Value: "ws-1"},
				{Key: "name", Value: "All Mentions"},
				{Key: "type", Value: "system"},
			},
		)
		second := mtest.CreateCursorResponse(1, "db.listening_views", mtest.NextBatch,
			bson.D{
				{Key: "_id", Value: primitive.NewObjectID()},
				{Key: "workspace_id", Value: "ws-1"},
				{Key: "name", Value: "My View"},
				{Key: "type", Value: "user"},
			},
		)
		kill := mtest.CreateCursorResponse(0, "db.listening_views", mtest.NextBatch)
		mt.AddMockResponses(first, second, kill)

		views, err := repo.ListViews(context.Background(), "ws-1")
		require.NoError(t, err)
		assert.Len(t, views, 2)
	})

	mt.Run("returns empty slice when no views", func(mt *mtest.T) {
		repo := newTestListeningViewsRepo(mt)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.listening_views", mtest.FirstBatch))

		views, err := repo.ListViews(context.Background(), "ws-empty")
		require.NoError(t, err)
		assert.Empty(t, views)
	})
}

func TestListeningViewsRepository_CreateView(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("creates view and returns it with id", func(mt *mtest.T) {
		repo := newTestListeningViewsRepo(mt)
		newID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "insertedId", Value: newID},
		))

		view := &mongoModels.ListeningView{
			WorkspaceID: "ws-1",
			Name:        "My View",
			Icon:        "Star",
			Type:        "user",
		}
		created, err := repo.CreateView(context.Background(), view)
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.False(t, created.ID.IsZero())
		assert.Equal(t, "My View", created.Name)
		assert.False(t, created.CreatedAt.IsZero())
	})
}

func TestListeningViewsRepository_UpdateView(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("returns updated view", func(mt *mtest.T) {
		repo := newTestListeningViewsRepo(mt)
		id := primitive.NewObjectID()
		now := time.Now().UTC().Truncate(time.Millisecond)

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "value", Value: bson.D{
				{Key: "_id", Value: id},
				{Key: "workspace_id", Value: "ws-1"},
				{Key: "name", Value: "Updated"},
				{Key: "type", Value: "user"},
				{Key: "updated_at", Value: primitive.NewDateTimeFromTime(now)},
			}},
		))

		updated, err := repo.UpdateView(context.Background(), id.Hex(), bson.M{"name": "Updated"})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "Updated", updated.Name)
	})

	mt.Run("returns nil when view not found after update", func(mt *mtest.T) {
		repo := newTestListeningViewsRepo(mt)
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "value", Value: nil},
		))

		updated, err := repo.UpdateView(context.Background(), primitive.NewObjectID().Hex(), bson.M{"name": "X"})
		require.NoError(t, err)
		assert.Nil(t, updated)
	})

	mt.Run("invalid id returns error", func(mt *mtest.T) {
		repo := newTestListeningViewsRepo(mt)

		_, err := repo.UpdateView(context.Background(), "bad-id", bson.M{"name": "X"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ID")
	})
}

func TestListeningViewsRepository_DeleteView(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("deletes existing view", func(mt *mtest.T) {
		repo := newTestListeningViewsRepo(mt)

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: 1},
		})

		err := repo.DeleteView(context.Background(), primitive.NewObjectID().Hex())
		require.NoError(t, err)
	})

	mt.Run("returns error when view not found", func(mt *mtest.T) {
		repo := newTestListeningViewsRepo(mt)

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: 0},
		})

		err := repo.DeleteView(context.Background(), primitive.NewObjectID().Hex())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "view not found")
	})

	mt.Run("invalid id returns error", func(mt *mtest.T) {
		repo := newTestListeningViewsRepo(mt)

		err := repo.DeleteView(context.Background(), "not-hex")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ID")
	})
}

func TestListeningViewsRepository_SeedSystemViews(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	allSeedNames := []string{
		"All Mentions",
		"High Relevance",
		"Crisis Management",
		"Brand Monitoring",
		"Competitor Intel",
		"Buy Intent",
		"Brand Love",
	}

	tests := []struct {
		name            string
		existingDocs    []bson.D
		wantInsert      bool
		wantNames       []string
		wantAbsentNames []string
	}{
		{
			// Fresh workspace: seed both system slots AND the 5 brand presets.
			name:         "fresh workspace seeds both system and preset views",
			existingDocs: nil,
			wantInsert:   true,
			wantNames:    allSeedNames,
		},
		{
			// Workspace already has at least one view → do NOT re-seed presets.
			// Only fill in any missing system slot.
			name: "populated workspace only re-seeds missing system slots",
			existingDocs: []bson.D{
				existingListeningViewDoc("All Mentions", "system"),
				existingListeningViewDoc("Brand Love", "preset"),
			},
			wantInsert: true,
			wantNames:  []string{"High Relevance"},
			wantAbsentNames: []string{
				"All Mentions",
				"Crisis Management",
				"Brand Monitoring",
				"Competitor Intel",
				"Buy Intent",
				"Brand Love",
			},
		},
		{
			// Both system slots present → no inserts, regardless of which presets exist.
			name: "no insert when both system slots already exist",
			existingDocs: []bson.D{
				existingListeningViewDoc("All Mentions", "system"),
				existingListeningViewDoc("High Relevance", "system"),
				existingListeningViewDoc("Crisis Management", "preset"),
			},
			wantInsert: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		mt.Run(tc.name, func(mt *mtest.T) {
			repo := newTestListeningViewsRepo(mt)

			// Migration UpdateMany — runs unconditionally, so always mock it.
			responses := []bson.D{
				{{Key: "ok", Value: 1}, {Key: "n", Value: 0}, {Key: "nModified", Value: 0}},
				mtest.CreateCursorResponse(0, "db.listening_views", mtest.FirstBatch, tc.existingDocs...),
			}
			existingSystemCount := 0
			for _, doc := range tc.existingDocs {
				for _, field := range doc {
					if field.Key == "type" && field.Value == "system" {
						existingSystemCount++
						break
					}
				}
			}
			for i := 0; i < existingSystemCount; i++ {
				responses = append(responses, mtest.CreateSuccessResponse())
			}
			if tc.wantInsert {
				responses = append(responses, mtest.CreateSuccessResponse())
			}
			mt.AddMockResponses(responses...)

			err := repo.SeedSystemViews(context.Background(), "ws-1")
			require.NoError(t, err)

			startedEvents := mt.GetAllStartedEvents()
			insertEvents := 0
			for _, event := range startedEvents {
				if event.CommandName != "insert" {
					continue
				}

				insertEvents++
				commandJSON := mustMarshalCommandJSON(t, event.Command)

				assert.Equal(t, len(tc.wantNames), strings.Count(commandJSON, `"workspace_id":"ws-1"`))
				for _, name := range tc.wantNames {
					assert.Contains(t, commandJSON, `"name":"`+name+`"`)
				}
				for _, name := range tc.wantAbsentNames {
					assert.NotContains(t, commandJSON, `"name":"`+name+`"`)
				}
			}

			if tc.wantInsert {
				assert.Equal(t, 1, insertEvents)
			} else {
				assert.Equal(t, 0, insertEvents)
			}
		})
	}
}

// migration test: legacy workspaces stored brand presets as type="system";
// SeedSystemViews must flip them to type="preset" on the fly.
func TestListeningViewsRepository_SeedSystemViews_MigratesLegacyPresets(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("issues UpdateMany targeting legacy preset names with type=system", func(mt *mtest.T) {
		repo := newTestListeningViewsRepo(mt)

		mt.AddMockResponses(
			bson.D{{Key: "ok", Value: 1}, {Key: "n", Value: 5}, {Key: "nModified", Value: 5}},
			mtest.CreateCursorResponse(0, "db.listening_views", mtest.FirstBatch,
				existingListeningViewDoc("All Mentions", "system"),
				existingListeningViewDoc("High Relevance", "system"),
			),
			mtest.CreateSuccessResponse(),
			mtest.CreateSuccessResponse(),
		)

		err := repo.SeedSystemViews(context.Background(), "ws-1")
		require.NoError(t, err)

		var updateCmd bson.Raw
		for _, event := range mt.GetAllStartedEvents() {
			if event.CommandName == "update" {
				updateCmd = event.Command
				break
			}
		}
		require.NotNil(t, updateCmd, "expected an update command for legacy preset migration")

		commandJSON := mustMarshalCommandJSON(t, updateCmd)
		assert.Contains(t, commandJSON, `"type":"system"`)
		assert.Contains(t, commandJSON, `"type":"preset"`)
		for _, name := range []string{"Crisis Management", "Brand Monitoring", "Competitor Intel", "Buy Intent", "Brand Love"} {
			assert.Contains(t, commandJSON, `"`+name+`"`)
		}
	})
}

func TestListeningViewsRepository_SeedSystemViews_AssignsSystemKeysAndDemotesDuplicateSystems(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	tests := []struct {
		name            string
		existingDocs    []bson.D
		wantSystemKeys  []string
		wantDemotion    bool
		wantInsertCount int
	}{
		{
			name: "renamed system views receive canonical keys and extras are demoted",
			existingDocs: []bson.D{
				{
					{Key: "_id", Value: primitive.NewObjectID()},
					{Key: "workspace_id", Value: "ws-1"},
					{Key: "name", Value: "Inbox Zero"},
					{Key: "type", Value: "system"},
					{Key: "created_at", Value: time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)},
				},
				{
					{Key: "_id", Value: primitive.NewObjectID()},
					{Key: "workspace_id", Value: "ws-1"},
					{Key: "name", Value: "Priority Mentions"},
					{Key: "type", Value: "system"},
					{Key: "filter_preset", Value: bson.D{
						{Key: "exclude_ai_tags", Value: bson.A{"Irrelevant"}},
					}},
					{Key: "created_at", Value: time.Date(2026, 5, 2, 9, 0, 0, 0, time.UTC)},
				},
				{
					{Key: "_id", Value: primitive.NewObjectID()},
					{Key: "workspace_id", Value: "ws-1"},
					{Key: "name", Value: "Duplicate System"},
					{Key: "type", Value: "system"},
					{Key: "created_at", Value: time.Date(2026, 5, 3, 9, 0, 0, 0, time.UTC)},
				},
			},
			wantSystemKeys:  []string{`"system_key":"all_mentions"`, `"system_key":"high_relevance"`},
			wantDemotion:    true,
			wantInsertCount: 0,
		},
	}

	for _, tc := range tests {
		tc := tc
		mt.Run(tc.name, func(mt *mtest.T) {
			repo := newTestListeningViewsRepo(mt)

			responses := []bson.D{
				{{Key: "ok", Value: 1}, {Key: "n", Value: 0}, {Key: "nModified", Value: 0}},
				mtest.CreateCursorResponse(0, "db.listening_views", mtest.FirstBatch, tc.existingDocs...),
			}
			for range tc.wantSystemKeys {
				responses = append(responses, mtest.CreateSuccessResponse())
			}
			if tc.wantDemotion {
				responses = append(responses, mtest.CreateSuccessResponse())
			}
			mt.AddMockResponses(responses...)

			err := repo.SeedSystemViews(context.Background(), "ws-1")
			require.NoError(t, err)

			var seenSystemKeys []string
			var sawDemotion bool
			var insertCount int

			for _, event := range mt.GetAllStartedEvents() {
				if event.CommandName == "insert" {
					insertCount++
					continue
				}
				if event.CommandName != "update" {
					continue
				}

				commandJSON := mustMarshalCommandJSON(t, event.Command)
				for _, wantSystemKey := range tc.wantSystemKeys {
					if strings.Contains(commandJSON, wantSystemKey) {
						seenSystemKeys = append(seenSystemKeys, wantSystemKey)
					}
				}
				if strings.Contains(commandJSON, `"type":"user"`) && strings.Contains(commandJSON, `"system_key":""`) {
					sawDemotion = true
					assert.Contains(t, commandJSON, `"$nin"`)
				}
			}

			assert.ElementsMatch(t, tc.wantSystemKeys, seenSystemKeys)
			assert.Equal(t, tc.wantDemotion, sawDemotion)
			assert.Equal(t, tc.wantInsertCount, insertCount)
		})
	}
}
