package feishubot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// GenSign generates a signature for Feishu webhook signature verification.
//
// The signing algorithm follows the Feishu documentation:
// 1. Create sign string: timestamp + "\n" + secret
// 2. Use HmacSHA256 algorithm with the sign string as key
// 3. Sign an empty byte array
// 4. Base64 encode the result
//
// Parameters:
//   - secret: The signing secret configured in the Feishu bot settings
//   - timestamp: Unix timestamp (seconds since epoch), must be within 1 hour of current time
//
// Returns:
//   - The Base64-encoded signature string
//   - An error if HMAC creation fails (should never happen with sha256)
//
// Example:
//
//	timestamp := time.Now().Unix()
//	sign, err := GenSign("your_secret", timestamp)
//	if err != nil {
//	    // handle error
//	}
//	message.Sign = sign
//	message.Timestamp = timestamp
func GenSign(secret string, timestamp int64) (string, error) {
	// Create the string to sign: timestamp + "\n" + secret
	// This matches the Feishu documentation format
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret

	// Create HMAC-SHA256 with the sign string as the key
	h := hmac.New(sha256.New, []byte(stringToSign))

	// Sign an empty byte array (as per Feishu documentation)
	_, err := h.Write([]byte{})
	if err != nil {
		return "", fmt.Errorf("failed to write to hmac: %w", err)
	}

	// Base64 encode the signature
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature, nil
}
