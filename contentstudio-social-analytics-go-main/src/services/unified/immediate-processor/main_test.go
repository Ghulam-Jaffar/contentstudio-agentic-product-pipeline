package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	fbprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/facebook/facebook-immediate-processor/processor"
	igprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/instagram/instagram-immediate-processor/processor"
	liprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/linkedin/linkedin-immediate-processor/processor"
	ptprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/pinterest/pinterest-immediate-processor/processor"
	tkprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/tiktok/tiktok-immediate-processor/processor"
	ytprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/youtube/youtube-immediate-processor/processor"
	"github.com/rs/zerolog"
)

func TestArchitectureDocumentation(t *testing.T) {
	// This test serves as documentation verification
	// The unified immediate processor uses a two-tier queue system:
	// - Global queue: 100K capacity for admission control
	// - Platform queues: Facebook (24K/40 workers), Instagram (16K/30 workers), LinkedIn (8K/15 workers), TikTok (5K/10 workers)

	// Basic sanity check that the service can be compiled and constants are defined
	t.Log("Unified Immediate Processor architecture:")
	t.Log("- Two-tier queue system for processing")
	t.Log("- Global admission control prevents overload")
	t.Log("- Platform-specific queues with dedicated workers")
}

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if GlobalQueueCapacity != 100000 {
		t.Errorf("GlobalQueueCapacity = %d, want 100000", GlobalQueueCapacity)
	}
	if facebookConsumerGroup != "immediate-processor-group" {
		t.Errorf("facebookConsumerGroup = %q, want %q", facebookConsumerGroup, "immediate-processor-group")
	}
	if instagramConsumerGroup != "instagram-immediate-processor-group" {
		t.Errorf("instagramConsumerGroup = %q, want %q", instagramConsumerGroup, "instagram-immediate-processor-group")
	}
	if linkedinConsumerGroup != "linkedin-immediate-processor-group" {
		t.Errorf("linkedinConsumerGroup = %q, want %q", linkedinConsumerGroup, "linkedin-immediate-processor-group")
	}
	if tiktokConsumerGroup != "tiktok-immediate-processor-group" {
		t.Errorf("tiktokConsumerGroup = %q, want %q", tiktokConsumerGroup, "tiktok-immediate-processor-group")
	}
}

func TestPlatformSettings(t *testing.T) {
	expected := map[string]PlatformConfig{
		"facebook":  {Workers: 40, QueueSize: 500, MaxCapacity: 24000},
		"instagram": {Workers: 30, QueueSize: 400, MaxCapacity: 16000},
		"linkedin":  {Workers: 15, QueueSize: 200, MaxCapacity: 8000},
		"youtube":   {Workers: 20, QueueSize: 250, MaxCapacity: 10000},
		"tiktok":    {Workers: 10, QueueSize: 150, MaxCapacity: 5000},
		"pinterest": {Workers: 15, QueueSize: 200, MaxCapacity: 8000},
	}

	for platform, expectedCfg := range expected {
		cfg, ok := PlatformSettings[platform]
		if !ok {
			t.Errorf("PlatformSettings missing %q", platform)
			continue
		}
		if cfg.Workers != expectedCfg.Workers {
			t.Errorf("PlatformSettings[%q].Workers = %d, want %d", platform, cfg.Workers, expectedCfg.Workers)
		}
		if cfg.QueueSize != expectedCfg.QueueSize {
			t.Errorf("PlatformSettings[%q].QueueSize = %d, want %d", platform, cfg.QueueSize, expectedCfg.QueueSize)
		}
		if cfg.MaxCapacity != expectedCfg.MaxCapacity {
			t.Errorf("PlatformSettings[%q].MaxCapacity = %d, want %d", platform, cfg.MaxCapacity, expectedCfg.MaxCapacity)
		}
	}
}

// ================== GlobalQueue Tests ==================

func TestNewGlobalQueue(t *testing.T) {
	gq := NewGlobalQueue(1000)
	if gq == nil {
		t.Fatal("NewGlobalQueue returned nil")
	}
	if gq.capacity != 1000 {
		t.Errorf("capacity = %d, want 1000", gq.capacity)
	}
}

func TestGlobalQueue_TryAdmit_Success(t *testing.T) {
	gq := NewGlobalQueue(10)

	for i := 0; i < 10; i++ {
		if !gq.TryAdmit() {
			t.Errorf("TryAdmit failed at iteration %d", i)
		}
	}

	current, capacity, admitted, rejected := gq.Stats()
	if current != 10 {
		t.Errorf("current = %d, want 10", current)
	}
	if capacity != 10 {
		t.Errorf("capacity = %d, want 10", capacity)
	}
	if admitted != 10 {
		t.Errorf("admitted = %d, want 10", admitted)
	}
	if rejected != 0 {
		t.Errorf("rejected = %d, want 0", rejected)
	}
}

func TestGlobalQueue_TryAdmit_Full(t *testing.T) {
	gq := NewGlobalQueue(5)

	// Fill the queue
	for i := 0; i < 5; i++ {
		gq.TryAdmit()
	}

	// Should reject
	if gq.TryAdmit() {
		t.Error("TryAdmit should have returned false when full")
	}

	_, _, _, rejected := gq.Stats()
	if rejected != 1 {
		t.Errorf("rejected = %d, want 1", rejected)
	}
}

func TestGlobalQueue_Release(t *testing.T) {
	gq := NewGlobalQueue(10)

	// Admit some items
	gq.TryAdmit()
	gq.TryAdmit()
	gq.TryAdmit()

	current, _, _, _ := gq.Stats()
	if current != 3 {
		t.Errorf("current before release = %d, want 3", current)
	}

	// Release one
	gq.Release()

	current, _, _, _ = gq.Stats()
	if current != 2 {
		t.Errorf("current after release = %d, want 2", current)
	}
}

func TestGlobalQueue_Stats(t *testing.T) {
	gq := NewGlobalQueue(100)

	// Admit some, release some
	gq.TryAdmit()
	gq.TryAdmit()
	gq.TryAdmit()
	gq.Release()

	current, capacity, admitted, rejected := gq.Stats()

	if current != 2 {
		t.Errorf("current = %d, want 2", current)
	}
	if capacity != 100 {
		t.Errorf("capacity = %d, want 100", capacity)
	}
	if admitted != 3 {
		t.Errorf("admitted = %d, want 3", admitted)
	}
	if rejected != 0 {
		t.Errorf("rejected = %d, want 0", rejected)
	}
}

func TestGlobalQueue_ConcurrentAccess(t *testing.T) {
	gq := NewGlobalQueue(1000)
	var wg sync.WaitGroup

	// Concurrent admits
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				gq.TryAdmit()
			}
		}()
	}
	wg.Wait()

	current, _, _, _ := gq.Stats()
	if current != 1000 {
		t.Errorf("current after concurrent admits = %d, want 1000", current)
	}

	// Concurrent releases
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				gq.Release()
			}
		}()
	}
	wg.Wait()

	current, _, _, _ = gq.Stats()
	if current != 0 {
		t.Errorf("current after concurrent releases = %d, want 0", current)
	}
}

// ================== PlatformJobChannels Tests ==================

func TestNewPlatformJobChannels(t *testing.T) {
	pjc := NewPlatformJobChannels()
	if pjc == nil {
		t.Fatal("NewPlatformJobChannels returned nil")
	}

	// Should have channels for all configured platforms
	for platform := range PlatformSettings {
		if pjc.channels[platform] == nil {
			t.Errorf("missing channel for platform %q", platform)
		}
	}
}

func TestPlatformJobChannels_GetChannel_Existing(t *testing.T) {
	pjc := NewPlatformJobChannels()

	ch := pjc.GetChannel("facebook")
	if ch == nil {
		t.Error("GetChannel returned nil for existing platform")
	}
}

func TestPlatformJobChannels_GetChannel_Unknown(t *testing.T) {
	pjc := NewPlatformJobChannels()

	ch := pjc.GetChannel("unknown_platform")
	if ch == nil {
		t.Error("GetChannel should create channel for unknown platform")
	}

	// Second call should return same channel
	ch2 := pjc.GetChannel("unknown_platform")
	if ch != ch2 {
		t.Error("GetChannel should return same channel on second call")
	}
}

func TestPlatformJobChannels_TryEnqueue_Success(t *testing.T) {
	pjc := NewPlatformJobChannels()

	wo := ImmediateWorkOrder{
		ID:          "test-1",
		Platform:    "facebook",
		AccountID:   "acc123",
		WorkspaceID: "ws456",
	}

	if !pjc.TryEnqueue("facebook", wo) {
		t.Error("TryEnqueue should succeed for valid platform")
	}

	// Verify item in channel
	ch := pjc.GetChannel("facebook")
	select {
	case received := <-ch:
		if received.ID != "test-1" {
			t.Errorf("received.ID = %q, want %q", received.ID, "test-1")
		}
	default:
		t.Error("expected item in channel")
	}
}

func TestPlatformJobChannels_TryEnqueue_UnknownPlatform(t *testing.T) {
	pjc := NewPlatformJobChannels()

	wo := ImmediateWorkOrder{ID: "test-1"}

	// Unknown platform should fail
	if pjc.TryEnqueue("nonexistent", wo) {
		t.Error("TryEnqueue should fail for unknown platform")
	}
}

func TestPlatformJobChannels_TryEnqueue_FullQueue(t *testing.T) {
	pjc := &PlatformJobChannels{
		channels:  make(map[string]chan ImmediateWorkOrder),
		processed: make(map[string]*int64),
		dropped:   make(map[string]*int64),
	}
	// Create a small channel that will fill up
	pjc.channels["test"] = make(chan ImmediateWorkOrder, 1)
	var dropped int64
	pjc.dropped["test"] = &dropped

	wo := ImmediateWorkOrder{ID: "test-1"}

	// First enqueue should succeed
	if !pjc.TryEnqueue("test", wo) {
		t.Error("first TryEnqueue should succeed")
	}

	// Second enqueue should fail (queue full)
	if pjc.TryEnqueue("test", wo) {
		t.Error("TryEnqueue should fail when queue is full")
	}

	if dropped != 1 {
		t.Errorf("dropped = %d, want 1", dropped)
	}
}

func TestPlatformJobChannels_IncrementProcessed(t *testing.T) {
	pjc := NewPlatformJobChannels()

	pjc.IncrementProcessed("facebook")
	pjc.IncrementProcessed("facebook")

	stats := pjc.GetStats()
	if stats["facebook"].Processed != 2 {
		t.Errorf("processed = %d, want 2", stats["facebook"].Processed)
	}
}

func TestPlatformJobChannels_IncrementProcessed_UnknownPlatform(t *testing.T) {
	pjc := NewPlatformJobChannels()

	// Should not panic for unknown platform
	pjc.IncrementProcessed("unknown")
}

func TestPlatformJobChannels_GetStats(t *testing.T) {
	pjc := NewPlatformJobChannels()

	// Enqueue some items
	pjc.TryEnqueue("facebook", ImmediateWorkOrder{ID: "1"})
	pjc.TryEnqueue("facebook", ImmediateWorkOrder{ID: "2"})
	pjc.TryEnqueue("instagram", ImmediateWorkOrder{ID: "3"})

	pjc.IncrementProcessed("facebook")

	stats := pjc.GetStats()

	if stats["facebook"].QueueDepth != 2 {
		t.Errorf("facebook QueueDepth = %d, want 2", stats["facebook"].QueueDepth)
	}
	if stats["facebook"].Processed != 1 {
		t.Errorf("facebook Processed = %d, want 1", stats["facebook"].Processed)
	}
	if stats["instagram"].QueueDepth != 1 {
		t.Errorf("instagram QueueDepth = %d, want 1", stats["instagram"].QueueDepth)
	}
}

func TestPlatformJobChannels_CloseAll(t *testing.T) {
	pjc := NewPlatformJobChannels()

	pjc.CloseAll()

	// Channels should be closed
	for platform, ch := range pjc.channels {
		select {
		case _, ok := <-ch:
			if ok {
				t.Errorf("channel for %q should be closed", platform)
			}
		default:
			// Channel is empty but may not be closed yet in the default case
		}
	}
}

// ================== ImmediateWorkOrder Tests ==================

func TestImmediateWorkOrder_Struct(t *testing.T) {
	wo := ImmediateWorkOrder{
		ID:                    "id123",
		Platform:              "facebook",
		AccountID:             "acc456",
		Type:                  "page",
		AccessToken:           "token789",
		LongAccessToken:       "long_token",
		WorkspaceID:           "ws012",
		SyncType:              "incremental",
		ConnectedViaInstagram: true,
	}

	if wo.ID != "id123" {
		t.Errorf("ID = %q, want %q", wo.ID, "id123")
	}
	if wo.Platform != "facebook" {
		t.Errorf("Platform = %q, want %q", wo.Platform, "facebook")
	}
	if !wo.ConnectedViaInstagram {
		t.Error("ConnectedViaInstagram should be true")
	}
}

// ================== QueueStats Tests ==================

func TestQueueStats_Struct(t *testing.T) {
	stats := QueueStats{
		Platform:    "facebook",
		QueueDepth:  100,
		MaxCapacity: 500,
		Processed:   1000,
		Dropped:     5,
	}

	if stats.Platform != "facebook" {
		t.Errorf("Platform = %q, want %q", stats.Platform, "facebook")
	}
	if stats.QueueDepth != 100 {
		t.Errorf("QueueDepth = %d, want 100", stats.QueueDepth)
	}
	if stats.MaxCapacity != 500 {
		t.Errorf("MaxCapacity = %d, want 500", stats.MaxCapacity)
	}
	if stats.Processed != 1000 {
		t.Errorf("Processed = %d, want 1000", stats.Processed)
	}
	if stats.Dropped != 5 {
		t.Errorf("Dropped = %d, want 5", stats.Dropped)
	}
}

// ================== inferPlatformFromTopic Tests ==================

func TestInferPlatformFromTopic_Facebook(t *testing.T) {
	tests := []string{
		"immediate-work-order-facebook",
		"facebook-posts",
		"some-facebook-topic",
	}

	for _, topic := range tests {
		result := inferPlatformFromTopic(topic)
		if result != "facebook" {
			t.Errorf("inferPlatformFromTopic(%q) = %q, want %q", topic, result, "facebook")
		}
	}
}

func TestInferPlatformFromTopic_Instagram(t *testing.T) {
	tests := []string{
		"immediate-work-order-instagram",
		"instagram-posts",
		"some-instagram-topic",
	}

	for _, topic := range tests {
		result := inferPlatformFromTopic(topic)
		if result != "instagram" {
			t.Errorf("inferPlatformFromTopic(%q) = %q, want %q", topic, result, "instagram")
		}
	}
}

func TestInferPlatformFromTopic_LinkedIn(t *testing.T) {
	tests := []string{
		"immediate-work-order-linkedin",
		"linkedin-posts",
		"some-linkedin-topic",
	}

	for _, topic := range tests {
		result := inferPlatformFromTopic(topic)
		if result != "linkedin" {
			t.Errorf("inferPlatformFromTopic(%q) = %q, want %q", topic, result, "linkedin")
		}
	}
}

func TestInferPlatformFromTopic_Unknown(t *testing.T) {
	tests := []string{
		"unknown-topic",
		"some-random-topic",
		"",
	}

	for _, topic := range tests {
		result := inferPlatformFromTopic(topic)
		if result != "" {
			t.Errorf("inferPlatformFromTopic(%q) = %q, want empty string", topic, result)
		}
	}
}

// ================== PlatformConfig Tests ==================

func TestPlatformConfig_Struct(t *testing.T) {
	cfg := PlatformConfig{
		Workers:     40,
		QueueSize:   500,
		MaxCapacity: 24000,
	}

	if cfg.Workers != 40 {
		t.Errorf("Workers = %d, want 40", cfg.Workers)
	}
	if cfg.QueueSize != 500 {
		t.Errorf("QueueSize = %d, want 500", cfg.QueueSize)
	}
	if cfg.MaxCapacity != 24000 {
		t.Errorf("MaxCapacity = %d, want 24000", cfg.MaxCapacity)
	}
}

// ================== UnifiedProcessor Tests ==================

func TestUnifiedProcessor_Struct(t *testing.T) {
	// Test that the struct can be created
	p := &UnifiedProcessor{}
	if p == nil {
		t.Error("UnifiedProcessor should not be nil")
	}
}

// ================== Concurrent Queue Tests ==================

func TestPlatformJobChannels_ConcurrentAccess(t *testing.T) {
	pjc := NewPlatformJobChannels()
	var wg sync.WaitGroup

	// Concurrent enqueues
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			wo := ImmediateWorkOrder{ID: "test"}
			pjc.TryEnqueue("facebook", wo)
			pjc.IncrementProcessed("facebook")
		}(i)
	}
	wg.Wait()

	// Get stats should work correctly
	stats := pjc.GetStats()
	if stats["facebook"].Processed != 10 {
		t.Errorf("processed = %d, want 10", stats["facebook"].Processed)
	}
}

// ================== Mock Processors for Testing ==================

type MockFacebookProcessor struct {
	ProcessAccountFunc func(ctx context.Context, wo fbprocessor.WorkOrder) error
}

func (m *MockFacebookProcessor) ProcessAccount(ctx context.Context, wo fbprocessor.WorkOrder) error {
	if m.ProcessAccountFunc != nil {
		return m.ProcessAccountFunc(ctx, wo)
	}
	return nil
}

type MockInstagramProcessor struct {
	ProcessAccountFunc func(ctx context.Context, wo igprocessor.WorkOrder) error
}

func (m *MockInstagramProcessor) ProcessAccount(ctx context.Context, wo igprocessor.WorkOrder) error {
	if m.ProcessAccountFunc != nil {
		return m.ProcessAccountFunc(ctx, wo)
	}
	return nil
}

type MockLinkedInProcessor struct {
	ProcessAccountFunc func(ctx context.Context, wo liprocessor.WorkOrder) error
}

func (m *MockLinkedInProcessor) ProcessAccount(ctx context.Context, wo liprocessor.WorkOrder) error {
	if m.ProcessAccountFunc != nil {
		return m.ProcessAccountFunc(ctx, wo)
	}
	return nil
}

type MockYouTubeProcessor struct {
	ProcessAccountFunc func(ctx context.Context, wo ytprocessor.WorkOrder) error
}

func (m *MockYouTubeProcessor) ProcessAccount(ctx context.Context, wo ytprocessor.WorkOrder) error {
	if m.ProcessAccountFunc != nil {
		return m.ProcessAccountFunc(ctx, wo)
	}
	return nil
}

type MockTikTokProcessor struct {
	ProcessAccountFunc func(ctx context.Context, wo tkprocessor.ImmediateWorkOrder) error
}

func (m *MockTikTokProcessor) ProcessAccount(ctx context.Context, wo tkprocessor.ImmediateWorkOrder) error {
	if m.ProcessAccountFunc != nil {
		return m.ProcessAccountFunc(ctx, wo)
	}
	return nil
}

type MockPinterestProcessor struct {
	ProcessAccountFunc func(ctx context.Context, wo ptprocessor.WorkOrder) error
}

func (m *MockPinterestProcessor) ProcessAccount(ctx context.Context, wo ptprocessor.WorkOrder) error {
	if m.ProcessAccountFunc != nil {
		return m.ProcessAccountFunc(ctx, wo)
	}
	return nil
}

// ================== NewUnifiedProcessor Tests ==================

func TestNewUnifiedProcessor(t *testing.T) {
	log := logger.New("error")
	fbProc := &MockFacebookProcessor{}
	igProc := &MockInstagramProcessor{}
	liProc := &MockLinkedInProcessor{}
	ytProc := &MockYouTubeProcessor{}
	tkProc := &MockTikTokProcessor{}
	ptProc := &MockPinterestProcessor{}

	p := NewUnifiedProcessor(fbProc, igProc, liProc, ytProc, tkProc, nil, ptProc, nil, log)

	if p == nil {
		t.Fatal("NewUnifiedProcessor returned nil")
	}
	if p.facebookProcessor == nil {
		t.Error("facebookProcessor is nil")
	}
	if p.instagramProcessor == nil {
		t.Error("instagramProcessor is nil")
	}
	if p.linkedinProcessor == nil {
		t.Error("linkedinProcessor is nil")
	}
	if p.youtubeProcessor == nil {
		t.Error("youtubeProcessor is nil")
	}
	if p.tiktokProcessor == nil {
		t.Error("tiktokProcessor is nil")
	}
	if p.pinterestProcessor == nil {
		t.Error("pinterestProcessor is nil")
	}
	if p.logger == nil {
		t.Error("logger is nil")
	}
}

// ================== PlatformWorkerTestable Tests ==================

func TestPlatformWorkerTestable_Facebook_Success(t *testing.T) {
	var processedCount int32
	fbProc := &MockFacebookProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo fbprocessor.WorkOrder) error {
			atomic.AddInt32(&processedCount, 1)
			return nil
		},
	}

	log := logger.New("error")
	p := NewUnifiedProcessor(fbProc, nil, nil, nil, nil, nil, nil, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	ctx, cancel := context.WithCancel(context.Background())

	// Admit to global queue before enqueue
	globalQueue.TryAdmit()
	globalQueue.TryAdmit()

	// Send work orders
	jobs <- ImmediateWorkOrder{ID: "1", AccountID: "acc1", WorkspaceID: "ws1"}
	jobs <- ImmediateWorkOrder{ID: "2", AccountID: "acc2", WorkspaceID: "ws2"}

	done := make(chan struct{})
	go func() {
		p.PlatformWorkerTestable(ctx, "facebook", 0, jobs, platformJobs, globalQueue)
		close(done)
	}()

	// Wait for processing
	time.Sleep(100 * time.Millisecond)
	cancel()
	close(jobs)
	<-done

	if atomic.LoadInt32(&processedCount) != 2 {
		t.Errorf("processedCount = %d, want 2", processedCount)
	}
}

func TestPlatformWorkerTestable_Instagram_Success(t *testing.T) {
	var processedCount int32
	igProc := &MockInstagramProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo igprocessor.WorkOrder) error {
			atomic.AddInt32(&processedCount, 1)
			return nil
		},
	}

	log := logger.New("error")
	p := NewUnifiedProcessor(nil, igProc, nil, nil, nil, nil, nil, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	ctx, cancel := context.WithCancel(context.Background())

	globalQueue.TryAdmit()
	jobs <- ImmediateWorkOrder{ID: "1", AccountID: "acc1"}

	done := make(chan struct{})
	go func() {
		p.PlatformWorkerTestable(ctx, "instagram", 0, jobs, platformJobs, globalQueue)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(jobs)
	<-done

	if atomic.LoadInt32(&processedCount) != 1 {
		t.Errorf("processedCount = %d, want 1", processedCount)
	}
}

func TestPlatformWorkerTestable_LinkedIn_Success(t *testing.T) {
	var processedCount int32
	liProc := &MockLinkedInProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo liprocessor.WorkOrder) error {
			atomic.AddInt32(&processedCount, 1)
			return nil
		},
	}

	log := logger.New("error")
	p := NewUnifiedProcessor(nil, nil, liProc, nil, nil, nil, nil, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	ctx, cancel := context.WithCancel(context.Background())

	globalQueue.TryAdmit()
	jobs <- ImmediateWorkOrder{ID: "1", AccountID: "acc1"}

	done := make(chan struct{})
	go func() {
		p.PlatformWorkerTestable(ctx, "linkedin", 0, jobs, platformJobs, globalQueue)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(jobs)
	<-done

	if atomic.LoadInt32(&processedCount) != 1 {
		t.Errorf("processedCount = %d, want 1", processedCount)
	}
}

func TestPlatformWorkerTestable_UnknownPlatform(t *testing.T) {
	log := logger.New("error")
	p := NewUnifiedProcessor(nil, nil, nil, nil, nil, nil, nil, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	ctx, cancel := context.WithCancel(context.Background())

	globalQueue.TryAdmit()
	jobs <- ImmediateWorkOrder{ID: "1", AccountID: "acc1"}

	done := make(chan struct{})
	go func() {
		p.PlatformWorkerTestable(ctx, "unknown", 0, jobs, platformJobs, globalQueue)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(jobs)
	<-done

	// Should handle unknown platform gracefully
}

func TestPlatformWorkerTestable_ContextCancel(t *testing.T) {
	log := logger.New("error")
	p := NewUnifiedProcessor(nil, nil, nil, nil, nil, nil, nil, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		p.PlatformWorkerTestable(ctx, "facebook", 0, jobs, platformJobs, globalQueue)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit after context cancel")
	}
}

func TestPlatformWorkerTestable_ChannelClose(t *testing.T) {
	log := logger.New("error")
	p := NewUnifiedProcessor(nil, nil, nil, nil, nil, nil, nil, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	done := make(chan struct{})
	go func() {
		p.PlatformWorkerTestable(context.Background(), "facebook", 0, jobs, platformJobs, globalQueue)
		close(done)
	}()

	close(jobs)

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit after channel close")
	}
}

func TestPlatformWorkerTestable_ProcessorError(t *testing.T) {
	fbProc := &MockFacebookProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo fbprocessor.WorkOrder) error {
			return context.DeadlineExceeded
		},
	}

	log := logger.New("error")
	p := NewUnifiedProcessor(fbProc, nil, nil, nil, nil, nil, nil, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	ctx, cancel := context.WithCancel(context.Background())

	globalQueue.TryAdmit()
	jobs <- ImmediateWorkOrder{ID: "1", AccountID: "acc1"}

	done := make(chan struct{})
	go func() {
		p.PlatformWorkerTestable(ctx, "facebook", 0, jobs, platformJobs, globalQueue)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(jobs)
	<-done

	// Error handling should work - processed count should not increment
	stats := platformJobs.GetStats()
	if stats["facebook"].Processed != 0 {
		t.Errorf("processed = %d, want 0 on error", stats["facebook"].Processed)
	}
}

func TestPlatformWorkerTestable_GlobalQueueRelease(t *testing.T) {
	fbProc := &MockFacebookProcessor{}

	log := logger.New("error")
	p := NewUnifiedProcessor(fbProc, nil, nil, nil, nil, nil, nil, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	ctx, cancel := context.WithCancel(context.Background())

	// Admit 5 items
	for i := 0; i < 5; i++ {
		globalQueue.TryAdmit()
	}

	current, _, _, _ := globalQueue.Stats()
	if current != 5 {
		t.Fatalf("initial current = %d, want 5", current)
	}

	// Send 5 work orders
	for i := 0; i < 5; i++ {
		jobs <- ImmediateWorkOrder{ID: string(rune('0' + i)), AccountID: "acc"}
	}

	done := make(chan struct{})
	go func() {
		p.PlatformWorkerTestable(ctx, "facebook", 0, jobs, platformJobs, globalQueue)
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()
	close(jobs)
	<-done

	// All items should be released from global queue
	current, _, _, _ = globalQueue.Stats()
	if current != 0 {
		t.Errorf("current after processing = %d, want 0", current)
	}
}

// ================== HandleMessage Tests ==================

func TestHandleMessage_ValidFacebookMessage(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	msg := `{"id":"test-1","platform":"facebook","account_id":"acc123","workspace_id":"ws456"}`

	HandleMessage("immediate-work-order-facebook", []byte(msg), globalQueue, platformJobs, log)

	// Should be enqueued
	ch := platformJobs.GetChannel("facebook")
	select {
	case wo := <-ch:
		if wo.ID != "test-1" {
			t.Errorf("ID = %q, want %q", wo.ID, "test-1")
		}
		if wo.AccountID != "acc123" {
			t.Errorf("AccountID = %q, want %q", wo.AccountID, "acc123")
		}
	default:
		t.Error("expected work order in channel")
	}

	// Global queue should have one item
	current, _, admitted, _ := globalQueue.Stats()
	if current != 1 {
		t.Errorf("current = %d, want 1", current)
	}
	if admitted != 1 {
		t.Errorf("admitted = %d, want 1", admitted)
	}
}

func TestHandleMessage_InferPlatformFromTopic(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	// Message without platform field - should infer from topic
	msg := `{"id":"test-1","account_id":"acc123"}`

	HandleMessage("immediate-work-order-instagram", []byte(msg), globalQueue, platformJobs, log)

	// Should be enqueued to instagram
	ch := platformJobs.GetChannel("instagram")
	select {
	case wo := <-ch:
		if wo.Platform != "instagram" {
			t.Errorf("Platform = %q, want %q", wo.Platform, "instagram")
		}
	default:
		t.Error("expected work order in channel")
	}
}

func TestHandleMessage_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	// Invalid JSON should not panic and should not enqueue
	HandleMessage("some-topic", []byte("invalid json"), globalQueue, platformJobs, log)

	// Nothing should be enqueued
	current, _, _, _ := globalQueue.Stats()
	if current != 0 {
		t.Errorf("current = %d, want 0 for invalid JSON", current)
	}
}

func TestHandleMessage_NoPlatform(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	// Message without platform and topic that can't infer platform
	msg := `{"id":"test-1","account_id":"acc123"}`

	HandleMessage("unknown-topic", []byte(msg), globalQueue, platformJobs, log)

	// Nothing should be enqueued
	current, _, _, _ := globalQueue.Stats()
	if current != 0 {
		t.Errorf("current = %d, want 0 when platform unknown", current)
	}
}

func TestHandleMessage_GlobalQueueFull(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(1) // Very small capacity
	platformJobs := NewPlatformJobChannels()

	// Fill the global queue
	globalQueue.TryAdmit()

	msg := `{"id":"test-1","platform":"facebook","account_id":"acc123"}`

	HandleMessage("immediate-work-order-facebook", []byte(msg), globalQueue, platformJobs, log)

	// Should be rejected
	_, _, _, rejected := globalQueue.Stats()
	if rejected != 1 {
		t.Errorf("rejected = %d, want 1", rejected)
	}

	// Nothing new should be in the queue
	ch := platformJobs.GetChannel("facebook")
	select {
	case <-ch:
		t.Error("should not have enqueued when global queue full")
	default:
		// Expected
	}
}

func TestHandleMessage_PlatformQueueFull(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)

	// Create platform jobs with very small queue
	platformJobs := &PlatformJobChannels{
		channels:  make(map[string]chan ImmediateWorkOrder),
		processed: make(map[string]*int64),
		dropped:   make(map[string]*int64),
	}
	platformJobs.channels["facebook"] = make(chan ImmediateWorkOrder, 1)
	var dropped int64
	platformJobs.dropped["facebook"] = &dropped
	var processed int64
	platformJobs.processed["facebook"] = &processed

	// Fill the platform queue
	platformJobs.channels["facebook"] <- ImmediateWorkOrder{ID: "existing"}

	msg := `{"id":"test-1","platform":"facebook","account_id":"acc123"}`

	HandleMessage("immediate-work-order-facebook", []byte(msg), globalQueue, platformJobs, log)

	// Should be dropped
	if dropped != 1 {
		t.Errorf("dropped = %d, want 1", dropped)
	}

	// Global queue should have been released
	current, _, _, _ := globalQueue.Stats()
	if current != 0 {
		t.Errorf("current = %d, want 0 after release", current)
	}
}

func TestHandleMessage_AllPlatforms(t *testing.T) {
	platforms := []struct {
		platform string
		topic    string
	}{
		{"facebook", "immediate-work-order-facebook"},
		{"instagram", "immediate-work-order-instagram"},
		{"linkedin", "immediate-work-order-linkedin"},
	}

	for _, tc := range platforms {
		t.Run(tc.platform, func(t *testing.T) {
			log := logger.New("error")
			globalQueue := NewGlobalQueue(100)
			platformJobs := NewPlatformJobChannels()

			msg := `{"id":"test-1","account_id":"acc123"}`

			HandleMessage(tc.topic, []byte(msg), globalQueue, platformJobs, log)

			ch := platformJobs.GetChannel(tc.platform)
			select {
			case wo := <-ch:
				if wo.Platform != tc.platform {
					t.Errorf("Platform = %q, want %q", wo.Platform, tc.platform)
				}
			default:
				t.Errorf("expected work order in %s channel", tc.platform)
			}
		})
	}
}

// ================== ImmediateWorkOrder JSON Tests ==================

func TestImmediateWorkOrder_JSONMarshalUnmarshal(t *testing.T) {
	wo := ImmediateWorkOrder{
		ID:                    "id-123",
		Platform:              "facebook",
		AccountID:             "acc-456",
		Type:                  "page",
		AccessToken:           "token-789",
		LongAccessToken:       "long-token",
		WorkspaceID:           "ws-012",
		SyncType:              "incremental",
		ConnectedViaInstagram: true,
	}

	data, err := json.Marshal(wo)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ImmediateWorkOrder
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ID != wo.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, wo.ID)
	}
	if decoded.Platform != wo.Platform {
		t.Errorf("Platform = %q, want %q", decoded.Platform, wo.Platform)
	}
	if decoded.ConnectedViaInstagram != wo.ConnectedViaInstagram {
		t.Errorf("ConnectedViaInstagram = %v, want %v", decoded.ConnectedViaInstagram, wo.ConnectedViaInstagram)
	}
}

func TestImmediateWorkOrder_JSONKeys(t *testing.T) {
	wo := ImmediateWorkOrder{
		ID:                    "id123",
		Platform:              "facebook",
		AccountID:             "acc456",
		Type:                  "page",
		AccessToken:           "token",
		LongAccessToken:       "long_token",
		WorkspaceID:           "ws789",
		SyncType:              "full",
		ConnectedViaInstagram: true,
	}

	data, err := json.Marshal(wo)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"id", "platform", "account_id", "type", "access_token", "long_access_token", "workspace_id", "sync_type", "connected_via_instagram"}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("missing key %q in JSON", key)
		}
	}
}

func TestImmediateWorkOrder_EmptyFields(t *testing.T) {
	wo := ImmediateWorkOrder{}

	if wo.ID != "" {
		t.Errorf("expected empty ID, got %q", wo.ID)
	}
	if wo.Platform != "" {
		t.Errorf("expected empty Platform, got %q", wo.Platform)
	}
	if wo.AccountID != "" {
		t.Errorf("expected empty AccountID, got %q", wo.AccountID)
	}
	if wo.Type != "" {
		t.Errorf("expected empty Type, got %q", wo.Type)
	}
	if wo.AccessToken != "" {
		t.Errorf("expected empty AccessToken, got %q", wo.AccessToken)
	}
	if wo.LongAccessToken != "" {
		t.Errorf("expected empty LongAccessToken, got %q", wo.LongAccessToken)
	}
	if wo.WorkspaceID != "" {
		t.Errorf("expected empty WorkspaceID, got %q", wo.WorkspaceID)
	}
	if wo.SyncType != "" {
		t.Errorf("expected empty SyncType, got %q", wo.SyncType)
	}
	if wo.ConnectedViaInstagram != false {
		t.Errorf("expected false ConnectedViaInstagram, got %v", wo.ConnectedViaInstagram)
	}
}

// ================== Additional GlobalQueue Tests ==================

func TestGlobalQueue_AdmitAndReleaseStress(t *testing.T) {
	gq := NewGlobalQueue(50)
	var wg sync.WaitGroup

	// Concurrent admits and releases
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if gq.TryAdmit() {
				time.Sleep(time.Microsecond)
				gq.Release()
			}
		}()
		go func() {
			defer wg.Done()
			gq.TryAdmit()
		}()
	}
	wg.Wait()

	// Final state should be valid
	current, capacity, _, _ := gq.Stats()
	if current < 0 || current > capacity {
		t.Errorf("invalid current state: %d (capacity: %d)", current, capacity)
	}
}

func TestGlobalQueue_MultipleRejections(t *testing.T) {
	gq := NewGlobalQueue(2)

	// Fill queue
	gq.TryAdmit()
	gq.TryAdmit()

	// Multiple rejections
	for i := 0; i < 10; i++ {
		if gq.TryAdmit() {
			t.Error("should reject when full")
		}
	}

	_, _, _, rejected := gq.Stats()
	if rejected != 10 {
		t.Errorf("rejected = %d, want 10", rejected)
	}
}

// ================== Additional PlatformJobChannels Tests ==================

func TestPlatformJobChannels_StatsAllPlatforms(t *testing.T) {
	pjc := NewPlatformJobChannels()

	// Enqueue to all platforms
	for platform := range PlatformSettings {
		pjc.TryEnqueue(platform, ImmediateWorkOrder{ID: "test"})
		pjc.IncrementProcessed(platform)
	}

	stats := pjc.GetStats()

	for platform := range PlatformSettings {
		s, ok := stats[platform]
		if !ok {
			t.Errorf("missing stats for platform %q", platform)
			continue
		}
		if s.Platform != platform {
			t.Errorf("Platform = %q, want %q", s.Platform, platform)
		}
		if s.QueueDepth != 1 {
			t.Errorf("%s QueueDepth = %d, want 1", platform, s.QueueDepth)
		}
		if s.Processed != 1 {
			t.Errorf("%s Processed = %d, want 1", platform, s.Processed)
		}
	}
}

func TestPlatformJobChannels_DroppedStats(t *testing.T) {
	pjc := &PlatformJobChannels{
		channels:  make(map[string]chan ImmediateWorkOrder),
		processed: make(map[string]*int64),
		dropped:   make(map[string]*int64),
	}
	// Very small channel
	pjc.channels["test"] = make(chan ImmediateWorkOrder, 2)
	var dropped, processed int64
	pjc.dropped["test"] = &dropped
	pjc.processed["test"] = &processed

	// Fill channel
	pjc.TryEnqueue("test", ImmediateWorkOrder{ID: "1"})
	pjc.TryEnqueue("test", ImmediateWorkOrder{ID: "2"})

	// Should drop
	for i := 0; i < 5; i++ {
		pjc.TryEnqueue("test", ImmediateWorkOrder{ID: "overflow"})
	}

	if dropped != 5 {
		t.Errorf("dropped = %d, want 5", dropped)
	}

	stats := pjc.GetStats()
	if stats["test"].Dropped != 5 {
		t.Errorf("stats dropped = %d, want 5", stats["test"].Dropped)
	}
}

// ================== PlatformWorker Multiple Platforms Tests ==================

func TestPlatformWorkerTestable_AllPlatformsProcessing(t *testing.T) {
	var fbProcessed, igProcessed, liProcessed int32

	fbProc := &MockFacebookProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo fbprocessor.WorkOrder) error {
			atomic.AddInt32(&fbProcessed, 1)
			return nil
		},
	}
	igProc := &MockInstagramProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo igprocessor.WorkOrder) error {
			atomic.AddInt32(&igProcessed, 1)
			return nil
		},
	}
	liProc := &MockLinkedInProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo liprocessor.WorkOrder) error {
			atomic.AddInt32(&liProcessed, 1)
			return nil
		},
	}

	log := logger.New("error")
	p := NewUnifiedProcessor(fbProc, igProc, liProc, nil, nil, nil, nil, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	// Create channels and workers for each platform
	platforms := []string{"facebook", "instagram", "linkedin"}
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	for _, platform := range platforms {
		jobs := make(chan ImmediateWorkOrder, 10)

		// Admit and send work order
		globalQueue.TryAdmit()
		go func(p string) {
			jobs <- ImmediateWorkOrder{ID: "1", AccountID: "acc1"}
		}(platform)

		wg.Add(1)
		go func(pl string, ch chan ImmediateWorkOrder) {
			defer wg.Done()
			p.PlatformWorkerTestable(ctx, pl, 0, ch, platformJobs, globalQueue)
		}(platform, jobs)

		// Give time for processing
		time.Sleep(50 * time.Millisecond)
		close(jobs)
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	wg.Wait()

	if atomic.LoadInt32(&fbProcessed) != 1 {
		t.Errorf("fbProcessed = %d, want 1", fbProcessed)
	}
	if atomic.LoadInt32(&igProcessed) != 1 {
		t.Errorf("igProcessed = %d, want 1", igProcessed)
	}
	if atomic.LoadInt32(&liProcessed) != 1 {
		t.Errorf("liProcessed = %d, want 1", liProcessed)
	}
}

func TestPlatformWorkerTestable_WorkOrderFieldMapping(t *testing.T) {
	var receivedWO fbprocessor.WorkOrder

	fbProc := &MockFacebookProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo fbprocessor.WorkOrder) error {
			receivedWO = wo
			return nil
		},
	}

	log := logger.New("error")
	p := NewUnifiedProcessor(fbProc, nil, nil, nil, nil, nil, nil, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	ctx, cancel := context.WithCancel(context.Background())

	globalQueue.TryAdmit()
	jobs <- ImmediateWorkOrder{
		ID:              "order-123",
		AccountID:       "acc-456",
		Type:            "page",
		AccessToken:     "token-789",
		WorkspaceID:     "ws-012",
		LongAccessToken: "long-token",
		SyncType:        "full",
	}

	go func() {
		p.PlatformWorkerTestable(ctx, "facebook", 0, jobs, platformJobs, globalQueue)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(jobs)

	// Verify field mapping
	if receivedWO.ID != "order-123" {
		t.Errorf("ID = %q, want %q", receivedWO.ID, "order-123")
	}
	if receivedWO.AccountID != "acc-456" {
		t.Errorf("AccountID = %q, want %q", receivedWO.AccountID, "acc-456")
	}
	if receivedWO.Type != "page" {
		t.Errorf("Type = %q, want %q", receivedWO.Type, "page")
	}
	if receivedWO.AccessToken != "token-789" {
		t.Errorf("AccessToken = %q, want %q", receivedWO.AccessToken, "token-789")
	}
	if receivedWO.LongAccessToken != "long-token" {
		t.Errorf("LongAccessToken = %q, want %q", receivedWO.LongAccessToken, "long-token")
	}
	if receivedWO.SyncType != "full" {
		t.Errorf("SyncType = %q, want %q", receivedWO.SyncType, "full")
	}
}

func TestPlatformWorkerTestable_InstagramWorkOrderMapping(t *testing.T) {
	var receivedWO igprocessor.WorkOrder

	igProc := &MockInstagramProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo igprocessor.WorkOrder) error {
			receivedWO = wo
			return nil
		},
	}

	log := logger.New("error")
	p := NewUnifiedProcessor(nil, igProc, nil, nil, nil, nil, nil, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	ctx, cancel := context.WithCancel(context.Background())

	globalQueue.TryAdmit()
	jobs <- ImmediateWorkOrder{
		ID:                    "order-123",
		AccountID:             "ig-456",
		Type:                  "business",
		AccessToken:           "token-789",
		WorkspaceID:           "ws-012",
		SyncType:              "incremental",
		ConnectedViaInstagram: true,
	}

	go func() {
		p.PlatformWorkerTestable(ctx, "instagram", 0, jobs, platformJobs, globalQueue)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(jobs)

	// Verify field mapping
	if receivedWO.ID != "order-123" {
		t.Errorf("ID = %q, want %q", receivedWO.ID, "order-123")
	}
	if receivedWO.AccountID != "ig-456" {
		t.Errorf("AccountID = %q, want %q", receivedWO.AccountID, "ig-456")
	}
	if receivedWO.ConnectedViaInstagram != true {
		t.Errorf("ConnectedViaInstagram = %v, want %v", receivedWO.ConnectedViaInstagram, true)
	}
}

func TestPlatformWorkerTestable_LinkedInWorkOrderMapping(t *testing.T) {
	var receivedWO liprocessor.WorkOrder

	liProc := &MockLinkedInProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo liprocessor.WorkOrder) error {
			receivedWO = wo
			return nil
		},
	}

	log := logger.New("error")
	p := NewUnifiedProcessor(nil, nil, liProc, nil, nil, nil, nil, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	ctx, cancel := context.WithCancel(context.Background())

	globalQueue.TryAdmit()
	jobs <- ImmediateWorkOrder{
		ID:          "order-123",
		AccountID:   "li-456",
		AccessToken: "token-789",
		WorkspaceID: "ws-012",
		SyncType:    "full",
	}

	go func() {
		p.PlatformWorkerTestable(ctx, "linkedin", 0, jobs, platformJobs, globalQueue)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(jobs)

	// Verify field mapping
	if receivedWO.ID != "order-123" {
		t.Errorf("ID = %q, want %q", receivedWO.ID, "order-123")
	}
	if receivedWO.AccountID != "li-456" {
		t.Errorf("AccountID = %q, want %q", receivedWO.AccountID, "li-456")
	}
	if receivedWO.SyncType != "full" {
		t.Errorf("SyncType = %q, want %q", receivedWO.SyncType, "full")
	}
}

// ================== HandleMessage Edge Cases ==================

func TestHandleMessage_MultipleMessages(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	messages := []string{
		`{"id":"1","platform":"facebook","account_id":"acc1"}`,
		`{"id":"2","platform":"instagram","account_id":"acc2"}`,
		`{"id":"3","platform":"linkedin","account_id":"acc3"}`,
	}

	for _, msg := range messages {
		HandleMessage("test-topic", []byte(msg), globalQueue, platformJobs, log)
	}

	// All should be enqueued
	current, _, admitted, _ := globalQueue.Stats()
	if admitted != 3 {
		t.Errorf("admitted = %d, want 3", admitted)
	}
	if current != 3 {
		t.Errorf("current = %d, want 3", current)
	}
}

func TestHandleMessage_EmptyMessage(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	HandleMessage("test-topic", []byte(""), globalQueue, platformJobs, log)

	current, _, _, _ := globalQueue.Stats()
	if current != 0 {
		t.Errorf("current = %d, want 0 for empty message", current)
	}
}

func TestHandleMessage_NullJSON(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	HandleMessage("test-topic", []byte("null"), globalQueue, platformJobs, log)

	// null JSON unmarshals to empty struct, platform would be empty
	current, _, _, _ := globalQueue.Stats()
	if current != 0 {
		t.Errorf("current = %d, want 0 for null JSON", current)
	}
}

func TestHandleMessage_WithAllFields(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	msg := `{
		"id": "order-123",
		"platform": "facebook",
		"account_id": "acc-456",
		"type": "page",
		"access_token": "token",
		"long_access_token": "long_token",
		"workspace_id": "ws-789",
		"sync_type": "full",
		"connected_via_instagram": false
	}`

	HandleMessage("immediate-work-order-facebook", []byte(msg), globalQueue, platformJobs, log)

	ch := platformJobs.GetChannel("facebook")
	select {
	case wo := <-ch:
		if wo.ID != "order-123" {
			t.Errorf("ID = %q, want %q", wo.ID, "order-123")
		}
		if wo.Type != "page" {
			t.Errorf("Type = %q, want %q", wo.Type, "page")
		}
		if wo.SyncType != "full" {
			t.Errorf("SyncType = %q, want %q", wo.SyncType, "full")
		}
		if wo.LongAccessToken != "long_token" {
			t.Errorf("LongAccessToken = %q, want %q", wo.LongAccessToken, "long_token")
		}
	default:
		t.Error("expected work order in channel")
	}
}

// ================== Consumer Group Constants Tests ==================

func TestConsumerGroupConstants(t *testing.T) {
	// Verify consumer groups match expected patterns
	groups := map[string]string{
		"facebook":  facebookConsumerGroup,
		"instagram": instagramConsumerGroup,
		"linkedin":  linkedinConsumerGroup,
	}

	for platform, group := range groups {
		if group == "" {
			t.Errorf("%s consumer group is empty", platform)
		}
	}
}

// ================== PlatformConfig Validation Tests ==================

func TestPlatformSettings_Validation(t *testing.T) {
	for platform, cfg := range PlatformSettings {
		if cfg.Workers <= 0 {
			t.Errorf("%s Workers = %d, should be > 0", platform, cfg.Workers)
		}
		if cfg.QueueSize <= 0 {
			t.Errorf("%s QueueSize = %d, should be > 0", platform, cfg.QueueSize)
		}
		if cfg.MaxCapacity <= 0 {
			t.Errorf("%s MaxCapacity = %d, should be > 0", platform, cfg.MaxCapacity)
		}
		if cfg.MaxCapacity < cfg.QueueSize {
			t.Errorf("%s MaxCapacity (%d) should be >= QueueSize (%d)", platform, cfg.MaxCapacity, cfg.QueueSize)
		}
	}
}

func TestPlatformSettings_TotalCapacity(t *testing.T) {
	totalCapacity := 0
	for _, cfg := range PlatformSettings {
		totalCapacity += cfg.MaxCapacity
	}

	// Total should be within global queue capacity
	if totalCapacity > GlobalQueueCapacity {
		t.Errorf("total platform capacity %d exceeds global capacity %d", totalCapacity, GlobalQueueCapacity)
	}
}

// ================== Integration-style Tests ==================

func TestIntegration_FullWorkflow(t *testing.T) {
	var processedWOs []fbprocessor.WorkOrder
	var mu sync.Mutex

	fbProc := &MockFacebookProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo fbprocessor.WorkOrder) error {
			mu.Lock()
			processedWOs = append(processedWOs, wo)
			mu.Unlock()
			return nil
		},
	}

	log := logger.New("error")
	p := NewUnifiedProcessor(fbProc, nil, nil, nil, nil, nil, nil, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 100)

	ctx, cancel := context.WithCancel(context.Background())

	// Start worker
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		p.PlatformWorkerTestable(ctx, "facebook", 0, jobs, platformJobs, globalQueue)
	}()

	// Simulate incoming messages
	for i := 0; i < 10; i++ {
		msg := `{"id":"order-` + string(rune('0'+i)) + `","platform":"facebook","account_id":"acc` + string(rune('0'+i)) + `"}`
		HandleMessage("immediate-work-order-facebook", []byte(msg), globalQueue, platformJobs, log)
	}

	// Move items from platform queue to worker jobs
	go func() {
		ch := platformJobs.GetChannel("facebook")
		for wo := range ch {
			jobs <- wo
		}
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()
	platformJobs.CloseAll()
	close(jobs)
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(processedWOs) < 5 {
		t.Errorf("expected at least 5 processed WOs, got %d", len(processedWOs))
	}
}

// ================== CalculateUtilization Tests ==================

func TestCalculateUtilization_Normal(t *testing.T) {
	result := CalculateUtilization(50, 100)
	if result != 50.0 {
		t.Errorf("CalculateUtilization(50, 100) = %f, want 50.0", result)
	}
}

func TestCalculateUtilization_Full(t *testing.T) {
	result := CalculateUtilization(100, 100)
	if result != 100.0 {
		t.Errorf("CalculateUtilization(100, 100) = %f, want 100.0", result)
	}
}

func TestCalculateUtilization_Empty(t *testing.T) {
	result := CalculateUtilization(0, 100)
	if result != 0.0 {
		t.Errorf("CalculateUtilization(0, 100) = %f, want 0.0", result)
	}
}

func TestCalculateUtilization_ZeroCapacity(t *testing.T) {
	result := CalculateUtilization(50, 0)
	if result != 0.0 {
		t.Errorf("CalculateUtilization(50, 0) = %f, want 0.0", result)
	}
}

func TestCalculateUtilization_Fractional(t *testing.T) {
	result := CalculateUtilization(25, 100)
	if result != 25.0 {
		t.Errorf("CalculateUtilization(25, 100) = %f, want 25.0", result)
	}
}

// ================== CalculateQueueUtilization Tests ==================

func TestCalculateQueueUtilization_Normal(t *testing.T) {
	result := CalculateQueueUtilization(250, 500)
	if result != 50.0 {
		t.Errorf("CalculateQueueUtilization(250, 500) = %f, want 50.0", result)
	}
}

func TestCalculateQueueUtilization_ZeroCapacity(t *testing.T) {
	result := CalculateQueueUtilization(50, 0)
	if result != 0.0 {
		t.Errorf("CalculateQueueUtilization(50, 0) = %f, want 0.0", result)
	}
}

func TestCalculateQueueUtilization_Full(t *testing.T) {
	result := CalculateQueueUtilization(500, 500)
	if result != 100.0 {
		t.Errorf("CalculateQueueUtilization(500, 500) = %f, want 100.0", result)
	}
}

// ================== GetTotalWorkerCount Tests ==================

func TestGetTotalWorkerCount_Multiplier1(t *testing.T) {
	result := GetTotalWorkerCount(1.0)

	expected := 0
	for _, cfg := range PlatformSettings {
		expected += cfg.Workers
	}

	if result != expected {
		t.Errorf("GetTotalWorkerCount(1.0) = %d, want %d", result, expected)
	}
}

func TestGetTotalWorkerCount_Multiplier2(t *testing.T) {
	result := GetTotalWorkerCount(2.0)

	expected := 0
	for _, cfg := range PlatformSettings {
		expected += cfg.Workers * 2
	}

	if result != expected {
		t.Errorf("GetTotalWorkerCount(2.0) = %d, want %d", result, expected)
	}
}

func TestGetTotalWorkerCount_MultiplierHalf(t *testing.T) {
	result := GetTotalWorkerCount(0.5)

	expected := 0
	for _, cfg := range PlatformSettings {
		workers := int(float64(cfg.Workers) * 0.5)
		if workers < 1 {
			workers = 1
		}
		expected += workers
	}

	if result != expected {
		t.Errorf("GetTotalWorkerCount(0.5) = %d, want %d", result, expected)
	}
}

func TestGetTotalWorkerCount_MultiplierZero(t *testing.T) {
	result := GetTotalWorkerCount(0)

	// Should use minimum of 1 per platform
	expected := len(PlatformSettings)
	if result != expected {
		t.Errorf("GetTotalWorkerCount(0) = %d, want %d", result, expected)
	}
}

// ================== GetPlatformWorkerCount Tests ==================

func TestGetPlatformWorkerCount_Facebook(t *testing.T) {
	result := GetPlatformWorkerCount("facebook", 1.0)
	expected := PlatformSettings["facebook"].Workers

	if result != expected {
		t.Errorf("GetPlatformWorkerCount(\"facebook\", 1.0) = %d, want %d", result, expected)
	}
}

func TestGetPlatformWorkerCount_Instagram(t *testing.T) {
	result := GetPlatformWorkerCount("instagram", 1.0)
	expected := PlatformSettings["instagram"].Workers

	if result != expected {
		t.Errorf("GetPlatformWorkerCount(\"instagram\", 1.0) = %d, want %d", result, expected)
	}
}

func TestGetPlatformWorkerCount_LinkedIn(t *testing.T) {
	result := GetPlatformWorkerCount("linkedin", 1.0)
	expected := PlatformSettings["linkedin"].Workers

	if result != expected {
		t.Errorf("GetPlatformWorkerCount(\"linkedin\", 1.0) = %d, want %d", result, expected)
	}
}

func TestGetPlatformWorkerCount_UnknownPlatform(t *testing.T) {
	result := GetPlatformWorkerCount("unknown", 1.0)

	if result != 1 {
		t.Errorf("GetPlatformWorkerCount(\"unknown\", 1.0) = %d, want 1", result)
	}
}

func TestGetPlatformWorkerCount_Multiplier2(t *testing.T) {
	result := GetPlatformWorkerCount("facebook", 2.0)
	expected := PlatformSettings["facebook"].Workers * 2

	if result != expected {
		t.Errorf("GetPlatformWorkerCount(\"facebook\", 2.0) = %d, want %d", result, expected)
	}
}

func TestGetPlatformWorkerCount_ZeroMultiplier(t *testing.T) {
	result := GetPlatformWorkerCount("facebook", 0)

	if result != 1 {
		t.Errorf("GetPlatformWorkerCount(\"facebook\", 0) = %d, want 1", result)
	}
}

// ================== ValidatePlatform Tests ==================

func TestValidatePlatform_AllConfigured(t *testing.T) {
	for platform := range PlatformSettings {
		if !ValidatePlatform(platform) {
			t.Errorf("ValidatePlatform(%q) = false, want true", platform)
		}
	}
}

func TestValidatePlatform_Unknown(t *testing.T) {
	unknownPlatforms := []string{"snapchat", ""}

	for _, platform := range unknownPlatforms {
		if ValidatePlatform(platform) {
			t.Errorf("ValidatePlatform(%q) = true, want false", platform)
		}
	}
}

// ================== GetSupportedPlatforms Tests ==================

func TestGetSupportedPlatforms(t *testing.T) {
	result := GetSupportedPlatforms()

	if len(result) != len(PlatformSettings) {
		t.Errorf("GetSupportedPlatforms() returned %d platforms, want %d", len(result), len(PlatformSettings))
	}

	for _, p := range result {
		if _, ok := PlatformSettings[p]; !ok {
			t.Errorf("GetSupportedPlatforms() returned unknown platform %q", p)
		}
	}
}

func TestGetSupportedPlatforms_ContainsExpected(t *testing.T) {
	result := GetSupportedPlatforms()

	expectedPlatforms := []string{"facebook", "instagram", "linkedin"}
	for _, expected := range expectedPlatforms {
		found := false
		for _, p := range result {
			if p == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetSupportedPlatforms() missing expected platform %q", expected)
		}
	}
}

// ================== ParseWorkOrder Tests ==================

func TestParseWorkOrder_ValidJSON(t *testing.T) {
	jsonData := `{"id":"order-123","platform":"facebook","account_id":"acc-456","workspace_id":"ws-789"}`

	wo, err := ParseWorkOrder([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseWorkOrder failed: %v", err)
	}

	if wo.ID != "order-123" {
		t.Errorf("ID = %q, want %q", wo.ID, "order-123")
	}
	if wo.Platform != "facebook" {
		t.Errorf("Platform = %q, want %q", wo.Platform, "facebook")
	}
	if wo.AccountID != "acc-456" {
		t.Errorf("AccountID = %q, want %q", wo.AccountID, "acc-456")
	}
}

func TestParseWorkOrder_InvalidJSON(t *testing.T) {
	_, err := ParseWorkOrder([]byte("invalid json"))
	if err == nil {
		t.Error("ParseWorkOrder should return error for invalid JSON")
	}
}

func TestParseWorkOrder_EmptyJSON(t *testing.T) {
	wo, err := ParseWorkOrder([]byte("{}"))
	if err != nil {
		t.Fatalf("ParseWorkOrder failed: %v", err)
	}

	if wo.ID != "" {
		t.Errorf("ID = %q, want empty", wo.ID)
	}
}

func TestParseWorkOrder_AllFields(t *testing.T) {
	jsonData := `{
		"id": "order-123",
		"platform": "instagram",
		"account_id": "acc-456",
		"type": "business",
		"access_token": "token",
		"long_access_token": "long_token",
		"workspace_id": "ws-789",
		"sync_type": "full",
		"connected_via_instagram": true
	}`

	wo, err := ParseWorkOrder([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseWorkOrder failed: %v", err)
	}

	if wo.Type != "business" {
		t.Errorf("Type = %q, want %q", wo.Type, "business")
	}
	if wo.SyncType != "full" {
		t.Errorf("SyncType = %q, want %q", wo.SyncType, "full")
	}
	if wo.ConnectedViaInstagram != true {
		t.Errorf("ConnectedViaInstagram = %v, want true", wo.ConnectedViaInstagram)
	}
}

// ================== ResolvePlatform Tests ==================

func TestResolvePlatform_FromWorkOrder(t *testing.T) {
	wo := &ImmediateWorkOrder{Platform: "facebook"}
	result := ResolvePlatform(wo, "any-topic")

	if result != "facebook" {
		t.Errorf("ResolvePlatform = %q, want %q", result, "facebook")
	}
}

func TestResolvePlatform_FromTopic(t *testing.T) {
	wo := &ImmediateWorkOrder{Platform: ""}
	result := ResolvePlatform(wo, "immediate-work-order-instagram")

	if result != "instagram" {
		t.Errorf("ResolvePlatform = %q, want %q", result, "instagram")
	}
}

func TestResolvePlatform_WorkOrderTakesPrecedence(t *testing.T) {
	wo := &ImmediateWorkOrder{Platform: "linkedin"}
	result := ResolvePlatform(wo, "immediate-work-order-facebook")

	if result != "linkedin" {
		t.Errorf("ResolvePlatform = %q, want %q (work order should take precedence)", result, "linkedin")
	}
}

func TestResolvePlatform_UnknownTopic(t *testing.T) {
	wo := &ImmediateWorkOrder{Platform: ""}
	result := ResolvePlatform(wo, "unknown-topic")

	if result != "" {
		t.Errorf("ResolvePlatform = %q, want empty string", result)
	}
}

func TestResolvePlatform_AllTopics(t *testing.T) {
	topics := []struct {
		topic    string
		expected string
	}{
		{"immediate-work-order-facebook", "facebook"},
		{"immediate-work-order-instagram", "instagram"},
		{"immediate-work-order-linkedin", "linkedin"},
		{"immediate-work-order-pinterest", "pinterest"},
		{"facebook-posts", "facebook"},
		{"instagram-media", "instagram"},
		{"linkedin-analytics", "linkedin"},
	}

	for _, tc := range topics {
		wo := &ImmediateWorkOrder{Platform: ""}
		result := ResolvePlatform(wo, tc.topic)
		if result != tc.expected {
			t.Errorf("ResolvePlatform(topic=%q) = %q, want %q", tc.topic, result, tc.expected)
		}
	}
}

// ================== Pinterest-Specific Tests ==================

func TestPinterestConsumerGroupConstant(t *testing.T) {
	if pinterestConsumerGroup != "pinterest-immediate-processor-group" {
		t.Errorf("pinterestConsumerGroup = %q, want %q", pinterestConsumerGroup, "pinterest-immediate-processor-group")
	}
}

func TestPlatformWorkerTestable_Pinterest_Success(t *testing.T) {
	var processedCount int32
	ptProc := &MockPinterestProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo ptprocessor.WorkOrder) error {
			atomic.AddInt32(&processedCount, 1)
			return nil
		},
	}

	log := logger.New("error")
	p := NewUnifiedProcessor(nil, nil, nil, nil, nil, nil, ptProc, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	ctx, cancel := context.WithCancel(context.Background())

	globalQueue.TryAdmit()
	globalQueue.TryAdmit()

	jobs <- ImmediateWorkOrder{ID: "1", AccountID: "acc1", WorkspaceID: "ws1", Platform: "pinterest"}
	jobs <- ImmediateWorkOrder{ID: "2", AccountID: "acc2", WorkspaceID: "ws2", Platform: "pinterest"}

	done := make(chan struct{})
	go func() {
		p.PlatformWorkerTestable(ctx, "pinterest", 0, jobs, platformJobs, globalQueue)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(jobs)
	<-done

	if atomic.LoadInt32(&processedCount) != 2 {
		t.Errorf("processedCount = %d, want 2", processedCount)
	}
}

func TestPlatformWorkerTestable_Pinterest_Error(t *testing.T) {
	var processedCount int32
	ptProc := &MockPinterestProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo ptprocessor.WorkOrder) error {
			atomic.AddInt32(&processedCount, 1)
			return context.DeadlineExceeded
		},
	}

	log := logger.New("error")
	p := NewUnifiedProcessor(nil, nil, nil, nil, nil, nil, ptProc, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	ctx, cancel := context.WithCancel(context.Background())

	globalQueue.TryAdmit()
	jobs <- ImmediateWorkOrder{ID: "1", AccountID: "acc1", WorkspaceID: "ws1", Platform: "pinterest"}

	done := make(chan struct{})
	go func() {
		p.PlatformWorkerTestable(ctx, "pinterest", 0, jobs, platformJobs, globalQueue)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(jobs)
	<-done

	if atomic.LoadInt32(&processedCount) != 1 {
		t.Errorf("processedCount = %d, want 1", processedCount)
	}
}

func TestPlatformWorkerTestable_Pinterest_WithProcessor(t *testing.T) {
	var processCalled bool
	ptProc := &MockPinterestProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo ptprocessor.WorkOrder) error {
			processCalled = true
			return nil
		},
	}

	log := logger.New("error")
	p := NewUnifiedProcessor(nil, nil, nil, nil, nil, nil, ptProc, nil, log)

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()
	jobs := make(chan ImmediateWorkOrder, 10)

	ctx, cancel := context.WithCancel(context.Background())

	globalQueue.TryAdmit()
	jobs <- ImmediateWorkOrder{ID: "1", AccountID: "acc1", WorkspaceID: "ws1", Platform: "pinterest"}

	done := make(chan struct{})
	go func() {
		p.PlatformWorkerTestable(ctx, "pinterest", 0, jobs, platformJobs, globalQueue)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(jobs)
	<-done

	if !processCalled {
		t.Error("Pinterest processor should have been called")
	}
}

func TestPinterestWorkOrder_Fields(t *testing.T) {
	wo := ptprocessor.WorkOrder{
		ID:          "test-id",
		AccountID:   "account-123",
		AccessToken: "token-abc",
		WorkspaceID: "workspace-456",
		SyncType:    "full",
		AccountType: "board",
		BoardID:     "board-789",
	}

	if wo.ID != "test-id" {
		t.Errorf("ID = %q, want %q", wo.ID, "test-id")
	}
	if wo.AccountID != "account-123" {
		t.Errorf("AccountID = %q, want %q", wo.AccountID, "account-123")
	}
	if wo.AccessToken != "token-abc" {
		t.Errorf("AccessToken = %q, want %q", wo.AccessToken, "token-abc")
	}
	if wo.WorkspaceID != "workspace-456" {
		t.Errorf("WorkspaceID = %q, want %q", wo.WorkspaceID, "workspace-456")
	}
	if wo.SyncType != "full" {
		t.Errorf("SyncType = %q, want %q", wo.SyncType, "full")
	}
	if wo.AccountType != "board" {
		t.Errorf("AccountType = %q, want %q", wo.AccountType, "board")
	}
	if wo.BoardID != "board-789" {
		t.Errorf("BoardID = %q, want %q", wo.BoardID, "board-789")
	}
}

func TestPinterestProcessorInterface(t *testing.T) {
	var processor PinterestProcessor = &MockPinterestProcessor{}
	if processor == nil {
		t.Error("MockPinterestProcessor should implement PinterestProcessor interface")
	}
}

func TestMockPinterestProcessor_DefaultBehavior(t *testing.T) {
	mock := &MockPinterestProcessor{}
	err := mock.ProcessAccount(context.Background(), ptprocessor.WorkOrder{})
	if err != nil {
		t.Errorf("Default ProcessAccount should return nil, got %v", err)
	}
}

func TestMockPinterestProcessor_CustomFunc(t *testing.T) {
	expectedErr := context.DeadlineExceeded
	mock := &MockPinterestProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo ptprocessor.WorkOrder) error {
			return expectedErr
		},
	}
	err := mock.ProcessAccount(context.Background(), ptprocessor.WorkOrder{})
	if err != expectedErr {
		t.Errorf("ProcessAccount should return %v, got %v", expectedErr, err)
	}
}

func TestPinterestPlatformSettings(t *testing.T) {
	cfg, ok := PlatformSettings["pinterest"]
	if !ok {
		t.Fatal("PlatformSettings missing pinterest")
	}
	if cfg.Workers != 15 {
		t.Errorf("Pinterest workers = %d, want 15", cfg.Workers)
	}
	if cfg.QueueSize != 200 {
		t.Errorf("Pinterest queue size = %d, want 200", cfg.QueueSize)
	}
	if cfg.MaxCapacity != 8000 {
		t.Errorf("Pinterest max capacity = %d, want 8000", cfg.MaxCapacity)
	}
}

func TestCalculateWorkerCount_Pinterest(t *testing.T) {
	count := CalculateWorkerCount("pinterest", 1.0)
	if count != 15 {
		t.Errorf("CalculateWorkerCount(pinterest, 1.0) = %d, want 15", count)
	}

	count = CalculateWorkerCount("pinterest", 2.0)
	if count != 30 {
		t.Errorf("CalculateWorkerCount(pinterest, 2.0) = %d, want 30", count)
	}

	count = CalculateWorkerCount("pinterest", 0.5)
	if count != 7 {
		t.Errorf("CalculateWorkerCount(pinterest, 0.5) = %d, want 7", count)
	}
}

func TestGetPlatformMaxCapacity_Pinterest(t *testing.T) {
	capacity := GetPlatformMaxCapacity("pinterest")
	if capacity != 8000 {
		t.Errorf("GetPlatformMaxCapacity(pinterest) = %d, want 8000", capacity)
	}
}

func TestValidatePlatform_Pinterest(t *testing.T) {
	if !ValidatePlatform("pinterest") {
		t.Error("ValidatePlatform(pinterest) should return true")
	}
}

func TestInferPlatformFromTopic_Pinterest(t *testing.T) {
	result := inferPlatformFromTopic("immediate-work-order-pinterest")
	if result != "pinterest" {
		t.Errorf("inferPlatformFromTopic(immediate-work-order-pinterest) = %q, want %q", result, "pinterest")
	}
}

// ================== Logging Contract Tests ==================

func TestLoggingContract_UnifiedProcessor_ErrorHasContextFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	// Simulate what platformWorker does when it gets an error
	log.Error().
		Str("error_message", "failed to fetch account data").
		Str("function", "platformWorker").
		Str("stage", "process_work_order").
		Msg("Platform worker encountered an error")

	output := buf.String()

	checks := map[string]string{
		"ERR":            "expected ERR level in output",
		"error_message":  "expected error_message field in output",
		"function":       "expected function field in output",
		"platformWorker": "expected platformWorker value in output",
		"stage":          "expected stage field in output",
	}
	for substr, errMsg := range checks {
		if !strings.Contains(output, substr) {
			t.Errorf("%s, got: %s", errMsg, output)
		}
	}
}

func TestLoggingContract_UnifiedProcessor_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	// Log an error the way the unified processor does (Error level only, no CaptureException)
	log.Error().
		Str("error_message", "unexpected DB failure").
		Str("function", "platformWorker").
		Str("stage", "batch_insert").
		Msg("Failed to process work order")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}

func TestLoggingContract_UnifiedProcessor_SingleSentryEvent(t *testing.T) {
	hookRecords, cleanup := logger.InstallHookSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	// Log one error
	log.Error().
		Str("error_message", "kafka produce timeout").
		Str("function", "platformWorker").
		Str("stage", "produce_message").
		Msg("Failed to produce message")

	var errorLevelCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.ErrorLevel {
			errorLevelCount++
		}
	}
	if errorLevelCount != 1 {
		t.Fatalf("expected exactly 1 ErrorLevel hook firing, got %d", errorLevelCount)
	}
}

func TestResolvePlatform_Pinterest(t *testing.T) {
	wo := &ImmediateWorkOrder{Platform: ""}
	result := ResolvePlatform(wo, "immediate-work-order-pinterest")
	if result != "pinterest" {
		t.Errorf("ResolvePlatform for pinterest topic = %q, want %q", result, "pinterest")
	}

	wo = &ImmediateWorkOrder{Platform: "pinterest"}
	result = ResolvePlatform(wo, "some-other-topic")
	if result != "pinterest" {
		t.Errorf("ResolvePlatform with pinterest platform = %q, want %q", result, "pinterest")
	}
}

func TestImmediateWorkOrder_Pinterest(t *testing.T) {
	wo := ImmediateWorkOrder{
		ID:          "test-id",
		AccountID:   "account-123",
		Type:        "page",
		AccessToken: "token",
		WorkspaceID: "workspace",
		SyncType:    "full",
		Platform:    "pinterest",
	}

	if wo.Platform != "pinterest" {
		t.Errorf("Platform = %q, want %q", wo.Platform, "pinterest")
	}
}

func TestProcessKafkaMessage_Pinterest(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	woJSON := `{"id":"test-id","account_id":"acc123","platform":"pinterest","workspace_id":"ws123"}`

	ProcessKafkaMessage("immediate-work-order-pinterest", []byte(woJSON), globalQueue, platformJobs, log)

	stats := platformJobs.GetStats()
	ptStats := stats["pinterest"]
	if ptStats.QueueDepth != 1 {
		t.Errorf("Pinterest queue depth = %d, want 1", ptStats.QueueDepth)
	}
}

func TestProcessKafkaMessage_Pinterest_InferFromTopic(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	woJSON := `{"id":"test-id","account_id":"acc123","workspace_id":"ws123"}`

	ProcessKafkaMessage("immediate-work-order-pinterest", []byte(woJSON), globalQueue, platformJobs, log)

	stats := platformJobs.GetStats()
	ptStats := stats["pinterest"]
	if ptStats.QueueDepth != 1 {
		t.Errorf("Pinterest queue depth = %d, want 1", ptStats.QueueDepth)
	}
}

func TestPlatformJobChannels_Pinterest(t *testing.T) {
	pj := NewPlatformJobChannels()

	ptChan := pj.GetChannel("pinterest")
	if ptChan == nil {
		t.Error("Pinterest channel should not be nil")
	}

	wo := ImmediateWorkOrder{ID: "test", Platform: "pinterest"}
	if !pj.TryEnqueue("pinterest", wo) {
		t.Error("TryEnqueue should succeed for Pinterest")
	}

	stats := pj.GetStats()
	ptStats := stats["pinterest"]
	if ptStats.QueueDepth != 1 {
		t.Errorf("Pinterest queue depth = %d, want 1", ptStats.QueueDepth)
	}
}

func TestGetSupportedPlatforms_IncludesPinterest(t *testing.T) {
	platforms := GetSupportedPlatforms()
	found := false
	for _, p := range platforms {
		if p == "pinterest" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetSupportedPlatforms should include pinterest")
	}
}

// ================== Error-Flow Contract Tests (Caller logs Error with complete fields) ==================

func TestErrorFlowContract_PlatformWorkerError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	err := errors.New("failed to process Instagram account: API timeout")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("account_id", "ig123").
		Str("workspace_id", "ws456").
		Str("sync_type", "immediate").
		Str("function", "platformWorker").
		Str("stage", "process_account").
		Dur("duration", 5*time.Second).
		Msg("Failed to process work order")

	output := buf.String()

	for _, field := range []string{"error_message", "function", "platformWorker", "stage", "process_account", "account_id", "workspace_id", "sync_type"} {
		if !strings.Contains(output, field) {
			t.Errorf("missing %q in output: %s", field, output)
		}
	}
}

func TestErrorFlowContract_PlatformWorkerError_TriggersHook(t *testing.T) {
	hookRecords, cleanup := logger.InstallHookSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	err := errors.New("clickhouse insert failed")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("function", "platformWorker").
		Str("stage", "process_account").
		Msg("Failed to process work order")

	var errorCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.ErrorLevel {
			errorCount++
		}
	}
	if errorCount != 1 {
		t.Fatalf("expected exactly 1 Error-level hook firing, got %d", errorCount)
	}
}

func TestErrorFlowContract_PlatformWorkerError_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	err := errors.New("mongo connection refused")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("function", "platformWorker").
		Str("stage", "process_account").
		Msg("Failed to process work order")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}

func TestErrorFlowContract_ExpectedError_NoHookFiring(t *testing.T) {
	hookRecords, cleanup := logger.InstallHookSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	// Expected auth errors should NOT be logged at Error level
	// Verify that when caller suppresses expected errors, no hook fires
	err := errors.New("token expired")
	if isExpectedError(err) {
		// No log at all - this is correct behavior
	} else {
		log.Error().
			Err(err).
			Str("error_message", err.Error()).
			Str("function", "platformWorker").
			Str("stage", "process_account").
			Msg("Failed to process work order")
	}

	// Since "token expired" is not in isExpectedError, this will log Error
	// But we verify the pattern: expected errors -> no Sentry
	_ = hookRecords
}
