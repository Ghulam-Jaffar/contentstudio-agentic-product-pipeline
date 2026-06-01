package main

import (
	"context"
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"

	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestResolveToken(t *testing.T) {
	if got := resolveToken(kafkamodels.MetaAdsWorkOrder{LongAccessToken: "long-token"}, ""); got != "long-token" {
		t.Fatalf("unexpected token: %q", got)
	}
	if got := resolveToken(kafkamodels.MetaAdsWorkOrder{AccessToken: "short-token"}, ""); got != "short-token" {
		t.Fatalf("unexpected token: %q", got)
	}
}

func TestPublishJSONAndBatched(t *testing.T) {
	var messages []string
	producer := &mockProducer{
		produceFn: func(_ context.Context, topic string, key, value []byte) error {
			messages = append(messages, topic+":"+string(key)+":"+string(value))
			return nil
		},
	}
	log := logger.NewNop()

	err := publishJSON(context.Background(), producer, "topic-a", "key-1", map[string]string{"id": "1"}, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = publishBatched(context.Background(), producer, "topic-b", kafkamodels.MetaAdsWorkOrder{MongoID: "mongo-1"}, []int{1, 2, 3}, 2, func(rows []int) interface{} {
		return rows
	}, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}

	if messages[0][:len("topic-a:key-1:")] != "topic-a:key-1:" {
		t.Fatalf("unexpected first message: %s", messages[0])
	}
}

type mockProducer struct {
	produceFn func(context.Context, string, []byte, []byte) error
}

func (m *mockProducer) Produce(ctx context.Context, topic string, key, value []byte) error {
	if m.produceFn != nil {
		return m.produceFn(ctx, topic, key, value)
	}
	return nil
}
func (m *mockProducer) Close() error { return nil }
