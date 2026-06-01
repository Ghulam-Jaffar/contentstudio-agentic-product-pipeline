package main

import (
	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
)

// GMBAPI is an alias to the shared interface in clients/social package.
type GMBAPI = social.GMBAPI

// Verify that GMBClient implements GMBAPI
var _ GMBAPI = (*social.GMBClient)(nil)

// UnifiedSocialRepository is an alias to the shared interface in db/mongodb package.
type UnifiedSocialRepository = mongodb.UnifiedSocialRepository

// KafkaProducer is an alias to the shared interface in kafka package.
type KafkaProducer = kafka.Producer

// KafkaConsumer is an alias to the shared interface in kafka package.
type KafkaConsumer = kafka.Consumer
