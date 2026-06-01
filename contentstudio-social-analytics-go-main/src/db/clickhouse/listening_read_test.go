package clickhouse

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

type mockCountRow struct {
	value uint64
	err   error
}

func (m *mockCountRow) Err() error { return m.err }

func (m *mockCountRow) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}

	if len(dest) > 0 {
		if count, ok := dest[0].(*uint64); ok {
			*count = m.value
		}
	}

	return nil
}

func (m *mockCountRow) ScanStruct(dest any) error { return m.err }

func TestMissingEnrichmentWhereClause(t *testing.T) {
	t.Parallel()

	clause := missingEnrichmentWhereClause()

	if !strings.Contains(clause, "sentiment_label = ''") {
		t.Fatalf("expected clause to include missing sentiment_label check, got %q", clause)
	}
	if !strings.Contains(clause, "length(ai_tags) = 0") {
		t.Fatalf("expected clause to include empty ai_tags check, got %q", clause)
	}
}

// --- placeholders ---

func TestPlaceholders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		n    int
		want string
	}{
		{0, ""},
		{1, "?"},
		{2, "?,?"},
		{3, "?,?,?"},
		{5, "?,?,?,?,?"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run("", func(t *testing.T) {
			t.Parallel()
			got := placeholders(tc.n)
			if got != tc.want {
				t.Errorf("placeholders(%d): want %q, got %q", tc.n, tc.want, got)
			}
		})
	}
}

// --- normalizeTags ---

func TestNormalizeTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "lowercases tags",
			input: []string{"BuyIntent", "SUPPORT"},
			want:  []string{"buyintent", "support"},
		},
		{
			name:  "strips leading hash",
			input: []string{"#buy_intent", "#support"},
			want:  []string{"buy_intent", "support"},
		},
		{
			name:  "trims whitespace",
			input: []string{"  buy_intent  "},
			want:  []string{"buy_intent"},
		},
		{
			name:  "deduplicates",
			input: []string{"buy_intent", "Buy_Intent", "#buy_intent"},
			want:  []string{"buy_intent"},
		},
		{
			name:  "skips empty strings",
			input: []string{"", "  ", "#"},
			want:  []string{},
		},
		{
			name:  "nil input returns empty",
			input: nil,
			want:  []string{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeTags(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("len: want %d, got %d (%v)", len(tc.want), len(got), got)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("[%d]: want %q, got %q", i, tc.want[i], got[i])
				}
			}
		})
	}
}

// --- normalizeLanguage ---

func TestNormalizeLanguage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"english", "en"},
		{"English", "en"},
		{"ENGLISH", "en"},
		{"spanish", "es"},
		{"french", "fr"},
		{"german", "de"},
		{"italian", "it"},
		{"portuguese", "pt"},
		{"dutch", "nl"},
		{"arabic", "ar"},
		{"turkish", "tr"},
		{"hindi", "hi"},
		{"urdu", "ur"},
		{"indonesian", "id"},
		{"malay", "ms"},
		{"japanese", "ja"},
		{"korean", "ko"},
		{"chinese", "zh"},
		// ISO codes pass through unchanged
		{"en", "en"},
		{"fr", "fr"},
		{"de", "de"},
		// Unknown values pass through lowercased
		{"pirate", "pirate"},
		// Empty returns empty
		{"", ""},
		{"   ", ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := normalizeLanguage(tc.input)
			if got != tc.want {
				t.Errorf("normalizeLanguage(%q): want %q, got %q", tc.input, tc.want, got)
			}
		})
	}
}

// --- normalizeLanguages ---

func TestNormalizeLanguages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "converts full names to codes",
			input: []string{"English", "French"},
			want:  []string{"en", "fr"},
		},
		{
			name:  "deduplicates same code via different spellings",
			input: []string{"English", "english", "en"},
			want:  []string{"en"},
		},
		{
			name:  "skips empty entries",
			input: []string{"", "  ", "en"},
			want:  []string{"en"},
		},
		{
			name:  "nil returns empty",
			input: nil,
			want:  []string{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeLanguages(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("len: want %d, got %d (%v)", len(tc.want), len(got), got)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("[%d]: want %q, got %q", i, tc.want[i], got[i])
				}
			}
		})
	}
}

// --- encodeCursor / decodeCursor round-trip ---

func TestCursorRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	original := MentionCursor{PostedAt: now, MentionID: "m-abc-123", TotalEngagement: 42}

	encoded := encodeCursor(original)
	if encoded == "" {
		t.Fatal("encoded cursor is empty")
	}

	decoded, err := decodeCursor(encoded)
	if err != nil {
		t.Fatalf("decodeCursor error: %v", err)
	}
	if !decoded.PostedAt.Equal(original.PostedAt) {
		t.Errorf("PostedAt: want %v, got %v", original.PostedAt, decoded.PostedAt)
	}
	if decoded.MentionID != original.MentionID {
		t.Errorf("MentionID: want %q, got %q", original.MentionID, decoded.MentionID)
	}
	if decoded.TotalEngagement != original.TotalEngagement {
		t.Errorf("TotalEngagement: want %d, got %d", original.TotalEngagement, decoded.TotalEngagement)
	}
}

func TestDecodeCursor_InvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"not base64", "not-valid-base64!!"},
		{"valid base64 but not JSON", "aW52YWxpZC1qc29u"},
		{"empty string", ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := decodeCursor(tc.input)
			if err == nil {
				t.Errorf("expected error for input %q, got nil", tc.input)
			}
		})
	}
}

// --- buildOrderBy ---

func TestBuildOrderBy(t *testing.T) {
	t.Parallel()

	r := &ListeningReadRepository{}

	tests := []struct {
		sort string
		want string
	}{
		{"oldest", "posted_at ASC, mention_id ASC"},
		{"most_engaged", "total_engagement DESC, posted_at DESC, mention_id DESC"},
		{"", "posted_at DESC, mention_id DESC"},
		{"latest", "posted_at DESC, mention_id DESC"},
		{"unknown", "posted_at DESC, mention_id DESC"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.sort, func(t *testing.T) {
			t.Parallel()
			got := r.buildOrderBy(tc.sort)
			if got != tc.want {
				t.Errorf("sort=%q: want %q, got %q", tc.sort, tc.want, got)
			}
		})
	}
}

// --- buildWhereClause ---

func TestBuildWhereClause(t *testing.T) {
	t.Parallel()

	r := &ListeningReadRepository{}
	boolTrue := true
	boolFalse := false

	tests := []struct {
		name            string
		filter          *MentionFilter
		wantContains    []string
		wantNotContains []string
		wantArgCount    int
	}{
		{
			name:   "empty filter keeps default irrelevant filter",
			filter: &MentionFilter{},
			wantContains: []string{
				"post_irrelevant = false",
				"NOT arrayExists(x -> x = 'Irrelevant', ai_tags)",
			},
			wantArgCount: 0,
		},
		{
			name:         "filters by topic ids",
			filter:       &MentionFilter{TopicIDs: []string{"t1", "t2"}},
			wantContains: []string{"topic_id IN (?,?)"},
			wantArgCount: 2,
		},
		{
			name:         "filters by platforms",
			filter:       &MentionFilter{Platforms: []string{"twitter"}},
			wantContains: []string{"platform IN (?)"},
			wantArgCount: 1,
		},
		{
			name:         "filters by sentiments",
			filter:       &MentionFilter{Sentiments: []string{"positive", "neutral"}},
			wantContains: []string{"sentiment_label IN (?,?)"},
			wantArgCount: 2,
		},
		{
			name:         "filters by ai tags",
			filter:       &MentionFilter{AITags: []string{"buy_intent"}},
			wantContains: []string{"arrayExists"},
			wantArgCount: 1,
		},
		{
			name:         "filters by min_followers",
			filter:       &MentionFilter{MinFollowers: 1000},
			wantContains: []string{"author_followers >= ?"},
			wantArgCount: 1,
		},
		{
			name:         "filters by min_total_engagement",
			filter:       &MentionFilter{MinTotalEngagement: 250},
			wantContains: []string{"total_engagement >= ?"},
			wantArgCount: 1,
		},
		{
			name: "filters by date range",
			filter: &MentionFilter{
				DateFrom: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				DateTo:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			},
			wantContains: []string{"posted_at >= ?", "posted_at <= ?"},
			wantArgCount: 2,
		},
		{
			name:         "filters by is_bookmarked=true",
			filter:       &MentionFilter{IsBookmarked: &boolTrue},
			wantContains: []string{"bookmark = ?"},
			wantArgCount: 1,
		},
		{
			name:         "filters by is_read=false",
			filter:       &MentionFilter{IsRead: &boolFalse},
			wantContains: []string{"post_read = ?"},
			wantArgCount: 1,
		},
		{
			name:   "excludes irrelevant by default",
			filter: &MentionFilter{},
			wantContains: []string{
				"post_irrelevant = false",
				"NOT arrayExists(x -> x = 'Irrelevant', ai_tags)",
			},
			wantArgCount: 0,
		},
		{
			name:         "includes irrelevant when flag set",
			filter:       &MentionFilter{IncludeIrrelevant: true},
			wantContains: []string{},
			wantNotContains: []string{
				"post_irrelevant = false",
				"arrayExists",
			},
			wantArgCount: 0,
		},
		{
			name:         "filters by search text",
			filter:       &MentionFilter{Search: "golang"},
			wantContains: []string{"positionCaseInsensitive"},
			wantArgCount: 1,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			clause, args := r.buildWhereClause(tc.filter)

			for _, expected := range tc.wantContains {
				if !containsSubstr(clause, expected) {
					t.Errorf("WHERE clause missing %q\ngot: %s", expected, clause)
				}
			}
			for _, unwanted := range tc.wantNotContains {
				if containsSubstr(clause, unwanted) {
					t.Errorf("WHERE clause should not contain %q\ngot: %s", unwanted, clause)
				}
			}
			if len(args) != tc.wantArgCount {
				t.Errorf("arg count: want %d, got %d", tc.wantArgCount, len(args))
			}
		})
	}
}

func TestCountMentions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		filter  *MentionFilter
		row     *mockCountRow
		want    int
		wantErr string
	}{
		{
			name:   "counts mentions with nil filter",
			filter: nil,
			row:    &mockCountRow{value: 12},
			want:   12,
		},
		{
			name: "counts mentions with topic filter",
			filter: &MentionFilter{
				TopicIDs:  []string{"topic-1", "topic-2"},
				Platforms: []string{"twitter"},
			},
			row:  &mockCountRow{value: 7},
			want: 7,
		},
		{
			name:    "wraps scan errors",
			filter:  &MentionFilter{},
			row:     &mockCountRow{err: errors.New("scan failed")},
			wantErr: "ListeningReadRepository.CountMentions: scan failed",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &ListeningReadRepository{
				client: newTestClient(&mockConn{queryRowResult: tc.row}),
				logger: zerolog.Nop(),
			}

			got, err := repo.CountMentions(context.Background(), tc.filter)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want %d, got %d", tc.want, got)
			}
		})
	}
}

// --- buildCursorClause ---

func TestBuildCursorClause(t *testing.T) {
	t.Parallel()

	r := &ListeningReadRepository{logger: zerolog.Nop()}

	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	validCursor := encodeCursor(MentionCursor{PostedAt: now, MentionID: "m1", TotalEngagement: 99})

	tests := []struct {
		name         string
		filter       *MentionFilter
		wantEmpty    bool
		wantContains string
	}{
		{
			name:      "empty cursor returns no clause",
			filter:    &MentionFilter{},
			wantEmpty: true,
		},
		{
			name:      "invalid cursor is ignored and returns no clause",
			filter:    &MentionFilter{Cursor: "not-valid-base64!!"},
			wantEmpty: true,
		},
		{
			name:         "default sort uses less-than comparison",
			filter:       &MentionFilter{Cursor: validCursor},
			wantContains: "posted_at, mention_id) < (?, ?)",
		},
		{
			name:         "oldest sort uses greater-than comparison",
			filter:       &MentionFilter{Cursor: validCursor, Sort: "oldest"},
			wantContains: "posted_at, mention_id) > (?, ?)",
		},
		{
			name:         "most engaged sort uses full sort tuple",
			filter:       &MentionFilter{Cursor: validCursor, Sort: "most_engaged"},
			wantContains: "total_engagement, posted_at, mention_id) < (?, ?, ?)",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			clause, args := r.buildCursorClause(tc.filter)

			if tc.wantEmpty {
				if clause != "" || len(args) != 0 {
					t.Errorf("want empty clause, got %q with %d args", clause, len(args))
				}
				return
			}
			if !containsSubstr(clause, tc.wantContains) {
				t.Errorf("want clause containing %q, got %q", tc.wantContains, clause)
			}
			wantArgs := 2
			if tc.filter.Sort == "most_engaged" {
				wantArgs = 3
			}
			if len(args) != wantArgs {
				t.Errorf("want %d args, got %d", wantArgs, len(args))
			}
		})
	}
}

func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
