// Package parser normalises provider-specific raw mention payloads into the
// canonical kafkamodels.ParsedMention shape before downstream enrichment.
// Content is truncated to maxTextBytes and hashed so dedup in the sink stage
// can rely on a stable key across retries and re-ingestion.
package parser

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const maxTextBytes = 10 * 1024 // 10KB

type flexibleString string

func (s *flexibleString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = ""
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = flexibleString(str)
		return nil
	}

	var number json.Number
	if err := json.Unmarshal(data, &number); err == nil {
		*s = flexibleString(number.String())
		return nil
	}

	var boolean bool
	if err := json.Unmarshal(data, &boolean); err == nil {
		*s = flexibleString(strconv.FormatBool(boolean))
		return nil
	}

	return fmt.Errorf("unsupported flexibleString payload: %s", string(data))
}

func (s flexibleString) String() string {
	return string(s)
}

type flexibleInt64 int64

func (n *flexibleInt64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*n = 0
		return nil
	}

	var number int64
	if err := json.Unmarshal(data, &number); err == nil {
		*n = flexibleInt64(number)
		return nil
	}

	var floatNumber float64
	if err := json.Unmarshal(data, &floatNumber); err == nil {
		*n = flexibleInt64(int64(floatNumber))
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		str = strings.TrimSpace(str)
		if str == "" {
			*n = 0
			return nil
		}
		parsed, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("unsupported flexibleInt64 payload: %s", string(data))
		}
		*n = flexibleInt64(parsed)
		return nil
	}

	return fmt.Errorf("unsupported flexibleInt64 payload: %s", string(data))
}

func (n flexibleInt64) Int64() int64 {
	return int64(n)
}

type flexibleStringList []string

func (l *flexibleStringList) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*l = nil
		return nil
	}

	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		if single == "" {
			*l = nil
			return nil
		}
		*l = []string{single}
		return nil
	}

	var many []string
	if err := json.Unmarshal(data, &many); err == nil {
		*l = compactStrings(many)
		return nil
	}

	var rawMany []json.RawMessage
	if err := json.Unmarshal(data, &rawMany); err == nil {
		values := make([]string, 0, len(rawMany))
		for _, item := range rawMany {
			var fs flexibleString
			if err := json.Unmarshal(item, &fs); err == nil {
				if fs.String() != "" {
					values = append(values, fs.String())
				}
				continue
			}

			var object map[string]json.RawMessage
			if err := json.Unmarshal(item, &object); err != nil {
				return err
			}

			for _, key := range []string{"tag", "text", "name", "value", "hashtag"} {
				if rawValue, ok := object[key]; ok {
					if err := json.Unmarshal(rawValue, &fs); err == nil && fs.String() != "" {
						values = append(values, fs.String())
						break
					}
				}
			}
		}
		*l = values
		return nil
	}

	return fmt.Errorf("unsupported flexibleStringList payload: %s", string(data))
}

type rawAuthor struct {
	Handle         flexibleString `json:"handle"`
	URL            flexibleString `json:"url"`
	FollowersCount flexibleInt64  `json:"followers_count"`
	Region         flexibleString `json:"region"`
	Country        flexibleString `json:"country"`
	CountryCode    flexibleString `json:"country_code"`
}

// DedupChecker abstracts Redis dedup operations for testability.
type DedupChecker interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
}

// TopicStatusChecker abstracts topic-state reads for parser-side limit enforcement.
type TopicStatusChecker interface {
	GetTopicByID(ctx context.Context, topicID string) (*mongomodels.ListeningTopic, error)
}

// ParserService consumes raw Data365 payloads, normalizes them into
// ListeningMention records, applies filters, and emits parsed mentions.
type ParserService struct {
	producer    kafka.Producer
	dedup       DedupChecker
	topicStatus TopicStatusChecker
	dedupTTL    time.Duration
	log         *logger.Logger
}

// NewParserService creates a new ParserService.
func NewParserService(
	producer kafka.Producer,
	dedup DedupChecker,
	topicStatus TopicStatusChecker,
	log *logger.Logger,
	dedupTTLHours int,
) *ParserService {
	ttl := time.Duration(dedupTTLHours) * time.Hour
	if ttl == 0 {
		ttl = 48 * time.Hour
	}
	return &ParserService{
		producer:    producer,
		dedup:       dedup,
		topicStatus: topicStatus,
		dedupTTL:    ttl,
		log:         log,
	}
}

// HandleRawPayload is a kafka.MessageHandler that parses a raw Data365 payload.
func (s *ParserService) HandleRawPayload(ctx context.Context, _ string, _ []byte, value []byte) error {
	var raw kafkamodels.ListeningRawPayload
	if err := json.Unmarshal(value, &raw); err != nil {
		return fmt.Errorf("ParserService.HandleRawPayload: unmarshal: %w", err)
	}

	if s.topicStatus != nil {
		topic, err := s.topicStatus.GetTopicByID(ctx, raw.TopicID)
		if err != nil {
			s.log.Warn().Err(err).
				Str("topic_id", raw.TopicID).
				Msg("Failed to check topic status in parser, proceeding (fail-open)")
		} else if topic == nil {
			s.log.Info().
				Str("topic_id", raw.TopicID).
				Str("platform", raw.Platform).
				Str("keyword", raw.Keyword).
				Msg("Topic no longer exists, skipping raw payload")
			return nil
		} else if topic.Status != "" && topic.Status != "active" {
			s.log.Info().
				Str("topic_id", raw.TopicID).
				Str("platform", raw.Platform).
				Str("keyword", raw.Keyword).
				Str("status", topic.Status).
				Msg("Topic is no longer active, skipping raw payload")
			return nil
		}
	}

	// Data365 wraps results in {"items": [...], "page_info": ...}.
	// Fall back to treating RawData as a bare array for forward compatibility.
	var envelope struct {
		Items []json.RawMessage `json:"items"`
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw.RawData, &envelope); err == nil && envelope.Items != nil {
		items = envelope.Items
	} else if err := json.Unmarshal(raw.RawData, &items); err != nil {
		return fmt.Errorf("ParserService.HandleRawPayload: unmarshal items: %w", err)
	}

	order := raw.WorkOrder
	log := s.log.With().
		Str("topic_id", raw.TopicID).
		Str("platform", raw.Platform).
		Str("keyword", raw.Keyword).
		Int("raw_items", len(items)).
		Logger()

	log.Debug().
		Strs("include_any", order.IncludeAny).
		Strs("include_all", order.IncludeAll).
		Strs("exclude_keywords", order.ExcludeKeywords).
		Strs("include_authors", order.IncludeAuthors).
		Msg("Applying filters")

	var emitted, filtered, deduped int
	for _, item := range items {
		mention, err := normalizeMention(raw.Platform, raw.TopicID, item)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to normalize mention")
			continue
		}

		if !passesFilters(mention, order) {
			filtered++
			continue
		}

		dedupKey := fmt.Sprintf("listening:dedup:%s:%s", raw.TopicID, mention.ContentHash)
		isNew, err := s.dedup.SetNX(ctx, dedupKey, "1", s.dedupTTL)
		if err != nil {
			// Fail-open: process the mention on Redis errors to avoid silent data loss.
			log.Warn().Err(err).Str("mention_id", mention.MentionID).Msg("Dedup check failed, processing mention (fail-open)")
		} else if !isNew {
			deduped++
			continue
		}

		mention.MatchedKeywords = computeMatchedKeywords(mention.PostText, order.IncludeKeywords, order.ExactMatch, order.CaseSensitive)
		mention.MentionsLimit = order.MentionsLimit
		mention.WorkspaceID = order.WorkspaceID
		mention.SuperAdminID = order.SuperAdminID

		data, err := json.Marshal(mention)
		if err != nil {
			log.Warn().Err(err).Str("mention_id", mention.MentionID).Msg("Marshal failed")
			continue
		}

		if err := s.producer.Produce(ctx, kafkamodels.TopicListeningParsed, []byte(mention.MentionID), data); err != nil {
			log.Error().Err(err).Str("mention_id", mention.MentionID).Msg("Produce failed")
			continue
		}
		emitted++
	}

	log.Info().
		Int("emitted", emitted).
		Int("filtered", filtered).
		Int("deduped", deduped).
		Msg("Parsed raw payload")

	return nil
}

// rawMention is a generic shape that covers fields across all 6 platforms.
type rawMention struct {
	ID                   flexibleString     `json:"id"`
	Text                 flexibleString     `json:"text"`
	Body                 flexibleString     `json:"body"`
	Caption              flexibleString     `json:"caption"`
	Description          flexibleString     `json:"description"`
	Title                flexibleString     `json:"title"`
	AuthorID             flexibleString     `json:"author_id"`
	OwnerID              flexibleString     `json:"owner_id"`
	UserID               flexibleString     `json:"user_id"`
	AuthorName           flexibleString     `json:"author_name"`
	OwnerFullName        flexibleString     `json:"owner_full_name"`
	Username             flexibleString     `json:"username"`
	OwnerUsername        flexibleString     `json:"owner_username"`
	AuthorHandle         flexibleString     `json:"author_handle"` // Fallback
	Author               rawAuthor          `json:"author"`
	AuthorAvatar         flexibleString     `json:"author_avatar_url"`
	AvatarURL            flexibleString     `json:"avatar_url"`
	OwnerAvatar          flexibleString     `json:"owner_avatar_url"`
	OwnerProfilePicURL   flexibleString     `json:"owner_profile_pic_url"`
	FollowersCount       flexibleInt64      `json:"followers_count"`
	AuthorFollowers      flexibleInt64      `json:"author_followers"`
	AuthorFollowersCount flexibleInt64      `json:"author_followers_count"`
	Language             flexibleString     `json:"language"`
	Lang                 flexibleString     `json:"lang"`
	LanguageCode         flexibleString     `json:"language_code"`
	Region               flexibleString     `json:"region"`
	Country              flexibleString     `json:"country"`
	CountryCode          flexibleString     `json:"country_code"`
	Hashtags             flexibleStringList `json:"hashtags"`
	Tags                 flexibleStringList `json:"tags"`
	URL                  flexibleString     `json:"url"`
	Permalink            flexibleString     `json:"permalink"`
	Link                 flexibleString     `json:"link"`
	PostURL              flexibleString     `json:"post_url"`
	Shortcode            flexibleString     `json:"shortcode"`
	CreatedAt            flexibleString     `json:"created_at"`
	CreatedTime          flexibleString     `json:"created_time"`
	PostedAt             flexibleString     `json:"posted_at"`
	Published            flexibleString     `json:"published"`
	Timestamp            float64            `json:"timestamp"`
	Likes                int                `json:"likes"`
	LikeCount            int                `json:"like_count"`
	LikesCount           int                `json:"likes_count"`
	Favorites            int                `json:"favorites"`
	LikeCnt              int                `json:"reactions_like_count"`
	ReactionsTotalCount  int                `json:"reactions_total_count"`
	Comments             int                `json:"comments"`
	CommentCount         int                `json:"comment_count"`
	CommentsCount        int                `json:"comments_count"`
	ReplyCount           int                `json:"reply_count"`
	Shares               int                `json:"shares"`
	ShareCount           int                `json:"share_count"`
	SharesCount          int                `json:"shares_count"`
	RepostsCount         int                `json:"reposts_count"`
	ResharesCount        int                `json:"reshares_count"`
	RetweetCount         int                `json:"retweet_count"`
	Views                int                `json:"views"`
	ViewCount            int                `json:"view_count"`
	PlayCount            int                `json:"play_count"`
	QuoteCount           int                `json:"quote_count"`
	QuotesCount          int                `json:"quotes_count"`
	Score                int                `json:"score"`
	Type                 flexibleString     `json:"type"`
	MediaType            flexibleString     `json:"media_type"`
	IsVideo              bool               `json:"is_video"`
	Images               flexibleStringList `json:"attached_image_urls"`
	ImageURL             flexibleStringList `json:"attached_image_url"`
}

// normalizeMention converts a raw platform-specific JSON into a ListeningMention.
func normalizeMention(platform, topicID string, raw json.RawMessage) (*kafkamodels.ListeningMention, error) {
	var r rawMention
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("normalizeMention: unmarshal: %w", err)
	}

	if r.ID.String() == "" {
		return nil, fmt.Errorf("normalizeMention: missing id")
	}

	text := coalesce(r.Text.String(), r.Body.String(), r.Caption.String(), r.Description.String(), r.Title.String())
	text = truncateText(text, maxTextBytes)

	authorID := coalesce(r.AuthorID.String(), r.OwnerID.String(), r.UserID.String(), r.Username.String(), r.OwnerUsername.String(), r.AuthorHandle.String(), r.Author.Handle.String())
	authorName := coalesce(r.AuthorName.String(), r.OwnerFullName.String(), r.Username.String(), r.OwnerUsername.String(), r.AuthorHandle.String(), r.Author.Handle.String())
	authorHandle := coalesce(r.Username.String(), r.OwnerUsername.String(), r.AuthorHandle.String(), r.Author.Handle.String(), r.AuthorID.String(), r.OwnerID.String(), r.UserID.String())
	authorImage := coalesce(r.AuthorAvatar.String(), r.AvatarURL.String(), r.OwnerAvatar.String(), r.OwnerProfilePicURL.String())
	authorURL := ConstructAuthorProfileURL(platform, authorHandle, authorID, r.Author.URL.String())

	postURL := coalesce(r.PostURL.String(), r.Link.String(), r.URL.String(), r.Permalink.String())
	if postURL == "" {
		postURL = constructPlatformURL(platform, r.ID.String(), r.Shortcode.String(), authorHandle, authorID)
	}

	mediaURLs := make([]string, 0, len(r.ImageURL)+len(r.Images))
	mediaURLs = append(mediaURLs, r.ImageURL...)
	mediaURLs = append(mediaURLs, r.Images...)
	mediaURLs = compactStrings(mediaURLs)
	authorFollowers := maxInt64(r.FollowersCount.Int64(), r.AuthorFollowers.Int64(), r.AuthorFollowersCount.Int64(), r.Author.FollowersCount.Int64())
	language := normalizeMentionLanguage(coalesce(r.Lang.String(), r.Language.String(), r.LanguageCode.String()))
	authorCountry := normalizeCountryCode(coalesce(
		r.Region.String(), r.Country.String(), r.CountryCode.String(),
		r.Author.Region.String(), r.Author.Country.String(), r.Author.CountryCode.String(),
	))

	postedAt := parseTimestamp(r.PostedAt.String(), r.CreatedAt.String(), r.CreatedTime.String(), r.Published.String(), int64(r.Timestamp))
	now := time.Now().UTC()

	mentionID := fmt.Sprintf("%s:%s", platform, r.ID.String())
	contentHash := computeHash(topicID, mentionID, text)

	engagement := computeEngagement(r)
	contentType := normalizeContentType(platform, r)
	mediaType := normalizeMediaType(r)

	likes, comments, shares := extractDiscreteEngagement(r)

	return &kafkamodels.ListeningMention{
		MentionID:       mentionID,
		TopicID:         topicID,
		Platform:        platform,
		NativeID:        r.ID.String(),
		ContentHash:     contentHash,
		AuthorID:        authorID,
		AuthorName:      authorName,
		AuthorHandle:    authorHandle,
		AuthorImageURL:  authorImage,
		AuthorURL:       authorURL,
		AuthorFollowers: authorFollowers,
		PostText:        text,
		Language:        language,
		AuthorCountry:   authorCountry,
		PostedAt:        postedAt,
		TotalEngagement: engagement,
		LikesCount:      likes,
		CommentsCount:   comments,
		SharesCount:     shares,
		ContentType:     contentType,
		MediaType:       mediaType,
		URL:             postURL,
		MediaURLs:       mediaURLs,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func constructPlatformURL(platform, id, shortcode, authorHandle, authorID string) string {
	switch platform {
	case "instagram":
		if shortcode != "" {
			return fmt.Sprintf("https://www.instagram.com/p/%s/", shortcode)
		}
		return fmt.Sprintf("https://www.instagram.com/p/%s/", id)
	case "facebook":
		return fmt.Sprintf("https://www.facebook.com/%s", id)
	case "tiktok":
		handle := strings.TrimPrefix(authorHandle, "@")
		if handle != "" {
			return fmt.Sprintf("https://www.tiktok.com/@%s/video/%s", handle, id)
		}
		// Inference: when the author handle is missing, TikTok's canonical page URL
		// cannot be reconstructed reliably from the video ID alone.
		return fmt.Sprintf("https://www.tiktok.com/embed/v2/%s", id)
	case "twitter":
		handle := strings.TrimPrefix(authorHandle, "@")
		if handle != "" {
			return fmt.Sprintf("https://x.com/%s/status/%s", handle, id)
		}
		if authorID != "" {
			return fmt.Sprintf("https://x.com/%s/status/%s", authorID, id)
		}
		return fmt.Sprintf("https://x.com/status/%s", id)
	}
	return ""
}

func extractDiscreteEngagement(r rawMention) (int64, int64, int64) {
	likes := int64(max(r.Likes, r.LikeCount, r.LikesCount, r.Favorites, r.LikeCnt))
	comments := int64(max(r.Comments, r.CommentCount, r.CommentsCount, r.ReplyCount))
	shares := int64(max(r.Shares, r.ShareCount, r.SharesCount, r.RepostsCount, r.ResharesCount, r.RetweetCount, r.QuoteCount, r.QuotesCount))
	return likes, comments, shares
}

// passesFilters checks exclude_keywords, include_any, include_all, include_authors,
// exclude_authors, global_excluded_subreddits, and language.
func passesFilters(m *kafkamodels.ListeningMention, order kafkamodels.ListeningWorkOrder) bool {
	// ExcludeKeywords — always case-insensitive substring match.
	// Skip blank entries (Laravel stores [""] when none are set).
	textLower := strings.ToLower(m.PostText)
	for _, kw := range order.ExcludeKeywords {
		if kw == "" {
			continue
		}
		if strings.Contains(textLower, strings.ToLower(kw)) {
			return false
		}
	}

	// IncludeAny: at least one non-blank term must appear in the text.
	hasIncludeAny := false
	for _, term := range order.IncludeAny {
		if term != "" {
			hasIncludeAny = true
			break
		}
	}
	if hasIncludeAny {
		found := false
		for _, term := range order.IncludeAny {
			if term != "" && containsTerm(m.PostText, term, order.ExactMatch, order.CaseSensitive) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// IncludeAll: every non-blank term must appear in the text.
	for _, term := range order.IncludeAll {
		if term == "" {
			continue
		}
		if !containsTerm(m.PostText, term, order.ExactMatch, order.CaseSensitive) {
			return false
		}
	}

	if len(order.IncludeAuthors) > 0 {
		found := false
		authorLower := strings.ToLower(m.AuthorID)
		for _, a := range order.IncludeAuthors {
			if strings.EqualFold(a, authorLower) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for _, a := range order.ExcludeAuthors {
		if strings.EqualFold(a, m.AuthorID) {
			return false
		}
	}

	if m.Platform == "reddit" && len(order.GlobalExcludedSubreddits) > 0 {
		sub := extractSubreddit(m.URL)
		if sub != "" {
			for _, excluded := range order.GlobalExcludedSubreddits {
				if strings.EqualFold(excluded, sub) {
					return false
				}
			}
		}
	}

	// Data365 filters server-side for Twitter (trigger) and Facebook/Twitter (results),
	// but Instagram, TikTok, Reddit, and Threads have no Data365 language param — so we
	// enforce it post-fetch here. Mentions without a detected language pass through.
	if len(order.Languages) > 0 && m.Language != "" {
		match := false
		for _, lang := range order.Languages {
			if lang != "" && strings.EqualFold(lang, m.Language) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	// Data365 has no region filter on any platform, so enforce post-fetch.
	// Only platforms that expose author country in raw data can be filtered reliably
	// (primarily TikTok); mentions without a country pass through.
	if len(order.Regions) > 0 && m.AuthorCountry != "" {
		match := false
		for _, region := range order.Regions {
			if region != "" && strings.EqualFold(region, m.AuthorCountry) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	return true
}

// normalizeCountryCode returns an uppercase ISO 3166-1 alpha-2 country code,
// or "" when the input isn't a valid 2-letter code. Data365 occasionally returns
// free-form region strings (e.g. "San Francisco, CA") which must be rejected so
// the regions filter stays permissive rather than dropping everything.
func normalizeCountryCode(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) != 2 {
		return ""
	}
	for _, r := range trimmed {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return ""
		}
	}
	return strings.ToUpper(trimmed)
}

// extractSubreddit parses a Reddit URL and returns the subreddit name.
// Handles URLs like "https://reddit.com/r/golang/..." or "https://www.reddit.com/r/golang/...".
func extractSubreddit(url string) string {
	idx := strings.Index(strings.ToLower(url), "/r/")
	if idx < 0 {
		return ""
	}
	rest := url[idx+3:]
	// Trim at the first path separator, query string, or fragment.
	for _, sep := range []string{"/", "?", "#"} {
		if i := strings.Index(rest, sep); i >= 0 {
			rest = rest[:i]
		}
	}
	return rest
}

// containsTerm checks whether text contains term, respecting exactMatch and caseSensitive flags.
// exactMatch requires the term to appear as a whole-phrase boundary (not mid-word).
func containsTerm(text, term string, exactMatch, caseSensitive bool) bool {
	if !caseSensitive {
		text = strings.ToLower(text)
		term = strings.ToLower(term)
	}

	if term == "" {
		return false
	}

	searchFrom := 0
	for {
		relIdx := strings.Index(text[searchFrom:], term)
		if relIdx < 0 {
			return false
		}

		idx := searchFrom + relIdx
		if !exactMatch {
			return true
		}

		// Exact match: term must not be bordered by word characters.
		before := idx == 0 || !isWordChar(rune(text[idx-1]))
		afterIdx := idx + len(term)
		after := afterIdx == len(text) || !isWordChar(rune(text[afterIdx]))
		if before && after {
			return true
		}

		searchFrom = idx + len(term)
	}
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// computeMatchedKeywords returns which include_keywords appear in the text,
// respecting the topic's exactMatch and caseSensitive flags.
func computeMatchedKeywords(text string, keywords []string, exactMatch, caseSensitive bool) []string {
	var matched []string
	for _, kw := range keywords {
		if containsTerm(text, kw, exactMatch, caseSensitive) {
			matched = append(matched, kw)
		}
	}
	return matched
}

func computeHash(topicID, _ string, text string) string {
	h := sha256.Sum256([]byte(topicID + ":" + text))
	return hex.EncodeToString(h[:])
}

func computeEngagement(r rawMention) int64 {
	likes := max(r.Likes, r.LikeCount, r.LikesCount, r.Favorites, r.ReactionsTotalCount)
	comments := max(r.Comments, r.CommentCount, r.CommentsCount, r.ReplyCount)
	shares := max(r.Shares, r.ShareCount, r.SharesCount, r.RepostsCount, r.ResharesCount, r.RetweetCount, r.QuoteCount, r.QuotesCount)
	views := max(r.Views, r.ViewCount, r.PlayCount)
	return int64(likes + comments + shares + views + r.Score)
}

func normalizeContentType(platform string, r rawMention) string {
	if r.Type.String() != "" {
		return r.Type.String()
	}
	if r.IsVideo || platform == "tiktok" {
		return "video"
	}
	return "post"
}

func normalizeMediaType(r rawMention) string {
	if r.MediaType.String() != "" {
		return r.MediaType.String()
	}
	if r.IsVideo {
		return "video"
	}
	return "text"
}

func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		result = append(result, value)
	}
	return result
}

func normalizeMentionLanguage(value string) string {
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

func truncateText(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}

	truncated := s[:maxBytes]
	for !utf8.ValidString(truncated) && len(truncated) > 0 {
		truncated = truncated[:len(truncated)-1]
	}
	return truncated
}

func parseTimestamp(postedAt, createdAt, createdTime, published string, ts int64) time.Time {
	for _, s := range []string{postedAt, createdAt, createdTime, published} {
		if s == "" {
			continue
		}
		for _, layout := range []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02 15:04:05",
		} {
			if t, err := time.Parse(layout, s); err == nil {
				return t.UTC()
			}
		}
	}
	if ts > 0 {
		return time.Unix(ts, 0).UTC()
	}
	return time.Now().UTC()
}

func max(vals ...int) int {
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

func maxInt64(vals ...int64) int64 {
	if len(vals) == 0 {
		return 0
	}

	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}
