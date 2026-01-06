package v2_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/heba920908/osdu-sdk-go/pkg/config"
	"github.com/heba920908/osdu-sdk-go/pkg/models/dataset"
	"github.com/heba920908/osdu-sdk-go/pkg/models/info"
	v2 "github.com/heba920908/osdu-sdk-go/pkg/osdu/v2"
)

func TestDatasetService_GetInfo(t *testing.T) {
	t.Run("successful GetInfo without auth", func(t *testing.T) {
		expectedVersionInfo := &info.VersionInfo{
			GroupID:    "org.opengroup.osdu",
			ArtifactID: "dataset",
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

				expectedURL := "https://test.osdu.com/api/dataset/v1/info"
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
			DatasetUrl:  "https://test.osdu.com/api/dataset/v1",
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

		versionInfo, err := client.Dataset().GetInfo()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if versionInfo == nil {
			t.Fatal("Expected version info, got nil")
		}

		if versionInfo.GroupID != "org.opengroup.osdu" {
			t.Errorf("Expected groupId 'org.opengroup.osdu', got '%s'", versionInfo.GroupID)
		}

		if versionInfo.ArtifactID != "dataset" {
			t.Errorf("Expected artifactId 'dataset', got '%s'", versionInfo.ArtifactID)
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
			DatasetUrl:  "https://test.osdu.com/api/dataset/v1",
			RegisterUrl: "https://test.osdu.com/api/register/v1",
		}

		client, err := v2.NewClient(
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		_, err = client.Dataset().GetInfo()
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
			DatasetUrl:  "https://test.osdu.com/api/dataset/v1",
			RegisterUrl: "https://test.osdu.com/api/register/v1",
		}

		client, err := v2.NewClient(
			v2.WithHTTPClient(mockHTTP),
			v2.WithOsduSettings(osduSettings),
		)
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		_, err = client.Dataset().GetInfo()
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestDatasetService_GetStorageInstructions(t *testing.T) {
	t.Run("successful GetStorageInstructions with auth", func(t *testing.T) {
		expectedResponse := &dataset.DatasetStorageInstructionsResponse{
			ProviderKey: "azure",
			StorageLocation: dataset.DatasetStorageLocation{
				SignedUrl:  "https://storage.blob.core.windows.net/container/file?sig=xxx",
				FileSource: "osdu://opendes/data/test-file.csv",
				CreatedBy:  "user@test.com",
				ExpiryTime: "2025-12-01T12:00:00Z",
			},
		}

		responseBody, _ := json.Marshal(expectedResponse)

		mockAuth := &mockAuthProvider{}
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				// Verify the request
				if req.Method != http.MethodPost {
					t.Errorf("Expected POST method, got %s", req.Method)
				}

				expectedURL := "https://test.osdu.com/api/dataset/v1/getStorageInstructions?kindSubType=dataset--File.Generic"
				if req.URL.String() != expectedURL {
					t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
				}

				// Verify authorization header is present
				if req.Header.Get("Authorization") == "" {
					t.Error("Expected Authorization header for GetStorageInstructions")
				}

				// Verify partition header
				if req.Header.Get("data-partition-id") == "" {
					t.Error("Expected data-partition-id header for GetStorageInstructions")
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			DatasetUrl:  "https://test.osdu.com/api/dataset/v1",
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

		result, err := client.Dataset().GetStorageInstructions("dataset--File.Generic")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.ProviderKey != "azure" {
			t.Errorf("Expected providerKey 'azure', got '%s'", result.ProviderKey)
		}

		if result.StorageLocation.SignedUrl == "" {
			t.Error("Expected signedUrl to be non-empty")
		}

		if result.StorageLocation.CreatedBy != "user@test.com" {
			t.Errorf("Expected createdBy 'user@test.com', got '%s'", result.StorageLocation.CreatedBy)
		}
	})

	t.Run("GetStorageInstructions with error response", func(t *testing.T) {
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
			DatasetUrl:  "https://test.osdu.com/api/dataset/v1",
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

		_, err = client.Dataset().GetStorageInstructions("dataset--File.Generic")
		if err == nil {
			t.Error("Expected error for 500 status code")
		}
	})

	t.Run("GetStorageInstructions with invalid JSON", func(t *testing.T) {
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
			DatasetUrl:  "https://test.osdu.com/api/dataset/v1",
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

		_, err = client.Dataset().GetStorageInstructions("dataset--File.Generic")
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestDatasetService_GetRetrievalInstructions(t *testing.T) {
	t.Run("successful GetRetrievalInstructions with auth", func(t *testing.T) {
		expectedResponse := &dataset.DatasetRetrievalInstructionsResponse{
			Datasets: []dataset.Dataset{
				{
					DatasetRegistryId: "opendes:dataset--File.Generic:test-file-1",
					RetrievalProperties: dataset.DatasetStorageLocation{
						SignedUrl:  "https://storage.blob.core.windows.net/container/file1?sig=xxx",
						CreatedBy:  "user@test.com",
						ExpiryTime: "2025-12-01T12:00:00Z",
					},
					ProviderKey: "azure",
				},
				{
					DatasetRegistryId: "opendes:dataset--File.Generic:test-file-2",
					RetrievalProperties: dataset.DatasetStorageLocation{
						SignedUrl:  "https://storage.blob.core.windows.net/container/file2?sig=yyy",
						CreatedBy:  "user@test.com",
						ExpiryTime: "2025-12-01T12:00:00Z",
					},
					ProviderKey: "azure",
				},
			},
		}

		responseBody, _ := json.Marshal(expectedResponse)

		mockAuth := &mockAuthProvider{}
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				// Verify the request
				if req.Method != http.MethodPost {
					t.Errorf("Expected POST method, got %s", req.Method)
				}

				expectedURL := "https://test.osdu.com/api/dataset/v1/getRetrievalInstructions"
				if req.URL.String() != expectedURL {
					t.Errorf("Expected URL %s, got %s", expectedURL, req.URL.String())
				}

				// Verify authorization header is present
				if req.Header.Get("Authorization") == "" {
					t.Error("Expected Authorization header for GetRetrievalInstructions")
				}

				// Verify partition header
				if req.Header.Get("data-partition-id") == "" {
					t.Error("Expected data-partition-id header for GetRetrievalInstructions")
				}

				// Verify request body
				body, _ := io.ReadAll(req.Body)
				var reqBody dataset.DatasetRetrievalRequest
				if err := json.Unmarshal(body, &reqBody); err != nil {
					t.Errorf("Failed to unmarshal request body: %v", err)
				}

				if len(reqBody.DatasetRegistryIds) != 2 {
					t.Errorf("Expected 2 dataset registry IDs, got %d", len(reqBody.DatasetRegistryIds))
				}

				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			DatasetUrl:  "https://test.osdu.com/api/dataset/v1",
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

		datasetIds := []string{
			"opendes:dataset--File.Generic:test-file-1",
			"opendes:dataset--File.Generic:test-file-2",
		}
		result, err := client.Dataset().GetRetrievalInstructions(datasetIds)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(result.Datasets) != 2 {
			t.Errorf("Expected 2 datasets, got %d", len(result.Datasets))
		}

		if result.Datasets[0].DatasetRegistryId != "opendes:dataset--File.Generic:test-file-1" {
			t.Errorf("Expected first dataset ID 'opendes:dataset--File.Generic:test-file-1', got '%s'", result.Datasets[0].DatasetRegistryId)
		}

		if result.Datasets[1].DatasetRegistryId != "opendes:dataset--File.Generic:test-file-2" {
			t.Errorf("Expected second dataset ID 'opendes:dataset--File.Generic:test-file-2', got '%s'", result.Datasets[1].DatasetRegistryId)
		}

		if result.Datasets[0].ProviderKey != "azure" {
			t.Errorf("Expected providerKey 'azure', got '%s'", result.Datasets[0].ProviderKey)
		}
	})

	t.Run("GetRetrievalInstructions with error response", func(t *testing.T) {
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
			DatasetUrl:  "https://test.osdu.com/api/dataset/v1",
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

		_, err = client.Dataset().GetRetrievalInstructions([]string{"test-id"})
		if err == nil {
			t.Error("Expected error for 500 status code")
		}
	})

	t.Run("GetRetrievalInstructions with invalid JSON", func(t *testing.T) {
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
			DatasetUrl:  "https://test.osdu.com/api/dataset/v1",
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

		_, err = client.Dataset().GetRetrievalInstructions([]string{"test-id"})
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("GetRetrievalInstructions with empty dataset IDs", func(t *testing.T) {
		expectedResponse := &dataset.DatasetRetrievalInstructionsResponse{
			Datasets: []dataset.Dataset{},
		}

		responseBody, _ := json.Marshal(expectedResponse)

		mockAuth := &mockAuthProvider{}
		mockHTTP := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
				}, nil
			},
		}

		osduSettings := config.OsduSettings{
			DatasetUrl:  "https://test.osdu.com/api/dataset/v1",
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

		result, err := client.Dataset().GetRetrievalInstructions([]string{})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(result.Datasets) != 0 {
			t.Errorf("Expected 0 datasets, got %d", len(result.Datasets))
		}
	})
}

