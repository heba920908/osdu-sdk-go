package v2

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/heba920908/osdu-sdk-go/pkg/models/notification"
)

// NotificationHandler provides utilities for handling incoming notification webhooks
// and challenge-response validation from OSDU notification service
type NotificationHandler interface {
	// ComputeChallengeResponse computes the response hash for the challenge-response validation
	// This implements the logic from HashingUtil.java
	ComputeChallengeResponse(crc, secret string) string

	// HandleChallenge handles the GET challenge request from notification service
	// Expected query params: crc=xxx&hmac=yyy
	// Returns the responseHash that should be sent back
	HandleChallenge(crc, hmac, secret string) (notification.ChallengeResponse, error)

	// ParseNotificationPayload parses the incoming notification payload
	// Supports both legacy (pre-M18) and current (M18+) formats
	ParseNotificationPayload(body []byte) (*notification.NotificationMessage, error)
}

// notificationHandler implements NotificationHandler interface
type notificationHandler struct{}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler() NotificationHandler {
	return &notificationHandler{}
}

// ComputeChallengeResponse implements the hashing logic from HashingUtil.java
// Java implementation:
//
//	String response = secret + crc;
//	response = Hashing.sha256().hashString(response, StandardCharsets.UTF_8).toString();
//	return Base64.getEncoder().encodeToString(response.getBytes());
func (h *notificationHandler) ComputeChallengeResponse(crc, secret string) string {
	// Step 1: Concatenate secret + crc
	input := secret + crc

	// Step 2: Compute SHA-256 hash and get hex string
	hash := sha256.Sum256([]byte(input))
	hexString := hex.EncodeToString(hash[:])

	// Step 3: Base64 encode the hex string (as bytes)
	encoded := base64.StdEncoding.EncodeToString([]byte(hexString))

	return encoded
}

// HandleChallenge handles the challenge-response validation
// GET /?crc=xxxxxx&hmac=ssxxsxsxsxx
// TODO: Implement HMAC signature verification
func (h *notificationHandler) HandleChallenge(crc, hmac, secret string) (notification.ChallengeResponse, error) {
	if crc == "" || hmac == "" {
		slog.Warn("Challenge validation failed: missing parameters")
		return notification.ChallengeResponse{}, fmt.Errorf("crc and hmac parameters are required")
	}

	if secret == "" {
		slog.Warn("Challenge validation failed: missing secret")
		return notification.ChallengeResponse{}, fmt.Errorf("secret is required")
	}

	// TODO: Verify HMAC signature
	// See: https://community.opengroup.org/osdu/platform/system/lib/core/os-core-common/-/blob/v0.21.0/src/main/java/org/opengroup/osdu/core/common/cryptographic/SignatureService.java

	// Compute the challenge response
	responseHash := h.ComputeChallengeResponse(crc, secret)

	slog.Debug("Challenge response computed successfully", "crcLength", len(crc))
	return notification.ChallengeResponse{
		ResponseHash: responseHash,
	}, nil
}

// ParseNotificationPayload parses incoming notification payload
// Supports both formats:
// - M18+: Direct array of records [{id, kind, op}, ...]
// - Legacy: Full message structure with id, eventTime, subject, data
func (h *notificationHandler) ParseNotificationPayload(body []byte) (*notification.NotificationMessage, error) {
	// Try M18+ format first (direct array)
	var records []notification.NotificationRecord
	err := json.Unmarshal(body, &records)

	if err == nil && len(records) > 0 {
		// M18+ format detected - construct NotificationMessage
		slog.Debug("Parsing notification payload", "format", "M18+", "recordCount", len(records))
		dataPartitionID := extractDataPartitionID(records[0].ID)

		msg := &notification.NotificationMessage{
			ID:        fmt.Sprintf("%s-%d", dataPartitionID, time.Now().UnixMilli()),
			EventTime: time.Now().Format(time.RFC3339),
			Subject:   "RecordsChanged",
			Topic:     "unknown",
			Data: notification.NotificationMessageData{
				DataPartitionID: dataPartitionID,
				Records:         records,
			},
			RecordEvents: records,
		}
		return msg, nil
	}

	// Try legacy format (pre-M18)
	slog.Debug("Parsing notification payload", "format", "legacy")
	var msg notification.NotificationMessage
	err = json.Unmarshal(body, &msg)
	if err != nil {
		slog.Error("Failed to parse notification payload", "error", err)
		return nil, fmt.Errorf("failed to parse notification payload: %w", err)
	}

	return &msg, nil
}

// extractDataPartitionID extracts the data partition ID from a record ID
// Record ID format: "datapartition:kind:version"
func extractDataPartitionID(recordID string) string {
	parts := strings.Split(recordID, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

// NotificationWebhookHandler provides HTTP handler functions for notification webhooks
// These can be used with any HTTP router (net/http, gorilla/mux, chi, etc.)
type NotificationWebhookHandler struct {
	handler NotificationHandler
	secret  string
	logger  *slog.Logger
	// Optional callback to process received notifications
	OnNotification func(*notification.NotificationMessage) error
}

// NewNotificationWebhookHandler creates a new webhook handler
func NewNotificationWebhookHandler(secret string, onNotification func(*notification.NotificationMessage) error) *NotificationWebhookHandler {
	return &NotificationWebhookHandler{
		handler:        NewNotificationHandler(),
		secret:         secret,
		logger:         slog.Default(),
		OnNotification: onNotification,
	}
}

// WithLogger sets a custom logger for the webhook handler
func (nwh *NotificationWebhookHandler) WithLogger(logger *slog.Logger) *NotificationWebhookHandler {
	nwh.logger = logger
	return nwh
}

// HandleChallengeHTTP is an HTTP handler for challenge-response validation
// GET /?crc=xxxxxx&hmac=ssxxsxsxsxx
func (nwh *NotificationWebhookHandler) HandleChallengeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		nwh.logger.Warn("Invalid HTTP method for challenge", "method", r.Method)
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	crc := r.URL.Query().Get("crc")
	hmac := r.URL.Query().Get("hmac")

	nwh.logger.Debug("Handling challenge request", "crcLength", len(crc), "hmacLength", len(hmac))

	response, err := nwh.handler.HandleChallenge(crc, hmac, nwh.secret)
	if err != nil {
		nwh.logger.Error("Challenge validation failed", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nwh.logger.Info("Challenge validated successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleNotificationHTTP is an HTTP handler for receiving notification events
// POST / with notification payload in body
func (nwh *NotificationWebhookHandler) HandleNotificationHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		nwh.logger.Warn("Invalid HTTP method for notification", "method", r.Method)
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		nwh.logger.Error("Failed to read notification request body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	nwh.logger.Debug("Received notification payload", "bodySize", len(body))

	msg, err := nwh.handler.ParseNotificationPayload(body)
	if err != nil {
		nwh.logger.Error("Failed to parse notification payload", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call the callback if provided
	if nwh.OnNotification != nil {
		nwh.logger.Debug("Processing notification",
			"notificationId", msg.ID,
			"subject", msg.Subject)
		if err := nwh.OnNotification(msg); err != nil {
			nwh.logger.Error("Failed to process notification",
				"error", err,
				"notificationId", msg.ID)
			http.Error(w, "Failed to process notification", http.StatusInternalServerError)
			return
		}
	}

	nwh.logger.Info("Notification processed successfully",
		"notificationId", msg.ID,
		"subject", msg.Subject,
		"recordCount", len(msg.RecordEvents))

	// Send success response
	response := notification.NotificationHandlerResponse{
		NotificationID: msg.ID,
		Status:         "OK",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleHealthHTTP is a simple health check endpoint
func (nwh *NotificationWebhookHandler) HandleHealthHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "Status OK")
}
