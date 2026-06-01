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

// UserRepository handles lookups against the api_keys collection on behalf of a user.
type UserRepository struct {
	collection *mongo.Collection
	log        *logger.Logger
}

func NewUserRepository(db *mongo.Database, log *logger.Logger) *UserRepository {
	return &UserRepository{
		collection: db.Collection("api_keys"),
		log:        log,
	}
}

// FindAPIKeyByUserID returns the active API key for the given user ID.
// Queries the api_keys collection (same as Laravel) for a non-revoked, non-deleted key.
// Returns an empty string (no error) when no key exists.
func (r *UserRepository) FindAPIKeyByUserID(ctx context.Context, userID string) (string, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", fmt.Errorf("UserRepository.FindAPIKeyByUserID: invalid user ID %q: %w", userID, err)
	}

	var apiKey mongoModels.ApiKey
	err = r.collection.FindOne(
		ctx,
		bson.M{
			"user_id":    bson.M{"$in": bson.A{oid, userID}},
			"revoked":    false,
			"deleted_at": nil, // nil matches both null and missing field (Laravel soft deletes store null)
		},
		options.FindOne().SetProjection(bson.M{"key": 1}),
	).Decode(&apiKey)

	if err == mongo.ErrNoDocuments {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("UserRepository.FindAPIKeyByUserID: %w", err)
	}
	return apiKey.Key, nil
}
