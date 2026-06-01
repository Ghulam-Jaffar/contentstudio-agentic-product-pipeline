package competitortokens

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"github.com/redis/go-redis/v9"
)

const maxDistinctTokenFetchAttempts = 6

type RedisCommander interface {
	Do(ctx context.Context, args ...interface{}) *redis.Cmd
}

type Candidate struct {
	AccessToken string
	PlatformID  string
}

func (c Candidate) Key() string {
	return CandidateKey(c.AccessToken, c.PlatformID)
}

func CandidateKey(accessToken, platformID string) string {
	return strings.TrimSpace(accessToken) + "|" + strings.TrimSpace(platformID)
}

func FetchCandidate(
	ctx context.Context,
	redisClient RedisCommander,
	queueName string,
	decryptionKey string,
	requirePlatformID bool,
	exclude map[string]struct{},
) (Candidate, error) {
	var lastErr error

	for attempt := 0; attempt < maxDistinctTokenFetchAttempts; attempt++ {
		cmd := redisClient.Do(ctx, "SRANDMEMBER", queueName)
		if err := cmd.Err(); err != nil {
			return Candidate{}, err
		}

		tokenStr, err := cmd.Text()
		if err != nil {
			return Candidate{}, err
		}

		candidate, err := parseCandidate(tokenStr, decryptionKey, requirePlatformID)
		if err != nil {
			lastErr = err
			continue
		}

		if _, seen := exclude[candidate.Key()]; seen {
			lastErr = fmt.Errorf("FetchCandidate: distinct token not found in %s", queueName)
			continue
		}

		return candidate, nil
	}

	if lastErr != nil {
		return Candidate{}, lastErr
	}

	return Candidate{}, fmt.Errorf("FetchCandidate: unable to fetch token from %s", queueName)
}

func IsInstagramTokenIssue(err error) bool {
	return social.IsAuthError(err)
}

func IsFacebookTokenIssue(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	tokenPatterns := []string{
		"error validating access token",
		"session has been invalidated",
		"oauthexception/190",
		"code: 190",
		"application does not have permission",
		"does not have permission for this action",
	}

	for _, pattern := range tokenPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

func parseCandidate(tokenStr, decryptionKey string, requirePlatformID bool) (Candidate, error) {
	var tokenData map[string]interface{}
	if err := json.Unmarshal([]byte(tokenStr), &tokenData); err != nil {
		return Candidate{}, err
	}

	encryptedToken, ok := tokenData["token"].(string)
	if !ok || strings.TrimSpace(encryptedToken) == "" {
		return Candidate{}, fmt.Errorf("parseCandidate: invalid token data")
	}

	platformID := stringifyValue(tokenData["platform_id"])
	if requirePlatformID && strings.TrimSpace(platformID) == "" {
		return Candidate{}, fmt.Errorf("parseCandidate: invalid token data: missing platform_id")
	}

	accessToken := encryptedToken
	if decryptedToken, err := crypto.DecryptToken(encryptedToken, decryptionKey); err == nil && strings.TrimSpace(decryptedToken) != "" {
		accessToken = decryptedToken
	}

	return Candidate{
		AccessToken: accessToken,
		PlatformID:  platformID,
	}, nil
}

func stringifyValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}
