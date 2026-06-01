package mongo

import "testing"

func TestAIContext_IsEmpty(t *testing.T) {
	cases := []struct {
		name string
		ctx  AIContext
		want bool
	}{
		{"all empty", AIContext{}, true},
		{"only brand_name", AIContext{BrandName: "Acme"}, false},
		{"only competitors", AIContext{Competitors: []AICompetitor{{Name: "x"}}}, false},
		{"only industry", AIContext{Industry: "SaaS"}, false},
		{"only brand_keywords", AIContext{BrandKeywords: []string{"acme"}}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.ctx.IsEmpty(); got != tc.want {
				t.Fatalf("IsEmpty=%v want %v", got, tc.want)
			}
		})
	}
}
