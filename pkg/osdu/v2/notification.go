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
	"github.com/heba920908/osdu-sdk-go/pkg/utils"
)

// ComputeChallengeResponse computes the response hash for the challenge-response validation
// This implements the logic from HashingUtil.java - public static String hashString(String crc, String secret)
// Can be used independently without creating a NotificationHandler
//
// Java implementation:
//
//	String response = secret + crc;
//	response = Hashing.sha256().hashString(response, StandardCharsets.UTF_8).toString();
//	return Base64.getEncoder().encodeToString(response.getBytes());
func ComputeChallengeResponse(crc, secret string) string {
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
// Verifies the HMAC signature (warns if invalid) and computes the challenge response
// Can be used independently without creating a NotificationHandler
//
// Note: HMAC verification failure is logged as warning but doesn't block the response
// This matches the Java ChallengeResponseCheck behavior where the challenge response is always computed
func HandleChallenge(crc, hmac, secret string) (notification.ChallengeResponse, error) {
	if crc == "" {
		slog.Warn("Challenge validation failed: missing crc parameter")
		return notification.ChallengeResponse{}, fmt.Errorf("crc parameter is required")
	}

	if secret == "" {
		slog.Warn("Challenge validation failed: missing secret")
		return notification.ChallengeResponse{}, fmt.Errorf("secret is required")
	}

	// Verify HMAC signature if provided (warn if invalid, but continue)
	if hmac != "" {
		signatureService := utils.NewSignatureService()
		if err := signatureService.VerifyHmacSignature(hmac, secret); err != nil {
			slog.Warn("HMAC signature verification failed", "error", err)
			// Don't return error - continue with challenge response
		} else {
			slog.Debug("HMAC signature verified successfully")
		}
	} else {
		slog.Warn("Challenge received without HMAC signature")
	}

	// Compute the challenge response
	responseHash := ComputeChallengeResponse(crc, secret)

	slog.Debug("Challenge response computed successfully", "crcLength", len(crc))
	return notification.ChallengeResponse{
		ResponseHash: responseHash,
	}, nil
}

// NotificationHandler provides utilities for handling incoming notification webhooks
// and challenge-response validation from OSDU notification service
type NotificationHandler interface {
	// ComputeChallengeResponse computes the response hash for the challenge-response validation
	// This implements the logic from HashingUtil.java it is the same as public static String hashString(String crc, String secret)
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
type notificationHandler struct {
	signatureService *utils.SignatureService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler() NotificationHandler {
	return &notificationHandler{
		signatureService: utils.NewSignatureService(),
	}
}

// ComputeChallengeResponse delegates to the package-level function
func (h *notificationHandler) ComputeChallengeResponse(crc, secret string) string {
	return ComputeChallengeResponse(crc, secret)
}

// HandleChallenge delegates to the package-level function
func (h *notificationHandler) HandleChallenge(crc, hmac, secret string) (notification.ChallengeResponse, error) {
	return HandleChallenge(crc, hmac, secret)
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
// This is a reference implementation - clients should implement their own handlers for production use
type NotificationWebhookHandler struct {
	handler NotificationHandler
	secret  string
	logger  *slog.Logger
	// Optional callback to process received notifications
	// Returns (response interface{}, error)
	// If response is nil, a default NotificationHandlerResponse will be used
	// The response will be JSON encoded and sent back to the caller
	OnNotification func(*notification.NotificationMessage) (interface{}, error)
}

// NewNotificationWebhookHandler creates a new webhook handler
// This is a reference implementation. For production use, clients should:
// 1. Use NotificationHandler directly for parsing/validation
// 2. Implement their own HTTP handlers with custom response logic
// 3. Handle errors and responses according to their requirements
func NewNotificationWebhookHandler(secret string, onNotification func(*notification.NotificationMessage) (interface{}, error)) *NotificationWebhookHandler {
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
// The hmac parameter is optional - if missing or invalid, a warning is logged
func (nwh *NotificationWebhookHandler) HandleChallengeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		nwh.logger.Warn("Invalid HTTP method for challenge", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	crc := r.URL.Query().Get("crc")
	hmac := r.URL.Query().Get("hmac")

	nwh.logger.Info("Handling challenge request", "crcLength", len(crc), "hmacLength", len(hmac))

	// Use the standalone function directly
	response, err := HandleChallenge(crc, hmac, nwh.secret)
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
	var response interface{}
	if nwh.OnNotification != nil {
		nwh.logger.Debug("Processing notification",
			"notificationId", msg.ID,
			"subject", msg.Subject)
		customResponse, err := nwh.OnNotification(msg)
		if err != nil {
			nwh.logger.Error("Failed to process notification",
				"error", err,
				"notificationId", msg.ID)
			http.Error(w, "Failed to process notification", http.StatusInternalServerError)
			return
		}
		response = customResponse
	}

	nwh.logger.Info("Notification processed successfully",
		"notificationId", msg.ID,
		"subject", msg.Subject,
		"recordCount", len(msg.RecordEvents))

	// Use custom response if provided, otherwise use default
	if response == nil {
		response = notification.NotificationHandlerResponse{
			NotificationID: msg.ID,
			Status:         "OK",
		}
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
