package main

import (
	"context"
	"sync"
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

func createTestLogger() *logger.Logger {
	return logger.New("debug")
}

func TestWorker_StopsOnChannelClose(t *testing.T) {
	log := createTestLogger()
	workChan := make(chan workMessage, 10)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		worker(1, workChan, nil, log)
	}()

	close(workChan)
	wg.Wait()
}

func TestWorkMessage_Struct(t *testing.T) {
	ctx := context.Background()
	msg := workMessage{
		ctx:   ctx,
		value: []byte("test"),
	}

	if msg.ctx != ctx {
		t.Fatal("expected context to match")
	}
	if string(msg.value) != "test" {
		t.Fatal("expected value to match")
	}
}

func TestConstants(t *testing.T) {
	if WorkerPoolSize != 10 {
		t.Fatalf("expected WorkerPoolSize to be 10, got %d", WorkerPoolSize)
	}
	if WorkChannelBuffer != 100 {
		t.Fatalf("expected WorkChannelBuffer to be 100, got %d", WorkChannelBuffer)
	}
	if ConsumerGroup != "youtube-immediate-processor-group" {
		t.Fatalf("unexpected ConsumerGroup: %s", ConsumerGroup)
	}
	if Topic != "immediate-work-order-youtube" {
		t.Fatalf("unexpected Topic: %s", Topic)
	}
}
