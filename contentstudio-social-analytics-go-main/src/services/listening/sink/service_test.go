package sink

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// --- mocks ---

type mockMentionWriter struct {
	insertErr error
	failCount int32 // number of times to fail before succeeding
	callCount int32
	rows      []chmodels.ListeningMentionRow
}

func (m *mockMentionWriter) InsertMentions(_ context.Context, mentions []chmodels.ListeningMentionRow) error {
	n := atomic.AddInt32(&m.callCount, 1)
	if n <= atomic.LoadInt32(&m.failCount) {
		return m.insertErr
	}
	m.rows = append(m.rows, mentions...)
	return nil
}

type mockTopicUpdater struct {
	incrementCalls    int
	limitReached      bool
	lastFetchCalls    int
	mentionsCount     int
	incrementErr      error
	countErr          error
	updateErr         error
	limitErr          error
	reserveCalls      int
	releaseCalls      int
	reserveErr        error
	releaseErr        error
	reserveDenied     bool
	firstBatchCalls   int
	firstBatchAlready bool
	firstBatchErr     error
}

func (m *mockTopicUpdater) GetMentionsCount(_ context.Context, _ string) (int, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}

	return m.mentionsCount, nil
}

func (m *mockTopicUpdater) IncrementMentionsCount(_ context.Context, _ string, count int) error {
	m.incrementCalls++
	m.mentionsCount += count
	return m.incrementErr
}

func (m *mockTopicUpdater) SetMentionsLimitReached(_ context.Context, _ string) error {
	m.limitReached = true
	return m.limitErr
}

func (m *mockTopicUpdater) UpdateLastFetched(_ context.Context, _ string, _ time.Time, _ map[string]string) error {
	m.lastFetchCalls++
	return m.updateErr
}

func (m *mockTopicUpdater) TryReserveMentionSlot(_ context.Context, _ string, _ int) (bool, int, error) {
	if m.reserveErr != nil {
		return false, 0, m.reserveErr
	}

	m.reserveCalls++
	if m.reserveDenied {
		return false, m.mentionsCount, nil
	}

	m.mentionsCount++
	return true, m.mentionsCount, nil
}

func (m *mockTopicUpdater) ReleaseMentionSlot(_ context.Context, _ string, _ int) error {
	m.releaseCalls++
	if m.mentionsCount > 0 {
		m.mentionsCount--
	}
	return m.releaseErr
}

func (m *mockTopicUpdater) MarkFirstMentionsReceived(_ context.Context, _ string) (bool, error) {
	if m.firstBatchErr != nil {
		return false, m.firstBatchErr
	}
	m.firstBatchCalls++
	if m.firstBatchAlready {
		return false, nil
	}
	m.firstBatchAlready = true
	return true, nil
}

// mockWorkspaceUpdater satisfies the WorkspaceUpdater interface for tests.
type mockWorkspaceUpdater struct {
	limitReached    bool
	mentionsCount   int
	mentionLimit    int
	incrementCalls  int
	limitReachedSet bool
	checkErr        error
	incrementErr    error
	setLimitErr     error
	reserveCalls    int
	releaseCalls    int
	reserveErr      error
	releaseErr      error
	reserveDenied   bool
}

func (m *mockWorkspaceUpdater) IsWorkspaceMentionLimitReached(_ context.Context, _ string) (bool, error) {
	return m.limitReached, m.checkErr
}

func (m *mockWorkspaceUpdater) IncrementWorkspaceMentionsCount(_ context.Context, _ string, count int) (int, int, error) {
	if m.incrementErr != nil {
		return 0, 0, m.incrementErr
	}
	m.incrementCalls++
	m.mentionsCount += count
	return m.mentionsCount, m.mentionLimit, nil
}

func (m *mockWorkspaceUpdater) SetWorkspaceMentionLimitReached(_ context.Context, _ string) error {
	m.limitReachedSet = true
	return m.setLimitErr
}

func (m *mockWorkspaceUpdater) TryReserveWorkspaceMention(_ context.Context, _ string) (bool, int, int, error) {
	if m.reserveErr != nil {
		return false, 0, 0, m.reserveErr
	}

	m.reserveCalls++
	if m.reserveDenied {
		return false, m.mentionsCount, m.mentionLimit, nil
	}

	m.mentionsCount++
	return true, m.mentionsCount, m.mentionLimit, nil
}

func (m *mockWorkspaceUpdater) ReleaseWorkspaceMentionReservation(_ context.Context, _ string) error {
	m.releaseCalls++
	if m.mentionsCount > 0 {
		m.mentionsCount--
	}
	return m.releaseErr
}

type firstBatchTriggerCall struct {
	channel string
	event   string
	data    interface{}
}

type mockFirstBatchNotifier struct {
	triggerErr error
	calls      []firstBatchTriggerCall
}

func (m *mockFirstBatchNotifier) Trigger(channel, event string, data interface{}) error {
	m.calls = append(m.calls, firstBatchTriggerCall{channel: channel, event: event, data: data})
	return m.triggerErr
}

func newTestSink(writer *mockMentionWriter, updater *mockTopicUpdater, prod *mockProducerRecorder) *SinkService {
	log, _ := logger.NewTestLogger()
	return NewSinkService(writer, updater, nil, prod, log, 3)
}

func makeEnrichedMention(id, topicID string, mentionsLimit int) []byte {
	m := kafkamodels.ListeningMention{
		MentionID:      id,
		TopicID:        topicID,
		MentionsLimit:  mentionsLimit,
		Platform:       "twitter",
		NativeID:       id,
		PostText:       "Test mention",
		SentimentLabel: "positive",
		SentimentScore: 0.8,
		PostedAt:       time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
	data, _ := json.Marshal(m)
	return data
}

// makeEnrichedMentionWithWorkspace mirrors makeEnrichedMention and additionally
// stamps the workspace ID so workspace-quota paths are exercised.
func makeEnrichedMentionWithWorkspace(id, topicID, workspaceID string, mentionsLimit int) []byte {
	payload := makeEnrichedMention(id, topicID, mentionsLimit)
	if workspaceID == "" {
		return payload
	}
	var m kafkamodels.ListeningMention
	_ = json.Unmarshal(payload, &m)
	m.WorkspaceID = workspaceID
	out, _ := json.Marshal(m)
	return out
}

// --- tests ---

func TestSinkService_HandleParsedMention(t *testing.T) {
	t.Parallel()

	type writerCfg struct {
		insertErr error
		failCount int32
	}
	type updaterCfg struct {
		mentionsCount int
		incrementErr  error
	}
	type workspaceCfg struct {
		attached      bool
		mentionLimit  int
		mentionsCount int
	}
	type want struct {
		handlerErrSubstr string // empty = no handler error
		writerCallCount  int32
		insertedRows     int
		firstBatchCalls  int
		topicIncCalls    int
		topicLimitFlag   bool
		wsIncCalls       int
		wsLimitFlag      bool
		dlqCount         int
		dlqStage         string
		dlqAttempt       int
	}

	tests := []struct {
		name       string
		rawPayload []byte // when non-nil, sent verbatim instead of generated
		mentionID  string
		topicID    string
		topicLimit int
		wsID       string
		writer     writerCfg
		updater    updaterCfg
		ws         workspaceCfg
		want       want
	}{
		{
			name:      "happy path persists row and increments topic only",
			mentionID: "tw:1", topicID: "topic-1",
			want: want{
				writerCallCount: 1, insertedRows: 1, firstBatchCalls: 1, topicIncCalls: 1,
			},
		},
		{
			name:       "invalid JSON returns unmarshal error",
			rawPayload: []byte("{bad"),
			want:       want{handlerErrSubstr: "unmarshal"},
		},
		{
			name:      "transient insert errors retry then succeed",
			mentionID: "tw:2", topicID: "topic-1",
			writer: writerCfg{insertErr: fmt.Errorf("connection reset"), failCount: 2},
			want: want{
				writerCallCount: 3, insertedRows: 1, firstBatchCalls: 1, topicIncCalls: 1,
			},
		},
		{
			name:      "insert exhausts retries and message is DLQd at sink stage",
			mentionID: "tw:3", topicID: "topic-1",
			writer: writerCfg{insertErr: fmt.Errorf("disk full"), failCount: 10},
			want: want{
				writerCallCount: 3, insertedRows: 0, topicIncCalls: 0,
				dlqCount: 1, dlqStage: "sink", dlqAttempt: 3,
			},
		},
		{
			name:      "topic increment failure DLQs at sink-bookkeeping after row is persisted",
			mentionID: "tw:4", topicID: "topic-1",
			updater: updaterCfg{incrementErr: fmt.Errorf("mongo down")},
			want: want{
				writerCallCount: 1, insertedRows: 1, firstBatchCalls: 1, topicIncCalls: 1,
				dlqCount: 1, dlqStage: "sink-bookkeeping",
			},
		},
		{
			name:      "topic threshold crossed never flips topic limit flag",
			mentionID: "tw:6", topicID: "topic-1", topicLimit: 100,
			updater: updaterCfg{mentionsCount: 99},
			want: want{
				writerCallCount: 1, insertedRows: 1, firstBatchCalls: 1, topicIncCalls: 1,
				topicLimitFlag: false,
			},
		},
		{
			name:      "topic over historical limit still inserts and counts",
			mentionID: "tw:7", topicID: "topic-1", topicLimit: 100,
			updater: updaterCfg{mentionsCount: 100},
			want: want{
				writerCallCount: 1, insertedRows: 1, firstBatchCalls: 1, topicIncCalls: 1,
				topicLimitFlag: false,
			},
		},
		{
			name:      "workspace limit reached flips workspace flag after insert",
			mentionID: "tw:8", topicID: "topic-1", topicLimit: 100, wsID: "ws-1",
			ws: workspaceCfg{attached: true, mentionLimit: 10, mentionsCount: 9},
			want: want{
				writerCallCount: 1, insertedRows: 1, firstBatchCalls: 1, topicIncCalls: 1,
				wsIncCalls: 1, wsLimitFlag: true,
			},
		},
		{
			name:      "topic-not-found error discards quietly without charging workspace",
			mentionID: "tw:9", topicID: "topic-1", topicLimit: 100, wsID: "ws-1",
			updater: updaterCfg{incrementErr: fmt.Errorf("topic not found: topic-1")},
			ws:      workspaceCfg{attached: true, mentionLimit: 10},
			want: want{
				writerCallCount: 1, insertedRows: 1, firstBatchCalls: 1, topicIncCalls: 1,
				wsIncCalls: 0, dlqCount: 0,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			writer := &mockMentionWriter{insertErr: tc.writer.insertErr, failCount: tc.writer.failCount}
			updater := &mockTopicUpdater{
				mentionsCount: tc.updater.mentionsCount,
				incrementErr:  tc.updater.incrementErr,
			}
			prod := &mockProducerRecorder{}
			svc := newTestSink(writer, updater, prod)

			var workspace *mockWorkspaceUpdater
			if tc.ws.attached {
				workspace = &mockWorkspaceUpdater{
					mentionLimit:  tc.ws.mentionLimit,
					mentionsCount: tc.ws.mentionsCount,
				}
				svc.workspaceUpdater = workspace
			}

			payload := tc.rawPayload
			if payload == nil {
				payload = makeEnrichedMentionWithWorkspace(tc.mentionID, tc.topicID, tc.wsID, tc.topicLimit)
			}

			err := svc.HandleParsedMention(context.Background(), "", nil, payload)
			if tc.want.handlerErrSubstr != "" {
				if err == nil {
					t.Fatalf("expected handler error containing %q, got nil", tc.want.handlerErrSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected handler error: %v", err)
			}

			if got := atomic.LoadInt32(&writer.callCount); got != tc.want.writerCallCount {
				t.Errorf("writer call count: want %d, got %d", tc.want.writerCallCount, got)
			}
			if len(writer.rows) != tc.want.insertedRows {
				t.Errorf("inserted rows: want %d, got %d", tc.want.insertedRows, len(writer.rows))
			}
			if updater.firstBatchCalls != tc.want.firstBatchCalls {
				t.Errorf("first-batch calls: want %d, got %d", tc.want.firstBatchCalls, updater.firstBatchCalls)
			}
			if updater.incrementCalls != tc.want.topicIncCalls {
				t.Errorf("topic increment calls: want %d, got %d", tc.want.topicIncCalls, updater.incrementCalls)
			}
			if updater.limitReached != tc.want.topicLimitFlag {
				t.Errorf("topic limit flag: want %v, got %v", tc.want.topicLimitFlag, updater.limitReached)
			}
			if updater.lastFetchCalls != 0 {
				t.Errorf("sink must never call UpdateLastFetched, got %d", updater.lastFetchCalls)
			}

			if tc.ws.attached {
				if workspace.incrementCalls != tc.want.wsIncCalls {
					t.Errorf("workspace increment calls: want %d, got %d", tc.want.wsIncCalls, workspace.incrementCalls)
				}
				if workspace.limitReachedSet != tc.want.wsLimitFlag {
					t.Errorf("workspace limit flag: want %v, got %v", tc.want.wsLimitFlag, workspace.limitReachedSet)
				}
			}

			prod.mu.Lock()
			defer prod.mu.Unlock()
			if len(prod.messages) != tc.want.dlqCount {
				t.Fatalf("DLQ message count: want %d, got %d", tc.want.dlqCount, len(prod.messages))
			}
			if tc.want.dlqCount > 0 {
				if prod.messages[0].Topic != kafkamodels.TopicListeningDLQ {
					t.Errorf("DLQ topic: want %q, got %q", kafkamodels.TopicListeningDLQ, prod.messages[0].Topic)
				}
				var dlq kafkamodels.ListeningDLQMessage
				if err := json.Unmarshal(prod.messages[0].Value, &dlq); err != nil {
					t.Fatalf("unmarshal DLQ: %v", err)
				}
				if tc.want.dlqStage != "" && dlq.Stage != tc.want.dlqStage {
					t.Errorf("DLQ stage: want %q, got %q", tc.want.dlqStage, dlq.Stage)
				}
				if tc.want.dlqAttempt != 0 && dlq.AttemptCount != tc.want.dlqAttempt {
					t.Errorf("DLQ attempt count: want %d, got %d", tc.want.dlqAttempt, dlq.AttemptCount)
				}
			}
		})
	}
}

func TestMentionToRow(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	m := kafkamodels.ListeningMention{
		MentionID:       "twitter:123",
		TopicID:         "topic-1",
		Platform:        "twitter",
		NativeID:        "123",
		ContentHash:     "abc",
		AuthorID:        "user1",
		AuthorName:      "User One",
		PostText:        "Hello",
		PostedAt:        now,
		MatchedKeywords: []string{"hello"},
		TotalEngagement: 100,
		ContentType:     "post",
		MediaType:       "text",
		URL:             "https://twitter.com/123",
		SentimentLabel:  "positive",
		SentimentScore:  0.9,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	row := mentionToRow(m)
	if row.MentionID != m.MentionID {
		t.Errorf("MentionID mismatch")
	}
	if row.TotalEngagement != 100 {
		t.Errorf("TotalEngagement mismatch")
	}
	if row.SentimentScore != 0.9 {
		t.Errorf("SentimentScore mismatch")
	}
	if len(row.MatchedKeywords) != 1 || row.MatchedKeywords[0] != "hello" {
		t.Errorf("MatchedKeywords mismatch")
	}
}

func TestSinkService_NotifyFirstBatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		attachNotifier    bool
		topicID           string
		workspaceID       string
		firstBatchAlready bool
		firstBatchErr     error
		triggerErr        error
		wantFirstBatchOps int
		wantTriggerCalls  int
	}{
		{
			name:              "no notifier wired still persists first-batch state for polling fallback",
			attachNotifier:    false,
			topicID:           "topic-1",
			workspaceID:       "ws-1",
			wantFirstBatchOps: 1,
			wantTriggerCalls:  0,
		},
		{
			name:              "empty topic id is treated as no-op",
			attachNotifier:    true,
			topicID:           "",
			workspaceID:       "ws-1",
			wantFirstBatchOps: 0,
			wantTriggerCalls:  0,
		},
		{
			name:              "first batch wins race and fires Pusher event",
			attachNotifier:    true,
			topicID:           "topic-1",
			workspaceID:       "ws-1",
			wantFirstBatchOps: 1,
			wantTriggerCalls:  1,
		},
		{
			name:              "subsequent batch loses race and does not fire",
			attachNotifier:    true,
			topicID:           "topic-1",
			workspaceID:       "ws-1",
			firstBatchAlready: true,
			wantFirstBatchOps: 1,
			wantTriggerCalls:  0,
		},
		{
			name:              "MarkFirstMentionsReceived error is swallowed without firing",
			attachNotifier:    true,
			topicID:           "topic-1",
			workspaceID:       "ws-1",
			firstBatchErr:     fmt.Errorf("mongo down"),
			wantFirstBatchOps: 0,
			wantTriggerCalls:  0,
		},
		{
			name:              "won race with empty workspace id defers transition for a later valid mention",
			attachNotifier:    true,
			topicID:           "topic-1",
			workspaceID:       "",
			wantFirstBatchOps: 0,
			wantTriggerCalls:  0,
		},
		{
			name:              "Pusher Trigger error is swallowed without failing handler",
			attachNotifier:    true,
			topicID:           "topic-1",
			workspaceID:       "ws-1",
			triggerErr:        fmt.Errorf("pusher 500"),
			wantFirstBatchOps: 1,
			wantTriggerCalls:  1,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			writer := &mockMentionWriter{}
			updater := &mockTopicUpdater{
				firstBatchAlready: tc.firstBatchAlready,
				firstBatchErr:     tc.firstBatchErr,
			}
			prod := &mockProducerRecorder{}
			svc := newTestSink(writer, updater, prod)

			var notifier *mockFirstBatchNotifier
			if tc.attachNotifier {
				notifier = &mockFirstBatchNotifier{triggerErr: tc.triggerErr}
				svc.WithFirstBatchNotifier(notifier)
			}

			payload := makeEnrichedMentionWithWorkspace("tw:fb", tc.topicID, tc.workspaceID, 0)

			if err := svc.HandleParsedMention(context.Background(), "", nil, payload); err != nil {
				t.Fatalf("HandleParsedMention should never fail on first-batch issues: %v", err)
			}
			if updater.firstBatchCalls != tc.wantFirstBatchOps {
				t.Fatalf("expected %d first-batch transitions, got %d", tc.wantFirstBatchOps, updater.firstBatchCalls)
			}

			if notifier == nil {
				return
			}

			if got := len(notifier.calls); got != tc.wantTriggerCalls {
				t.Fatalf("expected %d Trigger calls, got %d", tc.wantTriggerCalls, got)
			}

			if tc.wantTriggerCalls == 1 {
				call := notifier.calls[0]
				wantChannel := firstBatchChannelPrefix + tc.workspaceID
				if call.channel != wantChannel {
					t.Errorf("channel mismatch: want %q, got %q", wantChannel, call.channel)
				}
				if call.event != firstBatchEvent {
					t.Errorf("event mismatch: want %q, got %q", firstBatchEvent, call.event)
				}
				data, ok := call.data.(map[string]any)
				if !ok {
					t.Fatalf("expected map payload, got %T", call.data)
				}
				if data["topic_id"] != tc.topicID {
					t.Errorf("payload topic_id mismatch: want %q, got %v", tc.topicID, data["topic_id"])
				}
				if data["workspace_id"] != tc.workspaceID {
					t.Errorf("payload workspace_id mismatch: want %q, got %v", tc.workspaceID, data["workspace_id"])
				}
				if _, hasReceived := data["received_at"]; !hasReceived {
					t.Error("payload missing received_at")
				}
			}
		})
	}
}
