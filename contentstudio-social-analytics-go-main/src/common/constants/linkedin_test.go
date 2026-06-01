package constants

import (
	"testing"
)

func Test_GetSeniorities_Table(t *testing.T) {
	cases := []struct {
		name          string
		checkKey      string
		expectValue   string
		expectExists  bool
		checkNotEmpty bool
	}{
		{
			name:          "returns non-empty map",
			checkNotEmpty: true,
		},
		{
			name:         "contains Training seniority",
			checkKey:     "urn:li:seniority:1",
			expectValue:  "Training",
			expectExists: true,
		},
		{
			name:         "contains Entry seniority",
			checkKey:     "urn:li:seniority:2",
			expectValue:  "Entry",
			expectExists: true,
		},
		{
			name:         "contains Executive seniority",
			checkKey:     "urn:li:seniority:6",
			expectValue:  "Executive",
			expectExists: true,
		},
		{
			name:         "non-existent key returns empty",
			checkKey:     "urn:li:seniority:999",
			expectExists: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := GetSeniorities()

			if tc.checkNotEmpty {
				if len(result) == 0 {
					t.Fatal("expected non-empty map")
				}
				return
			}

			val, exists := result[tc.checkKey]
			if tc.expectExists {
				if !exists {
					t.Fatalf("expected key %s to exist", tc.checkKey)
				}
				if val != tc.expectValue {
					t.Fatalf("expected value %s, got %s", tc.expectValue, val)
				}
			} else {
				if exists {
					t.Fatalf("expected key %s to not exist", tc.checkKey)
				}
			}
		})
	}
}

func Test_GetSeniorities_ReturnsCopy(t *testing.T) {
	result1 := GetSeniorities()
	result2 := GetSeniorities()

	result1["urn:li:seniority:1"] = "Modified"

	if result2["urn:li:seniority:1"] == "Modified" {
		t.Fatal("GetSeniorities should return a copy, not a reference")
	}
}

func Test_GetIndustries_Table(t *testing.T) {
	cases := []struct {
		name          string
		checkKey      string
		expectValue   string
		expectExists  bool
		checkNotEmpty bool
	}{
		{
			name:          "returns non-empty map",
			checkNotEmpty: true,
		},
		{
			name:         "contains Defense & Space",
			checkKey:     "urn:li:industry:1",
			expectValue:  "Defense & Space",
			expectExists: true,
		},
		{
			name:         "contains Computer Software",
			checkKey:     "urn:li:industry:4",
			expectValue:  "Computer Software",
			expectExists: true,
		},
		{
			name:         "contains Internet",
			checkKey:     "urn:li:industry:6",
			expectValue:  "Internet",
			expectExists: true,
		},
		{
			name:         "contains IT Services",
			checkKey:     "urn:li:industry:96",
			expectValue:  "Information Technology & Services",
			expectExists: true,
		},
		{
			name:         "non-existent key returns empty",
			checkKey:     "urn:li:industry:9999",
			expectExists: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := GetIndustries()

			if tc.checkNotEmpty {
				if len(result) == 0 {
					t.Fatal("expected non-empty map")
				}
				return
			}

			val, exists := result[tc.checkKey]
			if tc.expectExists {
				if !exists {
					t.Fatalf("expected key %s to exist", tc.checkKey)
				}
				if val != tc.expectValue {
					t.Fatalf("expected value %s, got %s", tc.expectValue, val)
				}
			} else {
				if exists {
					t.Fatalf("expected key %s to not exist", tc.checkKey)
				}
			}
		})
	}
}

func Test_GetIndustries_ReturnsCopy(t *testing.T) {
	result1 := GetIndustries()
	result2 := GetIndustries()

	result1["urn:li:industry:1"] = "Modified"

	if result2["urn:li:industry:1"] == "Modified" {
		t.Fatal("GetIndustries should return a copy, not a reference")
	}
}

func Test_GetCountries_Table(t *testing.T) {
	cases := []struct {
		name          string
		checkKey      string
		expectValue   string
		expectExists  bool
		checkNotEmpty bool
	}{
		{
			name:          "returns non-empty map",
			checkNotEmpty: true,
		},
		{
			name:         "contains United States",
			checkKey:     "urn:li:country:us",
			expectValue:  "United States",
			expectExists: true,
		},
		{
			name:         "contains United Kingdom",
			checkKey:     "urn:li:country:gb",
			expectValue:  "United Kingdom",
			expectExists: true,
		},
		{
			name:         "contains Germany",
			checkKey:     "urn:li:country:de",
			expectValue:  "Germany",
			expectExists: true,
		},
		{
			name:         "contains Pakistan",
			checkKey:     "urn:li:country:pk",
			expectValue:  "Pakistan",
			expectExists: true,
		},
		{
			name:         "non-existent key returns empty",
			checkKey:     "urn:li:country:xx",
			expectExists: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := GetCountries()

			if tc.checkNotEmpty {
				if len(result) == 0 {
					t.Fatal("expected non-empty map")
				}
				return
			}

			val, exists := result[tc.checkKey]
			if tc.expectExists {
				if !exists {
					t.Fatalf("expected key %s to exist", tc.checkKey)
				}
				if val != tc.expectValue {
					t.Fatalf("expected value %s, got %s", tc.expectValue, val)
				}
			} else {
				if exists {
					t.Fatalf("expected key %s to not exist", tc.checkKey)
				}
			}
		})
	}
}

func Test_GetCountries_ReturnsCopy(t *testing.T) {
	result1 := GetCountries()
	result2 := GetCountries()

	result1["urn:li:country:us"] = "Modified"

	if result2["urn:li:country:us"] == "Modified" {
		t.Fatal("GetCountries should return a copy, not a reference")
	}
}

func Test_GetFunctions_Table(t *testing.T) {
	cases := []struct {
		name          string
		checkKey      string
		expectValue   string
		expectExists  bool
		checkNotEmpty bool
	}{
		{
			name:          "returns non-empty map",
			checkNotEmpty: true,
		},
		{
			name:         "contains Accounting",
			checkKey:     "urn:li:function:1",
			expectValue:  "Accounting",
			expectExists: true,
		},
		{
			name:         "contains Engineering",
			checkKey:     "urn:li:function:8",
			expectValue:  "Engineering",
			expectExists: true,
		},
		{
			name:         "contains Information Technology",
			checkKey:     "urn:li:function:13",
			expectValue:  "Information Technology",
			expectExists: true,
		},
		{
			name:         "contains Sales",
			checkKey:     "urn:li:function:25",
			expectValue:  "Sales",
			expectExists: true,
		},
		{
			name:         "non-existent key returns empty",
			checkKey:     "urn:li:function:999",
			expectExists: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := GetFunctions()

			if tc.checkNotEmpty {
				if len(result) == 0 {
					t.Fatal("expected non-empty map")
				}
				return
			}

			val, exists := result[tc.checkKey]
			if tc.expectExists {
				if !exists {
					t.Fatalf("expected key %s to exist", tc.checkKey)
				}
				if val != tc.expectValue {
					t.Fatalf("expected value %s, got %s", tc.expectValue, val)
				}
			} else {
				if exists {
					t.Fatalf("expected key %s to not exist", tc.checkKey)
				}
			}
		})
	}
}

func Test_GetFunctions_ReturnsCopy(t *testing.T) {
	result1 := GetFunctions()
	result2 := GetFunctions()

	result1["urn:li:function:1"] = "Modified"

	if result2["urn:li:function:1"] == "Modified" {
		t.Fatal("GetFunctions should return a copy, not a reference")
	}
}
