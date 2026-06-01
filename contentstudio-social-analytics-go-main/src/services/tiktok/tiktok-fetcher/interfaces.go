package main

import (
	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
)

// TikTokAPI is an alias to the shared interface in clients/social package.
// This allows services to use the common interface for TikTok API operations.
type TikTokAPI = social.TikTokAPI

// Verify that TikTokClient implements TikTokAPI
var _ TikTokAPI = (*social.TikTokClient)(nil)

// UnifiedSocialRepository is an alias to the shared interface in db/mongodb package.
// This allows services to use the common interface for MongoDB operations.
type UnifiedSocialRepository = mongodb.UnifiedSocialRepository

// KafkaConsumer is an alias to the shared interface in kafka package.
// This allows services to use the common interface for Kafka consumer operations.
type KafkaConsumer = kafka.Consumer

// KafkaConsumerInterface is an alias for backward compatibility.
// Prefer using KafkaConsumer for new code.
type KafkaConsumerInterface = kafka.Consumer

// KafkaProducer is an alias to the shared interface in kafka package.
// This allows services to use the common interface for Kafka producer operations.
type KafkaProducer = kafka.Producer

// KafkaProducerInterface is an alias for backward compatibility.
// Prefer using KafkaProducer for new code.
type KafkaProducerInterface = kafka.Producer

// MessageHandler is an alias to the shared type in kafka package.
// This allows services to use the common message handler type.
type MessageHandler = kafka.MessageHandler
