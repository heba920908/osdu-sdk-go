package v2_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/heba920908/osdu-sdk-go/pkg/config"
	"github.com/heba920908/osdu-sdk-go/pkg/models/info"
	"github.com/heba920908/osdu-sdk-go/pkg/models/notification"
	v2 "github.com/heba920908/osdu-sdk-go/pkg/osdu/v2"
)

func TestRegisterService_GetInfo(t *testing.T) {
	t.Run("successful GetInfo without auth", func(t *testing.T) {
		expectedVersionInfo := &info.VersionInfo{
			GroupID:    "org.opengroup.osdu",
			ArtifactID: "register",
			Version:    "1.0.0",
			BuildTime:  "2025-11-25T12:00:00Z",
			Branch:     "main",
			CommitID:   "abc123",
		}

		responseBody, _ := json.Marshal(expectedVersionInfo)

		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				// Verify the request
				if req.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", req.Method)
				}

				expectedURL := "https://test.osdu.com/api/register/v1/info"
				if req.URL.String() != expectedURL {
					t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
				}

				// Verify no authorization header (GetInfo should not require auth)
				if req.Header.Get("Authorization") != "" {
					t.Error("Expected no Authorization header for GetInfo")
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			RegisterUrl: "https://test.osdu.com/api/register/v1",
		}

		client, err := v2.NewClient(
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
			v2.WithContext(context.Background()),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		versionInfo, err := client.Register().GetInfo()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if versionInfo == nil {
			t.Fatal("Expected version info, got nil")
		}

		if versionInfo.GroupID != "org.opengroup.osdu" {
			t.Errorf("Expected groupId 'org.opengroup.osdu', got '%s'", versionInfo.GroupID)
		}

		if versionInfo.ArtifactID != "register" {
			t.Errorf("Expected artifactId 'register', got '%s'", versionInfo.ArtifactID)
		}

		if versionInfo.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", versionInfo.Version)
		}
	})

	t.Run("GetInfo with error response", func(t *testing.T) {
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(bytes.NewBufferString(`{"error":"internal server error"}`)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			RegisterUrl: "https://test.osdu.com/api/register/v1",
		}

		client, err := v2.NewClient(
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		_, err = client.Register().GetInfo()
		if err == nil {
			t.Error("Expected error for 500 status code")
		}
	})

	t.Run("GetInfo with invalid JSON", func(t *testing.T) {
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(`invalid json`)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			RegisterUrl: "https://test.osdu.com/api/register/v1",
		}

		client, err := v2.NewClient(
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		_, err = client.Register().GetInfo()
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestRegisterService_GetTopics(t *testing.T) {
	t.Run("successful GetTopics with auth", func(t *testing.T) {
		expectedTopics := []notification.Topic{
			{
				Name:        "records-changed",
				Description: "This notification is sent whenever a new record or record version is created",
				State:       "ACTIVE",
			},
			{
				Name:        "schema-changed",
				Description: "This notification is sent whenever a new schema is created",
				State:       "ACTIVE",
			},
		}

		responseBody, _ := json.Marshal(expectedTopics)

		mockAuth := &mockAuthProvider{}
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				// Verify the request
				if req.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", req.Method)
				}

				expectedURL := "https://test.osdu.com/api/register/v1/topics"
				if req.URL.String() != expectedURL {
					t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
				}

				// Verify authorization header is present (GetTopics requires auth)
				if req.Header.Get("Authorization") == "" {
					t.Error("Expected Authorization header for GetTopics")
				}

				// Verify partition header
				if req.Header.Get("data-partition-id") == "" {
					t.Error("Expected data-partition-id header for GetTopics")
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			RegisterUrl: "https://test.osdu.com/api/register/v1",
			PartitionId: "test-partition",
		}

		client, err := v2.NewClient(
			v2.WithAuthProvider(mockAuth),
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
			v2.WithContext(context.Background()),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		topics, err := client.Register().GetTopics()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(topics) != 2 {
			t.Errorf("Expected 2 topics, got %d", len(topics))
		}

		if topics[0].Name != "records-changed" {
			t.Errorf("Expected first topic name 'records-changed', got '%s'", topics[0].Name)
		}

		if topics[1].Name != "schema-changed" {
			t.Errorf("Expected second topic name 'schema-changed', got '%s'", topics[1].Name)
		}
	})

	t.Run("GetTopics with error response", func(t *testing.T) {
		mockAuth := &mockAuthProvider{}
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 500,
					Body:       io.NopCloser(bytes.NewBufferString(`{"error":"internal server error"}`)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			RegisterUrl: "https://test.osdu.com/api/register/v1",
			PartitionId: "test-partition",
		}

		client, err := v2.NewClient(
			v2.WithAuthProvider(mockAuth),
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		_, err = client.Register().GetTopics()
		if err == nil {
			t.Error("Expected error for 500 status code")
		}
	})

	t.Run("GetTopics with invalid JSON", func(t *testing.T) {
		mockAuth := &mockAuthProvider{}
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(`invalid json`)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			RegisterUrl: "https://test.osdu.com/api/register/v1",
			PartitionId: "test-partition",
		}

		client, err := v2.NewClient(
			v2.WithAuthProvider(mockAuth),
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		_, err = client.Register().GetTopics()
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestRegisterService_CreateSubscription(t *testing.T) {
	t.Run("successful create subscription", func(t *testing.T) {
		inputSub := &notification.Subscription{
			Name:         "test-subscription",
			Description:  "Test subscription",
			Topic:        "records-changed",
			PushEndpoint: "https://test.com/webhook",
		}

		expectedResponse := &notification.Subscription{
			ID:           "sub-123",
			Name:         "test-subscription",
			Description:  "Test subscription",
			Topic:        "records-changed",
			PushEndpoint: "https://test.com/webhook",
			CreatedBy:    "user@test.com",
		}

		responseBody, _ := json.Marshal(expectedResponse)

		mockAuth := &mockAuthProvider{}
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				// Verify method and URL
				if req.Method != http.MethodPost {
					t.Errorf("Expected POST method, got %s", req.Method)
				}

				// Verify authorization header is present
				if req.Header.Get("Authorization") == "" {
					t.Error("Expected Authorization header for CreateSubscription")
				}

				// Verify partition header
				if req.Header.Get("data-partition-id") == "" {
					t.Error("Expected data-partition-id header for CreateSubscription")
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			RegisterUrl: "https://test.osdu.com/api/register/v1",
			PartitionId: "test-partition",
		}

		client, err := v2.NewClient(
			v2.WithAuthProvider(mockAuth),
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		result, err := client.Register().CreateSubscription(inputSub)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ID != "sub-123" {
			t.Errorf("Expected subscription ID 'sub-123', got '%s'", result.ID)
		}

		if result.CreatedBy != "user@test.com" {
			t.Errorf("Expected created by 'user@test.com', got '%s'", result.CreatedBy)
		}
	})
}

func TestRegisterService_GetSubscription(t *testing.T) {
	t.Run("successful get subscription", func(t *testing.T) {
		expectedSub := &notification.Subscription{
			ID:           "sub-123",
			Name:         "test-subscription",
			Description:  "Test subscription",
			Topic:        "records-changed",
			PushEndpoint: "https://test.com/webhook",
		}

		responseBody, _ := json.Marshal(expectedSub)

		mockAuth := &mockAuthProvider{}
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", req.Method)
				}

				expectedURL := "https://test.osdu.com/api/register/v1/subscription/sub-123"
				if req.URL.String() != expectedURL {
					t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			RegisterUrl: "https://test.osdu.com/api/register/v1",
			PartitionId: "test-partition",
		}

		client, err := v2.NewClient(
			v2.WithAuthProvider(mockAuth),
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		result, err := client.Register().GetSubscription("sub-123")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("Expected result to not be nil")
		}

		if result.ID != "sub-123" {
			t.Errorf("Expected subscription ID 'sub-123', got '%s'", result.ID)
		}
	})

	t.Run("subscription not found", func(t *testing.T) {
		mockAuth := &mockAuthProvider{}
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(bytes.NewBufferString(`{"error":"not found"}`)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			RegisterUrl: "https://test.osdu.com/api/register/v1",
			PartitionId: "test-partition",
		}

		client, err := v2.NewClient(
			v2.WithAuthProvider(mockAuth),
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		result, err := client.Register().GetSubscription("nonexistent")
		if err != nil {
			t.Fatalf("Expected no error for 404, got %v", err)
		}

		if result != nil {
			t.Error("Expected nil result for 404")
		}
	})
}

func TestRegisterService_DeleteSubscription(t *testing.T) {
	t.Run("successful delete subscription", func(t *testing.T) {
		mockAuth := &mockAuthProvider{}
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodDelete {
					t.Errorf("Expected DELETE method, got %s", req.Method)
				}

				expectedURL := "https://test.osdu.com/api/register/v1/subscription/sub-123"
				if req.URL.String() != expectedURL {
					t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
				}

				return &http.Response{
					StatusCode: 204,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			RegisterUrl: "https://test.osdu.com/api/register/v1",
			PartitionId: "test-partition",
		}

		client, err := v2.NewClient(
			v2.WithAuthProvider(mockAuth),
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		err = client.Register().DeleteSubscription("sub-123")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
	})
}

func TestRegisterService_QuerySubscriptions(t *testing.T) {
	t.Run("successful query subscriptions", func(t *testing.T) {
		expectedSubs := []notification.Subscription{
			{
				ID:    "sub-123",
				Name:  "test-subscription-1",
				Topic: "records-changed",
			},
			{
				ID:    "sub-456",
				Name:  "test-subscription-2",
				Topic: "records-changed",
			},
		}

		responseBody, _ := json.Marshal(expectedSubs)

		mockAuth := &mockAuthProvider{}
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Errorf("Expected GET method, got %s", req.Method)
				}

				expectedURL := "https://test.osdu.com/api/register/v1/subscription?notificationId=test-notification"
				if req.URL.String() != expectedURL {
					t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			RegisterUrl: "https://test.osdu.com/api/register/v1",
			PartitionId: "test-partition",
		}

		client, err := v2.NewClient(
			v2.WithAuthProvider(mockAuth),
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		results, err := client.Register().QuerySubscriptions("test-notification")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 subscriptions, got %d", len(results))
		}

		if results[0].ID != "sub-123" {
			t.Errorf("Expected first subscription ID 'sub-123', got '%s'", results[0].ID)
		}

		if results[1].ID != "sub-456" {
			t.Errorf("Expected second subscription ID 'sub-456', got '%s'", results[1].ID)
		}
	})
}
