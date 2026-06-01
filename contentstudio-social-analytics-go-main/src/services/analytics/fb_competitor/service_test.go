package fb_competitor

import (
"context"
"errors"
"io"
"testing"

"github.com/rs/zerolog"

chRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/fb_competitor"
types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/fb_competitor"
mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// --- Stubs ---

type stubClickHouseRepo struct {
getDataTableMetricsFn              func(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
getPostingActivityGraphByTypesFn   func(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
getPostingActivityBySpecificTypeFn func(ctx context.Context, params *chRepo.CompetitorQueryParams, mediaType, sortOrder string) ([]map[string]interface{}, error)
getTopAndLeastPerformingPostsFn    func(ctx context.Context, params *chRepo.CompetitorQueryParams) ([]map[string]interface{}, error)
getTopHashtagsFn                   func(ctx context.Context, params *chRepo.CompetitorQueryParams, limit int) ([]map[string]interface{}, error)
getIndividualHashtagDataFn         func(ctx context.Context, params *chRepo.CompetitorQueryParams, hashtag string) ([]map[string]interface{}, error)
getBiographyDataFn                 func(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
getFollowersGrowthComparisonFn     func(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
getPostReactDistributionFn         func(ctx context.Context, params *chRepo.CompetitorQueryParams, facebookID, sortOrder string) ([]map[string]interface{}, error)
getPostReactDistByCompanyFn        func(ctx context.Context, params *chRepo.CompetitorQueryParams, facebookID, sortOrder string) ([]map[string]interface{}, error)
getPostTypeDistributionFn          func(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
getPostEngagementOverTimeFn        func(ctx context.Context, params *chRepo.CompetitorQueryParams, facebookID, sortOrder string) ([]map[string]interface{}, error)
getPostEngagementByCompetitorFn    func(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
}

func (s *stubClickHouseRepo) GetDataTableMetrics(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
if s.getDataTableMetricsFn != nil {
return s.getDataTableMetricsFn(ctx, params, sortOrder)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetPostingActivityGraphByTypes(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
if s.getPostingActivityGraphByTypesFn != nil {
return s.getPostingActivityGraphByTypesFn(ctx, params, sortOrder)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetPostingActivityBySpecificType(ctx context.Context, params *chRepo.CompetitorQueryParams, mediaType, sortOrder string) ([]map[string]interface{}, error) {
if s.getPostingActivityBySpecificTypeFn != nil {
return s.getPostingActivityBySpecificTypeFn(ctx, params, mediaType, sortOrder)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetTopAndLeastPerformingPosts(ctx context.Context, params *chRepo.CompetitorQueryParams) ([]map[string]interface{}, error) {
if s.getTopAndLeastPerformingPostsFn != nil {
return s.getTopAndLeastPerformingPostsFn(ctx, params)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetTopHashtags(ctx context.Context, params *chRepo.CompetitorQueryParams, limit int) ([]map[string]interface{}, error) {
if s.getTopHashtagsFn != nil {
return s.getTopHashtagsFn(ctx, params, limit)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetIndividualHashtagData(ctx context.Context, params *chRepo.CompetitorQueryParams, hashtag string) ([]map[string]interface{}, error) {
if s.getIndividualHashtagDataFn != nil {
return s.getIndividualHashtagDataFn(ctx, params, hashtag)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetBiographyData(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
if s.getBiographyDataFn != nil {
return s.getBiographyDataFn(ctx, params, sortOrder)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetFollowersGrowthComparison(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
if s.getFollowersGrowthComparisonFn != nil {
return s.getFollowersGrowthComparisonFn(ctx, params, sortOrder)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetPostReactDistribution(ctx context.Context, params *chRepo.CompetitorQueryParams, facebookID, sortOrder string) ([]map[string]interface{}, error) {
if s.getPostReactDistributionFn != nil {
return s.getPostReactDistributionFn(ctx, params, facebookID, sortOrder)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetPostReactDistributionByCompany(ctx context.Context, params *chRepo.CompetitorQueryParams, facebookID, sortOrder string) ([]map[string]interface{}, error) {
if s.getPostReactDistByCompanyFn != nil {
return s.getPostReactDistByCompanyFn(ctx, params, facebookID, sortOrder)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetPostTypeDistribution(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
if s.getPostTypeDistributionFn != nil {
return s.getPostTypeDistributionFn(ctx, params, sortOrder)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetPostEngagementOverTime(ctx context.Context, params *chRepo.CompetitorQueryParams, facebookID, sortOrder string) ([]map[string]interface{}, error) {
if s.getPostEngagementOverTimeFn != nil {
return s.getPostEngagementOverTimeFn(ctx, params, facebookID, sortOrder)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetPostEngagementByCompetitor(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
if s.getPostEngagementByCompetitorFn != nil {
return s.getPostEngagementByCompetitorFn(ctx, params, sortOrder)
}
return []map[string]interface{}{}, nil
}

type stubMongoRepo struct {
competitors map[string]*mongoModels.Competitor
err         error
}

func (s *stubMongoRepo) GetReportCompetitors(_ context.Context, _ string) (map[string]*mongoModels.Competitor, error) {
if s.err != nil {
return nil, s.err
}
return s.competitors, nil
}

func newTestService(ch ClickHouseRepo, mongo CompetitorRepo) *FacebookCompetitorService {
return NewFacebookCompetitorService(ch, mongo, zerolog.New(io.Discard))
}

func validRequest() *types.CompetitorRequest {
return &types.CompetitorRequest{
ReportID:  "report123",
StartDate: "2025-01-01",
EndDate:   "2025-01-31",
Timezone:  "UTC",
}
}

func defaultMongo() *stubMongoRepo {
return &stubMongoRepo{
competitors: map[string]*mongoModels.Competitor{
"fb_123": {Name: "Page One", Slug: "page-one", State: "active", Image: "img1.jpg"},
"fb_456": {Name: "Page Two", Slug: "page-two", State: "active", Image: "img2.jpg"},
},
}
}

// --- Tests ---

func TestNewFacebookCompetitorService(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
if svc == nil {
t.Fatal("expected non-nil service")
}
}

func TestResolveParams_MongoError(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, &stubMongoRepo{err: errors.New("mongo down")})
_, err := svc.GetPostingActivityGraphByTypes(context.Background(), validRequest())
if err == nil {
t.Fatal("expected error when mongo fails, got nil")
}
}

func TestGetDataTableMetrics_Success(t *testing.T) {
chStub := &stubClickHouseRepo{
getDataTableMetricsFn: func(_ context.Context, _ *chRepo.CompetitorQueryParams, _ string) ([]map[string]interface{}, error) {
return []map[string]interface{}{
{"facebook_id": "fb_123", "followersCount": float64(1000), "engagementRate": float64(2.5), "averagePostsPerWeek": float64(3)},
}, nil
},
}
svc := newTestService(chStub, defaultMongo())
result, err := svc.GetDataTableMetrics(context.Background(), validRequest())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m, ok := result.(map[string]interface{})
if !ok {
t.Fatalf("expected map result, got %T", result)
}
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key in result")
}
if _, exists := m["data_prev"]; !exists {
t.Fatal("expected 'data_prev' key in result")
}
if _, exists := m["data_table_metrics"]; !exists {
t.Fatal("expected 'data_table_metrics' key in result")
}
metrics := m["data_table_metrics"].([]map[string]interface{})
if len(metrics) != 1 {
t.Fatalf("expected 1 metric row, got %d", len(metrics))
}
if _, exists := metrics[0]["followersCountDiff"]; !exists {
t.Fatal("expected followersCountDiff in metrics")
}
}

func TestGetDataTableMetrics_CHError(t *testing.T) {
chStub := &stubClickHouseRepo{
getDataTableMetricsFn: func(_ context.Context, _ *chRepo.CompetitorQueryParams, _ string) ([]map[string]interface{}, error) {
return nil, errors.New("ch failed")
},
}
svc := newTestService(chStub, defaultMongo())
_, err := svc.GetDataTableMetrics(context.Background(), validRequest())
if err == nil {
t.Fatal("expected error, got nil")
}
}

func TestGetPostingActivityGraphByTypes_Success(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
result, err := svc.GetPostingActivityGraphByTypes(context.Background(), validRequest())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key")
}
}

func TestGetPostingActivityBySpecificType_Success(t *testing.T) {
req := validRequest()
req.MediaType = "video"
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
result, err := svc.GetPostingActivityBySpecificType(context.Background(), req)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key")
}
}

func TestGetTopAndLeastPerformingPosts_EmptyRows(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
result, err := svc.GetTopAndLeastPerformingPosts(context.Background(), validRequest())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
data := m["data"].([]map[string]interface{})
if len(data) != 0 {
t.Fatalf("expected 0 competitors for empty rows, got %d", len(data))
}
}

func TestGetTopAndLeastPerformingPosts_CHError(t *testing.T) {
chStub := &stubClickHouseRepo{
getTopAndLeastPerformingPostsFn: func(_ context.Context, _ *chRepo.CompetitorQueryParams) ([]map[string]interface{}, error) {
return nil, errors.New("ch failed")
},
}
svc := newTestService(chStub, defaultMongo())
_, err := svc.GetTopAndLeastPerformingPosts(context.Background(), validRequest())
if err == nil {
t.Fatal("expected error, got nil")
}
}

func TestGetTopHashtags_Success(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
result, err := svc.GetTopHashtags(context.Background(), validRequest())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key")
}
}

func TestGetIndividualHashtagData_Success(t *testing.T) {
req := validRequest()
req.Hashtag = "marketing"
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
result, err := svc.GetIndividualHashtagData(context.Background(), req)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key")
}
}

func TestGetBiographyData_Success(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
result, err := svc.GetBiographyData(context.Background(), validRequest())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key")
}
}

func TestGetFollowersGrowthComparison_Success(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
result, err := svc.GetFollowersGrowthComparison(context.Background(), validRequest())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key")
}
}

func TestGetPostReactDistribution_RequiresFacebookID(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
req := validRequest()
_, err := svc.GetPostReactDistribution(context.Background(), req)
if err == nil {
t.Fatal("expected error when facebook_id is missing")
}
}

func TestGetPostReactDistribution_Success(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
req := validRequest()
req.FacebookID = "fb_123"
result, err := svc.GetPostReactDistribution(context.Background(), req)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key")
}
}

func TestGetPostReactDistributionByCompany_RequiresFacebookID(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
_, err := svc.GetPostReactDistributionByCompany(context.Background(), validRequest())
if err == nil {
t.Fatal("expected error when facebook_id is missing")
}
}

func TestGetPostReactDistributionByCompany_Success(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
req := validRequest()
req.FacebookID = "fb_123"
result, err := svc.GetPostReactDistributionByCompany(context.Background(), req)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key")
}
}

func TestGetPostTypeDistribution_Success(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
result, err := svc.GetPostTypeDistribution(context.Background(), validRequest())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key")
}
}

func TestGetPostEngagementOverTime_RequiresFacebookID(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
_, err := svc.GetPostEngagementOverTime(context.Background(), validRequest())
if err == nil {
t.Fatal("expected error when facebook_id is missing")
}
}

func TestGetPostEngagementOverTime_Success(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
req := validRequest()
req.FacebookID = "fb_123"
result, err := svc.GetPostEngagementOverTime(context.Background(), req)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key")
}
}

func TestGetPostEngagementByCompetitor_Success(t *testing.T) {
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
result, err := svc.GetPostEngagementByCompetitor(context.Background(), validRequest())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key")
}
}

func TestCalculateGrowth(t *testing.T) {
tests := []struct {
name     string
current  float64
previous float64
expected interface{}
}{
{name: "zero_previous", current: 10, previous: 0, expected: "N/A"},
{name: "growth", current: 20, previous: 10, expected: "100.00"},
{name: "decline", current: 5, previous: 10, expected: "-50.00"},
}
for _, tc := range tests {
t.Run(tc.name, func(t *testing.T) {
got := calculateGrowth(tc.current, tc.previous)
if got != tc.expected {
t.Fatalf("expected %v, got %v", tc.expected, got)
}
})
}
}

func TestToFloat64(t *testing.T) {
tests := []struct {
input    interface{}
expected float64
}{
{float64(1.5), 1.5},
{float32(2.5), 2.5},
{int64(3), 3.0},
{int32(4), 4.0},
{int(5), 5.0},
{uint64(6), 6.0},
{"not a number", 0},
}
for _, tc := range tests {
got := toFloat64(tc.input)
if got != tc.expected {
t.Fatalf("toFloat64(%v) = %f, want %f", tc.input, got, tc.expected)
}
}
}

func TestTransformTopAndLeastPosts_EmptyInput(t *testing.T) {
result := transformTopAndLeastPosts(nil)
if len(result) != 0 {
t.Fatalf("expected empty result for nil input, got %d", len(result))
}
result = transformTopAndLeastPosts([]map[string]interface{}{})
if len(result) != 0 {
t.Fatalf("expected empty result for empty input, got %d", len(result))
}
}

func TestGroupPostsByCategory(t *testing.T) {
rows := []map[string]interface{}{
{"category": "top_5_posts", "post_id": "p1"},
{"category": "least_5_posts", "post_id": "p2"},
{"category": "top_5_posts", "post_id": "p3"},
}
order, byPost := groupPostsByCategory(rows, "top_5_posts")
if len(order) != 2 {
t.Fatalf("expected 2 top posts, got %d", len(order))
}
if _, exists := byPost["p1"]; !exists {
t.Fatal("expected p1 in top posts")
}
if _, exists := byPost["p2"]; exists {
t.Fatal("p2 should not be in top posts")
}
}

func TestBuildPostsList_NilMediaSkipped(t *testing.T) {
byPost := map[string][]map[string]interface{}{
"p1": {
{"post_id": "p1", "media_id": nil, "post_engagement": 10},
},
}
posts := buildPostsList(byPost, []string{"p1"})
if len(posts) != 1 {
t.Fatalf("expected 1 post, got %d", len(posts))
}
media := posts[0]["media"].([]map[string]interface{})
if len(media) != 0 {
t.Fatalf("expected 0 media for nil media_id, got %d", len(media))
}
}
