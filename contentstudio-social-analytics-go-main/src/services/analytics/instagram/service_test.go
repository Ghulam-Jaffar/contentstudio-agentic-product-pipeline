package instagram

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/instagram"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/instagram"
)

// --- Mock ClickHouse infrastructure ---

type mockRow struct{}

func (m *mockRow) Err() error { return nil }
func (m *mockRow) Scan(dest ...any) error {
	for _, d := range dest {
		switch v := d.(type) {
		case *int32:
			*v = 0
		case *int64:
			*v = 0
		case *float64:
			*v = 0
		case *float32:
			*v = 0
		case *string:
			*v = ""
		case *uint8:
			*v = 0
		}
	}
	return nil
}
func (m *mockRow) ScanStruct(_ any) error { return nil }

type mockRows struct{}

func (m *mockRows) Next() bool                       { return false }
func (m *mockRows) Scan(dest ...any) error           { return nil }
func (m *mockRows) ScanStruct(_ any) error           { return nil }
func (m *mockRows) ColumnTypes() []driver.ColumnType { return nil }
func (m *mockRows) Totals(dest ...any) error         { return nil }
func (m *mockRows) Columns() []string                { return nil }
func (m *mockRows) Close() error                     { return nil }
func (m *mockRows) Err() error                       { return nil }

type mockBatch struct{}

func (m *mockBatch) Abort() error                        { return nil }
func (m *mockBatch) Append(_ ...any) error               { return nil }
func (m *mockBatch) AppendStruct(_ any) error            { return nil }
func (m *mockBatch) Column(_ int) driver.BatchColumn     { return nil }
func (m *mockBatch) Columns() []column.Interface         { return nil }
func (m *mockBatch) Flush() error                        { return nil }
func (m *mockBatch) Send() error                         { return nil }
func (m *mockBatch) IsSent() bool                        { return false }
func (m *mockBatch) Rows() int                           { return 0 }
func (m *mockBatch) Close() error                        { return nil }

type mockConn struct{}

func (m *mockConn) Contributors() []string                        { return nil }
func (m *mockConn) ServerVersion() (*driver.ServerVersion, error) { return nil, nil }
func (m *mockConn) Select(_ context.Context, _ any, _ string, _ ...any) error {
	return nil
}
func (m *mockConn) Query(_ context.Context, _ string, _ ...any) (driver.Rows, error) {
	return &mockRows{}, nil
}
func (m *mockConn) QueryRow(_ context.Context, _ string, _ ...any) driver.Row {
	return &mockRow{}
}
func (m *mockConn) PrepareBatch(_ context.Context, _ string, _ ...driver.PrepareBatchOption) (driver.Batch, error) {
	return &mockBatch{}, nil
}
func (m *mockConn) Exec(_ context.Context, _ string, _ ...any) error              { return nil }
func (m *mockConn) AsyncInsert(_ context.Context, _ string, _ bool, _ ...any) error { return nil }
func (m *mockConn) Ping(_ context.Context) error                                   { return nil }
func (m *mockConn) Stats() driver.Stats                                            { return driver.Stats{} }
func (m *mockConn) Close() error                                                   { return nil }

func newTestClient(conn *mockConn) *ch.Client {
	return &ch.Client{
		Conn:   conn,
		Config: config.ClickHouseConfig{Database: "test"},
		Logger: zerolog.New(io.Discard),
	}
}

func newTestService() *InstagramAnalyticsService {
	client := newTestClient(&mockConn{})
	r := repo.NewRepository(client)
	return NewInstagramAnalyticsService(r, zerolog.Nop())
}

func validRequest() *types.InstagramRequest {
	return &types.InstagramRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		StartDate:   "2025-01-01",
		EndDate:     "2025-01-31",
		Timezone:    "UTC",
	}
}

// --- GetSummary ---

func TestGetSummary_InvalidRequest(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetSummary(context.Background(), &types.InstagramRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestGetSummary_ValidRequest(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetSummary(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
	if _, ok := resp.Overview["current"]; !ok {
		t.Fatal("expected 'current' key in overview")
	}
	if _, ok := resp.Overview["previous"]; !ok {
		t.Fatal("expected 'previous' key in overview")
	}
}

// --- GetAudienceGrowth ---

func TestGetAudienceGrowth_InvalidRequest(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetAudienceGrowth(context.Background(), &types.InstagramRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestGetAudienceGrowth_ValidRequest(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetAudienceGrowth(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
	if resp.AudienceGrowth == nil {
		t.Fatal("expected non-nil AudienceGrowth")
	}
	if _, ok := resp.AudienceGrowthRollup["current"]; !ok {
		t.Fatal("expected 'current' key in rollup")
	}
	if _, ok := resp.AudienceGrowthRollup["previous"]; !ok {
		t.Fatal("expected 'previous' key in rollup")
	}
}

// --- GetPublishingBehaviour ---

func TestGetPublishingBehaviour_InvalidRequest(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetPublishingBehaviour(context.Background(), &types.PublishingBehaviourRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestGetPublishingBehaviour_ValidRequest(t *testing.T) {
	svc := newTestService()
	req := &types.PublishingBehaviourRequest{
		InstagramRequest: *validRequest(),
		MediaType:        []string{"REELS", "IMAGE"},
	}
	resp, err := svc.GetPublishingBehaviour(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
	if resp.PublishingBehaviour == nil {
		t.Fatal("expected non-nil PublishingBehaviour")
	}
	if _, ok := resp.PublishingBehaviourRollup["current"]; !ok {
		t.Fatal("expected 'current' key in rollup")
	}
}

func TestGetPublishingBehaviour_DefaultMediaTypes(t *testing.T) {
	svc := newTestService()
	req := &types.PublishingBehaviourRequest{InstagramRequest: *validRequest()}
	resp, err := svc.GetPublishingBehaviour(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
}

// --- GetTopPosts ---

func TestGetTopPosts_InvalidRequest(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetTopPosts(context.Background(), &types.TopPostsRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestGetTopPosts_ValidRequest(t *testing.T) {
	svc := newTestService()
	req := &types.TopPostsRequest{
		InstagramRequest: *validRequest(),
		Limit:            10,
	}
	resp, err := svc.GetTopPosts(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
	if resp.TopPosts == nil {
		t.Fatal("expected non-nil TopPosts slice")
	}
}

func TestGetTopPosts_WithHashtags(t *testing.T) {
	svc := newTestService()
	req := &types.TopPostsRequest{
		InstagramRequest: *validRequest(),
		Limit:            5,
		Hashtags:         []string{"#go", "#analytics"},
	}
	resp, err := svc.GetTopPosts(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
}

// --- GetActiveUsers ---

func TestGetActiveUsers_InvalidRequest(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetActiveUsers(context.Background(), &types.InstagramRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestGetActiveUsers_ValidRequest(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetActiveUsers(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
	if resp.ActiveUsersHours == nil {
		t.Fatal("expected non-nil ActiveUsersHours")
	}
	if resp.ActiveUsersDays == nil {
		t.Fatal("expected non-nil ActiveUsersDays")
	}
}

// --- GetImpressions ---

func TestGetImpressions_InvalidRequest(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetImpressions(context.Background(), &types.InstagramRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestGetImpressions_ValidRequest(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetImpressions(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
	if resp.Impressions == nil {
		t.Fatal("expected non-nil Impressions")
	}
	if _, ok := resp.ImpressionsRollup["current"]; !ok {
		t.Fatal("expected 'current' key in rollup")
	}
}

// --- GetEngagement ---

func TestGetEngagement_InvalidRequest(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetEngagement(context.Background(), &types.InstagramRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestGetEngagement_ValidRequest(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetEngagement(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
	if resp.Engagements == nil {
		t.Fatal("expected non-nil Engagements")
	}
	if _, ok := resp.EngagementsRollup["current"]; !ok {
		t.Fatal("expected 'current' key in rollup")
	}
}

// --- GetHashtags ---

func TestGetHashtags_InvalidRequest(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetHashtags(context.Background(), &types.InstagramRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestGetHashtags_ValidRequest(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetHashtags(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
	if resp.TopHashtags == nil {
		t.Fatal("expected non-nil TopHashtags")
	}
	if _, ok := resp.TopHashtagsRollup["current"]; !ok {
		t.Fatal("expected 'current' key in rollup")
	}
}

// --- GetStoriesPerformance ---

func TestGetStoriesPerformance_InvalidRequest(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetStoriesPerformance(context.Background(), &types.InstagramRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestGetStoriesPerformance_ValidRequest(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetStoriesPerformance(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
	if resp.StoriesPerformance == nil {
		t.Fatal("expected non-nil StoriesPerformance")
	}
	if _, ok := resp.StoriesRollup["current"]; !ok {
		t.Fatal("expected 'current' key in rollup")
	}
}

// --- GetReelsPerformance ---

func TestGetReelsPerformance_InvalidRequest(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetReelsPerformance(context.Background(), &types.InstagramRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestGetReelsPerformance_ValidRequest(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetReelsPerformance(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
	if resp.Reels == nil {
		t.Fatal("expected non-nil Reels")
	}
	if _, ok := resp.ReelsRollup["current"]; !ok {
		t.Fatal("expected 'current' key in rollup")
	}
}

// --- GetDemographicsAge ---

func TestGetDemographicsAge_InvalidRequest(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetDemographicsAge(context.Background(), &types.InstagramRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestGetDemographicsAge_ValidRequest(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetDemographicsAge(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AudienceAge == nil {
		t.Fatal("expected non-nil AudienceAge map")
	}
	if resp.AudienceGender == nil {
		t.Fatal("expected non-nil AudienceGender map")
	}
}

// --- GetCountryCity ---

func TestGetCountryCity_InvalidRequest(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetCountryCity(context.Background(), &types.InstagramRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestGetCountryCity_ValidRequest(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetCountryCity(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AudienceCity == nil {
		t.Fatal("expected non-nil AudienceCity map")
	}
	if resp.AudienceCountry == nil {
		t.Fatal("expected non-nil AudienceCountry map")
	}
}

// --- Helper function unit tests ---

func TestComputeDailyDelta_Empty(t *testing.T) {
	result := computeDailyDelta([]int32{})
	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %v", result)
	}
}

func TestComputeDailyDelta_SingleElement(t *testing.T) {
	result := computeDailyDelta([]int32{100})
	if len(result) != 1 {
		t.Fatalf("expected length 1, got %d", len(result))
	}
	if result[0] != 0 {
		t.Fatalf("expected delta[0]=0, got %d", result[0])
	}
}

func TestComputeDailyDelta_MultipleElements(t *testing.T) {
	input := []int32{100, 110, 105, 120}
	result := computeDailyDelta(input)
	expected := []int32{0, 10, -5, 15}
	if len(result) != len(expected) {
		t.Fatalf("expected length %d, got %d", len(expected), len(result))
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Fatalf("delta[%d]: expected %d, got %d", i, expected[i], result[i])
		}
	}
}

func TestComputeDailyDelta_ConstantValues(t *testing.T) {
	input := []int32{50, 50, 50}
	result := computeDailyDelta(input)
	for i, v := range result {
		if i > 0 && v != 0 {
			t.Fatalf("expected 0 delta for constant followers at index %d, got %d", i, v)
		}
	}
}

func TestParseDemographicArray_Empty(t *testing.T) {
	result := parseDemographicArray(nil)
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %v", result)
	}
}

func TestParseDemographicArray_JSONFormat(t *testing.T) {
	entries := []string{
		`{"key":"18-24","value":150}`,
		`{"key":"25-34","value":200}`,
	}
	result := parseDemographicArray(entries)
	if result["18-24"] != 150 {
		t.Fatalf("expected 18-24=150, got %d", result["18-24"])
	}
	if result["25-34"] != 200 {
		t.Fatalf("expected 25-34=200, got %d", result["25-34"])
	}
}

func TestParseDemographicArray_ColonFormat(t *testing.T) {
	entries := []string{
		"US:500",
		"UK:200",
	}
	result := parseDemographicArray(entries)
	if result["US"] != 500 {
		t.Fatalf("expected US=500, got %d", result["US"])
	}
	if result["UK"] != 200 {
		t.Fatalf("expected UK=200, got %d", result["UK"])
	}
}

func TestParseDemographicArray_MixedFormats(t *testing.T) {
	entries := []string{
		`{"key":"F.18-24","value":300}`,
		"M.25-34:400",
		"",
	}
	result := parseDemographicArray(entries)
	if result["F.18-24"] != 300 {
		t.Fatalf("expected F.18-24=300, got %d", result["F.18-24"])
	}
	if result["M.25-34"] != 400 {
		t.Fatalf("expected M.25-34=400, got %d", result["M.25-34"])
	}
}

func TestParseDemographicArray_DuplicateKeys(t *testing.T) {
	entries := []string{
		`{"key":"US","value":100}`,
		"US:200",
	}
	result := parseDemographicArray(entries)
	if result["US"] != 300 {
		t.Fatalf("expected US=300 (summed), got %d", result["US"])
	}
}

func TestFindMaxAudienceAge_Nil(t *testing.T) {
	result := findMaxAudienceAge(nil)
	if result != nil {
		t.Fatalf("expected nil for empty map, got %v", result)
	}
}

func TestFindMaxAudienceAge_Empty(t *testing.T) {
	result := findMaxAudienceAge(map[string]int64{})
	if result != nil {
		t.Fatalf("expected nil for empty map, got %v", result)
	}
}

func TestFindMaxAudienceAge_FindsMax(t *testing.T) {
	genderAge := map[string]int64{
		"F.18-24": 300,
		"M.25-34": 500,
		"F.35-44": 100,
	}
	result := findMaxAudienceAge(genderAge)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Gender != "M" {
		t.Fatalf("expected gender=M, got %q", result.Gender)
	}
	if result.Age != "25-34" {
		t.Fatalf("expected age=25-34, got %q", result.Age)
	}
	if result.Value != 500 {
		t.Fatalf("expected value=500, got %d", result.Value)
	}
}

func TestFindMaxAudienceAge_KeyWithoutDot(t *testing.T) {
	genderAge := map[string]int64{
		"unknown": 999,
	}
	result := findMaxAudienceAge(genderAge)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Gender != "" {
		t.Fatalf("expected empty gender, got %q", result.Gender)
	}
	if result.Age != "unknown" {
		t.Fatalf("expected age=unknown, got %q", result.Age)
	}
}

// --- prevPeriodParams ---

func TestPrevPeriodParams(t *testing.T) {
	params := &ch.QueryParams{
		AccountIDs:   []string{"ig_123"},
		DateFrom:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTo:       time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		PrevDateFrom: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		PrevDateTo:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		Timezone:     "UTC",
		DayCount:     31,
	}
	prev := prevPeriodParams(params)
	if !prev.DateFrom.Equal(params.PrevDateFrom) {
		t.Fatalf("expected DateFrom=%v, got %v", params.PrevDateFrom, prev.DateFrom)
	}
	if !prev.DateTo.Equal(params.PrevDateTo) {
		t.Fatalf("expected DateTo=%v, got %v", params.PrevDateTo, prev.DateTo)
	}
	if len(prev.AccountIDs) != len(params.AccountIDs) {
		t.Fatalf("expected AccountIDs preserved")
	}
}
