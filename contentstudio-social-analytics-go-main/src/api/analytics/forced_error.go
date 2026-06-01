package analytics

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/rs/zerolog"
)

const (
	forcedErrorStatusEnv    = "ANALYTICS_FORCE_ERROR_STATUS"
	forcedErrorStatusesEnv  = "ANALYTICS_FORCE_ERROR_STATUSES"
	forcedErrorPlatformsEnv = "ANALYTICS_FORCE_ERROR_PLATFORMS"
	forcedErrorMessageEnv   = "ANALYTICS_FORCE_ERROR_MESSAGE"
	defaultForcedMessage    = "Forced analytics error for UI testing"
)

type ForcedErrorConfig struct {
	StatusCodes []int
	Message     string
	Platforms   []string
	Paths       []string
}

func ForcedErrorConfigFromEnv(log zerolog.Logger) *ForcedErrorConfig {
	statusCodes := parseStatusCodes(log)
	if len(statusCodes) == 0 {
		return nil
	}

	platforms := parsePlatforms(log)
	if len(platforms) == 0 {
		platforms = []string{"facebook", "linkedin", "gmb"}
	}

	message := strings.TrimSpace(os.Getenv(forcedErrorMessageEnv))
	return &ForcedErrorConfig{
		StatusCodes: statusCodes,
		Message:     message,
		Platforms:   platforms,
	}
}

func WithForcedError(next http.Handler, cfg *ForcedErrorConfig) http.Handler {
	if cfg == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.matchesPath(r.URL.Path) {
			statusCode := cfg.pickStatusCode()
			httputil.WriteStatusError(w, statusCode, cfg.messageForStatus(statusCode))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseStatusCodes(log zerolog.Logger) []int {
	statusesRaw := strings.TrimSpace(os.Getenv(forcedErrorStatusesEnv))
	if statusesRaw == "" {
		statusesRaw = strings.TrimSpace(os.Getenv(forcedErrorStatusEnv))
	}
	if statusesRaw == "" {
		return nil
	}

	parts := strings.Split(statusesRaw, ",")
	statusCodes := make([]int, 0, len(parts))

	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}

		statusCode, err := strconv.Atoi(value)
		if err != nil || statusCode < http.StatusBadRequest || statusCode > http.StatusNetworkAuthenticationRequired {
			log.Warn().
				Str("value", value).
				Msg("Ignoring invalid analytics forced error status")
			continue
		}

		statusCodes = append(statusCodes, statusCode)
	}

	return statusCodes
}

func parsePlatforms(log zerolog.Logger) []string {
	platformsRaw := strings.TrimSpace(os.Getenv(forcedErrorPlatformsEnv))
	if platformsRaw == "" {
		return nil
	}

	allowed := []string{"facebook", "linkedin", "gmb"}
	seen := make(map[string]struct{}, len(allowed))
	platforms := make([]string, 0, len(allowed))

	for _, part := range strings.Split(platformsRaw, ",") {
		platform := strings.ToLower(strings.TrimSpace(part))
		if platform == "" {
			continue
		}
		if !slices.Contains(allowed, platform) {
			log.Warn().
				Str("platform", platform).
				Msg("Ignoring invalid analytics forced error platform")
			continue
		}
		if _, ok := seen[platform]; ok {
			continue
		}
		seen[platform] = struct{}{}
		platforms = append(platforms, platform)
	}

	return platforms
}

func (c *ForcedErrorConfig) pickStatusCode() int {
	if len(c.StatusCodes) == 1 {
		return c.StatusCodes[0]
	}

	return c.StatusCodes[rand.Intn(len(c.StatusCodes))]
}

func (c *ForcedErrorConfig) messageForStatus(statusCode int) string {
	if c.Message != "" {
		return c.Message
	}

	return fmt.Sprintf("%s (%d)", defaultForcedMessage, statusCode)
}

func (c *ForcedErrorConfig) matchesPath(path string) bool {
	if len(c.Paths) > 0 {
		return slices.Contains(c.Paths, path)
	}

	const prefix = "/analytics/overview/"
	if !strings.HasPrefix(path, prefix) {
		return false
	}

	trimmed := strings.TrimPrefix(path, prefix)
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return false
	}

	platform := strings.ToLower(parts[0])
	return slices.Contains(c.Platforms, platform)
}
