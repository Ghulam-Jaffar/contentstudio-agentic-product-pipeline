package main

import (
	"sync/atomic"
)

// ServiceConfig holds configuration for the GMB analytics sink service
type ServiceConfig struct {
	DataParserWorkers      int
	BatchProcessorsPerType int
	MaxBatchSize           int
	BatchTimeoutSeconds    int
	IdleTimeoutMinutes     int
	MessageChanSize        int
}

// DefaultServiceConfig returns the default service configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		DataParserWorkers:      5,
		BatchProcessorsPerType: batchProcessorsPerType,
		MaxBatchSize:           maxBatchSize,
		BatchTimeoutSeconds:    int(batchTimeout.Seconds()),
		IdleTimeoutMinutes:     int(idleTimeout.Minutes()),
		MessageChanSize:        messageChanSize,
	}
}

// ServiceMetrics holds runtime metrics for the GMB analytics sink
type ServiceMetrics struct {
	PickedCount            uint64
	ParsedDailyMetrics     uint64
	ParsedMediaAssets      uint64
	ParsedSearchKeywords   uint64
	ParsedLocalPosts       uint64
	ParsedReviews          uint64
	InsertedDailyMetrics   uint64
	InsertedMediaAssets    uint64
	InsertedSearchKeywords uint64
	InsertedLocalPosts     uint64
	InsertedReviews        uint64
}

// GetMetrics returns current service metrics as a map
func GetMetrics(m *ServiceMetrics) map[string]uint64 {
	return map[string]uint64{
		"picked":                   atomic.LoadUint64(&m.PickedCount),
		"parsed_daily_metrics":     atomic.LoadUint64(&m.ParsedDailyMetrics),
		"parsed_media_assets":      atomic.LoadUint64(&m.ParsedMediaAssets),
		"parsed_search_keywords":   atomic.LoadUint64(&m.ParsedSearchKeywords),
		"parsed_local_posts":       atomic.LoadUint64(&m.ParsedLocalPosts),
		"parsed_reviews":           atomic.LoadUint64(&m.ParsedReviews),
		"inserted_daily_metrics":   atomic.LoadUint64(&m.InsertedDailyMetrics),
		"inserted_media_assets":    atomic.LoadUint64(&m.InsertedMediaAssets),
		"inserted_search_keywords": atomic.LoadUint64(&m.InsertedSearchKeywords),
		"inserted_local_posts":     atomic.LoadUint64(&m.InsertedLocalPosts),
		"inserted_reviews":         atomic.LoadUint64(&m.InsertedReviews),
	}
}
