package listening

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	chDB "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	chModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	mentionsSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/mentions"
)

func newAnalyticsHandler(reader *stubMentionReader) *AnalyticsHandler {
	svc := mentionsSvc.NewService(reader, zerolog.Nop())
	resolver := NewMentionFilterResolver(
		stubViewResolver{},
		stubTopicWorkspaceResolver{},
	)
	return NewAnalyticsHandler(svc, resolver, zerolog.Nop())
}

func TestHandleGetAnalytics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		url        string
		reader     *stubMentionReader
		wantStatus int
	}{
		{
			name: "returns 200 with analytics data",
			url:  "/api/listening/analytics?topic_ids[]=t1",
			reader: &stubMentionReader{
				getAnalyticsData: &chDB.AnalyticsData{
					TotalMentions:   10,
					SentimentCounts: map[string]int{"positive": 6, "negative": 4},
					PlatformCounts:  map[string]int{"twitter": 10},
					TagCounts:       map[string]int{},
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 500 on analytics service error",
			url:        "/api/listening/analytics?topic_ids[]=t1",
			reader:     &stubMentionReader{getAnalyticsErr: errors.New("db error")},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "returns 400 on invalid limit param",
			url:        "/api/listening/analytics?topic_ids[]=t1&limit=abc",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when no topic_ids and no workspace_id",
			url:        "/api/listening/analytics",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "maps topic_names[] to analytics response",
			url:  "/api/listening/analytics?topic_ids[]=t1&topic_ids[]=t2&topic_names[]=Alpha&topic_names[]=Beta",
			reader: &stubMentionReader{
				getAnalyticsData: &chDB.AnalyticsData{
					TotalMentions:   2,
					SentimentCounts: map[string]int{"positive": 2},
					PlatformCounts:  map[string]int{"reddit": 2},
					TagCounts:       map[string]int{},
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "returns 400 on invalid min_followers param",
			url:        "/api/listening/analytics?topic_ids[]=t1&min_followers=bad",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid min_total_engagement param",
			url:        "/api/listening/analytics?topic_ids[]=t1&min_total_engagement=bad",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newAnalyticsHandler(tc.reader)
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rec := httptest.NewRecorder()
			h.HandleGetAnalytics(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleExportMentions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		url             string
		reader          *stubMentionReader
		wantStatus      int
		wantContentType string
	}{
		{
			name: "exports CSV by default (no format param)",
			url:  "/api/listening/analytics/export?topic_ids[]=t1",
			reader: &stubMentionReader{
				queryMentionsRows: []chModels.ListeningMentionRow{
					{MentionID: "m1", Platform: "twitter", SentimentLabel: "positive"},
				},
			},
			wantStatus:      http.StatusOK,
			wantContentType: "text/csv",
		},
		{
			name: "exports CSV when format=csv",
			url:  "/api/listening/analytics/export?topic_ids[]=t1&format=csv",
			reader: &stubMentionReader{
				queryMentionsRows: []chModels.ListeningMentionRow{},
			},
			wantStatus:      http.StatusOK,
			wantContentType: "text/csv",
		},
		{
			name: "exports PDF when format=pdf",
			url:  "/api/listening/analytics/export?topic_ids[]=t1&format=pdf",
			reader: &stubMentionReader{
				getAnalyticsData: &chDB.AnalyticsData{
					TotalMentions:   0,
					SentimentCounts: map[string]int{},
					PlatformCounts:  map[string]int{},
					TagCounts:       map[string]int{},
				},
				queryMentionsRows: []chModels.ListeningMentionRow{},
			},
			wantStatus:      http.StatusOK,
			wantContentType: "application/pdf",
		},
		{
			name:       "returns 400 on invalid limit param",
			url:        "/api/listening/analytics/export?topic_ids[]=t1&limit=abc",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when no topic_ids and no workspace_id",
			url:        "/api/listening/analytics/export",
			reader:     &stubMentionReader{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newAnalyticsHandler(tc.reader)
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rec := httptest.NewRecorder()
			h.HandleExportMentions(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
			if tc.wantContentType != "" {
				ct := rec.Header().Get("Content-Type")
				if ct != tc.wantContentType {
					t.Errorf("Content-Type: want %q, got %q", tc.wantContentType, ct)
				}
			}
		})
	}
}
