package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// mockListeningRepo implements the methods used by HandleListeningWork for testing.
type mockListeningRepo struct {
	topic *mongoModels.ListeningTopic
	err   error
}

func (m *mockListeningRepo) GetTopicByID(_ context.Context, _ string) (*mongoModels.ListeningTopic, error) {
	return m.topic, m.err
}

type mockListeningWorkspaceRepo struct {
	mentionsCount int
	mentionLimit  int
	exists        bool
	err           error
}

func (m *mockListeningWorkspaceRepo) GetWorkspaceUsage(_ context.Context, _ string) (int, int, bool, error) {
	return m.mentionsCount, m.mentionLimit, m.exists, m.err
}

func (m *mockListeningWorkspaceRepo) IsWorkspaceMentionLimitReached(_ context.Context, _ string) (bool, error) {
	return false, m.err
}

// mockProducer implements kafka.Producer for testing.
type mockProducer struct {
	produced []producedMessage
	err      error
}

type producedMessage struct {
	topic string
	key   []byte
	value []byte
}

func (m *mockProducer) Produce(_ context.Context, topic string, key, value []byte) error {
	if m.err != nil {
		return m.err
	}
	m.produced = append(m.produced, producedMessage{topic: topic, key: key, value: value})
	return nil
}

func (m *mockProducer) Close() error { return nil }

func newTestServer(repo *mockListeningRepo, workspaceRepo *mockListeningWorkspaceRepo, producer *mockProducer) *APIServer {
	log, _ := logger.NewTestLogger()
	s := &APIServer{
		Producer:               producer,
		Logger:                 log,
		ListeningWorkspaceRepo: workspaceRepo,
	}
	if repo != nil {
		s.ListeningRepo = repo
	}
	return s
}

func TestHandleListeningWork(t *testing.T) {
	t.Parallel()

	validTopic := &mongoModels.ListeningTopic{
		TopicID:          "t1",
		WorkspaceID:      "ws1",
		IncludeKeywords:  []string{"go", "golang"},
		EnabledPlatforms: []string{"twitter", "reddit"},
		MentionsLimit:    1000,
	}

	tests := []struct {
		name          string
		method        string
		body          any
		repo          *mockListeningRepo
		workspaceRepo *mockListeningWorkspaceRepo
		wantStatus    int
		wantKey       string
	}{
		{
			name:       "method not allowed",
			method:     http.MethodGet,
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "invalid body",
			method:     http.MethodPost,
			body:       "{",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing topic_id",
			method:     http.MethodPost,
			body:       apiModels.ListeningWorkRequest{WorkspaceID: "ws1"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "repo not configured",
			method:     http.MethodPost,
			body:       apiModels.ListeningWorkRequest{TopicID: "t1", WorkspaceID: "ws1"},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "topic not found",
			method:     http.MethodPost,
			body:       apiModels.ListeningWorkRequest{TopicID: "t1", WorkspaceID: "ws1"},
			repo:       &mockListeningRepo{topic: nil},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "workspace mismatch",
			method: http.MethodPost,
			body:   apiModels.ListeningWorkRequest{TopicID: "t1", WorkspaceID: "ws1"},
			repo: &mockListeningRepo{
				topic: &mongoModels.ListeningTopic{TopicID: "t1", WorkspaceID: "ws-other"},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:   "topic historical mention count does not block dispatch",
			method: http.MethodPost,
			body:   apiModels.ListeningWorkRequest{TopicID: "t1", WorkspaceID: "ws1"},
			repo: &mockListeningRepo{
				topic: &mongoModels.ListeningTopic{
					TopicID:              "t1",
					WorkspaceID:          "ws1",
					MentionsLimitReached: true,
					MentionsLimit:        1000,
				},
			},
			workspaceRepo: &mockListeningWorkspaceRepo{
				mentionsCount: 100,
				mentionLimit:  1000,
				exists:        true,
			},
			wantStatus: http.StatusAccepted,
			wantKey:    "t1",
		},
		{
			name:       "missing workspace_id",
			method:     http.MethodPost,
			body:       apiModels.ListeningWorkRequest{TopicID: "t1"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "inactive topic blocked",
			method: http.MethodPost,
			body:   apiModels.ListeningWorkRequest{TopicID: "t1", WorkspaceID: "ws1"},
			repo: &mockListeningRepo{
				topic: &mongoModels.ListeningTopic{
					TopicID:     "t1",
					WorkspaceID: "ws1",
					Status:      "paused",
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:   "workspace quota exhausted",
			method: http.MethodPost,
			body:   apiModels.ListeningWorkRequest{TopicID: "t1", WorkspaceID: "ws1"},
			repo: &mockListeningRepo{
				topic: &mongoModels.ListeningTopic{
					TopicID:     "t1",
					WorkspaceID: "ws1",
				},
			},
			workspaceRepo: &mockListeningWorkspaceRepo{
				mentionsCount: 1000,
				mentionLimit:  1000,
				exists:        true,
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:   "workspace repo error fails open",
			method: http.MethodPost,
			body:   apiModels.ListeningWorkRequest{TopicID: "t1", WorkspaceID: "ws1"},
			repo:   &mockListeningRepo{topic: validTopic},
			workspaceRepo: &mockListeningWorkspaceRepo{
				err: fmt.Errorf("quota service unavailable"),
			},
			wantStatus: http.StatusAccepted,
			wantKey:    "t1",
		},
		{
			name:   "no subscription blocks dispatch",
			method: http.MethodPost,
			body:   apiModels.ListeningWorkRequest{TopicID: "t1", WorkspaceID: "ws1", SuperAdminID: "owner-1"},
			repo: &mockListeningRepo{
				topic: &mongoModels.ListeningTopic{
					TopicID:      "t1",
					WorkspaceID:  "ws1",
					SuperAdminID: "owner-1",
				},
			},
			workspaceRepo: &mockListeningWorkspaceRepo{exists: false},
			wantStatus:    http.StatusConflict,
		},
		{
			name:   "success dispatches work order",
			method: http.MethodPost,
			body:   apiModels.ListeningWorkRequest{TopicID: "t1", WorkspaceID: "ws1"},
			repo:   &mockListeningRepo{topic: validTopic},
			workspaceRepo: &mockListeningWorkspaceRepo{
				mentionsCount: 10,
				mentionLimit:  1000,
				exists:        true,
			},
			wantStatus: http.StatusAccepted,
			wantKey:    "t1",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prod := &mockProducer{}
			s := newTestServer(tc.repo, tc.workspaceRepo, prod)

			var reqBody []byte
			switch v := tc.body.(type) {
			case string:
				reqBody = []byte(v)
			case nil:
			default:
				reqBody, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(tc.method, "/api/v1/listening-work", bytes.NewReader(reqBody))
			rec := httptest.NewRecorder()
			s.HandleListeningWork(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d", tc.wantStatus, rec.Code)
			}
			if tc.wantKey != "" {
				if len(prod.produced) == 0 {
					t.Fatal("expected a produced Kafka message")
				}
				if string(prod.produced[0].key) != tc.wantKey {
					t.Errorf("kafka key: want %q, got %q", tc.wantKey, string(prod.produced[0].key))
				}
			}
		})
	}
}
