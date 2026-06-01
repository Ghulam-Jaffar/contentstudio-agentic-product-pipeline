package main

import (
	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
)

// MockInstagramClient is an alias to the shared mock in clients/social package.
// This allows services to use the common mock for testing Instagram API operations.
type MockInstagramClient = social.MockInstagramClient

// Verify mock implements interface at compile time
var _ InstagramAPI = (*MockInstagramClient)(nil)
