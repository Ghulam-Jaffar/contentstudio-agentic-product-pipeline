package ig_competitor

import (
"context"
"errors"
"io"
"testing"

"github.com/rs/zerolog"

chRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/ig_competitor"
types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/ig_competitor"
mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// --- Stubs ---

type stubClickHouseRepo struct {
getDataTableMetricsFn              func(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
getPostingActivityGraphByTypesFn   func(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
getPostingActivityBySpecificTypeFn func(ctx context.Context, params *chRepo.CompetitorQueryParams, mediaType, mediaProductType, sortOrder string) ([]map[string]interface{}, error)
getPostingActivityTableByTypeFn    func(ctx context.Context, params *chRepo.CompetitorQueryParams, mediaType, mediaProductType, sortOrder string) ([]map[string]interface{}, error)
getFollowersGrowthComparisonFn     func(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
getTopAndLeastPerformingPostsFn    func(ctx context.Context, params *chRepo.CompetitorQueryParams) ([]map[string]interface{}, error)
getTopHashtagsFn                   func(ctx context.Context, params *chRepo.CompetitorQueryParams, limit int) ([]map[string]interface{}, error)
getIndividualHashtagDataFn         func(ctx context.Context, params *chRepo.CompetitorQueryParams, hashtag string) ([]map[string]interface{}, error)
getBiographyDataFn                 func(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
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
func (s *stubClickHouseRepo) GetPostingActivityBySpecificType(ctx context.Context, params *chRepo.CompetitorQueryParams, mediaType, mediaProductType, sortOrder string) ([]map[string]interface{}, error) {
if s.getPostingActivityBySpecificTypeFn != nil {
return s.getPostingActivityBySpecificTypeFn(ctx, params, mediaType, mediaProductType, sortOrder)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetPostingActivityTableByType(ctx context.Context, params *chRepo.CompetitorQueryParams, mediaType, mediaProductType, sortOrder string) ([]map[string]interface{}, error) {
if s.getPostingActivityTableByTypeFn != nil {
return s.getPostingActivityTableByTypeFn(ctx, params, mediaType, mediaProductType, sortOrder)
}
return []map[string]interface{}{}, nil
}
func (s *stubClickHouseRepo) GetFollowersGrowthComparison(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
if s.getFollowersGrowthComparisonFn != nil {
return s.getFollowersGrowthComparisonFn(ctx, params, sortOrder)
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

func newTestService(ch ClickHouseRepo, mongo CompetitorRepo) *InstagramCompetitorService {
return NewInstagramCompetitorService(ch, mongo, zerolog.New(io.Discard))
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
"ig_123": {Name: "Account One", Slug: "account-one", State: "active", Image: "img1.jpg"},
"ig_456": {Name: "Account Two", Slug: "account-two", State: "active", Image: "img2.jpg"},
},
}
}

// --- Tests ---

func TestNewInstagramCompetitorService(t *testing.T) {
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
{"business_account_id": "ig_123", "followersCount": float64(1000), "followingCount": float64(500), "engagementRate": float64(2.5), "averagePostsPerWeek": float64(3)},
}, nil
},
}
svc := newTestService(chStub, defaultMongo())
result, err := svc.GetDataTableMetrics(context.Background(), validRequest())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key")
}
if _, exists := m["data_prev"]; !exists {
t.Fatal("expected 'data_prev' key")
}
if _, exists := m["data_table_metrics"]; !exists {
t.Fatal("expected 'data_table_metrics' key")
}
metrics := m["data_table_metrics"].([]map[string]interface{})
if len(metrics) != 1 {
t.Fatalf("expected 1 metric row, got %d", len(metrics))
}
if _, exists := metrics[0]["followersCountDiff"]; !exists {
t.Fatal("expected followersCountDiff in metrics")
}
if _, exists := metrics[0]["followingCountDiff"]; !exists {
t.Fatal("expected followingCountDiff in metrics")
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
req.MediaType = "VIDEO"
req.MediaProductType = "REELS"
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

func TestGetPostingActivityTableByType_Success(t *testing.T) {
req := validRequest()
req.MediaType = "IMAGE"
svc := newTestService(&stubClickHouseRepo{}, defaultMongo())
result, err := svc.GetPostingActivityTableByType(context.Background(), req)
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

func TestGetTopAndLeastPerformingPosts_Success(t *testing.T) {
chStub := &stubClickHouseRepo{
getTopAndLeastPerformingPostsFn: func(_ context.Context, _ *chRepo.CompetitorQueryParams) ([]map[string]interface{}, error) {
return []map[string]interface{}{
{"business_account_id": "ig_123", "top_5_posts": []interface{}{}, "least_5_posts": []interface{}{}},
}, nil
},
}
svc := newTestService(chStub, defaultMongo())
result, err := svc.GetTopAndLeastPerformingPosts(context.Background(), validRequest())
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m := result.(map[string]interface{})
if _, exists := m["data"]; !exists {
t.Fatal("expected 'data' key")
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
req.Hashtag = "travel"
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
{name: "no_change", current: 10, previous: 10, expected: "0.00"},
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
{nil, 0},
}
for _, tc := range tests {
got := toFloat64(tc.input)
if got != tc.expected {
t.Fatalf("toFloat64(%v) = %f, want %f", tc.input, got, tc.expected)
}
}
}
