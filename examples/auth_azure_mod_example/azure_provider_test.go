package azureprovider

import (
	"context"
	"testing"

	"github.com/heba920908/osdu-sdk-go/pkg/config"
)

func TestNewAzureProvider(t *testing.T) {
	tests := []struct {
		name     string
		config   config.AuthSettings
		wantErr  bool
		testName string
	}{
		{
			name: "valid service principal config",
			config: config.AuthSettings{
				ClientId:     "test-client-id",
				ClientSecret: "test-client-secret",
				TenantId:     "test-tenant-id",
				Scopes:       []string{"https://graph.microsoft.com/.default"},
			},
			wantErr:  false,
			testName: "Should create provider with service principal",
		},
		{
			name: "sdk auth enabled",
			config: config.AuthSettings{
				SdkAuth:  true,
				TenantId: "test-tenant-id",
				Scopes:   []string{"https://management.azure.com/.default"},
			},
			wantErr:  false,
			testName: "Should create provider with sdk auth",
		},
		{
			name: "minimal config",
			config: config.AuthSettings{
				Scopes: []string{"https://graph.microsoft.com/.default"},
			},
			wantErr:  false,
			testName: "Should create provider with minimal config",
		},
		{
			name:   "empty config with default scope",
			config: config.AuthSettings{
				// No scopes specified - should use default
			},
			wantErr:  false,
			testName: "Should create provider with default scope",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewAzureProvider(tt.config)

			if tt.wantErr && err == nil {
				t.Errorf("NewAzureProvider() expected error but got none")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("NewAzureProvider() unexpected error: %v", err)
				return
			}

			if !tt.wantErr {
				if provider == nil {
					t.Errorf("NewAzureProvider() returned nil provider")
					return
				}

				// Test token validity check (should be false for new provider)
				if provider.IsTokenValid() {
					t.Errorf("New provider should not have valid token")
				}

				// Test GetScopes method
				scopes := provider.GetScopes()
				if len(scopes) == 0 {
					t.Errorf("Provider should have at least one scope")
				}
			}
		})
	}
}

func TestAzureProviderScopes(t *testing.T) {
	tests := []struct {
		name           string
		inputScopes    []string
		expectedScopes []string
	}{
		{
			name:           "single scope",
			inputScopes:    []string{"https://graph.microsoft.com/.default"},
			expectedScopes: []string{"https://graph.microsoft.com/.default"},
		},
		{
			name:           "multiple scopes",
			inputScopes:    []string{"https://graph.microsoft.com/.default", "https://management.azure.com/.default"},
			expectedScopes: []string{"https://graph.microsoft.com/.default", "https://management.azure.com/.default"},
		},
		{
			name:           "empty scopes defaults to Microsoft Graph",
			inputScopes:    []string{},
			expectedScopes: []string{"https://graph.microsoft.com/.default"}, // default
		},
		{
			name:           "nil scopes defaults to Microsoft Graph",
			inputScopes:    nil,
			expectedScopes: []string{"https://graph.microsoft.com/.default"}, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.AuthSettings{
				Scopes: tt.inputScopes,
			}

			provider, err := NewAzureProvider(config)
			if err != nil {
				t.Fatalf("Failed to create provider: %v", err)
			}

			scopes := provider.GetScopes()
			if len(scopes) != len(tt.expectedScopes) {
				t.Errorf("Expected %d scopes, got %d", len(tt.expectedScopes), len(scopes))
			}

			for i, expectedScope := range tt.expectedScopes {
				if i >= len(scopes) || scopes[i] != expectedScope {
					t.Errorf("Expected scope[%d] = %s, got %s", i, expectedScope, scopes[i])
				}
			}
		})
	}
}

func TestAzureProviderTokenValidity(t *testing.T) {
	config := config.AuthSettings{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	}

	provider, err := NewAzureProvider(config)
	if err != nil {
		t.Fatalf("Failed to create Azure provider: %v", err)
	}

	// New provider should not have a valid token
	if provider.IsTokenValid() {
		t.Error("New provider should not have valid token")
	}

	// Test that GetAccessToken can be called (may fail without proper credentials)
	ctx := context.Background()
	_, err = provider.GetAccessToken(ctx)
	// We don't check for error here as we may not have valid credentials in test environment
	// The important thing is that it doesn't panic
}

func TestAzureProviderGetScopes(t *testing.T) {
	originalScopes := []string{"https://graph.microsoft.com/.default", "https://management.azure.com/.default"}

	config := config.AuthSettings{
		Scopes: originalScopes,
	}

	provider, err := NewAzureProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Get scopes
	scopes := provider.GetScopes()

	// Verify we got the right scopes
	if len(scopes) != len(originalScopes) {
		t.Errorf("Expected %d scopes, got %d", len(originalScopes), len(scopes))
	}

	// Verify GetScopes returns a copy (modifying returned slice shouldn't affect provider)
	if len(scopes) > 0 {
		scopes[0] = "modified"
		newScopes := provider.GetScopes()
		if newScopes[0] == "modified" {
			t.Error("GetScopes should return a copy, not the original slice")
		}
	}
}

func TestAzureProviderRefreshToken(t *testing.T) {
	config := config.AuthSettings{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	}

	provider, err := NewAzureProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	ctx := context.Background()

	// Try to refresh token (will likely fail without credentials, but shouldn't panic)
	_, err = provider.RefreshToken(ctx)
	// Error is expected in test environment without real credentials
	// The test passes if it doesn't panic
}

func TestAzureProviderWithServicePrincipal(t *testing.T) {
	config := config.AuthSettings{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		TenantId:     "test-tenant-id",
		Scopes:       []string{"https://graph.microsoft.com/.default"},
	}

	provider, err := NewAzureProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider with service principal: %v", err)
	}

	if provider == nil {
		t.Error("Provider should not be nil")
	}

	// Verify scopes were set
	scopes := provider.GetScopes()
	if len(scopes) != 1 || scopes[0] != "https://graph.microsoft.com/.default" {
		t.Errorf("Expected scope 'https://graph.microsoft.com/.default', got %v", scopes)
	}
}

func TestAzureProviderWithManagedIdentity(t *testing.T) {
	config := config.AuthSettings{
		SdkAuth:  true,
		TenantId: "test-tenant-id",
		Scopes:   []string{"https://management.azure.com/.default"},
	}

	provider, err := NewAzureProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider with managed identity: %v", err)
	}

	if provider == nil {
		t.Error("Provider should not be nil")
	}

	// Verify scopes were set
	scopes := provider.GetScopes()
	if len(scopes) != 1 || scopes[0] != "https://management.azure.com/.default" {
		t.Errorf("Expected scope 'https://management.azure.com/.default', got %v", scopes)
	}
}

// Example test showing how to mock Azure provider for testing
func ExampleNewAzureProvider() {
	// Create a provider with minimal configuration
	config := config.AuthSettings{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	}

	provider, err := NewAzureProvider(config)
	if err != nil {
		// Handle error
		return
	}

	// Check if provider was created
	if provider != nil {
		// Provider created successfully
		_ = provider.GetScopes()
	}
}

// Example test showing how to test with service principal configuration
func ExampleNewAzureProvider_servicePrincipal() {
	config := config.AuthSettings{
		ClientId:     "your-client-id",
		ClientSecret: "your-client-secret",
		TenantId:     "your-tenant-id",
		Scopes:       []string{"https://your-resource.com/.default"},
	}

	provider, err := NewAzureProvider(config)
	if err != nil {
		// Handle error
		return
	}

	// Use the provider
	ctx := context.Background()
	_, _ = provider.GetAccessToken(ctx)
}
