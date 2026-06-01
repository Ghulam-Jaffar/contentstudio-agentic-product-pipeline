package parser

import (
	"fmt"
	"strings"
)

// ConstructAuthorProfileURL builds a platform profile URL from Data365-derived
// handle/id fields, preferring an explicit provider URL when one is available.
func ConstructAuthorProfileURL(platform, authorHandle, authorID, explicitURL string) string {
	if explicitURL != "" {
		return explicitURL
	}

	handle := strings.TrimSpace(authorHandle)
	handle = strings.TrimPrefix(handle, "@")

	switch platform {
	case "instagram":
		if handle != "" {
			return fmt.Sprintf("https://www.instagram.com/%s/", handle)
		}
	case "facebook":
		if handle != "" {
			return fmt.Sprintf("https://www.facebook.com/%s", handle)
		}
		if authorID != "" {
			return fmt.Sprintf("https://www.facebook.com/profile.php?id=%s", authorID)
		}
	case "twitter":
		if handle != "" {
			return fmt.Sprintf("https://x.com/%s", handle)
		}
		if authorID != "" {
			return fmt.Sprintf("https://x.com/%s", authorID)
		}
	case "tiktok":
		if handle != "" {
			return fmt.Sprintf("https://www.tiktok.com/@%s", handle)
		}
	case "threads":
		if handle != "" {
			return fmt.Sprintf("https://www.threads.net/@%s", handle)
		}
	case "reddit":
		user := handle
		if user == "" {
			user = strings.TrimSpace(authorID)
		}
		user = strings.TrimPrefix(user, "u/")
		user = strings.TrimPrefix(user, "user/")
		if user != "" {
			return fmt.Sprintf("https://www.reddit.com/user/%s", user)
		}
	}

	return ""
}
