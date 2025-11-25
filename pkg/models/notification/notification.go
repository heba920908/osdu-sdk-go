package notification

import (
	"encoding/json"
)

// Topic represents a notification topic with its metadata and examples
type Topic struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	State       string          `json:"state"`
	Example     json.RawMessage `json:"example,omitempty"`
}

// Subscription represents a notification subscription
type Subscription struct {
	Name           string  `json:"name,omitempty"`
	Description    string  `json:"description,omitempty"`
	Topic          string  `json:"topic,omitempty"`
	PushEndpoint   string  `json:"pushEndpoint,omitempty"`
	NotificationID string  `json:"notificationId,omitempty"`
	ID             string  `json:"id,omitempty"`
	CreatedBy      string  `json:"createdBy,omitempty"`
	Secret         *Secret `json:"secret,omitempty"`
}

// SubscriptionInfo represents subscription information without secret details
type SubscriptionInfo struct {
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	Topic          string `json:"topic,omitempty"`
	PushEndpoint   string `json:"pushEndpoint,omitempty"`
	NotificationID string `json:"notificationId,omitempty"`
	ID             string `json:"id,omitempty"`
	CreatedBy      string `json:"createdBy,omitempty"`
}

// Secret represents an abstract secret with a type discriminator
type Secret struct {
	SecretType string          `json:"secretType"`
	Value      json.RawMessage `json:"value,omitempty"`
}

// HmacSecret represents an HMAC secret
type HmacSecret struct {
	SecretType string `json:"secretType"`
	Value      string `json:"value"`
}

// GsaSecret represents a GSA (Google Service Account) secret
type GsaSecret struct {
	SecretType string          `json:"secretType"`
	Value      *GsaSecretValue `json:"value"`
}

// GsaSecretValue represents the value of a GSA secret
type GsaSecretValue struct {
	Audience string `json:"audience"`
	Key      string `json:"key"`
}

// Constants for secret types
const (
	SecretTypeHMAC = "HMAC"
	SecretTypeGSA  = "GSA"
)

// Notification event models based on OSDU notification topics

// RecordsChangedEvent represents a records-changed notification event
type RecordsChangedEvent struct {
	ID            string `json:"id,omitempty"`
	Kind          string `json:"kind"`
	Op            string `json:"op"` // create, create_schema, delete, update
	RecordUpdated string `json:"recordUpdated,omitempty"`
}

// SchemaChangedEvent represents a schema-changed notification event
type SchemaChangedEvent struct {
	Kind string `json:"kind"`
	Op   string `json:"op"` // create, update
}

// StatusChangedEvent represents a status-changed notification event
type StatusChangedEvent struct {
	Kind       string                 `json:"kind"`
	Properties map[string]interface{} `json:"properties"`
}

// DataSetDetails represents dataset details in a status-changed event
type DataSetDetails struct {
	CorrelationID    string `json:"correlationId"`
	DataSetID        string `json:"dataSetId"`
	DataSetVersionID string `json:"dataSetVersionId"`
	DataSetType      string `json:"dataSetType"`
	RecordCount      int    `json:"recordCount"`
	Timestamp        int64  `json:"timestamp"`
}

// StatusDetails represents status details in a status-changed event
type StatusDetails struct {
	CorrelationID   string `json:"correlationId"`
	RecordID        string `json:"recordId"`
	RecordIDVersion string `json:"recordIdVersion"`
	Stage           string `json:"stage"`
	Status          string `json:"status"`
	Message         string `json:"message"`
	ErrorCode       int    `json:"errorCode"`
	Timestamp       int64  `json:"timestamp"`
}

// LegalTagsChangedEvent represents a legaltags-changed notification event
type LegalTagsChangedEvent struct {
	StatusChangedTags []LegalTagStatus `json:"statusChangedTags"`
}

// LegalTagStatus represents the status of a changed legal tag
type LegalTagStatus struct {
	ChangedTagName   string `json:"changedTagName"`
	ChangedTagStatus string `json:"changedTagStatus"` // compliant, incompliant
}

// Helper functions

// NewHmacSecret creates a new HMAC secret
func NewHmacSecret(value string) *HmacSecret {
	return &HmacSecret{
		SecretType: SecretTypeHMAC,
		Value:      value,
	}
}

// NewGsaSecret creates a new GSA secret
func NewGsaSecret(audience, key string) *GsaSecret {
	return &GsaSecret{
		SecretType: SecretTypeGSA,
		Value: &GsaSecretValue{
			Audience: audience,
			Key:      key,
		},
	}
}

// ToSubscriptionInfo converts a Subscription to SubscriptionInfo
func (s *Subscription) ToSubscriptionInfo() *SubscriptionInfo {
	return &SubscriptionInfo{
		Name:           s.Name,
		Description:    s.Description,
		Topic:          s.Topic,
		PushEndpoint:   s.PushEndpoint,
		NotificationID: s.NotificationID,
		ID:             s.ID,
		CreatedBy:      s.CreatedBy,
	}
}

// Webhook and Handler models

// ChallengeResponse represents the response to a challenge request
type ChallengeResponse struct {
	ResponseHash string `json:"responseHash"`
}

// NotificationMessage represents the structure of an incoming notification webhook
// This supports both legacy (pre-M18) and current (M18+) formats
type NotificationMessage struct {
	ID           string                  `json:"id,omitempty"`
	EventTime    string                  `json:"eventTime,omitempty"`
	Subject      string                  `json:"subject,omitempty"`
	Topic        string                  `json:"topic,omitempty"`
	Data         NotificationMessageData `json:"data,omitempty"`
	RecordEvents []NotificationRecord    `json:"-"` // For M18+ format (direct array)
}

// NotificationMessageData contains the actual notification data
type NotificationMessageData struct {
	DataPartitionID string               `json:"dataPartitionId,omitempty"`
	Records         []NotificationRecord `json:"data,omitempty"`
}

// NotificationRecord represents a single record change notification in webhook payload
type NotificationRecord struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Op   string `json:"op"` // "create", "update", "delete"
}

// NotificationHandlerResponse represents the response to send back after processing notification
type NotificationHandlerResponse struct {
	NotificationID string `json:"notificationId"`
	Status         string `json:"status"`
}
