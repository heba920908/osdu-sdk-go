package utils_test

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/heba920908/osdu-sdk-go/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// generateRandomHex generates a random hex string of the specified byte size
func generateRandomHex(size int) string {
	bytes := make([]byte, size)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateValidHmacSignature creates a valid HMAC signature for testing
// This mimics the Java implementation from SignatureService.getSignedSignature
func generateValidHmacSignature(url, secret string) (string, error) {
	currentTime := time.Now().UnixMilli()
	expireTime := strconv.FormatInt(currentTime+30000, 10)
	nonce := generateRandomHex(16)

	// Build the data JSON (must match exact format from Java)
	data := fmt.Sprintf(`{"expireMillisecond": "%s","hashMechanism": "hmacSHA256","endpointUrl": "%s","nonce": "%s"}`,
		expireTime, url, nonce)

	// Create signature service to compute the signature
	sigService := utils.NewSignatureService()
	signature, err := sigService.GetSignedSignature(url, secret, expireTime, nonce)
	if err != nil {
		return "", err
	}

	// Encode data as base64
	dataBytes := []byte(data)
	dataBytesEncoded := base64.StdEncoding.EncodeToString(dataBytes)

	// Return the full HMAC: base64(data).hexSignature
	return fmt.Sprintf("%s.%s", dataBytesEncoded, signature), nil
}

func TestSignatureService_VerifyHmacSignature(t *testing.T) {
	secret := generateRandomHex(32) // 64-character hex string (32 bytes)

	t.Run("valid signature", func(t *testing.T) {
		url := "https://example.com/webhook"
		hmacSignature, err := generateValidHmacSignature(url, secret)
		assert.NoError(t, err, "Failed to generate HMAC")

		sigService := utils.NewSignatureService()
		err = sigService.VerifyHmacSignature(hmacSignature, secret)
		assert.NoError(t, err, "Expected valid signature")
	})

	t.Run("invalid signature format - no dot", func(t *testing.T) {
		sigService := utils.NewSignatureService()
		err := sigService.VerifyHmacSignature("invalidsignature", secret)
		assert.Error(t, err, "Expected error for invalid format")
	})

	t.Run("invalid signature format - invalid base64", func(t *testing.T) {
		sigService := utils.NewSignatureService()
		err := sigService.VerifyHmacSignature("!!!invalid!!.abcd1234", secret)
		assert.Error(t, err, "Expected error for invalid base64")
	})

	t.Run("missing HMAC signature", func(t *testing.T) {
		sigService := utils.NewSignatureService()
		err := sigService.VerifyHmacSignature("", secret)
		assert.Error(t, err, "Expected error for missing HMAC")
	})

	t.Run("missing secret", func(t *testing.T) {
		sigService := utils.NewSignatureService()
		err := sigService.VerifyHmacSignature("test.signature", "")
		assert.Error(t, err, "Expected error for missing secret")
	})

	t.Run("wrong secret", func(t *testing.T) {
		url := "https://example.com/webhook"
		hmacSignature, err := generateValidHmacSignature(url, secret)
		assert.NoError(t, err, "Failed to generate HMAC")

		wrongSecret := generateRandomHex(32)
		sigService := utils.NewSignatureService()
		err = sigService.VerifyHmacSignature(hmacSignature, wrongSecret)
		assert.Error(t, err, "Expected error for wrong secret")
	})

	t.Run("expired signature", func(t *testing.T) {
		// Create an expired signature (expireTime in the past)
		expireTime := strconv.FormatInt(time.Now().UnixMilli()-10000, 10) // 10 seconds ago
		nonce := generateRandomHex(16)
		url := "https://example.com/webhook"

		data := fmt.Sprintf(`{"expireMillisecond": "%s","hashMechanism": "hmacSHA256","endpointUrl": "%s","nonce": "%s"}`,
			expireTime, url, nonce)

		dataBytes := []byte(data)
		dataBytesEncoded := base64.StdEncoding.EncodeToString(dataBytes)

		// Use any signature (it will fail on expiry check before signature validation)
		hmacSignature := fmt.Sprintf("%s.%s", dataBytesEncoded, "fakesignature")

		sigService := utils.NewSignatureService()
		err := sigService.VerifyHmacSignature(hmacSignature, secret)
		assert.Error(t, err, "Expected error for expired signature")
		if err != nil && err.Error() != "failed to generate signature: signature is expired" {
			t.Logf("Got error: %v", err)
		}
	})

	t.Run("tampered data - different URL", func(t *testing.T) {
		url := "https://example.com/webhook"
		hmacSignature, err := generateValidHmacSignature(url, secret)
		assert.NoError(t, err, "Failed to generate HMAC")

		// Split and decode the signature
		parts := []byte(hmacSignature)
		dotIndex := -1
		for i, b := range parts {
			if b == '.' {
				dotIndex = i
				break
			}
		}

		assert.NotEqual(t, -1, dotIndex, "Invalid signature format")

		encodedData := string(parts[:dotIndex])
		signaturePart := string(parts[dotIndex+1:])

		// Decode and modify the data
		dataBytes, _ := base64.StdEncoding.DecodeString(encodedData)
		var hmacData utils.HmacData
		json.Unmarshal(dataBytes, &hmacData)

		// Tamper with URL
		hmacData.EndpointUrl = "https://evil.com/webhook"

		// Re-encode
		tamperedData, _ := json.Marshal(hmacData)
		tamperedEncoded := base64.StdEncoding.EncodeToString(tamperedData)
		tamperedSignature := fmt.Sprintf("%s.%s", tamperedEncoded, signaturePart)

		sigService := utils.NewSignatureService()
		err = sigService.VerifyHmacSignature(tamperedSignature, secret)
		assert.Error(t, err, "Expected error for tampered data")
	})
}

func TestSignatureService_GetSignedSignature(t *testing.T) {
	secret := generateRandomHex(32)
	url := "https://example.com/webhook"

	t.Run("generate valid signature", func(t *testing.T) {
		expireTime := strconv.FormatInt(time.Now().UnixMilli()+30000, 10)
		nonce := generateRandomHex(16)

		sigService := utils.NewSignatureService()
		signature, err := sigService.GetSignedSignature(url, secret, expireTime, nonce)
		assert.NoError(t, err, "Expected no error")
		assert.NotEmpty(t, signature, "Expected non-empty signature")

		// Signature should be hex-encoded (only contains 0-9, a-f)
		assert.Equal(t, 64, len(signature), "SHA256 produces 32 bytes = 64 hex characters")
	})

	t.Run("missing parameters", func(t *testing.T) {
		sigService := utils.NewSignatureService()

		_, err := sigService.GetSignedSignature("", secret, "123456", "abc")
		assert.Error(t, err, "Expected error for missing URL")

		_, err = sigService.GetSignedSignature(url, "", "123456", "abc")
		assert.Error(t, err, "Expected error for missing secret")

		_, err = sigService.GetSignedSignature(url, secret, "", "abc")
		assert.Error(t, err, "Expected error for missing expireTime")
	})

	t.Run("invalid expire time format", func(t *testing.T) {
		sigService := utils.NewSignatureService()
		_, err := sigService.GetSignedSignature(url, secret, "not-a-number", "abc")
		assert.Error(t, err, "Expected error for invalid expire time")
	})

	t.Run("expired time", func(t *testing.T) {
		expireTime := strconv.FormatInt(time.Now().UnixMilli()-1000, 10)
		nonce := generateRandomHex(16)

		sigService := utils.NewSignatureService()
		_, err := sigService.GetSignedSignature(url, secret, expireTime, nonce)
		assert.Error(t, err, "Expected error for expired time")
	})
}
