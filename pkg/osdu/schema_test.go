package osdu_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heba920908/osdu-sdk-go/pkg/auth"
	"github.com/heba920908/osdu-sdk-go/pkg/config"
	"github.com/heba920908/osdu-sdk-go/pkg/osdu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPutSystemSchema_Success(t *testing.T) {
	// Create a mock server that simulates successful schema upload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/schemas/system", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer mock-access-token")

		// Return successful response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	// Create mock client
	client, mockAuth := createMockSchemaClient(server.URL)

	// Create test schema payload
	schemaPayload := []byte(`{
		"id": "test-schema-123",
		"kind": "test",
		"schema": {
			"type": "object",
			"properties": {
				"id": {"type": "string"}
			}
		}
	}`)

	// Execute test
	err := client.PutSystemSchema(schemaPayload)

	// Verify results
	assert.NoError(t, err)
	mockAuth.AssertExpectations(t)
}

func TestPutSystemSchema_HttpError(t *testing.T) {
	// Create a mock server that simulates HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return error response
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	// Create mock client
	client, mockAuth := createMockSchemaClient(server.URL)

	// Create test schema payload
	schemaPayload := []byte(`{
		"id": "test-schema-456",
		"kind": "test",
		"schema": {
			"type": "object"
		}
	}`)

	// Execute test
	err := client.PutSystemSchema(schemaPayload)

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "[test-schema-456] schema unexpected status code: 500")
	mockAuth.AssertExpectations(t)
}

func TestPutSystemSchema_BadRequest_SchemaExists(t *testing.T) {
	// Create a mock server that simulates schema already exists (400)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return bad request (schema exists)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Schema already exists"}`))
	}))
	defer server.Close()

	// Create mock client
	client, mockAuth := createMockSchemaClient(server.URL)

	// Create test schema payload
	schemaPayload := []byte(`{
		"id": "existing-schema-789",
		"kind": "test",
		"schema": {
			"type": "object"
		}
	}`)

	// Execute test - should succeed (400 is treated as "already exists")
	err := client.PutSystemSchema(schemaPayload)

	// Verify results - no error because 400 is handled as "schema already exists"
	assert.NoError(t, err)
	mockAuth.AssertExpectations(t)
}

func TestPutSystemSchema_InvalidJSON(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create mock client
	client, mockAuth := createMockSchemaClient(server.URL)

	// Create invalid JSON payload
	schemaPayload := []byte(`{"invalid": json`)

	// Execute test - should still work because JSON parsing is only for logging
	err := client.PutSystemSchema(schemaPayload)

	// Verify results - should succeed even with invalid JSON (schema ID will be "unknown")
	assert.NoError(t, err)
	mockAuth.AssertExpectations(t)
}

func TestPutSystemSchema_AuthenticationFailure(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create mock auth that fails
	mockAuth := &MockAuthProvider{}
	mockAuth.On("GetAccessToken", mock.Anything).Return((*auth.Token)(nil), fmt.Errorf("authentication failed"))

	// Create test configuration
	osduSettings := config.OsduSettings{
		PartitionId: "test-partition",
		SchemaUrl:   server.URL,
	}

	// Create client with failing auth provider
	client := osdu.NewClientWithConfig(mockAuth, osduSettings)

	// Create test schema payload
	schemaPayload := []byte(`{
		"id": "test-schema-auth-fail",
		"kind": "test"
	}`)

	// Execute test
	err := client.PutSystemSchema(schemaPayload)

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
	mockAuth.AssertExpectations(t)
}

func TestPutSystemSchema_EmptyPayload(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify empty body
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create mock client
	client, mockAuth := createMockSchemaClient(server.URL)

	// Execute test with empty payload
	err := client.PutSystemSchema([]byte{})

	// Verify results - should work (schema ID will be "unknown")
	assert.NoError(t, err)
	mockAuth.AssertExpectations(t)
}

func TestPutSystemSchema_WithRetry(t *testing.T) {
	callCount := 0
	// Create a mock server that fails twice then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 2 {
			// Fail the first two attempts
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Temporary server error"}`))
		} else {
			// Succeed on the third attempt
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		}
	}))
	defer server.Close()

	// Create mock client
	client, mockAuth := createMockSchemaClient(server.URL)

	// Create test schema payload
	schemaPayload := []byte(`{
		"id": "retry-schema-123",
		"kind": "test"
	}`)

	// Execute test
	err := client.PutSystemSchema(schemaPayload)

	// Verify results - should succeed after retries
	assert.NoError(t, err)
	assert.Equal(t, 3, callCount) // Should have retried 3 times total
	mockAuth.AssertExpectations(t)
}

// createMockSchemaClient creates an OSDU client with a mock auth provider for schema testing
func createMockSchemaClient(schemaURL string) (osdu.OsduApiRequest, *MockAuthProvider) {
	mockAuth := &MockAuthProvider{}

	// Mock successful token response
	mockAuth.On("GetAccessToken", mock.Anything).Return(&auth.Token{
		AccessToken: "mock-access-token",
		TokenType:   "Bearer",
	}, nil)

	// Create test configuration
	osduSettings := config.OsduSettings{
		PartitionId: "test-partition",
		SchemaUrl:   schemaURL,
	}

	// Create client with mock provider and test settings
	client := osdu.NewClientWithConfig(mockAuth, osduSettings)
	return client, mockAuth
}
