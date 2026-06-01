package listening

import (
	"context"
	"testing"

	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type stubViewResolver struct {
	view *mongoModels.ListeningView
	err  error
}

func (s stubViewResolver) GetViewByID(_ context.Context, _ string) (*mongoModels.ListeningView, error) {
	return s.view, s.err
}

type stubTopicWorkspaceResolver struct {
	topics []*mongoModels.ListeningTopic
	err    error
}

func (s stubTopicWorkspaceResolver) ListTopicsByWorkspace(_ context.Context, _ string) ([]*mongoModels.ListeningTopic, error) {
	return s.topics, s.err
}

func TestMentionFilterResolver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		viewResolver   stubViewResolver
		topicResolver  stubTopicWorkspaceResolver
		input          *apiModels.MentionFilter
		wantWorkspace  string
		wantTopicCount int
		wantPlatforms  []string
		wantSentiments []string
		wantAITags     []string
		wantLanguage   []string
		wantFollowers  int
		wantEngagement int
	}{
		{
			name: "applies view preset and expands workspace topics",
			viewResolver: stubViewResolver{
				view: &mongoModels.ListeningView{
					WorkspaceID: "ws-1",
					FilterPreset: mongoModels.ListeningViewFilterPreset{
						Platforms:          []string{"reddit"},
						Sentiments:         []string{"negative"},
						AITags:             []string{"buy_intent"},
						MinFollowers:       5000,
						MinTotalEngagement: 250,
						Language:           []string{"en"},
					},
				},
			},
			topicResolver: stubTopicWorkspaceResolver{
				topics: []*mongoModels.ListeningTopic{
					{TopicID: "topic-a"},
					{ID: primitive.NewObjectID()},
				},
			},
			input:          &apiModels.MentionFilter{ViewID: "view-1"},
			wantWorkspace:  "ws-1",
			wantTopicCount: 2,
			wantPlatforms:  []string{"reddit"},
			wantSentiments: []string{"negative"},
			wantAITags:     []string{"buy_intent"},
			wantLanguage:   []string{"en"},
			wantFollowers:  5000,
			wantEngagement: 250,
		},
		{
			name: "explicit request filters override view preset",
			viewResolver: stubViewResolver{
				view: &mongoModels.ListeningView{
					WorkspaceID: "ws-1",
					FilterPreset: mongoModels.ListeningViewFilterPreset{
						TopicIDs:           []string{"topic-view"},
						Platforms:          []string{"reddit"},
						Sentiments:         []string{"negative"},
						AITags:             []string{"buy_intent"},
						MinFollowers:       5000,
						MinTotalEngagement: 250,
						Language:           []string{"en"},
					},
				},
			},
			topicResolver: stubTopicWorkspaceResolver{},
			input: &apiModels.MentionFilter{
				WorkspaceID:        "ws-1",
				ViewID:             "view-1",
				TopicIDs:           []string{"topic-request"},
				Platforms:          []string{"twitter"},
				Sentiments:         []string{"positive"},
				AITags:             []string{"support"},
				MinFollowers:       100,
				MinTotalEngagement: 15,
				Language:           []string{"fr"},
			},
			wantWorkspace:  "ws-1",
			wantTopicCount: 1,
			wantPlatforms:  []string{"twitter"},
			wantSentiments: []string{"positive"},
			wantAITags:     []string{"support"},
			wantLanguage:   []string{"fr"},
			wantFollowers:  100,
			wantEngagement: 15,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resolver := NewMentionFilterResolver(tc.viewResolver, tc.topicResolver)
			filter, err := resolver.Resolve(context.Background(), tc.input)
			if err != nil {
				t.Fatalf("Resolve returned error: %v", err)
			}

			if filter.WorkspaceID != tc.wantWorkspace {
				t.Errorf("workspace_id: want %q, got %q", tc.wantWorkspace, filter.WorkspaceID)
			}
			if len(filter.TopicIDs) != tc.wantTopicCount {
				t.Errorf("topic count: want %d, got %d (%#v)", tc.wantTopicCount, len(filter.TopicIDs), filter.TopicIDs)
			}
			assertStringSlice(t, "platforms", filter.Platforms, tc.wantPlatforms)
			assertStringSlice(t, "sentiments", filter.Sentiments, tc.wantSentiments)
			assertStringSlice(t, "ai_tags", filter.AITags, tc.wantAITags)
			assertStringSlice(t, "language", filter.Language, tc.wantLanguage)
			if filter.MinFollowers != tc.wantFollowers {
				t.Errorf("min_followers: want %d, got %d", tc.wantFollowers, filter.MinFollowers)
			}
			if filter.MinTotalEngagement != tc.wantEngagement {
				t.Errorf("min_total_engagement: want %d, got %d", tc.wantEngagement, filter.MinTotalEngagement)
			}
		})
	}
}

func assertStringSlice(t *testing.T, field string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: want %v, got %v", field, want, got)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s[%d]: want %q, got %q", field, i, want[i], got[i])
		}
	}
}
