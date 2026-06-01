package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSocialIntegration_GetPlatformID(t *testing.T) {
	cases := []struct {
		name     string
		account  SocialIntegration
		expected string
	}{
		{
			name: "uses platform_identifier when set",
			account: SocialIntegration{
				PlatformType:       PlatformFacebook,
				PlatformIdentifier: "123456789",
				FacebookID:         "legacy_id",
			},
			expected: "123456789",
		},
		{
			name: "fallback to FacebookID",
			account: SocialIntegration{
				PlatformType:       PlatformFacebook,
				PlatformIdentifier: "",
				FacebookID:         "fb_legacy_id",
			},
			expected: "fb_legacy_id",
		},
		{
			name: "fallback to InstagramID",
			account: SocialIntegration{
				PlatformType:       PlatformInstagram,
				PlatformIdentifier: "",
				InstagramID:        "ig_legacy_id",
			},
			expected: "ig_legacy_id",
		},
		{
			name: "fallback to LinkedinID",
			account: SocialIntegration{
				PlatformType:       PlatformLinkedIn,
				PlatformIdentifier: "",
				LinkedinID:         "li_legacy_id",
			},
			expected: "li_legacy_id",
		},
		{
			name: "fallback to TwitterID",
			account: SocialIntegration{
				PlatformType:       PlatformTwitter,
				PlatformIdentifier: "",
				TwitterID:          "tw_legacy_id",
			},
			expected: "tw_legacy_id",
		},
		{
			name: "fallback to LocationID for GMB",
			account: SocialIntegration{
				PlatformType:       PlatformGMB,
				PlatformIdentifier: "",
				LocationID:         "gmb_location",
			},
			expected: "gmb_location",
		},
		{
			name: "fallback to PinterestID",
			account: SocialIntegration{
				PlatformType:       PlatformPinterest,
				PlatformIdentifier: "",
				PinterestID:        "pin_id",
			},
			expected: "pin_id",
		},
		{
			name: "unknown platform returns empty",
			account: SocialIntegration{
				PlatformType:       "unknown",
				PlatformIdentifier: "",
			},
			expected: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := tc.account.GetPlatformID()
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestSocialIntegration_GetAccessToken(t *testing.T) {
	cases := []struct {
		name     string
		account  SocialIntegration
		expected string
	}{
		{
			name: "Facebook prefers long_access_token",
			account: SocialIntegration{
				PlatformType:    PlatformFacebook,
				AccessToken:     "short_token",
				LongAccessToken: "long_token",
			},
			expected: "long_token",
		},
		{
			name: "Facebook falls back to access_token",
			account: SocialIntegration{
				PlatformType:    PlatformFacebook,
				AccessToken:     "short_token",
				LongAccessToken: "",
			},
			expected: "short_token",
		},
		{
			name: "Instagram from user_details",
			account: SocialIntegration{
				PlatformType: PlatformInstagram,
				AccessToken:  "default_token",
				UserDetails: map[string]interface{}{
					"access_token": "ig_token",
				},
			},
			expected: "ig_token",
		},
		{
			name: "Instagram fallback to access_token",
			account: SocialIntegration{
				PlatformType: PlatformInstagram,
				AccessToken:  "default_token",
				UserDetails:  nil,
			},
			expected: "default_token",
		},
		{
			name: "Twitter uses OAuth token",
			account: SocialIntegration{
				PlatformType: PlatformTwitter,
				AccessToken:  "default_token",
				OAuthToken:   "oauth_token",
			},
			expected: "oauth_token",
		},
		{
			name: "Twitter fallback to access_token",
			account: SocialIntegration{
				PlatformType: PlatformTwitter,
				AccessToken:  "default_token",
				OAuthToken:   "",
			},
			expected: "default_token",
		},
		{
			name: "Other platforms use access_token",
			account: SocialIntegration{
				PlatformType: PlatformLinkedIn,
				AccessToken:  "linkedin_token",
			},
			expected: "linkedin_token",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := tc.account.GetAccessToken()
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestSocialIntegration_IsValid(t *testing.T) {
	cases := []struct {
		name     string
		account  SocialIntegration
		expected bool
	}{
		{
			name: "valid and added",
			account: SocialIntegration{
				Validity: ValidityValid,
				State:    StateAdded,
			},
			expected: true,
		},
		{
			name: "valid and syncing",
			account: SocialIntegration{
				Validity: ValidityValid,
				State:    StateSyncing,
			},
			expected: true,
		},
		{
			name: "valid and processed",
			account: SocialIntegration{
				Validity: ValidityValid,
				State:    StateProcessed,
			},
			expected: true,
		},
		{
			name: "invalid validity",
			account: SocialIntegration{
				Validity: ValidityInvalid,
				State:    StateAdded,
			},
			expected: false,
		},
		{
			name: "expired validity",
			account: SocialIntegration{
				Validity: ValidityExpired,
				State:    StateAdded,
			},
			expected: false,
		},
		{
			name: "valid but deleted state",
			account: SocialIntegration{
				Validity: ValidityValid,
				State:    StateDeleted,
			},
			expected: false,
		},
		{
			name: "valid but disabled state",
			account: SocialIntegration{
				Validity: ValidityValid,
				State:    StateDisabled,
			},
			expected: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := tc.account.IsValid()
			if result != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestSocialIntegration_GetDisplayName(t *testing.T) {
	cases := []struct {
		name     string
		account  SocialIntegration
		expected string
	}{
		{
			name: "uses platform_name when set",
			account: SocialIntegration{
				PlatformName:       "My Page",
				Username:           "myusername",
				PlatformIdentifier: "123456",
			},
			expected: "My Page",
		},
		{
			name: "fallback to username",
			account: SocialIntegration{
				PlatformName:       "",
				Username:           "myusername",
				PlatformIdentifier: "123456",
			},
			expected: "myusername",
		},
		{
			name: "fallback to screen_name",
			account: SocialIntegration{
				PlatformName:       "",
				Username:           "",
				ScreenName:         "myscreenname",
				PlatformIdentifier: "123456",
			},
			expected: "myscreenname",
		},
		{
			name: "fallback to platform_identifier",
			account: SocialIntegration{
				PlatformName:       "",
				Username:           "",
				ScreenName:         "",
				PlatformIdentifier: "final_fallback",
			},
			expected: "final_fallback",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := tc.account.GetDisplayName()
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestToString(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"int", 42, "42"},
		{"int32", int32(100), "100"},
		{"int64", int64(1000), "1000"},
		{"float32", float32(3.14), "3"},
		{"float64", float64(2.718), "3"},
		{"string", "hello", "hello"},
		{"nil", nil, ""},
		{"bool", true, ""},
		{"slice", []int{1, 2, 3}, ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := ToString(tc.input)
			if result != tc.expected {
				t.Fatalf("ToString(%v) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestConvertDBToSocialIntegration(t *testing.T) {
	objectID := primitive.NewObjectID()
	workspaceID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	now := &MongoTime{Time: time.Now()}

	db := DBSocialIntegration{
		ID:                 objectID,
		PlatformType:       PlatformFacebook,
		PlatformIdentifier: "123456789",
		PlatformName:       "Test Page",
		Type:               TypePage,
		WorkspaceID:        workspaceID,
		UserID:             userID,
		State:              StateAdded,
		Validity:           ValidityValid,
		CreatedAt:          now,
		AccessToken:        "test_token",
		LongAccessToken:    "long_test_token",
		FanCount:           1000,
		IsBusiness:         true,
		Username:           "testuser",
	}

	result := ConvertDBToSocialIntegration(db)

	if result.ID != objectID {
		t.Fatal("ID mismatch")
	}
	if result.PlatformType != PlatformFacebook {
		t.Fatal("PlatformType mismatch")
	}
	if result.PlatformIdentifier != "123456789" {
		t.Fatalf("PlatformIdentifier mismatch: got %s", result.PlatformIdentifier)
	}
	if result.PlatformName != "Test Page" {
		t.Fatal("PlatformName mismatch")
	}
	if result.State != StateAdded {
		t.Fatal("State mismatch")
	}
	if result.AccessToken != "test_token" {
		t.Fatal("AccessToken mismatch")
	}
	if result.FanCount != 1000 {
		t.Fatal("FanCount mismatch")
	}
}

func TestConvertDBToSocialIntegration_InterfaceTypes(t *testing.T) {
	db := DBSocialIntegration{
		PlatformIdentifier: int64(123456),
		InstagramID:        int64(789012),
		LinkedinID:         "linkedin_str_id",
	}

	result := ConvertDBToSocialIntegration(db)

	if result.PlatformIdentifier != "123456" {
		t.Fatalf("expected PlatformIdentifier '123456', got %q", result.PlatformIdentifier)
	}
	if result.InstagramID != "789012" {
		t.Fatalf("expected InstagramID '789012', got %q", result.InstagramID)
	}
	if result.LinkedinID != "linkedin_str_id" {
		t.Fatalf("expected LinkedinID 'linkedin_str_id', got %q", result.LinkedinID)
	}
}

func TestConvertDBToSocialIntegration_EmbeddedAccessToken(t *testing.T) {
	cases := []struct {
		name                 string
		accessToken          interface{}
		refreshToken         interface{}
		expectedAccessToken  string
		expectedRefreshToken string
	}{
		{
			name:                 "String access token",
			accessToken:          "plain_access_token",
			refreshToken:         "plain_refresh_token",
			expectedAccessToken:  "plain_access_token",
			expectedRefreshToken: "plain_refresh_token",
		},
		{
			name: "YouTube embedded document - map[string]interface{} format",
			accessToken: map[string]interface{}{
				"access_token":  "embedded_access_token",
				"refresh_token": "embedded_refresh_token",
				"expires_in":    3599,
				"token_type":    "Bearer",
			},
			refreshToken:         nil,
			expectedAccessToken:  "embedded_access_token",
			expectedRefreshToken: "embedded_refresh_token",
		},
		{
			name: "YouTube embedded document with token key",
			accessToken: map[string]interface{}{
				"token":         "token_key_access",
				"refresh_token": "token_key_refresh",
			},
			refreshToken:         nil,
			expectedAccessToken:  "token_key_access",
			expectedRefreshToken: "token_key_refresh",
		},
		{
			name: "Embedded document - primitive.M format",
			accessToken: primitive.M{
				"access_token":  "primitive_m_access",
				"refresh_token": "primitive_m_refresh",
			},
			refreshToken:         nil,
			expectedAccessToken:  "primitive_m_access",
			expectedRefreshToken: "primitive_m_refresh",
		},
		{
			name: "Embedded document - primitive.D format",
			accessToken: primitive.D{
				{Key: "access_token", Value: "primitive_d_access"},
				{Key: "refresh_token", Value: "primitive_d_refresh"},
			},
			refreshToken:         nil,
			expectedAccessToken:  "primitive_d_access",
			expectedRefreshToken: "primitive_d_refresh",
		},
		{
			name:                 "Nil tokens",
			accessToken:          nil,
			refreshToken:         nil,
			expectedAccessToken:  "",
			expectedRefreshToken: "",
		},
		{
			name:                 "Empty string tokens",
			accessToken:          "",
			refreshToken:         "",
			expectedAccessToken:  "",
			expectedRefreshToken: "",
		},
		{
			name: "Refresh token as separate embedded document",
			accessToken: map[string]interface{}{
				"access_token": "access_only",
			},
			refreshToken: map[string]interface{}{
				"token": "separate_refresh",
			},
			expectedAccessToken:  "access_only",
			expectedRefreshToken: "separate_refresh",
		},
		{
			name: "Refresh token from embedded with refresh_token key",
			accessToken: map[string]interface{}{
				"access_token": "access_only",
			},
			refreshToken: map[string]interface{}{
				"refresh_token": "separate_refresh_key",
			},
			expectedAccessToken:  "access_only",
			expectedRefreshToken: "separate_refresh_key",
		},
		{
			name: "Mixed: string access token, embedded refresh token",
			accessToken:  "string_access",
			refreshToken: map[string]interface{}{
				"token": "embedded_refresh",
			},
			expectedAccessToken:  "string_access",
			expectedRefreshToken: "embedded_refresh",
		},
		{
			name:                 "Refresh token precedence from access_token embedded",
			accessToken: map[string]interface{}{
				"access_token":  "embedded_access",
				"refresh_token": "embedded_refresh_from_access_doc",
			},
			refreshToken:         "plain_refresh_ignored",
			expectedAccessToken:  "embedded_access",
			expectedRefreshToken: "embedded_refresh_from_access_doc",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db := DBSocialIntegration{
				ID:                 primitive.NewObjectID(),
				PlatformType:       PlatformYouTube,
				PlatformIdentifier: "youtube_channel_123",
				AccessToken:        tc.accessToken,
				RefreshToken:       tc.refreshToken,
			}

			result := ConvertDBToSocialIntegration(db)

			if result.AccessToken != tc.expectedAccessToken {
				t.Errorf("expected AccessToken %q, got %q", tc.expectedAccessToken, result.AccessToken)
			}
			if result.RefreshToken != tc.expectedRefreshToken {
				t.Errorf("expected RefreshToken %q, got %q", tc.expectedRefreshToken, result.RefreshToken)
			}
		})
	}
}

func TestParseMongoDateToString(t *testing.T) {
	testTime := time.Date(2025, 6, 9, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name        string
		input       interface{}
		shouldParse bool
	}{
		{
			name:        "string input",
			input:       "2025-06-09T12:00:00Z",
			shouldParse: true,
		},
		{
			name:        "nil input",
			input:       nil,
			shouldParse: false,
		},
		{
			name:        "time.Time input",
			input:       testTime,
			shouldParse: true,
		},
		{
			name:        "primitive.DateTime input",
			input:       primitive.NewDateTimeFromTime(testTime),
			shouldParse: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parseMongoDateToString(tc.input)
			if tc.shouldParse {
				if result == "" {
					t.Fatalf("expected non-empty result for %v", tc.input)
				}
			} else {
				if result != "" {
					t.Fatalf("expected empty result for nil, got %q", result)
				}
			}
		})
	}
}

func TestLinkedToInfo(t *testing.T) {
	info := LinkedToInfo{Zapier: true}
	if !info.Zapier {
		t.Fatal("expected Zapier to be true")
	}

	info2 := LinkedToInfo{Zapier: false}
	if info2.Zapier {
		t.Fatal("expected Zapier to be false")
	}
}

func TestPermissions(t *testing.T) {
	perm := Permissions{Permission: "read"}
	if perm.Permission != "read" {
		t.Fatalf("expected Permission 'read', got %s", perm.Permission)
	}
}

func TestPlatformConstants(t *testing.T) {
	platforms := []string{
		PlatformFacebook,
		PlatformInstagram,
		PlatformLinkedIn,
		PlatformTwitter,
		PlatformGMB,
		PlatformPinterest,
		PlatformYouTube,
		PlatformTikTok,
	}

	expected := []string{
		"facebook",
		"instagram",
		"linkedin",
		"twitter",
		"gmb",
		"pinterest",
		"youtube",
		"tiktok",
	}

	for i, p := range platforms {
		if p != expected[i] {
			t.Fatalf("platform constant mismatch: expected %q, got %q", expected[i], p)
		}
	}
}

func TestStateConstants(t *testing.T) {
	states := []string{
		StateAdded,
		StateDeleted,
		StateDisabled,
		StatePaused,
		StateSyncing,
		StateProcessed,
		StateNotFound,
	}

	if len(states) != 7 {
		t.Fatalf("expected 7 state constants, got %d", len(states))
	}
}

func TestValidityConstants(t *testing.T) {
	validities := []string{
		ValidityValid,
		ValidityInvalid,
		ValidityExpired,
		ValidityExpiringSoon,
	}

	expected := []string{"valid", "invalid", "expired", "expiring_soon"}
	for i, v := range validities {
		if v != expected[i] {
			t.Fatalf("validity constant mismatch: expected %q, got %q", expected[i], v)
		}
	}
}

func TestSuperAdminStateConstants(t *testing.T) {
	states := []string{
		SuperAdminStateActive,
		SuperAdminStatePastDue,
	}

	expected := []string{"active", "past_due"}
	for i, s := range states {
		if s != expected[i] {
			t.Fatalf("super admin state constant mismatch: expected %q, got %q", expected[i], s)
		}
	}
}

func TestTypeConstants(t *testing.T) {
	types := []string{
		TypePage,
		TypeProfile,
		TypeBusiness,
		TypeGroup,
		TypeBoard,
		TypeCreator,
	}

	expected := []string{"Page", "Profile", "Business", "Group", "Board", "Creator"}
	for i, typ := range types {
		if typ != expected[i] {
			t.Fatalf("type constant mismatch: expected %q, got %q", expected[i], typ)
		}
	}
}
