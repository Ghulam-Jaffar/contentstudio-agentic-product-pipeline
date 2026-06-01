package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestPkcs7Unpad(t *testing.T) {
	cases := []struct {
		name      string
		input     []byte
		expected  []byte
		expectErr bool
	}{
		{
			name:      "empty input returns error",
			input:     []byte{},
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "valid padding of 1",
			input:     []byte("hello\x01"),
			expected:  []byte("hello"),
			expectErr: false,
		},
		{
			name:      "valid padding of 4",
			input:     []byte("test\x04\x04\x04\x04"),
			expected:  []byte("test"),
			expectErr: false,
		},
		{
			name:      "valid padding of 16 (full block)",
			input:     append([]byte("0123456789abcdef"), bytes16Padding()...),
			expected:  []byte("0123456789abcdef"),
			expectErr: false,
		},
		{
			name:      "invalid padding length zero",
			input:     []byte("hello\x00"),
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "invalid padding length exceeds data length",
			input:     []byte("hi\x10"),
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "invalid padding bytes mismatch",
			input:     []byte("test\x04\x04\x04\x03"),
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "valid single byte with padding 1",
			input:     []byte("a\x01"),
			expected:  []byte("a"),
			expectErr: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := pkcs7Unpad(tc.input)

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(result) != string(tc.expected) {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func bytes16Padding() []byte {
	b := make([]byte, 16)
	for i := range b {
		b[i] = 16
	}
	return b
}

func TestIsLikelyBase64(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty string returns false",
			input:    "",
			expected: false,
		},
		{
			name:     "valid base64 string",
			input:    "SGVsbG8gV29ybGQ=",
			expected: true,
		},
		{
			name:     "valid base64 without padding",
			input:    "SGVsbG8gV29ybGQ",
			expected: false,
		},
		{
			name:     "valid base64 with double padding",
			input:    "SGVsbG8=",
			expected: true,
		},
		{
			name:     "contains invalid characters",
			input:    "Hello World!",
			expected: false,
		},
		{
			name:     "URL-safe base64 characters",
			input:    "SGVs-bG8_V29ybGQ=",
			expected: true,
		},
		{
			name:     "only alphanumeric characters with valid length",
			input:    "abcd",
			expected: true,
		},
		{
			name:     "contains newline character",
			input:    "SGVs\nbG8=",
			expected: false,
		},
		{
			name:     "contains space character",
			input:    "SGVs bG8=",
			expected: false,
		},
		{
			name:     "long valid base64",
			input:    "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY3ODkw",
			expected: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := isLikelyBase64(tc.input)
			if result != tc.expected {
				t.Fatalf("expected %v for input %q, got %v", tc.expected, tc.input, result)
			}
		})
	}
}

func TestIsValidJSON(t *testing.T) {
	cases := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{
			name:     "valid JSON object",
			input:    []byte(`{"key": "value"}`),
			expected: true,
		},
		{
			name:     "valid JSON array",
			input:    []byte(`[1, 2, 3]`),
			expected: true,
		},
		{
			name:     "valid JSON string",
			input:    []byte(`"hello"`),
			expected: true,
		},
		{
			name:     "valid JSON number",
			input:    []byte(`123`),
			expected: true,
		},
		{
			name:     "valid JSON boolean",
			input:    []byte(`true`),
			expected: true,
		},
		{
			name:     "valid JSON null",
			input:    []byte(`null`),
			expected: true,
		},
		{
			name:     "invalid JSON - missing brace",
			input:    []byte(`{"key": "value"`),
			expected: false,
		},
		{
			name:     "invalid JSON - plain text",
			input:    []byte(`hello world`),
			expected: false,
		},
		{
			name:     "empty input",
			input:    []byte{},
			expected: false,
		},
		{
			name:     "nested JSON object",
			input:    []byte(`{"outer": {"inner": "value"}}`),
			expected: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := isValidJSON(tc.input)
			if result != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGenerateAppsecretProof(t *testing.T) {
	cases := []struct {
		name        string
		accessToken string
		appSecret   string
		expected    string
	}{
		{
			name:        "generates valid HMAC-SHA256",
			accessToken: "test_access_token",
			appSecret:   "test_app_secret",
			expected:    "e8d9f5e3d8f08fce9f9f1d4b0e8e6c4b9f8e7d6c5b4a3928f7e6d5c4b3a29182",
		},
		{
			name:        "empty access token",
			accessToken: "",
			appSecret:   "secret",
			expected:    "f9e66e179b6747ae54108f82f8ade8b3c25d76fd30afde6c395822c530196169",
		},
		{
			name:        "empty app secret",
			accessToken: "token",
			appSecret:   "",
			expected:    "17fd917f39db05a1d64c0dce3016ff109f7a56e24ca1e2a45c8f2e8a5b5d8c38",
		},
		{
			name:        "both empty",
			accessToken: "",
			appSecret:   "",
			expected:    "b613679a0814d9ec772f95d778c35fc5ff1697c493715653c6c712144292c5ad",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := GenerateAppsecretProof(tc.accessToken, tc.appSecret)

			if len(result) != 64 {
				t.Fatalf("expected 64 character hex string, got %d characters", len(result))
			}

			for _, c := range result {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Fatalf("expected lowercase hex characters, found %c", c)
				}
			}
		})
	}
}

func TestDecryptToken(t *testing.T) {
	cases := []struct {
		name           string
		encryptedToken string
		decryptionKey  string
		expected       string
		expectErr      bool
		errContains    string
	}{
		{
			name:           "empty encrypted token",
			encryptedToken: "",
			decryptionKey:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
			expected:       "",
			expectErr:      true,
			errContains:    "encrypted token is empty",
		},
		{
			name:           "empty decryption key",
			encryptedToken: "sometoken",
			decryptionKey:  "",
			expected:       "",
			expectErr:      true,
			errContains:    "decryption key is empty",
		},
		{
			name:           "plaintext token passthrough (not base64)",
			encryptedToken: "plain_token_with_special!chars",
			decryptionKey:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
			expected:       "plain_token_with_special!chars",
			expectErr:      false,
		},
		{
			name:           "invalid base64 in key",
			encryptedToken: createValidEncryptedPayload(t, "test", make([]byte, 32)),
			decryptionKey:  "not-valid-base64!!!",
			expected:       "",
			expectErr:      true,
			errContains:    "failed to base64 decode decryption key",
		},
		{
			name:           "key wrong length",
			encryptedToken: createValidEncryptedPayload(t, "test", make([]byte, 32)),
			decryptionKey:  base64.StdEncoding.EncodeToString(make([]byte, 16)),
			expected:       "",
			expectErr:      true,
			errContains:    "decryption key must be 32 bytes",
		},
		{
			name:           "invalid JSON payload after base64 decode",
			encryptedToken: base64.StdEncoding.EncodeToString([]byte("not json")),
			decryptionKey:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
			expected:       "",
			expectErr:      true,
			errContains:    "not valid JSON",
		},
		{
			name:           "missing iv field",
			encryptedToken: base64.StdEncoding.EncodeToString([]byte(`{"value":"test"}`)),
			decryptionKey:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
			expected:       "",
			expectErr:      true,
			errContains:    "missing required fields",
		},
		{
			name:           "missing value field",
			encryptedToken: base64.StdEncoding.EncodeToString([]byte(`{"iv":"test"}`)),
			decryptionKey:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
			expected:       "",
			expectErr:      true,
			errContains:    "missing required fields",
		},
		{
			name:           "invalid base64 in IV field",
			encryptedToken: base64.StdEncoding.EncodeToString([]byte(`{"iv":"not-valid-base64!!!","value":"dGVzdA=="}`)),
			decryptionKey:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
			expected:       "",
			expectErr:      true,
			errContains:    "failed to base64 decode IV",
		},
		{
			name:           "invalid base64 in value field",
			encryptedToken: base64.StdEncoding.EncodeToString([]byte(`{"iv":"dGVzdA==","value":"not-valid-base64!!!"}`)),
			decryptionKey:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
			expected:       "",
			expectErr:      true,
			errContains:    "failed to base64 decode encrypted value",
		},
		{
			name:           "IV wrong length",
			encryptedToken: base64.StdEncoding.EncodeToString([]byte(`{"iv":"dGVzdA==","value":"` + base64.StdEncoding.EncodeToString(make([]byte, 32)) + `"}`)),
			decryptionKey:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
			expected:       "",
			expectErr:      true,
			errContains:    "IV length must equal AES block size",
		},
		{
			name:           "encrypted value not multiple of block size",
			encryptedToken: base64.StdEncoding.EncodeToString([]byte(`{"iv":"` + base64.StdEncoding.EncodeToString(make([]byte, 16)) + `","value":"` + base64.StdEncoding.EncodeToString(make([]byte, 17)) + `"}`)),
			decryptionKey:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
			expected:       "",
			expectErr:      true,
			errContains:    "encrypted value is not a multiple of the block size",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := DecryptToken(tc.encryptedToken, tc.decryptionKey)

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("expected error containing %q, got %q", tc.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestDecryptToken_ValidEncryption(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	base64Key := base64.StdEncoding.EncodeToString(key)

	plaintext := "my_secret_token"
	encryptedToken := createValidEncryptedPayload(t, plaintext, key)

	result, err := DecryptToken(encryptedToken, base64Key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, result)
	}
}

func TestDecryptToken_WrongKey(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	wrongKey := make([]byte, 32)
	for i := range wrongKey {
		wrongKey[i] = byte(i + 1)
	}
	base64WrongKey := base64.StdEncoding.EncodeToString(wrongKey)

	encryptedToken := createValidEncryptedPayload(t, "test", key)

	result, err := DecryptToken(encryptedToken, base64WrongKey)
	if err == nil && result == "test" {
		t.Fatal("expected wrong key to fail or return different plaintext, got original plaintext")
	}
}

func TestDecryptToken_URLSafeBase64(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	base64Key := base64.StdEncoding.EncodeToString(key)

	plaintext := "my_secret_token"

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		t.Fatalf("failed to generate IV: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	paddedPlaintext := pkcs7Pad([]byte(plaintext), aes.BlockSize)

	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	payload := EncryptedPayload{
		IV:    base64.StdEncoding.EncodeToString(iv),
		Value: base64.StdEncoding.EncodeToString(ciphertext),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	encryptedToken := base64.URLEncoding.EncodeToString(jsonPayload)

	result, err := DecryptToken(encryptedToken, base64Key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, result)
	}
}

func TestDecryptToken_FailedUnpad(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	base64Key := base64.StdEncoding.EncodeToString(key)

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		t.Fatalf("failed to generate IV: %v", err)
	}

	invalidCiphertext := make([]byte, 32)
	for i := range invalidCiphertext {
		invalidCiphertext[i] = 0xFF
	}

	payload := EncryptedPayload{
		IV:    base64.StdEncoding.EncodeToString(iv),
		Value: base64.StdEncoding.EncodeToString(invalidCiphertext),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	encryptedToken := base64.StdEncoding.EncodeToString(jsonPayload)

	_, err = DecryptToken(encryptedToken, base64Key)
	if err == nil {
		t.Fatal("expected error for invalid padding, got nil")
	}

	if !strings.Contains(err.Error(), "failed to unpad decrypted data") {
		t.Fatalf("expected error containing 'failed to unpad decrypted data', got %q", err.Error())
	}
}

func TestEncryptedPayload_Struct(t *testing.T) {
	payload := EncryptedPayload{
		IV:    "test-iv",
		Value: "test-value",
	}

	if payload.IV != "test-iv" {
		t.Fatalf("expected IV %q, got %q", "test-iv", payload.IV)
	}

	if payload.Value != "test-value" {
		t.Fatalf("expected Value %q, got %q", "test-value", payload.Value)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded EncryptedPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.IV != payload.IV || decoded.Value != payload.Value {
		t.Fatal("round-trip failed")
	}
}

func createValidEncryptedPayload(t *testing.T, plaintext string, key []byte) string {
	t.Helper()

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		t.Fatalf("failed to generate IV: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	paddedPlaintext := pkcs7Pad([]byte(plaintext), aes.BlockSize)

	ciphertext := make([]byte, len(paddedPlaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	payload := EncryptedPayload{
		IV:    base64.StdEncoding.EncodeToString(iv),
		Value: base64.StdEncoding.EncodeToString(ciphertext),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	return base64.StdEncoding.EncodeToString(jsonPayload)
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}
