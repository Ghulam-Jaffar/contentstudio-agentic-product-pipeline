package main

import (
	"reflect"
	"testing"
)

func Test_parseAccountTypes_Table(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string returns nil",
			input:    "",
			expected: nil,
		},
		{
			name:     "single value",
			input:    "page",
			expected: []string{"page"},
		},
		{
			name:     "two values comma separated",
			input:    "profile,page",
			expected: []string{"profile", "page"},
		},
		{
			name:     "multiple values",
			input:    "profile,page,group",
			expected: []string{"profile", "page", "group"},
		},
		{
			name:     "values with spaces are trimmed",
			input:    " profile , page , group ",
			expected: []string{"profile", "page", "group"},
		},
		{
			name:     "empty values in middle are skipped",
			input:    "profile,,page",
			expected: []string{"profile", "page"},
		},
		{
			name:     "only commas returns empty slice",
			input:    ",,,",
			expected: []string{},
		},
		{
			name:     "single value with trailing comma",
			input:    "page,",
			expected: []string{"page"},
		},
		{
			name:     "single value with leading comma",
			input:    ",page",
			expected: []string{"page"},
		},
		{
			name:     "whitespace only values are skipped",
			input:    "page,   ,group",
			expected: []string{"page", "group"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parseAccountTypes(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("parseAccountTypes(%q) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}
