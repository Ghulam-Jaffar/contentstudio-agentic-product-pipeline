package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

type captureConn struct {
	*mockConn
	lastQuery     string
	lastQueryArgs []any
	lastExecQuery string
	lastExecArgs  []any
	execCount     int
	execQueries   []string
	execArgs      [][]any
}

func (c *captureConn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	c.lastQuery = query
	c.lastQueryArgs = append([]any(nil), args...)
	if c.mockConn == nil {
		return nil, nil
	}
	return c.mockConn.Query(ctx, query, args...)
}

func (c *captureConn) Exec(ctx context.Context, query string, args ...any) error {
	c.lastExecQuery = query
	c.lastExecArgs = append([]any(nil), args...)
	c.execCount++
	c.execQueries = append(c.execQueries, query)
	c.execArgs = append(c.execArgs, append([]any(nil), args...))
	if c.mockConn == nil {
		return nil
	}
	return c.mockConn.Exec(ctx, query, args...)
}

func newCaptureClient(conn *captureConn) *Client {
	return &Client{
		Conn:   conn,
		Config: config.ClickHouseConfig{Database: "test_db"},
		Logger: testLogger(),
	}
}

func Test_GetMinimalOlderThan20DaysByPage_UsesTenDayWindow(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{queryRows: &mockRows{nextCount: 0}}}
	client := newCaptureClient(conn)

	_, err := client.GetMinimalOlderThan20DaysByPage(context.Background(), "", "page_123", 500, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "INTERVAL 10 DAY") {
		t.Fatalf("expected 10 day interval, got query: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "FROM facebook_posts") {
		t.Fatalf("expected default table in query, got: %s", conn.lastQuery)
	}
}

func Test_GetMinimalInstagramOlderThan20DaysByAccount_DefaultTable(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{queryRows: &mockRows{nextCount: 0}}}
	client := newCaptureClient(conn)

	posts, err := client.GetMinimalInstagramOlderThan20DaysByAccount(context.Background(), "", "ig_123", 500, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 0 {
		t.Fatalf("expected empty result, got %d", len(posts))
	}
	if !strings.Contains(conn.lastQuery, "FROM instagram_posts") {
		t.Fatalf("expected default table in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "url_refreshed_at < now() - INTERVAL 10 DAY") {
		t.Fatalf("expected url_refreshed_at 10 day window in query, got: %s", conn.lastQuery)
	}
}

func Test_GetMinimalInstagramOlderThan20DaysByAccount_EmptyID(t *testing.T) {
	client := newCaptureClient(&captureConn{mockConn: &mockConn{queryRows: &mockRows{nextCount: 0}}})

	_, err := client.GetMinimalInstagramOlderThan20DaysByAccount(context.Background(), "instagram_posts", "", 500, 0)
	if err == nil {
		t.Fatal("expected error for empty instagramID")
	}
}

func Test_UpdateInstagramMediaURLs_DeduplicatesAndFilters(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{}}
	client := newCaptureClient(conn)

	rows := []clickhousemodels.InstagramMinimalPost{
		{InstagramID: "ig_123", MediaID: "media_1", MediaURL: []string{"https://example.com/media_1.jpg"}, VideoURL: []string{"https://example.com/media_1.mp4"}},
		{InstagramID: "ig_123", MediaID: "media_1", MediaURL: []string{"https://example.com/duplicate.jpg"}},
		{InstagramID: "other", MediaID: "media_2", MediaURL: []string{"https://example.com/skip.jpg"}},
		{InstagramID: "ig_123", MediaID: "", MediaURL: []string{"https://example.com/skip-empty-id.jpg"}},
		{InstagramID: "ig_123", MediaID: "media_2", MediaURL: []string{"https://example.com/media_2.jpg"}},
	}

	count, err := client.UpdateInstagramMediaURLs(context.Background(), "", "ig_123", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 updated rows, got %d", count)
	}
	if !strings.Contains(conn.lastExecQuery, "ALTER TABLE instagram_posts") {
		t.Fatalf("expected default table in update query, got: %s", conn.lastExecQuery)
	}
	if !strings.Contains(conn.lastExecQuery, "has(CAST(? AS Array(String)), media_id)") {
		t.Fatalf("expected has() WHERE guard in update query, got: %s", conn.lastExecQuery)
	}
	if got := conn.lastExecArgs[0].([]string); len(got) != 2 || got[0] != "media_1" || got[1] != "media_2" {
		t.Fatalf("unexpected media ids: %#v", got)
	}
	if got := conn.lastExecArgs[1].([][]string); len(got) != 2 || len(got[0]) != 1 || got[0][0] != "https://example.com/media_1.jpg" || got[1][0] != "https://example.com/media_2.jpg" {
		t.Fatalf("unexpected media urls: %#v", got)
	}
	if got := conn.lastExecArgs[4].([][]string); len(got) != 2 || len(got[0]) != 1 || got[0][0] != "https://example.com/media_1.mp4" || len(got[1]) != 0 {
		t.Fatalf("unexpected video urls: %#v", got)
	}
	if got := conn.lastExecArgs[6].(string); got != "ig_123" {
		t.Fatalf("unexpected instagramID arg: %s", got)
	}
}

func Test_UpdateInstagramMediaURLs_ExecError(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{execErr: errors.New("exec failed")}}
	client := newCaptureClient(conn)

	rows := []clickhousemodels.InstagramMinimalPost{
		{InstagramID: "ig_123", MediaID: "media_1", MediaURL: []string{"https://example.com/media_1.jpg"}},
	}

	_, err := client.UpdateInstagramMediaURLs(context.Background(), "instagram_posts", "ig_123", rows)
	if err == nil {
		t.Fatal("expected error from exec")
	}
}

func Test_UpdateFullPictures_DeduplicatesAndUsesBoundTimestamp(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{}}
	client := newCaptureClient(conn)

	rows := []clickhousemodels.MinimalPost{
		{PageID: "page_123", PostID: "post_1", FullPicture: "https://example.com/post_1.jpg"},
		{PageID: "page_123", PostID: "post_1", FullPicture: "https://example.com/duplicate.jpg"},
		{PageID: "other", PostID: "post_2", FullPicture: "https://example.com/skip.jpg"},
		{PageID: "page_123", PostID: "", FullPicture: "https://example.com/skip-empty-id.jpg"},
		{PageID: "page_123", PostID: "post_2", FullPicture: "https://example.com/post_2.jpg"},
	}

	count, err := client.UpdateFullPictures(context.Background(), "", "page_123", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 updated rows, got %d", count)
	}
	if !strings.Contains(conn.lastExecQuery, "ALTER TABLE facebook_posts") {
		t.Fatalf("expected default table in update query, got: %s", conn.lastExecQuery)
	}
	if !strings.Contains(conn.lastExecQuery, "updated_time = toDateTime64(") {
		t.Fatalf("expected bound timestamp in update query, got: %s", conn.lastExecQuery)
	}
	if got := conn.lastExecArgs[0].([]string); len(got) != 2 || got[0] != "post_1" || got[1] != "post_2" {
		t.Fatalf("unexpected post ids: %#v", got)
	}
	if got := conn.lastExecArgs[1].([]string); len(got) != 2 || got[0] != "https://example.com/post_1.jpg" || got[1] != "https://example.com/post_2.jpg" {
		t.Fatalf("unexpected urls: %#v", got)
	}
	if got := conn.lastExecArgs[2].(string); got == "" {
		t.Fatal("expected bound timestamp argument")
	}
	if got := conn.lastExecArgs[3].(string); got != "page_123" {
		t.Fatalf("unexpected pageID arg: %s", got)
	}
}

func Test_UpdateFacebookCompetitorMediaAssetURLs_DeduplicatesAndUsesBoundTimestamp(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{}}
	client := newCaptureClient(conn)

	rows := []clickhousemodels.FacebookCompetitorMinimalMediaAsset{
		{PageID: "page_123", MediaID: "media_1", Link: "https://example.com/media_1.jpg"},
		{PageID: "page_123", MediaID: "media_1", Link: "https://example.com/duplicate.jpg"},
		{PageID: "other", MediaID: "media_2", Link: "https://example.com/skip.jpg"},
		{PageID: "page_123", MediaID: "", Link: "https://example.com/skip-empty-id.jpg"},
		{PageID: "page_123", MediaID: "media_2", Link: "https://example.com/media_2.jpg"},
	}

	count, err := client.UpdateFacebookCompetitorMediaAssetURLs(context.Background(), "", "page_123", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 updated rows, got %d", count)
	}
	if !strings.Contains(conn.lastExecQuery, "ALTER TABLE facebook_competitor_media_assets") {
		t.Fatalf("expected default table in update query, got: %s", conn.lastExecQuery)
	}
	if !strings.Contains(conn.lastExecQuery, "WHERE page_id = ?") || !strings.Contains(conn.lastExecQuery, "has(CAST(? AS Array(String)), media_id)") {
		t.Fatalf("expected media-id scoped WHERE clause in update query, got: %s", conn.lastExecQuery)
	}
	if got := conn.lastExecArgs[0].([]string); len(got) != 2 || got[0] != "media_1" || got[1] != "media_2" {
		t.Fatalf("unexpected media ids: %#v", got)
	}
	if got := conn.lastExecArgs[1].([]string); len(got) != 2 || got[0] != "https://example.com/media_1.jpg" || got[1] != "https://example.com/media_2.jpg" {
		t.Fatalf("unexpected links: %#v", got)
	}
	if got := conn.lastExecArgs[2].(string); got == "" {
		t.Fatal("expected bound timestamp argument")
	}
	if got := conn.lastExecArgs[3].(string); got != "page_123" {
		t.Fatalf("unexpected pageID arg: %s", got)
	}
	if got := conn.lastExecArgs[4].([]string); len(got) != 2 || got[0] != "media_1" || got[1] != "media_2" {
		t.Fatalf("unexpected WHERE media ids arg: %#v", got)
	}
}

func Test_UpdateFacebookCompetitorMediaAssetURLs_BatchesLargeUpdates(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{}}
	client := newCaptureClient(conn)

	rows := make([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, 0, 101)
	for i := 0; i < 101; i++ {
		rows = append(rows, clickhousemodels.FacebookCompetitorMinimalMediaAsset{
			PageID:  "page_123",
			MediaID: fmt.Sprintf("media_%03d", i),
			Link:    fmt.Sprintf("https://example.com/media_%03d.jpg", i),
		})
	}

	count, err := client.UpdateFacebookCompetitorMediaAssetURLs(context.Background(), "", "page_123", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 101 {
		t.Fatalf("expected 101 updated rows, got %d", count)
	}
	if conn.execCount != 3 {
		t.Fatalf("expected 3 batched exec calls, got %d", conn.execCount)
	}
	if got := conn.execArgs[0][0].([]string); len(got) != competitorURLUpdateBatchSize {
		t.Fatalf("expected first batch size %d, got %d", competitorURLUpdateBatchSize, len(got))
	}
	if got := conn.execArgs[2][0].([]string); len(got) != 1 {
		t.Fatalf("expected final batch size 1, got %d", len(got))
	}
}

func Test_UpdateFacebookCompetitorSharedPictures_DeduplicatesAndUsesBoundTimestamp(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{}}
	client := newCaptureClient(conn)

	rows := []clickhousemodels.FacebookCompetitorMinimalSharedPost{
		{FacebookID: "page_123", PostID: "post_1", SharedFromPic: "https://example.com/post_1.jpg"},
		{FacebookID: "page_123", PostID: "post_1", SharedFromPic: "https://example.com/duplicate.jpg"},
		{FacebookID: "other", PostID: "post_2", SharedFromPic: "https://example.com/skip.jpg"},
		{FacebookID: "page_123", PostID: "", SharedFromPic: "https://example.com/skip-empty-id.jpg"},
		{FacebookID: "page_123", PostID: "post_2", SharedFromPic: "https://example.com/post_2.jpg"},
	}

	count, err := client.UpdateFacebookCompetitorSharedPictures(context.Background(), "", "page_123", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 updated rows, got %d", count)
	}
	if !strings.Contains(conn.lastExecQuery, "ALTER TABLE facebook_competitor_posts") {
		t.Fatalf("expected default table in update query, got: %s", conn.lastExecQuery)
	}
	if !strings.Contains(conn.lastExecQuery, "WHERE facebook_id = ?") || !strings.Contains(conn.lastExecQuery, "has(CAST(? AS Array(String)), post_id)") {
		t.Fatalf("expected post-id scoped WHERE clause in update query, got: %s", conn.lastExecQuery)
	}
	if got := conn.lastExecArgs[0].([]string); len(got) != 2 || got[0] != "post_1" || got[1] != "post_2" {
		t.Fatalf("unexpected post ids: %#v", got)
	}
	if got := conn.lastExecArgs[1].([]string); len(got) != 2 || got[0] != "https://example.com/post_1.jpg" || got[1] != "https://example.com/post_2.jpg" {
		t.Fatalf("unexpected shared pics: %#v", got)
	}
	if got := conn.lastExecArgs[2].(string); got == "" {
		t.Fatal("expected bound timestamp argument")
	}
	if got := conn.lastExecArgs[3].(string); got != "page_123" {
		t.Fatalf("unexpected facebookID arg: %s", got)
	}
	if got := conn.lastExecArgs[4].([]string); len(got) != 2 || got[0] != "post_1" || got[1] != "post_2" {
		t.Fatalf("unexpected WHERE post ids arg: %#v", got)
	}
}

func Test_UpdateFacebookCompetitorSharedPictures_BatchesLargeUpdates(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{}}
	client := newCaptureClient(conn)

	rows := make([]clickhousemodels.FacebookCompetitorMinimalSharedPost, 0, 101)
	for i := 0; i < 101; i++ {
		rows = append(rows, clickhousemodels.FacebookCompetitorMinimalSharedPost{
			FacebookID:    "page_123",
			PostID:        fmt.Sprintf("post_%03d", i),
			SharedFromPic: fmt.Sprintf("https://example.com/post_%03d.jpg", i),
		})
	}

	count, err := client.UpdateFacebookCompetitorSharedPictures(context.Background(), "", "page_123", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 101 {
		t.Fatalf("expected 101 updated rows, got %d", count)
	}
	if conn.execCount != 3 {
		t.Fatalf("expected 3 batched exec calls, got %d", conn.execCount)
	}
	if got := conn.execArgs[0][0].([]string); len(got) != competitorURLUpdateBatchSize {
		t.Fatalf("expected first batch size %d, got %d", competitorURLUpdateBatchSize, len(got))
	}
	if got := conn.execArgs[2][0].([]string); len(got) != 1 {
		t.Fatalf("expected final batch size 1, got %d", len(got))
	}
}

func Test_GetMinimalLinkedInOlderThan7DaysByAccount_DefaultTable(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{queryRows: &mockRows{nextCount: 0}}}
	client := newCaptureClient(conn)

	posts, err := client.GetMinimalLinkedInOlderThan7DaysByAccount(context.Background(), "", "li_123", 500, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 0 {
		t.Fatalf("expected empty result, got %d", len(posts))
	}
	if !strings.Contains(conn.lastQuery, "FROM linkedin_posts") {
		t.Fatalf("expected default table in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "url_refreshed_at < now() - INTERVAL 10 DAY") {
		t.Fatalf("expected url_refreshed_at 10 day window in query, got: %s", conn.lastQuery)
	}
}

func Test_GetMinimalLinkedInOlderThan7DaysByAccount_EmptyID(t *testing.T) {
	client := newCaptureClient(&captureConn{mockConn: &mockConn{queryRows: &mockRows{nextCount: 0}}})

	_, err := client.GetMinimalLinkedInOlderThan7DaysByAccount(context.Background(), "linkedin_posts", "", 500, 0)
	if err == nil {
		t.Fatal("expected error for empty linkedinID")
	}
}

func Test_UpdateLinkedInPostURLs_DeduplicatesAndFilters(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{}}
	client := newCaptureClient(conn)

	rows := []clickhousemodels.LinkedInMinimalPost{
		{LinkedinID: "li_123", PostID: "post_1", Image: "https://example.com/post_1.jpg", Media: []string{"https://example.com/post_1.jpg"}},
		{LinkedinID: "li_123", PostID: "post_1", Image: "https://example.com/duplicate.jpg", Media: []string{"https://example.com/duplicate.jpg"}},
		{LinkedinID: "other", PostID: "post_2", Image: "https://example.com/skip.jpg"},
		{LinkedinID: "li_123", PostID: "", Image: "https://example.com/skip-empty-id.jpg"},
		{LinkedinID: "li_123", PostID: "post_2", Image: "https://example.com/post_2.jpg", Media: []string{"https://example.com/post_2.jpg", "https://example.com/post_2b.jpg"}},
	}

	count, err := client.UpdateLinkedInPostURLs(context.Background(), "", "li_123", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 updated rows, got %d", count)
	}
	if !strings.Contains(conn.lastExecQuery, "ALTER TABLE linkedin_posts") {
		t.Fatalf("expected default table in update query, got: %s", conn.lastExecQuery)
	}
	if got := conn.lastExecArgs[0].([]string); len(got) != 2 || got[0] != "post_1" || got[1] != "post_2" {
		t.Fatalf("unexpected post ids: %#v", got)
	}
	if got := conn.lastExecArgs[1].([]string); len(got) != 2 || got[0] != "https://example.com/post_1.jpg" || got[1] != "https://example.com/post_2.jpg" {
		t.Fatalf("unexpected images: %#v", got)
	}
	if got := conn.lastExecArgs[3].([][]string); len(got) != 2 || got[0][0] != "https://example.com/post_1.jpg" || got[1][1] != "https://example.com/post_2b.jpg" {
		t.Fatalf("unexpected media arrays: %#v", got)
	}
	if got := conn.lastExecArgs[5].(string); got != "li_123" {
		t.Fatalf("unexpected linkedinID arg: %s", got)
	}
}

func Test_UpdateLinkedInPostURLs_ExecError(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{execErr: errors.New("exec failed")}}
	client := newCaptureClient(conn)

	rows := []clickhousemodels.LinkedInMinimalPost{
		{LinkedinID: "li_123", PostID: "post_1", Image: "https://example.com/post_1.jpg"},
	}

	_, err := client.UpdateLinkedInPostURLs(context.Background(), "linkedin_posts", "li_123", rows)
	if err == nil {
		t.Fatal("expected error from exec")
	}
}

func Test_UpdateInstagramCompetitorMediaURLs_DeduplicatesAndUsesBoundTimestamp(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{}}
	client := newCaptureClient(conn)

	rows := []clickhousemodels.InstagramCompetitorMinimalPost{
		{InstagramID: 123, PostID: "post_1", MediaURL: "https://example.com/post_1.jpg", ProfilePictureURL: "https://example.com/profile_1.jpg"},
		{InstagramID: 123, PostID: "post_1", MediaURL: "https://example.com/duplicate.jpg"},
		{InstagramID: 456, PostID: "post_2", MediaURL: "https://example.com/skip.jpg"},
		{InstagramID: 123, PostID: "", MediaURL: "https://example.com/skip-empty-id.jpg"},
		{InstagramID: 123, PostID: "post_2", MediaURL: "https://example.com/post_2.jpg"},
	}

	count, err := client.UpdateInstagramCompetitorMediaURLs(context.Background(), "", 123, "https://example.com/profile_123.jpg", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 updated rows, got %d", count)
	}
	// Two exec calls: one for media_url+inserted_at, one for profile_picture_url.
	if conn.execCount != 2 {
		t.Fatalf("expected 2 exec calls (media + profile_pic), got %d", conn.execCount)
	}

	// --- First exec: media_url + inserted_at (has() guard only, no OR broadcast) ---
	mediaQuery := conn.execQueries[0]
	if !strings.Contains(mediaQuery, "ALTER TABLE instagram_competitor_posts") {
		t.Fatalf("expected default table in media update query, got: %s", mediaQuery)
	}
	if !strings.Contains(mediaQuery, "has(CAST(? AS Array(String)), post_id)") {
		t.Fatalf("expected has() WHERE guard in media update, got: %s", mediaQuery)
	}
	if strings.Contains(mediaQuery, "OR length") {
		t.Fatalf("media update must not have OR length() broadcast, got: %s", mediaQuery)
	}
	// args: postIDs, mediaURLs, updatedAt, instagramID, instagramID, postIDs
	mediaArgs := conn.execArgs[0]
	if got := mediaArgs[0].([]string); len(got) != 2 || got[0] != "post_1" || got[1] != "post_2" {
		t.Fatalf("unexpected post ids: %#v", got)
	}
	if got := mediaArgs[1].([]string); len(got) != 2 || got[0] != "https://example.com/post_1.jpg" || got[1] != "https://example.com/post_2.jpg" {
		t.Fatalf("unexpected media urls: %#v", got)
	}
	if got := mediaArgs[2].(string); got == "" {
		t.Fatal("expected bound timestamp argument")
	}
	if got := mediaArgs[3].(int64); got != 123 {
		t.Fatalf("unexpected instagramID arg: %d", got)
	}
	if got := mediaArgs[4].(int64); got != 123 {
		t.Fatalf("unexpected business_account_id instagramID arg: %d", got)
	}
	if got := mediaArgs[5].([]string); len(got) != 2 || got[0] != "post_1" || got[1] != "post_2" {
		t.Fatalf("unexpected WHERE post ids arg: %#v", got)
	}

	// --- Second exec: profile_picture_url only (no inserted_at change) ---
	picQuery := conn.execQueries[1]
	if !strings.Contains(picQuery, "profile_picture_url") {
		t.Fatalf("expected profile_picture_url in second query, got: %s", picQuery)
	}
	if strings.Contains(picQuery, "inserted_at") {
		t.Fatalf("profile picture update must NOT change inserted_at, got: %s", picQuery)
	}
	// args: profilePictureURL, instagramID, instagramID
	picArgs := conn.execArgs[1]
	if got := picArgs[0].(string); got != "https://example.com/profile_123.jpg" {
		t.Fatalf("unexpected profile picture arg: %s", got)
	}
	if got := picArgs[1].(int64); got != 123 {
		t.Fatalf("unexpected instagramID arg in profile pic update: %d", got)
	}
	if got := picArgs[2].(int64); got != 123 {
		t.Fatalf("unexpected business_account_id instagramID arg in profile pic update: %d", got)
	}
}

func Test_UpdateInstagramCompetitorMediaURLs_BatchesLargeUpdates(t *testing.T) {
	conn := &captureConn{mockConn: &mockConn{}}
	client := newCaptureClient(conn)

	rows := make([]clickhousemodels.InstagramCompetitorMinimalPost, 0, 101)
	for i := 0; i < 101; i++ {
		rows = append(rows, clickhousemodels.InstagramCompetitorMinimalPost{
			InstagramID:       123,
			PostID:            fmt.Sprintf("post_%03d", i),
			MediaURL:          fmt.Sprintf("https://example.com/post_%03d.jpg", i),
			ProfilePictureURL: "https://example.com/profile_123.jpg",
		})
	}

	count, err := client.UpdateInstagramCompetitorMediaURLs(context.Background(), "", 123, "https://example.com/profile_123.jpg", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 101 {
		t.Fatalf("expected 101 updated rows, got %d", count)
	}
	// 3 media URL batches (50+50+1) + 1 profile picture exec = 4 total
	if conn.execCount != 4 {
		t.Fatalf("expected 4 exec calls (3 media batches + 1 profile pic), got %d", conn.execCount)
	}
	if got := conn.execArgs[0][0].([]string); len(got) != competitorURLUpdateBatchSize {
		t.Fatalf("expected first batch size %d, got %d", competitorURLUpdateBatchSize, len(got))
	}
	if got := conn.execArgs[2][0].([]string); len(got) != 1 {
		t.Fatalf("expected final batch size 1, got %d", len(got))
	}
	// Last exec is the profile picture update with no postIDs array
	picArgs := conn.execArgs[3]
	if got := picArgs[0].(string); got != "https://example.com/profile_123.jpg" {
		t.Fatalf("unexpected profile picture arg: %s", got)
	}
	if got := picArgs[1].(int64); got != 123 {
		t.Fatalf("unexpected instagramID arg in profile pic: %d", got)
	}
}
