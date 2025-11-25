package v2_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heba920908/osdu-sdk-go/pkg/models/notification"
	v2 "github.com/heba920908/osdu-sdk-go/pkg/osdu/v2"
)

func TestComputeChallengeResponse(t *testing.T) {
	handler := v2.NewNotificationHandler()

	testCases := []struct {
		name     string
		crc      string
		secret   string
		expected string // We'll verify the format, not exact value
	}{
		{
			name:   "basic challenge",
			crc:    "12345",
			secret: "mysecret",
		},
		{
			name:   "empty crc",
			crc:    "",
			secret: "mysecret",
		},
		{
			name:   "long secret",
			crc:    "abcdef",
			secret: "verylongsecretkeythatisused",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := handler.ComputeChallengeResponse(tc.crc, tc.secret)

			// Verify it returns a non-empty base64 string
			if result == "" {
				t.Error("Expected non-empty result")
			}

			// Verify it's a valid base64 string (should not panic)
			// Base64 encoded SHA256 hex string should be 88 characters
			if len(result) != 88 {
				t.Errorf("Expected base64 length of 88, got %d", len(result))
			}
		})
	}
}

func TestHandleChallenge(t *testing.T) {
	handler := v2.NewNotificationHandler()

	t.Run("successful challenge", func(t *testing.T) {
		response, err := handler.HandleChallenge("12345", "hmac123", "secret")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.ResponseHash == "" {
			t.Error("Expected non-empty responseHash")
		}
	})

	t.Run("missing crc", func(t *testing.T) {
		_, err := handler.HandleChallenge("", "hmac123", "secret")
		if err == nil {
			t.Error("Expected error for missing crc")
		}
	})

	t.Run("missing hmac", func(t *testing.T) {
		_, err := handler.HandleChallenge("12345", "", "secret")
		if err == nil {
			t.Error("Expected error for missing hmac")
		}
	})

	t.Run("missing secret", func(t *testing.T) {
		_, err := handler.HandleChallenge("12345", "hmac123", "")
		if err == nil {
			t.Error("Expected error for missing secret")
		}
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
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if msg == nil {
		t.Fatal("Expected message, got nil")
	}

	if msg.Subject != "RecordsChanged" {
		t.Errorf("Expected Subject 'RecordsChanged', got '%s'", msg.Subject)
	}

	if msg.Data.DataPartitionID != "opendes" {
		t.Errorf("Expected DataPartitionID 'opendes', got '%s'", msg.Data.DataPartitionID)
	}

	if len(msg.Data.Records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(msg.Data.Records))
	}

	if msg.Data.Records[0].Op != "create" {
		t.Errorf("Expected op 'create', got '%s'", msg.Data.Records[0].Op)
	}
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
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if msg == nil {
		t.Fatal("Expected message, got nil")
	}

	if msg.ID != "msg-123" {
		t.Errorf("Expected ID 'msg-123', got '%s'", msg.ID)
	}

	if msg.Subject != "RecordsChanged" {
		t.Errorf("Expected Subject 'RecordsChanged', got '%s'", msg.Subject)
	}

	if msg.Data.DataPartitionID != "opendes" {
		t.Errorf("Expected DataPartitionID 'opendes', got '%s'", msg.Data.DataPartitionID)
	}

	if len(msg.Data.Records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(msg.Data.Records))
	}
}

func TestParseNotificationPayload_InvalidJSON(t *testing.T) {
	handler := v2.NewNotificationHandler()

	payload := `invalid json`

	_, err := handler.ParseNotificationPayload([]byte(payload))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestNotificationWebhookHandler_HandleChallengeHTTP(t *testing.T) {
	secret := "test-secret"
	handler := v2.NewNotificationWebhookHandler(secret, nil)

	t.Run("successful challenge", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?crc=12345&hmac=abc123", nil)
		w := httptest.NewRecorder()

		handler.HandleChallengeHTTP(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var response notification.ChallengeResponse
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.ResponseHash == "" {
			t.Error("Expected non-empty responseHash")
		}
	})

	t.Run("missing parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.HandleChallengeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("wrong method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/?crc=12345&hmac=abc123", nil)
		w := httptest.NewRecorder()

		handler.HandleChallengeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", resp.StatusCode)
		}
	})
}

func TestNotificationWebhookHandler_HandleNotificationHTTP(t *testing.T) {
	var receivedMsg *notification.NotificationMessage
	onNotification := func(msg *notification.NotificationMessage) error {
		receivedMsg = msg
		return nil
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

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, string(body))
		}

		var response notification.NotificationHandlerResponse
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Status != "OK" {
			t.Errorf("Expected status 'OK', got '%s'", response.Status)
		}

		if receivedMsg == nil {
			t.Fatal("Expected callback to be called")
		}

		if receivedMsg.Data.DataPartitionID != "opendes" {
			t.Errorf("Expected partition 'opendes', got '%s'", receivedMsg.Data.DataPartitionID)
		}
	})

	t.Run("wrong method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.HandleNotificationHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("invalid"))
		w := httptest.NewRecorder()

		handler.HandleNotificationHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
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

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "Status OK" {
			t.Errorf("Expected 'Status OK', got '%s'", string(body))
		}
	})

	t.Run("wrong method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/health", nil)
		w := httptest.NewRecorder()

		handler.HandleHealthHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", resp.StatusCode)
		}
	})
}
