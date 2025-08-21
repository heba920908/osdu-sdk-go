package auth_test

import (
	"context"
	"testing"

	"github.com/heba920908/osdu-sdk-go/pkg/auth"
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
				Scopes:       "https://graph.microsoft.com/.default",
			},
			wantErr:  false,
			testName: "Should create provider with service principal",
		},
		{
			name: "pod auth enabled",
			config: config.AuthSettings{
				PodAuthEnabled: true,
				TenantId:       "test-tenant-id",
				Scopes:         "https://management.azure.com/.default",
			},
			wantErr:  false,
			testName: "Should create provider with pod auth",
		},
		{
			name: "minimal config",
			config: config.AuthSettings{
				Scopes: "https://graph.microsoft.com/.default",
			},
			wantErr:  false,
			testName: "Should create provider with minimal config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := auth.NewAzureProvider(tt.config)

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

				// Test that the provider implements the interface
				var _ auth.AuthProvider = provider

				// Test token validity check (should be false for new provider)
				if provider.IsTokenValid() {
					t.Errorf("New provider should not have valid token")
				}
			}
		})
	}
}

func TestAzureProviderInterface(t *testing.T) {
	// Test that AzureProvider implements AuthProvider interface
	config := config.AuthSettings{
		Scopes: "https://graph.microsoft.com/.default",
	}

	provider, err := auth.NewAzureProvider(config)
	if err != nil {
		t.Fatalf("Failed to create Azure provider: %v", err)
	}

	// Verify interface methods exist
	ctx := context.Background()

	// These will fail without proper credentials, but should not panic
	_, err = provider.GetAccessToken(ctx)
	if err == nil {
		t.Log("GetAccessToken succeeded (unexpected but ok)")
	}

	valid := provider.IsTokenValid()
	if valid {
		t.Error("New provider should not have valid token")
	}

	_, err = provider.RefreshToken(ctx)
	if err == nil {
		t.Log("RefreshToken succeeded (unexpected but ok)")
	}
}

func TestAzureProviderScopes(t *testing.T) {
	tests := []struct {
		name           string
		inputScopes    string
		expectedScopes []string
	}{
		{
			name:           "single scope",
			inputScopes:    "https://graph.microsoft.com/.default",
			expectedScopes: []string{"https://graph.microsoft.com/.default"},
		},
		{
			name:           "multiple scopes",
			inputScopes:    "https://graph.microsoft.com/.default https://management.azure.com/.default",
			expectedScopes: []string{"https://graph.microsoft.com/.default", "https://management.azure.com/.default"},
		},
		{
			name:           "empty scopes",
			inputScopes:    "",
			expectedScopes: []string{"https://graph.microsoft.com/.default"}, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.AuthSettings{
				Scopes: tt.inputScopes,
			}

			provider, err := auth.NewAzureProvider(config)
			if err != nil {
				t.Fatalf("Failed to create provider: %v", err)
			}

			if len(provider.GetScopes()) != len(tt.expectedScopes) {
				t.Errorf("Expected %d scopes, got %d", len(tt.expectedScopes), len(provider.GetScopes()))
			}

			scopes := provider.GetScopes()
			for i, expectedScope := range tt.expectedScopes {
				if i >= len(scopes) || scopes[i] != expectedScope {
					t.Errorf("Expected scope[%d] = %s, got %s", i, expectedScope, scopes[i])
				}
			}
		})
	}
}
