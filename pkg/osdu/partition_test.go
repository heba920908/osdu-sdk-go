package osdu_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/heba920908/osdu-sdk-go/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestMockPartitionRegistration(t *testing.T) {
	tests := []struct {
		name           string
		isSystem       bool
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    bool
		errorContains  string
	}{
		{
			name:     "successful partition registration",
			isSystem: false,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/partitions/")
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"message": "Partition created successfully"}`))
			},
			expectError: false,
		},
		{
			name:     "partition already exists - conflict handling",
			isSystem: false,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "POST" {
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte(`{"message": "Partition already exists"}`))
				} else if r.Method == "PATCH" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"message": "Partition updated successfully"}`))
				}
			},
			expectError: false,
		},
		{
			name:     "system partition registration",
			isSystem: true,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/partitions/system")
				w.WriteHeader(http.StatusCreated)
			},
			expectError: false,
		},
		{
			name:     "server error",
			isSystem: false,
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Internal server error"}`))
			},
			expectError:   true,
			errorContains: "partition service response - 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			partitionServer := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer partitionServer.Close()

			// Create mock client
			client, _ := createMockClient(partitionServer.URL, "http://mock-entitlements")

			// Create test partition using the models structure
			partitionProperties := models.GetDefaultPartitionPropertiesCI("test-partition-id")
			partition := models.Partition{
				Properties: partitionProperties,
			}

			// Execute test
			err := client.RegisterPartition(partition)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockPartitionWithOverride(t *testing.T) {
	partitionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)

		// Verify the partition override properties are included in the request
		var partition map[string]interface{}
		json.NewDecoder(r.Body).Decode(&partition)

		properties, ok := partition["properties"].(map[string]interface{})
		assert.True(t, ok)

		// Check for override properties
		minioEndpoint, exists := properties["obm.minio.endpoint"]
		assert.True(t, exists)

		endpointMap, ok := minioEndpoint.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "http://test-minion:9000", endpointMap["value"])
		assert.Equal(t, false, endpointMap["sensitive"])

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "Partition with overrides created"}`))
	}))
	defer partitionServer.Close()

	// Create mock client
	client, _ := createMockClient(partitionServer.URL, "http://mock-entitlements")

	// Create test partition with overrides using the proper model structure
	partitionProperties := models.GetDefaultPartitionPropertiesCI("test-partition-override")
	partition := models.Partition{
		Properties: partitionProperties,
	}

	// Apply overrides using JSON unmarshaling (simulating the original test)
	partitionOverride := `{
		"properties": {
			"obm.minio.endpoint": {
				"sensitive": false,
				"value": "http://test-minion:9000"
			},
			"obm.minio.external.endpoint": {
				"sensitive": false,
				"value": "https://s3-some-external.test-minion:9000"
			}
		}
	}`

	// Merge the override into the existing partition
	var override map[string]interface{}
	err := json.NewDecoder(strings.NewReader(partitionOverride)).Decode(&override)
	assert.NoError(t, err)

	// Convert partition to map for merging
	partitionJSON, _ := json.Marshal(partition)
	var partitionMap map[string]interface{}
	json.Unmarshal(partitionJSON, &partitionMap)

	// Merge properties
	if overrideProps, ok := override["properties"].(map[string]interface{}); ok {
		if partitionProps, ok := partitionMap["properties"].(map[string]interface{}); ok {
			for key, value := range overrideProps {
				partitionProps[key] = value
			}
		}
	}

	// Convert back to Partition struct
	finalJSON, _ := json.Marshal(partitionMap)
	var finalPartition models.Partition
	json.Unmarshal(finalJSON, &finalPartition)

	// Execute test
	err = client.RegisterPartition(finalPartition)
	assert.NoError(t, err)
}
