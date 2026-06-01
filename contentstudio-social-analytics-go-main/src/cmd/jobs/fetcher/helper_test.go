package fetcher

import (
	"testing"
)

func Test_GetStringFromExtraData_Table(t *testing.T) {
	cases := []struct {
		name      string
		extraData map[string]interface{}
		key       string
		expected  string
	}{
		{
			name:      "nil map returns empty string",
			extraData: nil,
			key:       "test",
			expected:  "",
		},
		{
			name:      "empty map returns empty string",
			extraData: map[string]interface{}{},
			key:       "test",
			expected:  "",
		},
		{
			name: "key not found returns empty string",
			extraData: map[string]interface{}{
				"other_key": "value",
			},
			key:      "test",
			expected: "",
		},
		{
			name: "key found with string value",
			extraData: map[string]interface{}{
				"test": "my_value",
			},
			key:      "test",
			expected: "my_value",
		},
		{
			name: "key found with non-string value returns empty",
			extraData: map[string]interface{}{
				"test": 123,
			},
			key:      "test",
			expected: "",
		},
		{
			name: "key found with nil value returns empty",
			extraData: map[string]interface{}{
				"test": nil,
			},
			key:      "test",
			expected: "",
		},
		{
			name: "key found with bool value returns empty",
			extraData: map[string]interface{}{
				"test": true,
			},
			key:      "test",
			expected: "",
		},
		{
			name: "key found with float value returns empty",
			extraData: map[string]interface{}{
				"test": 3.14,
			},
			key:      "test",
			expected: "",
		},
		{
			name: "empty string key",
			extraData: map[string]interface{}{
				"": "value",
			},
			key:      "",
			expected: "value",
		},
		{
			name: "empty string value",
			extraData: map[string]interface{}{
				"test": "",
			},
			key:      "test",
			expected: "",
		},
		{
			name: "multiple keys with target found",
			extraData: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			key:      "key2",
			expected: "value2",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := GetStringFromExtraData(tc.extraData, tc.key)
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func Test_GetBoolFromExtraData_Table(t *testing.T) {
	cases := []struct {
		name      string
		extraData map[string]interface{}
		key       string
		expected  bool
	}{
		{
			name:      "nil map returns false",
			extraData: nil,
			key:       "test",
			expected:  false,
		},
		{
			name:      "empty map returns false",
			extraData: map[string]interface{}{},
			key:       "test",
			expected:  false,
		},
		{
			name: "key not found returns false",
			extraData: map[string]interface{}{
				"other_key": true,
			},
			key:      "test",
			expected: false,
		},
		{
			name: "key found with true value",
			extraData: map[string]interface{}{
				"test": true,
			},
			key:      "test",
			expected: true,
		},
		{
			name: "key found with false value",
			extraData: map[string]interface{}{
				"test": false,
			},
			key:      "test",
			expected: false,
		},
		{
			name: "key found with non-bool value returns false",
			extraData: map[string]interface{}{
				"test": "true",
			},
			key:      "test",
			expected: false,
		},
		{
			name: "key found with int value returns false",
			extraData: map[string]interface{}{
				"test": 1,
			},
			key:      "test",
			expected: false,
		},
		{
			name: "key found with nil value returns false",
			extraData: map[string]interface{}{
				"test": nil,
			},
			key:      "test",
			expected: false,
		},
		{
			name: "multiple keys with target true",
			extraData: map[string]interface{}{
				"key1": false,
				"key2": true,
				"key3": false,
			},
			key:      "key2",
			expected: true,
		},
		{
			name: "empty string key with bool value",
			extraData: map[string]interface{}{
				"": true,
			},
			key:      "",
			expected: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := GetBoolFromExtraData(tc.extraData, tc.key)
			if result != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
