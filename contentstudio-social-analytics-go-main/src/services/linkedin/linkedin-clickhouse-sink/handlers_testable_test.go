package main

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== NewMessageHandler Tests ==================

func TestNewMessageHandler(t *testing.T) {
	log := logger.New("error")
	h := NewMessageHandler(log)

	if h == nil {
		t.Fatal("NewMessageHandler returned nil")
	}
	if h.log == nil {
		t.Error("log is nil")
	}
}

// ================== HandleParsedPostTestable Tests ==================

func TestMessageHandler_HandleParsedPostTestable_Success(t *testing.T) {
	log := logger.New("error")
	h := NewMessageHandler(log)

	postsChan := make(chan *kafkamodels.ParsedLinkedinPost, 10)

	post := kafkamodels.ParsedLinkedinPost{
		PostID:     "post123",
		LinkedinID: "li456",
	}
	value, _ := json.Marshal(post)

	err := h.HandleParsedPostTestable(context.Background(), []byte("key"), value, postsChan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case received := <-postsChan:
		if received.PostID != "post123" {
			t.Errorf("PostID = %q, want %q", received.PostID, "post123")
		}
	default:
		t.Error("expected post in channel")
	}
}

func TestMessageHandler_HandleParsedPostTestable_UnmarshalError(t *testing.T) {
	log := logger.New("error")
	h := NewMessageHandler(log)

	postsChan := make(chan *kafkamodels.ParsedLinkedinPost, 10)

	err := h.HandleParsedPostTestable(context.Background(), []byte("key"), []byte("invalid json"), postsChan)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMessageHandler_HandleParsedPostTestable_ContextCancel(t *testing.T) {
	log := logger.New("error")
	h := NewMessageHandler(log)

	postsChan := make(chan *kafkamodels.ParsedLinkedinPost) // Unbuffered

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	post := kafkamodels.ParsedLinkedinPost{PostID: "post123"}
	value, _ := json.Marshal(post)

	err := h.HandleParsedPostTestable(ctx, []byte("key"), value, postsChan)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// ================== HandleParsedInsightsTestable Tests ==================

func TestMessageHandler_HandleParsedInsightsTestable_Success(t *testing.T) {
	log := logger.New("error")
	h := NewMessageHandler(log)

	insightsChan := make(chan *kafkamodels.ParsedLinkedinInsights, 10)

	insights := kafkamodels.ParsedLinkedinInsights{
		LinkedinID:       "li456",
		RecordID:         "li456_2024-01-15",
		TotalFollowerCount: 1000,
	}
	value, _ := json.Marshal(insights)

	err := h.HandleParsedInsightsTestable(context.Background(), []byte("key"), value, insightsChan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case received := <-insightsChan:
		if received.LinkedinID != "li456" {
			t.Errorf("LinkedinID = %q, want %q", received.LinkedinID, "li456")
		}
		if received.TotalFollowerCount != 1000 {
			t.Errorf("TotalFollowerCount = %d, want 1000", received.TotalFollowerCount)
		}
	default:
		t.Error("expected insights in channel")
	}
}

func TestMessageHandler_HandleParsedInsightsTestable_UnmarshalError(t *testing.T) {
	log := logger.New("error")
	h := NewMessageHandler(log)

	insightsChan := make(chan *kafkamodels.ParsedLinkedinInsights, 10)

	err := h.HandleParsedInsightsTestable(context.Background(), []byte("key"), []byte("invalid json"), insightsChan)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMessageHandler_HandleParsedInsightsTestable_ContextCancel(t *testing.T) {
	log := logger.New("error")
	h := NewMessageHandler(log)

	insightsChan := make(chan *kafkamodels.ParsedLinkedinInsights) // Unbuffered

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	insights := kafkamodels.ParsedLinkedinInsights{LinkedinID: "li456"}
	value, _ := json.Marshal(insights)

	err := h.HandleParsedInsightsTestable(ctx, []byte("key"), value, insightsChan)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// ================== PostsWorkerTestable Tests ==================

func TestMessageHandler_PostsWorkerTestable_ProcessesMessages(t *testing.T) {
	log := logger.New("error")
	h := NewMessageHandler(log)

	msgChan := make(chan Message, 10)
	postsChan := make(chan *kafkamodels.ParsedLinkedinPost, 10)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		h.PostsWorkerTestable(ctx, msgChan, postsChan)
		close(done)
	}()

	// Send messages
	post1 := kafkamodels.ParsedLinkedinPost{PostID: "post1"}
	post2 := kafkamodels.ParsedLinkedinPost{PostID: "post2"}
	value1, _ := json.Marshal(post1)
	value2, _ := json.Marshal(post2)

	msgChan <- Message{Key: []byte("key1"), Value: value1}
	msgChan <- Message{Key: []byte("key2"), Value: value2}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(msgChan)
	<-done

	if len(postsChan) != 2 {
		t.Errorf("postsChan len = %d, want 2", len(postsChan))
	}
}

func TestMessageHandler_PostsWorkerTestable_ChannelClose(t *testing.T) {
	log := logger.New("error")
	h := NewMessageHandler(log)

	msgChan := make(chan Message, 10)
	postsChan := make(chan *kafkamodels.ParsedLinkedinPost, 10)

	done := make(chan struct{})
	go func() {
		h.PostsWorkerTestable(context.Background(), msgChan, postsChan)
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

func TestMessageHandler_PostsWorkerTestable_ContextCancel(t *testing.T) {
	log := logger.New("error")
	h := NewMessageHandler(log)

	msgChan := make(chan Message, 10)
	postsChan := make(chan *kafkamodels.ParsedLinkedinPost, 10)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		h.PostsWorkerTestable(ctx, msgChan, postsChan)
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

// ================== InsightsWorkerTestable Tests ==================

func TestMessageHandler_InsightsWorkerTestable_ProcessesMessages(t *testing.T) {
	log := logger.New("error")
	h := NewMessageHandler(log)

	msgChan := make(chan Message, 10)
	insightsChan := make(chan *kafkamodels.ParsedLinkedinInsights, 10)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		h.InsightsWorkerTestable(ctx, msgChan, insightsChan)
		close(done)
	}()

	// Send messages
	insights1 := kafkamodels.ParsedLinkedinInsights{LinkedinID: "li1"}
	insights2 := kafkamodels.ParsedLinkedinInsights{LinkedinID: "li2"}
	value1, _ := json.Marshal(insights1)
	value2, _ := json.Marshal(insights2)

	msgChan <- Message{Key: []byte("key1"), Value: value1}
	msgChan <- Message{Key: []byte("key2"), Value: value2}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(msgChan)
	<-done

	if len(insightsChan) != 2 {
		t.Errorf("insightsChan len = %d, want 2", len(insightsChan))
	}
}

func TestMessageHandler_InsightsWorkerTestable_ChannelClose(t *testing.T) {
	log := logger.New("error")
	h := NewMessageHandler(log)

	msgChan := make(chan Message, 10)
	insightsChan := make(chan *kafkamodels.ParsedLinkedinInsights, 10)

	done := make(chan struct{})
	go func() {
		h.InsightsWorkerTestable(context.Background(), msgChan, insightsChan)
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

// ================== Concurrent Tests ==================

func TestMessageHandler_ConcurrentProcessing(t *testing.T) {
	log := logger.New("error")
	h := NewMessageHandler(log)

	msgChan := make(chan Message, 100)
	postsChan := make(chan *kafkamodels.ParsedLinkedinPost, 100)

	ctx, cancel := context.WithCancel(context.Background())

	// Start multiple workers
	var processed int64
	for i := 0; i < 3; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case _, ok := <-postsChan:
					if !ok {
						return
					}
					atomic.AddInt64(&processed, 1)
				}
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		h.PostsWorkerTestable(ctx, msgChan, postsChan)
		close(done)
	}()

	// Send many messages
	for i := 0; i < 50; i++ {
		post := kafkamodels.ParsedLinkedinPost{PostID: "post"}
		value, _ := json.Marshal(post)
		msgChan <- Message{Key: []byte("key"), Value: value}
	}

	time.Sleep(200 * time.Millisecond)

	cancel()
	close(msgChan)
	<-done

	if atomic.LoadInt64(&processed) != 50 {
		t.Errorf("processed = %d, want 50", processed)
	}
}
