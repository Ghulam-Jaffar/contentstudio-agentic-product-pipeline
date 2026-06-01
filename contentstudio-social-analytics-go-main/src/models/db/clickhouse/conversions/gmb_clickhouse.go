package conversions

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// buildGmbID constructs the combined gmb_id from account and location IDs.
// Format: "accounts/{accountID}/locations/{locationID}"
func buildGmbID(accountID, locationID string) string {
	return fmt.Sprintf("accounts/%s/locations/%s", accountID, locationID)
}

// starRatingToUint64 converts a Google API star rating string (e.g., "FIVE") to a uint64.
func starRatingToUint64(rating string) uint64 {
	switch strings.ToUpper(rating) {
	case "ONE":
		return 1
	case "TWO":
		return 2
	case "THREE":
		return 3
	case "FOUR":
		return 4
	case "FIVE":
		return 5
	default:
		// Try parsing as numeric string
		if v, err := strconv.ParseUint(rating, 10, 64); err == nil {
			return v
		}
		return 0
	}
}

// ConvertGMBDailyMetrics converts the parsed Kafka model to ClickHouse model
func ConvertGMBDailyMetrics(p *kafkamodels.ParsedGMBDailyMetrics) *chmodels.GMBDailyMetrics {
	if p == nil {
		return nil
	}

	createdAt, _ := time.Parse("2006-01-02", p.Date)

	return &chmodels.GMBDailyMetrics{
		GmbID:                            buildGmbID(p.AccountID, p.LocationID),
		AccountID:                        p.AccountID,
		LocationID:                       p.LocationID,
		AccountName:                      p.AccountName,
		LocationName:                     p.LocationName,
		PlatformName:                     "gmb",
		InsertedAt:                       time.Now().UTC(),
		CreatedAt:                        createdAt,
		BusinessImpressionsDesktopMaps:   p.BusinessImpressionsDesktopMaps,
		BusinessImpressionsDesktopSearch: p.BusinessImpressionsDesktopSearch,
		BusinessImpressionsMobileMaps:    p.BusinessImpressionsMobileMaps,
		BusinessImpressionsMobileSearch:  p.BusinessImpressionsMobileSearch,
		CallClicks:                       p.CallClicks,
		WebsiteClicks:                    p.WebsiteClicks,
		BusinessDirectionRequests:        p.BusinessDirectionRequests,
		BusinessConversations:            p.BusinessConversations,
		BusinessBookings:                 p.BusinessBookings,
		BusinessFoodOrders:               p.BusinessFoodOrders,
		BusinessFoodMenuClicks:           p.BusinessFoodMenuClicks,
	}
}

// GMBDailyMetricsBuilder aggregates multiple metric time series into a single daily record.
// The GMB API returns each metric as a separate time series; this builder collects them
// into one ClickHouse row per date.
type GMBDailyMetricsBuilder struct {
	AccountID    string
	LocationID   string
	AccountName  string
	LocationName string
	Date         string // YYYY-MM-DD

	businessImpressionsDesktopMaps   uint64
	businessImpressionsDesktopSearch uint64
	businessImpressionsMobileMaps    uint64
	businessImpressionsMobileSearch  uint64
	callClicks                       uint64
	websiteClicks                    uint64
	businessDirectionRequests        uint64
	businessConversations            uint64
	businessBookings                 uint64
	businessFoodOrders               uint64
	businessFoodMenuClicks           uint64
}

// SetMetric sets a metric value by its API metric name string.
func (b *GMBDailyMetricsBuilder) SetMetric(metric, value string) {
	v, _ := strconv.ParseUint(value, 10, 64)
	switch metric {
	case "BUSINESS_IMPRESSIONS_DESKTOP_MAPS":
		b.businessImpressionsDesktopMaps = v
	case "BUSINESS_IMPRESSIONS_DESKTOP_SEARCH":
		b.businessImpressionsDesktopSearch = v
	case "BUSINESS_IMPRESSIONS_MOBILE_MAPS":
		b.businessImpressionsMobileMaps = v
	case "BUSINESS_IMPRESSIONS_MOBILE_SEARCH":
		b.businessImpressionsMobileSearch = v
	case "CALL_CLICKS":
		b.callClicks = v
	case "WEBSITE_CLICKS":
		b.websiteClicks = v
	case "BUSINESS_DIRECTION_REQUESTS":
		b.businessDirectionRequests = v
	case "BUSINESS_CONVERSATIONS":
		b.businessConversations = v
	case "BUSINESS_BOOKINGS":
		b.businessBookings = v
	case "BUSINESS_FOOD_ORDERS":
		b.businessFoodOrders = v
	case "BUSINESS_FOOD_MENU_CLICKS":
		b.businessFoodMenuClicks = v
	}
}

// Build creates a ClickHouse GMBDailyMetrics from the aggregated builder state.
func (b *GMBDailyMetricsBuilder) Build() *chmodels.GMBDailyMetrics {
	createdAt, _ := time.Parse("2006-01-02", b.Date)
	return &chmodels.GMBDailyMetrics{
		GmbID:                            buildGmbID(b.AccountID, b.LocationID),
		AccountID:                        b.AccountID,
		LocationID:                       b.LocationID,
		AccountName:                      b.AccountName,
		LocationName:                     b.LocationName,
		PlatformName:                     "gmb",
		InsertedAt:                       time.Now().UTC(),
		CreatedAt:                        createdAt,
		BusinessImpressionsDesktopMaps:   b.businessImpressionsDesktopMaps,
		BusinessImpressionsDesktopSearch: b.businessImpressionsDesktopSearch,
		BusinessImpressionsMobileMaps:    b.businessImpressionsMobileMaps,
		BusinessImpressionsMobileSearch:  b.businessImpressionsMobileSearch,
		CallClicks:                       b.callClicks,
		WebsiteClicks:                    b.websiteClicks,
		BusinessDirectionRequests:        b.businessDirectionRequests,
		BusinessConversations:            b.businessConversations,
		BusinessBookings:                 b.businessBookings,
		BusinessFoodOrders:               b.businessFoodOrders,
		BusinessFoodMenuClicks:           b.businessFoodMenuClicks,
	}
}

// BulkInsertGMBDailyMetrics is a thin wrapper delegating to the ClickHouse client
func (s *ClickHouseSink) BulkInsertGMBDailyMetrics(ctx context.Context, metrics []*chmodels.GMBDailyMetrics) error {
	return s.ClickhouseClient.BulkInsertGMBDailyMetrics(ctx, metrics)
}

// ConvertGMBMediaAsset converts the parsed Kafka model to ClickHouse model
func ConvertGMBMediaAsset(p *kafkamodels.ParsedGMBMediaAsset) *chmodels.GMBMediaAssets {
	if p == nil {
		return nil
	}

	createdAt, _ := time.Parse(time.RFC3339, p.CreateTime)

	return &chmodels.GMBMediaAssets{
		GmbID:                       buildGmbID(p.AccountID, p.LocationID),
		AccountID:                   p.AccountID,
		LocationID:                  p.LocationID,
		AccountName:                 p.AccountName,
		LocationName:                p.LocationName,
		PlatformName:                "gmb",
		LanguageCode:                p.LanguageCode,
		InsertedAt:                  time.Now().UTC(),
		CreatedAt:                   createdAt,
		MediaName:                   p.MediaName,
		SourceURL:                   p.SourceURL,
		MediaFormat:                 p.MediaFormat,
		LocationAssociationCategory: p.LocationAssociationCategory,
		GoogleURL:                   p.GoogleURL,
		ThumbnailURL:                p.ThumbnailURL,
		WidthPixels:                 p.WidthPixels,
		HeightPixels:                p.HeightPixels,
	}
}

// BulkInsertGMBMediaAssets is a thin wrapper delegating to the ClickHouse client
func (s *ClickHouseSink) BulkInsertGMBMediaAssets(ctx context.Context, assets []*chmodels.GMBMediaAssets) error {
	return s.ClickhouseClient.BulkInsertGMBMediaAssets(ctx, assets)
}

// ConvertGMBSearchKeyword converts the parsed Kafka model to ClickHouse model
func ConvertGMBSearchKeyword(p *kafkamodels.ParsedGMBSearchKeyword) *chmodels.GMBSearchKeywordsMonthly {
	if p == nil {
		return nil
	}

	keywordMonth, _ := time.Parse("2006-01", p.KeywordMonth)

	return &chmodels.GMBSearchKeywordsMonthly{
		GmbID:                buildGmbID(p.AccountID, p.LocationID),
		AccountID:            p.AccountID,
		LocationID:           p.LocationID,
		AccountName:          p.AccountName,
		LocationName:         p.LocationName,
		PlatformName:         "gmb",
		InsertedAt:           time.Now().UTC(),
		KeywordMonth:         keywordMonth,
		Keyword:              p.Keyword,
		ImpressionsValue:     p.ImpressionsValue,
		ImpressionsThreshold: p.ImpressionsThreshold,
	}
}

// BulkInsertGMBSearchKeywordsMonthly is a thin wrapper delegating to the ClickHouse client
func (s *ClickHouseSink) BulkInsertGMBSearchKeywordsMonthly(ctx context.Context, keywords []*chmodels.GMBSearchKeywordsMonthly) error {
	return s.ClickhouseClient.BulkInsertGMBSearchKeywordsMonthly(ctx, keywords)
}

// ConvertGMBLocalPost converts the parsed Kafka model to ClickHouse model
func ConvertGMBLocalPost(p *kafkamodels.ParsedGMBLocalPost) *chmodels.GMBLocalPosts {
	if p == nil {
		return nil
	}

	createdAt, _ := time.Parse(time.RFC3339, p.CreateTime)
	updatedAt, _ := time.Parse(time.RFC3339, p.UpdateTime)

	return &chmodels.GMBLocalPosts{
		GmbID:           buildGmbID(p.AccountID, p.LocationID),
		AccountID:       p.AccountID,
		LocationID:      p.LocationID,
		AccountName:     p.AccountName,
		LocationName:    p.LocationName,
		PlatformName:    "gmb",
		LanguageCode:    p.LanguageCode,
		InsertedAt:      time.Now().UTC(),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		PostName:        p.PostName,
		Summary:         p.Summary,
		State:           p.State,
		TopicType:       p.TopicType,
		SearchURL:       p.SearchURL,
		MediaNames:      p.MediaNames,
		MediaFormats:    p.MediaFormats,
		MediaGoogleURLs: p.MediaGoogleURLs,
	}
}

// BulkInsertGMBLocalPosts is a thin wrapper delegating to the ClickHouse client
func (s *ClickHouseSink) BulkInsertGMBLocalPosts(ctx context.Context, posts []*chmodels.GMBLocalPosts) error {
	return s.ClickhouseClient.BulkInsertGMBLocalPosts(ctx, posts)
}

// ConvertGMBReview converts the parsed Kafka model to ClickHouse model
func ConvertGMBReview(p *kafkamodels.ParsedGMBReview) *chmodels.GMBReviews {
	if p == nil {
		return nil
	}

	createdAt, _ := time.Parse(time.RFC3339, p.CreateTime)
	updatedAt, _ := time.Parse(time.RFC3339, p.UpdateTime)
	replyUpdateTime, _ := time.Parse(time.RFC3339, p.ReplyUpdateTime)

	return &chmodels.GMBReviews{
		GmbID:                   buildGmbID(p.AccountID, p.LocationID),
		AccountID:               p.AccountID,
		LocationID:              p.LocationID,
		AccountName:             p.AccountName,
		LocationName:            p.LocationName,
		PlatformName:            "gmb",
		InsertedAt:              time.Now().UTC(),
		CreatedAt:               createdAt,
		UpdatedAt:               updatedAt,
		ReviewID:                p.ReviewID,
		ReviewName:              p.ReviewName,
		ReviewerDisplayName:     p.ReviewerDisplayName,
		ReviewerProfilePhotoURL: p.ReviewerProfilePhotoURL,
		StarRating:              starRatingToUint64(p.StarRating),
		Comment:                 p.Comment,
		ReplyComment:            p.ReplyComment,
		ReplyUpdateTime:         replyUpdateTime,
	}
}

// BulkInsertGMBReviews is a thin wrapper delegating to the ClickHouse client
func (s *ClickHouseSink) BulkInsertGMBReviews(ctx context.Context, reviews []*chmodels.GMBReviews) error {
	return s.ClickhouseClient.BulkInsertGMBReviews(ctx, reviews)
}
