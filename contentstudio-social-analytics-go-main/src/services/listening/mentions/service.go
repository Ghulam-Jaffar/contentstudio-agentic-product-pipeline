// Package mentions serves the user-facing listening API: querying stored
// mentions from ClickHouse and exporting them as CSV or PDF. PDF generation
// uses maroto rather than HTML-to-PDF so exports work in headless containers
// without a Chromium install.
package mentions

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	chModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/parser"
)

type MentionReader interface {
	QueryMentions(ctx context.Context, filter *clickhouse.MentionFilter) ([]chModels.ListeningMentionRow, string, error)
	CountUnread(ctx context.Context, filter *clickhouse.MentionFilter) (int, error)
	GetMention(ctx context.Context, mentionID, topicID string) (*chModels.ListeningMentionRow, error)
	UpdateMention(ctx context.Context, existing chModels.ListeningMentionRow) error
	MarkAllRead(ctx context.Context, filter *clickhouse.MentionFilter) (int, error)
	GetAnalytics(ctx context.Context, filter *clickhouse.MentionFilter) (*clickhouse.AnalyticsData, error)
}

type Service struct {
	reader MentionReader
	logger zerolog.Logger
}

func NewService(reader MentionReader, logger zerolog.Logger) *Service {
	return &Service{
		reader: reader,
		logger: logger.With().Str("service", "listening-mentions").Logger(),
	}
}

func (s *Service) ListMentions(ctx context.Context, filter *apiModels.MentionFilter) (*apiModels.MentionListResponse, error) {
	chFilter := toClickHouseFilter(filter)

	rows, nextCursor, err := s.reader.QueryMentions(ctx, chFilter)
	if err != nil {
		return nil, fmt.Errorf("mentions.Service.ListMentions: %w", err)
	}

	unreadCount, err := s.reader.CountUnread(ctx, chFilter)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to count unread mentions")
		unreadCount = 0
	}

	data := make([]apiModels.MentionResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, toMentionResponse(row))
	}

	return &apiModels.MentionListResponse{
		Status:      true,
		Data:        data,
		NextCursor:  nextCursor,
		HasMore:     nextCursor != "",
		TotalUnread: unreadCount,
	}, nil
}

func (s *Service) GetUnreadCount(ctx context.Context, filter *apiModels.MentionFilter) (int, error) {
	chFilter := toClickHouseFilter(filter)
	count, err := s.reader.CountUnread(ctx, chFilter)
	if err != nil {
		return 0, fmt.Errorf("mentions.Service.GetUnreadCount: %w", err)
	}
	return count, nil
}

func (s *Service) PatchMention(ctx context.Context, mentionID string, patch *apiModels.MentionPatchRequest) error {
	existing, err := s.reader.GetMention(ctx, mentionID, patch.TopicID)
	if err != nil {
		return fmt.Errorf("mentions.Service.PatchMention: get: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("mentions.Service.PatchMention: mention not found: %s", mentionID)
	}

	if patch.IsRead != nil {
		existing.PostRead = *patch.IsRead
	}
	if patch.IsBookmarked != nil {
		existing.Bookmark = *patch.IsBookmarked
	}
	if patch.IsIrrelevant != nil {
		existing.PostIrrelevant = *patch.IsIrrelevant
	}
	if patch.SentimentOverride != "" {
		existing.SentimentOverride = patch.SentimentOverride
	}

	if err := s.reader.UpdateMention(ctx, *existing); err != nil {
		return fmt.Errorf("mentions.Service.PatchMention: update: %w", err)
	}

	s.logger.Debug().
		Str("mention_id", mentionID).
		Msg("Patched mention")

	return nil
}

func (s *Service) MarkAllRead(ctx context.Context, filter *apiModels.MentionFilter) (int, error) {
	chFilter := toClickHouseFilter(filter)
	count, err := s.reader.MarkAllRead(ctx, chFilter)
	if err != nil {
		return 0, fmt.Errorf("mentions.Service.MarkAllRead: %w", err)
	}

	s.logger.Info().
		Int("count", count).
		Msg("Marked mentions as read")

	return count, nil
}

func (s *Service) GetAnalytics(ctx context.Context, filter *apiModels.MentionFilter, topicNames map[string]string) (*apiModels.ListeningAnalyticsResponse, error) {
	chFilter := toClickHouseFilter(filter)
	data, err := s.reader.GetAnalytics(ctx, chFilter)
	if err != nil {
		return nil, fmt.Errorf("mentions.Service.GetAnalytics: %w", err)
	}

	// Calculate metrics
	positive := data.SentimentCounts["positive"]
	positiveShare := 0.0
	if data.TotalMentions > 0 {
		positiveShare = (float64(positive) / float64(data.TotalMentions)) * 100
	}

	// Mentions Over Time mapping
	topicDataMap := make(map[string]*apiModels.TopicTimeSeries)
	for _, p := range data.MentionsOverTime {
		if _, ok := topicDataMap[p.TopicID]; !ok {
			topicDataMap[p.TopicID] = &apiModels.TopicTimeSeries{
				TopicID:   p.TopicID,
				TopicName: topicNames[p.TopicID],
				Data:      []apiModels.TimeSeriesPoint{},
			}
			if topicDataMap[p.TopicID].TopicName == "" {
				topicDataMap[p.TopicID].TopicName = "Topic " + p.TopicID
			}
		}
		topicDataMap[p.TopicID].Data = append(topicDataMap[p.TopicID].Data, apiModels.TimeSeriesPoint{
			Date:  p.Date.Format("2006-01-02"),
			Value: p.Count,
		})
	}

	mentionsOverTime := make([]apiModels.TopicTimeSeries, 0, len(topicDataMap))
	for _, v := range topicDataMap {
		mentionsOverTime = append(mentionsOverTime, *v)
	}

	// Sentiment Trend mapping
	sentimentDataMap := make(map[string]*apiModels.SentimentTimeSeries)
	for _, p := range data.SentimentTrend {
		if _, ok := sentimentDataMap[p.Sentiment]; !ok {
			sentimentDataMap[p.Sentiment] = &apiModels.SentimentTimeSeries{
				Sentiment: p.Sentiment,
				Data:      []apiModels.TimeSeriesPoint{},
			}
		}
		sentimentDataMap[p.Sentiment].Data = append(sentimentDataMap[p.Sentiment].Data, apiModels.TimeSeriesPoint{
			Date:  p.Date.Format("2006-01-02"),
			Value: p.Count,
		})
	}

	sentimentTrend := make([]apiModels.SentimentTimeSeries, 0, len(sentimentDataMap))
	for _, v := range sentimentDataMap {
		sentimentTrend = append(sentimentTrend, *v)
	}

	// Distributions
	sentimentDist := make([]apiModels.DistributionItem, 0, len(data.SentimentCounts))
	for label, val := range data.SentimentCounts {
		sentimentDist = append(sentimentDist, apiModels.DistributionItem{Label: label, Value: val})
	}

	platformDist := make([]apiModels.DistributionItem, 0, len(data.PlatformCounts))
	for label, val := range data.PlatformCounts {
		platformDist = append(platformDist, apiModels.DistributionItem{Label: label, Value: val})
	}

	tagDist := make([]apiModels.DistributionItem, 0, len(data.TagCounts))
	for label, val := range data.TagCounts {
		tagDist = append(tagDist, apiModels.DistributionItem{Label: label, Value: val})
	}

	// Days for average calculation
	days := 1.0
	if !chFilter.DateFrom.IsZero() && !chFilter.DateTo.IsZero() {
		days = chFilter.DateTo.Sub(chFilter.DateFrom).Hours() / 24
		if days < 1 {
			days = 1
		}
	} else if len(mentionsOverTime) > 0 {
		// Fallback to series length if no dates provided
		maxPoints := 0
		for _, t := range mentionsOverTime {
			if len(t.Data) > maxPoints {
				maxPoints = len(t.Data)
			}
		}
		if maxPoints > 0 {
			days = float64(maxPoints)
		}
	}

	avgDaily := float64(data.TotalMentions) / days

	return &apiModels.ListeningAnalyticsResponse{
		Status: true,
		Engagement: apiModels.EngagementMetrics{
			TotalMentions:          data.TotalMentions,
			PositiveSentimentShare: positiveShare,
			AvgDailyMentions:       float64(int(avgDaily*10)) / 10,
			TopicsTracked:          len(chFilter.TopicIDs),
		},
		MentionsOverTime:      mentionsOverTime,
		SentimentTrend:        sentimentTrend,
		SentimentDistribution: sentimentDist,
		PlatformDistribution:  platformDist,
		TagDistribution:       tagDist,
	}, nil
}

func (s *Service) ExportMentionsCSV(ctx context.Context, filter *apiModels.MentionFilter, topicNames map[string]string, w io.Writer) error {
	chFilter := toClickHouseFilter(filter)
	chFilter.Limit = 100

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write BOM for Excel compatibility with UTF-8
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	// Header
	header := []string{
		"Date",
		"Platform",
		"Topic",
		"Author",
		"Followers",
		"Content",
		"Sentiment",
		"Engagement",
		"Tags",
		"URL",
	}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	for {
		rows, nextCursor, err := s.reader.QueryMentions(ctx, chFilter)
		if err != nil {
			return fmt.Errorf("mentions.Service.ExportMentionsCSV: %w", err)
		}

		for _, row := range rows {
			topicName := topicNames[row.TopicID]
			if topicName == "" {
				topicName = row.TopicID
			}

			tags := strings.Join(row.AITags, ", ")
			engagement := strconv.FormatInt(int64(row.TotalEngagement), 10)
			followers := strconv.FormatInt(int64(row.AuthorFollowers), 10)

			record := []string{
				row.PostedAt.Format("2006-01-02 15:04:05"),
				row.Platform,
				topicName,
				row.AuthorName,
				followers,
				row.PostText,
				row.SentimentLabel,
				engagement,
				tags,
				row.URL,
			}

			if err := csvWriter.Write(record); err != nil {
				return err
			}
		}

		if nextCursor == "" {
			break
		}
		chFilter.Cursor = nextCursor
	}

	return nil
}

func (s *Service) ExportMentionsPDF(ctx context.Context, filter *apiModels.MentionFilter, topicNames map[string]string, w io.Writer) error {
	chFilter := toClickHouseFilter(filter)

	// 1. Fetch Analytics Data
	analytics, err := s.GetAnalytics(ctx, filter, topicNames)
	if err != nil {
		return fmt.Errorf("mentions.Service.ExportMentionsPDF: analytics: %w", err)
	}

	// 2. Fetch Top Mentions (latest 50)
	chFilter.Limit = 50
	rows, _, err := s.reader.QueryMentions(ctx, chFilter)
	if err != nil {
		return fmt.Errorf("mentions.Service.ExportMentionsPDF: mentions: %w", err)
	}

	// 3. Create PDF
	cfg := config.NewBuilder().Build()
	m := maroto.New(cfg)

	// Header
	m.AddRow(20,
		text.NewCol(12, "Social Listening Analytics Report", props.Text{
			Size:  24,
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: &props.Color{Red: 37, Green: 99, Blue: 235}, // Primary Blue
		}),
	)
	m.AddRow(10,
		text.NewCol(12, fmt.Sprintf("Generated on %s", time.Now().Format("Jan 02, 2006")), props.Text{
			Size:  10,
			Align: align.Center,
			Style: fontstyle.Italic,
		}),
	)
	m.AddRow(1,
		col.New(12).Add(line.New(props.Line{Thickness: 0.5, Color: &props.Color{Red: 200, Green: 200, Blue: 200}})),
	)

	// Overview Section
	m.AddRow(20, text.NewCol(12, "Performance Overview", props.Text{Size: 16, Style: fontstyle.Bold}))

	m.AddRow(15,
		text.NewCol(3, "Total Mentions", props.Text{Style: fontstyle.Bold, Size: 10}),
		text.NewCol(3, "Avg Daily Mentions", props.Text{Style: fontstyle.Bold, Size: 10}),
		text.NewCol(3, "Topics Tracked", props.Text{Style: fontstyle.Bold, Size: 10}),
		text.NewCol(3, "Positive Share", props.Text{Style: fontstyle.Bold, Size: 10}),
	)
	m.AddRow(15,
		text.NewCol(3, strconv.Itoa(analytics.Engagement.TotalMentions), props.Text{Size: 14}),
		text.NewCol(3, fmt.Sprintf("%.1f", analytics.Engagement.AvgDailyMentions), props.Text{Size: 14}),
		text.NewCol(3, strconv.Itoa(analytics.Engagement.TopicsTracked), props.Text{Size: 14}),
		text.NewCol(3, fmt.Sprintf("%.1f%%", analytics.Engagement.PositiveSentimentShare), props.Text{Size: 14}),
	)

	// Distributions
	m.AddRow(20, text.NewCol(12, "Distribution Highlights", props.Text{Size: 16, Style: fontstyle.Bold}))

	// Create a simple table-like distribution for Sentiment
	m.AddRow(10, text.NewCol(6, "Sentiment Distribution", props.Text{Style: fontstyle.Bold}))
	for _, item := range analytics.SentimentDistribution {
		m.AddRow(8,
			text.NewCol(1, ""), // indentation
			text.NewCol(3, strings.Title(item.Label), props.Text{Size: 9}),
			text.NewCol(2, strconv.Itoa(item.Value), props.Text{Size: 9, Align: align.Right}),
		)
	}

	// Top Mentions Highlights
	m.AddRow(20, text.NewCol(12, "Recent Mentions Highlights", props.Text{Size: 16, Style: fontstyle.Bold}))

	for i, row := range rows {
		if i >= 10 { // Limit to top 10 for the PDF summary
			break
		}

		topicName := topicNames[row.TopicID]
		if topicName == "" {
			topicName = row.TopicID
		}

		// Mention Card
		m.AddRow(15,
			text.NewCol(2, row.Platform, props.Text{Style: fontstyle.Bold, Size: 9}),
			text.NewCol(4, topicName, props.Text{Size: 9, Color: &props.Color{Red: 100, Green: 100, Blue: 100}}),
			text.NewCol(6, row.PostedAt.Format("Jan 02, 15:04"), props.Text{Size: 8, Align: align.Right}),
		)

		content := row.PostText
		if len(content) > 200 {
			content = content[:197] + "..."
		}
		m.AddAutoRow(text.NewCol(12, content, props.Text{Size: 9}))
		m.AddRow(5, text.NewCol(12, "", props.Text{})) // Spacer
		m.AddRow(1,
			col.New(12).Add(line.New(props.Line{Thickness: 0.2, Color: &props.Color{Red: 230, Green: 230, Blue: 230}})),
		)
		m.AddRow(5, text.NewCol(12, "", props.Text{})) // Spacer
	}

	// Generate and write
	doc, err := m.Generate()
	if err != nil {
		return fmt.Errorf("mentions.Service.ExportMentionsPDF: generate: %w", err)
	}

	_, err = w.Write(doc.GetBytes())
	return err
}

func toClickHouseFilter(f *apiModels.MentionFilter) *clickhouse.MentionFilter {
	chf := &clickhouse.MentionFilter{
		TopicIDs:           f.TopicIDs,
		Platforms:          f.Platforms,
		Sentiments:         f.Sentiments,
		AITags:             f.AITags,
		ExcludeAITags:      f.ExcludeAITags,
		Language:           f.Language,
		MinFollowers:       f.MinFollowers,
		MinTotalEngagement: f.MinTotalEngagement,
		Sort:               f.Sort,
		Cursor:             f.Cursor,
		Limit:              f.Limit,
		IsBookmarked:       f.IsBookmarked,
		IsRead:             f.IsRead,
		IncludeIrrelevant:  f.IncludeIrrelevant,
		Search:             f.Search,
	}

	if f.DateFrom != "" {
		if t, ok := parseFilterTime(f.DateFrom, false); ok {
			chf.DateFrom = t
		}
	}
	if f.DateTo != "" {
		if t, ok := parseFilterTime(f.DateTo, true); ok {
			chf.DateTo = t
		}
	}
	if chf.Limit <= 0 {
		chf.Limit = 25
	}

	return chf
}

func parseFilterTime(value string, endOfDay bool) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC(), true
	}

	if t, err := time.Parse("2006-01-02", value); err == nil {
		if endOfDay {
			return t.Add(24*time.Hour - time.Second).UTC(), true
		}
		return t.UTC(), true
	}

	return time.Time{}, false
}

func toMentionResponse(row chModels.ListeningMentionRow) apiModels.MentionResponse {
	return apiModels.MentionResponse{
		ID:                row.MentionID,
		TopicID:           row.TopicID,
		Platform:          row.Platform,
		AuthorID:          row.AuthorID,
		AuthorName:        row.AuthorName,
		AuthorHandle:      row.AuthorHandle,
		AuthorImageURL:    row.AuthorImageURL,
		AuthorURL:         parser.ConstructAuthorProfileURL(row.Platform, row.AuthorHandle, row.AuthorID, row.AuthorURL),
		AuthorFollowers:   row.AuthorFollowers,
		Content:           row.PostText,
		URL:               row.URL,
		MediaURLs:         row.MediaURLs,
		Language:          row.Language,
		PublishedAt:       row.PostedAt,
		Sentiment:         row.SentimentLabel,
		SentimentOverride: row.SentimentOverride,
		AITags:            row.AITags,
		Engagement: apiModels.MentionEngagment{
			Total:    row.TotalEngagement,
			Likes:    row.LikesCount,
			Comments: row.CommentsCount,
			Shares:   row.SharesCount,
		},
		IsRead:         row.PostRead,
		IsBookmarked:   row.Bookmark,
		IsIrrelevant:   row.PostIrrelevant,
		KeywordMatches: row.MatchedKeywords,
		ContentType:    row.ContentType,
		MediaType:      row.MediaType,
	}
}
