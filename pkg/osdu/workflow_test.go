package osdu_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heba920908/osdu-sdk-go/pkg/auth"
	"github.com/heba920908/osdu-sdk-go/pkg/config"
	"github.com/heba920908/osdu-sdk-go/pkg/models"
	"github.com/heba920908/osdu-sdk-go/pkg/osdu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterWorkflow_Success(t *testing.T) {
	// Create a mock server that simulates successful workflow registration
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/workflow", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer mock-access-token")
		assert.Equal(t, "test-partition", r.Header.Get("data-partition-id"))

		// Return successful response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	// Create mock client
	client, mockAuth := createMockWorkflowClient(server.URL)

	// Create test workflow
	workflow := models.RegisterWorkflow{
		WorkflowName: "test-workflow",
		Description:  "Test workflow description",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "test-dag",
		},
	}

	// Execute test
	err := client.RegisterWorkflow(workflow)

	// Verify results
	assert.NoError(t, err)
	mockAuth.AssertExpectations(t)
}

func TestRegisterWorkflow_Conflict_AlreadyRegistered(t *testing.T) {
	// Create a mock server that simulates workflow already exists (409)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return conflict response
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error": "Workflow already exists"}`))
	}))
	defer server.Close()

	// Create mock client
	client, mockAuth := createMockWorkflowClient(server.URL)

	// Create test workflow
	workflow := models.RegisterWorkflow{
		WorkflowName: "existing-workflow",
		Description:  "Already existing workflow",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "existing-dag",
		},
	}

	// Execute test - should succeed (409 is treated as "already registered")
	err := client.RegisterWorkflow(workflow)

	// Verify results - no error because 409 is handled as "already registered"
	assert.NoError(t, err)
	mockAuth.AssertExpectations(t)
}

func TestRegisterWorkflow_HttpError(t *testing.T) {
	// Create a mock server that simulates HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return error response
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	// Create mock client
	client, mockAuth := createMockWorkflowClient(server.URL)

	// Create test workflow
	workflow := models.RegisterWorkflow{
		WorkflowName: "test-workflow-error",
		Description:  "Test workflow that will fail",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "test-dag-error",
		},
	}

	// Execute test
	err := client.RegisterWorkflow(workflow)

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workflow service response - 500")
	mockAuth.AssertExpectations(t)
}

func TestRegisterWorkflow_BadRequest(t *testing.T) {
	// Create a mock server that simulates bad request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return bad request response
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Invalid workflow data"}`))
	}))
	defer server.Close()

	// Create mock client
	client, mockAuth := createMockWorkflowClient(server.URL)

	// Create test workflow
	workflow := models.RegisterWorkflow{
		WorkflowName: "invalid-workflow",
		Description:  "Invalid workflow data",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "",
		},
	}

	// Execute test
	err := client.RegisterWorkflow(workflow)

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workflow service response - 400")
	mockAuth.AssertExpectations(t)
}

func TestRegisterWorkflow_AuthenticationFailure(t *testing.T) {
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
		WorkflowUrl: server.URL,
	}

	authSettings := config.AuthSettings{
		InternalService: false,
	}

	// Create client with failing auth provider
	client := osdu.NewClientWithConfig(mockAuth, osduSettings, authSettings)

	// Create test workflow
	workflow := models.RegisterWorkflow{
		WorkflowName: "test-workflow-auth-fail",
		Description:  "Test workflow auth failure",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "test-dag-auth",
		},
	}

	// Execute test
	err := client.RegisterWorkflow(workflow)

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
	mockAuth.AssertExpectations(t)
}

func TestRegisterWorkflow_NetworkError(t *testing.T) {
	// Create mock client with invalid URL
	client, mockAuth := createMockWorkflowClient("http://invalid-url-that-should-not-exist.local")

	// Create test workflow
	workflow := models.RegisterWorkflow{
		WorkflowName: "test-workflow-network",
		Description:  "Test workflow network error",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "test-dag-network",
		},
	}

	// Execute test
	err := client.RegisterWorkflow(workflow)

	// Verify results
	assert.Error(t, err)
	// The error should be a network-related error
	assert.Contains(t, err.Error(), "no such host")
	mockAuth.AssertExpectations(t)
}

func TestRegisterWorkflow_WithRetry(t *testing.T) {
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
	client, mockAuth := createMockWorkflowClient(server.URL)

	// Create test workflow
	workflow := models.RegisterWorkflow{
		WorkflowName: "retry-workflow",
		Description:  "Test workflow with retry",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "retry-dag",
		},
	}

	// Execute test
	err := client.RegisterWorkflow(workflow)

	// Verify results - should succeed after retries
	assert.NoError(t, err)
	assert.Equal(t, 3, callCount) // Should have tried 3 times total
	mockAuth.AssertExpectations(t)
}

func TestRegisterWorkflow_EmptyWorkflowName(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create mock client
	client, mockAuth := createMockWorkflowClient(server.URL)

	// Create test workflow with empty name
	workflow := models.RegisterWorkflow{
		WorkflowName: "",
		Description:  "Workflow with empty name",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "test-dag",
		},
	}

	// Execute test - should work (server will validate)
	err := client.RegisterWorkflow(workflow)

	// Verify results
	assert.NoError(t, err)
	mockAuth.AssertExpectations(t)
}

// createMockWorkflowClient creates an OSDU client with a mock auth provider for workflow testing
func createMockWorkflowClient(workflowURL string) (osdu.OsduApiRequest, *MockAuthProvider) {
	mockAuth := &MockAuthProvider{}

	// Mock successful token response
	mockAuth.On("GetAccessToken", mock.Anything).Return(&auth.Token{
		AccessToken: "mock-access-token",
		TokenType:   "Bearer",
	}, nil)

	// Create test configuration
	osduSettings := config.OsduSettings{
		PartitionId: "test-partition",
		WorkflowUrl: workflowURL,
	}

	authSettings := config.AuthSettings{
		InternalService: false,
	}

	// Create client with mock provider and test settings
	client := osdu.NewClientWithConfig(mockAuth, osduSettings, authSettings)
	return client, mockAuth
}
