package main

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== NewParsingService Tests ==================

func TestNewParsingService(t *testing.T) {
	log := logger.New("error")
	svc := NewParsingService(log)

	if svc == nil {
		t.Fatal("NewParsingService returned nil")
	}
	if svc.parsePost == nil {
		t.Error("parsePost is nil")
	}
	if svc.parseInsights == nil {
		t.Error("parseInsights is nil")
	}
}

func TestNewParsingServiceWithFuncs(t *testing.T) {
	log := logger.New("error")

	customParsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return &kafkamodels.ParsedLinkedinPost{PostID: "custom"}, nil
	}
	customParseInsights := func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
		return []*kafkamodels.ParsedLinkedinInsights{{LinkedinID: "custom"}}, nil
	}

	svc := NewParsingServiceWithFuncs(customParsePost, customParseInsights, log)

	if svc == nil {
		t.Fatal("NewParsingServiceWithFuncs returned nil")
	}

	// Test custom parsePost
	post, err := svc.parsePost(json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("parsePost error: %v", err)
	}
	if post.PostID != "custom" {
		t.Errorf("PostID = %q, want %q", post.PostID, "custom")
	}
}

// ================== ParseAndQueuePostTestable Tests ==================

func TestParsingService_ParseAndQueuePostTestable_Success(t *testing.T) {
	log := logger.New("error")

	parsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return &kafkamodels.ParsedLinkedinPost{
			PostID:     "post123",
			LinkedinID: "li456",
		}, nil
	}

	svc := NewParsingServiceWithFuncs(parsePost, nil, log)

	postsChan := make(chan *kafkamodels.ParsedLinkedinPost, 10)
	var counter uint64

	err := svc.ParseAndQueuePostTestable(context.Background(), "li456", []byte(`{}`), postsChan, &counter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atomic.LoadUint64(&counter) != 1 {
		t.Errorf("counter = %d, want 1", counter)
	}

	select {
	case post := <-postsChan:
		if post.PostID != "post123" {
			t.Errorf("PostID = %q, want %q", post.PostID, "post123")
		}
	default:
		t.Error("expected post in channel")
	}
}

func TestParsingService_ParseAndQueuePostTestable_ParseError(t *testing.T) {
	log := logger.New("error")

	parsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return nil, errors.New("parse error")
	}

	svc := NewParsingServiceWithFuncs(parsePost, nil, log)

	postsChan := make(chan *kafkamodels.ParsedLinkedinPost, 10)
	var counter uint64

	err := svc.ParseAndQueuePostTestable(context.Background(), "li456", []byte(`{}`), postsChan, &counter)
	if err == nil {
		t.Error("expected error")
	}

	if atomic.LoadUint64(&counter) != 0 {
		t.Errorf("counter = %d, want 0", counter)
	}
}

func TestParsingService_ParseAndQueuePostTestable_NilResult(t *testing.T) {
	log := logger.New("error")

	parsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return nil, nil
	}

	svc := NewParsingServiceWithFuncs(parsePost, nil, log)

	postsChan := make(chan *kafkamodels.ParsedLinkedinPost, 10)
	var counter uint64

	err := svc.ParseAndQueuePostTestable(context.Background(), "li456", []byte(`{}`), postsChan, &counter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atomic.LoadUint64(&counter) != 0 {
		t.Errorf("counter = %d, want 0", counter)
	}
}

func TestParsingService_ParseAndQueuePostTestable_SetsLinkedinID(t *testing.T) {
	log := logger.New("error")

	parsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return &kafkamodels.ParsedLinkedinPost{
			PostID:     "post123",
			LinkedinID: "", // Empty - should be filled in
		}, nil
	}

	svc := NewParsingServiceWithFuncs(parsePost, nil, log)

	postsChan := make(chan *kafkamodels.ParsedLinkedinPost, 10)

	err := svc.ParseAndQueuePostTestable(context.Background(), "li789", []byte(`{}`), postsChan, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case post := <-postsChan:
		if post.LinkedinID != "li789" {
			t.Errorf("LinkedinID = %q, want %q", post.LinkedinID, "li789")
		}
	default:
		t.Error("expected post in channel")
	}
}

func TestParsingService_ParseAndQueuePostTestable_ContextCancel(t *testing.T) {
	log := logger.New("error")

	parsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return &kafkamodels.ParsedLinkedinPost{PostID: "post123"}, nil
	}

	svc := NewParsingServiceWithFuncs(parsePost, nil, log)

	postsChan := make(chan *kafkamodels.ParsedLinkedinPost) // Unbuffered - will block

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.ParseAndQueuePostTestable(ctx, "li456", []byte(`{}`), postsChan, nil)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// ================== ParseAndQueueInsightsTestable Tests ==================

func TestParsingService_ParseAndQueueInsightsTestable_Success(t *testing.T) {
	log := logger.New("error")

	parseInsights := func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
		return []*kafkamodels.ParsedLinkedinInsights{
			{LinkedinID: "li456", CreatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
			{LinkedinID: "li456", CreatedAt: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)},
		}, nil
	}

	svc := NewParsingServiceWithFuncs(nil, parseInsights, log)

	insightsChan := make(chan *kafkamodels.ParsedLinkedinInsights, 10)
	var counter uint64

	queued, err := svc.ParseAndQueueInsightsTestable(context.Background(), "li456", []byte(`{}`), insightsChan, &counter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if queued != 2 {
		t.Errorf("queued = %d, want 2", queued)
	}

	if atomic.LoadUint64(&counter) != 2 {
		t.Errorf("counter = %d, want 2", counter)
	}
}

func TestParsingService_ParseAndQueueInsightsTestable_ParseError(t *testing.T) {
	log := logger.New("error")

	parseInsights := func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
		return nil, errors.New("parse error")
	}

	svc := NewParsingServiceWithFuncs(nil, parseInsights, log)

	insightsChan := make(chan *kafkamodels.ParsedLinkedinInsights, 10)
	var counter uint64

	_, err := svc.ParseAndQueueInsightsTestable(context.Background(), "li456", []byte(`{}`), insightsChan, &counter)
	if err == nil {
		t.Error("expected error")
	}
}

func TestParsingService_ParseAndQueueInsightsTestable_EmptyResult(t *testing.T) {
	log := logger.New("error")

	parseInsights := func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
		return nil, nil
	}

	svc := NewParsingServiceWithFuncs(nil, parseInsights, log)

	insightsChan := make(chan *kafkamodels.ParsedLinkedinInsights, 10)
	var counter uint64

	queued, err := svc.ParseAndQueueInsightsTestable(context.Background(), "li456", []byte(`{}`), insightsChan, &counter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if queued != 0 {
		t.Errorf("queued = %d, want 0", queued)
	}
}

func TestParsingService_ParseAndQueueInsightsTestable_SetsRecordID(t *testing.T) {
	log := logger.New("error")

	parseInsights := func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
		return []*kafkamodels.ParsedLinkedinInsights{
			{LinkedinID: "li456", CreatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		}, nil
	}

	svc := NewParsingServiceWithFuncs(nil, parseInsights, log)

	insightsChan := make(chan *kafkamodels.ParsedLinkedinInsights, 10)

	_, _ = svc.ParseAndQueueInsightsTestable(context.Background(), "li456", []byte(`{}`), insightsChan, nil)

	select {
	case insights := <-insightsChan:
		expectedRecordID := "li456_2024-01-15"
		if insights.RecordID != expectedRecordID {
			t.Errorf("RecordID = %q, want %q", insights.RecordID, expectedRecordID)
		}
	default:
		t.Error("expected insights in channel")
	}
}

// ================== PostsParserWorkerTestable Tests ==================

func TestParsingService_PostsParserWorkerTestable_ProcessesMessages(t *testing.T) {
	log := logger.New("error")

	var parseCalls int32
	parsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		atomic.AddInt32(&parseCalls, 1)
		return &kafkamodels.ParsedLinkedinPost{PostID: "post123"}, nil
	}

	svc := NewParsingServiceWithFuncs(parsePost, nil, log)

	msgChan := make(chan RawMessage, 10)
	postsChan := make(chan *kafkamodels.ParsedLinkedinPost, 10)
	var counter uint64

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.PostsParserWorkerTestable(ctx, msgChan, postsChan, &counter)
		close(done)
	}()

	// Send messages
	msgChan <- RawMessage{Key: []byte("li1"), Value: []byte(`{}`)}
	msgChan <- RawMessage{Key: []byte("li2"), Value: []byte(`{}`)}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(msgChan)
	<-done

	if atomic.LoadInt32(&parseCalls) != 2 {
		t.Errorf("parseCalls = %d, want 2", parseCalls)
	}
}

func TestParsingService_PostsParserWorkerTestable_ChannelClose(t *testing.T) {
	log := logger.New("error")
	svc := NewParsingService(log)

	msgChan := make(chan RawMessage, 10)
	postsChan := make(chan *kafkamodels.ParsedLinkedinPost, 10)

	done := make(chan struct{})
	go func() {
		svc.PostsParserWorkerTestable(context.Background(), msgChan, postsChan, nil)
		close(done)
	}()

	close(msgChan)

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit after channel close")
	}
}

// ================== InsightsParserWorkerTestable Tests ==================

func TestParsingService_InsightsParserWorkerTestable_ProcessesMessages(t *testing.T) {
	log := logger.New("error")

	var parseCalls int32
	parseInsights := func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
		atomic.AddInt32(&parseCalls, 1)
		return []*kafkamodels.ParsedLinkedinInsights{
			{LinkedinID: "li456", CreatedAt: time.Now()},
		}, nil
	}

	svc := NewParsingServiceWithFuncs(nil, parseInsights, log)

	msgChan := make(chan RawMessage, 10)
	insightsChan := make(chan *kafkamodels.ParsedLinkedinInsights, 10)
	var counter uint64

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.InsightsParserWorkerTestable(ctx, msgChan, insightsChan, &counter)
		close(done)
	}()

	// Send messages
	msgChan <- RawMessage{Key: []byte("li1"), Value: []byte(`{}`)}
	msgChan <- RawMessage{Key: []byte("li2"), Value: []byte(`{}`)}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(msgChan)
	<-done

	if atomic.LoadInt32(&parseCalls) != 2 {
		t.Errorf("parseCalls = %d, want 2", parseCalls)
	}
}

func TestParsingService_InsightsParserWorkerTestable_ChannelClose(t *testing.T) {
	log := logger.New("error")
	svc := NewParsingService(log)

	msgChan := make(chan RawMessage, 10)
	insightsChan := make(chan *kafkamodels.ParsedLinkedinInsights, 10)

	done := make(chan struct{})
	go func() {
		svc.InsightsParserWorkerTestable(context.Background(), msgChan, insightsChan, nil)
		close(done)
	}()

	close(msgChan)

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit after channel close")
	}
}
