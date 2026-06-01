package fetcher

// getStringFromExtraData safely extracts a string value from the ExtraData map.
// Returns an empty string if the key is not found or the value is not a string.
func GetStringFromExtraData(extraData map[string]interface{}, key string) string {
	if val, ok := extraData[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

// getBoolFromExtraData safely extracts a boolean value from the ExtraData map.
// Returns false if the key is not found or the value is not a boolean.
func GetBoolFromExtraData(extraData map[string]interface{}, key string) bool {
	if val, ok := extraData[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return false
}
