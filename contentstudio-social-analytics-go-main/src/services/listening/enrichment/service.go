// Package enrichment attaches AI-generated tags and context hints to mentions.
// It batches requests (see defaultBatchSize) because the AI agents API is rate
// limited per call, and caches per-topic context so repeated mentions from the
// same topic reuse a single expensive context lookup.
package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	defaultBatchSize        = 50
	defaultFlushInterval    = 30 * time.Second
	defaultMaxBufferAge     = 5 * time.Minute
	defaultCacheTTL         = 5 * time.Minute
	defaultBackfillInterval = 5 * time.Minute
	defaultBackfillLookback = 24 * time.Hour
	defaultBackfillDelay    = 2 * time.Minute
	defaultBackfillLimit    = 200
)

// AIAnalyzer abstracts the HTTP call to the AI agents batch endpoint.
type AIAnalyzer interface {
	AnalyzeBatch(ctx context.Context, mentions []MentionPayload, topicCtx TopicContext) ([]MentionResult, error)
}

// MentionWriter abstracts ClickHouse batch inserts for enriched mentions.
type MentionWriter interface {
	InsertMentions(ctx context.Context, mentions []chmodels.ListeningMentionRow) error
}

// BackfillSource abstracts reading stale, still-unenriched mentions from ClickHouse.
type BackfillSource interface {
	ListMentionsMissingEnrichment(
		ctx context.Context,
		updatedAfter time.Time,
		updatedBefore time.Time,
		limit int,
	) ([]chmodels.ListeningMentionRow, error)
}

// ContextProvider abstracts reading the AI context for a topic.
type ContextProvider interface {
	GetAIContext(ctx context.Context, topicID string) (mongoModels.TopicContextSnapshot, error)
}

// MentionPayload is sent to the AI agents endpoint.
type MentionPayload struct {
	MentionID string `json:"mention_id"`
	Text      string `json:"text"`
}

// MentionResult is returned from the AI agents endpoint.
type MentionResult struct {
	MentionID      string   `json:"mention_id"`
	SentimentLabel string   `json:"sentiment_label"`
	SentimentScore float64  `json:"sentiment_score"`
	AITags         []string `json:"ai_tags"`
}

type cachedContext struct {
	snap      mongoModels.TopicContextSnapshot
	expiresAt time.Time
}

type bufferedMention struct {
	mention    kafkamodels.ListeningMention
	bufferedAt time.Time
	source     string
}

type BackfillMetrics struct {
	Candidates uint64
	Enriched   uint64
	Failed     uint64
}

const (
	bufferSourceLive     = "live"
	bufferSourceBackfill = "backfill"
)

// EnrichmentService buffers parsed mentions by topic and batch-enriches them
// via the AI agents endpoint, then writes enriched rows to ClickHouse.
type EnrichmentService struct {
	analyzer AIAnalyzer
	writer   MentionWriter
	ctx      ContextProvider
	log      *logger.Logger

	batchSize     int
	flushInterval time.Duration
	maxBufferAge  time.Duration
	cacheTTL      time.Duration

	mu      sync.Mutex
	buffers map[string][]bufferedMention

	cacheMu      sync.RWMutex
	contextCache map[string]cachedContext

	backfillSource   BackfillSource
	backfillInterval time.Duration
	backfillLookback time.Duration
	backfillDelay    time.Duration
	backfillLimit    int
	backfillMetrics  BackfillMetrics
}

// NewEnrichmentService creates a new EnrichmentService.
func NewEnrichmentService(
	analyzer AIAnalyzer,
	writer MentionWriter,
	ctx ContextProvider,
	log *logger.Logger,
) *EnrichmentService {
	return &EnrichmentService{
		analyzer:         analyzer,
		writer:           writer,
		ctx:              ctx,
		log:              log,
		batchSize:        defaultBatchSize,
		flushInterval:    defaultFlushInterval,
		maxBufferAge:     defaultMaxBufferAge,
		cacheTTL:         defaultCacheTTL,
		buffers:          make(map[string][]bufferedMention),
		contextCache:     make(map[string]cachedContext),
		backfillInterval: defaultBackfillInterval,
		backfillLookback: defaultBackfillLookback,
		backfillDelay:    defaultBackfillDelay,
		backfillLimit:    defaultBackfillLimit,
	}
}

// WithBackfillSource enables periodic enrichment backfill for rows that missed
// the live AI pass.
func (s *EnrichmentService) WithBackfillSource(
	source BackfillSource,
	interval time.Duration,
	lookback time.Duration,
) *EnrichmentService {
	s.backfillSource = source
	if interval > 0 {
		s.backfillInterval = interval
	}
	if lookback > 0 {
		s.backfillLookback = lookback
	}
	return s
}

// HandleParsedMention is a kafka.MessageHandler that buffers mentions for batch enrichment.
func (s *EnrichmentService) HandleParsedMention(ctx context.Context, _ string, _ []byte, value []byte) error {
	if s.analyzer == nil {
		return nil
	}

	var mention kafkamodels.ListeningMention
	if err := json.Unmarshal(value, &mention); err != nil {
		return fmt.Errorf("EnrichmentService.HandleParsedMention: unmarshal: %w", err)
	}
	if mention.PostText == "" {
		return nil
	}

	shouldFlush := s.bufferMention(mention, time.Now().UTC(), bufferSourceLive)

	if shouldFlush {
		s.flushTopic(ctx, mention.TopicID)
	}

	return nil
}

// StartFlushLoop runs a periodic flush of all topic buffers until ctx is canceled.
func (s *EnrichmentService) StartFlushLoop(ctx context.Context) {
	if s.analyzer == nil {
		<-ctx.Done()
		return
	}

	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.flushAll(context.Background())
			return
		case <-ticker.C:
			s.flushAll(ctx)
		}
	}
}

// StartBackfillLoop periodically requeues rows that were inserted into
// ClickHouse without AI enrichment.
func (s *EnrichmentService) StartBackfillLoop(ctx context.Context) {
	if s.analyzer == nil || s.backfillSource == nil {
		<-ctx.Done()
		return
	}

	s.log.Info().
		Dur("interval", s.backfillInterval).
		Dur("lookback", s.backfillLookback).
		Dur("delay", s.backfillDelay).
		Int("limit", s.backfillLimit).
		Msg("Enrichment backfill loop started")

	s.RunBackfillOnce(ctx)

	ticker := time.NewTicker(s.backfillInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.RunBackfillOnce(ctx)
		}
	}
}

// RunBackfillOnce scans ClickHouse for mentions still missing sentiment and
// routes them through the normal enrichment path.
func (s *EnrichmentService) RunBackfillOnce(ctx context.Context) {
	if s.analyzer == nil || s.backfillSource == nil {
		return
	}

	cutoff := time.Now().UTC().Add(-s.backfillDelay)
	from := cutoff.Add(-s.backfillLookback)

	rows, err := s.backfillSource.ListMentionsMissingEnrichment(ctx, from, cutoff, s.backfillLimit)
	if err != nil {
		s.log.Error().Err(err).Msg("Enrichment backfill query failed")
		return
	}
	if len(rows) == 0 {
		return
	}
	atomic.AddUint64(&s.backfillMetrics.Candidates, uint64(len(rows)))

	buffered := 0
	bufferedAt := time.Now().UTC()
	for _, row := range rows {
		mention := rowToMention(row)
		if mention.PostText == "" || mention.TopicID == "" || mention.MentionID == "" {
			continue
		}
		if s.bufferMention(mention, bufferedAt, bufferSourceBackfill) {
			s.flushTopic(ctx, mention.TopicID)
		}
		buffered++
	}

	if buffered == 0 {
		return
	}

	s.log.Info().
		Int("candidate_rows", len(rows)).
		Int("buffered", buffered).
		Uint64("backfill_candidates_total", atomic.LoadUint64(&s.backfillMetrics.Candidates)).
		Uint64("backfill_enriched_total", atomic.LoadUint64(&s.backfillMetrics.Enriched)).
		Uint64("backfill_failed_total", atomic.LoadUint64(&s.backfillMetrics.Failed)).
		Msg("Queued mentions for enrichment backfill")

	s.flushAll(ctx)
}

func (s *EnrichmentService) flushAll(ctx context.Context) {
	s.mu.Lock()
	topicIDs := make([]string, 0, len(s.buffers))
	for topicID := range s.buffers {
		topicIDs = append(topicIDs, topicID)
	}
	s.mu.Unlock()

	for _, topicID := range topicIDs {
		s.flushTopic(ctx, topicID)
	}
}

func (s *EnrichmentService) flushTopic(ctx context.Context, topicID string) {
	s.mu.Lock()
	items := s.buffers[topicID]
	if len(items) == 0 {
		s.mu.Unlock()
		return
	}
	delete(s.buffers, topicID)
	s.mu.Unlock()

	for start := 0; start < len(items); start += s.batchSize {
		end := start + s.batchSize
		if end > len(items) {
			end = len(items)
		}
		s.processBatch(ctx, topicID, items[start:end])
	}
}

func (s *EnrichmentService) processBatch(ctx context.Context, topicID string, items []bufferedMention) {
	log := s.log.With().Str("topic_id", topicID).Int("batch_size", len(items)).Logger()
	backfillItems := countBufferedItemsBySource(items, bufferSourceBackfill)

	for _, item := range items {
		if time.Since(item.bufferedAt) > s.maxBufferAge {
			log.Warn().Msg("Mentions exceeded max buffer age, enriching now")
			break
		}
	}

	payloads := make([]MentionPayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, MentionPayload{
			MentionID: item.mention.MentionID,
			Text:      item.mention.PostText,
		})
	}

	snap := s.getAIContext(ctx, topicID)
	results, err := s.analyzer.AnalyzeBatch(ctx, payloads, TopicContext{
		AIContext:     snap.AIContext,
		TopicName:     snap.TopicName,
		TopicType:     snap.TopicType,
		TopicKeywords: snap.TopicKeywords,
		RelevanceHint: snap.Hint,
	})
	if err != nil {
		if backfillItems > 0 {
			atomic.AddUint64(&s.backfillMetrics.Failed, uint64(backfillItems))
		}
		log.Error().Err(err).Msg("AI batch analysis failed, re-buffering eligible mentions")
		s.rebuffer(topicID, items)
		return
	}

	resultMap := make(map[string]MentionResult, len(results))
	for _, result := range results {
		resultMap[result.MentionID] = result
	}

	rows := make([]chmodels.ListeningMentionRow, 0, len(items))
	now := time.Now().UTC()
	backfillEnriched := 0
	for _, item := range items {
		mention := item.mention
		if result, ok := resultMap[mention.MentionID]; ok {
			mention.SentimentLabel = result.SentimentLabel
			mention.SentimentScore = result.SentimentScore
			mention.AITags = result.AITags
			if item.source == bufferSourceBackfill {
				backfillEnriched++
			}
		}
		mention.UpdatedAt = now
		rows = append(rows, mentionToRow(mention))
	}

	if err := s.writer.InsertMentions(ctx, rows); err != nil {
		if backfillItems > 0 {
			atomic.AddUint64(&s.backfillMetrics.Failed, uint64(backfillItems))
		}
		log.Error().Err(err).Msg("ClickHouse insert failed for enriched mentions, re-buffering eligible mentions")
		s.rebuffer(topicID, items)
		return
	}
	if backfillEnriched > 0 {
		atomic.AddUint64(&s.backfillMetrics.Enriched, uint64(backfillEnriched))
	}

	log.Info().
		Int("enriched", len(rows)).
		Int("backfill_enriched", backfillEnriched).
		Uint64("backfill_candidates_total", atomic.LoadUint64(&s.backfillMetrics.Candidates)).
		Uint64("backfill_enriched_total", atomic.LoadUint64(&s.backfillMetrics.Enriched)).
		Uint64("backfill_failed_total", atomic.LoadUint64(&s.backfillMetrics.Failed)).
		Msg("Flushed enriched mentions to ClickHouse")
}

func (s *EnrichmentService) bufferMention(mention kafkamodels.ListeningMention, bufferedAt time.Time, source string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.buffers[mention.TopicID]
	for _, item := range items {
		if item.mention.MentionID == mention.MentionID {
			return len(items) >= s.batchSize
		}
	}

	s.buffers[mention.TopicID] = append(items, bufferedMention{
		mention:    mention,
		bufferedAt: bufferedAt,
		source:     source,
	})

	return len(s.buffers[mention.TopicID]) >= s.batchSize
}

func (s *EnrichmentService) GetBackfillMetrics() BackfillMetrics {
	return BackfillMetrics{
		Candidates: atomic.LoadUint64(&s.backfillMetrics.Candidates),
		Enriched:   atomic.LoadUint64(&s.backfillMetrics.Enriched),
		Failed:     atomic.LoadUint64(&s.backfillMetrics.Failed),
	}
}

func countBufferedItemsBySource(items []bufferedMention, source string) int {
	count := 0
	for _, item := range items {
		if item.source == source {
			count++
		}
	}
	return count
}

func (s *EnrichmentService) rebuffer(topicID string, items []bufferedMention) {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range items {
		if now.Sub(item.bufferedAt) >= s.maxBufferAge {
			s.log.Warn().
				Str("mention_id", item.mention.MentionID).
				Str("topic_id", topicID).
				Msg("Dropping aged mention from enrichment buffer")
			continue
		}
		s.buffers[topicID] = append(s.buffers[topicID], item)
	}
}

func (s *EnrichmentService) getAIContext(ctx context.Context, topicID string) mongoModels.TopicContextSnapshot {
	s.cacheMu.RLock()
	if cached, ok := s.contextCache[topicID]; ok && time.Now().Before(cached.expiresAt) {
		s.cacheMu.RUnlock()
		return cached.snap
	}
	s.cacheMu.RUnlock()

	snap, err := s.ctx.GetAIContext(ctx, topicID)
	if err != nil {
		s.log.Warn().Err(err).Str("topic_id", topicID).Msg("Failed to get AI context")
		return mongoModels.TopicContextSnapshot{}
	}

	s.cacheMu.Lock()
	s.contextCache[topicID] = cachedContext{
		snap:      snap,
		expiresAt: time.Now().Add(s.cacheTTL),
	}
	s.cacheMu.Unlock()

	return snap
}

func mentionToRow(m kafkamodels.ListeningMention) chmodels.ListeningMentionRow {
	return chmodels.ListeningMentionRow{
		MentionID:         m.MentionID,
		TopicID:           m.TopicID,
		Platform:          m.Platform,
		NativeID:          m.NativeID,
		ContentHash:       m.ContentHash,
		AuthorID:          m.AuthorID,
		AuthorName:        m.AuthorName,
		AuthorHandle:      m.AuthorHandle,
		AuthorImageURL:    m.AuthorImageURL,
		AuthorURL:         m.AuthorURL,
		AuthorFollowers:   m.AuthorFollowers,
		PostText:          m.PostText,
		Language:          m.Language,
		PostedAt:          m.PostedAt,
		MatchedKeywords:   m.MatchedKeywords,
		TotalEngagement:   m.TotalEngagement,
		LikesCount:        m.LikesCount,
		CommentsCount:     m.CommentsCount,
		SharesCount:       m.SharesCount,
		ContentType:       m.ContentType,
		MediaType:         m.MediaType,
		URL:               m.URL,
		MediaURLs:         m.MediaURLs,
		AITags:            m.AITags,
		SentimentLabel:    m.SentimentLabel,
		SentimentScore:    m.SentimentScore,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
		PostRead:          m.PostRead,
		PostIrrelevant:    m.PostIrrelevant,
		Bookmark:          m.Bookmark,
		SentimentOverride: m.SentimentOverride,
	}
}

func rowToMention(row chmodels.ListeningMentionRow) kafkamodels.ListeningMention {
	return kafkamodels.ListeningMention{
		MentionID:         row.MentionID,
		TopicID:           row.TopicID,
		Platform:          row.Platform,
		NativeID:          row.NativeID,
		ContentHash:       row.ContentHash,
		AuthorID:          row.AuthorID,
		AuthorName:        row.AuthorName,
		AuthorHandle:      row.AuthorHandle,
		AuthorImageURL:    row.AuthorImageURL,
		AuthorURL:         row.AuthorURL,
		AuthorFollowers:   row.AuthorFollowers,
		PostText:          row.PostText,
		Language:          row.Language,
		AITags:            row.AITags,
		PostedAt:          row.PostedAt,
		MatchedKeywords:   row.MatchedKeywords,
		TotalEngagement:   row.TotalEngagement,
		LikesCount:        row.LikesCount,
		CommentsCount:     row.CommentsCount,
		SharesCount:       row.SharesCount,
		ContentType:       row.ContentType,
		MediaType:         row.MediaType,
		URL:               row.URL,
		MediaURLs:         row.MediaURLs,
		SentimentLabel:    row.SentimentLabel,
		SentimentScore:    row.SentimentScore,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
		PostRead:          row.PostRead,
		PostIrrelevant:    row.PostIrrelevant,
		Bookmark:          row.Bookmark,
		SentimentOverride: row.SentimentOverride,
	}
}
