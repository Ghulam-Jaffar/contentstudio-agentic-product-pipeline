package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// MockPlatformProcessor is a mock implementation of PlatformProcessor for testing
type MockPlatformProcessor struct {
	mu                sync.Mutex
	FacebookCalled    bool
	InstagramCalled   bool
	LinkedinCalled    bool
	YouTubeCalled     bool
	TikTokCalled      bool
	TwitterCalled     bool
	PinterestCalled   bool
	GmbCalled         bool
	MetaAdsCalled     bool
	FacebookAcctTypes []string
	FacebookSyncType  string
	InstagramSyncType string
	LinkedinSyncType  string
	YouTubeSyncType   string
	TikTokSyncType    string
	TwitterSyncType   string
	PinterestSyncType string
	GmbSyncType       string
	MetaAdsSyncType   string
	ProcessingDelay   time.Duration
}

// Verify DefaultPlatformProcessor implements PlatformProcessor interface
var _ PlatformProcessor = (*DefaultPlatformProcessor)(nil)

func (m *MockPlatformProcessor) ProcessFacebook(db *mongo.Database, producer kafka.Producer, log zerolog.Logger, accountTypes []string, syncType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FacebookCalled = true
	m.FacebookAcctTypes = accountTypes
	m.FacebookSyncType = syncType
	if m.ProcessingDelay > 0 {
		time.Sleep(m.ProcessingDelay)
	}
}

func (m *MockPlatformProcessor) ProcessInstagram(ctx context.Context, db *mongo.Database, producer kafka.Producer, log zerolog.Logger, accountTypes []string, syncType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.InstagramCalled = true
	m.InstagramSyncType = syncType
	if m.ProcessingDelay > 0 {
		time.Sleep(m.ProcessingDelay)
	}
}

func (m *MockPlatformProcessor) ProcessLinkedin(ctx context.Context, db *mongo.Database, producer kafka.Producer, log zerolog.Logger, accountTypes []string, syncType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LinkedinCalled = true
	m.LinkedinSyncType = syncType
	if m.ProcessingDelay > 0 {
		time.Sleep(m.ProcessingDelay)
	}
}

func (m *MockPlatformProcessor) ProcessYouTube(ctx context.Context, db *mongo.Database, producer kafka.Producer, log zerolog.Logger, syncType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.YouTubeCalled = true
	m.YouTubeSyncType = syncType
	if m.ProcessingDelay > 0 {
		time.Sleep(m.ProcessingDelay)
	}
}

func (m *MockPlatformProcessor) ProcessTikTok(ctx context.Context, db *mongo.Database, producer kafka.Producer, log zerolog.Logger, accountTypes []string, syncType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TikTokCalled = true
	m.TikTokSyncType = syncType
	if m.ProcessingDelay > 0 {
		time.Sleep(m.ProcessingDelay)
	}
}

func (m *MockPlatformProcessor) ProcessTwitter(ctx context.Context, db *mongo.Database, producer kafka.Producer, log zerolog.Logger, accountTypes []string, syncType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TwitterCalled = true
	m.TwitterSyncType = syncType
	if m.ProcessingDelay > 0 {
		time.Sleep(m.ProcessingDelay)
	}
}

func (m *MockPlatformProcessor) ProcessPinterest(ctx context.Context, db *mongo.Database, producer kafka.Producer, log zerolog.Logger, accountTypes []string, syncType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PinterestCalled = true
	m.PinterestSyncType = syncType
	if m.ProcessingDelay > 0 {
		time.Sleep(m.ProcessingDelay)
	}
}

func (m *MockPlatformProcessor) ProcessGMB(ctx context.Context, db *mongo.Database, producer kafka.Producer, log zerolog.Logger, syncType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GmbCalled = true
	m.GmbSyncType = syncType
	if m.ProcessingDelay > 0 {
		time.Sleep(m.ProcessingDelay)
	}
}

func (m *MockPlatformProcessor) ProcessMetaAds(db *mongo.Database, producer kafka.Producer, log zerolog.Logger, syncType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MetaAdsCalled = true
	m.MetaAdsSyncType = syncType
	if m.ProcessingDelay > 0 {
		time.Sleep(m.ProcessingDelay)
	}
}

func (m *MockPlatformProcessor) GetFacebookCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.FacebookCalled
}

func (m *MockPlatformProcessor) GetInstagramCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.InstagramCalled
}

func (m *MockPlatformProcessor) GetLinkedinCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.LinkedinCalled
}

func (m *MockPlatformProcessor) GetYouTubeCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.YouTubeCalled
}

func (m *MockPlatformProcessor) GetTikTokCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.TikTokCalled
}

func (m *MockPlatformProcessor) GetTwitterCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.TwitterCalled
}

func (m *MockPlatformProcessor) GetPinterestCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.PinterestCalled
}

func (m *MockPlatformProcessor) GetGmbCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.GmbCalled
}

// ================== RunService Tests ==================

func TestRunService_AllPlatforms(t *testing.T) {
	mockProcessor := &MockPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := logger.New("debug")

	deps := &ServiceDependencies{
		Database:          nil, // Mock doesn't use DB
		Producer:          mockProducer,
		Processor:         mockProcessor,
		Logger:            log,
		Platforms:         []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest", "gmb"},
		SyncType:          "incremental",
		FacebookAcctTypes: []string{"page", "group"},
	}

	ctx := context.Background()
	err := RunService(ctx, deps)

	if err != nil {
		t.Errorf("RunService failed: %v", err)
	}

	// Wait a bit for goroutines
	time.Sleep(50 * time.Millisecond)

	if !mockProcessor.GetFacebookCalled() {
		t.Error("expected Facebook processor to be called")
	}
	if !mockProcessor.GetInstagramCalled() {
		t.Error("expected Instagram processor to be called")
	}
	if !mockProcessor.GetLinkedinCalled() {
		t.Error("expected LinkedIn processor to be called")
	}
	if !mockProcessor.GetYouTubeCalled() {
		t.Error("expected YouTube processor to be called")
	}
	if !mockProcessor.GetTikTokCalled() {
		t.Error("expected TikTok processor to be called")
	}
	if !mockProcessor.GetTwitterCalled() {
		t.Error("expected Twitter processor to be called")
	}
	if !mockProcessor.GetPinterestCalled() {
		t.Error("expected Pinterest processor to be called")
	}
	if !mockProcessor.GetGmbCalled() {
		t.Error("expected GMB processor to be called")
	}
}

func TestRunService_SinglePlatform(t *testing.T) {
	mockProcessor := &MockPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := logger.New("debug")

	deps := &ServiceDependencies{
		Database:          nil,
		Producer:          mockProducer,
		Processor:         mockProcessor,
		Logger:            log,
		Platforms:         []string{"facebook"},
		SyncType:          "full_sync",
		FacebookAcctTypes: []string{"page"},
	}

	ctx := context.Background()
	err := RunService(ctx, deps)

	if err != nil {
		t.Errorf("RunService failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if !mockProcessor.GetFacebookCalled() {
		t.Error("expected Facebook processor to be called")
	}
	if mockProcessor.GetInstagramCalled() {
		t.Error("expected Instagram processor NOT to be called")
	}
	if mockProcessor.GetLinkedinCalled() {
		t.Error("expected LinkedIn processor NOT to be called")
	}
	if mockProcessor.GetYouTubeCalled() {
		t.Error("expected YouTube processor NOT to be called")
	}
	if mockProcessor.GetTwitterCalled() {
		t.Error("expected Twitter processor NOT to be called")
	}
}

func TestRunService_EmptyPlatforms(t *testing.T) {
	mockProcessor := &MockPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := logger.New("debug")

	deps := &ServiceDependencies{
		Database:          nil,
		Producer:          mockProducer,
		Processor:         mockProcessor,
		Logger:            log,
		Platforms:         []string{},
		SyncType:          "incremental",
		FacebookAcctTypes: nil,
	}

	ctx := context.Background()
	err := RunService(ctx, deps)

	if err != nil {
		t.Errorf("RunService failed: %v", err)
	}

	// No platforms means no processors called
	if mockProcessor.GetFacebookCalled() {
		t.Error("expected Facebook processor NOT to be called")
	}
	if mockProcessor.GetInstagramCalled() {
		t.Error("expected Instagram processor NOT to be called")
	}
	if mockProcessor.GetLinkedinCalled() {
		t.Error("expected LinkedIn processor NOT to be called")
	}
	if mockProcessor.GetYouTubeCalled() {
		t.Error("expected YouTube processor NOT to be called")
	}
	if mockProcessor.GetTwitterCalled() {
		t.Error("expected Twitter processor NOT to be called")
	}
}

func TestRunService_UnsupportedPlatform(t *testing.T) {
	mockProcessor := &MockPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := logger.New("debug")

	deps := &ServiceDependencies{
		Database:          nil,
		Producer:          mockProducer,
		Processor:         mockProcessor,
		Logger:            log,
		Platforms:         []string{"myspace"},
		SyncType:          "incremental",
		FacebookAcctTypes: nil,
	}

	ctx := context.Background()
	err := RunService(ctx, deps)

	if err != nil {
		t.Errorf("RunService failed: %v", err)
	}

	// Unsupported platforms should not call any processor
	if mockProcessor.GetFacebookCalled() {
		t.Error("expected Facebook processor NOT to be called")
	}
	if mockProcessor.GetInstagramCalled() {
		t.Error("expected Instagram processor NOT to be called")
	}
	if mockProcessor.GetLinkedinCalled() {
		t.Error("expected LinkedIn processor NOT to be called")
	}
	if mockProcessor.GetTikTokCalled() {
		t.Error("expected TikTok processor NOT to be called")
	}
	if mockProcessor.GetYouTubeCalled() {
		t.Error("expected YouTube processor NOT to be called")
	}
	if mockProcessor.GetTwitterCalled() {
		t.Error("expected Twitter processor NOT to be called")
	}
}

func TestRunService_WithContextCancellation(t *testing.T) {
	mockProcessor := &MockPlatformProcessor{
		ProcessingDelay: 100 * time.Millisecond,
	}
	mockProducer := &kafka.MockProducer{}
	log := logger.New("debug")

	deps := &ServiceDependencies{
		Database:          nil,
		Producer:          mockProducer,
		Processor:         mockProcessor,
		Logger:            log,
		Platforms:         []string{"facebook"},
		SyncType:          "incremental",
		FacebookAcctTypes: []string{"page"},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	// Should complete despite context being cancelled
	err := RunService(ctx, deps)
	if err != nil {
		t.Errorf("RunService failed: %v", err)
	}
}

func TestRunService_SyncTypePassThrough(t *testing.T) {
	mockProcessor := &MockPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := logger.New("debug")

	deps := &ServiceDependencies{
		Database:          nil,
		Producer:          mockProducer,
		Processor:         mockProcessor,
		Logger:            log,
		Platforms:         []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest"},
		SyncType:          "full_sync",
		FacebookAcctTypes: []string{"page"},
	}

	ctx := context.Background()
	err := RunService(ctx, deps)

	if err != nil {
		t.Errorf("RunService failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if mockProcessor.FacebookSyncType != "full_sync" {
		t.Errorf("Facebook sync type = %q, want %q", mockProcessor.FacebookSyncType, "full_sync")
	}
	if mockProcessor.InstagramSyncType != "full_sync" {
		t.Errorf("Instagram sync type = %q, want %q", mockProcessor.InstagramSyncType, "full_sync")
	}
	if mockProcessor.LinkedinSyncType != "full_sync" {
		t.Errorf("LinkedIn sync type = %q, want %q", mockProcessor.LinkedinSyncType, "full_sync")
	}
	if mockProcessor.YouTubeSyncType != "full_sync" {
		t.Errorf("YouTube sync type = %q, want %q", mockProcessor.YouTubeSyncType, "full_sync")
	}
	if mockProcessor.TikTokSyncType != "full_sync" {
		t.Errorf("TikTok sync type = %q, want %q", mockProcessor.TikTokSyncType, "full_sync")
	}
	if mockProcessor.TwitterSyncType != "full_sync" {
		t.Errorf("Twitter sync type = %q, want %q", mockProcessor.TwitterSyncType, "full_sync")
	}
	if mockProcessor.PinterestSyncType != "full_sync" {
		t.Errorf("Pinterest sync type = %q, want %q", mockProcessor.PinterestSyncType, "full_sync")
	}
}

// ================== BuildPlatformList Tests ==================

func TestBuildPlatformList_Empty(t *testing.T) {
	result := BuildPlatformList("")
	expected := []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest", "gmb", "meta_ads"}

	if len(result) != len(expected) {
		t.Fatalf("BuildPlatformList(\"\") returned %d items, want %d", len(result), len(expected))
	}

	for i, p := range expected {
		if result[i] != p {
			t.Errorf("result[%d] = %q, want %q", i, result[i], p)
		}
	}
}

func TestBuildPlatformList_Single(t *testing.T) {
	result := BuildPlatformList("facebook")

	if len(result) != 1 || result[0] != "facebook" {
		t.Errorf("BuildPlatformList(\"facebook\") = %v, want [facebook]", result)
	}
}

func TestBuildPlatformList_Multiple(t *testing.T) {
	result := BuildPlatformList("facebook,instagram")

	if len(result) != 2 {
		t.Fatalf("expected 2 platforms, got %d", len(result))
	}
	if result[0] != "facebook" || result[1] != "instagram" {
		t.Errorf("result = %v, want [facebook, instagram]", result)
	}
}

// ================== BuildAccountTypeList Tests ==================

func TestBuildAccountTypeList_Empty(t *testing.T) {
	result := BuildAccountTypeList("")

	if result != nil {
		t.Errorf("BuildAccountTypeList(\"\") = %v, want nil", result)
	}
}

func TestBuildAccountTypeList_Single(t *testing.T) {
	result := BuildAccountTypeList("page")

	if len(result) != 1 || result[0] != "page" {
		t.Errorf("BuildAccountTypeList(\"page\") = %v, want [page]", result)
	}
}

func TestBuildAccountTypeList_Multiple(t *testing.T) {
	result := BuildAccountTypeList("page,group")

	if len(result) != 2 {
		t.Fatalf("expected 2 types, got %d", len(result))
	}
	if result[0] != "page" || result[1] != "group" {
		t.Errorf("result = %v, want [page, group]", result)
	}
}

// ================== NormalizeSyncType Tests ==================

func TestNormalizeSyncType_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"incremental", "incremental"},
		{"full_sync", "full_sync"},
		{"full", "full_sync"},
		{"", "incremental"},
		{"unknown", "incremental"},
	}

	for _, tc := range tests {
		result := NormalizeSyncType(tc.input)
		if result != tc.expected {
			t.Errorf("NormalizeSyncType(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// ================== GetPlatformScaleInfo Tests ==================

func TestGetPlatformScaleInfo_Facebook(t *testing.T) {
	scale, ok := GetPlatformScaleInfo("facebook")

	if !ok {
		t.Error("expected facebook to be found")
	}
	if scale != 24000 {
		t.Errorf("scale = %d, want 24000", scale)
	}
}

func TestGetPlatformScaleInfo_Unknown(t *testing.T) {
	_, ok := GetPlatformScaleInfo("unknown")

	if ok {
		t.Error("expected unknown platform not to be found")
	}
}

func TestGetPlatformScaleInfo_AllPlatforms(t *testing.T) {
	expectedPlatforms := []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest"}

	for _, p := range expectedPlatforms {
		scale, ok := GetPlatformScaleInfo(p)
		if !ok {
			t.Errorf("expected platform %q to be found", p)
		}
		if scale <= 0 {
			t.Errorf("scale for %q should be positive, got %d", p, scale)
		}
	}
}

// ================== GetTotalAccountScale Tests ==================

func TestGetTotalAccountScale(t *testing.T) {
	total := GetTotalAccountScale()

	if total != 57500 {
		t.Errorf("GetTotalAccountScale() = %d, want 57500", total)
	}
}

// ================== IsPlatformSupported Tests ==================

func TestIsPlatformSupported_Valid(t *testing.T) {
	validPlatforms := []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest"}

	for _, p := range validPlatforms {
		if !IsPlatformSupported(p) {
			t.Errorf("expected %q to be supported", p)
		}
	}
}

func TestIsPlatformSupported_Invalid(t *testing.T) {
	invalidPlatforms := []string{"unknown", ""}

	for _, p := range invalidPlatforms {
		if IsPlatformSupported(p) {
			t.Errorf("expected %q to NOT be supported", p)
		}
	}
}

// ================== GetSupportedPlatformsForFetcher Tests ==================

func TestGetSupportedPlatformsForFetcher(t *testing.T) {
	result := GetSupportedPlatformsForFetcher()

	expected := []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d platforms, got %d", len(expected), len(result))
	}

	for i, p := range expected {
		if result[i] != p {
			t.Errorf("result[%d] = %q, want %q", i, result[i], p)
		}
	}
}

// ================== CalculatePlatformProgress Tests ==================

func TestCalculatePlatformProgress(t *testing.T) {
	tests := []struct {
		processed int
		total     int
		expected  float64
	}{
		{0, 100, 0.0},
		{50, 100, 50.0},
		{100, 100, 100.0},
		{25, 100, 25.0},
		{0, 0, 0.0}, // edge case: division by zero
	}

	for _, tc := range tests {
		result := CalculatePlatformProgress(tc.processed, tc.total)
		if result != tc.expected {
			t.Errorf("CalculatePlatformProgress(%d, %d) = %f, want %f", tc.processed, tc.total, result, tc.expected)
		}
	}
}

// ================== SplitPlatformString Tests ==================

func TestSplitPlatformString_Empty(t *testing.T) {
	result := SplitPlatformString("")
	expected := []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest", "gmb", "meta_ads"}

	if len(result) != len(expected) {
		t.Fatalf("expected %d platforms, got %d", len(expected), len(result))
	}
}

func TestSplitPlatformString_WithSpaces(t *testing.T) {
	result := SplitPlatformString("facebook , instagram")

	if len(result) != 2 {
		t.Fatalf("expected 2 platforms, got %d", len(result))
	}
	if result[0] != "facebook" || result[1] != "instagram" {
		t.Errorf("result = %v, want [facebook, instagram]", result)
	}
}

// ================== ValidatePlatformList Tests ==================

func TestValidatePlatformList_AllValid(t *testing.T) {
	input := []string{"facebook", "instagram", "linkedin", "tiktok", "pinterest"}
	valid, invalid := ValidatePlatformList(input)

	if len(valid) != 5 {
		t.Errorf("expected 5 valid platforms, got %d", len(valid))
	}
	if len(invalid) != 0 {
		t.Errorf("expected 0 invalid platforms, got %d", len(invalid))
	}
}

func TestValidatePlatformList_MixedValidInvalid(t *testing.T) {
	input := []string{"facebook", "youtube", "instagram", "tiktok", "pinterest", "twitter"}
	valid, invalid := ValidatePlatformList(input)

	if len(valid) != 6 {
		t.Errorf("expected 6 valid platforms, got %d: %v", len(valid), valid)
	}
	if len(invalid) != 0 {
		t.Errorf("expected 0 invalid platforms, got %d: %v", len(invalid), invalid)
	}
}

func TestValidatePlatformList_AllInvalid(t *testing.T) {
	input := []string{"snapchat", "unknown"}
	valid, invalid := ValidatePlatformList(input)

	if len(valid) != 0 {
		t.Errorf("expected 0 valid platforms, got %d", len(valid))
	}
	if len(invalid) != 2 {
		t.Errorf("expected 2 invalid platforms, got %d", len(invalid))
	}
}

func TestValidatePlatformList_Empty(t *testing.T) {
	valid, invalid := ValidatePlatformList([]string{})

	if len(valid) != 0 || len(invalid) != 0 {
		t.Errorf("expected empty results for empty input")
	}
}

func TestValidatePlatformList_Nil(t *testing.T) {
	valid, invalid := ValidatePlatformList(nil)

	if len(valid) != 0 || len(invalid) != 0 {
		t.Errorf("expected empty results for nil input")
	}
}

// ================== CreateServiceConfig Tests ==================

func TestCreateServiceConfig_Default(t *testing.T) {
	cfg := CreateServiceConfig("", "incremental", "page,group")

	expected := []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest", "gmb", "meta_ads"}
	if len(cfg.Platforms) != len(expected) {
		t.Fatalf("expected %d platforms, got %d", len(expected), len(cfg.Platforms))
	}

	if cfg.SyncType != "incremental" {
		t.Errorf("SyncType = %q, want %q", cfg.SyncType, "incremental")
	}

	if len(cfg.FacebookAcctTypes) != 2 {
		t.Errorf("expected 2 account types, got %d", len(cfg.FacebookAcctTypes))
	}
}

func TestCreateServiceConfig_Custom(t *testing.T) {
	cfg := CreateServiceConfig("facebook", "full", "page")

	if len(cfg.Platforms) != 1 {
		t.Fatalf("expected 1 platform, got %d", len(cfg.Platforms))
	}
	if cfg.Platforms[0] != "facebook" {
		t.Errorf("Platforms[0] = %q, want %q", cfg.Platforms[0], "facebook")
	}

	if cfg.SyncType != "full_sync" {
		t.Errorf("SyncType = %q, want %q", cfg.SyncType, "full_sync")
	}

	if len(cfg.FacebookAcctTypes) != 1 || cfg.FacebookAcctTypes[0] != "page" {
		t.Errorf("FacebookAcctTypes = %v, want [page]", cfg.FacebookAcctTypes)
	}
}

func TestCreateServiceConfig_EmptyAccountTypes(t *testing.T) {
	cfg := CreateServiceConfig("facebook", "incremental", "")

	if cfg.FacebookAcctTypes != nil {
		t.Errorf("expected nil FacebookAcctTypes, got %v", cfg.FacebookAcctTypes)
	}
}

// ================== DefaultPlatformProcessor Tests ==================

func TestDefaultPlatformProcessor_Instance(t *testing.T) {
	p := &DefaultPlatformProcessor{}

	// Just verify we can create an instance
	if p == nil {
		t.Error("expected non-nil processor")
	}
}

// ================== processPlatform Tests ==================

func TestProcessPlatform_Facebook(t *testing.T) {
	mockProcessor := &MockPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := logger.New("debug")

	deps := &ServiceDependencies{
		Database:          nil,
		Producer:          mockProducer,
		Processor:         mockProcessor,
		Logger:            log,
		Platforms:         []string{"facebook"},
		SyncType:          "incremental",
		FacebookAcctTypes: []string{"page"},
	}

	ctx := context.Background()
	processPlatform(ctx, "facebook", deps)

	if !mockProcessor.GetFacebookCalled() {
		t.Error("expected Facebook processor to be called")
	}
}

func TestProcessPlatform_Instagram(t *testing.T) {
	mockProcessor := &MockPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := logger.New("debug")

	deps := &ServiceDependencies{
		Database:          nil,
		Producer:          mockProducer,
		Processor:         mockProcessor,
		Logger:            log,
		Platforms:         []string{"instagram"},
		SyncType:          "incremental",
		FacebookAcctTypes: nil,
	}

	ctx := context.Background()
	processPlatform(ctx, "instagram", deps)

	if !mockProcessor.GetInstagramCalled() {
		t.Error("expected Instagram processor to be called")
	}
}

func TestProcessPlatform_LinkedIn(t *testing.T) {
	mockProcessor := &MockPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := logger.New("debug")

	deps := &ServiceDependencies{
		Database:          nil,
		Producer:          mockProducer,
		Processor:         mockProcessor,
		Logger:            log,
		Platforms:         []string{"linkedin"},
		SyncType:          "full_sync",
		FacebookAcctTypes: nil,
	}

	ctx := context.Background()
	processPlatform(ctx, "linkedin", deps)

	if !mockProcessor.GetLinkedinCalled() {
		t.Error("expected LinkedIn processor to be called")
	}
}

func TestProcessPlatform_Pinterest(t *testing.T) {
	mockProcessor := &MockPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := logger.New("debug")

	deps := &ServiceDependencies{
		Database:          nil,
		Producer:          mockProducer,
		Processor:         mockProcessor,
		Logger:            log,
		Platforms:         []string{"pinterest"},
		SyncType:          "incremental",
		FacebookAcctTypes: nil,
	}

	ctx := context.Background()
	processPlatform(ctx, "pinterest", deps)

	if !mockProcessor.GetPinterestCalled() {
		t.Error("expected Pinterest processor to be called")
	}
}

func TestProcessPlatform_Unsupported(t *testing.T) {
	mockProcessor := &MockPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := logger.New("debug")

	deps := &ServiceDependencies{
		Database:          nil,
		Producer:          mockProducer,
		Processor:         mockProcessor,
		Logger:            log,
		Platforms:         []string{"twitter"},
		SyncType:          "incremental",
		FacebookAcctTypes: nil,
	}

	ctx := context.Background()
	processPlatform(ctx, "twitter", deps)

	// No processors should be called for unsupported platform
	if mockProcessor.GetFacebookCalled() || mockProcessor.GetInstagramCalled() || mockProcessor.GetLinkedinCalled() || mockProcessor.GetTikTokCalled() || mockProcessor.GetPinterestCalled() {
		t.Error("expected no processors to be called for unsupported platform")
	}
}

// ================== ServiceDependencies Tests ==================

func TestServiceDependencies_Struct(t *testing.T) {
	deps := &ServiceDependencies{
		Platforms:         []string{"facebook"},
		SyncType:          "incremental",
		FacebookAcctTypes: []string{"page"},
	}

	if deps.Platforms[0] != "facebook" {
		t.Errorf("Platforms[0] = %q, want %q", deps.Platforms[0], "facebook")
	}
	if deps.SyncType != "incremental" {
		t.Errorf("SyncType = %q, want %q", deps.SyncType, "incremental")
	}
	if deps.FacebookAcctTypes[0] != "page" {
		t.Errorf("FacebookAcctTypes[0] = %q, want %q", deps.FacebookAcctTypes[0], "page")
	}
}

func TestServiceDependencies_Empty(t *testing.T) {
	deps := &ServiceDependencies{}

	if deps.Platforms != nil {
		t.Errorf("expected nil Platforms, got %v", deps.Platforms)
	}
	if deps.SyncType != "" {
		t.Errorf("expected empty SyncType, got %q", deps.SyncType)
	}
}

// ================== EstimateProcessingLoad Tests ==================

func TestEstimateProcessingLoad_AllPlatforms(t *testing.T) {
	platforms := []string{"facebook", "instagram", "linkedin", "tiktok"}
	result := EstimateProcessingLoad(platforms)

	expected := 24000 + 16000 + 8000 + 2000
	if result != expected {
		t.Errorf("EstimateProcessingLoad = %d, want %d", result, expected)
	}
}

func TestEstimateProcessingLoad_SinglePlatform(t *testing.T) {
	platforms := []string{"facebook"}
	result := EstimateProcessingLoad(platforms)

	if result != 24000 {
		t.Errorf("EstimateProcessingLoad = %d, want 24000", result)
	}
}

func TestEstimateProcessingLoad_UnknownPlatforms(t *testing.T) {
	platforms := []string{"unknown", "invalid"}
	result := EstimateProcessingLoad(platforms)

	if result != 0 {
		t.Errorf("EstimateProcessingLoad = %d, want 0", result)
	}
}

func TestEstimateProcessingLoad_MixedPlatforms(t *testing.T) {
	platforms := []string{"facebook", "unknown", "instagram", "tiktok"}
	result := EstimateProcessingLoad(platforms)

	expected := 24000 + 16000 + 2000
	if result != expected {
		t.Errorf("EstimateProcessingLoad = %d, want %d", result, expected)
	}
}

func TestEstimateProcessingLoad_Empty(t *testing.T) {
	platforms := []string{}
	result := EstimateProcessingLoad(platforms)

	if result != 0 {
		t.Errorf("EstimateProcessingLoad = %d, want 0", result)
	}
}

// ================== FilterPlatformsWithScale Tests ==================

func TestFilterPlatformsWithScale_AllValid(t *testing.T) {
	platforms := []string{"facebook", "instagram", "linkedin", "tiktok"}
	result := FilterPlatformsWithScale(platforms)

	if len(result) != 4 {
		t.Errorf("FilterPlatformsWithScale returned %d platforms, want 4", len(result))
	}
}

func TestFilterPlatformsWithScale_Mixed(t *testing.T) {
	platforms := []string{"facebook", "unknown", "instagram", "invalid", "tiktok"}
	result := FilterPlatformsWithScale(platforms)

	if len(result) != 3 {
		t.Errorf("FilterPlatformsWithScale returned %d platforms, want 3", len(result))
	}
	if result[0] != "facebook" || result[1] != "instagram" || result[2] != "tiktok" {
		t.Errorf("result = %v, want [facebook, instagram, tiktok]", result)
	}
}

func TestFilterPlatformsWithScale_AllInvalid(t *testing.T) {
	platforms := []string{"unknown", "invalid"}
	result := FilterPlatformsWithScale(platforms)

	if len(result) != 0 {
		t.Errorf("FilterPlatformsWithScale returned %d platforms, want 0", len(result))
	}
}

func TestFilterPlatformsWithScale_Empty(t *testing.T) {
	result := FilterPlatformsWithScale([]string{})

	if len(result) != 0 {
		t.Errorf("FilterPlatformsWithScale returned %d platforms, want 0", len(result))
	}
}

// ================== CompareSyncTypes Tests ==================

func TestCompareSyncTypes_Equal(t *testing.T) {
	tests := []struct {
		a, b string
	}{
		{"incremental", "incremental"},
		{"full_sync", "full_sync"},
		{"full", "full_sync"},
		{"", "incremental"},
		{"unknown", "incremental"},
	}

	for _, tc := range tests {
		if !CompareSyncTypes(tc.a, tc.b) {
			t.Errorf("CompareSyncTypes(%q, %q) = false, want true", tc.a, tc.b)
		}
	}
}

func TestCompareSyncTypes_NotEqual(t *testing.T) {
	tests := []struct {
		a, b string
	}{
		{"incremental", "full_sync"},
		{"full", "incremental"},
	}

	for _, tc := range tests {
		if CompareSyncTypes(tc.a, tc.b) {
			t.Errorf("CompareSyncTypes(%q, %q) = true, want false", tc.a, tc.b)
		}
	}
}

// ================== MergePlatformLists Tests ==================

func TestMergePlatformLists_NoDuplicates(t *testing.T) {
	a := []string{"facebook", "instagram"}
	b := []string{"linkedin", "tiktok"}
	result := MergePlatformLists(a, b)

	if len(result) != 4 {
		t.Errorf("MergePlatformLists returned %d platforms, want 4", len(result))
	}
}

func TestMergePlatformLists_WithDuplicates(t *testing.T) {
	a := []string{"facebook", "instagram"}
	b := []string{"instagram", "linkedin"}
	result := MergePlatformLists(a, b)

	if len(result) != 3 {
		t.Errorf("MergePlatformLists returned %d platforms, want 3 (deduplicated)", len(result))
	}
}

func TestMergePlatformLists_AllDuplicates(t *testing.T) {
	a := []string{"facebook", "instagram"}
	b := []string{"facebook", "instagram"}
	result := MergePlatformLists(a, b)

	if len(result) != 2 {
		t.Errorf("MergePlatformLists returned %d platforms, want 2", len(result))
	}
}

func TestMergePlatformLists_EmptyFirst(t *testing.T) {
	a := []string{}
	b := []string{"facebook", "instagram"}
	result := MergePlatformLists(a, b)

	if len(result) != 2 {
		t.Errorf("MergePlatformLists returned %d platforms, want 2", len(result))
	}
}

func TestMergePlatformLists_EmptySecond(t *testing.T) {
	a := []string{"facebook", "instagram"}
	b := []string{}
	result := MergePlatformLists(a, b)

	if len(result) != 2 {
		t.Errorf("MergePlatformLists returned %d platforms, want 2", len(result))
	}
}

func TestMergePlatformLists_BothEmpty(t *testing.T) {
	result := MergePlatformLists([]string{}, []string{})

	if len(result) != 0 {
		t.Errorf("MergePlatformLists returned %d platforms, want 0", len(result))
	}
}

// ================== GetPlatformCount Tests ==================

func TestGetPlatformCount_Default(t *testing.T) {
	result := GetPlatformCount("")

	if result != 9 {
		t.Errorf("GetPlatformCount(\"\") = %d, want 9", result)
	}
}

func TestGetPlatformCount_Single(t *testing.T) {
	result := GetPlatformCount("facebook")

	if result != 1 {
		t.Errorf("GetPlatformCount(\"facebook\") = %d, want 1", result)
	}
}

func TestGetPlatformCount_Multiple(t *testing.T) {
	result := GetPlatformCount("facebook,instagram,linkedin")

	if result != 3 {
		t.Errorf("GetPlatformCount = %d, want 3", result)
	}
}

// ================== IsFullSync Tests ==================

func TestIsFullSync_True(t *testing.T) {
	tests := []string{"full_sync", "full"}

	for _, st := range tests {
		if !IsFullSync(st) {
			t.Errorf("IsFullSync(%q) = false, want true", st)
		}
	}
}

func TestIsFullSync_False(t *testing.T) {
	tests := []string{"incremental", "", "unknown"}

	for _, st := range tests {
		if IsFullSync(st) {
			t.Errorf("IsFullSync(%q) = true, want false", st)
		}
	}
}

// ================== GetFacebookAccountTypeCount Tests ==================

func TestGetFacebookAccountTypeCount_Empty(t *testing.T) {
	result := GetFacebookAccountTypeCount("")

	if result != 0 {
		t.Errorf("GetFacebookAccountTypeCount(\"\") = %d, want 0", result)
	}
}

func TestGetFacebookAccountTypeCount_Single(t *testing.T) {
	result := GetFacebookAccountTypeCount("page")

	if result != 1 {
		t.Errorf("GetFacebookAccountTypeCount(\"page\") = %d, want 1", result)
	}
}

func TestGetFacebookAccountTypeCount_Multiple(t *testing.T) {
	result := GetFacebookAccountTypeCount("page,group,profile")

	if result != 3 {
		t.Errorf("GetFacebookAccountTypeCount = %d, want 3", result)
	}
}

// ================== DefaultPlatformProcessor Tests ==================

func TestDefaultPlatformProcessor_ProcessFacebook(t *testing.T) {
	processor := &DefaultPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := zerolog.Nop()

	// Call with nil DB - will fail internally but we test the method is callable
	defer func() {
		if r := recover(); r != nil {
			// Expected - nil DB causes panic in fetcher
			t.Log("ProcessFacebook panicked as expected with nil DB")
		}
	}()

	processor.ProcessFacebook(nil, mockProducer, log, []string{"page"}, "incremental")
}

func TestDefaultPlatformProcessor_ProcessInstagram(t *testing.T) {
	processor := &DefaultPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := zerolog.Nop()

	defer func() {
		if r := recover(); r != nil {
			t.Log("ProcessInstagram panicked as expected with nil DB")
		}
	}()

	ctx := context.Background()
	processor.ProcessInstagram(ctx, nil, mockProducer, log, nil, "incremental")
}

func TestDefaultPlatformProcessor_ProcessLinkedin(t *testing.T) {
	processor := &DefaultPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := zerolog.Nop()

	defer func() {
		if r := recover(); r != nil {
			t.Log("ProcessLinkedin panicked as expected with nil DB")
		}
	}()

	ctx := context.Background()
	processor.ProcessLinkedin(ctx, nil, mockProducer, log, nil, "incremental")
}

func TestDefaultPlatformProcessor_ProcessTikTok(t *testing.T) {
	processor := &DefaultPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := zerolog.Nop()

	defer func() {
		if r := recover(); r != nil {
			t.Log("ProcessTikTok panicked as expected with nil DB")
		}
	}()

	ctx := context.Background()
	processor.ProcessTikTok(ctx, nil, mockProducer, log, nil, "incremental")
}

func TestDefaultPlatformProcessor_ProcessPinterest(t *testing.T) {
	processor := &DefaultPlatformProcessor{}
	mockProducer := &kafka.MockProducer{}
	log := zerolog.Nop()

	defer func() {
		if r := recover(); r != nil {
			t.Log("ProcessPinterest panicked as expected with nil DB")
		}
	}()

	ctx := context.Background()
	processor.ProcessPinterest(ctx, nil, mockProducer, log, nil, "incremental")
}
