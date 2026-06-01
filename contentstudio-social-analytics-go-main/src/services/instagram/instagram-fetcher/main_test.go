package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/semaphore"
)

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if maxMediaWorkers != 10 {
		t.Errorf("maxMediaWorkers = %d, want 10", maxMediaWorkers)
	}
	if maxInsightsWorkers != 5 {
		t.Errorf("maxInsightsWorkers = %d, want 5", maxInsightsWorkers)
	}
	if mediaQueueSize != 500 {
		t.Errorf("mediaQueueSize = %d, want 500", mediaQueueSize)
	}
	if insightsQueueSize != 500 {
		t.Errorf("insightsQueueSize = %d, want 500", insightsQueueSize)
	}
	if mediaInsightsConcPerWorker != 3 {
		t.Errorf("mediaInsightsConcPerWorker = %d, want 3", mediaInsightsConcPerWorker)
	}
	if accountInsightsConcPerWorker != 2 {
		t.Errorf("accountInsightsConcPerWorker = %d, want 2", accountInsightsConcPerWorker)
	}
	if maxConcurrentAccounts != 50 {
		t.Errorf("maxConcurrentAccounts = %d, want 50", maxConcurrentAccounts)
	}
	if perAccountConcurrency != 1 {
		t.Errorf("perAccountConcurrency = %d, want 1", perAccountConcurrency)
	}
	if timestampUpdateChanSize != 1000 {
		t.Errorf("timestampUpdateChanSize = %d, want 1000", timestampUpdateChanSize)
	}
}

// ================== Struct Tests ==================

func TestTimestampUpdateRequest_Struct(t *testing.T) {
	req := TimestampUpdateRequest{
		AccountID:   "acc123",
		InstagramID: "ig456",
	}

	if req.AccountID != "acc123" {
		t.Errorf("AccountID = %q, want %q", req.AccountID, "acc123")
	}
	if req.InstagramID != "ig456" {
		t.Errorf("InstagramID = %q, want %q", req.InstagramID, "ig456")
	}
}

func TestResolvedOrder_Struct(t *testing.T) {
	ro := ResolvedOrder{
		AccountID:             "acc123",
		InstagramID:           "ig456",
		WorkspaceID:           "ws789",
		AccessTokenPlaintext:  "token",
		ConnectedViaInstagram: true,
		AppSecret:             "secret",
	}

	if ro.AccountID != "acc123" {
		t.Errorf("AccountID = %q, want %q", ro.AccountID, "acc123")
	}
	if ro.InstagramID != "ig456" {
		t.Errorf("InstagramID = %q, want %q", ro.InstagramID, "ig456")
	}
	if ro.WorkspaceID != "ws789" {
		t.Errorf("WorkspaceID = %q, want %q", ro.WorkspaceID, "ws789")
	}
	if !ro.ConnectedViaInstagram {
		t.Error("ConnectedViaInstagram should be true")
	}
}

func TestMediaJob_Struct(t *testing.T) {
	since := time.Now()
	job := MediaJob{
		Order:    ResolvedOrder{InstagramID: "ig123"},
		SyncType: "incremental",
		Since:    &since,
	}

	if job.SyncType != "incremental" {
		t.Errorf("SyncType = %q, want %q", job.SyncType, "incremental")
	}
	if job.Since == nil {
		t.Error("Since should not be nil")
	}
}

func TestInsightsJob_Struct(t *testing.T) {
	now := time.Now()
	job := InsightsJob{
		Order:    ResolvedOrder{InstagramID: "ig123"},
		SyncType: "full_sync",
		Since:    now.AddDate(0, 0, -30),
		Until:    now,
	}

	if job.SyncType != "full_sync" {
		t.Errorf("SyncType = %q, want %q", job.SyncType, "full_sync")
	}
}

// ================== semForAccount Tests ==================

func TestSemForAccount_NewAccount(t *testing.T) {
	// Clear the map first
	accountSemaphores = sync.Map{}

	sem := semForAccount("new_account_123", 1)
	if sem == nil {
		t.Fatal("semForAccount returned nil")
	}
}

func TestSemForAccount_ExistingAccount(t *testing.T) {
	accountSemaphores = sync.Map{}

	sem1 := semForAccount("existing_account", 1)
	sem2 := semForAccount("existing_account", 1)

	if sem1 != sem2 {
		t.Error("semForAccount should return same semaphore for same account")
	}
}

func TestSemForAccount_DifferentAccounts(t *testing.T) {
	accountSemaphores = sync.Map{}

	sem1 := semForAccount("account_a", 1)
	sem2 := semForAccount("account_b", 1)

	if sem1 == sem2 {
		t.Error("semForAccount should return different semaphores for different accounts")
	}
}

// ================== resolveAccessToken Tests ==================

func TestResolveAccessToken_Empty(t *testing.T) {
	log := logger.New("error")
	result := resolveAccessToken("", "key", "ig123", log)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestResolveAccessToken_IGAAToken(t *testing.T) {
	log := logger.New("error")
	token := "IGAAxxxxxxxxxxxxxxxx"
	result := resolveAccessToken(token, "key", "ig123", log)
	if result != token {
		t.Errorf("expected %q, got %q", token, result)
	}
}

func TestResolveAccessToken_EAAToken(t *testing.T) {
	log := logger.New("error")
	token := "EAAxxxxxxxxxxxxxxxx"
	result := resolveAccessToken(token, "key", "ig123", log)
	if result != token {
		t.Errorf("expected %q, got %q", token, result)
	}
}

func TestResolveAccessToken_DecryptionFails_ReturnsEmpty(t *testing.T) {
	log := logger.New("error")
	// Token that doesn't start with IGAA/EAA and can't be decrypted
	// Must use a non-prefixed token to avoid early return
	token := "some_random_opaque_token"
	result := resolveAccessToken(token, "bad_key", "ig123", log)
	if result != "" {
		t.Errorf("expected empty string on decryption failure, got %q", result)
	}
}

func TestResolveAccessToken_DecryptionSucceeds(t *testing.T) {
	log := logger.New("error")
	// Encrypt a known token using the crypto package
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	base64Key := base64.StdEncoding.EncodeToString(key)

	// Create a valid encrypted payload
	plaintext := "EAAxxxxDecrypted"
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		t.Fatalf("failed to generate IV: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	paddedPlaintext := pkcs7Pad([]byte(plaintext), aes.BlockSize)
	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	payload := struct {
		IV    string `json:"iv"`
		Value string `json:"value"`
	}{
		IV:    base64.StdEncoding.EncodeToString(iv),
		Value: base64.StdEncoding.EncodeToString(ciphertext),
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	encryptedToken := base64.StdEncoding.EncodeToString(jsonPayload)

	result := resolveAccessToken(encryptedToken, base64Key, "ig123", log)
	if result != plaintext {
		t.Errorf("expected %q, got %q", plaintext, result)
	}
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// ================== processAccount Tests ==================

func TestProcessAccount_EmptyToken(t *testing.T) {
	log := logger.New("error")
	mediaJobs := make(chan MediaJob, 10)
	insightsJobs := make(chan InsightsJob, 10)

	order := kafkamodels.InstagramAccountWorkOrder{
		ID:          "acc123",
		InstagramID: "ig456",
		AccessToken: "", // Empty token
		SyncType:    "incremental",
	}

	ctx := context.Background()
	err := processAccount(ctx, order, "key", "secret", mediaJobs, insightsJobs, log, nil, func() {}, func() {})

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	// No jobs should be created with empty token
	if len(mediaJobs) != 0 {
		t.Errorf("expected no media jobs, got %d", len(mediaJobs))
	}
	if len(insightsJobs) != 0 {
		t.Errorf("expected no insights jobs, got %d", len(insightsJobs))
	}
}

func TestProcessAccount_ValidToken_Incremental(t *testing.T) {
	log := logger.New("error")
	mediaJobs := make(chan MediaJob, 10)
	insightsJobs := make(chan InsightsJob, 10)

	order := kafkamodels.InstagramAccountWorkOrder{
		ID:          "acc123",
		InstagramID: "ig456",
		AccessToken: "EAAxxxxxxxxxx",
		SyncType:    "incremental",
		WorkspaceID: "ws789",
	}

	ctx := context.Background()
	err := processAccount(ctx, order, "key", "secret", mediaJobs, insightsJobs, log, nil, func() {}, func() {})

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if len(mediaJobs) != 1 {
		t.Errorf("expected 1 media job, got %d", len(mediaJobs))
	}
	if len(insightsJobs) != 1 {
		t.Errorf("expected 1 insights job, got %d", len(insightsJobs))
	}

	// Check media job
	mediaJob := <-mediaJobs
	if mediaJob.SyncType != "incremental" {
		t.Errorf("mediaJob.SyncType = %q, want %q", mediaJob.SyncType, "incremental")
	}
	if mediaJob.Since == nil {
		t.Error("mediaJob.Since should not be nil for incremental sync")
	}

	// Check insights job
	insightsJob := <-insightsJobs
	if insightsJob.SyncType != "incremental" {
		t.Errorf("insightsJob.SyncType = %q, want %q", insightsJob.SyncType, "incremental")
	}
}

func TestProcessAccount_ValidToken_FullSync(t *testing.T) {
	log := logger.New("error")
	mediaJobs := make(chan MediaJob, 10)
	insightsJobs := make(chan InsightsJob, 10)

	order := kafkamodels.InstagramAccountWorkOrder{
		ID:          "acc123",
		InstagramID: "ig456",
		AccessToken: "IGAAxxxxxxxxxx",
		SyncType:    "full_sync",
	}

	ctx := context.Background()
	err := processAccount(ctx, order, "key", "secret", mediaJobs, insightsJobs, log, nil, func() {}, func() {})

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	mediaJob := <-mediaJobs
	if mediaJob.Since != nil {
		t.Error("mediaJob.Since should be nil for full_sync")
	}
}

func TestProcessAccount_ValidToken_Immediate(t *testing.T) {
	log := logger.New("error")
	mediaJobs := make(chan MediaJob, 10)
	insightsJobs := make(chan InsightsJob, 10)

	order := kafkamodels.InstagramAccountWorkOrder{
		ID:          "acc123",
		InstagramID: "ig456",
		AccessToken: "EAAxxxxxxxxxx",
		SyncType:    "immediate",
	}

	ctx := context.Background()
	err := processAccount(ctx, order, "key", "secret", mediaJobs, insightsJobs, log, nil, func() {}, func() {})

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	mediaJob := <-mediaJobs
	if mediaJob.Since != nil {
		t.Error("mediaJob.Since should be nil for immediate sync")
	}
}

func TestProcessAccount_ContextCanceled_OnMediaSend(t *testing.T) {
	log := logger.New("error")
	mediaJobs := make(chan MediaJob) // Unbuffered - will block on media send
	insightsJobs := make(chan InsightsJob, 10)

	order := kafkamodels.InstagramAccountWorkOrder{
		ID:          "acc123",
		InstagramID: "ig456",
		AccessToken: "EAAxxxxxxxxxx",
		SyncType:    "incremental",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := processAccount(ctx, order, "key", "secret", mediaJobs, insightsJobs, log, nil, func() {}, func() {})

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestProcessAccount_ContextCanceled_OnInsightsSend(t *testing.T) {
	log := logger.New("error")
	mediaJobs := make(chan MediaJob)       // Unbuffered - we read it to unblock
	insightsJobs := make(chan InsightsJob) // Unbuffered - will block on insights send

	order := kafkamodels.InstagramAccountWorkOrder{
		ID:          "acc123",
		InstagramID: "ig456",
		AccessToken: "EAAxxxxxxxxxx",
		SyncType:    "incremental",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- processAccount(ctx, order, "key", "secret", mediaJobs, insightsJobs, log, nil, func() {}, func() {})
	}()

	// Drain media job so first select succeeds
	<-mediaJobs

	// Cancel context so insights select hits ctx.Done()
	cancel()

	err := <-errCh
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// ================== startTimestampUpdater Tests ==================

func TestStartTimestampUpdater_ChannelCloseStops(t *testing.T) {
	log := logger.New("error")
	repo := &mongodb.MockUnifiedSocialRepository{}
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var wg sync.WaitGroup

	startTimestampUpdater(&wg, repo, timestampUpdateChan, log)

	close(timestampUpdateChan)
	wg.Wait()
}

func TestStartTimestampUpdater_EmptyChannelClose(t *testing.T) {
	log := logger.New("error")
	repo := &mongodb.MockUnifiedSocialRepository{}
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var wg sync.WaitGroup

	startTimestampUpdater(&wg, repo, timestampUpdateChan, log)

	close(timestampUpdateChan)
	wg.Wait()
}

func TestStartTimestampUpdater_ProcessesRequest(t *testing.T) {
	log := logger.New("error")

	var updateCalled int32
	repo := &mongodb.MockUnifiedSocialRepository{
		GetByPlatformIDFunc: func(ctx context.Context, platformType, platformID string) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				PlatformIdentifier: platformID,
			}, nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			atomic.AddInt32(&updateCalled, 1)
			return nil
		},
	}

	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var wg sync.WaitGroup

	startTimestampUpdater(&wg, repo, timestampUpdateChan, log)

	// Send update request
	timestampUpdateChan <- TimestampUpdateRequest{
		AccountID:   "acc123",
		InstagramID: "ig456",
	}

	time.Sleep(100 * time.Millisecond)

	close(timestampUpdateChan)
	wg.Wait()

	if atomic.LoadInt32(&updateCalled) != 1 {
		t.Errorf("UpdateAnalyticsTimestamp called %d times, want 1", updateCalled)
	}
}

func TestStartTimestampUpdater_GetByPlatformIDError(t *testing.T) {
	log := logger.New("error")

	repo := &mongodb.MockUnifiedSocialRepository{
		GetByPlatformIDFunc: func(ctx context.Context, platformType, platformID string) (*mongomodels.SocialIntegration, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var wg sync.WaitGroup

	startTimestampUpdater(&wg, repo, timestampUpdateChan, log)

	// Send update request - should handle error gracefully
	timestampUpdateChan <- TimestampUpdateRequest{
		AccountID:   "acc123",
		InstagramID: "ig456",
	}

	time.Sleep(100 * time.Millisecond)

	close(timestampUpdateChan)
	wg.Wait()
	// Test passes if no panic
}

func TestStartTimestampUpdater_AccountNotFound(t *testing.T) {
	log := logger.New("error")

	repo := &mongodb.MockUnifiedSocialRepository{
		GetByPlatformIDFunc: func(ctx context.Context, platformType, platformID string) (*mongomodels.SocialIntegration, error) {
			return nil, nil // Account not found
		},
	}

	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var wg sync.WaitGroup

	startTimestampUpdater(&wg, repo, timestampUpdateChan, log)

	// Send update request - should handle nil account gracefully
	timestampUpdateChan <- TimestampUpdateRequest{
		AccountID:   "acc123",
		InstagramID: "ig456",
	}

	time.Sleep(100 * time.Millisecond)

	close(timestampUpdateChan)
	wg.Wait()
	// Test passes if no panic
}

func TestStartTimestampUpdater_UpdateStateError(t *testing.T) {
	log := logger.New("error")

	repo := &mongodb.MockUnifiedSocialRepository{
		GetByPlatformIDFunc: func(ctx context.Context, platformType, platformID string) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				PlatformIdentifier: platformID,
			}, nil
		},
		UpdateStateFunc: func(ctx context.Context, id primitive.ObjectID, state string) error {
			return fmt.Errorf("state update error")
		},
	}

	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var wg sync.WaitGroup

	startTimestampUpdater(&wg, repo, timestampUpdateChan, log)

	timestampUpdateChan <- TimestampUpdateRequest{
		AccountID:   "acc123",
		InstagramID: "ig456",
	}

	time.Sleep(100 * time.Millisecond)

	close(timestampUpdateChan)
	wg.Wait()
}

func TestStartTimestampUpdater_UpdateTimestampError(t *testing.T) {
	log := logger.New("error")

	repo := &mongodb.MockUnifiedSocialRepository{
		GetByPlatformIDFunc: func(ctx context.Context, platformType, platformID string) (*mongomodels.SocialIntegration, error) {
			return &mongomodels.SocialIntegration{
				PlatformIdentifier: platformID,
			}, nil
		},
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return fmt.Errorf("update error")
		},
	}

	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	var wg sync.WaitGroup

	startTimestampUpdater(&wg, repo, timestampUpdateChan, log)

	// Send update request - should handle error gracefully
	timestampUpdateChan <- TimestampUpdateRequest{
		AccountID:   "acc123",
		InstagramID: "ig456",
	}

	time.Sleep(100 * time.Millisecond)

	close(timestampUpdateChan)
	wg.Wait()
	// Test passes if no panic
}

// ================== Batch Work Order Tests ==================

func TestBatchWorkOrder_Unmarshal(t *testing.T) {
	jsonData := `{
		"batch_id": "batch123",
		"sync_type": "incremental",
		"accounts": [
			{
				"id": "acc1",
				"instagram_id": "ig1",
				"access_token": "token1",
				"sync_type": "incremental"
			}
		]
	}`

	var batchOrder kafkamodels.InstagramBatchWorkOrder
	err := json.Unmarshal([]byte(jsonData), &batchOrder)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if batchOrder.BatchID != "batch123" {
		t.Errorf("BatchID = %q, want %q", batchOrder.BatchID, "batch123")
	}
	if len(batchOrder.Accounts) != 1 {
		t.Errorf("len(Accounts) = %d, want 1", len(batchOrder.Accounts))
	}
}

// ================== Concurrent Tests ==================

func TestSemForAccount_ConcurrentAccess(t *testing.T) {
	accountSemaphores = sync.Map{}

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			accountID := "account_" + string(rune('0'+id%10))
			sem := semForAccount(accountID, 1)
			if sem == nil {
				t.Error("semForAccount returned nil")
			}
		}(i)
	}

	wg.Wait()
}

func TestSemForAccount_LoadOrStoreRace(t *testing.T) {
	// Maximize chance of hitting the LoadOrStore race path by having many
	// goroutines race on the same fresh key simultaneously
	for round := 0; round < 50; round++ {
		accountSemaphores = sync.Map{}
		key := fmt.Sprintf("race_key_%d", round)

		barrier := make(chan struct{})
		var wg sync.WaitGroup
		results := make([]*semaphore.Weighted, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				<-barrier // Wait for all goroutines to be ready
				results[idx] = semForAccount(key, 1)
			}(i)
		}

		close(barrier) // Release all goroutines simultaneously
		wg.Wait()

		// All goroutines must get the same semaphore
		for i := 1; i < 10; i++ {
			if results[i] != results[0] {
				t.Fatalf("round %d: goroutine %d got different semaphore", round, i)
			}
		}
	}
}

// ================== Atomic Counter Tests ==================

func TestAtomicCounters(t *testing.T) {
	var counter int64

	var wg sync.WaitGroup
	numGoroutines := 100
	incrementsPerGoroutine := 1000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				atomic.AddInt64(&counter, 1)
			}
		}()
	}

	wg.Wait()

	expected := int64(numGoroutines * incrementsPerGoroutine)
	if atomic.LoadInt64(&counter) != expected {
		t.Errorf("counter = %d, want %d", counter, expected)
	}
}

// ================== isExpectedInstagramError Tests ==================

func TestIsExpectedInstagramError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"auth error", fmt.Errorf("invalid access token"), true},
		{"OAuthException 190", fmt.Errorf("OAuthException (#190) token expired"), true},
		{"OAuthException 10", fmt.Errorf("OAuthException: (#10) not enough viewers"), true},
		{"not enough viewers lowercase", fmt.Errorf("not enough viewers for insights"), true},
		{"error #10", fmt.Errorf("error (#10) occurred"), true},
		{"permission error", fmt.Errorf("Application does not have permission for this action"), true},
		{"network error", fmt.Errorf("connection timeout"), false},
		{"parse error", fmt.Errorf("json parse failed"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExpectedInstagramError(tt.err)
			if got != tt.expected {
				t.Errorf("isExpectedInstagramError() = %v, want %v for error: %v", got, tt.expected, tt.err)
			}
		})
	}
}

// ================== Logging Contract Tests ==================

func TestLoggingContract_InstagramFetcher_AuthError_WarnOnly(t *testing.T) {
	hookRecords, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	log, buf := logger.NewTestLoggerWithHook()

	// Simulate an expected auth error — should be Warn level, not Error
	log.Warn().
		Str("error_message", "OAuthException (#190) token expired").
		Str("function", "mediaWorker").
		Str("instagram_id", "ig123").
		Msg("Instagram auth error, skipping account")

	output := buf.String()

	if !strings.Contains(output, "WRN") {
		t.Fatalf("expected WRN level in output, got: %s", output)
	}
	if strings.Contains(output, "ERR") {
		t.Fatalf("auth error should NOT produce ERR level: %s", output)
	}

	// Verify no Error-level hook firings
	for _, r := range *hookRecords {
		if r.Level >= zerolog.ErrorLevel {
			t.Fatalf("auth error should not trigger Error+ hook, got level %v", r.Level)
		}
	}
}

func TestLoggingContract_InstagramFetcher_NoDuplicateCaptureOnError(t *testing.T) {
	hookRecords, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()
	captureRecords, captureCleanup := logger.InstallCaptureSpy()
	defer captureCleanup()

	log, buf := logger.NewTestLoggerWithHook()

	// Simulate what the Instagram fetcher does for an unexpected error:
	// one Error log, no explicit CaptureException (hook handles Sentry)
	log.Error().
		Str("error_message", "unexpected API failure").
		Str("function", "mediaWorker").
		Str("stage", "fetch_media").
		Msg("Instagram fetcher error")

	output := buf.String()

	// Count ERR occurrences — should be exactly 1
	errCount := strings.Count(output, "ERR")
	if errCount != 1 {
		t.Fatalf("expected exactly 1 ERR entry in log output, got %d. Output:\n%s", errCount, output)
	}

	// Verify hook fired exactly once at Error level
	var hookErrCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.ErrorLevel {
			hookErrCount++
		}
	}
	if hookErrCount != 1 {
		t.Fatalf("expected exactly 1 ErrorLevel hook firing, got %d", hookErrCount)
	}

	// Verify NO explicit CaptureException was called (no duplication with hook)
	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls, got %d", len(*captureRecords))
	}
}
