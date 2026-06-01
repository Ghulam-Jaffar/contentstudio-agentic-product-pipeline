package main

import (
	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
)

// InstagramAPI is an alias to the shared interface in clients/social package.
// This allows services to use the common interface for Instagram API operations.
type InstagramAPI = social.InstagramAPI

// Verify that InstagramClient implements InstagramAPI
var _ InstagramAPI = (*social.InstagramClient)(nil)
