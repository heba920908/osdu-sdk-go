package osdu_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/heba920908/osdu-sdk-go/pkg/auth"
	"github.com/heba920908/osdu-sdk-go/pkg/config"
	"github.com/heba920908/osdu-sdk-go/pkg/osdu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMockEntitlementsBootstrap(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    bool
	}{
		{
			name: "successful bootstrap",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/tenant-provisioning")
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Verify request body contains expected structure
				var reqBody map[string]interface{}
				json.NewDecoder(r.Body).Decode(&reqBody)
				assert.Contains(t, reqBody, "aliasMappings")

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message": "Bootstrap completed"}`))
			},
			expectError: false,
		},
		{
			name: "retry and eventual success",
			serverResponse: func() func(http.ResponseWriter, *http.Request) {
				callCount := 0
				return func(w http.ResponseWriter, r *http.Request) {
					callCount++
					if callCount < 2 {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(`{"error": "Temporary error"}`))
					} else {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"message": "Bootstrap completed"}`))
					}
				}
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			entitlementsServer := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer entitlementsServer.Close()

			// Create mock client
			client, _ := createMockClient("http://mock-partition", entitlementsServer.URL)

			// Execute test
			err := client.EntitlementsBootstrap()

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockCreateEntitlementsAdminUser(t *testing.T) {
	tests := []struct {
		name           string
		userEmail      string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    bool
	}{
		{
			name:      "successful user creation",
			userEmail: "admin@example.com",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Contains(t, r.URL.Path, "/groups/")
				assert.Contains(t, r.URL.Path, "/members")
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				// Verify request body
				var reqBody map[string]interface{}
				json.NewDecoder(r.Body).Decode(&reqBody)
				assert.Equal(t, "admin@example.com", reqBody["email"])
				assert.Equal(t, "OWNER", reqBody["role"])

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message": "User added successfully"}`))
			},
			expectError: false,
		},
		{
			name:      "invalid email format",
			userEmail: "invalid-email",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "Invalid email format"}`))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			entitlementsServer := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer entitlementsServer.Close()

			// Create mock client
			client, _ := createMockClient("http://mock-partition", entitlementsServer.URL)

			// Execute test
			err := client.EntitlementsCreateAdminUser(tt.userEmail)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockEntitlementsCreateGroup(t *testing.T) {
	tests := []struct {
		name           string
		groupId        string
		userIds        []string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    bool
	}{
		{
			name:    "successful group creation",
			groupId: "test-group",
			userIds: []string{"user1@example.com", "user2@example.com"},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/groups" {
					// Group creation request
					assert.Equal(t, "POST", r.Method)
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

					var reqBody map[string]interface{}
					json.NewDecoder(r.Body).Decode(&reqBody)
					assert.Equal(t, "test-group", reqBody["name"])
					assert.Contains(t, reqBody["description"], "Group test-group bootstrapped")

					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`{"message": "Group created successfully"}`))
				} else if r.URL.Path == "/groups/test-group@test-partition./members" {
					// Member addition request
					assert.Equal(t, "POST", r.Method)
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

					var reqBody map[string]interface{}
					json.NewDecoder(r.Body).Decode(&reqBody)
					assert.Equal(t, "OWNER", reqBody["role"])
					assert.Contains(t, []string{"user1@example.com", "user2@example.com"}, reqBody["email"])

					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`{"message": "User added successfully"}`))
				}
			},
			expectError: false,
		},
		{
			name:    "group already exists",
			groupId: "existing-group",
			userIds: []string{"user@example.com"},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/groups" {
					w.WriteHeader(http.StatusConflict)
					w.Write([]byte(`{"error": "Group already exists"}`))
				} else if r.URL.Path == "/groups/existing-group@test-partition./members" {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`{"message": "User added successfully"}`))
				}
			},
			expectError: false, // Should not error on conflict
		},
		{
			name:    "group creation error",
			groupId: "error-group",
			userIds: []string{"user@example.com"},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/groups" {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "Internal server error"}`))
				}
			},
			expectError: true,
		},
		{
			name:    "member addition error",
			groupId: "member-error-group",
			userIds: []string{"user@example.com"},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/groups" {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte(`{"message": "Group created successfully"}`))
				} else if r.URL.Path == "/groups/member-error-group@test-partition./members" {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "Failed to add member"}`))
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			entitlementsServer := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer entitlementsServer.Close()

			// Create mock client with EntitlementsDomain set
			client, _ := createMockEntitlementsClient(entitlementsServer.URL)

			// Execute test
			err := client.EntitlementsCreateGroup(tt.groupId, tt.userIds)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// createMockEntitlementsClient creates a client specifically for entitlements testing with EntitlementsDomain
func createMockEntitlementsClient(entitlementsURL string) (osdu.OsduApiRequest, *MockAuthProvider) {
	mockAuth := &MockAuthProvider{}

	// Mock successful token response
	mockAuth.On("GetAccessToken", mock.Anything).Return(&auth.Token{
		AccessToken: "mock-access-token",
		TokenType:   "Bearer",
	}, nil)

	// Create test configuration with EntitlementsDomain
	osduSettings := config.OsduSettings{
		PartitionId:        "test-partition",
		EntitlementsUrl:    entitlementsURL,
		EntitlementsDomain: "", // Empty domain for testing
	}

	authSettings := config.AuthSettings{
		InternalService: false,
	}

	// Create client with mock provider and test settings
	client := osdu.NewClientWithConfig(mockAuth, osduSettings, authSettings)
	return client, mockAuth
}
