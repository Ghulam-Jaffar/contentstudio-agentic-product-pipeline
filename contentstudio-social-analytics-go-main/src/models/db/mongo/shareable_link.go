package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AnalyticsShareLink represents a document in analytics_share_links collection.
type AnalyticsShareLink struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	LinkID     string             `bson:"link_id"`
	UserID     interface{}        `bson:"user_id"`
	IsDisabled bool               `bson:"is_disabled"`
	CreatedAt  *time.Time         `bson:"created_at,omitempty"`
	UpdatedAt  *time.Time         `bson:"updated_at,omitempty"`
}
