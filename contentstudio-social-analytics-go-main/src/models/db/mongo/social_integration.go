package mongo

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SocialIntegration represents a unified social platform account from the social_integrations collection
type SocialIntegration struct {
	// Core fields (all platforms)
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	PlatformType       string             `bson:"platform_type"`
	PlatformIdentifier string             `bson:"platform_identifier"`
	PlatformName       string             `bson:"platform_name"`
	PlatformURL        string             `bson:"platform_url,omitempty"`
	PlatformLogo       string             `bson:"platform_logo,omitempty"`
	Type               string             `bson:"type"`
	WorkspaceID        primitive.ObjectID `bson:"workspace_id"`
	UserID             primitive.ObjectID `bson:"user_id"`
	// Raw string representations preserved from MongoDB to handle documents where
	// user_id or workspace_id is stored as a BSON String instead of ObjectID.
	UserIDStr      string     `bson:"-" json:"-"`
	WorkspaceIDStr string     `bson:"-" json:"-"`
	AddedBy        string     `bson:"added_by"`
	State          string     `bson:"state"`
	Validity       string     `bson:"validity"`
	CreatedAt      *MongoTime `bson:"created_at,omitempty"`
	UpdatedAt      *MongoTime `bson:"updated_at,omitempty"`

	// Token fields
	AccessToken      string      `bson:"access_token,omitempty"`
	RefreshToken     string      `bson:"refresh_token,omitempty"`
	LongAccessToken  string      `bson:"long_access_token,omitempty"`
	OAuthToken       string      `bson:"oauth_token,omitempty"`
	OAuthTokenSecret string      `bson:"oauth_token_secret,omitempty"`
	TokenExpiresAt   interface{} `bson:"token_expires_at,omitempty"`
	TokenIssuedAt    interface{} `bson:"token_issued_at,omitempty"`
	ExpiresIn        interface{} `bson:"expires_in,omitempty"`
	RefreshExpiresIn interface{} `bson:"refresh_expires_in,omitempty"`
	Scope            string      `bson:"scope,omitempty"`

	// Analytics timestamps
	LastAnalyticsUpdatedAt         string `bson:"last_analytics_updated_at,omitempty"` // we need to confirm why it was mongodb time
	LastInsightsAnalyticsUpdatedAt string `bson:"last_insights_analytics_updated_at,omitempty"`
	LastFansAnalyticsUpdatedAt     string `bson:"last_fans_analytics_updated_at,omitempty"`
	LastVideoAnalyticsUpdatedAt    string `bson:"last_video_analytics_updated_at,omitempty"`
	LastGroupAnalyticsUpdatedAt    string `bson:"last_group_analytics_updated_at,omitempty"`
	LastLinkPreviewUpdatedAt       string `bson:"last_link_preview_updated_at,omitempty"`

	// Facebook-specific fields
	FanCount int `bson:"fan_count,omitempty"`
	//Permission []Permissions `bson:"permissions,omitempty"`
	PostedAs interface{} `bson:"posted_as,omitempty"`

	// Instagram-specific fields
	UserDetails interface{} `bson:"user_details,omitempty"`
	IsBusiness  bool        `bson:"is_business,omitempty"`
	Username    string      `bson:"username,omitempty"`

	// Twitter-specific fields
	ScreenName       string `bson:"screen_name,omitempty"`
	Verified         bool   `bson:"verified,omitempty"`
	VerifiedType     string `bson:"verified_type,omitempty"`
	SubscriptionType string `bson:"subscription_type,omitempty"`
	DeveloperAppID   string `bson:"developer_app_id,omitempty"`
	APIKey           string `bson:"api_key,omitempty"`
	APISecret        string `bson:"api_secret,omitempty"`

	// LinkedIn-specific fields
	Headline          string `bson:"headline,omitempty"`
	LinkedinProfileID string `bson:"linkedin_profile_id,omitempty"`

	// Google My Business-specific fields
	MetaData     interface{} `bson:"meta_data,omitempty"`
	Locality     string      `bson:"locality,omitempty"`
	PostalCode   string      `bson:"postal_code,omitempty"`
	RegionCode   string      `bson:"region_code,omitempty"`
	LanguageCode string      `bson:"language_code,omitempty"`
	LocationID   string      `bson:"location_id,omitempty"`

	// Pinterest-specific fields
	ProfileID   string        `bson:"profile_id,omitempty"`
	BoardID     string        `bson:"board_id,omitempty"`
	LinkedTo    *LinkedToInfo `bson:"linked_to,omitempty" json:"linked_to,omitempty"`
	URL         string        `bson:"url,omitempty"`
	PinterestID string        `bson:"pinterest_id,omitempty"`

	// Scheduling
	QueueSlots []interface{} `bson:"QueueSlots,omitempty"`

	// Validity tracking
	InvalidTries     int    `bson:"invalid_tries,omitempty"`
	SentInvalidEmail int    `bson:"sent_invalid_email,omitempty"`
	ValidityError    string `bson:"validity_error,omitempty"`
	ValidityStatus   int    `bson:"validity_status,omitempty"`
	LimitExceedTries int    `bson:"limit_exceed_tries,omitempty"`

	// External connections
	ConnectionViaLink bool               `bson:"connection_via_link,omitempty"`
	ConnectionLinkID  primitive.ObjectID `bson:"connection_link_id,omitempty"`

	// Preferences and extra data
	Preferences map[string]interface{} `bson:"preferences,omitempty"`
	ExtraData   map[string]interface{} `bson:",inline"`

	// Legacy field mappings for backward compatibility during migration
	// These map to platform_identifier but are kept for Python services compatibility
	FacebookID  string `bson:"facebook_id,omitempty"`
	InstagramID string `bson:"instagram_id,omitempty"`
	LinkedinID  string `bson:"linkedin_id,omitempty"`
	TwitterID   string `bson:"twitter_id,omitempty"`
}

type LinkedToInfo struct {
	Zapier bool `bson:"zapier,omitempty" json:"zapier,omitempty"`
}

type DBSocialIntegration struct {
	// Core fields (all platforms)
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	PlatformType       string             `bson:"platform_type"`
	PlatformIdentifier interface{}        `bson:"platform_identifier"`
	PlatformName       string             `bson:"platform_name"`
	PlatformURL        string             `bson:"platform_url,omitempty"`
	PlatformLogo       string             `bson:"platform_logo,omitempty"`
	Type               string             `bson:"type"`
	WorkspaceID        interface{}        `bson:"workspace_id"`
	UserID             interface{}        `bson:"user_id"`
	AddedBy            string             `bson:"added_by"`
	State              string             `bson:"state"`
	Validity           string             `bson:"validity"`
	CreatedAt          *MongoTime         `bson:"created_at,omitempty"`
	UpdatedAt          *MongoTime         `bson:"updated_at,omitempty"`

	// Token fields - interface{} to handle both string and embedded document formats (YouTube)
	AccessToken      interface{} `bson:"access_token,omitempty"`
	RefreshToken     interface{} `bson:"refresh_token,omitempty"`
	LongAccessToken  string      `bson:"long_access_token,omitempty"`
	OAuthToken       string      `bson:"oauth_token,omitempty"`
	OAuthTokenSecret string      `bson:"oauth_token_secret,omitempty"`
	TokenExpiresAt   interface{} `bson:"token_expires_at,omitempty"`
	TokenIssuedAt    interface{} `bson:"token_issued_at,omitempty"`
	ExpiresIn        interface{} `bson:"expires_in,omitempty"`
	RefreshExpiresIn interface{} `bson:"refresh_expires_in,omitempty"`
	Scope            string      `bson:"scope,omitempty"`

	// Analytics timestamps
	LastAnalyticsUpdatedAt         interface{} `bson:"last_analytics_updated_at,omitempty"` // we need to confirm why it was mongodb time
	LastInsightsAnalyticsUpdatedAt interface{} `bson:"last_insights_analytics_updated_at,omitempty"`
	LastFansAnalyticsUpdatedAt     interface{} `bson:"last_fans_analytics_updated_at,omitempty"`
	LastVideoAnalyticsUpdatedAt    interface{} `bson:"last_video_analytics_updated_at,omitempty"`
	LastGroupAnalyticsUpdatedAt    interface{} `bson:"last_group_analytics_updated_at,omitempty"`
	LastLinkPreviewUpdatedAt       interface{} `bson:"last_link_preview_updated_at,omitempty"`

	// Facebook-specific fields
	FanCount int `bson:"fan_count,omitempty"`
	//Permission []Permissions `bson:"permissions,omitempty"`
	PostedAs interface{} `bson:"posted_as,omitempty"`

	// Instagram-specific fields
	UserDetails interface{} `bson:"user_details,omitempty"`
	IsBusiness  bool        `bson:"is_business,omitempty"`
	Username    string      `bson:"username,omitempty"`

	// Twitter-specific fields
	ScreenName       string `bson:"screen_name,omitempty"`
	Verified         bool   `bson:"verified,omitempty"`
	VerifiedType     string `bson:"verified_type,omitempty"`
	SubscriptionType string `bson:"subscription_type,omitempty"`
	DeveloperAppID   string `bson:"developer_app_id,omitempty"`
	APIKey           string `bson:"api_key,omitempty"`
	APISecret        string `bson:"api_secret,omitempty"`

	// LinkedIn-specific fields
	Headline          string `bson:"headline,omitempty"`
	LinkedinProfileID string `bson:"linkedin_profile_id,omitempty"`

	// Google My Business-specific fields
	MetaData     interface{} `bson:"meta_data,omitempty"`
	Locality     string      `bson:"locality,omitempty"`
	PostalCode   string      `bson:"postal_code,omitempty"`
	RegionCode   string      `bson:"region_code,omitempty"`
	LanguageCode string      `bson:"language_code,omitempty"`
	LocationID   string      `bson:"location_id,omitempty"`

	// Pinterest-specific fields
	ProfileID   string        `bson:"profile_id,omitempty"`
	BoardID     string        `bson:"board_id,omitempty"`
	LinkedTo    *LinkedToInfo `bson:"linked_to,omitempty" json:"linked_to,omitempty"`
	URL         string        `bson:"url,omitempty"`
	PinterestID string        `bson:"pinterest_id,omitempty"`

	// Scheduling
	QueueSlots []interface{} `bson:"QueueSlots,omitempty"`

	// Validity tracking
	InvalidTries     int    `bson:"invalid_tries,omitempty"`
	SentInvalidEmail int    `bson:"sent_invalid_email,omitempty"`
	ValidityError    string `bson:"validity_error,omitempty"`
	ValidityStatus   int    `bson:"validity_status,omitempty"`
	LimitExceedTries int    `bson:"limit_exceed_tries,omitempty"`

	// External connections
	ConnectionViaLink bool               `bson:"connection_via_link,omitempty"`
	ConnectionLinkID  primitive.ObjectID `bson:"connection_link_id,omitempty"`

	// Preferences and extra data
	Preferences map[string]interface{} `bson:"preferences,omitempty"`
	ExtraData   map[string]interface{} `bson:",inline"`

	// Legacy field mappings for backward compatibility during migration
	// These map to platform_identifier but are kept for Python services compatibility
	FacebookID  string      `bson:"facebook_id,omitempty"`
	InstagramID interface{} `bson:"instagram_id,omitempty"`
	LinkedinID  interface{} `bson:"linkedin_id,omitempty"`
	TwitterID   string      `bson:"twitter_id,omitempty"`
}

type Permissions struct {
	Permission string `bson:"permission,omitempty"`
	staus      string `bson:"staus,omitempty"`
}

// Platform type constants
const (
	PlatformFacebook  = "facebook"
	PlatformInstagram = "instagram"
	PlatformLinkedIn  = "linkedin"
	PlatformTwitter   = "twitter"
	PlatformGMB       = "gmb"
	PlatformPinterest = "pinterest"
	PlatformYouTube   = "youtube"
	PlatformTikTok    = "tiktok"
	PlatformMetaAds   = "meta_ads"
)

// Account state constants
const (
	StateAdded     = "Added"
	StateDeleted   = "Deleted"
	StateDisabled  = "Disabled"
	StateFailed    = "Failed"
	StatePaused    = "Paused"
	StateSyncing   = "Syncing"
	StateProcessed = "Processed"
	StateNotFound  = "NotFound"
)

// Validity constants
const (
	ValidityValid        = "valid"
	ValidityInvalid      = "invalid"
	ValidityExpired      = "expired"
	ValidityExpiringSoon = "expiring_soon"
)

// Super admin state constants
const (
	SuperAdminStateActive  = "active"
	SuperAdminStatePastDue = "past_due"
)

// Account type constants
const (
	TypePage     = "Page"
	TypeProfile  = "Profile"
	TypeBusiness = "Business"
	TypeGroup    = "Group"
	TypeBoard    = "Board"
	TypeCreator  = "Creator"
)

// GetUserIDHex returns the user ID as a 24-char hex string.
// It first tries the parsed ObjectID; if that is zero (e.g. because the field
// was stored as a BSON String and toObjectID failed), it falls back to the raw
// string value captured at decode time.
func (s *SocialIntegration) GetUserIDHex() string {
	if !s.UserID.IsZero() {
		return s.UserID.Hex()
	}
	return s.UserIDStr
}

// GetWorkspaceIDHex returns the workspace ID as a 24-char hex string.
// Same fallback strategy as GetUserIDHex.
func (s *SocialIntegration) GetWorkspaceIDHex() string {
	if !s.WorkspaceID.IsZero() {
		return s.WorkspaceID.Hex()
	}
	return s.WorkspaceIDStr
}

// GetPlatformID returns the platform-specific ID based on platform type
func (s *SocialIntegration) GetPlatformID() string {
	// Always use platform_identifier as the primary source
	if s.PlatformIdentifier != "" {
		return ToString(s.PlatformIdentifier)
	}

	// Fallback to legacy fields if platform_identifier is empty (for backward compatibility)
	switch s.PlatformType {
	case PlatformFacebook:
		return s.FacebookID
	case PlatformInstagram:
		return s.InstagramID
	case PlatformLinkedIn:
		return ToString(s.LinkedinID)
	case PlatformTwitter:
		return s.TwitterID
	case PlatformGMB:
		return s.LocationID
	case PlatformPinterest:
		return s.PinterestID
	default:
		return ""
	}
}

// GetAccessToken returns the appropriate access token for the platform
func (s *SocialIntegration) GetAccessToken() string {
	// For Facebook, prefer long_access_token if available
	if s.PlatformType == PlatformFacebook && s.LongAccessToken != "" {
		return s.LongAccessToken
	}

	// For Instagram, check if token is in user_details
	if s.PlatformType == PlatformInstagram && s.UserDetails != nil {
		if details, ok := s.UserDetails.(map[string]interface{}); ok {
			if token, exists := details["access_token"].(string); exists && token != "" {
				return token
			}
		}
	}

	// For Twitter, use OAuth token
	if s.PlatformType == PlatformTwitter && s.OAuthToken != "" {
		return s.OAuthToken
	}

	// Default to main access_token field
	return s.AccessToken
}

// IsValid checks if the account is valid and active
func (s *SocialIntegration) IsValid() bool {
	return s.Validity == ValidityValid &&
		(s.State == StateAdded || s.State == StateSyncing || s.State == StateProcessed)
}

//// NeedsTokenRefresh checks if the token needs refreshing
//func (s *SocialIntegration) NeedsTokenRefresh() bool {
//	if s.TokenExpiresAt == nil {
//		return false
//	}
//
//	// Check if token expires in the next 7 days
//	// Implementation would check against current time
//	// This is a placeholder - actual implementation would use time.Now()
//	return false
//}

// GetDisplayName returns the display name for the account
func (s *SocialIntegration) GetDisplayName() string {
	if s.PlatformName != "" {
		return s.PlatformName
	}

	// Fallback to username for Instagram/Twitter
	if s.Username != "" {
		return s.Username
	}

	if s.ScreenName != "" {
		return s.ScreenName
	}

	return s.PlatformIdentifier
}

func ToString(v interface{}) string {
	switch val := v.(type) {
	case int, int32, int64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%.0f", val) // optional: trim decimals
	case string:
		return val
	default:
		return ""
	}
}

func toObjectID(val interface{}) primitive.ObjectID {
	switch v := val.(type) {
	case primitive.ObjectID:
		return v
	case string:
		if oid, err := primitive.ObjectIDFromHex(v); err == nil {
			return oid
		}
	}
	return primitive.NilObjectID
}

// toIDString converts a MongoDB ID field (stored as either a BSON ObjectID or
// a plain string) into its hex string representation. Used to preserve the raw
// value when toObjectID cannot parse it.
func toIDString(val interface{}) string {
	switch v := val.(type) {
	case primitive.ObjectID:
		if !v.IsZero() {
			return v.Hex()
		}
	case string:
		return v
	}
	return ""
}

func toString(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case int, int32, int64:
		return fmt.Sprintf("%v", v)
	case float64, float32:
		return fmt.Sprintf("%.0f", v)
	default:
		return ""
	}
}

// extractTokens handles both string tokens and embedded document tokens (YouTube format)
// YouTube stores tokens as embedded document:
//
//	access_token: {
//	  access_token: 'encrypted_token',
//	  refresh_token: 'encrypted_refresh_token',
//	  expires_in: 3599,
//	  scope: '...',
//	  token_type: 'Bearer',
//	  created: timestamp
//	}
func extractTokens(accessTokenVal, refreshTokenVal interface{}, platformType string) (accessToken, refreshToken string) {
	// Handle access_token field - can be string or embedded document
	switch v := accessTokenVal.(type) {
	case string:
		accessToken = v
	case primitive.M:
		// YouTube embedded document format (unordered map): access_token.access_token
		if token, ok := v["access_token"].(string); ok {
			accessToken = token
		} else if token, ok := v["token"].(string); ok {
			accessToken = token
		}
		// Extract refresh_token from embedded document: access_token.refresh_token
		if rt, ok := v["refresh_token"].(string); ok {
			refreshToken = rt
		}
	case primitive.D:
		// BSON decodes embedded documents as primitive.D (ordered slice of key-value pairs)
		m := v.Map()
		if token, ok := m["access_token"].(string); ok {
			accessToken = token
		} else if token, ok := m["token"].(string); ok {
			accessToken = token
		}
		if rt, ok := m["refresh_token"].(string); ok {
			refreshToken = rt
		}
	case map[string]interface{}:
		// Alternative map format (after JSON unmarshaling)
		if token, ok := v["access_token"].(string); ok {
			accessToken = token
		} else if token, ok := v["token"].(string); ok {
			accessToken = token
		}
		if rt, ok := v["refresh_token"].(string); ok {
			refreshToken = rt
		}
	}

	// Handle refresh_token (if not already extracted from embedded document)
	if refreshToken == "" {
		switch v := refreshTokenVal.(type) {
		case string:
			refreshToken = v
		case primitive.M:
			if token, ok := v["token"].(string); ok {
				refreshToken = token
			} else if token, ok := v["refresh_token"].(string); ok {
				refreshToken = token
			}
		case primitive.D:
			m := v.Map()
			if token, ok := m["token"].(string); ok {
				refreshToken = token
			} else if token, ok := m["refresh_token"].(string); ok {
				refreshToken = token
			}
		case map[string]interface{}:
			if token, ok := v["token"].(string); ok {
				refreshToken = token
			} else if token, ok := v["refresh_token"].(string); ok {
				refreshToken = token
			}
		}
	}

	return accessToken, refreshToken
}

func ConvertDBToSocialIntegration(db DBSocialIntegration) SocialIntegration {
	// Extract access token - handle both string and embedded document formats
	accessToken, refreshToken := extractTokens(db.AccessToken, db.RefreshToken, db.PlatformType)

	return SocialIntegration{
		// Core fields
		ID:                 db.ID,
		PlatformType:       db.PlatformType,
		PlatformIdentifier: toString(db.PlatformIdentifier),
		PlatformName:       db.PlatformName,
		PlatformURL:        db.PlatformURL,
		PlatformLogo:       db.PlatformLogo,
		Type:               db.Type,
		WorkspaceID:        toObjectID(db.WorkspaceID),
		WorkspaceIDStr:     toIDString(db.WorkspaceID),
		UserID:             toObjectID(db.UserID),
		UserIDStr:          toIDString(db.UserID),
		AddedBy:            db.AddedBy,
		State:              db.State,
		Validity:           db.Validity,
		CreatedAt:          db.CreatedAt,
		UpdatedAt:          db.UpdatedAt,

		// Token fields
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		LongAccessToken:  db.LongAccessToken,
		OAuthToken:       db.OAuthToken,
		OAuthTokenSecret: db.OAuthTokenSecret,
		ExpiresIn:        db.ExpiresIn,
		RefreshExpiresIn: db.RefreshExpiresIn,
		Scope:            db.Scope,

		// Analytics
		LastAnalyticsUpdatedAt:         parseMongoDateToString(db.LastAnalyticsUpdatedAt),
		LastInsightsAnalyticsUpdatedAt: parseMongoDateToString(db.LastInsightsAnalyticsUpdatedAt),
		LastFansAnalyticsUpdatedAt:     parseMongoDateToString(db.LastFansAnalyticsUpdatedAt),
		LastVideoAnalyticsUpdatedAt:    parseMongoDateToString(db.LastVideoAnalyticsUpdatedAt),
		LastGroupAnalyticsUpdatedAt:    parseMongoDateToString(db.LastGroupAnalyticsUpdatedAt),
		LastLinkPreviewUpdatedAt:       parseMongoDateToString(db.LastLinkPreviewUpdatedAt),

		// Facebook-specific
		FanCount: db.FanCount,
		//Permission: db.Permission,
		PostedAs: db.PostedAs,

		// Instagram-specific
		UserDetails: db.UserDetails,
		IsBusiness:  db.IsBusiness,
		Username:    db.Username,

		// Twitter-specific
		ScreenName:       db.ScreenName,
		Verified:         db.Verified,
		VerifiedType:     db.VerifiedType,
		SubscriptionType: db.SubscriptionType,
		DeveloperAppID:   db.DeveloperAppID,
		APIKey:           db.APIKey,
		APISecret:        db.APISecret,

		// LinkedIn-specific
		Headline:          db.Headline,
		LinkedinProfileID: db.LinkedinProfileID,
		LinkedinID:        toString(db.LinkedinID), // normalized here

		// Google My Business-specific
		MetaData:     db.MetaData,
		Locality:     db.Locality,
		PostalCode:   db.PostalCode,
		RegionCode:   db.RegionCode,
		LanguageCode: db.LanguageCode,
		LocationID:   db.LocationID,

		// Pinterest-specific
		ProfileID:   db.ProfileID,
		BoardID:     db.BoardID,
		LinkedTo:    db.LinkedTo,
		URL:         db.URL,
		PinterestID: db.PinterestID,

		// Scheduling
		QueueSlots: db.QueueSlots,

		// Validity
		InvalidTries:     db.InvalidTries,
		SentInvalidEmail: db.SentInvalidEmail,
		ValidityError:    db.ValidityError,
		ValidityStatus:   db.ValidityStatus,
		LimitExceedTries: db.LimitExceedTries,

		// External connections
		ConnectionViaLink: db.ConnectionViaLink,
		ConnectionLinkID:  db.ConnectionLinkID,

		// Preferences
		Preferences: db.Preferences,
		ExtraData:   db.ExtraData,

		// Legacy fields
		FacebookID:  db.FacebookID,
		InstagramID: toString(db.InstagramID),
		TwitterID:   db.TwitterID,
	}
}

func parseMongoDateToString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case primitive.DateTime:
		return t.Time().Format(time.RFC3339) // you can change layout if needed.
	case time.Time:
		return t.Format(time.RFC3339)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", t) // fallback (rare case)
	}
}
