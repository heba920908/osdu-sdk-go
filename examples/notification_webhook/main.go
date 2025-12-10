package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/heba920908/osdu-sdk-go/pkg/models/notification"
	v2 "github.com/heba920908/osdu-sdk-go/pkg/osdu/v2"
)

func main() {
	// Setup structured logger with JSON output
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Get the secret from environment variable (or configuration)
	secret := os.Getenv("NOTIFICATION_SECRET")
	if secret == "" {
		secret = "default-secret-for-development"
		logger.Warn("Using default secret", "reason", "NOTIFICATION_SECRET env var not set")
	}

	// Define a callback to process received notifications
	onNotificationReceived := func(msg *notification.NotificationMessage) (interface{}, error) {
		logger.Info("Processing notification",
			"notificationId", msg.ID,
			"subject", msg.Subject,
			"dataPartitionId", msg.Data.DataPartitionID,
			"recordCount", len(msg.Data.Records))

		// Process the notification records
		for i, record := range msg.Data.Records {
			logger.Debug("Processing record",
				"index", i+1,
				"recordId", record.ID,
				"kind", record.Kind,
				"operation", record.Op)
		}

		// Here you can add your custom processing logic:
		// - Store in database
		// - Trigger workflows
		// - Send to message queue
		// - Update cache
		// etc.

		// Return nil to use the default response, or return a custom response structure
		// Example custom response:
		// return map[string]interface{}{
		//     "status": "PROCESSED",
		//     "message": "Successfully processed notification",
		//     "recordCount": len(msg.Data.Records),
		// }, nil

		return nil, nil
	}

	// Create the webhook handler with custom logger
	webhookHandler := v2.NewNotificationWebhookHandler(secret, onNotificationReceived).
		WithLogger(logger)

	// Setup HTTP routes
	http.HandleFunc("/health", webhookHandler.HandleHealthHTTP)
	http.HandleFunc("/webhook/challenge", webhookHandler.HandleChallengeHTTP)
	http.HandleFunc("/webhook/notification", webhookHandler.HandleNotificationHTTP)

	// For convenience, handle both challenge and notification on the same endpoint
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Challenge request
			webhookHandler.HandleChallengeHTTP(w, r)
		case http.MethodPost:
			// Notification request
			webhookHandler.HandleNotificationHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting notification webhook server on %s", addr)
	log.Printf("Endpoints:")
	log.Printf("  - Health check: GET http://localhost%s/health", addr)
	log.Printf("  - Challenge: GET http://localhost%s/webhook?crc=xxx&hmac=yyy", addr)
	log.Printf("  - Notification: POST http://localhost%s/webhook", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// Example usage with custom router (e.g., gorilla/mux):
/*
func setupWithMux() {
	secret := os.Getenv("NOTIFICATION_SECRET")

	onNotification := func(msg *notification.NotificationMessage) (interface{}, error) {
		// Process notification
		// Return nil, nil to use default response
		// Or return custom response structure
		return nil, nil
	}

	webhookHandler := v2.NewNotificationWebhookHandler(secret, onNotification)

	r := mux.NewRouter()
	r.HandleFunc("/health", webhookHandler.HandleHealthHTTP).Methods("GET")
	r.HandleFunc("/webhook", webhookHandler.HandleChallengeHTTP).Methods("GET")
	r.HandleFunc("/webhook", webhookHandler.HandleNotificationHTTP).Methods("POST")

	http.ListenAndServe(":8080", r)
}
*/

// Example of using standalone functions (without creating NotificationHandler):
/*
func testStandaloneFunctions() {
	// Compute challenge response directly
	crc := "12345"
	secret := "my-secret"
	responseHash := v2.ComputeChallengeResponse(crc, secret)
	fmt.Printf("Response hash: %s\n", responseHash)

	// Handle full challenge directly
	response, err := v2.HandleChallenge(crc, "hmac-value", secret)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Challenge response: %+v\n", response)
}
*/

// Example of using the NotificationHandler interface:
/*
func testChallengeResponse() {
	handler := v2.NewNotificationHandler()

	// Compute challenge response
	crc := "12345"
	secret := "my-secret"
	responseHash := handler.ComputeChallengeResponse(crc, secret)
	fmt.Printf("Response hash: %s\n", responseHash)

	// Handle full challenge
	response, err := handler.HandleChallenge(crc, "hmac-value", secret)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Challenge response: %+v\n", response)
}
*/

// Example of parsing notification payloads:
/*
func testParsePayload() {
	handler := v2.NewNotificationHandler()

	// M18+ format (direct array)
	m18Payload := []byte(`[
		{
			"id": "opendes:dataset:1.0.0",
			"kind": "opendes:wks:dataset:1.0.0",
			"op": "create"
		}
	]`)

	msg, err := handler.ParseNotificationPayload(m18Payload)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parsed M18+ message: %+v\n", msg)

	// Legacy format (full message)
	legacyPayload := []byte(`{
		"id": "msg-123",
		"eventTime": "2025-11-25T12:00:00Z",
		"subject": "RecordsChanged",
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
	}`)

	msg, err = handler.ParseNotificationPayload(legacyPayload)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parsed legacy message: %+v\n", msg)
}
*/
