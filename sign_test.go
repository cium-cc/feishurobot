package feishubot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

// TestGenSign tests the GenSign function with known expected values.
func TestGenSign(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		timestamp int64
		want      string
		wantErr   bool
	}{
		{
			name:      "demo secret with timestamp from documentation",
			secret:    "demo",
			timestamp: 1599360473,
			want:      calculateExpectedSign("demo", 1599360473),
			wantErr:   false,
		},
		{
			name:      "empty secret",
			secret:    "",
			timestamp: 1599360473,
			want:      calculateExpectedSign("", 1599360473),
			wantErr:   false,
		},
		{
			name:      "special characters in secret",
			secret:    "abc123!@#$%^&*()",
			timestamp: 1234567890,
			want:      calculateExpectedSign("abc123!@#$%^&*()", 1234567890),
			wantErr:   false,
		},
		{
			name:      "unicode characters in secret",
			secret:    "密钥demo",
			timestamp: 1599360473,
			want:      calculateExpectedSign("密钥demo", 1599360473),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenSign(tt.secret, tt.timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenSign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenSign() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGenSignConsistency tests that GenSign produces consistent results.
func TestGenSignConsistency(t *testing.T) {
	secret := "test_secret_123"
	timestamp := int64(1234567890)

	// Generate the same signature multiple times
	var signatures []string
	for i := 0; i < 10; i++ {
		sign, err := GenSign(secret, timestamp)
		require.NoError(t, err)
		signatures = append(signatures, sign)
	}

	// All signatures should be identical
	for i := 1; i < len(signatures); i++ {
		if signatures[i] != signatures[0] {
			t.Errorf("GenSign() produced different results: %v vs %v", signatures[i], signatures[0])
		}
	}
}

// TestGenSignDifferentInputs tests that different inputs produce different signatures.
func TestGenSignDifferentInputs(t *testing.T) {
	baseSecret := "secret"
	baseTimestamp := int64(1234567890)

	tests := []struct {
		name      string
		secret    string
		timestamp int64
	}{
		{"different secret", "different_secret", baseTimestamp},
		{"different timestamp", baseSecret, baseTimestamp + 1},
		{"both different", "another_secret", baseTimestamp + 100},
	}

	baseSign, err := GenSign(baseSecret, baseTimestamp)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenSign(tt.secret, tt.timestamp)
			require.NoError(t, err)
			if got == baseSign {
				t.Errorf("GenSign() produced same signature for different inputs: %v", got)
			}
		})
	}
}

// TestGenSignFormat tests the format of the generated signature.
func TestGenSignFormat(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		timestamp int64
	}{
		{"basic", "secret", 1234567890},
		{"long secret", "a_very_long_secret_key_for_testing_purposes", 1599360473},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenSign(tt.secret, tt.timestamp)
			require.NoError(t, err)

			// Signature should be Base64 encoded (44 characters for 32-byte HMAC-SHA256)
			decoded, err := base64.StdEncoding.DecodeString(got)
			require.NoError(t, err, "signature should be valid Base64")

			// Decoded signature should be 32 bytes (SHA256 output)
			if len(decoded) != sha256.Size {
				t.Errorf("GenSign() decoded signature length = %d, want %d", len(decoded), sha256.Size)
			}
		})
	}
}

// TestGenSignStringEncoding tests that the string to sign is formatted correctly.
func TestGenSignStringEncoding(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		timestamp int64
	}{
		{"basic", "secret", 1234567890},
		{"timestamp with newline separator", "demo", 1599360473},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The signature should use timestamp + "\n" + secret as the signing string
			expectedSign := calculateExpectedSign(tt.secret, tt.timestamp)
			got, err := GenSign(tt.secret, tt.timestamp)

			require.NoError(t, err)
			if !cmp.Equal(got, expectedSign) {
				t.Errorf("GenSign() = %v, want %v", got, expectedSign)
			}
		})
	}
}

// Helper function to calculate expected signature using the documented algorithm.
// This is a reference implementation that matches the Feishu documentation.
func calculateExpectedSign(secret string, timestamp int64) string {
	// Reference implementation from documentation:
	// stringToSign := timestamp + "\n" + secret
	// h := hmac.New(sha256.New, []byte(stringToSign))
	// _, err := h.Write([]byte{})
	// signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	h := hmac.New(sha256.New, []byte(stringToSign))
	h.Write(nil)
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature
}
