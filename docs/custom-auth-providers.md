# Creating Custom Authentication Providers

This guide explains how to create custom authentication providers for the OSDU SDK Go.

## Overview

The OSDU SDK uses an interface-based authentication system that allows you to implement custom authentication providers. This keeps the core library lightweight while providing flexibility for different authentication methods.

## The AuthProvider Interface

All authentication providers must implement the `auth.AuthProvider` interface:

```go
type AuthProvider interface {
    // GetAccessToken retrieves an access token for the given context
    GetAccessToken(ctx context.Context) (*Token, error)

    // IsTokenValid checks if the current token is still valid
    IsTokenValid() bool

    // RefreshToken attempts to refresh the current token
    RefreshToken(ctx context.Context) (*Token, error)
}
```

## Built-in Provider

The SDK includes one built-in provider:

- **OpenID Connect / OAuth2**: Standard OAuth2 authentication (no additional dependencies)

## Example Providers

Additional authentication providers are available as examples in the `examples/` directory:

- **Azure Active Directory**: See `examples/auth_azure_mod_example/`

These examples can be copied into your project and customized as needed.

## Creating a Custom Provider

### Step 1: Define Your Provider Struct

```go
package myauth

import (
    "context"
    "github.com/heba920908/osdu-sdk-go/pkg/auth"
    "github.com/heba920908/osdu-sdk-go/pkg/config"
)

type MyCustomProvider struct {
    config       config.AuthSettings
    currentToken *auth.Token
    // Add any provider-specific fields
}
```

### Step 2: Implement the Constructor

```go
func NewMyCustomProvider(authConfig config.AuthSettings) (*MyCustomProvider, error) {
    // Validate configuration
    if authConfig.ClientId == "" {
        return nil, fmt.Errorf("client_id is required")
    }

    return &MyCustomProvider{
        config: authConfig,
    }, nil
}
```

### Step 3: Implement GetAccessToken

```go
func (p *MyCustomProvider) GetAccessToken(ctx context.Context) (*auth.Token, error) {
    // Check if we have a valid cached token
    if p.IsTokenValid() {
        return p.currentToken, nil
    }

    // Implement your token retrieval logic here
    // This could be:
    // - HTTP request to an OAuth2 endpoint
    // - SDK call to a cloud provider
    // - Reading from a file or environment variable
    // - etc.

    token := &auth.Token{
        AccessToken: "your-access-token",
        TokenType:   "Bearer",
        ExpiresAt:   time.Now().Add(1 * time.Hour),
        ExpiresIn:   3600,
    }

    p.currentToken = token
    return token, nil
}
```

### Step 4: Implement IsTokenValid

```go
func (p *MyCustomProvider) IsTokenValid() bool {
    if p.currentToken == nil {
        return false
    }
    return len(p.currentToken.AccessToken) > 0 && !p.currentToken.IsExpired()
}
```

### Step 5: Implement RefreshToken

```go
func (p *MyCustomProvider) RefreshToken(ctx context.Context) (*auth.Token, error) {
    // Implement token refresh logic
    // This could use a refresh token, or simply get a new token
    
    // For simple cases, just get a new token
    return p.GetAccessToken(ctx)
}
```

### Step 6: Use Your Provider

```go
package main

import (
    "github.com/heba920908/osdu-sdk-go/pkg/config"
    "github.com/heba920908/osdu-sdk-go/pkg/osdu"
    "yourproject/myauth"
)

func main() {
    authSettings, _ := config.GetAuthSettings()
    
    // Create your custom provider
    provider, err := myauth.NewMyCustomProvider(authSettings)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use with OSDU client
    client := osdu.NewClientWithProvider(provider)
}
```

## Real-World Examples

### Example 1: Azure Active Directory

See `examples/auth_azure_mod_example/` for a complete implementation that supports:
- Azure Managed Identity
- Service Principal authentication
- Azure CLI authentication
- OAuth2 fallback

### Example 2: Static Token Provider

For testing or simple use cases:

```go
type StaticTokenProvider struct {
    token string
}

func NewStaticTokenProvider(token string) *StaticTokenProvider {
    return &StaticTokenProvider{token: token}
}

func (p *StaticTokenProvider) GetAccessToken(ctx context.Context) (*auth.Token, error) {
    return &auth.Token{
        AccessToken: p.token,
        TokenType:   "Bearer",
        ExpiresAt:   time.Now().Add(24 * time.Hour),
        ExpiresIn:   86400,
    }, nil
}

func (p *StaticTokenProvider) IsTokenValid() bool {
    return len(p.token) > 0
}

func (p *StaticTokenProvider) RefreshToken(ctx context.Context) (*auth.Token, error) {
    return p.GetAccessToken(ctx)
}
```

### Example 3: File-Based Token Provider

Read tokens from a file that's updated by an external process:

```go
type FileTokenProvider struct {
    tokenFile string
    currentToken *auth.Token
}

func (p *FileTokenProvider) GetAccessToken(ctx context.Context) (*auth.Token, error) {
    data, err := os.ReadFile(p.tokenFile)
    if err != nil {
        return nil, fmt.Errorf("failed to read token file: %w", err)
    }

    var token auth.Token
    if err := json.Unmarshal(data, &token); err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }

    p.currentToken = &token
    return &token, nil
}
```

## Best Practices

### 1. Token Caching

Always cache tokens to avoid unnecessary authentication requests:

```go
func (p *MyProvider) GetAccessToken(ctx context.Context) (*auth.Token, error) {
    if p.IsTokenValid() {
        return p.currentToken, nil  // Return cached token
    }
    // Only fetch new token if needed
    return p.fetchNewToken(ctx)
}
```

### 2. Error Handling

Provide clear error messages:

```go
if response.StatusCode == 401 {
    return nil, fmt.Errorf("authentication failed: invalid credentials")
}
if response.StatusCode >= 500 {
    return nil, fmt.Errorf("authentication service unavailable: %d", response.StatusCode)
}
```

### 3. Context Handling

Respect context cancellation:

```go
func (p *MyProvider) GetAccessToken(ctx context.Context) (*auth.Token, error) {
    req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, body)
    // ...
}
```

### 4. Thread Safety

If your provider will be used concurrently, add mutex protection:

```go
type MyProvider struct {
    mu           sync.RWMutex
    currentToken *auth.Token
}

func (p *MyProvider) GetAccessToken(ctx context.Context) (*auth.Token, error) {
    p.mu.RLock()
    if p.IsTokenValid() {
        token := p.currentToken
        p.mu.RUnlock()
        return token, nil
    }
    p.mu.RUnlock()

    p.mu.Lock()
    defer p.mu.Unlock()
    // Fetch new token
}
```

### 5. Logging

Add appropriate logging for debugging:

```go
import "log/slog"

func (p *MyProvider) GetAccessToken(ctx context.Context) (*auth.Token, error) {
    slog.InfoContext(ctx, "Fetching new access token")
    // ...
    slog.InfoContext(ctx, "Token retrieved successfully", "expires_at", token.ExpiresAt)
}
```

## Testing Your Provider

Create tests to verify your provider implementation:

```go
func TestMyProvider_GetAccessToken(t *testing.T) {
    provider := NewMyCustomProvider(config.AuthSettings{
        ClientId: "test-client",
    })

    ctx := context.Background()
    token, err := provider.GetAccessToken(ctx)
    
    if err != nil {
        t.Fatalf("GetAccessToken failed: %v", err)
    }

    if token.AccessToken == "" {
        t.Error("Expected non-empty access token")
    }

    if !provider.IsTokenValid() {
        t.Error("Token should be valid after retrieval")
    }
}
```

## Why This Approach?

### Benefits

1. **Lightweight Core**: The main SDK doesn't include heavy dependencies
2. **Flexibility**: Implement authentication exactly as you need it
3. **Maintainability**: You control the authentication code
4. **Testability**: Easy to create mock providers for testing
5. **Future-Proof**: New authentication methods don't require SDK updates

### When to Create a Custom Provider

- You need to support a specific cloud provider (AWS, GCP, etc.)
- You have custom authentication requirements
- You want to integrate with an existing authentication system
- You need to add caching, metrics, or other custom behavior

## Additional Resources

- [OpenID Provider Implementation](../pkg/auth/openid.go) - Built-in provider source code
- [Azure Provider Example](../examples/auth_azure_mod_example/) - Complete Azure AD implementation
- [Authentication Restructure](./authentication-restructure.md) - Architecture overview

## Need Help?

If you create a useful custom provider, consider contributing it as an example to help other users!

