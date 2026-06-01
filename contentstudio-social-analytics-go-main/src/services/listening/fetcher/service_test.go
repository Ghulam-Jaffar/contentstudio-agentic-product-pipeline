package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/quota"
)

// --- mocks ---

type mockData365 struct {
	triggerErr     error
	triggerDelay   time.Duration
	pollErr        error
	fetchResults   []*social.Data365SearchResult
	fetchErr       error
	fetchCallCount int
	triggerCalls   []triggerCall
	fetchCalls     []fetchCall
	mu             sync.Mutex
}

type triggerCall struct {
	Platform string
	Keyword  string
	MaxPosts int
}

type fetchCall struct {
	Platform string
	Keyword  string
	Cursor   string
}

func (m *mockData365) TriggerSearch(_ context.Context, platform, keyword string, maxPosts int, _, _ time.Time, _ []string) error {
	m.mu.Lock()
	m.triggerCalls = append(m.triggerCalls, triggerCall{platform, keyword, maxPosts})
	delay := m.triggerDelay
	err := m.triggerErr
	m.mu.Unlock()

	if delay > 0 {
		time.Sleep(delay)
	}

	if err != nil {
		return err
	}
	return nil
}

func (m *mockData365) PollUntilFinished(_ context.Context, _, _ string, _ int, _, _ time.Time, _ []string) error {
	return m.pollErr
}

func (m *mockData365) FetchResults(_ context.Context, platform, keyword, cursor string, _, _ time.Time, _ []string) (*social.Data365SearchResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fetchCalls = append(m.fetchCalls, fetchCall{
		Platform: platform,
		Keyword:  keyword,
		Cursor:   cursor,
	})

	if m.fetchErr != nil {
		return nil, m.fetchErr
	}
	if m.fetchCallCount >= len(m.fetchResults) {
		return &social.Data365SearchResult{Data: json.RawMessage(`[]`)}, nil
	}
	result := m.fetchResults[m.fetchCallCount]
	m.fetchCallCount++
	return result, nil
}

type mockProducerRecorder struct {
	messages []producedMsg
	mu       sync.Mutex
	err      error
}

type producedMsg struct {
	Topic string
	Key   string
	Value []byte
}

func (m *mockProducerRecorder) Produce(_ context.Context, topic string, key, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.messages = append(m.messages, producedMsg{Topic: topic, Key: string(key), Value: value})
	return nil
}

func (m *mockProducerRecorder) Close() error { return nil }

type mockQuotaChecker struct {
	mu        sync.Mutex
	remaining []int
	err       error
	calls     int
	quotaIDs  []string
	topicIDs  []string
}

func (m *mockQuotaChecker) GetRemainingMentionBudget(_ context.Context, topicID string, quotaID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls++
	m.quotaIDs = append(m.quotaIDs, quotaID)
	m.topicIDs = append(m.topicIDs, topicID)
	if m.err != nil {
		return 0, m.err
	}
	if len(m.remaining) == 0 {
		return 0, nil
	}

	value := m.remaining[0]
	if len(m.remaining) > 1 {
		m.remaining = m.remaining[1:]
	}

	return value, nil
}

type mockTopicSyncMarker struct {
	mu          sync.Mutex
	callCount   int
	topicID     string
	workspaceID string
	eventAt     time.Time
	applied     bool
	err         error
}

func (m *mockTopicSyncMarker) MarkInitialSyncDone(_ context.Context, topicID, workspaceID string, eventAt time.Time) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callCount++
	m.topicID = topicID
	m.workspaceID = workspaceID
	m.eventAt = eventAt

	return m.applied, m.err
}

type mockEmptyBatchNotifier struct {
	mu        sync.Mutex
	calls     []emptyBatchCall
	triggered int
	err       error
}

type emptyBatchCall struct {
	Channel string
	Event   string
	Data    map[string]any
}

func (m *mockEmptyBatchNotifier) Trigger(channel, event string, data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	payload, _ := data.(map[string]any)
	m.calls = append(m.calls, emptyBatchCall{Channel: channel, Event: event, Data: payload})
	if m.err != nil {
		return m.err
	}
	m.triggered++
	return nil
}

type mockFetchProgressTracker struct {
	mu          sync.Mutex
	callCount   int
	topicID     string
	fetchedAt   time.Time
	cursors     map[string]string
	updateError error
}

func (m *mockFetchProgressTracker) UpdateLastFetched(_ context.Context, topicID string, fetchedAt time.Time, cursors map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callCount++
	m.topicID = topicID
	m.fetchedAt = fetchedAt
	m.cursors = cursors

	return m.updateError
}

type mockSuperAdminResolver struct {
	superAdminID string
	err          error
}

func (m *mockSuperAdminResolver) GetSuperAdminID(_ context.Context, _ string) (string, error) {
	return m.superAdminID, m.err
}

// mockRedisForLock satisfies redis.Client for DistributedLock.
type mockRedisForLock struct {
	store                map[string]string
	lastCompareDeleteErr error
	mu                   sync.Mutex
}

func newMockRedisForLock() *mockRedisForLock {
	return &mockRedisForLock{store: make(map[string]string)}
}

func (m *mockRedisForLock) Get(_ context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.store[key], nil
}
func (m *mockRedisForLock) Set(_ context.Context, key string, value interface{}, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[key] = fmt.Sprintf("%v", value)
	return nil
}
func (m *mockRedisForLock) Del(_ context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range keys {
		delete(m.store, k)
	}
	return nil
}

func (m *mockRedisForLock) SetNX(_ context.Context, key string, value interface{}, _ time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.store[key]; ok {
		return false, nil
	}
	m.store[key] = fmt.Sprintf("%v", value)
	return true, nil
}

func (m *mockRedisForLock) DecrBy(_ context.Context, key string, amount int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	valStr := m.store[key]
	var current int64
	if valStr != "" {
		fmt.Sscanf(valStr, "%d", &current)
	}
	newVal := current - amount
	m.store[key] = fmt.Sprintf("%d", newVal)
	return newVal, nil
}

func (m *mockRedisForLock) DecrByIfPositive(_ context.Context, key string, amount int64) (int64, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	valStr, exists := m.store[key]
	if !exists || valStr == "" {
		return -1, false, nil
	}
	var current int64
	fmt.Sscanf(valStr, "%d", &current)
	if current < amount {
		return current, false, nil
	}
	newVal := current - amount
	m.store[key] = fmt.Sprintf("%d", newVal)
	return newVal, true, nil
}

func (m *mockRedisForLock) IncrBy(_ context.Context, key string, amount int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	valStr := m.store[key]
	var current int64
	if valStr != "" {
		fmt.Sscanf(valStr, "%d", &current)
	}
	newVal := current + amount
	m.store[key] = fmt.Sprintf("%d", newVal)
	return newVal, nil
}

func (m *mockRedisForLock) IncrByIfExists(_ context.Context, key string, amount int64) (int64, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	valStr, exists := m.store[key]
	if !exists {
		return -1, false, nil
	}
	var current int64
	if valStr != "" {
		fmt.Sscanf(valStr, "%d", &current)
	}
	newVal := current + amount
	m.store[key] = fmt.Sprintf("%d", newVal)
	return newVal, true, nil
}

func (m *mockRedisForLock) DecrByIfExists(_ context.Context, key string, amount int64) (int64, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	valStr, exists := m.store[key]
	if !exists {
		return -1, false, nil
	}
	var current int64
	if valStr != "" {
		fmt.Sscanf(valStr, "%d", &current)
	}
	newVal := current - amount
	m.store[key] = fmt.Sprintf("%d", newVal)
	return newVal, true, nil
}

func (m *mockRedisForLock) Expire(_ context.Context, _ string, _ time.Duration) (bool, error) {
	return true, nil
}

func (m *mockRedisForLock) CompareAndDelete(ctx context.Context, key, expected string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastCompareDeleteErr = ctx.Err()
	if m.store[key] == expected {
		delete(m.store, key)
		return true, nil
	}
	return false, nil
}

func (m *mockRedisForLock) Close() error { return nil }

func newTestFetcher(d365 *mockData365, prod *mockProducerRecorder) *FetcherService {
	return newTestFetcherWithOptions(d365, prod, nil, nil)
}

func newTestFetcherWithOptions(
	d365 *mockData365,
	prod *mockProducerRecorder,
	quota QuotaChecker,
	progress FetchProgressTracker,
) *FetcherService {
	log, _ := logger.NewTestLogger()
	redisMock := newMockRedisForLock()
	lock := redis.NewDistributedLock(redisMock, log.Logger)
	return NewFetcherService(d365, prod, lock, log, 1, 50, 50, quota).WithProgressTracker(progress)
}

func sampleWorkOrder() kafkamodels.ListeningWorkOrder {
	return kafkamodels.ListeningWorkOrder{
		TopicID:          "topic-1",
		WorkspaceID:      "ws-1",
		IncludeKeywords:  []string{"golang", "rust"},
		EnabledPlatforms: []string{"twitter", "reddit"},
		MentionsLimit:    1000,
		ToDate:           time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC),
	}
}

func TestHandleWorkOrder_UsesSyncTypeCapsAndPersistsCursors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		syncType            string
		maxPostsIncremental int
		maxPostsInitial     int
		initialCursors      map[string]string
		fetchResults        []*social.Data365SearchResult
		wantTriggerMax      int
		wantFetchCursor     string
		wantProgressCursors map[string]string
	}{
		{
			name:                "incremental sync resumes from stored cursor and persists next cursor",
			syncType:            "recurring",
			maxPostsIncremental: 1,
			maxPostsInitial:     5,
			initialCursors: map[string]string{
				"twitter:golang": "cursor-start",
			},
			fetchResults: []*social.Data365SearchResult{
				{Data: json.RawMessage(`[{"id":"1"}]`), Cursor: "cursor-next"},
			},
			wantTriggerMax:  1,
			wantFetchCursor: "cursor-start",
			wantProgressCursors: map[string]string{
				"twitter:golang": "cursor-next",
			},
		},
		{
			name:                "initial sync uses initial cap and clears exhausted cursor",
			syncType:            "initial",
			maxPostsIncremental: 1,
			maxPostsInitial:     5,
			fetchResults: []*social.Data365SearchResult{
				{Data: json.RawMessage(`[{"id":"1"}]`), Cursor: ""},
			},
			wantTriggerMax:  5,
			wantFetchCursor: "",
			wantProgressCursors: map[string]string{
				"twitter:golang": "",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d365 := &mockData365{fetchResults: tc.fetchResults}
			prod := &mockProducerRecorder{}
			progress := &mockFetchProgressTracker{}
			log, _ := logger.NewTestLogger()
			redisMock := newMockRedisForLock()
			lock := redis.NewDistributedLock(redisMock, log.Logger)
			svc := NewFetcherService(
				d365,
				prod,
				lock,
				log,
				1,
				tc.maxPostsIncremental,
				tc.maxPostsInitial,
				nil,
			).WithProgressTracker(progress)

			order := sampleWorkOrder()
			order.SyncType = tc.syncType
			order.IncludeKeywords = []string{"golang"}
			order.EnabledPlatforms = []string{"twitter"}
			order.Cursors = tc.initialCursors

			data, _ := json.Marshal(order)
			if err := svc.HandleWorkOrder(context.Background(), "test", nil, data); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			d365.mu.Lock()
			if len(d365.triggerCalls) != 1 {
				t.Fatalf("expected 1 trigger call, got %d", len(d365.triggerCalls))
			}
			if d365.triggerCalls[0].MaxPosts != tc.wantTriggerMax {
				t.Fatalf("expected max_posts %d, got %d", tc.wantTriggerMax, d365.triggerCalls[0].MaxPosts)
			}
			if len(d365.fetchCalls) != 1 {
				t.Fatalf("expected 1 fetch call, got %d", len(d365.fetchCalls))
			}
			if d365.fetchCalls[0].Cursor != tc.wantFetchCursor {
				t.Fatalf("expected fetch cursor %q, got %q", tc.wantFetchCursor, d365.fetchCalls[0].Cursor)
			}
			d365.mu.Unlock()

			progress.mu.Lock()
			defer progress.mu.Unlock()
			if progress.callCount != 1 {
				t.Fatalf("expected progress update once, got %d", progress.callCount)
			}
			if !reflect.DeepEqual(progress.cursors, tc.wantProgressCursors) {
				t.Fatalf("expected cursors %+v, got %+v", tc.wantProgressCursors, progress.cursors)
			}
		})
	}
}

// --- tests ---

func TestPrepareKeyword_Instagram(t *testing.T) {
	t.Parallel()
	tests := []struct {
		platform string
		keyword  string
		want     string
	}{
		{"instagram", "travel", "#travel"},
		{"instagram", "#travel", "#travel"},
		{"twitter", "news", "news"},
		{"facebook", "golang", "golang"},
	}
	for _, tt := range tests {
		t.Run(tt.platform+"_"+tt.keyword, func(t *testing.T) {
			t.Parallel()
			got := prepareKeyword(tt.platform, tt.keyword)
			if got != tt.want {
				t.Errorf("prepareKeyword(%q, %q) = %q, want %q", tt.platform, tt.keyword, got, tt.want)
			}
		})
	}
}

func TestHandleWorkOrder_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		keywords          []string
		platforms         []string
		fetchResults      []*social.Data365SearchResult
		wantTriggerCount  int
		wantProducedExact int // when > 0 enforces exact count
		wantProducedMin   int // when > 0 enforces lower bound
		wantFirstKeyword  string
	}{
		{
			name:      "fans out across all keyword/platform pairs",
			keywords:  []string{"golang", "rust"},
			platforms: []string{"twitter", "reddit"},
			fetchResults: []*social.Data365SearchResult{
				{Data: json.RawMessage(`[{"id":"1"}]`), Cursor: ""},
			},
			wantTriggerCount: 4,
			wantProducedMin:  1,
		},
		{
			name:      "follows pagination cursor across pages",
			keywords:  []string{"golang"},
			platforms: []string{"twitter"},
			fetchResults: []*social.Data365SearchResult{
				{Data: json.RawMessage(`[{"id":"1"}]`), Cursor: "page2"},
				{Data: json.RawMessage(`[{"id":"2"}]`), Cursor: ""},
			},
			wantTriggerCount:  1,
			wantProducedExact: 2,
		},
		{
			name:      "instagram keywords are prefixed with hashtag",
			keywords:  []string{"travel"},
			platforms: []string{"instagram"},
			fetchResults: []*social.Data365SearchResult{
				{Data: json.RawMessage(`[{"id":"1"}]`)},
			},
			wantTriggerCount: 1,
			wantFirstKeyword: "#travel",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d365 := &mockData365{fetchResults: tc.fetchResults}
			prod := &mockProducerRecorder{}
			svc := newTestFetcher(d365, prod)

			order := sampleWorkOrder()
			order.IncludeKeywords = tc.keywords
			order.EnabledPlatforms = tc.platforms

			data, _ := json.Marshal(order)
			if err := svc.HandleWorkOrder(context.Background(), "test", nil, data); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			d365.mu.Lock()
			triggerCount := len(d365.triggerCalls)
			firstKeyword := ""
			if triggerCount > 0 {
				firstKeyword = d365.triggerCalls[0].Keyword
			}
			d365.mu.Unlock()

			if triggerCount != tc.wantTriggerCount {
				t.Errorf("trigger count: want %d, got %d", tc.wantTriggerCount, triggerCount)
			}
			if tc.wantFirstKeyword != "" && firstKeyword != tc.wantFirstKeyword {
				t.Errorf("first keyword: want %q, got %q", tc.wantFirstKeyword, firstKeyword)
			}

			prod.mu.Lock()
			defer prod.mu.Unlock()
			if tc.wantProducedExact > 0 && len(prod.messages) != tc.wantProducedExact {
				t.Errorf("produced count: want exactly %d, got %d", tc.wantProducedExact, len(prod.messages))
			}
			if tc.wantProducedMin > 0 && len(prod.messages) < tc.wantProducedMin {
				t.Errorf("produced count: want at least %d, got %d", tc.wantProducedMin, len(prod.messages))
			}
			for i, msg := range prod.messages {
				if msg.Topic != kafkamodels.TopicListeningRaw {
					t.Errorf("message %d topic: want %q, got %q", i, kafkamodels.TopicListeningRaw, msg.Topic)
				}
			}
		})
	}
}

func TestHandleWorkOrder_ShortCircuit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		rawPayload            []byte // when non-nil, used verbatim (e.g. malformed JSON)
		triggerErr            error
		pollErr               error
		preLockToken          string // when non-empty, pre-fills the topic lock key
		quotaCheckerErr       error
		quotaRemaining        []int // when non-nil, attaches a quota checker
		superAdminResolverErr error
		wantHandlerErr        bool
		requireZeroTriggers   bool
		requireZeroProduced   bool
	}{
		{
			name:                "invalid JSON returns unmarshal error",
			rawPayload:          []byte("{bad"),
			wantHandlerErr:      true,
			requireZeroTriggers: true,
			requireZeroProduced: true,
		},
		{
			name:                "lock already held by another worker is a no-op",
			preLockToken:        "other-worker-token",
			requireZeroTriggers: true,
		},
		{
			name:                "trigger error fails the order so the offset is not committed",
			triggerErr:          fmt.Errorf("api down"),
			wantHandlerErr:      true,
			requireZeroProduced: true,
		},
		{
			name:                "poll error fails the order so the offset is not committed",
			pollErr:             fmt.Errorf("timeout"),
			wantHandlerErr:      true,
			requireZeroProduced: true,
		},
		{
			name:                  "super admin resolver failure short-circuits without erroring",
			quotaRemaining:        []int{100},
			superAdminResolverErr: fmt.Errorf("subscription lookup failed"),
			requireZeroTriggers:   true,
		},
		{
			name:                "quota checker failure fails closed without triggering",
			quotaCheckerErr:     fmt.Errorf("database unreachable"),
			requireZeroTriggers: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d365 := &mockData365{triggerErr: tc.triggerErr, pollErr: tc.pollErr}
			prod := &mockProducerRecorder{}
			log, _ := logger.NewTestLogger()

			redisMock := newMockRedisForLock()
			if tc.preLockToken != "" {
				redisMock.store["listening:fetch:topic-1"] = tc.preLockToken
			}
			lock := redis.NewDistributedLock(redisMock, log.Logger)

			var quotaCheck QuotaChecker
			if tc.quotaCheckerErr != nil || tc.quotaRemaining != nil {
				quotaCheck = &mockQuotaChecker{err: tc.quotaCheckerErr, remaining: tc.quotaRemaining}
			}

			svc := NewFetcherService(d365, prod, lock, log, 1, 50, 50, quotaCheck)
			if tc.superAdminResolverErr != nil {
				svc = svc.WithSuperAdminResolver(&mockSuperAdminResolver{err: tc.superAdminResolverErr})
			}

			data := tc.rawPayload
			if data == nil {
				order := sampleWorkOrder()
				data, _ = json.Marshal(order)
			}

			err := svc.HandleWorkOrder(context.Background(), "test", nil, data)
			if tc.wantHandlerErr && err == nil {
				t.Fatalf("expected handler error, got nil")
			}
			if !tc.wantHandlerErr && err != nil {
				t.Fatalf("expected no handler error, got %v", err)
			}

			if tc.requireZeroTriggers {
				d365.mu.Lock()
				if got := len(d365.triggerCalls); got != 0 {
					t.Errorf("expected 0 trigger calls, got %d", got)
				}
				d365.mu.Unlock()
			}

			if tc.requireZeroProduced {
				prod.mu.Lock()
				if got := len(prod.messages); got != 0 {
					t.Errorf("expected 0 produced messages, got %d", got)
				}
				prod.mu.Unlock()
			}
		})
	}
}

func TestHandleWorkOrder_UnsupportedSearchIsSkippedWhenCoverageIsComplete(t *testing.T) {
	t.Parallel()

	d365 := &mockData365{
		triggerErr: &social.UnsupportedSearchError{
			Platform:   "twitter",
			Keyword:    "golang",
			StatusCode: 403,
		},
	}
	prod := &mockProducerRecorder{}
	svc := newTestFetcher(d365, prod)

	order := sampleWorkOrder()
	order.SyncType = "recurring"
	order.IncludeKeywords = []string{"golang"}
	order.EnabledPlatforms = []string{"twitter"}

	data, _ := json.Marshal(order)
	if err := svc.HandleWorkOrder(context.Background(), "test", nil, data); err != nil {
		t.Fatalf("expected unsupported query to be skipped without failing handler, got %v", err)
	}

	prod.mu.Lock()
	defer prod.mu.Unlock()
	if len(prod.messages) != 0 {
		t.Fatalf("expected no produced messages, got %d", len(prod.messages))
	}
}

func TestHandleWorkOrder_DistributedBudget(t *testing.T) {
	tests := []struct {
		name           string
		quotaRemaining int
		fetchResults   []*social.Data365SearchResult
		triggerErr     error
		wantHandlerErr bool
		wantBudgetEnd  string
	}{
		{
			name:           "fan-out releases unused reservation when fewer items return than reserved",
			quotaRemaining: 200,
			fetchResults: []*social.Data365SearchResult{
				{Data: json.RawMessage(`{"items": [{},{},{},{},{}]}`), Cursor: ""},
			},
			wantBudgetEnd: "195",
		},
		{
			name:           "bare array response counts each item before releasing unused budget",
			quotaRemaining: 100,
			fetchResults: []*social.Data365SearchResult{
				{Data: json.RawMessage(`[{"id":"1"},{"id":"2"},{"id":"3"},{"id":"4"},{"id":"5"}]`), Cursor: ""},
			},
			wantBudgetEnd: "95",
		},
		{
			name:           "trigger error releases the entire reservation",
			quotaRemaining: 100,
			triggerErr:     fmt.Errorf("api down"),
			wantHandlerErr: true,
			wantBudgetEnd:  "100",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			log := logger.NewNop()
			store := make(map[string]string)
			rdb := &mockRedisForLock{store: store}
			lock := redis.NewDistributedLock(rdb, log.Logger)
			tracker := quota.NewDistributedQuotaTracker(rdb, log)

			d365 := &mockData365{
				fetchResults: tc.fetchResults,
				triggerErr:   tc.triggerErr,
			}
			prod := &mockProducerRecorder{}
			quotaCheck := &mockQuotaChecker{remaining: []int{tc.quotaRemaining}}

			svc := NewFetcherService(d365, prod, lock, log, 1, 50, 50, quotaCheck).
				WithDistributedQuota(tracker)

			order := kafkamodels.ListeningWorkOrder{
				TopicID:          "topic-1",
				WorkspaceID:      "ws-1",
				SuperAdminID:     "sa-1",
				EnabledPlatforms: []string{"twitter"},
				IncludeKeywords:  []string{"golang"},
			}
			val, _ := json.Marshal(order)

			err := svc.HandleWorkOrder(ctx, "topic", nil, val)
			if tc.wantHandlerErr && err == nil {
				t.Fatal("expected handler error")
			}
			if !tc.wantHandlerErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			rdb.mu.Lock()
			got := store["listening:budget:sa-1"]
			rdb.mu.Unlock()
			if got != tc.wantBudgetEnd {
				t.Errorf("budget end: want %q, got %q", tc.wantBudgetEnd, got)
			}
		})
	}
}

// Regression test for the staging retry-forever loop: a topic with several
// platforms where every (platform, keyword) returns 403 from Data365 used to
// trip the intra-work-order race in fetchAllPlatforms — one goroutine would
// flip coverageIncomplete via reserveKeywordBudget !ok, and HandleWorkOrder
// would return an error that left the scheduler re-emitting the work order
// indefinitely. With pre-partitioned per-platform budgets, every platform gets
// a private slice, all queries are skipped cleanly, and initial sync is marked
// done so the topic exits the awaiting-first-batch state.
func TestHandleWorkOrder_AllUnsupportedAcrossPlatformsMarksDone(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := logger.NewNop()
	store := make(map[string]string)
	rdb := &mockRedisForLock{store: store}
	lock := redis.NewDistributedLock(rdb, log.Logger)
	tracker := quota.NewDistributedQuotaTracker(rdb, log)

	d365 := &mockData365{
		triggerErr: &social.UnsupportedSearchError{StatusCode: 403},
	}
	prod := &mockProducerRecorder{}
	// 10k user budget, 6 platforms, 3 keywords — the exact shape that drove
	// the staging incident (10000 / 6 = 1666 per platform private slice, no
	// race for a shared counter).
	quotaCheck := &mockQuotaChecker{remaining: []int{10000}}
	marker := &mockTopicSyncMarker{applied: true}

	svc := NewFetcherService(d365, prod, lock, log, 1, 5000, 5000, quotaCheck).
		WithDistributedQuota(tracker).
		WithTopicSyncMarker(marker)

	order := kafkamodels.ListeningWorkOrder{
		TopicID:          "topic-staging",
		WorkspaceID:      "ws-staging",
		SuperAdminID:     "sa-staging",
		SyncType:         "initial",
		EnabledPlatforms: []string{"twitter", "threads", "reddit", "instagram", "facebook", "tiktok"},
		IncludeKeywords:  []string{"braintree", "square", "worldpay"},
		ToDate:           time.Date(2026, 5, 12, 6, 53, 13, 0, time.UTC),
	}
	val, _ := json.Marshal(order)

	if err := svc.HandleWorkOrder(ctx, "test", nil, val); err != nil {
		t.Fatalf("expected work order to complete without error so scheduler stops retrying, got %v", err)
	}

	marker.mu.Lock()
	defer marker.mu.Unlock()
	if marker.callCount != 1 {
		t.Fatalf("expected MarkInitialSyncDone to be called exactly once, got %d", marker.callCount)
	}
}

type mockQuotaTopicCounter struct {
	mu     sync.Mutex
	count  int
	err    error
	calls  int
	calledWith []string
}

func (m *mockQuotaTopicCounter) CountActiveTopicsForQuota(_ context.Context, quotaID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	m.calledWith = append(m.calledWith, quotaID)
	return m.count, m.err
}

// Asserts the fair-share clamp: with 5 topics sharing a 100k super-admin
// budget and 6 platforms × 5 keywords per topic, each per-keyword Data365
// request is capped at 100000/5/6/5 = 666 instead of the configured 5000.
// This is the math that lets 5 simultaneous initial syncs all complete
// within the budget, instead of topic 1 draining everything and topics 2-5
// staying permanently stuck at is_initial_sync_done=false.
func TestHandleWorkOrder_QuotaTopicCounterClampsPerKeywordCap(t *testing.T) {
	t.Parallel()

	d365 := &mockData365{
		fetchResults: []*social.Data365SearchResult{
			{Data: json.RawMessage(`[]`), Cursor: ""},
		},
	}
	prod := &mockProducerRecorder{}
	quotaCheck := &mockQuotaChecker{remaining: []int{100000}}
	counter := &mockQuotaTopicCounter{count: 5}
	marker := &mockTopicSyncMarker{applied: true}

	log, _ := logger.NewTestLogger()
	redisMock := newMockRedisForLock()
	lock := redis.NewDistributedLock(redisMock, log.Logger)
	svc := NewFetcherService(d365, prod, lock, log, 1, 500, 5000, quotaCheck).
		WithQuotaTopicCounter(counter).
		WithTopicSyncMarker(marker)

	order := kafkamodels.ListeningWorkOrder{
		TopicID:          "topic-fair",
		WorkspaceID:      "ws-fair",
		SuperAdminID:     "sa-fair",
		SyncType:         "initial",
		EnabledPlatforms: []string{"twitter", "threads", "reddit", "instagram", "facebook", "tiktok"},
		IncludeKeywords:  []string{"k1", "k2", "k3", "k4", "k5"},
		ToDate:           time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC),
	}
	data, _ := json.Marshal(order)

	if err := svc.HandleWorkOrder(context.Background(), "test", nil, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedCap := 100000 / 5 / 6 / 5 // = 666
	d365.mu.Lock()
	defer d365.mu.Unlock()
	if len(d365.triggerCalls) == 0 {
		t.Fatal("expected at least one TriggerSearch call")
	}
	for _, call := range d365.triggerCalls {
		if call.MaxPosts != expectedCap {
			t.Errorf("expected MaxPosts=%d (clamped fair share), got %d on platform=%s keyword=%s",
				expectedCap, call.MaxPosts, call.Platform, call.Keyword)
		}
	}

	counter.mu.Lock()
	defer counter.mu.Unlock()
	if counter.calls == 0 {
		t.Fatal("expected QuotaTopicCounter to be consulted")
	}
	if counter.calledWith[0] != "sa-fair" {
		t.Errorf("expected counter to be called with super_admin_id 'sa-fair', got %q", counter.calledWith[0])
	}
}

// When the counter returns 1 (single-topic super-admin), the clamp must
// not apply — the topic should get the configured cap unchanged.
func TestHandleWorkOrder_QuotaTopicCounterSkipsClampForSingleTopic(t *testing.T) {
	t.Parallel()

	d365 := &mockData365{
		fetchResults: []*social.Data365SearchResult{
			{Data: json.RawMessage(`[]`), Cursor: ""},
		},
	}
	prod := &mockProducerRecorder{}
	quotaCheck := &mockQuotaChecker{remaining: []int{100000}}
	counter := &mockQuotaTopicCounter{count: 1}
	marker := &mockTopicSyncMarker{applied: true}

	log, _ := logger.NewTestLogger()
	redisMock := newMockRedisForLock()
	lock := redis.NewDistributedLock(redisMock, log.Logger)
	svc := NewFetcherService(d365, prod, lock, log, 1, 500, 5000, quotaCheck).
		WithQuotaTopicCounter(counter).
		WithTopicSyncMarker(marker)

	order := kafkamodels.ListeningWorkOrder{
		TopicID:          "topic-solo",
		WorkspaceID:      "ws-solo",
		SuperAdminID:     "sa-solo",
		SyncType:         "initial",
		EnabledPlatforms: []string{"twitter"},
		IncludeKeywords:  []string{"k1"},
		ToDate:           time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC),
	}
	data, _ := json.Marshal(order)

	if err := svc.HandleWorkOrder(context.Background(), "test", nil, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	d365.mu.Lock()
	defer d365.mu.Unlock()
	if len(d365.triggerCalls) == 0 {
		t.Fatal("expected at least one TriggerSearch call")
	}
	if got := d365.triggerCalls[0].MaxPosts; got != 5000 {
		t.Errorf("expected MaxPosts=5000 (configured cap unchanged for single-topic quota), got %d", got)
	}
}

func TestHandleWorkOrder_UpdatesLastFetchedAfterSuccessfulRunWithoutMentions(t *testing.T) {
	t.Parallel()

	d365 := &mockData365{
		fetchResults: []*social.Data365SearchResult{
			{Data: json.RawMessage(`[]`), Cursor: ""},
		},
	}
	prod := &mockProducerRecorder{}
	progress := &mockFetchProgressTracker{}
	svc := newTestFetcherWithOptions(d365, prod, nil, progress)

	order := sampleWorkOrder()
	order.IncludeKeywords = []string{"golang"}
	order.EnabledPlatforms = []string{"twitter"}
	order.ToDate = time.Date(2026, 4, 10, 18, 0, 0, 0, time.UTC)

	data, _ := json.Marshal(order)
	if err := svc.HandleWorkOrder(context.Background(), "test", nil, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	progress.mu.Lock()
	defer progress.mu.Unlock()
	if progress.callCount != 1 {
		t.Fatalf("expected progress update once, got %d", progress.callCount)
	}
	if progress.topicID != order.TopicID {
		t.Fatalf("expected progress topic %q, got %q", order.TopicID, progress.topicID)
	}
	if !progress.fetchedAt.Equal(order.ToDate) {
		t.Fatalf("expected fetched_at %s, got %s", order.ToDate, progress.fetchedAt)
	}
}

func TestReleaseLock_UsesFreshContext(t *testing.T) {
	t.Parallel()

	log, _ := logger.NewTestLogger()
	redisMock := newMockRedisForLock()
	redisMock.store["listening:fetch:topic-1"] = "token-1"
	lock := redis.NewDistributedLock(redisMock, log.Logger)
	svc := NewFetcherService(&mockData365{}, &mockProducerRecorder{}, lock, log, 1, 50, 50, nil)

	svc.releaseLock("listening:fetch:topic-1", "token-1")

	redisMock.mu.Lock()
	defer redisMock.mu.Unlock()
	if redisMock.lastCompareDeleteErr != nil {
		t.Fatalf("expected release to avoid canceled context, got %v", redisMock.lastCompareDeleteErr)
	}
	if _, ok := redisMock.store["listening:fetch:topic-1"]; ok {
		t.Fatal("expected lock to be removed")
	}
}

func TestFetchAndEmitResults_StopsWhenBudgetExhausted(t *testing.T) {
	t.Parallel()

	d365 := &mockData365{
		fetchResults: []*social.Data365SearchResult{
			{Data: json.RawMessage(`{"items":[{"id":"1"},{"id":"2"},{"id":"3"}]}`), Cursor: "page-2"},
			{Data: json.RawMessage(`{"items":[{"id":"4"}]}`), Cursor: ""},
		},
	}
	prod := &mockProducerRecorder{}
	svc := newTestFetcherWithOptions(d365, prod, nil, nil)

	order := sampleWorkOrder()
	order.IncludeKeywords = []string{"golang"}
	order.EnabledPlatforms = []string{"twitter"}

	var budgetLeft int64
	if err := svc.fetchAndEmitResults(context.Background(), order, "twitter", "golang", order.WorkspaceID, &budgetLeft, 2, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prod.mu.Lock()
	defer prod.mu.Unlock()
	if len(prod.messages) != 1 {
		t.Fatalf("expected exactly one raw payload before budget stopped pagination, got %d", len(prod.messages))
	}

	if budgetLeft != 0 {
		t.Errorf("expected shared local budget to remain unchanged after a fully spent reservation, got %d", budgetLeft)
	}
}

func TestHandleWorkOrder_UsesSuperAdminQuotaID(t *testing.T) {
	t.Parallel()

	d365 := &mockData365{}
	prod := &mockProducerRecorder{}
	quota := &mockQuotaChecker{remaining: []int{0}}
	svc := newTestFetcherWithOptions(d365, prod, quota, nil)

	order := sampleWorkOrder()
	order.SuperAdminID = "owner-1"

	data, _ := json.Marshal(order)
	if err := svc.HandleWorkOrder(context.Background(), "test", nil, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	quota.mu.Lock()
	defer quota.mu.Unlock()
	if len(quota.quotaIDs) == 0 {
		t.Fatal("expected quota checker to be called")
	}
	if quota.quotaIDs[0] != "owner-1" {
		t.Fatalf("expected super admin quota id, got %q", quota.quotaIDs[0])
	}
}

func TestHandleWorkOrder_InitialSyncMark(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		syncType          string
		triggerErr        error
		triggerDelay      time.Duration
		pollErr           error
		quotaRemaining    []int
		markerApplied     bool
		markerErr         error
		wantHandlerErr    bool
		wantMarkCallCount int
	}{
		{
			name:              "marks done after successful initial sync",
			syncType:          "initial",
			markerApplied:     true,
			wantMarkCallCount: 1,
		},
		{
			name:              "skips mark for recurring sync",
			syncType:          "recurring",
			markerApplied:     true,
			wantMarkCallCount: 0,
		},
		{
			name:              "skips mark when fetch fails so the order retries",
			syncType:          "initial",
			pollErr:           fmt.Errorf("timeout"),
			wantHandlerErr:    true,
			wantMarkCallCount: 0,
		},
		{
			name:              "marks initial sync done when all queries are unsupported within private slices",
			syncType:          "initial",
			triggerErr:        &social.UnsupportedSearchError{Platform: "twitter", Keyword: "golang", StatusCode: 403},
			triggerDelay:      2 * time.Millisecond,
			quotaRemaining:    []int{100},
			markerApplied:     true,
			wantHandlerErr:    false,
			wantMarkCallCount: 1,
		},
		{
			name:              "swallows stale-mark result without surfacing an error",
			syncType:          "initial",
			markerApplied:     false,
			wantMarkCallCount: 1,
		},
		{
			name:              "swallows mongo errors without failing the work order",
			syncType:          "initial",
			markerErr:         fmt.Errorf("mongo down"),
			wantMarkCallCount: 1,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d365 := &mockData365{
				fetchResults: []*social.Data365SearchResult{
					{Data: json.RawMessage(`[{"id":"1"}]`), Cursor: ""},
				},
				triggerErr:   tc.triggerErr,
				triggerDelay: tc.triggerDelay,
				pollErr:      tc.pollErr,
			}
			prod := &mockProducerRecorder{}
			marker := &mockTopicSyncMarker{applied: tc.markerApplied, err: tc.markerErr}
			var quotaChecker QuotaChecker
			if tc.quotaRemaining != nil {
				quotaChecker = &mockQuotaChecker{remaining: append([]int(nil), tc.quotaRemaining...)}
			}
			svc := newTestFetcherWithOptions(d365, prod, quotaChecker, nil).WithTopicSyncMarker(marker)

			order := sampleWorkOrder()
			order.SyncType = tc.syncType
			if tc.quotaRemaining != nil {
				order.IncludeKeywords = []string{"golang"}
				order.EnabledPlatforms = []string{"twitter", "reddit", "threads"}
			}

			data, _ := json.Marshal(order)
			err := svc.HandleWorkOrder(context.Background(), "test", nil, data)
			if tc.wantHandlerErr && err == nil {
				t.Fatal("expected error so the work order retries instead of being marked complete")
			}
			if !tc.wantHandlerErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			marker.mu.Lock()
			defer marker.mu.Unlock()
			if marker.callCount != tc.wantMarkCallCount {
				t.Fatalf("expected %d mark calls, got %d", tc.wantMarkCallCount, marker.callCount)
			}
			if tc.wantMarkCallCount > 0 {
				if marker.topicID != order.TopicID {
					t.Errorf("expected topic ID %q, got %q", order.TopicID, marker.topicID)
				}
				if marker.workspaceID != order.WorkspaceID {
					t.Errorf("expected workspace ID %q, got %q", order.WorkspaceID, marker.workspaceID)
				}
				if !marker.eventAt.Equal(order.ToDate) {
					t.Errorf("expected eventAt %s, got %s", order.ToDate, marker.eventAt)
				}
			}
		})
	}
}

func TestHandleWorkOrder_FirstBatchEmptyNotification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		syncType          string
		fetchResults      []*social.Data365SearchResult
		markerApplied     bool
		markerErr         error
		notifierErr       error
		omitNotifier      bool
		omitWorkspace     bool
		quotaRemaining    []int
		wantTriggered     int
		wantPusherChannel string
		wantPusherReason  string
	}{
		{
			name:              "fires empty event when initial sync emits zero items",
			syncType:          "initial",
			fetchResults:      []*social.Data365SearchResult{{Data: json.RawMessage(`[]`), Cursor: ""}},
			markerApplied:     true,
			wantTriggered:     1,
			wantPusherChannel: "listening-channel-ws-1",
			wantPusherReason:  "no_matches",
		},
		{
			name:          "skips empty event when initial sync emitted items",
			syncType:      "initial",
			fetchResults:  []*social.Data365SearchResult{{Data: json.RawMessage(`[{"id":"1"}]`), Cursor: ""}},
			markerApplied: true,
			wantTriggered: 0,
		},
		{
			name:          "skips empty event when sync mark was stale (newer event already landed)",
			syncType:      "initial",
			fetchResults:  []*social.Data365SearchResult{{Data: json.RawMessage(`[]`), Cursor: ""}},
			markerApplied: false,
			wantTriggered: 0,
		},
		{
			name:          "skips empty event for recurring syncs even with zero items",
			syncType:      "recurring",
			fetchResults:  []*social.Data365SearchResult{{Data: json.RawMessage(`[]`), Cursor: ""}},
			markerApplied: true,
			wantTriggered: 0,
		},
		{
			name:          "skips empty event when MarkInitialSyncDone errored",
			syncType:      "initial",
			fetchResults:  []*social.Data365SearchResult{{Data: json.RawMessage(`[]`), Cursor: ""}},
			markerErr:     fmt.Errorf("mongo down"),
			wantTriggered: 0,
		},
		{
			name:          "swallows pusher trigger errors without failing the work order",
			syncType:      "initial",
			fetchResults:  []*social.Data365SearchResult{{Data: json.RawMessage(`[]`), Cursor: ""}},
			markerApplied: true,
			notifierErr:   fmt.Errorf("pusher down"),
			wantTriggered: 0,
		},
		{
			name:          "no panic when notifier is omitted",
			syncType:      "initial",
			fetchResults:  []*social.Data365SearchResult{{Data: json.RawMessage(`[]`), Cursor: ""}},
			markerApplied: true,
			omitNotifier:  true,
			wantTriggered: 0,
		},
		{
			name:          "skips empty event when workspace_id missing (cannot derive channel)",
			syncType:      "initial",
			fetchResults:  []*social.Data365SearchResult{{Data: json.RawMessage(`[]`), Cursor: ""}},
			markerApplied: true,
			omitWorkspace: true,
			wantTriggered: 0,
		},
		{
			name:           "skips empty event when quota prevents any fetch attempt",
			syncType:       "initial",
			markerApplied:  true,
			quotaRemaining: []int{0},
			wantTriggered:  0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d365 := &mockData365{fetchResults: tc.fetchResults}
			prod := &mockProducerRecorder{}
			marker := &mockTopicSyncMarker{applied: tc.markerApplied, err: tc.markerErr}
			notifier := &mockEmptyBatchNotifier{err: tc.notifierErr}

			var quotaChecker QuotaChecker
			if tc.quotaRemaining != nil {
				quotaChecker = &mockQuotaChecker{remaining: append([]int(nil), tc.quotaRemaining...)}
			}

			svc := newTestFetcherWithOptions(d365, prod, quotaChecker, nil).WithTopicSyncMarker(marker)
			if !tc.omitNotifier {
				svc = svc.WithEmptyBatchNotifier(notifier)
			}

			order := sampleWorkOrder()
			order.IncludeKeywords = []string{"golang"}
			order.EnabledPlatforms = []string{"twitter"}
			order.SyncType = tc.syncType
			if tc.omitWorkspace {
				order.WorkspaceID = ""
			}

			data, _ := json.Marshal(order)
			if err := svc.HandleWorkOrder(context.Background(), "test", nil, data); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			notifier.mu.Lock()
			defer notifier.mu.Unlock()
			if notifier.triggered != tc.wantTriggered {
				t.Fatalf("expected %d successful triggers, got %d (calls=%v)", tc.wantTriggered, notifier.triggered, notifier.calls)
			}
			if tc.wantTriggered > 0 {
				if len(notifier.calls) == 0 {
					t.Fatal("expected at least one Trigger call")
				}
				call := notifier.calls[0]
				if call.Channel != tc.wantPusherChannel {
					t.Errorf("expected channel %q, got %q", tc.wantPusherChannel, call.Channel)
				}
				if call.Event != "listening.mentions.first_batch_empty" {
					t.Errorf("expected event %q, got %q", "listening.mentions.first_batch_empty", call.Event)
				}
				if call.Data["reason"] != tc.wantPusherReason {
					t.Errorf("expected reason %q, got %v", tc.wantPusherReason, call.Data["reason"])
				}
				if call.Data["topic_id"] != order.TopicID {
					t.Errorf("expected topic_id %q, got %v", order.TopicID, call.Data["topic_id"])
				}
				if call.Data["workspace_id"] != order.WorkspaceID {
					t.Errorf("expected workspace_id %q, got %v", order.WorkspaceID, call.Data["workspace_id"])
				}
			}
		})
	}
}

func TestHandleWorkOrder_DistributedQuotaCoordination(t *testing.T) {
	ctx := context.Background()
	log := logger.NewNop()
	store := make(map[string]string)
	rdb := &mockRedisForLock{store: store}
	lock := redis.NewDistributedLock(rdb, log.Logger)
	tracker := quota.NewDistributedQuotaTracker(rdb, log)

	d365 := &mockData365{
		fetchResults: []*social.Data365SearchResult{
			{Data: json.RawMessage(`{"items": [{},{},{},{},{},{},{},{},{},{}]}`), Cursor: ""},
		},
	}
	prod := &mockProducerRecorder{}
	quota := &mockQuotaChecker{remaining: []int{100}}

	svc := NewFetcherService(d365, prod, lock, log, 1, 50, 50, quota).
		WithDistributedQuota(tracker)

	order := kafkamodels.ListeningWorkOrder{
		TopicID:          "topic-1",
		WorkspaceID:      "ws-1",
		SuperAdminID:     "sa-1",
		EnabledPlatforms: []string{"twitter"},
		IncludeKeywords:  []string{"golang"},
	}
	val, _ := json.Marshal(order)

	// Run Worker 1
	err := svc.HandleWorkOrder(ctx, "topic", nil, val)
	if err != nil {
		t.Fatalf("Worker 1 failed: %v", err)
	}

	// Verify budget was reserved (50) and unused released (50-10=40 back).
	// Net consumption: 10 items → Redis should show 90.
	budgetKey := "listening:budget:sa-1"
	if store[budgetKey] != "90" {
		t.Errorf("Expected Redis budget to be 90, got %s", store[budgetKey])
	}

	// Worker 2 with fresh data
	d365.mu.Lock()
	d365.fetchCallCount = 0
	d365.triggerCalls = nil
	d365.mu.Unlock()
	d365.fetchResults = []*social.Data365SearchResult{
		{Data: json.RawMessage(`{"items": [{},{},{},{},{}]}`), Cursor: ""},
	}

	err = svc.HandleWorkOrder(ctx, "topic", nil, val)
	if err != nil {
		t.Fatalf("Worker 2 failed: %v", err)
	}

	// Worker 2 consumed 5 more → 90 - 5 = 85
	if store[budgetKey] != "85" {
		t.Errorf("Expected Redis budget to be 85, got %s", store[budgetKey])
	}
}

func TestHandleWorkOrder_FailClosedOnDistributedQuotaError(t *testing.T) {
	t.Parallel()

	log, _ := logger.NewTestLogger()
	// Redis mock that returns errors
	rdb := &redis.MockRedisClient{
		GetFunc: func(_ context.Context, _ string) (string, error) {
			return "", fmt.Errorf("redis connection refused")
		},
	}
	tracker := quota.NewDistributedQuotaTracker(rdb, log)

	d365 := &mockData365{}
	prod := &mockProducerRecorder{}
	quota := &mockQuotaChecker{remaining: []int{100}}

	redisMock := newMockRedisForLock()
	lock := redis.NewDistributedLock(redisMock, log.Logger)

	svc := NewFetcherService(d365, prod, lock, log, 1, 50, 50, quota).
		WithDistributedQuota(tracker)

	order := sampleWorkOrder()
	order.SuperAdminID = "sa-1"
	data, _ := json.Marshal(order)

	if err := svc.HandleWorkOrder(context.Background(), "test", nil, data); err != nil {
		t.Fatalf("expected graceful skip, got error: %v", err)
	}

	d365.mu.Lock()
	defer d365.mu.Unlock()
	if len(d365.triggerCalls) != 0 {
		t.Errorf("expected no trigger calls when Redis is down (fail-closed), got %d", len(d365.triggerCalls))
	}
}

// budgetReseedFallback's contract has two shapes: nil when QuotaChecker is
// unwired (so Reserve falls back to fail-closed instead of dereffing nil),
// and a closure that plumbs topicID + quotaID to the checker when wired —
// the latter is the shape that broke prod when the topic was missing.
func TestBudgetReseedFallback(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		quota       *mockQuotaChecker
		wantNil     bool
		topicID     string
		quotaID     string
		wantValue   int
		wantTopicID string
		wantQuotaID string
	}{
		{
			name:    "nil when QuotaChecker unwired",
			quota:   nil,
			wantNil: true,
			topicID: "topic-x",
			quotaID: "quota-x",
		},
		{
			name:        "plumbs topicID and quotaID through",
			quota:       &mockQuotaChecker{remaining: []int{700}},
			topicID:     "topic-abc",
			quotaID:     "quota-xyz",
			wantValue:   700,
			wantTopicID: "topic-abc",
			wantQuotaID: "quota-xyz",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log, _ := logger.NewTestLogger()
			redisMock := newMockRedisForLock()
			lock := redis.NewDistributedLock(redisMock, log.Logger)

			var checker QuotaChecker
			if tc.quota != nil {
				checker = tc.quota
			}
			svc := NewFetcherService(&mockData365{}, &mockProducerRecorder{}, lock, log, 1, 50, 50, checker)

			fb := svc.budgetReseedFallback(context.Background(), tc.topicID, tc.quotaID)
			if tc.wantNil {
				if fb != nil {
					t.Fatal("expected nil fallback when QuotaChecker is unwired")
				}
				return
			}
			if fb == nil {
				t.Fatal("expected non-nil fallback when QuotaChecker is wired")
			}

			got, err := fb()
			if err != nil {
				t.Fatalf("fallback returned unexpected error: %v", err)
			}
			if got != tc.wantValue {
				t.Fatalf("expected fallback to surface remaining=%d, got %d", tc.wantValue, got)
			}

			tc.quota.mu.Lock()
			defer tc.quota.mu.Unlock()
			if tc.quota.calls != 1 {
				t.Fatalf("expected exactly 1 QuotaChecker call from the closure, got %d", tc.quota.calls)
			}
			if len(tc.quota.topicIDs) != 1 || tc.quota.topicIDs[0] != tc.wantTopicID {
				t.Fatalf("expected topicID %q to be plumbed through, got %v", tc.wantTopicID, tc.quota.topicIDs)
			}
			if len(tc.quota.quotaIDs) != 1 || tc.quota.quotaIDs[0] != tc.wantQuotaID {
				t.Fatalf("expected quotaID %q to be plumbed through, got %v", tc.wantQuotaID, tc.quota.quotaIDs)
			}
		})
	}
}

// Integration: when the budget key has expired mid-fan-out (Redis empty but
// MongoDB still has remaining quota), reserveKeywordBudget must succeed via
// the inline reseed-and-retry path rather than fail-closing. This is the
// production scenario from the 2026-04-29 traces.
func TestReserveKeywordBudget_RecoversWhenBudgetKeyMissing(t *testing.T) {
	t.Parallel()

	log := logger.NewNop()
	rdb := &mockRedisForLock{store: make(map[string]string)} // budget key missing
	lock := redis.NewDistributedLock(rdb, log.Logger)
	tracker := quota.NewDistributedQuotaTracker(rdb, log)

	quotaCheck := &mockQuotaChecker{remaining: []int{300}}
	svc := NewFetcherService(&mockData365{}, &mockProducerRecorder{}, lock, log, 1, 50, 50, quotaCheck).
		WithDistributedQuota(tracker)

	reserved, ok, err := svc.reserveKeywordBudget(context.Background(), "topic-mid", "quota-mid", nil, 50)
	if err != nil {
		t.Fatalf("expected reseed to recover; got error: %v", err)
	}
	if !ok {
		t.Fatal("expected reservation to succeed after reseed")
	}
	if reserved != 50 {
		t.Fatalf("expected reserved=50 (maxPosts), got %d", reserved)
	}

	// 300 reseeded - 50 reserved = 250 remaining in Redis
	rdb.mu.Lock()
	stored := rdb.store["listening:budget:quota-mid"]
	rdb.mu.Unlock()
	if stored != "250" {
		t.Fatalf("expected reseeded budget to land at 250 (300-50), got %q", stored)
	}

	// QuotaChecker should have been invoked exactly once via the reseed closure
	// (no GetRemaining at fan-out level since we called reserveKeywordBudget directly).
	quotaCheck.mu.Lock()
	defer quotaCheck.mu.Unlock()
	if quotaCheck.calls != 1 {
		t.Fatalf("expected 1 QuotaChecker call from reseed, got %d", quotaCheck.calls)
	}
	if len(quotaCheck.topicIDs) != 1 || quotaCheck.topicIDs[0] != "topic-mid" {
		t.Fatalf("expected topicID 'topic-mid' captured by reseed, got %v", quotaCheck.topicIDs)
	}
}
