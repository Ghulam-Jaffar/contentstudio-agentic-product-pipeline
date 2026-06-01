package main

import (
	"context"
	"sync"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"golang.org/x/sync/semaphore"
)

// LinkedInAccountWorkOrder mirrors the scheduler payload from Kafka.
// Contains all necessary information to process analytics for a LinkedIn account.
type LinkedInAccountWorkOrder = kafkamodels.LinkedinAccountWorkOrder

// LinkedInBatchWorkOrder represents a batch of accounts received from scheduler.
type LinkedInBatchWorkOrder = kafkamodels.LinkedinBatchWorkOrder

// WorkOrderMessage wraps a single account work order for internal processing.
// Distributed to workers after unpacking batch messages.
type WorkOrderMessage struct {
	AccountID  string // MongoDB _id for timestamp update callback
	LinkedinID string // LinkedIn social ID for logging
	Value      []byte // JSON payload of LinkedInAccountWorkOrder
	Ack        func() // called when fully processed; may be nil
}

// BatchMessage wraps a Kafka batch message for processing.
type BatchMessage struct {
	Key   []byte
	Value []byte
}

// enrichedPost holds post data with associated asset IDs for enrichment.
// Used during the post processing pipeline to collect and merge assets.
type enrichedPost struct {
	Post        map[string]any // Raw post data from LinkedIn API
	ActivityID  string         // Unique identifier for the post (ugcPost or share URN)
	ImageIDs    []string       // Image asset URNs to fetch
	VideoIDs    []string       // Video asset URNs to fetch
	DocumentIDs []string       // Document asset URNs to fetch (carousel/PDF)
}

// profileAnalyticsResults holds all fetched profile analytics data.
// Each field corresponds to a different LinkedIn Creator Analytics API response.
type profileAnalyticsResults struct {
	ImpressionData     []byte // Daily impression counts (IMPRESSION query type)
	MembersReachedData []byte // Total unique members reached (MEMBERS_REACHED query type)
	ReshareData        []byte // Daily reshare counts (RESHARE query type)
	ReactionData       []byte // Daily reaction counts (REACTION query type)
	CommentData        []byte // Daily comment counts (COMMENT query type)
	FollowerData       []byte // Daily follower count changes (q=dateRange)
	TotalFollowerData  []byte // Current total follower count (q=me)
}

// pageAnalyticsResults holds all fetched page/organization analytics data.
// Each field corresponds to a different LinkedIn Organization API response.
type pageAnalyticsResults struct {
	FollowerData []byte // Follower demographics and total count
	PageStats    []byte // Page view statistics
	ShareStats   []byte // Post engagement statistics
}

// assetFetchConfig holds configuration for a generic asset fetch operation.
// Used by the generic fetchAssetsInChunks function to reduce code duplication.
type assetFetchConfig struct {
	name       string                // Human-readable name for logging (e.g., "images", "videos")
	ids        []string              // Asset IDs to fetch
	chunkSize  int                   // Number of IDs per API call
	semaphore  *semaphore.Weighted   // Concurrency limiter
	fetchFunc  assetFetchFunc        // Function to call the API
	parseFunc  func([]byte) assetMap // Function to parse API response
	mu         *sync.Mutex           // Mutex for thread-safe map updates
	resultMap  assetMap              // Map to store results
	linkedinID string                // LinkedIn ID for logging
	log        *logger.Logger        // Logger instance
}

// assetFetchFunc is a function type for fetching assets from LinkedIn API.
// Returns raw JSON response bytes and any error.
type assetFetchFunc func(ctx context.Context, ids []string) ([]byte, error)

// assetMap is a type alias for the map structure used to store fetched assets.
// Key is the asset ID, value is the parsed asset data.
type assetMap = map[string]map[string]any

// analyticsQueryConfig holds configuration for a profile analytics API call.
// Used to reduce duplication when fetching multiple analytics query types.
type analyticsQueryConfig struct {
	queryType string // LinkedIn API query type (IMPRESSION, MEMBERS_REACHED, etc.)
	stageName string // Stage name for error capture
	logMsg    string // Log message on error
}

// linkedinClients holds all LinkedIn-related client instances.
// Passed to worker functions to avoid parameter bloat.
type linkedinClients struct {
	li         *social.LinkedInClient
	producer   kafka2.Producer
	decryptKey string
}

// ProcessingResult represents the result of processing a single account.
// Used to track success/failure for batch completion and timestamp updates.
type ProcessingResult struct {
	AccountID  string // MongoDB _id
	LinkedinID string // LinkedIn platform identifier
	Success    bool   // Whether processing succeeded
	Error      error  // Error if failed
}

// TimestampUpdateRequest represents a request to update last_analytics_updated_at.
// Sent to the timestamp update channel after successful processing.
type TimestampUpdateRequest struct {
	AccountID  string // MongoDB _id (hex)
	LinkedinID string // LinkedIn social ID for logging
}
