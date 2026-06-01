package clickhouse

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"

	chModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

type MentionFilter struct {
	TopicIDs           []string
	Platforms          []string
	Sentiments         []string
	AITags             []string
	ExcludeAITags      []string
	Language           []string
	MinFollowers       int
	MinTotalEngagement int
	DateFrom           time.Time
	DateTo             time.Time
	Sort               string
	Cursor             string
	Limit              int
	IsBookmarked       *bool
	IsRead             *bool
	IncludeIrrelevant  bool
	Search             string
}

type MentionCursor struct {
	PostedAt        time.Time `json:"p"`
	MentionID       string    `json:"m"`
	TotalEngagement int64     `json:"e,omitempty"`
}

type ListeningReadRepository struct {
	client *Client
	logger zerolog.Logger
}

type AnalyticsData struct {
	TotalMentions    int
	SentimentCounts  map[string]int
	MentionsOverTime []MentionPoint
	SentimentTrend   []SentimentPoint
	PlatformCounts   map[string]int
	TagCounts        map[string]int
}

type MentionPoint struct {
	Date    time.Time
	TopicID string
	Count   int
}

type SentimentPoint struct {
	Date      time.Time
	Sentiment string
	Count     int
}

func missingEnrichmentWhereClause() string {
	return "(sentiment_label = '' OR length(ai_tags) = 0)"
}

func NewListeningReadRepository(client *Client, logger zerolog.Logger) *ListeningReadRepository {
	return &ListeningReadRepository{
		client: client,
		logger: logger.With().Str("component", "listening_read_repo").Logger(),
	}
}

// ListMentionsMissingEnrichment returns the latest mention rows that still have
// incomplete AI enrichment. Rows qualify when the sentiment label is empty or
// the normalized AI tag array is empty. We intentionally do not treat a zero
// sentiment_score as missing because 0.0 is also a valid Float64 value and
// would cause valid low-confidence rows to be reprocessed indefinitely.
func (r *ListeningReadRepository) ListMentionsMissingEnrichment(
	ctx context.Context,
	updatedAfter time.Time,
	updatedBefore time.Time,
	limit int,
) ([]chModels.ListeningMentionRow, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	query := `
SELECT
    mention_id, topic_id, platform, native_id, content_hash,
    author_id, author_name, author_handle, author_image_url, author_url, author_followers,
    post_text, language, posted_at, matched_keywords,
    total_engagement, likes_count, comments_count, shares_count,
    content_type, media_type, url, media_urls, ai_tags,
    sentiment_label, sentiment_score, created_at, updated_at,
    post_read, post_irrelevant, bookmark, sentiment_override
FROM listening_mentions FINAL
WHERE post_text != ''
  AND ` + missingEnrichmentWhereClause() + `
  AND updated_at >= ?
  AND updated_at < ?
ORDER BY updated_at ASC
LIMIT ?`

	rows, err := r.client.Conn.Query(ctx, query, updatedAfter, updatedBefore, limit)
	if err != nil {
		return nil, fmt.Errorf("ListeningReadRepository.ListMentionsMissingEnrichment: %w", err)
	}
	defer rows.Close()

	results := make([]chModels.ListeningMentionRow, 0, limit)
	for rows.Next() {
		var row chModels.ListeningMentionRow
		if err := rows.Scan(
			&row.MentionID, &row.TopicID, &row.Platform, &row.NativeID, &row.ContentHash,
			&row.AuthorID, &row.AuthorName, &row.AuthorHandle, &row.AuthorImageURL, &row.AuthorURL, &row.AuthorFollowers,
			&row.PostText, &row.Language, &row.PostedAt, &row.MatchedKeywords,
			&row.TotalEngagement, &row.LikesCount, &row.CommentsCount, &row.SharesCount,
			&row.ContentType, &row.MediaType, &row.URL, &row.MediaURLs, &row.AITags,
			&row.SentimentLabel, &row.SentimentScore, &row.CreatedAt, &row.UpdatedAt,
			&row.PostRead, &row.PostIrrelevant, &row.Bookmark, &row.SentimentOverride,
		); err != nil {
			return nil, fmt.Errorf("ListeningReadRepository.ListMentionsMissingEnrichment: scan: %w", err)
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListeningReadRepository.ListMentionsMissingEnrichment: rows: %w", err)
	}

	return results, nil
}

func (r *ListeningReadRepository) QueryMentions(ctx context.Context, filter *MentionFilter) ([]chModels.ListeningMentionRow, string, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	where, args := r.buildWhereClause(filter)
	orderBy := r.buildOrderBy(filter.Sort)
	cursorWhere, cursorArgs := r.buildCursorClause(filter)
	if cursorWhere != "" {
		where += " AND " + cursorWhere
		args = append(args, cursorArgs...)
	}

	query := fmt.Sprintf(`
SELECT
    mention_id, topic_id, platform, native_id, content_hash,
    author_id, author_name, author_handle, author_image_url, author_url, author_followers,
    post_text, language, posted_at, matched_keywords,
    total_engagement, likes_count, comments_count, shares_count,
    content_type, media_type, url, media_urls, ai_tags,
    sentiment_label, sentiment_score, created_at, updated_at,
    post_read, post_irrelevant, bookmark, sentiment_override
FROM listening_mentions FINAL
WHERE %s
ORDER BY %s
LIMIT %d
`, where, orderBy, limit+1)

	rows, err := r.client.Conn.Query(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("ListeningReadRepository.QueryMentions: %w", err)
	}
	defer rows.Close()

	var results []chModels.ListeningMentionRow
	for rows.Next() {
		var row chModels.ListeningMentionRow
		if err := rows.Scan(
			&row.MentionID, &row.TopicID, &row.Platform, &row.NativeID, &row.ContentHash,
			&row.AuthorID, &row.AuthorName, &row.AuthorHandle, &row.AuthorImageURL, &row.AuthorURL, &row.AuthorFollowers,
			&row.PostText, &row.Language, &row.PostedAt, &row.MatchedKeywords,
			&row.TotalEngagement, &row.LikesCount, &row.CommentsCount, &row.SharesCount,
			&row.ContentType, &row.MediaType, &row.URL, &row.MediaURLs, &row.AITags,
			&row.SentimentLabel, &row.SentimentScore, &row.CreatedAt, &row.UpdatedAt,
			&row.PostRead, &row.PostIrrelevant, &row.Bookmark, &row.SentimentOverride,
		); err != nil {
			return nil, "", fmt.Errorf("ListeningReadRepository.QueryMentions: scan: %w", err)
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("ListeningReadRepository.QueryMentions: rows: %w", err)
	}

	var nextCursor string
	if len(results) > limit {
		results = results[:limit]
		last := results[limit-1]
		nextCursor = encodeCursor(MentionCursor{
			PostedAt:        last.PostedAt,
			MentionID:       last.MentionID,
			TotalEngagement: last.TotalEngagement,
		})
	}

	return results, nextCursor, nil
}

func (r *ListeningReadRepository) GetAnalytics(ctx context.Context, filter *MentionFilter) (*AnalyticsData, error) {
	where, args := r.buildWhereClause(filter)

	data := &AnalyticsData{
		SentimentCounts: make(map[string]int),
		PlatformCounts:  make(map[string]int),
		TagCounts:       make(map[string]int),
	}

	// 1. Sentiment & Total Mentions
	sentimentQuery := fmt.Sprintf(`
		SELECT sentiment_label, count()
		FROM listening_mentions FINAL
		WHERE %s
		GROUP BY sentiment_label
	`, where)

	rows, err := r.client.Conn.Query(ctx, sentimentQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("ListeningReadRepository.GetAnalytics: sentiment: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var label string
		var count uint64
		if err := rows.Scan(&label, &count); err != nil {
			return nil, err
		}
		if label == "" {
			label = "neutral"
		}
		data.SentimentCounts[label] = int(count)
		data.TotalMentions += int(count)
	}

	// 2. Mentions Over Time
	mentionsTimeSeriesQuery := fmt.Sprintf(`
		SELECT toStartOfDay(posted_at) as day, topic_id, count()
		FROM listening_mentions FINAL
		WHERE %s
		GROUP BY day, topic_id
		ORDER BY day ASC
	`, where)

	rowsMT, err := r.client.Conn.Query(ctx, mentionsTimeSeriesQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("ListeningReadRepository.GetAnalytics: mentions_over_time: %w", err)
	}
	defer rowsMT.Close()

	for rowsMT.Next() {
		var day time.Time
		var topicID string
		var count uint64
		if err := rowsMT.Scan(&day, &topicID, &count); err != nil {
			return nil, err
		}
		data.MentionsOverTime = append(data.MentionsOverTime, MentionPoint{
			Date:    day,
			TopicID: topicID,
			Count:   int(count),
		})
	}

	// 3. Sentiment Trend
	sentimentTrendQuery := fmt.Sprintf(`
		SELECT toStartOfDay(posted_at) as day, sentiment_label, count()
		FROM listening_mentions FINAL
		WHERE %s
		GROUP BY day, sentiment_label
		ORDER BY day ASC
	`, where)

	rowsST, err := r.client.Conn.Query(ctx, sentimentTrendQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("ListeningReadRepository.GetAnalytics: sentiment_trend: %w", err)
	}
	defer rowsST.Close()

	for rowsST.Next() {
		var day time.Time
		var sentiment string
		var count uint64
		if err := rowsST.Scan(&day, &sentiment, &count); err != nil {
			return nil, err
		}
		if sentiment == "" {
			sentiment = "neutral"
		}
		data.SentimentTrend = append(data.SentimentTrend, SentimentPoint{
			Date:      day,
			Sentiment: sentiment,
			Count:     int(count),
		})
	}

	// 4. Platform Distribution
	platformQuery := fmt.Sprintf(`
		SELECT platform, count()
		FROM listening_mentions FINAL
		WHERE %s
		GROUP BY platform
	`, where)

	rowsP, err := r.client.Conn.Query(ctx, platformQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("ListeningReadRepository.GetAnalytics: platform: %w", err)
	}
	defer rowsP.Close()

	for rowsP.Next() {
		var platform string
		var count uint64
		if err := rowsP.Scan(&platform, &count); err != nil {
			return nil, err
		}
		data.PlatformCounts[platform] = int(count)
	}

	// 5. Tag Distribution
	tagQuery := fmt.Sprintf(`
		SELECT tag, count()
		FROM (
			SELECT arrayJoin(ai_tags) as tag
			FROM listening_mentions FINAL
			WHERE %s
		)
		GROUP BY tag
		ORDER BY count() DESC
		LIMIT 15
	`, where)

	rowsT, err := r.client.Conn.Query(ctx, tagQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("ListeningReadRepository.GetAnalytics: tags: %w", err)
	}
	defer rowsT.Close()

	for rowsT.Next() {
		var tag string
		var count uint64
		if err := rowsT.Scan(&tag, &count); err != nil {
			return nil, err
		}
		data.TagCounts[tag] = int(count)
	}

	return data, nil
}

func (r *ListeningReadRepository) CountUnread(ctx context.Context, filter *MentionFilter) (int, error) {
	readFalse := false
	unreadFilter := *filter
	unreadFilter.IsRead = &readFalse

	where, args := r.buildWhereClause(&unreadFilter)
	query := fmt.Sprintf(`SELECT count() FROM listening_mentions FINAL WHERE %s`, where)

	var count uint64
	if err := r.client.Conn.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("ListeningReadRepository.CountUnread: %w", err)
	}
	return int(count), nil
}

func (r *ListeningReadRepository) CountMentions(ctx context.Context, filter *MentionFilter) (int, error) {
	effectiveFilter := filter
	if effectiveFilter == nil {
		effectiveFilter = &MentionFilter{}
	}

	where, args := r.buildWhereClause(effectiveFilter)
	query := fmt.Sprintf(`SELECT count() FROM listening_mentions FINAL WHERE %s`, where)

	var count uint64
	if err := r.client.Conn.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("ListeningReadRepository.CountMentions: %w", err)
	}
	return int(count), nil
}

func (r *ListeningReadRepository) GetMention(ctx context.Context, mentionID, topicID string) (*chModels.ListeningMentionRow, error) {
	query := `
SELECT
    mention_id, topic_id, platform, native_id, content_hash,
    author_id, author_name, author_handle, author_image_url, author_url, author_followers,
    post_text, language, posted_at, matched_keywords,
    total_engagement, likes_count, comments_count, shares_count,
    content_type, media_type, url, media_urls, ai_tags,
    sentiment_label, sentiment_score, created_at, updated_at,
    post_read, post_irrelevant, bookmark, sentiment_override
FROM listening_mentions FINAL
WHERE mention_id = ?`

	var args []interface{}
	args = append(args, mentionID)

	if topicID != "" {
		query += " AND topic_id = ?"
		args = append(args, topicID)
	}

	query += " LIMIT 1"

	var row chModels.ListeningMentionRow
	if err := r.client.Conn.QueryRow(ctx, query, args...).Scan(
		&row.MentionID, &row.TopicID, &row.Platform, &row.NativeID, &row.ContentHash,
		&row.AuthorID, &row.AuthorName, &row.AuthorHandle, &row.AuthorImageURL, &row.AuthorURL, &row.AuthorFollowers,
		&row.PostText, &row.Language, &row.PostedAt, &row.MatchedKeywords,
		&row.TotalEngagement, &row.LikesCount, &row.CommentsCount, &row.SharesCount,
		&row.ContentType, &row.MediaType, &row.URL, &row.MediaURLs, &row.AITags,
		&row.SentimentLabel, &row.SentimentScore, &row.CreatedAt, &row.UpdatedAt,
		&row.PostRead, &row.PostIrrelevant, &row.Bookmark, &row.SentimentOverride,
	); err != nil {
		return nil, fmt.Errorf("ListeningReadRepository.GetMention: %w", err)
	}
	return &row, nil
}

func (r *ListeningReadRepository) UpdateMention(ctx context.Context, existing chModels.ListeningMentionRow) error {
	existing.UpdatedAt = time.Now().UTC()

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
		return fmt.Errorf("ListeningReadRepository.UpdateMention: prepare: %w", err)
	}

	if err := batch.Append(
		existing.MentionID, existing.TopicID, existing.Platform, existing.NativeID, existing.ContentHash,
		existing.AuthorID, existing.AuthorName, existing.AuthorHandle, existing.AuthorImageURL, existing.AuthorURL, existing.AuthorFollowers,
		existing.PostText, existing.Language, existing.PostedAt, existing.MatchedKeywords,
		existing.TotalEngagement, existing.LikesCount, existing.CommentsCount, existing.SharesCount,
		existing.ContentType, existing.MediaType, existing.URL, existing.MediaURLs, existing.AITags,
		existing.SentimentLabel, existing.SentimentScore, existing.CreatedAt, existing.UpdatedAt,
		existing.PostRead, existing.PostIrrelevant, existing.Bookmark, existing.SentimentOverride,
	); err != nil {
		return fmt.Errorf("ListeningReadRepository.UpdateMention: append: %w", err)
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("ListeningReadRepository.UpdateMention: send: %w", err)
	}
	return nil
}

// MarkAllRead reinserts full rows because listening_mentions uses
// ReplacingMergeTree(updated_at): mutating the read flag means appending a new
// record version rather than patching a single column in place.
func (r *ListeningReadRepository) MarkAllRead(ctx context.Context, filter *MentionFilter) (int, error) {
	readFalse := false
	readFilter := *filter
	readFilter.IsRead = &readFalse

	where, args := r.buildWhereClause(&readFilter)
	selectQuery := fmt.Sprintf(`
SELECT
    mention_id, topic_id, platform, native_id, content_hash,
    author_id, author_name, author_handle, author_image_url, author_url, author_followers,
    post_text, language, posted_at, matched_keywords,
    total_engagement, likes_count, comments_count, shares_count,
    content_type, media_type, url, media_urls, ai_tags,
    sentiment_label, sentiment_score, created_at, updated_at,
    post_read, post_irrelevant, bookmark, sentiment_override
FROM listening_mentions FINAL
WHERE %s`, where)

	rows, err := r.client.Conn.Query(ctx, selectQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("ListeningReadRepository.MarkAllRead: query: %w", err)
	}
	defer rows.Close()

	now := time.Now().UTC()
	var mentions []chModels.ListeningMentionRow
	for rows.Next() {
		var row chModels.ListeningMentionRow
		if err := rows.Scan(
			&row.MentionID, &row.TopicID, &row.Platform, &row.NativeID, &row.ContentHash,
			&row.AuthorID, &row.AuthorName, &row.AuthorHandle, &row.AuthorImageURL, &row.AuthorURL, &row.AuthorFollowers,
			&row.PostText, &row.Language, &row.PostedAt, &row.MatchedKeywords,
			&row.TotalEngagement, &row.LikesCount, &row.CommentsCount, &row.SharesCount,
			&row.ContentType, &row.MediaType, &row.URL, &row.MediaURLs, &row.AITags,
			&row.SentimentLabel, &row.SentimentScore, &row.CreatedAt, &row.UpdatedAt,
			&row.PostRead, &row.PostIrrelevant, &row.Bookmark, &row.SentimentOverride,
		); err != nil {
			return 0, fmt.Errorf("ListeningReadRepository.MarkAllRead: scan: %w", err)
		}
		row.PostRead = true
		row.UpdatedAt = now
		mentions = append(mentions, row)
	}

	if len(mentions) == 0 {
		return 0, nil
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
		return 0, fmt.Errorf("ListeningReadRepository.MarkAllRead: prepare: %w", err)
	}

	for i := range mentions {
		m := &mentions[i]
		if err := batch.Append(
			m.MentionID, m.TopicID, m.Platform, m.NativeID, m.ContentHash,
			m.AuthorID, m.AuthorName, m.AuthorHandle, m.AuthorImageURL, m.AuthorURL, m.AuthorFollowers,
			m.PostText, m.Language, m.PostedAt, m.MatchedKeywords,
			m.TotalEngagement, m.LikesCount, m.CommentsCount, m.SharesCount,
			m.ContentType, m.MediaType, m.URL, m.MediaURLs, m.AITags,
			m.SentimentLabel, m.SentimentScore, m.CreatedAt, m.UpdatedAt,
			m.PostRead, m.PostIrrelevant, m.Bookmark, m.SentimentOverride,
		); err != nil {
			return 0, fmt.Errorf("ListeningReadRepository.MarkAllRead: append: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return 0, fmt.Errorf("ListeningReadRepository.MarkAllRead: send: %w", err)
	}

	return len(mentions), nil
}

func (r *ListeningReadRepository) buildWhereClause(filter *MentionFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if len(filter.TopicIDs) > 0 {
		conditions = append(conditions, "topic_id IN ("+placeholders(len(filter.TopicIDs))+")")
		for _, id := range filter.TopicIDs {
			args = append(args, id)
		}
	}

	if len(filter.Platforms) > 0 {
		conditions = append(conditions, "platform IN ("+placeholders(len(filter.Platforms))+")")
		for _, p := range filter.Platforms {
			args = append(args, p)
		}
	}

	if len(filter.Sentiments) > 0 {
		conditions = append(conditions, "sentiment_label IN ("+placeholders(len(filter.Sentiments))+")")
		for _, s := range filter.Sentiments {
			args = append(args, s)
		}
	}

	if normalizedTags := normalizeTags(filter.AITags); len(normalizedTags) > 0 {
		tagConditions := make([]string, 0, len(normalizedTags))
		for _, tag := range normalizedTags {
			tagConditions = append(tagConditions, "arrayExists(x -> lower(x) = ?, ai_tags)")
			args = append(args, tag)
		}
		conditions = append(conditions, "("+strings.Join(tagConditions, " OR ")+")")
	}

	if normalizedExcluded := normalizeTags(filter.ExcludeAITags); len(normalizedExcluded) > 0 {
		for _, tag := range normalizedExcluded {
			conditions = append(conditions, "NOT arrayExists(x -> lower(x) = ?, ai_tags)")
			args = append(args, tag)
		}
	}

	if normalizedLanguages := normalizeLanguages(filter.Language); len(normalizedLanguages) > 0 {
		conditions = append(conditions, "language IN ("+placeholders(len(normalizedLanguages))+")")
		for _, language := range normalizedLanguages {
			args = append(args, language)
		}
	}

	if filter.MinFollowers > 0 {
		conditions = append(conditions, "author_followers >= ?")
		args = append(args, filter.MinFollowers)
	}

	if filter.MinTotalEngagement > 0 {
		conditions = append(conditions, "total_engagement >= ?")
		args = append(args, filter.MinTotalEngagement)
	}

	if !filter.DateFrom.IsZero() {
		conditions = append(conditions, "posted_at >= ?")
		args = append(args, filter.DateFrom)
	}

	if !filter.DateTo.IsZero() {
		conditions = append(conditions, "posted_at <= ?")
		args = append(args, filter.DateTo)
	}

	if filter.IsBookmarked != nil {
		conditions = append(conditions, "bookmark = ?")
		args = append(args, *filter.IsBookmarked)
	}

	if filter.IsRead != nil {
		conditions = append(conditions, "post_read = ?")
		args = append(args, *filter.IsRead)
	}

	if !filter.IncludeIrrelevant {
		// Hide AI-tagged Irrelevant mentions and user-marked irrelevant mentions
		// by default. ?include_irrelevant=1 reveals both. Both signals are
		// orthogonal: AI signal lives in ai_tags, user signal in post_irrelevant.
		conditions = append(conditions, "NOT arrayExists(x -> x = 'Irrelevant', ai_tags)")
		conditions = append(conditions, "post_irrelevant = false")
	}

	if filter.Search != "" {
		conditions = append(conditions, "positionCaseInsensitive(post_text, ?) > 0")
		args = append(args, filter.Search)
	}

	if len(conditions) == 0 {
		return "1 = 1", args
	}
	return strings.Join(conditions, " AND "), args
}

func (r *ListeningReadRepository) buildOrderBy(sort string) string {
	switch sort {
	case "oldest":
		return "posted_at ASC, mention_id ASC"
	case "most_engaged":
		return "total_engagement DESC, posted_at DESC, mention_id DESC"
	default:
		return "posted_at DESC, mention_id DESC"
	}
}

func (r *ListeningReadRepository) buildCursorClause(filter *MentionFilter) (string, []interface{}) {
	if filter.Cursor == "" {
		return "", nil
	}

	cursor, err := decodeCursor(filter.Cursor)
	if err != nil {
		r.logger.Warn().Err(err).Msg("Invalid cursor, ignoring")
		return "", nil
	}

	switch filter.Sort {
	case "oldest":
		return "(posted_at, mention_id) > (?, ?)", []interface{}{cursor.PostedAt, cursor.MentionID}
	case "most_engaged":
		return "(total_engagement, posted_at, mention_id) < (?, ?, ?)", []interface{}{
			cursor.TotalEngagement,
			cursor.PostedAt,
			cursor.MentionID,
		}
	default:
		return "(posted_at, mention_id) < (?, ?)", []interface{}{cursor.PostedAt, cursor.MentionID}
	}
}

func encodeCursor(c MentionCursor) string {
	data, _ := json.Marshal(c)
	return base64.URLEncoding.EncodeToString(data)
}

func decodeCursor(s string) (MentionCursor, error) {
	data, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return MentionCursor{}, fmt.Errorf("decodeCursor: %w", err)
	}
	var c MentionCursor
	if err := json.Unmarshal(data, &c); err != nil {
		return MentionCursor{}, fmt.Errorf("decodeCursor: %w", err)
	}
	return c, nil
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

func normalizeTags(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))

	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(value, "#")))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	return result
}

func normalizeLanguages(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))

	for _, value := range values {
		normalized := normalizeLanguage(value)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}

	return result
}

func normalizeLanguage(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return ""
	}

	switch normalized {
	case "english":
		return "en"
	case "spanish":
		return "es"
	case "french":
		return "fr"
	case "german":
		return "de"
	case "italian":
		return "it"
	case "portuguese":
		return "pt"
	case "dutch":
		return "nl"
	case "arabic":
		return "ar"
	case "turkish":
		return "tr"
	case "hindi":
		return "hi"
	case "urdu":
		return "ur"
	case "indonesian":
		return "id"
	case "malay":
		return "ms"
	case "japanese":
		return "ja"
	case "korean":
		return "ko"
	case "chinese":
		return "zh"
	default:
		return normalized
	}
}
