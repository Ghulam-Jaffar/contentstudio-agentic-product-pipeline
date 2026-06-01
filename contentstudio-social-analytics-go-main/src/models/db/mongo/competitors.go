package mongo

import (
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CompetitorReport represents a competitor report in MongoDB
type CompetitorReport struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	WorkspaceID     primitive.ObjectID `bson:"workspace_id" json:"workspace_id"`
	Name            string             `bson:"name" json:"name"`
	PlatformType    string             `bson:"platform_type,omitempty" json:"platform_type,omitempty"`
	Competitors     []string           `bson:"competitors" json:"competitors"`
	CreatedByUserID primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id"`
	UpdatedByUserID primitive.ObjectID `bson:"updated_by_user_id,omitempty" json:"updated_by_user_id,omitempty"`
	UpdatedAt       time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

// Competitor represents a competitor document in MongoDB
type Competitor struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty"`
	CompetitorID           interface{}        `bson:"competitor_id"` // Can be int64 or string in MongoDB
	Name                   string             `bson:"name"`
	Slug                   string             `bson:"slug"`
	State                  string             `bson:"state"`
	Image                  string             `bson:"image,omitempty"`
	Error                  string             `bson:"error,omitempty"`
	IsActive               bool               `bson:"is_active"`
	PlatformType           string             `bson:"platform_type"`
	LastAnalyticsUpdatedAt time.Time          `bson:"last_analytics_updated_at,omitempty"`
}

// GetCompetitorIDAsString returns the competitor ID as a string, converting from int64 if necessary
func (c *Competitor) GetCompetitorIDAsString() string {
	switch v := c.CompetitorID.(type) {
	case string:
		return v
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatInt(int64(v), 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ParseObjectID parses a hex string into a primitive.ObjectID.
func ParseObjectID(hex string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(hex)
}

// User represents a user document in MongoDB
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Email     string             `bson:"email"`
	FirstName string             `bson:"first_name"`
	LastName  string             `bson:"last_name"`
	APIKey    string             `bson:"api_key"`
}
