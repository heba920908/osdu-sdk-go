# Azure Authentication Provider Example

This is a standalone example implementation of Azure Active Directory authentication for the OSDU SDK Go. It demonstrates how to create a custom authentication provider that implements the `auth.AuthProvider` interface.

## Why is this a separate module?

To keep the main OSDU SDK lightweight, Azure-specific authentication is provided as an example rather than a built-in feature. This approach:

- **Reduces dependencies**: Users who don't need Azure authentication don't have to download ~50MB of Azure SDK dependencies
- **Provides flexibility**: You can customize this implementation for your specific needs
- **Demonstrates the pattern**: Shows how to implement custom authentication providers

## Features

This Azure provider supports:

- **Azure Managed Identity**: For applications running in Azure (AKS, App Service, etc.)
- **Service Principal**: Using client ID and client secret
- **Azure CLI**: For local development
- **OAuth2 Fallback**: Standard OAuth2 flow when Azure SDK is not suitable

## Usage

### Option 1: Copy into Your Project (Recommended)

Copy `azure_provider.go` into your project and customize as needed:

```bash
# Copy the provider into your project
cp azure_provider.go /path/to/your/project/internal/auth/

# Update imports in the copied file to match your project structure
```

Then use it in your code:

```go
import (
    "github.com/heba920908/osdu-sdk-go/pkg/config"
    "github.com/heba920908/osdu-sdk-go/pkg/osdu"
    "yourproject/internal/auth"  // Your copied version
)

func main() {
    authSettings, _ := config.GetAuthSettings()
    
    // Create Azure provider
    azureProvider, err := auth.NewAzureProvider(authSettings)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use with OSDU client
    client := osdu.NewClientWithProvider(azureProvider)
}
```

### Option 2: Use as a Go Module

You can also use this example directly as a module:

```bash
# In your project
go get github.com/heba920908/osdu-sdk-go/examples/auth_azure_mod_example
```

```go
import (
    azureprovider "github.com/heba920908/osdu-sdk-go/examples/auth_azure_mod_example"
    "github.com/heba920908/osdu-sdk-go/pkg/config"
    "github.com/heba920908/osdu-sdk-go/pkg/osdu"
)

func main() {
    authSettings, _ := config.GetAuthSettings()
    
    azureProvider, err := azureprovider.NewAzureProvider(authSettings)
    if err != nil {
        log.Fatal(err)
    }
    
    client := osdu.NewClientWithProvider(azureProvider)
}
```

### Option 3: Run the Example

```bash
cd examples/auth_azure_mod_example/example
go run main.go
```

## Configuration

The Azure provider uses the same configuration structure as the main SDK:

```yaml
osdu:
  provider: azure  # Note: This won't work with the main SDK anymore
  auth:
    sdk_auth: true  # Use Azure Managed Identity
    tenant_id: "your-tenant-id"
    client_id: "your-client-id"
    client_secret: "your-client-secret"
    scopes:
      - "https://your-osdu-instance.com/.default"
```

Or via environment variables:

```bash
export OSDU_AUTH_SDK_AUTH=true
export OSDU_AUTH_TENANT_ID=your-tenant-id
export OSDU_AUTH_CLIENT_ID=your-client-id
export OSDU_AUTH_CLIENT_SECRET=your-client-secret
export OSDU_AUTH_SCOPES=https://your-osdu-instance.com/.default
```

## Authentication Methods

### 1. Azure Managed Identity (Recommended for Production)

Best for applications running in Azure:

```go
authSettings := config.AuthSettings{
    SdkAuth:  true,
    TenantId: "your-tenant-id",
    Scopes:   []string{"https://your-osdu-instance.com/.default"},
}
```

### 2. Service Principal

For service-to-service authentication:

```go
authSettings := config.AuthSettings{
    ClientId:     "your-client-id",
    ClientSecret: "your-client-secret",
    TenantId:     "your-tenant-id",
    Scopes:       []string{"https://your-osdu-instance.com/.default"},
}
```

### 3. OAuth2 Flow

Standard OAuth2 when Azure SDK is not suitable:

```go
authSettings := config.AuthSettings{
    ClientId:     "your-client-id",
    ClientSecret: "your-client-secret",
    GrantType:    "client_credentials",
    TokenUrl:     "https://login.microsoftonline.com/your-tenant-id/oauth2/v2.0/token",
    Scopes:       []string{"https://your-osdu-instance.com/.default"},
}
```

## Customization Ideas

This example can be customized for your needs:

- **Add certificate-based authentication**: Use `azidentity.NewClientCertificateCredential`
- **Implement token caching**: Save tokens to disk or Redis
- **Add retry logic**: Handle transient failures
- **Custom logging**: Replace `slog` with your logging framework
- **Add metrics**: Track authentication success/failure rates
- **Multi-tenant support**: Handle multiple Azure tenants

## Dependencies

This example requires:

```
github.com/Azure/azure-sdk-for-go/sdk/azcore v1.18.2
github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.11.0
```

## Testing

The module includes comprehensive unit tests in `azure_provider_test.go` that cover:

- Provider initialization with different configurations
- Scope handling (default, single, multiple)
- Token validity checks
- Service Principal authentication setup
- Managed Identity authentication setup
- Thread safety of GetScopes()

Run the tests:

```bash
# From the example module directory
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestAzureProviderScopes
```

These tests are useful as:
- **Reference**: See how to use the provider in different scenarios
- **Validation**: Ensure your customizations don't break functionality
- **Examples**: Learn the provider's API through working code

## Related Documentation

- [Azure SDK for Go Authentication](https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication)
- [OSDU Authentication Guide](../../docs/authentication-restructure.md)
- [Creating Custom Auth Providers](../../docs/custom-auth-providers.md)

## License

This example is provided as-is for reference and can be freely modified for your needs.

