package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

type mockAnalyzer struct {
	mu             sync.Mutex
	err            error
	batchSizes     []int
	receivedTopics []TopicContext
	results        map[string]MentionResult
}

func (m *mockAnalyzer) AnalyzeBatch(
	_ context.Context,
	mentions []MentionPayload,
	topicCtx TopicContext,
) ([]MentionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	m.batchSizes = append(m.batchSizes, len(mentions))
	m.receivedTopics = append(m.receivedTopics, topicCtx)
	results := make([]MentionResult, 0, len(mentions))
	for _, mention := range mentions {
		if result, ok := m.results[mention.MentionID]; ok {
			results = append(results, result)
			continue
		}
		results = append(results, MentionResult{
			MentionID:      mention.MentionID,
			SentimentLabel: "positive",
			SentimentScore: 0.9,
			AITags:         []string{"Brand Mention"},
		})
	}
	return results, nil
}

type mockContextProvider struct {
	mu    sync.Mutex
	calls int
	snap  mongoModels.TopicContextSnapshot
	err   error
}

func (m *mockContextProvider) GetAIContext(_ context.Context, _ string) (mongoModels.TopicContextSnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	return m.snap, m.err
}

type mockWriter struct {
	mu    sync.Mutex
	err   error
	rows  []chmodels.ListeningMentionRow
	calls int
}

func (m *mockWriter) InsertMentions(_ context.Context, mentions []chmodels.ListeningMentionRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	if m.err != nil {
		return m.err
	}
	m.rows = append(m.rows, mentions...)
	return nil
}

type mockBackfillSource struct {
	rows  []chmodels.ListeningMentionRow
	err   error
	calls int
}

func (m *mockBackfillSource) ListMentionsMissingEnrichment(
	_ context.Context,
	_ time.Time,
	_ time.Time,
	_ int,
) ([]chmodels.ListeningMentionRow, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	return append([]chmodels.ListeningMentionRow(nil), m.rows...), nil
}

func makeParsedMention(topicID, mentionID string) []byte {
	payload, _ := json.Marshal(kafkamodels.ListeningMention{
		MentionID: mentionID,
		TopicID:   topicID,
		Platform:  "twitter",
		NativeID:  mentionID,
		PostText:  "Need help with ContentStudio",
		PostedAt:  time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	return payload
}

func newTestService(
	analyzer AIAnalyzer,
	writer MentionWriter,
	ctxProvider ContextProvider,
) *EnrichmentService {
	log, _ := logger.NewTestLogger()
	return NewEnrichmentService(analyzer, writer, ctxProvider, log)
}

func TestHandleParsedMention_FlushesInCappedBatches(t *testing.T) {
	t.Parallel()

	analyzer := &mockAnalyzer{}
	writer := &mockWriter{}
	ctxProvider := &mockContextProvider{snap: mongoModels.TopicContextSnapshot{AIContext: mongoModels.AIContext{BrandName: "ContentStudio"}}}
	svc := newTestService(analyzer, writer, ctxProvider)

	for i := 0; i < 55; i++ {
		payload := makeParsedMention("topic-1", fmt.Sprintf("mention-%02d", i))
		if err := svc.HandleParsedMention(context.Background(), "", nil, payload); err != nil {
			t.Fatalf("HandleParsedMention returned error: %v", err)
		}
	}

	svc.flushTopic(context.Background(), "topic-1")

	analyzer.mu.Lock()
	defer analyzer.mu.Unlock()
	if len(analyzer.batchSizes) != 2 {
		t.Fatalf("expected 2 analyzer calls, got %d", len(analyzer.batchSizes))
	}
	if analyzer.batchSizes[0] != 50 || analyzer.batchSizes[1] != 5 {
		t.Fatalf("expected batch sizes [50 5], got %v", analyzer.batchSizes)
	}

	writer.mu.Lock()
	defer writer.mu.Unlock()
	if len(writer.rows) != 55 {
		t.Fatalf("expected 55 enriched rows, got %d", len(writer.rows))
	}
}

func TestFlushTopic_RebuffersOnAnalyzerFailure(t *testing.T) {
	t.Parallel()

	analyzer := &mockAnalyzer{err: fmt.Errorf("agent unavailable")}
	writer := &mockWriter{}
	ctxProvider := &mockContextProvider{}
	svc := newTestService(analyzer, writer, ctxProvider)

	if err := svc.HandleParsedMention(context.Background(), "", nil, makeParsedMention("topic-1", "mention-1")); err != nil {
		t.Fatalf("HandleParsedMention returned error: %v", err)
	}

	svc.flushTopic(context.Background(), "topic-1")

	svc.mu.Lock()
	defer svc.mu.Unlock()
	if len(svc.buffers["topic-1"]) != 1 {
		t.Fatalf("expected mention to be rebuffered, got %d", len(svc.buffers["topic-1"]))
	}
}

func TestGetAIContext_UsesCache(t *testing.T) {
	t.Parallel()

	analyzer := &mockAnalyzer{}
	writer := &mockWriter{}
	ctxProvider := &mockContextProvider{snap: mongoModels.TopicContextSnapshot{AIContext: mongoModels.AIContext{BrandName: "ContentStudio"}}}
	svc := newTestService(analyzer, writer, ctxProvider)

	for i := 0; i < 2; i++ {
		got := svc.getAIContext(context.Background(), "topic-1")
		if got.AIContext.IsEmpty() {
			t.Fatal("expected cached context value")
		}
	}

	ctxProvider.mu.Lock()
	defer ctxProvider.mu.Unlock()
	if ctxProvider.calls != 1 {
		t.Fatalf("expected context provider to be called once, got %d", ctxProvider.calls)
	}
}

func TestHandleParsedMention_NoAnalyzerSkipsBuffering(t *testing.T) {
	t.Parallel()

	writer := &mockWriter{}
	ctxProvider := &mockContextProvider{}
	svc := newTestService(nil, writer, ctxProvider)

	if err := svc.HandleParsedMention(context.Background(), "", nil, makeParsedMention("topic-1", "mention-1")); err != nil {
		t.Fatalf("HandleParsedMention returned error: %v", err)
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()
	if len(svc.buffers) != 0 {
		t.Fatalf("expected no buffered mentions when analyzer is disabled, got %d topics", len(svc.buffers))
	}
}

func TestRunBackfillOnce_EnrichesRowsMissingSentiment(t *testing.T) {
	t.Parallel()

	analyzer := &mockAnalyzer{
		results: map[string]MentionResult{
			"mention-1": {
				MentionID:      "mention-1",
				SentimentLabel: "negative",
				SentimentScore: 0.82,
				AITags:         []string{"User Feedback"},
			},
		},
	}
	writer := &mockWriter{}
	ctxProvider := &mockContextProvider{snap: mongoModels.TopicContextSnapshot{AIContext: mongoModels.AIContext{BrandName: "ContentStudio"}}}
	backfill := &mockBackfillSource{
		rows: []chmodels.ListeningMentionRow{
			{
				MentionID:   "mention-1",
				TopicID:     "topic-1",
				Platform:    "twitter",
				NativeID:    "native-1",
				PostText:    "ContentStudio broke my workflow",
				CreatedAt:   time.Now().UTC().Add(-10 * time.Minute),
				UpdatedAt:   time.Now().UTC().Add(-10 * time.Minute),
				PostedAt:    time.Now().UTC().Add(-10 * time.Minute),
				ContentType: "post",
			},
		},
	}

	svc := newTestService(analyzer, writer, ctxProvider).WithBackfillSource(backfill, time.Minute, time.Hour)
	svc.RunBackfillOnce(context.Background())

	if backfill.calls != 1 {
		t.Fatalf("expected backfill source to be called once, got %d", backfill.calls)
	}

	writer.mu.Lock()
	defer writer.mu.Unlock()
	if len(writer.rows) != 1 {
		t.Fatalf("expected 1 backfilled row, got %d", len(writer.rows))
	}
	if writer.rows[0].SentimentLabel != "negative" {
		t.Fatalf("expected enriched sentiment label, got %q", writer.rows[0].SentimentLabel)
	}
	if len(writer.rows[0].AITags) != 1 || writer.rows[0].AITags[0] != "User Feedback" {
		t.Fatalf("expected enriched tags, got %v", writer.rows[0].AITags)
	}

	metrics := svc.GetBackfillMetrics()
	if metrics.Candidates != 1 || metrics.Enriched != 1 || metrics.Failed != 0 {
		t.Fatalf("unexpected backfill metrics: %+v", metrics)
	}
}

func TestRunBackfillOnce_DedupesWithBufferedMention(t *testing.T) {
	t.Parallel()

	analyzer := &mockAnalyzer{}
	writer := &mockWriter{}
	ctxProvider := &mockContextProvider{}
	backfill := &mockBackfillSource{
		rows: []chmodels.ListeningMentionRow{
			{
				MentionID: "mention-1",
				TopicID:   "topic-1",
				Platform:  "twitter",
				NativeID:  "native-1",
				PostText:  "Need help with ContentStudio",
				CreatedAt: time.Now().UTC().Add(-10 * time.Minute),
				UpdatedAt: time.Now().UTC().Add(-10 * time.Minute),
				PostedAt:  time.Now().UTC().Add(-10 * time.Minute),
			},
		},
	}

	svc := newTestService(analyzer, writer, ctxProvider).WithBackfillSource(backfill, time.Minute, time.Hour)
	if err := svc.HandleParsedMention(context.Background(), "", nil, makeParsedMention("topic-1", "mention-1")); err != nil {
		t.Fatalf("HandleParsedMention returned error: %v", err)
	}

	svc.RunBackfillOnce(context.Background())
	svc.flushTopic(context.Background(), "topic-1")

	analyzer.mu.Lock()
	defer analyzer.mu.Unlock()
	if len(analyzer.batchSizes) != 1 || analyzer.batchSizes[0] != 1 {
		t.Fatalf("expected one deduped analyzer call, got %v", analyzer.batchSizes)
	}
}

func TestRunBackfillOnce_TracksFailedBackfillBatches(t *testing.T) {
	t.Parallel()

	analyzer := &mockAnalyzer{err: fmt.Errorf("agent unavailable")}
	writer := &mockWriter{}
	ctxProvider := &mockContextProvider{}
	backfill := &mockBackfillSource{
		rows: []chmodels.ListeningMentionRow{
			{
				MentionID: "mention-1",
				TopicID:   "topic-1",
				Platform:  "twitter",
				NativeID:  "native-1",
				PostText:  "Need help with ContentStudio",
				CreatedAt: time.Now().UTC().Add(-10 * time.Minute),
				UpdatedAt: time.Now().UTC().Add(-10 * time.Minute),
				PostedAt:  time.Now().UTC().Add(-10 * time.Minute),
			},
		},
	}

	svc := newTestService(analyzer, writer, ctxProvider).WithBackfillSource(backfill, time.Minute, time.Hour)
	svc.RunBackfillOnce(context.Background())

	metrics := svc.GetBackfillMetrics()
	if metrics.Candidates != 1 || metrics.Enriched != 0 || metrics.Failed != 1 {
		t.Fatalf("unexpected backfill metrics: %+v", metrics)
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()
	if len(svc.buffers["topic-1"]) != 1 {
		t.Fatalf("expected failed backfill mention to be rebuffered, got %d", len(svc.buffers["topic-1"]))
	}
}

func TestEnrichmentService_PassesTopicContextToAnalyzer(t *testing.T) {
	t.Parallel()

	analyzer := &mockAnalyzer{results: map[string]MentionResult{}}
	provider := &mockContextProvider{
		snap: mongoModels.TopicContextSnapshot{
			AIContext: mongoModels.AIContext{
				BrandName: "Acme",
				Industry:  "SaaS",
			},
			Hint:          "B2B only",
			TopicName:     "Acme Brand",
			TopicType:     "own_brand",
			TopicKeywords: []string{"acme"},
		},
	}
	writer := &mockWriter{}
	svc := newTestService(analyzer, writer, provider)

	if err := svc.HandleParsedMention(context.Background(), "", nil, makeParsedMention("topic-ctx", "m-1")); err != nil {
		t.Fatalf("HandleParsedMention err: %v", err)
	}
	svc.flushTopic(context.Background(), "topic-ctx")

	analyzer.mu.Lock()
	defer analyzer.mu.Unlock()
	if len(analyzer.receivedTopics) != 1 {
		t.Fatalf("receivedTopics len=%d want 1", len(analyzer.receivedTopics))
	}
	tc := analyzer.receivedTopics[0]
	if tc.TopicName != "Acme Brand" {
		t.Fatalf("TopicName=%q want 'Acme Brand'", tc.TopicName)
	}
	if tc.RelevanceHint != "B2B only" {
		t.Fatalf("RelevanceHint=%q want 'B2B only'", tc.RelevanceHint)
	}
	if tc.AIContext.BrandName != "Acme" {
		t.Fatalf("AIContext.BrandName=%q want Acme", tc.AIContext.BrandName)
	}
	if tc.TopicType != "own_brand" {
		t.Fatalf("TopicType=%q want own_brand", tc.TopicType)
	}
}
