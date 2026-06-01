package clickhouse

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	chModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

func sampleMentions(n int) []chModels.ListeningMentionRow {
	mentions := make([]chModels.ListeningMentionRow, n)
	now := time.Now()
	for i := range mentions {
		mentions[i] = chModels.ListeningMentionRow{
			MentionID:       "twitter:post-" + string(rune('0'+i)),
			TopicID:         "topic-1",
			Platform:        "twitter",
			NativeID:        "post-" + string(rune('0'+i)),
			ContentHash:     "hash-" + string(rune('0'+i)),
			AuthorID:        "author-1",
			AuthorName:      "Test Author",
			PostText:        "sample post text",
			PostedAt:        now,
			MatchedKeywords: []string{"go"},
			TotalEngagement: 100,
			ContentType:     "post",
			MediaType:       "text",
			URL:             "https://twitter.com/post",
			SentimentLabel:  "pending",
			SentimentScore:  0,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
	}
	return mentions
}

func TestListeningWriteRepository_InsertMentions(t *testing.T) {
	t.Parallel()

	t.Run("empty slice does nothing", func(t *testing.T) {
		t.Parallel()
		repo := NewListeningWriteRepository(&Client{Conn: &mockConn{}}, testLogger())
		err := repo.InsertMentions(context.Background(), nil)
		require.NoError(t, err)
	})

	t.Run("successful insert", func(t *testing.T) {
		t.Parallel()
		batch := &mockBatch{}
		conn := &mockConn{prepareBatchMock: batch}
		repo := NewListeningWriteRepository(&Client{Conn: conn}, testLogger())

		mentions := sampleMentions(3)
		err := repo.InsertMentions(context.Background(), mentions)
		require.NoError(t, err)
		assert.Equal(t, 3, batch.appendCount)
	})

	t.Run("prepare batch error", func(t *testing.T) {
		t.Parallel()
		conn := &mockConn{prepareBatchErr: errors.New("prepare failed")}
		repo := NewListeningWriteRepository(&Client{Conn: conn}, testLogger())

		err := repo.InsertMentions(context.Background(), sampleMentions(1))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "prepare batch failed")
	})

	t.Run("append error", func(t *testing.T) {
		t.Parallel()
		conn := &mockConn{batchAppendErr: errors.New("append failed")}
		repo := NewListeningWriteRepository(&Client{Conn: conn}, testLogger())

		err := repo.InsertMentions(context.Background(), sampleMentions(1))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "append row")
	})

	t.Run("send error", func(t *testing.T) {
		t.Parallel()
		conn := &mockConn{batchSendErr: errors.New("send failed")}
		repo := NewListeningWriteRepository(&Client{Conn: conn}, testLogger())

		err := repo.InsertMentions(context.Background(), sampleMentions(2))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "send batch failed")
	})
}
