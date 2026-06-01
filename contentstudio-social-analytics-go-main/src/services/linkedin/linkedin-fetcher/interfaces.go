package main

import (
	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
)

// LinkedInAPI is an alias to the shared interface in clients/social package.
// This allows services to use the common interface for LinkedIn API operations.
type LinkedInAPI = social.LinkedInAPI

// GeoResolverAPI is an alias to the shared interface in clients/social package.
// This allows services to use the common interface for geo resolution operations.
type GeoResolverAPI = social.GeoResolverAPI

// Verify that LinkedInClient implements LinkedInAPI
var _ LinkedInAPI = (*social.LinkedInClient)(nil)

// Verify that GeoResolver implements GeoResolverAPI
var _ GeoResolverAPI = (*social.GeoResolver)(nil)
