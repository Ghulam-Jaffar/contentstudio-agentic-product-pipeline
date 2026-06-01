package main

import (
	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
)

// FacebookAPI is an alias to the shared interface in clients/social package.
// This allows services to use the common interface for Facebook API operations.
type FacebookAPI = social.FacebookAPI

// KafkaProducer is an alias to the shared interface in kafka package.
// This allows services to use the common interface for Kafka producer operations.
type KafkaProducer = kafka.Producer

// KafkaConsumer is an alias to the shared interface in kafka package.
// This allows services to use the common interface for Kafka consumer operations.
type KafkaConsumer = kafka.Consumer

// Verify that FacebookClient implements FacebookAPI
var _ FacebookAPI = (*social.FacebookClient)(nil)
