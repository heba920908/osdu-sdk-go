package v2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/heba920908/osdu-sdk-go/pkg/models/info"
	"github.com/heba920908/osdu-sdk-go/pkg/models/notification"
)

// RegisterService defines the interface for register operations
type RegisterService interface {
	// GetInfo retrieves version information without authorization
	GetInfo() (*info.VersionInfo, error)

	// GetTopics retrieves information about available notification topics with authorization
	GetTopics() ([]notification.Topic, error)

	// CreateSubscription creates a new notification subscription
	CreateSubscription(sub *notification.Subscription) (*notification.Subscription, error)

	// GetSubscription retrieves a subscription by ID
	GetSubscription(subscriptionID string) (*notification.SubscriptionInfo, error)

	// DeleteSubscription deletes a subscription by ID
	DeleteSubscription(subscriptionID string) error

	// QuerySubscriptions queries subscriptions by notification ID
	QuerySubscriptions(notificationID string) ([]notification.Subscription, error)
}

// registerService implements the RegisterService interface
type registerService struct {
	client *Client
}

// GetInfo retrieves version information
// This endpoint does not require authorization
func (s *registerService) GetInfo() (*info.VersionInfo, error) {
	url := fmt.Sprintf("%s/info", s.client.osduSettings.RegisterUrl)
	s.client.logger.Debug("Getting register info", "url", url)

	resp, err := s.client.doRequest(http.MethodGet, url, nil, false, false)
	if err != nil {
		s.client.logger.Error("Failed to get register info", "error", err)
		return nil, fmt.Errorf("failed to get register info: %w", err)
	}

	body, err := readResponse(resp)
	if err != nil {
		s.client.logger.Error("Failed to read register info response", "error", err, "status", resp.StatusCode)
		return nil, fmt.Errorf("failed to read register info response: %w", err)
	}

	var versionInfo info.VersionInfo
	if err := json.Unmarshal(body, &versionInfo); err != nil {
		s.client.logger.Error("Failed to unmarshal version info", "error", err)
		return nil, fmt.Errorf("failed to unmarshal version info: %w", err)
	}

	s.client.logger.Info("Retrieved register info",
		"version", versionInfo.Version,
		"artifactId", versionInfo.ArtifactID)
	return &versionInfo, nil
}

// GetTopics retrieves information about available notification topics
// This endpoint requires authorization
func (s *registerService) GetTopics() ([]notification.Topic, error) {
	url := fmt.Sprintf("%s/topics", s.client.osduSettings.RegisterUrl)
	s.client.logger.Debug("Getting notification topics", "url", url)

	resp, err := s.client.doRequest(http.MethodGet, url, nil, true, true)
	if err != nil {
		s.client.logger.Error("Failed to get topics", "error", err)
		return nil, fmt.Errorf("failed to get topics: %w", err)
	}

	body, err := readResponse(resp)
	if err != nil {
		s.client.logger.Error("Failed to read topics response", "error", err, "status", resp.StatusCode)
		return nil, fmt.Errorf("failed to read topics response: %w", err)
	}

	var topics []notification.Topic
	if err := json.Unmarshal(body, &topics); err != nil {
		s.client.logger.Error("Failed to unmarshal topics", "error", err)
		return nil, fmt.Errorf("failed to unmarshal topics: %w", err)
	}

	s.client.logger.Info("Retrieved notification topics", "count", len(topics))
	return topics, nil
}

// CreateSubscription creates a new notification subscription
func (s *registerService) CreateSubscription(sub *notification.Subscription) (*notification.Subscription, error) {
	url := fmt.Sprintf("%s/subscription", s.client.osduSettings.RegisterUrl)
	s.client.logger.Debug("Creating subscription",
		"url", url,
		"name", sub.Name,
		"topic", sub.Topic)

	reqBody, err := json.Marshal(sub)
	if err != nil {
		s.client.logger.Error("Failed to marshal subscription", "error", err)
		return nil, fmt.Errorf("failed to marshal subscription: %w", err)
	}

	resp, err := s.client.doRequest(http.MethodPost, url, reqBody, true, true)
	if err != nil {
		s.client.logger.Error("Failed to create subscription", "error", err)
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	body, err := readResponse(resp)
	if err != nil {
		s.client.logger.Error("Failed to read create subscription response", "error", err, "status", resp.StatusCode)
		return nil, fmt.Errorf("failed to read create subscription response: %w", err)
	}

	var result notification.Subscription
	if err := json.Unmarshal(body, &result); err != nil {
		s.client.logger.Error("Failed to unmarshal subscription", "error", err)
		return nil, fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	s.client.logger.Info("Created subscription",
		"id", result.ID,
		"name", result.Name)
	return &result, nil
}

// GetSubscription retrieves a subscription by ID
func (s *registerService) GetSubscription(subscriptionID string) (*notification.SubscriptionInfo, error) {
	url := fmt.Sprintf("%s/subscription/%s", s.client.osduSettings.RegisterUrl, subscriptionID)
	s.client.logger.Debug("Getting subscription", "url", url, "subscriptionId", subscriptionID)

	resp, err := s.client.doRequest(http.MethodGet, url, nil, true, true)
	if err != nil {
		s.client.logger.Error("Failed to get subscription", "error", err, "subscriptionId", subscriptionID)
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Handle 404 as not found
	if resp.StatusCode == http.StatusNotFound {
		s.client.logger.Debug("Subscription not found", "subscriptionId", subscriptionID)
		return nil, nil
	}

	body, err := readResponse(resp)
	if err != nil {
		s.client.logger.Error("Failed to read get subscription response", "error", err, "status", resp.StatusCode)
		return nil, fmt.Errorf("failed to read get subscription response: %w", err)
	}

	var sub notification.Subscription
	if err := json.Unmarshal(body, &sub); err != nil {
		s.client.logger.Error("Failed to unmarshal subscription", "error", err)
		return nil, fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	s.client.logger.Info("Retrieved subscription",
		"id", sub.ID,
		"name", sub.Name)
	return sub.ToSubscriptionInfo(), nil
}

// DeleteSubscription deletes a subscription by ID
func (s *registerService) DeleteSubscription(subscriptionID string) error {
	url := fmt.Sprintf("%s/subscription/%s", s.client.osduSettings.RegisterUrl, subscriptionID)
	s.client.logger.Debug("Deleting subscription", "url", url, "subscriptionId", subscriptionID)

	resp, err := s.client.doRequest(http.MethodDelete, url, nil, true, true)
	if err != nil {
		s.client.logger.Error("Failed to delete subscription", "error", err, "subscriptionId", subscriptionID)
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	_, err = readResponse(resp)
	if err != nil {
		s.client.logger.Error("Failed to read delete subscription response", "error", err, "status", resp.StatusCode)
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	s.client.logger.Info("Deleted subscription", "subscriptionId", subscriptionID)
	return nil
}

// QuerySubscriptions queries subscriptions by notification ID
func (s *registerService) QuerySubscriptions(notificationID string) ([]notification.Subscription, error) {
	url := fmt.Sprintf("%s/subscription?notificationId=%s",
		s.client.osduSettings.RegisterUrl, notificationID)
	s.client.logger.Debug("Querying subscriptions", "url", url, "notificationId", notificationID)

	resp, err := s.client.doRequest(http.MethodGet, url, nil, true, true)
	if err != nil {
		s.client.logger.Error("Failed to query subscriptions", "error", err, "notificationId", notificationID)
		return nil, fmt.Errorf("failed to query subscriptions: %w", err)
	}

	body, err := readResponse(resp)
	if err != nil {
		s.client.logger.Error("Failed to read query subscriptions response", "error", err, "status", resp.StatusCode)
		return nil, fmt.Errorf("failed to read query subscriptions response: %w", err)
	}

	var subscriptions []notification.Subscription
	if err := json.Unmarshal(body, &subscriptions); err != nil {
		s.client.logger.Error("Failed to unmarshal subscriptions", "error", err)
		return nil, fmt.Errorf("failed to unmarshal subscriptions: %w", err)
	}

	s.client.logger.Info("Queried subscriptions",
		"notificationId", notificationID,
		"count", len(subscriptions))
	return subscriptions, nil
}
