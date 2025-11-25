# OSDU Notification Webhook Handler

This package provides a complete implementation for handling OSDU notification webhooks, including challenge-response validation and notification payload processing.

## Overview

OSDU notifications use a **push model** where the notification service sends HTTP requests to your configured webhook endpoint. This is different from a pull model where you would make API calls to fetch notifications.

## Key Features

- ✅ Challenge-response validation (implements `HashingUtil.java` logic)
- ✅ Support for both M18+ and legacy notification formats
- ✅ HTTP handler functions compatible with any Go router
- ✅ Type-safe models in `pkg/models/notification`
- ✅ Comprehensive unit tests

## Architecture

### Models (`pkg/models/notification`)

All notification-related data structures are defined in the models package:

- `ChallengeResponse` - Response to challenge requests
- `NotificationMessage` - Incoming notification payload
- `NotificationMessageData` - Notification data container
- `NotificationRecord` - Individual record change
- `NotificationHandlerResponse` - Response after processing notification

### Handler (`pkg/osdu/v2/notification.go`)

The handler provides:

1. **NotificationHandler interface** - Core logic for challenge and payload parsing
2. **NotificationWebhookHandler** - HTTP handlers for webhooks

## Usage

### 1. Basic Setup

```go
import (
    "github.com/heba920908/osdu-sdk-go/pkg/models/notification"
    v2 "github.com/heba920908/osdu-sdk-go/pkg/osdu/v2"
)

// Define your notification processing logic
onNotification := func(msg *notification.NotificationMessage) error {
    log.Printf("Received %d records from partition %s", 
        len(msg.Data.Records), msg.Data.DataPartitionID)
    
    // Process records...
    for _, record := range msg.Data.Records {
        log.Printf("  %s: %s (%s)", record.Op, record.ID, record.Kind)
    }
    
    return nil
}

// Create webhook handler
secret := os.Getenv("NOTIFICATION_SECRET")
webhookHandler := v2.NewNotificationWebhookHandler(secret, onNotification)
```

### 2. Setup HTTP Routes

#### With net/http (standard library)

```go
http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        // Challenge validation
        webhookHandler.HandleChallengeHTTP(w, r)
    case http.MethodPost:
        // Notification processing
        webhookHandler.HandleNotificationHTTP(w, r)
    }
})

http.ListenAndServe(":8080", nil)
```

#### With gorilla/mux

```go
r := mux.NewRouter()
r.HandleFunc("/webhook", webhookHandler.HandleChallengeHTTP).Methods("GET")
r.HandleFunc("/webhook", webhookHandler.HandleNotificationHTTP).Methods("POST")
r.HandleFunc("/health", webhookHandler.HandleHealthHTTP).Methods("GET")

http.ListenAndServe(":8080", r)
```

### 3. Challenge-Response Flow

When you register a subscription with OSDU, it will validate your endpoint:

```
GET /webhook?crc=12345&hmac=abc123
```

Your handler will:
1. Extract `crc` and `hmac` parameters
2. Compute response hash: `Base64(SHA256(secret + crc))`
3. Return JSON response:

```json
{
    "responseHash": "computed_hash_value"
}
```

This implements the same logic as the Java `HashingUtil.java`:
```java
String response = secret + crc;
response = Hashing.sha256().hashString(response, StandardCharsets.UTF_8).toString();
return Base64.getEncoder().encodeToString(response.getBytes());
```

### 4. Notification Payload Processing

The handler supports both notification formats:

#### M18+ Format (Current)
Direct array of record changes:
```json
[
    {
        "id": "opendes:dataset:1.0.0",
        "kind": "opendes:wks:dataset:1.0.0",
        "op": "create"
    }
]
```

#### Legacy Format (Pre-M18)
Full message structure:
```json
{
    "id": "msg-123",
    "eventTime": "2025-11-25T12:00:00Z",
    "subject": "RecordsChanged",
    "topic": "records-changed",
    "data": {
        "dataPartitionId": "opendes",
        "data": [...]
    }
}
```

Both formats are automatically parsed and normalized to `NotificationMessage`.

## Testing

### Unit Tests

Run the comprehensive test suite:

```bash
go test ./pkg/osdu/v2/... -v -run "Notification|Challenge"
```

Tests cover:
- ✅ Challenge response computation
- ✅ Challenge handling with validation
- ✅ M18+ payload parsing
- ✅ Legacy payload parsing
- ✅ HTTP handlers for all endpoints
- ✅ Error handling

### Manual Testing

#### Test Challenge Endpoint

```bash
curl "http://localhost:8080/webhook?crc=12345&hmac=test"
```

Expected response:
```json
{
    "responseHash": "computed_hash"
}
```

#### Test Notification Endpoint

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '[{"id":"opendes:dataset:1.0.0","kind":"opendes:wks:dataset:1.0.0","op":"create"}]'
```

Expected response:
```json
{
    "notificationId": "opendes-1234567890",
    "status": "OK"
}
```

## Configuration

### Environment Variables

- `NOTIFICATION_SECRET` - Secret used for challenge validation (required)
- `PORT` - HTTP server port (default: 8080)

### OSDU Subscription Setup

To receive notifications, create a subscription in OSDU:

```go
subscription := &notification.Subscription{
    Name:         "my-webhook",
    Description:  "Webhook for record changes",
    Topic:        "records-changed",
    PushEndpoint: "https://my-domain.com/webhook",
    Secret: &notification.Secret{
        SecretType: notification.SecretTypeHMAC,
        Value:      json.RawMessage(`{"value":"your-secret"}`),
    },
}

client := v2.NewClient()
created, err := client.Register().CreateSubscription(subscription)
```

## Advanced Usage

### Using NotificationHandler Interface Directly

For custom implementations or testing:

```go
handler := v2.NewNotificationHandler()

// Compute challenge response
responseHash := handler.ComputeChallengeResponse(crc, secret)

// Parse payload
msg, err := handler.ParseNotificationPayload(payloadBytes)
```

### Custom Processing Logic

Implement complex workflows in your callback:

```go
onNotification := func(msg *notification.NotificationMessage) error {
    // Validate
    if msg.Data.DataPartitionID == "" {
        return fmt.Errorf("missing partition ID")
    }
    
    // Store in database
    if err := db.StoreNotification(msg); err != nil {
        return err
    }
    
    // Send to message queue
    if err := queue.Publish(msg); err != nil {
        return err
    }
    
    // Update cache
    cache.Invalidate(msg.Data.Records)
    
    return nil
}
```

## Example Application

See the complete example in `examples/notification_webhook/main.go`:

```bash
cd examples/notification_webhook
go run main.go
```

## References

- [OSDU Register Service](https://community.opengroup.org/osdu/platform/system/register)
- [HashingUtil.java](https://community.opengroup.org/osdu/platform/system/register/-/blob/v0.21.0/register-core/src/main/java/org/opengroup/osdu/register/utils/HashingUtil.java)
- [SignatureService.java](https://community.opengroup.org/osdu/platform/system/lib/core/os-core-common/-/blob/v0.21.0/src/main/java/org/opengroup/osdu/core/common/cryptographic/SignatureService.java)

## TODO

- [ ] Implement HMAC signature verification
- [ ] Add retry logic for failed notifications
- [ ] Support for additional secret types (GSA)
- [ ] Metrics and monitoring integration
