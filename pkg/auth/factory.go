package auth

import (
	"fmt"

	"github.com/heba920908/osdu-sdk-go/pkg/config"
)

// ProviderFactory creates authentication providers based on configuration
type ProviderFactory struct{}

// NewProviderFactory creates a new provider factory
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// CreateProvider creates an appropriate authentication provider based on the provider type
func (f *ProviderFactory) CreateProvider(providerType ProviderType, authConfig config.AuthSettings) (AuthProvider, error) {
	switch providerType {
	case ProviderTypeOpenID:
		return NewOpenIDProvider(authConfig), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s. Note: Azure provider has been moved to examples/auth_azure_mod_example - see examples for custom provider implementation", providerType)
	}
}

// GetProviderFromConfig determines the provider type from configuration and creates the appropriate provider
func (f *ProviderFactory) GetProviderFromConfig(osduConfig config.OsduClient) (AuthProvider, error) {
	providerType := ProviderType(osduConfig.Provider)
	if providerType == "" {
		providerType = ProviderTypeOpenID // Default to OpenID
	}

	return f.CreateProvider(providerType, osduConfig.AuthSettings)
}
