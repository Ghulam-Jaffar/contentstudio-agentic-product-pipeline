package main

import (
	"testing"
	"time"

	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Kafka Topics Tests ==================

func TestKafkaTopicsConstants(t *testing.T) {
	if kafkamodels.PinterestKafkaTopics.RawUsers == "" {
		t.Error("RawUsers topic should not be empty")
	}
	if kafkamodels.PinterestKafkaTopics.RawBoards == "" {
		t.Error("RawBoards topic should not be empty")
	}
	if kafkamodels.PinterestKafkaTopics.RawPins == "" {
		t.Error("RawPins topic should not be empty")
	}
	if kafkamodels.PinterestKafkaTopics.RawPinInsights == "" {
		t.Error("RawPinInsights topic should not be empty")
	}
	if kafkamodels.PinterestKafkaTopics.RawUserInsights == "" {
		t.Error("RawUserInsights topic should not be empty")
	}
}

// ================== Batch Configuration Tests ==================

func TestBatchConfigurationValues(t *testing.T) {
	if maxBatchSize <= 0 {
		t.Errorf("maxBatchSize should be positive, got %d", maxBatchSize)
	}
	if batchTimeout <= 0 {
		t.Errorf("batchTimeout should be positive, got %v", batchTimeout)
	}
	if batchProcessorsPerType <= 0 {
		t.Errorf("batchProcessorsPerType should be positive, got %d", batchProcessorsPerType)
	}
	if messageChanSize <= 0 {
		t.Errorf("messageChanSize should be positive, got %d", messageChanSize)
	}
}

func TestBatchConfigurationReasonableValues(t *testing.T) {
	if maxBatchSize < 100 || maxBatchSize > 100000 {
		t.Errorf("maxBatchSize = %d, should be between 100 and 100000", maxBatchSize)
	}
	if batchTimeout < time.Second || batchTimeout > time.Minute {
		t.Errorf("batchTimeout = %v, should be between 1s and 1m", batchTimeout)
	}
	if batchProcessorsPerType < 1 || batchProcessorsPerType > 20 {
		t.Errorf("batchProcessorsPerType = %d, should be between 1 and 20", batchProcessorsPerType)
	}
}

// ================== Idle Configuration Tests ==================

func TestIdleConfigurationValues(t *testing.T) {
	if idleTimeout <= 0 {
		t.Errorf("idleTimeout should be positive, got %v", idleTimeout)
	}
	if idleCheckInterval <= 0 {
		t.Errorf("idleCheckInterval should be positive, got %v", idleCheckInterval)
	}
	if idleCheckInterval >= idleTimeout {
		t.Errorf("idleCheckInterval (%v) should be less than idleTimeout (%v)", idleCheckInterval, idleTimeout)
	}
}

// ================== Consumer Group Tests ==================

func TestConsumerGroupValue(t *testing.T) {
	if consumerGroup == "" {
		t.Error("consumerGroup should not be empty")
	}
	expectedPrefix := "pinterest-"
	if len(consumerGroup) < len(expectedPrefix) {
		t.Errorf("consumerGroup should start with %q", expectedPrefix)
	}
}
