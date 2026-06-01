package views

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type stubMentionCounter struct {
	count     int
	err       error
	callCount int
	filters   []*clickhouse.MentionFilter
}

func (s *stubMentionCounter) CountMentions(_ context.Context, filter *clickhouse.MentionFilter) (int, error) {
	s.callCount++
	if filter != nil {
		cloned := *filter
		cloned.TopicIDs = append([]string(nil), filter.TopicIDs...)
		cloned.Platforms = append([]string(nil), filter.Platforms...)
		cloned.Sentiments = append([]string(nil), filter.Sentiments...)
		cloned.AITags = append([]string(nil), filter.AITags...)
		cloned.ExcludeAITags = append([]string(nil), filter.ExcludeAITags...)
		cloned.Language = append([]string(nil), filter.Language...)
		s.filters = append(s.filters, &cloned)
	}

	if s.err != nil {
		return 0, s.err
	}

	return s.count, nil
}

type stubTopicResolver struct {
	topics []*mongoModels.ListeningTopic
	err    error
}

func (s *stubTopicResolver) ListTopicsByWorkspace(_ context.Context, _ string) ([]*mongoModels.ListeningTopic, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.topics, nil
}

func TestService_listWorkspaceTopicIDs(t *testing.T) {
	t.Parallel()

	fallbackID := primitive.NewObjectID()

	tests := []struct {
		name     string
		resolver topicWorkspaceResolver
		want     []string
		wantErr  string
	}{
		{
			name:     "returns nil when topic resolver is not configured",
			resolver: nil,
			want:     nil,
		},
		{
			name: "uses topic_id when present and falls back to object id when missing",
			resolver: &stubTopicResolver{
				topics: []*mongoModels.ListeningTopic{
					{TopicID: "topic-1"},
					nil,
					{ID: fallbackID},
				},
			},
			want: []string{"topic-1", fallbackID.Hex()},
		},
		{
			name: "propagates resolver errors",
			resolver: &stubTopicResolver{
				err: errors.New("mongo unavailable"),
			},
			wantErr: "mongo unavailable",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := &Service{topics: tc.resolver}

			got, err := service.listWorkspaceTopicIDs(context.Background(), "ws-1")
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if err.Error() != tc.wantErr {
					t.Fatalf("expected error %q, got %q", tc.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestService_countForView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		counter           *stubMentionCounter
		view              mongoModels.ListeningView
		workspaceTopicIDs []string
		want              int
		wantCalls         int
		wantFilter        *clickhouse.MentionFilter
		wantErr           string
	}{
		{
			name: "returns zero when counter is not configured",
			view: mongoModels.ListeningView{
				FilterPreset: mongoModels.ListeningViewFilterPreset{
					TopicIDs: []string{"topic-1"},
				},
			},
			want:      0,
			wantCalls: 0,
		},
		{
			name: "returns zero without querying when no topics are available",
			counter: &stubMentionCounter{
				count: 99,
			},
			view:      mongoModels.ListeningView{},
			want:      0,
			wantCalls: 0,
		},
		{
			name: "intersects view topic ids with workspace and forwards filter fields",
			counter: &stubMentionCounter{
				count: 21,
			},
			view: mongoModels.ListeningView{
				FilterPreset: mongoModels.ListeningViewFilterPreset{
					TopicIDs:           []string{"topic-a", "topic-b"},
					Platforms:          []string{"twitter"},
					Sentiments:         []string{"negative"},
					AITags:             []string{"Own Brand Mention"},
					ExcludeAITags:      []string{"Buy Intent"},
					Language:           []string{"en"},
					MinFollowers:       100,
					MinTotalEngagement: 250,
				},
			},
			// Only topic-a is in this workspace; topic-b must be dropped to
			// prevent cross-tenant counting.
			workspaceTopicIDs: []string{"topic-a", "topic-c"},
			want:              21,
			wantCalls:         1,
			wantFilter: &clickhouse.MentionFilter{
				TopicIDs:           []string{"topic-a"},
				Platforms:          []string{"twitter"},
				Sentiments:         []string{"negative"},
				AITags:             []string{"Own Brand Mention"},
				ExcludeAITags:      []string{"Buy Intent"},
				Language:           []string{"en"},
				MinFollowers:       100,
				MinTotalEngagement: 250,
			},
		},
		{
			name: "returns zero when view topic ids are all outside the workspace",
			counter: &stubMentionCounter{
				count: 42,
			},
			view: mongoModels.ListeningView{
				FilterPreset: mongoModels.ListeningViewFilterPreset{
					TopicIDs: []string{"foreign-topic"},
				},
			},
			workspaceTopicIDs: []string{"topic-a"},
			want:              0,
			wantCalls:         0,
		},
		{
			name: "passes view topic ids through when workspace lookup was unavailable",
			counter: &stubMentionCounter{
				count: 7,
			},
			view: mongoModels.ListeningView{
				FilterPreset: mongoModels.ListeningViewFilterPreset{
					TopicIDs: []string{"view-topic"},
				},
			},
			// workspaceTopicIDs=nil simulates the upstream ListViews fallback
			// where topic lookup errored; we still render counts using whatever
			// the saved view requested so the UI isn't empty.
			workspaceTopicIDs: nil,
			want:              7,
			wantCalls:         1,
			wantFilter: &clickhouse.MentionFilter{
				TopicIDs: []string{"view-topic"},
			},
		},
		{
			name: "falls back to workspace topic ids when the view has none",
			counter: &stubMentionCounter{
				count: 5,
			},
			view: mongoModels.ListeningView{
				FilterPreset: mongoModels.ListeningViewFilterPreset{
					Platforms: []string{"reddit"},
				},
			},
			workspaceTopicIDs: []string{"topic-1", "topic-2"},
			want:              5,
			wantCalls:         1,
			wantFilter: &clickhouse.MentionFilter{
				TopicIDs:  []string{"topic-1", "topic-2"},
				Platforms: []string{"reddit"},
			},
		},
		{
			name: "propagates counter errors",
			counter: &stubMentionCounter{
				err: errors.New("clickhouse down"),
			},
			view: mongoModels.ListeningView{
				FilterPreset: mongoModels.ListeningViewFilterPreset{
					TopicIDs: []string{"topic-1"},
				},
			},
			workspaceTopicIDs: []string{"topic-1"},
			wantCalls:         1,
			wantErr:           "clickhouse down",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := &Service{}
			if tc.counter != nil {
				service.mentionCounter = tc.counter
			}

			got, err := service.countForView(context.Background(), tc.view, tc.workspaceTopicIDs)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if err.Error() != tc.wantErr {
					t.Fatalf("expected error %q, got %q", tc.wantErr, err.Error())
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Fatalf("want count %d, got %d", tc.want, got)
			}

			actualCalls := 0
			if tc.counter != nil {
				actualCalls = tc.counter.callCount
			}
			if actualCalls != tc.wantCalls {
				t.Fatalf("want %d CountMentions calls, got %d", tc.wantCalls, actualCalls)
			}

			if tc.wantFilter != nil {
				if len(tc.counter.filters) != 1 {
					t.Fatalf("expected one captured filter, got %d", len(tc.counter.filters))
				}
				if !reflect.DeepEqual(tc.counter.filters[0], tc.wantFilter) {
					t.Fatalf("want filter %+v, got %+v", tc.wantFilter, tc.counter.filters[0])
				}
			}
		})
	}
}

func TestToViewResponse_SystemKey(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		view          mongoModels.ListeningView
		count         int
		wantSystemKey string
	}{
		{
			name: "maps empty system key for user views",
			view: mongoModels.ListeningView{
				ID:          primitive.NewObjectID(),
				WorkspaceID: "ws-1",
				Name:        "My View",
				Icon:        "Star",
				Type:        "user",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			count:         3,
			wantSystemKey: "",
		},
		{
			name: "maps system key for system views",
			view: mongoModels.ListeningView{
				ID:          primitive.NewObjectID(),
				WorkspaceID: "ws-1",
				Name:        "All Mentions",
				Icon:        "Inbox",
				Type:        "system",
				SystemKey:   "all_mentions",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			count:         9,
			wantSystemKey: "all_mentions",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := toViewResponse(tc.view, tc.count)
			if got.SystemKey != tc.wantSystemKey {
				t.Fatalf("expected system_key %q, got %q", tc.wantSystemKey, got.SystemKey)
			}
			if got.Count != tc.count {
				t.Fatalf("expected count %d, got %d", tc.count, got.Count)
			}
			if got.Type != tc.view.Type {
				t.Fatalf("expected type %q, got %q", tc.view.Type, got.Type)
			}
		})
	}
}
