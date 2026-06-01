package social

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	metaadsmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestMetaAdsRateManager(t *testing.T) {
	rm := NewMetaAdsRateManager(MetaAdsRateLimits{})
	if rm == nil {
		t.Fatal("expected non-nil rate manager")
	}

	if rm.tokenLimiter("token-a") != rm.tokenLimiter("token-a") {
		t.Fatal("expected token limiter reuse")
	}

	if err := rm.Wait(context.Background(), "token-a"); err != nil {
		t.Fatalf("unexpected wait error: %v", err)
	}
}

func TestMetaAdsClientHelpers(t *testing.T) {
	log := logger.NewNop()
	client := NewMetaAdsClientWithRates("app-secret", NewMetaAdsRateManager(MetaAdsRateLimits{}), log)
	if client == nil {
		t.Fatal("expected client")
	}

	proof := client.appsecretProof("access-token")
	if len(proof) != 64 {
		t.Fatalf("expected 64-char proof, got %d", len(proof))
	}

	u := client.buildURL("act_1/insights", "access-token", url.Values{"fields": []string{"spend"}})
	if !strings.Contains(u, "access_token=access-token") {
		t.Fatalf("expected access token in url: %s", u)
	}
	if !strings.Contains(u, "appsecret_proof=") {
		t.Fatalf("expected appsecret proof in url: %s", u)
	}
	if !strings.HasPrefix(u, metaAdsBaseURL+metaAdsAPIVersion+"/act_1/insights?") {
		t.Fatalf("unexpected url prefix: %s", u)
	}
}

func TestMetaAdsParsingHelpers(t *testing.T) {
	if got := ParseMetaAdsFloat64(" 12.5 "); got != 12.5 {
		t.Fatalf("expected 12.5, got %v", got)
	}
	if got := ParseMetaAdsInt64(" 42 "); got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}

	actions := []metaadsmodels.RawMetaAdsAction{
		{ActionType: "purchase", Value: "7"},
		{ActionType: "lead", Value: "3"},
	}
	if got := MetaAdsActionValue(actions, "lead"); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}

	truncated := TruncateToHour(time.Date(2025, 5, 13, 17, 42, 11, 123, time.FixedZone("PKT", 5*60*60)))
	if truncated.Minute() != 0 || truncated.Second() != 0 || truncated.Location().String() != "PKT" {
		t.Fatalf("expected hour truncation in original timezone, got %v", truncated)
	}
}
