package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FlexStringMap decodes both BSON Document (map) and BSON Array (PHP empty [])
// into a map[string]string. PHP serialises an empty associative array as a
// BSON Array, which the standard map decoder rejects.
type FlexStringMap map[string]string

func (f *FlexStringMap) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t == bsontype.Array {
		*f = make(FlexStringMap)
		return nil
	}
	var m map[string]string
	if err := bson.Unmarshal(data, &m); err != nil {
		return err
	}
	if m == nil {
		*f = make(FlexStringMap)
	} else {
		*f = FlexStringMap(m)
	}
	return nil
}

// FlexPlatforms decodes either a BSON Array ([]string, Go-written documents)
// or a BSON Document (map[string]bool, Laravel's format where keys are platform
// names and values indicate whether the platform is enabled) into a string slice
// of enabled platform names.
type FlexPlatforms []string

func (f *FlexPlatforms) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t == bsontype.Array {
		var s []string
		if err := bson.Unmarshal(data, &s); err != nil {
			return err
		}
		*f = FlexPlatforms(s)
		return nil
	}
	if t == bsontype.EmbeddedDocument {
		var m map[string]bool
		if err := bson.Unmarshal(data, &m); err != nil {
			return err
		}
		result := make(FlexPlatforms, 0, len(m))
		for platform, enabled := range m {
			if enabled {
				result = append(result, platform)
			}
		}
		*f = result
		return nil
	}
	*f = nil
	return nil
}

// ListeningTopicUsage tracks mention usage counters for a listening topic.
type ListeningTopicUsage struct {
	MentionsCount int `bson:"mentions_count" json:"mentions_count"`
	MentionsLimit int `bson:"mentions_limit" json:"mentions_limit"`
}

// ListeningWorkspaceUsage tracks workspace-level subscription limits for social listening.
// Stored in the "listening_workspace_usage" collection with workspace_id as the unique key.
// topic_limit and mention_limit are set by the billing/subscription system when the
// social_listening addon is purchased; this service only reads limits and increments counts.
type ListeningWorkspaceUsage struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty"          json:"id"`
	WorkspaceID         string             `bson:"workspace_id"           json:"workspace_id"`
	SuperAdminID        string             `bson:"super_admin_id"         json:"super_admin_id"`
	MentionLimit        int                `bson:"mention_limit_monthly"  json:"mention_limit"`
	TopicLimit          int                `bson:"topic_limit"            json:"topic_limit"`
	MentionsCount       int                `bson:"mentions_this_month"    json:"mentions_count"`
	MentionLimitReached bool               `bson:"mention_limit_reached"  json:"mention_limit_reached"`
	UpdatedAt           time.Time          `bson:"updated_at"             json:"updated_at"`
}

// laravelQuery mirrors the nested `query` field written by the Laravel backend.
type laravelQuery struct {
	IncludeKeywords          []string  `bson:"include_keywords"`
	ExcludeKeywords          []string  `bson:"exclude_keywords"`
	IncludeAny               []string  `bson:"include_any"`
	IncludeAll               []string  `bson:"include_all"`
	ExactMatch               bool      `bson:"exact_match"`
	CaseSensitive            bool      `bson:"case_sensitive"`
	IncludeAuthors           []string  `bson:"include_authors"`
	ExcludeAuthors           []string  `bson:"exclude_authors"`
	Languages                []string  `bson:"language"`
	Regions                  []string  `bson:"regions"`
	GlobalExcludedSubreddits []string  `bson:"global_excluded_subreddits"`
	AIContext                AIContext `bson:"ai_context"`
	AIContextHint            string    `bson:"ai_context_hint"`
}

// ListeningTopic represents a social listening topic stored in MongoDB.
// The Go pipeline writes flat fields (include_keywords, enabled_platforms, …).
// The Laravel backend writes a nested query object and uses "platforms" instead
// of "enabled_platforms". Normalize() reconciles both layouts after decode.
type ListeningTopic struct {
	ID                       primitive.ObjectID  `bson:"_id,omitempty"          json:"id"`
	TopicID                  string              `bson:"topic_id"               json:"topic_id"`
	WorkspaceID              string              `bson:"workspace_id"           json:"workspace_id"`
	SuperAdminID             string              `bson:"super_admin_id"         json:"super_admin_id"`
	Name                     string              `bson:"name"                   json:"name"`
	Status                   string              `bson:"status"                 json:"status"`
	IncludeKeywords          []string            `bson:"include_keywords"       json:"include_keywords"`
	ExcludeKeywords          []string            `bson:"exclude_keywords"       json:"exclude_keywords"`
	IncludeAny               []string            `bson:"include_any"            json:"include_any"`
	IncludeAll               []string            `bson:"include_all"            json:"include_all"`
	ExactMatch               bool                `bson:"exact_match"            json:"exact_match"`
	CaseSensitive            bool                `bson:"case_sensitive"         json:"case_sensitive"`
	IncludeAuthors           []string            `bson:"include_authors"        json:"include_authors"`
	ExcludeAuthors           []string            `bson:"exclude_authors"        json:"exclude_authors"`
	Languages                []string            `bson:"language"               json:"language"`
	Regions                  []string            `bson:"regions"                json:"regions"`
	CustomTypeID             string              `bson:"custom_type_id"         json:"custom_type_id"`
	GlobalExcludedSubreddits []string            `bson:"global_excluded_subreddits" json:"global_excluded_subreddits"`
	EnabledPlatforms         []string            `bson:"enabled_platforms"      json:"enabled_platforms"`
	IsInitialSyncDone        bool                `bson:"is_initial_sync_done"   json:"is_initial_sync_done"`
	MentionsLimit            int                 `bson:"mentions_limit"         json:"mentions_limit"`
	MentionsLimitReached     bool                `bson:"mentions_limit_reached" json:"mentions_limit_reached"`
	Usage                    ListeningTopicUsage `bson:"usage"                  json:"usage"`
	CurrentPeriodStart       time.Time           `bson:"current_period_start"   json:"current_period_start"`
	LastFetchedAt            time.Time           `bson:"last_fetched_at"        json:"last_fetched_at"`
	LastFetchedCursors       FlexStringMap       `bson:"last_fetched_cursors"   json:"last_fetched_cursors"`
	// Set the first time a non-empty mention batch lands in ClickHouse for
	// this topic. Drives the feed-page progress animation. Pointer/null
	// distinguishes "never received" from a zero-value timestamp.
	FirstMentionsReceivedAt *time.Time `bson:"first_mentions_received_at,omitempty" json:"first_mentions_received_at,omitempty"`
	CreatedAt               time.Time  `bson:"created_at"             json:"created_at"`
	UpdatedAt               time.Time  `bson:"updated_at"             json:"updated_at"`
	AIContextHint           string     `bson:"ai_context_hint"        json:"ai_context_hint"`
	AIContext               AIContext  `bson:"ai_context"             json:"ai_context"`

	// Laravel-written fields — not used directly; consumed by Normalize().
	Platforms FlexPlatforms `bson:"platforms" json:"-"`
	Query     laravelQuery  `bson:"query"     json:"-"`
}

// Normalize reconciles the two document layouts (Go-written flat fields vs
// Laravel-written nested query + platforms). Call after any MongoDB decode.
func (t *ListeningTopic) Normalize() {
	if t.TopicID == "" {
		t.TopicID = t.ID.Hex()
	}
	if t.MentionsLimit == 0 {
		t.MentionsLimit = t.Usage.MentionsLimit
	}
	if len(t.EnabledPlatforms) == 0 {
		t.EnabledPlatforms = t.Platforms
	}
	if len(t.IncludeKeywords) == 0 {
		t.IncludeKeywords = t.Query.IncludeKeywords
	}
	if len(t.ExcludeKeywords) == 0 {
		t.ExcludeKeywords = t.Query.ExcludeKeywords
	}
	if len(t.IncludeAny) == 0 {
		t.IncludeAny = t.Query.IncludeAny
	}
	if len(t.IncludeAll) == 0 {
		t.IncludeAll = t.Query.IncludeAll
	}
	if len(t.IncludeAuthors) == 0 {
		t.IncludeAuthors = t.Query.IncludeAuthors
	}
	if len(t.ExcludeAuthors) == 0 {
		t.ExcludeAuthors = t.Query.ExcludeAuthors
	}
	if len(t.Languages) == 0 {
		t.Languages = t.Query.Languages
	}
	if len(t.Regions) == 0 {
		t.Regions = t.Query.Regions
	}
	if !t.ExactMatch {
		t.ExactMatch = t.Query.ExactMatch
	}
	if !t.CaseSensitive {
		t.CaseSensitive = t.Query.CaseSensitive
	}
	if len(t.GlobalExcludedSubreddits) == 0 {
		t.GlobalExcludedSubreddits = t.Query.GlobalExcludedSubreddits
	}
	if t.AIContext.IsEmpty() {
		t.AIContext = t.Query.AIContext
	}
	if t.AIContextHint == "" {
		t.AIContextHint = t.Query.AIContextHint
	}
	if !t.MentionsLimitReached && t.MentionsLimit > 0 && t.Usage.MentionsCount >= t.MentionsLimit {
		t.MentionsLimitReached = true
	}
}
