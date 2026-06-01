package tokenstore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mt "go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

func testLogger() *logger.Logger {
	zl := zerolog.New(io.Discard)
	return &logger.Logger{Logger: zl}
}

func runMT(t *testing.T, name string, fn func(*mt.T)) {
	t.Helper()
	h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
	h.Run(name, fn)
}

func setupMiniredis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return mr, client
}

func TestPlatformConstants(t *testing.T) {
	if PlatformFacebook != "facebook" {
		t.Fatalf("expected PlatformFacebook 'facebook', got %q", PlatformFacebook)
	}
	if PlatformInstagram != "instagram" {
		t.Fatalf("expected PlatformInstagram 'instagram', got %q", PlatformInstagram)
	}
}

func TestNewTokenStore(t *testing.T) {
	ts := NewTokenStore(PlatformFacebook, nil, nil, testLogger())
	if ts == nil {
		t.Fatal("expected non-nil TokenStore")
	}
	if ts.platform != PlatformFacebook {
		t.Fatalf("expected platform %q, got %q", PlatformFacebook, ts.platform)
	}
	if ts.queueName != "facebook_valid_token_set" {
		t.Fatalf("expected queue name facebook_valid_token_set, got %q", ts.queueName)
	}
}

func TestTokenStore_GetValidAccounts_Facebook(t *testing.T) {
	doc := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongomodels.PlatformFacebook},
		{Key: "platform_identifier", Value: "fb_123"},
		{Key: "type", Value: "Page"},
		{Key: "validity", Value: mongomodels.ValidityValid},
		{Key: "access_token", Value: "token123"},
	}

	runMT(t, "facebook accounts", func(m *mt.T) {
		ns := mt.TestDb + ".social_integrations"
		m.AddMockResponses(
			mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc),
			mt.CreateCursorResponse(0, ns, mt.NextBatch),
		)

		ts := &TokenStore{
			platform: PlatformFacebook,
			mongoDB:  m.DB,
			log:      testLogger(),
		}

		accounts, err := ts.GetValidAccounts(context.Background())
		if err != nil {
			m.Fatalf("unexpected error: %v", err)
		}
		if len(accounts) != 1 {
			m.Fatalf("expected 1 account, got %d", len(accounts))
		}
		if accounts[0].PlatformIdentifier != "fb_123" {
			m.Fatalf("expected platform identifier fb_123, got %q", accounts[0].PlatformIdentifier)
		}
	})
}

func TestTokenStore_GetValidAccounts_Instagram(t *testing.T) {
	doc := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongomodels.PlatformInstagram},
		{Key: "platform_identifier", Value: "ig_123"},
		{Key: "validity", Value: mongomodels.ValidityValid},
		{Key: "facebook_page_id", Value: "fb_page_123"},
		{Key: "access_token", Value: "token456"},
	}

	runMT(t, "instagram accounts", func(m *mt.T) {
		ns := mt.TestDb + ".social_integrations"
		m.AddMockResponses(
			mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc),
			mt.CreateCursorResponse(0, ns, mt.NextBatch),
		)

		ts := &TokenStore{
			platform: PlatformInstagram,
			mongoDB:  m.DB,
			log:      testLogger(),
		}

		accounts, err := ts.GetValidAccounts(context.Background())
		if err != nil {
			m.Fatalf("unexpected error: %v", err)
		}
		if len(accounts) != 1 {
			m.Fatalf("expected 1 account, got %d", len(accounts))
		}
	})
}

func TestTokenStore_GetValidAccounts_UnsupportedPlatform(t *testing.T) {
	runMT(t, "unsupported platform", func(m *mt.T) {
		ts := &TokenStore{
			platform: "unsupported",
			mongoDB:  m.DB,
			log:      testLogger(),
		}

		_, err := ts.GetValidAccounts(context.Background())
		if err == nil {
			m.Fatal("expected error for unsupported platform")
		}
	})
}

func TestTokenStore_PopulateRedisSet_FromMongoOnly(t *testing.T) {
	mr, redisClient := setupMiniredis(t)
	defer mr.Close()

	// Pre-populate Redis to verify PopulateRedisSet ignores it now.
	mr.SAdd("facebook_valid_token_set", `{"platform_id":"stale","token":"stale-token"}`)

	doc := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongomodels.PlatformFacebook},
		{Key: "platform_identifier", Value: "fb_789"},
		{Key: "type", Value: "Page"},
		{Key: "validity", Value: mongomodels.ValidityValid},
		{Key: "access_token", Value: "token3"},
	}

	runMT(t, "populate from mongo only", func(m *mt.T) {
		ns := mt.TestDb + ".social_integrations"
		m.AddMockResponses(
			mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc),
			mt.CreateCursorResponse(0, ns, mt.NextBatch),
		)

		ts := &TokenStore{
			platform:    PlatformFacebook,
			queueName:   "facebook_valid_token_set",
			redisClient: redisClient,
			mongoDB:     m.DB,
			log:         testLogger(),
		}

		tokenSet, err := ts.PopulateRedisSet(context.Background())
		if err != nil {
			m.Fatalf("unexpected error: %v", err)
		}
		if len(tokenSet) != 1 {
			m.Fatalf("expected 1 token, got %d", len(tokenSet))
		}
	})
}

func TestTokenStore_PopulateRedisSet_InstagramFallbackToLegacyID(t *testing.T) {
	mr, redisClient := setupMiniredis(t)
	defer mr.Close()

	doc := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongomodels.PlatformInstagram},
		{Key: "instagram_id", Value: "ig_legacy_123"},
		{Key: "validity", Value: mongomodels.ValidityValid},
		{Key: "facebook_page_id", Value: "fb_page_123"},
		{Key: "access_token", Value: "ig_token123"},
	}

	runMT(t, "instagram fallback to instagram_id", func(m *mt.T) {
		ns := mt.TestDb + ".social_integrations"
		m.AddMockResponses(
			mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc),
			mt.CreateCursorResponse(0, ns, mt.NextBatch),
		)

		ts := &TokenStore{
			platform:    PlatformInstagram,
			queueName:   "instagram_valid_token_set",
			redisClient: redisClient,
			mongoDB:     m.DB,
			log:         testLogger(),
		}

		tokenSet, err := ts.PopulateRedisSet(context.Background())
		if err != nil {
			m.Fatalf("unexpected error: %v", err)
		}
		if len(tokenSet) != 1 {
			m.Fatalf("expected 1 token, got %d", len(tokenSet))
		}

		found := false
		for _, td := range tokenSet {
			if td.PlatformID == "ig_legacy_123" {
				found = true
			}
		}
		if !found {
			m.Fatal("expected token with instagram_id fallback")
		}
	})
}

func TestTokenStore_PopulateRedisSet_NoPlatformIdentifier(t *testing.T) {
	mr, redisClient := setupMiniredis(t)
	defer mr.Close()

	doc := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongomodels.PlatformFacebook},
		{Key: "type", Value: "Page"},
		{Key: "validity", Value: mongomodels.ValidityValid},
		{Key: "access_token", Value: "token123"},
	}

	runMT(t, "skip account with no platform id", func(m *mt.T) {
		ns := mt.TestDb + ".social_integrations"
		m.AddMockResponses(
			mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc),
			mt.CreateCursorResponse(0, ns, mt.NextBatch),
		)

		ts := &TokenStore{
			platform:    PlatformFacebook,
			queueName:   "facebook_valid_token_set",
			redisClient: redisClient,
			mongoDB:     m.DB,
			log:         testLogger(),
		}

		tokenSet, err := ts.PopulateRedisSet(context.Background())
		if err != nil {
			m.Fatalf("unexpected error: %v", err)
		}
		if len(tokenSet) != 0 {
			m.Fatalf("expected 0 tokens, got %d", len(tokenSet))
		}
	})
}

func TestTokenStore_Validate_ReplacesRedisSet(t *testing.T) {
	mr, redisClient := setupMiniredis(t)
	defer mr.Close()

	staleToken := `{"platform_id":"stale","token":"old"}`
	mr.SAdd("facebook_valid_token_set", staleToken)

	ts := &TokenStore{
		platform:    PlatformFacebook,
		queueName:   "facebook_valid_token_set",
		redisClient: redisClient,
		log:         testLogger(),
	}

	tokenSet := map[string]TokenData{
		`{"platform_id":"fb_123","token":"token1"}`: {PlatformID: "fb_123", Token: "token1"},
		`{"platform_id":"fb_456","token":"token2"}`: {PlatformID: "fb_456", Token: "token2"},
	}

	if err := ts.Validate(context.Background(), tokenSet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	members, _ := redisClient.SMembers(context.Background(), "facebook_valid_token_set").Result()
	if len(members) != 2 {
		t.Fatalf("expected 2 members in Redis set, got %d", len(members))
	}
	for _, member := range members {
		if member == staleToken {
			t.Fatal("expected stale Redis token to be removed during refresh")
		}
	}
}

func TestTokenStore_Validate_ContextCancelled(t *testing.T) {
	mr, redisClient := setupMiniredis(t)
	defer mr.Close()

	ts := &TokenStore{
		platform:    PlatformFacebook,
		queueName:   "facebook_valid_token_set",
		redisClient: redisClient,
		log:         testLogger(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tokenSet := map[string]TokenData{
		`{"platform_id":"fb_123","token":"token1"}`: {PlatformID: "fb_123", Token: "token1"},
	}

	err := ts.Validate(ctx, tokenSet)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestTokenStore_Validate_LogsProgressEvery50(t *testing.T) {
	mr, redisClient := setupMiniredis(t)
	defer mr.Close()

	log, buf := logger.NewTestLogger()

	ts := &TokenStore{
		platform:    PlatformFacebook,
		queueName:   "facebook_valid_token_set",
		redisClient: redisClient,
		log:         log,
	}

	tokenSet := make(map[string]TokenData, 120)
	for i := 0; i < 120; i++ {
		td := TokenData{
			PlatformID: fmt.Sprintf("fb_%03d", i),
			Token:      fmt.Sprintf("token_%03d", i),
		}
		b, err := json.Marshal(td)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}
		tokenSet[string(b)] = td
	}

	if err := ts.Validate(context.Background(), tokenSet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if count := strings.Count(output, "Token sync progress"); count != 2 {
		t.Fatalf("expected 2 progress log entries at 50 and 100 inserts, got %d\noutput=%s", count, output)
	}
	if !strings.Contains(output, "\"inserted_tokens\":50") && !strings.Contains(output, "inserted_tokens=50") {
		t.Fatalf("expected progress log for 50 inserted tokens, got %s", output)
	}
	if !strings.Contains(output, "\"inserted_tokens\":100") && !strings.Contains(output, "inserted_tokens=100") {
		t.Fatalf("expected progress log for 100 inserted tokens, got %s", output)
	}
}

func TestTokenStore_ProcessJob_Success(t *testing.T) {
	mr, redisClient := setupMiniredis(t)
	defer mr.Close()

	doc := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongomodels.PlatformFacebook},
		{Key: "platform_identifier", Value: "fb_123"},
		{Key: "type", Value: "Page"},
		{Key: "validity", Value: mongomodels.ValidityValid},
		{Key: "access_token", Value: "token123"},
	}

	runMT(t, "process job success", func(m *mt.T) {
		ns := mt.TestDb + ".social_integrations"
		m.AddMockResponses(
			mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc),
			mt.CreateCursorResponse(0, ns, mt.NextBatch),
		)

		ts := &TokenStore{
			platform:    PlatformFacebook,
			queueName:   "facebook_valid_token_set",
			redisClient: redisClient,
			mongoDB:     m.DB,
			log:         testLogger(),
		}

		if err := ts.ProcessJob(context.Background()); err != nil {
			m.Fatalf("unexpected error: %v", err)
		}

		members, _ := redisClient.SMembers(context.Background(), "facebook_valid_token_set").Result()
		if len(members) != 1 {
			m.Fatalf("expected 1 member in Redis set, got %d", len(members))
		}
	})
}

func TestTokenStore_ProcessJob_PopulateError(t *testing.T) {
	mr, redisClient := setupMiniredis(t)
	defer mr.Close()

	runMT(t, "process job populate error", func(m *mt.T) {
		m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{
			Message: "find failed",
			Code:    1,
		}))

		ts := &TokenStore{
			platform:    PlatformFacebook,
			queueName:   "facebook_valid_token_set",
			redisClient: redisClient,
			mongoDB:     m.DB,
			log:         testLogger(),
		}

		err := ts.ProcessJob(context.Background())
		if err == nil {
			m.Fatal("expected error for populate failure")
		}
	})
}

func TestTokenStore_ProcessJob_ValidateContextCancelled(t *testing.T) {
	mr, redisClient := setupMiniredis(t)
	defer mr.Close()

	doc := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongomodels.PlatformFacebook},
		{Key: "platform_identifier", Value: "fb_123"},
		{Key: "type", Value: "Page"},
		{Key: "validity", Value: mongomodels.ValidityValid},
		{Key: "access_token", Value: "token123"},
	}

	runMT(t, "process job validate context cancelled", func(m *mt.T) {
		ns := mt.TestDb + ".social_integrations"
		m.AddMockResponses(
			mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc),
			mt.CreateCursorResponse(0, ns, mt.NextBatch),
		)

		ts := &TokenStore{
			platform:    PlatformFacebook,
			queueName:   "facebook_valid_token_set",
			redisClient: redisClient,
			mongoDB:     m.DB,
			log:         testLogger(),
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := ts.ProcessJob(ctx)
		if err == nil {
			m.Fatal("expected error for cancelled context during validation")
		}
	})
}

func TestLoggingContract_TokenStore_WarnLevelOnly(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:1",
	})
	defer redisClient.Close()

	runMT(t, "logging contract warn level", func(m *mt.T) {
		ts := &TokenStore{
			platform:    PlatformFacebook,
			queueName:   "facebook_valid_token_set",
			redisClient: redisClient,
			mongoDB:     m.DB,
			log:         log,
		}

		_, _ = ts.PopulateRedisSet(context.Background())

		output := buf.String()
		if strings.Contains(output, "ERR") {
			m.Fatalf("TokenStore should not emit ERR-level logs, got: %s", output)
		}
		if len(*captureRecords) != 0 {
			m.Fatalf("expected 0 CaptureException calls, got %d", len(*captureRecords))
		}
	})
}

func TestLoggingContract_TokenStore_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:1",
	})
	defer redisClient.Close()

	runMT(t, "no capture exception", func(m *mt.T) {
		ts := &TokenStore{
			platform:    PlatformFacebook,
			queueName:   "facebook_valid_token_set",
			redisClient: redisClient,
			mongoDB:     m.DB,
			log:         log,
		}

		_ = ts.ProcessJob(context.Background())
	})

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls, got %d", len(*captureRecords))
	}
}
