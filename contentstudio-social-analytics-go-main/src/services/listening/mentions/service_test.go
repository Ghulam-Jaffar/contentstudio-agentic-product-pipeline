package mentions

import (
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/parser"
)

func TestConstructAuthorProfileURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		platform    string
		handle      string
		authorID    string
		expectedURL string
	}{
		{
			name:        "instagram handle",
			platform:    "instagram",
			handle:      "photographer",
			expectedURL: "https://www.instagram.com/photographer/",
		},
		{
			name:        "facebook fallback to profile id",
			platform:    "facebook",
			authorID:    "123456789",
			expectedURL: "https://www.facebook.com/profile.php?id=123456789",
		},
		{
			name:        "twitter trims at-sign",
			platform:    "twitter",
			handle:      "@korrssk",
			expectedURL: "https://x.com/korrssk",
		},
		{
			name:        "twitter falls back to author id",
			platform:    "twitter",
			authorID:    "2029325572719972759",
			expectedURL: "https://x.com/2029325572719972759",
		},
		{
			name:        "tiktok handle",
			platform:    "tiktok",
			handle:      "creator",
			expectedURL: "https://www.tiktok.com/@creator",
		},
		{
			name:        "threads handle",
			platform:    "threads",
			handle:      "threader",
			expectedURL: "https://www.threads.net/@threader",
		},
		{
			name:        "reddit handle",
			platform:    "reddit",
			handle:      "u/product_hunter",
			expectedURL: "https://www.reddit.com/user/product_hunter",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := parser.ConstructAuthorProfileURL(tc.platform, tc.handle, tc.authorID, "")
			if got != tc.expectedURL {
				t.Fatalf("expected %q, got %q", tc.expectedURL, got)
			}
		})
	}
}
