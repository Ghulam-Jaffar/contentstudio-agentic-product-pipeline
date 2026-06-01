package enrichment

import (
	"encoding/json"
	"testing"

	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

func TestBuildWireContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   TopicContext
		want aiContextWire
	}{
		{
			name: "zero TopicContext yields empty wire with non-nil slices",
			in:   TopicContext{},
			want: aiContextWire{
				BrandKeywords: []string{},
				Competitors:   []competitorWire{},
				TopicKeywords: []string{},
			},
		},
		{
			name: "all AIContext fields populated, no topic fields",
			in: TopicContext{
				AIContext: mongoModels.AIContext{
					BrandName:     "Acme",
					BrandKeywords: []string{"acme", "ACME corp"},
					Industry:      "SaaS",
					Competitors: []mongoModels.AICompetitor{
						{Name: "Foo", Keywords: []string{"foo", "f00"}},
						{Name: "Bar", Keywords: []string{"bar"}},
					},
				},
			},
			want: aiContextWire{
				BrandName:     "Acme",
				BrandKeywords: []string{"acme", "ACME corp"},
				Industry:      "SaaS",
				Competitors: []competitorWire{
					{Name: "Foo", Keywords: []string{"foo", "f00"}},
					{Name: "Bar", Keywords: []string{"bar"}},
				},
				TopicKeywords: []string{},
			},
		},
		{
			name: "topic fields populated, no AIContext",
			in: TopicContext{
				TopicName:     "Acme Brand",
				TopicType:     "own_brand",
				TopicKeywords: []string{"acme", "saas"},
				RelevanceHint: "B2B only",
			},
			want: aiContextWire{
				BrandKeywords: []string{},
				Competitors:   []competitorWire{},
				TopicName:     "Acme Brand",
				TopicType:     "own_brand",
				TopicKeywords: []string{"acme", "saas"},
				RelevanceHint: "B2B only",
			},
		},
		{
			name: "competitor with nil keywords serializes as empty array",
			in: TopicContext{
				AIContext: mongoModels.AIContext{
					BrandName: "Acme",
					Competitors: []mongoModels.AICompetitor{
						{Name: "Solo", Keywords: nil},
					},
				},
			},
			want: aiContextWire{
				BrandName:     "Acme",
				BrandKeywords: []string{},
				Competitors: []competitorWire{
					{Name: "Solo", Keywords: []string{}},
				},
				TopicKeywords: []string{},
			},
		},
		{
			name: "nil BrandKeywords becomes empty slice (JSON-friendly)",
			in: TopicContext{
				AIContext: mongoModels.AIContext{
					BrandName:     "Acme",
					BrandKeywords: nil,
				},
			},
			want: aiContextWire{
				BrandName:     "Acme",
				BrandKeywords: []string{},
				Competitors:   []competitorWire{},
				TopicKeywords: []string{},
			},
		},
		{
			name: "every field set together",
			in: TopicContext{
				AIContext: mongoModels.AIContext{
					BrandName:     "Acme",
					BrandKeywords: []string{"acme"},
					Industry:      "SaaS",
					Competitors: []mongoModels.AICompetitor{
						{Name: "Foo", Keywords: []string{"foo"}},
					},
				},
				TopicName:     "Acme Brand",
				TopicType:     "own_brand",
				TopicKeywords: []string{"acme"},
				RelevanceHint: "B2B only",
			},
			want: aiContextWire{
				BrandName:     "Acme",
				BrandKeywords: []string{"acme"},
				Industry:      "SaaS",
				Competitors: []competitorWire{
					{Name: "Foo", Keywords: []string{"foo"}},
				},
				TopicName:     "Acme Brand",
				TopicType:     "own_brand",
				TopicKeywords: []string{"acme"},
				RelevanceHint: "B2B only",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := buildWireContext(tc.in)
			assertWireEqual(t, got, tc.want)
		})
	}
}

func TestBuildWireContext_DefensiveCopy(t *testing.T) {
	t.Parallel()

	original := []string{"acme"}
	in := TopicContext{
		AIContext: mongoModels.AIContext{
			BrandName:     "Acme",
			BrandKeywords: original,
			Competitors: []mongoModels.AICompetitor{
				{Name: "Foo", Keywords: []string{"foo"}},
			},
		},
		TopicKeywords: original,
	}

	got := buildWireContext(in)
	got.BrandKeywords[0] = "MUTATED"
	got.TopicKeywords[0] = "MUTATED"
	got.Competitors[0].Keywords[0] = "MUTATED"

	if original[0] != "acme" {
		t.Errorf("buildWireContext mutated caller's slice: got %q, want %q", original[0], "acme")
	}
	if in.AIContext.Competitors[0].Keywords[0] != "foo" {
		t.Errorf("buildWireContext mutated caller's competitor keywords: got %q, want %q",
			in.AIContext.Competitors[0].Keywords[0], "foo")
	}
}

func TestBuildWireContext_OmitsEmptyTopicFieldsInJSON(t *testing.T) {
	t.Parallel()

	wire := buildWireContext(TopicContext{
		AIContext: mongoModels.AIContext{BrandName: "Acme"},
	})

	body, err := json.Marshal(wire)
	if err != nil {
		t.Fatalf("Marshal err: %v", err)
	}
	s := string(body)

	for _, omitted := range []string{"topic_name", "topic_type", "topic_keywords", "relevance_hint"} {
		if contains(s, omitted) {
			t.Errorf("expected %q to be omitted from JSON, got: %s", omitted, s)
		}
	}
	for _, present := range []string{"brand_name", "brand_keywords", "competitors"} {
		if !contains(s, present) {
			t.Errorf("expected %q to appear in JSON, got: %s", present, s)
		}
	}
}

func assertWireEqual(t *testing.T, got, want aiContextWire) {
	t.Helper()
	if got.BrandName != want.BrandName {
		t.Errorf("BrandName: got %q, want %q", got.BrandName, want.BrandName)
	}
	if got.Industry != want.Industry {
		t.Errorf("Industry: got %q, want %q", got.Industry, want.Industry)
	}
	if got.TopicName != want.TopicName {
		t.Errorf("TopicName: got %q, want %q", got.TopicName, want.TopicName)
	}
	if got.TopicType != want.TopicType {
		t.Errorf("TopicType: got %q, want %q", got.TopicType, want.TopicType)
	}
	if got.RelevanceHint != want.RelevanceHint {
		t.Errorf("RelevanceHint: got %q, want %q", got.RelevanceHint, want.RelevanceHint)
	}
	assertStringSliceEqual(t, "BrandKeywords", got.BrandKeywords, want.BrandKeywords)
	assertStringSliceEqual(t, "TopicKeywords", got.TopicKeywords, want.TopicKeywords)

	if len(got.Competitors) != len(want.Competitors) {
		t.Errorf("Competitors len: got %d, want %d", len(got.Competitors), len(want.Competitors))
		return
	}
	for i := range got.Competitors {
		if got.Competitors[i].Name != want.Competitors[i].Name {
			t.Errorf("Competitors[%d].Name: got %q, want %q",
				i, got.Competitors[i].Name, want.Competitors[i].Name)
		}
		assertStringSliceEqual(t, "Competitors[].Keywords",
			got.Competitors[i].Keywords, want.Competitors[i].Keywords)
	}
}

func assertStringSliceEqual(t *testing.T, name string, got, want []string) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: nil slice (want non-nil for JSON safety)", name)
		return
	}
	if len(got) != len(want) {
		t.Errorf("%s len: got %d, want %d (got=%v want=%v)", name, len(got), len(want), got, want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s[%d]: got %q, want %q", name, i, got[i], want[i])
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
