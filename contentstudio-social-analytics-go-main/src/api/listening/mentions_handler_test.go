package listening

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	chDB "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	chModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	mentionsSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/mentions"
)

// stubMentionReader is a test double for mentionsSvc.MentionReader.
// Used by mentions_handler_test.go, analytics_handler_test.go, and router_test.go.
type stubMentionReader struct {
	queryMentionsRows   []chModels.ListeningMentionRow
	queryMentionsCursor string
	queryMentionsErr    error

	countUnreadVal int
	countUnreadErr error

	getMentionRow *chModels.ListeningMentionRow
	getMentionErr error

	updateMentionErr error

	markAllReadCount int
	markAllReadErr   error

	getAnalyticsData *chDB.AnalyticsData
	getAnalyticsErr  error
}

func (s *stubMentionReader) QueryMentions(_ context.Context, _ *chDB.MentionFilter) ([]chModels.ListeningMentionRow, string, error) {
	return s.queryMentionsRows, s.queryMentionsCursor, s.queryMentionsErr
}

func (s *stubMentionReader) CountUnread(_ context.Context, _ *chDB.MentionFilter) (int, error) {
	return s.countUnreadVal, s.countUnreadErr
}

func (s *stubMentionReader) GetMention(_ context.Context, _, _ string) (*chModels.ListeningMentionRow, error) {
	return s.getMentionRow, s.getMentionErr
}

func (s *stubMentionReader) UpdateMention(_ context.Context, _ chModels.ListeningMentionRow) error {
	return s.updateMentionErr
}

func (s *stubMentionReader) MarkAllRead(_ context.Context, _ *chDB.MentionFilter) (int, error) {
	return s.markAllReadCount, s.markAllReadErr
}

func (s *stubMentionReader) GetAnalytics(_ context.Context, _ *chDB.MentionFilter) (*chDB.AnalyticsData, error) {
	return s.getAnalyticsData, s.getAnalyticsErr
}

// newMentionsHandler builds a MentionsHandler with stub dependencies.
// The resolver requires topic_ids[] in the request URL to avoid workspace topic lookup,
// or a configured stubTopicWorkspaceResolver for workspace-based tests.
func newMentionsHandler(reader *stubMentionReader) *MentionsHandler {
	svc := mentionsSvc.NewService(reader, zerolog.Nop())
	resolver := NewMentionFilterResolver(
		stubViewResolver{},
		stubTopicWorkspaceResolver{},
	)
	return NewMentionsHandler(svc, resolver, zerolog.Nop())
}

func bodyOf(v interface{}) *bytes.Reader {
	if v == nil {
		return bytes.NewReader(nil)
	}
	if s, ok := v.(string); ok {
		return bytes.NewReader([]byte(s))
	}
	b, _ := json.Marshal(v)
	return bytes.NewReader(b)
}

func TestQueryArray(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		values map[string][]string
		key    string
		want   []string
	}{
		{
			name:   "single value",
			values: map[string][]string{"ids[]": {"a"}},
			key:    "ids[]",
			want:   []string{"a"},
		},
		{
			name:   "multiple repeated values",
			values: map[string][]string{"ids[]": {"a", "b", "c"}},
			key:    "ids[]",
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "comma-separated in single value",
			values: map[string][]string{"ids[]": {"a,b,c"}},
			key:    "ids[]",
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "mixed repeated and comma-separated",
			values: map[string][]string{"ids[]": {"a,b", "c"}},
			key:    "ids[]",
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "trims whitespace around comma-parts",
			values: map[string][]string{"ids[]": {"a, b , c"}},
			key:    "ids[]",
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "absent key returns nil",
			values: map[string][]string{},
			key:    "ids[]",
			want:   nil,
		},
		{
			name:   "skips empty parts from consecutive commas",
			values: map[string][]string{"ids[]": {"a,,b"}},
			key:    "ids[]",
			want:   []string{"a", "b"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := queryArray(tc.values, tc.key)
			assertStringSlice(t, "result", got, tc.want)
		})
	}
}

func TestParseMentionFilter(t *testing.T) {
	t.Parallel()

	boolTrue := true
	boolFalse := false

	tests := []struct {
		name        string
		queryString string
		check       func(t *testing.T, got *apiModels.MentionFilter)
		wantErrMsg  string
	}{
		{
			name:        "empty query returns zero-value filter",
			queryString: "",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if got.WorkspaceID != "" || got.ViewID != "" || got.Limit != 0 {
					t.Errorf("want zero filter, got %+v", got)
				}
			},
		},
		{
			name:        "parses workspace_id",
			queryString: "workspace_id=ws-1",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if got.WorkspaceID != "ws-1" {
					t.Errorf("workspace_id: want %q, got %q", "ws-1", got.WorkspaceID)
				}
			},
		},
		{
			name:        "parses topic_ids array",
			queryString: "topic_ids[]=t1&topic_ids[]=t2",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				assertStringSlice(t, "topic_ids", got.TopicIDs, []string{"t1", "t2"})
			},
		},
		{
			name:        "parses platforms array",
			queryString: "platforms[]=twitter&platforms[]=reddit",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				assertStringSlice(t, "platforms", got.Platforms, []string{"twitter", "reddit"})
			},
		},
		{
			name:        "parses sentiments array",
			queryString: "sentiments[]=positive&sentiments[]=neutral",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				assertStringSlice(t, "sentiments", got.Sentiments, []string{"positive", "neutral"})
			},
		},
		{
			name:        "parses ai_tags array",
			queryString: "ai_tags[]=buy_intent",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				assertStringSlice(t, "ai_tags", got.AITags, []string{"buy_intent"})
			},
		},
		{
			name:        "parses language array",
			queryString: "language[]=en&language[]=fr",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				assertStringSlice(t, "language", got.Language, []string{"en", "fr"})
			},
		},
		{
			name:        "parses limit",
			queryString: "limit=50",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if got.Limit != 50 {
					t.Errorf("limit: want 50, got %d", got.Limit)
				}
			},
		},
		{
			name:        "parses min_followers",
			queryString: "min_followers=1000",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if got.MinFollowers != 1000 {
					t.Errorf("min_followers: want 1000, got %d", got.MinFollowers)
				}
			},
		},
		{
			name:        "parses min_total_engagement",
			queryString: "min_total_engagement=250",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if got.MinTotalEngagement != 250 {
					t.Errorf("min_total_engagement: want 250, got %d", got.MinTotalEngagement)
				}
			},
		},
		{
			name:        "parses is_bookmarked=true",
			queryString: "is_bookmarked=true",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if got.IsBookmarked == nil || *got.IsBookmarked != boolTrue {
					t.Errorf("is_bookmarked: want true, got %v", got.IsBookmarked)
				}
			},
		},
		{
			name:        "parses is_bookmarked=false",
			queryString: "is_bookmarked=false",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if got.IsBookmarked == nil || *got.IsBookmarked != boolFalse {
					t.Errorf("is_bookmarked: want false, got %v", got.IsBookmarked)
				}
			},
		},
		{
			name:        "parses is_read=true",
			queryString: "is_read=true",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if got.IsRead == nil || *got.IsRead != boolTrue {
					t.Errorf("is_read: want true, got %v", got.IsRead)
				}
			},
		},
		{
			name:        "parses include_irrelevant=true",
			queryString: "include_irrelevant=true",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if !got.IncludeIrrelevant {
					t.Errorf("include_irrelevant: want true, got false")
				}
			},
		},
		{
			name:        "parses include_irrelevant=1",
			queryString: "include_irrelevant=1",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if !got.IncludeIrrelevant {
					t.Errorf("include_irrelevant=1: want true, got false")
				}
			},
		},
		{
			name:        "ignores unrecognized include_irrelevant value",
			queryString: "include_irrelevant=yes",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if got.IncludeIrrelevant {
					t.Errorf("include_irrelevant=yes: want false, got true")
				}
			},
		},
		{
			name:        "parses date range",
			queryString: "date_from=2024-01-01&date_to=2024-01-31",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if got.DateFrom != "2024-01-01" {
					t.Errorf("date_from: want %q, got %q", "2024-01-01", got.DateFrom)
				}
				if got.DateTo != "2024-01-31" {
					t.Errorf("date_to: want %q, got %q", "2024-01-31", got.DateTo)
				}
			},
		},
		{
			name:        "parses view_id and cursor",
			queryString: "view_id=v1&cursor=tok123",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if got.ViewID != "v1" {
					t.Errorf("view_id: want %q, got %q", "v1", got.ViewID)
				}
				if got.Cursor != "tok123" {
					t.Errorf("cursor: want %q, got %q", "tok123", got.Cursor)
				}
			},
		},
		{
			name:        "parses search",
			queryString: "search=hello+world",
			check: func(t *testing.T, got *apiModels.MentionFilter) {
				t.Helper()
				if got.Search != "hello world" {
					t.Errorf("search: want %q, got %q", "hello world", got.Search)
				}
			},
		},
		{
			name:        "invalid limit returns validation error",
			queryString: "limit=abc",
			wantErrMsg:  "limit must be a valid integer",
		},
		{
			name:        "invalid min_followers returns validation error",
			queryString: "min_followers=abc",
			wantErrMsg:  "min_followers must be a valid integer",
		},
		{
			name:        "invalid min_total_engagement returns validation error",
			queryString: "min_total_engagement=abc",
			wantErrMsg:  "min_total_engagement must be a valid integer",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			url := "/?"
			if tc.queryString != "" {
				url += tc.queryString
			}
			r := httptest.NewRequest(http.MethodGet, url, nil)
			got, err := parseMentionFilter(r)

			if tc.wantErrMsg != "" {
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErrMsg)
				}
				if err.Error() != tc.wantErrMsg {
					t.Errorf("error: want %q, got %q", tc.wantErrMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tc.check(t, got)
		})
	}
}

func TestHandleListMentions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		url        string
		reader     *stubMentionReader
		wantStatus int
	}{
		{
			name: "returns 200 with mentions when topic_ids provided",
			url:  "/api/listening/mentions?topic_ids[]=t1",
			reader: &stubMentionReader{
				queryMentionsRows: []chModels.ListeningMentionRow{{MentionID: "m1"}},
				countUnreadVal:    3,
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 500 on query error",
			url:        "/api/listening/mentions?topic_ids[]=t1",
			reader:     &stubMentionReader{queryMentionsErr: errors.New("db down")},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "returns 400 on invalid limit",
			url:        "/api/listening/mentions?topic_ids[]=t1&limit=abc",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid min_followers",
			url:        "/api/listening/mentions?topic_ids[]=t1&min_followers=xyz",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid min_total_engagement",
			url:        "/api/listening/mentions?topic_ids[]=t1&min_total_engagement=xyz",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when no topic_ids and no workspace_id",
			url:        "/api/listening/mentions",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newMentionsHandler(tc.reader)
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rec := httptest.NewRecorder()
			h.HandleListMentions(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandlePatchMention(t *testing.T) {
	t.Parallel()

	validRow := &chModels.ListeningMentionRow{MentionID: "m1"}
	boolTrue := true

	tests := []struct {
		name       string
		mentionID  string
		body       interface{}
		reader     *stubMentionReader
		wantStatus int
	}{
		{
			name:       "returns 200 on valid patch",
			mentionID:  "m1",
			body:       apiModels.MentionPatchRequest{IsRead: &boolTrue, TopicID: "t1"},
			reader:     &stubMentionReader{getMentionRow: validRow},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 on missing mention id",
			mentionID:  "",
			body:       apiModels.MentionPatchRequest{},
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid JSON body",
			mentionID:  "m1",
			body:       "not-json",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid sentiment_override value",
			mentionID:  "m1",
			body:       apiModels.MentionPatchRequest{SentimentOverride: "happy"},
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "accepts valid sentiment_override positive",
			mentionID:  "m1",
			body:       apiModels.MentionPatchRequest{SentimentOverride: "positive", TopicID: "t1"},
			reader:     &stubMentionReader{getMentionRow: validRow},
			wantStatus: http.StatusOK,
		},
		{
			name:       "accepts valid sentiment_override neutral",
			mentionID:  "m1",
			body:       apiModels.MentionPatchRequest{SentimentOverride: "neutral", TopicID: "t1"},
			reader:     &stubMentionReader{getMentionRow: validRow},
			wantStatus: http.StatusOK,
		},
		{
			name:       "accepts valid sentiment_override negative",
			mentionID:  "m1",
			body:       apiModels.MentionPatchRequest{SentimentOverride: "negative", TopicID: "t1"},
			reader:     &stubMentionReader{getMentionRow: validRow},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 500 on service get error",
			mentionID:  "m1",
			body:       apiModels.MentionPatchRequest{TopicID: "t1"},
			reader:     &stubMentionReader{getMentionErr: errors.New("db error")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newMentionsHandler(tc.reader)
			req := httptest.NewRequest(http.MethodPatch, "/api/listening/mentions/"+tc.mentionID, bodyOf(tc.body))
			if tc.mentionID != "" {
				req.SetPathValue("id", tc.mentionID)
			}
			rec := httptest.NewRecorder()
			h.HandlePatchMention(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleMarkAllRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       interface{}
		reader     *stubMentionReader
		wantStatus int
	}{
		{
			name:       "returns 200 with marked count",
			body:       apiModels.MarkAllReadRequest{TopicIDs: []string{"t1"}, Platforms: []string{"twitter"}},
			reader:     &stubMentionReader{markAllReadCount: 5},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 200 with empty filter body",
			body:       apiModels.MarkAllReadRequest{},
			reader:     &stubMentionReader{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 on invalid JSON",
			body:       "not-json",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 500 on service error",
			body:       apiModels.MarkAllReadRequest{TopicIDs: []string{"t1"}},
			reader:     &stubMentionReader{markAllReadErr: errors.New("db error")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newMentionsHandler(tc.reader)
			req := httptest.NewRequest(http.MethodPost, "/api/listening/mentions/mark-all-read", bodyOf(tc.body))
			rec := httptest.NewRecorder()
			h.HandleMarkAllRead(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleUnreadCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		url        string
		reader     *stubMentionReader
		wantStatus int
	}{
		{
			name:       "returns 200 with unread count",
			url:        "/api/listening/mentions/unread-count?topic_ids[]=t1",
			reader:     &stubMentionReader{countUnreadVal: 3},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 500 on count service error",
			url:        "/api/listening/mentions/unread-count?topic_ids[]=t1",
			reader:     &stubMentionReader{countUnreadErr: errors.New("db error")},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "returns 400 on invalid limit param",
			url:        "/api/listening/mentions/unread-count?topic_ids[]=t1&limit=abc",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when no topic_ids and no workspace_id",
			url:        "/api/listening/mentions/unread-count",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newMentionsHandler(tc.reader)
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rec := httptest.NewRecorder()
			h.HandleUnreadCount(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
