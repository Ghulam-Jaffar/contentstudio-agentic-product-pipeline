package main

import (
	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
)

// MockLinkedInClient is an alias to the shared mock in clients/social package.
// This allows services to use the common mock for testing LinkedIn API operations.
type MockLinkedInClient = social.MockLinkedInClient

// MockGeoResolver is an alias to the shared mock in clients/social package.
// This allows services to use the common mock for testing geo resolution operations.
type MockGeoResolver = social.MockGeoResolver

// Verify mocks implement interfaces at compile time
var _ LinkedInAPI = (*MockLinkedInClient)(nil)
var _ GeoResolverAPI = (*MockGeoResolver)(nil)
