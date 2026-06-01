package main

import (
	"context"
	"io"
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	"github.com/rs/zerolog"
)

func TestConstants(t *testing.T) {
	if consumerGroupSuffix != "meta-ads-sink-group" {
		t.Fatalf("unexpected consumer group: %s", consumerGroupSuffix)
	}
	if topicAccountInfo != "raw-meta-ads-account-info" {
		t.Fatalf("unexpected topic: %s", topicAccountInfo)
	}
}

func TestDispatchInvalidJSON(t *testing.T) {
	appLog := logger.NewNop()
	zlog := zerolog.New(io.Discard)
	sink := conversions.NewClickHouseSinkWithClient(&zlog, nil)
	if err := dispatch(context.Background(), topicAccountInfo, []byte("{invalid"), sink, appLog); err != nil {
		t.Fatalf("expected nil error on invalid json path, got %v", err)
	}
}
