# Authentication System Restructure

This document outlines the new pluggable authentication system for the OSDU SDK Go.

## Overview

The authentication system has been restructured to support multiple authentication providers through a common interface. This allows for easy switching between different authentication methods like OpenID Connect/OAuth2 and Azure Active Directory.

## Architecture

```
pkg/auth/
├── interface.go    # AuthProvider interface and common types
├── openid.go      # OpenID Connect/OAuth2 implementation
├── azure.go       # Azure Active Directory implementation
└── factory.go     # Provider factory for creating auth providers
```

## Key Components

### 1. AuthProvider Interface
All authentication providers implement this interface:
- `GetAccessToken(ctx context.Context) (*Token, error)`
- `IsTokenValid() bool`
- `RefreshToken(ctx context.Context) (*Token, error)`

### 2. Provider Types
- `openid`: OpenID Connect/OAuth2 (default)
- `azure`: Azure Active Directory (requires Azure SDK)

### 3. Configuration
The configuration now includes a `provider` field to specify which authentication method to use:

```yaml
osdu:
  provider: openid  # or "azure"
  auth:
    # ... auth settings
```

## Usage Examples

### Using Factory Pattern (Recommended)
```go
// Automatically selects provider based on configuration
client := osdu.NewClient()
```

### Using Specific Provider
```go
authSettings, _ := config.GetAuthSettings()

// OpenID provider
openidProvider := auth.NewOpenIDProvider(authSettings)
client := osdu.NewClientWithProvider(openidProvider)

// Azure provider (when dependencies are available)
azureProvider, err := auth.NewAzureProvider(authSettings)
if err == nil {
    client := osdu.NewClientWithProvider(azureProvider)
}
```

## Migration Path

### Phase 1: Immediate Migration
- Old `auth.go` and `http.go` have been replaced with new pluggable system
- `NewClient()` now uses the new authentication architecture
- Existing calling code works without changes

### Phase 2: Azure Implementation
- Add Azure SDK dependencies to `go.mod`
- Enable Azure provider in factory
- Configure with `provider: azure`

## Azure Setup

To enable Azure authentication:

1. Add dependencies:
```bash
go get github.com/Azure/azure-sdk-for-go/sdk/azcore@latest
go get github.com/Azure/azure-sdk-for-go/sdk/azidentity@latest
```

2. Update configuration:
```yaml
osdu:
  provider: azure
  auth:
    clientId: "your-app-id"
    clientSecret: "your-client-secret"  # for service principal
    tenantId: "your-tenant-id"
    scopes: "https://graph.microsoft.com/.default"
```

3. Enable Azure provider in `pkg/auth/factory.go`

## Benefits

1. **Pluggable Architecture**: Easy to add new authentication providers
2. **Backward Compatibility**: Existing code continues to work
3. **Type Safety**: Interface ensures consistent behavior
4. **Testability**: Easy to mock authentication providers
5. **Configuration-Driven**: Switch providers via configuration

## Testing

```go
// Create a mock provider for testing
type MockProvider struct{}
func (m *MockProvider) GetAccessToken(ctx context.Context) (*auth.Token, error) {
    return &auth.Token{AccessToken: "mock-token"}, nil
}

// Use in tests
mockProvider := &MockProvider{}
client := osdu.NewClientWithProvider(mockProvider)
```

## Future Enhancements

- JWT token validation
- Token caching strategies
- Certificate-based authentication
- Custom token refresh logic
- Metrics and monitoring hooks
