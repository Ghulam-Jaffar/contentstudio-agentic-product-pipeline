package main

import (
	"testing"
)

func TestPlatformScale(t *testing.T) {
	expectedPlatforms := []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest", "gmb"}

	for _, platform := range expectedPlatforms {
		if _, ok := PlatformScale[platform]; !ok {
			t.Errorf("expected platform '%s' in PlatformScale", platform)
		}
	}

	if PlatformScale["facebook"] != 24000 {
		t.Errorf("expected facebook scale 24000, got %d", PlatformScale["facebook"])
	}
	if PlatformScale["instagram"] != 16000 {
		t.Errorf("expected instagram scale 16000, got %d", PlatformScale["instagram"])
	}
	if PlatformScale["linkedin"] != 8000 {
		t.Errorf("expected linkedin scale 8000, got %d", PlatformScale["linkedin"])
	}
	if PlatformScale["tiktok"] != 2000 {
		t.Errorf("expected tiktok scale 2000, got %d", PlatformScale["tiktok"])
	}
	if PlatformScale["twitter"] != 1000 {
		t.Errorf("expected twitter scale 1000, got %d", PlatformScale["twitter"])
	}
	if PlatformScale["pinterest"] != 1000 {
		t.Errorf("expected pinterest scale 1000, got %d", PlatformScale["pinterest"])
	}
	if PlatformScale["gmb"] != 500 {
		t.Errorf("expected gmb scale 500, got %d", PlatformScale["gmb"])
	}
}

func TestPlatformScale_TotalCount(t *testing.T) {
	total := 0
	for _, count := range PlatformScale {
		total += count
	}
	// Total should be approximately 57500 (24000+16000+8000+4000+2000+1000+1000+500+1000)
	if total != 57500 {
		t.Errorf("expected total scale ~57500, got %d", total)
	}
}

// ================== determinePlatforms Tests ==================

func TestDeterminePlatforms_EmptyString(t *testing.T) {
	result := determinePlatforms("")
	expected := []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest", "gmb", "meta_ads"}

	if len(result) != len(expected) {
		t.Fatalf("determinePlatforms(\"\") returned %d items, want %d", len(result), len(expected))
	}

	for i, p := range expected {
		if result[i] != p {
			t.Errorf("determinePlatforms(\"\")[%d] = %q, want %q", i, result[i], p)
		}
	}
}

func TestDeterminePlatforms_SinglePlatform(t *testing.T) {
	result := determinePlatforms("facebook")

	if len(result) != 1 {
		t.Fatalf("expected 1 platform, got %d", len(result))
	}
	if result[0] != "facebook" {
		t.Errorf("result[0] = %q, want %q", result[0], "facebook")
	}
}

func TestDeterminePlatforms_MultiplePlatforms(t *testing.T) {
	result := determinePlatforms("facebook,instagram,linkedin,tiktok")

	if len(result) != 4 {
		t.Fatalf("expected 4 platforms, got %d", len(result))
	}

	expected := []string{"facebook", "instagram", "linkedin", "tiktok"}
	for i, p := range expected {
		if result[i] != p {
			t.Errorf("result[%d] = %q, want %q", i, result[i], p)
		}
	}
}

func TestDeterminePlatforms_WithSpaces(t *testing.T) {
	result := determinePlatforms("facebook, instagram , linkedin, tiktok")

	if len(result) != 4 {
		t.Fatalf("expected 4 platforms, got %d", len(result))
	}

	expected := []string{"facebook", "instagram", "linkedin", "tiktok"}
	for i, p := range expected {
		if result[i] != p {
			t.Errorf("result[%d] = %q, want %q", i, result[i], p)
		}
	}
}

func TestDeterminePlatforms_EmptyItems(t *testing.T) {
	result := determinePlatforms("facebook,,instagram")

	if len(result) != 2 {
		t.Fatalf("expected 2 platforms (empty items filtered), got %d", len(result))
	}
}

func TestDeterminePlatforms_OnlyCommas(t *testing.T) {
	result := determinePlatforms(",,,")

	if len(result) != 0 {
		t.Errorf("expected 0 platforms, got %d", len(result))
	}
}

func TestDeterminePlatforms_LeadingTrailingSpaces(t *testing.T) {
	result := determinePlatforms("  facebook  ")

	if len(result) != 1 {
		t.Fatalf("expected 1 platform, got %d", len(result))
	}
	if result[0] != "facebook" {
		t.Errorf("result[0] = %q, want %q", result[0], "facebook")
	}
}

func TestDeterminePlatforms_MixedEmptyAndValid(t *testing.T) {
	result := determinePlatforms(",facebook,,instagram,")

	if len(result) != 2 {
		t.Fatalf("expected 2 platforms, got %d", len(result))
	}
	if result[0] != "facebook" || result[1] != "instagram" {
		t.Errorf("result = %v, want [facebook, instagram]", result)
	}
}

// ================== parseAccountTypes Tests ==================

func TestParseAccountTypes_EmptyString(t *testing.T) {
	result := parseAccountTypes("")

	if result != nil {
		t.Errorf("parseAccountTypes(\"\") = %v, want nil", result)
	}
}

func TestParseAccountTypes_SingleType(t *testing.T) {
	result := parseAccountTypes("page")

	if len(result) != 1 {
		t.Fatalf("expected 1 type, got %d", len(result))
	}
	if result[0] != "page" {
		t.Errorf("result[0] = %q, want %q", result[0], "page")
	}
}

func TestParseAccountTypes_MultipleTypes(t *testing.T) {
	result := parseAccountTypes("page,group")

	if len(result) != 2 {
		t.Fatalf("expected 2 types, got %d", len(result))
	}

	if result[0] != "page" || result[1] != "group" {
		t.Errorf("result = %v, want [page, group]", result)
	}
}

func TestParseAccountTypes_WithSpaces(t *testing.T) {
	result := parseAccountTypes("page , group , profile , creator")

	if len(result) != 4 {
		t.Fatalf("expected 4 types, got %d", len(result))
	}

	expected := []string{"page", "group", "profile", "creator"}
	for i, p := range expected {
		if result[i] != p {
			t.Errorf("result[%d] = %q, want %q", i, result[i], p)
		}
	}
}

func TestParseAccountTypes_EmptyItems(t *testing.T) {
	result := parseAccountTypes("page,,group")

	if len(result) != 2 {
		t.Fatalf("expected 2 types (empty items filtered), got %d", len(result))
	}
}

func TestParseAccountTypes_OnlyCommas(t *testing.T) {
	result := parseAccountTypes(",,,")

	if len(result) != 0 {
		t.Errorf("expected 0 types, got %d", len(result))
	}
}

func TestParseAccountTypes_LeadingTrailingSpaces(t *testing.T) {
	result := parseAccountTypes("  page  ")

	if len(result) != 1 {
		t.Fatalf("expected 1 type, got %d", len(result))
	}
	if result[0] != "page" {
		t.Errorf("result[0] = %q, want %q", result[0], "page")
	}
}

// ================== Edge Cases ==================

func TestDeterminePlatforms_WhitespaceOnly(t *testing.T) {
	result := determinePlatforms("   ")

	// Only whitespace should be filtered out
	if len(result) != 0 {
		t.Errorf("expected 0 platforms for whitespace-only input, got %d", len(result))
	}
}

func TestParseAccountTypes_WhitespaceOnly(t *testing.T) {
	result := parseAccountTypes("   ")

	// Only whitespace should be filtered out
	if len(result) != 0 {
		t.Errorf("expected 0 types for whitespace-only input, got %d", len(result))
	}
}

func TestDeterminePlatforms_SinglePlatformWithTrailingComma(t *testing.T) {
	result := determinePlatforms("facebook,")

	if len(result) != 1 {
		t.Fatalf("expected 1 platform, got %d", len(result))
	}
	if result[0] != "facebook" {
		t.Errorf("result[0] = %q, want %q", result[0], "facebook")
	}
}

func TestParseAccountTypes_SingleTypeWithTrailingComma(t *testing.T) {
	result := parseAccountTypes("page,")

	if len(result) != 1 {
		t.Fatalf("expected 1 type, got %d", len(result))
	}
	if result[0] != "page" {
		t.Errorf("result[0] = %q, want %q", result[0], "page")
	}
}

func TestDeterminePlatforms_AllSupportedPlatforms(t *testing.T) {
	result := determinePlatforms("facebook,instagram,linkedin,youtube,tiktok")

	if len(result) != 5 {
		t.Fatalf("expected 5 platforms, got %d", len(result))
	}
}

func TestParseAccountTypes_MultipleTypesNoSpaces(t *testing.T) {
	result := parseAccountTypes("page,group,profile,user")

	if len(result) != 4 {
		t.Fatalf("expected 4 types, got %d", len(result))
	}
}

// ================== ValidatePlatform Tests ==================

func TestValidatePlatform_Facebook(t *testing.T) {
	if !ValidatePlatform("facebook") {
		t.Error("expected facebook to be valid")
	}
}

func TestValidatePlatform_Instagram(t *testing.T) {
	if !ValidatePlatform("instagram") {
		t.Error("expected instagram to be valid")
	}
}

func TestValidatePlatform_LinkedIn(t *testing.T) {
	if !ValidatePlatform("linkedin") {
		t.Error("expected linkedin to be valid")
	}
}

func TestValidatePlatform_Invalid(t *testing.T) {
	invalidPlatforms := []string{"tiktok-invalid", "twitter-x", "", "unknown"}

	for _, p := range invalidPlatforms {
		if ValidatePlatform(p) {
			t.Errorf("expected %q to be invalid", p)
		}
	}
}

func TestValidatePlatform_AllValid(t *testing.T) {
	validPlatforms := []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "pinterest"}

	for _, p := range validPlatforms {
		if !ValidatePlatform(p) {
			t.Errorf("expected %q to be valid", p)
		}
	}
}

// ================== GetDefaultPlatforms Tests ==================

func TestGetDefaultPlatforms(t *testing.T) {
	result := GetDefaultPlatforms()

	expected := []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest", "gmb", "meta_ads"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d platforms, got %d", len(expected), len(result))
	}

	for i, p := range expected {
		if result[i] != p {
			t.Errorf("result[%d] = %q, want %q", i, result[i], p)
		}
	}
}

func TestGetDefaultPlatforms_AllValid(t *testing.T) {
	platforms := GetDefaultPlatforms()

	for _, p := range platforms {
		if !ValidatePlatform(p) {
			t.Errorf("default platform %q is invalid", p)
		}
	}
}

// ================== FilterValidPlatforms Tests ==================

func TestFilterValidPlatforms_AllValid(t *testing.T) {
	input := []string{"facebook", "instagram", "linkedin", "tiktok"}
	result := FilterValidPlatforms(input)

	if len(result) != 4 {
		t.Fatalf("expected 4 platforms, got %d", len(result))
	}
}

func TestFilterValidPlatforms_MixedValidInvalid(t *testing.T) {
	input := []string{"facebook", "youtube", "instagram", "tiktok", "linkedin"}
	result := FilterValidPlatforms(input)

	expected := []string{"facebook", "youtube", "instagram", "tiktok", "linkedin"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d platforms, got %d", len(expected), len(result))
	}

	for i, p := range expected {
		if result[i] != p {
			t.Errorf("result[%d] = %q, want %q", i, result[i], p)
		}
	}
}

func TestFilterValidPlatforms_AllInvalid(t *testing.T) {
	input := []string{"snapchat", "abc", "unknown"}
	result := FilterValidPlatforms(input)

	if len(result) != 0 {
		t.Errorf("expected 0 platforms, got %d: %v", len(result), result)
	}
}

func TestFilterValidPlatforms_Empty(t *testing.T) {
	result := FilterValidPlatforms([]string{})

	if len(result) != 0 {
		t.Errorf("expected 0 platforms, got %d", len(result))
	}
}

func TestFilterValidPlatforms_Nil(t *testing.T) {
	result := FilterValidPlatforms(nil)

	if len(result) != 0 {
		t.Errorf("expected 0 platforms, got %d", len(result))
	}
}

// ================== ParseSyncType Tests ==================

func TestParseSyncType_FullSync(t *testing.T) {
	result := ParseSyncType("full_sync")

	if result != "full_sync" {
		t.Errorf("ParseSyncType(\"full_sync\") = %q, want %q", result, "full_sync")
	}
}

func TestParseSyncType_Full(t *testing.T) {
	result := ParseSyncType("full")

	if result != "full_sync" {
		t.Errorf("ParseSyncType(\"full\") = %q, want %q", result, "full_sync")
	}
}

func TestParseSyncType_Incremental(t *testing.T) {
	result := ParseSyncType("incremental")

	if result != "incremental" {
		t.Errorf("ParseSyncType(\"incremental\") = %q, want %q", result, "incremental")
	}
}

func TestParseSyncType_Empty(t *testing.T) {
	result := ParseSyncType("")

	if result != "incremental" {
		t.Errorf("ParseSyncType(\"\") = %q, want %q", result, "incremental")
	}
}

func TestParseSyncType_Unknown(t *testing.T) {
	unknownTypes := []string{"partial", "daily", "weekly", "random"}

	for _, st := range unknownTypes {
		result := ParseSyncType(st)
		if result != "incremental" {
			t.Errorf("ParseSyncType(%q) = %q, want %q", st, result, "incremental")
		}
	}
}

// ================== ProcessPlatformConfig Tests ==================

func TestProcessPlatformConfig_Struct(t *testing.T) {
	cfg := ProcessPlatformConfig{
		Platform:             "facebook",
		SyncType:             "incremental",
		FacebookAccountTypes: []string{"page", "group"},
	}

	if cfg.Platform != "facebook" {
		t.Errorf("Platform = %q, want %q", cfg.Platform, "facebook")
	}
	if cfg.SyncType != "incremental" {
		t.Errorf("SyncType = %q, want %q", cfg.SyncType, "incremental")
	}
	if len(cfg.FacebookAccountTypes) != 2 {
		t.Errorf("FacebookAccountTypes length = %d, want 2", len(cfg.FacebookAccountTypes))
	}
}

func TestProcessPlatformConfig_EmptyFields(t *testing.T) {
	cfg := ProcessPlatformConfig{}

	if cfg.Platform != "" {
		t.Errorf("expected empty Platform, got %q", cfg.Platform)
	}
	if cfg.SyncType != "" {
		t.Errorf("expected empty SyncType, got %q", cfg.SyncType)
	}
	if cfg.FacebookAccountTypes != nil {
		t.Errorf("expected nil FacebookAccountTypes, got %v", cfg.FacebookAccountTypes)
	}
}

// ================== PlatformScale Additional Tests ==================

func TestPlatformScale_AllPositive(t *testing.T) {
	for platform, scale := range PlatformScale {
		if scale <= 0 {
			t.Errorf("PlatformScale[%q] = %d, should be positive", platform, scale)
		}
	}
}

func TestPlatformScale_KeysExist(t *testing.T) {
	requiredPlatforms := []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest"}

	for _, platform := range requiredPlatforms {
		if _, ok := PlatformScale[platform]; !ok {
			t.Errorf("PlatformScale missing required platform %q", platform)
		}
	}
}

// ================== Integration Tests ==================

func TestDeterminePlatforms_AllCaseCombinations(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 9},                            // default
		{"facebook", 1},                    // single
		{"facebook,instagram", 2},          // double
		{"facebook,instagram,linkedin", 3}, // triple
		{"facebook,,instagram", 2},         // with empty
		{",facebook,", 1},                  // leading/trailing commas
		{"  facebook  ,  instagram  ", 2},  // with spaces
		{" , , facebook , , ", 1},          // mixed
	}

	for _, tc := range tests {
		result := determinePlatforms(tc.input)
		if len(result) != tc.expected {
			t.Errorf("determinePlatforms(%q) returned %d items, want %d", tc.input, len(result), tc.expected)
		}
	}
}

func TestParseAccountTypes_AllCaseCombinations(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},                   // empty returns nil (length 0)
		{"page", 1},               // single
		{"page,group", 2},         // double
		{"page,group,profile", 3}, // triple
		{"page,,group", 2},        // with empty
		{",page,", 1},             // leading/trailing commas
		{"  page  ,  group  ", 2}, // with spaces
		{" , , page , , ", 1},     // mixed
	}

	for _, tc := range tests {
		result := parseAccountTypes(tc.input)
		if tc.expected == 0 && tc.input == "" {
			if result != nil {
				t.Errorf("parseAccountTypes(%q) = %v, want nil", tc.input, result)
			}
		} else if len(result) != tc.expected {
			t.Errorf("parseAccountTypes(%q) returned %d items, want %d", tc.input, len(result), tc.expected)
		}
	}
}
