package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	fbprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/facebook/facebook-immediate-processor/processor"
	igprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/instagram/instagram-immediate-processor/processor"
	liprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/linkedin/linkedin-immediate-processor/processor"
)

// ================== ServiceDependencies Tests ==================

func TestServiceDependencies_Struct(t *testing.T) {
	log := logger.New("error")
	deps := &ServiceDependencies{
		Logger:           log,
		WorkerMultiplier: 1.0,
	}

	if deps.Logger == nil {
		t.Error("Logger should not be nil")
	}
	if deps.WorkerMultiplier != 1.0 {
		t.Errorf("WorkerMultiplier = %f, want 1.0", deps.WorkerMultiplier)
	}
}

// ================== RunService Tests ==================

func TestRunService_WithMockConsumers(t *testing.T) {
	log := logger.New("error")

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

	tkProc := &MockTikTokProcessor{}

	processor := NewUnifiedProcessor(fbProc, igProc, liProc, nil, tkProc, nil, nil, nil, log)

	// Create mock consumers that deliver messages
	fbMessages := []kafka.MockMessage{
		{Topic: "immediate-work-order-facebook", Value: []byte(`{"id":"1","platform":"facebook","account_id":"fb1"}`)},
		{Topic: "immediate-work-order-facebook", Value: []byte(`{"id":"2","platform":"facebook","account_id":"fb2"}`)},
	}
	igMessages := []kafka.MockMessage{
		{Topic: "immediate-work-order-instagram", Value: []byte(`{"id":"3","platform":"instagram","account_id":"ig1"}`)},
	}
	liMessages := []kafka.MockMessage{
		{Topic: "immediate-work-order-linkedin", Value: []byte(`{"id":"4","platform":"linkedin","account_id":"li1"}`)},
	}
	tkMessages := []kafka.MockMessage{
		{Topic: "immediate-work-order-tiktok", Value: []byte(`{"id":"5","platform":"tiktok","account_id":"tk1"}`)},
	}

	deps := &ServiceDependencies{
		FacebookConsumer:  &kafka.MockConsumerWithMessages{Messages: fbMessages},
		InstagramConsumer: &kafka.MockConsumerWithMessages{Messages: igMessages},
		LinkedInConsumer:  &kafka.MockConsumerWithMessages{Messages: liMessages},
		TikTokConsumer:    &kafka.MockConsumerWithMessages{Messages: tkMessages},
		Processor:         processor,
		Logger:            log,
		WorkerMultiplier:  0.1, // Use minimal workers for testing
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := RunService(ctx, deps)
	if err != nil {
		t.Fatalf("RunService failed: %v", err)
	}

	// Give some time for processing
	time.Sleep(100 * time.Millisecond)

	// Verify messages were processed
	if atomic.LoadInt32(&fbProcessed) < 1 {
		t.Errorf("fbProcessed = %d, expected at least 1", fbProcessed)
	}
}

func TestRunService_NilConsumers(t *testing.T) {
	log := logger.New("error")
	processor := NewUnifiedProcessor(nil, nil, nil, nil, nil, nil, nil, nil, log)

	deps := &ServiceDependencies{
		FacebookConsumer:  nil,
		InstagramConsumer: nil,
		LinkedInConsumer:  nil,
		TikTokConsumer:    nil,
		Processor:         processor,
		Logger:            log,
		WorkerMultiplier:  0.1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := RunService(ctx, deps)
	if err != nil {
		t.Fatalf("RunService failed with nil consumers: %v", err)
	}
}

// ================== StartWorkerPools Tests ==================

func TestStartWorkerPools(t *testing.T) {
	log := logger.New("error")

	var processedCount int32
	fbProc := &MockFacebookProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo fbprocessor.WorkOrder) error {
			atomic.AddInt32(&processedCount, 1)
			return nil
		},
	}

	processor := NewUnifiedProcessor(fbProc, nil, nil, nil, nil, nil, nil, nil, log)
	platformJobs := NewPlatformJobChannels()
	globalQueue := NewGlobalQueue(100)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	StartWorkerPools(ctx, processor, platformJobs, globalQueue, 0.1, &wg, log)

	// Send work to facebook queue
	globalQueue.TryAdmit()
	platformJobs.TryEnqueue("facebook", ImmediateWorkOrder{ID: "1", Platform: "facebook"})

	time.Sleep(100 * time.Millisecond)
	cancel()
	platformJobs.CloseAll()
	wg.Wait()

	if atomic.LoadInt32(&processedCount) != 1 {
		t.Errorf("processedCount = %d, want 1", processedCount)
	}
}

func TestStartWorkerPools_MultipleWorkers(t *testing.T) {
	log := logger.New("error")

	var processedCount int32
	fbProc := &MockFacebookProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo fbprocessor.WorkOrder) error {
			atomic.AddInt32(&processedCount, 1)
			time.Sleep(10 * time.Millisecond) // Simulate work
			return nil
		},
	}

	processor := NewUnifiedProcessor(fbProc, nil, nil, nil, nil, nil, nil, nil, log)
	platformJobs := NewPlatformJobChannels()
	globalQueue := NewGlobalQueue(100)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Use multiplier 1.0 to get all workers
	StartWorkerPools(ctx, processor, platformJobs, globalQueue, 1.0, &wg, log)

	// Send multiple work orders
	for i := 0; i < 10; i++ {
		globalQueue.TryAdmit()
		platformJobs.TryEnqueue("facebook", ImmediateWorkOrder{ID: string(rune('0' + i)), Platform: "facebook"})
	}

	time.Sleep(200 * time.Millisecond)
	cancel()
	platformJobs.CloseAll()
	wg.Wait()

	if atomic.LoadInt32(&processedCount) < 5 {
		t.Errorf("processedCount = %d, expected at least 5", processedCount)
	}
}

// ================== RunStatsLogger Tests ==================

func TestRunStatsLogger_ContextCancel(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		RunStatsLogger(ctx, globalQueue, platformJobs, log)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("RunStatsLogger did not exit after context cancel")
	}
}

func TestRunStatsLoggerWithInterval_TickerFires(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	// Add some data to make stats interesting
	globalQueue.TryAdmit()
	platformJobs.TryEnqueue("facebook", ImmediateWorkOrder{ID: "1"})

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		// Use a very short interval to trigger the ticker
		RunStatsLoggerWithInterval(ctx, globalQueue, platformJobs, log, 10*time.Millisecond)
		close(done)
	}()

	// Wait for at least one tick
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("RunStatsLoggerWithInterval did not exit after context cancel")
	}
}

func TestRunStatsLoggerWithInterval_MultipleTicks(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		RunStatsLoggerWithInterval(ctx, globalQueue, platformJobs, log, 5*time.Millisecond)
		close(done)
	}()

	// Wait for multiple ticks
	time.Sleep(30 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("RunStatsLoggerWithInterval did not exit after context cancel")
	}
}

// ================== LogQueueStats Tests ==================

func TestLogQueueStats(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	// Add some data
	globalQueue.TryAdmit()
	globalQueue.TryAdmit()
	platformJobs.TryEnqueue("facebook", ImmediateWorkOrder{ID: "1"})

	// Should not panic
	LogQueueStats(globalQueue, platformJobs, log)
}

func TestLogQueueStats_EmptyQueues(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	// Should not panic with empty queues
	LogQueueStats(globalQueue, platformJobs, log)
}

func TestLogQueueStats_FullQueue(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(5)
	platformJobs := NewPlatformJobChannels()

	// Fill the global queue
	for i := 0; i < 5; i++ {
		globalQueue.TryAdmit()
	}

	// Should not panic
	LogQueueStats(globalQueue, platformJobs, log)
}

// ================== StartConsumers Tests ==================

func TestStartConsumers_AllPlatforms(t *testing.T) {
	log := logger.New("error")

	var fbCalled, igCalled, liCalled, tkCalled int32

	fbConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			atomic.AddInt32(&fbCalled, 1)
			<-ctx.Done()
			return ctx.Err()
		},
	}
	igConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			atomic.AddInt32(&igCalled, 1)
			<-ctx.Done()
			return ctx.Err()
		},
	}
	liConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			atomic.AddInt32(&liCalled, 1)
			<-ctx.Done()
			return ctx.Err()
		},
	}
	tkConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			atomic.AddInt32(&tkCalled, 1)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := &ServiceDependencies{
		FacebookConsumer:  fbConsumer,
		InstagramConsumer: igConsumer,
		LinkedInConsumer:  liConsumer,
		TikTokConsumer:    tkConsumer,
		Logger:            log,
	}

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	StartConsumers(ctx, deps, globalQueue, platformJobs, &wg)

	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()

	if atomic.LoadInt32(&fbCalled) != 1 {
		t.Errorf("fbCalled = %d, want 1", fbCalled)
	}
	if atomic.LoadInt32(&igCalled) != 1 {
		t.Errorf("igCalled = %d, want 1", igCalled)
	}
	if atomic.LoadInt32(&liCalled) != 1 {
		t.Errorf("liCalled = %d, want 1", liCalled)
	}
	if atomic.LoadInt32(&tkCalled) != 1 {
		t.Errorf("tkCalled = %d, want 1", tkCalled)
	}
}

func TestStartConsumers_NilConsumers(t *testing.T) {
	log := logger.New("error")

	deps := &ServiceDependencies{
		FacebookConsumer:  nil,
		InstagramConsumer: nil,
		LinkedInConsumer:  nil,
		TikTokConsumer:    nil,
		Logger:            log,
	}

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	StartConsumers(ctx, deps, globalQueue, platformJobs, &wg)

	cancel()
	wg.Wait()

	// Should complete without error
}

func TestStartConsumers_WithMessageHandler(t *testing.T) {
	log := logger.New("error")

	var messagesHandled int32

	fbConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Simulate receiving a message
			err := handler(ctx, topics[0], nil, []byte(`{"id":"1","platform":"facebook"}`))
			if err == nil {
				atomic.AddInt32(&messagesHandled, 1)
			}
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := &ServiceDependencies{
		FacebookConsumer: fbConsumer,
		Logger:           log,
	}

	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	StartConsumers(ctx, deps, globalQueue, platformJobs, &wg)

	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()

	if atomic.LoadInt32(&messagesHandled) != 1 {
		t.Errorf("messagesHandled = %d, want 1", messagesHandled)
	}

	// Verify message was queued
	current, _, _, _ := globalQueue.Stats()
	if current != 1 {
		t.Errorf("global queue current = %d, want 1", current)
	}
}

// ================== ProcessKafkaMessage Tests ==================

func TestProcessKafkaMessage_ValidMessage(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	msg := `{"id":"test-1","platform":"facebook","account_id":"acc123","workspace_id":"ws456"}`
	ProcessKafkaMessage("immediate-work-order-facebook", []byte(msg), globalQueue, platformJobs, log)

	current, _, admitted, _ := globalQueue.Stats()
	if admitted != 1 {
		t.Errorf("admitted = %d, want 1", admitted)
	}
	if current != 1 {
		t.Errorf("current = %d, want 1", current)
	}

	// Verify in queue
	ch := platformJobs.GetChannel("facebook")
	select {
	case wo := <-ch:
		if wo.ID != "test-1" {
			t.Errorf("ID = %q, want %q", wo.ID, "test-1")
		}
	default:
		t.Error("expected work order in channel")
	}
}

func TestProcessKafkaMessage_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	ProcessKafkaMessage("test-topic", []byte("invalid json"), globalQueue, platformJobs, log)

	current, _, _, _ := globalQueue.Stats()
	if current != 0 {
		t.Errorf("current = %d, want 0 for invalid JSON", current)
	}
}

func TestProcessKafkaMessage_NoPlatform(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)
	platformJobs := NewPlatformJobChannels()

	msg := `{"id":"test-1","account_id":"acc123"}`
	ProcessKafkaMessage("unknown-topic", []byte(msg), globalQueue, platformJobs, log)

	current, _, _, _ := globalQueue.Stats()
	if current != 0 {
		t.Errorf("current = %d, want 0 when platform unknown", current)
	}
}

func TestProcessKafkaMessage_GlobalQueueFull(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(1)
	platformJobs := NewPlatformJobChannels()

	// Fill global queue
	globalQueue.TryAdmit()

	msg := `{"id":"test-1","platform":"facebook","account_id":"acc123"}`
	ProcessKafkaMessage("immediate-work-order-facebook", []byte(msg), globalQueue, platformJobs, log)

	_, _, _, rejected := globalQueue.Stats()
	if rejected != 1 {
		t.Errorf("rejected = %d, want 1", rejected)
	}
}

func TestProcessKafkaMessage_PlatformQueueFull(t *testing.T) {
	log := logger.New("error")
	globalQueue := NewGlobalQueue(100)

	// Create platform jobs with very small queue
	platformJobs := &PlatformJobChannels{
		channels:  make(map[string]chan ImmediateWorkOrder),
		processed: make(map[string]*int64),
		dropped:   make(map[string]*int64),
	}
	platformJobs.channels["facebook"] = make(chan ImmediateWorkOrder, 1)
	var dropped, processed int64
	platformJobs.dropped["facebook"] = &dropped
	platformJobs.processed["facebook"] = &processed

	// Fill platform queue
	platformJobs.channels["facebook"] <- ImmediateWorkOrder{ID: "existing"}

	msg := `{"id":"test-1","platform":"facebook","account_id":"acc123"}`
	ProcessKafkaMessage("immediate-work-order-facebook", []byte(msg), globalQueue, platformJobs, log)

	if dropped != 1 {
		t.Errorf("dropped = %d, want 1", dropped)
	}

	// Global queue should be released
	current, _, _, _ := globalQueue.Stats()
	if current != 0 {
		t.Errorf("current = %d, want 0 after release", current)
	}
}

func TestProcessKafkaMessage_InferPlatformFromTopic(t *testing.T) {
	topics := []struct {
		topic    string
		platform string
	}{
		{"immediate-work-order-facebook", "facebook"},
		{"immediate-work-order-instagram", "instagram"},
		{"immediate-work-order-linkedin", "linkedin"},
	}

	for _, tc := range topics {
		t.Run(tc.platform, func(t *testing.T) {
			log := logger.New("error")
			globalQueue := NewGlobalQueue(100)
			platformJobs := NewPlatformJobChannels()

			msg := `{"id":"test-1","account_id":"acc123"}`
			ProcessKafkaMessage(tc.topic, []byte(msg), globalQueue, platformJobs, log)

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

// ================== CreateProcessorFromDeps Tests ==================

func TestCreateProcessorFromDeps(t *testing.T) {
	log := logger.New("error")
	fbProc := &MockFacebookProcessor{}
	igProc := &MockInstagramProcessor{}
	liProc := &MockLinkedInProcessor{}
	tkProc := &MockTikTokProcessor{}

	processor := CreateProcessorFromDeps(fbProc, igProc, liProc, tkProc, log)

	if processor == nil {
		t.Fatal("CreateProcessorFromDeps returned nil")
	}
	if processor.facebookProcessor == nil {
		t.Error("facebookProcessor is nil")
	}
	if processor.instagramProcessor == nil {
		t.Error("instagramProcessor is nil")
	}
	if processor.linkedinProcessor == nil {
		t.Error("linkedinProcessor is nil")
	}
	if processor.tiktokProcessor == nil {
		t.Error("tiktokProcessor is nil")
	}
}

func TestCreateProcessorFromDeps_NilProcessors(t *testing.T) {
	log := logger.New("error")

	processor := CreateProcessorFromDeps(nil, nil, nil, nil, log)

	if processor == nil {
		t.Fatal("CreateProcessorFromDeps returned nil")
	}
	// nil processors are allowed
}

// ================== Convert Work Order Tests ==================

func TestConvertToFacebookWorkOrder(t *testing.T) {
	wo := ImmediateWorkOrder{
		ID:              "order-123",
		AccountID:       "acc-456",
		Type:            "page",
		AccessToken:     "token",
		WorkspaceID:     "ws-789",
		LongAccessToken: "long_token",
		SyncType:        "full",
	}

	result := ConvertToFacebookWorkOrder(wo)

	if result.ID != wo.ID {
		t.Errorf("ID = %q, want %q", result.ID, wo.ID)
	}
	if result.AccountID != wo.AccountID {
		t.Errorf("AccountID = %q, want %q", result.AccountID, wo.AccountID)
	}
	if result.Type != wo.Type {
		t.Errorf("Type = %q, want %q", result.Type, wo.Type)
	}
	if result.AccessToken != wo.AccessToken {
		t.Errorf("AccessToken = %q, want %q", result.AccessToken, wo.AccessToken)
	}
	if result.WorkspaceID != wo.WorkspaceID {
		t.Errorf("WorkspaceID = %q, want %q", result.WorkspaceID, wo.WorkspaceID)
	}
	if result.LongAccessToken != wo.LongAccessToken {
		t.Errorf("LongAccessToken = %q, want %q", result.LongAccessToken, wo.LongAccessToken)
	}
	if result.SyncType != wo.SyncType {
		t.Errorf("SyncType = %q, want %q", result.SyncType, wo.SyncType)
	}
}

func TestConvertToInstagramWorkOrder(t *testing.T) {
	wo := ImmediateWorkOrder{
		ID:                    "order-123",
		AccountID:             "acc-456",
		Type:                  "business",
		AccessToken:           "token",
		WorkspaceID:           "ws-789",
		SyncType:              "incremental",
		ConnectedViaInstagram: true,
	}

	result := ConvertToInstagramWorkOrder(wo)

	if result.ID != wo.ID {
		t.Errorf("ID = %q, want %q", result.ID, wo.ID)
	}
	if result.AccountID != wo.AccountID {
		t.Errorf("AccountID = %q, want %q", result.AccountID, wo.AccountID)
	}
	if result.Type != wo.Type {
		t.Errorf("Type = %q, want %q", result.Type, wo.Type)
	}
	if result.ConnectedViaInstagram != wo.ConnectedViaInstagram {
		t.Errorf("ConnectedViaInstagram = %v, want %v", result.ConnectedViaInstagram, wo.ConnectedViaInstagram)
	}
}

func TestConvertToLinkedInWorkOrder(t *testing.T) {
	wo := ImmediateWorkOrder{
		ID:          "order-123",
		AccountID:   "acc-456",
		AccessToken: "token",
		WorkspaceID: "ws-789",
		SyncType:    "full",
	}

	result := ConvertToLinkedInWorkOrder(wo)

	if result.ID != wo.ID {
		t.Errorf("ID = %q, want %q", result.ID, wo.ID)
	}
	if result.AccountID != wo.AccountID {
		t.Errorf("AccountID = %q, want %q", result.AccountID, wo.AccountID)
	}
	if result.AccessToken != wo.AccessToken {
		t.Errorf("AccessToken = %q, want %q", result.AccessToken, wo.AccessToken)
	}
	if result.SyncType != wo.SyncType {
		t.Errorf("SyncType = %q, want %q", result.SyncType, wo.SyncType)
	}
}

func TestConvertToFacebookWorkOrder_EmptyFields(t *testing.T) {
	wo := ImmediateWorkOrder{}
	result := ConvertToFacebookWorkOrder(wo)

	if result.ID != "" {
		t.Errorf("ID = %q, want empty", result.ID)
	}
}

func TestConvertToInstagramWorkOrder_EmptyFields(t *testing.T) {
	wo := ImmediateWorkOrder{}
	result := ConvertToInstagramWorkOrder(wo)

	if result.ID != "" {
		t.Errorf("ID = %q, want empty", result.ID)
	}
	if result.ConnectedViaInstagram != false {
		t.Errorf("ConnectedViaInstagram = %v, want false", result.ConnectedViaInstagram)
	}
}

func TestConvertToLinkedInWorkOrder_EmptyFields(t *testing.T) {
	wo := ImmediateWorkOrder{}
	result := ConvertToLinkedInWorkOrder(wo)

	if result.ID != "" {
		t.Errorf("ID = %q, want empty", result.ID)
	}
}

// ================== Integration Tests with Mocks ==================

func TestIntegration_FullPipelineWithMocks(t *testing.T) {
	log := logger.New("error")

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

	processor := NewUnifiedProcessor(fbProc, igProc, liProc, nil, nil, nil, nil, nil, log)

	// Create mock consumers with messages
	fbConsumer := &kafka.MockConsumerWithMessages{
		Messages: []kafka.MockMessage{
			{Topic: "immediate-work-order-facebook", Value: []byte(`{"id":"fb1","platform":"facebook","account_id":"acc1"}`)},
			{Topic: "immediate-work-order-facebook", Value: []byte(`{"id":"fb2","platform":"facebook","account_id":"acc2"}`)},
		},
	}
	igConsumer := &kafka.MockConsumerWithMessages{
		Messages: []kafka.MockMessage{
			{Topic: "immediate-work-order-instagram", Value: []byte(`{"id":"ig1","platform":"instagram","account_id":"acc3"}`)},
		},
	}
	liConsumer := &kafka.MockConsumerWithMessages{
		Messages: []kafka.MockMessage{
			{Topic: "immediate-work-order-linkedin", Value: []byte(`{"id":"li1","platform":"linkedin","account_id":"acc4"}`)},
		},
	}

	deps := &ServiceDependencies{
		FacebookConsumer:  fbConsumer,
		InstagramConsumer: igConsumer,
		LinkedInConsumer:  liConsumer,
		TikTokConsumer:    nil,
		Processor:         processor,
		Logger:            log,
		WorkerMultiplier:  0.1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	RunService(ctx, deps)

	// Verify all platforms received messages
	t.Logf("FB processed: %d, IG processed: %d, LI processed: %d",
		atomic.LoadInt32(&fbProcessed),
		atomic.LoadInt32(&igProcessed),
		atomic.LoadInt32(&liProcessed))
}

func TestIntegration_ConsumerError(t *testing.T) {
	log := logger.New("error")
	processor := NewUnifiedProcessor(nil, nil, nil, nil, nil, nil, nil, nil, log)

	fbConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			return context.DeadlineExceeded
		},
	}

	deps := &ServiceDependencies{
		FacebookConsumer: fbConsumer,
		Processor:        processor,
		Logger:           log,
		WorkerMultiplier: 0.1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Should handle consumer error gracefully
	err := RunService(ctx, deps)
	if err != nil {
		t.Fatalf("RunService should not return error: %v", err)
	}
}

// ================== StartWorkerPools Edge Case Tests ==================

func TestStartWorkerPools_ZeroMultiplier(t *testing.T) {
	log := logger.New("error")

	var processedCount int32
	fbProc := &MockFacebookProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo fbprocessor.WorkOrder) error {
			atomic.AddInt32(&processedCount, 1)
			return nil
		},
	}

	processor := NewUnifiedProcessor(fbProc, nil, nil, nil, nil, nil, nil, nil, log)
	platformJobs := NewPlatformJobChannels()
	globalQueue := NewGlobalQueue(100)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Use zero multiplier - should still create at least 1 worker per platform
	StartWorkerPools(ctx, processor, platformJobs, globalQueue, 0.0, &wg, log)

	// Send work to facebook queue
	globalQueue.TryAdmit()
	platformJobs.TryEnqueue("facebook", ImmediateWorkOrder{ID: "1", Platform: "facebook"})

	time.Sleep(100 * time.Millisecond)
	cancel()
	platformJobs.CloseAll()
	wg.Wait()

	// Even with 0.0 multiplier, workerCount should be at least 1
	if atomic.LoadInt32(&processedCount) != 1 {
		t.Errorf("processedCount = %d, want 1 (at least 1 worker should exist)", processedCount)
	}
}

func TestStartWorkerPools_NegativeMultiplier(t *testing.T) {
	log := logger.New("error")

	var processedCount int32
	fbProc := &MockFacebookProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo fbprocessor.WorkOrder) error {
			atomic.AddInt32(&processedCount, 1)
			return nil
		},
	}

	processor := NewUnifiedProcessor(fbProc, nil, nil, nil, nil, nil, nil, nil, log)
	platformJobs := NewPlatformJobChannels()
	globalQueue := NewGlobalQueue(100)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Use negative multiplier - should still create at least 1 worker per platform
	StartWorkerPools(ctx, processor, platformJobs, globalQueue, -0.5, &wg, log)

	// Send work to facebook queue
	globalQueue.TryAdmit()
	platformJobs.TryEnqueue("facebook", ImmediateWorkOrder{ID: "1", Platform: "facebook"})

	time.Sleep(100 * time.Millisecond)
	cancel()
	platformJobs.CloseAll()
	wg.Wait()

	// Even with negative multiplier, workerCount should be at least 1
	if atomic.LoadInt32(&processedCount) != 1 {
		t.Errorf("processedCount = %d, want 1 (at least 1 worker should exist)", processedCount)
	}
}

func TestStartWorkerPools_VerySmallMultiplier(t *testing.T) {
	log := logger.New("error")
	processor := NewUnifiedProcessor(nil, nil, nil, nil, nil, nil, nil, nil, log)
	platformJobs := NewPlatformJobChannels()
	globalQueue := NewGlobalQueue(100)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Use very small multiplier (0.001) - should still create at least 1 worker
	// 40 workers * 0.001 = 0.04 -> rounds to 0, but min is 1
	StartWorkerPools(ctx, processor, platformJobs, globalQueue, 0.001, &wg, log)

	cancel()
	platformJobs.CloseAll()
	wg.Wait()

	// Test passes if we don't panic/hang
}

// ================== GetChannel Race Condition Test ==================

func TestPlatformJobChannels_GetChannel_ConcurrentCreation(t *testing.T) {
	pjc := NewPlatformJobChannels()

	var wg sync.WaitGroup
	channels := make([]chan ImmediateWorkOrder, 10)

	// Concurrently get the same new channel from multiple goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			channels[idx] = pjc.GetChannel("new_platform")
		}(i)
	}
	wg.Wait()

	// All should return the same channel
	for i := 1; i < 10; i++ {
		if channels[i] != channels[0] {
			t.Errorf("GetChannel returned different channel on concurrent access: %p vs %p", channels[i], channels[0])
		}
	}
}

func TestPlatformJobChannels_GetChannel_RaceCondition(t *testing.T) {
	// Run multiple times to increase chance of triggering race condition
	for run := 0; run < 10; run++ {
		pjc := NewPlatformJobChannels()
		platformName := "race_test_platform"

		var wg sync.WaitGroup
		start := make(chan struct{})
		numGoroutines := 100
		channels := make([]chan ImmediateWorkOrder, numGoroutines)

		// Setup all goroutines
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				<-start // Wait for signal
				channels[idx] = pjc.GetChannel(platformName)
			}(i)
		}

		// Start all goroutines at the same time
		close(start)
		wg.Wait()

		// All should return the same channel
		for i := 1; i < numGoroutines; i++ {
			if channels[i] != channels[0] {
				t.Errorf("run %d: GetChannel returned different channel: %p vs %p", run, channels[i], channels[0])
			}
		}
	}
}

// ================== GetDefaultWorkerMultiplier Tests ==================

func TestGetDefaultWorkerMultiplier(t *testing.T) {
	result := GetDefaultWorkerMultiplier()

	if result != 1.0 {
		t.Errorf("GetDefaultWorkerMultiplier() = %f, want 1.0", result)
	}
}

// ================== CalculateWorkerCount Tests ==================

func TestCalculateWorkerCount_Facebook(t *testing.T) {
	result := CalculateWorkerCount("facebook", 1.0)

	if result != 40 {
		t.Errorf("CalculateWorkerCount(\"facebook\", 1.0) = %d, want 40", result)
	}
}

func TestCalculateWorkerCount_WithMultiplier(t *testing.T) {
	result := CalculateWorkerCount("facebook", 0.5)

	if result != 20 {
		t.Errorf("CalculateWorkerCount(\"facebook\", 0.5) = %d, want 20", result)
	}
}

func TestCalculateWorkerCount_UnknownPlatform(t *testing.T) {
	result := CalculateWorkerCount("unknown", 1.0)

	if result != 1 {
		t.Errorf("CalculateWorkerCount(\"unknown\", 1.0) = %d, want 1", result)
	}
}

func TestCalculateWorkerCount_ZeroMultiplier(t *testing.T) {
	result := CalculateWorkerCount("facebook", 0.0)

	if result != 1 {
		t.Errorf("CalculateWorkerCount with 0.0 multiplier should return 1, got %d", result)
	}
}

func TestCalculateWorkerCount_NegativeMultiplier(t *testing.T) {
	result := CalculateWorkerCount("facebook", -1.0)

	if result != 1 {
		t.Errorf("CalculateWorkerCount with negative multiplier should return 1, got %d", result)
	}
}

// ================== GetPlatformQueueCapacity Tests ==================

func TestGetPlatformQueueCapacity_Facebook(t *testing.T) {
	result := GetPlatformQueueCapacity("facebook")

	if result != 500 {
		t.Errorf("GetPlatformQueueCapacity(\"facebook\") = %d, want 500", result)
	}
}

func TestGetPlatformQueueCapacity_Instagram(t *testing.T) {
	result := GetPlatformQueueCapacity("instagram")

	if result != 400 {
		t.Errorf("GetPlatformQueueCapacity(\"instagram\") = %d, want 400", result)
	}
}

func TestGetPlatformQueueCapacity_LinkedIn(t *testing.T) {
	result := GetPlatformQueueCapacity("linkedin")

	if result != 200 {
		t.Errorf("GetPlatformQueueCapacity(\"linkedin\") = %d, want 200", result)
	}
}

func TestGetPlatformQueueCapacity_Unknown(t *testing.T) {
	result := GetPlatformQueueCapacity("unknown")

	if result != 50 {
		t.Errorf("GetPlatformQueueCapacity(\"unknown\") = %d, want 50 (default)", result)
	}
}

// ================== GetPlatformMaxCapacity Tests ==================

func TestGetPlatformMaxCapacity_Facebook(t *testing.T) {
	result := GetPlatformMaxCapacity("facebook")

	if result != 24000 {
		t.Errorf("GetPlatformMaxCapacity(\"facebook\") = %d, want 24000", result)
	}
}

func TestGetPlatformMaxCapacity_Unknown(t *testing.T) {
	result := GetPlatformMaxCapacity("unknown")

	if result != 0 {
		t.Errorf("GetPlatformMaxCapacity(\"unknown\") = %d, want 0", result)
	}
}

// ================== GetGlobalQueueDefaultCapacity Tests ==================

func TestGetGlobalQueueDefaultCapacity(t *testing.T) {
	result := GetGlobalQueueDefaultCapacity()

	if result != 100000 {
		t.Errorf("GetGlobalQueueDefaultCapacity() = %d, want 100000", result)
	}
}

// ================== ValidateWorkOrder Tests ==================

func TestValidateWorkOrder_Valid(t *testing.T) {
	wo := ImmediateWorkOrder{Platform: "facebook", AccountID: "acc123"}

	if !ValidateWorkOrder(wo) {
		t.Error("ValidateWorkOrder should return true for valid work order")
	}
}

func TestValidateWorkOrder_MissingPlatform(t *testing.T) {
	wo := ImmediateWorkOrder{AccountID: "acc123"}

	if ValidateWorkOrder(wo) {
		t.Error("ValidateWorkOrder should return false when platform is missing")
	}
}

func TestValidateWorkOrder_MissingAccountID(t *testing.T) {
	wo := ImmediateWorkOrder{Platform: "facebook"}

	if ValidateWorkOrder(wo) {
		t.Error("ValidateWorkOrder should return false when account_id is missing")
	}
}

func TestValidateWorkOrder_Empty(t *testing.T) {
	wo := ImmediateWorkOrder{}

	if ValidateWorkOrder(wo) {
		t.Error("ValidateWorkOrder should return false for empty work order")
	}
}

// ================== GetConsumerGroupForPlatform Tests ==================

func TestGetConsumerGroupForPlatform_Facebook(t *testing.T) {
	result := GetConsumerGroupForPlatform("facebook")

	if result != "immediate-processor-group" {
		t.Errorf("GetConsumerGroupForPlatform(\"facebook\") = %q, want %q", result, "immediate-processor-group")
	}
}

func TestGetConsumerGroupForPlatform_Instagram(t *testing.T) {
	result := GetConsumerGroupForPlatform("instagram")

	if result != "instagram-immediate-processor-group" {
		t.Errorf("GetConsumerGroupForPlatform(\"instagram\") = %q", result)
	}
}

func TestGetConsumerGroupForPlatform_LinkedIn(t *testing.T) {
	result := GetConsumerGroupForPlatform("linkedin")

	if result != "linkedin-immediate-processor-group" {
		t.Errorf("GetConsumerGroupForPlatform(\"linkedin\") = %q", result)
	}
}

func TestGetConsumerGroupForPlatform_TikTok(t *testing.T) {
	result := GetConsumerGroupForPlatform("tiktok")

	if result != "tiktok-immediate-processor-group" {
		t.Errorf("GetConsumerGroupForPlatform(\"tiktok\") = %q", result)
	}
}

func TestGetConsumerGroupForPlatform_Unknown(t *testing.T) {
	result := GetConsumerGroupForPlatform("unknown")

	if result != "" {
		t.Errorf("GetConsumerGroupForPlatform(\"unknown\") = %q, want empty", result)
	}
}

// ================== GetTopicForPlatform Tests ==================

func TestGetTopicForPlatform_Facebook(t *testing.T) {
	result := GetTopicForPlatform("facebook")

	if result != "immediate-work-order-facebook" {
		t.Errorf("GetTopicForPlatform(\"facebook\") = %q", result)
	}
}

func TestGetTopicForPlatform_Instagram(t *testing.T) {
	result := GetTopicForPlatform("instagram")

	if result != "immediate-work-order-instagram" {
		t.Errorf("GetTopicForPlatform(\"instagram\") = %q", result)
	}
}

func TestGetTopicForPlatform_LinkedIn(t *testing.T) {
	result := GetTopicForPlatform("linkedin")

	if result != "immediate-work-order-linkedin" {
		t.Errorf("GetTopicForPlatform(\"linkedin\") = %q", result)
	}
}

func TestGetTopicForPlatform_TikTok(t *testing.T) {
	result := GetTopicForPlatform("tiktok")

	if result != "immediate-work-order-tiktok" {
		t.Errorf("GetTopicForPlatform(\"tiktok\") = %q", result)
	}
}

func TestGetTopicForPlatform_Unknown(t *testing.T) {
	result := GetTopicForPlatform("unknown")

	if result != "" {
		t.Errorf("GetTopicForPlatform(\"unknown\") = %q, want empty", result)
	}
}

// ================== IsValidWorkOrderJSON Tests ==================

func TestIsValidWorkOrderJSON_Valid(t *testing.T) {
	data := []byte(`{"id":"1","platform":"facebook","account_id":"acc123"}`)

	if !IsValidWorkOrderJSON(data) {
		t.Error("IsValidWorkOrderJSON should return true for valid JSON")
	}
}

func TestIsValidWorkOrderJSON_Invalid(t *testing.T) {
	data := []byte(`invalid json`)

	if IsValidWorkOrderJSON(data) {
		t.Error("IsValidWorkOrderJSON should return false for invalid JSON")
	}
}

func TestIsValidWorkOrderJSON_Empty(t *testing.T) {
	data := []byte(`{}`)

	if !IsValidWorkOrderJSON(data) {
		t.Error("IsValidWorkOrderJSON should return true for empty JSON object")
	}
}

// ================== GetAllConsumerGroups Tests ==================

func TestGetAllConsumerGroups(t *testing.T) {
	result := GetAllConsumerGroups()

	if len(result) != 8 {
		t.Errorf("GetAllConsumerGroups returned %d groups, want 8", len(result))
	}

	if result["facebook"] != "immediate-processor-group" {
		t.Errorf("facebook consumer group = %q", result["facebook"])
	}
	if result["tiktok"] != "tiktok-immediate-processor-group" {
		t.Errorf("tiktok consumer group = %q", result["tiktok"])
	}
	if result["twitter"] != "twitter-immediate-processor-group" {
		t.Errorf("twitter consumer group = %q", result["twitter"])
	}
	if result["youtube"] != "youtube-immediate-processor-group" {
		t.Errorf("youtube consumer group = %q", result["youtube"])
	}
	if result["pinterest"] != "pinterest-immediate-processor-group" {
		t.Errorf("pinterest consumer group = %q", result["pinterest"])
	}
	if result["gmb"] != "gmb-immediate-processor-group" {
		t.Errorf("gmb consumer group = %q", result["gmb"])
	}
}

// ================== CalculateQueueFillPercentage Tests ==================

func TestCalculateQueueFillPercentage_Half(t *testing.T) {
	result := CalculateQueueFillPercentage(50, 100)

	if result != 50.0 {
		t.Errorf("CalculateQueueFillPercentage(50, 100) = %f, want 50.0", result)
	}
}

func TestCalculateQueueFillPercentage_Full(t *testing.T) {
	result := CalculateQueueFillPercentage(100, 100)

	if result != 100.0 {
		t.Errorf("CalculateQueueFillPercentage(100, 100) = %f, want 100.0", result)
	}
}

func TestCalculateQueueFillPercentage_Empty(t *testing.T) {
	result := CalculateQueueFillPercentage(0, 100)

	if result != 0.0 {
		t.Errorf("CalculateQueueFillPercentage(0, 100) = %f, want 0.0", result)
	}
}

func TestCalculateQueueFillPercentage_ZeroCapacity(t *testing.T) {
	result := CalculateQueueFillPercentage(50, 0)

	if result != 0.0 {
		t.Errorf("CalculateQueueFillPercentage(50, 0) = %f, want 0.0", result)
	}
}

// ================== EstimateTotalWorkers Tests ==================

func TestEstimateTotalWorkers_DefaultMultiplier(t *testing.T) {
	result := EstimateTotalWorkers(1.0)

	expected := 40 + 30 + 15 + 10 + 20 + 25 + 15 + 10 + 15 // facebook + instagram + linkedin + tiktok + youtube + twitter + pinterest + gmb + meta_ads
	if result != expected {
		t.Errorf("EstimateTotalWorkers(1.0) = %d, want %d", result, expected)
	}
}

func TestEstimateTotalWorkers_HalfMultiplier(t *testing.T) {
	result := EstimateTotalWorkers(0.5)

	expected := 20 + 15 + 7 + 5 + 10 + 12 + 7 + 5 + 7 // half of each (rounded), including meta_ads
	if result != expected {
		t.Errorf("EstimateTotalWorkers(0.5) = %d, want %d", result, expected)
	}
}

func TestEstimateTotalWorkers_ZeroMultiplier(t *testing.T) {
	result := EstimateTotalWorkers(0.0)

	// With 0 multiplier, each platform should have at least 1 worker
	if result != 9 {
		t.Errorf("EstimateTotalWorkers(0.0) = %d, want 9 (1 per platform minimum)", result)
	}
}

// ================== GetAllPlatformTopics Tests ==================

func TestGetAllPlatformTopics(t *testing.T) {
	result := GetAllPlatformTopics()

	if len(result) != 8 {
		t.Errorf("GetAllPlatformTopics returned %d topics, want 8", len(result))
	}

	expected := []string{
		"immediate-work-order-facebook",
		"immediate-work-order-instagram",
		"immediate-work-order-linkedin",
		"immediate-work-order-youtube",
		"immediate-work-order-tiktok",
		"immediate-work-order-twitter",
		"immediate-work-order-pinterest",
		"immediate-work-order-gmb",
	}

	for i, topic := range expected {
		if result[i] != topic {
			t.Errorf("result[%d] = %q, want %q", i, result[i], topic)
		}
	}
}

func TestGetConsumerGroupForPlatform_Pinterest(t *testing.T) {
	result := GetConsumerGroupForPlatform("pinterest")

	if result != "pinterest-immediate-processor-group" {
		t.Errorf("GetConsumerGroupForPlatform(\"pinterest\") = %q", result)
	}
}

func TestGetTopicForPlatform_Pinterest(t *testing.T) {
	result := GetTopicForPlatform("pinterest")

	if result != "immediate-work-order-pinterest" {
		t.Errorf("GetTopicForPlatform(\"pinterest\") = %q", result)
	}
}

func TestGetPlatformQueueCapacity_Pinterest(t *testing.T) {
	result := GetPlatformQueueCapacity("pinterest")

	if result != 200 {
		t.Errorf("GetPlatformQueueCapacity(\"pinterest\") = %d, want 200", result)
	}
}

func TestGetConsumerGroupForPlatform_GMB(t *testing.T) {
	result := GetConsumerGroupForPlatform("gmb")

	if result != "gmb-immediate-processor-group" {
		t.Errorf("GetConsumerGroupForPlatform(\"gmb\") = %q", result)
	}
}

func TestGetTopicForPlatform_GMB(t *testing.T) {
	result := GetTopicForPlatform("gmb")

	if result != "immediate-work-order-gmb" {
		t.Errorf("GetTopicForPlatform(\"gmb\") = %q", result)
	}
}

func TestGetPlatformQueueCapacity_GMB(t *testing.T) {
	result := GetPlatformQueueCapacity("gmb")

	if result != 150 {
		t.Errorf("GetPlatformQueueCapacity(\"gmb\") = %d, want 150", result)
	}
}
