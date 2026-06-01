package mongodb

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// ListeningRepository handles MongoDB operations for listening topics.
type ListeningRepository struct {
	collection *mongo.Collection
	log        *logger.Logger
}

// NewListeningRepository initializes a new listening topic repository.
func NewListeningRepository(db *mongo.Database, log *logger.Logger) *ListeningRepository {
	return &ListeningRepository{
		collection: db.Collection("listening_topics"),
		log:        log,
	}
}

// GetTopicByID retrieves a listening topic by its MongoDB _id.
// It also normalises TopicID from the ObjectId when the topic_id field is absent
// (topics created by Laravel before the backfill migration).
func (r *ListeningRepository) GetTopicByID(ctx context.Context, topicID string) (*mongoModels.ListeningTopic, error) {
	objID, err := primitive.ObjectIDFromHex(topicID)
	if err != nil {
		return nil, fmt.Errorf("ListeningRepository.GetTopicByID: invalid topic ID: %w", err)
	}
	var topic mongoModels.ListeningTopic
	err = r.collection.FindOne(ctx, bson.M{"_id": objID, "deleted_at": nil}).Decode(&topic)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("ListeningRepository.GetTopicByID: failed to find topic: %w", err)
	}
	topic.Normalize()
	return &topic, nil
}

// GetMentionsCount returns the current usage.mentions_count for a topic.
func (r *ListeningRepository) GetMentionsCount(ctx context.Context, topicID string) (int, error) {
	topic, err := r.GetTopicByID(ctx, topicID)
	if err != nil {
		return 0, fmt.Errorf("ListeningRepository.GetMentionsCount: failed to get topic: %w", err)
	}

	if topic == nil {
		return 0, fmt.Errorf("ListeningRepository.GetMentionsCount: topic not found: %s", topicID)
	}

	return topic.Usage.MentionsCount, nil
}

// GetActiveTopics retrieves all topics. Status and quota checks are applied by
// higher-level scheduler logic.
func (r *ListeningRepository) GetActiveTopics(ctx context.Context) ([]*mongoModels.ListeningTopic, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"deleted_at": nil})
	if err != nil {
		return nil, fmt.Errorf("ListeningRepository.GetActiveTopics: failed to find active topics: %w", err)
	}
	defer cursor.Close(ctx)

	var topics []*mongoModels.ListeningTopic
	if err := cursor.All(ctx, &topics); err != nil {
		return nil, fmt.Errorf("ListeningRepository.GetActiveTopics: failed to decode topics: %w", err)
	}

	for _, t := range topics {
		t.Normalize()
	}

	r.log.Info().Int("count", len(topics)).Msg("Fetched active listening topics")
	return topics, nil
}

// CountActiveTopicsForQuota returns the number of active, non-deleted topics
// that share the given quota namespace. The fetcher uses this to scale the
// per-keyword max_posts cap to a topic's fair share of the shared mention
// budget — see fetcher.FetcherService.WithQuotaTopicCounter for context.
//
// quotaID may be either a super_admin_id (preferred — matches how the budget
// is keyed in Redis and used_mention_credits in MongoDB) or a workspace_id
// (legacy fallback for topics created before super_admin_id was denormalised
// onto the topic document). The $or filter covers both so the count matches
// however the fetcher resolved the quota for a given work order.
//
// Paused and limit-reached topics are excluded because they will not consume
// budget this cycle — including them in the denominator would unfairly
// shrink the active topics' fair-share cap.
func (r *ListeningRepository) CountActiveTopicsForQuota(ctx context.Context, quotaID string) (int, error) {
	if quotaID == "" {
		return 0, nil
	}
	filter := bson.M{
		"deleted_at": nil,
		"status":     bson.M{"$in": bson.A{"active", "", nil}},
		"$or": bson.A{
			bson.M{"super_admin_id": quotaID},
			bson.M{"workspace_id": quotaID},
		},
	}
	n, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("ListeningRepository.CountActiveTopicsForQuota: %w", err)
	}
	return int(n), nil
}

// ListTopicsByWorkspace retrieves all listening topics for a workspace.
func (r *ListeningRepository) ListTopicsByWorkspace(ctx context.Context, workspaceID string) ([]*mongoModels.ListeningTopic, error) {
	cursor, err := r.collection.Find(
		ctx,
		bson.M{"workspace_id": workspaceID, "deleted_at": nil},
		options.Find().SetSort(bson.D{{Key: "name", Value: 1}}),
	)
	if err != nil {
		return nil, fmt.Errorf("ListeningRepository.ListTopicsByWorkspace: failed to find workspace topics: %w", err)
	}
	defer cursor.Close(ctx)

	var topics []*mongoModels.ListeningTopic
	if err := cursor.All(ctx, &topics); err != nil {
		return nil, fmt.Errorf("ListeningRepository.ListTopicsByWorkspace: failed to decode topics: %w", err)
	}

	for _, topic := range topics {
		topic.Normalize()
	}

	return topics, nil
}

// IncrementMentionsCount atomically increments the usage.mentions_count for a topic.
func (r *ListeningRepository) IncrementMentionsCount(ctx context.Context, topicID string, count int) error {
	objID, err := primitive.ObjectIDFromHex(topicID)
	if err != nil {
		return fmt.Errorf("ListeningRepository.IncrementMentionsCount: invalid topic ID: %w", err)
	}
	update := bson.M{
		"$inc": bson.M{"usage.mentions_count": count},
		"$set": bson.M{"updated_at": time.Now()},
	}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("ListeningRepository.IncrementMentionsCount: failed to increment: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("ListeningRepository.IncrementMentionsCount: topic not found: %s", topicID)
	}
	return nil
}

// TryReserveMentionSlot atomically increments usage.mentions_count for a topic.
// It enforces a hard limit: if the count has reached mentions_limit, it returns false.
func (r *ListeningRepository) TryReserveMentionSlot(ctx context.Context, topicID string, limit int) (bool, int, error) {
	objID, err := primitive.ObjectIDFromHex(topicID)
	if err != nil {
		return false, 0, fmt.Errorf("ListeningRepository.TryReserveMentionSlot: invalid topic ID: %w", err)
	}

	filter := bson.M{
		"_id": objID,
		"$or": []bson.M{
			{"mentions_limit": bson.M{"$lte": 0}},
			{"$expr": bson.M{"$lt": bson.A{"$usage.mentions_count", "$mentions_limit"}}},
		},
	}
	update := bson.M{
		"$inc": bson.M{"usage.mentions_count": 1},
		"$set": bson.M{"updated_at": time.Now()},
	}

	var topic mongoModels.ListeningTopic
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&topic); err != nil {
		if err == mongo.ErrNoDocuments {
			// Check if we are over the limit without atomic increment
			err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&topic)
			if err == mongo.ErrNoDocuments {
				return false, 0, nil
			}
			if err != nil {
				return false, 0, fmt.Errorf("ListeningRepository.TryReserveMentionSlot: lookup: %w", err)
			}
			return false, topic.Usage.MentionsCount, nil
		}
		return false, 0, fmt.Errorf("ListeningTopicRepository.TryReserveMentionSlot: %w", err)
	}

	topic.Normalize()
	return true, topic.Usage.MentionsCount, nil
}

// ReleaseMentionSlot rolls back a previously reserved topic mention count.
func (r *ListeningRepository) ReleaseMentionSlot(ctx context.Context, topicID string, _ int) error {
	objID, err := primitive.ObjectIDFromHex(topicID)
	if err != nil {
		return fmt.Errorf("ListeningRepository.ReleaseMentionSlot: invalid topic ID: %w", err)
	}

	update := bson.M{
		"$inc": bson.M{"usage.mentions_count": -1},
		"$set": bson.M{"updated_at": time.Now()},
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{
		"_id":                  objID,
		"usage.mentions_count": bson.M{"$gt": 0},
	}, update)
	if err != nil {
		return fmt.Errorf("ListeningRepository.ReleaseMentionSlot: %w", err)
	}

	return nil
}

// SetMentionsLimitReached sets the mentions_limit_reached flag to true.
func (r *ListeningRepository) SetMentionsLimitReached(ctx context.Context, topicID string) error {
	objID, err := primitive.ObjectIDFromHex(topicID)
	if err != nil {
		return fmt.Errorf("ListeningRepository.SetMentionsLimitReached: invalid topic ID: %w", err)
	}
	update := bson.M{
		"$set": bson.M{
			"mentions_limit_reached": true,
			"updated_at":             time.Now(),
		},
	}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("ListeningRepository.SetMentionsLimitReached: failed to update: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("ListeningRepository.SetMentionsLimitReached: topic not found: %s", topicID)
	}
	return nil
}

// MarkFirstMentionsReceived atomically transitions a topic from "never received
// mentions" to "received mentions now". Returns true only when this caller won
// the race — the document had first_mentions_received_at == null and we set it.
// Subsequent calls return false. Idempotent and concurrency-safe by design;
// callers should treat the boolean as the authoritative "fire the first-batch
// notification" signal.
func (r *ListeningRepository) MarkFirstMentionsReceived(ctx context.Context, topicID string) (bool, error) {
	objID, err := primitive.ObjectIDFromHex(topicID)
	if err != nil {
		return false, fmt.Errorf("ListeningRepository.MarkFirstMentionsReceived: invalid topic ID: %w", err)
	}
	now := time.Now().UTC()
	filter := bson.M{
		"_id": objID,
		"$or": []bson.M{
			{"first_mentions_received_at": nil},
			{"first_mentions_received_at": bson.M{"$exists": false}},
		},
	}
	update := bson.M{
		"$set": bson.M{
			"first_mentions_received_at": now,
			"updated_at":                 now,
		},
	}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, fmt.Errorf("ListeningRepository.MarkFirstMentionsReceived: failed to update: %w", err)
	}
	return result.ModifiedCount == 1, nil
}

// MarkInitialSyncDone flips is_initial_sync_done to true once the fetcher
// completes the first crawl for a topic. Mirrors the idempotency guard the
// Laravel sync-callback used to perform: writes only when the incoming
// event_timestamp is strictly newer than the stored one. Returns true when
// the document was modified, false when the call was stale (a newer event
// already landed).
func (r *ListeningRepository) MarkInitialSyncDone(
	ctx context.Context,
	topicID, workspaceID string,
	eventAt time.Time,
) (bool, error) {
	objID, err := primitive.ObjectIDFromHex(topicID)
	if err != nil {
		return false, fmt.Errorf("ListeningRepository.MarkInitialSyncDone: invalid topic ID: %w", err)
	}
	now := time.Now().UTC()
	eventAt = eventAt.UTC()
	filter := bson.M{
		"_id":          objID,
		"workspace_id": workspaceID,
		"$or": []bson.M{
			{"last_event_timestamp": bson.M{"$exists": false}},
			{"last_event_timestamp": nil},
			{"last_event_timestamp": bson.M{"$lt": eventAt}},
		},
	}
	update := bson.M{
		"$set": bson.M{
			"is_initial_sync_done": true,
			"last_event_timestamp": eventAt,
			"updated_at":           now,
		},
	}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, fmt.Errorf("ListeningRepository.MarkInitialSyncDone: failed to update: %w", err)
	}
	return result.ModifiedCount == 1, nil
}

// UpdateLastFetched updates the last_fetched_at timestamp and per-platform cursors.
func (r *ListeningRepository) UpdateLastFetched(
	ctx context.Context,
	topicID string,
	fetchedAt time.Time,
	cursors map[string]string,
) error {
	objID, err := primitive.ObjectIDFromHex(topicID)
	if err != nil {
		return fmt.Errorf("ListeningRepository.UpdateLastFetched: invalid topic ID: %w", err)
	}
	fields := bson.M{
		"last_fetched_at": fetchedAt,
		"updated_at":      time.Now(),
	}
	if cursors != nil {
		fields["last_fetched_cursors"] = cursors
	}
	update := bson.M{"$set": fields}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("ListeningRepository.UpdateLastFetched: failed to update: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("ListeningRepository.UpdateLastFetched: topic not found: %s", topicID)
	}
	return nil
}

// ResetUsage resets the mentions count, limit flag, and period start for a topic.
func (r *ListeningRepository) ResetUsage(ctx context.Context, topicID string, periodStart time.Time) error {
	objID, err := primitive.ObjectIDFromHex(topicID)
	if err != nil {
		return fmt.Errorf("ListeningRepository.ResetUsage: invalid topic ID: %w", err)
	}
	update := bson.M{
		"$set": bson.M{
			"usage.mentions_count":   0,
			"mentions_limit_reached": false,
			"current_period_start":   periodStart,
			"updated_at":             time.Now(),
		},
	}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return fmt.Errorf("ListeningRepository.ResetUsage: failed to reset: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("ListeningRepository.ResetUsage: topic not found: %s", topicID)
	}
	return nil
}

type ListeningViewsRepository struct {
	collection *mongo.Collection
	log        *logger.Logger
}

func NewListeningViewsRepository(db *mongo.Database, log *logger.Logger) *ListeningViewsRepository {
	return &ListeningViewsRepository{
		collection: db.Collection("listening_views"),
		log:        log,
	}
}

func (r *ListeningViewsRepository) ListViews(ctx context.Context, workspaceID string) ([]mongoModels.ListeningView, error) {
	opts := options.Find().SetSort(bson.D{
		{Key: "type", Value: 1},
		{Key: "name", Value: 1},
	})

	cursor, err := r.collection.Find(ctx, bson.M{"workspace_id": workspaceID}, opts)
	if err != nil {
		return nil, fmt.Errorf("ListeningViewsRepository.ListViews: %w", err)
	}
	defer cursor.Close(ctx)

	var views []mongoModels.ListeningView
	if err := cursor.All(ctx, &views); err != nil {
		return nil, fmt.Errorf("ListeningViewsRepository.ListViews: decode: %w", err)
	}

	return views, nil
}

func (r *ListeningViewsRepository) GetViewByID(ctx context.Context, viewID string) (*mongoModels.ListeningView, error) {
	objID, err := primitive.ObjectIDFromHex(viewID)
	if err != nil {
		return nil, fmt.Errorf("ListeningViewsRepository.GetViewByID: invalid ID: %w", err)
	}

	var view mongoModels.ListeningView
	if err := r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&view); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("ListeningViewsRepository.GetViewByID: %w", err)
	}
	return &view, nil
}

func (r *ListeningViewsRepository) CreateView(ctx context.Context, view *mongoModels.ListeningView) (*mongoModels.ListeningView, error) {
	now := time.Now().UTC()
	view.CreatedAt = now
	view.UpdatedAt = now

	result, err := r.collection.InsertOne(ctx, view)
	if err != nil {
		return nil, fmt.Errorf("ListeningViewsRepository.CreateView: %w", err)
	}

	view.ID = result.InsertedID.(primitive.ObjectID)
	return view, nil
}

func (r *ListeningViewsRepository) UpdateView(ctx context.Context, viewID string, update bson.M) (*mongoModels.ListeningView, error) {
	objID, err := primitive.ObjectIDFromHex(viewID)
	if err != nil {
		return nil, fmt.Errorf("ListeningViewsRepository.UpdateView: invalid ID: %w", err)
	}

	update["updated_at"] = time.Now().UTC()
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var view mongoModels.ListeningView
	if err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": update},
		opts,
	).Decode(&view); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("ListeningViewsRepository.UpdateView: %w", err)
	}

	return &view, nil
}

func (r *ListeningViewsRepository) DeleteView(ctx context.Context, viewID string) error {
	objID, err := primitive.ObjectIDFromHex(viewID)
	if err != nil {
		return fmt.Errorf("ListeningViewsRepository.DeleteView: invalid ID: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("ListeningViewsRepository.DeleteView: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("ListeningViewsRepository.DeleteView: view not found: %s", viewID)
	}
	return nil
}

// presetSeedNames are the brand-preset views shipped at workspace creation.
// They are seeded once with type="preset" and behave like user views afterward
// (editable, deletable). Listed here so the migration step can recognise legacy
// workspaces that stored them as type="system" and flip them to "preset".
var presetSeedNames = []string{
	"Crisis Management",
	"Brand Monitoring",
	"Competitor Intel",
	"Buy Intent",
	"Brand Love",
}

const (
	listeningSystemKeyAllMentions   = "all_mentions"
	listeningSystemKeyHighRelevance = "high_relevance"
)

func (r *ListeningViewsRepository) SeedSystemViews(ctx context.Context, workspaceID string) error {
	// Migration: legacy workspaces seeded the brand presets as type="system",
	// which made them un-editable. Flip them to "preset" on the fly so existing
	// data picks up the new behaviour without a separate migration job. Idempotent.
	if _, err := r.collection.UpdateMany(ctx,
		bson.M{
			"workspace_id": workspaceID,
			"type":         "system",
			"name":         bson.M{"$in": presetSeedNames},
		},
		bson.M{"$set": bson.M{"type": "preset", "updated_at": time.Now().UTC()}},
	); err != nil {
		return fmt.Errorf("ListeningViewsRepository.SeedSystemViews: migrate presets: %w", err)
	}

	// Snapshot what already exists so system slots can be keyed stably even if
	// the user renames them. Name-based seeding created accidental duplicates.
	cursor, err := r.collection.Find(ctx, bson.M{
		"workspace_id": workspaceID,
	}, options.Find().SetProjection(bson.M{
		"name":                          1,
		"type":                          1,
		"system_key":                    1,
		"filter_preset.ai_tags":         1,
		"filter_preset.exclude_ai_tags": 1,
		"created_at":                    1,
	}))
	if err != nil {
		return fmt.Errorf("ListeningViewsRepository.SeedSystemViews: find: %w", err)
	}
	defer cursor.Close(ctx)

	existingNames := make(map[string]bool)
	type existingViewDoc struct {
		ID           primitive.ObjectID                    `bson:"_id"`
		Name         string                                `bson:"name"`
		Type         string                                `bson:"type"`
		SystemKey    string                                `bson:"system_key"`
		FilterPreset mongoModels.ListeningViewFilterPreset `bson:"filter_preset"`
		CreatedAt    time.Time                             `bson:"created_at"`
	}
	var systemDocs []existingViewDoc
	for cursor.Next(ctx) {
		var doc existingViewDoc
		// Fail loud on a decode error: silently skipping would let us treat an
		// existing view as absent and re-insert a duplicate default on the next
		// seed pass.
		if err := cursor.Decode(&doc); err != nil {
			return fmt.Errorf("ListeningViewsRepository.SeedSystemViews: decode: %w", err)
		}
		if doc.Name != "" {
			existingNames[doc.Name] = true
		}
		if doc.Type == "system" {
			systemDocs = append(systemDocs, doc)
		}
	}
	if err := cursor.Err(); err != nil {
		return fmt.Errorf("ListeningViewsRepository.SeedSystemViews: cursor: %w", err)
	}

	now := time.Now().UTC()

	var highRelevanceDoc *existingViewDoc
	var remainingSystemDocs []existingViewDoc
	for i := range systemDocs {
		doc := systemDocs[i]
		if doc.SystemKey == listeningSystemKeyHighRelevance {
			if highRelevanceDoc == nil || doc.CreatedAt.Before(highRelevanceDoc.CreatedAt) {
				highRelevanceDoc = &doc
			}
			continue
		}

		if doc.Name == "High Relevance" {
			if highRelevanceDoc == nil || doc.CreatedAt.Before(highRelevanceDoc.CreatedAt) {
				highRelevanceDoc = &doc
				continue
			}
		}

		if containsString(doc.FilterPreset.ExcludeAITags, "Irrelevant") {
			if highRelevanceDoc == nil || doc.CreatedAt.Before(highRelevanceDoc.CreatedAt) {
				highRelevanceDoc = &doc
				continue
			}
		}

		remainingSystemDocs = append(remainingSystemDocs, doc)
	}

	var allMentionsDoc *existingViewDoc
	for i := range remainingSystemDocs {
		doc := remainingSystemDocs[i]
		if allMentionsDoc == nil || doc.CreatedAt.Before(allMentionsDoc.CreatedAt) {
			allMentionsDoc = &doc
		}
	}

	var canonicalSystemIDs []primitive.ObjectID
	if allMentionsDoc != nil {
		canonicalSystemIDs = append(canonicalSystemIDs, allMentionsDoc.ID)
		if _, err := r.collection.UpdateOne(ctx,
			bson.M{"_id": allMentionsDoc.ID},
			bson.M{"$set": bson.M{
				"system_key": listeningSystemKeyAllMentions,
				"updated_at": now,
			}},
		); err != nil {
			return fmt.Errorf("ListeningViewsRepository.SeedSystemViews: set all_mentions key: %w", err)
		}
	}
	if highRelevanceDoc != nil {
		canonicalSystemIDs = append(canonicalSystemIDs, highRelevanceDoc.ID)
		if _, err := r.collection.UpdateOne(ctx,
			bson.M{"_id": highRelevanceDoc.ID},
			bson.M{"$set": bson.M{
				"system_key": listeningSystemKeyHighRelevance,
				"updated_at": now,
			}},
		); err != nil {
			return fmt.Errorf("ListeningViewsRepository.SeedSystemViews: set high_relevance key: %w", err)
		}
	}

	if len(systemDocs) > len(canonicalSystemIDs) {
		filter := bson.M{
			"workspace_id": workspaceID,
			"type":         "system",
		}
		if len(canonicalSystemIDs) > 0 {
			filter["_id"] = bson.M{"$nin": canonicalSystemIDs}
		}
		if _, err := r.collection.UpdateMany(ctx,
			filter,
			bson.M{"$set": bson.M{
				"type":       "user",
				"updated_at": now,
			}, "$unset": bson.M{
				"system_key": "",
			}},
		); err != nil {
			return fmt.Errorf("ListeningViewsRepository.SeedSystemViews: demote duplicate system views: %w", err)
		}
	}

	// System slots: always re-seed if missing. These two are protected — they
	// can be edited but never deleted, so re-inserting a missing one is the
	// only path that should fire (e.g. brand-new workspace).
	systemSeeds := []mongoModels.ListeningView{
		{
			WorkspaceID:  workspaceID,
			Name:         "All Mentions",
			Icon:         "Inbox",
			Type:         "system",
			SystemKey:    listeningSystemKeyAllMentions,
			FilterPreset: mongoModels.ListeningViewFilterPreset{},
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			WorkspaceID: workspaceID,
			Name:        "High Relevance",
			Icon:        "Star",
			Type:        "system",
			SystemKey:   listeningSystemKeyHighRelevance,
			FilterPreset: mongoModels.ListeningViewFilterPreset{
				ExcludeAITags: []string{"Irrelevant"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// Preset views: seed only on a fresh workspace (no views exist at all).
	// Once seeded, presets behave like user views — if the user deletes one
	// it must stay deleted, so we never re-add presets to a populated workspace.
	presetSeeds := []mongoModels.ListeningView{
		{
			WorkspaceID: workspaceID,
			Name:        "Crisis Management",
			Icon:        "AlertTriangle",
			Type:        "preset",
			FilterPreset: mongoModels.ListeningViewFilterPreset{
				Sentiments: []string{"negative"},
				AITags:     []string{"Own Brand Mention"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			WorkspaceID: workspaceID,
			Name:        "Brand Monitoring",
			Icon:        "Shield",
			Type:        "preset",
			FilterPreset: mongoModels.ListeningViewFilterPreset{
				AITags: []string{"Own Brand Mention"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			WorkspaceID: workspaceID,
			Name:        "Competitor Intel",
			Icon:        "Crosshair",
			Type:        "preset",
			FilterPreset: mongoModels.ListeningViewFilterPreset{
				AITags: []string{"Competitor Mention"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			WorkspaceID: workspaceID,
			Name:        "Buy Intent",
			Icon:        "ShoppingCart",
			Type:        "preset",
			FilterPreset: mongoModels.ListeningViewFilterPreset{
				AITags: []string{"Buy Intent"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			WorkspaceID: workspaceID,
			Name:        "Brand Love",
			Icon:        "Heart",
			Type:        "preset",
			FilterPreset: mongoModels.ListeningViewFilterPreset{
				Sentiments: []string{"positive"},
				AITags:     []string{"Own Brand Mention"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	var toInsert []interface{}
	existingSystemKeys := map[string]bool{}
	if allMentionsDoc != nil {
		existingSystemKeys[listeningSystemKeyAllMentions] = true
	}
	if highRelevanceDoc != nil {
		existingSystemKeys[listeningSystemKeyHighRelevance] = true
	}
	for _, v := range systemSeeds {
		if !existingSystemKeys[v.SystemKey] {
			toInsert = append(toInsert, v)
		}
	}
	// Fresh workspace = nothing exists yet (after migration). Only then add presets.
	if len(existingNames) == 0 {
		for _, v := range presetSeeds {
			toInsert = append(toInsert, v)
		}
	}

	if len(toInsert) == 0 {
		return nil
	}

	if _, err := r.collection.InsertMany(ctx, toInsert); err != nil {
		return fmt.Errorf("ListeningViewsRepository.SeedSystemViews: insert: %w", err)
	}

	r.log.Info().
		Str("workspace_id", workspaceID).
		Int("count", len(toInsert)).
		Msg("Seeded system listening views")
	return nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}

// ListeningWorkspaceRepository resolves listening quotas from workspace, user,
// and subscription plan documents. Owner-level mention usage is stored on
// workspace documents as used_mention_credits and mirrored across all
// workspaces belonging to the same super admin.
type ListeningWorkspaceRepository struct {
	workspaceCollection *mongo.Collection
	userCollection      *mongo.Collection
	planCollection      *mongo.Collection
	log                 *logger.Logger
}

// NewListeningWorkspaceRepository initialises the workspace quota repository.
func NewListeningWorkspaceRepository(db *mongo.Database, log *logger.Logger) *ListeningWorkspaceRepository {
	return &ListeningWorkspaceRepository{
		workspaceCollection: db.Collection("workspace"),
		userCollection:      db.Collection("users"),
		planCollection:      db.Collection("subscription_plans"),
		log:                 log,
	}
}

// IsWorkspaceMentionLimitReached returns true when the owner has consumed the
// plan-level listening mention quota.
func (r *ListeningWorkspaceRepository) IsWorkspaceMentionLimitReached(ctx context.Context, id string) (bool, error) {
	mentionsCount, mentionLimit, exists, err := r.GetWorkspaceUsage(ctx, id)
	if err != nil {
		return false, err
	}
	if !exists || mentionLimit <= 0 {
		return false, nil
	}
	return mentionsCount >= mentionLimit, nil
}

// IncrementWorkspaceMentionsCount atomically increments the canonical workspace
// counter, then mirrors the resulting total to the owner's sibling workspaces.
func (r *ListeningWorkspaceRepository) IncrementWorkspaceMentionsCount(
	ctx context.Context,
	id string,
	count int,
) (newCount int, mentionLimit int, err error) {
	ownerCtx, err := r.resolveOwnerQuotaContext(ctx, id)
	if err != nil {
		return 0, 0, fmt.Errorf("ListeningWorkspaceRepository.IncrementWorkspaceMentionsCount: %w", err)
	}
	if ownerCtx == nil {
		return 0, 0, nil
	}

	var updated workspaceQuotaDoc
	if err := r.workspaceCollection.FindOneAndUpdate(
		ctx,
		buildIdentifierFilter("_id", ownerCtx.CanonicalWorkspaceID),
		bson.M{"$inc": bson.M{"used_mention_credits": count}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated); err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("ListeningWorkspaceRepository.IncrementWorkspaceMentionsCount: %w", err)
	}

	newCount = updated.UsedMentionCredits
	if err := r.propagateUsedMentionCredits(ctx, ownerCtx.UserID, stringifyMongoID(updated.ID), newCount); err != nil {
		if r.log != nil {
			r.log.Warn().
				Err(err).
				Str("user_id", ownerCtx.UserID).
				Str("canonical_workspace_id", stringifyMongoID(updated.ID)).
				Int("used_mention_credits", newCount).
				Msg("Failed to mirror used_mention_credits to sibling workspaces after canonical increment")
		}
	}

	return newCount, ownerCtx.MentionLimit, nil
}

// TryReserveWorkspaceMention increments usage when quota remains. It is kept as
// a fail-closed helper for callers that want a boolean result.
func (r *ListeningWorkspaceRepository) TryReserveWorkspaceMention(
	ctx context.Context,
	id string,
) (bool, int, int, error) {
	mentionsCount, mentionLimit, exists, err := r.GetWorkspaceUsage(ctx, id)
	if err != nil {
		return false, 0, 0, err
	}
	if !exists || mentionLimit <= 0 {
		return false, 0, 0, nil
	}
	if mentionsCount >= mentionLimit {
		return false, mentionsCount, mentionLimit, nil
	}

	newCount, mentionLimit, err := r.IncrementWorkspaceMentionsCount(ctx, id, 1)
	if err != nil {
		return false, 0, 0, fmt.Errorf("ListeningWorkspaceRepository.TryReserveWorkspaceMention: %w", err)
	}

	return true, newCount, mentionLimit, nil
}

// ReleaseWorkspaceMentionReservation rolls back a previously reserved workspace mention count.
func (r *ListeningWorkspaceRepository) ReleaseWorkspaceMentionReservation(ctx context.Context, id string) error {
	ownerCtx, err := r.resolveOwnerQuotaContext(ctx, id)
	if err != nil {
		return fmt.Errorf("ListeningWorkspaceRepository.ReleaseWorkspaceMentionReservation: %w", err)
	}
	if ownerCtx == nil {
		return nil
	}

	var updated workspaceQuotaDoc
	if err := r.workspaceCollection.FindOneAndUpdate(ctx, bson.M{
		"$and": []bson.M{
			buildIdentifierFilter("_id", ownerCtx.CanonicalWorkspaceID),
			{"used_mention_credits": bson.M{"$gt": 0}},
		},
	}, bson.M{"$inc": bson.M{"used_mention_credits": -1}}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updated); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return fmt.Errorf("ListeningWorkspaceRepository.ReleaseWorkspaceMentionReservation: %w", err)
	}

	if err := r.propagateUsedMentionCredits(ctx, ownerCtx.UserID, stringifyMongoID(updated.ID), updated.UsedMentionCredits); err != nil {
		return fmt.Errorf("ListeningWorkspaceRepository.ReleaseWorkspaceMentionReservation: propagate: %w", err)
	}

	return nil
}

// SetWorkspaceMentionLimitReached is kept for interface compatibility. Limit
// state is derived from used_mention_credits and plan limits, so nothing is persisted here.
func (r *ListeningWorkspaceRepository) SetWorkspaceMentionLimitReached(ctx context.Context, id string) error {
	r.log.Debug().Str("id", id).Msg("Workspace mention limit reached")
	return nil
}

// GetTopicLimit returns the mentions_limit for a topic.
// Reuses GetTopicByID to read the full document and extract the limit.
func (r *ListeningRepository) GetTopicLimit(ctx context.Context, topicID string) (int, error) {
	topic, err := r.GetTopicByID(ctx, topicID)
	if err != nil {
		return 0, fmt.Errorf("ListeningRepository.GetTopicLimit: %w", err)
	}
	if topic == nil {
		return 0, fmt.Errorf("ListeningRepository.GetTopicLimit: topic not found: %s", topicID)
	}
	return topic.MentionsLimit, nil
}

// GetWorkspaceUsage returns the current owner-level mention count, mention
// limit, and whether listening is enabled for the owner. The explicit Laravel
// billing flag has_social_listening_subscription must be present and true.
func (r *ListeningWorkspaceRepository) GetWorkspaceUsage(
	ctx context.Context,
	id string,
) (mentionsCount, mentionLimit int, exists bool, err error) {
	ownerCtx, err := r.resolveOwnerQuotaContext(ctx, id)
	if err != nil {
		return 0, 0, false, fmt.Errorf("ListeningWorkspaceRepository.GetWorkspaceUsage: %w", err)
	}
	if ownerCtx == nil || !ownerCtx.HasSocialListeningSubscription || ownerCtx.MentionLimit <= 0 {
		return 0, 0, false, nil
	}
	return ownerCtx.UsedMentionCredits, ownerCtx.MentionLimit, true, nil
}

// GetSuperAdminID resolves the super admin from the workspace owner.
func (r *ListeningWorkspaceRepository) GetSuperAdminID(ctx context.Context, workspaceID string) (string, error) {
	workspace, err := r.findWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return "", fmt.Errorf("ListeningWorkspaceRepository.GetSuperAdminID: %w", err)
	}
	if workspace == nil {
		return "", nil
	}
	return stringifyMongoID(workspace.UserID), nil
}

type workspaceQuotaDoc struct {
	ID                 interface{} `bson:"_id"`
	UserID             interface{} `bson:"user_id"`
	UsedMentionCredits int         `bson:"used_mention_credits"`
}

type workspaceUserQuotaDoc struct {
	ID                             interface{}            `bson:"_id"`
	SubscriptionID                 interface{}            `bson:"subscription_id"`
	Stackable                      interface{}            `bson:"stackable"`
	Addons                         map[string]interface{} `bson:"addons"`
	HasSocialListeningSubscription bool                   `bson:"has_social_listening_subscription"`
}

type subscriptionPlanQuotaDoc struct {
	Limits map[string]interface{} `bson:"limits"`
}

type ownerQuotaContext struct {
	UserID                         string
	CanonicalWorkspaceID           string
	UsedMentionCredits             int
	MentionLimit                   int
	HasSocialListeningSubscription bool
}

func (r *ListeningWorkspaceRepository) resolveOwnerQuotaContext(ctx context.Context, id string) (*ownerQuotaContext, error) {
	workspace, err := r.findWorkspaceByID(ctx, id)
	if err != nil {
		return nil, err
	}

	userID := id
	if workspace != nil {
		userID = stringifyMongoID(workspace.UserID)
	}
	if userID == "" {
		return nil, nil
	}

	workspaces, err := r.findOwnerWorkspaces(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(workspaces) == 0 {
		return nil, nil
	}

	mentionLimit, hasSocialListeningSubscription, err := r.resolveMentionLimit(ctx, userID)
	if err != nil {
		return nil, err
	}

	usedMentionCredits := 0
	for _, ownerWorkspace := range workspaces {
		if ownerWorkspace.UsedMentionCredits > usedMentionCredits {
			usedMentionCredits = ownerWorkspace.UsedMentionCredits
		}
	}

	return &ownerQuotaContext{
		UserID:                         userID,
		CanonicalWorkspaceID:           stringifyMongoID(workspaces[0].ID),
		UsedMentionCredits:             usedMentionCredits,
		MentionLimit:                   mentionLimit,
		HasSocialListeningSubscription: hasSocialListeningSubscription,
	}, nil
}

func (r *ListeningWorkspaceRepository) findWorkspaceByID(ctx context.Context, workspaceID string) (*workspaceQuotaDoc, error) {
	var workspace workspaceQuotaDoc
	err := r.workspaceCollection.FindOne(ctx, buildIdentifierFilter("_id", workspaceID)).Decode(&workspace)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &workspace, nil
}

func (r *ListeningWorkspaceRepository) findOwnerWorkspaces(ctx context.Context, userID string) ([]workspaceQuotaDoc, error) {
	cursor, err := r.workspaceCollection.Find(
		ctx,
		bson.M{"$or": buildIdentifierClauses("user_id", userID)},
		options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var workspaces []workspaceQuotaDoc
	if err := cursor.All(ctx, &workspaces); err != nil {
		return nil, err
	}

	return workspaces, nil
}

func (r *ListeningWorkspaceRepository) resolveMentionLimit(ctx context.Context, userID string) (int, bool, error) {
	var user workspaceUserQuotaDoc
	err := r.userCollection.FindOne(ctx, buildIdentifierFilter("_id", userID)).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, false, nil
		}
		return 0, false, err
	}

	planLimit := 0
	if subscriptionID := stringifyMongoID(user.SubscriptionID); subscriptionID != "" {
		var plan subscriptionPlanQuotaDoc
		err := r.planCollection.FindOne(ctx, buildIdentifierFilter("_id", subscriptionID)).Decode(&plan)
		if err != nil && err != mongo.ErrNoDocuments {
			return 0, false, err
		}
		planLimit = toInt(plan.Limits["listening_mentions"])
	}

	if stackable := toInt(user.Stackable); stackable > 1 {
		planLimit *= stackable
	}

	return planLimit + toInt(user.Addons["listening_mentions"]), user.HasSocialListeningSubscription, nil
}

func (r *ListeningWorkspaceRepository) propagateUsedMentionCredits(ctx context.Context, userID, canonicalWorkspaceID string, usedMentionCredits int) error {
	_, err := r.workspaceCollection.UpdateMany(ctx, bson.M{
		"$and": []bson.M{
			{"$or": buildIdentifierClauses("user_id", userID)},
			{"$nor": buildIdentifierClauses("_id", canonicalWorkspaceID)},
		},
	}, bson.M{
		"$set": bson.M{"used_mention_credits": usedMentionCredits},
	})
	return err
}

func buildIdentifierFilter(field, identifier string) bson.M {
	return bson.M{"$or": buildIdentifierClauses(field, identifier)}
}

func buildIdentifierClauses(field, identifier string) []bson.M {
	clauses := []bson.M{{field: identifier}}
	if objectID, err := primitive.ObjectIDFromHex(identifier); err == nil {
		clauses = append(clauses, bson.M{field: objectID})
	}
	return clauses
}

func stringifyMongoID(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case primitive.ObjectID:
		return typed.Hex()
	default:
		return fmt.Sprint(typed)
	}
}

// toInt coerces values pulled from heterogeneous Mongo documents into ints.
// Laravel/PHP and Stripe webhook payloads write integer-shaped fields as
// strings ("5000"), and aggregation pipelines occasionally surface them as
// Decimal128, so a strict int-only switch silently zeroed out plan limits and
// addons after the first non-numeric write. Treat anything number-like as a
// number; everything else falls back to 0.
func toInt(value interface{}) int {
	switch typed := value.(type) {
	case nil:
		return 0
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	case bool:
		if typed {
			return 1
		}
		return 0
	case string:
		s := strings.TrimSpace(typed)
		if s == "" {
			return 0
		}
		if n, err := strconv.Atoi(s); err == nil {
			return n
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return int(f)
		}
		return 0
	case primitive.Decimal128:
		if f, err := strconv.ParseFloat(typed.String(), 64); err == nil {
			return int(f)
		}
		return 0
	default:
		return 0
	}
}

// GetAIContext retrieves the typed AI context, free-text hint, and topic
// fields needed by the enrichment service in a single Mongo round trip.
// Reads both the top-level (Go-written) and nested query.* (Laravel-written)
// shapes; whichever is non-empty wins.
func (r *ListeningRepository) GetAIContext(ctx context.Context, topicID string) (mongoModels.TopicContextSnapshot, error) {
	objID, err := primitive.ObjectIDFromHex(topicID)
	if err != nil {
		return mongoModels.TopicContextSnapshot{}, fmt.Errorf("ListeningRepository.GetAIContext: invalid topic ID %q: %w", topicID, err)
	}

	var doc struct {
		Name            string                `bson:"name"`
		Type            string                `bson:"type"`
		AIContext       mongoModels.AIContext `bson:"ai_context"`
		AIContextHint   string                `bson:"ai_context_hint"`
		IncludeKeywords []string              `bson:"include_keywords"`
		Query           struct {
			AIContext       mongoModels.AIContext `bson:"ai_context"`
			AIContextHint   string                `bson:"ai_context_hint"`
			IncludeKeywords []string              `bson:"include_keywords"`
		} `bson:"query"`
	}

	err = r.collection.FindOne(
		ctx,
		bson.M{"_id": objID},
		options.FindOne().SetProjection(bson.M{
			"name":                   1,
			"type":                   1,
			"ai_context":             1,
			"ai_context_hint":        1,
			"include_keywords":       1,
			"query.ai_context":       1,
			"query.ai_context_hint":  1,
			"query.include_keywords": 1,
		}),
	).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return mongoModels.TopicContextSnapshot{}, nil
		}
		return mongoModels.TopicContextSnapshot{}, fmt.Errorf("ListeningRepository.GetAIContext: %w", err)
	}

	aiCtx := doc.AIContext
	if aiCtx.IsEmpty() {
		aiCtx = doc.Query.AIContext
	}
	hint := doc.AIContextHint
	if hint == "" {
		hint = doc.Query.AIContextHint
	}
	keywords := doc.IncludeKeywords
	if len(keywords) == 0 {
		keywords = doc.Query.IncludeKeywords
	}

	return mongoModels.TopicContextSnapshot{
		AIContext:     aiCtx,
		Hint:          hint,
		TopicName:     doc.Name,
		TopicType:     doc.Type,
		TopicKeywords: keywords,
	}, nil
}

// GetTopicStatus returns the status and existence of a topic.
// Returns ("", false, nil) when the topic doesn't exist (deleted).
func (r *ListeningRepository) GetTopicStatus(ctx context.Context, topicID string) (status string, exists bool, err error) {
	topic, err := r.GetTopicByID(ctx, topicID)
	if err != nil {
		return "", false, fmt.Errorf("ListeningRepository.GetTopicStatus: %w", err)
	}
	if topic == nil {
		return "", false, nil
	}
	return topic.Status, true, nil
}
