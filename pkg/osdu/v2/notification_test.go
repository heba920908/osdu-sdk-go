package v2_test

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/heba920908/osdu-sdk-go/pkg/models/notification"
	v2 "github.com/heba920908/osdu-sdk-go/pkg/osdu/v2"
	"github.com/heba920908/osdu-sdk-go/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// generateRandomHex generates a random hex string of the specified byte size
func generateRandomHex(size int) string {
	bytes := make([]byte, size)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getExpectedChallengeResponse mimics the Java getExpectedResponse method
// from ChallengeResponseCheck.java
// This computes what the challenge response should be for a given crc and secret
func getExpectedChallengeResponse(secret, crc string) string {
	// Step 1: Concatenate secret + crc (same as HashingUtil.hashString)
	combined := secret + crc

	// Step 2: Compute SHA-256 hash
	hash := sha256.Sum256([]byte(combined))

	// Step 3: Convert hash to hex string
	hexString := hex.EncodeToString(hash[:])

	// Step 4: Base64 encode the hex string (as bytes)
	return base64.StdEncoding.EncodeToString([]byte(hexString))
}

// generateValidHmacSignature creates a valid HMAC signature for testing
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

func TestComputeChallengeResponse(t *testing.T) {
	handler := v2.NewNotificationHandler()

	testCases := []struct {
		name   string
		crc    string
		secret string
	}{
		{
			name:   "uuid crc mimicking real scenario",
			crc:    uuid.New().String(),
			secret: generateRandomHex(32), // 64-char hex secret
		},
		{
			name:   "real-world example with uuid",
			crc:    "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			secret: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := handler.ComputeChallengeResponse(tc.crc, tc.secret)

			// Compute expected response using the same algorithm as Java
			expected := getExpectedChallengeResponse(tc.secret, tc.crc)

			// Verify the result matches the expected response
			assert.Equal(t, expected, result, "ComputeChallengeResponse() should match expected response")

			// Verify it returns a non-empty base64 string
			assert.NotEmpty(t, result, "ComputeChallengeResponse() returned empty string")

			// Verify it's a valid base64 string (should not panic)
			// Base64 encoded SHA256 hex string should be 88 characters
			assert.Equal(t, 88, len(result), "ComputeChallengeResponse() length should be 88")
		})
	}
}

func TestHandleChallenge(t *testing.T) {
	handler := v2.NewNotificationHandler()
	secret := generateRandomHex(32) // Valid 64-char hex secret
	// Mimic the real challenge scenario
	crc := uuid.New().String()
	// Compute expected response using the same algorithm as Java
	expected := getExpectedChallengeResponse(secret, crc)

	t.Run("successful challenge with valid HMAC", func(t *testing.T) {
		url := "https://example.com/webhook"

		// Generate valid HMAC signature
		hmacSignature, err := generateValidHmacSignature(url, secret)
		assert.NoError(t, err, "Failed to generate HMAC")
		response, err := handler.HandleChallenge(crc, hmacSignature, secret)
		assert.NoError(t, err, "HandleChallenge() should not error")
		assert.NotEmpty(t, response.ResponseHash, "HandleChallenge() ResponseHash should not be empty")

		// Verify the result matches the expected response
		t.Logf("Response Hash --- Expected: %s, Got: %s", expected, response.ResponseHash)
		assert.Equal(t, expected, response.ResponseHash, "ComputeChallengeResponse() should match expected response")
	})

	t.Run("missing crc", func(t *testing.T) {
		url := "https://example.com/webhook"
		hmacSignature, _ := generateValidHmacSignature(url, secret)

		_, err := handler.HandleChallenge("", hmacSignature, secret)
		assert.Error(t, err, "HandleChallenge() should error for missing crc")
	})

	t.Run("missing hmac", func(t *testing.T) {
		response, err := handler.HandleChallenge(crc, "", secret)
		assert.NoError(t, err, "HandleChallenge() should succeed even without hmac")
		assert.NotEmpty(t, response.ResponseHash, "HandleChallenge() should return response hash")
		assert.Equal(t, expected, response.ResponseHash, "ComputeChallengeResponse() should match expected response")
	})

	t.Run("missing secret", func(t *testing.T) {
		url := "https://example.com/webhook"
		hmacSignature, _ := generateValidHmacSignature(url, secret)

		_, err := handler.HandleChallenge(crc, hmacSignature, "")
		assert.Error(t, err, "HandleChallenge() should error for missing secret")
	})

	t.Run("invalid HMAC signature", func(t *testing.T) {
		invalidHmac := "invalid.signature.format"

		response, err := handler.HandleChallenge(crc, invalidHmac, secret)
		assert.NoError(t, err, "HandleChallenge() should succeed even with invalid HMAC")
		assert.NotEmpty(t, response.ResponseHash, "HandleChallenge() should return response hash")
		assert.Equal(t, expected, response.ResponseHash, "ComputeChallengeResponse() should match expected response")
	})

	t.Run("wrong secret for HMAC", func(t *testing.T) {
		url := "https://example.com/webhook"

		// Generate HMAC with one secret
		hmacSignature, _ := generateValidHmacSignature(url, secret)

		// Try to validate with different secret - should still compute response
		wrongSecret := generateRandomHex(32)
		response, err := handler.HandleChallenge(crc, hmacSignature, wrongSecret)
		assert.NoError(t, err, "HandleChallenge() should succeed even with wrong secret")
		assert.NotEmpty(t, response.ResponseHash, "HandleChallenge() should return response hash")
		// The generated responseHash will differ since secret is different
		t.Logf("%s should be different than %s", expected, response.ResponseHash)
		assert.NotEqual(t, expected, response.ResponseHash, "ComputeChallengeResponse() should match expected response")
	})
}

func TestParseNotificationPayload_M18Format(t *testing.T) {
	handler := v2.NewNotificationHandler()

	// M18+ format: direct array of records
	payload := `[
		{
			"id": "opendes:dataset:1.0.0",
			"kind": "opendes:wks:dataset:1.0.0",
			"op": "create"
		},
		{
			"id": "opendes:dataset:2.0.0",
			"kind": "opendes:wks:dataset:2.0.0",
			"op": "update"
		}
	]`

	msg, err := handler.ParseNotificationPayload([]byte(payload))
	assert.NoError(t, err, "ParseNotificationPayload() should not error")
	assert.NotNil(t, msg, "ParseNotificationPayload() should not return nil")
	assert.Equal(t, "RecordsChanged", msg.Subject, "ParseNotificationPayload() Subject")
	assert.Equal(t, "opendes", msg.Data.DataPartitionID, "ParseNotificationPayload() DataPartitionID")
	assert.Equal(t, 2, len(msg.Data.Records), "ParseNotificationPayload() record count")
	assert.Equal(t, "create", msg.Data.Records[0].Op, "ParseNotificationPayload() first record op")
}

func TestParseNotificationPayload_LegacyFormat(t *testing.T) {
	handler := v2.NewNotificationHandler()

	// Legacy format: full message structure
	payload := `{
		"id": "msg-123",
		"eventTime": "2025-11-25T12:00:00Z",
		"subject": "RecordsChanged",
		"topic": "records-changed",
		"data": {
			"dataPartitionId": "opendes",
			"data": [
				{
					"id": "opendes:dataset:1.0.0",
					"kind": "opendes:wks:dataset:1.0.0",
					"op": "create"
				}
			]
		}
	}`

	msg, err := handler.ParseNotificationPayload([]byte(payload))
	assert.NoError(t, err, "ParseNotificationPayload() should not error")
	assert.NotNil(t, msg, "ParseNotificationPayload() should not return nil")
	assert.Equal(t, "msg-123", msg.ID, "ParseNotificationPayload() ID")
	assert.Equal(t, "RecordsChanged", msg.Subject, "ParseNotificationPayload() Subject")
	assert.Equal(t, "opendes", msg.Data.DataPartitionID, "ParseNotificationPayload() DataPartitionID")
	assert.Equal(t, 1, len(msg.Data.Records), "ParseNotificationPayload() record count")
}

func TestParseNotificationPayload_InvalidJSON(t *testing.T) {
	handler := v2.NewNotificationHandler()

	payload := `invalid json`

	_, err := handler.ParseNotificationPayload([]byte(payload))
	assert.Error(t, err, "ParseNotificationPayload() should error for invalid JSON")
}

func TestNotificationWebhookHandler_HandleChallengeHTTP(t *testing.T) {
	secret := generateRandomHex(32) // Valid 64-char hex secret
	handler := v2.NewNotificationWebhookHandler(secret, nil)

	t.Run("successful challenge with valid HMAC", func(t *testing.T) {
		crc := uuid.New().String()
		url := "https://example.com/webhook"

		// Generate valid HMAC signature
		hmacSignature, err := generateValidHmacSignature(url, secret)
		assert.NoError(t, err, "Failed to generate HMAC")

		t.Logf("Generated HMAC: %s", hmacSignature)
		t.Logf("Generated CRC: %s", crc)
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?crc=%s&hmac=%s", crc, hmacSignature), nil)
		w := httptest.NewRecorder()

		handler.HandleChallengeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if got, want := resp.StatusCode, http.StatusOK; got != want {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("HandleChallengeHTTP() status = %d, want %d; body: %s", got, want, string(body))
		}

		var response notification.ChallengeResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err, "json.Decode() should not error")
		assert.NotEmpty(t, response.ResponseHash, "HandleChallengeHTTP() ResponseHash should not be empty")
	})

	t.Run("missing crc parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.HandleChallengeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "HandleChallengeHTTP() status")
	})

	t.Run("challenge without hmac parameter", func(t *testing.T) {
		// Should succeed even without HMAC
		crc := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?crc=%s", crc), nil)
		w := httptest.NewRecorder()

		handler.HandleChallengeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "HandleChallengeHTTP() should succeed without HMAC")

		var response notification.ChallengeResponse
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err, "json.Decode() should not error")
		assert.NotEmpty(t, response.ResponseHash, "Response should contain hash")
	})

	t.Run("invalid HMAC signature", func(t *testing.T) {
		crc := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?crc=%s&hmac=invalid", crc), nil)
		w := httptest.NewRecorder()

		handler.HandleChallengeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()
		// Should succeed with warning logged
		assert.Equal(t, http.StatusOK, resp.StatusCode, "HandleChallengeHTTP() should return 200 even with invalid HMAC")

		var response notification.ChallengeResponse
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err, "json.Decode() should not error")
		assert.NotEmpty(t, response.ResponseHash, "Response should contain hash")
	})

	t.Run("wrong method", func(t *testing.T) {
		crc := uuid.New().String()
		url := "https://example.com/webhook"
		hmacSignature, _ := generateValidHmacSignature(url, secret)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/?crc=%s&hmac=%s", crc, hmacSignature), nil)
		w := httptest.NewRecorder()

		handler.HandleChallengeHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "HandleChallengeHTTP() status")
	})
}

func TestNotificationWebhookHandler_HandleNotificationHTTP(t *testing.T) {
	var receivedMsg *notification.NotificationMessage
	onNotification := func(msg *notification.NotificationMessage) (interface{}, error) {
		receivedMsg = msg
		return nil, nil // Return nil to use default response
	}

	handler := v2.NewNotificationWebhookHandler("secret", onNotification)

	t.Run("successful notification M18+", func(t *testing.T) {
		payload := `[
			{
				"id": "opendes:dataset:1.0.0",
				"kind": "opendes:wks:dataset:1.0.0",
				"op": "create"
			}
		]`

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(payload))
		w := httptest.NewRecorder()

		handler.HandleNotificationHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if got, want := resp.StatusCode, http.StatusOK; got != want {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("HandleNotificationHTTP() status = %d, want %d; body: %s", got, want, string(body))
		}

		var response notification.NotificationHandlerResponse
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err, "json.Decode() should not error")
		assert.Equal(t, "OK", response.Status, "HandleNotificationHTTP() Status")
		assert.NotNil(t, receivedMsg, "HandleNotificationHTTP() callback should be called")
		assert.Equal(t, "opendes", receivedMsg.Data.DataPartitionID, "HandleNotificationHTTP() DataPartitionID")
	})

	t.Run("wrong method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.HandleNotificationHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "HandleNotificationHTTP() status")
	})

	t.Run("invalid payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("invalid"))
		w := httptest.NewRecorder()

		handler.HandleNotificationHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "HandleNotificationHTTP() status")
	})

	t.Run("custom response", func(t *testing.T) {
		// Test with custom response structure
		customHandler := v2.NewNotificationWebhookHandler("secret", func(msg *notification.NotificationMessage) (interface{}, error) {
			// Return custom response
			return map[string]interface{}{
				"notificationId": msg.ID,
				"status":         "PROCESSED",
				"message":        "Successfully processed notification",
				"recordCount":    len(msg.RecordEvents),
			}, nil
		})

		payload := `[{"id": "opendes:dataset:1.0.0", "kind": "opendes:wks:dataset:1.0.0", "op": "create"}]`
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(payload))
		w := httptest.NewRecorder()

		customHandler.HandleNotificationHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "HandleNotificationHTTP() status")

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err, "json.Decode() should not error")
		assert.Equal(t, "PROCESSED", response["status"], "Custom status should be returned")
		assert.Equal(t, "Successfully processed notification", response["message"], "Custom message should be returned")
		assert.Equal(t, float64(1), response["recordCount"], "Record count should be included")
	})
}

func TestNotificationWebhookHandler_HandleHealthHTTP(t *testing.T) {
	handler := v2.NewNotificationWebhookHandler("secret", nil)

	t.Run("health check", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.HandleHealthHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "HandleHealthHTTP() status")

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "Status OK", string(body), "HandleHealthHTTP() body")
	})

	t.Run("wrong method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/health", nil)
		w := httptest.NewRecorder()

		handler.HandleHealthHTTP(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "HandleHealthHTTP() status")
	})
}

// https://community.opengroup.org/osdu/platform/system/register/-/blob/v0.26.0/provider/register-azure/src/main/resources/application.properties?ref_type=tags#L73
// Test with hardcoded secret, which should return always the same response for given crc
func TestComputeChallengeResponse_DefaultSecret(t *testing.T) {
	handler := v2.NewNotificationHandler()
	secret := "395f1b05e95171d7c0dde0b19fd6cf"
	crc := "test-crc-12345"

	result := handler.ComputeChallengeResponse(crc, secret)
	expected := getExpectedChallengeResponse(secret, crc)

	t.Logf("Response hash: %s", result)
	assert.Equal(t, expected, result)
}
