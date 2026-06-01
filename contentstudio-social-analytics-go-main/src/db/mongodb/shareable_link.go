package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// ShareableLinkRepository validates analytics shared links from MongoDB.
type ShareableLinkRepository struct {
	shareLinksCollection *mongo.Collection
	usersCollection      *mongo.Collection
	log                  *logger.Logger
}

func NewShareableLinkRepository(db *mongo.Database, log *logger.Logger) *ShareableLinkRepository {
	return &ShareableLinkRepository{
		shareLinksCollection: db.Collection("analytics_share_links"),
		usersCollection:      db.Collection("users"),
		log:                  log,
	}
}

// FindActiveUserIDByLinkID returns the associated user ID for a valid, enabled link.
// Returns empty string (no error) when link is missing/disabled or user no longer exists.
func (r *ShareableLinkRepository) FindActiveUserIDByLinkID(ctx context.Context, linkID string) (string, error) {
	filter := bson.M{
		"link_id":     linkID,
		"is_disabled": bson.M{"$ne": true},
	}

	var link mongoModels.AnalyticsShareLink
	err := r.shareLinksCollection.FindOne(
		ctx,
		filter,
		options.FindOne().SetProjection(bson.M{"user_id": 1, "link_id": 1}),
	).Decode(&link)
	if err == mongo.ErrNoDocuments {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("ShareableLinkRepository.FindActiveUserIDByLinkID: %w", err)
	}

	userID := normalizeUserID(link.UserID)
	if userID == "" {
		return "", nil
	}

	userFilter := bson.M{"_id": userID}
	if oid, err := primitive.ObjectIDFromHex(userID); err == nil {
		userFilter = bson.M{"_id": bson.M{"$in": bson.A{oid, userID}}}
	}

	var userDoc bson.M
	err = r.usersCollection.FindOne(
		ctx,
		userFilter,
		options.FindOne().SetProjection(bson.M{"_id": 1}),
	).Decode(&userDoc)
	if err == mongo.ErrNoDocuments {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("ShareableLinkRepository.FindActiveUserIDByLinkID (user check): %w", err)
	}

	return userID, nil
}

func normalizeUserID(id interface{}) string {
	switch v := id.(type) {
	case primitive.ObjectID:
		return v.Hex()
	case string:
		return v
	default:
		return ""
	}
}
