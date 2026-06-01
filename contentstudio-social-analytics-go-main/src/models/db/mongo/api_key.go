package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ApiKey represents a document in the api_keys collection.
type ApiKey struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Key        string             `bson:"key"`
	UserID     interface{}        `bson:"user_id"`
	Revoked    bool               `bson:"revoked"`
	AICreation bool               `bson:"ai_creation"`
	LastUsedAt *time.Time         `bson:"last_used_at,omitempty"`
	DeletedAt  *time.Time         `bson:"deleted_at,omitempty"`
	CreatedAt  *time.Time         `bson:"created_at,omitempty"`
	UpdatedAt  *time.Time         `bson:"updated_at,omitempty"`
}
