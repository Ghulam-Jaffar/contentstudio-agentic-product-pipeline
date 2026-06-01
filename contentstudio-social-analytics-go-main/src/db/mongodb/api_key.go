package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// ApiKeyRepository handles lookups against the api_keys collection.
type ApiKeyRepository struct {
	collection *mongo.Collection
	log        *logger.Logger
}

func NewApiKeyRepository(db *mongo.Database, log *logger.Logger) *ApiKeyRepository {
	return &ApiKeyRepository{
		collection: db.Collection("api_keys"),
		log:        log,
	}
}

// FindValidByKey returns the ApiKey document for the given key if it exists,
// is not revoked, and has not been soft-deleted. Returns nil (no error) when
// the key simply does not match — the caller treats that as unauthorised.
func (r *ApiKeyRepository) FindValidByKey(ctx context.Context, key string) (*mongoModels.ApiKey, error) {
	filter := bson.M{
		"key":        key,
		"revoked":    false,
		"deleted_at": bson.M{"$exists": false},
	}

	var apiKey mongoModels.ApiKey
	err := r.collection.FindOne(ctx, filter).Decode(&apiKey)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ApiKeyRepository.FindValidByKey: %w", err)
	}
	return &apiKey, nil
}

// FindActiveByUserID returns the most recently created, non-revoked, non-deleted
// API key for the given user ID string. Returns nil (no error) when none exists.
func (r *ApiKeyRepository) FindActiveByUserID(ctx context.Context, userID string) (*mongoModels.ApiKey, error) {
	filter := bson.M{
		"user_id":    userID,
		"revoked":    false,
		"deleted_at": bson.M{"$exists": false},
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var apiKey mongoModels.ApiKey
	err := r.collection.FindOne(ctx, filter, opts).Decode(&apiKey)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ApiKeyRepository.FindActiveByUserID: %w", err)
	}
	return &apiKey, nil
}
