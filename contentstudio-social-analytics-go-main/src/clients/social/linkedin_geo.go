package social

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	"github.com/rs/zerolog/log"
)

// GeoResolver handles resolving LinkedIn geo IDs to human-readable names.
// It uses ClickHouse as a cache and falls back to LinkedIn API for unknown IDs.
type GeoResolver struct {
	liClient *LinkedInClient
	chClient *clickhouse.Client
}

// NewGeoResolver creates a new GeoResolver with LinkedIn and ClickHouse clients.
func NewGeoResolver(liClient *LinkedInClient, chClient *clickhouse.Client) *GeoResolver {
	return &GeoResolver{
		liClient: liClient,
		chClient: chClient,
	}
}

// ResolveGeoIDs resolves geo IDs to names using ClickHouse cache first, then LinkedIn API.
// Returns a map of geo_id -> geo_name.
// New mappings from LinkedIn API are automatically saved to ClickHouse cache.
func (r *GeoResolver) ResolveGeoIDs(ctx context.Context, geoIDs []string, accessToken string) (map[string]string, error) {
	// Convert to GeoIDWithType with empty type (for backward compatibility)
	geoIDsWithType := make([]GeoIDWithType, len(geoIDs))
	for i, id := range geoIDs {
		geoIDsWithType[i] = GeoIDWithType{ID: id, Type: ""}
	}
	return r.ResolveGeoIDsWithType(ctx, geoIDsWithType, accessToken)
}

// ResolveGeoIDsWithType resolves geo IDs to names using ClickHouse cache first, then LinkedIn API.
// Includes geo type (country/city) when saving to cache.
// Returns a map of geo_id -> geo_name.
func (r *GeoResolver) ResolveGeoIDsWithType(ctx context.Context, geoIDsWithType []GeoIDWithType, accessToken string) (map[string]string, error) {
	if len(geoIDsWithType) == 0 {
		return map[string]string{}, nil
	}

	// Extract just the IDs for cache lookup
	geoIDs := make([]string, len(geoIDsWithType))
	geoTypeMap := make(map[string]string) // id -> type
	for i, g := range geoIDsWithType {
		geoIDs[i] = g.ID
		geoTypeMap[g.ID] = g.Type
	}

	log.Info().Int("geo_ids_count", len(geoIDs)).Msg("GeoResolver: Starting resolution")

	// Step 1: Check ClickHouse cache for existing mappings
	cachedMappings, err := r.chClient.GetGeoMappings(ctx, geoIDs)
	if err != nil {
		log.Warn().Err(err).Msg("GeoResolver: Failed to get cached mappings, will try API")
		cachedMappings = map[string]string{}
	} else {
		log.Info().Int("cached_count", len(cachedMappings)).Msg("GeoResolver: Found cached mappings")
	}

	// Step 2: Find IDs that are not in cache
	missingIDs := make([]string, 0)
	for _, id := range geoIDs {
		if _, found := cachedMappings[id]; !found {
			missingIDs = append(missingIDs, id)
		}
	}

	// Step 3: If all IDs are cached, return cached results
	if len(missingIDs) == 0 {
		log.Info().Int("total_resolved", len(cachedMappings)).Msg("GeoResolver: All IDs found in cache")
		return cachedMappings, nil
	}

	log.Info().Int("missing_count", len(missingIDs)).Msg("GeoResolver: Calling LinkedIn API for missing IDs")

	// Step 4: Call LinkedIn API for missing IDs
	apiMappings, err := r.liClient.ResolveGeoIDs(ctx, missingIDs, accessToken)
	if err != nil {
		log.Warn().Err(err).Msg("GeoResolver: LinkedIn API failed")
		// API failed, return what we have from cache
		return cachedMappings, nil
	}

	log.Info().Int("api_resolved_count", len(apiMappings)).Msg("GeoResolver: LinkedIn API returned mappings")

	// Step 5: Save new mappings to ClickHouse cache with geo type (async, don't block on errors)
	if len(apiMappings) > 0 {
		// Build GeoMappingWithType slice with type info
		mappingsToSave := make([]clickhouse.GeoMappingWithType, 0, len(apiMappings))
		for id, name := range apiMappings {
			mappingsToSave = append(mappingsToSave, clickhouse.GeoMappingWithType{
				ID:   id,
				Name: name,
				Type: geoTypeMap[id],
			})
		}

		go func() {
			// Use background context since the original might be cancelled
			bgCtx := context.Background()
			if err := r.chClient.InsertGeoMappingsWithType(bgCtx, mappingsToSave); err != nil {
				log.Warn().Err(err).Int("count", len(mappingsToSave)).Msg("GeoResolver: Failed to save mappings to cache")
			} else {
				log.Info().Int("count", len(mappingsToSave)).Msg("GeoResolver: Saved new mappings to cache")
			}
		}()
	}

	// Step 6: Merge cached and API results
	result := make(map[string]string, len(cachedMappings)+len(apiMappings))
	for id, name := range cachedMappings {
		result[id] = name
	}
	for id, name := range apiMappings {
		result[id] = name
	}

	log.Info().Int("total_resolved", len(result)).Msg("GeoResolver: Resolution complete")

	return result, nil
}

// ResolveGeoIDsFromFollowerData extracts geo IDs from follower data JSON and resolves them.
// This is a convenience method that combines extraction and resolution.
func (r *GeoResolver) ResolveGeoIDsFromFollowerData(ctx context.Context, followerDataJSON []byte, accessToken string) (map[string]string, error) {
	geoIDs := ExtractGeoIDsFromJSON(followerDataJSON)
	if len(geoIDs) == 0 {
		return map[string]string{}, nil
	}
	return r.ResolveGeoIDs(ctx, geoIDs, accessToken)
}

// ExtractGeoIDsFromJSON extracts unique geo IDs from follower data JSON.
func ExtractGeoIDsFromJSON(data []byte) []string {
	var parsed struct {
		Elements []struct {
			FollowerCountsByGeoCountry []struct {
				Geo string `json:"geo"`
			} `json:"followerCountsByGeoCountry"`
			FollowerCountsByGeo []struct {
				Geo string `json:"geo"`
			} `json:"followerCountsByGeo"`
		} `json:"elements"`
	}

	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil
	}

	geoIDSet := make(map[string]struct{})
	for _, el := range parsed.Elements {
		for _, gc := range el.FollowerCountsByGeoCountry {
			if idx := strings.LastIndex(gc.Geo, ":"); idx >= 0 {
				geoIDSet[gc.Geo[idx+1:]] = struct{}{}
			}
		}
		for _, g := range el.FollowerCountsByGeo {
			if idx := strings.LastIndex(g.Geo, ":"); idx >= 0 {
				geoIDSet[g.Geo[idx+1:]] = struct{}{}
			}
		}
	}

	geoIDs := make([]string, 0, len(geoIDSet))
	for id := range geoIDSet {
		geoIDs = append(geoIDs, id)
	}
	return geoIDs
}

// ExtractGeoIDsWithTypeFromJSON extracts unique geo IDs with their types from follower data JSON.
// Returns a slice of GeoIDWithType with type "country" or "city".
func ExtractGeoIDsWithTypeFromJSON(data []byte) []GeoIDWithType {
	var parsed struct {
		Elements []struct {
			FollowerCountsByGeoCountry []struct {
				Geo string `json:"geo"`
			} `json:"followerCountsByGeoCountry"`
			FollowerCountsByGeo []struct {
				Geo string `json:"geo"`
			} `json:"followerCountsByGeo"`
		} `json:"elements"`
	}

	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil
	}

	geoIDMap := make(map[string]string) // id -> type
	for _, el := range parsed.Elements {
		for _, gc := range el.FollowerCountsByGeoCountry {
			if idx := strings.LastIndex(gc.Geo, ":"); idx >= 0 {
				geoIDMap[gc.Geo[idx+1:]] = "country"
			}
		}
		for _, g := range el.FollowerCountsByGeo {
			if idx := strings.LastIndex(g.Geo, ":"); idx >= 0 {
				geoIDMap[g.Geo[idx+1:]] = "city"
			}
		}
	}

	result := make([]GeoIDWithType, 0, len(geoIDMap))
	for id, geoType := range geoIDMap {
		result = append(result, GeoIDWithType{ID: id, Type: geoType})
	}
	return result
}
