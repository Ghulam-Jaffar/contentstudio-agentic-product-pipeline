// Package campaign_label provides the business logic layer for campaign & label analytics.
// It orchestrates MongoDB post-ID resolution (cached campaign_analytics/label_analytics + fallback
// via plans → postings) with ClickHouse cross-platform analytics queries.
//
// Migrated from PHP: CampaignLabelAnalyticsController (contentstudio-backend).
package campaign_label

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/campaign_label"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/campaign_label"
)

// Service defines the campaign/label analytics business operations exposed to the API layer.
type Service interface {
	SetPostIds(ctx context.Context, req *types.CampaignLabelRequest) (*types.SetPostIdsResponse, error)
	GetSummaryAnalytics(ctx context.Context, req *types.CampaignLabelRequest) (*types.SummaryResponse, error)
	GetBreakdownData(ctx context.Context, req *types.CampaignLabelRequest) (map[string]interface{}, error)
	GetInsightsBreakdown(ctx context.Context, req *types.CampaignLabelRequest) (map[string]interface{}, error)
	GetPlannerAnalytics(ctx context.Context, req *types.PlannerAnalyticsRequest) (map[string]interface{}, error)
}

// CampaignLabelAnalyticsService orchestrates MongoDB and ClickHouse operations for
// campaign/label analytics.
type CampaignLabelAnalyticsService struct {
	repo              *repo.Repository
	campaignAnalytics *mongo.Collection // campaign_analytics collection
	labelAnalytics    *mongo.Collection // label_analytics collection
	plansCollection   *mongo.Collection // plans collection
	postingCollection *mongo.Collection // posting collection
	logger            zerolog.Logger
}

var _ Service = (*CampaignLabelAnalyticsService)(nil)

// NewCampaignLabelAnalyticsService constructs a campaign/label analytics service with
// ClickHouse repo and MongoDB database access.
func NewCampaignLabelAnalyticsService(r *repo.Repository, mongoDB *mongo.Database, logger zerolog.Logger) *CampaignLabelAnalyticsService {
	svc := &CampaignLabelAnalyticsService{
		repo:   r,
		logger: logger.With().Str("service", "campaign-label-analytics").Logger(),
	}
	if mongoDB != nil {
		svc.campaignAnalytics = mongoDB.Collection("campaign_analytics")
		svc.labelAnalytics = mongoDB.Collection("label_analytics")
		svc.plansCollection = mongoDB.Collection("plans")
		svc.postingCollection = mongoDB.Collection("posting")
	}
	return svc
}

func round2(x float64) float64 {
	return math.Round(x*100) / 100
}

// roundFloatValues walks a response payload and rounds any float values to 2 decimals.
// Count-like integer values are left untouched.
func roundFloatValues(v interface{}) interface{} {
	switch value := v.(type) {
	case map[string]interface{}:
		for k, child := range value {
			value[k] = roundFloatValues(child)
		}
		return value
	case []interface{}:
		for i, child := range value {
			value[i] = roundFloatValues(child)
		}
		return value
	case float64:
		return round2(value)
	case float32:
		return float32(round2(float64(value)))
	default:
		return v
	}
}

// ---- Post ID resolution from MongoDB ----
// These methods replicate the PHP CampaignLabelAnalyticsController's MongoDB resolution logic:
// 1. Check campaign_analytics / label_analytics for cached post IDs.
// 2. For unmatched: query plans → posting collections to find post IDs, then cache them.

// FlexStringSlice is a []string that can unmarshal BSON arrays containing mixed types
// (string, int64, int32, double). MongoDB's posted_ids field may store numeric IDs that
// need to be decoded as strings in Go.
type FlexStringSlice []string

// UnmarshalBSONValue implements bson.ValueUnmarshaler, handling mixed-type arrays.
func (f *FlexStringSlice) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t == bson.TypeNull || t == bson.TypeUndefined {
		*f = nil
		return nil
	}
	if t != bson.TypeArray {
		return fmt.Errorf("FlexStringSlice: expected array, got BSON type %v", t)
	}

	rawArr := bson.Raw(data)
	elems, err := rawArr.Elements()
	if err != nil {
		return fmt.Errorf("FlexStringSlice: %w", err)
	}

	result := make([]string, 0, len(elems))
	for _, elem := range elems {
		val := elem.Value()
		switch val.Type {
		case bson.TypeString:
			result = append(result, val.StringValue())
		case bson.TypeInt64:
			result = append(result, strconv.FormatInt(val.Int64(), 10))
		case bson.TypeInt32:
			result = append(result, strconv.FormatInt(int64(val.Int32()), 10))
		case bson.TypeDouble:
			result = append(result, strconv.FormatFloat(val.Double(), 'f', -1, 64))
		case bson.TypeObjectID:
			result = append(result, val.ObjectID().Hex())
		case bson.TypeNull, bson.TypeUndefined:
			// skip null/undefined elements inside the array
			continue
		default:
			return fmt.Errorf("FlexStringSlice: unsupported element type %v", val.Type)
		}
	}
	*f = result
	return nil
}

// FlexString is a string that can unmarshal BSON scalars stored as either strings or numbers.
// Legacy MongoDB documents may store platform_id / posted_id as numeric values.
type FlexString string

// UnmarshalBSONValue implements bson.ValueUnmarshaler for mixed scalar types.
func (f *FlexString) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	switch t {
	case bson.TypeNull, bson.TypeUndefined:
		*f = ""
		return nil
	case bson.TypeString:
		var value string
		if err := bson.UnmarshalValue(t, data, &value); err != nil {
			return fmt.Errorf("FlexString: %w", err)
		}
		*f = FlexString(value)
		return nil
	case bson.TypeInt64:
		var value int64
		if err := bson.UnmarshalValue(t, data, &value); err != nil {
			return fmt.Errorf("FlexString: %w", err)
		}
		*f = FlexString(strconv.FormatInt(value, 10))
		return nil
	case bson.TypeInt32:
		var value int32
		if err := bson.UnmarshalValue(t, data, &value); err != nil {
			return fmt.Errorf("FlexString: %w", err)
		}
		*f = FlexString(strconv.FormatInt(int64(value), 10))
		return nil
	case bson.TypeDouble:
		var value float64
		if err := bson.UnmarshalValue(t, data, &value); err != nil {
			return fmt.Errorf("FlexString: %w", err)
		}
		*f = FlexString(strconv.FormatFloat(value, 'f', -1, 64))
		return nil
	case bson.TypeObjectID:
		var value primitive.ObjectID
		if err := bson.UnmarshalValue(t, data, &value); err != nil {
			return fmt.Errorf("FlexString: %w", err)
		}
		*f = FlexString(value.Hex())
		return nil
	default:
		return fmt.Errorf("FlexString: unsupported BSON type %v", t)
	}
}

// mongoPostMapping represents a document in campaign_analytics or label_analytics.
type mongoPostMapping struct {
	CampaignID   string          `bson:"campaign_id,omitempty"`
	LabelID      string          `bson:"label_id,omitempty"`
	PlatformID   FlexString      `bson:"platform_id"`
	Platform     string          `bson:"platform"`
	PlatformType string          `bson:"platform_type"`
	PostedIDs    FlexStringSlice `bson:"posted_ids"`
}

// findResult holds the result of a MongoDB lookup for campaign/label analytics.
type findResult struct {
	MatchedPostedIDs []string            // flat list of all post IDs
	Unmatched        []string            // campaign/label IDs not found in cache
	GroupedPostedIDs map[string][]string // per campaign/label ID → post IDs (for breakdown)
}

func filterMappingsByAccounts(docs []mongoPostMapping, accountIDs []string, includeAll bool) []mongoPostMapping {
	if includeAll || len(accountIDs) == 0 {
		return docs
	}

	accountSet := make(map[string]struct{}, len(accountIDs))
	for _, accountID := range accountIDs {
		accountSet[accountID] = struct{}{}
	}

	filtered := make([]mongoPostMapping, 0, len(docs))
	for _, doc := range docs {
		if _, ok := accountSet[string(doc.PlatformID)]; ok {
			filtered = append(filtered, doc)
		}
	}
	return filtered
}

func isMongoCursorError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "CursorNotFound") ||
		strings.Contains(errStr, "cursor") ||
		strings.Contains(errStr, "Cursor")
}

func (s *CampaignLabelAnalyticsService) findAllWithRetry(ctx context.Context, coll *mongo.Collection, filter interface{}, dest interface{}, op string) error {
	opts := options.Find().SetNoCursorTimeout(true)

	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		cursor, err := coll.Find(ctx, filter, opts)
		if err != nil {
			if !isMongoCursorError(err) {
				return err
			}
			lastErr = err
		} else {
			func() {
				defer cursor.Close(ctx)
				lastErr = cursor.All(ctx, dest)
			}()

			if lastErr == nil {
				return nil
			}
			if !isMongoCursorError(lastErr) {
				return lastErr
			}
		}

		s.logger.Warn().
			Err(lastErr).
			Str("operation", op).
			Int("attempt", attempt).
			Int("max_retries", maxRetries).
			Msg("Mongo cursor error, retrying")

		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
	}

	return lastErr
}

// findCampaignAnalytics queries campaign_analytics for cached post IDs.
// If includeAllAccounts is true, returns all post IDs for the campaigns.
// Otherwise, filters by the provided account IDs.
func (s *CampaignLabelAnalyticsService) findCampaignAnalytics(ctx context.Context, campaigns []string, accountIDs []string, includeAll bool, grouped bool) (*findResult, error) {
	if len(campaigns) == 0 {
		return &findResult{GroupedPostedIDs: make(map[string][]string)}, nil
	}

	// Find all documents for these campaigns (to determine unmatched)
	allFilter := bson.M{"campaign_id": bson.M{"$in": campaigns}}
	var allDocs []mongoPostMapping
	if err := s.findAllWithRetry(ctx, s.campaignAnalytics, allFilter, &allDocs, "findCampaignAnalytics"); err != nil {
		return nil, fmt.Errorf("findCampaignAnalytics decode: %w", err)
	}

	// Determine which campaigns have at least one document
	foundIDs := make(map[string]bool)
	for _, doc := range allDocs {
		foundIDs[doc.CampaignID] = true
	}
	var unmatched []string
	for _, c := range campaigns {
		if !foundIDs[c] {
			unmatched = append(unmatched, c)
		}
	}

	result := &findResult{
		Unmatched:        unmatched,
		GroupedPostedIDs: make(map[string][]string),
	}
	for _, doc := range filterMappingsByAccounts(allDocs, accountIDs, includeAll) {
		result.MatchedPostedIDs = append(result.MatchedPostedIDs, doc.PostedIDs...)
		if grouped {
			result.GroupedPostedIDs[doc.CampaignID] = append(result.GroupedPostedIDs[doc.CampaignID], doc.PostedIDs...)
		}
	}
	return result, nil
}

// findLabelAnalytics queries label_analytics for cached post IDs.
func (s *CampaignLabelAnalyticsService) findLabelAnalytics(ctx context.Context, labels []string, accountIDs []string, includeAll bool, grouped bool) (*findResult, error) {
	if len(labels) == 0 {
		return &findResult{GroupedPostedIDs: make(map[string][]string)}, nil
	}

	allFilter := bson.M{"label_id": bson.M{"$in": labels}}
	var allDocs []mongoPostMapping
	if err := s.findAllWithRetry(ctx, s.labelAnalytics, allFilter, &allDocs, "findLabelAnalytics"); err != nil {
		return nil, fmt.Errorf("findLabelAnalytics decode: %w", err)
	}

	foundIDs := make(map[string]bool)
	for _, doc := range allDocs {
		foundIDs[doc.LabelID] = true
	}
	var unmatched []string
	for _, l := range labels {
		if !foundIDs[l] {
			unmatched = append(unmatched, l)
		}
	}

	result := &findResult{
		Unmatched:        unmatched,
		GroupedPostedIDs: make(map[string][]string),
	}
	for _, doc := range filterMappingsByAccounts(allDocs, accountIDs, includeAll) {
		result.MatchedPostedIDs = append(result.MatchedPostedIDs, doc.PostedIDs...)
		if grouped {
			result.GroupedPostedIDs[doc.LabelID] = append(result.GroupedPostedIDs[doc.LabelID], doc.PostedIDs...)
		}
	}
	return result, nil
}

// postingGroupResult holds grouped postings per platform_id from the posting collection.
type postingGroupResult struct {
	Platform     string   `bson:"platform"`
	PlatformType string   `bson:"platform_type"`
	PlatformID   string   `bson:"platform_id"`
	PostedIDs    []string `bson:"posted_ids"`
}

// resolvePostIDsFromPlans resolves post IDs for unmatched campaigns/labels by querying
// the plans and posting MongoDB collections. This is the "fallback" path.
func (s *CampaignLabelAnalyticsService) resolvePostIDsFromPlans(ctx context.Context, planFilter bson.M, workspaceID string) ([]postingGroupResult, error) {
	// Step 1: Find plan IDs from the plans collection
	var plans []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err := s.findAllWithRetry(ctx, s.plansCollection, planFilter, &plans, "resolvePostIDsFromPlans plans"); err != nil {
		return nil, fmt.Errorf("resolvePostIDsFromPlans plans decode: %w", err)
	}
	if len(plans) == 0 {
		return nil, nil
	}

	// Build a mixed-type list of plan ID values using both ObjectID and string (hex) forms.
	// The posting collection's plan_id field may be stored as a BSON ObjectID or as a plain
	// hex string depending on the dispatch path:
	//   - Cron dispatch (PlanPostingCommand): json_decode(json_encode($plan)) converts _id to string
	//   - API dispatch (PostingController, ApprovalBuilder): plan model passed directly → ObjectID
	// Including both forms ensures we match posting documents regardless of storage format.
	planIDValues := make(bson.A, 0, len(plans)*2)
	for _, p := range plans {
		planIDValues = append(planIDValues, p.ID)       // ObjectID form
		planIDValues = append(planIDValues, p.ID.Hex()) // string (hex) form
	}

	// Step 2: Find postings for those plan IDs
	postFilter := bson.M{
		"workspace_id": workspaceID,
		"plan_id":      bson.M{"$in": planIDValues},
		"posted_id":    bson.M{"$ne": nil},
	}
	var postings []struct {
		PlatformID   FlexString `bson:"platform_id"`
		Platform     string     `bson:"platform"`
		PlatformType string     `bson:"platform_type"`
		PostedID     FlexString `bson:"posted_id"`
	}
	if err := s.findAllWithRetry(ctx, s.postingCollection, postFilter, &postings, "resolvePostIDsFromPlans postings"); err != nil {
		return nil, fmt.Errorf("resolvePostIDsFromPlans postings decode: %w", err)
	}

	// Step 3: Group by platform_id (same as PHP getPostedIdsByPlanIds)
	grouped := make(map[string]*postingGroupResult)
	for _, p := range postings {
		platformID := string(p.PlatformID)
		postedID := string(p.PostedID)
		if postedID == "" {
			continue
		}
		if g, ok := grouped[platformID]; ok {
			g.PostedIDs = append(g.PostedIDs, postedID)
		} else {
			grouped[platformID] = &postingGroupResult{
				Platform:     p.Platform,
				PlatformType: p.PlatformType,
				PlatformID:   platformID,
				PostedIDs:    []string{postedID},
			}
		}
	}

	results := make([]postingGroupResult, 0, len(grouped))
	for _, g := range grouped {
		results = append(results, *g)
	}
	return results, nil
}

// resolveUnmatchedCampaigns resolves post IDs for campaigns not found in campaign_analytics cache.
// For each unmatched campaign: finds plans → postings, creates cache entries, returns post IDs.
func (s *CampaignLabelAnalyticsService) resolveUnmatchedCampaigns(ctx context.Context, unmatched []string, workspaceID string, grouped bool) ([]string, map[string][]string, error) {
	var allPostIDs []string
	groupedIDs := make(map[string][]string)

	for _, campaignID := range unmatched {
		planFilter := bson.M{
			"folderId":     campaignID,
			"workspace_id": workspaceID,
		}
		postings, err := s.resolvePostIDsFromPlans(ctx, planFilter, workspaceID)
		if err != nil {
			s.logger.Warn().Err(err).Str("campaign_id", campaignID).Msg("failed to resolve plans for campaign")
			if grouped {
				groupedIDs[campaignID] = []string{}
			}
			continue
		}
		if len(postings) == 0 {
			if grouped {
				groupedIDs[campaignID] = []string{}
			}
			continue
		}

		for _, pg := range postings {
			// Cache the resolved mapping in campaign_analytics
			doc := bson.M{
				"campaign_id":   campaignID,
				"platform_id":   pg.PlatformID,
				"platform":      pg.Platform,
				"platform_type": pg.PlatformType,
				"posted_ids":    pg.PostedIDs,
			}
			if _, err := s.campaignAnalytics.InsertOne(ctx, doc); err != nil {
				s.logger.Warn().Err(err).Str("campaign_id", campaignID).Msg("failed to cache campaign analytics")
			}

			allPostIDs = append(allPostIDs, pg.PostedIDs...)
			if grouped {
				groupedIDs[campaignID] = append(groupedIDs[campaignID], pg.PostedIDs...)
			}
		}
	}
	return allPostIDs, groupedIDs, nil
}

// resolveUnmatchedLabels resolves post IDs for labels not found in label_analytics cache.
func (s *CampaignLabelAnalyticsService) resolveUnmatchedLabels(ctx context.Context, unmatched []string, workspaceID string, grouped bool) ([]string, map[string][]string, error) {
	var allPostIDs []string
	groupedIDs := make(map[string][]string)

	for _, labelID := range unmatched {
		planFilter := bson.M{
			"labels":       bson.M{"$in": []string{labelID}},
			"workspace_id": workspaceID,
		}
		postings, err := s.resolvePostIDsFromPlans(ctx, planFilter, workspaceID)
		if err != nil {
			s.logger.Warn().Err(err).Str("label_id", labelID).Msg("failed to resolve plans for label")
			if grouped {
				groupedIDs[labelID] = []string{}
			}
			continue
		}
		if len(postings) == 0 {
			if grouped {
				groupedIDs[labelID] = []string{}
			}
			continue
		}

		for _, pg := range postings {
			// Cache the resolved mapping in label_analytics
			doc := bson.M{
				"label_id":      labelID,
				"platform_id":   pg.PlatformID,
				"platform":      pg.Platform,
				"platform_type": pg.PlatformType,
				"posted_ids":    pg.PostedIDs,
			}
			if _, err := s.labelAnalytics.InsertOne(ctx, doc); err != nil {
				s.logger.Warn().Err(err).Str("label_id", labelID).Msg("failed to cache label analytics")
			}

			allPostIDs = append(allPostIDs, pg.PostedIDs...)
			if grouped {
				groupedIDs[labelID] = append(groupedIDs[labelID], pg.PostedIDs...)
			}
		}
	}
	return allPostIDs, groupedIDs, nil
}

// ---- Public Service Methods ----

// SetPostIds resolves campaign/label IDs to their aggregated post IDs.
// This is called internally by other endpoints and can also be used standalone.
func (s *CampaignLabelAnalyticsService) SetPostIds(ctx context.Context, req *types.CampaignLabelRequest) (*types.SetPostIdsResponse, error) {
	accountIDs := req.GetAllAccountIDs()
	includeAll := req.IncludeAllAccounts

	// Resolve campaigns and labels concurrently — they are independent MongoDB lookups
	var campaignPostIDs, labelPostIDs []string
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		campaignResult, err := s.findCampaignAnalytics(egCtx, req.Campaigns, accountIDs, includeAll, false)
		if err != nil {
			return err
		}
		campaignPostIDs = append(campaignPostIDs, campaignResult.MatchedPostedIDs...)

		// Resolve unmatched campaigns via plans → postings fallback
		if len(campaignResult.Unmatched) > 0 {
			extraIDs, _, err := s.resolveUnmatchedCampaigns(egCtx, campaignResult.Unmatched, req.WorkspaceID, false)
			if err != nil {
				s.logger.Error().Err(err).Msg("SetPostIds: failed to resolve unmatched campaigns")
			} else {
				campaignPostIDs = append(campaignPostIDs, extraIDs...)
			}
		}
		return nil
	})

	eg.Go(func() error {
		labelResult, err := s.findLabelAnalytics(egCtx, req.Labels, accountIDs, includeAll, false)
		if err != nil {
			return err
		}
		labelPostIDs = append(labelPostIDs, labelResult.MatchedPostedIDs...)

		// Resolve unmatched labels via plans → postings fallback
		if len(labelResult.Unmatched) > 0 {
			extraIDs, _, err := s.resolveUnmatchedLabels(egCtx, labelResult.Unmatched, req.WorkspaceID, false)
			if err != nil {
				s.logger.Error().Err(err).Msg("SetPostIds: failed to resolve unmatched labels")
			} else {
				labelPostIDs = append(labelPostIDs, extraIDs...)
			}
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	allPostIDs := make([]string, 0, len(campaignPostIDs)+len(labelPostIDs))
	allPostIDs = append(allPostIDs, campaignPostIDs...)
	allPostIDs = append(allPostIDs, labelPostIDs...)

	return &types.SetPostIdsResponse{MatchedPostedIds: allPostIDs}, nil
}

// GetSummaryAnalytics returns current/previous period cross-platform summary with diffs and percentages.
// Flow: resolve post IDs → ClickHouse summary for current period → previous period → compute diffs.
func (s *CampaignLabelAnalyticsService) GetSummaryAnalytics(ctx context.Context, req *types.CampaignLabelRequest) (*types.SummaryResponse, error) {
	// Resolve post IDs
	postIdsResp, err := s.SetPostIds(ctx, req)
	if err != nil {
		return nil, err
	}

	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	flagSetup := req.GetFlagSetup()

	// Query current and previous periods concurrently
	var current, previous *repo.SummaryResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetSummary(egCtx, postIdsResp.MatchedPostedIds, params, flagSetup)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummaryAnalytics: current period failed")
			r = &repo.SummaryResult{}
		}
		current = r
		return nil
	})

	// Build previous period params
	prevParams := s.prevPeriodParams(params)
	eg.Go(func() error {
		r, err := s.repo.GetSummary(egCtx, postIdsResp.MatchedPostedIds, prevParams, flagSetup)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummaryAnalytics: previous period failed")
			r = &repo.SummaryResult{}
		}
		previous = r
		return nil
	})

	_ = eg.Wait()

	resp := &types.SummaryResponse{
		Current: map[string]interface{}{
			"total_posts":                          current.TotalPosts,
			"total_engagement":                     current.TotalEngagement,
			"total_impressions":                    current.TotalImpressions,
			"total_engagement_rate_per_impression": current.TotalEngagementRatePerImpression,
		},
		Previous: map[string]interface{}{
			"total_posts":                          previous.TotalPosts,
			"total_engagement":                     previous.TotalEngagement,
			"total_impressions":                    previous.TotalImpressions,
			"total_engagement_rate_per_impression": previous.TotalEngagementRatePerImpression,
		},
		Difference: make(map[string]interface{}),
		Percentage: make(map[string]interface{}),
	}

	// Compute differences and percentages (same logic as PHP getSummaryDiff)
	computeDiffAndPct(resp, "total_posts", int64(current.TotalPosts), int64(previous.TotalPosts))
	computeDiffAndPct(resp, "total_engagement", int64(current.TotalEngagement), int64(previous.TotalEngagement))
	computeDiffAndPct(resp, "total_impressions", int64(current.TotalImpressions), int64(previous.TotalImpressions))
	computeDiffAndPctFloat(resp, "total_engagement_rate_per_impression", current.TotalEngagementRatePerImpression, previous.TotalEngagementRatePerImpression)

	return resp, nil
}

// GetBreakdownData returns per-campaign/label breakdown data for current and previous periods.
// Groups results by campaign/label ID and era (current/previous).
func (s *CampaignLabelAnalyticsService) GetBreakdownData(ctx context.Context, req *types.CampaignLabelRequest) (map[string]interface{}, error) {
	// Resolve post IDs grouped by campaign/label
	campaignLabelObjects, err := s.resolveGroupedPostIDs(ctx, req)
	if err != nil {
		return nil, err
	}

	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	// Query current and previous periods concurrently
	var currentRows, previousRows []repo.BreakdownResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		rows, err := s.repo.GetBreakdownData(egCtx, campaignLabelObjects, params, "current")
		if err != nil {
			s.logger.Error().Err(err).Msg("GetBreakdownData: current period failed")
			rows = []repo.BreakdownResult{}
		}
		currentRows = rows
		return nil
	})

	eg.Go(func() error {
		rows, err := s.repo.GetBreakdownData(egCtx, campaignLabelObjects, params, "previous")
		if err != nil {
			s.logger.Error().Err(err).Msg("GetBreakdownData: previous period failed")
			rows = []repo.BreakdownResult{}
		}
		previousRows = rows
		return nil
	})

	_ = eg.Wait()

	// Group by id → era (same structure as PHP collect($res)->groupBy(['id','era']))
	allRows := append(currentRows, previousRows...)
	grouped := make(map[string]interface{})
	for _, row := range allRows {
		idGroup, ok := grouped[row.ID]
		if !ok {
			idGroup = make(map[string]interface{})
			grouped[row.ID] = idGroup
		}
		eraMap := idGroup.(map[string]interface{})
		eraMap[row.Era] = []types.BreakdownRow{{
			ID:               row.ID,
			Era:              row.Era,
			TotalPosts:       row.TotalPosts,
			TotalEngagement:  row.TotalEngagement,
			TotalImpressions: row.TotalImpressions,
		}}
	}

	return grouped, nil
}

// GetInsightsBreakdown returns time-series insights data grouped by campaign/label ID.
func (s *CampaignLabelAnalyticsService) GetInsightsBreakdown(ctx context.Context, req *types.CampaignLabelRequest) (map[string]interface{}, error) {
	campaignLabelObjects, err := s.resolveGroupedPostIDs(ctx, req)
	if err != nil {
		return nil, err
	}

	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.GetInsightsData(ctx, campaignLabelObjects, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetInsightsBreakdown: failed")
		rows = []repo.InsightsResult{}
	}

	// Group by id (same as PHP collect($res)->groupBy('id'))
	grouped := make(map[string]interface{})
	for _, row := range rows {
		grouped[row.ID] = []types.InsightsRow{{
			ID:               row.ID,
			TotalEngagement:  row.TotalEngagement,
			TotalImpressions: row.TotalImpressions,
			TotalPosts:       row.TotalPosts,
			CreatedAt:        row.CreatedAt,
		}}
	}

	return grouped, nil
}

// GetPlannerAnalytics returns detailed per-post analytics for a single platform.
// Used by the planner view for individual post performance.
func (s *CampaignLabelAnalyticsService) GetPlannerAnalytics(ctx context.Context, req *types.PlannerAnalyticsRequest) (map[string]interface{}, error) {
	result, err := s.repo.GetPlannerAnalytics(ctx, req.AllPostIDs, req.Platforms)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetPlannerAnalytics: failed")
		return map[string]interface{}{}, nil
	}
	if result == nil {
		return nil, nil
	}
	normalized, _ := roundFloatValues(result).(map[string]interface{})
	return normalized, nil
}

// resolveGroupedPostIDs resolves post IDs grouped by campaign/label ID.
// Used by breakdown and insights endpoints that need per-entity post ID mapping.
func (s *CampaignLabelAnalyticsService) resolveGroupedPostIDs(ctx context.Context, req *types.CampaignLabelRequest) (map[string][]string, error) {
	accountIDs := req.GetAllAccountIDs()
	includeAll := req.IncludeAllAccounts

	// Resolve campaigns and labels concurrently — they are independent MongoDB lookups
	campaignGrouped := make(map[string][]string)
	labelGrouped := make(map[string][]string)
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		campaignResult, err := s.findCampaignAnalytics(egCtx, req.Campaigns, accountIDs, includeAll, true)
		if err != nil {
			return err
		}
		for id, ids := range campaignResult.GroupedPostedIDs {
			campaignGrouped[id] = ids
		}

		// Resolve unmatched campaigns
		if len(campaignResult.Unmatched) > 0 {
			_, groupedIDs, err := s.resolveUnmatchedCampaigns(egCtx, campaignResult.Unmatched, req.WorkspaceID, true)
			if err != nil {
				s.logger.Error().Err(err).Msg("resolveGroupedPostIDs: failed campaigns")
			} else {
				for id, ids := range groupedIDs {
					campaignGrouped[id] = ids
				}
			}
		}
		return nil
	})

	eg.Go(func() error {
		labelResult, err := s.findLabelAnalytics(egCtx, req.Labels, accountIDs, includeAll, true)
		if err != nil {
			return err
		}
		for id, ids := range labelResult.GroupedPostedIDs {
			labelGrouped[id] = ids
		}

		// Resolve unmatched labels
		if len(labelResult.Unmatched) > 0 {
			_, groupedIDs, err := s.resolveUnmatchedLabels(egCtx, labelResult.Unmatched, req.WorkspaceID, true)
			if err != nil {
				s.logger.Error().Err(err).Msg("resolveGroupedPostIDs: failed labels")
			} else {
				for id, ids := range groupedIDs {
					labelGrouped[id] = ids
				}
			}
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Merge campaign and label results
	result := make(map[string][]string, len(campaignGrouped)+len(labelGrouped))
	for id, ids := range campaignGrouped {
		result[id] = ids
	}
	for id, ids := range labelGrouped {
		result[id] = ids
	}

	return result, nil
}

// prevPeriodParams builds query params pointing at the previous comparison period.
func (s *CampaignLabelAnalyticsService) prevPeriodParams(params *clickhouse.QueryParams) *clickhouse.QueryParams {
	prev := *params
	prev.DateFrom = params.PrevDateFrom
	prev.DateTo = params.PrevDateTo
	return &prev
}

// computeDiffAndPct computes the difference and percentage change between two int values.
func computeDiffAndPct(resp *types.SummaryResponse, key string, current, previous int64) {
	resp.Difference[key] = current - previous
	if previous == 0 {
		if current == 0 {
			resp.Percentage[key] = float64(0)
		} else {
			resp.Percentage[key] = float64(100)
		}
	} else {
		resp.Percentage[key] = math.Round(float64(current-previous)*10000/float64(previous)) / 100
	}
}

// computeDiffAndPctFloat computes the difference and percentage change between two float values.
func computeDiffAndPctFloat(resp *types.SummaryResponse, key string, current, previous float64) {
	resp.Difference[key] = math.Round((current-previous)*100) / 100
	if previous == 0 {
		if current == 0 {
			resp.Percentage[key] = float64(0)
		} else {
			resp.Percentage[key] = float64(100)
		}
	} else {
		resp.Percentage[key] = math.Round((current-previous)*10000/previous) / 100
	}
}

// unused but kept for reference: prevDate computes the previous date range
func prevDate(startDate, endDate time.Time) (time.Time, time.Time) {
	diff := endDate.Sub(startDate)
	if diff == 0 {
		diff = 24 * time.Hour
	}
	prevEnd := startDate
	prevStart := prevEnd.Add(-diff)
	return prevStart, prevEnd
}
