package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// EncryptedPayload defines the structure of the JSON payload for encrypted tokens.
type EncryptedPayload struct {
	IV    string `json:"iv"`
	Value string `json:"value"`
}

// pkcs7Unpad removes PKCS7 padding from a byte slice.
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("pkcs7Unpad: input data is empty")
	}
	paddingLen := int(data[len(data)-1])
	if paddingLen > len(data) || paddingLen == 0 {
		return nil, errors.New("pkcs7Unpad: invalid padding length")
	}
	for i := 0; i < paddingLen; i++ {
		if data[len(data)-paddingLen+i] != byte(paddingLen) {
			return nil, errors.New("pkcs7Unpad: invalid padding bytes")
		}
	}
	return data[:len(data)-paddingLen], nil
}

// DecryptToken decrypts an AES-256-CBC encrypted token.
// The encryptedToken is expected to be a base64 encoded JSON string containing
// an "iv" field and a "value" field, both base64 encoded.
// The base64EncodedDecryptionKey must be a base64 encoded 32-byte AES key.
func DecryptToken(encryptedToken string, base64EncodedDecryptionKey string) (string, error) {
	// Validate inputs

	if encryptedToken == "" {
		return "", errors.New("encrypted token is empty")
	}
	if base64EncodedDecryptionKey == "" {
		return "", errors.New("decryption key is empty")
	}

	// First, check if the token is already plaintext (for backward compatibility)
	// If it doesn't look like base64, it might be a plaintext token
	if !isLikelyBase64(encryptedToken) {
		// Log this case but return the token as-is
		return encryptedToken, nil
	}

	// Decode the encrypted payload string (which itself is base64 encoded JSON)
	decodedPayloadBytes, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		// Try URL-safe base64 decoding as fallback
		decodedPayloadBytes, err = base64.URLEncoding.DecodeString(encryptedToken)
		if err != nil {
			return "", fmt.Errorf("DecryptToken: failed to base64 decode encrypted token: %w", err)
		}
	}
	// Check if the decoded payload is valid UTF-8 and looks like JSON
	if !isValidJSON(decodedPayloadBytes) {
		return "", fmt.Errorf("DecryptToken: decoded payload is not valid JSON")
	}

	var payload EncryptedPayload
	err = json.Unmarshal(decodedPayloadBytes, &payload)
	if err != nil {
		return "", fmt.Errorf("DecryptToken: failed to unmarshal JSON payload: %w", err)
	}

	// Validate payload structure
	if payload.IV == "" || payload.Value == "" {
		return "", errors.New("encrypted payload missing required fields (iv or value)")
	}

	iv, err := base64.StdEncoding.DecodeString(payload.IV)
	if err != nil {
		return "", fmt.Errorf("DecryptToken: failed to base64 decode IV: %w", err)
	}

	encryptedValue, err := base64.StdEncoding.DecodeString(payload.Value)
	if err != nil {
		return "", fmt.Errorf("DecryptToken: failed to base64 decode encrypted value: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(base64EncodedDecryptionKey)

	if err != nil {
		return "", fmt.Errorf("DecryptToken: failed to base64 decode decryption key: %w", err)
	}

	if len(key) != 32 { // AES-256 requires a 32-byte key
		return "", errors.New("decryption key must be 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("DecryptToken: failed to create AES cipher: %w", err)
	}

	if len(encryptedValue)%aes.BlockSize != 0 {
		return "", errors.New("encrypted value is not a multiple of the block size")
	}
	if len(iv) != aes.BlockSize {
		return "", errors.New("IV length must equal AES block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decryptedPadded := make([]byte, len(encryptedValue))
	mode.CryptBlocks(decryptedPadded, encryptedValue)

	decrypted, err := pkcs7Unpad(decryptedPadded)
	if err != nil {
		// If unpadding fails, it might indicate a bad key or corrupted data.
		// The Python version returns the original encrypted_token in case of error.
		// We will return an error here for clarity, but you can adapt this.
		return "", fmt.Errorf("DecryptToken: failed to unpad decrypted data: %w. This could be due to an incorrect decryption key or corrupted data.", err)
	}

	return string(decrypted), nil
}

// isLikelyBase64 checks if a string looks like base64 encoded data
func isLikelyBase64(s string) bool {
	// Basic checks for base64 format
	if len(s) == 0 {
		return false
	}

	// Base64 strings should be properly padded or have a length that's a multiple of 4
	if len(s)%4 != 0 && !strings.HasSuffix(s, "=") && !strings.HasSuffix(s, "==") {
		return false
	}

	// Check if the string contains only valid base64 characters
	for _, char := range s {
		if !((char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '+' || char == '/' || char == '=' || char == '-' || char == '_') {
			return false
		}
	}

	return true
}

// isValidJSON checks if the given bytes represent valid JSON
func isValidJSON(data []byte) bool {
	var js interface{}
	return json.Unmarshal(data, &js) == nil
}

// GenerateAppsecretProof creates an appsecret_proof for Facebook API calls.
// It takes an access token and an app secret, and returns the HMAC-SHA256 hash of the access token
// using the app secret as the key, hex-encoded.
func GenerateAppsecretProof(accessToken string, appSecret string) string {
	h := hmac.New(sha256.New, []byte(appSecret))
	h.Write([]byte(accessToken))
	return hex.EncodeToString(h.Sum(nil))
}
