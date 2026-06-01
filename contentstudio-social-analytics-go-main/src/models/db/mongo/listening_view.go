package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ListeningViewFilterPreset struct {
	TopicIDs           []string `bson:"topic_ids"             json:"topic_ids"`
	Platforms          []string `bson:"platforms"            json:"platforms"`
	Sentiments         []string `bson:"sentiments"           json:"sentiments"`
	AITags             []string `bson:"ai_tags"              json:"ai_tags"`
	ExcludeAITags      []string `bson:"exclude_ai_tags"      json:"exclude_ai_tags"`
	MinFollowers       int      `bson:"min_followers"        json:"min_followers"`
	MinTotalEngagement int      `bson:"min_total_engagement" json:"min_total_engagement"`
	Language           []string `bson:"language"             json:"language"`
}

type ListeningView struct {
	ID           primitive.ObjectID        `bson:"_id,omitempty"   json:"id"`
	WorkspaceID  string                    `bson:"workspace_id"    json:"workspace_id"`
	Name         string                    `bson:"name"            json:"name"`
	Icon         string                    `bson:"icon"            json:"icon"`
	Type         string                    `bson:"type"            json:"type"`
	SystemKey    string                    `bson:"system_key,omitempty" json:"system_key,omitempty"`
	FilterPreset ListeningViewFilterPreset `bson:"filter_preset"   json:"filter_preset"`
	CreatedAt    time.Time                 `bson:"created_at"      json:"created_at"`
	UpdatedAt    time.Time                 `bson:"updated_at"      json:"updated_at"`
}
