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

// MockWorkflowService implements the WorkflowService interface for testing
type MockWorkflowService struct {
	mock.Mock
}

func (m *MockWorkflowService) RegisterWorkflow(workflow models.RegisterWorkflow) error {
	args := m.Called(workflow)
	return args.Error(0)
}

func TestWorkflowService_RegisterWorkflow_Success(t *testing.T) {
	// Create a mock server that simulates successful workflow registration
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/workflow", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer mock-access-token")
		assert.Equal(t, "test-partition", r.Header.Get("data-partition-id"))

		// Return successful response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	// Create client and workflow service
	client := createMockWorkflowClient(server.URL)
	workflowService := client.Workflow()

	// Create test workflow
	workflow := models.RegisterWorkflow{
		WorkflowName: "test-workflow-service",
		Description:  "Test workflow via service interface",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "test-dag-service",
		},
	}

	// Execute test via interface
	err := workflowService.RegisterWorkflow(workflow)

	// Verify results
	assert.NoError(t, err)
}

func TestWorkflowService_RegisterWorkflow_ConflictHandling(t *testing.T) {
	// Create a mock server that returns conflict (workflow already exists)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error": "Workflow already exists"}`))
	}))
	defer server.Close()

	// Create client and workflow service
	client := createMockWorkflowClient(server.URL)
	workflowService := client.Workflow()

	// Create test workflow
	workflow := models.RegisterWorkflow{
		WorkflowName: "existing-workflow-service",
		Description:  "Already existing workflow via service",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "existing-dag-service",
		},
	}

	// Execute test via interface - should succeed (409 handled as "already exists")
	err := workflowService.RegisterWorkflow(workflow)

	// Verify results
	assert.NoError(t, err)
}

func TestWorkflowService_RegisterWorkflow_HttpError(t *testing.T) {
	// Create a mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	// Create client and workflow service
	client := createMockWorkflowClient(server.URL)
	workflowService := client.Workflow()

	// Create test workflow
	workflow := models.RegisterWorkflow{
		WorkflowName: "error-workflow-service",
		Description:  "Workflow that will fail via service",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "error-dag-service",
		},
	}

	// Execute test via interface
	err := workflowService.RegisterWorkflow(workflow)

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workflow service response - 500")
}

func TestWorkflowService_RegisterWorkflow_AuthenticationFailure(t *testing.T) {
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

	// Create client with failing auth provider and get workflow service
	client := osdu.NewClientWithConfig(mockAuth, osduSettings)
	workflowService := client.Workflow()

	// Create test workflow
	workflow := models.RegisterWorkflow{
		WorkflowName: "test-auth-fail-workflow",
		Description:  "Test workflow auth failure",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "auth-fail-dag",
		},
	}

	// Execute test
	err := workflowService.RegisterWorkflow(workflow)

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
	mockAuth.AssertExpectations(t)
}

func TestWorkflowService_RegisterWorkflow_WithRetry(t *testing.T) {
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

	// Create client and workflow service
	client := createMockWorkflowClient(server.URL)
	workflowService := client.Workflow()

	// Create test workflow
	workflow := models.RegisterWorkflow{
		WorkflowName: "retry-workflow-service",
		Description:  "Test workflow with retry via service",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "retry-dag-service",
		},
	}

	// Execute test
	err := workflowService.RegisterWorkflow(workflow)

	// Verify results - should succeed after retries
	assert.NoError(t, err)
	assert.Equal(t, 3, callCount) // Should have retried 3 times total
}

func TestWorkflowService_RegisterWorkflow_EmptyPayload(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client and workflow service
	client := createMockWorkflowClient(server.URL)
	workflowService := client.Workflow()

	// Create workflow with empty fields
	workflow := models.RegisterWorkflow{
		WorkflowName: "",
		Description:  "",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "",
		},
	}

	// Execute test - should work (server will validate)
	err := workflowService.RegisterWorkflow(workflow)

	// Verify results
	assert.NoError(t, err)
}

// Test the mock workflow service directly
func TestMockWorkflowService_RegisterWorkflow_Success(t *testing.T) {
	// Create mock workflow service
	mockWorkflowService := &MockWorkflowService{}

	// Setup expectations
	workflow := models.RegisterWorkflow{
		WorkflowName: "mock-test-workflow",
		Description:  "Mock test workflow",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "mock-test-dag",
		},
	}

	mockWorkflowService.On("RegisterWorkflow", workflow).Return(nil)

	// Execute test
	err := mockWorkflowService.RegisterWorkflow(workflow)

	// Verify results
	assert.NoError(t, err)
	mockWorkflowService.AssertExpectations(t)
}

func TestMockWorkflowService_RegisterWorkflow_Error(t *testing.T) {
	// Create mock workflow service
	mockWorkflowService := &MockWorkflowService{}

	// Setup expectations for error case
	workflow := models.RegisterWorkflow{
		WorkflowName: "mock-error-workflow",
		Description:  "Mock error workflow",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "mock-error-dag",
		},
	}

	expectedError := fmt.Errorf("mock registration failed")
	mockWorkflowService.On("RegisterWorkflow", workflow).Return(expectedError)

	// Execute test
	err := mockWorkflowService.RegisterWorkflow(workflow)

	// Verify results
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockWorkflowService.AssertExpectations(t)
}

// Demonstrate dependency injection pattern
func TestWorkflowService_DependencyInjection(t *testing.T) {
	// This test demonstrates how the interface can be used for dependency injection
	mockWorkflowService := &MockWorkflowService{}

	// Setup expectations
	workflow := models.RegisterWorkflow{
		WorkflowName: "dependency-injection-test",
		Description:  "Test dependency injection pattern",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "di-test-dag",
		},
	}

	mockWorkflowService.On("RegisterWorkflow", workflow).Return(nil)

	// Execute test using mock service directly (avoiding import cycle)
	err := mockWorkflowService.RegisterWorkflow(workflow)

	// Verify results
	assert.NoError(t, err)
	mockWorkflowService.AssertExpectations(t)
}

// Benchmark test for the new interface approach
func BenchmarkWorkflowService_RegisterWorkflow(b *testing.B) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	// Create client and workflow service
	client := createMockWorkflowClient(server.URL)
	workflowService := client.Workflow()

	// Create test workflow
	workflow := models.RegisterWorkflow{
		WorkflowName: "benchmark-workflow",
		Description:  "Benchmark test workflow",
		RegistrationInstructions: models.RegistrationInstructions{
			DagName: "benchmark-dag",
		},
	}

	// Run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = workflowService.RegisterWorkflow(workflow)
	}
}

// Helper function to create mock workflow client
func createMockWorkflowClient(workflowURL string) osdu.OsduApiRequest {
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

	// Create client with mock provider and test settings
	return osdu.NewClientWithConfig(mockAuth, osduSettings)
}
