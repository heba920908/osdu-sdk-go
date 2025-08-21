package osdu_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
			err := client.CreateEntitlementsAdminUser(tt.userEmail)

			// Verify results
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
