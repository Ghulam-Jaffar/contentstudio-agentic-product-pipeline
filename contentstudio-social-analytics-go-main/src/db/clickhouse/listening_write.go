package clickhouse

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	chModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// ListeningWriteRepository handles ClickHouse write operations for listening mentions.
type ListeningWriteRepository struct {
	client *Client
	logger zerolog.Logger
}

// NewListeningWriteRepository creates a new listening write repository.
func NewListeningWriteRepository(client *Client, logger zerolog.Logger) *ListeningWriteRepository {
	return &ListeningWriteRepository{
		client: client,
		logger: logger.With().Str("component", "listening_write_repo").Logger(),
	}
}

// InsertMentions performs a batch insert of listening mentions into ClickHouse.
func (r *ListeningWriteRepository) InsertMentions(ctx context.Context, mentions []chModels.ListeningMentionRow) error {
	if len(mentions) == 0 {
		return nil
	}

	batch, err := r.client.Conn.PrepareBatch(ctx, `
		INSERT INTO listening_mentions (
			mention_id, topic_id, platform, native_id, content_hash,
			author_id, author_name, author_handle, author_image_url, author_url, author_followers,
			post_text, language, posted_at, matched_keywords,
			total_engagement, likes_count, comments_count, shares_count,
			content_type, media_type, url, media_urls, ai_tags,
			sentiment_label, sentiment_score, created_at, updated_at,
			post_read, post_irrelevant, bookmark, sentiment_override
		)`)
	if err != nil {
		return fmt.Errorf("ListeningWriteRepository.InsertMentions: prepare batch failed: %w", err)
	}

	for i := range mentions {
		m := &mentions[i]
		if err := batch.Append(
			m.MentionID,
			m.TopicID,
			m.Platform,
			m.NativeID,
			m.ContentHash,
			m.AuthorID,
			m.AuthorName,
			m.AuthorHandle,
			m.AuthorImageURL,
			m.AuthorURL,
			m.AuthorFollowers,
			m.PostText,
			m.Language,
			m.PostedAt,
			m.MatchedKeywords,
			m.TotalEngagement,
			m.LikesCount,
			m.CommentsCount,
			m.SharesCount,
			m.ContentType,
			m.MediaType,
			m.URL,
			m.MediaURLs,
			m.AITags,
			m.SentimentLabel,
			m.SentimentScore,
			m.CreatedAt,
			m.UpdatedAt,
			m.PostRead,
			m.PostIrrelevant,
			m.Bookmark,
			m.SentimentOverride,
		); err != nil {
			return fmt.Errorf("ListeningWriteRepository.InsertMentions: append row %d failed: %w", i, err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("ListeningWriteRepository.InsertMentions: send batch failed: %w", err)
	}

	r.logger.Info().Int("count", len(mentions)).Msg("Inserted listening mentions batch")
	return nil
}
