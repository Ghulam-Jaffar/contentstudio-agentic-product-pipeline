package sentiment

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// --- mocks ---

type mockSentimentAgent struct {
	label string
	score float64
	err   error
	mu    sync.Mutex
	calls int
}

func (m *mockSentimentAgent) Request(_ context.Context, _ string, _ map[string]interface{}) (map[string]interface{}, error) {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()
	if m.err != nil {
		return nil, m.err
	}
	return map[string]interface{}{
		"label": m.label,
		"score": m.score,
	}, nil
}

func makeParsedMention(id, text string) []byte {
	m := kafkamodels.ListeningMention{
		MentionID: id,
		TopicID:   "topic-1",
		Platform:  "twitter",
		NativeID:  id,
		PostText:  text,
	}
	data, _ := json.Marshal(m)
	return data
}

// --- tests ---

func TestHandleParsedMention(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		nilAgent       bool
		agentLabel     string
		agentScore     float64
		agentErr       error
		produceErr     error
		payload        []byte
		wantErr        bool
		wantMsgCount   int
		wantLabel      string
		wantScore      float64
		wantAgentCalls int
		wantTopic      string
	}{
		{
			name:           "success enriches with sentiment",
			agentLabel:     "positive",
			agentScore:     0.85,
			payload:        makeParsedMention("tw:1", "I love this product"),
			wantMsgCount:   1,
			wantLabel:      "positive",
			wantScore:      0.85,
			wantAgentCalls: 1,
			wantTopic:      kafkamodels.TopicListeningEnriched,
		},
		{
			name:           "agent failure falls back to empty label",
			agentErr:       fmt.Errorf("AI service down"),
			payload:        makeParsedMention("tw:2", "Some text"),
			wantMsgCount:   1,
			wantLabel:      "",
			wantScore:      0,
			wantAgentCalls: 1,
		},
		{
			name:           "empty text skips agent call",
			agentLabel:     "positive",
			agentScore:     0.9,
			payload:        makeParsedMention("tw:3", ""),
			wantMsgCount:   1,
			wantAgentCalls: 0,
		},
		{
			name:         "nil agent skips analysis and still produces",
			nilAgent:     true,
			payload:      makeParsedMention("tw:4", "Hello world"),
			wantMsgCount: 1,
		},
		{
			name:    "invalid JSON returns error",
			payload: []byte("{bad"),
			wantErr: true,
		},
		{
			name:           "empty agent label produces message with empty label",
			agentLabel:     "",
			agentScore:     0.0,
			payload:        makeParsedMention("tw:5", "Some neutral text"),
			wantMsgCount:   1,
			wantLabel:      "",
			wantAgentCalls: 1,
		},
		{
			name:           "produce error is propagated",
			agentLabel:     "negative",
			agentScore:     -0.5,
			produceErr:     fmt.Errorf("kafka down"),
			payload:        makeParsedMention("tw:6", "Bad stuff"),
			wantErr:        true,
			wantAgentCalls: 1,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prod := &mockProducerRecorder{err: tc.produceErr}
			log, _ := logger.NewTestLogger()

			var agent *mockSentimentAgent
			var svc *SentimentService
			if tc.nilAgent {
				svc = NewSentimentService(nil, prod, log)
			} else {
				agent = &mockSentimentAgent{label: tc.agentLabel, score: tc.agentScore, err: tc.agentErr}
				svc = NewSentimentService(agent, prod, log)
			}

			err := svc.HandleParsedMention(context.Background(), "", nil, tc.payload)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v, wantErr=%v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}

			prod.mu.Lock()
			msgCount := len(prod.messages)
			prod.mu.Unlock()

			if msgCount != tc.wantMsgCount {
				t.Fatalf("message count: want %d, got %d", tc.wantMsgCount, msgCount)
			}

			if tc.wantMsgCount > 0 {
				prod.mu.Lock()
				msg := prod.messages[0]
				prod.mu.Unlock()

				if tc.wantTopic != "" && msg.Topic != tc.wantTopic {
					t.Errorf("topic: want %q, got %q", tc.wantTopic, msg.Topic)
				}

				var enriched kafkamodels.ListeningMention
				if err := json.Unmarshal(msg.Value, &enriched); err != nil {
					t.Fatalf("unmarshal enriched mention: %v", err)
				}
				if enriched.SentimentLabel != tc.wantLabel {
					t.Errorf("sentiment_label: want %q, got %q", tc.wantLabel, enriched.SentimentLabel)
				}
				if enriched.SentimentScore != tc.wantScore {
					t.Errorf("sentiment_score: want %f, got %f", tc.wantScore, enriched.SentimentScore)
				}
			}

			if agent != nil {
				agent.mu.Lock()
				calls := agent.calls
				agent.mu.Unlock()
				if calls != tc.wantAgentCalls {
					t.Errorf("agent calls: want %d, got %d", tc.wantAgentCalls, calls)
				}
			}
		})
	}
}
