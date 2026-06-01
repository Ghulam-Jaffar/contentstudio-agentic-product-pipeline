package mongo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestListeningTopicNormalize_LaravelUsageLimit(t *testing.T) {
	topic := ListeningTopic{
		ID:                   primitive.NewObjectID(),
		MentionsLimitReached: false,
		Usage: ListeningTopicUsage{
			MentionsCount: 10140,
			MentionsLimit: 10000,
		},
		Platforms: FlexPlatforms{"twitter", "reddit"},
		Query: laravelQuery{
			IncludeKeywords: []string{"iran"},
		},
	}

	topic.Normalize()

	assert.Equal(t, topic.ID.Hex(), topic.TopicID)
	assert.Equal(t, 10000, topic.MentionsLimit)
	assert.Equal(t, []string{"iran"}, topic.IncludeKeywords)
	assert.ElementsMatch(t, []string{"twitter", "reddit"}, topic.EnabledPlatforms)
	assert.True(t, topic.MentionsLimitReached)
}

func TestNormalize_LiftsAIContextFromQuery(t *testing.T) {
	topic := ListeningTopic{
		Query: laravelQuery{
			AIContext: AIContext{
				BrandName:     "Acme",
				BrandKeywords: []string{"acme"},
				Industry:      "SaaS",
				Competitors: []AICompetitor{
					{Name: "Foo", Keywords: []string{"foo"}},
				},
			},
			AIContextHint: "Only B2B",
		},
	}

	topic.Normalize()

	assert.Equal(t, "Acme", topic.AIContext.BrandName)
	assert.Equal(t, "Only B2B", topic.AIContextHint)
	assert.Len(t, topic.AIContext.Competitors, 1)
}

func TestNormalize_PreservesTopLevelAIContextWhenQueryEmpty(t *testing.T) {
	topic := ListeningTopic{
		AIContext: AIContext{BrandName: "Already-Set"},
	}
	topic.Normalize()
	assert.Equal(t, "Already-Set", topic.AIContext.BrandName)
}
